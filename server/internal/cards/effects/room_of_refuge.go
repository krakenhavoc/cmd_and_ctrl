package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Room of Refuge — Land:
//
//	"This land enters tapped. As it enters, choose a color.
//	 {T}: Add one mana of the chosen color.
//	 {5}, {T}, Sacrifice this land: Put two +1/+1 counters on target
//	 creature. Activate only as a sorcery."
//
// The one-colour chosen-colour land shape (chosen_color_lands.go: a
// tapped entry, an as-enters colour prompt, a mana ability read off the
// stored colour) plus a sorcery-speed activated ability whose cost is
// mana, tap and sacrifice. The land is sacrificed as the cost is paid,
// so the counters land even if the Room is gone by resolution.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "27d5b2eb-94fe-45fb-806c-354044de7ba5",
		Name:         "Room of Refuge",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		AsEnters:     ChooseColorAsEnters(game.ColorForMana, "Room of Refuge"),
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			ProducedFunc: ProducedChosenColor(),
			Label:        "Add one mana of the chosen color",
		}},
		Activated: []ActivatedAbility{{
			Label:        "{5}, {T}, Sacrifice this land: Put two +1/+1 counters on target creature. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{5}"), TapCost(), SacrificeThis()),
			SorcerySpeed: true,
			Targets:      TargetCreature("target creature"),
			Effect:       b31CountersOnChosenCreature,
		}},
	})
}
