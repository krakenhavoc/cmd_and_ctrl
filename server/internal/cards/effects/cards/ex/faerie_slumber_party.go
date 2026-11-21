package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Faerie Slumber Party — Sorcery {4}{U}{U} (Wilds of Eldraine, #4471):
//
//	"Return all creatures to their owners' hands. For each opponent who
//	 controlled a creature returned this way, you create two 1/1 blue
//	 Faerie creature tokens with flying and "This token can block only
//	 creatures with flying.""
//
// Evacuation's sweep with Flood of Tears's tail: once the creatures are in
// hands "who controlled it" is gone, so the tail reads the swept set (the
// pre-move copies, CR 603.10) and counts the distinct opponents among their
// controllers. A token returned this way counts: it was returned, then
// ceased to exist. The Faerie is Into the Fae Court's token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "48ee9007-2d1c-446f-aee3-795d23028e06",
		Name:         "Faerie Slumber Party",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepBounce}, Tokens: 1},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BounceAllMatching{Match: effects.Creature(), ThenTail: ctx.ItemTail(faerieSlumberPartyFaeries)}.Apply(ctx)
		},
	})
}

// faerieSlumberPartyFaeries is "for each opponent who controlled a creature
// returned this way, you create two Faeries", keyed.
var faerieSlumberPartyFaeries = effects.MassTail("mass/faerie-slumber-party-faeries", func(ctx *effects.Context, swept []game.Card, _ int) error {
	you := ctx.Controller()
	seen := map[uuid.UUID]bool{}
	for _, c := range swept {
		if c.Controller != you {
			seen[c.Controller] = true
		}
	}
	if len(seen) == 0 {
		return nil
	}
	return effects.CreateToken{Template: effects.FlyingBlockerFaerieToken(), N: 2 * len(seen)}.Apply(ctx)
})
