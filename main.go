package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/urfave/cli.v1"
)

var args = make([]string, 5, 5)
var currentPath = ""
var port = ":8080"
var dir http.Dir
var ssize = false
var modt = false
var full = false
var er error

// Fily struct to hold single file
type Fily struct {
	filepath, filename string
	size               int64
	moddate            time.Time
	isdir              bool
}

func main() {
	// app params
	currentPath, err := os.Getwd()
	dir = http.Dir(currentPath)
	//file params

	//prefiks := http.StripPrefix("/static/", http.FileServer(dir))

	if err != nil {
		er = err
		log.Println(err)
	}

	//http.Handle("/static/", prefiks)

	app := cli.NewApp()
	app.Name = "sfl"
	app.Usage = "Serve any folder to local network"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "directory, d",
			Usage:       "Path to folder to serve to",
			EnvVar:      "",
			Hidden:      false,
			Value:       "./",
			Destination: new(string),
		},
		cli.StringFlag{
			Name:        "port, p",
			Usage:       "Port to open",
			EnvVar:      "",
			Hidden:      false,
			Value:       ":8080",
			Destination: new(string),
		},
		cli.BoolFlag{
			Name:        "size, s",
			Usage:       "Show the sizes of files",
			EnvVar:      "",
			Hidden:      false,
			Destination: new(bool),
		},
		cli.BoolFlag{
			Name:        "modified, m",
			Usage:       "Show the last modification time",
			EnvVar:      "",
			Hidden:      false,
			Destination: new(bool),
		},
		cli.BoolFlag{
			Name:        "full, f",
			Usage:       "Print full path",
			EnvVar:      "",
			Hidden:      false,
			Destination: new(bool),
		},
	}
	app.Action = func(c *cli.Context) error {
		currentPath = c.GlobalString("directory")
		dir = http.Dir(currentPath)
		port = c.GlobalString("port")
		ssize = c.GlobalBool("size")
		full = c.GlobalBool("full")
		//modt = c.GlobalString("modt")
		fmt.Printf("Path: %s, Port %s\n", dir, port)
		return nil
	}
	app.Run(os.Args)

	//diveIntoFolder(currentPath)
	diveDirTree(currentPath)

	if err != nil {
		log.Println(err)
	}

	http.ListenAndServe(port, http.FileServer(dir))
}

func diveDirTree(path string) {
	fmt.Println("\nfilepath.Walk------------------")
	args = nil
	er = filepath.Walk(path,
		func(path string, info os.FileInfo, errr error) error {
			if er != nil {
				return er
			}
			if info.IsDir() {
				args = append(args, "+")
			}
			if full != false {
				args = append(args, path)
			} else {
				args = append(args, info.Name())
			}
			if ssize != false && !info.IsDir() {
				args = append(args, strconv.FormatInt(info.Size()/1000, 10)+"Kb")
			}
			if modt != false {
				args = append(args, info.ModTime().Local().String())
			}

			args = append(args, "\n")
			return nil
		})

	printList(args)
	args = nil
}

func diveIntoFolder(dir string) {
	fmt.Println("\nioutil.ReadDir------------------")
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {

		args = append(args, f.Name())
		if ssize != false {
			args = append(args, strconv.FormatInt(f.Size()/1000, 10)+"Kb")
		}
		if modt != false {
			args = append(args, f.ModTime().Local().String())
		}
		if f.IsDir() != false {
			args = append(args, dir+"/"+f.Name())
			args = append(args, "\n")
			diveIntoFolder(dir + "/" + f.Name())
		}

		args = append(args, "\n")
		printList(args)
		args = nil
	}
}

func printList(args []string) {
	s := strings.Join(args, " ")
	fmt.Printf(s)
}
