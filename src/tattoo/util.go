package main

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"time"
)

type KeyValuePair struct {
	Key   int64
	Value string
}

type KeyPairs struct {
	Items []*KeyValuePair
}

func (kp KeyPairs) Len() int {
	return len(kp.Items)
}

func (kp KeyPairs) Swap(i, j int) {
	kp.Items[i], kp.Items[j] = kp.Items[j], kp.Items[i]
}

func (kp KeyPairs) Less(i, j int) bool {
	if kp.Items[i].Key == kp.Items[j].Key {
		return kp.Items[i].Value < kp.Items[j].Value
	}
	return kp.Items[i].Key < kp.Items[j].Key
}

func UUID() string {
	b := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		log.Fatal(err)
	}
	b[6] = (b[6] & 0x0F) | 0x40
	b[8] = (b[8] &^ 0x40) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func TimeHumanReading(t1 int64) string {
	t := time.Unix(t1, 0).Local()
	return fmt.Sprintf("%v-%v-%v %d:%d:%d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
}

func TimeRFC3339(t1 int64) string {
	t := time.Unix(t1, 0).Local()
	return t.Format("2006-01-02T15:04:05Z")
}

func MD5Sum(in string) string {
	md5 := md5.New()
	md5.Write([]byte(in))
	return fmt.Sprintf("%x", md5.Sum(nil))
}

func SHA256Sum(in string) string {
	sha := sha256.New()
	sha.Write([]byte(in))
	return fmt.Sprintf("%x", sha.Sum(nil))
}
