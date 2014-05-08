package fs

import (
	"encoding/base64"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/koding/kite"
	"github.com/koding/kite/dnode"
)

var (
	fs     *kite.Kite
	remote *kite.Client
)

func init() {
	fs = kite.New("fs", "0.0.1")
	fs.Config.DisableAuthentication = true
	fs.Config.Port = 3636
	fs.HandleFunc("readDirectory", ReadDirectory)
	fs.HandleFunc("glob", Glob)
	fs.HandleFunc("readFile", ReadFile)
	fs.HandleFunc("writeFile", WriteFile)
	fs.HandleFunc("uniquePath", UniquePath)
	fs.HandleFunc("getInfo", GetInfo)
	fs.HandleFunc("setPermissions", SetPermissions)
	fs.HandleFunc("remove", Remove)
	fs.HandleFunc("rename", Rename)
	fs.HandleFunc("createDirectory", CreateDirectory)
	fs.HandleFunc("move", Move)
	fs.HandleFunc("copy", Copy)

	go fs.Run()
	<-fs.ServerReadyNotify()

	client := kite.New("client", "0.0.1")
	client.Config.DisableAuthentication = true
	remote = client.NewClientString("ws://127.0.0.1:3636")
	err := remote.Dial()
	if err != nil {
		log.Fatal("err")
	}
}

func TestReadDirectory(t *testing.T) {
	testDir := "."

	files, err := ioutil.ReadDir(testDir)
	if err != nil {
		t.Fatal(err)
	}

	currentFiles := make([]string, len(files))
	for i, f := range files {
		currentFiles[i] = f.Name()
	}

	resp, err := remote.Tell("readDirectory", struct {
		Path     string
		OnChange dnode.Function
	}{
		Path:     testDir,
		OnChange: dnode.Function{},
	})

	if err != nil {
		t.Fatal(err)
	}

	f, err := resp.Map()
	if err != nil {
		t.Fatal(err)
	}

	entries, err := f["files"].SliceOfLength(len(files))
	if err != nil {
		t.Fatal(err)
	}

	respFiles := make([]string, len(files))
	for i, e := range entries {
		f := &FileEntry{}
		err := e.Unmarshal(f)
		if err != nil {
			t.Fatal(err)
		}

		respFiles[i] = f.Name
	}

	if !reflect.DeepEqual(respFiles, currentFiles) {
		t.Error("got %+v, expected %+v", respFiles, currentFiles)
	}
}

func TestGlob(t *testing.T) {
	testGlob := "*"

	files, err := glob(testGlob)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := remote.Tell("glob", struct {
		Pattern string
	}{
		Pattern: testGlob,
	})
	if err != nil {
		t.Fatal(err)
	}

	var r []string
	err = resp.Unmarshal(&r)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(r, files) {
		t.Errorf("got %+v, expected %+v", r, files)
	}
}

func TestReadFile(t *testing.T) {
	testFile := "testdata/testfile1.txt"

	content, err := ioutil.ReadFile(testFile)
	if err != nil {
		t.Error(err)
	}

	resp, err := remote.Tell("readFile", struct {
		Path string
	}{
		Path: testFile,
	})
	if err != nil {
		t.Fatal(err)
	}

	buf := resp.MustMap()["content"].MustString()

	s, err := base64.StdEncoding.DecodeString(buf)
	if err != nil {
		t.Error(err)
	}

	if string(s) != string(content) {
		t.Errorf("got %s, expecting %s", string(s), string(content))
	}

}

func TestWriteFile(t *testing.T) {
	testFile, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(testFile.Name())

	content := []byte("hello kite")

	resp, err := remote.Tell("writeFile", struct {
		Path           string
		Content        []byte
		DoNotOverwrite bool
		Append         bool
	}{
		Path:    testFile.Name(),
		Content: content,
	})
	if err != nil {
		t.Fatal(err)
	}

	if int(resp.MustFloat64()) != len(content) {
		t.Errorf("content len is wrong. got %d expected %d", int(resp.MustFloat64()), len(content))
	}

	buf, err := ioutil.ReadFile(testFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(buf, content) {
		t.Errorf("content is wrong. got '%s' expected '%s'", string(buf), string(content))
	}
}

func TestUniquePath(t *testing.T)      {}
func TestGetInfo(t *testing.T)         {}
func TestSetPermissions(t *testing.T)  {}
func TestRemove(t *testing.T)          {}
func TestRename(t *testing.T)          {}
func TestCreateDirectory(t *testing.T) {}
func TestMove(t *testing.T)            {}
func TestCopy(t *testing.T)            {}
