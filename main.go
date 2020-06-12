package main

import (
	"bytes"
	"fmt"

	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/saintfish/chardet"
	lg "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"golang.org/x/net/html/charset"
)

var log = lg.New()

func todo() {
	app := &cli.App{
		Name: "file-charset-convert",
		//Version: version,
		Usage: "file-charset-convert is a tool based xorm package",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "include",
				Usage: "",

				Hidden:      false,
				Value:       "",
				Destination: nil,
			},
		},
		//Action: func(ctx *cli.Context) error {
		//	return cli.ShowAppHelp(ctx)
		//},
	}
	//sort.Sort(cli.FlagsByName(app.Flags))
	//sort.Sort(cli.CommandsByName(app.Commands))

	if err := app.Run(os.Args); err != nil {
		//log.Fatal(err)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s + [filename]\n", os.Args[0])
		return
	}
	allFiles, _ := getAllFiles()
	files := os.Args[1:]
	for _, filePattern := range files {
		fileList, _ := getFileList(filePattern, allFiles)
		for _, file := range fileList {
			fText, err := ioutil.ReadFile(file)
			if err != nil {
				log.Error("ioutil.ReadFile %s failed: %s", file, err)
				continue
			}

			charCode, err := detectCode(fText)
			if err != nil {
				log.Error("detectCode failed: %s", err)
				continue
			}

			if charCode == "GB-18030" {
				//oriFile, err := os.OpenFile(file+".ori", os.O_RDWR|os.O_CREATE, 0666)
				oriFile, err := os.Create(file + ".bak")
				if err != nil {
					log.Error("OpenFile %s failed: %s", file+".bak", err)
					continue
				}

				//newFile, err := os.OpenFile(file, os.O_RDWR, 0666)
				newFile, err := os.Create(file)
				if err != nil {
					log.Error("OpenFile %s failed: %s", file, err)
					oriFile.Close()
					continue
				}

				_, err = oriFile.Write(fText)
				if err != nil {
					log.Error("oriFile.Write failed: %s", err)
					oriFile.Close()
					newFile.Close()
					continue
				}

				// github.com/saintfish/chardet 只检测 GB-18030
				// golang.org/x/net/html/charset 只能用gbk
				newContent, err := convertToUtf8(fText, "gbk")
				if err != nil {
					log.Error("convertToUtf8 failed: %s", err)
					oriFile.Close()
					newFile.Close()
					continue
				}

				_, err = newFile.Write(newContent)
				if err != nil {
					log.Error("newFile.Write failed: %s", err)
					oriFile.Close()
					newFile.Close()
					continue
				}

				oriFile.Close()
				newFile.Close()
				fmt.Printf("%s convert from %s to UTF-8 success!\n", file, charCode)
			}
		}
	}

}

func convertToUtf8(src []byte, encode string) ([]byte, error) {
	byteReader := bytes.NewReader(src)
	reader, err := charset.NewReaderLabel(encode, byteReader)
	if err != nil {
		log.Errorf("charset.NewReaderLabel failed : %s", err)
		return nil, err
	}

	dst, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Errorf("ioutil.ReadAll failed : %s", err)
		return nil, err
	}
	return dst, nil
}

func detectCode(src []byte) (string, error) {
	detector := chardet.NewTextDetector()
	var result *chardet.Result

	result, err := detector.DetectBest(src)
	if err != nil {
		log.Errorf("detector.DetectBest failed: %s", err)
		return "", err
	}

	log.Infof("charset: %s, language: %s, confidence: %d",
		result.Charset, result.Language, result.Confidence)

	return result.Charset, nil
}

func getFileList(filename string, fileList []string) ([]string, error) {
	var res = make([]string, 0, 10)
	for _, file := range fileList {
		//file = filepath.ToSlash(file)
		if match, _ := filepath.Match(filename, file); match {
			res = append(res, file)
		}
	}
	return res, nil
}

func getAllFiles() ([]string, error) {
	var allFiles = make([]string, 0)
	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			allFiles = append(allFiles, path)
		}
		if err != nil {
			log.Error("Walk err: %s", err)
			return err
		}
		return nil
	})

	return allFiles, nil
}

func getFilesWithPattern(targetPath, pattern string) []string {
	ret := make([]string, 0)
	err := filepath.Walk(".", func(file string, info os.FileInfo, err error) error {
		if !info.IsDir() {

			ret = append(ret, file)
		}
		if err != nil {
			log.Error("Walk err: %s", err)
			return err
		}
		return nil
	})
	if err != nil {
		log.Error(err)
	}
	return ret
}
