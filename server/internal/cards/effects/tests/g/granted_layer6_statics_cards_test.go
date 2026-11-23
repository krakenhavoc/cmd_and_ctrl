package testsg

import (
	"testing"

	"github.com/google/uuid"

	cardsj "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/j"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Legendary Creature — Human",
		Power: 3, Toughness: 3, Owner: owner, Controller: owner, IsCommander: true,
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
	if !effectstest.AttachedTo(t, g, equipment, creature) {
		t.Fatal("the Equipment did not attach")
	}
}

func TestDancersChakramsBuffsOtherCommandersYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	mine := gl6SeedCommander(g, me.ID, "My Commander")
	theirs := gl6SeedCommander(g, opp.ID, "Their Commander")
	chakrams := effectstest.SeedEquipment(g, me.ID, "Dancer's Chakrams", dancersChakramsOracle)
	effectstest.EquipTo(t, g, me.ID, chakrams, bear)

	// The equipped creature: +2/+2, lifelink, a Performer.
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("equipped bear is %d/%d, want 4/4", p, tough)
	}
	if !effectstest.Gl6HasKeyword(t, g, bear, "lifelink") {
		t.Error("the equipped creature has no lifelink")
	}
	if !effectstest.ContainsString(effectstest.EffectiveSubtypes(t, g, bear), "Performer") {
		t.Errorf("subtypes %v, want a Performer", effectstest.EffectiveSubtypes(t, g, bear))
	}
	// The quoted ability: your other commanders.
	if p, tough := effectstest.EffectivePower(t, g, mine), effectstest.EffectiveToughness(t, g, mine); p != 5 || tough != 5 {
		t.Errorf("your commander is %d/%d, want 5/5", p, tough)
	}
	if !effectstest.Gl6HasKeyword(t, g, mine, "lifelink") {
		t.Error("your commander has no lifelink")
	}
	// Not an opponent's.
	if p := effectstest.EffectivePower(t, g, theirs); p != 3 {
		t.Errorf("their commander's power is %d, want 3", p)
	}
	if effectstest.Gl6HasKeyword(t, g, theirs, "lifelink") {
		t.Error("an opponent's commander got lifelink")
	}
}

// "Other": a commander wearing the Chakrams gets the Equipment's own
// +2/+2 once, not the commanders' buff on top.
func TestDancersChakramsOnACommanderBuffsTheOtherOneOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	first := gl6SeedCommander(g, me.ID, "First Commander")
	second := gl6SeedCommander(g, me.ID, "Second Commander")
	chakrams := effectstest.SeedEquipment(g, me.ID, "Dancer's Chakrams", dancersChakramsOracle)
	effectstest.EquipTo(t, g, me.ID, chakrams, first)
	if p := effectstest.EffectivePower(t, g, first); p != 5 {
		t.Errorf("the equipped commander's power is %d, want 5 (+2/+2 once)", p)
	}
	if p := effectstest.EffectivePower(t, g, second); p != 5 || !effectstest.Gl6HasKeyword(t, g, second, "lifelink") {
		t.Errorf("the other commander's power is %d (lifelink %v), want 5 with lifelink", p, effectstest.Gl6HasKeyword(t, g, second, "lifelink"))
	}
}

// The quoted ability is the CREATURE's: when the equipped creature loses
// all its abilities, your commanders lose the buff, though the Chakrams
// stay attached and the Equipment's own +2/+2 still applies.
func TestDancersChakramsHostThatLosesItsAbilitiesStopsBuffingCommanders(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	mine := gl6SeedCommander(g, me.ID, "My Commander")
	chakrams := effectstest.SeedEquipment(g, me.ID, "Dancer's Chakrams", dancersChakramsOracle)
	effectstest.EquipTo(t, g, me.ID, chakrams, bear)
	if !effectstest.Gl6HasKeyword(t, g, mine, "lifelink") {
		t.Fatal("setup: no lifelink before the removal")
	}

	effectstest.Gl6LoseAllAbilities(g, bear)

	if effectstest.Gl6HasKeyword(t, g, mine, "lifelink") {
		t.Error("the equipped creature lost all abilities, but the commander kept the granted lifelink")
	}
	if p := effectstest.EffectivePower(t, g, mine); p != 3 {
		t.Errorf("commander power %d, want 3: the +2/+2 went with the ability", p)
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 4 {
		t.Errorf("bear power %d, want 4: the Equipment's own +2/+2 is not the bear's ability", p)
	}
	if !effectstest.AttachedTo(t, g, chakrams, bear) {
		t.Error("the Chakrams fell off")
	}
}

// CR 613.7: a commander that lost its abilities BEFORE the Chakrams were
// attached gets the lifelink back from the later grant; one that loses
// them afterwards does not keep it.
func TestDancersChakramsLifelinkSortsAgainstARemovalOnTheCommander(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	early := gl6SeedCommander(g, me.ID, "Early Commander")
	late := gl6SeedCommander(g, me.ID, "Late Commander")
	effectstest.Gl6LoseAllAbilities(g, early)
	chakrams := effectstest.SeedEquipment(g, me.ID, "Dancer's Chakrams", dancersChakramsOracle)
	effectstest.EquipTo(t, g, me.ID, chakrams, bear)
	effectstest.Gl6LoseAllAbilities(g, late)

	if !effectstest.Gl6HasKeyword(t, g, early, "lifelink") {
		t.Error("the commander lost its abilities before the grant, and should have lifelink from it")
	}
	if effectstest.Gl6HasKeyword(t, g, late, "lifelink") {
		t.Error("the commander lost its abilities after the grant, and kept the lifelink")
	}
}

// Job select: the Chakrams make a Hero and attach to it.
func TestDancersChakramsJobSelectEquipsAHero(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	mine := gl6SeedCommander(g, me.ID, "My Commander")
	chakrams := effectstest.SeedEquipment(g, me.ID, "Dancer's Chakrams", dancersChakramsOracle)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, Actor: me.ID, CardID: chakrams})
	})
	effectstest.PassPriorityAroundTable(t, g)
	host := effectstest.AttachmentHostOf(t, g, chakrams)
	if host.Kind != game.TargetCard {
		t.Fatalf("the Chakrams are attached to %+v, want a Hero", host)
	}
	if p, tough := effectstest.EffectivePower(t, g, host.ID), effectstest.EffectiveToughness(t, g, host.ID); p != 3 || tough != 3 {
		t.Errorf("the Hero is %d/%d, want 3/3", p, tough)
	}
	if !effectstest.Gl6HasKeyword(t, g, mine, "lifelink") {
		t.Error("the Hero's granted static gives your commander no lifelink")
	}
}

// --- the Kaldheim Runes --------------------------------------------------

func TestKaldheimRunesGiveACreatureAndAnEquipmentsCreatureTheirKeyword(t *testing.T) {
	for _, r := range cardsj.KaldheimRunes {
		t.Run(r.Name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.AdvanceToMain(t, g)

			// On a creature: the keyword and the pump.
			bear := effectstest.SeedBear(g, me.ID)
			effectstest.CastCatalogSpell(t, g, r.Name, "Enchantment — Aura Rune", r.OracleID,
				[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
			effectstest.PassPriorityAroundTable(t, g)
			if !effectstest.Gl6HasKeyword(t, g, bear, r.Keyword) {
				t.Errorf("enchanted creature has no %s", r.Keyword)
			}
			if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 2+r.Power || tough != 2+r.Toughness {
				t.Errorf("enchanted creature is %d/%d, want %d/%d", p, tough, 2+r.Power, 2+r.Toughness)
			}

			// On an Equipment: the equipped creature has it, the
			// Equipment does not.
			wearer := effectstest.SeedBear(g, me.ID)
			sword := effectstest.SeedEquipment(g, me.ID, "Plain Equipment", "")
			effectstest.CastCatalogSpell(t, g, r.Name, "Enchantment — Aura Rune", r.OracleID,
				[]game.TargetRef{{Kind: game.TargetCard, ID: sword}})
			effectstest.PassPriorityAroundTable(t, g)
			gl6Attach(t, g, sword, wearer)
			if !effectstest.Gl6HasKeyword(t, g, wearer, r.Keyword) {
				t.Errorf("the creature equipped with an enchanted Equipment has no %s", r.Keyword)
			}
			if p, tough := effectstest.EffectivePower(t, g, wearer), effectstest.EffectiveToughness(t, g, wearer); p != 2+r.Power || tough != 2+r.Toughness {
				t.Errorf("equipped creature is %d/%d, want %d/%d", p, tough, 2+r.Power, 2+r.Toughness)
			}
			if effectstest.Gl6HasKeyword(t, g, sword, r.Keyword) {
				t.Errorf("the Equipment itself has %s", r.Keyword)
			}

			// The quoted ability is the Equipment's: when it loses all
			// abilities, the equipped creature loses the keyword.
			effectstest.Gl6LoseAllAbilities(g, sword)
			if effectstest.Gl6HasKeyword(t, g, wearer, r.Keyword) {
				t.Errorf("the Equipment lost all abilities, but its creature kept %s", r.Keyword)
			}
		})
	}
}

// --- Giant's Amulet -----------------------------------------------------

func TestGiantsAmuletHexproofFollowsTheEquippedCreaturesTapState(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	amulet := effectstest.SeedEquipment(g, me.ID, "Giant's Amulet", giantsAmuletOracle)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	effectstest.EquipTo(t, g, me.ID, amulet, bear)

	if tough := effectstest.EffectiveToughness(t, g, bear); tough != 3 {
		t.Errorf("toughness %d, want 3", tough)
	}
	if !effectstest.Gl6HasKeyword(t, g, bear, game.KeywordHexproof) {
		t.Error("the untapped equipped creature has no hexproof")
	}
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(bear) })
	if effectstest.Gl6HasKeyword(t, g, bear, game.KeywordHexproof) {
		t.Error("the tapped equipped creature still has hexproof")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(bear) })
	if !effectstest.Gl6HasKeyword(t, g, bear, game.KeywordHexproof) {
		t.Error("the untapped-again creature has no hexproof")
	}
	if effectstest.Gl6HasKeyword(t, g, amulet, game.KeywordHexproof) {
		t.Error("the Amulet itself has hexproof")
	}
	effectstest.Gl6LoseAllAbilities(g, bear)
	if effectstest.Gl6HasKeyword(t, g, bear, game.KeywordHexproof) {
		t.Error("the creature lost all abilities, but kept the granted hexproof")
	}
}

func TestGiantsAmuletMayPayForAGiantToWearIt(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		amulet := effectstest.SeedEquipment(g, me.ID, "Giant's Amulet", giantsAmuletOracle)
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventETB, Actor: me.ID, CardID: amulet})
		})
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.HasPayUnlessFor(g, me.ID) {
			t.Fatal("no {3}{U} offer")
		}
		if pay {
			effectstest.FloatMana(t, g, me, "{C}{C}{C}{U}")
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		host := effectstest.AttachmentHostOf(t, g, amulet)
		if !pay {
			if host.Kind == game.TargetCard {
				t.Errorf("declined, but the Amulet is attached to %s", host.ID)
			}
			continue
		}
		if host.Kind != game.TargetCard {
			t.Fatal("paid, but the Amulet is attached to nothing")
		}
		if p, tough := effectstest.EffectivePower(t, g, host.ID), effectstest.EffectiveToughness(t, g, host.ID); p != 4 || tough != 5 {
			t.Errorf("the Giant is %d/%d, want 4/5", p, tough)
		}
		if !effectstest.ContainsString(effectstest.EffectiveSubtypes(t, g, host.ID), "Giant") || !effectstest.Gl6HasKeyword(t, g, host.ID, game.KeywordHexproof) {
			t.Errorf("the token is %v with %v, want an untapped Giant Wizard with hexproof",
				effectstest.EffectiveSubtypes(t, g, host.ID), effectstest.EffectiveAbilities(t, g, host.ID))
		}
	}
}
