package to

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
)

var (
	output  = flag.String("o", "./output", "转换后文件存放目录")
	to      = flag.String("t", "jpg", "需要转换的目标格式")
	replace = flag.Bool("r", false, "转换存储时是否覆盖已存在的文件")
)

var (
	exitCode          = 0
	defaultSourcePath = "./source"

	convertTo    = map[string]convert{}
	convertCount = 0
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: tiff2 [flags] [path ...]\n")
	flag.PrintDefaults()
}

func report(err interface{}) {
	fmt.Fprintf(os.Stderr, "%s\n", err)
	exitCode = 2
}

// Reg 注册转换器
func Reg(tpe string, c convert) {
	if _, ok := convertTo[tpe]; ok {
		report(tpe + "，已注册的转换类型")
		return
	}
	convertTo[tpe] = c
}

// Tiff2Main 入口
func Tiff2Main() {
	if exitCode != 0 {
		os.Exit(exitCode)
	}
	flag.Usage = usage
	flag.Parse()

	tiff2Main()
	os.Exit(exitCode)
}

func tiff2Main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	if err := checkFlag(); err != nil {
		report(err)
		return
	}
	run()

	if fileSumCount == 0 {
		log.Println("无TIFF格式图片需处理")
	} else {
		log.Printf("共需处理文件%d个，成功转换%d个，失败%d个\n", fileSumCount, fileSuccessCount, fileSumCount-fileSuccessCount)
	}
}
func checkFlag() error {
	if _, ok := convertTo[*to]; !ok {
		all := []string{}
		for key := range convertTo {
			all = append(all, key)
		}
		return fmt.Errorf("暂不支持%q格式转换，当前仅支持%q", *to, all)
	}

	//创建 output dir
	if f, err := os.Stat(*output); err != nil {
		if os.IsNotExist(err) || !f.IsDir() {
			return os.Mkdir(*output, 0700)
		}
		return err
	}
	return nil
}
