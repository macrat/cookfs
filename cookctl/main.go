package main

import (
	"time"
	"bytes"
	"path"
	"net/url"
	"fmt"
	"os"
	"net/http"
	"math/rand"
	"encoding/json"

	"github.com/alecthomas/kingpin"
	"github.com/go-yaml/yaml"
	"github.com/macrat/cookfs/cookfs"
)

func Request(servers []*url.URL, method, endpoint string, data interface{}) (*http.Response, error) {
	var err error

	var body []byte
	if data != nil {
		body, err = json.Marshal(data)
		if err != nil {
			return nil, err
		}
	}

	used := []int{}

	for  {
		var i int
		for {
			i = rand.Intn(len(servers))
			ok := true
			for _, x := range used {
				if i == x {
					ok = false
					break
				}
			}
			if ok {
				break
			}
		}

		used = append(used, i)
		server := *servers[i]
		server.Path = path.Join(server.Path, endpoint)

		req, err := http.NewRequest(method, (&server).String(), bytes.NewReader(body))
		if err != nil {
			return nil, err
		}

		var resp *http.Response
		resp, err = (&http.Client{}).Do(req)
		if err == nil {
			return resp, nil
		}

		if len(used) == len(servers) {
			return nil, err
		}
	}
}

func Info(servers []*url.URL) error {
	resp, err := Request(servers, "GET", "/leader", nil)
	if err != nil {
		return err
	}

	var status cookfs.TermStatus
	json.NewDecoder(resp.Body).Decode(&status)

	y, _ := yaml.Marshal(status)
	fmt.Println(string(y))

	return nil
}

func Upload(servers []*url.URL, tag string, file *os.File) error {
	return fmt.Errorf("not implemented")
}

func Download(servers []*url.URL, tag string, file *os.File) error {
	return fmt.Errorf("not implemented")
}

func main() {
	rand.Seed(time.Now().Unix())

	server := kingpin.Flag("server", "Server address.").Default("http://localhost:8080").URLList()

	kingpin.Command("info", "Get server information.").Action(func(c *kingpin.ParseContext) error {
		return Info(*server)
	})

	uploadCommand := kingpin.Command("upload", "Upload file.")
	uploadTag := uploadCommand.Arg("tag", "Tag name.").Required().String()
	uploadFile := uploadCommand.Arg("file", "File name.").File()
	uploadCommand.Action(func(c *kingpin.ParseContext) error {
		return Upload(*server, *uploadTag, *uploadFile)
	})


	downloadCommand := kingpin.Command("download", "Download file.")
	downloadTag := downloadCommand.Arg("tag", "Tag name.").Required().String()
	downloadFile := downloadCommand.Arg("file", "File name.").File()
	downloadCommand.Action(func(c *kingpin.ParseContext) error {
		return Download(*server, *downloadTag, *downloadFile)
	})

	kingpin.Parse()
}
