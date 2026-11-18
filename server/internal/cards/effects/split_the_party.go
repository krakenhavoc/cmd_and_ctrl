package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Split the Party — Sorcery {3}{U}{U} (Adventures in the Forgotten
// Realms):
//
//	"Choose target player. Return half the creatures they control to
//	 their owner's hand, rounded up."
//
// The number is read as the spell resolves: half of the creatures the
// target player controls then, rounded up (one creature is one, three are
// two). The spell makes the choice of which, since nothing in the text
// gives it to the target player (CR 608.2d); each returns to its owner's
// hand. A target player who is no longer legal makes it do nothing (CR
// 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9ad61841-9894-41da-bdd8-248c2ce21b00",
		Name:         "Split the Party",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			var player uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetPlayer {
					player = t.ID
				}
			}
			if player == uuid.Nil {
				return nil
			}
			var creatures []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller == player && c.IsCreature() {
					creatures = append(creatures, c.InstanceID)
				}
			}
			n := (len(creatures) + 1) / 2
			if n == 0 {
				return nil
			}
			ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:  ctx.Controller(),
				Source:   ctx.Source(),
				Question: "Split the Party — choose half the creatures they control, rounded up, to return to their owners' hands",
				Cards:    creatures,
				Min:      n,
				Max:      n,
				LiveZone: game.ZoneBattlefield,
				ThenKey:  splitThePartyPick,
				Asked:    game.NewAskedItem(item),
			})
			return nil
		},
	})
}

// splitThePartyPick returns the chosen creatures. An on-disk key.
var splitThePartyPick = askedPick("choose-cards/split-the-party-return", func(ctx *Context, r game.CardsChosen) error {
	for _, id := range r.Picked {
		if err := (BounceToHand{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
