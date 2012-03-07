package main

type Export int

func (e * Export) GetPrevArticleName(name string) string {
    return TattooDB.GetPrevArticleName(name)
}

func (e * Export) GetNextArticleName(name string) string {
    return TattooDB.GetNextArticleName(name)
}

func (e * Export) GetPrevTLPos(offset int, count int) int {
	if count <= 0 {
		count = GetConfig().TimelineCount
	}
	prev := offset - count
    if prev < 0 {
        return 0
    }
    if prev > TattooDB.GetArticleCount() - 1 {
        return TattooDB.GetArticleCount() - 1
    }
    return prev
}

func (e * Export) GetNextTLPos(offset int, count int) int {
	if count <= 0 {
		count = GetConfig().TimelineCount
	}
    next := offset + count
    if next < GetConfig().TimelineCount {
        return 0
    }
    if next > TattooDB.GetArticleCount() - 1 {
        return TattooDB.GetArticleCount() - 1
    }
    return next
}

func (e * Export) GetPrevCommentTLPos(offset int, count int) int {
	if count <= 0 {
		count = GetConfig().TimelineCount
	}
	prev := offset - count
    if prev < 0 {
        return 0
    }
    if prev > TattooDB.GetCommentCount() - 1 {
        return TattooDB.GetCommentCount() - 1
    }
    return prev
}

func (e * Export) GetNextCommentTLPos(offset int, count int) int {
	if count <= 0 {
		count = GetConfig().TimelineCount
	}
    next := offset + count
    if next < GetConfig().TimelineCount {
        return 0
    }
    if next > TattooDB.GetCommentCount() - 1 {
        return TattooDB.GetCommentCount() - 1
    }
    return next
}


func (e * Export) GetCommentTimeline(offset int, count int) []*Comment {
    comments, _ := TattooDB.GetCommentTimeline(offset, count)
    return comments
}

func (e * Export) GetArticleCommentCount(name string) int {
    return TattooDB.GetArticleCommentCount(name)
}

func (e * Export) GetArticleTimeline(offset int, count int) []*Article {
	articles, _ := TattooDB.GetArticleTimeline(offset, count)
    return articles
}

func (e * Export) GetArticleTimelineByTag(offset int, count int, tag string) []*Article {
	articles, _ := TattooDB.GetArticleTimelineByTag(offset, count, tag)
    return articles
}

func (e * Export) GetArticle(name string) *Article {
    article, _ := TattooDB.GetArticleFull(name)
    return article
}

func (e * Export) GetArticleComments(name string) []*Comment {
    return TattooDB.GetComments(name)
}

func (e * Export) GetArticleMetadata(name string) *ArticleMetadata {
    meta, _ := TattooDB.GetMetadata(name)
    return meta
}

func (e * Export) GetArticleTags(name string) []string {
    meta, _ := TattooDB.GetMetadata(name)
	return meta.Tags
}
