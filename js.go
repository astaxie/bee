package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var cmdJs = &Command{
	UsageLine: "js",
	Short:     "install js lib from network",
	Long:      `install js lib from network`,
}

var currentDir string

const (
	COMPONENTS_URL = "https://bower.herokuapp.com/packages"
)

func init() {
	currentDir, _ = os.Getwd()
	cmdJs.Run = jsSubCommand
}

func echoSubCommandUsage() {

	usage := `
Usage:

	bee js [command] [arguments]

The commands are:
    search  <packagename>
    install <packagename>
    list
    uninstall <packagename>

The arguments are:
    packagename
    `

	fmt.Println(usage)
}

func jsSubCommand(cmd *Command, args []string) {
	if len(args) < 1 {
		echoSubCommandUsage()
	} else {
		subCommand := args[0]
		manager := NewJsManager()
		manager.ParsePkgInfo(args[1:])

		if subCommand == "search" {
			manager.Search()
		} else if subCommand == "install" {
			manager.Install()
		} else if subCommand == "uninstall" {
			manager.Uninstall()
		} else if subCommand == "list" {
			manager.List()
		} else {
			echoSubCommandUsage()
		}
	}
}

type PkgInfo struct {
	Name    string
	Version string
	Url     string
}

type JsManager struct {
	searchInfos []PkgInfo
	pkgInfos    []PkgInfo
}

func NewJsManager() *JsManager {
	return &JsManager{}
}

func (this *JsManager) ParsePkgInfo(pkgs []string) {
	var searchInfos []PkgInfo

	var name string
	var version string
	for _, pkg := range pkgs {
		name = ""
		version = ""
		sp := strings.Split(pkg, "#")
		if len(sp) > 1 {
			name = sp[0]
			version = sp[1]
		} else {
			name = sp[0]
		}

		searchInfos = append(searchInfos, PkgInfo{Name: name, Version: version})
	}

	this.searchInfos = searchInfos
}

func (this *JsManager) search() {
	log.Println("Waiting...")
	req, err := http.NewRequest("GET", COMPONENTS_URL, nil)
	if err != nil {
		log.Fatalln(err.Error())
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalln(err.Error())
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err.Error())
	}

	pkgInfoList := make([]PkgInfo, 0)
	err = json.Unmarshal(body, &pkgInfoList)
	if err != nil {
		log.Fatalln(err.Error())
	}

	this.pkgInfos = pkgInfoList
	log.Printf("Found %d packages\n", len(pkgInfoList))
}

func (this JsManager) Search() {
	this.search()
	for _, pkgInfo := range this.pkgInfos {
		fmt.Println(pkgInfo.Name, pkgInfo.Url)
	}
}

func (this JsManager) Install() {
	this.search()
	var name string
	var url string
	for _, searchLib := range this.searchInfos {
		for _, pkgInfo := range this.pkgInfos {
			if pkgInfo.Name == searchLib.Name {
				name = pkgInfo.Name
				url = pkgInfo.Url
				break
			}
		}
	}

	targetDir := filepath.Join(currentDir, "static/js/", name)
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		os.MkdirAll(targetDir, 0755)
	} else {
		os.RemoveAll(targetDir)
	}

	clone := exec.Command("git", "clone", url, filepath.Join("./static/js/", name))
	clone.Stdout = os.Stdout
	clone.Stderr = os.Stderr
	clone.Run()
	fmt.Println("Download finished")
}

func (this JsManager) List() {
	files, _ := ioutil.ReadDir(filepath.Join(currentDir, "./static/js"))
	for _, file := range files {
		if file.IsDir() {
			bowerJson := filepath.Join(currentDir, "./static/js", file.Name(), "bower.json")
			buf, _ := ioutil.ReadFile(bowerJson)
			var pkgInfo PkgInfo
			err := json.Unmarshal(buf, &pkgInfo)
			if err != nil {
				log.Fatalln(err.Error())
			}
			log.Println(pkgInfo.Name + ":" + pkgInfo.Version)
		}
	}
}

func (this JsManager) Uninstall() {

	toRmDir := filepath.Join(currentDir, "./static/js/", this.searchInfos[0].Name)
	if _, err := os.Stat(toRmDir); err == nil {
		err = os.RemoveAll(toRmDir)
		if err != nil {
			log.Fatalln(err.Error())
		}
		log.Println("Remove pkg success")
	}

}
