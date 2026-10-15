package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tome of Legends — Artifact — Book {2}:
//
//	"This artifact enters with a page counter on it.
//	 Whenever your commander enters or attacks, put a page counter on
//	 this artifact.
//	 {1}, {T}, Remove a page counter from this artifact: Draw a card."
//
// The page counter at entry is a self-replacement, so it is there before
// any trigger sees the artifact. "Your commander" is a card you own that
// is designated as a commander, whoever controls it now. Removing the
// counter is a cost, paid at announce, so a response cannot spend the
// same counter twice.
//
// No simplification.
func init() {
	const page = "page"
	Register(Spec{
		OracleID:     "cada2a1c-5db2-4702-9a57-cbe7da1bf208",
		Name:         "Tome of Legends",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(page, 1, "Tome of Legends: enters with a page counter"),
		},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventETB, game.EventAttack},
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if ev.CardID == source.InstanceID {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && c.IsCommander && c.Owner == source.Controller
				},
				"Tome of Legends — put a page counter on it",
				func(g *game.Game, item *game.StackItem) error {
					return AddCounter{Target: item.SourceCardID, Kind: page, N: 1}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}, Remove a page counter from this artifact: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}"), TapCost(), RemoveCountersFromThis(page, 1)),
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}
