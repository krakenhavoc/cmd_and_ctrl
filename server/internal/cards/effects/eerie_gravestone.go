package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "cbf5bd0a-20ef-4dce-9f70-13339d438ff7",
		Name:         "Eerie Gravestone",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Eerie Gravestone — draw a card", Do(DrawCards{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}, Sacrifice this artifact: Mill four cards. You may put a creature card from among them into your hand.",
			Cost:    Plus(ManaCost("{1}{B}"), SacrificeThis()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  eerieGravestoneMill,
		}},
	})
}

func eerieGravestoneMill(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return MillToZone{N: 4, ThenTail: ctx.ItemTail(eerieGravestoneMilled)}.Apply(ctx)
}

var eerieGravestoneMilled = CardTail("eerie-gravestone/after-mill", func(ctx *Context, o game.TailOutcome) error {
	return mayTakeOneFromAmongThem(ctx, o.IDs, Creature(),
		"Eerie Gravestone — you may put a creature card from among them into your hand")
})
