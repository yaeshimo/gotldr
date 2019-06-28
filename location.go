package main

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
)

func FindPage(candidateDirs []string, name string) (string, error) {
	base := filepath.Base(name) + ".md"
	var path string
	for _, dir := range candidateDirs {
		path = filepath.Join(dir, base)
		fi, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return "", err
		}
		if fi.Mode().IsRegular() {
			return path, nil
		}
	}
	return "", errors.New("not found pages")
}

func CacheHome() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "gotldr"), nil
}

// expected location: CacheHome/users/
func UserPageDir() (string, error) {
	h, err := CacheHome()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(h, "users")
	fi, err := os.Stat(dir)
	if err == nil {
		if fi.IsDir() {
			return dir, nil
		}
		return "", errors.New("is not directory: " + dir)
	}
	if os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0700)
		if err != nil {
			return "", err
		}
		return dir, nil
	}
	return "", err
}

// TODO: remove this? for treat pages.pt-BR
// expected ISO 639-1 codes
// lazy validate
var IsValidLang func(lang string) bool = regexp.MustCompile("^[a-z][a-z]$").MatchString

// lang is ISO 639-1 codes
// always priority the users cache
func CandidateCacheDirs(platform, lang string) ([]string, error) {
	ud, err := UserPageDir()
	if err != nil {
		return nil, err
	}
	dirs := []string{ud}

	// validate
	if !IsValidLang(lang) {
		return nil, errors.New("invalid language code: " + lang)
	}
	platform = filepath.Base(platform)

	// upstream's cache
	upd, err := UpstreamCacheDir()
	if err != nil {
		return nil, err
	}

	// with lang
	if lang != "" {
		dirs = append(dirs, filepath.Join(upd, "pages."+lang, "common"))
		if platform != "" {
			dirs = append(dirs, filepath.Join(upd, "pages."+lang, platform))
		}
	}

	// default english pages
	dirs = append(dirs, filepath.Join(upd, "pages", "common"))
	if platform != "" {
		dirs = append(dirs, filepath.Join(upd, "pages", platform))
	}

	return dirs, nil
}
