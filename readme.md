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

## 补充

detect 已经存在乱码的文件, 将匹配情况可能性大的编码格式.

## install

`go get github.com/AJenpan/fcc`

## help

`fcc --help`

## usage

`fcc --help`

## 使用例子

### 将当前目录下的 `*.h` 文件转 `utf-8` 编码

`fcc -p *.h`

### 将当前目录下,并且包括全部子文件夹的 `*.md` 文件由`utf-8`转`gbk`编码, 并且备份

`fcc -i ./ -r -p *.md -s utf-8 -t gbk --backup`

### 只探测(不转换)

`fcc -p *.h detect`