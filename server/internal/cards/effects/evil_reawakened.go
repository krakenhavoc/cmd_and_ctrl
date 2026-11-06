package effects

// Evil Reawakened — Sorcery {4}{B} (Final Fantasy, #4462):
//
//	"Return target creature card from your graveyard to the battlefield
//	 with two additional +1/+1 counters on it."
//
// Generous Revival's return with two counters in place of one. The
// counters are a CR 614.1c "enters with", so they ride the entry event:
// an enters trigger finds them already there and a counter replacement
// sees them. A card that left the graveyard in response is an illegal
// target and the spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "264f944b-386b-4378-9764-4a8bde246809",
		Name:         "Evil Reawakened",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature()),
		Purpose:      ForTargets(Reanimates(0)),
		OnResolve:    returnTargetCreatureWithTwoPlusOneCounters,
	})
}
