package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Matoya, Archon Elder — Legendary Creature — Human Warlock {2}{U}, 1/4
// (Final Fantasy, #4462):
//
//	"Whenever you scry or surveil, draw a card. (Draw after you scry or
//	 surveil.)"
//
// Prudent Fateseer's trigger on the scry and surveil events (CR 701.22,
// 701.25), with no once-a-turn gate: it triggers once per scry or
// surveil you do, and the draw waits on the stack until that scry or
// surveil is finished, as the reminder text says.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ef6c0054-3a89-4fac-b822-37c7d2514489",
		Name:         "Matoya, Archon Elder",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventScry, game.EventSurveil}, ByYou,
				"Matoya, Archon Elder — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
