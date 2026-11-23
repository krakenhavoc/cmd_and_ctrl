package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestWrathDoesNotDestroyAnAvacynBoard is the headline case from
// #470: eight mana of "my permanents cannot be destroyed" versus the
// format's most-played sweeper. Avacyn's own printed indestructible
// and the layer-6 grant she hands the rest of her controller's board
// both have to reach the sweep.
func TestWrathDoesNotDestroyAnAvacynBoard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	avacyn := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Avacyn, Angel of Hope",
		TypeLine:   "Legendary Creature — Angel",
		OracleID:   effectstest.AvacynOracle,
		Power:      8,
		Toughness:  8,
		Owner:      opp.ID,
		Controller: opp.ID,
	})
	protected := effectstest.PushWipeCreature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.PushWipeCreature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	effectstest.CastCatalogSpell(t, g, "Wrath of God", "Sorcery", effectstest.WrathOfGodOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(avacyn) {
		t.Error("Avacyn was destroyed by a board wipe (#470 / #446)")
	}
	if !g.Battlefield.Contains(protected) {
		t.Error("a creature with Avacyn's granted indestructible was destroyed by a board wipe")
	}
	if g.Battlefield.Contains(mine) {
		t.Error("an unprotected creature survived the wrath — control case broken")
	}
}

// TestFumigateCountsOnlyTheCreaturesItDestroyed pins the count.
// "You gain 1 life for each creature destroyed this way" — the golem
// that shrugged the wipe off was not destroyed and must not be paid
// for.
func TestFumigateCountsOnlyTheCreaturesItDestroyed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushWipeCreature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.PushWipeCreature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	golem := effectstest.PushIndestructibleWipeCreature(g, opp.ID, "Darksteel Golem")

	before := me.Life
	effectstest.CastCatalogSpell(t, g, "Fumigate", "Sorcery", effectstest.FumigateOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(golem) {
		t.Fatal("Fumigate destroyed an indestructible creature")
	}
	if want := before + 2; me.Life != want {
		t.Errorf("life %d -> %d, want %d (two creatures died; the indestructible one did not)", before, me.Life, want)
	}
}

// TestDeadlyTempestDoesNotChargeForSurvivors is the same rule through
// the other door. Deadly Tempest's Then clause reads the swept CARDS
// rather than the count, so filtering only the destruction would
// still have billed its controller for a creature that is visibly
// still on the battlefield.
func TestDeadlyTempestDoesNotChargeForSurvivors(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushWipeCreature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.PushWipeCreature(g, opp.ID, "Theirs A", "Creature — Bear", 2, 2)
	effectstest.PushWipeCreature(g, opp.ID, "Theirs B", "Creature — Bear", 2, 2)
	golem := effectstest.PushIndestructibleWipeCreature(g, opp.ID, "Darksteel Golem")

	meBefore, oppBefore := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "Deadly Tempest", "Sorcery", effectstest.DeadlyTempestOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(golem) {
		t.Fatal("Deadly Tempest destroyed an indestructible creature")
	}
	if want := meBefore - 1; me.Life != want {
		t.Errorf("caster life %d -> %d, want %d", meBefore, me.Life, want)
	}
	if want := oppBefore - 2; opp.Life != want {
		t.Errorf("opponent life %d -> %d, want %d (two of their three creatures died)", oppBefore, opp.Life, want)
	}
}

// TestExileWipeStillGetsThroughIndestructible is the negative space,
// and the reason Merciless Eviction and Farewell cost what they cost:
// exile is not destruction, so CR 702.12b never applies and
// ExileAllMatching deliberately has no filter.
func TestExileWipeStillGetsThroughIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	golem := effectstest.PushIndestructibleWipeCreature(g, me.ID, "Darksteel Golem")

	g.WithWriteLock(func() {
		_ = g.ExileCardsForEffect([]uuid.UUID{golem})
	})

	if g.Battlefield.Contains(golem) {
		t.Error("an indestructible permanent survived being exiled — exile is not destruction")
	}
	// The destination IS the thesis of this test: a route that sent
	// the Golem to a graveyard instead would pass the check above.
	if !g.Exile.Contains(golem) {
		t.Error("the indestructible permanent left the battlefield but was not exiled")
	}
}
