package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b545_helpers.go — shared bodies for OLD batch 545 (#5593, Lorwyn).

// S82OldB545OtherAttackingOfSubtype counts the attacking creatures with the
// given creature type, other than the source, read as the ability resolves
// through the layered view, so a changeling counts (CR 702.73a). No
// controller clause: "each other attacking Kithkin" is every player's.
func S82OldB545OtherAttackingOfSubtype(g *game.Game, source uuid.UUID, subtype string) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID != source && c.IsCreature() && c.AttackingTarget != uuid.Nil && c.HasSubtypeWord(subtype) {
			n++
		}
	}
	return n
}

// S82OldB545ActivatedAbilityOfSubtype is "whenever you activate an ability of
// a <subtype>": the ability's source is read on the battlefield, or as it
// last existed there when the activation cost sacrificed it (CR 608.2h).
func S82OldB545ActivatedAbilityOfSubtype(ev game.Event, source *game.Card, g *game.Game, subtype string) bool {
	if ev.Actor == uuid.Nil || ev.Actor != source.Controller {
		return false
	}
	if c, ok := g.LookupCardForEffect(ev.CardID); ok && OnBattlefield(g, ev.CardID) {
		return c.HasSubtypeWord(subtype)
	}
	if lki, left := g.LastKnownPermanentForEffect(ev.CardID); left {
		ch := lki.Characteristic
		return slices.Contains(ch.Subtypes, subtype) || ch.AllCreatureTypes && slices.Contains(ch.Types, "Creature")
	}
	return false
}

// S82OldB545SelfWhileYouControlSubtype is "As long as you control a
// <subtype>, this creature gets +P/+T and has <keyword>": a layer 7c pump and
// a layer 6 grant on the source alone, both gated on its controller
// controlling a creature of the type (the source itself counts when it is
// one), read through the layered view so a changeling counts.
func S82OldB545SelfWhileYouControlSubtype(subtype string, power, toughness int, keyword, label string) []game.StaticAbility {
	cond := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return SelfCreature(target, g, source) &&
			CountPermanentsYouControl(g, source, func(o *game.Card) bool {
				return o.IsCreature() && o.HasSubtypeWord(subtype)
			}) > 0
	}
	pump := game.StaticAbility{
		Label:     label,
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: cond,
		Apply:     PumpBy(power, toughness),
	}
	grant := KeywordGrant(cond, keyword)
	grant.Label = label
	return []game.StaticAbility{pump, grant}
}
