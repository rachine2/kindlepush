package kindlepush

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"text/template"
	"time"

	"github.com/antchfx/xquery/html"
	"github.com/nfnt/resize"
	"github.com/zhengchun/objectid"
	"golang.org/x/net/html"
	"golang.org/x/text/encoding/unicode"
)

var (
	htmltmpl, _     = template.New("html").Parse(htmlTemplateStr)
	contentstmpl, _ = template.New("contents").Parse(contentsTemplateStr)
	ncxtmpl, _      = template.New("ncx").Parse(ncxTemplateStr)
	opftmpl, _      = template.New("opf").Parse(opfTemplateStr)
)

type fileInfo struct {
	id   string
	name string
	typ  string
}

var allowedTags = map[string]bool{
	"body": true,
	"div":  true,
	"img":  true,
	"p":    true,
	"br":   true,
	"hr":   true,
	"a":    true,
	"b":    true,
	"font": true,
	"h1":   true,
	"h2":   true,
	"h3":   true,
}

var allowedAttrs = map[string]bool{
	"src":  true,
	"size": true,
	"href": true,
}

func isSelfClosingTag(t string) bool {
	switch t {
	case "hr", "img", "br":
		return true
	default:
		return false
	}
}

var seqNum int32

func sequence() string {
	return fmt.Sprintf("%d", atomic.AddInt32(&seqNum, 1))
}

func createHtmlFile(post *Post) (htmlFile *fileInfo, imageFiles []*fileInfo) {
	doc, err := html.Parse(strings.NewReader(post.Body))
	if err != nil {
		return
	}
	body := htmlquery.FindOne(doc, "//body")
	// iteration all elements of HTML document.
	var fn func(*bytes.Buffer, *html.Node, bool)
	fn = func(buf *bytes.Buffer, n *html.Node, includeSelf bool) {
		if n == nil {
			return
		}
		if n.Type == html.TextNode || n.Type == html.CommentNode {
			buf.WriteString(strings.TrimSpace(n.Data))
			return
		}
		if ok := allowedTags[n.Data]; !ok {
			return
		}

		selfClosing := isSelfClosingTag(n.Data)
		if includeSelf {
			buf.WriteString("<" + n.Data)
			for _, attr := range n.Attr {
				if ok := allowedAttrs[attr.Key]; ok {
					val := attr.Val

					// If an element is `src` element that means need
					// download this image from remote server.
					if n.Data == "img" && attr.Key == "src" {
						if f, err := downloadImage(val); err == nil {
							imageFiles = append(imageFiles, f)
							val = f.name
						}
					}

					if val != "" {
						buf.WriteString(fmt.Sprintf(` %s="%s"`, attr.Key, val))
					}
				}
			}

			if selfClosing {
				buf.WriteString("/>")
			} else {
				buf.WriteString(">")
			}
		}

		for child := n.FirstChild; child != nil; child = child.NextSibling {
			fn(buf, child, true)
		}
		if includeSelf {
			if !selfClosing {
				buf.WriteString(fmt.Sprintf("</%s>", n.Data))
			}
		}

	}
	var buf bytes.Buffer
	fn(&buf, body, false)

	fname := filepath.Join(os.TempDir(), fmt.Sprintf("%s.html", objectid.New()))
	if f, err := os.Create(fname); err == nil {
		defer f.Close()
		w := unicode.UTF8.NewEncoder().Writer(f)
		// Passed data to template file.
		type Model struct {
			Title       string
			Author      string
			Body        string
			Description string
			Date        string
			Link        string
		}

		err := htmltmpl.Execute(w, &Model{
			Title:       post.Title,
			Author:      post.Author,
			Body:        buf.String(),
			Description: post.FormatDescription(250),
			Date:        post.FormatDate(),
			Link:        post.Link,
		})
		if err != nil {
			panic(err)
		}
		htmlFile = &fileInfo{
			id:   sequence(),
			name: f.Name(),
			typ:  "application/xhtml+xml",
		}
	}
	return
}

// downloadImage downloads an image via HTTP and returns local file path.
var client = &http.Client{
	Timeout: 15 * time.Second,
}

func downloadImage(link string) (*fileInfo, error) {
	fmt.Println(link)
	resp, err := client.Get(link)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// resize image
	img, typ, err := image.Decode(resp.Body)
	if err != nil {
		return nil, err
	}

	const (
		maxWidth  = uint(500)
		maxHeight = uint(600)
	)
	img = resize.Thumbnail(maxWidth, maxHeight, img, resize.NearestNeighbor)

	ext := ".jpg"
	switch typ {
	case "gif":
		ext = ".gif"
	case "png":
		ext = ".png"
	}

	name := objectid.New().String() + ext
	f, err := os.Create(filepath.Join(os.TempDir(), name))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	switch typ {
	case "gif":
		if err := gif.Encode(f, img, &gif.Options{NumColors: 256}); err != nil {
			return nil, err
		}
	case "png":
		if err := png.Encode(f, img); err != nil {
			return nil, err
		}
	default: // jpeg or other types
		if err := jpeg.Encode(f, img, &jpeg.Options{Quality: jpeg.DefaultQuality}); err != nil {
			return nil, err
		}
	}

	return &fileInfo{
		id:   sequence(),
		name: f.Name(),
		typ:  typ,
	}, nil
}

func createContentsFile(data map[string]map[string]string) *fileInfo {
	f, err := os.Create(filepath.Join(os.TempDir(), "contents.html"))
	if err != nil {
		return nil
	}
	defer f.Close()

	w := unicode.UTF8.NewEncoder().Writer(f)
	err = contentstmpl.Execute(w, struct {
		Title    string
		Sections map[string]map[string]string
	}{
		Title:    "Table of Contents",
		Sections: data,
	})
	if err != nil {
		panic(err)
	}
	return &fileInfo{id: "contents", name: f.Name(), typ: "application/xhtml+xml"}
}

func createNcxFile(channels []*channel, postFileMaps map[*Post]*fileInfo) *fileInfo {
	type Item struct {
		Id          string
		Link        string
		Author      string
		Description string
		Title       string
	}
	type Section struct {
		Id    int
		Name  string
		Items []Item
	}

	var sections []*Section
	for i, channel := range channels {
		sect := &Section{Id: i, Name: channel.name, Items: make([]Item, 0)}
		sections = append(sections, sect)
		for _, post := range channel.posts {
			f := postFileMaps[post]
			item := Item{
				Id:          f.id,
				Link:        f.name,
				Author:      post.Author,
				Description: post.FormatDescription(250),
				Title:       post.Title,
			}
			sect.Items = append(sect.Items, item)
		}
	}

	f, err := os.Create(filepath.Join(os.TempDir(), "contents.ncx"))
	if err != nil {
		return nil
	}
	defer f.Close()
	w := unicode.UTF8.NewEncoder().Writer(f)
	err = ncxtmpl.Execute(w, struct {
		Date     string
		Sections []*Section
	}{
		Date:     time.Now().Format("2006-01-02"),
		Sections: sections,
	})
	if err != nil {
		panic(err)
	}
	return &fileInfo{id: "nav-contents", name: f.Name(), typ: "application/x-dtbncx+xml"}
}

func createOpfFile(files []*fileInfo) *fileInfo {
	type File struct {
		Id    string
		Type  string
		Name  string
		IsRef bool
	}

	var list []File
	for _, f := range files {
		ref := false
		if f.typ == "application/xhtml+xml" {
			ref = true
		}
		list = append(list, File{
			Id:    f.id,
			Type:  f.typ,
			Name:  f.name,
			IsRef: ref,
		})
	}

	f, err := os.Create(filepath.Join(os.TempDir(), fmt.Sprintf("%s.opf", objectid.New())))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	w := unicode.UTF8.NewEncoder().Writer(f)
	err = opftmpl.Execute(w, struct {
		Date  string
		Files []File
	}{
		Date:  time.Now().Format("2006-01-02"),
		Files: list,
	})
	if err != nil {
		panic(err)
	}
	return &fileInfo{id: sequence(), name: f.Name(), typ: ""}
}

const (
	htmlTemplateStr = `<html lang="en" xmlns="http://www.w3.org/1999/xhtml" xml:lang="en">
<head>
	<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
	<title>{{.Title}}</title>
	<meta content="{{.Author}}" name="author" />
	<meta content="{{.Description}}" name="description" />
</head>
<body>
	<h1><b>{{.Title}}</b></h1>
	<p>{{.Author}} - {{.Date}}</p>
	<p>{{.Body}}</p>
	<p>——————</p>
	<p height="1em" width="0">{{.Link}}</p>
	</body>
</html>`

	contentsTemplateStr = `<html lang="en" xmlns="http://www.w3.org/1999/xhtml" xml:lang="en">
	<head>
		<meta content="text/html; charset=utf-8" http-equiv="Content-Type"/>
		<title>{{.Title}}</title>
	</head>
	<body>
		<h1>{{.Title}}</h1>
        {{ range $name,$links := .Sections }}
		<h4>{{ $name }}</h4>
		<ul>
            {{ range $href,$text:= $links}}
			<li><a href="{{ $href }}">{{ $text }}</a></li>
            {{end}}
        </ul>
		{{ end }}
	</body>
</html>`

	ncxTemplateStr = `<?xml version='1.0' encoding='utf-8'?>
<!DOCTYPE ncx PUBLIC "-//NISO//DTD ncx 2005-1//EN" "http://www.daisy.org/z3986/2005/ncx-2005-1.dtd">
<ncx xmlns:mbp="http://mobipocket.com/ns/mbp" xmlns="http://www.daisy.org/z3986/2005/ncx/" version="2005-1" xml:lang="en-GB">
	<head>
		<meta content="KindlePush-{{.Date}}" name="dtb:uid" />
		<meta content="2" name="dtb:depth" />
		<meta content="0" name="dtb:totalPageCount" />
		<meta content="0" name="dtb:maxPageNumber" />
	</head>	
	<docTitle>
		<text>KindlePush-{{.Date}}</text>
	</docTitle>
	<docAuthor>
		<text>KindlePush</text>
	</docAuthor>	
	<navMap>
		<navPoint playOrder="0" class="periodical" id="periodical">
			<navLabel>
				<text>Table of Contents</text>
			</navLabel>
			<content src="contents.html" />

			{{ range $num,$section:=.Sections}}
			<navPoint playOrder="{{ $section.Id }}" class="section" id="s-{{$num}}">				
				<navLabel>
					<text>{{$section.Name}}</text>
				</navLabel>						
				<content src="{{  (index $section.Items 0).Link }}" />
				{{ range $num,$item:=$section.Items }}
				<navPoint playOrder="{{$item.Id}}" class="article" id="item-{{$num}}">					
					<navLabel>
						<text>{{$item.Title}}</text>
					</navLabel>					
					<content src="{{$item.Link}}" />					
					<mbp:meta name="description">{{$item.Description}}</mbp:meta>					
					<mbp:meta name="author">{{$item.Author}}</mbp:meta>
				</navPoint>				
				{{end}}
			</navPoint>
			{{end}}
		</navPoint>
	</navMap>
</ncx>`

	opfTemplateStr = `<?xml version='1.0' encoding='utf-8'?>
<package xmlns="http://www.idpf.org/2007/opf" version="2.0" unique-identifier="KindleFere_2010-10-15">
<metadata>
	<dc-metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
		<dc:title>KindlePush</dc:title>
		<dc:language>en-us</dc:language>
		<dc:Identifier id="uid">02FFA518EB</dc:Identifier>
		<dc:creator>KindlePush</dc:creator>
		<dc:publisher>KindlePush</dc:publisher> 
		<dc:date>{{.Date}}</dc:date>		
	</dc-metadata>
	<x-metadata>
		<output content-type="application/x-mobipocket-subscription-magazine" encoding="utf-8"/>
	</x-metadata>
</metadata>
<manifest>
	{{range $,$file:= .Files }}
	<item href="{{$file.Name}}" media-type="{{$file.Type}}" id="{{$file.Id}}"/>
	{{end}}
</manifest>
<spine toc="nav-contents">
	<itemref idref="contents"/>
	{{range $,$file:= .Files }}
		{{if $file.IsRef}}
		<itemref idref="{{$file.Id}}"/>
		{{end}}
	{{end}}	
</spine>
<guide>
	<reference href="contents.html" type="toc" title="Table of Contents" />	
</guide>
</package>`
)
