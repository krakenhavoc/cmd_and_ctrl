package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 old sets, batch 006, group C (#5462): Ashiok, Dream Render; Jaya
// Ballard; Ajani, Mentor of Heroes.

const (
	s82OldB6AshiokOracle     = "93723b12-db34-4047-885e-8606415b1553"
	s82OldB6JayaOracle       = "a56c5ca5-70eb-4d5f-8116-2acbe5f5a3cb"
	s82OldB6AjaniOracle      = "a612bd38-069f-4bdc-9aa6-c172435e3841"
	s82OldB6SylvanTutor      = "935e0cac-51ee-4cb7-a209-f085e0f099ed"
	s82OldB6DivinationOracle = "273b339c-964b-4a18-8eb5-ceb8abcdfd9e"
)

func TestS82OldB006CCardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		s82OldB6AshiokOracle: "Ashiok, Dream Render",
		s82OldB6JayaOracle:   "Jaya Ballard",
		s82OldB6AjaniOracle:  "Ajani, Mentor of Heroes",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q completeness %v caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Ashiok, Dream Render --------------------------------------------------

func TestS82OldB006CAshiokStopsAnOpponentsSearch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.S82Walker(g, opp.ID, "Ashiok, Dream Render", s82OldB6AshiokOracle, 5)
	bear := effectstest.SeedSearchLibrary(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})[0]
	effectstest.CastCatalogSpell(t, g, "Sylvan Tutor", "Sorcery", s82OldB6SylvanTutor, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Fatal("an opponent's Ashiok left a search prompt open for the caster")
	}
	if me.Hand.Contains(bear) || !me.Library.Contains(bear) {
		t.Error("the card left the library although the search cannot happen")
	}
}

func TestS82OldB006CAshiokDoesNotStopItsControllersOwnSearch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.S82Walker(g, me.ID, "Ashiok, Dream Render", s82OldB6AshiokOracle, 5)
	bear := effectstest.SeedSearchLibrary(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})[0]
	effectstest.CastCatalogSpell(t, g, "Sylvan Tutor", "Sorcery", s82OldB6SylvanTutor, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Library.Contains(bear) {
		t.Fatal("Sylvan Tutor leaves the creature on top of the library")
	}
}

func TestS82OldB006CASearchWithoutAshiokStillHappens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.SeedSearchLibrary(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})[0]
	effectstest.CastCatalogSpell(t, g, "Sylvan Tutor", "Sorcery", s82OldB6SylvanTutor, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, me.ID) == nil {
		t.Fatal("no search prompt although no Ashiok is in play")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, bear)
}

func TestS82OldB006CAshiokMillsTheTargetThenExilesEachOpponentsGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	other := g.Seats[(g.Turn.ActiveSeat+2)%4]
	ashiok := effectstest.S82Walker(g, me.ID, "Ashiok, Dream Render", s82OldB6AshiokOracle, 5)
	effectstest.SeedLibrary(opp, "A", "B", "C", "D", "E", "F")
	libBefore := opp.Library.Size()
	oldOpp := effectstest.PtGraveCard(opp, "Old Opp Card", "Sorcery")
	oldOther := effectstest.PtGraveCard(other, "Old Other Card", "Sorcery")
	mine := effectstest.PtGraveCard(me, "My Card", "Sorcery")
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, ashiok, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoyaltyOf(g, ashiok); got != 4 {
		t.Errorf("loyalty %d, want 4", got)
	}
	if opp.Library.Size() != libBefore-4 {
		t.Errorf("the target's library has %d cards, want %d (four milled)", opp.Library.Size(), libBefore-4)
	}
	if opp.Graveyard.Size() != 0 || other.Graveyard.Contains(oldOther) || opp.Graveyard.Contains(oldOpp) {
		t.Error("an opponent's graveyard survived, milled cards included")
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("my own graveyard was exiled")
	}
	if !g.Exile.Contains(oldOpp) || !g.Exile.Contains(oldOther) {
		t.Error("the graveyard cards are not in exile")
	}
}

func TestS82OldB006CAshiokMillTargetingMyselfLeavesMyGraveyardMilled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ashiok := effectstest.S82Walker(g, me.ID, "Ashiok, Dream Render", s82OldB6AshiokOracle, 5)
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, ashiok, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Size() != 4 {
		t.Errorf("my graveyard has %d cards, want the 4 milled (it is not an opponent's)", me.Graveyard.Size())
	}
}

// --- Jaya Ballard -----------------------------------------------------------

func TestS82OldB006CJayaAddsRestrictedRedMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	jaya := effectstest.S82Walker(g, me.ID, "Jaya Ballard", s82OldB6JayaOracle, 5)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, jaya, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoyaltyOf(g, jaya); got != 6 {
		t.Errorf("loyalty %d, want 6", got)
	}
	if me.ManaPool.Total() != 3 {
		t.Fatalf("pool has %d mana, want 3", me.ManaPool.Total())
	}
	for _, tok := range me.ManaPool {
		if tok.Color != "R" {
			t.Errorf("mana %q, want R", tok.Color)
		}
		if len(tok.Restrictions) < 2 {
			t.Errorf("restrictions %v, want cast + instant-or-sorcery", tok.Restrictions)
		}
	}
}

func TestS82OldB006CJayaLootsUpToThreeThenDrawsThatMany(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	jaya := effectstest.S82Walker(g, me.ID, "Jaya Ballard", s82OldB6JayaOracle, 5)
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")
	me.Hand.Cards = nil
	h := make([]uuid.UUID, 3)
	for i := range h {
		h[i] = uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: h[i], Name: "Card", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, jaya, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID, h[0], h[1])
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != 3 {
		t.Errorf("hand has %d cards, want 3 (discarded two, drew two)", me.Hand.Size())
	}
	if me.Graveyard.Size() != 2 {
		t.Errorf("graveyard has %d cards, want 2", me.Graveyard.Size())
	}
}

func TestS82OldB006CJayaEmblemCastsSpellsFromTheGraveyardAndExilesThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	jaya := effectstest.S82Walker(g, me.ID, "Jaya Ballard", s82OldB6JayaOracle, 8)
	effectstest.SeedCheapLibrary(g, me, 6)
	div := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: div, Name: "Divination", TypeLine: "Sorcery", ManaCost: "{0}",
		OracleID: s82OldB6DivinationOracle, Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
	bear := effectstest.PtGraveCard(me, "Bear", "Creature — Bear")
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, div, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Fatal("cast from the graveyard before the emblem")
	}
	effectstest.S82Activate(t, g, me.ID, jaya, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if len(effectstest.EmblemsOf(g, me.ID)) != 1 {
		t.Fatal("no emblem")
	}
	if got := effectstest.LoyaltyOf(g, jaya); got != 0 && got != -1 {
		t.Errorf("loyalty %d, want 0 after -8", got)
	}
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Error("a creature spell was cast from the graveyard")
	}
	hand := me.Hand.Size()
	if err := g.CastSpell(me.ID, div, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "jaya_ballard_emblem"}); err != nil {
		t.Fatalf("cast Divination from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(div) || me.Graveyard.Contains(div) {
		t.Error("the spell cast this way is not exiled instead of going to the graveyard")
	}
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d, want %d (Divination resolved)", me.Hand.Size(), hand+2)
	}
}

// --- Ajani, Mentor of Heroes -----------------------------------------------

func TestS82OldB006CAjaniDistributesThreeCountersAmongTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	ajani := effectstest.S82Walker(g, me.ID, "Ajani, Mentor of Heroes", s82OldB6AjaniOracle, 4)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, ajani, 0, game.ActivateAbilityParams{Strict: true,
		Targets:      effectstest.S82TargetCard(theirs),
		Distribution: map[uuid.UUID]int{theirs: 3}}); err == nil {
		t.Fatal("an opponent's creature was targeted")
	}
	effectstest.S82Activate(t, g, me.ID, ajani, 0, game.ActivateAbilityParams{
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}},
		Distribution: map[uuid.UUID]int{a: 2, b: 1}})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, a, "+1/+1"); n != 2 {
		t.Errorf("A has %d counters, want 2", n)
	}
	if n := effectstest.CounterCount(g, b, "+1/+1"); n != 1 {
		t.Errorf("B has %d counters, want 1", n)
	}
	if got := effectstest.LoyaltyOf(g, ajani); got != 5 {
		t.Errorf("loyalty %d, want 5", got)
	}
}

func TestS82OldB006CAjaniLooksAtFourAndTakesACreatureAuraOrPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ajani := effectstest.S82Walker(g, me.ID, "Ajani, Mentor of Heroes", s82OldB6AjaniOracle, 4)
	me.Library.Cards = nil
	mk := func(name, tl string) uuid.UUID {
		id := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: tl, Owner: me.ID, Controller: me.ID})
		return id
	}
	deep := mk("Deep", "Creature — Bear") // fifth card: not looked at
	forest := mk("Forest", "Basic Land — Forest")
	aura := mk("Pacifism", "Enchantment — Aura")
	wolf := mk("Wolf", "Creature — Wolf")
	shock := mk("Shock", "Instant")
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, ajani, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no card-set prompt")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.ChooseCards {
		offered[id] = true
	}
	if !offered[aura] || !offered[wolf] || offered[forest] || offered[shock] || offered[deep] {
		t.Errorf("offered %v: want the Aura and the creature only", offered)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, wolf)
	if !me.Hand.Contains(wolf) {
		t.Error("the chosen creature is not in hand")
	}
	ord := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if ord == nil || len(ord.ScryCards) != 3 {
		t.Fatalf("the other three are ordered onto the bottom: %+v", ord)
	}
	if err := g.ResolvePutInLibrary(ord.ID, me.ID, ord.ScryCards, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if top, _ := me.Library.Top(); top.InstanceID != deep {
		t.Error("the unseen fifth card is no longer on top of the library")
	}
	bottom := map[uuid.UUID]bool{}
	for _, c := range me.Library.Cards[:3] {
		bottom[c.InstanceID] = true
	}
	if !bottom[forest] || !bottom[aura] || !bottom[shock] {
		t.Error("the cards that were not taken are not the bottom three")
	}
}

func TestS82OldB006CAjaniUltimateGainsAHundredLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ajani := effectstest.S82Walker(g, me.ID, "Ajani, Mentor of Heroes", s82OldB6AjaniOracle, 8)
	life := me.Life
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, ajani, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+100 {
		t.Errorf("life %d, want %d", me.Life, life+100)
	}
}
