package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lord of the Forsaken — Creature — Demon {4}{B}{B}, 6/6 (Tarkir: Dragonstorm Commander)
// Commander):
//
//	"Flying, trample
//	 {B}, Sacrifice another creature: Target player mills three cards.
//	 Pay 1 life: Add {C}. Spend this mana only to cast a spell from your
//	 graveyard."
//
// The mill is a plain activated ability whose sacrifice is a cost. The
// mana ability costs life, not a tap, and its mana carries the
// graveyard-cast tag (ManaRestrictCastFromGraveyard, the twin of Web
// Watch's exile tag): it pays for a cast whose zone is a graveyard and
// for nothing else.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3a25d1e2-63d8-4a93-b0ce-f9307736b3f9",
		Name:            "Lord of the Forsaken",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		ManaAbilities: []effects.ManaAbility{{
			Cost:         effects.ManaAbilityCost{Life: 1},
			Produced:     "{C}",
			Label:        "Pay 1 life: Add {C}. Spend this mana only to cast a spell from your graveyard",
			Restrictions: []string{effects.ManaRestrictCastFromGraveyard},
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}, Sacrifice another creature: Target player mills three cards.",
			Cost:    effects.Plus(effects.ManaCost("{B}"), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			Targets: effects.TargetPlayer("target player"),
			Effect:  effects.TargetPlayerMills(3),
		}},
	})
}
