package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsg "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/g"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func two(a, b uuid.UUID) []game.TargetRef { return effectstest.CardRefs(a, b) }

// --- sorceries ------------------------------------------------------

func TestEriettesWhisperDiscardsTwoAndRolesMyCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 2, 2)
	for i := 0; i < 3; i++ {
		opp.Hand.PushTop(game.NewCard("Filler", opp.ID))
	}
	effectstest.CastCatalogSpell(t, g, "Eriette's Whisper", "Sorcery", effectstest.ErietteOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: mine}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DiscardOwed(g, opp.ID); got != 2 {
		t.Errorf("opponent owes %d discards, want 2", got)
	}
	// The Role is the discard's continuation: it appears once it is answered.
	effectstest.WantNoRole(t, g, mine)
	effectstest.DiscardFromHand(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.WantRole(t, g, mine, "Wicked Role")
}

func TestEriettesWhisperRefusesMyselfAsTheDiscarder(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	if err := effectstest.CastCatalogSpellErr(t, g, "Eriette's Whisper", "Sorcery", effectstest.ErietteOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}); err == nil {
		t.Fatal("\"target opponent\" accepted the caster")
	}
}

func TestShatterTheOathDestroysAndRoles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 2, 2)
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Shatter the Oath", "Sorcery", effectstest.ShatterOathOracle, two(victim, mine))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldByName(g, "Victim") != uuid.Nil {
		t.Error("the target creature survived")
	}
	effectstest.WantRole(t, g, mine, "Wicked Role")
}

func TestShatterTheOathDestroysAnEnchantmentWithoutARoleTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	ench := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Enchantment", TypeLine: "Enchantment",
		Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.CastCatalogSpell(t, g, "Shatter the Oath", "Sorcery", effectstest.ShatterOathOracle, effectstest.CardRefs(ench))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldByName(g, "Their Enchantment") != uuid.Nil {
		t.Error("the enchantment survived")
	}
}

func TestShatterTheOathRefusesALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest",
		Owner: opp.ID, Controller: opp.ID,
	})
	if err := effectstest.CastCatalogSpellErr(t, g, "Shatter the Oath", "Sorcery", effectstest.ShatterOathOracle, effectstest.CardRefs(land)); err == nil {
		t.Fatal("a land was a legal target")
	}
}

func TestTwistedFealtyStealsAndRolesAnyCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
	other := effectstest.PushVanillaCreature(g, opp.ID, "Other", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Twisted Fealty", "Sorcery", effectstest.TwistedFealtyOracle, two(theirs, other))
	effectstest.PassPriorityAroundTable(t, g)
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == theirs && c.Controller != me.ID {
				t.Error("the creature was not stolen")
			}
		}
	})
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, theirs), "haste") {
		t.Error("the stolen creature lacks haste")
	}
	effectstest.WantRole(t, g, other, "Wicked Role")
}

func TestWitchsMarkRolesAndOffersTheDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 2, 2)
	pitch := game.NewCard("Pitch", me.ID)
	me.Hand.PushTop(pitch)
	effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.CastCatalogSpell(t, g, "Witch's Mark", "Sorcery", effectstest.WitchsMarkOracle, effectstest.CardRefs(mine))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.WantRole(t, g, mine, "Wicked Role")
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("no optional discard prompt")
	}
	before := effectstest.HandCount(g, me)
	effectstest.AnswerDiscard(t, g, me.ID, pitch.InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	// One card out, two in.
	if got := effectstest.HandCount(g, me); got != before+1 {
		t.Errorf("hand = %d, want %d (discard one, draw two)", got, before+1)
	}
}

func TestWitchsMarkDecliningDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	me.Hand.PushTop(game.NewCard("Keep", me.ID))
	effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.CastCatalogSpell(t, g, "Witch's Mark", "Sorcery", effectstest.WitchsMarkOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	before := effectstest.HandCount(g, me)
	effectstest.AnswerDiscard(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandCount(g, me); got != before {
		t.Errorf("hand changed from %d to %d after declining", before, got)
	}
}

func TestBecomeBrutesGivesHasteAndAMonsterRoleToEach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.PushVanillaCreature(g, me.ID, "A", 2, 2)
	b := effectstest.PushVanillaCreature(g, opp.ID, "B", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Become Brutes", "Sorcery", effectstest.BecomeBrutesOracle, two(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		effectstest.WantRole(t, g, id, "Monster Role")
		if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, id), "haste") {
			t.Errorf("creature %v lacks haste", id)
		}
	}
	if effectstest.RolesOn(g, b)[0].Controller != me.ID {
		t.Error("the Role on the opponent's creature is not the caster's")
	}
}

func TestBecomeBrutesRefusesThreeTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.PushVanillaCreature(g, me.ID, "A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "B", 2, 2)
	c := effectstest.PushVanillaCreature(g, me.ID, "C", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Become Brutes", "Sorcery", effectstest.BecomeBrutesOracle, effectstest.CardRefs(a, b, c)); err == nil {
		t.Fatal("three targets accepted for \"one or two\"")
	}
}

func TestAsinineAnticsCursesEveryOpposingCreatureAndNoneOfMine(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 2, 2)
	a := effectstest.PushVanillaCreature(g, opp.ID, "A", 4, 4)
	b := effectstest.PushVanillaCreature(g, g.Seats[2].ID, "B", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Asinine Antics", "Sorcery", effectstest.AsinineAnticsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.WantNoRole(t, g, mine)
	for _, id := range []uuid.UUID{a, b} {
		effectstest.WantRole(t, g, id, "Cursed Role")
		if p, tough := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id); p != 1 || tough != 1 {
			t.Errorf("cursed creature is %d/%d, want 1/1", p, tough)
		}
	}
}

func TestCurseOfTheWerefoxRolesThenFights(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Curse of the Werefox", "Sorcery", effectstest.CurseWerefoxOracle, effectstest.CardRefs(mine))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.WantRole(t, g, mine, "Monster Role")
	if effectstest.LatestPickTarget(g, me.ID) == nil {
		t.Fatal("no target prompt for the reflexive fight")
	}
	effectstest.PickTriggerTarget(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	// 3/3 (Role) fights a 3/3: both die.
	if effectstest.FindBattlefieldByName(g, "Theirs") != uuid.Nil || effectstest.FindBattlefieldByName(g, "Mine") != uuid.Nil {
		t.Error("the fight did not resolve between 3/3s")
	}
}

func TestCurseOfTheWerefoxRefusesAnOpponentsCreatureAsTheRoleTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 3, 3)
	if err := effectstest.CastCatalogSpellErr(t, g, "Curse of the Werefox", "Sorcery", effectstest.CurseWerefoxOracle, effectstest.CardRefs(theirs)); err == nil {
		t.Fatal("the Role clause accepted a creature I don't control")
	}
}

// --- creatures ------------------------------------------------------

func TestDiminisherWitchUnbargainedDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 4, 4)
	effectstest.CastAndResolveCreature(t, g, "Diminisher Witch", "Creature — Human Warlock", effectstest.DiminisherOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, g.Seats[0].ID) != nil {
		t.Error("an unbargained Witch asked for a target")
	}
	effectstest.WantNoRole(t, g, theirs)
}

func TestDiminisherWitchBargainedCursesAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 4, 4)
	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Treasure", TypeLine: "Token Artifact — Treasure",
		Owner: me.ID, Controller: me.ID,
	})
	effectstest.P1g2Cast(t, g, "Diminisher Witch", "Creature — Human Warlock", effectstest.DiminisherOracle, "", game.CastSpellParams{
		OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{token},
	})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.WantRole(t, g, theirs, "Cursed Role")
	if p := effectstest.EffectivePower(t, g, theirs); p != 1 {
		t.Errorf("cursed creature has power %d, want 1", p)
	}
}

func TestTwistedSewerWitchRolesTheNewRatAndEveryOtherRat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oldRat := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Old Rat", TypeLine: "Creature — Rat",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	witch := effectstest.CastAndResolveCreature(t, g, "Twisted Sewer-Witch", "Creature — Human Warlock", effectstest.SewerWitchOracle)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.WantRole(t, g, oldRat, "Wicked Role")
	effectstest.WantNoRole(t, g, bear)
	effectstest.WantNoRole(t, g, witch)
	var newRat game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Rat" && c.IsToken() {
				newRat = c
			}
		}
	})
	if newRat.InstanceID == uuid.Nil {
		t.Fatal("no Rat token was created")
	}
	effectstest.WantRole(t, g, newRat.InstanceID, "Wicked Role")
	// 1/1 + 1/+0.
	if p := effectstest.EffectivePower(t, g, newRat.InstanceID); p != 2 {
		t.Errorf("the new Rat has power %d, want 2", p)
	}
	effectstest.AssertRestrictions(t, g, newRat.InstanceID, game.CantBlock)
}

func TestCharmingScoundrelEachBullet(t *testing.T) {
	pick := func(t *testing.T, mode int) (*game.Game, *game.Player, uuid.UUID) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		id := effectstest.CastAndResolveCreature(t, g, "Charming Scoundrel", "Creature — Human Rogue", effectstest.ScoundrelOracle)
		c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if c == nil {
			t.Fatalf("no mode prompt: %+v", g.PendingChoices)
		}
		if err := g.ResolveModePick(c.ID, me.ID, []int{mode}); err != nil {
			t.Fatalf("ResolveModePick: %v", err)
		}
		return g, me, id
	}
	t.Run("discard then draw", func(t *testing.T) {
		g, me, _ := pick(t, 0)
		me.Hand.PushTop(game.NewCard("Pitch", me.ID))
		effectstest.SeedLibrary(me, "A", "B")
		effectstest.PassPriorityAroundTable(t, g)
		before := effectstest.HandCount(g, me)
		effectstest.DiscardFromHand(t, g, me.ID)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.HandCount(g, me); got != before {
			t.Errorf("hand %d, want %d (discard one, draw one)", got, before)
		}
	})
	t.Run("treasure", func(t *testing.T) {
		g, _, _ := pick(t, 1)
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.FindBattlefieldByName(g, "Treasure") == uuid.Nil {
			t.Error("no Treasure")
		}
	})
	t.Run("role", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
		effectstest.CastAndResolveCreature(t, g, "Charming Scoundrel", "Creature — Human Rogue", effectstest.ScoundrelOracle)
		c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if c == nil {
			t.Fatal("no mode prompt")
		}
		if err := g.ResolveModePick(c.ID, me.ID, []int{2}); err != nil {
			t.Fatalf("ResolveModePick: %v", err)
		}
		effectstest.PickTriggerTarget(t, g, me.ID, bear)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.WantRole(t, g, bear, "Wicked Role")
	})
}

func TestGylwainRolesTheEnteringCreatureAndItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gylwain := effectstest.CastAndResolveCreature(t, g, "Gylwain, Casting Director", "Legendary Creature — Human Bard", effectstest.GylwainOracle)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if c == nil {
		t.Fatalf("no mode prompt for Gylwain's own entry: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.WantRole(t, g, gylwain, "Royal Role")

	// Another nontoken creature entering asks again, and each bullet is a
	// different Role.
	effectstest.CastAndResolveCreature(t, g, "Grizzly Bears", "Creature — Bear", "")
	c = effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if c == nil {
		t.Fatal("no mode prompt for another creature entering")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	bears := effectstest.FindBattlefieldByName(g, "Grizzly Bears")
	effectstest.WantRole(t, g, bears, "Sorcerer Role")
}

func TestGylwainIgnoresTokensAndOpponentsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Gylwain, Casting Director", "Legendary Creature — Human Bard", effectstest.GylwainOracle, false)
	// An opponent's creature entering does not trigger it.
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Theirs", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID); c != nil {
		t.Error("an opponent's creature entering prompted Gylwain")
	}
}

// --- sagas ----------------------------------------------------------

func TestTheWitchsVanityChapters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cheap := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Cheap", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	pricey := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Pricey", TypeLine: "Creature — Bear", ManaCost: "{2}{G}",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 2, 2)
	seat := g.Turn.ActiveSeat

	effectstest.CastCatalogSpell(t, g, "The Witch's Vanity", "Enchantment — Saga", effectstest.WitchsVanityOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickTriggerTarget(t, g, me.ID, cheap)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if effectstest.FindBattlefieldByName(g, "Cheap") != uuid.Nil {
		t.Error("chapter I did not destroy the mana value 2 creature")
	}
	if effectstest.FindBattlefieldByName(g, "Pricey") == uuid.Nil {
		t.Error("chapter I destroyed a mana value 3 creature")
	}
	_ = pricey

	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldByName(g, "Food") == uuid.Nil {
		t.Error("chapter II made no Food")
	}

	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickTriggerTarget(t, g, me.ID, mine)
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.WantRole(t, g, mine, "Wicked Role")
}

func TestTheWitchsVanityChapterOneRefusesAThreeDrop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Pricey", TypeLine: "Creature — Bear", ManaCost: "{2}{G}",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.CastCatalogSpell(t, g, "The Witch's Vanity", "Enchantment — Saga", effectstest.WitchsVanityOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, g.Seats[0].ID) != nil {
		t.Error("chapter I offered a target when none has mana value 2 or less")
	}
	if effectstest.FindBattlefieldByName(g, "Pricey") == uuid.Nil {
		t.Error("a mana value 3 creature was destroyed")
	}
}

func TestGadwicksFirstDuelChapters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 4, 4)
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	seat := g.Turn.ActiveSeat

	effectstest.CastCatalogSpell(t, g, "Gadwick's First Duel", "Enchantment — Saga", effectstest.GadwicksDuelOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) == nil {
		t.Fatal("no target prompt for chapter I")
	}
	effectstest.PickTriggerTarget(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.WantRole(t, g, theirs, "Cursed Role")

	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Fatal("chapter II queued no scry")
	}
	effectstest.AnswerScryKeepAll(t, g, me.ID)

	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("chapter III left %d delayed triggers, want 1", len(g.DelayedTriggers))
	}
}

func TestGadwicksDelayedCopyOnlyFiresForAThreeOrLessSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0].ID, g.Seats[1].ID
	before := effectstest.LifeOf(g, victim)
	g.WithWriteLock(func() {
		err := effects.DelayedOnEvent{
			Label:      "Gadwick's First Duel — copy that spell",
			On:         []game.EventKind{game.EventCast},
			Condition:  cardsg.GadwickNextCastCondition,
			CondParams: game.EffectParams{Filter: game.CastFilter{Types: []string{"Instant", "Sorcery"}}, Amount: 3},
			Body:       effects.CopyTheSpellBody,
		}.Apply(effects.NewContext(g, &game.StackItem{Controller: me}))
		if err != nil {
			t.Fatal(err)
		}
	})

	// A four-mana instant does not use the trigger up.
	big := game.NewCard("Big Instant", me)
	big.TypeLine = "Instant"
	big.ManaCost = "{3}{R}"
	g.Seats[0].Hand.PushTop(big)
	if err := g.CastSpell(me, big.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the big instant: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("a mana value 4 spell used the trigger up (%d left)", len(g.DelayedTriggers))
	}

	// A Lightning Bolt (mana value 1) is copied.
	effectstest.CastBoltAtAndKeepTheCopysTarget(t, g, me, victim, 1)
	if got := effectstest.LifeOf(g, victim); got != before-6 {
		t.Errorf("victim life = %d, want %d (the Bolt and its copy)", got, before-6)
	}
}

// --- enchantments ---------------------------------------------------

func TestLordSkittersBlessingRolesOnEntryAndDrawsExtraWhileEnchanted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 2, 2)
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")
	effectstest.CastCatalogSpell(t, g, "Lord Skitter's Blessing", "Enchantment", effectstest.SkittersOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.WantRole(t, g, mine, "Wicked Role")

	// Next turn's draw step: lose 1 life, draw an additional card.
	seat := g.Turn.ActiveSeat
	life, hand := me.Life, effectstest.HandCount(g, me)
	for i := 0; i < 400; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
		if g.Turn.Step == game.StepDraw && g.Turn.ActiveSeat == seat {
			break
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-1 {
		t.Errorf("life = %d, want %d", me.Life, life-1)
	}
	if got := effectstest.HandCount(g, me); got != hand+2 {
		t.Errorf("hand = %d, want %d (the draw step's card and one more)", got, hand+2)
	}
}

func TestLordSkittersBlessingDoesNothingInTheDrawStepWithoutAnEnchantedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Lord Skitter's Blessing", "Enchantment", effectstest.SkittersOracle, false)
	effectstest.PushVanillaCreature(g, me.ID, "Bare", 2, 2)
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")
	seat := g.Turn.ActiveSeat
	life := me.Life
	for i := 0; i < 400; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
		if g.Turn.Step == game.StepDraw && g.Turn.ActiveSeat == seat {
			break
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life = %d, want %d — nothing is enchanted", me.Life, life)
	}
}

func TestGiantInheritanceGrowsTheCreatureAndRolesAnAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	friend := effectstest.PushVanillaCreature(g, me.ID, "Friend", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Giant Inheritance", "Enchantment — Aura", effectstest.GiantInheritOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 7 || tough != 7 {
		t.Fatalf("enchanted bear is %d/%d, want 7/7", p, tough)
	}

	effectstest.DeclareAttack(t, g, opp.ID, bear, friend)
	if effectstest.LatestPickTarget(g, me.ID) == nil {
		t.Fatal("no target prompt for the granted attack trigger")
	}
	effectstest.PickTriggerTarget(t, g, me.ID, friend)
	if effectstest.TriggersOnStackFrom(g, bear) != 1 {
		t.Fatalf("granted triggers from the bear = %d, want 1 (the Aura is not the source)", effectstest.TriggersOnStackFrom(g, bear))
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.WantRole(t, g, friend, "Monster Role")
}

func TestGiantInheritanceReturnsToHandWhenItIsPutIntoAGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	aura := effectstest.CastCatalogSpell(t, g, "Giant Inheritance", "Enchantment — Aura", effectstest.GiantInheritOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	inHand := false
	g.ReadSnapshot(func() {
		for _, c := range me.Hand.Cards {
			if c.InstanceID == aura {
				inHand = true
			}
		}
	})
	if !inHand {
		t.Error("Giant Inheritance did not return to its owner's hand")
	}
}

// --- instants -------------------------------------------------------

func TestNotDeadAfterAllReturnsTheCreatureWithAWickedRole(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Not Dead After All", "Instant", effectstest.NotDeadOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)

	back := effectstest.FindBattlefieldByName(g, "Bear")
	if back == uuid.Nil {
		t.Fatal("the creature did not come back")
	}
	tapped := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == back {
				tapped = c.Tapped
			}
		}
	})
	if !tapped {
		t.Error("the creature returned untapped")
	}
	effectstest.WantRole(t, g, back, "Wicked Role")
}

func TestNotDeadAfterAllGivesATokenNoRoleAndNoReturn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tok := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Tok", TypeLine: "Token Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	effectstest.CastCatalogSpell(t, g, "Not Dead After All", "Instant", effectstest.NotDeadOracle, effectstest.CardRefs(tok))
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(tok) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldByName(g, "Tok") != uuid.Nil {
		t.Error("a token came back")
	}
	if effectstest.FindBattlefieldByName(g, "Wicked Role") != uuid.Nil {
		t.Error("a Role was created for a token that never returned")
	}
}

func TestNotDeadAfterAllGrantEndsWithTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Not Dead After All", "Instant", effectstest.NotDeadOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToPrecombatMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldByName(g, "Bear") != uuid.Nil {
		t.Error("the grant outlived the turn")
	}
}

func TestAdventureHalvesCreateTheirRoles(t *testing.T) {
	cases := []effectstest.AdventureCase{
		{Name: "Besotted Knight", Adv: "Betroth the Beast", Oracle: effectstest.BesottedKnightOracle, Role: "Royal Role",
			Faces: [2]game.Face{
				{Name: "Besotted Knight", TypeLine: "Creature — Human Knight", ManaCost: "{3}{W}", Colors: []string{"W"}, Power: 3, Toughness: 3},
				{Name: "Betroth the Beast", TypeLine: "Sorcery — Adventure", ManaCost: "{W}", Colors: []string{"W"}},
			}},
		{Name: "Conceited Witch", Adv: "Price of Beauty", Oracle: effectstest.ConceitedOracle, Role: "Wicked Role",
			Faces: [2]game.Face{
				{Name: "Conceited Witch", TypeLine: "Creature — Human Warlock", ManaCost: "{2}{B}", Colors: []string{"B"}, Power: 2, Toughness: 3},
				{Name: "Price of Beauty", TypeLine: "Sorcery — Adventure", ManaCost: "{B}", Colors: []string{"B"}},
			}},
		{Name: "Ferocious Werefox", Adv: "Guard Change", Oracle: effectstest.WerefoxOracle, Role: "Monster Role",
			Faces: [2]game.Face{
				{Name: "Ferocious Werefox", TypeLine: "Creature — Elf Fox Warrior", ManaCost: "{3}{G}", Colors: []string{"G"}, Power: 4, Toughness: 3},
				{Name: "Guard Change", TypeLine: "Instant — Adventure", ManaCost: "{1}{G}", Colors: []string{"G"}},
			}},
	}
	for _, tc := range cases {
		t.Run(tc.Adv, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
			card := effectstest.AdventureCard(me.ID, tc)
			me.Hand.PushTop(card)
			effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
			if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, Targets: effectstest.CardRefs(bear)}); err != nil {
				t.Fatalf("cast %s: %v", tc.Adv, err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.WantRole(t, g, bear, tc.Role)
			if !g.Exile.Contains(card.InstanceID) {
				t.Error("the Adventure card is not in exile (CR 715.3d)")
			}
		})
	}
}

func TestCroakingCurseTapsAndCursesAnyCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 4, 4)
	card := effectstest.AdventureCard(me.ID, effectstest.AdventureCase{Oracle: effectstest.TransmuterOracle, Faces: [2]game.Face{
		{Name: "Vantress Transmuter", TypeLine: "Creature — Human Wizard", ManaCost: "{3}{U}", Colors: []string{"U"}, Power: 3, Toughness: 4},
		{Name: "Croaking Curse", TypeLine: "Sorcery — Adventure", ManaCost: "{1}{U}", Colors: []string{"U"}},
	}})
	me.Hand.PushTop(card)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, Targets: effectstest.CardRefs(theirs)}); err != nil {
		t.Fatalf("cast Croaking Curse: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.WantRole(t, g, theirs, "Cursed Role")
	tapped := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == theirs {
				tapped = c.Tapped
			}
		}
	})
	if !tapped {
		t.Error("the target was not tapped")
	}
}

func TestBetrothTheBeastRefusesAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
	card := effectstest.AdventureCard(me.ID, effectstest.AdventureCase{Oracle: effectstest.BesottedKnightOracle, Faces: [2]game.Face{
		{Name: "Besotted Knight", TypeLine: "Creature — Human Knight", ManaCost: "{3}{W}", Colors: []string{"W"}, Power: 3, Toughness: 3},
		{Name: "Betroth the Beast", TypeLine: "Sorcery — Adventure", ManaCost: "{W}", Colors: []string{"W"}},
	}})
	me.Hand.PushTop(card)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, Targets: effectstest.CardRefs(theirs)}); err == nil {
		t.Fatal("an opponent's creature was a legal target for \"creature you control\"")
	}
}
