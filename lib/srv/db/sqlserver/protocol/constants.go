package protocol

const (
	preloginVERSION         = 0
	preloginENCRYPTION      = 1
	preloginINSTOPT         = 2
	preloginTHREADID        = 3
	preloginMARS            = 4
	preloginTRACEID         = 5
	preloginFEDAUTHREQUIRED = 6
	preloginNONCEOPT        = 7
	preloginTERMINATOR      = 0xff
)

const (
	encryptOff    = 0 // Encryption is available but off.
	encryptOn     = 1 // Encryption is available and on.
	encryptNotSup = 2 // Encryption is not available.
	encryptReq    = 3 // Encryption is required.
)
