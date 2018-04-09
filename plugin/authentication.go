package plugin


package authorization

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"time"


	"dana-tech.com/wbw/logs"
)


type BasicToken struct {
	mutex         sync.RWMutex
	defaultToken  string
	previousToken string
	token         string
	timeStamp     int64
	watChan       chan string
}

var BscToken *BasicToken

func (t *BasicToken) startTask(timeleft int64) {
	for {
		now := time.Now()
		//计算剩余时间
		next := now.Add(time.Duration(timeleft) * time.Second)
		timer := time.NewTimer(next.Sub(now))
		<-timer.C

		//update token
		newT := newToken()
		t.setToken(newT.token, newT.timeStamp)
		writeToken(t.token, t.timeStamp)
		t.watChan <- newT.token
		timeleft = tokenTTL
		logs.Logger.Infof("update token finish token:%v, timestamp:%v", t.token, t.timeStamp)
	}
}

func (t *BasicToken) GetToken() string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.token
}

func (t *BasicToken) CheckToken(token string) bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	if token == t.token {
		return true
	} else if token == t.previousToken {
		return true
	} else if t.defaultToken != "" {
		return token == t.defaultToken
	}

	return false
}

func (t *BasicToken) WatchToken() <-chan string {
	return t.watChan
}

func (t *BasicToken) setToken(newToken string, ts int64) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.previousToken = t.token
	t.token = newToken
	t.timeStamp = ts
}

//生成随机字符串
func genRandomString(length int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

func newToken() *BasicToken {
	tk := genRandomString(32)
	return &BasicToken{
		token:     tk,
		timeStamp: time.Now().Unix(),
		watChan:   make(chan string),
	}
}

func writeToken(token string, timestamp int64) {
	tk := token + "+" + fmt.Sprintf("%v", timestamp)
	err := ioutil.WriteFile(getTokenFile(), []byte(tk), 0644)
	if err != nil {
		logs.Logger.Errorf("write token errror %v", err)
	}
}

func InitToken(defaultToken string) {
	//启动时先看当前目录是否有之前的token
	var timeleft int64
	var has bool

	dat, err := ioutil.ReadFile(getTokenFile())
	if err == nil {
		strtmp := strings.Split(string(dat), "+")
		if len(strtmp) == 2 {
			ts, _ := strconv.ParseInt(strtmp[1], 10, 64)
			tleft := ts + tokenTTL - time.Now().Unix()
			if tleft > 0 && ts > 0 {
				BscToken = &BasicToken{
					token:     strtmp[0],
					timeStamp: ts,
					watChan:   make(chan string),
				}
				timeleft = tleft
				has = true
			}
		}
	}
	if !has {
		BscToken = newToken()
		writeToken(BscToken.token, BscToken.timeStamp)
		timeleft = tokenTTL
	}
	BscToken.defaultToken = defaultToken
	go BscToken.startTask(timeleft)
}


func getTokenFile() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return ""
	}
	currentDir := strings.Replace(dir, "\\", "/", -1)

	path := currentDir + "/" + tokenFileName
	return path
}