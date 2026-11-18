package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Famished Foragers — Creature — Vampire {3}{R}, 4/3 (Innistrad: Midnight
// Hunt, #5497):
//
//	"When this creature enters, if an opponent lost life this turn, add
//	 {R}{R}{R}.
//	 {2}{R}, Discard a card: Draw a card."
//
// Flamecache Gecko's entry mana and rummage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e95f22dd-b89a-4836-86ad-5a9a1627af38",
		Name:         "Famished Foragers",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AllOf(Self, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return opponentLifeLostThisTurn(g, source.Controller) > 0
			}), "Famished Foragers — add {R}{R}{R} (an opponent lost life this turn)",
				Do(If{
					Cond: avB34CondAnOpponentLostLife{},
					Then: []Applier{AddMana{Produced: "{R}{R}{R}"}},
				})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{R}, Discard a card: Draw a card.",
			Cost:    Plus(ManaCost("{2}{R}"), DiscardACard()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}
