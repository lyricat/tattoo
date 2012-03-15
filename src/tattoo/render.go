package main

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"webapp"
)

var mainTPL *template.Template
var writerTPL *template.Template
var guardTPL *template.Template
var feedTPL *template.Template
var editorTPL *template.Template
var notFoundTPL *template.Template

func LoadSystemTemplates() error {
	var err error
	// system templates
	writerTPL, err = template.ParseFiles(
		"sys/template/bare.html",
		"sys/template/nav.html",
		"sys/template/tags.html",
		"sys/template/comments.html",
		"sys/template/overview.html",
		"sys/template/content.html")
	if err != nil {
		return err
	}
	editorTPL, err = template.ParseFiles("sys/template/editor.html")
	if err != nil {
		return err
	}
	guardTPL, err = template.ParseFiles("sys/template/guard.html")
	if err != nil {
		return err
	}
	feedTPL, err = template.ParseFiles("sys/template/feed_atom.html")
	if err != nil {
		return err
	}
	return err
}

func LoadThemeTemplates(themeName string) error {
	var err error
	// required templates
	mainTPL, err = template.ParseFiles(
		fmt.Sprintf("theme/%s/template/bare.html", themeName),
		fmt.Sprintf("theme/%s/template/header.html", themeName),
		fmt.Sprintf("theme/%s/template/footer.html", themeName),
		fmt.Sprintf("theme/%s/template/tag.html", themeName),
		fmt.Sprintf("theme/%s/template/article.html", themeName),
		fmt.Sprintf("theme/%s/template/articles.html", themeName),
		fmt.Sprintf("theme/%s/template/content.html", themeName))
	if err != nil {
		return err
	}
	// optional templates
	notFoundTPL, err = template.ParseFiles(
		fmt.Sprintf("theme/%s/template/404.html", themeName))
	if err != nil {
		return err
	}
	return err
}

func RenderSinglePage(ctx *webapp.Context, name string, lastMeta *CommentMetadata) error {
	vars := make(map[string]interface{})
	vars["Name"] = name
	vars["LastCommentMeta"] = lastMeta
	data := MakeData(ctx, vars)
	data.Flags.Single = true
	err := ctx.Execute(mainTPL, &data)
	return err
}

func RenderTagPage(ctx *webapp.Context, offset int, tag string) error {
	vars := make(map[string]interface{})
	tag = strings.Trim(tag, " ")
	if !TattooDB.HasTag(tag) {
		return errors.New(webapp.ErrNotFound)
	}

	vars["Offset"] = offset
	vars["Tag"] = tag
	vars["AtBegin"] = false
	vars["AtEnd"] = false
	if TattooDB.GetTagArticleCount(tag)-1-offset < GetConfig().TimelineCount {
		vars["AtEnd"] = true
	}
	if offset < GetConfig().TimelineCount {
		vars["AtBegin"] = true
	}
	data := MakeData(ctx, vars)
	data.Flags.HasTag = true
	err := ctx.Execute(mainTPL, &data)
	return err
}

func RenderHomePage(ctx *webapp.Context, offset int) error {
	vars := make(map[string]interface{})
	vars["Offset"] = offset

	vars["AtBegin"] = false
	vars["AtEnd"] = false
	vars["Offset"] = offset
	if TattooDB.GetArticleCount()-1-offset < GetConfig().TimelineCount {
		vars["AtEnd"] = true
	}
	if offset < GetConfig().TimelineCount {
		vars["AtBegin"] = true
	}
	data := MakeData(ctx, vars)
	data.Flags.Home = true
	err := ctx.Execute(mainTPL, &data)
	return err
}

func RenderGuard(ctx *webapp.Context, hint string) error {
	vars := make(map[string]interface{})
	vars["Error"] = ""
	if len(hint) != 0 {
		vars["Error"] = hint
	}
	data := MakeData(ctx, vars)
	err := ctx.Execute(guardTPL, &data)
	return err
}

func RenderFeedAtom(ctx *webapp.Context) error {
	vars := make(map[string]interface{})
	vars["Declaration"] = template.HTML("<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
	data := MakeData(ctx, vars)
	data.Flags.Feed = true
	ctx.SetHeader("Content-Type", "application/atom+xml")
	err := ctx.Execute(feedTPL, &data)
	return err
}

func RenderWriterEditor(ctx *webapp.Context, article *Article) error {
	vars := make(map[string]interface{})
	vars["Article"] = article
	data := MakeData(ctx, vars)
	data.Flags.WriterEditor = true
	err := ctx.Execute(editorTPL, &data)
	return err
}

func RenderWriterOverview(ctx *webapp.Context, offset int) error {
	vars := make(map[string]interface{})
	vars["Offset"] = offset

	vars["AtBegin"] = false
	vars["AtEnd"] = false
	vars["Offset"] = offset
	if TattooDB.GetArticleCount()-1-offset < 20 {
		vars["AtEnd"] = true
	}
	if offset < 20 {
		vars["AtBegin"] = true
	}
	data := MakeData(ctx, vars)
	data.Flags.WriterOverview = true
	err := ctx.Execute(writerTPL, &data)
	return err
}

func RenderWriterComments(ctx *webapp.Context, offset int) error {
	vars := make(map[string]interface{})
	vars["Offset"] = offset

	vars["AtBegin"] = false
	vars["AtEnd"] = false
	vars["Offset"] = offset
	if TattooDB.GetCommentCount()-1-offset < 20 {
		vars["AtEnd"] = true
	}
	if offset < 20 {
		vars["AtBegin"] = true
	}
	data := MakeData(ctx, vars)
	data.Flags.WriterComments = true
	err := ctx.Execute(writerTPL, &data)
	return err
}

func Render404page(ctx *webapp.Context, msg string) error {
	if notFoundTPL != nil {
		vars := make(map[string]interface{})
		vars["Message"] = msg
		vars["URL"] = ctx.Request.RequestURI
		vars["Referer"] = ctx.Request.Referer()
		data := MakeData(ctx, vars)
		err := ctx.Execute(notFoundTPL, &data)
		return err
	} else {
		ctx.Error(fmt.Sprintf("%s: %s", webapp.ErrNotFound, msg),
			http.StatusNotFound)
		return nil
	}
	return nil
}

func RenderWriterSettings(ctx *webapp.Context) (string, error) {
	return "", nil
}
