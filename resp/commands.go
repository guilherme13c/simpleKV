package resp

type RESPCommand string

const (
	CMD_SET     = "SET"
	CMD_GET     = "GET"
	CMD_DEL     = "DEL"
	CMD_COMMAND = "COMMAND"
	CMD_INFO    = "INFO"
	CMD_SCAN    = "SCAN"
)
