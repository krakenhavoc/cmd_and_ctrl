package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cryptid Inspector — Creature — Elf Warrior {2}{G}, 2/3 (Duskmourn:
// House of Horror, #4466):
//
//	"Vigilance
//	 Whenever a face-down permanent you control enters and whenever this
//	 creature or another permanent you control is turned face up, put a
//	 +1/+1 counter on this creature."
//
// One ability on two event kinds (OnAny over EventETB and
// EventTurnedFaceUp, Perimeter Enforcer's shape). A face-down permanent
// entering under your control counts whatever it is (Threats Around Every
// Corner reads the entering card's face-down state, and the Inspector
// itself entering face down has no abilities, CR 708.2a). A permanent
// turned face up counts when you control it, the Inspector included: it is
// face up and the same object by the time EventTurnedFaceUp is emitted
// (CR 708.8), so a face-down Inspector turned up puts the counter on
// itself. The counter is a Do primitive (CounterOnThis), so the one
// sentence carries a footprint.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c36f787e-d8d4-4bd5-9790-64766878fe46",
		Name:            "Cryptid Inspector",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventETB, game.EventTurnedFaceUp},
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					switch ev.Kind {
					case game.EventETB:
						c, ok := enteredUnderYourControl(ev, source, g, false)
						return ok && c.FaceDown
					case game.EventTurnedFaceUp:
						c, ok := g.LookupCardForEffect(ev.CardID)
						return ok && c.Controller == source.Controller
					}
					return false
				},
				"Cryptid Inspector — put a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
