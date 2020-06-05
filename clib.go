package main

// /* These comments before the import "C" are included in the C output. */
// #include <stdlib.h>
// typedef const char* Body;
// typedef int (*sendFunc)(int, int, Body);
// inline int invokeSendFunc(sendFunc send, int port, int reply, Body str) {
//    return send(port, reply, str);
// }
import "C"
import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/iqlusioninc/relayer/relayer"
)

type goReturn = struct {
	str string
	err error
}

var clibPort = 0
var replies = map[int]chan goReturn{}
var lastReply = 0

//export RunClib
func RunClib(nodePort C.int, toNode C.sendFunc, clibArgs []*C.char) C.int {
	sendToNode := func(needReply bool, str string) (string, error) {
		var rPort int
		if needReply {
			lastReply++
			rPort = lastReply
			replies[rPort] = make(chan goReturn)
		}
		// Send the message.
		C.invokeSendFunc(toNode, nodePort, C.int(rPort), C.CString(str))
		if !needReply {
			// Return immediately
			return "<no-reply-requested>", nil
		}

		// Block the sending goroutine while we wait for the reply
		ret := <-replies[rPort]
		delete(replies, rPort)
		return ret.str, ret.err
	}

	args := make([]string, len(clibArgs))
	for i, s := range clibArgs {
		args[i] = C.GoString(s)
	}
	fmt.Println("Starting Clib with args", args)
	go func() {
		for i := 0; i < 3; i++ {
			fmt.Println("Call", i)
			s, err := sendToNode(true, fmt.Sprintf("%d", i))
			if err != nil {
				fmt.Println("Error", err)
			} else {
				fmt.Println("Return", s)
			}
		}
		os.Exit(0)
	}()

	clibPort++
	return C.int(clibPort)
}

//export ReplyToClib
func ReplyToClib(replyPort C.int, isError C.int, str C.Body) C.int {
	goStr := C.GoString(str)
	returnCh := replies[int(replyPort)]
	if returnCh == nil {
		return C.int(0)
	}
	ret := goReturn{}
	if int(isError) == 0 {
		ret.str = goStr
	} else {
		ret.err = errors.New(goStr)
	}
	returnCh <- ret
	return C.int(0)
}

//export SendToClib
func SendToClib(port C.int, str C.Body) C.Body {
	goStr := C.GoString(str)
	var action relayer.DeliverMsgsAction
	if err := json.Unmarshal([]byte(goStr), &action); err == nil {
		switch action.Type {
		case "RELAYER_SEND":
			rm := relayer.RelayMsgs{
				Src: unmarshalMsgs(action.SrcMsgs),
				Dst: unmarshalMsgs(action.DstMsgs),
			}
			src := unmarshalChain(action.Src)
			dst := unmarshalChain(action.Dst)
			if src == nil || dst == nil {
				return C.CString("false")
			}
			rm.Send(src, dst)
			if !rm.success {
				return C.CString("0")
			}
			return C.CString(fmt.Sprintf("%d", len(rm.Src) + len(rm.Dst))
		}
	}
	return C.CString("false")
}

// Do nothing in main.
func main() {}
