package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// skyway_robber_test.go — Skyway Robber (#5880): escape's exiled cards are
// linked to the permanent (CR 607.2q), and the escaped-with combat-damage
// trigger (CR 702.138d) casts an artifact, instant or sorcery spell from
// among them for free.

const skywayRobberOracle = "58aac732-3f03-4551-a684-6cb0e3b31c09"

func TestSkywayRobberIsFull(t *testing.T) {
	spec, ok := effects.Lookup(skywayRobberOracle)
	if !ok {
		t.Fatal("Skyway Robber is not registered")
	}
	if spec.Name != "Skyway Robber" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
	}
	offers := game.AlternativeCostsOfferedFromZone(skywayRobberOracle, game.ZoneGraveyard)
	if len(offers) != 1 || offers[0].Key != "escape" || offers[0].ManaCost != "{3}{U}" ||
		offers[0].ExileFromGraveyard == nil || offers[0].ExileFromGraveyard.Min != 5 {
		t.Fatalf("graveyard offers = %+v, want escape {3}{U} exiling five", offers)
	}
}

type skywayRobberTable struct {
	g       *game.Game
	me, opp *game.Player
	robber  uuid.UUID
	// The five cards the escape exiled.
	instant, sorcery, artifact, creature, land uuid.UUID
}

func skywayRobberGraveCard(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

// skywayRobberEscape escapes the Robber, exiling an instant, a sorcery, an
// artifact, a creature and a land, and resolves it.
func skywayRobberEscape(t *testing.T) skywayRobberTable {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	w := skywayRobberTable{g: g, me: me, opp: opp}
	w.instant = skywayRobberGraveCard(me, "Fast Instant", "Instant", "{2}{U}")
	w.sorcery = skywayRobberGraveCard(me, "Big Sorcery", "Sorcery", "{5}{R}")
	w.artifact = skywayRobberGraveCard(me, "Iron Relic", "Artifact", "{4}")
	w.creature = skywayRobberGraveCard(me, "Plain Bear", "Creature — Bear", "{1}{G}")
	w.land = skywayRobberGraveCard(me, "A Forest", "Basic Land — Forest", "")
	w.robber = effectstest.SeedGraveyardCard(t, g, "Skyway Robber", "Creature — Bird Rogue", skywayRobberOracle)
	pay := []uuid.UUID{w.instant, w.sorcery, w.artifact, w.creature, w.land}
	if err := g.CastSpell(me.ID, w.robber, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "escape", AltCostIDs: pay,
	}); err != nil {
		t.Fatalf("escape cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(w.robber) {
		t.Fatal("the escaped Robber is not on the battlefield")
	}
	return w
}

func skywayRobberLinked(g *game.Game, robber uuid.UUID) map[uuid.UUID]bool {
	out := map[uuid.UUID]bool{}
	g.ReadSnapshot(func() {
		for _, c := range g.Exile.Cards {
			if c.ExiledWith.ID == robber {
				out[c.InstanceID] = true
			}
		}
	})
	return out
}

func TestSkywayRobberEscapeLinksTheFiveExiledCards(t *testing.T) {
	w := skywayRobberEscape(t)
	linked := skywayRobberLinked(w.g, w.robber)
	for _, id := range []uuid.UUID{w.instant, w.sorcery, w.artifact, w.creature, w.land} {
		if !linked[id] {
			t.Errorf("card %s exiled by the escape cost is not linked to the Robber: %v", id, linked)
		}
	}
	if len(linked) != 5 {
		t.Errorf("%d cards linked, want 5", len(linked))
	}
}

func TestSkywayRobberEscapedCastsAnInstantFreeAndRefusesCreaturesAndLands(t *testing.T) {
	w := skywayRobberEscape(t)
	// A card exiled by something else is not "exiled with this creature".
	stray := uuid.New()
	w.g.WithWriteLock(func() {
		w.g.Exile.PushTop(game.Card{InstanceID: stray, Name: "Stray Instant", TypeLine: "Instant",
			ManaCost: "{U}", Owner: w.me.ID, Controller: w.me.ID})
	})

	effectstest.DealCombatDamageToPlayer(w.g, w.robber, w.opp.ID, 3)
	effectstest.PassPriorityUntilChoice(t, w.g)

	offer := effectstest.MustResolutionCastOffer(t, w.g, w.me.ID, game.MayCastKeywordFreeCast, w.instant, w.sorcery, w.artifact)
	for _, bad := range []uuid.UUID{w.creature, w.land, stray} {
		if effectstest.HasID(offer.DiscardOptions, bad) {
			t.Errorf("card %s was offered: %v", bad, offer.DiscardOptions)
		}
	}
	// Strict: a creature card and a land among the linked cards are refused.
	for _, bad := range []uuid.UUID{w.creature, w.land} {
		if err := w.g.CastSpell(w.me.ID, bad, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
			t.Errorf("card %s was cast from among the cards exiled with the Robber", bad)
		}
		if !effectstest.MscSqInExile(w.g, bad) {
			t.Errorf("the refused cast moved card %s", bad)
		}
	}

	// The instant is cast for no mana (the pool is empty), inside the resolution.
	if err := effectstest.CastThroughResolutionCast(t, w.g, w.me.ID, w.instant, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("free cast of the instant: %v", err)
	}
	effectstest.CastInsideTheResolution(t, w.g, w.instant)
	if offer := effectstest.ResolutionCastOfferFor(w.g, w.me.ID); offer != nil {
		t.Error("the offer outlived its one cast")
	}
	if err := w.g.CastSpell(w.me.ID, w.sorcery, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("a second spell was cast from the same trigger")
	}
	effectstest.PassPriorityAroundTable(t, w.g)
	if effectstest.MscSqInExile(w.g, w.instant) {
		t.Error("the cast instant is still in exile")
	}
	for _, id := range []uuid.UUID{w.sorcery, w.artifact, w.creature, w.land} {
		if !effectstest.MscSqInExile(w.g, id) {
			t.Errorf("uncast card %s left exile", id)
		}
	}
}

func TestSkywayRobberCastsAnArtifactFree(t *testing.T) {
	w := skywayRobberEscape(t)
	effectstest.DealCombatDamageToPlayer(w.g, w.robber, w.opp.ID, 3)
	effectstest.PassPriorityUntilChoice(t, w.g)
	effectstest.MustResolutionCastOffer(t, w.g, w.me.ID, game.MayCastKeywordFreeCast, w.artifact)
	if err := effectstest.CastThroughResolutionCast(t, w.g, w.me.ID, w.artifact, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("free cast of the artifact: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, w.g)
	if !w.g.Battlefield.Contains(w.artifact) {
		t.Error("the free artifact should resolve onto the battlefield")
	}
}

func TestSkywayRobberDecliningLeavesTheCardsForTheNextHit(t *testing.T) {
	w := skywayRobberEscape(t)
	effectstest.DealCombatDamageToPlayer(w.g, w.robber, w.opp.ID, 3)
	effectstest.PassPriorityUntilChoice(t, w.g)
	effectstest.MustResolutionCastOffer(t, w.g, w.me.ID, game.MayCastKeywordFreeCast, w.instant)
	effectstest.DeclineResolutionCast(t, w.g, w.me.ID)
	if !effectstest.MscSqInExile(w.g, w.instant) {
		t.Fatal("declining should leave the instant in exile")
	}
	if err := w.g.CastSpell(w.me.ID, w.instant, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Fatal("the declined offer left the instant castable")
	}
	effectstest.DealCombatDamageToPlayer(w.g, w.robber, w.opp.ID, 3)
	effectstest.PassPriorityUntilChoice(t, w.g)
	effectstest.MustResolutionCastOffer(t, w.g, w.me.ID, game.MayCastKeywordFreeCast, w.instant)
}

func TestSkywayRobberCastFromHandHasNoTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	instant := uuid.New()
	g.WithWriteLock(func() {
		g.Exile.PushTop(game.Card{InstanceID: instant, Name: "Fast Instant", TypeLine: "Instant",
			ManaCost: "{2}{U}", Owner: me.ID, Controller: me.ID})
	})
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Skyway Robber", TypeLine: "Creature — Bird Rogue",
		Power: 3, Toughness: 3, OracleID: skywayRobberOracle, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("hand cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the hard-cast Robber is not on the battlefield")
	}
	before := g.Stack.Size()
	effectstest.DealCombatDamageToPlayer(g, id, opp.ID, 3)
	g.RunStateChecksForTest()
	if got := g.Stack.Size(); got != before {
		t.Fatalf("a hard-cast Robber put %d items on the stack", got-before)
	}
	if offer := effectstest.ResolutionCastOfferFor(g, me.ID); offer != nil {
		t.Error("a hard-cast Robber offered a cast")
	}
}

// A Robber that died and escaped again is a new object (CR 400.7): it reads
// only the five its new escape cost exiled, not the first pile.
func TestSkywayRobberEscapedAgainReadsOnlyItsNewPile(t *testing.T) {
	w := skywayRobberEscape(t)
	if err := w.g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: w.me.ID}, w.robber); err != nil {
		t.Fatalf("moving the Robber to the graveyard: %v", err)
	}
	var pay []uuid.UUID
	for i := 0; i < 5; i++ {
		pay = append(pay, skywayRobberGraveCard(w.me, "Second Instant", "Instant", "{1}{U}"))
	}
	if err := w.g.CastSpell(w.me.ID, w.robber, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "escape", AltCostIDs: pay,
	}); err != nil {
		t.Fatalf("second escape cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, w.g)
	effectstest.DealCombatDamageToPlayer(w.g, w.robber, w.opp.ID, 3)
	effectstest.PassPriorityUntilChoice(t, w.g)
	offer := effectstest.MustResolutionCastOffer(t, w.g, w.me.ID, game.MayCastKeywordFreeCast, pay...)
	for _, old := range []uuid.UUID{w.instant, w.sorcery, w.artifact} {
		if effectstest.HasID(offer.DiscardOptions, old) {
			t.Errorf("a card from the first escape's pile was offered: %s", old)
		}
	}
}
