package game

import "github.com/google/uuid"

// cant_be_enchanted.go — #6203: "can't be enchanted", "can't be enchanted
// by other Auras" and "other players can't gain control of".
//
//	Printed                                         Bit (ExtraRestriction)              On
//	Tetravite: "This token can't be enchanted."     CantBeEnchanted                     the host
//	Guardian Beast: "can't be enchanted … doesn't   CantBecomeEnchanted                 the host
//	  remove Auras already attached"
//	Anti-Magic Aura, Consecrate Land: "can't be     EnchantedCantBeEnchantedByOthers    the Aura itself
//	  enchanted by other Auras"
//	Guardian Beast: "other players can't gain       OthersCantGainControl               the host
//	  control of them"
//
// CR 303.4: "Other effects can limit what a permanent can be enchanted
// by." The limit is read at every point an Aura meets a host: the Aura
// spell's target (canBeTargetedBy), an Aura entering without a target
// (auraHostCandidatesLocked, CR 303.4f), an attach (AttachForEffect, CR
// 303.4j) and the state-based check (attachmentLegalLocked, CR 303.4c).
// The "new attachments only" form is not read by the state-based check.

// auraBarredFromHostLocked reports whether `aura` can't enchant `host`
// because of a can't-be-enchanted effect. `newAttach` is true for an
// attachment being made (a target, an entry, an attach) and false for
// the state-based check of one already made. False for a non-Aura.
//
// Caller must hold g.mu with fresh layers.
func (g *Game) auraBarredFromHostLocked(aura, host *Card, newAttach bool) bool {
	if aura == nil || host == nil || !aura.IsAura() {
		return false
	}
	x := host.Effective().ExtraRestrictions
	if x.Has(CantBeEnchanted) || (newAttach && x.Has(CantBecomeEnchanted)) {
		return true
	}
	if g == nil || g.Battlefield == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		other := &g.Battlefield.Cards[i]
		if other.InstanceID == aura.InstanceID || !other.IsAttachedTo(host.InstanceID) {
			continue
		}
		if other.Effective().ExtraRestrictions.Has(EnchantedCantBeEnchantedByOthers) {
			return true
		}
	}
	return false
}

// othersCantGainControlLocked reports whether `gainer` is barred from
// gaining control of `target` (#6203): the permanent has
// OthersCantGainControl and `gainer` is not its controller.
//
// Caller must hold g.mu with fresh layers.
func othersCantGainControlLocked(target *Card, gainer uuid.UUID) bool {
	if target == nil || gainer == uuid.Nil || gainer == target.Controller {
		return false
	}
	return target.Effective().ExtraRestrictions.Has(OthersCantGainControl)
}

// controlChangeBarredLocked reports whether a scoped record about to be
// registered hands control of a barred permanent to another player: one
// of its mods is a setController and one of its affected objects is a
// battlefield permanent that player can't gain control of.
//
// Caller must hold g.mu with fresh layers.
func (g *Game) controlChangeBarredLocked(affected []AffectedObject, mods []Mod) bool {
	for _, m := range mods {
		if m.Kind != ModSetController {
			continue
		}
		for _, a := range affected {
			if a.OnStack {
				continue
			}
			if c := findBattlefieldCard(g, a.ID); othersCantGainControlLocked(c, m.Player) {
				return true
			}
		}
	}
	return false
}
