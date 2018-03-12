package types

import "image"

// files (images, videos, pdfs, ...) can be attached to
// a post
type File struct {
	Filename      string
	OrigFilename  string
	FileMD5       string
	ThumbnailName string
	Mime          string
	FileDeleted   bool
	Spoiler       bool
	FileSize      int
	Image         image.Point
	Thumbnail     image.Point
}
