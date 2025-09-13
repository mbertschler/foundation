package foundation

import (
	"embed"
	_ "embed"
	"io/fs"
	"net/http"

	"github.com/pkg/errors"
)

var (
	//go:embed browser/dist/*
	embeddedDistAssets embed.FS
	devDistAssets      = http.Dir("./browser/dist")

	//go:embed browser/img/*
	embeddedImgAssets embed.FS
	devImgAssets      = http.Dir("./browser/img")
)

type Assets struct {
	Dist http.FileSystem
	Img  http.FileSystem
}

func (c *Config) Assets() (*Assets, error) {
	if c.DevFileServer {
		return &Assets{
			Dist: devDistAssets,
			Img:  devImgAssets,
		}, nil
	}

	dist, err := fs.Sub(embeddedDistAssets, "browser/dist")
	if err != nil {
		return nil, errors.Wrap(err, "sub embeddedDistAssets")
	}

	img, err := fs.Sub(embeddedImgAssets, "browser/img")
	if err != nil {
		return nil, errors.Wrap(err, "sub embeddedImgAssets")
	}

	return &Assets{
		Dist: http.FS(dist),
		Img:  http.FS(img),
	}, nil
}
