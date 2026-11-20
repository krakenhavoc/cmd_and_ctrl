package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Smaug, Wicked Worm — Legendary Creature — Dragon {3}{B}{R} (The Hobbit),
// 5/5:
//
//	"Flying
//	 When Smaug enters, create X tapped Treasure tokens, where X is the
//	 number of artifacts your opponents control.
//	 Whenever you cast a spell, if mana from a Treasure was spent to cast
//	 it, you draw a card and lose 1 life."
//
// #4454, S76. The cast trigger is Alchemist's Talent's level 3: the cast
// event's payment record keeps the source kinds of the mana spent (a Treasure
// that sacrificed itself to pay still counts), declared ReadsPoolManaOnly
// because Treasure mana is only ever pool mana. The draw-then-lose is Hired
// Hexblade's, in printed order, the loss a real life loss. X is read as the
// enters trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "20535126-f811-4386-bdce-d73f30691724",
		Name:            "Smaug, Wicked Worm",
		ReadsManaSpent:  game.ReadsPoolManaOnly,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Smaug — create X tapped Treasure tokens, X the artifacts your opponents control", smaugTreasures),
			effects.On(game.EventCast, effects.SpellPaidWithTreasureMana,
				"Smaug — draw a card and lose 1 life (mana from a Treasure was spent)",
				// The loss waits for a prompt the draw asks (CR 608.2c, #4155).
				effects.Do(
					effects.DrawCards{N: 1},
					effects.Instruction(func(ctx *effects.Context) error {
						return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -1)
					}),
				)),
		},
	})
}

func smaugTreasures(g *game.Game, item *game.StackItem) error {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != item.Controller && c.IsArtifact() {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	return effects.CreateTokenAdvanced{Spec: effects.Token(effects.TreasureToken()).EntersTapped(), N: n}.Apply(effects.NewContext(g, item))
}
