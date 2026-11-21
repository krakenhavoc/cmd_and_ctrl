package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cathar Commando — Creature — Human Soldier {1}{W}, 3/1 (EDHREC
// rank 1416):
//
//	"Flash
//	 {1}, Sacrifice this creature: Destroy target artifact or
//	 enchantment."
//
// A Disenchant that is also a flash 3/1 — the white deck's answer
// that can block first. Flash is the printed keyword; the removal is
// a CR 602 activated ability with a mana-plus-sacrifice-self cost
// (Haywire Mite's shape) and a real target clause, so it uses the
// stack and the sacrifice pays at announce — dies-triggers on the
// Commando land above the ability and resolve first.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "774dce79-67e0-4820-8013-c7a7347993ce",
		Name:            "Cathar Commando",
		Completeness:    effects.CompletenessFull,
		Purpose:         game.Purpose{Answers: game.AnswerMakesBlocker}, // #4734: a creature with flash is a blocker (sweep ruling 7)
		PrintedKeywords: []string{"flash"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice this creature: Destroy target artifact or enchantment.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
			Effect:  effects.DestroyChosenPermanent,
		}},
	})
}
