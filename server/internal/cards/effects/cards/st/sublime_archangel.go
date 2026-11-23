package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sublime Archangel — Creature — Angel {2}{W}{W}, 4/3 (Magic 2013, #5568):
//
//	"Flying
//	 Exalted (Whenever a creature you control attacks alone, that
//	 creature gets +1/+1 until end of turn.)
//	 Other creatures you control have exalted. (If a creature has multiple
//	 instances of exalted, each triggers separately.)"
//
// Exalted is cumulative (CR 702.83b), so the grant adds one instance to
// each other creature you control and each triggers on its own: a lone
// attacker gets +1/+1 for the Archangel and for every creature you
// control, itself included.
//
// No simplification.
func init() {
	grant := effects.KeywordGrant(effects.OtherCreaturesYouControl, game.KeywordExalted)
	grant.Label = "Other creatures you control have exalted."
	effects.Register(effects.Spec{
		OracleID:        "b5479c97-df19-49c7-b305-7a0336fd1cb3",
		Name:            "Sublime Archangel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", game.KeywordExalted},
		Static:          []game.StaticAbility{grant},
	})
}
