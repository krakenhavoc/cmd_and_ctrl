package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mai, Jaded Edge — Legendary Creature — Human Noble {1}{R}, 1/3:
//
//	"Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)
//	 Exhaust — {3}: Put a double strike counter on Mai. (Activate each
//	 exhaust ability only once.)"
//
// Prowess is the engine's keyword. The exhaust ability is CR 702.177a's
// "activate only once", keyed to the object (CR 400.7), so Mai back
// from a flicker may use it again, and a Mai that has left and come
// back in response is a new object that gets no counter. The counter is
// a keyword counter (CR 122.1b): Mai has double strike for as long as it
// stays on her.
func init() {
	Register(Spec{
		OracleID:        "b5029f21-e0d6-4f22-9afd-79452de26490",
		Name:            "Mai, Jaded Edge",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Activated: []ActivatedAbility{{
			Label:   "Exhaust — {3}: Put a double strike counter on Mai.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Exhaust: true,
			Cost:    ManaCost("{3}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				return AddCounter{Target: item.SourceCardID, Kind: game.CounterDoubleStrike, N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
