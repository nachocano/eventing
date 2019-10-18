package ingress

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"time"

	cloudevents "github.com/cloudevents/sdk-go"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
	"knative.dev/eventing/pkg/broker"
	"knative.dev/eventing/pkg/utils"
)

var (
	shutdownTimeout = 1 * time.Minute

	defaultTTL = 255
)

type Handler struct {
	Logger     *zap.Logger
	CeClient   cloudevents.Client
	ChannelURI *url.URL
	BrokerName string
	Namespace  string
	Reporter   StatsReporter
	QueueSize  int

	jobs chan Job
}

type Job struct {
	sendingCTX   *context.Context
	event        *cloudevents.Event
	reporterArgs *ReportArgs
}

func (h *Handler) Sender(jobs <-chan Job) {
	for job := range jobs {
		start := time.Now()
		rctx, _, _ := h.CeClient.Send(*job.sendingCTX, *job.event)
		rtctx := cloudevents.HTTPTransportContextFrom(rctx)
		// Record the dispatch time.
		h.Reporter.ReportEventDispatchTime(job.reporterArgs, rtctx.StatusCode, time.Since(start))
		// Record the event count.
		h.Reporter.ReportEventCount(job.reporterArgs, rtctx.StatusCode)
	}

}

func (h *Handler) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	workers := runtime.NumCPU()
	h.Logger.Info("", zap.Any("workers", workers))
	runtime.GOMAXPROCS(workers)
	h.jobs = make(chan Job, h.QueueSize)
	for w := 1; w <= workers; w++ {
		go h.Sender(h.jobs)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- h.CeClient.StartReceiver(ctx, h.serveHTTP)
	}()

	// Stop either if the receiver stops (sending to errCh) or if stopCh is closed.
	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		break
	}

	// stopCh has been closed, we need to gracefully shutdown h.ceClient. cancel() will start its
	// shutdown, if it hasn't finished in a reasonable amount of time, just return an error.
	cancel()
	select {
	case err := <-errCh:
		return err
	case <-time.After(shutdownTimeout):
		return errors.New("timeout shutting down ceClient")
	}
}

func (h *Handler) serveHTTP(ctx context.Context, event cloudevents.Event, resp *cloudevents.EventResponse) error {
	// Setting the extension as a string as the CloudEvents sdk does not support non-string extensions.
	event.SetExtension(broker.EventArrivalTime, time.Now().Format(time.RFC3339))
	tctx := cloudevents.HTTPTransportContextFrom(ctx)
	if tctx.Method != http.MethodPost {
		resp.Status = http.StatusMethodNotAllowed
		return nil
	}

	// tctx.URI is actually the path...
	if tctx.URI != "/" {
		resp.Status = http.StatusNotFound
		return nil
	}

	reporterArgs := &ReportArgs{
		ns:          h.Namespace,
		broker:      h.BrokerName,
		eventType:   event.Type(),
		eventSource: event.Source(),
	}

	send := h.decrementTTL(&event)
	if !send {
		// Record the event count.
		h.Reporter.ReportEventCount(reporterArgs, http.StatusBadRequest)
		return nil
	}

	sendingCTX := utils.ContextFrom(tctx, h.ChannelURI)
	// Due to an issue in utils.ContextFrom, we don't retain the original trace context from ctx, so
	// bring it in manually.
	sendingCTX = trace.NewContext(sendingCTX, trace.FromContext(ctx))

	job := Job{
		sendingCTX:   &sendingCTX,
		event:        &event,
		reporterArgs: reporterArgs,
	}
	h.jobs <- job

	// Always return accepted. No guarantees if we cannot send it to the channel.
	resp.Status = http.StatusAccepted
	return nil
}

func (h *Handler) decrementTTL(event *cloudevents.Event) bool {
	ttl := h.getTTLToSet(event)
	if ttl <= 0 {
		// TODO send to some form of dead letter queue rather than dropping.
		h.Logger.Error("Dropping message due to TTL", zap.Any("event", event))
		return false
	}

	var err error
	event.Context, err = broker.SetTTL(event.Context, ttl)
	if err != nil {
		h.Logger.Error("failed to set TTL", zap.Error(err))
	}
	return true
}

func (h *Handler) getTTLToSet(event *cloudevents.Event) int {
	ttlInterface, _ := broker.GetTTL(event.Context)
	if ttlInterface == nil {
		h.Logger.Debug("No TTL found, defaulting")
		return defaultTTL
	}
	// This should be a JSON number, which json.Unmarshalls as a float64.
	ttl, ok := ttlInterface.(float64)
	if !ok {
		h.Logger.Info("TTL attribute wasn't a float64, defaulting", zap.Any("ttlInterface", ttlInterface), zap.Any("typeOf(ttlInterface)", reflect.TypeOf(ttlInterface)))
	}
	return int(ttl) - 1
}
