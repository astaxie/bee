package main

import (
	"fmt"
	"os"
	path "path/filepath"
	"strings"
)

var cmdApiapp = &Command{
	// CustomFlags: true,
	UsageLine: "api [appname]",
	Short:     "create an api application base on beego framework",
	Long: `
create an api application base on beego framework

In the current path, will create a folder named [appname]

In the appname folder has the follow struct:

	├── conf
	│   └── app.conf
	├── controllers
	│   └── default.go
	├── main.go
	└── models
	    └── default.go             

`,
}

var apiconf = `
appname = {{.Appname}}
httpport = 8080
runmode = dev
autorender = false
`
var apiMaingo = `package main

import (
	"{{.Appname}}/controllers"
	"github.com/astaxie/beego"
)

func main() {
	beego.Router("/{{.Version}}/users/:objectId", &controllers.UserController{})
	beego.Router("/{{.Version}}/users", &controllers.UserController{})
	beego.Run()
}
`
var apiModels = `package models

type User struct {
	Id    interface{}
	Name  string
	Email string
}

func (this *User) One(id interface{}) (user User, err error) {
	// get user from DB
	// user, err = query(id)
	user = User{id, "astaxie", "astaxie@gmail.com"}
	return
}

func (this *User) All() (users []User, err error) {
	// get all users from DB
	// users, err = queryAll()
	users = append([]User{}, User{1, "astaxie", "astaxie@gmail.com"})
	users = append(users, User{2, "someone", "someone@gmail.com"})
	return
}

func (this *User) Update(id interface{}) (err error) {
	// user, err = update(id, this)
	return
}

`

var apiControllers = `package controllers

import (
	"encoding/json"
	"{{.Appname}}/models"
	"github.com/astaxie/beego"
	"io/ioutil"
)

type UserController struct {
	beego.Controller
}

func (this *UserController) Get() {
	var user models.User
	objectId := this.Ctx.Params[":objectId"]
	if objectId != "" {
		user, err := user.One(objectId)
		if err != nil {
			this.Data["json"] = err
		} else {
			this.Data["json"] = user
		}
	} else {
		users, err := user.All()
		if err != nil {
			this.Data["json"] = err
		} else {
			this.Data["json"] = users
		}
	}
	this.ServeJson()
}

func (this *UserController) Put() {
	defer this.ServeJson()
	var user models.User
	objectId := this.Ctx.Params[":objectId"]

	body, err := ioutil.ReadAll(this.Ctx.Request.Body)
	if err != nil {
		this.Data["json"] = err
		return
	}
	this.Ctx.Request.Body.Close()

	if err = json.Unmarshal(body, &user); err != nil {
		this.Data["json"] = err
		return
	}

	err = user.Update(objectId)
	if err != nil {
		this.Data["json"] = err
	} else {
		this.Data["json"] = "update success!"
	}
}

`

func init() {
	cmdApiapp.Run = createapi
}

func createapi(cmd *Command, args []string) {
	version := "1"
	if len(args) == 2 {
		version = args[1]
	} else if len(args) != 1 {
		fmt.Println("error args")
		os.Exit(2)
	}
	apppath, packpath, err := checkEnv(args[0])
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	os.MkdirAll(apppath, 0755)
	fmt.Println("create app folder:", apppath)
	os.Mkdir(path.Join(apppath, "conf"), 0755)
	fmt.Println("create conf:", path.Join(apppath, "conf"))
	os.Mkdir(path.Join(apppath, "controllers"), 0755)
	fmt.Println("create controllers:", path.Join(apppath, "controllers"))
	os.Mkdir(path.Join(apppath, "models"), 0755)
	fmt.Println("create models:", path.Join(apppath, "models"))

	fmt.Println("create conf app.conf:", path.Join(apppath, "conf", "app.conf"))
	writetofile(path.Join(apppath, "conf", "app.conf"),
		strings.Replace(apiconf, "{{.Appname}}", args[0], -1))

	fmt.Println("create controllers default.go:", path.Join(apppath, "controllers", "default.go"))
	writetofile(path.Join(apppath, "controllers", "default.go"),
		strings.Replace(apiControllers, "{{.Appname}}", packpath, -1))

	fmt.Println("create models default.go:", path.Join(apppath, "models", "default.go"))
	writetofile(path.Join(apppath, "models", "default.go"), apiModels)

	fmt.Println("create main.go:", path.Join(apppath, "main.go"))
	apiMaingo = strings.Replace(apiMaingo, "{{.Version}}", version, -1)
	writetofile(path.Join(apppath, "main.go"),
		strings.Replace(apiMaingo, "{{.Appname}}", packpath, -1))
}

func checkEnv(appname string) (apppath, packpath string, err error) {
	curpath, err := os.Getwd()
	if err != nil {
		return
	}

	gopath := os.Getenv("GOPATH")
	Debugf("gopath:%s", gopath)
	if gopath == "" {
		err = fmt.Errorf("you should set GOPATH in the env")
		return
	}

	appsrcpath := ""
	haspath := false
	wgopath := path.SplitList(gopath)
	for _, wg := range wgopath {
		wg = path.Join(wg, "src")

		if path.HasPrefix(strings.ToLower(curpath), strings.ToLower(wg)) {
			haspath = true
			appsrcpath = wg
			break
		}
	}

	if !haspath {
		err = fmt.Errorf("can't create application outside of GOPATH `%s`\n"+
			"you first should `cd $GOPATH%ssrc` then use create\n", gopath, string(path.Separator))
		return
	}
	apppath = path.Join(curpath, appname)

	if _, e := os.Stat(apppath); os.IsNotExist(e) == false {
		err = fmt.Errorf("path `%s` exists, can not create app without remove it\n", apppath)
		return
	}
	packpath = strings.Join(strings.Split(apppath[len(appsrcpath)+1:], string(path.Separator)), "/")
	return
}
