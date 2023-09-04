package server

const (
	opNone            = iota // 0x00
	opIdentify               // 0x01
	opSession                // 0x02
	opQuit                   // 0x03
	opCommand                // 0x04
	opCommandResponse        // 0x05
	opPing                   // 0x06
	opPong                   // 0x07
	opEcho                   // 0x08
)
