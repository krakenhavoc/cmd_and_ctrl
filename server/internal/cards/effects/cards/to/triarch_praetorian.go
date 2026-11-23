package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Triarch Praetorian — Artifact Creature — Necron {1}{B}, 2/1 (Warhammer 40,000
// Commander, #5948):
//
//	"Flying
//	 Dynastic Codes — When this creature enters from a graveyard, you draw two
//	 cards and you lose 2 life.
//	 Unearth {4}{B}"
//
// Unearth is the shared keyword; it returns the card from the graveyard, so the
// entry trigger sees it come from a graveyard (CR 603.6a). Cast from hand it
// enters from the stack and draws nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "52aa5a79-db0e-4eaa-b2df-e41674c47794",
		Name:            "Triarch Praetorian",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, effects.EnteredFromAGraveyard), "Triarch Praetorian — you draw two cards and you lose 2 life",
				effects.Do(effects.DrawCards{N: 2}, effects.GainLife{Amount: -2})),
		},
		Activated: []effects.ActivatedAbility{effects.Unearth("{4}{B}")},
	})
}
