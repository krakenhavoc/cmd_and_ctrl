package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// boltWardedCreature casts a Lightning Bolt at the warded creature and
// returns the spell's ID. The Glutton is a 4/4 and survives it, so a
// test reads whether the bolt resolved off the damage on the creature.
func BoltWardedCreature(t *testing.T, g *game.Game, caster *game.Player, victim uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Lightning Bolt", TypeLine: "Instant",
		OracleID: LightningBoltOracle, Owner: caster.ID, Controller: caster.ID,
	})
	if err := g.CastSpell(caster.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("a warded permanent is a LEGAL target — announce must succeed: %v", err)
	}
	return id
}

// boltResolved is true when the Lightning Bolt dealt its damage: the
// creature either died or carries 3 marked damage.
func BoltResolved(g *game.Game, victim uuid.UUID) bool {
	c, ok := g.LookupCardForEffect(victim)
	if !ok || !g.Battlefield.Contains(victim) {
		return true
	}
	return c.DamageMarked == 3
}

func FillHand(p *game.Player, names ...string) []uuid.UUID {
	var ids []uuid.UUID
	for _, n := range names {
		c := game.NewCard(n, p.ID)
		c.TypeLine = "Sorcery"
		p.Hand.PushTop(c)
		ids = append(ids, c.InstanceID)
	}
	return ids
}
