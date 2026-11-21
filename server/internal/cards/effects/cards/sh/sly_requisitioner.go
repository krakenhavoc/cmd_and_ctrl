package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sly Requisitioner — Creature — Human Artificer {4}{B}, 2/2:
//
//	"Improvise ... Whenever a nontoken artifact you control is put into
//	 a graveyard from the battlefield, create a 1/1 colorless Servo
//	 artifact creature token."
//
// Improvise is Spec.TapCost. The trigger reads the exit's last-known
// types (leftAsType), so an artifact that stopped being one is judged as
// it last existed; the controller is the one it left under.
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "15ea4698-86b4-47c5-a9a7-60af16063a61",
		Name:         "Sly Requisitioner",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Improvise(),
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.NewZone != game.ZoneGraveyard {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && !effects.IsToken(c) && effects.LeftAsType(ev, c, "artifact") &&
					effects.LeftUnderControlOf(ev, c) == source.Controller
			}, "Sly Requisitioner — create a Servo", effects.Do(effects.CreateToken{
				Template: effects.TokenCard("1/1 colorless Servo artifact"), N: 1,
			})),
		},
	})
}
