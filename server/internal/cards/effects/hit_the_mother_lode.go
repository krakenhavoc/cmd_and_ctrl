package effects

// Hit the Mother Lode — Sorcery {4}{R}{R}{R}:
//
//	"Discover 10. If the discovered card's mana value is less than 10,
//	 create a number of tapped Treasure tokens equal to the difference."
//
// The Treasures read CR 701.57c's discovered card, which the discover
// files as a step record (DiscoverStepKey) before it asks whether to
// cast it. They are the next step, so they wait for the answer (the cast
// is made during the resolution, CR 608.2g, and the card is cast or put
// into the hand either way), and a table restored while the question is
// open still makes them. An empty walk discovers no card, so it makes no
// Treasures (ADR 0099, owner decision 6).
func init() {
	Register(Spec{
		OracleID:     "a49900b3-cc34-428f-806d-71861dc8059d",
		Name:         "Hit the Mother Lode",
		Completeness: CompletenessFull,
		Discovers:    true,
		OnResolve:    ResolveSteps(Discover{N: 10}, Instruction(hitTheMotherLodeTreasures)),
	})
}

func hitTheMotherLodeTreasures(ctx *Context) error {
	rec := ctx.StepRecord(DiscoverStepKey)
	if len(rec.IDs) == 0 || rec.N() >= 10 {
		return nil
	}
	return b13CreateTappedTreasures(ctx, ctx.Controller(), 10-rec.N())
}
