package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hearthcage Giant — Creature — Giant Warrior {6}{R}{R}, 5/5 (Lorwyn, S82 OLD,
// #5593):
//
//	"When this creature enters, create two 3/1 red Elemental Shaman creature
//	 tokens.
//	 Sacrifice an Elemental: Target Giant creature gets +3/+1 until end of turn."
//
// Two 3/1 Elemental Shaman tokens, and Seething Pathblazer's sacrifice over
// effective creature types aimed at a Giant. A token just made may pay, and the
// Giant may be the Elemental sacrificed (it is then not a legal target as the
// ability resolves, CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7e523d42-c7d8-47d1-8874-230920f03345",
		Name:         "Hearthcage Giant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Hearthcage Giant — create two 3/1 red Elemental Shaman tokens",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("3/1 red Elemental Shaman"), N: 2})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice an Elemental: Target Giant creature gets +3/+1 until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 3, 1)),
			Cost:    effects.SacrificeN(1, "an Elemental", effects.OfCreatureType("Elemental")),
			Targets: effects.TargetCreature("target Giant creature", effects.OfCreatureType("Giant")),
			Effect:  effects.BoostTargetUntilEOT("Hearthcage Giant — +3/+1 until end of turn", 3, 1),
		}},
	})
}
