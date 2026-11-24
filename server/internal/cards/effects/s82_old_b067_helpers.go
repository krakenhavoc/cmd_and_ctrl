package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// AttachedCreatureBecomesBlocked is "whenever equipped creature becomes
// blocked" (Tormentor's Helm): once per blocked attacker
// (game.EventBecomesBlocked), the attacker being the event's CardID and
// the defending player its Actor, the attacker being the creature this
// Equipment is attached to.
func AttachedCreatureBecomesBlocked(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventBecomesBlocked && source.IsAttachedTo(ev.CardID)
}

// CreatureIsEnchantedOrEquipped is "creature that is enchanted or equipped"
// (Halvar, God of Battle): some Aura or Equipment, whoever controls it, is
// attached to the permanent.
func CreatureIsEnchantedOrEquipped(g *game.Game, target *game.Card) bool {
	if g == nil || g.Battlefield == nil || target == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		a := &g.Battlefield.Cards[i]
		if a.IsAttachedTo(target.InstanceID) && (a.IsAura() || a.HasSubtype("Equipment")) {
			return true
		}
	}
	return false
}

// AuraOrEquipmentOnACreatureYouControl is the first clause of Halvar's
// "attach target Aura or Equipment attached to a creature you control to
// target creature you control": an Aura or Equipment whose host is a
// creature the caster controls.
func AuraOrEquipmentOnACreatureYouControl() CardPredicate {
	return func(g *game.Game, caster uuid.UUID, c game.Card) bool {
		if !c.IsAura() && !c.HasSubtype("Equipment") {
			return false
		}
		host := g.AttachedHostOf(&c)
		return host != nil && host.IsCreature() && host.Controller == caster
	}
}

// ExcessNoncombatDamageToOpponentPermanent is the excess (CR 120.10) a
// noncombat damage event dealt to a creature or planeswalker an opponent
// of `source`'s controller controls, or zero when ev is not that
// (Toralf, God of Fury). The engine computes the excess as the damage is
// dealt (Event.ExcessDamage), so a planeswalker's loyalty before the damage
// is already accounted for.
func ExcessNoncombatDamageToOpponentPermanent(ev game.Event, source *game.Card, g *game.Game) int {
	if ev.Kind != game.EventDealDamage || ev.Combat || ev.ExcessDamage <= 0 {
		return 0
	}
	target, ok := g.LookupCardForEffect(ev.Target)
	if !ok || target.Controller == source.Controller || (!target.IsCreature() && !target.IsPlaneswalker()) {
		return 0
	}
	return ev.ExcessDamage
}

// AnyTargetOtherThan is "any target other than <that permanent>": TargetAny
// with one object excluded (Toralf, God of Fury).
func AnyTargetOtherThan(excluded uuid.UUID) *game.TargetSpec {
	spec := TargetAny()
	spec.Label = "any target other than that permanent"
	base := spec.CardOK
	spec.CardOK = func(g *game.Game, caster uuid.UUID, c game.Card, z game.ZoneKind) bool {
		return c.InstanceID != excluded && base(g, caster, c, z)
	}
	return spec
}
