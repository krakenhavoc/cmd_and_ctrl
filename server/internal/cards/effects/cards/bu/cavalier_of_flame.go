package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cavalier of Flame — Creature — Elemental Knight {2}{R}{R}{R}, 6/5 (Core Set
// 2020, #5518):
//
//	"{1}{R}: Creatures you control get +1/+0 and gain haste until end of turn.
//	 When this creature enters, discard any number of cards, then draw that
//	 many cards.
//	 When this creature dies, it deals X damage to each opponent and each
//	 planeswalker they control, where X is the number of land cards in your
//	 graveyard."
//
// The pump is Scurry of Gremlins's Overrun shape over the creatures you
// control as it resolves (CR 611.2c). The enter trigger is the discard prompt
// (up to the whole hand) whose keyed answer draws as many as were discarded;
// an empty hand asks nothing. The dies trigger counts the land cards in the
// graveyard as it resolves, deals the damage from the Cavalier as it last
// existed (CR 608.2h), and does it as one instruction, so one damage instance
// (CR 615.8).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "98b19afc-7696-46b9-96f0-7f7c60003e1c",
		Name:         "Cavalier of Flame",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}: Creatures you control get +1/+0 and gain haste until end of turn.",
			Purpose: game.Purpose{TeamPump: &game.Pump{Power: 1, Keywords: []string{"haste"}}, Answers: game.AnswerPump | game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{1}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				yours := effects.And(effects.Creature(), effects.YouControl())
				if err := (effects.BoostUntilEOT{Match: yours, Power: 1, Label: "Cavalier of Flame — +1/+0"}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Match: yours, Keywords: []string{"haste"}, Label: "Cavalier of Flame — haste"}.Apply(ctx)
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Cavalier of Flame — discard any number of cards, then draw that many", effects.OldB154CavalierDiscardDraw),
			effects.WhenThisDies("Cavalier of Flame — X damage to each opponent and each planeswalker they control", effects.OldB154CavalierDeath),
		},
	})
}
