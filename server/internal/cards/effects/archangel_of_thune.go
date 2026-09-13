package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Archangel of Thune — Creature — Angel {3}{W}{W}, 3/4 (EDHREC rank
// 1380):
//
//	"Flying
//	 Lifelink
//	 Whenever you gain life, put a +1/+1 counter on each creature you
//	 control."
//
// The lifegain deck's finisher: every point of lifegain — its own
// lifelink included — is an anthem. The condition is Marauding
// Blight-Priest's (every lifegain path emits a positive life-change
// event on the controller, lifelink included), and the body is
// Abandoned Air Temple's: the creature set is snapshotted first so a
// creature made mid-loop by a counter-placement trigger does not
// receive one. A single lifegain event is one trigger, however much
// life it was — the printed "whenever you gain life" is per event,
// not per point.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4f2d4538-dc1d-4c09-964b-b0d7c240fb7d",
		Name:            "Archangel of Thune",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventChangeLife},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b10YouGainedLife(ev, source)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Archangel of Thune — a +1/+1 counter on each creature you control",
					b11PutCounterOnEachCreatureYouControl)
			},
		}},
	})
}
