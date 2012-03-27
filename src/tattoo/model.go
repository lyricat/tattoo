package main

import (
	"html/template"
	"strings"
	"time"
)

type ArticleMetadata struct {
	Name           string
	Author         string
	IsPage		   bool
	Title          string
	Tags           []string
	FeaturedPicURL string
	Summary        string
	CreatedTime    int64
	ModifiedTime   int64
	Hits           int64
}

type Article struct {
	Metadata ArticleMetadata
	Text     template.HTML
	Comments []*Comment
}

func (meta *ArticleMetadata) CreatedTimeRFC3339() string {
	return TimeRFC3339(meta.CreatedTime)
}

func (meta *ArticleMetadata) ModifiedTimeRFC3339() string {
	return TimeRFC3339(meta.ModifiedTime)
}

func (meta *ArticleMetadata) CreatedTimeHumanReading() string {
	return TimeHumanReading(meta.CreatedTime)
}

func (meta *ArticleMetadata) GetCreatedTime() time.Time {
	t := time.Unix(meta.CreatedTime, 0).Local()
	return t
}

func (meta *ArticleMetadata) GetShortMonth(t1 time.Time) string {
	return t1.Month().String()[0:3]
}

func (meta *ArticleMetadata) ModifiedTimeHumanReading() string {
	return TimeHumanReading(meta.ModifiedTime)
}

func (meta *ArticleMetadata) TagRawList() string {
	return strings.Join(meta.Tags, ", ")
}

func (meta *ArticleMetadata) HasFeaturedPic() bool {
	if len(meta.FeaturedPicURL) == 0 {
		return false
	}
	return true
}

func (meta *ArticleMetadata) HasSummary() bool {
	if len(meta.Summary) == 0 {
		return false
	}
	return true
}

type CommentIndexItem struct {
	Name         string
	CommentNames []string
}

type CommentMetadata struct {
	Name        string
	Author      string
	ArticleName string
	UAgent      string
	URL         string
	IP          string
	Email       string
	EmailHash   string
	CreatedTime int64
}

func (meta *CommentMetadata) CreatedTimeHumanReading() string {
	return TimeHumanReading(meta.CreatedTime)
}

type Comment struct {
	Metadata CommentMetadata
	Text     template.HTML
}

type TagWrapper struct {
	Name  string
	Count int
}
