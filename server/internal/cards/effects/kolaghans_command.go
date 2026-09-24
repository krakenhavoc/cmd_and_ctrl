package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kolaghan's Command — Instant {1}{B}{R} (EDHREC rank 1288):
//
//	"Choose two —
//	 • Return target creature card from your graveyard to your hand.
//	 • Target player discards a card.
//	 • Destroy target artifact.
//	 • Kolaghan's Command deals 2 damage to any target."
//
// The card ADR 0019 §8 named as the reason per-mode target slots had
// to exist, and the first "choose two" whose bullets each target.
// Until #764 it could not be declared at all: effects.Register
// panicked at boot on a Max > 1 spec with more than one targeted
// option, and the cast path refused an announcement that derived two
// clauses.
//
// Now each chosen bullet is its own mode OCCURRENCE with its own
// target group (TargetRef.Mode), the picker walks them in the order
// they were chosen, and resolution runs the chosen bullets in that
// same order (CR 608.2c). Choosing "destroy target artifact" and
// "deals 2 damage to any target" announces two targets, and each is
// re-checked against ITS OWN clause at resolution (CR 608.2b) — an
// artifact that stopped being one is skipped while the damage still
// happens.
//
// The discard is the printed one: the targeted player chooses which
// card (CR 701.8a), through the ordinary discard prompt.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "45f1e957-09f0-4d46-8e32-238f26060a87",
		Name:         "Kolaghan's Command",
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose two", 2, 2,
			ModeDoing("Return target creature card from your graveyard to your hand.",
				TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature()),
				func(item *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}.Apply(ctx)
				}),
			ModeDoing("Target player discards a card.",
				TargetPlayer("target player"),
				func(item *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
						Player:   t.ID,
						Source:   item.SourceCardID,
						N:        1,
						Question: "Kolaghan's Command — discard a card",
					})
					return nil
				}),
			ModeDoing("Destroy target artifact.",
				TargetPermanent("target artifact", Artifact()),
				DestroyTheModesTarget),
			ModeDoing("Kolaghan's Command deals 2 damage to any target.",
				TargetAny(),
				DealFixedDamageToModesTarget(2)),
		),
	})
}
