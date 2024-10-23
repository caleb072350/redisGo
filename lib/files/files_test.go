package files

import (
	"testing"
)

func TestGetSize(t *testing.T) {
	// 测试用例1：空文件
	file1 := "test.txt"
	size1, err1 := GetSize(file1)
	if size1 != 0 || err1 == nil {
		t.Errorf("GetSize(%s) = (%d, %v), want (%d, err)", file1, size1, err1, 0)
	}

	file2 := "../../README.md"
	size2, err2 := GetSize(file2)
	t.Logf("GetSize(%s) = (%d, %v)", file2, size2, err2)
	if size2 == 0 || err2 != nil {
		t.Errorf("GetSize(%s) = (%d, %v), want (%d, err)", file2, size2, err2, 0)
	}

}

func TestGetExt(t *testing.T) {
	file := "../../README.md"
	ext := GetExt(file)
	if ext != ".md" {
		t.Errorf("GetExt(%s) = %s, want .md", file, ext)
	}
}

func TestCheckNotExist(t *testing.T) {
	file := "../../README.md"
	if CheckNotExist(file) {
		t.Errorf("CheckNotExist(%s) = true, want false", file)
	}

	file2 := "test.txt"
	if !CheckNotExist(file2) {
		t.Errorf("CheckNotExist(%s) = false, want true", file2)
	}
}
