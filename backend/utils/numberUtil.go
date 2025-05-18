package utils

import (
	"math/rand"
	"time"
)

// GenRandomNumber 生成指定范围内的随机数
// min: 最小值，max: 最大值
func GenRandomNumber(min, max int) int {
	// 设置随机种子确保每次执行的随机化不同
	rg := rand.New(rand.NewSource(time.Now().UnixNano()))
	// 生成1到100之间的随机数
	randomNumber := rg.Intn(max) + min // Intn(100)生成0到99之间的数，+1后变为1到100

	return randomNumber
}
