package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sicarian Infiltrator — Artifact Creature — Human Soldier {2}{U}, 1/2
// (Warhammer 40,000 Commander, S82 GAP batch 195, #5948):
//
//	"Flash
//	 Squad {2} (As an additional cost to cast this spell, you may pay {2} any
//	 number of times. When this creature enters, create that many tokens that are
//	 copies of it.)
//	 Benediction of the Omnissiah — When this creature enters, draw a card."
//
// Squad is Wasteland Raider's optional cost and its enters trigger; each
// token copy enters as a copy and has its own enters draw, so a squad of three
// draws three cards. The draw is its own enters trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "65db1ec4-70b9-45f9-a2c2-01f6c8e8fd89",
		Name:            "Sicarian Infiltrator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		OptionalCosts:   []game.AdditionalCost{effects.SquadMana("{2}", 20)},
		Triggered: []game.TriggeredAbility{
			effects.Squad("Sicarian Infiltrator"),
			effects.WhenThisEnters("Sicarian Infiltrator — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
