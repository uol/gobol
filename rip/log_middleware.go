package rip

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/pborman/uuid"
	"github.com/uol/gobol/snitch"
)

type key int

const (
	statsTagskey key = 0
)

type LogResponseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (w *LogResponseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	size, err := w.ResponseWriter.Write(b)
	w.size += size
	return size, err
}

func (w *LogResponseWriter) WriteHeader(s int) {
	w.ResponseWriter.WriteHeader(s)
	w.status = s
}

func (w *LogResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func NewLogMiddleware(service, system string, logger *logrus.Logger, sts *snitch.Stats, next http.Handler) *LogHandler {
	return &LogHandler{
		service: service,
		system:  system,
		next:    next,
		logger:  logger,
		stats:   sts,
	}
}

type LogHandler struct {
	service string
	system  string
	next    http.Handler
	logger  *logrus.Logger
	stats   *snitch.Stats
}

func (h *LogHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	rid := uuid.NewRandom().String()

	header := w.Header()

	header.Add(fmt.Sprintf("X-REQUEST-%s-ID", h.service), rid)

	if header.Get(fmt.Sprintf("X-REQUEST-%s-ID", h.system)) == "" {
		header.Add(fmt.Sprintf("X-REQUEST-%s-ID", h.system), rid)
	}

	logw := &LogResponseWriter{
		ResponseWriter: w,
	}

	ctx := context.Background()

	userTags := map[string]string{}

	h.next.ServeHTTP(logw, r.WithContext(context.WithValue(ctx, statsTagskey, userTags)))

	status := logw.status

	addr, _, _ := net.SplitHostPort(r.RemoteAddr)

	d := time.Since(start)

	fields := logrus.Fields{
		"id":         rid,
		"status":     status,
		"method":     r.Method,
		"path":       r.RequestURI,
		"remote":     addr,
		"service":    h.service,
		"system":     h.system,
		"duration":   d,
		"size":       logw.size,
		"user-agent": r.UserAgent(),
	}

	if f := r.Header.Get("X-FORWARDED-FOR"); f != "" {
		fields["forward"] = f
	}

	if status >= http.StatusBadRequest {
		h.logger.WithFields(fields).Error("completed handling request with errors")
	} else {
		h.logger.WithFields(fields).Info("completed handling request")
	}

	tags := map[string]string{
		"protocol": r.Proto,
		"method":   r.Method,
		"status":   strconv.Itoa(status),
	}

	if status != 404 && status != 405 {
		tags["path"] = r.URL.Path
	}

	for k, v := range userTags {
		tags[k] = v
	}

	h.increment("request.count", tags)
	h.valueAdd("request.duration", tags, float64(d.Nanoseconds())/float64(time.Millisecond))
}

func AddStatsMap(r *http.Request, tags map[string]string) {
	t, ok := r.Context().Value(statsTagskey).(map[string]string)
	if ok {
		for k, v := range tags {
			t[k] = v
		}
	}
}

func (h *LogHandler) increment(metric string, tags map[string]string) {
	err := h.stats.Increment(metric, tags, "@every 1m", false, true)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"package": "rip",
			"func":    "statsIncrement",
			"metric":  metric,
		}).Error(err)
	}
}

func (h *LogHandler) valueAdd(metric string, tags map[string]string, v float64) {
	err := h.stats.ValueAdd(metric, tags, "avg", "@every 1m", false, false, v)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"package": "rip",
			"func":    "statsValueAdd",
			"metric":  metric,
		}).Error(err)
	}
}
