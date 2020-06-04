package relayer

var SendToController func(needReply bool, str string) (string, error)
