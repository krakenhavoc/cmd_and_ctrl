package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Baleful Strix — Artifact Creature — Bird {U}{B}, 1/1 (EDHREC rank
// 360):
//
//	"Flying, deathtouch
//	 When this creature enters, draw a card."
//
// A cantrip that trades with anything: both keywords are printed
// and enforced by the combat engine (flying restricts blockers,
// deathtouch makes any damage lethal), and the ETB draw is
// Mulldrifter's trigger. Being an artifact creature, it also counts
// for artifact-count clauses, off the printed type line.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "37688720-03de-4eca-a82d-a0afe8d58adc",
		Name:            "Baleful Strix",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Baleful Strix — draw a card",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
