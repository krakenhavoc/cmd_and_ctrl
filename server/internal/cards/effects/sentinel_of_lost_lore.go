package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sentinel of Lost Lore — Creature — Elf Knight {1}{G}{G}, 3/2 (Wilds of
// Eldraine, #4471):
//
//	"When this creature enters, choose one or more —
//	 • Return target card you own in exile that has an Adventure to your
//	   hand.
//	 • Put target card you don't own in exile that has an Adventure on the
//	   bottom of its owner's library.
//	 • Exile target player's graveyard."
//
// An enters trigger with three bullets, each with its own target clause,
// chosen as it goes on the stack and run in printed order (CR 608.2c). The
// first two read any face-up exiled adventurer card, whether or not it was
// cast as an Adventure (the ruling), split by ownership: yours goes to your
// hand, someone else's goes to the bottom of its owner's library. A bullet
// with no legal target is not offered, so with none the player must take the
// graveyard bullet, and each target is re-judged as it resolves
// (CR 608.2b). The third is Bojuka Bog's graveyard exile.
//
// No simplification.
func init() {
	t := WhenThisEnters("Sentinel of Lost Lore — choose one or more", func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = ChooseOneOrMore(
		ModeDoing("Return target card you own in exile that has an Adventure to your hand.",
			TargetCardInExile("target card you own in exile that has an Adventure", FaceUp(), YouOwn(), HasAnAdventure()),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				id, ok := ModeTarget(ctx, occ)
				if !ok || !ctx.IsTargetLegal(id) {
					return nil
				}
				return (BounceToHand{Target: id.ID}).Apply(ctx)
			}),
		ModeDoing("Put target card you don't own in exile that has an Adventure on the bottom of its owner's library.",
			TargetCardInExile("target card you don't own in exile that has an Adventure", FaceUp(), Not(YouOwn()), HasAnAdventure()),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				id, ok := ModeTarget(ctx, occ)
				if !ok || !ctx.IsTargetLegal(id) {
					return nil
				}
				return ctx.Game.TuckToLibraryForEffect(id.ID, true)
			}),
		ModeDoing("Exile target player's graveyard.", TargetPlayer("target player"),
			func(item *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetPlayer {
					return nil
				}
				return exileGraveyardForEffect(ctx.Game, item, t.ID)
			}),
	)
	Register(Spec{
		OracleID:     "683775ac-481e-4c26-ab2f-5d2813af0268",
		Name:         "Sentinel of Lost Lore",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
