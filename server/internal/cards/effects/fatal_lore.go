package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fatal Lore — Sorcery {2}{B}{B}:
//
//	"An opponent chooses one —
//	 • You draw three cards.
//	 • You destroy up to two target creatures that player controls. They
//	   can't be regenerated. That player draws up to three cards."
//
// The opponent chooses the bullet as the spell is cast (CR 700.2e,
// OpponentChoosesOne), and the caster picks which opponent when there is
// more than one. The 2004 ruling: "The opponent chooses which option they
// want (this is a mode choice). If they choose the 'destroy' option, you
// choose the zero, one, or two targets after that." So the caster is
// asked for the targets once the bullet is known, among the creatures
// that opponent controls (ThatPlayerControls). "That player draws up to
// three cards" is that opponent's choice of how many, 0 to 3 (CR 121.2b
// caps it at the draws a draw limit leaves them). If both targets have
// become illegal the spell does not resolve (CR 608.2b), and the draw
// with it; with no target chosen it resolves.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "aeb85786-16fc-45d8-9383-60054a878c40",
		Name:         "Fatal Lore",
		Completeness: CompletenessFull,
		Modes: OpponentChoosesOne(
			ModeWithPurpose(Mode("You draw three cards."), game.Purpose{Draws: 3}),
			Mode("You destroy up to two target creatures that player controls. They can't be regenerated. That player draws up to three cards.",
				ThatPlayerControls(TargetCreature("up to two target creatures that player controls").WithCount(0, 2))),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if ctx.HasMode(0) {
				return DrawCards{Player: ctx.Controller(), N: 3}.Apply(ctx)
			}
			if !ctx.HasMode(1) {
				return nil
			}
			for _, t := range OptionTargets(ctx, 1) {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (DestroyTarget{Target: t.ID, CantBeRegenerated: true}).Apply(ctx); err != nil {
					return err
				}
			}
			that := ctx.ModeChooser()
			if ctx.Game.PlayerByIDForEffect(that) == nil {
				return nil
			}
			return ChooseNumber{
				Chooser:  that,
				Min:      0,
				Max:      ctx.Game.DrawChoiceCeilingForEffect(that, 3),
				Question: "Fatal Lore — draw up to three cards",
				Unit:     game.PayAmountCards,
				Goal: func(ctx *Context) int {
					if p := ctx.Game.PlayerByIDForEffect(that); p != nil && p.Library != nil && p.Library.Size() < 3 {
						return p.Library.Size()
					}
					return 3
				},
				Then: fatalLoreDraws,
			}.Apply(ctx)
		},
	})
}

// fatalLoreDraws is the opponent's number's branch: they draw that many.
// The chooser of the number is "that player".
var fatalLoreDraws = NumberThen("number/fatal-lore-draw-up-to-three",
	func(ctx *Context, r game.NumberAnswer) error {
		if r.Number <= 0 || ctx.Game.PlayerByIDForEffect(r.Chooser) == nil {
			return nil
		}
		return DrawCards{Player: r.Chooser, N: r.Number}.Apply(ctx)
	})
