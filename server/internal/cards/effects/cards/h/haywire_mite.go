package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Haywire Mite — Artifact Creature — Insect {1}, 1/1 (EDHREC rank
// 683):
//
//	"When this creature dies, you gain 2 life.
//	 {G}, Sacrifice this creature: Exile target noncreature artifact
//	 or noncreature enchantment."
//
// A one-mana answer to a Rhystic Study or a Sol Ring that also gains
// two life on the way out — the sacrifice IS a death, so the dies
// trigger fires off the activation and resolves above the exile.
// The exile is a CR 602 activated ability with a mana plus
// sacrifice-self cost and a target clause: artifact or enchantment,
// and NOT a creature (so an artifact creature is out of reach, as
// printed).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "749d2994-44e7-40d3-8630-7bebed239e9e",
		Name:         "Haywire Mite",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Haywire Mite — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{G}, Sacrifice this creature: Exile target noncreature artifact or noncreature enchantment.",
			Cost:  effects.Plus(effects.ManaCost("{G}"), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target noncreature artifact or noncreature enchantment",
				effects.Or(effects.Artifact(), effects.Enchantment()), effects.Noncreature()),
			Effect: effects.ExileFirstTarget,
		}},
	})
}
