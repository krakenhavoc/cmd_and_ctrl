package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// role_tokens_part3_test.go — #1945 slice roles-c: the remaining cards
// that create a Role and need no new machinery.

const (
	gylwainOracle        = "e7dfe47f-41fc-4071-9ad3-96cd8d9be52e"
	curseWerefoxOracle   = "64b41b54-ca22-41df-8bd3-dc4f7832fbb7"
	asinineAnticsOracle  = "0a0a051c-59ed-4286-b842-a34da7ef15d5"
	diminisherOracle     = "c2283655-f6b8-4399-ba1f-68a7107f9485"
	erietteOracle        = "27ca3201-a374-4a9d-b0ba-93b391f36844"
	shatterOathOracle    = "f545870d-b628-4acb-88c7-37c2d72a5a85"
	twistedFealtyOracle  = "a5ae75de-0dff-44bf-9dc6-5895ab3ae731"
	sewerWitchOracle     = "99c9920a-f84b-4626-9fe5-d94cc62cd277"
	witchsMarkOracle     = "00807825-88e5-4b88-8527-a49f553a91be"
	skittersOracle       = "d5bd40fc-6138-4752-b3ff-b1e5fa269214"
	witchsVanityOracle   = "ad2793f3-a36f-4a09-a38f-eb98ce0a2b79"
	gadwicksDuelOracle   = "322e3bc1-2dfa-4d5f-848f-a82d9ce02a67"
	giantInheritOracle   = "6be21185-f85c-49e8-988a-a76b6613fdaa"
	notDeadOracle        = "380c367f-73ea-485f-a37b-5af0ba160893"
	scoundrelOracle      = "9eec304d-ffcb-4287-b877-280ac1e4f496"
	becomeBrutesOracle   = "789e034c-7f4a-4fd5-9dd8-48313742126e"
	besottedKnightOracle = "c7b6d2cd-9105-404b-88f0-a67037fb2120"
	conceitedOracle      = "1c751201-24ba-4e5c-bf44-71c3e4693a92"
	werefoxOracle        = "fef67393-9b7b-495d-8b03-865537386d42"
	transmuterOracle     = "89ae3475-9063-4700-a7bd-47d5dcd43a3b"
)

func wantRole(t *testing.T, g *game.Game, host uuid.UUID, name string) {
	t.Helper()
	roles := rolesOn(g, host)
	if len(roles) != 1 || roles[0].Name != name {
		t.Fatalf("roles = %+v, want one %s", roles, name)
	}
}

func wantNoRole(t *testing.T, g *game.Game, host uuid.UUID) {
	t.Helper()
	if roles := rolesOn(g, host); len(roles) != 0 {
		t.Fatalf("roles = %+v, want none", roles)
	}
}

func two(a, b uuid.UUID) []game.TargetRef { return cardRefs(a, b) }

// --- sorceries ------------------------------------------------------

func TestEriettesWhisperDiscardsTwoAndRolesMyCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushVanillaCreature(g, me.ID, "Mine", 2, 2)
	for i := 0; i < 3; i++ {
		opp.Hand.PushTop(game.NewCard("Filler", opp.ID))
	}
	castCatalogSpell(t, g, "Eriette's Whisper", "Sorcery", erietteOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: mine}})
	passPriorityAroundTable(t, g)
	if got := discardOwed(g, opp.ID); got != 2 {
		t.Errorf("opponent owes %d discards, want 2", got)
	}
	// The Role is the discard's continuation: it appears once it is answered.
	wantNoRole(t, g, mine)
	discardFromHand(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	wantRole(t, g, mine, "Wicked Role")
}

func TestEriettesWhisperRefusesMyselfAsTheDiscarder(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	if err := castCatalogSpellErr(t, g, "Eriette's Whisper", "Sorcery", erietteOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}); err == nil {
		t.Fatal("\"target opponent\" accepted the caster")
	}
}

func TestShatterTheOathDestroysAndRoles(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushVanillaCreature(g, opp.ID, "Victim", 2, 2)
	mine := pushVanillaCreature(g, me.ID, "Mine", 2, 2)
	castCatalogSpell(t, g, "Shatter the Oath", "Sorcery", shatterOathOracle, two(victim, mine))
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Victim") != uuid.Nil {
		t.Error("the target creature survived")
	}
	wantRole(t, g, mine, "Wicked Role")
}

func TestShatterTheOathDestroysAnEnchantmentWithoutARoleTarget(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	ench := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Enchantment", TypeLine: "Enchantment",
		Owner: opp.ID, Controller: opp.ID,
	})
	castCatalogSpell(t, g, "Shatter the Oath", "Sorcery", shatterOathOracle, cardRefs(ench))
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Their Enchantment") != uuid.Nil {
		t.Error("the enchantment survived")
	}
}

func TestShatterTheOathRefusesALand(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	land := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest",
		Owner: opp.ID, Controller: opp.ID,
	})
	if err := castCatalogSpellErr(t, g, "Shatter the Oath", "Sorcery", shatterOathOracle, cardRefs(land)); err == nil {
		t.Fatal("a land was a legal target")
	}
}

func TestTwistedFealtyStealsAndRolesAnyCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
	other := pushVanillaCreature(g, opp.ID, "Other", 2, 2)
	castCatalogSpell(t, g, "Twisted Fealty", "Sorcery", twistedFealtyOracle, two(theirs, other))
	passPriorityAroundTable(t, g)
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == theirs && c.Controller != me.ID {
				t.Error("the creature was not stolen")
			}
		}
	})
	if !containsString(effectiveAbilities(t, g, theirs), "haste") {
		t.Error("the stolen creature lacks haste")
	}
	wantRole(t, g, other, "Wicked Role")
}

func TestWitchsMarkRolesAndOffersTheDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := pushVanillaCreature(g, me.ID, "Mine", 2, 2)
	pitch := game.NewCard("Pitch", me.ID)
	me.Hand.PushTop(pitch)
	seedLibrary(me, "A", "B", "C")
	castCatalogSpell(t, g, "Witch's Mark", "Sorcery", witchsMarkOracle, cardRefs(mine))
	passPriorityAroundTable(t, g)
	wantRole(t, g, mine, "Wicked Role")
	if discardChoiceFor(g, me.ID) == nil {
		t.Fatal("no optional discard prompt")
	}
	before := handCount(g, me)
	answerDiscard(t, g, me.ID, pitch.InstanceID)
	passPriorityAroundTable(t, g)
	// One card out, two in.
	if got := handCount(g, me); got != before+1 {
		t.Errorf("hand = %d, want %d (discard one, draw two)", got, before+1)
	}
}

func TestWitchsMarkDecliningDrawsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Hand.PushTop(game.NewCard("Keep", me.ID))
	seedLibrary(me, "A", "B", "C")
	castCatalogSpell(t, g, "Witch's Mark", "Sorcery", witchsMarkOracle, nil)
	passPriorityAroundTable(t, g)
	before := handCount(g, me)
	answerDiscard(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := handCount(g, me); got != before {
		t.Errorf("hand changed from %d to %d after declining", before, got)
	}
}

func TestBecomeBrutesGivesHasteAndAMonsterRoleToEach(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushVanillaCreature(g, me.ID, "A", 2, 2)
	b := pushVanillaCreature(g, opp.ID, "B", 2, 2)
	castCatalogSpell(t, g, "Become Brutes", "Sorcery", becomeBrutesOracle, two(a, b))
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		wantRole(t, g, id, "Monster Role")
		if !containsString(effectiveAbilities(t, g, id), "haste") {
			t.Errorf("creature %v lacks haste", id)
		}
	}
	if rolesOn(g, b)[0].Controller != me.ID {
		t.Error("the Role on the opponent's creature is not the caster's")
	}
}

func TestBecomeBrutesRefusesThreeTargets(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := pushVanillaCreature(g, me.ID, "A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "B", 2, 2)
	c := pushVanillaCreature(g, me.ID, "C", 2, 2)
	if err := castCatalogSpellErr(t, g, "Become Brutes", "Sorcery", becomeBrutesOracle, cardRefs(a, b, c)); err == nil {
		t.Fatal("three targets accepted for \"one or two\"")
	}
}

func TestAsinineAnticsCursesEveryOpposingCreatureAndNoneOfMine(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushVanillaCreature(g, me.ID, "Mine", 2, 2)
	a := pushVanillaCreature(g, opp.ID, "A", 4, 4)
	b := pushVanillaCreature(g, g.Seats[2].ID, "B", 3, 3)
	castCatalogSpell(t, g, "Asinine Antics", "Sorcery", asinineAnticsOracle, nil)
	passPriorityAroundTable(t, g)
	wantNoRole(t, g, mine)
	for _, id := range []uuid.UUID{a, b} {
		wantRole(t, g, id, "Cursed Role")
		if p, tough := effectivePower(t, g, id), effectiveToughness(t, g, id); p != 1 || tough != 1 {
			t.Errorf("cursed creature is %d/%d, want 1/1", p, tough)
		}
	}
}

func TestCurseOfTheWerefoxRolesThenFights(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushVanillaCreature(g, me.ID, "Mine", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 3, 3)
	castCatalogSpell(t, g, "Curse of the Werefox", "Sorcery", curseWerefoxOracle, cardRefs(mine))
	passPriorityAroundTable(t, g)
	wantRole(t, g, mine, "Monster Role")
	if latestPickTarget(g, me.ID) == nil {
		t.Fatal("no target prompt for the reflexive fight")
	}
	pickTriggerTarget(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	// 3/3 (Role) fights a 3/3: both die.
	if findBattlefieldByName(g, "Theirs") != uuid.Nil || findBattlefieldByName(g, "Mine") != uuid.Nil {
		t.Error("the fight did not resolve between 3/3s")
	}
}

func TestCurseOfTheWerefoxRefusesAnOpponentsCreatureAsTheRoleTarget(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 3, 3)
	if err := castCatalogSpellErr(t, g, "Curse of the Werefox", "Sorcery", curseWerefoxOracle, cardRefs(theirs)); err == nil {
		t.Fatal("the Role clause accepted a creature I don't control")
	}
}

// --- creatures ------------------------------------------------------

func TestDiminisherWitchUnbargainedDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 4, 4)
	castAndResolveCreature(t, g, "Diminisher Witch", "Creature — Human Warlock", diminisherOracle)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, g.Seats[0].ID) != nil {
		t.Error("an unbargained Witch asked for a target")
	}
	wantNoRole(t, g, theirs)
}

func TestDiminisherWitchBargainedCursesAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 4, 4)
	token := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Treasure", TypeLine: "Token Artifact — Treasure",
		Owner: me.ID, Controller: me.ID,
	})
	p1g2Cast(t, g, "Diminisher Witch", "Creature — Human Warlock", diminisherOracle, "", game.CastSpellParams{
		OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{token},
	})
	passPriorityAroundTable(t, g)
	pickTriggerTarget(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	wantRole(t, g, theirs, "Cursed Role")
	if p := effectivePower(t, g, theirs); p != 1 {
		t.Errorf("cursed creature has power %d, want 1", p)
	}
}

func TestTwistedSewerWitchRolesTheNewRatAndEveryOtherRat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	oldRat := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Old Rat", TypeLine: "Creature — Rat",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	witch := castAndResolveCreature(t, g, "Twisted Sewer-Witch", "Creature — Human Warlock", sewerWitchOracle)
	passPriorityAroundTable(t, g)

	wantRole(t, g, oldRat, "Wicked Role")
	wantNoRole(t, g, bear)
	wantNoRole(t, g, witch)
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
	wantRole(t, g, newRat.InstanceID, "Wicked Role")
	// 1/1 + 1/+0.
	if p := effectivePower(t, g, newRat.InstanceID); p != 2 {
		t.Errorf("the new Rat has power %d, want 2", p)
	}
	assertRestrictions(t, g, newRat.InstanceID, game.CantBlock)
}

func TestCharmingScoundrelEachBullet(t *testing.T) {
	pick := func(t *testing.T, mode int) (*game.Game, *game.Player, uuid.UUID) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		id := castAndResolveCreature(t, g, "Charming Scoundrel", "Creature — Human Rogue", scoundrelOracle)
		c := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
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
		seedLibrary(me, "A", "B")
		passPriorityAroundTable(t, g)
		before := handCount(g, me)
		discardFromHand(t, g, me.ID)
		passPriorityAroundTable(t, g)
		if got := handCount(g, me); got != before {
			t.Errorf("hand %d, want %d (discard one, draw one)", got, before)
		}
	})
	t.Run("treasure", func(t *testing.T) {
		g, _, _ := pick(t, 1)
		passPriorityAroundTable(t, g)
		if findBattlefieldByName(g, "Treasure") == uuid.Nil {
			t.Error("no Treasure")
		}
	})
	t.Run("role", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
		castAndResolveCreature(t, g, "Charming Scoundrel", "Creature — Human Rogue", scoundrelOracle)
		c := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if c == nil {
			t.Fatal("no mode prompt")
		}
		if err := g.ResolveModePick(c.ID, me.ID, []int{2}); err != nil {
			t.Fatalf("ResolveModePick: %v", err)
		}
		pickTriggerTarget(t, g, me.ID, bear)
		passPriorityAroundTable(t, g)
		wantRole(t, g, bear, "Wicked Role")
	})
}

func TestGylwainRolesTheEnteringCreatureAndItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	gylwain := castAndResolveCreature(t, g, "Gylwain, Casting Director", "Legendary Creature — Human Bard", gylwainOracle)
	c := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if c == nil {
		t.Fatalf("no mode prompt for Gylwain's own entry: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	passPriorityAroundTable(t, g)
	wantRole(t, g, gylwain, "Royal Role")

	// Another nontoken creature entering asks again, and each bullet is a
	// different Role.
	castAndResolveCreature(t, g, "Grizzly Bears", "Creature — Bear", "")
	c = latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if c == nil {
		t.Fatal("no mode prompt for another creature entering")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	passPriorityAroundTable(t, g)
	bears := findBattlefieldByName(g, "Grizzly Bears")
	wantRole(t, g, bears, "Sorcerer Role")
}

func TestGylwainIgnoresTokensAndOpponentsCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Gylwain, Casting Director", "Legendary Creature — Human Bard", gylwainOracle, false)
	// An opponent's creature entering does not trigger it.
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Theirs", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	passPriorityAroundTable(t, g)
	if c := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID); c != nil {
		t.Error("an opponent's creature entering prompted Gylwain")
	}
}

// --- sagas ----------------------------------------------------------

func TestTheWitchsVanityChapters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cheap := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Cheap", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	pricey := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Pricey", TypeLine: "Creature — Bear", ManaCost: "{2}{G}",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	mine := pushVanillaCreature(g, me.ID, "Mine", 2, 2)
	seat := g.Turn.ActiveSeat

	castCatalogSpell(t, g, "The Witch's Vanity", "Enchantment — Saga", witchsVanityOracle, nil)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		pickTriggerTarget(t, g, me.ID, cheap)
		passPriorityAroundTable(t, g)
	}
	if findBattlefieldByName(g, "Cheap") != uuid.Nil {
		t.Error("chapter I did not destroy the mana value 2 creature")
	}
	if findBattlefieldByName(g, "Pricey") == uuid.Nil {
		t.Error("chapter I destroyed a mana value 3 creature")
	}
	_ = pricey

	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Food") == uuid.Nil {
		t.Error("chapter II made no Food")
	}

	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		pickTriggerTarget(t, g, me.ID, mine)
		passPriorityAroundTable(t, g)
	}
	wantRole(t, g, mine, "Wicked Role")
}

func TestTheWitchsVanityChapterOneRefusesAThreeDrop(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Pricey", TypeLine: "Creature — Bear", ManaCost: "{2}{G}",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	castCatalogSpell(t, g, "The Witch's Vanity", "Enchantment — Saga", witchsVanityOracle, nil)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, g.Seats[0].ID) != nil {
		t.Error("chapter I offered a target when none has mana value 2 or less")
	}
	if findBattlefieldByName(g, "Pricey") == uuid.Nil {
		t.Error("a mana value 3 creature was destroyed")
	}
}

func TestGadwicksFirstDuelChapters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 4, 4)
	seedLibrary(me, "A", "B", "C", "D")
	seat := g.Turn.ActiveSeat

	castCatalogSpell(t, g, "Gadwick's First Duel", "Enchantment — Saga", gadwicksDuelOracle, nil)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) == nil {
		t.Fatal("no target prompt for chapter I")
	}
	pickTriggerTarget(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	wantRole(t, g, theirs, "Cursed Role")

	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) == nil {
		t.Fatal("chapter II queued no scry")
	}
	answerScryKeepAll(t, g, me.ID)

	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("chapter III left %d delayed triggers, want 1", len(g.DelayedTriggers))
	}
}

func TestGadwicksDelayedCopyOnlyFiresForAThreeOrLessSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0].ID, g.Seats[1].ID
	before := lifeOf(g, victim)
	g.WithWriteLock(func() {
		err := DelayedOnEvent{
			Label:      "Gadwick's First Duel — copy that spell",
			On:         []game.EventKind{game.EventCast},
			Condition:  gadwickNextCastCondition,
			CondParams: game.EffectParams{Filter: game.CastFilter{Types: []string{"Instant", "Sorcery"}}, Amount: 3},
			Body:       copyTheSpellBody,
		}.Apply(NewContext(g, &game.StackItem{Controller: me}))
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
	passPriorityAroundTable(t, g)
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("a mana value 4 spell used the trigger up (%d left)", len(g.DelayedTriggers))
	}

	// A Lightning Bolt (mana value 1) is copied.
	castBoltAtAndKeepTheCopysTarget(t, g, me, victim, 1)
	if got := lifeOf(g, victim); got != before-6 {
		t.Errorf("victim life = %d, want %d (the Bolt and its copy)", got, before-6)
	}
}

// --- enchantments ---------------------------------------------------

func TestLordSkittersBlessingRolesOnEntryAndDrawsExtraWhileEnchanted(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := pushVanillaCreature(g, me.ID, "Mine", 2, 2)
	seedLibrary(me, "A", "B", "C", "D", "E")
	castCatalogSpell(t, g, "Lord Skitter's Blessing", "Enchantment", skittersOracle, nil)
	passPriorityAroundTable(t, g)
	pickTriggerTarget(t, g, me.ID, mine)
	passPriorityAroundTable(t, g)
	wantRole(t, g, mine, "Wicked Role")

	// Next turn's draw step: lose 1 life, draw an additional card.
	seat := g.Turn.ActiveSeat
	life, hand := me.Life, handCount(g, me)
	for i := 0; i < 400; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
		if g.Turn.Step == game.StepDraw && g.Turn.ActiveSeat == seat {
			break
		}
	}
	passPriorityAroundTable(t, g)
	if me.Life != life-1 {
		t.Errorf("life = %d, want %d", me.Life, life-1)
	}
	if got := handCount(g, me); got != hand+2 {
		t.Errorf("hand = %d, want %d (the draw step's card and one more)", got, hand+2)
	}
}

func TestLordSkittersBlessingDoesNothingInTheDrawStepWithoutAnEnchantedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Lord Skitter's Blessing", "Enchantment", skittersOracle, false)
	pushVanillaCreature(g, me.ID, "Bare", 2, 2)
	seedLibrary(me, "A", "B", "C", "D", "E")
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
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life = %d, want %d — nothing is enchanted", me.Life, life)
	}
}

func TestGiantInheritanceGrowsTheCreatureAndRolesAnAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	friend := pushVanillaCreature(g, me.ID, "Friend", 2, 2)
	castCatalogSpell(t, g, "Giant Inheritance", "Enchantment — Aura", giantInheritOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 7 || tough != 7 {
		t.Fatalf("enchanted bear is %d/%d, want 7/7", p, tough)
	}

	declareAttack(t, g, opp.ID, bear, friend)
	if latestPickTarget(g, me.ID) == nil {
		t.Fatal("no target prompt for the granted attack trigger")
	}
	pickTriggerTarget(t, g, me.ID, friend)
	if triggersOnStackFrom(g, bear) != 1 {
		t.Fatalf("granted triggers from the bear = %d, want 1 (the Aura is not the source)", triggersOnStackFrom(g, bear))
	}
	passPriorityAroundTable(t, g)
	wantRole(t, g, friend, "Monster Role")
}

func TestGiantInheritanceReturnsToHandWhenItIsPutIntoAGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	aura := castCatalogSpell(t, g, "Giant Inheritance", "Enchantment — Aura", giantInheritOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	castCatalogSpell(t, g, "Not Dead After All", "Instant", notDeadOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)

	back := findBattlefieldByName(g, "Bear")
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
	wantRole(t, g, back, "Wicked Role")
}

func TestNotDeadAfterAllGivesATokenNoRoleAndNoReturn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tok := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Tok", TypeLine: "Token Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	castCatalogSpell(t, g, "Not Dead After All", "Instant", notDeadOracle, cardRefs(tok))
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(tok) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Tok") != uuid.Nil {
		t.Error("a token came back")
	}
	if findBattlefieldByName(g, "Wicked Role") != uuid.Nil {
		t.Error("a Role was created for a token that never returned")
	}
}

func TestNotDeadAfterAllGrantEndsWithTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	castCatalogSpell(t, g, "Not Dead After All", "Instant", notDeadOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	advanceToPrecombatMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Bear") != uuid.Nil {
		t.Error("the grant outlived the turn")
	}
}

// --- Adventures -----------------------------------------------------

type adventureCase struct {
	name, adv, oracle, role string
	faces                   [2]game.Face
}

func adventureCard(owner uuid.UUID, c adventureCase) game.Card {
	card := game.Card{
		InstanceID: uuid.New(), OracleID: c.oracle, Layout: game.LayoutAdventure,
		Owner: owner, Controller: owner,
		Faces: []game.Face{c.faces[0], c.faces[1]},
	}
	card.SetFace(0)
	return card
}

func TestAdventureHalvesCreateTheirRoles(t *testing.T) {
	cases := []adventureCase{
		{name: "Besotted Knight", adv: "Betroth the Beast", oracle: besottedKnightOracle, role: "Royal Role",
			faces: [2]game.Face{
				{Name: "Besotted Knight", TypeLine: "Creature — Human Knight", ManaCost: "{3}{W}", Colors: []string{"W"}, Power: 3, Toughness: 3},
				{Name: "Betroth the Beast", TypeLine: "Sorcery — Adventure", ManaCost: "{W}", Colors: []string{"W"}},
			}},
		{name: "Conceited Witch", adv: "Price of Beauty", oracle: conceitedOracle, role: "Wicked Role",
			faces: [2]game.Face{
				{Name: "Conceited Witch", TypeLine: "Creature — Human Warlock", ManaCost: "{2}{B}", Colors: []string{"B"}, Power: 2, Toughness: 3},
				{Name: "Price of Beauty", TypeLine: "Sorcery — Adventure", ManaCost: "{B}", Colors: []string{"B"}},
			}},
		{name: "Ferocious Werefox", adv: "Guard Change", oracle: werefoxOracle, role: "Monster Role",
			faces: [2]game.Face{
				{Name: "Ferocious Werefox", TypeLine: "Creature — Elf Fox Warrior", ManaCost: "{3}{G}", Colors: []string{"G"}, Power: 4, Toughness: 3},
				{Name: "Guard Change", TypeLine: "Instant — Adventure", ManaCost: "{1}{G}", Colors: []string{"G"}},
			}},
	}
	for _, tc := range cases {
		t.Run(tc.adv, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
			card := adventureCard(me.ID, tc)
			me.Hand.PushTop(card)
			advanceTo(t, g, game.StepPrecombatMain)
			if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, Targets: cardRefs(bear)}); err != nil {
				t.Fatalf("cast %s: %v", tc.adv, err)
			}
			passPriorityAroundTable(t, g)
			wantRole(t, g, bear, tc.role)
			if !g.Exile.Contains(card.InstanceID) {
				t.Error("the Adventure card is not in exile (CR 715.3d)")
			}
		})
	}
}

func TestCroakingCurseTapsAndCursesAnyCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 4, 4)
	card := adventureCard(me.ID, adventureCase{oracle: transmuterOracle, faces: [2]game.Face{
		{Name: "Vantress Transmuter", TypeLine: "Creature — Human Wizard", ManaCost: "{3}{U}", Colors: []string{"U"}, Power: 3, Toughness: 4},
		{Name: "Croaking Curse", TypeLine: "Sorcery — Adventure", ManaCost: "{1}{U}", Colors: []string{"U"}},
	}})
	me.Hand.PushTop(card)
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, Targets: cardRefs(theirs)}); err != nil {
		t.Fatalf("cast Croaking Curse: %v", err)
	}
	passPriorityAroundTable(t, g)
	wantRole(t, g, theirs, "Cursed Role")
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
	card := adventureCard(me.ID, adventureCase{oracle: besottedKnightOracle, faces: [2]game.Face{
		{Name: "Besotted Knight", TypeLine: "Creature — Human Knight", ManaCost: "{3}{W}", Colors: []string{"W"}, Power: 3, Toughness: 3},
		{Name: "Betroth the Beast", TypeLine: "Sorcery — Adventure", ManaCost: "{W}", Colors: []string{"W"}},
	}})
	me.Hand.PushTop(card)
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, Targets: cardRefs(theirs)}); err == nil {
		t.Fatal("an opponent's creature was a legal target for \"creature you control\"")
	}
}
