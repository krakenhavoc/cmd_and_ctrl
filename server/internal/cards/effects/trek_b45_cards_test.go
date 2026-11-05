package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b45_cards_test.go — Star Trek Commander, batch 45 (#4974). Each
// test plays the real catalog entry and asserts what its text says.

const (
	trekB45AggressorOracle = "8f1a39bf-7597-4655-b7af-4b1332edf529"
	trekB45DirectHitOracle = "f6b0adcb-4676-4d9f-83ae-a2ff3cdfd84b"
	trekB45GintakOracle    = "286213a8-ae53-4f97-9f81-e13bed462d0a"
	trekB45KrugeOracle     = "61cef71a-d38b-4b6f-8401-aa4fc861bf52"
	trekB45MeklethOracle   = "7063eb2b-81a4-4599-b1db-edf861e39bfb"
	trekB45FlightOracle    = "d55977f8-1655-4d97-adbf-ef6cb979ea55"
	trekB45GornOracle      = "60ff591d-170d-410a-9710-510e811f9827"
	trekB45GumatoOracle    = "67114cbb-cbf4-4250-a5e9-e1c4bc5841f6"
	trekB45SpiderOracle    = "369f086f-3497-496a-94a8-0faccd1d5931"
	trekB45PistolOracle    = "bb8c2c0f-fc63-40c6-98c6-691e42b3b8c9"
)

func TestTrekB45EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB45AggressorOracle: "Defense Force Aggressor",
		trekB45DirectHitOracle: "Direct Hit",
		trekB45GintakOracle:    "Gin'tak Charge",
		trekB45KrugeOracle:     "Kruge, Genesis Seeker",
		trekB45MeklethOracle:   "Mek'leth Berserker",
		trekB45FlightOracle:    "Warship Flight Crew",
		trekB45GornOracle:      "Gorn Captain",
		trekB45GumatoOracle:    "Gumato",
		trekB45SpiderOracle:    "Talarian Hook Spider",
		trekB45PistolOracle:    "Disruptor Pistol",
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

// --- vanilla and keyword bodies -------------------------------------------------

func TestTrekB45VanillaCreaturesResolveOntoTheBattlefield(t *testing.T) {
	for _, c := range []struct {
		name, typeLine, oracle string
		p, tough               int
	}{
		{"Defense Force Aggressor", "Creature — Klingon Warrior", trekB45AggressorOracle, 2, 1},
		{"Warship Flight Crew", "Creature — Klingon Pilot", trekB45FlightOracle, 2, 2},
	} {
		g := newCatalogGame(t)
		id := castHoldCreature(t, g, c.name, c.typeLine, c.oracle, c.p, c.tough)
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s did not resolve onto the battlefield", c.name)
			continue
		}
		if got := layeredCard(t, g, id); got.CurrentPower() != c.p || got.CurrentToughness() != c.tough || len(effectiveAbilities(t, g, id)) != 0 {
			t.Errorf("%s is %d/%d with %v, want a vanilla %d/%d", c.name, got.CurrentPower(), got.CurrentToughness(), effectiveAbilities(t, g, id), c.p, c.tough)
		}
	}
}

func TestTrekB45KeywordCreaturesCarryTheirKeyword(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	assertKeywords(t, g, b11Push(g, me.ID, "Mek'leth Berserker", "Creature — Klingon Berserker", trekB45MeklethOracle, 4, 3), "menace")
	assertKeywords(t, g, b11Push(g, me.ID, "Gorn Captain", "Creature — Gorn Soldier", trekB45GornOracle, 4, 3), "trample")
	assertKeywords(t, g, b11Push(g, me.ID, "Talarian Hook Spider", "Creature — Spider", trekB45SpiderOracle, 2, 2), "reach")
	assertKeywords(t, g, b11Push(g, me.ID, "Gumato", "Creature — Ape Beast", trekB45GumatoOracle, 4, 5), "vigilance")
}

func TestTrekB45MeklethBerserkerCantBeBlockedByOneCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	berserker := b11Push(g, me.ID, "Mek'leth Berserker", "Creature — Klingon Berserker", trekB45MeklethOracle, 4, 3)
	wall := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 5)
	declareAttack(t, g, opp.ID, berserker)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, berserker); err == nil {
		if err := g.FinishBlocks(opp.ID); err == nil {
			t.Fatal("one creature blocked a creature with menace")
		}
	}
}

func TestTrekB45TalarianHookSpiderBlocksAFlier(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	flier := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "My Bird", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Keywords: []string{"flying"}, Owner: me.ID, Controller: me.ID})
	spider := b11Push(g, opp.ID, "Talarian Hook Spider", "Creature — Spider", trekB45SpiderOracle, 2, 2)
	declareAttack(t, g, opp.ID, flier)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(spider, flier); err != nil {
		t.Fatalf("the reach Spider could not block a flier: %v", err)
	}
}

// --- Direct Hit ---------------------------------------------------------------

func TestTrekB45DirectHitDealsFiveToACreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	giant := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 9)
	if _, err := trekB21CastStrict(t, g, "Direct Hit", "Sorcery", "{2}{R}", trekB45DirectHitOracle, "{R}{C}{C}", cardRefs(giant)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := damageMarkedOn(g, giant); got != 5 {
		t.Errorf("damage %d, want 5", got)
	}
}

func TestTrekB45DirectHitCantTargetAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	if _, err := trekB21CastStrict(t, g, "Direct Hit", "Sorcery", "{2}{R}", trekB45DirectHitOracle, "{R}{C}{C}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err == nil {
		t.Fatal("Direct Hit targeted a player")
	}
}

// --- Gin'tak Charge -----------------------------------------------------------

func TestTrekB45GintakChargeHitsAPlayerOrACreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	life := opp.Life
	if _, err := trekB21CastStrict(t, g, "Gin'tak Charge", "Sorcery", "{R}", trekB45GintakOracle, "{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
		t.Fatalf("cast at a player: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d → %d, want -2", life, opp.Life)
	}
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 3)
	if _, err := trekB21CastStrict(t, g, "Gin'tak Charge", "Sorcery", "{R}", trekB45GintakOracle, "{R}", cardRefs(bear)); err != nil {
		t.Fatalf("cast at a creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := damageMarkedOn(g, bear); got != 2 {
		t.Errorf("damage %d, want 2", got)
	}
}

func TestTrekB45GintakChargeIsSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceTo(t, g, game.StepUpkeep)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Gin'tak Charge", TypeLine: "Sorcery", ManaCost: "{R}",
		OracleID: trekB45GintakOracle, Owner: me.ID, Controller: me.ID})
	floatMana(t, g, me, "{R}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, Strict: true}); err == nil {
		t.Fatal("a sorcery was cast in the upkeep")
	}
}

// --- Kruge, Genesis Seeker ------------------------------------------------------

func TestTrekB45KrugePumpsAttackingKlingonsYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	kruge := b11Push(g, me.ID, "Kruge, Genesis Seeker", "Legendary Creature — Klingon Officer", trekB45KrugeOracle, 2, 2)
	klingon := b12Creature(g, me.ID, "My Klingon", "Creature — Klingon Warrior", 2, 1)
	human := b12Creature(g, me.ID, "My Human", "Creature — Human Soldier", 2, 2)
	home := b12Creature(g, me.ID, "Home Klingon", "Creature — Klingon Warrior", 2, 1)
	theirs := b12Creature(g, opp.ID, "Their Klingon", "Creature — Klingon Warrior", 2, 1)

	declareAttack(t, g, opp.ID, kruge, klingon, human)
	passPriorityAroundTable(t, g)
	for _, c := range []struct {
		id   uuid.UUID
		name string
		want int
	}{{kruge, "Kruge", 3}, {klingon, "the attacking Klingon", 3}, {human, "the attacking Human", 2}, {home, "the Klingon at home", 2}, {theirs, "an opponent's Klingon", 2}} {
		if got := powerOf(t, g, c.id); got != c.want {
			t.Errorf("%s has power %d, want %d", c.name, got, c.want)
		}
	}

	endTurn(t, g)
	if got := powerOf(t, g, klingon); got != 2 {
		t.Errorf("the bonus outlived the turn: power %d", got)
	}
}

func TestTrekB45KrugeTriggersOnlyOnHisOwnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	kruge := b11Push(g, me.ID, "Kruge, Genesis Seeker", "Legendary Creature — Klingon Officer", trekB45KrugeOracle, 2, 2)
	klingon := b12Creature(g, me.ID, "My Klingon", "Creature — Klingon Warrior", 2, 1)
	declareAttack(t, g, opp.ID, klingon)
	if n := triggersOnStackFrom(g, kruge); n != 0 {
		t.Errorf("%d Kruge triggers when only another creature attacked", n)
	}
	passPriorityAroundTable(t, g)
	if got := powerOf(t, g, klingon); got != 2 {
		t.Errorf("the Klingon has power %d without Kruge attacking, want 2", got)
	}
}

// --- Gumato -------------------------------------------------------------------

func TestTrekB45GumatoTapsForThreeGreen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	gumato := b11Push(g, me.ID, "Gumato", "Creature — Ape Beast", trekB45GumatoOracle, 4, 5)
	tapForMana(t, g, me.ID, gumato)
	if got := trekB21GreenInPool(me); got != 3 || len(poolColors(me)) != 3 {
		t.Errorf("pool %v, want {G}{G}{G}", poolColors(me))
	}
}

func TestTrekB45GumatoCantTapWhileSummoningSick(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gumato := castHoldCreature(t, g, "Gumato", "Creature — Ape Beast", trekB45GumatoOracle, 4, 5)
	if err := g.ActivateManaAbility(me.ID, gumato, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("a summoning-sick Gumato tapped for mana")
	}
}

func TestTrekB45GumatoAttacksWithoutTapping(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	gumato := b11Push(g, me.ID, "Gumato", "Creature — Ape Beast", trekB45GumatoOracle, 4, 5)
	declareAttack(t, g, opp.ID, gumato)
	if layeredCard(t, g, gumato).Tapped {
		t.Error("attacking tapped a creature with vigilance")
	}
}

// --- Disruptor Pistol ----------------------------------------------------------

func TestTrekB45DisruptorPistolSacrificesForFiveDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pistol := b27Push(g, me.ID, "Disruptor Pistol", "Artifact", trekB45PistolOracle, "{2}", 0, 0)
	giant := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 9)
	advanceTo(t, g, game.StepUpkeep)
	floatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, pistol, 0, game.ActivateAbilityParams{Targets: cardRefs(giant), Strict: true}); err == nil {
		t.Fatal("{3} paid a {4} cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, pistol, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, Strict: true}); err == nil {
		t.Fatal("the Pistol targeted a player")
	}
	floatMana(t, g, me, "{C}")
	// Not sorcery-speed: the upkeep is fine.
	if err := g.ActivateCatalogAbility(me.ID, pistol, 0, game.ActivateAbilityParams{Targets: cardRefs(giant), Strict: true}); err != nil {
		t.Fatalf("activate for {4} in the upkeep: %v", err)
	}
	if g.Battlefield.Contains(pistol) {
		t.Error("the Pistol was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	if got := damageMarkedOn(g, giant); got != 5 {
		t.Errorf("damage %d, want 5", got)
	}
}
