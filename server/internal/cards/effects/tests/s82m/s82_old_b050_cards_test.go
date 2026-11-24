package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b050_cards_test.go — Strixhaven: School of Mages, batch 50 (#5502).
// Each test plays the real catalog entry through the engine and asserts what
// the printed text says.

const (
	oldB50Mastery    = "07bfb7d2-f171-4ce1-876a-f9fb5c7cd11a"
	oldB50Onslaught  = "22195e0e-a426-48d6-b100-480c7e2bd5cc"
	oldB50Glyphweave = "4507d35a-f7f3-4abe-b06b-90c5331b533a"
	oldB50Fumes      = "784108b9-dc85-4344-8c71-d527c4f850c0"
	oldB50BlotOut    = "ded78090-c6a1-4c9d-b7b9-b433b1d963b6"
	oldB50Muse       = "30b03dda-409f-4615-b392-477d66eb49f4"
	oldB50SpecMage   = "71920c05-7061-4ba5-8903-cffd3d2f85f4"
	oldB50Mentor     = "7c4b4dd4-4a80-436a-8526-5c58c1751ccd"
	oldB50Symmetrist = "ffce416c-90c6-43ed-9019-e5a32accad6d"
)

func TestOldB50AllRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		oldB50Mastery:           "Devastating Mastery",
		oldB50Onslaught:         "Mage Hunters' Onslaught",
		oldB50Glyphweave:        "Selfless Glyphweaver",
		oldB50Glyphweave + "#1": "Deadly Vanity",
		oldB50Fumes:             "Necrotic Fumes",
		oldB50BlotOut:           "Blot Out the Sky",
		oldB50Muse:              "Maelstrom Muse",
		oldB50SpecMage:          "Spectacle Mage",
		oldB50Mentor:            "Stonebound Mentor",
		oldB50Symmetrist:        "Master Symmetrist",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

// oldB50Walker puts a planeswalker with loyalty onto the battlefield (one
// without loyalty is put into the graveyard by state-based actions).
func oldB50Walker(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Walker", TypeLine: "Legendary Planeswalker — Test",
		Owner: owner, Controller: owner, Counters: map[string]int{game.CounterLoyalty: 3},
	})
}

func oldB50OnBattlefield(g *game.Game, id uuid.UUID) bool {
	z := g.FindCardZoneForEffect(id)
	return z != nil && z.Kind == game.ZoneBattlefield
}

// --- Devastating Mastery -----------------------------------------------------

func TestOldB50DevastatingMasteryPrintedCostDestroysEveryNonlandPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Push(g, opp.ID, "Their Relic", "Artifact", "", 0, 0)
	land := effectstest.SeedLand(g, opp.ID, "Their Forest", "Basic Land — Forest", "")
	effectstest.CastCatalogSpell(t, g, "Devastating Mastery", "Sorcery", oldB50Mastery, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB50OnBattlefield(g, mine) || oldB50OnBattlefield(g, theirs) {
		t.Error("a nonland permanent survived the wipe")
	}
	if !oldB50OnBattlefield(g, land) {
		t.Error("a land was destroyed")
	}
}

func TestOldB50DevastatingMasteryAlternativeCostLetsAnOpponentReturnTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var opp *game.Player
	for _, s := range g.Seats {
		if s.ID != me.ID {
			opp = s
			break
		}
	}
	a := effectstest.B12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "Bear C", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if _, err := effectstest.CastWithTapParams(t, g, "Devastating Mastery", "Sorcery", "", oldB50Mastery,
		game.CastSpellParams{AlternativeCost: "devastating-mastery"}); err != nil {
		t.Fatalf("cast for the alternative cost: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	pick := effectstest.LatestChooseCards(g, opp.ID)
	if pick == nil {
		t.Fatal("the chosen opponent was not asked for permanents")
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 2 || len(pick.ChooseCards) != 3 {
		t.Fatalf("offered %d cards (min %d max %d), want their three with 0 to 2", len(pick.ChooseCards), pick.ChooseMin, pick.ChooseMax)
	}
	for _, id := range pick.ChooseCards {
		if id == mine {
			t.Error("a permanent the opponent does not control was offered")
		}
	}
	if err := g.ResolveChooseCards(pick.ID, opp.ID, []uuid.UUID{a, b}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(a) || !opp.Hand.Contains(b) {
		t.Error("the two chosen permanents were not returned to their owner's hand")
	}
	if oldB50OnBattlefield(g, c) || oldB50OnBattlefield(g, mine) {
		t.Error("the unchosen permanent or my own survived the wipe")
	}
}

func TestOldB50DevastatingMasteryAlternativeCostWithNoPermanentsToReturnStillWipes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var opp *game.Player
	for _, s := range g.Seats {
		if s.ID != me.ID {
			opp = s
			break
		}
	}
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if _, err := effectstest.CastWithTapParams(t, g, "Devastating Mastery", "Sorcery", "", oldB50Mastery,
		game.CastSpellParams{AlternativeCost: "devastating-mastery"}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB50OnBattlefield(g, mine) {
		t.Error("my creature survived the wipe")
	}
}

// --- Mage Hunters' Onslaught ---------------------------------------------------

func TestOldB50MageHuntersOnslaughtDestroysThenDrainsBlockersControllers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var opp *game.Player
	for _, s := range g.Seats {
		if s.ID != me.ID {
			opp = s
			break
		}
	}
	effectstest.AdvanceToMain(t, g)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 1, 4)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 1, 1)
	effectstest.CastCatalogSpell(t, g, "Mage Hunters' Onslaught", "Sorcery", oldB50Onslaught, effectstest.CardTargets(victim))
	effectstest.PassPriorityAroundTable(t, g)
	if oldB50OnBattlefield(g, victim) {
		t.Fatal("the target was not destroyed")
	}
	life := effectstest.LifeOf(g, opp.ID)
	myLife := effectstest.LifeOf(g, me.ID)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != life-1 {
		t.Errorf("the blocking creature's controller is at %d, want %d", got, life-1)
	}
	if got := effectstest.LifeOf(g, me.ID); got != myLife {
		t.Errorf("the caster is at %d, want %d", got, myLife)
	}
}

// --- Selfless Glyphweaver // Deadly Vanity ---------------------------------------

func TestOldB50SelflessGlyphweaverExilesItselfToProtectYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	weaver := effectstest.PushCatalogPermanent(g, me.ID, "Selfless Glyphweaver", "Creature — Human Cleric", oldB50Glyphweave, false)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := effectstest.HobB4Activate(t, g, me.ID, weaver, 0, "", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.ZoneOfCard(g, weaver); z != game.ZoneExile {
		t.Errorf("the Glyphweaver is in %v, want exile", z)
	}
	if !effectstest.HasEffectiveKeyword(t, g, mine, "indestructible") {
		t.Error("my creature did not gain indestructible")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "indestructible") {
		t.Error("their creature gained indestructible")
	}
}

func TestOldB50DeadlyVanityKeepsOnlyTheChosenCreatureOrPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	keep := effectstest.B12Creature(g, opp.ID, "Keeper", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, opp.ID, "Other", "Creature — Bear", 2, 2)
	walker := oldB50Walker(g, me.ID)
	relic := effectstest.B12Push(g, me.ID, "Relic", "Artifact", "", 0, 0)
	effectstest.CastCatalogSpell(t, g, "Deadly Vanity", "Sorcery", oldB50Glyphweave+"#1", nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatal("the caster was not asked to choose")
	}
	if pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Errorf("the pick is %d to %d, want exactly one", pick.ChooseMin, pick.ChooseMax)
	}
	for _, id := range pick.ChooseCards {
		if id == relic {
			t.Error("an artifact was offered")
		}
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{keep}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !oldB50OnBattlefield(g, keep) {
		t.Error("the chosen creature was destroyed")
	}
	if oldB50OnBattlefield(g, other) {
		t.Error("another creature survived")
	}
	if oldB50OnBattlefield(g, walker) {
		t.Error("the other planeswalker survived")
	}
	if !oldB50OnBattlefield(g, relic) {
		t.Error("an artifact was destroyed")
	}
}

// --- Necrotic Fumes --------------------------------------------------------------

func TestOldB50NecroticFumesExilesAYourCreatureAsACostAndTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	_, err := effectstest.CastWithTapParams(t, g, "Necrotic Fumes", "Sorcery — Lesson", "", oldB50Fumes,
		game.CastSpellParams{Strict: true, Targets: effectstest.CardTargets(victim), ExilePermanentIDs: []uuid.UUID{fodder}})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.ZoneOfCard(g, fodder); z != game.ZoneExile {
		t.Errorf("the creature paid as a cost is in %v, want exile", z)
	}
	if z := effectstest.ZoneOfCard(g, victim); z != game.ZoneExile {
		t.Errorf("the target is in %v, want exile", z)
	}
}

func TestOldB50NecroticFumesCannotBeCastWithoutACreatureToExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	_, err := effectstest.CastWithTapParams(t, g, "Necrotic Fumes", "Sorcery — Lesson", "", oldB50Fumes,
		game.CastSpellParams{Strict: true, Targets: effectstest.CardTargets(victim)})
	if err == nil {
		t.Error("cast Necrotic Fumes with no creature to exile")
	}
}

// --- Blot Out the Sky ------------------------------------------------------------

func oldB50Inklings(g *game.Game, owner uuid.UUID) []game.Card {
	return effectstest.SocB17Names(g, owner, "Inkling")
}

func TestOldB50BlotOutTheSkyMakesXTappedInklingsAndWipesAtSix(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	relic := effectstest.B12Push(g, opp.ID, "Relic", "Artifact", "", 0, 0)
	walker := oldB50Walker(g, opp.ID)
	vehicle := effectstest.B12Push(g, opp.ID, "Golem", "Artifact Creature — Golem", "", 2, 2)
	land := effectstest.SeedLand(g, opp.ID, "Their Forest", "Basic Land — Forest", "")
	effectstest.CastXSpell(t, g, "Blot Out the Sky", "Sorcery", oldB50BlotOut, "{X}{W}{B}", 5, nil)
	effectstest.PassPriorityAroundTable(t, g)
	tokens := oldB50Inklings(g, me.ID)
	if len(tokens) != 5 {
		t.Fatalf("%d Inklings for X=5, want 5", len(tokens))
	}
	for _, tok := range tokens {
		if !tok.Tapped {
			t.Error("an Inkling entered untapped")
		}
	}
	if !oldB50OnBattlefield(g, relic) || !oldB50OnBattlefield(g, walker) {
		t.Error("X=5 destroyed noncreature permanents")
	}
	// X = 6: the wipe happens, sparing creatures (the Inklings included) and lands.
	effectstest.CastXSpell(t, g, "Blot Out the Sky", "Sorcery", oldB50BlotOut, "{X}{W}{B}", 6, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(oldB50Inklings(g, me.ID)); n != 11 {
		t.Errorf("%d Inklings after X=6, want 11", n)
	}
	if oldB50OnBattlefield(g, relic) || oldB50OnBattlefield(g, walker) {
		t.Error("X=6 left a noncreature, nonland permanent")
	}
	if !oldB50OnBattlefield(g, vehicle) || !oldB50OnBattlefield(g, land) {
		t.Error("X=6 destroyed an artifact creature or a land")
	}
}

// --- Maelstrom Muse --------------------------------------------------------------

func TestOldB50MaelstromMuseDiscountsTheNextInstantOrSorceryByItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	muse := effectstest.B12Push(g, me.ID, "Maelstrom Muse", "Creature — Djinn Wizard", oldB50Muse, 2, 4)
	effectstest.AttackWith(t, g, opp.ID, muse)
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	// A creature spell is not an instant or sorcery.
	bear := effectstest.SocB17Hand(me, "Bear", "Creature — Bear", "{3}{U}", "")
	effectstest.FloatForTest(g, me, "UUU")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("a creature spell was discounted")
	}
	me.ManaPool = nil
	first := effectstest.SocB17Hand(me, "Test Sorcery", "Sorcery", "{3}{U}", "")
	effectstest.FloatForTest(g, me, "UU")
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the next sorcery should cost {X} = {2} less, so {1}{U}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	me.ManaPool = nil
	second := effectstest.SocB17Hand(me, "Test Sorcery", "Sorcery", "{3}{U}", "")
	effectstest.FloatForTest(g, me, "UU")
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("the discount applied to a second spell")
	}
}

// --- Spectacle Mage ----------------------------------------------------------------

func TestOldB50SpectacleMageDiscountsOnlyInstantsAndSorceriesOfManaValueFiveOrMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Spectacle Mage", "Creature — Bird Shaman", oldB50SpecMage, 2, 2)
	cast := func(name, typeLine, cost, pool string) error {
		me.ManaPool = nil
		id := effectstest.SocB17Hand(me, name, typeLine, cost, "")
		effectstest.FloatForTest(g, me, pool)
		return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
	}
	if err := cast("Big Instant", "Instant", "{4}{U}", "UUUU"); err != nil {
		t.Errorf("a mana value 5 instant should cost {3}{U}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := cast("Big Instant", "Instant", "{4}{U}", "UUU"); err == nil {
		t.Error("a mana value 5 instant cost less than {3}{U}")
	}
	if err := cast("Small Instant", "Instant", "{3}{U}", "UUU"); err == nil {
		t.Error("a mana value 4 instant was discounted")
	}
	if err := cast("Big Creature", "Creature — Bear", "{4}{U}", "UUUU"); err == nil {
		t.Error("a mana value 5 creature spell was discounted")
	}
	if err := cast("Big Sorcery", "Sorcery", "{5}{U}", "UUUUU"); err != nil {
		t.Errorf("a mana value 6 sorcery should cost {4}{U}: %v", err)
	}
}

// --- Stonebound Mentor ---------------------------------------------------------------

func TestOldB50StoneboundMentorScriesWhenCardsLeaveYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Stonebound Mentor", "Creature — Spirit Advisor", oldB50Mentor, false)
	theirs := effectstest.SocB11Graveyard(opp, "Their Bear", "Creature — Bear", "{1}{G}")
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: opp.ID}, game.ZoneRef{Kind: game.ZoneExile}, theirs); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKind(g, game.PendingChoiceScry) != nil {
		t.Fatal("a card leaving an opponent's graveyard triggered the scry")
	}
	mine := effectstest.SocB11Graveyard(me, "My Bear", "Creature — Bear", "{1}{G}")
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, mine); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	scry := effectstest.LatestChoiceOfKind(g, game.PendingChoiceScry)
	if scry == nil || len(scry.ScryCards) != 1 {
		t.Fatalf("expected a scry 1 prompt, got %+v", scry)
	}
}

// --- Master Symmetrist ---------------------------------------------------------------

func TestOldB50MasterSymmetristGivesTrampleToAnAttackerWithPowerEqualToToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Master Symmetrist", "Creature — Rhino Druid", oldB50Symmetrist, 4, 4)
	square := effectstest.B12Creature(g, me.ID, "Square", "Creature — Bear", 2, 2)
	lopsided := effectstest.B12Creature(g, me.ID, "Lopsided", "Creature — Bear", 3, 1)
	effectstest.AttackWith(t, g, opp.ID, square, lopsided)
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, square, "trample") {
		t.Error("the 2/2 attacker did not gain trample")
	}
	if effectstest.HasEffectiveKeyword(t, g, lopsided, "trample") {
		t.Error("the 3/1 attacker gained trample")
	}
}
