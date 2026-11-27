package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brown Ouphe — Creature — Ouphe {G}, 1/1 (Mirrodin, #5606):
//
//	"{1}{G}, {T}: Counter target activated ability from an artifact source.
//	 (Mana abilities can't be targeted.)"
//
// Azorius Guildmage's counter over activated abilities on the stack (a spell, a
// triggered ability and a mana ability are no legal target), with the ability's
// source required to be an artifact, read as the object it is or last was.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5b42e91a-e3bd-4f69-9abf-ed6273715fff",
		Name:         "Brown Ouphe",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}, {T}: Counter target activated ability from an artifact source.",
			Cost:    effects.Plus(effects.ManaCost("{1}{G}"), effects.TapCost()),
			Targets: effects.TargetSpellOrAbility("target activated ability from an artifact source", effects.AnActivatedAbilityItem(), effects.OldB607FromAnArtifactSource()),
			Effect:  effects.CounterTheChosenSpell,
		}},
	})
}
