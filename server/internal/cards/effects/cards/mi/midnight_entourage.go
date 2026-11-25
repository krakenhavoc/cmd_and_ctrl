package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Midnight Entourage — Creature — Aetherborn Rogue {2}{B}{B}, 3/3 (Aether
// Revolt, S82 older sets, #5536):
//
//	"Other Aetherborn you control get +1/+1.
//	 Whenever this creature or another Aetherborn you control dies, you draw
//	 a card and you lose 1 life."
//
// The death is read off last-known information (CR 603.10a, 608.2h): a
// creature that was an Aetherborn as it died counts, whatever it is in the
// graveyard, so a changeling does too. The life loss waits for the prompt
// the draw may ask (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "02d6b7f6-1fe0-4a93-8385-69ab160ab732",
		Name:         "Midnight Entourage",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Aetherborn"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if effects.CardDied(ev, source) {
					return true
				}
				if ev.CardID == source.InstanceID {
					return false
				}
				dead, ok := effects.DiedCreature(ev, g)
				return ok && effects.LeftUnderControlOf(ev, dead) == source.Controller &&
					effects.LeftAsSubtype(ev, dead, source.TextSubtype("Aetherborn"))
			},
			Key:     "Midnight Entourage — you draw a card and you lose 1 life",
			Purpose: game.Purpose{DeathPayoff: true, Draws: 1},
			Effect: effects.Do(
				effects.DrawCards{N: 1},
				effects.Instruction(func(ctx *effects.Context) error {
					return ctx.Game.ChangePlayerLifeForEffect(ctx.Item.SourceCardID, ctx.Controller(), -1)
				}),
			),
		}},
	})
}
