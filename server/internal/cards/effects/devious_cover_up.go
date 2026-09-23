package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Devious Cover-Up — Instant {2}{U}{U}:
//
//	"Counter target spell. If that spell is countered this way, exile
//	 it instead of putting it into its owner's graveyard.
//	 You may shuffle up to four target cards from your graveyard into
//	 your library."
//
// Two separate "target" clauses in one statement — a mandatory spell
// and an independent up-to-four from the caster's own graveyard —
// which is the #764 multi-clause shape (Clauses / .WithCount), not a
// reflexive trigger: both are chosen at announce (CR 601.2c), same as
// every other target on the card, not picked later the way a "when
// you do" follow-up is.
//
// The exile clause is modelled with CounterTarget.Dest (#1230):
// counterSpellLocked already accepted a destination zone, and closing
// the *ForEffect wrapper gap that hard-coded the owner's graveyard was
// shared work with Reprieve and Narset's Reversal, which need the
// sibling "return to hand" verb off the same stack-exit primitive.
func init() {
	Register(Spec{
		OracleID:     "feb221fb-59bf-4671-a53f-1bbe8e9c2ca9",
		Name:         "Devious Cover-Up",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetSpell("target spell"),
			TargetCardInGraveyard("up to four target cards in your graveyard", YouOwn()).WithCount(0, 4),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
				if err := (CounterTarget{
					StackID: t.ID,
					Dest:    game.ZoneRef{Kind: game.ZoneExile},
				}).Apply(ctx); err != nil {
					return err
				}
			}
			var moved bool
			for _, t := range ctx.ClauseTargets(1) {
				if !ctx.IsTargetLegal(t) {
					continue
				}
				if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneLibrary}).Apply(ctx); err != nil {
					return err
				}
				moved = true
			}
			if moved {
				return ShuffleLibrary{Player: ctx.Controller()}.Apply(ctx)
			}
			return nil
		},
	})
}
