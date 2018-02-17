package kindlepush

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"sync"

	"io/ioutil"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/go-gomail/gomail"
	"github.com/zhengchun/objectid"
)

type Server struct {
	config Config
}

func getPlugins(id []string) []pluginEntry {
	var p []pluginEntry
	for _, plugin := range allPlugins {
		for _, v := range id {
			if v == plugin.id {
				p = append(p, plugin)
			}
		}
	}
	return p
}

func (s *Server) startPush() {
	logrus.Info("begin push service")

	plugins := getPlugins(s.config.Subscribes)
	if len(plugins) == 0 {
		return
	}

	var (
		wg       sync.WaitGroup
		channels []*channel
	)
	wg.Add(len(plugins))
	for _, plugin := range plugins {
		c := &channel{
			name: plugin.name,
		}
		go func(c *channel, plugin Plugin) {
			defer wg.Done()

			posts, err := plugin.FetchFeed()
			if err != nil {
				logrus.Errorf("%s rss feed update got error: %v", c.name, err)
				return
			}
			c.posts = posts
			if len(c.posts) > s.config.MaxNumber {
				c.posts = c.posts[:s.config.MaxNumber]
			}
			logrus.Infof("%s rss feed update completed", c.name)
			//
			if len(c.posts) > 0 {
				channels = append(channels, c)
			}
		}(c, plugin.handler)
	}
	wg.Wait()

	var files []*fileInfo
	defer func() {
		if !s.config.Debug {
			for _, f := range files {
				os.Remove(f.name)
			}
		}
	}()

	sections := make(map[string]map[string]string) // name:[filepath:title]
	postFileMaps := make(map[*Post]*fileInfo)

	for _, channel := range channels {
		sections[channel.name] = make(map[string]string)
		for _, post := range channel.posts {
			f1, f2 := createHtmlFile(post, "", s.config.ImageKeepSize)
			files = append(files, f1)
			files = append(files, f2...)
			postFileMaps[post] = f1
			sections[channel.name][f1.name] = post.Title
		}
	}
	// contents.html
	files = append(files, createContentsFile(sections))
	// ncx file
	files = append(files, createNcxFile(channels, postFileMaps))

	s.createMobi(files)
}

func getFilelist(check_path string) []string {
	var fileList []string
	err := filepath.Walk(check_path, func(check_path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		ext := path.Ext(f.Name())
		if (ext == ".htm") || (ext == ".html") {
			fileList = append(fileList, check_path)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("filepath.Walk() returned %v\n", err)
	}
	return fileList
}

// TODO, no accurate formula
func (s *Server) convertLocal2MobiSize(localSize int64) int64 {
	return localSize*4 + 3*1024*1024
}

func (s *Server) getFileSize(fileName string) int64 {
	file, err := os.Open(fileName)
	defer file.Close()
	if err != nil {
		logrus.Fatal(err)
		return 0
	}
	fi, err := file.Stat()
	if err != nil {
		logrus.Fatal(err)
		return 0
	}
	return fi.Size()
}

func (s *Server) exceedFileSizeLimit(totalSize int64, htmlFile *fileInfo, imageFiles []*fileInfo) bool {
	totalSize += s.getFileSize(htmlFile.name)
	for _, img := range imageFiles {
		totalSize += s.getFileSize(img.name)
	}

	if s.convertLocal2MobiSize(totalSize) > s.config.SizeLimit {
		return true
	}

	return false
}

func (s *Server) pushHtml(htmlPath string) {
	sections := make(map[string]map[string]string) // name:[filepath:title]
	postFileMaps := make(map[*Post]*fileInfo)

	var channels []*channel
	var files []*fileInfo
	var totalFileSize int64 = 0

	htmlFileList := getFilelist(htmlPath)
	for _, fileName := range htmlFileList {
		sections[path.Base(s.config.Htmlpath)] = make(map[string]string)

		htmlContent, err := ioutil.ReadFile(fileName)
		if err != nil {
			logrus.Fatal(err)
		}

		content := string(htmlContent)

		baseName := path.Base(fileName)
		post := &Post{
			Title: baseName,
			Body:  content,
		}
		f1, f2 := createHtmlFile(post, htmlPath, s.config.ImageKeepSize)

		if s.exceedFileSizeLimit(totalFileSize, f1, f2) {
			// contents.html
			files = append(files, createContentsFile(sections))
			// ncx file
			files = append(files, createNcxFile(channels, postFileMaps))

			s.createMobi(files)

			// clear works
			sections = make(map[string]map[string]string) // name:[filepath:title]
			sections[path.Base(s.config.Htmlpath)] = make(map[string]string)
			postFileMaps = make(map[*Post]*fileInfo)
			totalFileSize = 0
			channels = channels[:0]
			files = files[:0]
		}
		totalFileSize += s.getFileSize(f1.name)
		for _, img := range f2 {
			totalFileSize += s.getFileSize(img.name)
		}

		files = append(files, f1)
		files = append(files, f2...)
		postFileMaps[post] = f1
		sections["html"][f1.name] = baseName

		//postList := [...]*Post{post}
		//postList := [...]*Post{post}
		var postList []*Post
		postList = append(postList, post)
		channelNode := &channel{
			name:  baseName,
			posts: postList,
		}
		channels = append(channels, channelNode)
	}
	// contents.html
	files = append(files, createContentsFile(sections))
	// ncx file
	files = append(files, createNcxFile(channels, postFileMaps))

	s.createMobi(files)
}

func (s *Server) createMobi(files []*fileInfo) {
	// opf file
	opfFile := createOpfFile(files)
	files = append(files, opfFile)

	mobiName := fmt.Sprintf("%s.mobi", objectid.New())
	cmd := exec.Command("kindlegen", opfFile.name, "-o", mobiName)
	cmd.Run()

	// checking a file exists when after  kindlegen run
	mobiName = filepath.Join(filepath.Dir(opfFile.name), mobiName)
	if _, err := os.Stat(mobiName); os.IsNotExist(err) {
		logrus.Warn("checking a kindlegen program whether exists")
		return
	}
	logrus.Infof("%s created successfully", mobiName)
	send(s.config.KindleAddr, s.config.Email, mobiName)
}

func send(to string, email EmailConfig, attachFile string) {
	logrus.Info("sending file via email")
	m := gomail.NewMessage()
	m.SetHeader("From", email.From)
	m.SetHeader("To", to)
	m.SetBody("text/html", "Hello KindlePush")
	m.Attach(attachFile)

	host := email.SMTP
	port := 25
	if i := strings.Index(host, ":"); i > 0 {
		h, p, err := net.SplitHostPort(host)
		if err != nil {
			panic(err)
		}
		host = h
		p2, _ := strconv.ParseInt(p, 10, 64)
		port = int(p2)
	}

	d := gomail.NewDialer(host, port, email.Username, email.Password)
	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		logrus.Error(err)
		return
	}
	logrus.Info("send completed, you can see it on your kindle device")
}

func (s *Server) Run() error {
	if len(s.config.Htmlpath) == 0 {
		s.startPush()
	} else {
		s.pushHtml(s.config.Htmlpath)
	}
	return nil
}

func NewServer(cfg Config) *Server {
	return &Server{
		config: cfg,
	}
}
