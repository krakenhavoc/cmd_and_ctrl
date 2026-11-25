package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ral's Staticaster — Creature — Lizard Wizard {2}{U}{R}, 3/3 (Guilds of
// Ravnica, #5525):
//
//	"Trample
//	 Whenever this creature attacks, if you control a Ral planeswalker, this
//	 creature gets +1/+0 for each card in your hand until end of turn."
//
// An intervening "if" (CR 603.4): checked as it attacks and again as the
// ability resolves. "A Ral planeswalker" is the subtype on a planeswalker; the
// number of cards in hand is counted as the ability resolves (CR 608.2h) and
// locked for the turn.
//
// No simplification.
func init() {
	const name = "Ral's Staticaster"
	effects.Register(effects.Spec{
		OracleID:        "eae7dd41-3d5f-49a7-baaa-2c03bb341c3a",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIfYou("you control a Ral planeswalker",
				func(g *game.Game, you uuid.UUID) bool { return effects.B42ControlsPlaneswalkerNamed(g, you, "Ral") },
				effects.WheneverThisAttacks(name+" — +1/+0 for each card in your hand until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						return effects.BoostUntilEOT{
							Target: item.SourceCardID,
							Power:  effects.CardsInHand(g, item.Controller),
							Label:  name + " — +1/+0 for each card in your hand until end of turn",
						}.Apply(effects.NewContext(g, item))
					})),
		},
	})
}
