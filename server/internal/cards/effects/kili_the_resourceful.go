package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kíli the Resourceful — Legendary Creature — Dwarf Scout {1}{W}, 1/2:
//
//	"Storied (If you control three or more artifacts, legendaries,
//	 and/or Sagas, you have an enduring story for the rest of the game.)
//	 As long as you have an enduring story, you may pay {0} rather than
//	 pay the equip cost of the first equip ability you activate each
//	 turn.
//	 Whenever another Dwarf or Equipment you control enters, draw a card.
//	 This ability triggers only once each turn."
//
// Storied is the engine's (CR 702.195a). The "{0} rather than" is a CR
// 118.9 alternative cost on an activated ability
// (game/equip_cost_alternative.go): while you have an enduring story and
// have activated no equip ability this turn, each equip ability of your
// permanents has a twin row costing {0} and nothing else, with the
// mirrored row's targets, so the choice of paying the equip cost or {0}
// stays yours. Activating either row is your first equip of the turn. The
// draw is an entry trigger tallied once per turn per Kíli
// (TriggersOnlyOnceEachTurn).
//
// No simplification. The Hobbit (#4475).
func init() {
	Register(Spec{
		OracleID:        "840de40b-590e-4bf0-8d16-22b0eff61202",
		Name:            "Kíli the Resourceful",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordStoried},
		EquipCostAlternatives: []game.EquipCostAlternative{{
			Label:         "As long as you have an enduring story, you may pay {0} rather than pay the equip cost of the first equip ability you activate each turn.",
			Cost:          "{0}",
			FirstEachTurn: true,
			Condition:     YouHaveAnEnduringStory,
		}},
		Triggered: []game.TriggeredAbility{
			TriggersOnlyOnceEachTurn(On(game.EventETB, kiliAnotherDwarfOrEquipmentEntered,
				"Kíli the Resourceful — draw a card", Do(DrawCards{N: 1}))),
		},
	})
}

// kiliAnotherDwarfOrEquipmentEntered is "another Dwarf or Equipment you
// control enters".
func kiliAnotherDwarfOrEquipmentEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, true)
	return ok && (c.HasSubtype(source.TextSubtype("Dwarf")) || c.HasSubtype("Equipment"))
}
