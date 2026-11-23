package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bofur, Reliable Guardian // Concerted Care — an Adventure card (CR 715,
// The Hobbit, #4454), oracle 8a0e35ac…:
//
//	Bofur, Reliable Guardian — Legendary Creature — Dwarf Scout {W}, 1/1
//	  "Lifelink"
//	Concerted Care — Instant — Adventure {1}{W}
//	  "Target artifact or creature you control gains hexproof and
//	   indestructible until end of turn. (Then exile this card. You may
//	   cast the creature later from exile.)"
//
// No simplification.

// bofurReliableGuardianOracleID is shared by both faces of the card.
const BofurReliableGuardianOracleID = "8a0e35ac-6c03-4922-b3b4-e419419fe3d7"

func init() {
	effects.Register(effects.Spec{
		OracleID:        BofurReliableGuardianOracleID,
		Name:            "Bofur, Reliable Guardian",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
	})
	effects.Register(effects.Spec{
		OracleID:     BofurReliableGuardianOracleID + "#1",
		Name:         "Concerted Care",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target artifact or creature you control", effects.Or(effects.Artifact(), effects.Creature()), effects.YouControl()),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "hexproof", "indestructible")),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				return effects.GrantKeywordUntilEOT{
					Target:   t.ID,
					Keywords: []string{"hexproof", "indestructible"},
					Label:    "Concerted Care — hexproof and indestructible",
				}.Apply(ctx)
			}
			return nil
		},
	})
}
