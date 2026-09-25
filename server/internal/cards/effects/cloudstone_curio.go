package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cloudstone Curio — Artifact {3} (#387, #1223):
//
//	"Whenever a nonartifact permanent you control enters, you may
//	 return another permanent you control that shares a permanent
//	 type with it to its owner's hand."
//
// The second card on the trigger-data row, and the other half of what
// that row means. Scrap Trawler's clause reads a NUMBER off the
// event; this one reads a TYPE LINE — "shares a permanent type WITH
// IT", where "it" is the permanent that just entered.
//
// "You may return another permanent you control" is a CHOICE, not a
// target — there is no "target" in the printed text — made on
// resolution (CR 608.2), the ReturnOneYouControl / ChoosePermanents
// posture (#1214, #1337). It used to be TargetsFrom's announce-time
// clause, a declared simplification that let opponents see and
// answer the choice before the trigger resolved and excluded a
// permanent with shroud; neither applies to an untargeted choice, so
// both are gone. Because the candidate set depends on a FACT about
// what entered rather than a static predicate, this calls
// ChoosePermanents directly instead of the simpler
// ReturnOneYouControl wrapper: `entered` is captured as a plain ID
// when the trigger fires, and its CURRENT permanent types (CR 110.4a)
// are read fresh at resolution — the correct timing for an
// instruction with no target, and never less permissive than the old
// announce-time snapshot.
func init() {
	Register(Spec{
		OracleID:     "5cd2fd32-4da2-40eb-b003-c0b9a9ec91c1",
		Name:         "Cloudstone Curio",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: cloudstoneNonartifactEntered,
			Key:       "Cloudstone Curio — return another permanent you control to its owner's hand",
			Effect:    cloudstoneReturnAnother,
		}},
	})
}

// cloudstoneNonartifactEntered is the trigger condition: a
// NONARTIFACT permanent entered the battlefield under the Curio's
// controller's control.
//
// The Curio is itself an artifact, so it never triggers on its own
// entry, and the exclusion is the printed one rather than an
// "another" clause — a second nonartifact permanent entering does
// trigger it.
func cloudstoneNonartifactEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventETB || ev.CardID == source.InstanceID {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && !c.IsArtifact() && c.Controller == source.Controller
}

// cloudstoneReturnAnother is the resolution effect: "you may return
// another permanent you control that shares a permanent type with
// it". The entering permanent is read off item.Trigger.Event.CardID
// rather than a closure (ADR 0041 P9), so a restored item asks again
// about the same object.
//
// If the entered permanent has since left the battlefield — bounced,
// destroyed, or otherwise, in response to the trigger — nothing can
// share a type with a permanent that is not there any more, so there
// is nothing to offer (CR 608.2c: only as much of the ability as
// possible).
func cloudstoneReturnAnother(g *game.Game, item *game.StackItem) error {
	entered := item.Trigger.Event.CardID
	ctx := NewContext(g, item)
	enteredCard, ok := g.LookupCardForEffect(entered)
	if !ok {
		return nil
	}
	return ChoosePermanents{
		Question: "Cloudstone Curio — return another permanent you control to its owner's hand",
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			var out []uuid.UUID
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller != of || c.InstanceID == entered {
					continue
				}
				if cloudstoneSharesAPermanentType(enteredCard, c) {
					out = append(out, c.InstanceID)
				}
			}
			return out, 0, 1
		},
		Then: bouncePickedToHand,
	}.Apply(ctx)
}

// cloudstoneSharesAPermanentType is CR 110.4a's six permanent types,
// checked directly against two live cards rather than through
// game.ObjectSnapshot — that type is built from CR 603.10
// last-known-information at trigger-fire time and is not meant to be
// reconstructed later, so a resolution-time check reads both cards'
// CURRENT effective types instead.
func cloudstoneSharesAPermanentType(a, b game.Card) bool {
	return (a.IsArtifact() && b.IsArtifact()) ||
		(a.IsCreature() && b.IsCreature()) ||
		(a.IsEnchantment() && b.IsEnchantment()) ||
		(a.IsLand() && b.IsLand()) ||
		(a.IsPlaneswalker() && b.IsPlaneswalker()) ||
		(a.IsBattle() && b.IsBattle())
}
