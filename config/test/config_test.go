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

func (App) Prefix() string {
	return "app"
}

type Server struct {
	Addresses []string `yaml:"addresses" json:"addresses"`
}

func (Server) Prefix() string {
	return "server"
}

type NewApp struct {
}

func (NewApp) Prefix() string {
	return "app"
}

func TestConfig(t *testing.T) {
	var err error
	err = config.RegisterConfig(&App{})
	if err != nil {
		fmt.Println(err)
		return
	}

	appCfg := &App{}
	err = config.GetConfig(appCfg)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = config.RegisterConfig(&Server{})
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

	// srv2 := Server{}

	// err = config.GetConfig(srv2)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	err = config.RegisterConfig(NewApp{})
	if err != nil {
		fmt.Println(err)
		return
	}
}
