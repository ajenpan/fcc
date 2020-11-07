package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/transform"

	"github.com/AJenpan/fcc/chardet"
)

var Input = ""

//var Output = ""
var Backup = true
var Recurse = false
var ForceConvert = false
var SourceCharset = "gb18030"
var TargetCharset = "utf-8"
var Pattern = "*"
var DryRun = false

func main() {
	app := &cli.App{
		Name:    "fcc (file-charset-convert)",
		Version: "0.0.1",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "input-dir",
				Aliases:     []string{"i"},
				Usage:       "",
				Value:       "./",
				Destination: &Input,
			},
			//&cli.StringFlag{
			//	Name:        "output-dir",
			//	Aliases:     []string{"o"},
			//	Usage:       "",
			//	Destination: &Output,
			//},
			&cli.StringFlag{
				Name:        "source-charset",
				Aliases:     []string{"s"},
				Value:       "auto",
				Destination: &SourceCharset,
			},
			&cli.StringFlag{
				Name:        "target-charset",
				Aliases:     []string{"t"},
				Value:       "utf-8",
				Destination: &TargetCharset,
			},
			&cli.StringFlag{
				Name:    "pattern",
				Aliases: []string{"p"},
				// Value:       "*",
				Destination: &Pattern,
				Required:    true,
			},
			&cli.BoolFlag{
				Name:        "force-convert",
				Aliases:     []string{"f"},
				Value:       false,
				Destination: &ForceConvert,
			},
			&cli.BoolFlag{
				Name:        "backup",
				Value:       true,
				Destination: &Backup,
			},
			&cli.BoolFlag{
				Name:        "recurse",
				Aliases:     []string{"r"},
				Value:       false,
				Destination: &Recurse,
			},
			&cli.BoolFlag{
				Name:        "dry-run",
				Aliases:     []string{"d"},
				Value:       false,
				Destination: &DryRun,
			},
		},
		Action: func(context *cli.Context) error {
			return Run()
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Println(err)
	}
}

func Run() error {
	fileList, err := GetSourceFile(Input, Pattern, Recurse)
	if err != nil {
		log.Println(err)
		return err
	}
	for _, file := range fileList {
		var err error
		if !DryRun {
			err = ConvertFile(file)
		}
		if err != nil {
			log.Println(err)
		} else {
			log.Printf("convert file success %s\n", file)
		}
	}
	return nil
}

func GetSourceFile(dir, pattern string, recurse bool) ([]string, error) {
	if !filepath.IsAbs(dir) {
		current, err := os.Getwd()
		if err == nil {
			dir = path.Join(current, dir)
		}
	}
	dir = filepath.ToSlash(dir)
	rd, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	ret := make([]string, 0, len(rd))
	for _, fi := range rd {
		if fi.IsDir() {
			if recurse {
				nextDir := path.Join(dir, fi.Name())
				files, err := GetSourceFile(nextDir, pattern, recurse)
				if err != nil {
					return ret, err
				}
				ret = append(ret, files...)
			}
		} else {
			match, err := path.Match(pattern, fi.Name())
			if err == nil && match {
				fullName := path.Join(dir, fi.Name())
				ret = append(ret, fullName)
			}
		}
	}
	return ret, nil
}

func ConvertFile(filepath string) error {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}

	sCharset := ""

	result, err := chardet.NewTextDetector().DetectBest(data)
	if err != nil {
		if SourceCharset == "auto" {
			return fmt.Errorf("detector file:%s failed: %s", filepath, err)
		}
		sCharset = SourceCharset
	} else {
		if SourceCharset == "auto" {
			sCharset = result.Charset
		} else {
			if result.Charset != SourceCharset {
				log.Printf("detect the file:%s charset:%s, but SourceCharset:%s;ForceConvert:%v\n",
					filepath, result.Charset, SourceCharset, ForceConvert)
				if !ForceConvert {
					return nil
				}
			}
			sCharset = SourceCharset
		}
	}

	sCharset = strings.ToLower(sCharset)
	if sCharset == strings.ToLower(TargetCharset) {
		return nil
	}

	sourceEncoding, _ := htmlindex.Get(sCharset)
	if sourceEncoding == nil {
		return fmt.Errorf("unsupported input charset: %s", sCharset)
	}

	targetCoding, _ := htmlindex.Get(TargetCharset)
	if targetCoding == nil {
		return fmt.Errorf("unsupported output charset: %s", TargetCharset)
	}

	//backup if need.
	if Backup {
		err = ioutil.WriteFile(filepath+".bak", data, os.ModePerm)
		if err != nil {
			return err
		}
	}

	//begin to convert
	chain := transform.Chain(sourceEncoding.NewDecoder(), targetCoding.NewEncoder())
	targetReader := transform.NewReader(bytes.NewReader(data), chain)

	data, err = ioutil.ReadAll(targetReader)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filepath, data, os.ModePerm)
}
