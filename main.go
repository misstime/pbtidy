package main

import (
	"bytes"
	"flag"
	"io/fs"
	"io/ioutil"
	"log"
	"regexp"
	"time"
)

const (
	leftDelimiter  = "@{"
	rightDelimiter = "}"
)

var (
	reg1        = regexp.MustCompile(".+`protobuf:\".+?\" json:\".+?\"`\\s*//.*?" + leftDelimiter + ".+?" + rightDelimiter + ".*")           // 正则：validate 标签
	reg2        = regexp.MustCompile("(.+`protobuf:\".+?\" json:\".+?\"`)(\\s*//.*?)(" + leftDelimiter + "(.+?)" + rightDelimiter + ")(.*)") // reg2同reg1，仅多出子匹配
	regEnumType = regexp.MustCompile(`^type (\S+) int32$`)                                                                                   // 正则：枚举定义
)

// getPbFiles 获取指定文件夹下所有后缀为 .pb.go 的文件路径。
// dirPath 为入口根目录，rec 是否递归处理子文件夹，filePath 保存返回的文件路径
func getPbFiles(dirPath string, rec bool, filePaths *[]string) error {
	rd, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return err
	}
	for _, info := range rd {
		if info.IsDir() {
			if rec {
				err := getPbFiles(dirPath+"/"+info.Name(), rec, filePaths)
				if err != nil {
					return err
				}
			}
		} else {
			fileName := info.Name()
			if len(fileName) >= 6 && fileName[len(fileName)-6:] == ".pb.go" {
				*filePaths = append(*filePaths, dirPath+"/"+info.Name())
			}
		}
	}
	return nil
}

// tidyPbFile 对文件中的每一行进行正则处理，编辑为期望样式。done 通道用以在main协程中实现超时控制。
func tidyPbFile(filePath string, done chan<- struct{}) {
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}
	var (
		lines        = bytes.Split(fileBytes, []byte("\n"))
		regEnumConst *regexp.Regexp
	)
	for k, line := range lines {
		// validate 标签
		if reg1.Match(line) {
			subs := reg2.FindSubmatch(line)
			newSub1 := [][]byte{
				subs[1][:len(subs[1])-1],
				[]byte(" validate:\""),
				subs[4],
				[]byte("\"`"),
			}
			subs[1] = bytes.Join(newSub1, nil)
			newSubs := [][]byte{
				subs[1],
				subs[2],
				subs[5],
			}
			lines[k] = bytes.Join(newSubs, nil)
			continue
		}
		// 枚举
		if regEnumConst == nil {
			eSubs := regEnumType.FindSubmatch(line)
			if len(eSubs) > 0 {
				regEnumConst = regexp.MustCompile(`^(\s+?)` + string(eSubs[1]) + `_(\S+? ` + string(eSubs[1]) + ` = \d+)$`)
			}
		} else {
			if len(line) == 1 {
				if line[0] != ')' {
					log.Fatal("pbtidy: assert ( failed")
				}
				regEnumConst = nil
			} else {
				lines[k] = regEnumConst.ReplaceAll(line, []byte("$1$2"))
			}
		}
	}
	newFileBytes := bytes.Join(lines, []byte("\n"))
	err = ioutil.WriteFile(filePath, newFileBytes, fs.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	done <- struct{}{}
}

func main() {
	// 命令行参数解析
	var (
		dir string
		rec bool
	)
	flag.StringVar(&dir, "dir", "", "指定的目录")
	flag.BoolVar(&rec, "rec", true, "是否递归处理子文件夹")
	flag.Parse()

	// 获取所有 .pb.go 文件路径
	filePaths := make([]string, 0)
	err := getPbFiles(dir, rec, &filePaths)
	if err != nil {
		log.Fatal(err)
	}
	if len(filePaths) == 0 {
		log.Fatal("pbtidy: find none `.pb.go` files")
	}

	// 并发处理所有任务。当所任务处理完毕 或 超时，main协程退出
	var (
		done    = make(chan struct{}, len(filePaths))
		timeout = time.After(time.Second * 3)
		doneCnt = 0
	)
	for _, filePath := range filePaths {
		go tidyPbFile(filePath, done)
	}
	for {
		select {
		case <-timeout:
			log.Fatalf("pbtidy timeout: timeout goroutine count -- %v", len(filePaths)-doneCnt)
		case <-done:
			doneCnt++
			if doneCnt == len(filePaths) {
				log.Println("pbtidy done")
				return
			}
		}
	}
}
