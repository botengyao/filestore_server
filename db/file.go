package db

import (
	"database/sql"
	mydb "filestore_server/db/mysql"
	"fmt"
)

// TableFile
type TableFile struct {
	FileHash string
	FileName sql.NullString
	FileSize sql.NullInt64
	FileAddr sql.NullString
}

func OnFileUploadFinished(filehash string, filename string,
	filesize int64, fileaddr string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"replace into tbl_file (`file_sha1`,`file_name`,`file_size`," +
			"`file_addr`,`status`) values (?,?,?,?,1)")
	if err != nil {
		fmt.Println("Failed to prepare statement, err: " + err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(filehash, filename, filesize, fileaddr)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	if rf, err := ret.RowsAffected(); err == nil {
		if rf <= 0 {
			fmt.Printf("File with hash:%s has been uploaded before\n", filehash)
		}
		return true
	}
	return false
}

func GetFileMeta(filehash string) (*TableFile, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select file_sha1,file_addr,file_name,file_size from tbl_file " +
			"where file_sha1=? and status=1 limit 1")
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	tfile := TableFile{}
	err = stmt.QueryRow(filehash).Scan(
		&tfile.FileHash, &tfile.FileAddr, &tfile.FileName, &tfile.FileSize) //加&指针，会影响到Scan操作外tfile的值
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	return &tfile, nil
}

// IsFileUploaded
func IsFileUploaded(filehash string) bool {
	stmt, err := mydb.DBConn().Prepare("select 1 from tbl_file where file_sha1=? and status=1 limit 1")
	// TODO: 测试中文输入, 完成查询逻辑
	rows, err := stmt.Query(filehash)
	if err != nil {

		return false
	} else if rows == nil || !rows.Next() {
		return false
	}
	return true
}

// GetFileMetaList : get the list of meta from mysql
func GetFileMetaList(limit int) ([]TableFile, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select file_sha1,file_addr,file_name,file_size from tbl_file " +
			"where status=1 limit ?")
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(limit)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	cloumns, _ := rows.Columns() //Columns returns the column names. Columns returns an error if the rows are closed.
	values := make([]sql.RawBytes, len(cloumns))
	//RawBytes is a byte slice that holds a reference to memory owned by the database itself. After a Scan into a RawBytes, the slice is only valid until the next call to Next, Scan, or Close.
	fmt.Printf("RawBytes:")
	fmt.Println(len(values))
	fmt.Printf("cloumns:")
	fmt.Println(len(cloumns)) //output：4
	fmt.Println(cloumns)      // output： [file_sha1 file_addr file_name file_size]

	var tfiles []TableFile
	//for i := 0; i < len(values) && rows.Next(); i++ {
	for rows.Next() {
		tfile := TableFile{}
		err = rows.Scan(&tfile.FileHash, &tfile.FileAddr, //Scan copies the columns in the current row into the values pointed at by dest.
			&tfile.FileName, &tfile.FileSize) //The number of values in dest must be the same as the number of columns in Rows.

		if err != nil {
			fmt.Println(err.Error())
			break
		}
		tfiles = append(tfiles, tfile)
	}

	fmt.Println(len(tfiles))
	return tfiles, nil
}

// OnFileRemoved : delete file (just change status=2)
func OnFileRemoved(filehash string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"update tbl_file set status=2 where file_sha1=? and status=1 limit 1")
	if err != nil {
		fmt.Println("Failed to prepare statement, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(filehash)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	if rf, err := ret.RowsAffected(); nil == err {
		if rf <= 0 {
			fmt.Printf("File with hash:%s not uploaded", filehash)
		}
		return true
	}
	return false
}
