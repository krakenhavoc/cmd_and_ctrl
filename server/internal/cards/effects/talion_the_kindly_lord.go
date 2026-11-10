package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Talion, the Kindly Lord — Legendary Creature — Faerie Noble {2}{U}{B},
// 3/4 (Wilds of Eldraine, #4471):
//
//	"Flying
//	 As Talion enters, choose a number between 1 and 10.
//	 Whenever an opponent casts a spell with mana value, power, or
//	 toughness equal to the chosen number, that player loses 2 life and
//	 you draw a card."
//
// The number is Nyota Uhura's: a prompt as Talion enters, stored on the
// permanent (game.Card.ChosenNumber), here from 1 to 10 inclusive. The
// trigger reads the spell as it is cast: its mana value (with X as
// announced, the ruling; cost reductions don't change it) or, for a
// creature spell, its power or toughness. One trigger per spell, whichever
// of the three matched, and the casting opponent is the one who loses the
// life.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a019e34d-00c4-4687-9d41-fe550bf1dbdc",
		Name:            "Talion, the Kindly Lord",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		AsEnters:        talionChooseANumber,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, talionOpponentCastTheNumber,
				"Talion, the Kindly Lord — that player loses 2 life and you draw a card",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := g.ChangePlayerLifeForEffect(item.SourceCardID, item.Trigger.Event.Actor, -2); err != nil {
						return err
					}
					return DrawCards{N: 1}.Apply(ctx)
				}),
		},
	})
}

// talionChooseANumber is the as-enters prompt: 1 to 10, stored on Talion.
func talionChooseANumber(card *game.Card, ctx *Context) error {
	return ctx.Game.QueueChooseNumberForEffect(game.ChooseNumber{
		Chooser:  card.Controller,
		Source:   card.InstanceID,
		Resource: game.PayResourceNone,
		Min:      1,
		Max:      10,
		Goal:     3,
		Question: "Talion, the Kindly Lord — choose a number between 1 and 10",
		Branch:   game.NumberBranch{Then: storeChosenNumber, Carry: []uuid.UUID{card.Controller}},
	})
}

// talionOpponentCastTheNumber is "whenever an opponent casts a spell with
// mana value, power, or toughness equal to the chosen number".
func talionOpponentCastTheNumber(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	if !AnOpponentCast(nil)(ev, source, lki, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || source.ChosenNumber < 1 {
		return false
	}
	n := source.ChosenNumber
	if mv, known := g.ManaValueForEffect(c); known && mv == n {
		return true
	}
	return c.IsCreature() && (c.PowerForComparison() == n || c.CurrentToughness() == n)
}
