package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cathartic Parting — Sorcery {1}{G} (Duskmourn: House of Horror, #4466):
//
//	"The owner of target artifact or enchantment an opponent controls
//	 shuffles it into their library. You may shuffle up to four target cards
//	 from your graveyard into your library."
//
// Two target clauses in printed order, each its own instruction (CR 608.2c,
// the second waits for any prompt the first asks). The first is Watery Grasp's
// tuck: the permanent goes into its OWNER's library and that owner shuffles,
// whoever controls it (a commander's owner is offered the command zone first,
// CR 903.9). The second is "you may shuffle up to four target cards from your
// graveyard into your library": the "may" is the shuffle, so it is asked as
// the spell resolves, and a "yes" puts each of the targets still in your
// graveyard into your library and then shuffles it. With no card targeted you
// may still choose to shuffle your library (the ruling), and a "no" moves
// nothing. A target that has become illegal is left alone (CR 608.2b); if the
// artifact or enchantment is illegal the graveyard half still happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0ea80bd9-ab7e-4a8f-85b0-167fe7fae43a",
		Name:         "Cathartic Parting",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: Clauses(
			TargetPermanent("target artifact or enchantment an opponent controls", Or(Artifact(), Enchantment()), OpponentControls()),
			TargetCardInGraveyard("up to four target cards from your graveyard", YouOwn()).WithCount(0, 4),
		),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				c, found := ctx.Game.LookupCardForEffect(t.ID)
				if !found {
					return nil
				}
				return ctx.Game.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{t.ID}, game.TuckOptions{},
					game.KeyedTail{Ref: thenShuffle, Player: c.Owner})
			}),
			Instruction(func(ctx *Context) error {
				var ids []uuid.UUID
				for _, t := range ctx.ClauseTargets(1) {
					if t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
						ids = append(ids, t.ID)
					}
				}
				return MayChoice{
					Question: "Cathartic Parting — shuffle the targeted cards from your graveyard into your library?",
					Then:     catharticPartingShuffle,
					Carry:    ids,
				}.Apply(ctx)
			}),
		),
	})
}

// catharticPartingShuffle is the "yes": each carried card still in the
// controller's graveyard goes into their library, then it is shuffled. An
// on-disk key: never renamed, never reused.
var catharticPartingShuffle = MayYes("confirm/cathartic-parting-shuffle", func(ctx *Context, r game.ConfirmAnswer) error {
	for _, id := range r.Carry {
		z := ctx.Game.FindCardZoneForEffect(id)
		if z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneLibrary}).Apply(ctx); err != nil {
			return err
		}
	}
	return ShuffleLibrary{Player: ctx.Controller()}.Apply(ctx)
})
