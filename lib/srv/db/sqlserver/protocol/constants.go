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

const (
	verTDS70     = 0x70000000
	verTDS71     = 0x71000000
	verTDS71rev1 = 0x71000001
	verTDS72     = 0x72090002
	verTDS73A    = 0x730A0003
	verTDS73     = verTDS73A
	verTDS73B    = 0x730B0003
	verTDS74     = 0x74000004
)
