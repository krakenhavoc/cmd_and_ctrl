package cardsh

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Herald of Anafenza — Creature — Human Soldier {W}, 1/2 (Khans of Tarkir):
//
//	"Outlast {2}{W} ({2}{W}, {T}: Put a +1/+1 counter on this creature.
//	 Outlast only as a sorcery.)
//	 Whenever you activate this creature's outlast ability, create a 1/1 white
//	 Warrior creature token."
//
// Outlast is the shared sorcery-speed tap-and-pay row (CR 702.107a). The
// trigger watches the activation itself (CR 602.2b, EventActivateAbility), not
// its resolution: the token is made when the ability is put on the stack and
// stays even if the ability is countered. It fires only for this creature's
// own outlast row, by its label.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "889d1056-4d3d-42f0-96f2-16529a4a94e3",
		Name:         "Herald of Anafenza",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.RprB04Outlast("{2}{W}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventActivateAbility, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller && ev.CardID == source.InstanceID && strings.HasPrefix(ev.Label, "Outlast ")
			}, "Herald of Anafenza — create a 1/1 white Warrior creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Warrior"), N: 1})),
		},
	})
}
