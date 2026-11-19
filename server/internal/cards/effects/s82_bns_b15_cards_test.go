package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	bnsB15Roofstalker = "c2f2fd39-4a21-489b-893e-4258f8d2aa63"
	bnsB15Savant      = "6eaf5cdc-2843-4167-bd18-df4c2ac73ac0"
	bnsB15Sauroform   = "559c9949-faaf-41bc-9349-b6738b4763a2"
	bnsB15Scuttlegtr  = "9bc0ca04-26ab-4743-bc65-0b5f4115c8f7"
	bnsB15Sentinel    = "a9f1802b-f2a0-4902-acc2-e21c14960e4c"
	bnsB15Selesnya    = "8fbdca3a-7aa3-4b5a-98d7-9168856045ac"
	bnsB15Seller      = "87e4c6ff-f83f-412b-9d23-aac7a57ef6db"
	bnsB15Seraph      = "4a3acbf7-0699-4d7c-b4fa-aa15c7cde189"
	bnsB15Eel         = "75132177-0881-4158-9d65-73f5ba9d4402"
	bnsB15Titan       = "94dd397e-0127-4c79-b9d9-22ce62ccdfe7"
	bnsB15Spawn       = "6a290adf-7b75-451e-bb79-9c13e10b38aa"
	bnsB15Weird       = "550811ca-0995-4886-aa8c-5cee118aa54f"
)

func TestBnsB15Registered(t *testing.T) {
	for _, id := range []string{bnsB15Roofstalker, bnsB15Savant, bnsB15Sauroform, bnsB15Scuttlegtr, bnsB15Sentinel,
		bnsB15Selesnya, bnsB15Seller, bnsB15Seraph, bnsB15Eel, bnsB15Titan, bnsB15Spawn, bnsB15Weird} {
		spec, ok := Lookup(id)
		if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v completeness=%v caveats=%v", id, ok, spec.Completeness, spec.Caveats)
		}
	}
}

func TestBnsB15RoofstalkerWightGainsFlyingForOneBlue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	w := b12Push(g, me.ID, "Roofstalker Wight", "Creature — Zombie", bnsB15Roofstalker, 2, 1)
	advanceToMain(t, g)
	if hasEffectiveKeyword(t, g, w, "flying") {
		t.Fatal("flying before the ability")
	}
	floatMana(t, g, me, "{C}{U}")
	b16Activate(t, g, me.ID, w, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, w, "flying") {
		t.Error("no flying after the ability resolved")
	}
}

func TestBnsB15SagesRowSavantScrysTwoOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	advanceToMain(t, g)
	mm2Library(me, mm2Spell("A"), mm2Spell("B"), mm2Spell("C"))
	ecl6Cast(t, g, me, "Sage's Row Savant", "Creature — Vedalken Wizard", bnsB15Savant, 2, 1, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if len(g.PendingChoices) == 0 {
		t.Fatal("no scry prompt opened")
	}
}

func TestBnsB15SauroformHybridAdaptsOnce(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	h := b12Push(g, me.ID, "Sauroform Hybrid", "Creature — Human Lizard Warrior", bnsB15Sauroform, 2, 2)
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}{C}{C}{G}{G}")
	b16Activate(t, g, me.ID, h, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := counterCount(g, h, game.CounterPlusOne); got != 4 {
		t.Fatalf("counters = %d, want 4", got)
	}
	floatMana(t, g, me, "{C}{C}{C}{C}{G}{G}")
	b16Activate(t, g, me.ID, h, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := counterCount(g, h, game.CounterPlusOne); got != 4 {
		t.Errorf("a second adapt added counters: %d", got)
	}
}

func TestBnsB15SkitterEelAdaptsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	e := b12Push(g, me.ID, "Skitter Eel", "Creature — Fish Crab", bnsB15Eel, 3, 3)
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}{U}")
	b16Activate(t, g, me.ID, e, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := counterCount(g, e, game.CounterPlusOne); got != 2 {
		t.Errorf("counters = %d, want 2", got)
	}
}

func TestBnsB15ScuttlegatorAttacksOnlyOnceItHasAPlusOneCounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	s := b12Push(g, me.ID, "Scuttlegator", "Creature — Crab Turtle Crocodile", bnsB15Scuttlegtr, 6, 6)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(s, opp.ID); err == nil {
		t.Fatal("a Scuttlegator with no counter attacked")
	}

	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	s2 := b12Push(g2, me2.ID, "Scuttlegator", "Creature — Crab Turtle Crocodile", bnsB15Scuttlegtr, 6, 6)
	advanceToMain(t, g2)
	floatMana(t, g2, me2, "{C}{C}{C}{C}{C}{C}{G}{G}")
	b16Activate(t, g2, me2.ID, s2, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g2)
	if got := counterCount(g2, s2, game.CounterPlusOne); got != 3 {
		t.Fatalf("counters = %d, want 3", got)
	}
	advanceTo(t, g2, game.StepDeclareAttackers)
	if err := g2.DeclareAttacker(s2, opp2.ID); err != nil {
		t.Fatalf("a Scuttlegator with counters could not attack: %v", err)
	}
	if !hasEffectiveKeyword(t, g2, s2, "defender") {
		t.Error("defender was removed rather than waived")
	}
}

func TestBnsB15ScuttlingSentinelBoostsAnotherCreatureAndGivesItHexproof(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "Other Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	sent := ecl6Cast(t, g, me, "Scuttling Sentinel", "Creature — Crab Elf", bnsB15Sentinel, 3, 2, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		for _, bad := range []uuid.UUID{sent, theirs} {
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
				t.Errorf("an illegal target %s was accepted", bad)
			}
		}
		pickCard(t, g, me.ID, bear)
	} else {
		t.Fatal("no target pick")
	}
	passPriorityAroundTable(t, g)
	if counterCount(g, other, game.CounterPlusOne) != 0 {
		t.Error("the counter landed on the wrong creature")
	}
	if got := counterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Fatalf("bear counters = %d, want 1", got)
	}
	if got := counterCount(g, sent, game.CounterPlusOne); got != 0 {
		t.Errorf("the Sentinel put a counter on itself: %d", got)
	}
	if !hasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("the bear has no hexproof")
	}
	v := s58p6View(t, g, bear)
	if !v.HasSubtype("Crab") || !v.HasColor("U") || !v.HasSubtype("Bear") {
		t.Errorf("bear is not a blue Crab in addition: %+v", v.TypeLine)
	}
}

func TestBnsB15SelesnyaGuildmageMakesSaprolingsAndPumps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	gm := b12Push(g, me.ID, "Selesnya Guildmage", "Creature — Elf Wizard", bnsB15Selesnya, 2, 2)
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}{C}{G}")
	b16Activate(t, g, me.ID, gm, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Saproling"); n != 1 {
		t.Fatalf("%d Saprolings, want 1", n)
	}
	floatMana(t, g, me, "{C}{C}{C}{W}")
	b16Activate(t, g, me.ID, gm, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if v := s58p6View(t, g, gm); v.CurrentPower() != 3 || v.CurrentToughness() != 3 {
		t.Errorf("guildmage is %d/%d, want 3/3", v.CurrentPower(), v.CurrentToughness())
	}
}

func TestBnsB15SellerOfSongbirdsMakesABird(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	advanceToMain(t, g)
	ecl6Cast(t, g, me, "Seller of Songbirds", "Creature — Human", bnsB15Seller, 1, 2, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Bird"); n != 1 {
		t.Errorf("%d Birds, want 1", n)
	}
}

func TestBnsB15SeraphOfTheScalesGrantsAndLeavesTwoSpirits(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	s := b12Push(g, me.ID, "Seraph of the Scales", "Creature — Angel", bnsB15Seraph, 4, 3)
	advanceToMain(t, g)
	floatMana(t, g, me, "{W}")
	b16Activate(t, g, me.ID, s, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, s, "vigilance") {
		t.Error("no vigilance")
	}
	floatMana(t, g, me, "{B}")
	b16Activate(t, g, me.ID, s, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, s, "deathtouch") {
		t.Error("no deathtouch")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(s) })
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Spirit"); n != 2 {
		t.Errorf("%d Spirits, want 2", n)
	}
}

func TestBnsB15SpellgorgerWeirdGrowsOnNoncreatureSpellsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	w := b12Push(g, me.ID, "Spellgorger Weird", "Creature — Weird", bnsB15Weird, 2, 2)
	advanceToMain(t, g)
	castFromHandForTest(t, g, me, "Plain Sorcery", "Sorcery", "", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if got := counterCount(g, w, game.CounterPlusOne); got != 1 {
		t.Fatalf("counters after a sorcery = %d, want 1", got)
	}
	castFromHandForTest(t, g, me, "Bear", "Creature — Bear", "", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if got := counterCount(g, w, game.CounterPlusOne); got != 1 {
		t.Errorf("a creature spell added a counter: %d", got)
	}
}

func TestBnsB15SpawnOfMayhemUpkeepPingsEveryoneAndGrowsAtTenOrLess(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spawn := b12Push(g, me.ID, "Spawn of Mayhem", "Creature — Demon", bnsB15Spawn, 4, 4)
	g.WithWriteLock(func() { me.Life = 11 })
	advanceToUpkeepOf(t, g, 1)
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if me.Life != 10 || opp.Life != 39 {
		t.Fatalf("life %d / %d, want 10 / 39", me.Life, opp.Life)
	}
	if got := counterCount(g, spawn, game.CounterPlusOne); got != 1 {
		t.Errorf("counters at 10 life = %d, want 1", got)
	}
}

func TestBnsB15SpawnOfMayhemNoCounterAboveTen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spawn := b12Push(g, me.ID, "Spawn of Mayhem", "Creature — Demon", bnsB15Spawn, 4, 4)
	advanceToUpkeepOf(t, g, 1)
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, spawn, game.CounterPlusOne); got != 0 {
		t.Errorf("counters at high life = %d, want 0", got)
	}
}

func TestBnsB15SpawnOfMayhemSpectacleNeedsAnOpponentToHaveLostLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	card := game.Card{Name: "Spawn of Mayhem", TypeLine: "Creature — Demon", ManaCost: "{2}{B}{B}", OracleID: bnsB15Spawn, Power: 4, Toughness: 4}
	if _, err := sncB14Cast(t, g, me, card, "CBB", game.CastSpellParams{Strict: true, AlternativeCost: "spectacle"}); err == nil {
		t.Fatal("spectacle accepted with no life lost this turn")
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -1) })
	if _, err := sncB14Cast(t, g, me, card, "CBB", game.CastSpellParams{Strict: true, AlternativeCost: "spectacle"}); err != nil {
		t.Fatalf("spectacle after an opponent lost life: %v", err)
	}
}

func TestBnsB15SludgeTitanMillsFiveAndTakesACreatureAndALand(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	advanceToMain(t, g)
	ids := mm2Library(me, mm2CreatureCard("Bear", 2), mm2Land("Forest"), mm2Spell("A"), mm2CreatureCard("Ogre", 3), mm2Spell("B"), mm2Spell("Deep"))
	ecl6Cast(t, g, me, "Sludge Titan", "Creature — Zombie Giant", bnsB15Titan, 6, 6, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	pc := chooseCardsChoiceFor(g, me.ID)
	if pc == nil || len(pc.ChooseCards) != 2 {
		t.Fatalf("creature prompt = %+v, want the two creature cards", pc)
	}
	answerChooseCards(t, g, me.ID, ids[0])
	pc = chooseCardsChoiceFor(g, me.ID)
	if pc == nil || len(pc.ChooseCards) != 1 || pc.ChooseCards[0] != ids[1] {
		t.Fatalf("land prompt = %+v, want only the Forest", pc)
	}
	answerChooseCards(t, g, me.ID, ids[1])
	if mm2Zone(g, ids[0]) != game.ZoneHand || mm2Zone(g, ids[1]) != game.ZoneHand {
		t.Error("the picked creature and land are not in hand")
	}
	if mm2Zone(g, ids[3]) != game.ZoneGraveyard {
		t.Error("the unpicked creature left the graveyard")
	}
	if mm2Zone(g, ids[5]) != game.ZoneLibrary {
		t.Error("the sixth card was milled")
	}
}

func TestBnsB15SludgeTitanMayDeclineBothPicks(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	advanceToMain(t, g)
	ids := mm2Library(me, mm2CreatureCard("Bear", 2), mm2Land("Forest"), mm2Spell("A"), mm2Spell("B"), mm2Spell("C"))
	ecl6Cast(t, g, me, "Sludge Titan", "Creature — Zombie Giant", bnsB15Titan, 6, 6, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID)
	answerChooseCards(t, g, me.ID)
	for _, id := range ids {
		if mm2Zone(g, id) != game.ZoneGraveyard {
			t.Errorf("card %v is not in the graveyard", id)
		}
	}
}
