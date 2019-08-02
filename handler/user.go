package handler

import (
	dblayer "filestore_server/db"
	"filestore_server/util"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

const (
	pwdSalt = "@!#542"
)

//SignpHandler
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data, err := ioutil.ReadFile("./static/view/signup.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(data)
	}

	r.ParseForm()
	username := r.Form.Get("username")
	passwd := r.Form.Get("password")

	if len(username) < 3 || len(passwd) < 5 {
		w.Write([]byte("Invalid parameter"))
		return
	}
	EncPasswd := util.Sha1([]byte(passwd + pwdSalt))
	suc := dblayer.UserSignUp(username, EncPasswd)
	if suc {
		w.Write([]byte("SUCCESS"))
	} else {
		w.Write([]byte("FAILED"))
	}

}

// SigninHandler
func SigninHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// data, err := ioutil.ReadFile("./static/view/signin.html")
		// if err != nil {
		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	return
		// }
		// w.Write(data)
		http.Redirect(w, r, "/static/view/signin.html", http.StatusFound)
		return
	}

	r.ParseForm()
	username := r.Form.Get("username")
	password := r.Form.Get("password")

	encPasswd := util.Sha1([]byte(password + pwdSalt))

	// 1. verify username and password
	pwdChecked := dblayer.UserSignin(username, encPasswd)
	if !pwdChecked {
		//w.Write([]byte("FAILED"))
		io.WriteString(w, "FAILED, Please retry")
		return
	}

	// 2. generate and update token
	token := GenToken(username)
	upRes := dblayer.UpdateToken(username, token)
	if !upRes {
		w.Write([]byte("FAILED"))
		return
	}

	// 3. After login successfully, redirect to home page
	//w.Write([]byte("http://" + r.Host + "/static/view/home.html"))
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: struct {
			Location string
			Username string
			Token    string
		}{
			Location: "http://" + r.Host + "/static/view/home.html",
			Username: username,
			Token:    token,
		},
	}
	w.Write(resp.JSONBytes())
}

// UserInfoHandler: query user info and return the results to homepage
func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Get the username
	r.ParseForm()
	username := r.Form.Get("username")
	// achieved by HTTPInterceptor
	/*// 2. Verify token
	 token := r.Form.Get("token")
	 isValidToken := IsTokenValid(token)
	 if !isValidToken {
		w.WriteHeader(http.StatusForbidden)
	 	return
	 }
	*/
	// 3. Query user infomation
	user, err := dblayer.GetUserInfo(username)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// 4. Assemble and respond user's data
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: user,
	}
	w.Write(resp.JSONBytes())
}

// GenToken : generate token
func GenToken(username string) string {
	// 40位字符:md5(username+timestamp+token_salt)+timestamp[:8]
	ts := fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix := util.MD5([]byte(username + ts + "_tokensalt"))
	return tokenPrefix + ts[:8] //32 + 8
}

// IsTokenValid
func IsTokenValid(token string, username string) bool {
	if len(token) != 40 {
		return false
	}
	// 5 mins logout
	timeOld, _ := strconv.ParseInt("0x"+token[32:], 0, 64)
	timeNow := time.Now().Unix()
	if timeDiff := timeNow - timeOld; timeDiff > 300 {
		return false
	}
	// Compare with the token in mysql
	if !dblayer.UserTokenVerify(username, token) {
		return false
	}

	return true
}
