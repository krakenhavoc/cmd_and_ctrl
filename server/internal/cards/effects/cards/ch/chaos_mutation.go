package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chaos Mutation — Instant {3}{U}{R} (Warhammer 40,000 Commander, #5948):
//
//	"Exile any number of target creatures controlled by different players.
//	 For each creature exiled this way, its controller reveals cards from
//	 the top of their library until they reveal a creature card, puts that
//	 card onto the battlefield, then puts the rest on the bottom of their
//	 library in a random order."
//
// Divergent Transformations' shape with a set rule that no two picks share
// a controller (EachDifferentController, re-judged as the spell resolves).
// Every pick still legal is exiled first and its controller read as it was
// (CR 608.2h); then each of those players, in turn, reveals until a creature
// card, puts it onto the battlefield and puts the other revealed cards on the
// bottom in a random order. A pick that is gone (CR 608.2b) is left out.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1a240953-26e0-429f-9663-dcfe5bc2c1c7",
		Name:         "Chaos Mutation",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetCreature("any number of target creatures controlled by different players").
			WithCount(0, 0).EachDifferent(effects.EachDifferentController()),
		Purpose: game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var ids, controllers []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				c, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok {
					continue
				}
				ids = append(ids, t.ID)
				controllers = append(controllers, c.Controller)
			}
			for _, id := range ids {
				if err := (effects.ExileTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
			}
			for _, p := range controllers {
				run, hit := effects.RevealUntil(ctx, p, func(c game.Card) bool { return c.IsCreature() },
					"Chaos Mutation — reveal until a creature card")
				if err := (effects.PutFromLibraryOntoBattlefield{
					Player:   p,
					Cards:    run,
					Match:    func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return hit != uuid.Nil && c.InstanceID == hit },
					All:      true,
					ThenTail: effects.PutRestOnBottomInRandomOrderTail,
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
