package mcpseat

// The `ack` frame, ADR 0122 §6.4 (Delivery PR 5).
//
// Written here to the ADR because PR 5 has not merged into this branch.
// When it lands, kindAck and ackPayload give way to the protocol
// package's own constant and type; nothing else in mcpseat names them.
//
//	{"v": 0, "kind": "ack", "id": "<the action's id>",
//	 "payload": {"seq": <uint64>, "generation": <uint64>}}
//
// Sent to the originating connection only, after every applied action,
// and always after the snapshot of the state it acknowledges. A refused
// action still gets an `error` frame with the action's id, never an ack.

const kindAck = "ack"

// ackPayload names the state the acknowledged action produced.
type ackPayload struct {
	Seq        uint64 `json:"seq"`
	Generation uint64 `json:"generation"`
}
