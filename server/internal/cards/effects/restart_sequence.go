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
// Freerunning {1}{B} is an alternative cost gated on the turn tally of
// who dealt combat damage to a player (freerunning.go): an Assassin or a
// commander under your control, read as the damage was dealt.
//
// The reanimation is complete. The target is chosen from your own
// graveyard through the zone-browser picker and re-checked at
// resolution (CR 608.2b), so a card milled out of the graveyard in
// response fizzles the spell. The creature arrives under YOUR
// control, untapped and summoning sick.
func init() {
	Register(Spec{
		OracleID:         "c415bfdd-3d42-4f94-9f93-7e310becbc8b",
		Name:             "Restart Sequence",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Freerunning("{1}{B}", "Assassin")},
		Targets:          targetCreatureInYourGraveyard(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return returnFirstLegalGraveyardTargetToBattlefield(ctx.Game, item)
		},
	})
}
