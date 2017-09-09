package main

import "fmt"

type Engine interface {
	getName() string
	getHost() string
	getFile(string) string
	getTmb(string) string
	getStatic(string, string) string
	genThread(board, no string) (Thread, error)
	getThreads(string, chan<- struct{ n, l float64 }) error
}

func getEngine(host string) Engine {
	switch host {
	case "8ch", "8chan", "8ch.net":
		return b_8chan{e_vichan: e_vichan{host: "8ch.net"}}
	default:
		return e_vichan{host: host}
	}
}

func getFile(e Engine, file File) error {
	local := fmt.Sprintf("%s/%s", i_dir, file.Filename)
	remote := e.getFile(file.Filename)
	return dl(local, remote)
}

func getThumbnail(e Engine, file File) error {
	local := fmt.Sprintf("%s/%s", T_dir, file.Thumbnail)
	remote := e.getTmb(file.Thumbnail)
	if remote == "" {
		return nil
	}
	return dl(local, remote)
}
