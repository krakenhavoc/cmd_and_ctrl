package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Loxodon Warhammer: a 7c modify and a 6 grant on one card -----

func TestLoxodonWarhammerPumpsAndGrantsTrampleAndLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	hammer := effectstest.SeedEquipment(g, me.ID, "Loxodon Warhammer", effectstest.WarhammerOracle)

	effectstest.EquipTo(t, g, me.ID, hammer, bear)

	if got := effectstest.EffectivePower(t, g, bear); got != 5 {
		t.Errorf("power %d, want 5", got)
	}
	if got := effectstest.EffectiveToughness(t, g, bear); got != 2 {
		t.Errorf("the Warhammer is +3/+0; toughness %d, want 2", got)
	}
	abilities := effectstest.EffectiveAbilities(t, g, bear)
	for _, want := range []string{"trample", "lifelink"} {
		if !effectstest.ContainsString(abilities, want) {
			t.Errorf("abilities %v missing %q", abilities, want)
		}
	}
}

// The pump and the grant move together, because both are scoped by
// the same AttachedToSource predicate re-read every recompute.
func TestLoxodonWarhammerTakesEverythingWithItWhenItMoves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	first, second := effectstest.SeedBear(g, me.ID), effectstest.SeedBear(g, me.ID)
	hammer := effectstest.SeedEquipment(g, me.ID, "Loxodon Warhammer", effectstest.WarhammerOracle)

	effectstest.EquipTo(t, g, me.ID, hammer, first)
	effectstest.EquipTo(t, g, me.ID, hammer, second)

	if got := effectstest.EffectivePower(t, g, first); got != 2 {
		t.Errorf("the first creature kept the pump: power %d, want 2", got)
	}
	if abilities := effectstest.EffectiveAbilities(t, g, first); effectstest.ContainsString(abilities, "trample") {
		t.Errorf("the first creature kept trample: %v", abilities)
	}
	if got := effectstest.EffectivePower(t, g, second); got != 5 {
		t.Errorf("the second creature's power %d, want 5", got)
	}
	if abilities := effectstest.EffectiveAbilities(t, g, second); !effectstest.ContainsString(abilities, "lifelink") {
		t.Errorf("the second creature is missing lifelink: %v", abilities)
	}
}

// --- Basilisk Collar: grant and nothing else ----------------------

func TestBasiliskCollarGrantsDeathtouchAndLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	collar := effectstest.SeedEquipment(g, me.ID, "Basilisk Collar", effectstest.BasiliskCollarOracle)

	effectstest.EquipTo(t, g, me.ID, collar, bear)

	abilities := effectstest.EffectiveAbilities(t, g, bear)
	for _, want := range []string{"deathtouch", "lifelink"} {
		if !effectstest.ContainsString(abilities, want) {
			t.Errorf("abilities %v missing %q", abilities, want)
		}
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Errorf("the Collar has no stat line; power %d, want 2", got)
	}
}

// --- Darksteel Plate: the grant ADR 0036 decision 10 had to cut ---

func TestDarksteelPlateKeepsTheEquippedCreatureAlive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	plate := effectstest.SeedEquipment(g, me.ID, "Darksteel Plate", effectstest.DarksteelPlateOracle)

	effectstest.EquipTo(t, g, me.ID, plate, bear)
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, bear), "indestructible") {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	plate := effectstest.SeedEquipment(g, me.ID, "Darksteel Plate", effectstest.DarksteelPlateOracle)

	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, plate), "indestructible") {
		t.Fatal("the Plate prints indestructible on itself")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(plate) })
	if !g.Battlefield.Contains(plate) {
		t.Error("the Plate should have survived its own destruction")
	}
}

// --- Mask of Memory: the third attachment combat-damage trigger ---

func TestMaskOfMemoryDrawsTwoAndDiscardsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	mask := effectstest.SeedEquipment(g, me.ID, "Mask of Memory", effectstest.MaskOfMemoryOracle)
	effectstest.EquipTo(t, g, me.ID, mask, bear)
	before := me.Hand.Size()

	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)

	// "You MAY draw two cards" is asked at resolution (#796) before
	// anything is drawn.
	if got := me.Hand.Size(); got != before {
		t.Errorf("hand %d -> %d, want no draw before the question is answered", before, got)
	}
	effectstest.AnswerMayChoice(t, g, me.ID, true)

	// The draw lands on the accept branch; the linked discard is a
	// prompt over the post-draw hand and the controller picks (#651).
	if got := me.Hand.Size(); got != before+2 {
		t.Errorf("hand %d -> %d, want +2 before the discard is answered", before, got)
	}
	if got := effectstest.DiscardOwed(g, me.ID); got != 1 {
		t.Fatalf("the linked discard owes %d, want 1", got)
	}
	effectstest.DiscardFromHand(t, g, me.ID)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	carrier, other := effectstest.SeedBear(g, me.ID), effectstest.SeedBear(g, me.ID)
	mask := effectstest.SeedEquipment(g, me.ID, "Mask of Memory", effectstest.MaskOfMemoryOracle)
	effectstest.EquipTo(t, g, me.ID, mask, carrier)
	before := me.Hand.Size()

	effectstest.DealCombatDamageToPlayer(g, other, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != before {
		t.Errorf("hand %d -> %d; the unequipped creature must not trigger the Mask", before, got)
	}
}

// --- Whispersilk Cloak: shroud, honestly, and nothing else --------

func TestWhispersilkCloakGrantsShroudToBothSides(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	cloak := effectstest.SeedEquipment(g, me.ID, "Whispersilk Cloak", effectstest.WhispersilkOracle)

	effectstest.EquipTo(t, g, me.ID, cloak, bear)

	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, bear), "shroud") {
		t.Fatal("the Cloak grants shroud")
	}
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.InstanceID != bear {
				continue
			}
			if game.CanBeTargetedBy(c, game.ZoneBattlefield, game.SourceChooser(me.ID)) {
				t.Error("shroud stops the controller too (CR 702.18a)")
			}
			if game.CanBeTargetedBy(c, game.ZoneBattlefield, game.SourceChooser(opp.ID)) {
				t.Error("shroud stops an opponent")
			}
		}
	})
}

// --- Colossus Hammer: the first static that takes something away --

func TestColossusHammerPumpsAndStripsFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	flier := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
		Keywords: []string{"flying"},
	})
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, flier), "flying") {
		t.Fatal("setup: the Bird should print flying")
	}
	hammer := effectstest.SeedEquipment(g, me.ID, "Colossus Hammer", effectstest.ColossusHammerOracle)

	effectstest.EquipTo(t, g, me.ID, hammer, flier)

	if got := effectstest.EffectivePower(t, g, flier); got != 12 {
		t.Errorf("power %d, want 12", got)
	}
	if got := effectstest.EffectiveToughness(t, g, flier); got != 12 {
		t.Errorf("toughness %d, want 12", got)
	}
	if abilities := effectstest.EffectiveAbilities(t, g, flier); effectstest.ContainsString(abilities, "flying") {
		t.Errorf("the Hammer strips flying; abilities %v", abilities)
	}
}

// The removal is scoped by the attachment like every other
// attachment static — unequip and the flying comes back, because
// nothing was ever written to the card.
func TestColossusHammerGivesFlyingBackWhenItMovesOn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	flier := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
		Keywords: []string{"flying"},
	})
	ground := effectstest.SeedBear(g, me.ID)
	hammer := effectstest.SeedEquipment(g, me.ID, "Colossus Hammer", effectstest.ColossusHammerOracle)

	effectstest.EquipTo(t, g, me.ID, hammer, flier)
	effectstest.EquipTo(t, g, me.ID, hammer, ground)

	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, flier), "flying") {
		t.Error("the Bird should have its flying back once the Hammer moved")
	}
	if got := effectstest.EffectivePower(t, g, ground); got != 12 {
		t.Errorf("the new carrier's power %d, want 12", got)
	}
}

// --- Battle Mastery: an Aura that is only a grant -----------------

func TestBattleMasteryGrantsDoubleStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.SeedBear(g, me.ID)

	aura := effectstest.CastCatalogSpell(t, g, "Battle Mastery", effectstest.AuraTypeLine, effectstest.BattleMasteryOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)

	if host := effectstest.AttachmentHostOf(t, g, aura); host.Kind != game.TargetCard || host.ID != bear {
		t.Fatalf("AttachedTo = %+v, want card %s", host, bear)
	}
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, bear), "double strike") {
		t.Error("the enchanted creature should have double strike")
	}
}

// --- Control Magic: what a SECOND control-changer proves ----------

func TestControlMagicStealsTheCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.SeedBear(g, opp.ID)

	effectstest.CastCatalogSpell(t, g, "Control Magic", effectstest.AuraTypeLine, effectstest.ControlMagicOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.ControllerOf(t, g, theirs); got != me.ID {
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
	g := effectstest.NewCatalogGame(t)
	first, second, owner := g.Seats[0], g.Seats[1], g.Seats[2]
	theirs := effectstest.SeedBear(g, owner.ID)

	early := effectstest.CastCatalogSpell(t, g, "Control Magic", effectstest.AuraTypeLine, effectstest.ControlMagicOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, theirs); got != first.ID {
		t.Fatalf("after the first steal: controller %s, want %s", got, first.ID)
	}

	// The second Aura is cast by a different seat on a later turn, so
	// its layer-2 effect has the later timestamp.
	effectstest.AdvanceToPrecombatMainOf(t, g, 1)
	late := effectstest.CastCatalogSpell(t, g, "Mind Control", effectstest.AuraTypeLine, effectstest.MindControlOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, theirs); got != second.ID {
		t.Fatalf("the later Aura should win: controller %s, want %s", got, second.ID)
	}

	// Remove the later one and the EARLIER one takes over — not the
	// creature's owner.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(late) })
	if got := effectstest.ControllerOf(t, g, theirs); got != first.ID {
		t.Errorf("controller %s after the later Aura died, want the earlier Aura's controller %s", got, first.ID)
	}

	// Remove the earlier one too and control finally goes home.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(early) })
	if got := effectstest.ControllerOf(t, g, theirs); got != owner.ID {
		t.Errorf("controller %s with no Aura left, want the baseline %s", got, owner.ID)
	}
}

// --- CR 704.5m's "attached to nothing" branch, through a card -----

// An Aura put onto the battlefield by something that does not say
// "attached to" — Brilliant Restoration, Carmen — enters with no
// host. CR 704.5m puts it in its owner's graveyard on the next
// state-based-action pass. Before this branch shipped it sat on the
// battlefield permanently, which is a board state no sequence of
// legal plays can reach.
func TestACataloguedAuraWithNoHostGoesToTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	_ = effectstest.SeedBear(g, me.ID)
	aura := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Battle Mastery", TypeLine: effectstest.AuraTypeLine,
		OracleID: effectstest.BattleMasteryOracle, Owner: me.ID, Controller: me.ID,
	})

	// State-based actions run at every step boundary.
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}

	if g.Battlefield.Contains(aura) {
		t.Fatal("CR 704.5m: an Aura attached to nothing must leave the battlefield")
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	collar := effectstest.SeedEquipment(g, me.ID, "Basilisk Collar", effectstest.BasiliskCollarOracle)

	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}

	if !g.Battlefield.Contains(collar) {
		t.Error("CR 704.5n has no \"attached to nothing\" clause")
	}
}

// TestMaskOfMemoryCanDeclineTheDraw is the #796 half: the optional
// draw is a real question, and declining draws nothing and discards
// nothing — the discard is LINKED to the draw (CR 607.2), so a branch
// that skipped the draw must not still take the card.
func TestMaskOfMemoryCanDeclineTheDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	mask := effectstest.SeedEquipment(g, me.ID, "Mask of Memory", effectstest.MaskOfMemoryOracle)
	effectstest.EquipTo(t, g, me.ID, mask, bear)
	before, graves := me.Hand.Size(), me.Graveyard.Size()

	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)

	if me.Hand.Size() != before {
		t.Errorf("hand %d -> %d, want unchanged on the decline", before, me.Hand.Size())
	}
	if me.Graveyard.Size() != graves {
		t.Error("the discard is linked to the draw — declining discards nothing")
	}
	if effectstest.DiscardOwed(g, me.ID) != 0 {
		t.Error("declining owes no discard")
	}
}
