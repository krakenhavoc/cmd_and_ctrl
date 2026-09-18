package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Restart Sequence — Sorcery {3}{B} (EDHREC rank 4478):
//
//	"Freerunning {1}{B} (You may cast this spell for its freerunning
//	 cost if you dealt combat damage to a player this turn with an
//	 Assassin or commander.)
//	 Return target creature card from your graveyard to the
//	 battlefield."
//
// Reanimate for four, or for two in the deck it was printed for. The
// reanimation itself is unconditional — any creature card, no mana
// value cap, straight to the battlefield — which is what makes it a
// Commander card rather than a limited one.
//
// DECLARED SIMPLIFICATION (weaker than printed): FREERUNNING is not
// implemented, so the spell always costs its full {3}{B}. Freerunning
// is an alternative cost gated on a condition that has to be measured
// across the turn ("you dealt combat damage to a player this turn
// with an Assassin or commander"), and the alternative-cost slot
// carries a price, not a gate — an unconditional {1}{B} offer would
// be strictly STRONGER than the printed card, which is the direction
// #259 forbids, so the offer is left out entirely instead. Paying
// four mana where the card sometimes asks two is the weaker
// direction, which is the allowed one.
//
// The reanimation is complete. The target is chosen from your own
// graveyard through the zone-browser picker and re-checked at
// resolution (CR 608.2b), so a card milled out of the graveyard in
// response fizzles the spell. The creature arrives under YOUR
// control, untapped and summoning sick.
func init() {
	Register(Spec{
		OracleID:     "c415bfdd-3d42-4f94-9f93-7e310becbc8b",
		Name:         "Restart Sequence",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Freerunning isn't implemented — the spell always costs {3}{B}, never the cheaper {1}{B}.",
		},
		Targets: targetCreatureInYourGraveyard(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return returnFirstLegalGraveyardTargetToBattlefield(ctx.Game, item)
		},
	})
}
