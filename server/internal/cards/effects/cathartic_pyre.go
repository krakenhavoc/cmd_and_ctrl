package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cathartic Pyre — Instant {1}{R} (EDHREC rank 4112):
//
//	"Choose one —
//	 • Cathartic Pyre deals 3 damage to target creature or
//	   planeswalker.
//	 • Discard up to two cards, then draw that many cards."
//
// Two mana that is never dead. Early it is a Shock-and-a-half on the
// mana dork or the walker that is about to ultimate; late, with a
// hand full of the wrong half of the deck, it is a free rummage at
// instant speed. A reanimator deck plays it for the second mode and
// is happy that the first one exists.
//
// THE FIRST MODE CANNOT HIT A PLAYER, and that is the whole
// difference between this and a burn spell: "target creature or
// planeswalker" is two permanent types and nothing else. The target
// clause says exactly that, so a player is never offered.
//
// THE SECOND MODE IS A GENUINE "UP TO", including zero. Discarding
// nothing and drawing nothing is a legal way to resolve the spell
// (it is what you do when the mode was chosen and the board changed),
// so no set-level check is imposed on the pick — unlike Thrilling
// Discovery, which prints a fixed two and is all-or-nothing. "Then
// draw that many" means the count actually discarded, so pitching one
// draws one.
//
// The discard happens BEFORE the draw, which is observable and is
// what makes the mode good: the cards you pitch are chosen from a
// hand that has not yet been refilled, and a discard payoff (a
// Marauding Mako, an Archfiend of Ifnir) triggers before the new
// cards arrive. The count is measured off the hand — see
// b39MayDiscardThenDraw for why that is the reliable read.
//
// A first mode whose target became illegal in response is countered
// by the rules (CR 608.2b) and deals nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f7e55117-69fb-412e-b16e-d92e0f95f629",
		Name:         "Cathartic Pyre",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Cathartic Pyre deals 3 damage to target creature or planeswalker.",
				TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker()))),
			Mode("Discard up to two cards, then draw that many cards."),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if ctx.HasMode(0) {
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 3}.Apply(ctx)
				}
				return nil
			}
			return b39MayDiscardThenDraw(2, false,
				"Cathartic Pyre — discard up to two cards, then draw that many",
				func(discarded int) int { return discarded })(ctx)
		},
	})
}
