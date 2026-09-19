package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const brassSquireOracle = "5a10c3a5-6724-4e5a-ae4f-b27dde12735a"

// TestBrassSquireMovesEquipmentOntoAChosenCreature is the whole card:
// the two heterogeneous clauses (ADR 0065) each land the right kind
// of object, and the Equipment actually ends up attached to the
// creature — not just "the call didn't error".
func TestBrassSquireMovesEquipmentOntoAChosenCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	squire := pushCatalogPermanent(g, me.ID, "Brass Squire", "Artifact Creature — Myr", brassSquireOracle, false)
	sword := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bonesplitter", TypeLine: equipTypeLine,
		OracleID: bonesplitterOracle, Owner: me.ID, Controller: me.ID,
	})
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})

	if err := g.ActivateCatalogAbility(me.ID, squire, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: sword},
			{Kind: game.TargetCard, ID: bear},
		},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	if host := attachmentHostOf(t, g, sword); host.Kind != game.TargetCard || host.ID != bear {
		t.Fatalf("AttachedTo = %+v, want card %s", host, bear)
	}
	// The Equipment's own static (+2/+0) rides along, proof the engine
	// treats this exactly like an ordinary equip.
	if got := effectivePower(t, g, bear); got != 4 {
		t.Errorf("equipped power %d, want 4", got)
	}
}

// TestBrassSquireRejectsAnEquipmentYouDoNotControl pins the "you
// control" half of both clauses at announce (CR 601.2c): an
// opponent's Equipment cannot fill clause 0, whatever creature fills
// clause 1.
func TestBrassSquireRejectsAnEquipmentYouDoNotControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	squire := pushCatalogPermanent(g, me.ID, "Brass Squire", "Artifact Creature — Myr", brassSquireOracle, false)
	theirSword := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bonesplitter", TypeLine: equipTypeLine,
		OracleID: bonesplitterOracle, Owner: opp.ID, Controller: opp.ID,
	})
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})

	if err := g.ActivateCatalogAbility(me.ID, squire, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: theirSword},
			{Kind: game.TargetCard, ID: bear},
		},
	}); err == nil {
		t.Error("an opponent's Equipment should not fill 'target Equipment you control'")
	}
}
