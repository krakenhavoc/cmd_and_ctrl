package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Warlord's Fury — Sorcery {R} (Dominaria, #5530):
//
//	"Creatures you control gain first strike until end of turn.
//	 Draw a card."
//
// Two steps in printed order: first strike for the creatures you control
// as it resolves, then a card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7340199d-60c7-4c27-9a0f-047790273db0",
		Name:         "Warlord's Fury",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.GrantKeywordUntilEOT{
				Match:    effects.And(effects.Creature(), effects.YouControl()),
				Keywords: []string{"first strike"},
				Label:    "Warlord's Fury — first strike",
			},
			effects.DrawCards{N: 1},
		),
	})
}
