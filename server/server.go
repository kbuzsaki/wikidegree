package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/kbuzsaki/wikidegree/server/logic"
	"golang.org/x/net/context"
)

type Server interface {
	HandlePathLookup(writer http.ResponseWriter, request *http.Request)
	HandlePageLookup(writer http.ResponseWriter, request *http.Request)
}

type serverImpl struct {
	logic logic.Logic
}

func New() (Server, error) {
	l, err := logic.New()
	if err != nil {
		return nil, err
	}

	return &serverImpl{logic: l}, nil
}

func (s *serverImpl) HandlePathLookup(writer http.ResponseWriter, request *http.Request) {
	values := request.URL.Query()
	start := values.Get("start")
	end := values.Get("end")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	startTime := time.Now()
	path, err := s.logic.LookupPath(ctx, start, end)
	duration := time.Since(startTime)

	if err != nil {
		s.renderError(writer, err)
	} else if ctx.Err() != nil {
		s.renderError(writer, errors.New("Timed out after 10 seconds."))
	} else {
		s.renderJSON(writer, map[string]interface{}{
			"time": duration.String(),
			"path": path,
		})
	}
}

func (s *serverImpl) HandlePageLookup(writer http.ResponseWriter, request *http.Request) {
	values := request.URL.Query()
	title := values.Get("title")

	page, err := s.logic.LookupPage(context.Background(), title)
	if err != nil {
		s.renderError(writer, err)
	} else {
		s.renderJSON(writer, page)
	}
}

func (s *serverImpl) renderJSON(writer http.ResponseWriter, resp interface{}) {
	respBytes, _ := json.Marshal(&resp)
	io.WriteString(writer, string(respBytes))
}

func (s *serverImpl) renderError(writer http.ResponseWriter, err error) {
	s.renderJSON(writer, map[string]string{"error": err.Error()})
}
