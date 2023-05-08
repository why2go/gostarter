package test

import (
	"fmt"
	"testing"

	"github.com/why2go/gostarter/config"
)

type App struct {
	AppName string `yaml:"app_name" json:"app_name"`
	Author  string `yaml:"author" json:"author"`
	Version string `yaml:"version" json:"version"`
}

func (App) ConfigName() string {
	return "app"
}

type App2 struct {
	AppName string `yaml:"app_name" json:"app_name"`
	Author  string `yaml:"author" json:"author"`
	Version string `yaml:"version" json:"version"`
}

func (App2) ConfigName() string {
	return "app"
}

type Server struct {
	Addresses []string `yaml:"addresses" json:"addresses"`
}

func (Server) ConfigName() string {
	return "server"
}

type NewApp struct {
}

func (NewApp) ConfigName() string {
	return "app"
}

func TestConfig(t *testing.T) {
	var err error

	appCfg := &App{}
	err = config.GetConfig(appCfg)
	if err != nil {
		fmt.Println(err)
		return
	}

	srv := &Server{}
	err = config.GetConfig(srv)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("appCfg: %v\n", appCfg)
	fmt.Printf("srv: %v\n", srv)

	srv2 := &Server{}
	err = config.GetConfig(srv2)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("srv2: %v\n", srv2)

	app2 := &App2{}
	err = config.GetConfig(app2)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("app2: %v\n", app2)
}
