package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Exhibition Tidecaller — Creature — Djinn Wizard {U}, 0/2 (Secrets of
// Strixhaven, #4456):
//
//	"Opus — Whenever you cast an instant or sorcery spell, target player
//	 mills three cards. If five or more mana was spent to cast that spell,
//	 that player mills ten cards instead."
//
// Colorstorm Stallion's Opus with a player target: a cast trigger (a copy is
// not cast, CR 707.10) whose mana spent is counted as the spell is cast and
// carried on the item (the ruling: the ability resolves before the spell,
// and even if the spell is countered the count it was cast with is what the
// "if" reads). The target is any player, the controller included. "Instead"
// is one mill of three or of ten, never both; a library holding fewer mills
// what it has and nobody loses for it (CR 701.17b).
//
// No simplification.
func init() {
	const label = "Exhibition Tidecaller — Opus: target player mills three cards; with five or more mana spent, ten instead"
	opus := Targeting(
		On(game.EventCast, YouCast(Or(Instant(), Sorcery())), label,
			func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				t, ok := firstLegalTarget(ctx)
				if !ok || t.Kind != game.TargetPlayer {
					return nil
				}
				n := 3
				if item.Params.Amount >= 5 {
					n = 10
				}
				return MillCards{Player: t.ID, N: n}.Apply(ctx)
			}),
		TargetPlayer("target player"),
	)
	opus.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Amount = g.StackItemPaidForEffect(ev.CardID).Spent().Total()
		return item
	}
	Register(Spec{
		OracleID:       "d32c9439-4c97-412b-83cc-6d9acb48be3c",
		Name:           "Exhibition Tidecaller",
		ReadsManaSpent: game.ReadsYourCasts,
		Completeness:   CompletenessFull,
		Triggered:      []game.TriggeredAbility{opus},
	})
}
