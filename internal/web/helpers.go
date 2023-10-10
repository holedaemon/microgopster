package web

import (
	"context"
	"errors"
	"image"
	"net/http"

	_ "image/gif"
	_ "image/jpeg"
)

var errNonOK = errors.New("web: non-OK status")

func (s *Server) downloadImage(ctx context.Context, url string) (image.Image, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent)

	res, err := s.cli.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, errNonOK
	}

	m, _, err := image.Decode(res.Body)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func noCover() (image.Image, error) {
	file, err := staticDir.Open("no_cover.jpg")
	if err != nil {
		return nil, err
	}

	defer file.Close()

	m, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	return m, nil
}
