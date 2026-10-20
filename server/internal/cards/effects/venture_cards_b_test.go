package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// venture_cards_b_test.go — the second half of the venture cards (#4100):
// one test per card, the venture and every other printed clause.

const (
	intrepidOutlanderOracle     = "ff8d3532-bf4f-4209-ac01-351ddc9b9423"
	radiantSolarOracle          = "21a86dcd-03e3-409b-b214-5feed07db584"
	fiftyFeetOfRopeOracle       = "66aa6905-86f9-4800-988d-ffebafd4a9dc"
	barrowinOracle              = "c16d5daf-4995-4342-815c-a6b29d594f26"
	varisOracle                 = "bd751b25-9ea7-4ce4-b532-48be281c6da9"
	eccentricApprenticeOracle   = "27dbe95e-6274-416a-a45b-a2a0a64d5403"
	dungeonCrawlerOracle        = "26b53524-9ec5-4ec9-a1e2-73e78959481b"
	delversTorchOracle          = "a1c8b833-8c74-4f95-8a47-1dd41ef99361"
	sefrisOracle                = "9b21e056-57e7-4e39-b2b4-2fdf8719d0d2"
	findThePathOracle           = "860eb9eb-c318-4eae-9b54-088f667115ec"
	precipitousDropOracle       = "d60c8cfe-435c-419e-a44b-b47d49d1afef"
	thoroughInvestigationOracle = "e7504407-bb72-4baf-be4a-8152e4198e2b"
	flyOracle                   = "da1b633f-82b5-4f3e-884c-58c2fe4fcb0f"
	hamaPasharOracle            = "c7371040-160c-4502-b7f9-77b7f54cc020"
	immovableRodOracle          = "ec664e13-b7c4-4ca3-80d6-55a564d97937"
	acererakOracle              = "31b4e85d-89fc-4712-8f05-7ec097b1d3b6"
	ellywickOracle              = "d0375311-b0ea-4870-bc2c-4894da088dc3"
	keenEaredSentryOracle       = "a5575201-0d73-43c4-9290-940bbaefc61b"
)

// vcAttack declares `attackers` at `defender` and leaves the declaration
// so the attack triggers go on the stack.
func vcAttack(t *testing.T, g *game.Game, defender uuid.UUID, attackers ...uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range attackers {
		if err := g.DeclareAttacker(id, defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	_, _ = g.AdvanceStep()
}

// vcPermanent puts a catalog permanent onto the battlefield through the
// layer listener.
func vcPermanent(g *game.Game, owner uuid.UUID, name, oracle, typeLine string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, OracleID: oracle, TypeLine: typeLine,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// --- Intrepid Outlander ------------------------------------------------

func TestIntrepidOutlanderVenturesWithPackTactics(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	out := pushDiesCreatureForTest(g, me.ID, "Intrepid Outlander", intrepidOutlanderOracle, "Creature — Orc Ranger", 2, 3)
	big := pushVanillaCreature(g, me.ID, "Big", 4, 4)
	vcAttack(t, g, them.ID, out, big)
	vcVentured(t, g, me.ID)

	g2 := newCatalogGame(t)
	me2, them2 := g2.Seats[0], g2.Seats[1]
	out2 := pushDiesCreatureForTest(g2, me2.ID, "Intrepid Outlander", intrepidOutlanderOracle, "Creature — Orc Ranger", 2, 3)
	vcAttack(t, g2, them2.ID, out2)
	vcNotVentured(t, g2, me2.ID)
}

// --- Radiant Solar ------------------------------------------------------

func TestRadiantSolarVenturesForItselfAndNontokenCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castAndResolveCreature(t, g, "Radiant Solar", "Creature — Angel", radiantSolarOracle)
	vcVentured(t, g, me.ID)

	castAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	if left := frSettle(t, g, "goblin-lair"); len(left) > 0 {
		t.Fatalf("a nontoken creature entering did not venture")
	}
	g.WithWriteLock(func() {
		if err := (CreateToken{Controller: me.ID, Template: TokenCard("1/1 red Goblin"), N: 1}).Apply(&Context{Game: g}); err != nil {
			t.Fatal(err)
		}
	})
	frSettle(t, g)
	if c := latestOptionPickFor(g, me.ID); c != nil {
		t.Fatalf("a token entering ventured: %q", c.Reason)
	}
}

func TestRadiantSolarDiscardsToVentureAndGainThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	solar := handCardFull(me, "Radiant Solar", "Creature — Angel", "{5}{W}", radiantSolarOracle, nil)
	conduitMana(t, g, me, "{W}")
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, solar, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate from hand: %v", err)
	}
	if !me.Graveyard.Contains(solar) {
		t.Errorf("the Solar was not discarded")
	}
	vcVentured(t, g, me.ID)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

// --- Fifty Feet of Rope ---------------------------------------------------

func TestFiftyFeetOfRope(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	rope := vcPermanent(g, me.ID, "Fifty Feet of Rope", fiftyFeetOfRopeOracle, "Artifact", 0, 0)
	wall := pushDiesCreatureForTest(g, them.ID, "Their Wall", "", "Creature — Wall", 0, 4)
	bear := pushVanillaCreature(g, them.ID, "Their Bear", 2, 2)

	// Climb Over: only a Wall may be targeted, and it can't block.
	if err := activateOn(t, g, me.ID, rope, 0, bear); err == nil {
		t.Errorf("Climb Over targeted a non-Wall")
	}
	if err := activateOn(t, g, me.ID, rope, 0, wall); err != nil {
		t.Fatalf("Climb Over: %v", err)
	}
	frSettle(t, g)
	g.ReadSnapshot(func() {
		if findBattlefieldCardForTest(g, wall).Effective().Restrictions&game.CantBlock == 0 {
			t.Errorf("the Wall can still block")
		}
	})

	// Rappel Down, at sorcery speed.
	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	rope2 := vcPermanent(g2, me2.ID, "Fifty Feet of Rope", fiftyFeetOfRopeOracle, "Artifact", 0, 0)
	advanceToMain(t, g2)
	conduitMana(t, g2, me2, "{C}{C}{C}{C}")
	if err := activateOn(t, g2, me2.ID, rope2, 2, uuid.Nil); err != nil {
		t.Fatalf("Rappel Down: %v", err)
	}
	vcVentured(t, g2, me2.ID)

	// Tie Up: the tapped creature stays tapped through its controller's
	// next untap step.
	g3 := newCatalogGame(t)
	me3, them3 := g3.Seats[0], g3.Seats[1]
	rope3 := vcPermanent(g3, me3.ID, "Fifty Feet of Rope", fiftyFeetOfRopeOracle, "Artifact", 0, 0)
	tied := pushVanillaCreature(g3, them3.ID, "Tied Bear", 2, 2)
	for i := range g3.Battlefield.Cards {
		if g3.Battlefield.Cards[i].InstanceID == tied {
			g3.Battlefield.Cards[i].Tapped = true
		}
	}
	conduitMana(t, g3, me3, "{C}{C}{C}")
	if err := activateOn(t, g3, me3.ID, rope3, 1, tied); err != nil {
		t.Fatalf("Tie Up: %v", err)
	}
	frSettle(t, g3)
	advanceToUpkeepOf(t, g3, 1)
	if !findBattlefieldCardForTest(g3, tied).Tapped {
		t.Errorf("the tied creature untapped")
	}
}

// --- Barrowin of Clan Undurr ---------------------------------------------

func TestBarrowinVenturesAndReturnsASmallCreatureOnceADungeonIsDone(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	castAndResolveCreature(t, g, "Barrowin of Clan Undurr", "Legendary Creature — Dwarf Cleric", barrowinOracle)
	vcVentured(t, g, me.ID)

	// Attacking with no dungeon completed returns nothing.
	g2 := newCatalogGame(t)
	me2, them2 := g2.Seats[0], g2.Seats[1]
	b2 := pushDiesCreatureForTest(g2, me2.ID, "Barrowin of Clan Undurr", barrowinOracle, "Legendary Creature — Dwarf Cleric", 3, 3)
	b17GraveyardCard(me2, "Dead Bear", "Creature — Bear", "{1}{G}")
	vcAttack(t, g2, them2.ID, b2)
	frSettle(t, g2)
	if c := latestChooseCardsFor(g2, me2.ID); c != nil {
		t.Fatalf("asked to return a card with no dungeon completed")
	}

	// With one completed: a creature card of mana value 3 or less, not 4.
	g3 := newCatalogGame(t)
	me3, them3 := g3.Seats[0], g3.Seats[1]
	b3 := pushDiesCreatureForTest(g3, me3.ID, "Barrowin of Clan Undurr", barrowinOracle, "Legendary Creature — Dwarf Cleric", 3, 3)
	small := b17GraveyardCard(me3, "Dead Bear", "Creature — Bear", "{1}{G}")
	large := b17GraveyardCard(me3, "Dead Giant", "Creature — Giant", "{3}{G}")
	vcCompleteADungeon(t, g3, me3.ID)
	vcAttack(t, g3, them3.ID, b3)
	frSettle(t, g3)
	c := latestChooseCardsFor(g3, me3.ID)
	if c == nil {
		t.Fatalf("no return offered")
	}
	if len(c.ChooseCards) != 1 || c.ChooseCards[0] != small {
		t.Fatalf("offered %v, want only the mana value 2 creature (not %v)", c.ChooseCards, large)
	}
	if err := g3.ResolveChooseCards(c.ID, me3.ID, []uuid.UUID{small}); err != nil {
		t.Fatal(err)
	}
	frSettle(t, g3)
	if !g3.Battlefield.Contains(small) {
		t.Errorf("the creature card did not return to the battlefield")
	}
	_ = them
}

// --- Varis, Silverymoon Ranger --------------------------------------------

func TestVarisVenturesOnceEachTurnAndMakesAWolf(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushDiesCreatureForTest(g, me.ID, "Varis, Silverymoon Ranger", varisOracle, "Legendary Creature — Human Elf Ranger", 3, 3)
	castAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	vcVentured(t, g, me.ID)
	castAndResolveCreature(t, g, "Second Bear", "Creature — Bear", "")
	vcNotVenturedAgain(t, g, me.ID, "cave-entrance")

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	pushDiesCreatureForTest(g2, me2.ID, "Varis, Silverymoon Ranger", varisOracle, "Legendary Creature — Human Elf Ranger", 3, 3)
	vcCompleteADungeon(t, g2, me2.ID)
	frSettle(t, g2)
	if n := frTokens(g2, me2.ID, "Wolf"); n != 1 {
		t.Errorf("Wolves = %d after completing a dungeon, want 1", n)
	}
}

// vcNotVenturedAgain settles and fails unless `player`'s marker is still
// in `room` with no venture prompt open.
func vcNotVenturedAgain(t *testing.T, g *game.Game, player uuid.UUID, room string) {
	t.Helper()
	frSettle(t, g)
	if c := latestOptionPickFor(g, player); c != nil {
		t.Fatalf("ventured again: asked %q", c.Reason)
	}
	if got := initRoom(g, player); got != room {
		t.Fatalf("room = %q, want %q", got, room)
	}
}

// --- Eccentric Apprentice -----------------------------------------------

func TestEccentricApprenticeTurnsACreatureIntoABird(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	giant := pushDiesCreatureForTest(g, them.ID, "Their Giant", "", "Creature — Giant", 5, 5)
	vcCompleteADungeon(t, g, me.ID)
	castAndResolveCreature(t, g, "Eccentric Apprentice", "Creature — Tiefling Wizard", eccentricApprenticeOracle)
	vcVentured(t, g, me.ID)
	advanceTo(t, g, game.StepBeginCombat)
	if left := frSettle(t, g, frCard(giant)); len(left) > 0 {
		t.Fatalf("the beginning-of-combat trigger asked for no target")
	}
	if p, tough := effectivePower(t, g, giant), effectiveToughness(t, g, giant); p != 1 || tough != 1 {
		t.Errorf("the Giant is %d/%d, want 1/1", p, tough)
	}
	if !hasAbility(effectiveAbilities(t, g, giant), "flying") {
		t.Errorf("the Bird has no flying")
	}
	if subs := effectiveSubtypes(t, g, giant); len(subs) != 1 || subs[0] != "Bird" {
		t.Errorf("subtypes = %v, want [Bird]", subs)
	}
}

func TestEccentricApprenticeDoesNothingBeforeADungeonIsDone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushDiesCreatureForTest(g, me.ID, "Eccentric Apprentice", eccentricApprenticeOracle, "Creature — Tiefling Wizard", 2, 2)
	advanceTo(t, g, game.StepBeginCombat)
	frSettle(t, g)
	if c := latestChoiceOfKindFor(g, game.PendingChoicePickTarget, me.ID); c != nil {
		t.Fatalf("the trigger fired with no dungeon completed")
	}
}

// --- Dungeon Crawler ------------------------------------------------------

func TestDungeonCrawlerReturnsFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	crawler := game.Card{InstanceID: uuid.New(), Name: "Dungeon Crawler", OracleID: dungeonCrawlerOracle,
		TypeLine: "Creature — Zombie", Power: 2, Toughness: 1, Owner: me.ID, Controller: me.ID}
	me.Graveyard.PushTop(crawler)
	vcCompleteADungeon(t, g, me.ID)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	frSettle(t, g)
	if z := initZone(g, crawler.InstanceID); z != game.ZoneHand {
		t.Fatalf("Dungeon Crawler is in %v, want hand", z)
	}
}

// --- Delver's Torch -------------------------------------------------------

func TestDelversTorchPumpsAndVenturesOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	torch := vcPermanent(g, me.ID, "Delver's Torch", delversTorchOracle, "Artifact — Equipment", 0, 0)
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(torch, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
			t.Fatal(err)
		}
	})
	if p := effectivePower(t, g, bear); p != 3 {
		t.Errorf("equipped power %d, want 3", p)
	}
	vcAttack(t, g, them.ID, bear)
	vcVentured(t, g, me.ID)
}

// --- Sefris of the Hidden Ways ----------------------------------------------

func TestSefrisVenturesOnceEachTurnAndRaisesTheDead(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushDiesCreatureForTest(g, me.ID, "Sefris of the Hidden Ways", sefrisOracle, "Legendary Creature — Human Wizard", 2, 3)
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(a); err != nil {
			t.Fatal(err)
		}
	})
	vcVentured(t, g, me.ID)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(b); err != nil {
			t.Fatal(err)
		}
	})
	vcNotVenturedAgain(t, g, me.ID, "cave-entrance")

	// Create Undead: completing a dungeon returns a creature card.
	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	pushDiesCreatureForTest(g2, me2.ID, "Sefris of the Hidden Ways", sefrisOracle, "Legendary Creature — Human Wizard", 2, 3)
	dead := b17GraveyardCard(me2, "Dead Giant", "Creature — Giant", "{5}{G}")
	vcCompleteADungeon(t, g2, me2.ID)
	if left := frSettle(t, g2, frCard(dead)); len(left) > 0 {
		t.Fatalf("Create Undead asked for no target")
	}
	if !g2.Battlefield.Contains(dead) {
		t.Errorf("Create Undead did not return the creature card")
	}
}

// --- Find the Path ---------------------------------------------------------

func TestFindThePathVenturesAndGivesTheLandGG(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Wasteland Plain", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
	pushAuraOnLand(t, g, me.ID, "Find the Path", findThePathOracle, land)
	tapForMana(t, g, me.ID, land)
	if got := me.ManaPool.Total(); got != 2 {
		t.Errorf("the enchanted land made %d mana, want {G}{G}", got)
	}

	g2 := newCatalogGame(t)
	land2 := pushBattlefieldCardWithTimestamp(g2, game.Card{InstanceID: uuid.New(), Name: "Wasteland Plain", TypeLine: "Land", Owner: g2.Seats[0].ID, Controller: g2.Seats[0].ID})
	castCatalogSpell(t, g2, "Find the Path", "Enchantment — Aura", findThePathOracle, b16TargetCard(land2))
	vcVentured(t, g2, g2.Seats[0].ID)
}

// --- Precipitous Drop -------------------------------------------------------

func TestPrecipitousDropShrinksMoreOnceADungeonIsDone(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	giant := pushVanillaCreature(g, them.ID, "Their Giant", 6, 6)
	pushAuraOnLand(t, g, me.ID, "Precipitous Drop", precipitousDropOracle, giant)
	if p, tough := effectivePower(t, g, giant), effectiveToughness(t, g, giant); p != 4 || tough != 4 {
		t.Fatalf("enchanted Giant %d/%d, want 4/4", p, tough)
	}
	vcCompleteADungeon(t, g, me.ID)
	if p, tough := effectivePower(t, g, giant), effectiveToughness(t, g, giant); p != 1 || tough != 1 {
		t.Errorf("after completing, %d/%d, want 1/1", p, tough)
	}
}

// --- Thorough Investigation -------------------------------------------------

func TestThoroughInvestigationInvestigatesAndVenturesOnAClue(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	vcPermanent(g, me.ID, "Thorough Investigation", thoroughInvestigationOracle, "Enchantment", 0, 0)
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	vcAttack(t, g, them.ID, a, b)
	frSettle(t, g)
	if n := frTokens(g, me.ID, "Clue"); n != 1 {
		t.Fatalf("Clues = %d, want 1 for one attack", n)
	}
	clue := findBattlefieldByName(g, "Clue")
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(clue); err != nil {
			t.Fatal(err)
		}
	})
	vcVentured(t, g, me.ID)
}

// --- Fly --------------------------------------------------------------------

func TestFlyGivesFlyingAndAVentureOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	pushAuraOnLand(t, g, me.ID, "Fly", flyOracle, bear)
	if !hasAbility(effectiveAbilities(t, g, bear), "flying") {
		t.Fatalf("the enchanted creature has no flying")
	}
	attackWith(t, g, them.ID, bear)
	vcVentured(t, g, me.ID)
}

// --- Hama Pashar, Ruin Seeker -------------------------------------------------

func TestHamaPasharDoublesRoomAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	pushDiesCreatureForTest(g, me.ID, "Hama Pashar, Ruin Seeker", hamaPasharOracle, "Legendary Creature — Human Wizard", 2, 3)
	frVenture(t, g, me.ID, LostMineOfPhandelverKey)
	frVenture(t, g, me.ID, "goblin-lair")
	if n := frTokens(g, me.ID, "Goblin"); n != 2 {
		t.Fatalf("Goblin Lair made %d Goblins under Hama Pashar, want 2", n)
	}
	// An opponent's rooms are not doubled.
	frVenture(t, g, them.ID, LostMineOfPhandelverKey)
	frVenture(t, g, them.ID, "goblin-lair")
	if n := frTokens(g, them.ID, "Goblin"); n != 1 {
		t.Fatalf("an opponent's Goblin Lair made %d Goblins, want 1", n)
	}
}

// --- Immovable Rod ------------------------------------------------------------

func TestImmovableRodLocksAPermanentAndVenturesWhenItUntaps(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	rod := vcPermanent(g, me.ID, "Immovable Rod", immovableRodOracle, "Artifact", 0, 0)
	flier := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Flier", TypeLine: "Creature — Bird",
		Power: 2, Toughness: 2, Keywords: []string{"flying"}, Owner: them.ID, Controller: them.ID})
	conduitMana(t, g, me, "{W}{C}{C}{C}")
	if err := activateOn(t, g, me.ID, rod, 0, flier); err != nil {
		t.Fatalf("activate: %v", err)
	}
	frSettle(t, g)
	if hasAbility(effectiveAbilities(t, g, flier), "flying") {
		t.Errorf("the target kept its abilities")
	}
	g.ReadSnapshot(func() {
		if findBattlefieldCardForTest(g, flier).Effective().Restrictions&game.CantAttackOrBlock != game.CantAttackOrBlock {
			t.Errorf("the target can still attack or block")
		}
	})
	g.WithWriteLock(func() {
		if err := g.UntapTargetForEffect(rod); err != nil {
			t.Fatal(err)
		}
	})
	vcVentured(t, g, me.ID)
	if !hasAbility(effectiveAbilities(t, g, flier), "flying") {
		t.Errorf("the lock outlasted the Rod's untapping")
	}
}

// --- Acererak the Archlich ------------------------------------------------------

func TestAcererakBouncesAndVenturesUntilTheTombIsDone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := castAndResolveCreature(t, g, "Acererak the Archlich", "Legendary Creature — Zombie Wizard", acererakOracle)
	vcVentured(t, g, me.ID)
	if z := initZone(g, id); z != game.ZoneHand {
		t.Fatalf("Acererak is in %v, want hand", z)
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	me2.CompletedDungeons = append(me2.CompletedDungeons, TombOfAnnihilationKey)
	id2 := castAndResolveCreature(t, g2, "Acererak the Archlich", "Legendary Creature — Zombie Wizard", acererakOracle)
	vcNotVentured(t, g2, me2.ID)
	if !g2.Battlefield.Contains(id2) {
		t.Fatalf("Acererak left the battlefield with Tomb of Annihilation completed")
	}
}

func TestAcererakAttackAsksEachOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, first := g.Seats[0], g.Seats[1]
	ace := pushDiesCreatureForTest(g, me.ID, "Acererak the Archlich", acererakOracle, "Legendary Creature — Zombie Wizard", 5, 5)
	victim := pushVanillaCreature(g, first.ID, "Their Bear", 2, 2)
	vcAttack(t, g, first.ID, ace)
	frSettle(t, g)
	answerOptionPick(t, g, first.ID, 1)
	answerSacrifice(t, g, first.ID, victim)
	frSettle(t, g)
	if g.Battlefield.Contains(victim) {
		t.Errorf("the first opponent's creature was not sacrificed")
	}
	// The two opponents with no creature are not asked: two Zombies.
	if n := frTokens(g, me.ID, "Zombie"); n != 2 {
		t.Errorf("Zombies = %d, want 2", n)
	}
}

// --- Ellywick Tumblestrum ----------------------------------------------------------

func TestEllywickVenturesLooksAndMakesAnEmblem(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	elly := pushWalkerForTest(g, me.ID, "Ellywick Tumblestrum", ellywickOracle, 8)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, elly, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	vcVentured(t, g, me.ID)

	// −7: the emblem.
	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	elly2 := pushWalkerForTest(g2, me2.ID, "Ellywick Tumblestrum", ellywickOracle, 8)
	bear := pushVanillaCreature(g2, me2.ID, "Bear", 2, 2)
	advanceToMain(t, g2)
	if err := g2.ActivateCatalogAbility(me2.ID, elly2, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−7: %v", err)
	}
	frSettle(t, g2)
	abs := effectiveAbilities(t, g2, bear)
	if !hasAbility(abs, "trample") || !hasAbility(abs, "haste") {
		t.Errorf("the emblem gave %v, want trample and haste", abs)
	}
	if p := effectivePower(t, g2, bear); p != 2 {
		t.Fatalf("power %d with no dungeon completed, want 2", p)
	}
	vcCompleteADungeon(t, g2, me2.ID)
	if p := effectivePower(t, g2, bear); p != 4 {
		t.Errorf("power %d with one dungeon completed, want 4", p)
	}
	vcCompleteADungeon(t, g2, me2.ID)
	if p := effectivePower(t, g2, bear); p != 4 {
		t.Errorf("power %d after completing the same dungeon twice, want 4", p)
	}
}

func TestEllywickMinusTwoTakesALegendAndGainsThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	elly := pushWalkerForTest(g, me.ID, "Ellywick Tumblestrum", ellywickOracle, 4)
	var top []uuid.UUID
	for i := 0; i < 5; i++ {
		top = append(top, frPushTop(me, "Filler", "Sorcery", "{1}"))
	}
	legend := frPushTop(me, "Legend", "Legendary Creature — Elf", "{1}{G}")
	advanceToMain(t, g)
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, elly, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−2: %v", err)
	}
	frSettle(t, g)
	c := latestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatalf("no creature offered")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{legend}); err != nil {
		t.Fatal(err)
	}
	frSettle(t, g)
	if z := initZone(g, legend); z != game.ZoneHand {
		t.Fatalf("the legend is in %v", z)
	}
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
	// The rest went to the bottom: none of them is among the top five.
	for _, id := range top {
		if z := initZone(g, id); z != game.ZoneLibrary {
			t.Errorf("a looked-at card is in %v, want the library", z)
		}
		for _, c := range me.Library.Cards[len(me.Library.Cards)-5:] {
			if c.InstanceID == id {
				t.Errorf("a looked-at card stayed on top")
			}
		}
	}
}

// --- Keen-Eared Sentry ----------------------------------------------------------

func TestKeenEaredSentryLimitsOpponentsToOneVentureATurn(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	vcPermanent(g, them.ID, "Keen-Eared Sentry", keenEaredSentryOracle, "Creature — Human Soldier", 2, 1)
	frVenture(t, g, me.ID, LostMineOfPhandelverKey)
	frVenture(t, g, me.ID)
	if room := initRoom(g, me.ID); room != "cave-entrance" {
		t.Fatalf("an opponent ventured twice in a turn: room %q", room)
	}
	// The Sentry's controller is not limited.
	frVenture(t, g, them.ID, LostMineOfPhandelverKey)
	frVenture(t, g, them.ID, "goblin-lair")
	if room := initRoom(g, them.ID); room != "goblin-lair" {
		t.Fatalf("the Sentry's controller was limited: room %q", room)
	}
}
