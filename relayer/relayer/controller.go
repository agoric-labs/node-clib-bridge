package relayer

var SendToNode func(needReply bool, str string) (string, error)
