package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// granted_layer6_statics_cards_test.go — #3784: a static a layer-6 grant
// gives a permanent may itself be a layer-6 static (ADR 0093 amendment
// 2026-10-15). Dancer's Chakrams gives the equipped creature "Other
// commanders you control get +2/+2 and have lifelink"; each Kaldheim
// Rune gives an enchanted Equipment "Equipped creature has <keyword>";
// Giant's Amulet gives the equipped creature "This creature has hexproof
// as long as it's untapped." Every test drives the real equip, cast and
// layer paths and asserts what a player sees. The engine contract is
// pinned without cards in internal/game/granted_statics_test.go.

const (
	dancersChakramsOracle = "7465d02f-3eee-483c-8de3-404381aa21eb"
	giantsAmuletOracle    = "016a6ec3-f4f6-4b47-9931-e0b14a25bd5e"
)

// seedCommander puts a 3/3 commander creature on the battlefield.
func gl6SeedCommander(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Legendary Creature — Human",
		Power: 3, Toughness: 3, Owner: owner, Controller: owner, IsCommander: true,
	})
}

// loseAllAbilities pins a "loses all abilities" record to a permanent,
// with a timestamp later than everything already on the table.
func gl6LoseAllAbilities(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(id),
			[]game.Mod{game.LoseAllAbilitiesMod()}, game.IndefiniteDuration(), "test — loses all abilities")
	})
}

// attachForTest attaches an Equipment that has no equip ability of its
// own (the fixture under a Rune) to a creature.
func gl6Attach(t *testing.T, g *game.Game, equipment, creature uuid.UUID) {
	t.Helper()
	var err error
	g.WithWriteLock(func() {
		err = g.AttachForEffect(equipment, game.TargetRef{Kind: game.TargetCard, ID: creature})
	})
	if err != nil {
		t.Fatalf("attach: %v", err)
	}
	if !attachedTo(t, g, equipment, creature) {
		t.Fatal("the Equipment did not attach")
	}
}

func gl6HasKeyword(t *testing.T, g *game.Game, id uuid.UUID, kw string) bool {
	t.Helper()
	return containsString(effectiveAbilities(t, g, id), kw)
}

func TestDancersChakramsBuffsOtherCommandersYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	mine := gl6SeedCommander(g, me.ID, "My Commander")
	theirs := gl6SeedCommander(g, opp.ID, "Their Commander")
	chakrams := seedEquipment(g, me.ID, "Dancer's Chakrams", dancersChakramsOracle)
	equipTo(t, g, me.ID, chakrams, bear)

	// The equipped creature: +2/+2, lifelink, a Performer.
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("equipped bear is %d/%d, want 4/4", p, tough)
	}
	if !gl6HasKeyword(t, g, bear, "lifelink") {
		t.Error("the equipped creature has no lifelink")
	}
	if !containsString(effectiveSubtypes(t, g, bear), "Performer") {
		t.Errorf("subtypes %v, want a Performer", effectiveSubtypes(t, g, bear))
	}
	// The quoted ability: your other commanders.
	if p, tough := effectivePower(t, g, mine), effectiveToughness(t, g, mine); p != 5 || tough != 5 {
		t.Errorf("your commander is %d/%d, want 5/5", p, tough)
	}
	if !gl6HasKeyword(t, g, mine, "lifelink") {
		t.Error("your commander has no lifelink")
	}
	// Not an opponent's.
	if p := effectivePower(t, g, theirs); p != 3 {
		t.Errorf("their commander's power is %d, want 3", p)
	}
	if gl6HasKeyword(t, g, theirs, "lifelink") {
		t.Error("an opponent's commander got lifelink")
	}
}

// "Other": a commander wearing the Chakrams gets the Equipment's own
// +2/+2 once, not the commanders' buff on top.
func TestDancersChakramsOnACommanderBuffsTheOtherOneOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	first := gl6SeedCommander(g, me.ID, "First Commander")
	second := gl6SeedCommander(g, me.ID, "Second Commander")
	chakrams := seedEquipment(g, me.ID, "Dancer's Chakrams", dancersChakramsOracle)
	equipTo(t, g, me.ID, chakrams, first)
	if p := effectivePower(t, g, first); p != 5 {
		t.Errorf("the equipped commander's power is %d, want 5 (+2/+2 once)", p)
	}
	if p := effectivePower(t, g, second); p != 5 || !gl6HasKeyword(t, g, second, "lifelink") {
		t.Errorf("the other commander's power is %d (lifelink %v), want 5 with lifelink", p, gl6HasKeyword(t, g, second, "lifelink"))
	}
}

// The quoted ability is the CREATURE's: when the equipped creature loses
// all its abilities, your commanders lose the buff, though the Chakrams
// stay attached and the Equipment's own +2/+2 still applies.
func TestDancersChakramsHostThatLosesItsAbilitiesStopsBuffingCommanders(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	mine := gl6SeedCommander(g, me.ID, "My Commander")
	chakrams := seedEquipment(g, me.ID, "Dancer's Chakrams", dancersChakramsOracle)
	equipTo(t, g, me.ID, chakrams, bear)
	if !gl6HasKeyword(t, g, mine, "lifelink") {
		t.Fatal("setup: no lifelink before the removal")
	}

	gl6LoseAllAbilities(g, bear)

	if gl6HasKeyword(t, g, mine, "lifelink") {
		t.Error("the equipped creature lost all abilities, but the commander kept the granted lifelink")
	}
	if p := effectivePower(t, g, mine); p != 3 {
		t.Errorf("commander power %d, want 3: the +2/+2 went with the ability", p)
	}
	if p := effectivePower(t, g, bear); p != 4 {
		t.Errorf("bear power %d, want 4: the Equipment's own +2/+2 is not the bear's ability", p)
	}
	if !attachedTo(t, g, chakrams, bear) {
		t.Error("the Chakrams fell off")
	}
}

// CR 613.7: a commander that lost its abilities BEFORE the Chakrams were
// attached gets the lifelink back from the later grant; one that loses
// them afterwards does not keep it.
func TestDancersChakramsLifelinkSortsAgainstARemovalOnTheCommander(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	early := gl6SeedCommander(g, me.ID, "Early Commander")
	late := gl6SeedCommander(g, me.ID, "Late Commander")
	gl6LoseAllAbilities(g, early)
	chakrams := seedEquipment(g, me.ID, "Dancer's Chakrams", dancersChakramsOracle)
	equipTo(t, g, me.ID, chakrams, bear)
	gl6LoseAllAbilities(g, late)

	if !gl6HasKeyword(t, g, early, "lifelink") {
		t.Error("the commander lost its abilities before the grant, and should have lifelink from it")
	}
	if gl6HasKeyword(t, g, late, "lifelink") {
		t.Error("the commander lost its abilities after the grant, and kept the lifelink")
	}
}

// Job select: the Chakrams make a Hero and attach to it.
func TestDancersChakramsJobSelectEquipsAHero(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	mine := gl6SeedCommander(g, me.ID, "My Commander")
	chakrams := seedEquipment(g, me.ID, "Dancer's Chakrams", dancersChakramsOracle)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, Actor: me.ID, CardID: chakrams})
	})
	passPriorityAroundTable(t, g)
	host := attachmentHostOf(t, g, chakrams)
	if host.Kind != game.TargetCard {
		t.Fatalf("the Chakrams are attached to %+v, want a Hero", host)
	}
	if p, tough := effectivePower(t, g, host.ID), effectiveToughness(t, g, host.ID); p != 3 || tough != 3 {
		t.Errorf("the Hero is %d/%d, want 3/3", p, tough)
	}
	if !gl6HasKeyword(t, g, mine, "lifelink") {
		t.Error("the Hero's granted static gives your commander no lifelink")
	}
}

// --- the Kaldheim Runes --------------------------------------------------

func TestKaldheimRunesGiveACreatureAndAnEquipmentsCreatureTheirKeyword(t *testing.T) {
	for _, r := range kaldheimRunes {
		t.Run(r.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			advanceToMain(t, g)

			// On a creature: the keyword and the pump.
			bear := seedBear(g, me.ID)
			castCatalogSpell(t, g, r.name, "Enchantment — Aura Rune", r.oracleID,
				[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
			passPriorityAroundTable(t, g)
			if !gl6HasKeyword(t, g, bear, r.keyword) {
				t.Errorf("enchanted creature has no %s", r.keyword)
			}
			if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 2+r.power || tough != 2+r.toughness {
				t.Errorf("enchanted creature is %d/%d, want %d/%d", p, tough, 2+r.power, 2+r.toughness)
			}

			// On an Equipment: the equipped creature has it, the
			// Equipment does not.
			wearer := seedBear(g, me.ID)
			sword := seedEquipment(g, me.ID, "Plain Equipment", "")
			castCatalogSpell(t, g, r.name, "Enchantment — Aura Rune", r.oracleID,
				[]game.TargetRef{{Kind: game.TargetCard, ID: sword}})
			passPriorityAroundTable(t, g)
			gl6Attach(t, g, sword, wearer)
			if !gl6HasKeyword(t, g, wearer, r.keyword) {
				t.Errorf("the creature equipped with an enchanted Equipment has no %s", r.keyword)
			}
			if p, tough := effectivePower(t, g, wearer), effectiveToughness(t, g, wearer); p != 2+r.power || tough != 2+r.toughness {
				t.Errorf("equipped creature is %d/%d, want %d/%d", p, tough, 2+r.power, 2+r.toughness)
			}
			if gl6HasKeyword(t, g, sword, r.keyword) {
				t.Errorf("the Equipment itself has %s", r.keyword)
			}

			// The quoted ability is the Equipment's: when it loses all
			// abilities, the equipped creature loses the keyword.
			gl6LoseAllAbilities(g, sword)
			if gl6HasKeyword(t, g, wearer, r.keyword) {
				t.Errorf("the Equipment lost all abilities, but its creature kept %s", r.keyword)
			}
		})
	}
}

// --- Giant's Amulet -----------------------------------------------------

func TestGiantsAmuletHexproofFollowsTheEquippedCreaturesTapState(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	amulet := seedEquipment(g, me.ID, "Giant's Amulet", giantsAmuletOracle)
	floatMana(t, g, me, "{C}{C}")
	equipTo(t, g, me.ID, amulet, bear)

	if tough := effectiveToughness(t, g, bear); tough != 3 {
		t.Errorf("toughness %d, want 3", tough)
	}
	if !gl6HasKeyword(t, g, bear, game.KeywordHexproof) {
		t.Error("the untapped equipped creature has no hexproof")
	}
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(bear) })
	if gl6HasKeyword(t, g, bear, game.KeywordHexproof) {
		t.Error("the tapped equipped creature still has hexproof")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(bear) })
	if !gl6HasKeyword(t, g, bear, game.KeywordHexproof) {
		t.Error("the untapped-again creature has no hexproof")
	}
	if gl6HasKeyword(t, g, amulet, game.KeywordHexproof) {
		t.Error("the Amulet itself has hexproof")
	}
	gl6LoseAllAbilities(g, bear)
	if gl6HasKeyword(t, g, bear, game.KeywordHexproof) {
		t.Error("the creature lost all abilities, but kept the granted hexproof")
	}
}

func TestGiantsAmuletMayPayForAGiantToWearIt(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		amulet := seedEquipment(g, me.ID, "Giant's Amulet", giantsAmuletOracle)
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventETB, Actor: me.ID, CardID: amulet})
		})
		passPriorityAroundTable(t, g)
		if !hasPayUnlessFor(g, me.ID) {
			t.Fatal("no {3}{U} offer")
		}
		if pay {
			floatMana(t, g, me, "{C}{C}{C}{U}")
		}
		answerPayUnless(t, g, me.ID, pay)
		passPriorityAroundTable(t, g)
		host := attachmentHostOf(t, g, amulet)
		if !pay {
			if host.Kind == game.TargetCard {
				t.Errorf("declined, but the Amulet is attached to %s", host.ID)
			}
			continue
		}
		if host.Kind != game.TargetCard {
			t.Fatal("paid, but the Amulet is attached to nothing")
		}
		if p, tough := effectivePower(t, g, host.ID), effectiveToughness(t, g, host.ID); p != 4 || tough != 5 {
			t.Errorf("the Giant is %d/%d, want 4/5", p, tough)
		}
		if !containsString(effectiveSubtypes(t, g, host.ID), "Giant") || !gl6HasKeyword(t, g, host.ID, game.KeywordHexproof) {
			t.Errorf("the token is %v with %v, want an untapped Giant Wizard with hexproof",
				effectiveSubtypes(t, g, host.ID), effectiveAbilities(t, g, host.ID))
		}
	}
}
