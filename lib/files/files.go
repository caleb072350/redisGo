package files

import (
	"fmt"
	"os"
	"path"
)

// func GetSize(f multipart.File) (int, error) {
// 	content, err := io.ReadAll(f)
// 	return len(content), err
// }

func GetSize(fileName string) (int64, error) {
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		return 0, err
	}
	return fileInfo.Size(), nil
}

func GetExt(fileName string) string {
	return path.Ext(fileName)
}

func CheckNotExist(src string) bool {
	_, err := os.Stat(src)
	return os.IsNotExist(err)
}

func CheckPermission(src string) bool {
	_, err := os.Stat(src)
	return os.IsPermission(err)
}

func IsNotExistMkDir(src string) error {
	if CheckNotExist(src) {
		if err := MkDir(src); err != nil {
			return err
		}
	}
	return nil
}

func MkDir(src string) error {
	if err := os.MkdirAll(src, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func Open(name string, flag int, perm os.FileMode) (*os.File, error) {
	f, err := os.OpenFile(name, flag, perm)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func MustOpen(fileName, dir string) (*os.File, error) {
	perm := CheckPermission(dir)
	if perm {
		return nil, fmt.Errorf("permission denied dir: %s", dir)
	}

	err := IsNotExistMkDir(dir)
	if err != nil {
		return nil, fmt.Errorf("file mk dir: %s", err)
	}

	f, err := Open(dir+string(os.PathSeparator)+fileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("file open: %s", err)
	}

	return f, nil
}
