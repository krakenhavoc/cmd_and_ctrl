package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Repair and Recharge — Sorcery {3}{W}{W} (The Brothers' War, #5482):
//
//	"Return target artifact, enchantment, or planeswalker card from your
//	 graveyard to the battlefield. Create a tapped Powerstone token."
//
// The one-target reanimation (Agadeem's Awakening's return), then the
// Powerstone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4f42517b-e62e-4b8c-8c56-cb32fb62c478",
		Name:         "Repair and Recharge",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetCardInGraveyard("target artifact, enchantment, or planeswalker card from your graveyard",
			effects.YouOwn(), effects.Or(effects.Artifact(), effects.Enchantment(), effects.Planeswalker())),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(effects.ReturnLegalGraveyardTargetsToBattlefield),
			effects.CreateTokenAdvanced{Spec: effects.Token(effects.PowerstoneToken()).EntersTapped(), N: 1},
		),
	})
}
