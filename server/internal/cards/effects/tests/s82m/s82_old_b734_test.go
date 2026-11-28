package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b734_test.go — Portal Second Age, batch 734 (#5625): enters and
// dies triggers. Each test plays the real catalog entry.

const (
	oldB734Angel     = "07eb70cf-e6af-4924-a00d-bdf2c6ba78e1"
	oldB734Predatory = "9d2ad65f-fba0-45e9-a5e6-bec63a9db979"
	oldB734Brutal    = "a4003cd1-e98c-4346-b685-89c7891f9920"
	oldB734Foul      = "d8177d08-a603-42e9-9350-890eac05e184"
	oldB734Talas     = "da359465-f823-4989-9acd-249a2a1f3f17"
	oldB734Horror    = "db7c5a2f-9aea-4b14-86cf-7af79d6484af"

	oldB734BearType = "Creature — Bear"
	oldB734NsType   = "Creature — Nightstalker"
)

func TestOldB734EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB734Angel: "Angel of Fury", oldB734Predatory: "Predatory Nightstalker",
		oldB734Brutal: "Brutal Nightstalker", oldB734Foul: "Foul Spirit",
		oldB734Talas: "Talas Explorer", oldB734Horror: "Hidden Horror",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v; want Full with none", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB734InGraveyard(p *game.Player, id uuid.UUID) bool {
	for _, c := range p.Graveyard.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

func oldB734InLibrary(p *game.Player, id uuid.UUID) bool {
	for _, c := range p.Library.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

func oldB734Die(t *testing.T, g *game.Game, owner *game.Player, id uuid.UUID) {
	t.Helper()
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: owner.ID}, id); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
}

func TestOldB734AngelOfFuryFliesAndMayShuffleItselfAwayWhenItDies(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.SeedLibrary(me, "A", "B", "C")
		angel := effectstest.B12Push(g, me.ID, "Angel of Fury", "Creature — Angel", oldB734Angel, 3, 5)
		effectstest.AssertKeywords(t, g, angel, "flying")
		oldB734Die(t, g, me, angel)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		if got := oldB734InLibrary(me, angel); got != yes {
			t.Errorf("answered %v: in library = %v", yes, got)
		}
		if got := oldB734InGraveyard(me, angel); got == yes {
			t.Errorf("answered %v: in graveyard = %v", yes, got)
		}
	}
}

func TestOldB734PredatoryNightstalkerMakesAnOpponentSacrificeACreatureOfTheirChoice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	keep := effectstest.B12Creature(g, opp.ID, "Keeper", oldB734BearType, 3, 3)
	lose := effectstest.B12Creature(g, opp.ID, "Loser", oldB734BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Predatory Nightstalker", oldB734NsType, oldB734Predatory, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.AnswerSacrifice(t, g, opp.ID, lose)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldByID(g, lose) || !effectstest.OnBattlefieldByID(g, keep) {
		t.Errorf("the opponent's choice was not honoured: loser on battlefield %v, keeper %v", effectstest.OnBattlefieldByID(g, lose), effectstest.OnBattlefieldByID(g, keep))
	}
}

func TestOldB734PredatoryNightstalkerDeclinedSacrificesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB734BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Predatory Nightstalker", oldB734NsType, oldB734Predatory, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OnBattlefieldByID(g, bear) {
		t.Error("the opponent sacrificed after the controller declined")
	}
}

func TestOldB734BrutalNightstalkerMakesAnOpponentDiscardACardOfTheirChoice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	keep := uuid.New()
	pitch := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: keep, Name: "Keeper", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	opp.Hand.PushTop(game.Card{InstanceID: pitch, Name: "Pitched", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	effectstest.CastCatalogSpell(t, g, "Brutal Nightstalker", oldB734NsType, oldB734Brutal, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.AnswerDiscard(t, g, opp.ID, pitch)
	if !oldB734InGraveyard(opp, pitch) || oldB734InGraveyard(opp, keep) {
		t.Errorf("the opponent did not discard the card they chose")
	}
}

func TestOldB734FoulSpiritFliesAndSacrificesALandWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	forest := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	id := effectstest.CastCatalogSpell(t, g, "Foul Spirit", "Creature — Spirit", oldB734Foul, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if pick := effectstest.LatestChooseCardsFor(g, me.ID); pick != nil {
		if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{forest}); err != nil {
			t.Fatalf("ResolveChooseCards: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertKeywords(t, g, id, "flying")
	if effectstest.OnBattlefieldByID(g, forest) {
		t.Error("the land was not sacrificed")
	}
	if !effectstest.OnBattlefieldByID(g, id) {
		t.Error("the Spirit itself left the battlefield")
	}
}

func TestOldB734TalasExplorerFliesAndLooksAtAnOpponentsHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	other := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	effectstest.HandCardFull(opp, "Secret", "Instant", "{U}", "", nil)
	id := effectstest.CastCatalogSpell(t, g, "Talas Explorer", "Creature — Human Pirate Scout", oldB734Talas, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertKeywords(t, g, id, "flying")
	for i := range opp.Hand.Cards {
		if !opp.Hand.Cards[i].IsKnownTo(me.ID) {
			t.Error("the Explorer's controller did not see the opponent's hand")
		}
		if opp.Hand.Cards[i].IsKnownTo(other.ID) {
			t.Error("\"look at\" showed the hand to another player")
		}
	}
}

func oldB734CastHorror(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	id := effectstest.CastCatalogSpell(t, g, "Hidden Horror", "Creature — Horror", oldB734Horror, nil)
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestOldB734HiddenHorrorStaysWhenYouDiscardACreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: oldB734BearType, Owner: me.ID, Controller: me.ID})
	id := oldB734CastHorror(t, g, me)
	effectstest.AnswerDiscard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OnBattlefieldByID(g, id) {
		t.Error("the Horror was sacrificed although a creature card was discarded")
	}
	if !oldB734InGraveyard(me, bear) {
		t.Error("the creature card was not discarded")
	}
}

func TestOldB734HiddenHorrorIsSacrificedWithoutADiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	spell := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	id := oldB734CastHorror(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldByID(g, id) {
		t.Error("the Horror stayed with no creature card in hand")
	}
	if oldB734InGraveyard(me, spell) {
		t.Error("a noncreature card was discarded")
	}
}

func TestOldB734HiddenHorrorDeclinedDiscardSacrificesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: oldB734BearType, Owner: me.ID, Controller: me.ID})
	id := oldB734CastHorror(t, g, me)
	effectstest.AnswerDiscard(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldByID(g, id) {
		t.Error("the Horror stayed although no creature card was discarded")
	}
	if oldB734InGraveyard(me, bear) {
		t.Error("the creature card was discarded anyway")
	}
}
