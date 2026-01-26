package stream

import (
	"context"

	relay "github.com/yarlson/relay"
	relaystore "github.com/yarlson/relay/store"
	filestore "github.com/yarlson/turbine/internal/relay/store"
)

func AppendEvent(ctx context.Context, store *filestore.FileStore, workflowID string, evt relay.Event) {
	if store == nil || workflowID == "" {
		return
	}
	_ = store.AppendEvent(ctx, workflowID, toStoreEvent(evt))
}

func toStoreEvent(evt relay.Event) *relaystore.Event {
	return &relaystore.Event{
		Kind:      relaystore.EventKind(evt.Kind),
		Timestamp: evt.Timestamp,
		StepID:    evt.StepID,
		SessionID: evt.SessionID,
		Text:      evt.Text,
		Usage:     toStoreUsage(evt.Usage),
		Error:     evt.Error,
	}
}

func toStoreUsage(usage *relay.Usage) *relaystore.Usage {
	if usage == nil {
		return nil
	}
	return &relaystore.Usage{
		InputTokens:  usage.InputTokens,
		OutputTokens: usage.OutputTokens,
		CostUSD:      usage.CostUSD,
	}
}
