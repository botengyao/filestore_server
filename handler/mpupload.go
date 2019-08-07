package handler

import (
	"filestore_server/util"
	"fmt"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"

	rPool "filestore_server/cache/redis"
	dblayer "filestore_server/db"
)

// MultipartUploadInfo : 初始化信息
type MultipartUploadInfo struct {
	FileHash   string
	FileSize   int
	UploadID   string
	ChunkSize  int
	ChunkCount int
}

// InitialMultipartUploadHandler : 初始化分块上传
func InitialMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析用户请求参数
	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize, err := strconv.Atoi(r.Form.Get("filesize"))
	if err != nil {
		w.Write(util.NewRespMsg(-1, "params invalid", nil).JSONBytes())
		return
	}

	// 2. 获得redis的一个连接
	rConn := rPool.RedisPool().Get() //
	defer rConn.Close()

	// 3. 生成分块上传的初始化信息
	upInfo := MultipartUploadInfo{
		FileHash:   filehash,
		FileSize:   filesize,
		UploadID:   username + fmt.Sprintf("%x", time.Now().UnixNano()),
		ChunkSize:  5 * 1024 * 1024,                                       // 5MB
		ChunkCount: int(math.Ceil(float64(filesize) / (5 * 1024 * 1024))), //Ceil int
	}
	fmt.Printf("Init ChunkCount" + strconv.Itoa(upInfo.ChunkCount))
	// 4. 将初始化信息写入到redis缓存
	rConn.Do("HSET", "MP_"+upInfo.UploadID, "chunkcount", upInfo.ChunkCount)
	rConn.Do("HSET", "MP_"+upInfo.UploadID, "filehash", upInfo.FileHash)
	rConn.Do("HSET", "MP_"+upInfo.UploadID, "filesize", upInfo.FileSize)

	// 5. 将响应初始化数据返回到客户端
	w.Write(util.NewRespMsg(0, "Initialization OK", upInfo).JSONBytes())
}

// UploadPartHandler : 上传文件分块
func UploadPartHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析用户请求参数
	r.ParseForm()
	//	username := r.Form.Get("username")
	uploadID := r.Form.Get("uploadid")
	chunkIndex := r.Form.Get("index") //文件分块序号
	fmt.Println("Get index" + chunkIndex)
	// 2. 获得redis连接池中的一个连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	// 3. 获得文件句柄，用于存储分块内容
	fpath := "/Users/boteng/Documents/golang/upload_part/" + uploadID + "/" + chunkIndex
	os.MkdirAll(path.Dir(fpath), 0744)
	fd, err := os.Create(fpath)
	if err != nil {
		w.Write(util.NewRespMsg(-1, "Upload part failed", nil).JSONBytes())
		return
	}
	defer fd.Close()

	buf := make([]byte, 1024*1024) //1M
	for {
		n, err := r.Body.Read(buf) //request.Body
		fd.Write(buf[:n])
		if err != nil {
			break
		}
	}

	// 4. 更新redis缓存状态
	/*
		//加锁
		var islock bool
		for !islock {
			islock, err = rPool.Lock(rConn, "lock"+uploadID)
			if islock && err == nil {
				rConn.Do("HSET", "MP_"+uploadID, "chkidx_"+chunkIndex, 1)
				//time.Sleep(50 * time.Millisecond)
				fmt.Println(chunkIndex + " gets lock")
				break
			}
			fmt.Println(chunkIndex + " doesn't get lock")
			time.Sleep(1000 * time.Millisecond)
		}

		_ = rPool.Unlock(rConn, "lock"+uploadID)
	*/
	rConn.Do("HSET", "MP_"+uploadID, "chkidx_"+chunkIndex, 1)
	// 5. 返回处理结果到客户端
	w.Write(util.NewRespMsg(0, "Part OK"+chunkIndex, nil).JSONBytes())
}

// CompleteUploadHandler : 通知上传合并
func CompleteUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析请求参数
	r.ParseForm()
	upid := r.Form.Get("uploadid")
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize := r.Form.Get("filesize")
	filename := r.Form.Get("filename")

	// 2. 获得redis连接池中的一个连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	// 3. 通过uploadid查询redis并判断是否所有分块上传完成
	data, err := redis.Values(rConn.Do("HGETALL", "MP_"+upid))
	/*
		redis> HSET people jack "Jack Sparrow"
			   (integer) 1

			   redis> HSET people gump "Forrest Gump"
			   (integer) 1

			   redis> HGETALL people
					   1) "jack"          # 域
					   2) "Jack Sparrow"  # 值
					   3) "gump"
					   4) "Forrest Gump"
	*/

	if err != nil {
		w.Write(util.NewRespMsg(-1, "Complete upload failed", nil).JSONBytes())
		return
	}
	totalCount := 0
	chunkCount := 0
	fmt.Println("len redis" + strconv.Itoa(len(data)))

	for i := 0; i < len(data); i += 2 {
		k := string(data[i].([]byte))
		v := string(data[i+1].([]byte))
		if k == "chunkcount" {
			totalCount, _ = strconv.Atoi(v)
		} else if strings.HasPrefix(k, "chkidx_") && v == "1" {
			chunkCount++
		}
	}
	fmt.Println("Actual chunkCount" + strconv.Itoa(chunkCount))
	if totalCount != chunkCount {
		w.Write(util.NewRespMsg(-2, "Invalid request", nil).JSONBytes())
		return
	}

	// 4. TODO：合并分块

	// 5. 更新唯一文件表及用户文件表
	fsize, _ := strconv.Atoi(filesize)
	dblayer.OnFileUploadFinished(filehash, filename, int64(fsize), "")
	dblayer.OnUserFileUploadFinished(username, filehash, filename, int64(fsize))

	// 6. 响应处理结果
	w.Write(util.NewRespMsg(0, "Complete OK", nil).JSONBytes())
}
