package relayer

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type DeliverMsg struct {
	Msg  string `json:"msg"`
	Type string `json:"type"`
}

type DeliverMsgsAction struct {
	SrcMsgs []DeliverMsg `json:"src_msgs"`
	Src     PathEnd      `json:"src"`
	DstMsgs []DeliverMsg `json:"dst_msgs"`
	Dst     PathEnd      `json:"dst"`
	Type    string       `json:"type"`
}

// RelayMsgs contains the msgs that need to be sent to both a src and dst chain
// after a given relay round
type RelayMsgs struct {
	Src []sdk.Msg `json:"src"`
	Dst []sdk.Msg `json:"dst"`

	last    bool
	success bool
}

// Ready returns true if there are messages to relay
func (r *RelayMsgs) Ready() bool {
	if r == nil {
		return false
	}

	if len(r.Src) == 0 && len(r.Dst) == 0 {
		return false
	}
	return true
}

// Success returns the success var
func (r *RelayMsgs) Success() bool {
	return r.success
}

// Send sends the messages with appropriate output
func (r *RelayMsgs) Send(src, dst *Chain) bool {
	if SendToController != nil {
		action := &DeliverMsgsAction{
			SrcMsgs: MarshalMsgs(r.Src),
			Src:     MarshalChain(src),
			DstMsgs: MarshalMsgs(r.Dst),
			Dst:     MarshalChain(dst),
			Type:    "RELAYER_SEND",
		}

		// Get the messages that are actually sent.
		cont, err := ControllerUpcall(&action)
		if !cont {
			if err != nil {
				fmt.Println("Error calling controller", err)
				r.success = false
			} else {
				r.success = true
			}
			return r.success
		}
	}

	var failed = false
	// TODO: maybe figure out a better way to indicate error here?

	// TODO: Parallelize? Maybe?
	if len(r.Src) > 0 {
		// Submit the transactions to src chain
		res, err := src.SendMsgs(r.Src)
		if err != nil || res.Code != 0 {
			src.LogFailedTx(res, err, r.Src)
			failed = true
		} else {
			// NOTE: Add more data to this such as identifiers
			src.LogSuccessTx(res, r.Src)
		}
	}

	if len(r.Dst) > 0 {
		// Submit the transactions to dst chain
		res, err := dst.SendMsgs(r.Dst)
		if err != nil || res.Code != 0 {
			dst.LogFailedTx(res, err, r.Dst)
			failed = true
		} else {
			// NOTE: Add more data to this such as identifiers
			dst.LogSuccessTx(res, r.Dst)

		}
	}

	if failed {
		r.success = false
		return r.success
	}
	r.success = true
	return r.success
}

func getMsgAction(msgs []sdk.Msg) string {
	var out string
	for i, msg := range msgs {
		out += fmt.Sprintf("%d:%s,", i, msg.Type())
	}
	return strings.TrimSuffix(out, ",")
}
