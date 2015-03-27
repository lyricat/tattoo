package main

import (
	"github.com/shellex/tattoo/webapp"
	"time"
)

type T_FLAGS struct {
	Home     bool
	Articles bool
	Single   bool
	Tag      bool
	Page     bool
	Feed     bool

	WriterOverview bool
	WriterPages    bool
	WriterTags     bool
	WriterComments bool
	WriterSettings bool
	WriterEditor   bool
}

type T_DATA struct {
	Fn          Export
	Flags       T_FLAGS
	SiteConfig  Config
	ContextInfo webapp.ContextInfo
	Vars        interface{}
}

func MakeData(ctx *webapp.Context, vars interface{}) T_DATA {
	config := GetConfig()
	ctx.Info.During = time.Now().Sub(ctx.Info.StartTime).Nanoseconds() / 1000.0
	ctx.Info.URL = ctx.Request.URL.Path
	data := T_DATA{
		SiteConfig:  *config,
		ContextInfo: ctx.Info,
		Vars:        vars,
	}
	return data
}
