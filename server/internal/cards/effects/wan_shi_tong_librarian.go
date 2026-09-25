package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wan Shi Tong, Librarian — Legendary Creature — Bird Spirit {X}{U}{U},
// 1/1 (EDHREC rank unlisted):
//
//	"Flash
//	 Flying, vigilance
//	 When Wan Shi Tong enters, put X +1/+1 counters on him. Then draw
//	 half X cards, rounded down.
//	 Whenever an opponent searches their library, put a +1/+1 counter
//	 on Wan Shi Tong and draw a card."
//
// The proof card for two engine reads (#1312, #1335) found by the
// *Aang is so flashy* deck triage (#1306).
//
//   - "Put X +1/+1 counters on him. Then draw half X" is a printed
//     ENTERS TRIGGER, not a CR 614.1c entry clause ("enters WITH N
//     counters") — the counters go on AFTER the creature has already
//     landed, as an ordinary CR 603 trigger you can respond to, so
//     EntryCountersFromCast is the wrong shape (that primitive fires
//     INSIDE the CR 614 window, before there is a permanent to watch).
//     Until #1312 an enters trigger's Build had nowhere to read X
//     from: the resolving spell's StackItem is gone by the time a
//     trigger's Effect runs, which is why Goose Mother, Farmer Cotton,
//     Springleaf Parade and Spiteful Banditry all moved their own
//     X-reads into OnResolve instead — a beat early, with a declared
//     caveat that the tokens/counters/damage arrive without a trigger
//     on the stack to respond to. Wan Shi Tong reads source.CastX()
//     (CastProvenance.X, CR 107.3m) inside Build — synchronous, before
//     the trigger goes on the stack — and closes over the plain int
//     rather than the card, so the printed ordering is exact: the
//     trigger itself can be countered or responded to, and only THEN
//     does it put the counters on and draw.
//   - "Whenever an opponent searches THEIR library" needs
//     EventSearchLibrary's Target (#1335) to tell that apart from "an
//     opponent searches YOUR library" (Bribery) — AnOpponentSearches-
//     TheirOwnLibrary in triggers_common.go, the same condition
//     Archivist of Oghma uses.
func init() {
	Register(Spec{
		OracleID:        "2bace3ea-7d28-4a1c-857a-d3c3a1c11b59",
		Name:            "Wan Shi Tong, Librarian",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: Self,
				Key:       "Wan Shi Tong, Librarian — enters: counters and draw",
				// A fill-in Build (ADR 0041 P9): #1312's read is a
				// fact of the moment the trigger fired, so it is
				// captured into Params.Amount rather than a closure —
				// not the card, which the Effect below must not
				// capture (AGENTS.md: undo restores a cloned game and
				// the closure has to resolve against that one).
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, "Wan Shi Tong, Librarian — enters: counters and draw", nil)
					item.Params.Amount = source.CastX()
					return item
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					x := item.Params.Amount
					ctx := NewContext(g, item)
					if x > 0 {
						if err := (AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: x}).Apply(ctx); err != nil {
							return err
						}
					}
					// "Then draw half X cards, rounded down" —
					// integer division on a non-negative X
					// (CR 107.3) is exactly floor(X/2), and
					// DrawCards no-ops on N<=0.
					return DrawCards{Player: item.Controller, N: x / 2}.Apply(ctx)
				},
			},
			On(game.EventSearchLibrary, AnOpponentSearchesTheirOwnLibrary,
				"Wan Shi Tong, Librarian — opponent searches: counter and draw",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
						return err
					}
					return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
				}),
		},
	})
}
