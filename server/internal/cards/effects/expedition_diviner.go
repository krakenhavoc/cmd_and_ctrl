package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Expedition Diviner — Creature — Merfolk Wizard {3}{U}, 3/2 (Foundations
// Jumpstart):
//
//	"Flying
//	 As long as you control another Wizard, this creature has "When this
//	 creature dies, draw a card.""
//
// A layer 6 grant of the dies trigger to the source alone, gated on some other
// Wizard its controller controls. A Diviner that dies while another Wizard is
// out had the ability as it last existed (CR 603.10a), so the draw fires.
//
// No simplification.
const expeditionDivinerGrant = "expedition-diviner/draw"

func init() {
	Register(Spec{
		OracleID:        "bd6bc87d-fc75-43bc-a7d7-49d89dabcd3a",
		Name:            "Expedition Diviner",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Grants: []AbilityGrant{{
			Key: expeditionDivinerGrant,
			Triggered: []game.TriggeredAbility{
				WhenThisDies("Expedition Diviner — draw a card", Do(DrawCards{N: 1})),
			},
			Text: "When this creature dies, draw a card.",
		}},
		Static: []game.StaticAbility{
			Labelled(GrantAbilities(bnsB29SelfWhileAnother("Wizard"), expeditionDivinerGrant),
				"As long as you control another Wizard, this creature has \"When this creature dies, draw a card.\""),
		},
	})
}
