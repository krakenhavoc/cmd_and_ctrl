package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cave of Temptation — Land (Foundations Jumpstart, #5929):
//
//	"{T}: Add {C}.
//	 {1}, {T}: Add one mana of any color.
//	 {4}, {T}, Sacrifice this land: Put two +1/+1 counters on target
//	 creature. Activate only as a sorcery."
//
// The same three abilities as Captivating Cave.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "75540897-53f6-433b-bd70-9851551df6ef",
		Name:         "Cave of Temptation",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			effects.PainlessColorless(),
			{
				Cost:     effects.ManaAbilityCost{Tap: true, Mana: "{1}"},
				Produced: "{W|U|B|R|G}",
				Label:    "{1}, {T}: Add one mana of any color",
			},
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{4}, {T}, Sacrifice this land: Put two +1/+1 counters on target creature. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{4}"), effects.TapCost(), effects.SacrificeThis()),
			SorcerySpeed: true,
			Targets:      effects.TargetCreature("target creature"),
			Effect:       effects.B31CountersOnChosenCreature,
		}},
	})
}
