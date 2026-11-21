package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eerie Gravestone — Artifact {2} (Marvel Super Heroes Commander, #4455):
//
//	"When this artifact enters, draw a card.
//	 {1}{B}, Sacrifice this artifact: Mill four cards. You may put a
//	 creature card from among them into your hand."
//
// The mill is MillToZone with a registered tail, so "from among them"
// offers only the cards this mill put into the graveyard
// (mayTakeOneFromAmongThem), and the choice is a "may".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cbf5bd0a-20ef-4dce-9f70-13339d438ff7",
		Name:         "Eerie Gravestone",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Eerie Gravestone — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, Sacrifice this artifact: Mill four cards. You may put a creature card from among them into your hand.",
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.SacrificeThis()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  eerieGravestoneMill,
		}},
	})
}

func eerieGravestoneMill(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	return effects.MillToZone{N: 4, ThenTail: ctx.ItemTail(eerieGravestoneMilled)}.Apply(ctx)
}

var eerieGravestoneMilled = effects.CardTail("eerie-gravestone/after-mill", func(ctx *effects.Context, o game.TailOutcome) error {
	return effects.MayTakeOneFromAmongThem(ctx, o.IDs, effects.Creature(),
		"Eerie Gravestone — you may put a creature card from among them into your hand")
})
