package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Court Street Denizen — Creature — Human Soldier {2}{W}, 2/2 (Commander
// Legends, #5507):
//
//	"Whenever another white creature you control enters, tap target
//	 creature an opponent controls."
//
// Shadow Alley Denizen's trigger (the entering creature's effective colour,
// CR 105.2; "another" is object identity) with Rimefur Reindeer's tap of a
// target creature an opponent controls. The target is chosen as the trigger
// goes on the stack and rechecked as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "34312179-f3f9-47ca-bdf4-5abbd493876e",
		Name:         "Court Street Denizen",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !effects.B13AnotherCreatureYouControlEntered(ev, source, g) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.HasColor(source.TextColor("W"))
			},
			Targets: effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
			Key:     "Court Street Denizen — tap target creature an opponent controls",
			Effect:  effects.TapTheFirstLegalTarget,
		}},
	})
}
