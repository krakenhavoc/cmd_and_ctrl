package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// kuja_b_crystal_test.go — The Darkness Crystal (#3213, slice kuja-b).

const kbDarknessCrystalOracle = "660f370a-b25b-4ea5-b765-7abf96899c1c"

func kbCrystal(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Darkness Crystal", TypeLine: "Legendary Artifact",
		OracleID: kbDarknessCrystalOracle, Owner: owner, Controller: owner,
	})
}

// A nontoken creature an opponent controls is exiled instead of dying,
// linked to the Crystal, and its controller's opponent gains 2.
func TestDarknessCrystalExilesAnOpponentsDyingCreatureAndGainsLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crystal := kbCrystal(g, me.ID)
	bear := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	life := me.Life

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	passPriorityAroundTable(t, g)

	if opp.Graveyard.Contains(bear) {
		t.Fatal("the creature died instead of being exiled")
	}
	if !g.Exile.Contains(bear) {
		t.Fatal("the creature was not exiled")
	}
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
	var link game.PermissionCardRef
	for _, c := range g.Exile.Cards {
		if c.InstanceID == bear {
			link = c.ExiledWith
		}
	}
	if link.ID != crystal {
		t.Errorf("exiled-with link = %v, want the Crystal %v", link.ID, crystal)
	}
}

// Your own creature dies normally and a token is not exiled.
func TestDarknessCrystalLeavesYourCreaturesAndTokensAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kbCrystal(g, me.ID)
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	token := pushToken(g, opp.ID, TreasureToken())
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Token", TypeLine: "Token Creature — Goblin",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})
	_ = token
	life := me.Life

	g.WithWriteLock(func() {
		_ = g.DestroyPermanentForEffect(mine)
		_ = g.DestroyPermanentForEffect(theirs)
	})
	passPriorityAroundTable(t, g)

	if !me.Graveyard.Contains(mine) {
		t.Error("my own creature should die normally")
	}
	if g.Exile.Contains(theirs) {
		t.Error("a token should not be exiled by the replacement")
	}
	if me.Life != life {
		t.Errorf("life %d -> %d, want unchanged", life, me.Life)
	}
}

// The ability brings an exiled creature back tapped, under the
// Crystal's controller, with two extra +1/+1 counters; a creature
// exiled some other way is not offered.
func TestDarknessCrystalReturnsAnExiledCreatureTappedWithCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crystal := kbCrystal(g, me.ID)
	bear := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	passPriorityAroundTable(t, g)
	// An unrelated exiled creature card: no link, so never offered.
	stranger := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: stranger, Name: "Stranger", TypeLine: "Creature — Test",
		Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID})

	g.WithWriteLock(func() { _ = g.AddManaForEffect(me.ID, uuid.Nil, "{4}{B}{B}") })
	if err := g.ActivateCatalogAbility(me.ID, crystal, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)

	var prompt *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards && c.Chooser == me.ID {
			prompt = c
		}
	}
	if prompt == nil {
		t.Fatal("no choice for the exiled creature")
	}
	if len(prompt.ChooseCards) != 1 || prompt.ChooseCards[0] != bear {
		t.Fatalf("offered %v, want only the Crystal's own exile", prompt.ChooseCards)
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)

	var back *game.Card
	for i := range g.Battlefield.Cards {
		if c := g.Battlefield.Cards[i]; c.Name == "Their Bear" {
			back = &g.Battlefield.Cards[i]
		}
	}
	if back == nil {
		t.Fatal("the creature did not return")
	}
	if back.Controller != me.ID || !back.Tapped {
		t.Errorf("controller %v tapped %v, want me and tapped", back.Controller, back.Tapped)
	}
	if got := counterCount(g, back.InstanceID, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
	if !g.Exile.Contains(stranger) {
		t.Error("the unlinked creature should still be in exile")
	}
}

// Black spells you cast cost {1} less; others do not.
func TestDarknessCrystalDiscountsBlackSpellsYouCast(t *testing.T) {
	seat := uuid.New()
	crystal := game.Card{InstanceID: uuid.New(), Name: "The Darkness Crystal", TypeLine: "Legendary Artifact",
		OracleID: kbDarknessCrystalOracle, Owner: seat, Controller: seat}
	mods := game.CostModifiersForCard(crystal)
	if len(mods) != 1 {
		t.Fatalf("want one cost modifier, got %d", len(mods))
	}
	m := mods[0]
	q := game.CostQuery{Controller: seat, Source: crystal,
		Card: game.Card{Name: "Dark", TypeLine: "Instant", Colors: []string{"B"}}}
	if !m.AppliesTo(q) || m.Amount(q) != 1 {
		t.Error("a black spell you cast should cost {1} less")
	}
	q.Card.Colors = []string{"R"}
	if m.AppliesTo(q) {
		t.Error("a red spell must not be discounted")
	}
}
