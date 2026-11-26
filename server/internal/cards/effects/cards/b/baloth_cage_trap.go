package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Baloth Cage Trap — Instant — Trap {3}{G}{G} (Zendikar, #5583):
//
//	"If an opponent had an artifact enter the battlefield under their control
//	 this turn, you may pay {1}{G} rather than pay this spell's mana cost.
//	 Create a 4/4 green Beast creature token."
//
// The trap cost is offered when any opponent had an artifact permanent enter
// under their control this turn, read off the permanents as they entered
// (game.EnteredWithCardTypeThisTurn); a token counts, since it enters the
// battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1221bef9-5947-4c32-a74b-b60279bb79bf",
		Name:         "Baloth Cage Trap",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		AlternativeCosts: []game.AlternativeCost{{
			Key:       "trap",
			Label:     "If an opponent had an artifact enter the battlefield under their control this turn, pay {1}{G}",
			ManaCost:  "{1}{G}",
			Condition: OldB497AnOpponentHadAnArtifactEnter,
		}},
		OnResolve: effects.ResolveSteps(effects.CreateToken{Template: effects.TokenCard("4/4 green Beast"), N: 1}),
	})
}

// OldB497AnOpponentHadAnArtifactEnter is Baloth Cage Trap's "if an opponent
// had an artifact enter the battlefield under their control this turn".
func OldB497AnOpponentHadAnArtifactEnter(g *game.Game, controller uuid.UUID) bool {
	for _, p := range g.Seats {
		if p == nil || p.ID == controller {
			continue
		}
		if g.EnteredWithCardTypeThisTurn(p.ID, "artifact") > 0 {
			return true
		}
	}
	return false
}
