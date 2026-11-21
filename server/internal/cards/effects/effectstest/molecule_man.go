package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// molecule_man_test.go — Molecule Man's "Nonland cards in your hand have
// miracle {0}" (#4455), end to end on the real catalog card. The engine
// half is game/miracle.go (GrantedMiracleTrigger, AltCostKeyGrantedMiracle);
// what is pinned here is the whole road a player walks, and every way it is
// closed.

const MoleculeManOracle = "ba944437-0b55-47cb-92dc-2477ce6726c3"

// moleculeManOnTable puts the real Molecule Man on `owner`'s battlefield.
func MoleculeManOnTable(g *game.Game, owner uuid.UUID) uuid.UUID {
	return PushCatalogPermanent(g, owner, "Molecule Man", "Legendary Creature — Human Villain", MoleculeManOracle, false)
}

// hasGrantedMiracleGrant reports whether p holds a live permission under
// `key` naming cardID as it sits in p's hand now.
func HasMiracleGrantUnder(p *game.Player, cardID uuid.UUID, key string) bool {
	epoch := -1
	for _, c := range p.Hand.Cards {
		if c.InstanceID == cardID {
			epoch = c.ObjectEpoch
		}
	}
	if epoch < 0 {
		return false
	}
	for _, perm := range p.CastPermissions {
		if perm.Zone != game.ZoneHand || perm.AltCostKey != key {
			continue
		}
		for _, ref := range perm.Cards {
			if ref.ID == cardID && ref.Epoch == epoch {
				return true
			}
		}
	}
	return false
}
