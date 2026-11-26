package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Agent of the Fates — Creature — Human Assassin {1}{B}{B}, 3/2 (Theros, S82
// OLD b388):
//
//	"Deathtouch
//	 Heroic — Whenever you cast a spell that targets this creature, each
//	 opponent sacrifices a creature of their choice."
//
// Heroic (CR 207.2c) with the edict every opponent answers for themselves
// (EachOpponentSacrificesACreature): each chooses among their own creatures as it
// resolves, and an opponent with none sacrifices nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "45f88a6a-ac8a-4a53-a96b-f09ca33c79ba",
		Name:            "Agent of the Fates",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.Heroic("Agent of the Fates — each opponent sacrifices a creature of their choice (heroic)",
				effects.EachOpponentSacrificesACreature),
		},
	})
}
