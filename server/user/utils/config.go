package utils

import "github.com/microcosm-cc/bluemonday"

const (
	FrontIP        = "http://93.171.139.195:780"
	FrontIPChat    = "http://93.171.139.195:781"
	BackIP         = "http://93.171.139.196:780"
	ListenAddr     = "172.26.112.3:80"
	DataBaseConfig = "server:12345@tcp(localhost:3306)/users?"
	PathToImages   = `/root/golang/server_with_db/2019_2_TODO/server/`
	Secret         = `askhgashjasl;hjaojgh;asjha;shm;`
)

var Sanitizer = bluemonday.UGCPolicy()
