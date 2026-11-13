package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flayed Nim — Creature — Skeleton {3}{B}, 2/2 (Mirrodin, #5606):
//
//	"Whenever this creature deals combat damage to a creature, that
//	 creature's controller loses that much life.
//	 {2}{B}: Regenerate this creature."
//
// The trigger is Stinkweed Imp's shape (thisDealtCombatDamageToACreature):
// it names the damaged creature as the object it was when the damage was
// dealt, and "that much" is the event's amount, the damage actually
// dealt after prevention. "That creature's controller" is read as the
// trigger resolves (CR 608.2h): the controller now if the creature is
// still on the battlefield, and the one it had as it left if the damage
// killed it. It is life loss, not damage (CR 119.3). The regeneration is
// Albino Troll's.
//
// No simplification.
func init() {
	const label = "Flayed Nim — that creature's controller loses that much life"
	Register(Spec{
		OracleID:     "857818d4-6cf0-4234-b9e0-d7181ba6b54f",
		Name:         "Flayed Nim",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventDealDamage},
			AppliesTo: thisDealtCombatDamageToACreature,
			Key:       label,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, label)
				item.Params.Object = damagedCreatureRef(g, ev.Target)
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				info, ok := g.PermanentForEffect(item.Params.Object)
				if !ok || item.Trigger == nil {
					return nil
				}
				return GainLife{Player: info.Controller, Amount: -item.Trigger.Event.Amount}.Apply(NewContext(g, item))
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    ManaCost("{2}{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Regenerate{Target: item.SourceCardID}.Apply(NewContext(g, item))
			},
		}},
	})
}
