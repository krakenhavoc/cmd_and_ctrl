package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eldrazi Displacer — Creature — Eldrazi {2}{W}, 3/3:
//
//	"Devoid (This card has no color.)
//	 {2}{C}: Exile another target creature, then return it to the
//	 battlefield tapped under its owner's control. ({C} represents
//	 colorless mana.)"
//
// Devoid is the printed keyword the engine reads as CR 702.114a, a
// colour-setting static in every zone (game.KeywordDevoid). The {C} in
// the cost is a colourless symbol, which only colourless mana pays
// (CR 107.4c), never mana of a colour.
//
// The ability is the immediate blink with the return tapped (The Mighty
// Thor's mightyThorFlickerTapped): one resolution, no delayed trigger.
// The creature comes back as a new object under its OWNER's control
// (CR 400.7), tapped and no longer in combat. A target that left
// in response is a no-op (CR 608.2b). The ability has no tap cost, so
// it can be activated as often as the mana allows, summoning sickness
// notwithstanding.
//
// No simplification. Roon deck request (#4128).
func init() {
	Register(Spec{
		OracleID:        "9b1f552a-bddc-4fcb-ac67-b4a65b2f48ba",
		Name:            "Eldrazi Displacer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Activated: []ActivatedAbility{{
			Label:   "{2}{C}: Exile another target creature, then return it to the battlefield tapped under its owner's control.",
			Cost:    ManaCost("{2}{C}"),
			Targets: Another(TargetCreature("another target creature")),
			Effect:  mightyThorFlickerTapped,
		}},
	})
}
