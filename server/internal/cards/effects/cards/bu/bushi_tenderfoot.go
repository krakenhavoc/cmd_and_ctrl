package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bushi Tenderfoot // Kenzo the Hardhearted — {W} Creature — Human
// Soldier 1/1 // Legendary Creature — Human Samurai 3/4, a flip card
// (#3216, ADR 0151):
//
//	Top:     "When a creature dealt damage by this creature this turn
//	          dies, flip this creature."
//	Flipped: "Double strike; bushido 2 (Whenever this creature blocks or
//	          becomes blocked, it gets +2/+2 until end of turn.)"
//
// Audited against the engine (the issue's "check"): "dealt damage by
// this creature this turn" is read off the turn's record of which
// objects each source damaged (dealtDamageByThisTurn, #1946). The source
// and the victim are each named by object, so either one re-entering the
// battlefield is a new object (CR 400.7), and a creature
// that was dealt damage, blinked and died is not the one that was
// damaged. The trigger is "a creature dies", not "another", but a
// Tenderfoot that dies with its victim is gone and has nothing to flip.
//
// Kenzo is white and {W} (CR 710.1c), and has double strike and bushido
// 2 as its own half's text (ADR 0151 decision 3).
//
// No simplification.
const BushiTenderfootOracleID = "82959ca2-cd96-4cca-9ce0-afb8db209860"

func init() {
	effects.Register(effects.Spec{
		OracleID:     BushiTenderfootOracleID,
		Name:         "Bushi Tenderfoot",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, died := effects.DiedCreature(ev, g)
				return died && ev.CardID != source.InstanceID && effects.DealtDamageByThisTurn(g, source.InstanceID, ev.CardID)
			}, "Bushi Tenderfoot — flip it", effects.Do(effects.FlipThis{})),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        BushiTenderfootOracleID + "#1",
		Name:            "Kenzo the Hardhearted",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Triggered:       []game.TriggeredAbility{effects.Bushido(2, "Kenzo the Hardhearted")},
	})
}
