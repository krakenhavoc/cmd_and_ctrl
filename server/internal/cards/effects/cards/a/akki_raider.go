package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Akki Raider — Creature — Goblin Warrior {1}{R}, 2/1 (Betrayers of
// Kamigawa, S82 OLD, #5602):
//
//	"Whenever a land is put into a graveyard from the battlefield, this
//	 creature gets +1/+0 until end of turn."
//
// Any player's land counts, the Raider's own controller's or an
// opponent's, and one that was a land as it left the battlefield (CR
// 603.10a): a destroyed, sacrificed or otherwise dead land alike, a bounced
// or exiled one not. Each land is one trigger.
//
// No simplification.
func init() {
	const label = "Akki Raider — +1/+0 until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "bf9c5c0d-86a5-467a-863c-653279f2e6e7",
		Name:         "Akki Raider",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.NewZone != game.ZoneGraveyard {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && effects.LeftAsType(ev, c, "land")
			}, label, effects.ThisGetsUntilEndOfTurn(1, 0, label)),
		},
	})
}
