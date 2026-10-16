package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// zone_layers_cards_test.go — #3862: the three cards whose types or
// colours reach objects off the battlefield (game/zone_layers.go). The
// oracle constants are lattice_lord_test.go's and tribal_test.go's.

// zoneLayerCard reads an object wherever it is, after a recompute.
func zoneLayerCard(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	var out game.Card
	found := false
	g.ReadSnapshot(func() {
		out, found = g.LookupCardForEffect(id)
	})
	if !found {
		t.Fatalf("object %s is in no zone", id)
	}
	return out
}

func redCard(owner uuid.UUID, name, typeLine string) game.Card {
	return game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: "{1}{R}",
		Owner: owner, Controller: owner,
	}
}

// Mycosynth Lattice "makes everything, in every zone of the game,
// colorless" (its ruling): a card in a hand, a library, a graveyard,
// exile and the command zone, and a spell on the stack. None of them
// becomes an artifact, which is the permanents' line only. When the
// Lattice leaves, every one of them has its colour back.
func TestMycosynthLatticeMakesCardsAndSpellsInEveryZoneColorless(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	inHand := redCard(me.ID, "Hand Bolt", "Instant")
	inLibrary := redCard(me.ID, "Library Bolt", "Instant")
	inGraveyard := redCard(me.ID, "Graveyard Bolt", "Instant")
	inExile := redCard(me.ID, "Exiled Bolt", "Instant")
	me.Hand.PushTop(inHand)
	me.Library.PushTop(inLibrary)
	me.Graveyard.PushTop(inGraveyard)
	g.Exile.PushTop(inExile)
	commander := redCard(me.ID, "Red Commander", "Legendary Creature — Goblin")
	commander.IsCommander = true
	me.Command.PushTop(commander)

	lattice := pushCatalogPermanent(g, me.ID, "Mycosynth Lattice", "Artifact", mycosynthLatticeOracle, false)
	spell := castCatalogSpell(t, g, "Red Bear", "Creature — Bear", "", nil)
	g.WithWriteLock(func() {
		for i := range g.Stack.Cards {
			if g.Stack.Cards[i].InstanceID == spell {
				g.Stack.Cards[i].ManaCost = "{1}{R}"
			}
		}
		g.BumpLayerVersionForTest()
	})

	for _, tc := range []struct {
		where string
		id    uuid.UUID
	}{
		{"hand", inHand.InstanceID},
		{"library", inLibrary.InstanceID},
		{"graveyard", inGraveyard.InstanceID},
		{"exile", inExile.InstanceID},
		{"command zone", commander.InstanceID},
		{"stack", spell},
	} {
		c := zoneLayerCard(t, g, tc.id)
		if !c.IsColorless() {
			t.Errorf("%s: colours %v under Mycosynth Lattice, want colorless", tc.where, c.EffectiveColors())
		}
		if c.IsArtifact() {
			t.Errorf("%s: a card off the battlefield became an artifact; only permanents do", tc.where)
		}
	}
	if got := game.CommanderIdentityForTest(g, me); len(got) != 1 || got[0] != "R" {
		t.Errorf("commander identity %v, want [R]: colour identity is not colour (CR 903.4)", got)
	}

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(lattice); err != nil {
			t.Fatalf("destroy the Lattice: %v", err)
		}
	})
	for _, id := range []uuid.UUID{inHand.InstanceID, inLibrary.InstanceID, commander.InstanceID, spell} {
		if c := zoneLayerCard(t, g, id); !c.HasColor("R") {
			t.Errorf("%s: colours %v after the Lattice left, want red", c.Name, c.EffectiveColors())
		}
	}
}

// Maskwood Nexus: "creature spells you control and creature cards you
// own that aren't on the battlefield" are every creature type. Not an
// opponent's card, not a noncreature card, and not after it leaves.
func TestMaskwoodNexusMakesYourCreatureCardsAndSpellsEveryType(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	myBear := redCard(me.ID, "My Bear", "Creature — Bear")
	myLibraryBear := redCard(me.ID, "Library Bear", "Creature — Bear")
	myGraveBear := redCard(me.ID, "Graveyard Bear", "Creature — Bear")
	mySorcery := redCard(me.ID, "My Sorcery", "Sorcery")
	theirBear := redCard(opp.ID, "Their Bear", "Creature — Bear")
	me.Hand.PushTop(myBear)
	me.Hand.PushTop(mySorcery)
	me.Library.PushTop(myLibraryBear)
	me.Graveyard.PushTop(myGraveBear)
	opp.Hand.PushTop(theirBear)

	nexus := pushCatalogPermanent(g, me.ID, "Maskwood Nexus", "Artifact", maskwoodNexusOracle, false)
	spell := castCatalogSpell(t, g, "Cast Bear", "Creature — Bear", "", nil)
	g.WithWriteLock(func() { g.BumpLayerVersionForTest() })

	for _, tc := range []struct {
		name string
		id   uuid.UUID
		want bool
	}{
		{"a creature card in your hand", myBear.InstanceID, true},
		{"a creature card in your library", myLibraryBear.InstanceID, true},
		{"a creature card in your graveyard", myGraveBear.InstanceID, true},
		{"a creature spell you control", spell, true},
		{"a noncreature card", mySorcery.InstanceID, false},
		{"an opponent's creature card", theirBear.InstanceID, false},
	} {
		c := zoneLayerCard(t, g, tc.id)
		if got := c.HasSubtype("Elf"); got != tc.want {
			t.Errorf("%s: is an Elf = %v, want %v", tc.name, got, tc.want)
		}
		if tc.want && !c.HasSubtype("Bear") {
			t.Errorf("%s: lost its printed Bear", tc.name)
		}
	}
	// Cavern of Souls' restriction reads the spend context, which reads
	// the card in hand: an Elf creature spell, so Elf-only mana pays.
	if ctx := game.ManaSpendForCast(zoneLayerCard(t, g, myBear.InstanceID)); !ctx.AllCreatureTypes {
		t.Error("the spend context for a creature card in hand is not every creature type")
	}

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(nexus); err != nil {
			t.Fatalf("destroy the Nexus: %v", err)
		}
	})
	for _, id := range []uuid.UUID{myBear.InstanceID, spell} {
		if c := zoneLayerCard(t, g, id); c.HasSubtype("Elf") {
			t.Errorf("%s is still an Elf after the Nexus left", c.Name)
		}
	}
}

// A card drawn while the Nexus is out is a new object (CR 400.7) and is
// every creature type from the moment it is in the hand.
func TestMaskwoodNexusReachesACardDrawnWhileItIsOut(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Maskwood Nexus", "Artifact", maskwoodNexusOracle, false)
	top := redCard(me.ID, "Drawn Bear", "Creature — Bear")
	me.Library.PushTop(top)
	g.WithWriteLock(func() { g.BumpLayerVersionForTest() })
	if !zoneLayerCard(t, g, top.InstanceID).HasSubtype("Goblin") {
		t.Fatal("setup: the library card is not every creature type")
	}
	g.WithWriteLock(func() {
		if err := g.DrawNForEffect(me.ID, 1); err != nil {
			t.Fatalf("draw: %v", err)
		}
	})
	if c := zoneLayerCard(t, g, top.InstanceID); !c.HasSubtype("Goblin") {
		t.Error("the drawn creature card is not every creature type in hand")
	}
}

// Secret Arcade: "permanent spells you control are enchantments in
// addition to their other types" — on the stack, before the spell
// resolves. An instant is not a permanent spell.
func TestSecretArcadeMakesYourPermanentSpellsEnchantments(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	roomsBCast(t, g, me, roomsBCard(me.ID, secretArcadeOracle, "Secret Arcade", "{4}{W}", "Dusty Parlor", "{2}{W}"), 0)
	roomsBSettle(t, g, me)

	creature := castCatalogSpell(t, g, "Cast Bear", "Creature — Bear", "", nil)
	if c := zoneLayerCard(t, g, creature); !c.IsEnchantment() || !c.IsCreature() {
		t.Errorf("a creature spell you control: enchantment %v, creature %v; want both", c.IsEnchantment(), c.IsCreature())
	}
	instant := castCatalogSpell(t, g, "Cast Shock", "Instant", "", nil)
	if zoneLayerCard(t, g, instant).IsEnchantment() {
		t.Error("an instant spell became an enchantment; it is not a permanent spell")
	}
}
