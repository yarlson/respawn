package stream

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	relay "github.com/yarlson/relay"
)

func TestToStoreEvent(t *testing.T) {
	timestamp := time.Now()
	evt := relay.Event{
		Kind:      relay.EventKindText,
		Timestamp: timestamp,
		StepID:    "step-1",
		SessionID: "sess-1",
		Text:      "hello",
		Usage: &relay.Usage{
			InputTokens:  10,
			OutputTokens: 20,
			CostUSD:      0.5,
		},
	}

	storeEvt := toStoreEvent(evt)
	require.Equal(t, relay.EventKindText, relay.EventKind(storeEvt.Kind))
	require.Equal(t, timestamp, storeEvt.Timestamp)
	require.Equal(t, "step-1", storeEvt.StepID)
	require.Equal(t, "sess-1", storeEvt.SessionID)
	require.Equal(t, "hello", storeEvt.Text)
	require.NotNil(t, storeEvt.Usage)
	require.Equal(t, 10, storeEvt.Usage.InputTokens)
	require.Equal(t, 20, storeEvt.Usage.OutputTokens)
	require.Equal(t, 0.5, storeEvt.Usage.CostUSD)
}

func TestToStoreEvent_NilUsage(t *testing.T) {
	storeEvt := toStoreEvent(relay.Event{Kind: relay.EventKindWarning})
	require.Nil(t, storeEvt.Usage)
}
