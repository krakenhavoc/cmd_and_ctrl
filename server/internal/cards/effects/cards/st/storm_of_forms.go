package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Storm of Forms — Instant {3}{U} (Streets of New Capenna Commander,
// #5491):
//
//	"When you cast this spell, copy it for each kind of counter among
//	 permanents you control. You may choose new targets for the copies.
//	 Return target nonland permanent to its owner's hand."
//
// Storm's shape (a cast trigger that names the spell, so one countered in
// response is still copied from last-known information, CR 608.2h) with
// another count: the distinct kinds of counter on permanents the caster
// controls, read as the trigger resolves. Two +1/+1 counters are one kind,
// a +1/+1 and a -1/-1 counter are two. Each copy may take a new target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e9a08492-3d2c-4c7a-8198-ef4193f5758a",
		Name:         "Storm of Forms",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard && effects.OnBattlefield(ctx.Game, t.ID) {
					return effects.BounceToHand{Target: t.ID}.Apply(ctx)
				}
			}
			return nil
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenYouCastThisSpell("Storm of Forms — copy this spell for each kind of counter among permanents you control",
				func(g *game.Game, item *game.StackItem) error {
					n := len(effects.S82SncCounterKindsAmongYourPermanents(g, item.Controller))
					if n == 0 {
						return nil
					}
					return effects.CopySpell{StackID: item.SourceCardID, Controller: item.Controller, Count: n,
						ChooseNewTargets: true, FromLastKnown: true}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
