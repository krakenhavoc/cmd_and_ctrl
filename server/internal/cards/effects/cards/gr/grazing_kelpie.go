package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grazing Kelpie — Creature — Beast {3}{G/U}, 2/3:
//
//	"{G/U}, Sacrifice this creature: Put target card from a graveyard on
//	 the bottom of its owner's library.
//	 Persist"
//
// Persist is PrintedKeywords (#2075).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a2db1035-008e-4de3-b5e5-2a0de83082b2",
		Name:            "Grazing Kelpie",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordPersist},
		Activated: []effects.ActivatedAbility{{
			Label:   "{G/U}, Sacrifice this creature: Put target card from a graveyard on the bottom of its owner's library.",
			Cost:    effects.Plus(effects.ManaCost("{G/U}"), effects.SacrificeThis()),
			Targets: effects.TargetCardInGraveyard("target card from a graveyard"),
			Effect:  effects.SocB37PutTargetGraveyardCardOnBottom,
		}},
	})
}
