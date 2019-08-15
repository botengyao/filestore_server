package config

import (
	cmn "filestore_server/common"
)

const (
	// TempLocalRootDir : 本地临时存储地址的路径
	TempLocalRootDir = "/Users/boteng/Documents/golang/upload_file/"
	// CurrentStoreType : 设置当前文件的存储类型
	CurrentStoreType = cmn.StoreOSS
)
