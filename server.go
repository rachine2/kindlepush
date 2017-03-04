package kindlepush

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"

	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/go-gomail/gomail"
	objectid "github.com/zhengchun/go-objectid"
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
			f1, f2 := createHtmlFile(post)
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
	s.startPush()
	return nil
}

func NewServer(cfg Config) *Server {
	return &Server{
		config: cfg,
	}
}
