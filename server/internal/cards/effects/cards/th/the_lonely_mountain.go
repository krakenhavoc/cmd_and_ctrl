package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Lonely Mountain — Land — Mountain (The Hobbit, HOB):
//
//	"({T}: Add {R}.)
//	 This land enters tapped unless you control an Equipment.
//	 {4}{R}, {T}: Create a 2/2 red Dwarf creature token. This ability
//	 costs {1} less to activate for each Equipment you control. Activate
//	 only as a sorcery."
//
// Typed lands carry their mana as a row (the engine does not synthesise it
// for a nonbasic), and the tapped entry is the real CR 614 self-replacement,
// judged as the land enters. The token ability's own cost clause is Mobilized
// District's (ActivatedAbility.CostModifiers): {1} less for each Equipment
// you control as the cost is determined (CR 602.2b), floored at {R}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3678c06f-8a33-4a6d-bf20-5b92d5c05a95",
		Name:         "The Lonely Mountain",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTappedUnless(youControlAnEquipment)},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "Add {R}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label: "{4}{R}, {T}: Create a 2/2 red Dwarf creature token. This ability costs {1} less to activate for each Equipment you control. Activate only as a sorcery.",
			Cost:  effects.Plus(effects.ManaCost("{4}{R}"), effects.TapCost()),
			CostModifiers: []game.CostModifier{
				effects.CostsLessEach(effects.PermanentsYouControl(effects.HasSubtype("Equipment")),
					"This ability costs {1} less to activate for each Equipment you control."),
			},
			SorcerySpeed: true,
			Effect:       effects.CreateTheToken("2/2 red Dwarf"),
		}},
	})
}

// youControlAnEquipment is the "unless you control an Equipment" test of
// The Lonely Mountain's tapped entry.
func youControlAnEquipment(g *game.Game, controller uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.HasSubtype("Equipment") {
			return true
		}
	}
	return false
}
