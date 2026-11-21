package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82VowB06OtherAttackersThisCombat lists, once each, the creatures
// `controller` declared as attackers in the same combat as the source's
// latest attack, other than the source: "you attacked with exactly one
// other creature this combat". It reads the turn's attack records, which
// keep what was declared after a creature has left combat (CR 508.1) and
// tell the combats of one turn apart by PhaseID.
func S82VowB06OtherAttackersThisCombat(g *game.Game, controller, source uuid.UUID) []uuid.UUID {
	phase, found := 0, false
	for _, a := range g.TurnTally.Attacks {
		if a.Attacker == source {
			phase, found = a.PhaseID, true
		}
	}
	if !found {
		return nil
	}
	var out []uuid.UUID
	seen := map[uuid.UUID]bool{}
	for _, a := range g.TurnTally.Attacks {
		if a.PhaseID != phase || a.Attacker == source || a.Controller != controller || seen[a.Attacker] {
			continue
		}
		seen[a.Attacker] = true
		out = append(out, a.Attacker)
	}
	return out
}

// s82VowB06WithPartner is the printed "as long as an Equipment named
// <partner> is attached to a creature you control" of Bride's Gown and
// Groom's Finery: some Equipment with that name on the battlefield is
// attached to a creature the source's controller controls. The name is read
// off each Equipment as it is now, so a copy named so counts; the
// source itself can't satisfy it because it never carries its partner's name.
func s82VowB06WithPartner(partner string) func(target *game.Card, g *game.Game, source *game.Card) bool {
	return func(target *game.Card, g *game.Game, source *game.Card) bool {
		if !source.IsAttachedTo(target.InstanceID) {
			return false
		}
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.InstanceID == source.InstanceID || c.Effective().Name != partner || !c.HasSubtype("Equipment") {
				continue
			}
			if c.AttachedTo.Kind != game.TargetCard {
				continue
			}
			if host, ok := g.LookupCardForEffect(c.AttachedTo.ID); ok && host.IsCreature() && host.Controller == source.Controller {
				return true
			}
		}
		return false
	}
}

// s82VowB06EquipmentPair is the shared static text of the two Crimson Vow
// wedding Equipment: "Equipped creature gets +2/+0. It gets an additional
// +0/+2 and has <keyword> as long as an Equipment named <partner> is attached
// to a creature you control." Three statics: the plain pump, then the
// conditional toughness and keyword, which are read live.
func S82VowB06EquipmentPair(partner, keyword string) []game.StaticAbility {
	extra := PumpAttached(0, 2)
	extra.Label = "Equipped creature gets an additional +0/+2 while " + partner + " is attached to a creature you control."
	extra.AppliesTo = s82VowB06WithPartner(partner)
	grant := KeywordGrant(s82VowB06WithPartner(partner), keyword)
	grant.Label = "Equipped creature has " + keyword + " while " + partner + " is attached to a creature you control."
	return []game.StaticAbility{PumpAttached(2, 0), extra, grant}
}
