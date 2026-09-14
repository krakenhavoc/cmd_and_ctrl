package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attachments_batch2_test.go covers the S24 tail: six more Equipment
// and two more Auras on the surface #379 built, plus the CR 704.5n
// branch that shipped with them (an Aura attached to NOTHING goes to
// its owner's graveyard) seen through a real catalogued card rather
// than through a fixture.
//
// The shared helpers — advanceToMain, equipTo, attachmentHostOf,
// effectiveAbilities, dealCombatDamageToPlayer — all live in
// attachments_test.go and its siblings.

const (
	warhammerOracle      = "dba35ac5-7ad3-488a-a006-6b9a1d54eea5"
	basiliskCollarOracle = "f5f4dd28-f4ae-4d39-b9b8-6ebfd63c93fe"
	darksteelPlateOracle = "b5b4cf54-ed5e-42d0-9d98-5fec76b0b0b8"
	maskOfMemoryOracle   = "d6b2c998-a226-426c-a40d-6e6007041bfe"
	whispersilkOracle    = "9ad4f730-a18e-4a7c-a468-a926c718c741"
	colossusHammerOracle = "8ec03b88-8d3a-4a32-8b7c-7da59b0c03d0"
	battleMasteryOracle  = "a06c7c1a-8534-40fb-8bb8-59f8f2530567"
	controlMagicOracle   = "cd0d7141-46d2-4aa3-bc77-6b3b4513803e"
)

// seedBear puts a vanilla 2/2 on the battlefield under `owner`.
func seedBear(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
}

// seedEquipment puts an Equipment on the battlefield unattached,
// which is where every Equipment starts.
func seedEquipment(g *game.Game, owner uuid.UUID, name, oracle string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: equipTypeLine,
		OracleID: oracle, Owner: owner, Controller: owner,
	})
}

// --- Loxodon Warhammer: a 7c modify and a 6 grant on one card -----

func TestLoxodonWarhammerPumpsAndGrantsTrampleAndLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	hammer := seedEquipment(g, me.ID, "Loxodon Warhammer", warhammerOracle)

	equipTo(t, g, me.ID, hammer, bear)

	if got := effectivePower(t, g, bear); got != 5 {
		t.Errorf("power %d, want 5", got)
	}
	if got := effectiveToughness(t, g, bear); got != 2 {
		t.Errorf("the Warhammer is +3/+0; toughness %d, want 2", got)
	}
	abilities := effectiveAbilities(t, g, bear)
	for _, want := range []string{"trample", "lifelink"} {
		if !containsString(abilities, want) {
			t.Errorf("abilities %v missing %q", abilities, want)
		}
	}
}

// The pump and the grant move together, because both are scoped by
// the same AttachedToSource predicate re-read every recompute.
func TestLoxodonWarhammerTakesEverythingWithItWhenItMoves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	first, second := seedBear(g, me.ID), seedBear(g, me.ID)
	hammer := seedEquipment(g, me.ID, "Loxodon Warhammer", warhammerOracle)

	equipTo(t, g, me.ID, hammer, first)
	equipTo(t, g, me.ID, hammer, second)

	if got := effectivePower(t, g, first); got != 2 {
		t.Errorf("the first creature kept the pump: power %d, want 2", got)
	}
	if abilities := effectiveAbilities(t, g, first); containsString(abilities, "trample") {
		t.Errorf("the first creature kept trample: %v", abilities)
	}
	if got := effectivePower(t, g, second); got != 5 {
		t.Errorf("the second creature's power %d, want 5", got)
	}
	if abilities := effectiveAbilities(t, g, second); !containsString(abilities, "lifelink") {
		t.Errorf("the second creature is missing lifelink: %v", abilities)
	}
}

// --- Basilisk Collar: grant and nothing else ----------------------

func TestBasiliskCollarGrantsDeathtouchAndLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	collar := seedEquipment(g, me.ID, "Basilisk Collar", basiliskCollarOracle)

	equipTo(t, g, me.ID, collar, bear)

	abilities := effectiveAbilities(t, g, bear)
	for _, want := range []string{"deathtouch", "lifelink"} {
		if !containsString(abilities, want) {
			t.Errorf("abilities %v missing %q", abilities, want)
		}
	}
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("the Collar has no stat line; power %d, want 2", got)
	}
}

// --- Darksteel Plate: the grant ADR 0036 decision 10 had to cut ---

func TestDarksteelPlateKeepsTheEquippedCreatureAlive(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	plate := seedEquipment(g, me.ID, "Darksteel Plate", darksteelPlateOracle)

	equipTo(t, g, me.ID, plate, bear)
	if !containsString(effectiveAbilities(t, g, bear), "indestructible") {
		t.Fatal("the equipped creature should have indestructible")
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })

	if !g.Battlefield.Contains(bear) {
		t.Error("an indestructible creature is not destroyed (CR 702.12b)")
	}
}

// The Plate's OWN indestructible is a printed keyword, so it is true
// of the Equipment whether or not it is attached to anything — which
// is what makes it survive the wipe that kills its carrier.
func TestDarksteelPlateIsItselfIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	plate := seedEquipment(g, me.ID, "Darksteel Plate", darksteelPlateOracle)

	if !containsString(effectiveAbilities(t, g, plate), "indestructible") {
		t.Fatal("the Plate prints indestructible on itself")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(plate) })
	if !g.Battlefield.Contains(plate) {
		t.Error("the Plate should have survived its own destruction")
	}
}

// --- Mask of Memory: the third attachment combat-damage trigger ---

func TestMaskOfMemoryDrawsTwoAndDiscardsOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	mask := seedEquipment(g, me.ID, "Mask of Memory", maskOfMemoryOracle)
	equipTo(t, g, me.ID, mask, bear)
	before := me.Hand.Size()

	dealCombatDamageToPlayer(g, bear, opp.ID, 2)
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d -> %d, want +1 (draw two, discard one)", before, got)
	}
	if got := me.Graveyard.Size(); got == 0 {
		t.Error("the linked discard should have put a card in the graveyard")
	}
}

// The trigger is scoped to the EQUIPPED creature, not to every
// creature its controller attacks with.
func TestMaskOfMemoryIgnoresAnUnequippedAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	carrier, other := seedBear(g, me.ID), seedBear(g, me.ID)
	mask := seedEquipment(g, me.ID, "Mask of Memory", maskOfMemoryOracle)
	equipTo(t, g, me.ID, mask, carrier)
	before := me.Hand.Size()

	dealCombatDamageToPlayer(g, other, opp.ID, 2)
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != before {
		t.Errorf("hand %d -> %d; the unequipped creature must not trigger the Mask", before, got)
	}
}

// --- Whispersilk Cloak: shroud, honestly, and nothing else --------

func TestWhispersilkCloakGrantsShroudToBothSides(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	cloak := seedEquipment(g, me.ID, "Whispersilk Cloak", whispersilkOracle)

	equipTo(t, g, me.ID, cloak, bear)

	if !containsString(effectiveAbilities(t, g, bear), "shroud") {
		t.Fatal("the Cloak grants shroud")
	}
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.InstanceID != bear {
				continue
			}
			if game.CanBeTargetedBy(c, game.ZoneBattlefield, me.ID) {
				t.Error("shroud stops the controller too (CR 702.18a)")
			}
			if game.CanBeTargetedBy(c, game.ZoneBattlefield, opp.ID) {
				t.Error("shroud stops an opponent")
			}
		}
	})
}

// --- Colossus Hammer: the first static that takes something away --

func TestColossusHammerPumpsAndStripsFlying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	flier := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
		Keywords: []string{"flying"},
	})
	if !containsString(effectiveAbilities(t, g, flier), "flying") {
		t.Fatal("setup: the Bird should print flying")
	}
	hammer := seedEquipment(g, me.ID, "Colossus Hammer", colossusHammerOracle)

	equipTo(t, g, me.ID, hammer, flier)

	if got := effectivePower(t, g, flier); got != 12 {
		t.Errorf("power %d, want 12", got)
	}
	if got := effectiveToughness(t, g, flier); got != 12 {
		t.Errorf("toughness %d, want 12", got)
	}
	if abilities := effectiveAbilities(t, g, flier); containsString(abilities, "flying") {
		t.Errorf("the Hammer strips flying; abilities %v", abilities)
	}
}

// The removal is scoped by the attachment like every other
// attachment static — unequip and the flying comes back, because
// nothing was ever written to the card.
func TestColossusHammerGivesFlyingBackWhenItMovesOn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	flier := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
		Keywords: []string{"flying"},
	})
	ground := seedBear(g, me.ID)
	hammer := seedEquipment(g, me.ID, "Colossus Hammer", colossusHammerOracle)

	equipTo(t, g, me.ID, hammer, flier)
	equipTo(t, g, me.ID, hammer, ground)

	if !containsString(effectiveAbilities(t, g, flier), "flying") {
		t.Error("the Bird should have its flying back once the Hammer moved")
	}
	if got := effectivePower(t, g, ground); got != 12 {
		t.Errorf("the new carrier's power %d, want 12", got)
	}
}

// --- Battle Mastery: an Aura that is only a grant -----------------

func TestBattleMasteryGrantsDoubleStrike(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := seedBear(g, me.ID)

	aura := castCatalogSpell(t, g, "Battle Mastery", auraTypeLine, battleMasteryOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)

	if host := attachmentHostOf(t, g, aura); host.Kind != game.TargetCard || host.ID != bear {
		t.Fatalf("AttachedTo = %+v, want card %s", host, bear)
	}
	if !containsString(effectiveAbilities(t, g, bear), "double strike") {
		t.Error("the enchanted creature should have double strike")
	}
}

// --- Control Magic: what a SECOND control-changer proves ----------

func TestControlMagicStealsTheCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := seedBear(g, opp.ID)

	castCatalogSpell(t, g, "Control Magic", auraTypeLine, controlMagicOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, g)

	if got := controllerOf(t, g, theirs); got != me.ID {
		t.Errorf("controller %s, want the Aura's controller %s", got, me.ID)
	}
}

// CR 613.7 — two layer-2 effects on one creature sort by timestamp
// and the LATER one wins. When it goes away the earlier one takes
// over, rather than control going home to the original controller.
// Neither card file says any of this; it falls out of the recompute's
// sort, and until there were two control-changers nothing could pin
// it.
func TestTwoControlAurasSortByTimestamp(t *testing.T) {
	g := newCatalogGame(t)
	first, second, owner := g.Seats[0], g.Seats[1], g.Seats[2]
	theirs := seedBear(g, owner.ID)

	early := castCatalogSpell(t, g, "Control Magic", auraTypeLine, controlMagicOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, g)
	if got := controllerOf(t, g, theirs); got != first.ID {
		t.Fatalf("after the first steal: controller %s, want %s", got, first.ID)
	}

	// The second Aura is cast by a different seat on a later turn, so
	// its layer-2 effect has the later timestamp.
	advanceToPrecombatMainOf(t, g, 1)
	late := castCatalogSpell(t, g, "Mind Control", auraTypeLine, mindControlOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, g)
	if got := controllerOf(t, g, theirs); got != second.ID {
		t.Fatalf("the later Aura should win: controller %s, want %s", got, second.ID)
	}

	// Remove the later one and the EARLIER one takes over — not the
	// creature's owner.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(late) })
	if got := controllerOf(t, g, theirs); got != first.ID {
		t.Errorf("controller %s after the later Aura died, want the earlier Aura's controller %s", got, first.ID)
	}

	// Remove the earlier one too and control finally goes home.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(early) })
	if got := controllerOf(t, g, theirs); got != owner.ID {
		t.Errorf("controller %s with no Aura left, want the baseline %s", got, owner.ID)
	}
}

// --- CR 704.5n's "attached to nothing" branch, through a card -----

// An Aura put onto the battlefield by something that does not say
// "attached to" — Brilliant Restoration, Carmen — enters with no
// host. CR 704.5n puts it in its owner's graveyard on the next
// state-based-action pass. Before this branch shipped it sat on the
// battlefield permanently, which is a board state no sequence of
// legal plays can reach.
func TestACataloguedAuraWithNoHostGoesToTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	_ = seedBear(g, me.ID)
	aura := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Battle Mastery", TypeLine: auraTypeLine,
		OracleID: battleMasteryOracle, Owner: me.ID, Controller: me.ID,
	})

	// State-based actions run at every step boundary.
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}

	if g.Battlefield.Contains(aura) {
		t.Fatal("CR 704.5n: an Aura attached to nothing must leave the battlefield")
	}
	found := false
	for _, c := range me.Graveyard.Cards {
		if c.InstanceID == aura {
			found = true
		}
	}
	if !found {
		t.Error("the Aura should be in its owner's graveyard")
	}
}

// The same rule does not touch Equipment: an unattached Equipment is
// an ordinary artifact, which is where every Equipment starts.
func TestAnUnequippedEquipmentStaysOnTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	collar := seedEquipment(g, me.ID, "Basilisk Collar", basiliskCollarOracle)

	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}

	if !g.Battlefield.Contains(collar) {
		t.Error("CR 704.5m has no \"attached to nothing\" clause")
	}
}
