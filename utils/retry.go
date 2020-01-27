package utils

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/gok8s/k8swatch/utils/zlog"
)

func Retry(f func() error, describe string, attempts int, sleep int) error {
	err := f()
	if err == nil {
		return nil
	}
	if s, ok := err.(stop); ok {
		// Return the original error for later checking
		return s.error
	}
	if attempts--; attempts > 0 {
		// Add some randomness to prevent creating a Thundering Herd
		jitter := rand.Int63n(int64(sleep))
		fmt.Println(int(jitter))
		sleep = sleep + int(jitter/2)
		zlog.Errorf("执行:%s失败，错误为:%v，将在%d秒后重试，剩余%d次", describe, err.Error(), sleep, attempts)
		time.Sleep(time.Duration(sleep) * time.Second)
		return Retry(f, describe, attempts, 2*sleep)
	}
	return err
}

type stop struct {
	error
}
