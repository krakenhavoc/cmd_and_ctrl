package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Appa, the Vigilant — Legendary Creature — Bison Ally {5}{W}{W}, 6/6
// (Avatar: The Last Airbender Eternal, #5075):
//
//	"Flying, vigilance
//	 Whenever Appa or another Ally you control enters, creatures you
//	 control get +1/+1 and gain flying and vigilance until end of turn."
//
// One trigger for Appa itself or another Ally entering under your
// control (South Pole Voyager's condition). The creatures it affects are
// fixed as it resolves (CR 611.2c), so a creature that enters later in
// the turn gets nothing from this instance.
//
// No simplification.
func init() {
	you := effects.And(effects.Creature(), effects.YouControl())
	effects.Register(effects.Spec{
		OracleID:        "72994a7a-7c0d-4f28-8cc0-28161d098d7f",
		Name:            "Appa, the Vigilant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.CardID == source.InstanceID ||
					effects.B10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev, source, g, source.TextSubtype("Ally"))
			}, "Appa, the Vigilant — creatures you control get +1/+1 and gain flying and vigilance until end of turn",
				effects.Do(
					effects.BoostUntilEOT{Match: you, Power: 1, Toughness: 1, Label: "Appa, the Vigilant — +1/+1"},
					effects.GrantKeywordUntilEOT{Match: you, Keywords: []string{"flying", "vigilance"}, Label: "Appa, the Vigilant — flying and vigilance"},
				)),
		},
	})
}
