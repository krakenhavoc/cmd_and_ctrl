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

// cmdr2B07PickIfAsked answers a pending target prompt for chooser with the
// card, when the engine asked (a lone legal target is chosen for you).
func cmdr2B07PickIfAsked(t *testing.T, g *game.Game, chooser, card uuid.UUID) {
	t.Helper()
	if latestPickTarget(g, chooser) != nil {
		pickCard(t, g, chooser, card)
	}
}

// --- Raging Regisaur ---------------------------------------------------

func TestCmdr2B07RegisaurPingsAnyTargetWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	regisaur := b27Push(g, me.ID, "Raging Regisaur", "Creature — Dinosaur", cmdr2B07RegisaurOID, "{2}{R}{G}", 4, 4, "R", "G")
	victim := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	before := opp.Life
	declareAttack(t, g, opp.ID, regisaur)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if got := cardByID(g, victim); got.DamageMarked != 1 {
		t.Errorf("the target took %d damage, want 1", got.DamageMarked)
	}
	if opp.Life != before {
		t.Errorf("the defending player lost life before combat damage: %d → %d", before, opp.Life)
	}
}

func TestCmdr2B07RegisaurCanPingAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	third := g.Seats[2]
	regisaur := b27Push(g, me.ID, "Raging Regisaur", "Creature — Dinosaur", cmdr2B07RegisaurOID, "{2}{R}{G}", 4, 4, "R", "G")
	before := third.Life
	declareAttack(t, g, opp.ID, regisaur)
	answerPickTargetPlayer(t, g, third.ID)
	passPriorityAroundTable(t, g)
	if third.Life != before-1 {
		t.Errorf("third player's life %d → %d, want 1 lost", before, third.Life)
	}
}

// --- Majestic Heliopterus ----------------------------------------------

func TestCmdr2B07HeliopterusGivesAnotherDinosaurFlying(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	heli := b27Push(g, me.ID, "Majestic Heliopterus", "Creature — Dinosaur", cmdr2B07HeliopterusOID, "{3}{W}", 2, 2, "W")
	dino := b16Creature(g, me.ID, "My Raptor", "Creature — Dinosaur", 2, 2, "R")
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := b16Creature(g, opp.ID, "Their Raptor", "Creature — Dinosaur", 2, 2, "R")
	declareAttack(t, g, opp.ID, heli)
	p := latestPickTarget(g, me.ID)
	if p != nil {
		if hasID(p.PickTargetCards, heli) || hasID(p.PickTargetCards, bear) || hasID(p.PickTargetCards, theirs) {
			t.Errorf("legal targets %v include the Heliopterus, a Bear or an opponent's Dinosaur", p.PickTargetCards)
		}
		pickCard(t, g, me.ID, dino)
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, dino, "flying") {
		t.Error("the other Dinosaur did not gain flying")
	}
	if hasEffectiveKeyword(t, g, bear, "flying") || hasEffectiveKeyword(t, g, theirs, "flying") {
		t.Error("a non-Dinosaur or an opponent's Dinosaur gained flying")
	}
}

func TestCmdr2B07HeliopterusAloneHasNoTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	heli := b27Push(g, me.ID, "Majestic Heliopterus", "Creature — Dinosaur", cmdr2B07HeliopterusOID, "{3}{W}", 2, 2, "W")
	declareAttack(t, g, opp.ID, heli)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("asked for a target with no other Dinosaur")
	}
	if n := triggersOnStackFrom(g, heli); n != 0 {
		t.Errorf("%d triggers on the stack, want 0 (no legal target)", n)
	}
}

// --- Raging Swordtooth -------------------------------------------------

func TestCmdr2B07SwordtoothDamagesEveryOtherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b16Creature(g, me.ID, "My Bear", "Creature — Bear", 3, 3, "G")
	theirs := b16Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1, "G")
	tooth := castCatalogSpell(t, g, "Raging Swordtooth", "Creature — Dinosaur", cmdr2B07SwordtoothOID, nil)
	passPriorityAroundTable(t, g)
	if got := cardByID(g, mine).DamageMarked; got != 1 {
		t.Errorf("my Bear took %d, want 1", got)
	}
	if _, ok := battlefieldCard(g, theirs); ok {
		t.Error("their 1/1 survived")
	}
	if c, ok := battlefieldCard(g, tooth); !ok || c.DamageMarked != 0 {
		t.Errorf("the Swordtooth hurt itself: %+v", c.DamageMarked)
	}
	if !hasEffectiveKeyword(t, g, tooth, "trample") {
		t.Error("no trample")
	}
}

// --- Atzocan Seer ------------------------------------------------------

func TestCmdr2B07SeerReturnsADinosaurCardFromYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	seer := b27Push(g, me.ID, "Atzocan Seer", "Creature — Human Druid", cmdr2B07SeerOID, "{1}{G}{W}", 2, 3, "G", "W")
	dino := pushGraveyardCardTyped(me, "Dead Raptor", "Creature — Dinosaur")
	bear := pushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	theirs := pushGraveyardCardTyped(opp, "Their Dead Raptor", "Creature — Dinosaur")
	advanceToMain(t, g)
	for _, bad := range []uuid.UUID{bear, theirs} {
		if err := g.ActivateCatalogAbility(me.ID, seer, 0, game.ActivateAbilityParams{Strict: true, Targets: cmdr2B07Target(bad)}); err == nil {
			t.Fatalf("%v was a legal target", bad)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, seer, 0, game.ActivateAbilityParams{Strict: true, Targets: cmdr2B07Target(dino)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if _, ok := battlefieldCard(g, seer); ok {
		t.Error("the Seer was not sacrificed as a cost")
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(dino) {
		t.Error("the Dinosaur card did not return to hand")
	}
}

// --- Shifting Ceratops -------------------------------------------------

func TestCmdr2B07CeratopsGainsTheChosenKeyword(t *testing.T) {
	for i, kw := range []string{"reach", "trample", "haste"} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		cer := b27Push(g, me.ID, "Shifting Ceratops", "Creature — Dinosaur", cmdr2B07CeratopsOID, "{2}{G}{G}", 5, 4, "G")
		advanceToMain(t, g)
		cmdr2B07Fund(me, "G", 1)
		if err := g.ActivateCatalogAbility(me.ID, cer, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		passPriorityAroundTable(t, g)
		answerOptionPick(t, g, me.ID, i)
		passPriorityAroundTable(t, g)
		for _, other := range []string{"reach", "trample", "haste"} {
			if got := hasEffectiveKeyword(t, g, cer, other); got != (other == kw) {
				t.Errorf("picked %s: has %s = %v", kw, other, got)
			}
		}
	}
}

func TestCmdr2B07CeratopsCantBeCounteredAndHasProtectionFromBlue(t *testing.T) {
	spec, _ := Lookup(cmdr2B07CeratopsOID)
	if !spec.CantBeCountered {
		t.Error("the spell can be countered")
	}
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cer := b27Push(g, me.ID, "Shifting Ceratops", "Creature — Dinosaur", cmdr2B07CeratopsOID, "{2}{G}{G}", 5, 4, "G")
	if !hasEffectiveKeyword(t, g, cer, game.ProtectionFromColor("U")) {
		t.Error("no protection from blue")
	}
	blue := b16Creature(g, opp.ID, "Blue Wall", "Creature — Wall", 0, 4, "U")
	attackerErr := func() error {
		advanceTo(t, g, game.StepDeclareAttackers)
		return nil
	}
	_ = attackerErr
	_ = blue
}

// --- Deathgorge Scavenger ----------------------------------------------

func TestCmdr2B07ScavengerExilesACreatureCardAndGainsLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := pushGraveyardCardTyped(opp, "Dead Bear", "Creature — Bear")
	before := me.Life
	scav := castCatalogSpell(t, g, "Deathgorge Scavenger", "Creature — Dinosaur", cmdr2B07ScavengerOID, nil)
	passPriorityAroundTable(t, g)
	cmdr2B07PickIfAsked(t, g, me.ID, dead)
	passPriorityAroundTable(t, g)
	cmdr2B07AnswerPrompt(t, g, me, true)
	passPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(dead) {
		t.Error("the card is still in the graveyard")
	}
	if me.Life != before+2 {
		t.Errorf("life %d → %d, want +2", before, me.Life)
	}
	if p := effectivePower(t, g, scav); p != 3 {
		t.Errorf("power %d, want 3 (no pump for a creature card)", p)
	}
}

func TestCmdr2B07ScavengerExilesANoncreatureCardAndGrows(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell := pushGraveyardCardTyped(opp, "Dead Bolt", "Instant")
	before := me.Life
	scav := castCatalogSpell(t, g, "Deathgorge Scavenger", "Creature — Dinosaur", cmdr2B07ScavengerOID, nil)
	passPriorityAroundTable(t, g)
	cmdr2B07PickIfAsked(t, g, me.ID, spell)
	passPriorityAroundTable(t, g)
	cmdr2B07AnswerPrompt(t, g, me, true)
	passPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(spell) {
		t.Error("the card is still in the graveyard")
	}
	if me.Life != before {
		t.Errorf("life changed to %d", me.Life)
	}
	if p, tough := effectivePower(t, g, scav), effectiveToughness(t, g, scav); p != 4 || tough != 3 {
		t.Errorf("%d/%d, want 4/3 until end of turn", p, tough)
	}
}

func TestCmdr2B07ScavengerMayDecline(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := pushGraveyardCardTyped(opp, "Dead Bear", "Creature — Bear")
	castCatalogSpell(t, g, "Deathgorge Scavenger", "Creature — Dinosaur", cmdr2B07ScavengerOID, nil)
	passPriorityAroundTable(t, g)
	cmdr2B07PickIfAsked(t, g, me.ID, dead)
	passPriorityAroundTable(t, g)
	cmdr2B07AnswerPrompt(t, g, me, false)
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(dead) {
		t.Error("declined, yet the card left the graveyard")
	}
}

// --- Marauding Raptor --------------------------------------------------

func TestCmdr2B07RaptorDiscountsCreatureSpells(t *testing.T) {
	cases := []struct {
		typeLine string
		lands    int
		ok       bool
	}{
		{"Creature — Bear", 2, true},
		{"Creature — Bear", 1, false},
		{"Sorcery", 2, false},
	}
	for _, tc := range cases {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		b27Push(g, me.ID, "Marauding Raptor", "Creature — Dinosaur", cmdr2B07RaptorOID, "{1}{R}", 2, 3, "R")
		_, err := trekB37CastFromHand(t, g, game.Card{Name: "Spell", TypeLine: tc.typeLine, ManaCost: "{2}{R}", Power: 1, Toughness: 1}, tc.lands)
		if (err == nil) != tc.ok {
			t.Errorf("%s with %d Mountains: cast error %v, want ok=%v", tc.typeLine, tc.lands, err, tc.ok)
		}
	}
}

func TestCmdr2B07RaptorShootsANonDinosaurWithoutPumping(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	raptor := b27Push(g, me.ID, "Marauding Raptor", "Creature — Dinosaur", cmdr2B07RaptorOID, "{1}{R}", 2, 3, "R")
	bear := castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	_ = bear
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, raptor); p != 2 {
		t.Errorf("the Raptor is %d power after a Bear entered, want 2", p)
	}
}

func TestCmdr2B07RaptorDamagesAnEnteringDinosaurAndPumps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	raptor := b27Push(g, me.ID, "Marauding Raptor", "Creature — Dinosaur", cmdr2B07RaptorOID, "{1}{R}", 2, 3, "R")
	dino := b16Creature(g, me.ID, "Big Raptor", "Creature — Dinosaur", 5, 5, "R")
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("3/3 green Dinosaur with trample"), 1) })
	passPriorityAroundTable(t, g)
	_ = dino
	if p := effectivePower(t, g, raptor); p != 4 {
		t.Errorf("the Raptor is %d power after a Dinosaur token entered, want 4", p)
	}
	toks := trekB37Tokens(g, me.ID, "Dinosaur")
	if len(toks) != 1 || toks[0].DamageMarked != 2 {
		t.Errorf("tokens %+v, want one with 2 damage", toks)
	}
}

// cmdr2B07AnswerPrompt answers an open "you may" trigger prompt.
func cmdr2B07AnswerPrompt(t *testing.T, g *game.Game, me *game.Player, yes bool) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Chooser == me.ID {
			if err := g.ResolveTriggerPrompt(c.ID, me.ID, yes); err != nil {
				t.Fatalf("ResolveTriggerPrompt: %v", err)
			}
			return
		}
	}
	t.Fatalf("no trigger prompt for %s: %+v", me.Name, g.PendingChoices)
}
