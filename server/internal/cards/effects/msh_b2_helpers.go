package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msh_b2_helpers.go — shared bodies for Marvel Super Heroes, half B, second
// batch (#4455). Every name carries the mshB2 prefix because the other
// Marvel batches land beside this one.

// mshB2AnotherVillainEntered is the condition of "whenever another Villain
// you control enters" (Baron Strucker, Crossbones): a permanent under the
// source controller's control, not the source, with the Villain subtype as
// it enters (a changeling counts).
func mshB2AnotherVillainEntered(ev game.Event, source *game.Card, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, true)
	return ok && c.HasSubtype(source.TextSubtype("Villain"))
}

// mshB2YouPutPlusOneOnThisTurn reports whether the source's controller has
// put one or more +1/+1 counters on the source this turn (Beast, Erudite
// Aerialist): any placement in this turn's event log that names them as the
// placer and the source as the recipient.
func mshB2YouPutPlusOneOnThisTurn(g *game.Game, source *game.Card) bool {
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventCounterPlaced || ev.Target != source.InstanceID || ev.Label != game.CounterPlusOne {
			continue
		}
		if _, ok := countersPutBy(ev, source.Controller, g, true); ok {
			return true
		}
	}
	return false
}

// mshB2PlanCounterReached is the When of "when the Nth plan counter is put
// on this enchantment" (Claim the Kingdom): a plan-counter placement on the
// source whose new total reaches n from below.
func mshB2PlanCounterReached(n int) func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if ev.Kind != game.EventCounterPlaced || ev.Target != source.InstanceID || ev.Label != "plan" || ev.Amount < n {
			return false
		}
		return b12CounterTotalBefore(ev, g) < n
	}
}

// mshB2SacrificeThisThenWhenYouDo is "sacrifice it. When you do, <body>":
// the source goes first (nothing happens for a source that has left or come
// back as a new object), and only then is the reflexive trigger put on the
// stack, so its target is chosen after the sacrifice.
func mshB2SacrificeThisThenWhenYouDo(label string, body game.BodyRef) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		if !onBattlefield(g, item.SourceCardID) || ctx.isNewSourceObjectAsThis(item.SourceCardID) {
			return nil
		}
		if err := (SacrificePermanent{Target: item.SourceCardID}).Apply(ctx); err != nil {
			return err
		}
		return WhenYouDo(label, body).Apply(ctx)
	}
}

// mshB2TargetCreatureTheAttackerDefends is the TargetsFrom clause of
// "target creature defending player controls" on a trigger whose source is
// not the attacker (Captain America's Shield: the equipped creature
// attacked). The defending player is the one the attacker named by the
// triggering event is attacking.
func mshB2TargetCreatureTheAttackerDefends(tc game.TriggerContext, _ *game.Card, g *game.Game) *game.TargetSpec {
	attacker, ok := g.LookupCardForEffect(tc.Event.CardID)
	if !ok {
		return nil
	}
	defender := g.DefendingPlayerForAttackForEffect(attacker.AttackingTarget)
	if defender == uuid.Nil {
		return nil
	}
	return &game.TargetSpec{
		Mode: "creature", Label: "target creature defending player controls",
		Zones: []game.ZoneKind{game.ZoneBattlefield},
		CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return c.IsCreature() && c.Controller == defender
		},
		Min: 1, Max: 1,
	}
}
