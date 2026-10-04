package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gruesome Discovery — Sorcery {2}{B}{B}:
//
//	"Target player discards two cards.
//	 Morbid — If a creature died this turn, instead that player reveals
//	 their hand, you choose two cards from it, then that player
//	 discards those cards."
//
// Morbid is read as the spell resolves, off the same turn tally
// Deathreap Ritual reads (every creature that went to a graveyard from
// the battlefield this turn, tokens included). Without it the player
// chooses their own two cards (CR 701.9b). With it the hand is revealed
// to the table and the caster chooses two cards of any kind through the
// revealed-hand pick (ADR 0116); a hand of one card gives up that one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ecdae60b-c594-4e70-909b-83483104a42c",
		Name:         "Gruesome Discovery",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			player := TargetedPlayer(ctx)
			if player == uuid.Nil {
				return nil
			}
			if b11CreaturesDiedThisTurn(ctx.Game) > 0 {
				return ChooseFromRevealedHand{Player: player, Count: 2}.Apply(ctx)
			}
			ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
				Player: player,
				Source: item.SourceCardID,
				N:      2,
			})
			return nil
		},
	})
}
