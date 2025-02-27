package resp

type RESPType rune

// RESP type
const (
	SIMPLE_STRING   = '+'
	SIMPLE_ERROR    = '-'
	INTEGER         = ':'
	BULK_STRING     = '$'
	ARRAY           = '*'
	NULL            = '_'
	BOOLEAN         = '#'
	DOUBLE          = ','
	BIG_NUMBER      = '('
	VERBATIM_STRING = '='
	MAP             = '%'
	ATTRIBUTE       = '`'
	SET             = '~'
	PUSH            = '>'
)
