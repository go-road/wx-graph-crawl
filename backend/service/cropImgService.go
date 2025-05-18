package service

import (
	"bytes" // 用于处理字节缓冲区
	"io"
	"net/http"      // 用于处理HTTP请求，这里主要用于MIME类型检测
	"os"            // 用于文件系统操作
	"path/filepath" // 用于文件路径操作
	"strings"
	"sync" // 用于并发控制

	"github.com/disintegration/imaging" // 第三方图像处理库，提供更高级的图像处理功能
	"github.com/pkg/errors"
	"github.com/pudongping/wx-graph-crawl/backend/types"
	"go.uber.org/zap"
)

// 定义一个全局的bufferPool，使用sync.Pool来重用字节缓冲区，减少内存分配和垃圾回收开销
var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

type CropImgService struct {
	RootDir        string // 设置要处理的目录路径
	ConcurrencyMax int    // 设置并发处理的goroutines数量
	BottomPixel    int    // 裁剪底部的65像素
}

func NewCropImgService(rootDir string, concurrencyMax int, bottomPixel int) *CropImgService {
	return &CropImgService{
		RootDir:        rootDir,
		ConcurrencyMax: concurrencyMax,
		BottomPixel:    bottomPixel,
	}
}

func (svc *CropImgService) RunCropImg() ([]types.CropResult, error) {
	return svc.processImages(svc.RootDir, svc.BottomPixel, svc.ConcurrencyMax)
}

// isImage 函数检查给定文件是否为图片，通过读取文件的前512字节并检测其MIME类型实现
func (svc *CropImgService) isImage(file *os.File) bool {
	// 从bufferPool中获取一个字节缓冲区，并在函数返回时将其放回pool中
	buffer := bufferPool.Get().(*bytes.Buffer)
	defer bufferPool.Put(buffer)
	buffer.Reset()

	// 从文件中读取前512字节到缓冲区中
	_, err := buffer.ReadFrom(io.LimitReader(file, 512))
	if err != nil {
		return false
	}

	// 使用http.DetectContentType来检测缓冲区中数据的MIME类型
	contentType := http.DetectContentType(buffer.Bytes())
	// 如果MIME类型以"image/"开头，则认为文件是图片
	return strings.HasPrefix(contentType, "image/")
}

// processFile 函数处理单个图片文件，包括打开文件、检测是否为图片、读取图片、裁剪图片和保存图片
func (svc *CropImgService) processFile(path string, bottomPixel int, wg *sync.WaitGroup, semaphore chan struct{}, cropResultChan chan types.CropResult) {
	defer wg.Done()                // 在函数结束时通知WaitGroup，表示一个goroutine完成了工作
	defer func() { <-semaphore }() // 释放信号量，允许其他goroutine开始执行

	var err error
	cropResult := types.CropResult{
		ImgPath: path,
		Err:     nil,
	}

	// 打开图片文件
	file, err := os.Open(path)
	if err != nil {
		cropResult.Err = errors.Wrap(err, "打开文件失败")
		cropResultChan <- cropResult
		return
	}
	defer file.Close() // 确保文件在函数返回时关闭

	// 检查文件是否为图片
	if !svc.isImage(file) {
		zap.L().Info("该文件不是图片", zap.String("path", path))
		return
	}

	// 使用第三方库imaging打开并解码图片
	img, err := imaging.Open(path)
	if err != nil {
		cropResult.Err = errors.Wrap(err, "解码图片失败")
		cropResultChan <- cropResult
		return
	}

	// 使用imaging库裁剪图片，裁掉底部的80像素
	// 在这里要做一个判断，不能让高度减少成为0
	if img.Bounds().Dy() <= bottomPixel {
		cropResult.Err = errors.New("裁剪高度不能大于图片高度")
		cropResultChan <- cropResult
		return
	}
	croppedImg := imaging.CropAnchor(img, img.Bounds().Dx(), img.Bounds().Dy()-bottomPixel, imaging.Top)

	// 将裁剪后的图片保存回原文件
	if err = imaging.Save(croppedImg, path); err != nil {
		cropResult.Err = errors.Wrap(err, "保存图片失败")
		cropResultChan <- cropResult
		return
	}

	zap.L().Info("裁剪并保存成功", zap.String("path", path))
}

// processImages 函数遍历指定目录及其子目录，寻找图片文件并并发处理它们
func (svc *CropImgService) processImages(rootDir string, bottomPixel int, concurrency int) (cropResults []types.CropResult, err error) {
	// 创建一个信号量通道，用于限制并发数量
	semaphore := make(chan struct{}, concurrency)

	var wg sync.WaitGroup                              // WaitGroup用于等待所有goroutine完成
	cropResultChan := make(chan types.CropResult, 100) // 创建一个通道，用于接收裁剪结果

	// 遍历目录中的所有文件和子目录
	err = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// 如果是文件，则启动一个goroutine来处理它
		if !info.IsDir() {
			wg.Add(1)                                                             // 增加WaitGroup的计数
			semaphore <- struct{}{}                                               // 获取信号量，如果信号量用尽则阻塞，直到其他goroutine释放信号量
			go svc.processFile(path, bottomPixel, &wg, semaphore, cropResultChan) // 启动goroutine处理文件
		}
		return nil
	})

	if err != nil {
		return nil, errors.Wrap(err, "遍历目录时出错")
	}

	wg.Wait()             // 等待所有goroutine完成
	close(cropResultChan) // 关闭通道，表示没有更多的裁剪结果

	// 处理裁剪结果
	for cropResult := range cropResultChan {
		cropResults = append(cropResults, cropResult)
	}

	return
}
