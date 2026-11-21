package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vodalian Hexcatcher — Creature — Merfolk Wizard {1}{U}, 1/1 (Dominaria
// United, #5487):
//
//	"Flash
//	 Other Merfolk you control get +1/+1.
//	 Sacrifice a Merfolk: Counter target noncreature spell unless its
//	 controller pays {1}."
//
// The lord is Merfolk Mistbinder's tribal anthem (the layered view, so a
// changeling is a Merfolk). The cost is Goblin Chirurgeon's "Sacrifice a
// <type>": any Merfolk you control, the Hexcatcher itself included. The
// counter is Malevolent Hermit's CounterUnlessPaid at {1}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4246b503-0da8-4de7-9343-9dfb3c3672df",
		Name:            "Vodalian Hexcatcher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.Cmdr2B05MerfolkFilter(true, true), 1, 1),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a Merfolk: Counter target noncreature spell unless its controller pays {1}.",
			Cost:    game.AbilityCost{SacrificeOther: effects.SacrificeSpec("a Merfolk", effects.Subtype("Merfolk"))},
			Targets: effects.TargetSpell("target noncreature spell", effects.Noncreature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				return effects.CounterUnlessPaid{
					StackID:  item.Targets[0].ID,
					Cost:     "{1}",
					Question: "Vodalian Hexcatcher — pay {1} or your spell is countered",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
