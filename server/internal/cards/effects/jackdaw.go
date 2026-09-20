package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jackdaw — Legendary Artifact — Vehicle, {1}{U}{R}:
//
//	"Whenever Jackdaw deals combat damage to a player, you may discard
//	 your hand. If you do, draw a card for each artifact you control.
//	 Crew 3"
//
// "If you do" is NOT a CR 603.12 reflexive triggered ability — that
// rule's signal phrase is "when you do" (its own example,
// Heart-Piercer Manticore: "you may sacrifice another creature. WHEN
// YOU DO, [deal damage]"), and it means a second object that goes on
// the stack ABOVE the parent with its own response window. "If you
// do" has no rule of its own; it is ordinary same-resolution
// sequencing — CR 608.2c's "later text on the card may modify the
// meaning of earlier text", the same mechanism "Counter target spell.
// If that spell is countered this way, …" uses. One instruction, run
// top to bottom, nothing to respond to in between.
//
// This card originally shipped built on ReflexiveTrigger, following
// the (correctly-"when you do") Undead Butler / Generous Plunderer
// precedent without checking that Jackdaw's own printed word is
// different. That was wrong in two observable ways: it opened a
// response window the card doesn't print (an opponent could act
// between the discard and the draw and change the artifact count),
// and it meant the draw was priced off a game state one priority pass
// later than printed. Fixed by making jackdawDiscardHandThenDraw a
// single Effect that discards and then draws in the same resolution —
// no continuation machinery needed, because the discard offers no
// choice ("your hand", not "a card") and so nothing can pause between
// the two halves.
//
// The count is artifacts controlled AFTER the discard, which falls
// out of running the two lines in order rather than from any special
// primitive. Jackdaw counts itself: nothing on the card excludes it,
// and it is still on the battlefield (the ability doesn't sacrifice
// it).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ac161100-46b7-4f0f-a72b-84aaa45980ad",
		Name:         "Jackdaw",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(
				WheneverThisDealsCombatDamageToAPlayer("Jackdaw — discard your hand", jackdawDiscardHandThenDraw),
				"Jackdaw — discard your hand? (Draw a card for each artifact you control.)"),
		},
		Activated: []ActivatedAbility{{
			Label:  "Crew 3",
			Cost:   CrewCost(3),
			Effect: CrewEffect("Jackdaw"),
		}},
	})
}

// jackdawDiscardHandThenDraw is the whole ability's body, run as one
// resolution: discard the whole hand (no choice of what to pitch —
// it's all of it), then draw a card for each artifact controlled
// AFTER that discard. "If you do" needs no continuation object —
// Optional's "you may" already means this only runs on a yes, so the
// discard is unconditional by the time this Effect is called, and the
// draw simply comes next.
func jackdawDiscardHandThenDraw(g *game.Game, item *game.StackItem) error {
	if _, err := discardWholeHand(g, item.Controller); err != nil {
		return err
	}
	n := b03ArtifactsControlled(g, item.Controller)
	if n == 0 {
		return nil
	}
	return g.DrawNForEffect(item.Controller, n)
}
