package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sazh Katzroy — Legendary Creature — Human Pilot {3}{G}, 3/3 (Final
// Fantasy, #4462):
//
//	"When Sazh Katzroy enters, you may search your library for a Bird or
//	 basic land card, reveal it, put it into your hand, then shuffle.
//	 Whenever Sazh Katzroy attacks, put a +1/+1 counter on target
//	 creature, then double the number of +1/+1 counters on that
//	 creature."
//
// The search is an optional prompt (Demolition Field's), revealed, to
// the hand; "Bird" is read through the card's text (CR 612.2) and
// matches any card with that subtype. The attack trigger is Sage of the
// Fang's body: one counter, then as many more as the creature has once
// that counter lands, so a creature with none ends on two.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0e5b40dc-debf-48a1-8398-5ea00b5a5166",
		Name:         "Sazh Katzroy",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sazh Katzroy — you may search for a Bird or basic land card",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					bird := ctx.TextSubtype("Bird")
					return SearchLibrary{
						Player: ctx.Controller(),
						Predicate: func(c game.Card) bool {
							return c.HasSubtype(bird) || IsBasicLand(c)
						},
						Dest:     game.ZoneHand,
						Limit:    1,
						Reveal:   true,
						Shuffle:  true,
						Optional: true,
						Reason:   "Sazh Katzroy — you may search for a Bird or basic land card",
						Source:   item.SourceCardID,
					}.Apply(ctx)
				}),
			Targeting(WheneverThisAttacks("Sazh Katzroy — a +1/+1 counter on target creature, then double its +1/+1 counters",
				func(g *game.Game, item *game.StackItem) error {
					return b34PutCounterThenDoubleCounters(NewContext(g, item))
				}),
				TargetCreature("target creature")),
		},
	})
}
