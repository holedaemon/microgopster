package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/holedaemon/microgopster/internal/web"
	"github.com/zikaeroh/ctxlog"
	"go.uber.org/zap"
)

func die(msg string, args ...any) {
	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}

	fmt.Fprintf(os.Stderr, msg, args...)
}

func main() {
	addr := os.Getenv("TESTSERVER_ADDR")
	apiKey := os.Getenv("TESTSERVER_API_KEY")
	user := os.Getenv("TESTSERVER_LAST_USER")
	period := os.Getenv("TESTSERVER_LAST_PERIOD")

	if addr == "" || apiKey == "" || user == "" {
		die("missing required argument")
		return
	}

	if period == "" {
		period = "overall"
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger, err := zap.NewDevelopment()
	if err != nil {
		die("error creating logger: %s", err.Error())
		return
	}

	ctx = ctxlog.WithLogger(ctx, logger)

	srv, err := web.New(
		web.WithAddr(addr),
		web.WithAPIKey(apiKey),
	)
	if err != nil {
		die("error creating server: %s", err.Error())
		return
	}

	go func() {
		if err := srv.Run(ctx); err != nil {
			ctxlog.Error(ctx, "error shutting down server")
		}
	}()

	body := &web.GenerateBody{
		User:            user,
		Period:          period,
		ShowTitles:      true,
		Gap:             20,
		BackgroundColor: "#f5a442",
	}

	var input bytes.Buffer
	if err := json.NewEncoder(&input).Encode(&body); err != nil {
		die("error encoding json: %s", err.Error())
		return
	}

	res, err := http.Post("http://localhost"+addr, "application/json", &input)
	if err != nil {
		die("error POSTing: %s", err.Error())
		return
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		var e web.Error
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			die("error code %d returned, error decoding error: %s", res.StatusCode, err.Error())
			return
		}

		die("error code %d returned, msg: %s", res.StatusCode, e.Message)
		return
	}

	var gr *web.GeneratedResponse
	if err := json.NewDecoder(res.Body).Decode(&gr); err != nil {
		die("error decoding image: %s", err.Error())
		return
	}

	output, err := base64.StdEncoding.DecodeString(gr.Image)
	if err != nil {
		die("error decoding image from base64: %s", err.Error())
		return
	}

	unix := time.Now().Unix()
	seed := strconv.FormatInt(unix, 10)
	file, err := os.Create("out" + seed + ".jpg")
	if err != nil {
		die("error creating output file: %s", err.Error())
		return
	}

	defer file.Close()

	if _, err := file.Write(output); err != nil {
		die("error writing file: %s", err.Error())
		return
	}
}
