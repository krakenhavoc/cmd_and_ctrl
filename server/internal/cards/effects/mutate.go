package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// mutate.go — CR 702.140, mutate (ADR 0223, #5910). The rules side is
// game/mutate.go.

// Mutate is "Mutate [cost]" (CR 702.140a): an alternative cost that
// makes the spell a mutating creature spell targeting "a non-Human
// creature with the same owner as this spell". The engine does the
// rest: CR 702.140b's fall-back to a creature spell, the over/under
// choice and the merge (CR 702.140c), and the merged permanent's
// characteristics and abilities (CR 702.140e).
//
// The target is a creature on the battlefield, read through the
// layered view (a creature that became a Human is not a legal target),
// owned by the spell's owner (TargetSpec.OwnedBySourceOwner), whoever
// controls it.
func Mutate(cost string) game.AlternativeCost {
	t := TargetCreature("target non-Human creature you own", Not(Subtype("Human")))
	t.OwnedBySourceOwner = true
	return game.AlternativeCost{
		Key:      game.MutateKey,
		Label:    "Mutate " + cost,
		ManaCost: cost,
		Targets:  t,
	}
}

// MutateFromHandOrGraveyard is Mutate plus "You may cast this card from your
// graveyard using its mutate ability" (Brokkos, Apex of Forever): two offers
// with two keys, the shape BlitzFromHandOrGraveyard uses. "mutate" is bound to
// the hand and "mutate-graveyard" to the graveyard, with the same price and
// target. The card also lists game.ZoneGraveyard in CastableZones, and the
// graveyard opens only through its own offer: the card cannot be cast from
// there for its mana cost.
func MutateFromHandOrGraveyard(cost string) []game.AlternativeCost {
	hand := Mutate(cost)
	hand.FromZone = game.ZoneHand
	grave := hand
	grave.Key = game.MutateGraveyardKey
	grave.FromZone = game.ZoneGraveyard
	return []game.AlternativeCost{hand, grave}
}

// ThisCreatureMutates is "Whenever this creature mutates" (CR
// 702.140d): a mutating creature spell merged with the permanent this
// ability is on. The mutated permanent is the same object as before
// (CR 730.2c) and has every component's abilities (CR 702.140e), so the
// ability triggers whether its card went over, went under, or was the
// creature the spell merged with.
func ThisCreatureMutates(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return source != nil && ev.CardID == source.InstanceID
}
