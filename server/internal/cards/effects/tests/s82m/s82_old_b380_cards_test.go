package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b380_cards_test.go — Commander 2013, batch 380 (S82 older sets,
// #5560): eight cards of card draw and library. Every test plays the real
// catalog entry through the engine.

const (
	oldB380Spoils    = "852bd598-6e48-43c8-9211-740ae9e0c42e"
	oldB380Diviner   = "911b8849-dd0a-4383-8403-ea80227c5d7d"
	oldB380Scope     = "9633730d-c41f-4597-b806-55ce2dd848e9"
	oldB380Arrows    = "a3f10772-d055-419d-a535-4f094eaf437c"
	oldB380Altar     = "aa05900f-0f04-407e-931c-fea8f91e78e3"
	oldB380Command   = "d45a4924-daa0-4ac3-afd7-b66f636ce870"
	oldB380Baleful   = "f26fa9e4-f528-4e22-8f7c-329fc3b633e4"
	oldB380Foster    = "fb431500-152c-4524-b76b-de62922ff57f"
	oldB380Bear      = "Creature — Bear"
	oldB380BasicLand = "Basic Land — Forest"
)

var oldB380Cards = map[string]string{
	oldB380Spoils: "Spoils of Victory", oldB380Diviner: "Diviner Spirit", oldB380Scope: "Surveyor's Scope",
	oldB380Arrows: "Borrowing 100,000 Arrows", oldB380Altar: "Carnage Altar", oldB380Command: "Incendiary Command",
	oldB380Baleful: "Baleful Force", oldB380Foster: "Foster",
}

func oldB380Filler(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
	}
}

func oldB380Lands(g *game.Game, owner uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		effectstest.B12Permanent(g, owner, "Forest", oldB380BasicLand)
	}
}

func TestOldB380EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range oldB380Cards {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB380SpoilsOfVictoryFindsANonbasicDualButNotANonland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Library.Cards = nil
	dual := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Temple Garden", TypeLine: "Land — Forest Plains"})
	colorless := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Wastes Nonbasic", TypeLine: "Land"})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Bear", TypeLine: oldB380Bear})
	effectstest.CastCatalogSpell(t, g, "Spoils of Victory", "Sorcery", oldB380Spoils, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, dual)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dual) || effectstest.B16Tapped(t, g, dual) {
		t.Error("the dual land is not on the battlefield untapped")
	}
	if g.Battlefield.Contains(colorless) {
		t.Error("a land with no basic land type was found")
	}
}

func TestOldB380DivinerSpiritDrawsForBothPlayers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB380Filler(me, 4)
	oldB380Filler(opp, 4)
	spirit := effectstest.B12Push(g, me.ID, "Diviner Spirit", "Creature — Spirit", oldB380Diviner, 2, 4)
	mine, theirs := me.Hand.Size(), opp.Hand.Size()
	effectstest.AttackWith(t, g, opp.ID, spirit)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - mine; got != 2 {
		t.Errorf("I drew %d, want 2", got)
	}
	if got := opp.Hand.Size() - theirs; got != 2 {
		t.Errorf("the defender drew %d, want 2", got)
	}
}

func TestOldB380SurveyorsScopeFetchesOneLandPerPlayerWithTwoMoreLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Library.Cards = nil
	oldB380Lands(g, me.ID, 1)
	oldB380Lands(g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID, 3)
	oldB380Lands(g, g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)].ID, 2)
	a := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: oldB380BasicLand})
	b := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: oldB380BasicLand})
	scope := effectstest.PushCatalogPermanent(g, me.ID, "Surveyor's Scope", "Artifact", oldB380Scope, false)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, scope, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(scope) {
		t.Error("the Scope was not exiled as a cost")
	}
	if !g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Errorf("battlefield has A %v B %v, want exactly one land (one player has two more)", g.Battlefield.Contains(a), g.Battlefield.Contains(b))
	}
}

func TestOldB380SurveyorsScopeWithNoMoreLandsFindsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Library.Cards = nil
	oldB380Lands(g, me.ID, 2)
	a := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: oldB380BasicLand})
	scope := effectstest.PushCatalogPermanent(g, me.ID, "Surveyor's Scope", "Artifact", oldB380Scope, false)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, scope, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || !me.Library.Contains(a) {
		t.Error("a land was fetched with X = 0")
	}
}

func TestOldB380BorrowingArrowsDrawsForEachTappedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	oldB380Filler(me, 5)
	a := effectstest.B12Creature(g, opp.ID, "A", oldB380Bear, 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", oldB380Bear, 2, 2)
	effectstest.B12Creature(g, opp.ID, "C", oldB380Bear, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", oldB380Bear, 2, 2)
	g.WithWriteLock(func() {
		effectstest.BattlefieldCardPtr(t, g, a).Tapped = true
		effectstest.BattlefieldCardPtr(t, g, b).Tapped = true
		effectstest.BattlefieldCardPtr(t, g, mine).Tapped = true
	})
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Borrowing 100,000 Arrows", "Sorcery", oldB380Arrows, effectstest.PlayerRefs(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2 (two tapped creatures; mine does not count)", got)
	}
}

func TestOldB380CarnageAltarSacrificesACreatureToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oldB380Filler(me, 3)
	altar := effectstest.PushCatalogPermanent(g, me.ID, "Carnage Altar", "Artifact", oldB380Altar, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB380Bear, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, altar, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || me.Hand.Size() != hand+1 {
		t.Errorf("the Bear is on the battlefield %v, hand %d → %d, want gone and +1", g.Battlefield.Contains(bear), hand, me.Hand.Size())
	}
}

func TestOldB380IncendiaryCommandBurnsAndSweeps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	small := effectstest.B12Creature(g, opp.ID, "Small", oldB380Bear, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", oldB380Bear, 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big", oldB380Bear, 3, 3)
	life := effectstest.LifeOf(g, opp.ID)
	effectstest.CastCatalogSpellWithModesTargets(t, g, oldB380Command, []int{0, 1}, effectstest.PlayerRefs(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != life-4 {
		t.Errorf("opponent life %d → %d, want -4", life, got)
	}
	if g.Battlefield.Contains(small) || g.Battlefield.Contains(mine) {
		t.Error("a 2/2 survived 2 damage to each creature")
	}
	if !g.Battlefield.Contains(big) {
		t.Error("the 3/3 died to 2 damage")
	}
}

func TestOldB380IncendiaryCommandDestroysANonbasicLandAndWheelsEveryHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	oldB380Filler(me, 8)
	oldB380Filler(opp, 8)
	dual := effectstest.B12Permanent(g, opp.ID, "Temple Garden", "Land — Forest Plains")
	basic := effectstest.B12Permanent(g, opp.ID, "Forest", oldB380BasicLand)
	for i := 0; i < 2; i++ {
		opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "OppCard", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	}
	oppHand := opp.Hand.Size()
	var old []uuid.UUID
	for _, c := range opp.Hand.Cards {
		old = append(old, c.InstanceID)
	}
	effectstest.CastCatalogSpellWithModesTargets(t, g, oldB380Command, []int{2, 3}, effectstest.CardRefs(dual))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(dual) || !g.Battlefield.Contains(basic) {
		t.Errorf("nonbasic land on the battlefield %v, basic land %v, want false true", g.Battlefield.Contains(dual), g.Battlefield.Contains(basic))
	}
	if got := opp.Hand.Size(); got != oppHand {
		t.Errorf("the opponent's hand went %d → %d, want the same size after discarding and drawing that many", oppHand, got)
	}
	for _, id := range old {
		if opp.Hand.Contains(id) {
			t.Error("an old card is still in the opponent's hand")
		}
	}
	if len(opp.Graveyard.Cards) != oppHand+1 {
		t.Errorf("the opponent's graveyard has %d cards, want %d (the discards and the land)", len(opp.Graveyard.Cards), oppHand+1)
	}
}

func TestOldB380BalefulForceDrawsAndBleedsEveryUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oldB380Filler(me, 10)
	effectstest.B12Push(g, me.ID, "Baleful Force", "Creature — Elemental", oldB380Baleful, 7, 7)
	for _, seat := range []int{1, 0} {
		hand, life := me.Hand.Size(), effectstest.LifeOf(g, me.ID)
		effectstest.AdvanceToUpkeepOf(t, g, seat)
		effectstest.PassPriorityAroundTable(t, g)
		if me.Hand.Size() < hand+1 || effectstest.LifeOf(g, me.ID) > life-1 {
			t.Errorf("seat %d's upkeep: hand %d → %d, life %d → %d, want a card and -1", seat, hand, me.Hand.Size(), life, effectstest.LifeOf(g, me.ID))
		}
	}
}

func TestOldB380FosterRevealsUntilACreatureForOneMana(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		me.Library.Cards = nil
		// PushLibraryCardForTest pushes to the bottom: the first card pushed is the top.
		a := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Spell A", TypeLine: "Sorcery"})
		b := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Spell B", TypeLine: "Sorcery"})
		creature := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Library Bear", TypeLine: oldB380Bear, Power: 2, Toughness: 2})
		effectstest.PushCatalogPermanent(g, me.ID, "Foster", "Enchantment", oldB380Foster, false)
		victim := effectstest.B12Creature(g, me.ID, "Victim", oldB380Bear, 1, 1)
		effectstest.HobDestroy(t, g, victim)
		effectstest.PassPriorityAroundTable(t, g)
		if pay {
			effectstest.B06AddMana(me, "C")
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		inGrave := func(id uuid.UUID) bool { return me.Graveyard.Contains(id) }
		if pay {
			if !me.Hand.Contains(creature) || !inGrave(a) || !inGrave(b) {
				t.Errorf("paid: creature in hand %v, spells milled %v %v, want true true true", me.Hand.Contains(creature), inGrave(a), inGrave(b))
			}
		} else if me.Hand.Contains(creature) || inGrave(a) {
			t.Error("declined: the library was touched")
		}
	}
}
