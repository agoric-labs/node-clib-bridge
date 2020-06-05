package relayer

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type DeliverMsgsAction struct {
	SrcMsgs   []string `json:"src_msgs"`
	Src       PathEnd  `json:"src"`
	DstMsgs   []string `json:"dst_msgs"`
	Dst       PathEnd  `json:"dst"`
	Last      bool     `json:"last"`
	Succeeded bool     `json:"succeeded"`
	Type      string   `json:"type"`
}

// RelayMsgs contains the msgs that need to be sent to both a src and dst chain
// after a given relay round
type RelayMsgs struct {
	Src []sdk.Msg `json:"src"`
	Dst []sdk.Msg `json:"dst"`

	Last      bool `json:"last"`
	Succeeded bool `json:"success"`
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
	return r.Succeeded
}

// Send sends the messages with appropriate output
func (r *RelayMsgs) Send(src, dst *Chain) {
	r.SendWithController(src, dst, true)
}

func EncodeMsgs(c *Chain, msgs []sdk.Msg) []string {
	outMsgs := make([]string, 0, len(msgs))
	for _, msg := range msgs {
		bz, err := c.Cdc.MarshalJSON(msg)
		if err != nil {
			fmt.Println("Cannot marshal message", msg, err)
		} else {
			outMsgs = append(outMsgs, string(bz))
		}
	}
	return outMsgs
}

func DecodeMsgs(c *Chain, msgs []string) []sdk.Msg {
	outMsgs := make([]sdk.Msg, 0, len(msgs))
	for _, msg := range msgs {
		var sm sdk.Msg
		err := c.Cdc.UnmarshalJSON([]byte(msg), &sm)
		if err != nil {
			fmt.Println("Cannot unmarshal message", err)
		} else {
			outMsgs = append(outMsgs, sm)
		}
	}
	return outMsgs
}

func (r *RelayMsgs) SendWithController(src, dst *Chain, useController bool) {
	if useController && SendToController != nil {
		action := &DeliverMsgsAction{
			Src:       MarshalChain(src),
			Dst:       MarshalChain(dst),
			Last:      r.Last,
			Succeeded: r.Succeeded,
			Type:      "RELAYER_SEND",
		}

		action.SrcMsgs = EncodeMsgs(src, r.Src)
		action.DstMsgs = EncodeMsgs(dst, r.Dst)

		// Get the messages that are actually sent.
		cont, err := ControllerUpcall(&action)
		if !cont {
			if err != nil {
				fmt.Println("Error calling controller", err)
				r.Succeeded = false
			} else {
				r.Succeeded = true
			}
			return
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

	r.Succeeded = !failed
}

func getMsgAction(msgs []sdk.Msg) string {
	var out string
	for i, msg := range msgs {
		out += fmt.Sprintf("%d:%s,", i, msg.Type())
	}
	return strings.TrimSuffix(out, ",")
}
