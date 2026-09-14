package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s30_indestructible_wipes_test.go — #470 / #446 at the catalog
// level.
//
// The engine-side cases live in game/indestructible_test.go. These
// are the ones a player would file the bug from: a real Wrath of God,
// cast off a real Avacyn board, resolved through the real priority
// loop. They exist because the gap that shipped was not a missing
// rule — IsIndestructible had been correct for four sprints — it was
// one of two entry points not asking the question, and only an
// end-to-end cast goes through the one that was wrong.
//
// The second and third cases are the half of the fix that is easy to
// forget. Filtering the destruction is not enough on its own: "for
// each creature destroyed this way" reads a count, and Deadly
// Tempest's per-controller tally reads the swept CARDS, so a survivor
// left in either would pay out for a creature still standing.

// pushIndestructibleWipeCreature seeds a creature whose indestructible
// is PRINTED on it, via game.Card.Keywords — the same road
// Darksteel Citadel takes. No oracle ID needed, and it survives a
// layer recompute because printedCharacteristic merges the slice in.
func pushIndestructibleWipeCreature(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   "Creature — Golem",
		Power:      2,
		Toughness:  2,
		Keywords:   []string{"indestructible"},
		Owner:      owner,
		Controller: owner,
	})
	return id
}

// TestWrathDoesNotDestroyAnAvacynBoard is the headline case from
// #470: eight mana of "my permanents cannot be destroyed" versus the
// format's most-played sweeper. Avacyn's own printed indestructible
// and the layer-6 grant she hands the rest of her controller's board
// both have to reach the sweep.
func TestWrathDoesNotDestroyAnAvacynBoard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	avacyn := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Avacyn, Angel of Hope",
		TypeLine:   "Legendary Creature — Angel",
		OracleID:   avacynOracle,
		Power:      8,
		Toughness:  8,
		Owner:      opp.ID,
		Controller: opp.ID,
	})
	protected := pushWipeCreature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := pushWipeCreature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	castCatalogSpell(t, g, "Wrath of God", "Sorcery", wrathOfGodOracle, nil)
	passPriorityAroundTable(t, g)

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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushWipeCreature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	pushWipeCreature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	golem := pushIndestructibleWipeCreature(g, opp.ID, "Darksteel Golem")

	before := me.Life
	castCatalogSpell(t, g, "Fumigate", "Sorcery", fumigateOracle, nil)
	passPriorityAroundTable(t, g)

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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushWipeCreature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	pushWipeCreature(g, opp.ID, "Theirs A", "Creature — Bear", 2, 2)
	pushWipeCreature(g, opp.ID, "Theirs B", "Creature — Bear", 2, 2)
	golem := pushIndestructibleWipeCreature(g, opp.ID, "Darksteel Golem")

	meBefore, oppBefore := me.Life, opp.Life
	castCatalogSpell(t, g, "Deadly Tempest", "Sorcery", deadlyTempestOracle, nil)
	passPriorityAroundTable(t, g)

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
	g := newCatalogGame(t)
	me := g.Seats[0]
	golem := pushIndestructibleWipeCreature(g, me.ID, "Darksteel Golem")

	g.WithWriteLock(func() {
		_ = g.ExileCardsForEffect([]uuid.UUID{golem})
	})

	if g.Battlefield.Contains(golem) {
		t.Error("an indestructible permanent survived being exiled — exile is not destruction")
	}
}
