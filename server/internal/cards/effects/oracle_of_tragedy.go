package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Oracle of Tragedy — Creature — Human Wizard {1}{U}, 1/3 (March of the
// Machine, S82 MOM batch 13):
//
//	"When this creature enters or dies, choose one —
//	 • Draw a card, then discard a card.
//	 • Shuffle up to four target cards with mana value 3 or greater from
//	   your graveyard into your library."
//
// One modal trigger with two conditions (Mogg War Marshal's "enters or
// dies"): the mode is chosen as the trigger goes on the stack and only the
// shuffle bullet takes targets (up to four, so none is allowed). The loot
// is draw then discard, the discard being the controller's choice. The
// shuffle moves each target still in the graveyard (CR 608.2b) to the
// library and then shuffles it (Cathartic Parting's shape).
//
// No simplification.
func init() {
	t := WhenThisEntersOrDies("Oracle of Tragedy — choose one", func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = ChooseOne(
		ModeDoing("Draw a card, then discard a card.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return lootOne(ctx.Game, item, 1)
			}),
		ModeDoing("Shuffle up to four target cards with mana value 3 or greater from your graveyard into your library.",
			TargetCardInGraveyard("up to four target cards with mana value 3 or greater from your graveyard", YouOwn(), ManaValueGE(3)).WithCount(0, 4),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				var ids []uuid.UUID
				for _, tr := range ctx.ModeTargets(occ) {
					if tr.Kind == game.TargetCard && ctx.IsTargetLegal(tr) {
						ids = append(ids, tr.ID)
					}
				}
				for _, id := range ids {
					z := ctx.Game.FindCardZoneForEffect(id)
					if z == nil || z.Kind != game.ZoneGraveyard {
						continue
					}
					if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneLibrary}).Apply(ctx); err != nil {
						return err
					}
				}
				return ShuffleLibrary{Player: ctx.Controller()}.Apply(ctx)
			}),
	)
	Register(Spec{
		OracleID:     "f88e1616-a7c7-4122-b626-bfa799cccaa4",
		Name:         "Oracle of Tragedy",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
