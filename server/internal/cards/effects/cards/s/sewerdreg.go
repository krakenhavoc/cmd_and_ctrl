package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sewerdreg — Creature — Spirit {3}{B}{B}, 3/3 (Ravnica: City of Guilds):
//
//	"Swampwalk
//	 Sacrifice this creature: Exile target card from a graveyard."
//
// Swampwalk is the shared landwalk keyword. The activated ability's target is
// any card in any graveyard, chosen as the ability is activated (CR 602.2b); the
// sacrifice is a cost (CR 118.3), so the card goes even if the target is
// gone, and the target is checked again as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f05e3fb4-a7e0-4dfa-9f3e-5109be79ef16",
		Name:            "Sewerdreg",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"swampwalk"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: Exile target card from a graveyard.",
			Purpose: game.Purpose{Answers: game.AnswerGraveyard},
			Cost:    effects.SacrificeThis(),
			Targets: effects.TargetCardInGraveyard("target card from a graveyard"),
			Effect:  effects.S82OldB504ExileTheFirstTarget,
		}},
	})
}
