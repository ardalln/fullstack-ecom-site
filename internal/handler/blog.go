package handler

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"shop-api/internal/domain"
	"shop-api/internal/richtext"
	"shop-api/internal/service"
)

type BlogHandler struct{ svc *service.BlogService }

func NewBlogHandler(svc *service.BlogService) *BlogHandler { return &BlogHandler{svc: svc} }

func (r BlogPostRequest) toInput() service.BlogInput {
	return service.BlogInput{
		Title: r.Title, Slug: r.Slug, Summary: r.Summary, Content: r.Content,
		CoverImageURL: r.CoverImageURL, SEOTitle: r.SEOTitle,
		SEODescription: r.SEODescription, IsPublished: r.IsPublished,
	}
}

func (h *BlogHandler) ListPublished(c echo.Context) error {
	page, limit := pagination(c)
	search := strings.TrimSpace(c.QueryParam("q"))
	if len(search) > 100 {
		return fmt.Errorf("%w: search query must be at most 100 characters", domain.ErrInvalidInput)
	}
	posts, total, err := h.svc.ListPublished(c.Request().Context(), search, page, limit)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toBlogPostPage(posts, page, limit, total))
}

func (h *BlogHandler) GetPublished(c echo.Context) error {
	post, err := h.svc.GetPublished(c.Request().Context(), c.Param("slug"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toBlogPostResponse(post))
}

func (h *BlogHandler) GetAdmin(c echo.Context) error {
	id, err := idParam(c)
	if err != nil {
		return err
	}
	post, err := h.svc.GetAdmin(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toBlogPostResponse(post))
}

func (h *BlogHandler) ListAdmin(c echo.Context) error {
	page, limit := pagination(c)
	search := strings.TrimSpace(c.QueryParam("q"))
	if len(search) > 100 {
		return fmt.Errorf("%w: search query must be at most 100 characters", domain.ErrInvalidInput)
	}
	posts, total, err := h.svc.ListAdmin(c.Request().Context(), search, page, limit)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toBlogPostPage(posts, page, limit, total))
}

func (h *BlogHandler) Create(c echo.Context) error {
	var req BlogPostRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	post, err := h.svc.Create(c.Request().Context(), req.toInput())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, toBlogPostResponse(post))
}

func (h *BlogHandler) Update(c echo.Context) error {
	id, err := idParam(c)
	if err != nil {
		return err
	}
	var req BlogPostRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	post, err := h.svc.Update(c.Request().Context(), id, req.toInput())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toBlogPostResponse(post))
}

func (h *BlogHandler) Delete(c echo.Context) error {
	id, err := idParam(c)
	if err != nil {
		return err
	}
	if err := h.svc.Delete(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

type blogSEOPage struct {
	Title       string
	Description string
	Canonical   string
	Post        *domain.BlogPost
	Posts       []domain.BlogPost
	HTMLContent template.HTML
}

var blogPageTemplate = template.Must(template.New("blog-page").Parse(`<!doctype html>
<html lang="fa" dir="rtl"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Title}}</title><meta name="description" content="{{.Description}}"><link rel="canonical" href="{{.Canonical}}">
<meta property="og:type" content="{{if .Post}}article{{else}}website{{end}}"><meta property="og:title" content="{{.Title}}"><meta property="og:description" content="{{.Description}}"><meta property="og:url" content="{{.Canonical}}">
{{if .Post}}{{if .Post.CoverImageURL}}<meta property="og:image" content="{{.Post.CoverImageURL}}">{{end}}<meta name="author" content="OUTSIDE">{{end}}
<link rel="stylesheet" href="/assets/index.css">
<style>.seo-blog-header{padding:1rem;background:#18181b;color:#fff}.seo-blog-header>div{max-width:1320px;margin:auto;display:flex;justify-content:space-between;align-items:center}.seo-blog-header a{color:#fff;text-decoration:none;font-size:.8rem}.seo-blog-brand{font-size:1.2rem!important;font-weight:800;letter-spacing:.1em}.seo-blog-brand span{color:#c91e32}.seo-blog-container{max-width:1320px;margin:auto;padding:1px 1rem}.seo-blog-article,.seo-blog-list{max-width:52rem;margin:2rem auto;padding:1.5rem;border:1px solid #deddda;background:#fff}.seo-blog-article h1,.seo-blog-list h1{font-size:2rem}.seo-blog-cover{width:100%;max-height:28rem;object-fit:cover}.seo-blog-content{line-height:2}.seo-blog-content h2,.seo-blog-content h3,.seo-blog-content h4{line-height:1.6;margin:1.7em 0 .5em}.seo-blog-content ul,.seo-blog-content ol{padding-inline-start:2rem}.seo-blog-content a{color:#b51e32;text-decoration:underline}.seo-blog-card{padding:1rem 0;border-top:1px solid #deddda}.seo-blog-card a{font-weight:800;color:#202024}</style></head>
<body><div id="root"><header class="seo-blog-header"><div><a href="/#/" class="seo-blog-brand">OUTSIDE<span>.</span></a><a href="/blog">مجله OUTSIDE</a></div></header>
<main class="seo-blog-container" tabindex="-1">{{if .Post}}<article class="seo-blog-article" itemscope itemtype="https://schema.org/BlogPosting">
<a href="/blog">← وبلاگ</a><h1 itemprop="headline">{{.Post.Title}}</h1>{{if .Post.Summary}}<p itemprop="description">{{.Post.Summary}}</p>{{end}}
{{if .Post.PublishedAt}}<time itemprop="datePublished" datetime="{{.Post.PublishedAt.Format "2006-01-02T15:04:05Z07:00"}}">{{.Post.PublishedAt.Format "2006-01-02"}}</time>{{end}}
{{if .Post.CoverImageURL}}<img class="seo-blog-cover" src="{{.Post.CoverImageURL}}" alt="{{.Post.Title}}" itemprop="image">{{end}}
<div class="seo-blog-content" itemprop="articleBody">{{.HTMLContent}}</div></article>{{else}}<section class="seo-blog-list"><h1>مجله OUTSIDE</h1><p>راهنمای انتخاب کتونی، مراقبت از کفش و تازه‌ترین کالکشن‌ها.</p>
{{range .Posts}}<article class="seo-blog-card"><h2><a href="/blog/{{.Slug}}">{{.Title}}</a></h2><p>{{.Summary}}</p></article>{{else}}<p>مطلبی منتشر نشده است.</p>{{end}}</section>{{end}}</main></div>
<script type="module" src="/assets/app.js"></script></body></html>`))

func (h *BlogHandler) SEOIndex(c echo.Context) error {
	posts, _, err := h.svc.ListPublished(c.Request().Context(), "", 1, 50)
	if err != nil {
		return err
	}
	return renderBlogSEO(c, blogSEOPage{
		Title: "مجله OUTSIDE | راهنمای دنیای کتونی", Description: "راهنمای انتخاب کتونی، مراقبت از کفش و تازه‌ترین کالکشن‌ها در مجله OUTSIDE.",
		Canonical: blogCanonical(c, "/blog"), Posts: posts,
	})
}

func (h *BlogHandler) SEOPage(c echo.Context) error {
	post, err := h.svc.GetPublished(c.Request().Context(), c.Param("slug"))
	if err != nil {
		return err
	}
	description := post.SEODescription
	if description == "" {
		description = post.Summary
	}
	title := post.SEOTitle
	if title == "" {
		title = post.Title
	}
	return renderBlogSEO(c, blogSEOPage{Title: title + " | مجله OUTSIDE", Description: description,
		Canonical: blogCanonical(c, "/blog/"+post.Slug), Post: post, HTMLContent: template.HTML(richtext.Sanitize(post.Content))})
}

func renderBlogSEO(c echo.Context, page blogSEOPage) error {
	var output bytes.Buffer
	if err := blogPageTemplate.Execute(&output, page); err != nil {
		return fmt.Errorf("render blog SEO page: %w", err)
	}
	return c.Blob(http.StatusOK, "text/html; charset=utf-8", output.Bytes())
}

func blogCanonical(c echo.Context, path string) string {
	scheme := c.Scheme()
	if forwarded := strings.TrimSpace(strings.Split(c.Request().Header.Get("X-Forwarded-Proto"), ",")[0]); forwarded == "http" || forwarded == "https" {
		scheme = forwarded
	}
	if scheme == "" {
		scheme = "http"
	}
	return fmt.Sprintf("%s://%s%s", scheme, c.Request().Host, path)
}
