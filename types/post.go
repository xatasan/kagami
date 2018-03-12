package types

import (
	"html/template"
	"time"
)

// different kinds of flags for posts
type Flag struct {
	Icon, Name string
}

// one message, posted in a thread, either as OP
// or as a response
type Post struct {
	PostNumber int
	ReplyTo    *Post
	Sticky     bool
	Closed     bool
	OP         bool
	Time       time.Time
	Flag       *Flag
	Name       string
	Tripcode   string
	Id         string
	Capcode    string
	Subject    string
	Comment    template.HTML
	QuotedBy   []*Post
	Files      []*File
}

// a thread is a list of posts, where the first
// is the head
type Thread []*Post
