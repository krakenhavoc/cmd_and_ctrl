package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spectacular Skywhale — Creature — Elemental Whale {2}{U}{R}, 1/4 (Secrets
// of Strixhaven, #4456):
//
//	"Flying
//	 Opus — Whenever you cast an instant or sorcery spell, this creature gets
//	 +3/+0 until end of turn. If five or more mana was spent to cast that
//	 spell, put three +1/+1 counters on this creature instead."
//
// Opus is an ability word (CR 207.2c); the ability is one trigger on your
// instant and sorcery casts. The mana spent is counted as the ability
// triggers (ReadsYourCasts) and carried on the trigger, so the count the
// spell was cast with is what the "instead" reads even if that spell was
// countered. Exactly one of the two happens, never both. Both belong to
// this permanent and are skipped if it has left or come back as a new
// object (CR 400.7).
//
// No simplification.
func init() {
	const label = "Spectacular Skywhale — Opus: +3/+0 until end of turn, or three +1/+1 counters if five or more mana was spent"
	opus := On(game.EventCast, YouCast(Or(Instant(), Sorcery())), label,
		func(g *game.Game, item *game.StackItem) error {
			if !sourceIsStillThisPermanent(g, item) {
				return nil
			}
			ctx := NewContext(g, item)
			if item.Params.Amount >= 5 {
				return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 3}.Apply(ctx)
			}
			return BoostUntilEOT{
				Target: item.SourceCardID,
				Power:  3,
				Label:  "Spectacular Skywhale — +3/+0 until end of turn",
			}.Apply(ctx)
		})
	opus.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Amount = g.StackItemPaidForEffect(ev.CardID).Spent().Total()
		return item
	}
	Register(Spec{
		OracleID:        "6a6b19db-cdeb-4922-8642-5d872f90e7a1",
		Name:            "Spectacular Skywhale",
		Completeness:    CompletenessFull,
		ReadsManaSpent:  game.ReadsYourCasts,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{opus},
	})
}
