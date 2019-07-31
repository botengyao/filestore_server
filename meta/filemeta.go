package meta

import (
	"sort"
)

// FileMeta : 文件元信息结构
type FileMeta struct {
	FileSha1 string
	FileName string
	FileSize int64
	Location string
	UploadAt string
}

var fileMetas map[string]FileMeta

func init() {
	fileMetas = make(map[string]FileMeta)
}

// UpdateFileMeta : add/update file's meta
func UpdateFileMeta(fmeta FileMeta) {
	fileMetas[fmeta.FileSha1] = fmeta
}

// GetFileMeta : get meta by sha1
func GetFileMeta(fileSha1 string) (FileMeta, bool) {
	filemeta, ok := fileMetas[fileSha1]
	return filemeta, ok
}

// GetLastFileMetas : get many files' meta
func GetLastFileMetas(count int) []FileMeta {
	fMetaArray := make([]FileMeta, len(fileMetas))
	for _, v := range fileMetas {
		fMetaArray = append(fMetaArray, v)
	}

	if count >= len(fileMetas) {
		count = len(fileMetas)
	}

	sort.Sort(ByUploadTime(fMetaArray))
	return fMetaArray[0:count]
}

// RemoveFileMeta
func RemoveFileMeta(fileSha1 string) {
	delete(fileMetas, fileSha1)
}
