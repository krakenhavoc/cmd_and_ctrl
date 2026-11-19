package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	bnsGlimmerbellOracle   = "1a8b7792-e105-40c8-b1b4-ab1a8f3731d8"
	bnsGnawingZombieOracle = "7979c1e6-6bc5-4b8c-82b0-2770e67ba898"
	bnsArsonistOracle      = "c1177f22-a1cf-4da3-a68d-ff954e878403"
	bnsDarkDwellersOracle  = "bab0ab8d-74d2-49a3-b258-b93d19925d99"
	bnsFireslingerOracle   = "396564be-3747-49dd-98dc-afb81a375f18"
	bnsTrailblazerOracle   = "38f66ef9-6afc-478b-a1d4-a9d42c31dfc9"
	bnsGomaFadaOracle      = "74027d31-6efb-471d-9499-cdb1cefc9eae"
	bnsGravediggerOracle   = "1a2030cc-d7ee-4059-b2d7-fb95ea8e267b"
	bnsGladehartOracle     = "f19f28e5-9cad-4398-b2d4-9e7fefb23cb4"
	bnsSentinelOracle      = "8178c546-84f8-4563-a804-6489ec016660"
)

func TestBnsB31EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		bnsGlimmerbellOracle:   "Glimmerbell",
		bnsGnawingZombieOracle: "Gnawing Zombie",
		bnsArsonistOracle:      "Goblin Arsonist",
		bnsDarkDwellersOracle:  "Goblin Dark-Dwellers",
		bnsFireslingerOracle:   "Goblin Fireslinger",
		bnsTrailblazerOracle:   "Goblin Trailblazer",
		bnsGomaFadaOracle:      "Goma Fada Vanguard",
		bnsGravediggerOracle:   "Gravedigger",
		bnsGladehartOracle:     "Grazing Gladehart",
		bnsSentinelOracle:      "Greenwood Sentinel",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v registered=%v", name, spec, ok)
		}
	}
}

func TestBnsB31PrintedKeywords(t *testing.T) {
	for oracle, kw := range map[string]string{
		bnsGlimmerbellOracle: "flying", bnsDarkDwellersOracle: "menace",
		bnsTrailblazerOracle: "menace", bnsSentinelOracle: "vigilance",
	} {
		spec, _ := Lookup(oracle)
		if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != kw {
			t.Errorf("%s: keywords %v, want [%s]", spec.Name, spec.PrintedKeywords, kw)
		}
	}
}

func TestBnsB31GlimmerbellUntapsForOneBlue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogPermanent(g, me.ID, "Glimmerbell", "Creature — Elemental Jellyfish", bnsGlimmerbellOracle, false)
	g.WithWriteLock(func() { g.Battlefield.Cards[len(g.Battlefield.Cards)-1].Tapped = true })
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("unpaid activation: %v", err)
	}
	floatForTest(g, me, "UU")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, id); c.Tapped {
		t.Error("Glimmerbell is still tapped")
	}
}

func TestBnsB31GnawingZombieDrainsOneBySacrificingACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, foe := g.Seats[0], g.Seats[1]
	zombie := pushCatalogPermanent(g, me.ID, "Gnawing Zombie", "Creature — Zombie", bnsGnawingZombieOracle, false)
	fodder := pushVanillaCreature(g, me.ID, "Fodder", 1, 1)
	advanceToMain(t, g)
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: foe.ID}}
	if err := g.ActivateCatalogAbility(me.ID, zombie, 0, game.ActivateAbilityParams{Strict: true, Targets: target, SacrificeIDs: []uuid.UUID{fodder}}); !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("unpaid activation: %v", err)
	}
	floatForTest(g, me, "BB")
	mine, theirs := me.Life, foe.Life
	if err := g.ActivateCatalogAbility(me.ID, zombie, 0, game.ActivateAbilityParams{Strict: true, Targets: target, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) || foe.Life != theirs-1 || me.Life != mine+1 {
		t.Errorf("fodder on battlefield %v, foe %d->%d, me %d->%d", g.Battlefield.Contains(fodder), theirs, foe.Life, mine, me.Life)
	}
}

func TestBnsB31GoblinArsonistMayPingWhenItDies(t *testing.T) {
	for _, accept := range []bool{true, false} {
		g := newCatalogGame(t)
		me, foe := g.Seats[0], g.Seats[1]
		id := pushCatalogPermanent(g, me.ID, "Goblin Arsonist", "Creature — Goblin Shaman", bnsArsonistOracle, false)
		before := foe.Life
		killCreature(t, g, me.ID, id)
		if pick := latestChoiceOfKindFor(g, game.PendingChoicePickTarget, me.ID); pick != nil {
			if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: foe.ID}); err != nil {
				t.Fatal(err)
			}
		}
		passPriorityAroundTable(t, g)
		answerLatestTriggerPrompt(t, g, me.ID, accept)
		passPriorityAroundTable(t, g)
		want := before
		if accept {
			want--
		}
		if foe.Life != want {
			t.Errorf("accept=%v: foe life %d, want %d", accept, foe.Life, want)
		}
	}
}

func TestBnsB31GoblinDarkDwellersOffersOnlyCheapInstantsAndSorceries(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cheap := gyCard(me, "Old Sorcery", "Sorcery", "{2}{R}", "")
	instant := gyCard(me, "Old Instant", "Instant", "{R}", "")
	dear := gyCard(me, "Big Sorcery", "Sorcery", "{3}{R}", "")
	creature := gyCard(me, "Old Bear", "Creature — Bear", "{1}{G}", "")
	castHoldCreature(t, g, "Goblin Dark-Dwellers", "Creature — Goblin", bnsDarkDwellersOracle, 4, 4)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	for _, id := range []uuid.UUID{cheap, instant} {
		if !hasID(p.PickTargetCards, id) {
			t.Errorf("a legal card was not offered: %v", id)
		}
	}
	for _, id := range []uuid.UUID{dear, creature} {
		if hasID(p.PickTargetCards, id) {
			t.Errorf("an illegal card was offered: %v", id)
		}
	}
}

func TestBnsB31GoblinDarkDwellersCastsItFreeAndExilesIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sorcery := gyCard(me, "Old Sorcery", "Sorcery", "{2}{R}", "")
	castHoldCreature(t, g, "Goblin Dark-Dwellers", "Creature — Goblin", bnsDarkDwellersOracle, 4, 4)
	pickCard(t, g, me.ID, sorcery)
	passPriorityAroundTable(t, g)
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, sorcery)
	if err := castThroughResolutionCast(t, g, me.ID, sorcery, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("casting for free: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(sorcery) || !inExile(g, sorcery) {
		t.Error("the sorcery should be exiled, not back in the graveyard")
	}
}

func TestBnsB31GoblinFireslingerPingsAPlayerButNotACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, foe := g.Seats[0], g.Seats[1]
	id := pushCatalogPermanent(g, me.ID, "Goblin Fireslinger", "Creature — Goblin Warrior", bnsFireslingerOracle, false)
	bear := pushVanillaCreature(g, foe.ID, "Their Bear", 2, 2)
	advanceToMain(t, g)
	err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("a creature target: %v", err)
	}
	before := foe.Life
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: foe.ID}}}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if foe.Life != before-1 {
		t.Errorf("foe life %d -> %d", before, foe.Life)
	}
}

func TestBnsB31GomaFadaOnlyTargetsPowerAtMostItsWarriors(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	goma := b39Creature(g, me.ID, "Goma Fada Vanguard", "Creature — Human Warrior", 2, 2)
	g.WithWriteLock(func() { g.Battlefield.Cards[len(g.Battlefield.Cards)-1].OracleID = bnsGomaFadaOracle })
	b39Creature(g, me.ID, "Second Warrior", "Creature — Human Warrior", 1, 1)
	small := pushVanillaCreature(g, opp.ID, "Small", 2, 2)
	big := pushVanillaCreature(g, opp.ID, "Big", 3, 3)
	declareAttack(t, g, opp.ID, goma)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if !hasID(p.PickTargetCards, small) || hasID(p.PickTargetCards, big) {
		t.Fatalf("offered %v, want only the power-2 creature", p.PickTargetCards)
	}
	pickCard(t, g, me.ID, small)
	passPriorityAroundTable(t, g)
	if !game.Restricted(findBattlefieldCardByID(g, small), game.CantBlock) {
		t.Error("the targeted creature can still block")
	}
	if game.Restricted(findBattlefieldCardByID(g, big), game.CantBlock) {
		t.Error("an untargeted creature was restricted")
	}
}

func TestBnsB31GravediggerMayReturnACreatureCard(t *testing.T) {
	for _, accept := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		bear := pushGraveyardPermanent(me, "Bear", "Creature — Bear", "{1}{G}")
		spell := gyCard(me, "Bolt", "Instant", "{R}", "")
		castHoldCreature(t, g, "Gravedigger", "Creature — Zombie", bnsGravediggerOracle, 2, 2)
		p := latestPickTarget(g, me.ID)
		if p == nil || !hasID(p.PickTargetCards, bear) || hasID(p.PickTargetCards, spell) {
			t.Fatalf("pick: %+v", p)
		}
		pickCard(t, g, me.ID, bear)
		answerLatestTriggerPrompt(t, g, me.ID, accept)
		passPriorityAroundTable(t, g)
		if me.Hand.Contains(bear) != accept {
			t.Errorf("accept=%v: bear in hand = %v", accept, me.Hand.Contains(bear))
		}
	}
}

func TestBnsB31GrazingGladehartLandfallMayGainTwo(t *testing.T) {
	for _, accept := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		pushCatalogPermanent(g, me.ID, "Grazing Gladehart", "Creature — Antelope", bnsGladehartOracle, false)
		before := me.Life
		playLandFromHand(t, g, "Forest", "")
		answerLatestTriggerPrompt(t, g, me.ID, accept)
		passPriorityAroundTable(t, g)
		want := before
		if accept {
			want += 2
		}
		if me.Life != want {
			t.Errorf("accept=%v: life %d, want %d", accept, me.Life, want)
		}
	}
}
