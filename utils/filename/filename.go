package filename

import (
	"crypto/md5"
	"encoding/hex"
	"path"
	"strconv"
	"strings"
	"time"
)

// 为上传的文件生成一个文件名.
func getFileName(fileName string, ext string) string {
	h := md5.New()
	h.Write([]byte(fileName + strconv.FormatInt(time.Now().Unix(), 10)))
	return hex.EncodeToString(h.Sum(nil)) + ext
}

// 检查文件后缀合法性并生成一个新的文件名称
func CheckAndCreateFileName(oldName string) (newName string, isLegal bool) {
	ext := path.Ext(oldName)
	if strings.ToLower(ext) == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" {
		//随机生成一个文件名.
		newName = getFileName(oldName, ext)
		isLegal = true
	}
	return newName, isLegal
}
