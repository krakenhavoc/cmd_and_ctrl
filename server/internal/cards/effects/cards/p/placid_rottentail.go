package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Placid Rottentail — Creature — Fungus Rabbit {G}, 1/1 (March of the
// Machine, S82 MOM batch 13):
//
//	"Vigilance
//	 {2}{G}, Exile this card from your graveyard: Put two +1/+1 counters
//	 on target creature. Activate only as a sorcery."
//
// Adorned Crocodile's graveyard ability (Renew's shape, written out
// because the printed text carries no keyword name): activated only from
// the graveyard, the exile is part of the cost, and the target is chosen
// as it is activated. A target gone by resolution takes nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "502f8beb-0132-492f-bbfb-acf1ea532967",
		Name:            "Placid Rottentail",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []effects.ActivatedAbility{{
			Label:        "{2}{G}, Exile this card from your graveyard: Put two +1/+1 counters on target creature. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{2}{G}"), effects.ExileThis()),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Targets:      effects.TargetCreature("target creature"),
			Effect:       effects.RenewCountersOnTarget(game.CounterPlusOne, 2),
		}},
	})
}
