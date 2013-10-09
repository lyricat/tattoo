package main

import (
	"errors"
	"github.com/russross/blackfriday"
	"github.com/shellex/tattoo/webapp"
	"html/template"
	"log"
	"sort"
)

type TattooStorage struct {
	ArticleDB            webapp.FileStorage
	ArticleHTMLDB        webapp.FileStorage
	MetadataDB           webapp.FileStorage
	CommentDB            webapp.FileStorage
	CommentHTMLDB        webapp.FileStorage
	CommentMetadataDB    webapp.FileStorage
	CommentIndexDB       webapp.FileStorage
	TagIndexDB           webapp.FileStorage
	VarDB                webapp.FileStorage
	ArticleTimeline      []string
	ArticleTimelineIndex map[string]int
	PageTimeline         []string
	CommentTimeline      []string
}

var TattooDB *TattooStorage = nil

func init() {
	TattooDB = new(TattooStorage)
}

func (db *TattooStorage) Load(app *webapp.App) {
	app.Log("Tattoo DB", "Init DB: Article DB")
	db.ArticleDB.Init("storage/source/", webapp.FILE_STORAGE_MODE_MULIPLE)
	app.Log("Tattoo DB", "Init DB: Article HTML DB")
	db.ArticleHTMLDB.Init("storage/html/", webapp.FILE_STORAGE_MODE_MULIPLE)
	app.Log("Tattoo DB", "Init DB: Article Metadata DB")
	db.MetadataDB.Init("storage/metadata/", webapp.FILE_STORAGE_MODE_MULIPLE)
	app.Log("Tattoo DB", "Init DB: Vars DB")
	db.VarDB.Init("storage/var.json", webapp.FILE_STORAGE_MODE_SINGLE)

	app.Log("Tattoo DB", "Init DB: Comment DB")
	db.CommentDB.Init("storage/comment_source/", webapp.FILE_STORAGE_MODE_MULIPLE)
	app.Log("Tattoo DB", "Init DB: Comment HTML DB")
	db.CommentHTMLDB.Init("storage/comment_html/", webapp.FILE_STORAGE_MODE_MULIPLE)
	app.Log("Tattoo DB", "Init DB: Comment Metadata DB")
	db.CommentMetadataDB.Init("storage/comment_metadata/", webapp.FILE_STORAGE_MODE_MULIPLE)
	app.Log("Tattoo DB", "Init DB: Comment Index DB")
	db.CommentIndexDB.Init("storage/comment_index/", webapp.FILE_STORAGE_MODE_MULIPLE)

	app.Log("Tattoo DB", "Init DB: Tag Index DB")
	db.TagIndexDB.Init("storage/tag_index/", webapp.FILE_STORAGE_MODE_MULIPLE)

	app.Log("Tattoo DB", "Rebuild Article Timeline")
	db.RebuildTimeline()
	app.Log("Tattoo DB", "Rebuild Comment Timeline")
	db.RebuildCommentTimeline()
}

// TattooStorage.RebuildTimeline rebuilds an array Tattoo.ArticleTimeline
// which contains all articles' name, order by created time.
// And builds a mapping from articles' name to the position of according
// articles.
func (s *TattooStorage) RebuildTimeline() {
	s.ArticleTimeline = make([]string, 0)
	s.ArticleTimelineIndex = make(map[string]int)
	s.PageTimeline = make([]string, 0)
	tmp_a := new(KeyPairs)
	tmp_a.Items = make([]*KeyValuePair, 0)
	tmp_p := new(KeyPairs)
	tmp_p.Items = make([]*KeyValuePair, 0)
	var metadata *ArticleMetadata
	var err error
	for name, _ := range s.ArticleDB.Index {
		if name == "*" {
			continue
		}
		metadata, err = s.GetMeta(name)
		if err != nil {
			log.Printf("%v\n", err)
		}
		if metadata.IsPage {
			tmp_p.Items = append(tmp_p.Items, &KeyValuePair{Key: metadata.CreatedTime, Value: name})
		} else {
			tmp_a.Items = append(tmp_a.Items, &KeyValuePair{Key: metadata.CreatedTime, Value: name})
		}
	}
	sort.Sort(tmp_a)
	for i, j := 0, len(tmp_a.Items)-1; i < j; i, j = i+1, j-1 {
		tmp_a.Items[i], tmp_a.Items[j] = tmp_a.Items[j], tmp_a.Items[i]
	}
	sort.Sort(tmp_p)
	for i, j := 0, len(tmp_p.Items)-1; i < j; i, j = i+1, j-1 {
		tmp_p.Items[i], tmp_p.Items[j] = tmp_p.Items[j], tmp_p.Items[i]
	}
	for i := range tmp_a.Items {
		s.ArticleTimeline = append(s.ArticleTimeline, tmp_a.Items[i].Value)
		s.ArticleTimelineIndex[tmp_a.Items[i].Value] = i
	}
	for i := range tmp_p.Items {
		s.PageTimeline = append(s.PageTimeline, tmp_p.Items[i].Value)
	}
}

// TattooStorage.RebuildCommentTimeline rebuilds an array Tattoo.CommentTimeline
// which contains all articles' name, order by created time.
func (s *TattooStorage) RebuildCommentTimeline() {
	num := s.CommentDB.Count() - 1
	s.CommentTimeline = make([]string, num, num+1024)
	tmp := new(KeyPairs)
	tmp.Items = make([]*KeyValuePair, num)
	i := 0
	var metadata *CommentMetadata
	var err error
	for name, _ := range s.CommentDB.Index {
		if name == "*" {
			continue
		}
		metadata, err = s.GetCommentMetadata(name)
		if err != nil {
			log.Printf("err: %v\n", err)
		}
		tmp.Items[i] = &KeyValuePair{Key: metadata.CreatedTime, Value: name}
		i += 1
	}
	sort.Sort(tmp)
	// reverse
	for i, j := 0, len(tmp.Items)-1; i < j; i, j = i+1, j-1 {
		tmp.Items[i], tmp.Items[j] = tmp.Items[j], tmp.Items[i]
	}
	// copy
	for i = range tmp.Items {
		s.CommentTimeline[i] = tmp.Items[i].Value
	}
}

// TattooStorage.PrependCommentTimeline prepends the name of comment to current Timeline.
func (s *TattooStorage) PrependCommentTimeline(comment *Comment) {
	num := s.CommentDB.Count() - 1
	tmp := make([]string, num)
	copy(tmp[0:num], s.CommentTimeline[:])
	s.CommentTimeline = make([]string, num, num+1024)
	copy(s.CommentTimeline[1:num], tmp[0:num])
	s.CommentTimeline[0] = comment.Metadata.Name
}

// TattooStorage.DeleteComment deletes the name of specified comment from the timeline.
func (s *TattooStorage) DeleteCommentTimeline(comment *Comment) {
	num := s.CommentDB.Count() - 1
	idx := -1
	for i, j := range s.CommentTimeline {
		if j == comment.Metadata.Name {
			idx = i
			break
		}
	}
	if idx != -1 {
		copy(s.CommentTimeline[idx:num-1], s.CommentTimeline[idx+1:num])
		s.CommentTimeline[num] = ""
	}
}

// TattooStorage.Has checks if an article with specified name dosen't exists in the storage.
func (s *TattooStorage) Has(name string) bool {
	if s.ArticleDB.Has(name) {
		return true
	}
	return false
}

// TattooStorage.GetMetaJSON gets the JSON format metadata of an article specified by the name.
func (db *TattooStorage) GetMetaJSON(name string) (interface{}, error) {
	if !db.MetadataDB.Has(name) {
		return nil, errors.New(webapp.ErrNotFound)
	}
	meta, _ := db.MetadataDB.GetJSON(name)
	return meta, nil
}

// TattooStorage.GetMeta gets the metadata of an article specified by the name.
func (db *TattooStorage) GetMeta(name string) (*ArticleMetadata, error) {
	raw, err := db.GetMetaJSON(name)
	if err != nil {
		return nil, err
	}
	meta := new(ArticleMetadata)
	meta.BuildFromJson(raw)
	return meta, nil
}

// TattooStorage.UpdateMetadata updates a specified metadata
func (s *TattooStorage) UpdateMetadata(meta *ArticleMetadata) {
	name := meta.Name
	s.MetadataDB.SetJSON(name, meta)
	s.MetadataDB.SaveIndex()
}

// TattooStorage.DeleteMetadata deletes metadata by a specified name.
func (s *TattooStorage) DeleteMetadata(name string) {
	s.MetadataDB.Delete(name)
	s.MetadataDB.SaveIndex()
}

// TattooStorage.Dump saves all Indexes of article dbs
func (s *TattooStorage) Dump() {
	s.MetadataDB.SaveIndex()
	s.ArticleDB.SaveIndex()
	s.ArticleHTMLDB.SaveIndex()
}

// TattooStorage.DumpComment saves all Indexes of comment dbs
func (s *TattooStorage) DumpComment() {
	s.CommentIndexDB.SaveIndex()
	s.CommentMetadataDB.SaveIndex()
	s.CommentDB.SaveIndex()
	s.CommentHTMLDB.SaveIndex()
}

func (s *TattooStorage) GetArticle(name string) ([]byte, error) {
	return s.ArticleHTMLDB.Get(name)
}

func (s *TattooStorage) GetPrevArticleName(name string) string {
	idx := s.ArticleTimelineIndex[name]
	if idx == 0 {
		return ""
	}
	return s.ArticleTimeline[idx-1]
}

func (s *TattooStorage) GetNextArticleName(name string) string {
	idx := s.ArticleTimelineIndex[name]
	if idx == len(s.ArticleTimeline)-1 {
		return ""
	}
	return s.ArticleTimeline[idx+1]
}

func (s *TattooStorage) GetArticleSource(name string) ([]byte, error) {
	return s.ArticleDB.Get(name)
}

func (s *TattooStorage) GetFullArticleCount() int {
	return s.ArticleDB.Count()
}

func (s *TattooStorage) GetArticleCount() int {
	return len(s.ArticleTimeline)
}

func (s *TattooStorage) GetPageCount() int {
	return len(s.PageTimeline)
}

func (s *TattooStorage) GetArticleFull(name string) (*Article, error) {
	var err error
	var meta *ArticleMetadata
	ret := new(Article)
	meta, err = s.GetMeta(name)
	if err != nil {
		return nil, err
	}
	ret.Metadata = *meta
	text, err := s.GetArticle(name)
	ret.Text = template.HTML(text)
	return ret, err
}

func (s *TattooStorage) GetArticleTimeline(from int, count int) ([]*Article, error) {
	if from < 0 || from > len(s.ArticleTimeline)-1 {
		from = 0
	}
	if from+count > len(s.ArticleTimeline) {
		count = len(s.ArticleTimeline) - from
	}
	var err error
	var meta *ArticleMetadata
	var text []byte
	tlSlice := s.ArticleTimeline[from : from+count]
	ret := make([]*Article, count)
	for i := 0; i < count; i += 1 {
		name := tlSlice[i]
		ret[i] = new(Article)
		meta, err = s.GetMeta(name)
		ret[i].Metadata = *meta
		text, err = s.GetArticle(name)
		ret[i].Text = template.HTML(text)
	}
	return ret, err
}

func (s *TattooStorage) GetArticleTimelineByTag(from int, count int, tag string) ([]*Article, error) {
	// @TODO
	if from < 0 || from > len(s.ArticleTimeline)-1 {
		from = 0
	}
	if from+count > len(s.ArticleTimeline) {
		count = len(s.ArticleTimeline) - from
	}
	var err error
	var meta *ArticleMetadata
	var text []byte
	tlSlice := s.ArticleTimeline
	ret := make([]*Article, 0)
	for i := 0; i < len(s.ArticleTimeline); i += 1 {
		name := tlSlice[i]
		meta, err = s.GetMeta(name)
		for _, t := range meta.Tags {
			if tag == t {
				a := new(Article)
				a.Metadata = *meta
				text, err = s.GetArticle(name)
				a.Text = template.HTML(text)
				ret = append(ret, a)
				break
			}
		}
		if len(ret) > count+from {
			break
		}
	}
	if from+count > len(ret) {
		count = len(ret) - from
	}
	return ret[from : from+count], err
}

func (s *TattooStorage) GetPageTimeline(from int, count int) ([]*Article, error) {
	if from < 0 || from > len(s.PageTimeline)-1 {
		from = 0
	}
	if from+count > len(s.PageTimeline) {
		count = len(s.PageTimeline) - from
	}
	var err error
	var meta *ArticleMetadata
	var text []byte
	tlSlice := s.PageTimeline[from : from+count]
	ret := make([]*Article, count)
	for i := 0; i < count; i += 1 {
		name := tlSlice[i]
		ret[i] = new(Article)
		meta, err = s.GetMeta(name)
		ret[i].Metadata = *meta
		text, err = s.GetArticle(name)
		ret[i].Text = template.HTML(text)
	}
	return ret, err
}

func (s *TattooStorage) UpdateArticle(name string, text []byte) {
	s.ArticleDB.Set(name, text)
	md := blackfriday.MarkdownCommon(text)
	s.ArticleHTMLDB.Set(name, md)
}

func (s *TattooStorage) DeleteArticle(name string) {
	s.ArticleDB.Delete(name)
	s.ArticleHTMLDB.Delete(name)
}

// simple add an item to Tag Index DB if the tag doesn't exists
func (s *TattooStorage) AddTag(tagName string) {
	s.TagIndexDB.SetJSON(tagName, "[]")
	return
}

// for each article use the tag, update their metadata.
// and update the Tag Index DB.
// @TODO
func (s *TattooStorage) RenameTag(origName string, newName string) {
	var lst []interface{}
	var lst_buff interface{}
	lst_buff, err := s.TagIndexDB.GetJSON(origName)
	if err != nil {
		println("load tag index failed", err)
	} else {
		lst = lst_buff.([]interface{})
	}
	newList := make([]string, 0)
	for _, k := range lst {
		meta, err := s.GetMeta(k.(string))
		if err != nil {
			continue
		}
		for i, v := range meta.Tags {
			if v == origName {
				meta.Tags[i] = newName
			}
		}
		s.UpdateMetadata(meta)
		newList = append(newList, k.(string))
	}
	s.TagIndexDB.SetJSON(newName, newList)
	s.TagIndexDB.Delete(origName)
	s.TagIndexDB.SaveIndex()
	return
}

func (s *TattooStorage) GetTagArticleCount(tagName string) int {
	lst_buff, err := s.TagIndexDB.GetJSON(tagName)
	if err != nil {
		return 0
	}
	lst := lst_buff.([]interface{})
	return len(lst)
}

func (s *TattooStorage) HasTag(tagName string) bool {
	return s.TagIndexDB.Has(tagName)
}

// for each article use the tag, update their metadata.
// and remove it from the Tag Index DB.
func (s *TattooStorage) DeleteTag(tagName string) {
	return
}

func (s *TattooStorage) GetTags() []TagWrapper {
	tmp := new(KeyPairs)
	tmp.Items = make([]*KeyValuePair, 0)
	ret := make([]TagWrapper, 0)
	for name, _ := range s.TagIndexDB.Index {
		if name == "*" {
			continue
		}
		lst_buff, _ := s.TagIndexDB.GetJSON(name)
		count := len(lst_buff.([]interface{}))
		tmp.Items = append(tmp.Items, &KeyValuePair{Key: int64(count), Value: name})
	}
	sort.Sort(tmp)
	for i, j := 0, len(tmp.Items)-1; i < j; i, j = i+1, j-1 {
		tmp.Items[i], tmp.Items[j] = tmp.Items[j], tmp.Items[i]
	}
	for _, t := range tmp.Items {
		ret = append(ret, TagWrapper{Name: t.Value, Count: int(t.Key)})
	}
	return ret
}

// assigns several tags to a specified article
func (s *TattooStorage) UpdateArticleTagIndex(name string, tags []string) {
	// assign
	for _, t := range tags {
		if s.TagIndexDB.Has(t) {
			lst_buff, _ := s.TagIndexDB.GetJSON(t)
			articleList := lst_buff.([]interface{})
			newList := make([]string, 0)
			newList = append(newList, name)
			for _, n := range articleList {
				if n.(string) == name {
					continue
				}
				newList = append(newList, n.(string))
			}
			s.TagIndexDB.SetJSON(t, newList)
		} else {
			s.TagIndexDB.SetJSON(t, []string{})
		}
	}
	s.TagIndexDB.SaveIndex()
	return
}

// assigns several tags to a specified article
func (s *TattooStorage) DeleteArticleTagIndex(name string) {
	// detach
	meta, err := s.GetMeta(name)
	if err != nil {
		return
	}
	for _, t := range meta.Tags {
		if s.TagIndexDB.Has(t) {
			lst_buff, _ := s.TagIndexDB.GetJSON(t)
			articleList := lst_buff.([]interface{})
			newList := make([]string, 0)
			for _, n := range articleList {
				if n.(string) != name {
					newList = append(newList, n.(string))
				}
			}
			s.TagIndexDB.SetJSON(t, newList)
		}
	}
	s.TagIndexDB.SaveIndex()
}

func (s *TattooStorage) GetCommentTimeline(from int, count int) ([]*Comment, error) {
	if from < 0 || from > len(s.CommentTimeline)-1 {
		from = 0
	}
	if from+count > len(s.CommentTimeline) {
		count = len(s.CommentTimeline) - from
	}
	var err error
	var text []byte
	var meta *CommentMetadata
	tlSlice := s.CommentTimeline[from : from+count]
	ret := make([]*Comment, count)
	for i := 0; i < count; i += 1 {
		name := tlSlice[i]
		ret[i] = new(Comment)
		meta, err = s.GetCommentMetadata(name)
		ret[i].Metadata = *meta
		text, err = s.GetComment(name)
		ret[i].Text = template.HTML(text)
	}
	return ret, err
}

func (s *TattooStorage) HasComment(uuid string) bool {
	if s.CommentDB.Has(uuid) {
		return true
	}
	return false
}

// TattooStorage.GetCommentMetaJSON gets the JSON format metadata of a comment specified by the name.
func (db *TattooStorage) GetCommentMetaJSON(name string) (interface{}, error) {
	if !db.CommentMetadataDB.Has(name) {
		return nil, errors.New(webapp.ErrNotFound)
	}
	meta, _ := db.CommentMetadataDB.GetJSON(name)
	return meta, nil
}

// TattooStorage.GetCommentMetaJSON gets the metadata of a comment specified by the name.
func (db *TattooStorage) GetCommentMetadata(name string) (*CommentMetadata, error) {
	raw, err := db.GetCommentMetaJSON(name)
	if err != nil {
		return nil, err
	}
	meta := new(CommentMetadata)
	meta.BuildFromJson(raw)
	return meta, nil
}

func (s *TattooStorage) UpdateCommentMetadata(meta *CommentMetadata) {
	uuid := meta.Name
	s.CommentMetadataDB.SetJSON(uuid, meta)
	s.CommentMetadataDB.SaveIndex()
}

func (s *TattooStorage) DeleteCommentMetadata(uuid string) {
	s.CommentMetadataDB.Delete(uuid)
	s.CommentMetadataDB.SaveIndex()
}

func (s *TattooStorage) GetComment(uuid string) ([]byte, error) {
	return s.CommentHTMLDB.Get(uuid)
}

func (s *TattooStorage) GetCommentSource(uuid string) ([]byte, error) {
	return s.CommentDB.Get(uuid)
}

func (s *TattooStorage) UpdateComment(uuid string, text []byte) {
	s.CommentDB.Set(uuid, text)
	md := blackfriday.MarkdownCommon(text)
	s.CommentHTMLDB.Set(uuid, md)
}

/* High level operation about comment */

// DeleteComments deletes all comments under an article.
func (s *TattooStorage) DeleteComments(name string) {
	var lst_buff interface{}
	var lst []interface{}
	lst_buff, err := s.CommentIndexDB.GetJSON(name)
	if err != nil {
		log.Printf("load comment index failed (%v)!\n", err)
	} else {
		lst = lst_buff.([]interface{})
	}
	for _, k := range lst {
		s.DeleteComment(k.(string))
		s.DeleteCommentMetadata(k.(string))
	}
	s.CommentDB.SaveIndex()
	s.CommentHTMLDB.SaveIndex()
	s.CommentMetadataDB.SaveIndex()
	s.CommentIndexDB.Delete(name)
	s.CommentIndexDB.SaveIndex()
}

// RenameComments updates the .ArticleName field in all comments' meta under an article.
func (s *TattooStorage) RenameComments(origName, newName string) {
	var lst []interface{}
	var lst_buff interface{}
	lst_buff, err := s.CommentIndexDB.GetJSON(origName)
	if err != nil {
		log.Printf("load comment index failed (%v)!\n", err)
	} else {
		lst = lst_buff.([]interface{})
	}
	newList := make([]string, 0)
	for _, k := range lst {
		meta, err := s.GetCommentMetadata(k.(string))
		if err != nil {
			continue
		}
		meta.ArticleName = newName
		s.UpdateCommentMetadata(meta)
		newList = append(newList, k.(string))
	}
	s.CommentIndexDB.SetJSON(newName, newList)
	s.CommentIndexDB.Delete(origName)
	s.CommentIndexDB.SaveIndex()
	s.RebuildCommentTimeline()
}

// DeleteComment delete meta and content of a specified comment.
// And remove its uuid from Comment Index DB.
func (s *TattooStorage) DeleteComment(uuid string) {
	var lst []interface{}
	var lst_buff interface{}
	var err error
	var meta *CommentMetadata
	meta, err = s.GetCommentMetadata(uuid)
	if err != nil {
		return
	}
	lst_buff, err = s.CommentIndexDB.GetJSON(meta.ArticleName)
	if err != nil {
		log.Printf("load comment index failed: %v\n", err)
		return
	} else {
		lst = lst_buff.([]interface{})
	}
	newList := make([]string, 0)
	for _, k := range lst {
		if uuid != k {
			newList = append(newList, k.(string))
		}
	}
	s.CommentIndexDB.SetJSON(meta.ArticleName, newList)
	s.CommentIndexDB.SaveIndex()
	// save meta & text
	s.DeleteCommentMetadata(uuid)
	s.CommentDB.Delete(uuid)
	s.CommentHTMLDB.Delete(uuid)
	s.CommentMetadataDB.SaveIndex()
	s.CommentDB.SaveIndex()
	s.CommentHTMLDB.SaveIndex()
}

// AddComment adds a new comment, both meta and content.
// Also adds an item in Comment Index DB
func (s *TattooStorage) AddComment(comment *Comment) {
	var lst []interface{}
	var lst_buff interface{}
	lst_buff, err := s.CommentIndexDB.GetJSON(comment.Metadata.ArticleName)
	if err != nil {
		println("load comment index failed", err)
	} else {
		lst = lst_buff.([]interface{})
	}
	newList := make([]string, len(lst)+1)
	for i, k := range lst {
		newList[i] = k.(string)
	}
	newList[len(lst)] = comment.Metadata.Name
	s.CommentIndexDB.SetJSON(comment.Metadata.ArticleName, newList)
	s.CommentIndexDB.SaveIndex()
	// save meta & text
	s.UpdateCommentMetadata(&comment.Metadata)
	s.UpdateComment(comment.Metadata.Name, []byte(string(comment.Text)))
	s.CommentMetadataDB.SaveIndex()
	s.CommentDB.SaveIndex()
	s.CommentHTMLDB.SaveIndex()
}

// GetComments get all comments of a specified article.
func (s *TattooStorage) GetComments(name string) []*Comment {
	var lst_buff interface{}
	var lst []interface{}
	var err error
	var text []byte
	lst_buff, err = s.CommentIndexDB.GetJSON(name)
	if err != nil {
		log.Printf("load comment index failed (%v)!\n", err)
		return nil
	}
	lst = lst_buff.([]interface{})
	arr := make([]*Comment, len(lst))
	for i, k := range lst {
		var meta *CommentMetadata
		meta, err = s.GetCommentMetadata(k.(string))
		if err != nil {
			log.Printf("Get Comment Medata Failed (%s)!\n", k.(string))
			continue
		}
		comment := new(Comment)
		comment.Metadata = *meta
		text, err = s.CommentHTMLDB.Get(k.(string))
		comment.Text = template.HTML(text)
		arr[i] = comment
	}
	return arr
}

// GetArticleCommentCount gets the comment count of a specified article.
func (s *TattooStorage) GetArticleCommentCount(name string) int {
	lst_buff, err := s.CommentIndexDB.GetJSON(name)
	if err != nil {
		log.Printf("load comment index failed (%v)!\n", err)
		return 0
	}
	return len(lst_buff.([]interface{}))
}

func (s *TattooStorage) GetCommentCount() int {
	return s.CommentDB.Count()
}

func (s *TattooStorage) SetVar(name string, value string) {
	s.VarDB.SetString(name, value)
}

func (s *TattooStorage) GetVar(name string) (string, error) {
	ret, err := s.VarDB.GetString(name)
	return ret, err
}
