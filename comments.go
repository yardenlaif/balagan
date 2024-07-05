package main

import (
	"go/ast"
	"strings"
)

func shouldDeleteComment(comment *ast.Comment) bool {
	return !strings.HasPrefix(comment.Text, "//go:") && !strings.HasPrefix(comment.Text, "// +build")
}

func removeComments(astPkgs []*ast.Package) {
	for _, pkg := range astPkgs {
		for _, file := range pkg.Files {
			var groups []*ast.CommentGroup
			for _, group := range file.Comments {
				newGroup := &ast.CommentGroup{}
				for _, comment := range group.List {
					if !shouldDeleteComment(comment) {
						newGroup.List = append(newGroup.List, comment)
					}
				}
				groups = append(groups, newGroup)
			}
			file.Comments = groups
		}
	}
}
