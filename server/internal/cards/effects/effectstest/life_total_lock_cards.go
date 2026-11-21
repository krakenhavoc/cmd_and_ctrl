package effectstest

// life_total_lock_cards_test.go — #1200's three cards, each proved
// through the engine rather than by reading its own Spec back.
//
// The assertions are about what a PLAYER at the table would notice:
// the drain that drains nobody, the life gain that does not arrive,
// the shockland that enters tapped because you could not pay for it,
// and the board that comes back at the same moment the shield lifts.
// A test that only checked `spec.PlayerLifeTotalLocked` would pass
// with every consumer backed out.

const (
	PlatinumEmperionOracle = "bcd2f70c-36c2-44b2-9d4b-1000e9bb62b6"
	TeferisReproachOracle  = "9381b4a5-a8e0-412d-bc7f-ae15afa0f135"
)
