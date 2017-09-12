package main

import "fmt"

type Engine struct {
	name       string
	host       string
	getFile    func(string) string
	getTmb     func(string) string
	getStatic  func(string, string) string
	genThread  func(string, string) (Thread, error)
	getThreads func(string, chan<- struct{ n, l float64 }) error
}

func getEngine(host string) Engine {
	switch host {
	case "8ch", "8chan", "8ch.net":
		return b_8chan
	default:
		return newVichan(host)
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
