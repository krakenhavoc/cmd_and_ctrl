package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mycoid Shepherd — Creature — Fungus {1}{G}{G}{W}, 5/4 (Alara Reborn, #5586):
//
//	"Whenever this creature or another creature you control with power 5 or
//	 greater dies, you may gain 5 life."
//
// One ability for the Shepherd or any creature you controlled when it died; the
// power is the creature's last-known power on the battlefield (CR 603.10a,
// CR 608.2h: counters and anthems included), read when it dies. The life gain is
// optional.
//
// No simplification.
func init() {
	const label = "Mycoid Shepherd — you may gain 5 life"
	effects.Register(effects.Spec{
		OracleID:     "c4a4e607-7f37-4dc6-8fc0-f77bdafeb919",
		Name:         "Mycoid Shepherd",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventLTB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				if !effects.ACreatureYouControlDied(ev, source, lki, g) {
					return false
				}
				return effects.B13LastKnownPower(g, ev.CardID, lki) >= 5
			}, label, effects.Do(effects.GainLife{Amount: 5})), "Mycoid Shepherd — gain 5 life?"),
		},
	})
}
