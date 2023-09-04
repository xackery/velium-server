# velium-server
Velium


### Packet Message Overview

Every packet frame uses the following structure:

`| 1 byte opcode | variable length data | 1 byte \n |`

- [Opcode list](./server/op.go)
- Variable length data can be either a string or binary data, it depends on the opcode to declare structure.
- Newline is a single byte 0x0A, this is a reserved byte that should never be sent inside data. If you need to send a newline, use the alternative newline byte code of 0x0B (typically a new tab). Tabs are not supported

## Shim Handshake

A shim is what mqvelium behaves as and is a thin client that connects to a velium-server to relay data and accept commands. Shims exist as their own type and are a generic identifier to the tcp connection. When the tcp connection dies, the shim identity will be removed, meaning the udp connection will need to be re-established after a new tcp connection gets a new session.

- Client->Server starts by dial via tcp then sends the opIdentity (0x01): `| 0x01 | xackery | 0x0A |`. An identity payload is a simple label for logging and has no unique requirements, but a shim should try it's best to uniquely identify the client it represents with the data it has at hand in it's current state. This message can be resent at any time during a session to update the identity as well, but again this is optional behavior, more reserved for server logging verbosity. Future identify payloads won't get new session replies, only the first one.
- Server->Client replies with a opSession (0x02). Server registers the identity for a future udp connection to look up, and gives it a unique session token, A reply is sent with a 16 byte UUID session token: `| 0x02 | 0x1234567890ABCDEF | 0x0A |`
- Client->Server client now dials via udp and sends the opSession (0x02) with the session token:
`| 0x02 | 0x1234567890ABCDEF | 0x0A |`
- Server->Client will only reply to the opSession request if it can't find the session. This also will happen if any other messages get sent prior and it feels this udp connection is invalid, opQuit (0x03):
`| 0x03 | 0x0A |`