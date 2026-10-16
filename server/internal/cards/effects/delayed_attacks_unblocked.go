package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// delayed_attacks_unblocked.go — delayed triggers on "attacks and isn't
// blocked" (#3531, ADR 0187 amendment 2026-10-16):
//
//	"Until end of turn, whenever a creature you control attacks and
//	 isn't blocked, …"                                       Gaze of Pain
//	"This turn, when target creature you control attacks and isn't
//	 blocked, …"                               Delif's Cone, Delif's Cube
//
// Both are delayed triggered abilities a resolving spell or ability
// creates (CR 603.7a). They watch the completed block declaration of
// each defending player (game.EventBlockersDeclared), the event the
// printed "whenever this creature attacks and isn't blocked" watches
// (attacksAndIsNotBlocked, CR 509.3g), and are each-conditions
// (game.DelayedEachCondition): one instance for each creature the
// declaration left unblocked, with that creature as the instance's
// payload, its "it".

var (
	// "Whenever a creature you control attacks and isn't blocked": every
	// creature the trigger's controller controls that is attacking the
	// player who just declared and that nothing blocks.
	eachCreatureYouControlAttacksUnblockedCondition = game.DelayedEachCondition(
		"attack/each-creature-you-control-unblocked", eachCreatureYouControlAttacksUnblocked)

	// "When target creature you control attacks and isn't blocked": the
	// one creature named by CondParams.Object, while it is still that
	// object (CR 603.7c: one that left and came back is a new object and
	// never fires it).
	thatCreatureAttacksUnblockedCondition = game.DelayedEachCondition(
		"attack/that-creature-unblocked", thatCreatureAttacksUnblocked)
)

func eachCreatureYouControlAttacksUnblocked(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) []uuid.UUID {
	if ev.Kind != game.EventBlockersDeclared {
		return nil
	}
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == dt.Controller && c.IsCreature() && attacksPlayerUnblocked(g, c, ev.Actor) {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

func thatCreatureAttacksUnblocked(ev game.Event, _ *game.DelayedTrigger, g *game.Game, p game.EffectParams) []uuid.UUID {
	if ev.Kind != game.EventBlockersDeclared || p.Object.ID == uuid.Nil {
		return nil
	}
	c, ok := battlefieldCardForEffect(g, p.Object.ID)
	if !ok || c.ObjectEpoch != p.Object.Epoch || !attacksPlayerUnblocked(g, c, ev.Actor) {
		return nil
	}
	return []uuid.UUID{c.InstanceID}
}

// attacksPlayerUnblocked is CR 509.3g for one creature and the defending
// player whose declaration just completed: it is attacking that player
// (or a planeswalker or battle that player defends) and nothing blocks it.
func attacksPlayerUnblocked(g *game.Game, c game.Card, defender uuid.UUID) bool {
	return c.AttackingTarget != uuid.Nil &&
		g.DefendingPlayerForAttackerForEffect(c.InstanceID) == defender &&
		g.UnblockedAttackerForEffect(c.InstanceID)
}

// whenThatCreatureAttacksUnblocked creates "this turn, when <creature>
// attacks and isn't blocked, <body>" about the creature as the object it
// is now. A creature that is no longer on the battlefield creates
// nothing: the trigger could never fire.
func whenThatCreatureAttacksUnblocked(ctx *Context, creature uuid.UUID, label string, body game.BodyRef) error {
	if creature == uuid.Nil {
		return nil
	}
	c, ok := battlefieldCardForEffect(ctx.Game, creature)
	if !ok {
		return nil
	}
	return DelayedOnEvent{
		Label:      label,
		On:         []game.EventKind{game.EventBlockersDeclared},
		Condition:  thatCreatureAttacksUnblockedCondition,
		CondParams: game.EffectParams{Object: game.ObjectRef{ID: creature, Epoch: c.ObjectEpoch}},
		Body:       body,
	}.Apply(ctx)
}

// firstPayloadCard is the fired instance's "it": the first card on its
// payload, or uuid.Nil.
func firstPayloadCard(ctx *Context) uuid.UUID {
	if cards := ctx.PayloadCards(); len(cards) > 0 {
		return cards[0]
	}
	return uuid.Nil
}

// powerNowOrLastKnown is "its power" for a creature an effect names: its
// power now while it is on the battlefield, else as it last existed there
// (CR 608.2h).
func powerNowOrLastKnown(g *game.Game, id uuid.UUID) int {
	if onBattlefield(g, id) {
		return b31CurrentPowerOf(g, id)
	}
	return b17LastKnownPowerOffBattlefield(g, id)
}
