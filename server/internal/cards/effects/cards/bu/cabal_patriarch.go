package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cabal Patriarch — Legendary Creature — Human Wizard {3}{B}{B}{B}, 5/5
// (Commander Masters):
//
//	"{2}{B}, Sacrifice a creature: Target creature gets -2/-2 until end of
//	 turn.
//	 {2}{B}, Exile a creature card from your graveyard: Target creature gets
//	 -2/-2 until end of turn."
//
// Plagued Rusalka's sacrifice outlet and Ashnod's graveyard exile cost, each
// paying for the same -2/-2 until end of turn (layer 7c, the target's
// toughness read by state-based actions). Two printed abilities, two rows.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "70d1caee-1c35-4fa3-bbc3-a7dcc226be90",
		Name:         "Cabal Patriarch",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}{B}, Sacrifice a creature: Target creature gets -2/-2 until end of turn.",
				Purpose: effects.ForTargets(effects.HarmUntilEOT(0, -2, -2)),
				Cost:    effects.Plus(effects.ManaCost("{2}{B}"), effects.SacrificeACreature()),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.BoostTargetUntilEOT("Cabal Patriarch — -2/-2", -2, -2),
			},
			{
				Label:   "{2}{B}, Exile a creature card from your graveyard: Target creature gets -2/-2 until end of turn.",
				Purpose: effects.ForTargets(effects.HarmUntilEOT(0, -2, -2)),
				Cost:    effects.Plus(effects.ManaCost("{2}{B}"), effects.ExileFromGraveyard(1, "a creature card", effects.IsCreatureCard)),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.BoostTargetUntilEOT("Cabal Patriarch — -2/-2", -2, -2),
			},
		},
	})
}
