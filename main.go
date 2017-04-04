package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"

	"github.com/davinche/goviz/server"
	"github.com/urfave/cli"
)

var VERSION string
var port int
var id string

func main() {
	app := cli.NewApp()
	app.Name = "GoVIZ"
	app.Version = VERSION
	app.Usage = "Live Previewer for GraphViz"
	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:        "port, p",
			Value:       1338,
			Usage:       "the port for the preview server",
			Destination: &port,
		},

		cli.StringFlag{
			Name:        "id, i",
			Value:       "",
			Usage:       "the port for the preview server",
			Destination: &id,
		},
	}

	app.Commands = []cli.Command{
		cli.Command{
			Name:        "start",
			Description: "start the preview server",
			Action:      start,
		},

		cli.Command{
			Name:        "send",
			Description: "send graphviz dot data to preview server",
			Action:      send,
		},
	}

	app.Run(os.Args)
}

func start(c *cli.Context) error {
	server.ListenAndServe(":" + strconv.Itoa(port))
	return nil
}

func send(c *cli.Context) error {
	if id == "" {
		cli.ShowAppHelp(c)
		return nil
	}
	command := exec.Command("dot", "-T", "png")
	command.Stdin = os.Stdin
	data, err := command.Output()
	if err != nil {
		fmt.Printf("error: could not generate image\n")
		os.Exit(1)
		return err
	}
	return sendData(data)
}

func sendData(data []byte) error {
	b := bytes.NewBuffer(data)
	c := http.Client{}
	r, err := http.NewRequest("POST", "http://localhost:"+strconv.Itoa(port)+"/generate?id="+id, b)
	if err != nil {
		fmt.Printf("error: could not create http request: err=%q\n", err)
		return err
	}
	r.Header.Set("Content-Type", "application/octet-stream")
	_, err = c.Do(r)
	if err != nil {
		fmt.Printf("error: could send http request; err=%q\n", err)
		return err
	}
	return nil
}
