package effects

import (
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
const bushiTenderfootOracleID = "82959ca2-cd96-4cca-9ce0-afb8db209860"

func init() {
	Register(Spec{
		OracleID:     bushiTenderfootOracleID,
		Name:         "Bushi Tenderfoot",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, died := diedCreature(ev, g)
				return died && ev.CardID != source.InstanceID && dealtDamageByThisTurn(g, source.InstanceID, ev.CardID)
			}, "Bushi Tenderfoot — flip it", Do(FlipThis{})),
		},
	})
	Register(Spec{
		OracleID:        bushiTenderfootOracleID + "#1",
		Name:            "Kenzo the Hardhearted",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Triggered:       []game.TriggeredAbility{Bushido(2, "Kenzo the Hardhearted")},
	})
}
