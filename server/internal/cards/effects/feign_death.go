package effects

// Feign Death — Instant {B}:
//
//	"Until end of turn, target creature gains "When this creature
//	 dies, return it to the battlefield tapped under its owner's
//	 control with a +1/+1 counter on it.""
//
// A duration grant (ADR 0093 PR 4, #1584): the dies trigger is a
// catalog bundle and the spell gives it to the target until end of
// turn as a ScopedEffect record. The trigger is the CREATURE's — its
// controller controls it, a later "loses all abilities" takes it
// (CR 613.6), and it fires from last-known information as the
// creature dies (CR 603.10a). The +1/+1 counter is an "enters with"
// clause riding the return (CR 614.1c), so Hardened Scales sees it.
//
// The returned creature is a new object (CR 400.7) the grant never
// named, so it does not come back a second time.
//
// The spell is registerReturnWithCounterSpell, shared with Undying Malice.
//
// No simplification.
const feignDeathReturn = "feign-death/return"

func init() {
	registerReturnWithCounterSpell("f718e507-296b-4f22-842b-5fb91322069b", "Feign Death", feignDeathReturn)
}
