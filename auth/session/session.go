package session

import (
	context "context"
	"encoding/json"
	fmt "fmt"
	"log"
	"math/rand"

	"github.com/gomodule/redigo/redis"
)

const sessKeyLen int = 10

type SessionManager struct {
	redisConn redis.Conn
}

func NewSessionManager(conn redis.Conn) *SessionManager {
	return &SessionManager{
		redisConn: conn,
	}
}

func (sm *SessionManager) Create(ctx context.Context, in *Session) (*SessionID, error) {
	id := SessionID{
		ID: RandStringRunes(sessKeyLen),
	}
	dataSerialized, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	mkey := "sessions:" + id.ID
	result, err := redis.String(sm.redisConn.Do("SET", mkey, dataSerialized, "EX", 86400))
	if err != nil {
		return nil, err
	}
	if result != "OK" {
		return nil, fmt.Errorf("result not OK")
	}
	return &id, nil
}

func (sm *SessionManager) Check(ctx context.Context, in *SessionID) (*Session, error) {
	mkey := "sessions:" + in.ID
	data, err := redis.Bytes(sm.redisConn.Do("GET", mkey))
	if err != nil {
		log.Println("cant get data:", err)
		return nil, nil
	}
	sess := &Session{}
	err = json.Unmarshal(data, sess)
	if err != nil {
		log.Println("cant unpack session data:", err)
		return nil, nil
	}
	return sess, nil
}

func (sm *SessionManager) Delete(ctx context.Context, in *SessionID) (*Nothing, error) {
	mkey := "sessions:" + in.ID
	_, err := redis.Int(sm.redisConn.Do("DEL", mkey))
	if err != nil {
		log.Println("redis error:", err)
	}
	return nil, nil
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}