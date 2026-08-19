package service

import (
	"errors"
	"strings"
)

func pageValues(page, size int) (int, int) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	return size, (page - 1) * size
}
func required(values ...string) error {
	for _, v := range values {
		if strings.TrimSpace(v) == "" {
			return errors.New("required field is empty")
		}
	}
	return nil
}
