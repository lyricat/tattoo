package main

import (
    "encoding/json"
    "io/ioutil"
    "fmt"
    "time"
    "strconv"
    )

const CONFIG_NAME = "settings.json"

type Config struct {
    Port int
    Certificate string
    SiteBase string
    SiteURL string
    SiteTitle string
    SiteSubTitle string
    StaticURL string
    AuthorName string
    TimelineCount int
}

var config *Config = nil
var sessionToken string

func init() {
    config = new(Config)
    // default config
    config.Port = 8888
    config.Certificate = SHA256Sum("42")
    config.SiteBase = "localhost"
    config.SiteURL = "http://localhost:8888"
    config.SiteTitle = "TATTOO!"
    config.StaticURL = config.SiteURL + "/static"
    config.AuthorName = "root"
    config.TimelineCount = 3
    sessionToken = GenerateSessionToken()
}

func GetSessionToken() string {
    return sessionToken
}

func GenerateSessionToken() string {
    sessionToken = SHA256Sum(strconv.Itoa(time.Now().Nanosecond()))
    return sessionToken
}

func RevokeSessionTokon() {
    sessionToken = GenerateSessionToken()
}

func GetConfig() *Config {
    return config
}

func (config * Config) Load() error {
    buff, err := ioutil.ReadFile(CONFIG_NAME)
    if err != nil {
        fmt.Println("Read file failed:", err)
        config.Save()
    }
    buff, err = ioutil.ReadFile(CONFIG_NAME)
    if err != nil {
        fmt.Println("Read file failed:", err)
        return err
    }
    if err := json.Unmarshal(buff, config); err != nil {
        fmt.Println("Unmarshal json failed:", err)
        return err
    }
    return nil
}

func (config * Config) Save() error {
    jsobj, err := json.Marshal(config)
    if err != nil {
        fmt.Println("Marshal json failed:", err)
        return err
    }
    ioutil.WriteFile(CONFIG_NAME, jsobj, 0644)
    return nil
}

func (config * Config) Modify(newcfg *Config) bool {
    config.Port = newcfg.Port
    return true
}

func (config * Config) String() string{
    return fmt.Sprintf("{ Port: %v, Certificate: %v }", config.Port, config.Certificate)
}

