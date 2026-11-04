package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b29_moseo_test.go — Moseo, Vein's New Dean (batch 29, #4456).

const socB29MoseoOracle = "c2e2f21a-0338-44d7-96a5-5a277f0ec419"

func TestSocB29MoseoIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(socB29MoseoOracle)
	if !ok {
		t.Fatal("Moseo is not registered")
	}
	if spec.Name != "Moseo, Vein's New Dean" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
	}
}

func socB29GraveyardCreature(me *game.Player, name, cost string) uuid.UUID {
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: socB29BearType, ManaCost: cost,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	return id
}

func socB29GainLife(g *game.Game, p *game.Player, n int) {
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, p.ID, n) })
}

func TestSocB29MoseoFliesAndMakesAPestThatGainsLifeWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	moseo := castCatalogSpell(t, g, "Moseo, Vein's New Dean", "Legendary Creature — Bird Skeleton Warlock", socB29MoseoOracle, nil)
	passPriorityAroundTable(t, g)
	if !socB29HasKeyword(t, g, moseo, "flying") {
		t.Error("Moseo has flying")
	}
	pestID := findBattlefieldByName(g, "Pest")
	if pestID == uuid.Nil {
		t.Fatal("no Pest was created")
	}
	pest := s58p6View(t, g, pestID)
	if !pest.IsToken() || pest.CurrentPower() != 1 || pest.CurrentToughness() != 1 || len(pest.Colors) != 2 {
		t.Fatalf("the Pest is token %v %d/%d colors %v, want a 1/1 black and green token", pest.IsToken(), pest.CurrentPower(), pest.CurrentToughness(), pest.Colors)
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == pestID {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	life := me.Life
	declareAttack(t, g, opp.ID, pestID)
	passPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d -> %d when the Pest attacked, want +1", life, me.Life)
	}
}

func TestSocB29MoseoInfusionReturnsACreatureWithManaValueAtMostTheLifeGained(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Moseo, Vein's New Dean", "Legendary Creature — Bird Skeleton Warlock", socB29MoseoOracle, 2, 1)
	cheap := socB29GraveyardCreature(me, "Cheap Bear", "{1}{G}")
	three := socB29GraveyardCreature(me, "Three Bear", "{2}{G}")
	dear := socB29GraveyardCreature(me, "Dear Bear", "{3}{G}")
	advanceToMain(t, g)
	socB29GainLife(g, me, 3)
	advanceToEndStepOf(t, g, 0)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: dear}); err == nil {
		t.Fatal("a mana value 4 creature was accepted after gaining 3 life")
	}
	pickCard(t, g, me.ID, three)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(three) {
		t.Error("the chosen creature is not on the battlefield")
	}
	if !me.Graveyard.Contains(cheap) || !me.Graveyard.Contains(dear) {
		t.Error("an unchosen creature left the graveyard")
	}
}

func TestSocB29MoseoInfusionDoesNothingWithoutLifeGained(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Moseo, Vein's New Dean", "Legendary Creature — Bird Skeleton Warlock", socB29MoseoOracle, 2, 1)
	bear := socB29GraveyardCreature(me, "Land Bear", "")
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("the Infusion trigger asked for a target with no life gained")
	}
	if g.Battlefield.Contains(bear) {
		t.Error("a creature returned with no life gained this turn")
	}
}

func TestSocB29MoseoInfusionMayChooseNoTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Moseo, Vein's New Dean", "Legendary Creature — Bird Skeleton Warlock", socB29MoseoOracle, 2, 1)
	bear := socB29GraveyardCreature(me, "Cheap Bear", "{1}")
	advanceToMain(t, g)
	socB29GainLife(g, me, 1)
	advanceToEndStepOf(t, g, 0)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
		t.Fatalf("declining the \"up to one\": %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !me.Graveyard.Contains(bear) {
		t.Error("a creature returned although none was chosen")
	}
}

func TestSocB29MoseoInfusionIgnoresLifeAnOpponentGainedAndOtherTurns(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Moseo, Vein's New Dean", "Legendary Creature — Bird Skeleton Warlock", socB29MoseoOracle, 2, 1)
	bear := socB29GraveyardCreature(me, "Cheap Bear", "{1}")
	advanceToMain(t, g)
	socB29GainLife(g, opp, 5)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || latestPickTarget(g, me.ID) != nil {
		t.Error("an opponent's life gain triggered my Infusion")
	}
}
