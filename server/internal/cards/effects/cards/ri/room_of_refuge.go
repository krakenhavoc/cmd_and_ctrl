package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "27d5b2eb-94fe-45fb-806c-354044de7ba5",
		Name:         "Room of Refuge",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		AsEnters:     effects.ChooseColorAsEnters(game.ColorForMana, "Room of Refuge"),
		ManaAbilities: []effects.ManaAbility{{
			Cost:         effects.ManaAbilityCost{Tap: true},
			ProducedFunc: effects.ProducedChosenColor(),
			Label:        "Add one mana of the chosen color",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:        "{5}, {T}, Sacrifice this land: Put two +1/+1 counters on target creature. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{5}"), effects.TapCost(), effects.SacrificeThis()),
			SorcerySpeed: true,
			Targets:      effects.TargetCreature("target creature"),
			Effect:       effects.B31CountersOnChosenCreature,
		}},
	})
}
