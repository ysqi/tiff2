package to

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"sync"

	"golang.org/x/image/tiff"
)

var (
	fileChan          chan string
	exitChan          chan bool
	proWG             sync.WaitGroup
	convertConcurrent int //图片转换并发数

	completed          chan bool //标记是否完成
	fileSumCount       uint64    //需处理的图片数量
	fileCompletedCount uint64    //已完成处理的图片数量
	fileSuccessCount   uint64    //已成功处理的图片数量
	searchCompleted    bool      //文件检索数量
)

func run() {
	//根据CPU数开辟多个Go
	convertConcurrent = runtime.NumCPU() * 5
	fileChan = make(chan string, convertConcurrent)
	exitChan = make(chan bool, convertConcurrent)
	completed = make(chan bool, convertConcurrent)

	go startConvert()

	proWG = sync.WaitGroup{}

	if flag.NArg() == 0 {
		log.Printf("默认将%q目录下TIFF图片转换为%q,并输出到%q目录\n", defaultSourcePath, *to, *output)
		path, err := filepath.Abs(defaultSourcePath)
		if err != nil {
			report(err)
			return
		}
		proWG.Add(1)
		go func(path string) {
			defer proWG.Done()
			walkDir(path)
		}(path)
	}
	for i := 0; i < flag.NArg(); i++ {
		path := flag.Arg(i)
		path, err := filepath.Abs(path)
		if err != nil {
			log.Println(err)
			continue
		}
		switch f, err := os.Stat(path); {
		case err != nil:
			log.Println(err)
		case f.IsDir():
			proWG.Add(1)
			go func(path string) {
				defer proWG.Done()
				walkDir(path)
			}(path)
		default:
			visitFile(path, f, nil)
		}
	}

	//开启转换
	for i := 0; i < convertConcurrent; i++ {
		go startConvert()
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	//等待待处理文件工作完成
	proWG.Wait()
	searchCompleted = true
	needBreak := false
	for {
		if needBreak {
			break
		}
		select {
		case <-signalChan:
			for i := 0; i < convertConcurrent; i++ {
				exitChan <- true
			}
			log.Println("程序运行终止")
			needBreak = true
		case <-time.After(2 * time.Second):
		default:
			if searchCompleted && fileCompletedCount == fileSumCount {
				for i := 0; i < convertConcurrent; i++ {
					exitChan <- true
				}
				needBreak = true
			}
		}
	}
}

func startConvert() {
	defer func() {
		if err := recover(); err != nil {
			log.Println("转换时错误,", err)
		}
	}()
	for {
		select {
		case <-exitChan:
			return
		case f := <-fileChan:
			err := process(f)
			atomic.AddUint64(&fileCompletedCount, 1)
			if err != nil {
				log.Printf("文件%q转换失败,%s", f, err)
			} else {
				atomic.AddUint64(&fileSuccessCount, 1)
			}
		case <-time.After(5 * time.Second):
		default:
			if searchCompleted && fileCompletedCount == fileSumCount {
				return
			}
		}
	}

}

func isTIFFFile(f os.FileInfo) bool {
	name := f.Name()
	if f.IsDir() || strings.LastIndex(name, ".") == -1 {
		return false
	}
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".tiff" || ext == ".tif"
}

func visitFile(path string, f os.FileInfo, err error) error {
	if err == nil {
		if isTIFFFile(f) {
			joinFile(path)
			return nil
		}
	}
	if err != nil && !os.IsNotExist(err) {
		log.Println(err)
	}
	return nil
}

func walkDir(path string) {
	log.Printf("检索目录：%s中...\n", path)
	filepath.Walk(path, visitFile)
}

func joinFile(fielname string) {
	//记录需转换的文件数
	atomic.AddUint64(&fileSumCount, 1)
	fileChan <- fielname
}

func process(filename string) (err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("panic:%s", p)
		}
	}()

	//新文件名
	//time.Now().Format("20060102150405")
	newName := filepath.Join(*output, fmt.Sprintf("%s.%s", filepath.Base(filename), *to))
	_, err = os.Stat(newName)
	//如果不能覆盖，则检查文件是否存在，存在时则不进行处理
	if err == nil && !(*replace) {
		return nil
	}

	if err != nil && !os.IsNotExist(err) {
		return err
	}

	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	img, err := tiff.Decode(f)
	if err != nil {
		return err
	}
	out, err := os.Create(newName)
	if err != nil {
		return err
	}
	defer out.Close()
	err = convertTo[*to](out, img)
	if err != nil {

		return err
	}
	return nil
}
