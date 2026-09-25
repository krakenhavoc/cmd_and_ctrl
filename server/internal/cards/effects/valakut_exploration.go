package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Valakut Exploration — Enchantment {2}{R} (EDHREC rank ~1597):
//
//	"Landfall — Whenever a land you control enters, exile the top card
//	 of your library. You may play that card for as long as it remains
//	 exiled.
//	 At the beginning of your end step, if there are cards exiled with
//	 this enchantment, put them into their owner's graveyard, then
//	 this enchantment deals that much damage to each opponent."
//
// A landfall deck's card-advantage engine with a shot clock that
// resets every land drop instead of every turn: play lands, stack up
// impulse-exiled cards, and either play them before your end step or
// take (and deal) the damage for sitting on them.
//
// # This card needed nothing new — it needed b27ExiledWith to exist
//
// It was declared skipped on batch 14 (#307) for "linked exile
// ('exiled with this enchantment') and an exile → graveyard move".
// Both halves are ordinary composition of primitives that shipped
// later for other cards, and neither is a "post-departure" read: this
// card's own trigger requires ITSELF to still be on the battlefield
// (an end-step trigger has no source once it leaves), so nothing here
// needed game.LastKnownCountersForEffect or a record that survives
// past this permanent's own life. What it needed was:
//
//   - the LINKED EXILE identity — b27ExiledWith (Duplicant, Angel of
//     Serenity, Chrome Mox, Bag of Holding, Ossification), which
//     reads back off the event log which cards moved into exile while
//     THIS card's ability was resolving, keyed by (source, label);
//   - "you may play it for as long as it remains exiled" —
//     game.CastPermission{Duration: WhileInZoneDuration()}, the same
//     primitive airbend.go's ExileWithPermission uses for its own
//     "while it's exiled" grant, applied to the OTHER half of the
//     exile-with-permission pair (ExileTopWithPermissionForEffect,
//     the library-facing shape impulse draw already uses);
//   - the exile → graveyard move — PutCardsIntoGraveyardThenForEffect
//     (simultaneous.go), the one addition this card needed: a
//     continuation sibling of PutIntoGraveyardForEffect
//     (random_bottom.go's Genesis Wave primitive: "not a mill, not a
//     discard"), because "that much damage" has to count what
//     actually LANDED in the graveyard, not what was asked for before
//     the CR 903.9 window had a chance to redirect a leg — the same
//     lesson #911 already taught ExileCardsThenForEffect and
//     DestroyPermanentsThenForEffect.
//
// So the fix for the row this card sat on (docs/engine-seams.md,
// "Post-departure LKI / 'exiled with this' record") is The Ozolith's
// counters-at-departure LKI (#1218) — this card and Currency
// Converter are the two of the row's three that turned out to need no
// new engine primitive at all, just the composition above; see the
// tracking issue for the full accounting.
//
// # Re-checked at resolution, not just at trigger time
//
// The end-step ability's Build freezes the exiled-with list (CR
// 603.3d: the condition names the object as it is when the ability
// triggers). It freezes it as DATA — the item's Payload, the carrier a
// reflexive trigger's "cards revealed this way" already uses — so the
// row declares its Effect and a table with the ability waiting is a
// restore point (ADR 0041 P9, #1497). valakutExplorationSweep still
// confirms each card is
// STILL IN EXILE before moving it — a card someone cast in the
// priority window between the trigger and its resolution is not
// double-counted or yanked off the stack (CR 608.2b).
//
// No simplification.

const valakutExplorationOracle = "d6861319-ae16-4e6c-af87-a264f667d694"

// valakutExplorationExileLabel is the stack label the landfall
// trigger carries. b27ExiledWith keys the "exiled with this
// enchantment" record on it, so the trigger and the end-step sweep
// must agree — a const rather than two string literals, the same
// discipline Bag of Holding's own label follows.
const valakutExplorationExileLabel = "Valakut Exploration — exile the top card of your library"

// valakutExplorationSweepLabel is the end-step ability's stack label
// and its row's Key.
const valakutExplorationSweepLabel = "Valakut Exploration — graveyard and damage"

func init() {
	Register(Spec{
		OracleID:     valakutExplorationOracle,
		Name:         "Valakut Exploration",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Landfall(valakutExplorationExileLabel, valakutExplorationLandfall),
			{
				Watches: []game.EventKind{game.EventBeginEndStep},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.Actor == source.Controller &&
						len(b27ExiledWith(g, source.InstanceID, valakutExplorationExileLabel)) > 0
				},
				Key:    valakutExplorationSweepLabel,
				Build:  valakutExplorationRecordExiled,
				Effect: valakutExplorationSweep,
			},
		},
	})
}

// valakutExplorationLandfall is the landfall half: exile the
// controller's own top card and let them play it for as long as it
// remains exiled — no turn limit, unlike ordinary impulse draw.
func valakutExplorationLandfall(g *game.Game, item *game.StackItem) error {
	controller := item.Controller
	_, err := g.ExileTopWithPermissionForEffect(controller, controller, 1, game.CastPermission{
		Duration: game.WhileInZoneDuration(),
	})
	return err
}

// valakutExplorationRecordExiled is the end-step ability's fill-in
// Build: it writes the cards exiled with this enchantment, as they are
// when the ability triggers, onto the item's Payload and leaves the
// Effect to the row. No cards (the AppliesTo gate makes that
// unreachable in play) is no trigger, as before.
func valakutExplorationRecordExiled(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
	ids := b27ExiledWith(g, source.InstanceID, valakutExplorationExileLabel)
	if len(ids) == 0 {
		return nil
	}
	item := game.NewTriggeredItem(source, valakutExplorationSweepLabel, nil)
	item.Payload = make([]game.TargetRef, 0, len(ids))
	for _, id := range ids {
		item.Payload = append(item.Payload, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	return item
}

// valakutExplorationSweep is the end-step half: every card the item's
// Payload names that is
// still sitting in exile goes to its owner's graveyard, and the
// controller's opponents each take one damage per card that actually
// landed there.
//
// The damage total is read from the continuation's `landed` list, not
// counted before the move — PutCardsIntoGraveyardThenForEffect's own
// doc explains why: a leg can pause mid-move (CR 903.9), and reading
// a pre-move tally would pay out for a move the window has not
// answered yet (#1218, ADR 0013 §5t).
func valakutExplorationSweep(g *game.Game, item *game.StackItem) error {
	var stillExiled []uuid.UUID
	for _, id := range NewContext(g, item).PayloadCards() {
		if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneExile {
			stillExiled = append(stillExiled, id)
		}
	}
	if len(stillExiled) == 0 {
		return nil
	}
	return g.PutCardsIntoGraveyardThenForEffect(stillExiled, func(g *game.Game, landed []uuid.UUID) error {
		if len(landed) == 0 {
			return nil
		}
		ctx := NewContext(g, item)
		for _, opp := range ctx.Opponents() {
			if err := g.DealDamageToPlayerForEffect(item.SourceCardID, opp, len(landed)); err != nil {
				return err
			}
		}
		return nil
	})
}
