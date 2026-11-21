package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dreadbore — Sorcery {B}{R} (EDHREC rank 4494):
//
//	"Destroy target creature or planeswalker."
//
// Two mana, no conditions, no regeneration clause, and the
// planeswalker half printed rather than bolted on — which is what
// separates it from Terror and from Doom Blade and is why a Rakdos
// deck runs it over either. The sorcery speed is the price.
//
// Both halves are one target clause: Or(Creature(), Planeswalker())
// over the battlefield, re-checked at resolution (CR 608.2b), so a
// creature that gained hexproof or left in response fizzles the
// spell rather than destroying something else. "Destroy" is the
// ordinary destruction verb, so indestructible (CR 702.12) and
// regeneration shields both do what they say.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d685799f-cc1a-40d6-9df6-d05b8b1f5b13",
		Name:         "Dreadbore",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetPermanent("target creature or planeswalker",
			effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyChosenPermanent(ctx.Game, item)
		},
	})
}
