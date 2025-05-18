package service

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/pudongping/wx-graph-crawl/backend/utils"
	"go.uber.org/zap"
)

var (
	reImg = regexp.MustCompile(`(?i)\.(jpg|jpeg|png|gif)$`)
)

type MoveImgService struct {
	DirPath     string
	MaxNumImage int
}

func NewMoveImgService(dirPath string, maxNumImage int) *MoveImgService {
	return &MoveImgService{
		DirPath:     dirPath,     // 设置要处理的目录路径
		MaxNumImage: maxNumImage, // 每个文件夹中最多可放的图片数量
	}
}

func (svc *MoveImgService) RunMoveImg() error {
	zap.L().Info("开始处理图片", zap.String("dirPath", svc.DirPath))
	imgChanges, err := svc.WalkAllImages()
	if err != nil {
		return errors.Wrap(err, "遍历图片时出错")
	}

	zap.L().Info("开始打乱图片顺序并更改名称", zap.String("dirPath", svc.DirPath))
	errorSlices := svc.RenameImages(imgChanges)
	for _, errorItem := range errorSlices {
		zap.L().Info("图片改名失败", zap.Error(errorItem))
	}

	zap.L().Info("开始拆分图片", zap.String("dirPath", svc.DirPath))
	if err := svc.DisassembleDir(imgChanges); err != nil {
		return errors.Wrap(err, "拆分图片时出错")
	}

	zap.L().Info("图片处理完成", zap.String("dirPath", svc.DirPath))
	return nil
}

func (svc *MoveImgService) isImageFile(filename string) bool {
	// 支持 JPEG, PNG, GIF 等格式
	return reImg.MatchString(filename)
}

func (svc *MoveImgService) randomizeImgFiles(files []string) map[string]string {
	rg := rand.New(rand.NewSource(time.Now().UnixNano())) // 设置随机种子确保每次执行的随机化不同
	rg.Shuffle(len(files), func(i, j int) {
		files[i], files[j] = files[j], files[i]
	})

	// 创建一个临时映射，以避免直接覆盖文件
	tempMap := make(map[string]string)
	mux := sync.Mutex{}
	for index, file := range files {
		ext := filepath.Ext(file)
		// 重新生成新的文件名，避免修改文件名称时，出现命名冲突
		randomNumber := utils.GenRandomNumber(1, 100)
		newName := fmt.Sprintf("%s/%d_%d%s", filepath.Dir(file), index+1, randomNumber, ext)
		mux.Lock()
		tempMap[file] = newName
		mux.Unlock()
	}

	return tempMap
}

func (svc *MoveImgService) WalkAllImages() (imgChanges []map[string]string, err error) {
	// 使用filepath.Walk递归遍历目录
	err = filepath.Walk(svc.DirPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return errors.Wrapf(err, "Walk 出错：%s", path)
		}
		if info.IsDir() {
			// 因为这里只考虑了图片会放到文件夹下，所有只考虑文件夹下中的图片
			files, err := ioutil.ReadDir(path)
			if err != nil {
				return errors.Wrap(err, "读取目录出错")
			}
			var imageFiles []string
			for _, file := range files {
				if file.IsDir() {
					continue
				}
				if svc.isImageFile(file.Name()) {
					// 收集符合条件的图片文件
					imageFiles = append(imageFiles, filepath.Join(path, file.Name()))
				}
			}
			// 打乱图片文件，并修改文件名称
			if len(imageFiles) > 1 {
				ret := svc.randomizeImgFiles(imageFiles)
				imgChanges = append(imgChanges, ret)
			}
		}

		return nil
	})

	return
}

func (svc *MoveImgService) RenameImages(imgChanges []map[string]string) (errorSlices []error) {
	// 执行重命名
	for _, tempMap := range imgChanges {
		for original, newName := range tempMap {
			if err := os.Rename(original, newName); err != nil {
				errorSlices = append(errorSlices, errors.Wrapf(err, "%s 改名时，改成 %s 出错", original, newName))
			} else {
				zap.L().Info("图片改名成功", zap.String("original", original), zap.String("newName", newName))
			}
		}
	}

	return
}

func (svc *MoveImgService) DisassembleDir(imgChanges []map[string]string) error {
	for _, item := range imgChanges {
		if err := svc.disassembleDirItem(item); err != nil {
			return errors.Wrap(err, "拆分目录时出现异常")
		}
	}

	return nil
}

func (svc *MoveImgService) getRandomImgPath(imgs map[string]string) string {
	values := make([]string, 0, len(imgs))
	for _, v := range imgs {
		values = append(values, v)
	}

	rg := rand.New(rand.NewSource(time.Now().UnixNano())) // 初始化随机数生成器
	randomIndex := rg.Intn(len(values))                   // 获取一个随机索引
	return values[randomIndex]
}

func (svc *MoveImgService) disassembleDirItem(imgs map[string]string) error {
	imageCount := len(imgs)
	if svc.MaxNumImage <= 0 {
		return nil
	}
	if imageCount <= svc.MaxNumImage {
		return nil
	}

	var err error
	// 因为微信中一篇小绿书的图片最多只能有 20 张，因此就拆成 2 个文件夹好了
	halfCount := imageCount / 2
	imageFiles := make([]string, 0, imageCount)
	for _, v := range imgs {
		imageFiles = append(imageFiles, v)
	}

	// 创建新的子目录
	parentPath := filepath.Dir(imageFiles[0]) // 随便取一个图片，从而获得父目录
	subDir1Name := filepath.Join(parentPath, strconv.Itoa(imageCount)+"_1")
	subDir2Name := filepath.Join(parentPath, strconv.Itoa(imageCount)+"_2")

	err = utils.MkdirIfNotExist(subDir1Name)
	if err != nil {
		return errors.Wrap(err, "创建目录1出错")
	}
	err = utils.MkdirIfNotExist(subDir2Name)
	if err != nil {
		return errors.Wrap(err, "创建目录2出错")
	}

	// 对图片进行排序并移动
	sort.Strings(imageFiles)
	for i, imageFile := range imageFiles {
		imgFileName := filepath.Base(imageFile)
		var dstPath string
		if i < halfCount {
			dstPath = filepath.Join(subDir1Name, imgFileName)
		} else {
			dstPath = filepath.Join(subDir2Name, imgFileName)
		}

		// 移动图片
		if err = os.Rename(imageFile, dstPath); err != nil {
			return errors.Wrapf(err, "%s ==> %s 移动失败 %v", imageFile, dstPath, err)
		}

	}

	return nil
}
