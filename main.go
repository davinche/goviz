package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"

	"github.com/davinche/goviz/server"
	"github.com/urfave/cli"
)

var VERSION string
var port int
var id string
var shouldLaunch bool

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
		cli.BoolFlag{
			Name:        "launch, l",
			Usage:       "flag for launching in browser",
			Destination: &shouldLaunch,
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
		cli.Command{
			Name:        "shutdown",
			Description: "shutdown the preview server",
			Action:      shutdown,
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

func shutdown(c *cli.Context) error {
	client := http.Client{}
	r, err := http.NewRequest("POST", "http://localhost:"+strconv.Itoa(port)+"/shutdown", nil)
	if err != nil {
		fmt.Printf("error: could not create http request: err=%q\n", err)
		return err
	}
	client.Do(r)
	return nil
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

	if shouldLaunch {
		launchBrowser(id)
	}
	return nil
}

func launchBrowser(id string) {
	var args []string
	switch runtime.GOOS {
	case "darwin":
		args = append(args, "open", "-g")
		break
	case "linux":
		args = append(args, "xdg-open")
		break
	}

	if len(args) == 0 {
		log.Println("error: could not determine how to launch browser")
		os.Exit(1)
	}
	args = append(args, "http://localhost:"+strconv.Itoa(port)+"?id="+id)
	command := exec.Command(args[0], args[1:]...)
	if err := command.Start(); err != nil {
		log.Printf("error: could not launch browser: %v\n", err)
	}
}
