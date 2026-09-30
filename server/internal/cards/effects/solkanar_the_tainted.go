package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// solKanarLabel is the trigger's stack label, and with it the key its
// "hasn't been chosen" memory is kept under.
const solKanarLabel = "Sol'Kanar the Tainted — beginning of your end step"

// Sol'Kanar the Tainted — Legendary Creature — Elemental Demon
// {2}{U}{B}{R}, 5/5:
//
//	"At the beginning of your end step, choose one that hasn't been
//	 chosen —
//	 • Draw a card.
//	 • Each opponent loses 2 life and you gain 2 life.
//	 • Sol'Kanar deals 3 damage to up to one other target creature or
//	   planeswalker.
//	 • Exile Sol'Kanar, then return it to the battlefield under an
//	   opponent's control."
//
// ChooseOneNotChosen (ADR 0097): each bullet once for this object. The
// fourth bullet hands Sol'Kanar to an opponent as a NEW object (CR
// 400.7), so its new controller starts with no memory and may choose
// all four — the Demon passes round the table.
//
// "Up to one other target" may name nothing, so the bullet is always on
// offer. "Other" is object identity (effects.Another, CR 109.1): a
// second creature named Sol'Kanar the Tainted is a legal target. The
// drain is life loss, not damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4efcdefc-e49d-4bce-8581-69037bb48c0a",
		Name:         "Sol'Kanar the Tainted",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			solKanarTrigger(),
		},
	})
}

func solKanarTrigger() game.TriggeredAbility {
	t := AtYourEndStep(solKanarLabel, func(_ *game.Game, _ *game.StackItem) error { return nil })
	t.Modes = ChooseOneNotChosen(
		ModeDoing("Draw a card.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
			}),
		ModeDoing("Each opponent loses 2 life and you gain 2 life.", nil,
			eachOpponentLosesTwoYouGainTwo),
		ModeDoing("Sol'Kanar deals 3 damage to up to one other target creature or planeswalker.",
			Another(TargetPermanent("up to one other target creature or planeswalker",
				Or(Creature(), Planeswalker()))).WithCount(0, 1),
			func(item *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 3}.Apply(ctx)
			}),
		ModeDoing("Exile Sol'Kanar, then return it to the battlefield under an opponent's control.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return exileThisThenReturnUnderAnOpponentsControl(ctx, item, "Sol'Kanar the Tainted")
			}),
	)
	return t
}
