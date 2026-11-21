package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b141_test.go — S82 GAP batch 141 (Dominaria Remastered), #5948:
// the cards through the real catalog.

const (
	gapImprovisedArmor = "aca7c0a7-b365-421e-aeb0-49d3a9873e4f"
	gapSunClasp        = "cb8f5ee4-285a-4c37-adcb-514b3e4bdad3"
	gapTestEndurance   = "ffdad8fb-69ee-40d3-b2eb-b0aee0a5becc"
	gapHowl            = "403cf6ae-48a9-4ea9-894e-7135cfca4e1b"
	gapRecoil          = "1579bf86-4400-4467-88e8-1c5ba1209804"
	gapSpinalEmbrace   = "4cf3fb65-9107-428a-8853-029ec97112b5"
	gapEmeraldCharm    = "904b3db5-70bc-4d49-9887-f504aff8666f"
	gapLull            = "d53f4c24-8deb-485a-b03a-945211fdf10d"
	gapPrimalBoost     = "902e7d4b-7ab4-40b5-b3a5-c83ef604989a"
)

func gapB141PT(g *game.Game, id uuid.UUID) (p, tough int) {
	g.ReadSnapshot(func() {
		if c, ok := g.LookupCardForEffect(id); ok {
			p, tough = c.PowerForComparison(), c.CurrentToughness()
		}
	})
	return
}

func TestGapB141Registered(t *testing.T) {
	for oracle, name := range map[string]string{
		gapImprovisedArmor: "Improvised Armor", gapSunClasp: "Sun Clasp",
		gapTestEndurance: "Test of Endurance", gapHowl: "Howl from Beyond",
		gapRecoil: "Recoil", gapSpinalEmbrace: "Spinal Embrace",
		gapEmeraldCharm: "Emerald Charm", gapLull: "Lull",
		gapPrimalBoost: "Primal Boost",
	} {
		stdRegisteredFull(t, oracle, name)
	}
}

func TestGapB141AurasPumpAndSunClaspBounces(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	castCatalogSpell(t, g, "Improvised Armor", "Enchantment — Aura", gapImprovisedArmor, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if p, th := gapB141PT(g, bear); p != 4 || th != 7 {
		t.Fatalf("Improvised Armor: got %d/%d, want 4/7", p, th)
	}
	clasp := castCatalogSpell(t, g, "Sun Clasp", "Enchantment — Aura", gapSunClasp, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if p, th := gapB141PT(g, bear); p != 5 || th != 10 {
		t.Fatalf("with Sun Clasp: got %d/%d, want 5/10", p, th)
	}
	b06AddMana(me, "W")
	if err := g.ActivateCatalogAbility(me.ID, clasp, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate Sun Clasp: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the enchanted creature stayed on the battlefield")
	}
	if !me.Hand.Contains(bear) {
		t.Error("the enchanted creature did not return to its owner's hand")
	}
}

func TestGapB141SunClaspRefusesWithoutMana(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	clasp := castCatalogSpell(t, g, "Sun Clasp", "Enchantment — Aura", gapSunClasp, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, clasp, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("Sun Clasp activated with no {W}")
	}
}

func TestGapB141TestOfEnduranceWinsAtFiftyLife(t *testing.T) {
	for _, tc := range []struct {
		life int
		win  bool
	}{{49, false}, {50, true}} {
		g := newCatalogGame(t)
		me := g.Seats[1]
		pushPermanentForTest(g, me.ID, "Test of Endurance", gapTestEndurance, "Enchantment")
		me.Life = tc.life
		advanceToUpkeepOf(t, g, 1)
		passPriorityAroundTable(t, g)
		if got := g.State == game.StateEnded && g.Outcome != nil && g.Outcome.Winner == me.ID; got != tc.win {
			t.Errorf("life %d: won = %v, want %v", tc.life, got, tc.win)
		}
	}
}

func TestGapB141HowlFromBeyondPumpsByX(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	advanceToMain(t, g)
	b06AddMana(me, "B", "B", "B", "B")
	paCast(t, g, "Howl from Beyond", "Instant", "{X}{B}", gapHowl, game.CastSpellParams{XValue: 3, Targets: cardTarget(bear)})
	passPriorityAroundTable(t, g)
	if p, th := gapB141PT(g, bear); p != 5 || th != 2 {
		t.Errorf("got %d/%d, want 5/2", p, th)
	}
}

func TestGapB141RecoilBouncesThenTheOwnerDiscards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	their := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	pitch := pushCatalogHandCard(opp, "Pitch", "Sorcery", "")
	castCatalogSpell(t, g, "Recoil", "Instant", gapRecoil, cardTarget(their))
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(their) {
		t.Fatal("the permanent did not return to its owner's hand")
	}
	// The owner, not the caster, discards a card of their choice.
	answerDiscard(t, g, opp.ID, pitch)
	if !opp.Graveyard.Contains(pitch) || !opp.Hand.Contains(their) {
		t.Error("the owner did not discard the chosen card")
	}
	_ = me
}

func TestGapB141SpinalEmbraceStealsThenSacrificesForLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	troll := blbB1Push(g, opp.ID, "Their Troll", "Creature — Troll", "", 3, 4)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == troll {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	// Not castable in the main phase.
	var cant *game.CantCastError
	advanceToMain(t, g)
	id := pushCatalogHandCard(me, "Spinal Embrace", "Instant", gapSpinalEmbrace)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: cardTarget(troll), Strict: true}); !errors.As(err, &cant) {
		t.Fatalf("main phase cast = %v, want the timing refusal", err)
	}
	advanceTo(t, g, game.StepBeginCombat)
	life := lifeOf(g, me.ID)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: cardTarget(troll)}); err != nil {
		t.Fatalf("cast in combat: %v", err)
	}
	passPriorityAroundTable(t, g)
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(troll)
		if c.Controller != me.ID || c.Tapped {
			t.Errorf("controller %v tapped %v, want me and untapped", c.Controller, c.Tapped)
		}
	})
	if !p1g2HasKeyword(g, troll, "haste") {
		t.Error("the stolen creature lacks haste")
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(troll) {
		t.Error("the creature was not sacrificed at the end step")
	}
	if got := lifeOf(g, me.ID); got != life+4 {
		t.Errorf("life %d, want %d (its toughness gained)", got, life+4)
	}
}

func TestGapB141EmeraldCharmModes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	land := blbB1Push(g, me.ID, "Forest", "Land — Forest", "", 0, 0)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == land {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	advanceToMain(t, g)
	cast := func(mode int, target uuid.UUID) error {
		id := pushCatalogHandCard(me, "Emerald Charm", "Instant", gapEmeraldCharm)
		return g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{mode}, Targets: cardTarget(target), Strict: true})
	}
	// Mode 0: untap.
	b06AddMana(me, "G", "G", "G")
	if err := cast(0, land); err != nil {
		t.Fatalf("untap mode: %v", err)
	}
	passPriorityAroundTable(t, g)
	g.ReadSnapshot(func() {
		if c, _ := g.LookupCardForEffect(land); c.Tapped {
			t.Error("the land stayed tapped")
		}
	})
	// Mode 1: destroys a non-Aura enchantment, refuses an Aura.
	ench := blbB1Push(g, opp.ID, "Ench", "Enchantment", "", 0, 0)
	aura := blbB1Push(g, opp.ID, "Aura", "Enchantment — Aura", "", 0, 0)
	if err := cast(1, aura); err == nil {
		t.Fatal("Emerald Charm targeted an Aura")
	}
	if err := cast(1, ench); err != nil {
		t.Fatalf("destroy mode: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ench) {
		t.Error("the enchantment survived")
	}
	// Mode 2: loses flying.
	bird := blbB1Push(g, opp.ID, "Bird", "Creature — Bird", "", 1, 1)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bird {
				g.Battlefield.Cards[i].Keywords = append(g.Battlefield.Cards[i].Keywords, "flying")
			}
		}
	})
	if !p1g2HasKeyword(g, bird, "flying") {
		t.Fatal("test setup: bird has no flying")
	}
	if err := cast(2, bird); err != nil {
		t.Fatalf("flying mode: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p1g2HasKeyword(g, bird, "flying") {
		t.Error("the bird kept flying")
	}
}

func TestGapB141LullPreventsCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Lull", "Instant", gapLull, nil)
	passPriorityAroundTable(t, g)
	life := lifeOf(g, opp.ID)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	advanceTo(t, g, game.StepEndCombat)
	if got := lifeOf(g, opp.ID); got != life {
		t.Errorf("life %d, want %d: combat damage was not prevented", got, life)
	}
}

func TestGapB141PrimalBoostPumpsAndItsCycleTriggerPumps(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	castCatalogSpell(t, g, "Primal Boost", "Instant", gapPrimalBoost, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if p, th := gapB141PT(g, bear); p != 6 || th != 6 {
		t.Fatalf("cast: got %d/%d, want 6/6", p, th)
	}
	g2 := newCatalogGame(t)
	me2, _, _ := blbB1Seats(g2)
	bear2 := blbB1Push(g2, me2.ID, "Bear", "Creature — Bear", "", 2, 2)
	cycleFromHand(t, g2, "Primal Boost", "Instant", gapPrimalBoost, "{C}{C}{G}")
	assistCardSettle(t, g2, me2.ID, &game.TargetRef{Kind: game.TargetCard, ID: bear2}, true)
	if p, th := gapB141PT(g2, bear2); p != 3 || th != 3 {
		t.Errorf("cycled: got %d/%d, want 3/3", p, th)
	}
}
