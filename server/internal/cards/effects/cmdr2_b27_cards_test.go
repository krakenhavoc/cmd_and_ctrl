package effects

import (
	"github.com/google/uuid"
	"reflect"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	cmdr2B27StarfishOracle  = "58321467-41d6-4419-bcb2-cd3e60f677b3"
	cmdr2B27NotionOracle    = "f93dbe17-71d1-4542-96ff-81ac6da9513f"
	cmdr2B27CurateOracle    = "75e8907d-f726-4a63-ba8b-c0fd4936734e"
	cmdr2B27SpriteOracle    = "8ac23863-0aa7-41bc-b9b7-c5f0566172cf"
	cmdr2B27DispersalOracle = "1e0bb57a-f5c7-45e9-b5b9-5bb3ec930108"
)

func TestCmdr2B27SinisterStarfishTapsToSurveil(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	seedLibrary(me, "Bin Me", "Keep")
	fish := b12Push(g, me.ID, "Sinister Starfish", "Creature — Starfish", cmdr2B27StarfishOracle, 0, 3)
	if err := g.ActivateCatalogAbility(me.ID, fish, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !b16Tapped(t, g, fish) {
		t.Error("the Starfish did not tap")
	}
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("want a surveil 1 prompt, got %+v", c)
	}
	gy := me.Graveyard.Size()
	if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if me.Graveyard.Size() != gy+1 || me.Graveyard.Cards[me.Graveyard.Size()-1].Name != "Bin Me" {
		t.Errorf("graveyard %v, want Bin Me on top", graveyardNames(me))
	}
	// Tapped: it cannot activate again.
	if err := g.ActivateCatalogAbility(me.ID, fish, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated while tapped")
	}
}

func TestCmdr2B27NightveilSpriteSurveilsWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	seedLibrary(me, "Top", "Next")
	sprite := b12Push(g, me.ID, "Nightveil Sprite", "Creature — Faerie Rogue", cmdr2B27SpriteOracle, 1, 2)
	declareAttack(t, g, opp.ID, sprite)
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("want a surveil 1 prompt after attacking, got %+v", c)
	}
}

func TestCmdr2B27NightveilSpriteDoesNotSurveilUnlessItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Top", "Next")
	b12Push(g, me.ID, "Nightveil Sprite", "Creature — Faerie Rogue", cmdr2B27SpriteOracle, 1, 2)
	advanceToMain(t, g)
	passPriorityAroundTable(t, g)
	if c := surveilChoiceFor(g, me.ID); c != nil {
		t.Fatalf("surveil without an attack: %+v", c)
	}
}

func TestCmdr2B27CurateSurveilsTwoThenDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Bin One", "Bin Two", "Draw Me")
	hand, gy := me.Hand.Size(), me.Graveyard.Size()
	castCatalogSpell(t, g, "Curate", "Instant", cmdr2B27CurateOracle, nil)
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("want a surveil 2 prompt, got %+v", c)
	}
	if me.Hand.Size() != hand { // Curate left the hand, nothing drawn yet
		t.Fatalf("hand %d before answering, want %d", me.Hand.Size(), hand)
	}
	if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Graveyard.Size() != gy+3 {
		t.Errorf("graveyard %v, want two binned plus Curate", graveyardNames(me))
	}
	if got := me.Hand.Cards[me.Hand.Size()-1].Name; got != "Draw Me" {
		t.Errorf("drew %q, want Draw Me", got)
	}
}

func TestCmdr2B27NotionRainSurveilsDrawsAndHurtsYou(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Bin One", "Keep", "Draw A", "Draw B")
	hand, life := me.Hand.Size(), me.Life
	castCatalogSpell(t, g, "Notion Rain", "Sorcery", cmdr2B27NotionOracle, nil)
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("want a surveil 2 prompt, got %+v", c)
	}
	if me.Life != life {
		t.Fatalf("damage before the surveil was answered")
	}
	if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards[:1], c.ScryCards[1:]); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+2)
	}
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
}

func TestCmdr2B27EpharasDispersalBouncesThenSurveils(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ox := b12Creature(g, opp.ID, "Ox", "Creature — Ox", 2, 2)
	seedLibrary(me, "A", "B")
	advanceToMain(t, g)
	card := game.Card{Name: "Ephara's Dispersal", TypeLine: "Instant", ManaCost: "{2}{U}", OracleID: cmdr2B27DispersalOracle}
	if err := socB19CastWithPool(g, me, card, "U", 3, game.CastSpellParams{Targets: cardRefs(ox)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ox) {
		t.Error("the creature was not bounced")
	}
	c := surveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("want a surveil 2 prompt, got %+v", c)
	}
}

func TestCmdr2B27EpharasDispersalCostsTwoLessAgainstAnAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	// The opponent's creature attacks me: seat 1 needs the turn.
	_ = me
	_ = opp
	g = newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	victim := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	atk := b12Push(g, active.ID, "Raider", "Creature — Human", "", 2, 2)
	idle := b12Creature(g, active.ID, "Idle", "Creature — Human", 2, 2)
	declareAttack(t, g, victim.ID, atk)
	passPriorityAroundTable(t, g)
	card := game.Card{Name: "Ephara's Dispersal", TypeLine: "Instant", ManaCost: "{2}{U}", OracleID: cmdr2B27DispersalOracle}
	if err := socB19CastWithPool(g, victim, card, "U", 1, game.CastSpellParams{Targets: cardRefs(atk)}); err != nil {
		t.Fatalf("cast for {U} at an attacker: %v", err)
	}
	card2 := game.Card{Name: "Ephara's Dispersal", TypeLine: "Instant", ManaCost: "{2}{U}", OracleID: cmdr2B27DispersalOracle}
	if err := socB19CastWithPool(g, victim, card2, "U", 1, game.CastSpellParams{Targets: cardRefs(idle)}); err == nil {
		t.Fatal("cast for {U} at a non-attacking creature")
	}
}

const cmdr2B27MasterOracle = "3a16bdda-8499-47e4-b870-b9576f4c1a83"

func cmdr2B27MasterInGraveyard(g *game.Game, owner *game.Player) uuid.UUID {
	id := uuid.New()
	owner.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: "Master of Death", TypeLine: "Creature — Zombie Wizard", ManaCost: "{1}{U}{B}",
		OracleID: cmdr2B27MasterOracle, Power: 3, Toughness: 1, Owner: owner.ID, Controller: owner.ID,
	})
	return id
}

func TestCmdr2B27MasterOfDeathSurveilsTwoWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B", "C")
	castCatalogSpell(t, g, "Master of Death", "Creature — Zombie Wizard", cmdr2B27MasterOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("want a surveil 2 prompt, got %+v", c)
	}
}

func TestCmdr2B27MasterOfDeathReturnsFromTheGraveyardForALife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := cmdr2B27MasterInGraveyard(g, me)
	life := me.Life
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Life != life-1 {
		t.Errorf("life %d → %d, want 1 paid", life, me.Life)
	}
	if !me.Hand.Contains(id) || me.Graveyard.Contains(id) {
		t.Error("Master of Death did not return to hand")
	}
}

func TestCmdr2B27MasterOfDeathStaysIfYouDeclineToPay(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := cmdr2B27MasterInGraveyard(g, me)
	life := me.Life
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if me.Life != life || !me.Graveyard.Contains(id) || me.Hand.Contains(id) {
		t.Error("declining should leave the card in the graveyard and the life untouched")
	}
}

func TestCmdr2B27MasterOfDeathDoesNothingFromTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Master of Death", "Creature — Zombie Wizard", cmdr2B27MasterOracle, 3, 1)
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		t.Error("the graveyard trigger fired from the battlefield")
	}
}

const (
	cmdr2B27TellingTimeOracle = "192f3c8a-1120-4f7f-ae0f-fe9bde59cd08"
	cmdr2B27VisionOracle      = "6b7b1a71-812d-4ff5-b38b-94d1cf634143"
)

func TestCmdr2B27TellingTimeSplitsThreeCardsHandTopBottom(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedLibrary(me, "One", "Two", "Three", "Four")
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Telling Time", "Instant", cmdr2B27TellingTimeOracle, nil)
	passPriorityAroundTable(t, g)
	pick := latestChooseCards(g, me.ID)
	if pick == nil || len(pick.ChooseCards) != 3 || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Fatalf("want a mandatory pick of one of three, got %+v", pick)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(ids[1]) || me.Hand.Size() != hand+1 {
		t.Fatal("Two did not reach the hand")
	}
	order := putInLibraryChoiceFor(g, me.ID)
	if order == nil {
		t.Fatal("no top/bottom prompt for the other two")
	}
	// Three goes on top, One on the bottom.
	if err := g.ResolvePutInLibrary(order.ID, me.ID, []uuid.UUID{ids[0]}, []uuid.UUID{ids[2]}); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if got := me.Library.Cards[me.Library.Size()-1].InstanceID; got != ids[2] {
		t.Errorf("top of library is %v, want Three", got)
	}
	if got := me.Library.Cards[0].InstanceID; got != ids[0] {
		t.Errorf("bottom of library is %v, want One", got)
	}
	// Both on top is refused: exactly one goes on top.
	if err := g.ResolvePutInLibrary(order.ID, me.ID, nil, []uuid.UUID{ids[0], ids[2]}); err == nil {
		t.Error("two cards on top were accepted")
	}
}

func TestCmdr2B27DiabolicVisionTakesOneOfFiveAndReordersTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedLibrary(me, "A", "B", "C", "D", "E", "F")
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Diabolic Vision", "Sorcery", cmdr2B27VisionOracle, nil)
	passPriorityAroundTable(t, g)
	pick := latestChooseCards(g, me.ID)
	if pick == nil || len(pick.ChooseCards) != 5 || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Fatalf("want a mandatory pick of one of five, got %+v", pick)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[3]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(ids[3]) || me.Hand.Size() != hand+1 {
		t.Fatal("D did not reach the hand")
	}
	order := putInLibraryChoiceFor(g, me.ID)
	if order == nil {
		t.Fatal("no order prompt for the rest")
	}
	// Nothing may go to the bottom.
	if err := g.ResolvePutInLibrary(order.ID, me.ID, []uuid.UUID{ids[0]}, []uuid.UUID{ids[1], ids[2], ids[4]}); err == nil {
		t.Error("a card was put on the bottom")
	}
	// Top-first order: E, C, B, A.
	if err := g.ResolvePutInLibrary(order.ID, me.ID, nil, []uuid.UUID{ids[4], ids[2], ids[1], ids[0]}); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	want := []string{"E", "C", "B", "A", "F"}
	if got := libraryTopNames(me, 5); !reflect.DeepEqual(got, want) {
		t.Errorf("library top %v, want %v", got, want)
	}
}

const (
	cmdr2B27MeaningOracle = "70e58068-90d2-4720-a11c-5d243046b4a3"
	cmdr2B27WatchOracle   = "847b5be5-cb0c-4857-bcd9-c5cfe55fdaeb"
)

func TestCmdr2B27ThirstForMeaningDrawsThreeAndOwesTwoDiscards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Thirst for Meaning", "Instant", cmdr2B27MeaningOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+3 {
		t.Errorf("drew %d, want 3", me.Hand.Size()-hand)
	}
	if discardOwed(g, me.ID) != 2 {
		t.Errorf("owes %d discards, want 2", discardOwed(g, me.ID))
	}
}

func TestCmdr2B27ThirstForMeaningAcceptsOneEnchantmentOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	aura := handCardFull(me, "Pacifism", "Enchantment — Aura", "", "", nil)
	rock := handCardFull(me, "Sol Ring", "Artifact", "", "", nil)
	castCatalogSpell(t, g, "Thirst for Meaning", "Instant", cmdr2B27MeaningOracle, nil)
	passPriorityAroundTable(t, g)
	c := discardChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no discard prompt")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{rock}); err == nil {
		t.Fatal("a lone artifact was accepted")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{aura}); err != nil {
		t.Fatalf("one enchantment card is a complete discard: %v", err)
	}
	if me.Hand.Contains(aura) || !me.Hand.Contains(rock) {
		t.Error("only the enchantment was discarded")
	}
}

func TestCmdr2B27ThirstForMeaningAcceptsAnyTwoCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := handCardFull(me, "Sol Ring", "Artifact", "", "", nil)
	b := handCardFull(me, "Forest", "Basic Land — Forest", "", "", nil)
	castCatalogSpell(t, g, "Thirst for Meaning", "Instant", cmdr2B27MeaningOracle, nil)
	passPriorityAroundTable(t, g)
	c := discardChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no discard prompt")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{a, b}); err != nil {
		t.Fatalf("two cards of any kind: %v", err)
	}
	if me.Hand.Contains(a) || me.Hand.Contains(b) {
		t.Error("both cards should be discarded")
	}
}

func TestCmdr2B27KeepWatchDrawsForEachAttackingCreature(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	victim := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := b12Push(g, active.ID, "Raider A", "Creature — Human", "", 2, 2)
	b := b12Push(g, active.ID, "Raider B", "Creature — Human", "", 2, 2)
	b12Creature(g, active.ID, "Idle", "Creature — Human", 2, 2)
	declareAttack(t, g, victim.ID, a, b)
	passPriorityAroundTable(t, g)
	hand := victim.Hand.Size()
	card := game.Card{Name: "Keep Watch", TypeLine: "Instant", ManaCost: "{2}{U}", OracleID: cmdr2B27WatchOracle}
	if err := socB19CastWithPool(g, victim, card, "U", 3, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := victim.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2 (one per attacker, not the idle creature)", got)
	}
}

func TestCmdr2B27KeepWatchDrawsNothingOutsideCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Creature(g, me.ID, "Idle", "Creature — Human", 2, 2)
	advanceToMain(t, g)
	hand := me.Hand.Size()
	card := game.Card{Name: "Keep Watch", TypeLine: "Instant", ManaCost: "{2}{U}", OracleID: cmdr2B27WatchOracle}
	if err := socB19CastWithPool(g, me, card, "U", 3, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("drew %d with no attackers", me.Hand.Size()-hand)
	}
}
