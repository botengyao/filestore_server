package meta

import (
	mydb "filestore_server/db"
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

// UpdateFileMetaDB : add/update file meta to mysql
func UpdateFileMetaDB(fmeta FileMeta) bool {
	return mydb.OnFileUploadFinished(
		fmeta.FileSha1, fmeta.FileName, fmeta.FileSize, fmeta.Location)
}

// GetFileMetaDB
func GetFileMetaDB(fileSha1 string) (FileMeta, error) {
	tfile, err := mydb.GetFileMeta(fileSha1)
	if err != nil {
		return FileMeta{}, err
	}
	fmeta := FileMeta{
		FileSha1: tfile.FileHash,
		FileName: tfile.FileName.String,
		FileSize: tfile.FileSize.Int64,
		Location: tfile.FileAddr.String,
	}
	return fmeta, nil
}

// GetLastFileMetasDB : get the list of meta from mysql -> format(FileMeta)
func GetLastFileMetasDB(limit int) ([]FileMeta, error) {
	tfiles, err := mydb.GetFileMetaList(limit)
	if err != nil {
		return make([]FileMeta, 0), err
	}

	tfilesm := make([]FileMeta, len(tfiles))
	for i := 0; i < len(tfilesm); i++ {
		tfilesm[i] = FileMeta{
			FileSha1: tfiles[i].FileHash,
			FileName: tfiles[i].FileName.String,
			FileSize: tfiles[i].FileSize.Int64,
			Location: tfiles[i].FileAddr.String,
		}
	}
	return tfilesm, nil
}

// OnFileRemovedDB : just change the status to 2
func OnFileRemovedDB(filehash string) bool {
	return mydb.OnFileRemoved(filehash)
}
