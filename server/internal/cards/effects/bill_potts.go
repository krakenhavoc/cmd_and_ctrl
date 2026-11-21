package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bill Potts — Legendary Creature — Human {3}{R}, 2/4 (Doctor Who):
//
//	"Whenever you cast an instant or sorcery spell that targets only Bill Potts or
//	 activate an ability that targets only Bill Potts, copy that spell or ability.
//	 You may choose new targets for the copy. This ability triggers only once each
//	 turn.
//	 Doctor's companion (You can have two commanders if the other is the Doctor.)"
//
// One ability with two trigger conditions (a cast, an activation), each requiring
// that every target of the spell or ability is Bill ("only" is every target, in
// each instance of the word "target", being Bill, so one that names Bill twice
// qualifies and one that names anything else does not) and that you control it.
// The once-a-turn limit is on the trigger (TriggersOnlyOnceEachTurn). The copy is
// Ertha Jo's copy of an ability and Double Major's of a spell, each with CR 707.10c's
// choice of new targets. Doctor's companion is a deck-construction rule read off the
// oracle text.
//
// No simplification.
func init() {
	const label = "Bill Potts — copy that spell or ability; you may choose new targets for the copy"
	Register(Spec{
		OracleID:     "0ebe67ed-8a71-4bbc-8e58-cebbd7b2d83f",
		Name:         "Bill Potts",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggersOnlyOnceEachTurn(OnAny([]game.EventKind{game.EventCast, game.EventActivateAbility}, billPottsTargetsOnlyHim, label,
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					ev := ctx.Trigger().Event
					if ev.Kind == game.EventActivateAbility {
						return copyTheTriggeringAbility(g, item)
					}
					return CopySpell{StackID: ev.CardID, Controller: item.Controller, ChooseNewTargets: true}.Apply(ctx)
				})),
		},
	})
}

// billPottsTargetsOnlyHim is the trigger condition: its controller cast an instant
// or sorcery spell, or activated an ability, whose every target is Bill.
func billPottsTargetsOnlyHim(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil || ev.Actor != source.Controller {
		return false
	}
	var targets []game.TargetRef
	switch ev.Kind {
	case game.EventCast:
		spell, ok := g.LookupCardForEffect(ev.CardID)
		if !ok || !(spell.HasCardType("instant") || spell.HasCardType("sorcery")) {
			return false
		}
		if it := g.StackItemForEffect(ev.CardID); it != nil {
			targets = it.Targets
		}
	case game.EventActivateAbility:
		if ev.StackItemID == uuid.Nil {
			return false
		}
		if it := g.StackItemForEffect(ev.StackItemID); it != nil {
			targets = it.Targets
		}
	default:
		return false
	}
	if len(targets) == 0 {
		return false
	}
	for _, t := range targets {
		if t.Kind != game.TargetCard || t.ID != source.InstanceID {
			return false
		}
	}
	return true
}
