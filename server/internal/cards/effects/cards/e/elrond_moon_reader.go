package cardse

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Elrond, Moon-Reader — Legendary Creature — Elf Noble {2}{U}, 3/3:
//
//	"Whenever you activate an ability of a creature, draw a card. This
//	 ability triggers only once each turn.
//	 {5}{U}{U}: Exile up to two other target nonland permanents you
//	 control. Return those cards to the battlefield under their owner's
//	 control at the beginning of the next end step."
//
// The trigger watches both activation kinds, because a mana ability is an
// activated ability too (CR 605.1a), and reads the ability's source as a
// creature: live while it is on the battlefield, and as it last existed
// when the activation itself took it away (a creature sacrificed to pay
// for its own ability, whose last-known record must be the very object
// the ability names, CR 113.7a). Once each turn is the per-turn trigger
// record (b11TriggeredThisTurn), as Exemplar of Light reads it; Elrond
// triggers on his own abilities too ("a creature" does not say another).
//
// The flicker is Eerie Interlude's exile-then-return at the next end
// step (exileTargetsThenScheduleReturn); "up to two" lets the player
// pick none, one or two, and Another keeps Elrond out of his own target
// list.
//
// No simplification. The Hobbit (#4454).
const elrondDrawLabel = "Elrond, Moon-Reader — draw a card"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "3f4d6f91-95ad-4687-8899-5a21a0abb49e",
		Name:         "Elrond, Moon-Reader",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny(effects.TheTwoActivationKinds, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && elrondActivatedACreatureAbility(ev, g) &&
					!effects.B11TriggeredThisTurn(g, source.InstanceID, elrondDrawLabel)
			}, elrondDrawLabel, effects.Do(effects.DrawCards{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}{U}{U}: Exile up to two other target nonland permanents you control. Return those cards to the battlefield under their owner's control at the beginning of the next end step.",
			Cost:    effects.ManaCost("{5}{U}{U}"),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Targets: effects.Another(effects.TargetPermanent("up to two other target nonland permanents you control", effects.Nonland(), effects.YouControl()).WithCount(0, 2)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ExileTargetsThenScheduleReturn(effects.NewContext(g, item),
					"Elrond, Moon-Reader — return the exiled permanents")
			},
		}},
	})
}

// elrondActivatedACreatureAbility is "an ability of a creature" for an
// activation event: the source is a creature on the battlefield now, or
// (a sacrifice cost took it away) was one as the ability named it.
func elrondActivatedACreatureAbility(ev game.Event, g *game.Game) bool {
	if effects.ActivationSourceMatches(g, ev, effects.Creature()) {
		return true
	}
	if ev.Kind != game.EventActivateAbility {
		return false
	}
	item := g.StackItemForEffect(ev.StackItemID)
	if item == nil {
		return false
	}
	info, ok := g.LastKnownPermanentForEffect(item.SourceObject.ID)
	return ok && info.Epoch == item.SourceObject.Epoch && slices.Contains(info.Characteristic.Types, "Creature")
}
