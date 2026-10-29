package effects

import (
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
	Register(Spec{
		OracleID:        "20535126-f811-4386-bdce-d73f30691724",
		Name:            "Smaug, Wicked Worm",
		ReadsManaSpent:  game.ReadsPoolManaOnly,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Smaug — create X tapped Treasure tokens, X the artifacts your opponents control", smaugTreasures),
			On(game.EventCast, spellPaidWithTreasureMana,
				"Smaug — draw a card and lose 1 life (mana from a Treasure was spent)",
				// The loss waits for a prompt the draw asks (CR 608.2c, #4155).
				Do(
					DrawCards{N: 1},
					Instruction(func(ctx *Context) error {
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
	return CreateTokenAdvanced{Spec: Token(TreasureToken()).EntersTapped(), N: n}.Apply(NewContext(g, item))
}
