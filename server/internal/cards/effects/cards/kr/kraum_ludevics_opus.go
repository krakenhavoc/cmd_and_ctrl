package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kraum, Ludevic's Opus — Legendary Creature — Zombie Horror {3}{U}{R},
// 4/4: "Flying, haste. Whenever an opponent casts their second spell each
// turn, draw a card. Partner."
//
// The per-player cast tally is bumped before EventCast fires, so the spell
// that makes that opponent's count two is their second; the tally is the
// whole turn's, whether or not Kraum was out for the first.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5f694f2d-9349-4988-b2b4-788136d4293e",
		Name:            "Kraum, Ludevic's Opus",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, anOpponentCastTheirSecondSpellThisTurn,
				"Kraum, Ludevic's Opus — draw a card",
				effects.Do(effects.DrawCards{N: 1})),
		},
	})
}

func anOpponentCastTheirSecondSpellThisTurn(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor == source.Controller {
		return false
	}
	return g.CastTallyFor(ev.Actor).Total == 2
}
