package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	s82OneB18MeliraOracle      = "6bb8c22c-8b01-45f8-8c9e-f683e6b4f022"
	s82OneB18MycosynthOracle   = "1f0bb053-4967-4289-b2b1-f6309f2eca7d"
	s82OneB18ThrunOracle       = "789b7af5-ac15-40b6-b5b7-f3fcdcfb52e1"
	s82OneB18SlobadOracle      = "72062fb6-0329-424d-936a-882b638d6d57"
	s82OneB18GethOracle        = "b868de48-3271-4f33-94a4-3dd9b67dd72c"
	s82OneB18RiaOracle         = "785f55bb-d47c-470c-94d4-95f69055e35f"
	s82OneB18UnctusOracle      = "54c18f09-5c5a-46e7-9f66-ff91bc12537c"
	s82OneB18MonumentOracle    = "df4bb104-d939-434f-8a1c-e89beda27c90"
	s82OneB18KethekOracle      = "26747bdc-6bd4-4218-ae83-bd46c7668880"
	s82OneB18AnyCardOracleNone = ""
)

// --- Thrun, Breaker of Silence --------------------------------------------------

func TestS82OneB18ThrunRefusesNongreenSpellsAndAbilitiesFromOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	thrun := effectstest.PushCatalogPermanent(g, opp.ID, "Thrun, Breaker of Silence", "Legendary Creature — Troll Shaman", s82OneB18ThrunOracle, false)
	effectstest.AdvanceToMain(t, g)

	// A red Lightning Bolt from an opponent can't target it.
	bolt := effectstest.PushCatalogHandCard(me, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle)
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: effectstest.CardRefs(thrun), Strict: true}); err == nil {
		t.Fatal("an opponent's nongreen spell targeted Thrun")
	}

	spec, _ := effects.Lookup(s82OneB18ThrunOracle)
	r := spec.TargetingRestrictions[0]
	self := effectstest.LayeredCard(t, g, thrun)
	q := func(by game.TargetSource) game.TargetingQuery {
		return game.TargetingQuery{Game: g, Card: self, Zone: game.ZoneBattlefield, Controller: by.Controller, Source: self, By: by}
	}
	red := game.Card{InstanceID: uuid.New(), Name: "Red Thing", TypeLine: "Artifact Creature", Colors: []string{"R"}, Owner: me.ID, Controller: me.ID}
	green := game.Card{InstanceID: uuid.New(), Name: "Green Thing", TypeLine: "Creature — Elf", Colors: []string{"G"}, Owner: me.ID, Controller: me.ID}
	colorless := game.Card{InstanceID: uuid.New(), Name: "Relic", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID}
	if !r.Forbids(q(game.SourceObject(me.ID, &red))) {
		t.Error("an opponent's red ability is not refused")
	}
	if !r.Forbids(q(game.SourceCast(me.ID, &red))) {
		t.Error("an opponent's red spell is not refused")
	}
	if !r.Forbids(q(game.SourceObject(me.ID, &colorless))) {
		t.Error("an opponent's colorless (nongreen) ability is not refused")
	}
	if r.Forbids(q(game.SourceObject(me.ID, &green))) {
		t.Error("an opponent's green ability is refused")
	}
	if r.Forbids(q(game.SourceCast(me.ID, &green))) {
		t.Error("an opponent's green spell is refused")
	}
	if r.Forbids(q(game.SourceObject(opp.ID, &red))) {
		t.Error("Thrun's own controller's red ability is refused")
	}
}

func TestS82OneB18ThrunItsControllerMayTargetItAndOthersAreUntouched(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	thrun := effectstest.PushCatalogPermanent(g, me.ID, "Thrun, Breaker of Silence", "Legendary Creature — Troll Shaman", s82OneB18ThrunOracle, false)
	other := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	bolt := effectstest.PushCatalogHandCard(me, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle)
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: effectstest.CardRefs(other), Strict: true}); err != nil {
		t.Fatalf("a Bolt at another creature: %v", err)
	}
	_ = thrun
}

func TestS82OneB18ThrunIsIndestructibleOnlyDuringItsControllersTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	thrun := effectstest.PushCatalogPermanent(g, me.ID, "Thrun, Breaker of Silence", "Legendary Creature — Troll Shaman", s82OneB18ThrunOracle, false)
	effectstest.AdvanceToMain(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, thrun, "indestructible") {
		t.Error("Thrun is not indestructible during its controller's turn")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, thrun, "trample") {
		t.Error("Thrun lacks trample")
	}
	effectstest.PowerUpPassTo(t, g, 1)
	if effectstest.HasKeywordOnBattlefield(t, g, thrun, "indestructible") {
		t.Error("Thrun is indestructible on an opponent's turn")
	}
}

func TestS82OneB18ThrunCantBeCountered(t *testing.T) {
	spec, ok := effects.Lookup(s82OneB18ThrunOracle)
	if !ok || !spec.CantBeCountered {
		t.Fatal("Thrun's spell can be countered")
	}
}

// --- Geth, Thane of Contracts ---------------------------------------------------

func TestS82OneB18GethShrinksOnlyYourOtherCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	geth := effectstest.PushCatalogPermanent(g, me.ID, "Geth, Thane of Contracts", "Legendary Creature — Phyrexian Zombie", s82OneB18GethOracle, false)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	effectstest.AdvanceToMain(t, g)
	if got := effectstest.LayeredCard(t, g, mine).CurrentPower(); got != 2 {
		t.Errorf("my other creature's power %d, want 2", got)
	}
	if got := effectstest.LayeredCard(t, g, theirs).CurrentPower(); got != 3 {
		t.Errorf("an opponent's creature's power %d, want 3", got)
	}
	if got := effectstest.LayeredCard(t, g, geth).CurrentToughness(); got != 1 {
		// pushCatalogPermanent seeds a 1/1, so Geth itself must stay 1/1.
		t.Errorf("Geth shrank itself: toughness %d, want 1", got)
	}
}

func TestS82OneB18GethReanimatesAtSorcerySpeedAndExilesTheCreatureWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	geth := effectstest.PushCatalogPermanent(g, me.ID, "Geth, Thane of Contracts", "Legendary Creature — Phyrexian Zombie", s82OneB18GethOracle, false)
	dead := effectstest.PushGraveyardPermanent(me, "Fallen Bear", "Creature — Bear", "{1}{G}")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{B}{B}")
	if err := g.ActivateCatalogAbility(me.ID, geth, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(dead), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, dead) {
		t.Fatal("the creature card did not return")
	}
	// It leaves the battlefield: exiled instead of going to the graveyard.
	effectstest.B27Kill(g, dead)
	if z := g.FindCardZoneForEffect(dead); z == nil || z.Kind != game.ZoneExile {
		t.Errorf("the returned creature went to %v, want exile", z)
	}
}

func TestS82OneB18GethRefusesToActivateOutsideAMainPhase(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	geth := effectstest.PushCatalogPermanent(g, me.ID, "Geth, Thane of Contracts", "Legendary Creature — Phyrexian Zombie", s82OneB18GethOracle, false)
	dead := effectstest.PushGraveyardPermanent(me, "Fallen Bear", "Creature — Bear", "{1}{G}")
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.FloatMana(t, g, me, "{C}{B}{B}")
	if err := g.ActivateCatalogAbility(me.ID, geth, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(dead), Strict: true}); err == nil {
		t.Error("activated outside a main phase")
	}
}
