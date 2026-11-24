package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soulherder — Creature — Spirit {1}{W}{U}, 1/1 (Secret Lair Drop):
//
//	"Whenever a creature is exiled from the battlefield, put a +1/+1 counter
//	 on this creature.
//	 At the beginning of your end step, you may exile another target creature
//	 you control, then return that card to the battlefield under its owner's
//	 control."
//
// The counter trigger watches every battlefield-to-exile exit of a permanent
// that was a creature as it last existed there (CR 603.10a), yours or not,
// and never fires for Soulherder's own exile (it is gone by then).
//
// The end-step ability targets "another target creature you control" when it
// goes on the stack and asks "you may" as it resolves, as Conjurer's Closet
// does; the blink returns the card as a new object under its OWNER's control
// (CR 400.7), so the counter trigger sees the exile and Soulherder grows.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "92019547-f6db-4ea6-8356-d0a90ace5662",
		Name:         "Soulherder",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneExile {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && effects.LeftAsType(ev, c, "creature")
			}, "Soulherder — put a +1/+1 counter on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			effects.Optional(
				effects.Targeting(
					effects.AtYourEndStep("Soulherder — exile another target creature you control, then return that card to the battlefield under its owner's control",
						effects.FlickerFirstLegalTargetUnderOwner),
					effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
				"Soulherder — exile another target creature you control, then return it to the battlefield under its owner's control?"),
		},
	})
}
