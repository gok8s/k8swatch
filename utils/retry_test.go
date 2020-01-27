package utils

import (
	"errors"
	"fmt"
	"testing"
)

func Sqrt(f float64) (float64, error) {
	fmt.Println("start sqrt")
	if f < 0 {
		return 0, errors.New("math: square root of negative number")
	}
	fmt.Println("Ok")
	return 0, nil
}

func Test_Retry(t *testing.T) {
	Retry(func() error {
		_, error := Sqrt(-1)
		return error
	}, "test", 5, 3)
}
