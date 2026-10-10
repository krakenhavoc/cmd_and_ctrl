package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Surveillance Phantasm — Creature — Bird Illusion {1}{U}, 2/3:
//
//	"Defender, flying, vigilance
//	 As long as you've scried or surveilled this turn, this creature
//	 can attack as though it didn't have defender.
//	 {3}{U}: Surveil 1."
//
// Stalked Researcher's reading of "can attack as though it didn't have
// defender": defender is the only thing that stops a creature
// attacking, so the keyword is removed until end of turn.
//
// The condition is the part that is simplified. Printed, it is a static
// that reads "this turn" live. A static would need the layer cache
// invalidated when a scry or surveil happens (game/layer_listener.go
// bumps on neither EventScry nor EventSurveil) and again when the turn
// ends, which is engine work. So it is a trigger on EventScry and
// EventSurveil by the controller that removes defender until end of
// turn: identical for every scry or surveil that happens while the
// Phantasm is on the battlefield, and weaker than printed for one that
// happened earlier in the turn, before it entered.
func init() {
	Register(Spec{
		OracleID:     "b97648b7-ae98-45c9-8f0f-a52d6faf2d2d",
		Name:         "Surveillance Phantasm",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Looking at the top of your library earlier in the turn, before it entered the battlefield, doesn't let it attack.",
		},
		PrintedKeywords: []string{"defender", "flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventScry, game.EventSurveil}, ByYou,
				"Surveillance Phantasm — can attack this turn as though it didn't have defender",
				func(g *game.Game, item *game.StackItem) error {
					if !onBattlefield(g, item.SourceCardID) {
						return nil
					}
					return untilEndOfTurn(NewContext(g, item), item.SourceCardID, nil,
						"Surveillance Phantasm — can attack as though it didn't have defender",
						game.RemoveKeywordsMod("defender"))
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{U}: Surveil 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{3}{U}"),
			Effect:  Do(Surveil{N: 1}),
		}},
	})
}
