package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/transform"

	"github.com/AJenpan/fcc/chardet"
)

//var Output = ""
//var ForceConvert = false
var Input = ""
var Backup = false
var Recurse = false
var SourceCharset = "gb18030"
var TargetCharset = "utf-8"
var Pattern = "*"
var DryRun = false

func main() {
	app := &cli.App{
		Name:    "fcc (file-charset-convert)",
		Version: "0.1.0",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "input-dir",
				Aliases:     []string{"i"},
				Usage:       "",
				Value:       ".",
				Destination: &Input,
			},
			//&cli.StringFlag{
			//	Name:        "output-dir",
			//	Aliases:     []string{"o"},
			//	Usage:       "",
			//	Destination: &Output,
			//},
			// &cli.BoolFlag{
			// 	Name:        "force-convert",
			// 	Aliases:     []string{"f"},
			// 	Value:       false,
			// 	Destination: &ForceConvert,
			// },
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
				Name:        "pattern",
				Aliases:     []string{"p"},
				Usage:       "glob patterns, like: *.txt, filename.*, ???.txt",
				Destination: &Pattern,
				Required:    true,
			},
			&cli.BoolFlag{
				Name:        "backup",
				Value:       false,
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
		fmt.Println(err)
	}
}

func Run() error {
	SourceCharset = CharsetNameClean(SourceCharset)
	TargetCharset = CharsetNameClean(TargetCharset)

	input := filepath.Clean(Input)

	fileList, err := GetSourceFile(input, Pattern, Recurse)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if len(fileList) == 0 {
		fmt.Println("no file match:", filepath.Join(input, Pattern))
		return nil
	}

	var detectCharset string

	for _, filePath := range fileList {
		if detectCharset, err = DetectCharset(filePath); err != nil {
			fmt.Println(err)
			continue
		}

		if SourceCharset != "auto" && detectCharset != SourceCharset {
			// fmt.Printf("%s detect charset is %s\n", filePath, detectCharset)
			continue
		}

		if detectCharset == TargetCharset {
			// fmt.Printf("%s detect charset is %s\n", filePath, detectCharset)
			continue
		}

		if !DryRun {
			err = ConvertCharset(filePath, detectCharset, TargetCharset, Backup)
			if err != nil {
				fmt.Println(err)
				continue
			}
		}
		fmt.Printf("%s convert %s to %s success\n", filePath, detectCharset, TargetCharset)
	}
	return nil
}

func GetSourceFile(targetDir, pattern string, recurse bool) ([]string, error) {
	var err error
	if !filepath.IsAbs(targetDir) {
		if targetDir, err = filepath.Abs(targetDir); err != nil {
			return nil, err
		}
	}
	rd, err := ioutil.ReadDir(targetDir)
	if err != nil {
		return nil, err
	}
	ret := make([]string, 0, len(rd))
	for _, fi := range rd {
		if fi.IsDir() && recurse {
			nextDir := filepath.Join(targetDir, fi.Name())
			files, err := GetSourceFile(nextDir, pattern, recurse)
			if err != nil {
				return ret, err
			}
			ret = append(ret, files...)
		} else {
			match, err := filepath.Match(pattern, fi.Name())
			if err == nil && match {
				fullName := filepath.Join(targetDir, fi.Name())
				ret = append(ret, fullName)
			}
		}
	}
	return ret, nil
}

var gDetecter = chardet.NewTextDetector()

func DetectCharset(filepath string) (charset string, err error) {
	raw, err := ioutil.ReadFile(filepath)
	if err != nil {
		return
	}
	result, err := gDetecter.DetectBest(raw)
	if err != nil {
		return
	}
	charset = result.Charset
	charset = strings.ToLower(charset)
	return
}

func ConvertCharset(filepath, sourceCharset, targetCharset string, backup bool) error {
	if sourceCharset == targetCharset {
		return nil
	}

	raw, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}

	//backup if need.
	if backup {
		err = ioutil.WriteFile(filepath+".bak", raw, os.ModePerm)
		if err != nil {
			return err
		}
	}

	sourceEncoding, err := htmlindex.Get(sourceCharset)
	if err != nil {
		return fmt.Errorf("unsupported input charset: %s", sourceCharset)
	}

	targetCoding, err := htmlindex.Get(TargetCharset)
	if err != nil {
		return fmt.Errorf("unsupported output charset: %s", TargetCharset)
	}

	//start to convert
	chain := transform.Chain(sourceEncoding.NewDecoder(), targetCoding.NewEncoder())
	targetReader := transform.NewReader(bytes.NewReader(raw), chain)

	f, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}

	_, err = io.Copy(f, targetReader)
	if err1 := f.Close(); err1 != nil && err == nil {
		err = err1
	}

	return err
}

var mapCharsetAliases = map[string]string{
	"gbk":    "gb18030",
	"gb2312": "gb18030",
	"utf8":   "utf-8",
	"utf_8":  "utf-8",
}

func CharsetNameClean(charset string) string {
	charset = strings.ToLower(charset)
	if realName, has := mapCharsetAliases[charset]; has {
		return realName
	}
	return charset
}
