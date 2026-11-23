package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Necron Monolith — Artifact — Vehicle {7}, 7/7 (Warhammer 40,000
// Commander, S82 GAP batch 186, #5948):
//
//	"Flying, indestructible
//	 Eternity Gate — Whenever this Vehicle attacks, mill three cards. For
//	 each creature card milled this way, create a 2/2 black Necron
//	 Warrior artifact creature token.
//	 Crew 4"
//
// The mill reports the cards that really reached the graveyard (a
// replacement may send some elsewhere), and one Warrior is made for each
// creature card among them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "20932bca-52be-4288-9c7d-4154650c5ec9",
		Name:            "Necron Monolith",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "indestructible"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 4",
			Cost:    effects.CrewCost(4),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Necron Monolith"),
		}},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Necron Monolith — mill three cards; for each creature card milled this way, create a 2/2 black Necron Warrior artifact creature token",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.MillToZone{N: 3, ThenTail: ctx.ItemTail(necronMonolithMilled)}.Apply(ctx)
				}),
		},
	})
}

// necronMonolithMilled is the clause after the mill: one Warrior for each
// creature card the mill put into the graveyard. An on-disk key: never
// renamed, never reused.
var necronMonolithMilled = effects.CardTail("necron-monolith/after-mill", func(ctx *effects.Context, o game.TailOutcome) error {
	n := 0
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsCreature() {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	return effects.CreateToken{Template: effects.TokenCard("2/2 black Necron Warrior artifact"), N: n}.Apply(ctx)
})
