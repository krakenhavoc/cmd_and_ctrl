package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wear Down — Sorcery {1}{G} (EDHREC rank 2403):
//
//	"Gift a card (You may promise an opponent a gift as you cast this
//	 spell. If you do, they draw a card before its other effects.)
//	 Destroy target artifact or enchantment. If the gift was
//	 promised, instead destroy two target artifacts and/or
//	 enchantments."
//
// A Naturalize that can become a two-for-one.
//
// Gift (CR 702.174, ADR 0089): the promise swaps the clause for its
// two-target twin (`.Instead(… .WithCount(2, 2))`, CR 702.174m), so
// the resolution simply destroys every target that is still legal —
// one for the unpromised cast, two for the promised one, and one if
// the other left in response (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "27905301-333e-4cdd-90cf-188159fcf8e9",
		Name:         "Wear Down",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		Gift: effects.GiftACard().Instead(
			effects.TargetPermanent("two target artifacts and/or enchantments", effects.Or(effects.Artifact(), effects.Enchantment())).WithCount(2, 2)),
		OnResolve: effects.DestroyEachLegalTarget,
	})
}
