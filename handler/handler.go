package handler

import (
	"encoding/json"
	"filestore_server/meta"
	"filestore_server/util"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

// process - upload file
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		data, err := ioutil.ReadFile("./static/view/index.html")
		if err != nil {
			io.WriteString(w, "internal server error")
			return
		}
		io.WriteString(w, string(data))
	} else if r.Method == "POST" {
		file, head, err := r.FormFile("file")
		if err != nil {
			fmt.Printf("Failed to get data, err: %s\n", err.Error())
			return
		}
		defer file.Close()

		fileMeta := meta.FileMeta{
			FileName: head.Filename,
			Location: "/Users/boteng/Documents/golang/upload_file/" + head.Filename,
			UploadAt: time.Now().Format("2006-01-02 15:04:05"),
		}

		newFile, err := os.Create(fileMeta.Location)
		if err != nil {
			fmt.Printf("Failed to create file, err: %s\n", err.Error())
		}
		defer newFile.Close()

		fileMeta.FileSize, err = io.Copy(newFile, file) //return filesize
		if err != nil {
			fmt.Printf("Failed to save data into file, err: %s\n", err.Error())
			return
		}

		newFile.Seek(0, 0)
		fileMeta.FileSha1 = util.FileSha1(newFile)
		//meta.UpdateFileMeta(fileMeta)
		_ = meta.UpdateFileMetaDB(fileMeta) // use mysql to store meta

		http.Redirect(w, r, "/file/upload/suc", http.StatusFound)
	}

}

// Upload finished
func UploadSucHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Upload finished!")
}

// GetFileMetaHandler : get meta
func GetFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	filehash := r.Form["filehash"][0]
	//fMeta, ok := meta.GetFileMeta(filehash)
	fMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		io.WriteString(w, "Don't find this file")
		return
	}

	data, err := json.Marshal(fMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func FileQueryHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	limitCnt, _ := strconv.Atoi(r.Form.Get("limit"))
	//fileMetas := meta.GetLastFileMetas(limitCnt)

	fileMetas, err := meta.GetLastFileMetasDB(limitCnt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(fileMetas)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	filesha1 := r.Form.Get("filehash")
	fm, err := meta.GetFileMetaDB(filesha1)
	if err != nil {
		io.WriteString(w, "Don't find this file")
		return
	}

	f, err := os.Open(fm.Location)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f) //small file -> big file: use stream
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octect-stream") //.*（ 二进制流，不知道下载文件类型）	application/octet-stream
	w.Header().Set("content-disposition", "attachment; filename=\""+fm.FileName+"\"")
	w.Write(data)
}

//FileMetaUpdateHandler: update the filename in mysql
func FileMetaUpdateHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	opType := r.Form.Get("op")
	fileSha1 := r.Form.Get("filehash")
	newFileName := r.Form.Get("filename")

	if opType != "0" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Get file meta accord to filehash
	curFileMeta, err := meta.GetFileMetaDB(fileSha1)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// update fileName
	curFileMeta.FileName = newFileName
	suc := meta.UpdateFileMetaDB(curFileMeta)
	if !suc {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(curFileMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func FileDeleteHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	filesha1 := r.Form.Get("filehash")

	fmeta, err := meta.GetFileMetaDB(filesha1)
	if err != nil {
		io.WriteString(w, "Don't find this file")
		return
	}

	os.Remove(fmeta.Location)

	ok := meta.OnFileRemovedDB(filesha1)
	if ok != true {
		io.WriteString(w, "Have some errors")
		return
	}
}

/*
// FileMetaUpdateHandler ： process in memory
func FileMetaUpdateHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	opType := r.Form.Get("op")
	fileSha1 := r.Form.Get("filehash")
	newFileName := r.Form.Get("filename")

	if opType != "0" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	curFileMeta, ok := meta.GetFileMeta(fileSha1)
	if ok == false {
		io.WriteString(w, "Don't find this file")
		return
	}
	curFileMeta.FileName = newFileName
	meta.UpdateFileMeta(curFileMeta)

	data, err := json.Marshal(curFileMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}



// FileDeleteHandler : process in memory
func FileDeleteHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fileSha1 := r.Form.Get("filehash")

	fMeta, ok := meta.GetFileMeta(fileSha1)
	if ok == false {
		io.WriteString(w, "Don't find this file")
		return
	}

	os.Remove(fMeta.Location)

	meta.RemoveFileMeta(fileSha1)

	w.WriteHeader(http.StatusOK)
}
*/
