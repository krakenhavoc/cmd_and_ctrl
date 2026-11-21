package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Professor Hojo — Legendary Creature — Human Scientist {1}{G}, 2/2
// (Cloud, Ex-SOLDIER deck, #3656; seam #3779):
//
//	"The first activated ability you activate during your turn that
//	 targets a creature you control costs {2} less to activate.
//	 Whenever one or more creatures you control become the target of an
//	 activated ability, draw a card. This ability triggers only once
//	 each turn."
//
// The discount is an activation-scoped cost modifier (ADR 0048) that
// reads the announced targets (ReadsTargets) and one new fact: whether
// the activator has already made an activation this turn that targeted
// a creature of their own (Game.OwnCreatureTargetedActivationsThisTurn,
// ADR 0020 amendment 2026-10-15). The record is written at the announce,
// beside the activation tally, so a refused activation does not spend the
// discount and a countered one does. "During your turn" is the active
// player, "you activate" is the activator being Hojo's controller, and
// "costs {2} less" takes generic mana only (CR 118.7a) — an ability with
// no generic mana to give up simply is not discounted.
//
// The draw reads Event.TargeterKind (ADR 0018 amendment 2026-10-15), so
// the trigger says "activated ability" rather than "spell or ability",
// copies of an activated ability included (a copy is an activated
// ability, CR 707.10). One ability with two creatures targeted is two
// events; "one or more" and "only once each turn" are the same gate here,
// b11TriggeredThisTurn, which counts the trigger from the moment it is
// put on the stack (a countered one still counts, as printed).
//
// Whose ability it is does not matter: the printed clause says "the
// target of an activated ability", not "of an ability you control", so
// an opponent pointing an activation at your creature draws you a card.
//
// No simplification.
func init() {
	const label = "Professor Hojo — draw a card"
	effects.Register(effects.Spec{
		OracleID:     "4ca66b33-28c2-4d0d-8a44-9d891ab13e18",
		Name:         "Professor Hojo",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.ActivationCostsLessIfItTargets(2,
				"The first activated ability you activate during your turn that targets a creature you control costs {2} less to activate.",
				effects.And(effects.Creature(), effects.YouControl()),
				effects.YouActivate(), effects.ActivatedDuringYourTurn(), effects.TheFirstOwnCreatureTargetingActivationThisTurn()),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, effects.AllOf(
				creatureYouControlBecameTargetOfAnActivatedAbility,
				func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return source != nil && !effects.B11TriggeredThisTurn(g, source.InstanceID, label)
				},
			), label, effects.Do(effects.DrawCards{N: 1})),
		},
	})
}

// creatureYouControlBecameTargetOfAnActivatedAbility is "a creature you
// control becomes the target of an activated ability": the event is a
// card target, the targeter is an activated ability (Event.TargeterKind,
// #3779 — a spell or a triggered ability is not), and the target is a
// creature on the battlefield controlled by the source's controller.
func creatureYouControlBecameTargetOfAnActivatedAbility(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventBecomesTarget || ev.CardID == uuid.Nil || ev.TargeterKind != game.StackItemActivated {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && c.Controller == source.Controller
}
