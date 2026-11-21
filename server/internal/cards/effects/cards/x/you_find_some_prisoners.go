package cardsx

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// You Find Some Prisoners — Instant {1}{R} (Adventures in the Forgotten
// Realms, #5499):
//
//	"Choose one —
//	 • Break Their Chains — Destroy target artifact.
//	 • Interrogate Them — Exile the top three cards of target opponent's
//	   library. Choose one of them. Until the end of your next turn, you may
//	   play that card, and you may spend mana as though it were mana of any
//	   color to cast it."
//
// Abrade's destroy bullet. The second bullet exiles all three face up
// (Mishra's Research Desk's exile-and-choose, from the opponent's library),
// the caster chooses one, and only that card gets the window, which also
// lets mana of any color pay for it (CR 609.4). Interrogate targets a player.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "467c526a-02a7-405d-98de-c632ea83519b",
		Name:         "You Find Some Prisoners",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Break Their Chains — Destroy target artifact.",
				effects.TargetPermanent("target artifact", effects.Artifact()),
				effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error {
					return effects.DestroyTarget{Target: id}.Apply(ctx)
				})), game.Purpose{Answers: game.AnswerRemove}),
			effects.ModeDoing("Interrogate Them — Exile the top three cards of target opponent's library. Choose one of them. Until the end of your next turn, you may play that card, and you may spend mana as though it were mana of any color to cast it.",
				effects.TargetPlayer("target opponent", effects.Opponent()),
				effects.AfrB04Interrogate),
		),
	})
}
