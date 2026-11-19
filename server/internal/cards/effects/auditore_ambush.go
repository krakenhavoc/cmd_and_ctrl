package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Auditore Ambush — Sorcery {2}{U}{B}:
//
//	"Choose one or both —
//	 • Return target creature to its owner's hand.
//	 • Target player searches their library and/or graveyard for a card
//	   named Ezio, Blade of Vengeance, reveals it, and puts it into their
//	   hand. If they search their library this way, they shuffle."
//
// The first mode is a plain bounce. The second is Ajani's Aid's search by
// name over the library and the graveyard, done by the TARGET player, not
// by the caster, and not optional ("searches", not "may search"): the
// engine asks which zones to search before anything is shown, and shows
// and shuffles the library only when it was searched (#4516, #4519).
// Both modes run in printed order when both are chosen.
func init() {
	const named = "Ezio, Blade of Vengeance"
	Register(Spec{
		OracleID:     "72075231-e3b3-46cc-b994-2d95b24cf904",
		Name:         "Auditore Ambush",
		Completeness: CompletenessFull,
		Modes: ChooseOneOrMore(
			ModeDoing("Return target creature to its owner's hand.",
				TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					for _, t := range ctx.ModeTargets(occ) {
						if t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
							return BounceToHand{Target: t.ID}.Apply(ctx)
						}
					}
					return nil
				}),
			ModeDoing("Target player searches their library and/or graveyard for a card named Ezio, Blade of Vengeance, reveals it, and puts it into their hand. If they search their library this way, they shuffle.",
				TargetPlayer("target player"),
				func(item *game.StackItem, ctx *Context, occ int) error {
					var player uuid.UUID
					for _, t := range ctx.ModeTargets(occ) {
						if t.Kind == game.TargetPlayer && ctx.IsTargetLegal(t) {
							player = t.ID
							break
						}
					}
					if player == uuid.Nil {
						return nil
					}
					return SearchLibrary{
						Player:        player,
						Source:        item.SourceCardID,
						Predicate:     func(c game.Card) bool { return c.Name == named },
						Dest:          game.ZoneHand,
						Limit:         1,
						Reveal:        true,
						Shuffle:       true,
						AlsoGraveyard: true,
						Reason:        "Auditore Ambush — choose a card named " + named + " to put into your hand",
					}.Apply(ctx)
				}),
		),
	})
}
