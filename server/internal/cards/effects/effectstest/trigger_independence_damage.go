package effectstest

// trigger_independence_damage_test.go — the aristocrats batches the
// arena kept asking about: a damage step whose source died in the same
// event it triggers on (Syr Konrad in the wipe), "~ deals N damage to
// you" (City of Brass, Midnight Reaper), and the targeted drain (Blood
// Artist, Falkenrath Noble). Each now has a footprint, so a batch whose
// order cannot show goes on with no CR 603.3b prompt.

const (
	SyrKonradOracle         = "14c3ff84-1e82-4606-a433-869fc52cc382"
	FalkenrathNobleOracle   = "3739b179-bc81-4737-8376-66a57e16b942"
	TatyovaOracle           = "0715e860-3b3b-4331-9718-207973e94fee"
	VindictiveVampireOracle = "11ad664e-36d3-4d5b-8a87-59ea715877e0"
)
