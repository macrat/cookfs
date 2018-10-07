package main

import (
	"time"
	"net/url"
	"fmt"
	"os"
	"math/rand"
	"encoding/json"
	"context"

	"github.com/alecthomas/kingpin"
	"github.com/go-yaml/yaml"

	"github.com/macrat/cookfs/cooklib"
	"github.com/macrat/cookfs/plugins"
)

func Request(servers []*cooklib.Node, path string, data interface{}) cooklib.Response {
	handler := &plugins.HTTPHandler{}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	resp := make(chan cooklib.Response)

	for _, server := range servers {
		go func(server *cooklib.Node) {
			r := handler.Send(ctx, cooklib.Request{server, path, data})
			if r.StatusCode == 200 || r.StatusCode == 204 {
				resp <- r
				cancel()
			}
		}(server)
	}

	select {
	case r := <-resp:
		return r

	case <-ctx.Done():
		return cooklib.Response{StatusCode: 502}
	}
}

func Info(servers []*cooklib.Node, format string) error {
	resp := Request(servers, "/term", nil)
	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		return fmt.Errorf("failed to request: %d", resp.StatusCode)
	}

	if format == "yaml" {
		y, _ := yaml.Marshal(resp.Data)
		fmt.Println(string(y))
	} else {
		j, _ := json.Marshal(resp.Data)
		fmt.Println(string(j))
	}

	return nil
}

func Upload(servers []*cooklib.Node, tag string, file *os.File) error {
	return fmt.Errorf("not implemented")
}

func Download(servers []*cooklib.Node, tag string, file *os.File) error {
	return fmt.Errorf("not implemented")
}

func ConvertServers(servers []*url.URL) []*cooklib.Node {
	r := make([]*cooklib.Node, 0, len(servers))

	for _, s := range servers {
		r = append(r, (*cooklib.Node)(s))
	}

	return r
}

func main() {
	rand.Seed(time.Now().Unix())

	server := kingpin.Flag("server", "Server address.").Default("http://localhost:5790").URLList()

	infoCommand := kingpin.Command("info", "Get server information.")
	infoFormat := infoCommand.Flag("format", "Output format. yaml or json.").Default("yaml").Enum("yaml", "json")
	infoCommand.Action(func(c *kingpin.ParseContext) error {
		return Info(ConvertServers(*server), *infoFormat)
	})

	uploadCommand := kingpin.Command("upload", "Upload file.")
	uploadTag := uploadCommand.Arg("tag", "Tag name.").Required().String()
	uploadFile := uploadCommand.Arg("file", "File name.").File()
	uploadCommand.Action(func(c *kingpin.ParseContext) error {
		return Upload(ConvertServers(*server), *uploadTag, *uploadFile)
	})


	downloadCommand := kingpin.Command("download", "Download file.")
	downloadTag := downloadCommand.Arg("tag", "Tag name.").Required().String()
	downloadFile := downloadCommand.Arg("file", "File name.").File()
	downloadCommand.Action(func(c *kingpin.ParseContext) error {
		return Download(ConvertServers(*server), *downloadTag, *downloadFile)
	})

	kingpin.Parse()
}
