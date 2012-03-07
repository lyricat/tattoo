package webapp
import (
    "os"
    "path"
    "path/filepath"
    "encoding/json"
    "io/ioutil"
    "fmt"
    "time"
	"errors"
    )

const (
    OP_SAVE = 0
    )

/*
 Storage:
            - MODE_SINGLE                   MODE_MULIPLE     
    FILE      JSON(index)                   JSON(index), STRING(value)
    GET       JSON/String                   JSON/String
    SET       JSON/String                   JSON/String  
 */

type Storage interface {
    Get(string)string
    Set(string)string
    Delete(string)string
}

type StorageStat struct {
    GetCount int64
    SetCount int64
    DeleteCount int64
    HasCount int64
    SaveIndexCount int64
}


type StorageCmdItem struct {
    Op int
    Arg1 interface{}
}

type FileStorage struct {
    Path string;
    Mode int;
    Index map[string]string
    Stat StorageStat
// @TODO
    GetChan chan string
    SetChan chan string
    CmdChan chan *StorageCmdItem
}

const (
    FILE_STORAGE_MODE_SINGLE = 0
    FILE_STORAGE_MODE_MULIPLE = 1
    )

func (fs *FileStorage) Init(path string, mode int) error {
    fs.Path = path
    fs.Mode = mode
    fs.CmdChan = make(chan *StorageCmdItem)
    fs.GetChan = make(chan string)
    fs.SetChan = make(chan string)
    fs.Index = make(map[string]string)
    fs.Index["*"] = "placeholder"
    indexPath := fs.getIndexFilePath()
    _, err := ioutil.ReadFile(indexPath)
    if err != nil {
        fmt.Println("Read index file failed, create new one:", err)
        dir, _ := filepath.Split(indexPath)
        os.MkdirAll(dir, 0755)
        fs.SaveIndex()
    }
    fs.LoadIndex()
    go fs.Master()
    go func () {
        for {
            time.Sleep(120*1e9)
            cmd := new(StorageCmdItem)
            cmd.Op = OP_SAVE
            fs.CmdChan <- cmd
        }
    }()
    return nil
}

func (fs *FileStorage) Master() {
    for {
        select {
        case cmd:= <-fs.CmdChan:
            if cmd.Op == OP_SAVE {
                fs.SaveIndex()
            }
        case v:= <-fs.GetChan:
            // @TODO get value
            print(v)
        case v:= <-fs.SetChan:
            // @TODO set value
            print(v)
        }
    }
}

func (fs *FileStorage) Get(key string) ([]byte, error) {
    if key == "*" {
        return nil, errors.New(ErrNotFound)
    }
    fs.Stat.GetCount ++
    if fs.Mode == FILE_STORAGE_MODE_SINGLE {
        if val, ok := fs.Index[key]; ok {
            return []byte(val), nil
        } else {
            return nil, errors.New(ErrNotFound)
        }
    } else if fs.Mode == FILE_STORAGE_MODE_MULIPLE {
        if _, ok := fs.Index[key]; ok {
            valueFilePath := path.Join(fs.Path, key)
            value, err := ioutil.ReadFile(valueFilePath)
            if err != nil {
                fmt.Println("FileStorage.Get, Read file failed:", err)
                return nil, err
            }
            return value, nil
        } else {
            return nil, errors.New(ErrNotFound)
        }
    }
    return nil, errors.New(ErrNotFound)
}

func (fs *FileStorage) Has(key string) bool {
    if key == "*" {
        return false
    }
    fs.Stat.HasCount ++
    if _, ok := fs.Index[key]; ok {
        return true
    }
    return false
}

func (fs *FileStorage) Set(key string, value []byte) error {
    fs.Stat.SetCount ++
    if fs.Mode == FILE_STORAGE_MODE_SINGLE {
        fs.Index[key] = string(value)
    } else if fs.Mode == FILE_STORAGE_MODE_MULIPLE {
        valueFilePath := path.Join(fs.Path, key)
        fs.Index[key] = valueFilePath
        ioutil.WriteFile(valueFilePath, value, 0644)
    }
    return nil
}
func (fs *FileStorage) GetString(key string) (string, error) {
    str, err := fs.Get(key)
    if err != nil {
        return "", err
    }
	return string(str), err
}

func (fs *FileStorage) SetString(key string, value string) error {
	return fs.Set(key, []byte(value))
}

func (fs *FileStorage) GetJSON(key string) (interface{}, error) {
    var jsobj interface{}
    str, err := fs.Get(key)
    if err != nil {
        return nil, err
    }
	if err := json.Unmarshal(str, &jsobj); err != nil {
		fmt.Printf("FileStorage.GetJSON, Unmarshal json failed (%v):%s\n", key, err)
			return nil, err
	}
    return jsobj, nil
}

func (fs *FileStorage) SetJSON(key string, value interface{}) error {
	str, err := json.Marshal(value)
	if err != nil {
		fmt.Printf("FileStorage.SetJSON, Marshal json failed (%v):%s\n", value, err)
		return err
	}
	return fs.Set(key, str)
}

func (fs *FileStorage) Delete(key string) {
    fs.Stat.DeleteCount ++
    if fs.Mode == FILE_STORAGE_MODE_SINGLE {
        delete(fs.Index, key)
    } else if fs.Mode == FILE_STORAGE_MODE_MULIPLE {
        delete(fs.Index, key)
        valueFilePath := path.Join(fs.Path, key)
        os.Remove(valueFilePath)
    }
}

func (fs *FileStorage) Count() int {
    return len(fs.Index)
}

func (fs *FileStorage) SaveIndex() error {
    indexPath := fs.getIndexFilePath()
    buff, err := json.Marshal(fs.Index)
    if err != nil {
        fmt.Printf("FileStorage.SaveIndex, Marshal json failed (%v):%s\n", indexPath, err)
        return err
    }
    fs.Stat.SaveIndexCount ++
    ioutil.WriteFile(indexPath, buff, 0644)
    return nil
}

func (fs *FileStorage) LoadIndex() error {
    indexPath := fs.getIndexFilePath()
    buff, err := ioutil.ReadFile(indexPath)
    if err != nil {
        fmt.Println("FileStorage.LoadIndex, Read file failed:", err)
        return err
    }
    if err := json.Unmarshal(buff, &fs.Index); err != nil {
        fmt.Printf("FileStorage.LoadIndex, Unmarshal json failed (%v):%s\n", indexPath, err)
        return err
    }
    return nil
}

func (fs *FileStorage) getIndexFilePath() string {
	var indexPath string
    if (fs.Mode == FILE_STORAGE_MODE_SINGLE) {
        indexPath = fs.Path
    } else if (fs.Mode == FILE_STORAGE_MODE_MULIPLE) {
        indexPath = path.Join(fs.Path, "index.json")
    }
    return indexPath
}


