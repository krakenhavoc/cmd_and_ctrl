package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "cada2a1c-5db2-4702-9a57-cbe7da1bf208",
		Name:         "Tome of Legends",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(page, 1, "Tome of Legends: enters with a page counter"),
		},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventETB, game.EventAttack},
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if ev.CardID == source.InstanceID {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && c.IsCommander && c.Owner == source.Controller
				},
				"Tome of Legends — put a page counter on it",
				func(g *game.Game, item *game.StackItem) error {
					return effects.AddCounter{Target: item.SourceCardID, Kind: page, N: 1}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Remove a page counter from this artifact: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.RemoveCountersFromThis(page, 1)),
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
