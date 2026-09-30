package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flame of Anor — Instant {1}{U}{R} (EDHREC rank 1741):
//
//	"Choose one. If you control a Wizard as you cast this spell, you
//	 may choose two instead.
//	 • Target player draws two cards.
//	 • Destroy target artifact.
//	 • Flame of Anor deals 5 damage to target creature."
//
// The Wizard deck's charm. Three targeted options on a "choose one",
// each one primitive — the draw's target is a player, the other two
// are permanents — resolved in printed order.
//
// "If you control a Wizard as you cast this spell, you may choose two
// instead" is the conditional mode count #1590 built:
// `OrUpToIf(2, YouControlAWizard)`, read at announce (CR 601.2b) and
// fixed from then on. It reads effective subtypes, so a changeling
// counts. Each bullet has its own target group (#764) —
// the old body walked every target against every chosen mode, which
// with two chosen would have destroyed the creature meant for the
// 5 damage — and the bullets run in PRINTED order (CR 608.2c).
func init() {
	Register(Spec{
		OracleID:     "ecd49d85-9c8c-4cc9-9a83-072ecd433677",
		Name:         "Flame of Anor",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Target player draws two cards.", TargetPlayer("target player")),
			Mode("Destroy target artifact.", TargetPermanent("target artifact", Artifact())),
			Mode("Flame of Anor deals 5 damage to target creature.", TargetCreature("target creature")),
		).OrUpToIf(2, YouControlAWizard),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range OptionTargets(ctx, 0) {
				if err := (DrawCards{Player: t.ID, N: 2}).Apply(ctx); err != nil {
					return err
				}
			}
			for _, t := range OptionTargets(ctx, 1) {
				if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			for _, t := range OptionTargets(ctx, 2) {
				if err := (DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 5}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
