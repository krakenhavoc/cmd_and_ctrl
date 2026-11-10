package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_b07_cards_test.go — Commander deck cards wave 2, batch 7
// (Dinosaurs: lords, cost reducers, simple triggers; #5368). Each test
// plays the real catalog entry through the engine.

const (
	cmdr2B07HuntmasterOID  = "41623eab-38b5-4058-8de2-e6b8706de7a2"
	cmdr2B07DroverOID      = "6deba6e8-077b-4602-86af-64db766fbbd9"
	cmdr2B07SpinebackOID   = "c9db31d7-00cf-4200-a11f-74cc2d9e843e"
	cmdr2B07RegisaurOID    = "3d563149-7d4e-43e5-b701-28e3212051ec"
	cmdr2B07HeliopterusOID = "d3b49970-46b7-441b-bff8-8403d982b8a7"
	cmdr2B07SwordtoothOID  = "626709bc-6c29-434f-9e3f-b84bb8a5c431"
	cmdr2B07SeerOID        = "affced7f-797d-4e50-8a24-2da14de44e43"
	cmdr2B07CeratopsOID    = "6cac6f76-f3c6-4b00-a468-166b07853401"
	cmdr2B07ScavengerOID   = "17da3654-7517-473b-89e7-0303f6f030cb"
	cmdr2B07RaptorOID      = "b3fe2f61-c58e-4ebc-a5e3-99d6b6f62caf"
)

func TestCmdr2B07EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		cmdr2B07HuntmasterOID:  "Otepec Huntmaster",
		cmdr2B07DroverOID:      "Drover of the Mighty",
		cmdr2B07SpinebackOID:   "Thundering Spineback",
		cmdr2B07RegisaurOID:    "Raging Regisaur",
		cmdr2B07HeliopterusOID: "Majestic Heliopterus",
		cmdr2B07SwordtoothOID:  "Raging Swordtooth",
		cmdr2B07SeerOID:        "Atzocan Seer",
		cmdr2B07CeratopsOID:    "Shifting Ceratops",
		cmdr2B07ScavengerOID:   "Deathgorge Scavenger",
		cmdr2B07RaptorOID:      "Marauding Raptor",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// cmdr2B07Target is a one-card target list.
func cmdr2B07Target(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

// cmdr2B07Fund puts n generic-payable mana of colour c in me's pool.
func cmdr2B07Fund(me *game.Player, c string, n int) {
	for i := 0; i < n; i++ {
		me.ManaPool.AddMana(game.ManaToken{Color: c})
	}
}

// --- Otepec Huntmaster ---------------------------------------------------

func TestCmdr2B07HuntmasterDiscountsDinosaursOnly(t *testing.T) {
	cases := []struct {
		typeLine string
		lands    int
		ok       bool
	}{
		{"Creature — Dinosaur", 2, true},
		{"Creature — Dinosaur", 1, false},
		{"Creature — Bear", 2, false},
		{"Creature — Bear", 3, true},
	}
	for _, tc := range cases {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		b27Push(g, me.ID, "Otepec Huntmaster", "Creature — Human Shaman", cmdr2B07HuntmasterOID, "{1}{R}", 1, 2, "R")
		_, err := trekB37CastFromHand(t, g, game.Card{Name: "Spell", TypeLine: tc.typeLine, ManaCost: "{2}{R}", Power: 1, Toughness: 1}, tc.lands)
		if (err == nil) != tc.ok {
			t.Errorf("%s with %d Mountains: cast error %v, want ok=%v", tc.typeLine, tc.lands, err, tc.ok)
		}
	}
}

func TestCmdr2B07HuntmasterGivesADinosaurHaste(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hunter := b27Push(g, me.ID, "Otepec Huntmaster", "Creature — Human Shaman", cmdr2B07HuntmasterOID, "{1}{R}", 1, 2, "R")
	dino := b16Creature(g, me.ID, "Raptor", "Creature — Dinosaur", 2, 2, "R")
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, hunter, 0, game.ActivateAbilityParams{Targets: cmdr2B07Target(bear)}); err == nil {
		t.Fatal("a Bear was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, hunter, 0, game.ActivateAbilityParams{Targets: cmdr2B07Target(dino)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, hunter); !c.Tapped {
		t.Error("the Huntmaster did not tap")
	}
	if !hasEffectiveKeyword(t, g, dino, "haste") {
		t.Error("the Dinosaur did not gain haste")
	}
	if hasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("the Bear gained haste")
	}
}

// --- Drover of the Mighty ------------------------------------------------

func TestCmdr2B07DroverGrowsWhileYouControlADinosaur(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drover := b27Push(g, me.ID, "Drover of the Mighty", "Creature — Human Druid", cmdr2B07DroverOID, "{1}{G}", 1, 1, "G")
	if p, tough := effectivePower(t, g, drover), effectiveToughness(t, g, drover); p != 1 || tough != 1 {
		t.Fatalf("alone the Drover is %d/%d, want 1/1", p, tough)
	}
	b16Creature(g, opp.ID, "Their Raptor", "Creature — Dinosaur", 2, 2, "R")
	if p := effectivePower(t, g, drover); p != 1 {
		t.Errorf("an opponent's Dinosaur grew it to %d", p)
	}
	dino := b16Creature(g, me.ID, "My Raptor", "Creature — Dinosaur", 2, 2, "R")
	if p, tough := effectivePower(t, g, drover), effectiveToughness(t, g, drover); p != 3 || tough != 3 {
		t.Errorf("with a Dinosaur the Drover is %d/%d, want 3/3", p, tough)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(dino) })
	if p := effectivePower(t, g, drover); p != 1 {
		t.Errorf("after the Dinosaur left the Drover is %d/x, want 1", p)
	}
}

func TestCmdr2B07DroverTapsForAnyColour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	drover := b27Push(g, me.ID, "Drover of the Mighty", "Creature — Human Druid", cmdr2B07DroverOID, "{1}{G}", 1, 1, "G")
	spec, _ := Lookup(cmdr2B07DroverOID)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{W|U|B|R|G}" || !spec.ManaAbilities[0].Cost.Tap {
		t.Fatalf("mana abilities = %+v", spec.ManaAbilities)
	}
	b23TapMana(t, g, me.ID, drover)
	if c, _ := battlefieldCard(g, drover); !c.Tapped {
		t.Error("the Drover did not tap")
	}
}

// --- Thundering Spineback -------------------------------------------------

func TestCmdr2B07SpinebackPumpsOtherDinosaursYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spine := b27Push(g, me.ID, "Thundering Spineback", "Creature — Dinosaur", cmdr2B07SpinebackOID, "{5}{G}{G}", 5, 5, "G")
	dino := b16Creature(g, me.ID, "My Raptor", "Creature — Dinosaur", 2, 2, "R")
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := b16Creature(g, opp.ID, "Their Raptor", "Creature — Dinosaur", 2, 2, "R")
	for _, c := range []struct {
		id   uuid.UUID
		want int
	}{{spine, 5}, {dino, 3}, {bear, 2}, {theirs, 2}} {
		if p := effectivePower(t, g, c.id); p != c.want {
			t.Errorf("%v has power %d, want %d", c.id, p, c.want)
		}
	}
}

func TestCmdr2B07SpinebackMakesAThreeThreeTramplingDinosaur(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	spine := b27Push(g, me.ID, "Thundering Spineback", "Creature — Dinosaur", cmdr2B07SpinebackOID, "{5}{G}{G}", 5, 5, "G")
	advanceToMain(t, g)
	cmdr2B07Fund(me, "C", 5)
	if err := g.ActivateCatalogAbility(me.ID, spine, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without the {G}")
	}
	cmdr2B07Fund(me, "G", 1)
	if err := g.ActivateCatalogAbility(me.ID, spine, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	tokens := trekB37Tokens(g, me.ID, "Dinosaur")
	if len(tokens) != 1 {
		t.Fatalf("%d Dinosaur tokens, want 1", len(tokens))
	}
	id := tokens[0].InstanceID
	if effectivePower(t, g, id) != 4 || effectiveToughness(t, g, id) != 4 {
		t.Error("a 3/3 token with the Spineback's +1/+1 is 4/4")
	}
	if !hasEffectiveKeyword(t, g, id, "trample") {
		t.Error("the token lacks trample")
	}
}
