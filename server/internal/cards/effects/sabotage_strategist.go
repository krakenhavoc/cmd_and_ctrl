package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sabotage Strategist — Creature — Vedalken Ranger {2}{U}{U}, 2/2
// (Aetherdrift, #4464):
//
//	"Flying, vigilance
//	 Whenever one or more creatures attack you, those creatures get -1/-0
//	 until end of turn.
//	 Exhaust — {5}{U}{U}: Put three +1/+1 counters on this creature.
//	 (Activate each exhaust ability only once.)"
//
// One trigger per attack declaration (OncePerBatch, CR 603.2c) whose
// attackers include one declared against YOU: an attack at a planeswalker
// or a battle you control is not an attack on you (CR 508.3e), so it
// triggers nothing. "Those creatures" are the creatures of that
// declaration that attacked you, read from the declaration's own batch
// when the trigger resolves, never a creature that merely entered
// attacking you (it was not declared as an attacker, CR 508.4) and not one
// attacking someone else in the same declaration. Each gets -1/-0 until end
// of turn. Exhaust is the exhaust row of Afterburner Expert's shape.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1b788680-28e3-4378-9607-bd42acb4fbad",
		Name:            "Sabotage Strategist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor != uuid.Nil && ev.Actor != source.Controller && ev.Target == source.Controller
			}, "Sabotage Strategist — those creatures get -1/-0 until end of turn", sabotageStrategistShrink)),
		},
		Activated: []ActivatedAbility{{
			Label:   "Exhaust — {5}{U}{U}: Put three +1/+1 counters on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Exhaust: true,
			Cost:    ManaCost("{5}{U}{U}"),
			Effect:  plusOneCountersOnThis(3),
		}},
	})
}

// sabotageStrategistShrink gives -1/-0 to every creature of the batch that
// attacked the trigger's controller.
func sabotageStrategistShrink(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	batch := item.Trigger.Event.Batch
	seen := map[uuid.UUID]bool{}
	for _, ev := range g.Events {
		if ev.Kind != game.EventAttack || ev.Batch != batch || ev.Target != item.Controller || ev.CardID == uuid.Nil || seen[ev.CardID] {
			continue
		}
		seen[ev.CardID] = true
		if !onBattlefield(g, ev.CardID) {
			continue
		}
		if err := (BoostUntilEOT{Target: ev.CardID, Power: -1, Label: "Sabotage Strategist — -1/-0"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
