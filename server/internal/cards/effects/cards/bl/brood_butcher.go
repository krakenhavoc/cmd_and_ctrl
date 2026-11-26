package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Brood Butcher — Creature — Eldrazi Drone {3}{B}{G}, 3/3 (Battle for
// Zendikar, #5547):
//
//	"Devoid (This card has no color.)
//	 When this creature enters, create a 1/1 colorless Eldrazi Scion creature
//	 token. It has "Sacrifice this token: Add {C}."
//	 {B}{G}, Sacrifice a creature: Target creature gets -2/-2 until end of
//	 turn."
//
// Scion Summoner's token, and Plagued Rusalka's sacrifice outlet at -2/-2: the
// creature is sacrificed as a cost (CR 601.2h), the target is chosen as the
// ability goes on the stack and re-checked as it resolves; Brood Butcher
// itself, or the Scion, may be the sacrifice.
//
// No simplification.
func init() {
	const name = "Brood Butcher"
	effects.Register(effects.Spec{
		OracleID:        "a20edb8d-219d-4d0f-8346-0c31064eb49f",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Triggered:       []game.TriggeredAbility{effects.OldB307WhenThisEntersCreateScions(name, 1)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}{G}, Sacrifice a creature: Target creature gets -2/-2 until end of turn.",
			Purpose: effects.ForTargets(effects.HarmUntilEOT(0, -2, -2)),
			Cost:    effects.Plus(effects.ManaCost("{B}{G}"), effects.SacrificeACreature()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.BoostTargetUntilEOT("Brood Butcher — -2/-2", -2, -2),
		}},
	})
}
