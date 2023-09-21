package web

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"image"
	"image/jpeg"
	"net/http"

	"github.com/holedaemon/gopster"
	"github.com/holedaemon/lastfm"
	"github.com/zikaeroh/ctxlog"
	"go.uber.org/zap"
)

type GenerateBody struct {
	User            string  `json:"user"`
	Period          string  `json:"period"`
	Title           string  `json:"title"`
	BackgroundColor string  `json:"background_color"`
	TextColor       string  `json:"text_color"`
	Gap             float64 `json:"gap"`
	ShowNumbers     bool    `json:"show_numbers"`
	ShowTitles      bool    `json:"show_titles"`
}

type GeneratedResponse struct {
	Image string `json:"image,omitempty"`
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var b *GenerateBody

	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		respondError(w, r, http.StatusBadRequest, "unable to decode JSON body, make sure it's correct")
		return
	}

	if b.User == "" {
		respondError(w, r, http.StatusBadRequest, "lastfm user cannot be blank")
		return
	}

	switch b.Period {
	case "overall", "7day", "1month", "3month", "6month", "12month":
	default:
		b.Period = "overall"
	}

	switch b.Gap {
	case 0:
		b.Gap = 20
	default:
	}

	albums, err := s.LastFM.UserTopAlbums(ctx, &lastfm.UserQuery{
		User:   b.User,
		Limit:  9,
		Page:   1,
		Period: b.Period,
	})
	if err != nil {
		ctxlog.Error(ctx, "error fetching last.fm albums", zap.Error(err), zap.String("user", b.User))
		respondError(w, r, http.StatusInternalServerError, "unable to retrieve last.fm user data")
		return
	}

	if len(albums.Albums) == 0 {
		respond(w, r, http.StatusOK, &GeneratedResponse{Image: ""})
		return
	}

	opts := []gopster.Option{
		gopster.Title(b.Title),
		gopster.BackgroundColor(b.BackgroundColor),
		gopster.TextColor(b.TextColor),
		gopster.Gap(b.Gap),
	}

	if b.ShowTitles {
		opts = append(opts, gopster.ShowTitles())

		if b.ShowNumbers {
			opts = append(opts, gopster.ShowNumbers())
		}
	}

	chart, err := gopster.NewChart(opts...)
	if err != nil {
		if errors.Is(err, gopster.ErrorChart) {
			respondErrorf(w, r, http.StatusBadRequest, "unable to create chart: %s", err.Error())
			return
		}

		ctxlog.Error(ctx, "error creating chart", zap.Error(err))
		respondError(w, r, http.StatusBadRequest, "unable to create chart; try again")
		return
	}

	for _, a := range albums.Albums {
		var im image.Image

		if len(a.Image) != 0 {
			var url string

			for _, i := range a.Image {
				if i.Size == "extralarge" {
					url = i.Text
				}
			}

			if url != "" {
				di, err := s.downloadImage(ctx, url)
				if err != nil {
					ctxlog.Error(ctx, "error downloading album cover", zap.Error(err), zap.String("album", a.Name))
					respondError(w, r, http.StatusInternalServerError, "error downloading cover image")
					return
				}

				im = di
			}
		}

		if im == nil {
			nc, err := noCover()
			if err != nil {
				ctxlog.Error(ctx, "error decoding placeholder cover", zap.Error(err))
				respondError(w, r, http.StatusInternalServerError, "error creating chart; unable to decode placeholder cover")
				return
			}
			im = nc
		}

		name := a.Name
		if name == "" {
			name = "Unknown Album"
		}

		artist := a.Artist.Name
		if artist == "" {
			artist = "Unknown Artist"
		}

		err = chart.AddItem(name, artist, im)
		if err != nil {
			respondError(w, r, http.StatusInternalServerError, "error creating chart; unable to add item")
			return
		}
	}

	generated := chart.Generate()
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, generated, nil)
	if err != nil {
		ctxlog.Error(ctx, "error encoding chart to jpeg", zap.Error(err))
		respondError(w, r, http.StatusInternalServerError, "error encoding chart to jpeg")
		return
	}

	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	respond(w, r, http.StatusOK, &GeneratedResponse{
		Image: encoded,
	})
}
