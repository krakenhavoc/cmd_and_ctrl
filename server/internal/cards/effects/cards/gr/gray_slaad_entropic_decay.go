package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gray Slaad // Entropic Decay — Creature — Frog Horror {2}{B}, 4/1 //
// Sorcery — Adventure {1}{B} (Commander Legends: Battle for Baldur's Gate,
// #5948):
//
//	"Gray Slaad: As long as there are four or more creature cards in your
//	 graveyard, this creature has menace and deathtouch.
//	 Entropic Decay: Mill four cards. (Then exile this card. You may cast the
//	 creature later from exile.)"
//
// The creature half is a pair of layer 6 keyword grants (CR 613.1f) on itself,
// re-read on every recompute against the number of creature cards in its
// controller's graveyard, so the keywords come and go as the graveyard
// changes. The adventure half is a sorcery that mills its caster for four
// (CR 701.17a); the adventure's exile-and-cast-later is the engine's
// (CR 715.4), as on Beanstalk Giant's Fertile Footsteps.
//
// No simplification.
func init() {
	const oracle = "788897f5-d49d-4294-8006-0009a3124ad0"
	fourCreatures := func(g *game.Game, you uuid.UUID) bool {
		n, _ := effects.Cmdr2S30CreatureCardsIn(g, you)
		return n >= 4
	}
	grant := func(keyword string) game.StaticAbility {
		ab := effects.KeywordGrant(effects.WhileYouHave(fourCreatures, effects.SelfCreature), keyword)
		ab.Label = "As long as there are four or more creature cards in your graveyard, this creature has " + keyword + "."
		return ab
	}
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Gray Slaad",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{grant("menace"), grant("deathtouch")},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Entropic Decay",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.MillToZone{Player: ctx.Controller(), N: 4}.Apply(ctx)
		},
	})
}
