package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Divergent Transformations — Instant {6}{R} (Commander Masters, #5948):
//
//	"Undaunted (This spell costs {1} less to cast for each opponent.)
//	 Exile two target creatures. For each of those creatures, its controller
//	 reveals cards from the top of their library until they reveal a creature
//	 card, puts that card onto the battlefield, then shuffles the rest into
//	 their library."
//
// Both creatures are exiled first; the controller of each is read as it was
// before the exile (CR 608.2h). Each of those players reveals until a
// creature card, puts it onto the battlefield under their control, and
// shuffles the other revealed cards into their library. A target that is
// gone as the spell resolves (CR 608.2b) is left out, and its controller
// reveals nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f08e835a-7827-4fdc-8dcf-ae74507da052",
		Name:         "Divergent Transformations",
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(effects.OpponentsOfCaster(), "Undaunted"),
		},
		Targets: effects.TargetCreature("two target creatures").WithCount(2, 2),
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
					"Divergent Transformations — reveal until a creature card")
				if err := (effects.PutFromLibraryOntoBattlefield{
					Player:   p,
					Cards:    run,
					Match:    func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return hit != uuid.Nil && c.InstanceID == hit },
					All:      true,
					ThenTail: divergentTransformationsShuffleRest,
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}

// divergentTransformationsShuffleRest is "then shuffles the rest into their
// library": the revealed cards that were not put onto the battlefield are
// still in the library, so the shuffle mixes them back in. An on-disk key:
// never renamed, never reused.
var divergentTransformationsShuffleRest = effects.PutFromLibraryTail("put-from-library/divergent-transformations-shuffle",
	func(ctx *effects.Context, res effects.PutFromLibraryResult) error {
		return ctx.Game.ShuffleLibraryForEffect(res.Player)
	})
