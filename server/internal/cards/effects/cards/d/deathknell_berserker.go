package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Deathknell Berserker — Creature — Elf Berserker {1}{B}, 2/2:
//
//	"When this creature dies, if its power was 3 or greater, create a 2/2
//	 black Zombie Berserker creature token."
//
// An intervening "if" (CR 603.4) read from the departure record: the power the
// creature had as it last existed on the battlefield (CR 603.10a), counters and
// effects included, so a Berserker pumped to 3 or shrunk to 1 is judged as it
// died. The record doesn't change, so the check holds as the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4658ae19-d203-4c12-b56e-42cd8f0db91d",
		Name:         "Deathknell Berserker",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				if !effects.ThisDied(ev, source, lki, g) {
					return false
				}
				info, ok := g.LastKnownPermanentForEffect(ev.CardID)
				return ok && info.Power >= 3
			}, "Deathknell Berserker — create a 2/2 black Zombie Berserker creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie Berserker"), N: 1})),
		},
	})
}
