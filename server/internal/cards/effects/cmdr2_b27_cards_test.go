package effects

import (
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
