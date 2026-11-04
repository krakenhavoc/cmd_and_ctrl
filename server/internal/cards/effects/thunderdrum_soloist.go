package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thunderdrum Soloist — Creature — Dwarf Bard {1}{R}, 1/3 (Secrets of
// Strixhaven, #4456):
//
//	"Reach
//	 Opus — Whenever you cast an instant or sorcery spell, this creature
//	 deals 1 damage to each opponent. If five or more mana was spent to cast
//	 that spell, this creature deals 3 damage to each opponent instead."
//
// Opus is an ability word (CR 207.2c): one trigger on your instant and sorcery
// casts, not a second one for the big case. The mana spent is read off the
// spell's payment record as it is cast and carried on the trigger (a waived
// payment asks what was paid: ReadsYourCasts), so the ability resolves with
// the count the spell was cast with even if the spell was countered or has
// left the stack. The damage is one instance dealt by this creature to each
// opponent (a creature that left the battlefield still deals it, from its
// last-known information, CR 608.2h), and it is 1 or, with five or more mana,
// 3 instead, never 4.
//
// No simplification.
func init() {
	const label = "Thunderdrum Soloist — Opus: 1 damage to each opponent, or 3 instead if five or more mana was spent"
	opus := On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return instantOrSorceryCastByYou(ev, source, g)
	}, label, func(g *game.Game, item *game.StackItem) error {
		n := 1
		if item.Params.Amount >= 5 {
			n = 3
		}
		return damageToEachOpponent(g, item, n)
	})
	opus.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Amount = g.StackItemPaidForEffect(ev.CardID).Spent().Total()
		return item
	}
	Register(Spec{
		OracleID:        "43b5794a-4ed6-471d-bc45-22fac486895d",
		Name:            "Thunderdrum Soloist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		ReadsManaSpent:  game.ReadsYourCasts,
		Triggered:       []game.TriggeredAbility{opus},
	})
}
