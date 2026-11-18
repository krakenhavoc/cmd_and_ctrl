package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Contested War Zone — Land:
//
//	"Whenever a creature deals combat damage to you, that creature's
//	 controller gains control of this land.
//	 {T}: Add {C}.
//	 {1}, {T}: Attacking creatures get +1/+0 until end of turn."
//
// Contested Game Ball's steal on a land: one trigger per damage event,
// the new controller being the controller of the creature that dealt it,
// as it last existed (CR 608.2h). The control change lasts until the land
// leaves (CR 611.2a); a War Zone that left and came back is a new object
// and is not taken (CR 400.7). The pump reads the attacking creatures as
// it resolves (Trumpet Blast's shape), whoever controls them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "ed73de2b-d7f4-48d9-9be2-aa9d111b7aa7",
		Name:          "Contested War Zone",
		Completeness:  CompletenessFull,
		ManaAbilities: []ManaAbility{colorlessTapRow()},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, strixhavenCreatureDealtCombatDamageToYou,
				"Contested War Zone — the creature's controller gains control of this land",
				contestedWarZoneSteal),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Attacking creatures get +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return BoostUntilEOT{Match: AttackingCreature(), Power: 1, Label: "Contested War Zone — attacking creatures get +1/+0"}.Apply(NewContext(g, item))
			},
		}},
	})
}

func contestedWarZoneSteal(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || sourceIsNewObject(g, item) {
		return nil
	}
	dealer, ok := g.LookupCardForEffect(item.Trigger.Event.Source)
	if !ok {
		return nil
	}
	zone := item.SourceCardID
	g.GainControlForEffect(zone, zone, dealer.Controller, game.IndefiniteDuration(),
		"Contested War Zone — gains control of this land")
	return nil
}
