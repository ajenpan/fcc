# file charset convert

## 一个文件字符集批量转换工具

## 背景

因手头上有好几个十多年前的 vs c++ 项目需要维护, 这些项目的源文件都是用 gbk 写的.
因此需要将全部的源文件编码转为 utf-8, 所以有了本项目.

## 功能

- 支持批量转码
- 支持 recurse 子目录
- 支持备份
- 文件名匹配
- 自动探测原文件格式

## install

`go get github.com/AJenpan/fcc`

## help

`fcc --help`

## usage

- --input-dir value, -i value (default: "./")
- --source-charset value, -s value (default: "auto")
- --target-charset value, -t value (default: "utf-8")
- --pattern value, -p value
- --force-convert, -f (default: false)
- --backup (default: true)
- --recurse, -r (default: false)
- --dry-run, -d (default: false)
- --help, -h show help (default: false)
- --version, -v print the version (default: false)

## 使用例子

### 将当前目录下的 gbk \*.h 文件转 utf-8

`fcc -p *.h -i ./ -t utf-8 -b -r `
