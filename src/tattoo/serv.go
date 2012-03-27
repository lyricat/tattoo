package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"webapp"
)

const NOT_FOUND_MESSAGE = "Sorry, the page you were looking does not exist."

func isAuthorized(c *webapp.Context) bool {
	for _, cookie := range c.Request.Cookies() {
		if cookie.Name == "token" && cookie.Value == GetSessionToken() {
			return true
		}
	}
	return false
}
// Root Handler.
func HandleRoot(c *webapp.Context) {
	c.Info.UseGZip = strings.Index(c.Request.Header.Get("Accept-Encoding"), "gzip") > -1
	c.Info.StartTime = time.Now()

	urlPath := c.Request.URL.Path
	pathLevels := strings.Split(strings.Trim(urlPath, "/"), "/")
	if urlPath == "/" {
		// home page
		if HasTemplate("HOME") {
			HandleHome(c)
		} else {
			HandleArticles(c)
		}
	} else {
		if pathLevels[0] == "writer" {
			// writer
			HandleWriter(c, pathLevels)
		} else if pathLevels[0] == "guard" {
			// guard page
			HandleGuard(c)
		} else if pathLevels[0] == "comment" {
			// comment
			HandleComment(c)
		} else if pathLevels[0] == "feed" {
			// feed
			HandleFeed(c, pathLevels)
		} else if pathLevels[0] == "tag" {
			if len(pathLevels) >= 2 {
				// tag
				HandleTag(c, pathLevels[1])
			} else {
				Render404page(c, NOT_FOUND_MESSAGE)
			}
		} else {
			// single page
			HandleSingle(c, strings.ToLower(url.QueryEscape(pathLevels[0])))
		}
	}
}

func HandleHome(c *webapp.Context) {
	err := RenderHome(c)
	if err != nil {
		c.Error(fmt.Sprintf("%s: %s", webapp.ErrInternalServerError, err), http.StatusInternalServerError)
	}
}

func HandleArticles(c *webapp.Context) {
	pos, _ := strconv.Atoi(c.Request.FormValue("pos"))
	if pos > TattooDB.GetArticleCount()-1 {
		if HasTemplate("HOME") {
			c.Redirect("/post/", http.StatusFound)
		} else {
			c.Redirect("/", http.StatusFound)
		}
		return
	}
	err := RenderArticles(c, pos)
	if err != nil {
		c.Error(fmt.Sprintf("%s: %s", webapp.ErrInternalServerError, err), http.StatusInternalServerError)
	}
}

func HandleTag(c *webapp.Context, tag string) {
	tag = strings.Trim(tag, " ")
	if !TattooDB.HasTag(tag) {
		Render404page(c, NOT_FOUND_MESSAGE)
	}
	pos, _ := strconv.Atoi(c.Request.FormValue("pos"))
	if pos > TattooDB.GetTagArticleCount(tag)-1 {
		c.Redirect("/", http.StatusFound)
		return
	}
	err := RenderTagPage(c, pos, tag)
	if err != nil {
		c.Error(fmt.Sprintf("%s: %s",
			webapp.ErrInternalServerError, err),
			http.StatusInternalServerError)
	}
}

func HandleGuard(c *webapp.Context) {
	var err error
	action := c.Request.FormValue("action")
	if action == "logout" {
		RevokeSessionTokon()
		c.Redirect("/guard", http.StatusFound)
		return
	}
	if c.Request.Method == "POST" {
		cert := c.Request.FormValue("certificate")
		if len(cert) == 0 {
			c.Redirect("/guard", http.StatusFound)
			return
		}
		if SHA256Sum(cert) == GetConfig().Certificate {
			cookie := new(http.Cookie)
			cookie.Name = "token"
			cookie.Path = "/"
			cookie.Value = GenerateSessionToken()
			http.SetCookie(c.Writer, cookie)
			c.Redirect("/writer", http.StatusFound)
		} else {
			err = RenderGuard(c, "Your password is not correct")
			if err != nil {
				c.Error(fmt.Sprintf("%s: %s", webapp.ErrInternalServerError, err), http.StatusInternalServerError)
			}
		}
	} else if c.Request.Method == "GET" {
		err = RenderGuard(c, "")
		if err != nil {
			c.Error(fmt.Sprintf("%s: %s", webapp.ErrInternalServerError, err), http.StatusInternalServerError)
		}
	}
}

func HandleComment(c *webapp.Context) {
	if c.Request.Method == "POST" {
		IP := strings.Split(c.Request.RemoteAddr, ":")[0]
		comment := new(Comment)
		comment.Metadata.Name = UUID()
		comment.Metadata.IP = IP
		comment.Metadata.CreatedTime = time.Now().Unix()
		comment.Metadata.Author = strings.Trim(c.Request.FormValue("author"), " ")
		comment.Metadata.ArticleName = strings.Trim(c.Request.FormValue("article_name"), " /")
		comment.Metadata.UAgent = strings.Trim(c.Request.UserAgent(), " ")
		comment.Metadata.Email = strings.Trim(c.Request.FormValue("email"), " ")
		comment.Metadata.URL = strings.Trim(c.Request.FormValue("url"), " ")
		if strings.Index(comment.Metadata.URL, "http://") == -1 && strings.Index(comment.Metadata.URL, "https://") == -1 {
			comment.Metadata.URL = "http://" + comment.Metadata.URL;
		}
		comment.Text = template.HTML(strings.Trim(c.Request.FormValue("text"), " "))

		var cookie *http.Cookie

		expires := time.Date(2099, time.November, 10, 23, 0, 0, 0, time.UTC)
		cookie = new(http.Cookie)
		cookie.Path = "/"
		cookie.Expires = expires
		cookie.Name = "author"
		cookie.Value = comment.Metadata.Author
		http.SetCookie(c.Writer, cookie)
		cookie.Name = "email"
		cookie.Value = comment.Metadata.Email
		http.SetCookie(c.Writer, cookie)
		cookie.Name = "url"
		cookie.Value = comment.Metadata.URL
		http.SetCookie(c.Writer, cookie)

		// verify the form data
		if len(comment.Metadata.Author) == 0 || len(comment.Metadata.Email) < 3 || len(comment.Text) < 3 || len(comment.Metadata.Author) > 20 || len(comment.Metadata.Email) > 32 {
			c.Redirect("/"+comment.Metadata.ArticleName+"#respond", http.StatusFound)
			return
		}
		if !webapp.CheckEmailForm(comment.Metadata.Email) || (0 < len(comment.Metadata.URL) && !webapp.CheckURLForm(comment.Metadata.URL)) {
			c.Redirect("/"+comment.Metadata.ArticleName+"#respond", http.StatusFound)
			return
		}
		if !TattooDB.Has(comment.Metadata.ArticleName) {
			c.Redirect("/"+comment.Metadata.ArticleName+"#respond", http.StatusFound)
			return
		}
		comment.Text = template.HTML(webapp.TransformTags(string(comment.Text)))
		comment.Metadata.EmailHash = MD5Sum(comment.Metadata.Email)
		TattooDB.AddComment(comment)
		TattooDB.PrependCommentTimeline(comment)
		c.Redirect("/"+comment.Metadata.ArticleName+"#comment_"+comment.Metadata.Name, http.StatusFound)
	} else {
		c.Redirect("/"+c.Request.FormValue("article_name"), http.StatusFound)
	}
}

func HandleFeed(c *webapp.Context, pathLevels []string) {
	if len(pathLevels) < 2 {
		c.Redirect("/feed/atom", http.StatusFound)
		return
	}
	if pathLevels[1] == "atom" {
		var meta *ArticleMetadata
		var err error
		if len(TattooDB.ArticleTimeline) != 0 {
			meta, err = TattooDB.GetMetadata(TattooDB.ArticleTimeline[0])
			if err == nil {
				TattooDB.SetVar("LastUpdatedTime", TimeRFC3339(meta.ModifiedTime))
			}
		}
		err = RenderFeedAtom(c)
		if err != nil {
			c.Error(fmt.Sprintf("%s: %s", webapp.ErrInternalServerError, err), http.StatusInternalServerError)
			return
		}
	}
}

func HandleWriter(c *webapp.Context, pathLevels []string) {
	if ok := isAuthorized(c); !ok {
		c.Redirect("/guard", http.StatusFound)
		return
	}
	if c.Request.Method == "GET" {
		var err error
		if len(pathLevels) < 2 {
			c.Redirect("/writer/overview", http.StatusFound)
			return
		}
		if pathLevels[1] == "overview" {
			pos, _ := strconv.Atoi(c.Request.FormValue("pos"))
			if pos > TattooDB.GetArticleCount()-1 {
				c.Redirect("/writer/overview", http.StatusFound)
				return
			}
			err = RenderWriterOverview(c, pos)
		} else if pathLevels[1] == "comments" {
			pos, _ := strconv.Atoi(c.Request.FormValue("pos"))
			if pos > TattooDB.GetCommentCount()-1 {
				c.Redirect("/writer/comments", http.StatusFound)
				return
			}
			err = RenderWriterComments(c, pos)
		} else if pathLevels[1] == "settings" {
			err = RenderWriterSettings(c, "")
		} else if pathLevels[1] == "edit" {
			var article *Article = new(Article)
			var meta *ArticleMetadata = new(ArticleMetadata)
			var source []byte
			if len(pathLevels) >= 3 {
				name := strings.ToLower(url.QueryEscape(pathLevels[2]))
				meta, err = TattooDB.GetMetadata(name)
				if err == nil {
					source, err = TattooDB.GetArticleSource(name)
					if err == nil {
						article.Metadata = *meta
						article.Text = template.HTML(string(source))
					}
				}
			} else {
				article = new(Article)
			}
			err = RenderWriterEditor(c, article)
		} else if pathLevels[1] == "delete" {
			if len(pathLevels) >= 3 {
				name := strings.ToLower(url.QueryEscape(pathLevels[2]))
				if TattooDB.Has(name) {
					TattooDB.DeleteArticleTagIndex(name)
					TattooDB.DeleteArticle(name)
					TattooDB.DeleteMetadata(name)
					TattooDB.DeleteComments(name)
					TattooDB.Dump()
					TattooDB.RebuildTimeline()
					TattooDB.RebuildCommentTimeline()
				}
			}
			c.Redirect("/writer", http.StatusFound)
		} else if pathLevels[1] == "delete_comment" {
			if len(pathLevels) >= 3 {
				name := strings.ToLower(url.QueryEscape(pathLevels[2]))
				if TattooDB.HasComment(name) {
					TattooDB.DeleteComment(name)
					TattooDB.RebuildCommentTimeline()
				}
			}
			c.Redirect("/writer/comments", http.StatusFound)
		} else {
			Render404page(c, NOT_FOUND_MESSAGE)
		}
		if err != nil {
			c.Error(fmt.Sprintf("%s: %s", webapp.ErrInternalServerError, err), http.StatusInternalServerError)
		}
	} else if c.Request.Method == "POST" {
		if pathLevels[1] == "update" {
			HandleUpdateArticle(c)
		} else if pathLevels[1] == "settings" {
			HandleUpdateSystemSettings(c)
		} else {
			c.Redirect("/writer", http.StatusFound)
			return
		}
	}
}

func HandleUpdateArticle(c *webapp.Context) {
	isNew := false
	isRename := false
	origName := strings.Trim(c.Request.FormValue("orig_name"), " ")
	var err error
	var meta *ArticleMetadata
	article := new(Article)
	article.Metadata.Title = strings.Trim(c.Request.FormValue("title"), " ")
	article.Metadata.Name = strings.ToLower(strings.Trim(c.Request.FormValue("url"), " "))
	article.Metadata.FeaturedPicURL = strings.Trim(c.Request.FormValue("fpic"), " ")
	article.Metadata.Summary = strings.Trim(c.Request.FormValue("sum"), " ")
	article.Metadata.IsPage, err = strconv.ParseBool(c.Request.FormValue("ispage"))
	if err != nil {
		article.Metadata.IsPage = false
	}
	article.Metadata.Author = GetConfig().AuthorName
	article.Metadata.ModifiedTime = time.Now().Unix()
	article.Text = template.HTML(c.Request.FormValue("text"))

	if len(origName) == 0 {
		isNew = true
	} else if origName != article.Metadata.Name {
		isRename = true
	}
	if isNew {
		article.Metadata.CreatedTime = article.Metadata.ModifiedTime
	} else {
		meta, err = TattooDB.GetMetadata(origName)
		if err == nil {
			article.Metadata.CreatedTime = meta.CreatedTime
			article.Metadata.Hits = meta.Hits
		}
	}
	// check if the name is avaliable.
	meta, err = TattooDB.GetMetadata(article.Metadata.Name)
	if isNew && err == nil {
		article.Metadata.Name = ""
		err = RenderWriterEditor(c, article)
		return
	}
	// verify the form data
	if len(article.Metadata.Title) == 0 || len(article.Metadata.Name) == 0 {
		c.Redirect("/writer/edit", http.StatusFound)
		return
	}
	// handle tags
	tags := strings.Split(c.Request.FormValue("tags"), ",")
	tags_tmp := make(map[string]int)
	for _, t := range tags {
		tag := strings.ToLower(strings.Trim(t, " "))
		if len(tag) == 0 {
			continue
		}
		tags_tmp[tag] = 1
	}
	// remove duplication
	article.Metadata.Tags = make([]string, 0)
	for t := range tags_tmp {
		article.Metadata.Tags = append(article.Metadata.Tags, t)
	}
	// update tag index
	TattooDB.DeleteArticleTagIndex(article.Metadata.Name)
	TattooDB.UpdateArticleTagIndex(article.Metadata.Name, article.Metadata.Tags)
	// update metadata
	TattooDB.UpdateMetadata(&article.Metadata)
	TattooDB.UpdateArticle(article.Metadata.Name, []byte(string(article.Text)))
	if isRename {
		TattooDB.DeleteArticleTagIndex(origName)
		TattooDB.DeleteMetadata(origName)
		TattooDB.DeleteArticle(origName)
		TattooDB.RenameComments(origName, article.Metadata.Name)
	}
	TattooDB.Dump()
	if isNew || isRename {
		TattooDB.RebuildTimeline()
	}
	c.Redirect("/writer/overview", http.StatusFound)
	return
}

func HandleUpdateSystemSettings(c *webapp.Context) {
	portStr := strings.Trim(c.Request.FormValue("port"), " ")
	certificate := strings.Trim(c.Request.FormValue("certificate"), " ")
	sitebase := strings.Trim(c.Request.FormValue("sitebase"), " ")
	siteurl := strings.Trim(c.Request.FormValue("siteurl"), " ")
	sitetitle := strings.Trim(c.Request.FormValue("sitetitle"), " ")
	sitesubtitle := strings.Trim(c.Request.FormValue("sitesubtitle"), " ")
	path := strings.Trim(c.Request.FormValue("path"), " ")
	author := strings.Trim(c.Request.FormValue("author"), " ")
	timelinecountStr := strings.Trim(c.Request.FormValue("timelinecount"), " ")
	theme := strings.Trim(c.Request.FormValue("theme"), " ")
	// verify
	port, err := strconv.Atoi(portStr)
	if err != nil {
		RenderWriterSettings(c, "Port should be a positive integer!")
		return
	}
	timelinecount, err := strconv.Atoi(timelinecountStr)
	if err != nil {
		RenderWriterSettings(c, "Timeline Count should be a positive integer!")
		return
	}
	if err := LoadTheme(c.Application, theme); err != nil {
		RenderWriterSettings(c, fmt.Sprintf("Failed to load theme '%v': %v", theme, err))
		return
	}
	var newConfig Config
	newConfig.Port          = port
	newConfig.Certificate   = certificate
	newConfig.SiteBase      = sitebase
	newConfig.SiteURL       = siteurl
	newConfig.SiteTitle     = sitetitle
	newConfig.SiteSubTitle  = sitesubtitle
	newConfig.Path		    = path
	newConfig.AuthorName    = author
	newConfig.TimelineCount = timelinecount
	newConfig.ThemeName     = theme
	cfg := GetConfig()
	cfg.Update(&newConfig)
	cfg.Save()
	c.Redirect("/writer/settings", http.StatusFound)
}

func GetLastCommentMetadata(c *webapp.Context) (meta *CommentMetadata) {
	meta = new(CommentMetadata)
	for _, cookie := range c.Request.Cookies() {
		switch cookie.Name {
		case "author":
			meta.Author = cookie.Value
		case "email":
			meta.Email = cookie.Value
		case "url":
			meta.URL = cookie.Value
		}
	}
	return meta
}

func HandleSingle(c *webapp.Context, pagename string) {
	if TattooDB.Has(pagename) {
		lastMeta := GetLastCommentMetadata(c)
		err := RenderSinglePage(c, pagename, lastMeta)
		if err != nil {
			c.Error(fmt.Sprintf("%s: %s", webapp.ErrInternalServerError, err), http.StatusInternalServerError)
		}
		meta, err := TattooDB.GetMetadata(pagename)
		if err == nil {
			meta.Hits += 1
			TattooDB.UpdateMetadata(meta)
		}
	} else {
		Render404page(c, NOT_FOUND_MESSAGE)
	}
}
