package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 BNS batch 7 — Special Guests (#5929).

const (
	s82BnsAkromasMemorial = "967e8b3f-8266-4d9a-a38b-d0e10184e1e4"
	s82BnsCodie           = "67a9357c-4713-4e01-a60d-532bf0dd80b6"
	s82BnsKalamax         = "b8501a37-23e4-4873-80cb-d1c0f6c95155"
	s82BnsEyeOfUgin       = "10d13ff6-c4d0-4753-8939-a8a90f0e92bb"
	s82BnsWindgrace       = "38a55562-1e2d-4240-9454-219a4a25d38d"
	s82BnsFootfalls       = "2b166d97-c2c1-4642-b151-9bb1ad32e362"
	s82BnsHeatShimmer     = "5c80d39e-f3af-4d24-bca0-19f0bd0f5082"
	s82BnsUltimatum       = "76ca315a-4979-4fda-8d3f-b7f4e5d3a622"
	s82BnsMindTwist       = "78f9c223-9982-4282-a496-a6f892f0a5bf"
	s82BnsSylvanTutor     = "935e0cac-51ee-4cb7-a209-f085e0f099ed"
)

func TestS82BnsB07CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		s82BnsAkromasMemorial: "Akroma's Memorial", s82BnsCodie: "Codie, Vociferous Codex",
		s82BnsKalamax: "Kalamax, the Stormsire", s82BnsEyeOfUgin: "Eye of Ugin",
		s82BnsWindgrace: "Lord Windgrace", s82BnsFootfalls: "Crashing Footfalls",
		s82BnsHeatShimmer: "Heat Shimmer", s82BnsUltimatum: "Inspired Ultimatum",
		s82BnsMindTwist: "Mind Twist", s82BnsSylvanTutor: "Sylvan Tutor",
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

// --- Sylvan Tutor -----------------------------------------------------

func TestS82BnsSylvanTutorPutsACreatureOnTopAndRevealsIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushLibraryCardForTest(me, game.Card{Name: "Bolt", TypeLine: "Instant"})
	needle := pushLibraryCardForTest(me, game.Card{Name: "Craterhoof Behemoth", TypeLine: "Creature — Beast"})
	top := tutorTopOfLibrary(t, g, me, "Sylvan Tutor", "Sorcery", s82BnsSylvanTutor, needle)
	if top.InstanceID != needle {
		t.Errorf("top of library is %q, want the tutored creature", top.Name)
	}
	if me.Hand.Contains(needle) {
		t.Error("the creature went to hand; Sylvan Tutor puts it on top")
	}
	if !top.IsKnownTo(opp.ID) {
		t.Error("an opponent does not know the revealed card")
	}
}

// --- Mind Twist -------------------------------------------------------

func s82BnsCastMindTwist(t *testing.T, g *game.Game, target uuid.UUID, x int) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Mind Twist", TypeLine: "Sorcery", OracleID: s82BnsMindTwist,
		Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: x,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: target}}}); err != nil {
		t.Fatalf("cast Mind Twist: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func TestS82BnsMindTwistDiscardsXCardsAtRandom(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	for _, n := range []string{"A", "B", "C", "D", "E"} {
		opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: n, TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	}
	hand, yard := opp.Hand.Size(), opp.Graveyard.Size()
	s82BnsCastMindTwist(t, g, opp.ID, 3)
	if got := opp.Hand.Size(); got != hand-3 {
		t.Errorf("hand %d, want %d", got, hand-3)
	}
	if got := opp.Graveyard.Size(); got != yard+3 {
		t.Errorf("graveyard %d, want %d", got, yard+3)
	}
}

func TestS82BnsMindTwistEmptiesASmallerHandAndXZeroDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	hand := opp.Hand.Size()
	s82BnsCastMindTwist(t, g, opp.ID, 0)
	if got := opp.Hand.Size(); got != hand {
		t.Errorf("X=0: hand %d, want %d", got, hand)
	}
	s82BnsCastMindTwist(t, g, opp.ID, hand+4)
	if got := opp.Hand.Size(); got != 0 {
		t.Errorf("X larger than the hand: %d cards left, want 0", got)
	}
}

// --- Inspired Ultimatum -----------------------------------------------

func TestS82BnsInspiredUltimatumGainsDamagesAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 6; i++ {
		pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Creature — Bear"})
	}
	hand, mine, theirs := me.Hand.Size(), me.Life, opp.Life
	// The life goes to the targeted player (the caster's opponent here), the
	// damage to a different target, and the caster draws.
	castCatalogSpell(t, g, "Inspired Ultimatum", "Sorcery", s82BnsUltimatum, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: opp.ID, Slot: 0},
		{Kind: game.TargetPlayer, ID: me.ID, Slot: 1},
	})
	passPriorityAroundTable(t, g)
	if opp.Life != theirs+5 {
		t.Errorf("the targeted player's life %d, want %d", opp.Life, theirs+5)
	}
	if me.Life != mine-5 {
		t.Errorf("the damage target's life %d, want %d", me.Life, mine-5)
	}
	if got := me.Hand.Size(); got != hand+5 {
		t.Errorf("caster's hand %d, want %d", got, hand+5)
	}
}

// --- Heat Shimmer -----------------------------------------------------

func TestS82BnsHeatShimmerMakesAHastyCopyThatIsExiledAtEndStep(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	// "Target creature": an opponent's creature is a legal target too.
	bear := pushVanillaCreature(g, opp.ID, "Big Bear", 4, 4)
	castCatalogSpell(t, g, "Heat Shimmer", "Sorcery", s82BnsHeatShimmer, cardTargets(bear))
	passPriorityAroundTable(t, g)
	tokens, _ := copiesOnBattlefield(g, "Big Bear")
	if len(tokens) != 1 {
		t.Fatalf("%d token copies, want 1", len(tokens))
	}
	tok := tokens[0]
	if tok.Controller != g.Seats[0].ID {
		t.Errorf("the token is controlled by %v, want the caster", tok.Controller)
	}
	if tok.Power != 4 || tok.Toughness != 4 || !hasKeywordForTest(tok.Keywords, "haste") {
		t.Errorf("token %d/%d keywords %v, want a hasty 4/4", tok.Power, tok.Toughness, tok.Keywords)
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if onBattlefieldByID(g, tok.InstanceID) {
		t.Error("the token is still on the battlefield after the end step")
	}
	if g.Seats[0].Graveyard.Contains(tok.InstanceID) {
		t.Error("the token was sacrificed, not exiled")
	}
	if !onBattlefieldByID(g, bear) {
		t.Error("the original creature left")
	}
}

// --- Crashing Footfalls -----------------------------------------------

func TestS82BnsCrashingFootfallsSuspendsAndMakesTwoRhinos(t *testing.T) {
	spec, _ := Lookup(s82BnsFootfalls)
	if len(spec.SpecialActions) != 1 {
		t.Fatalf("%d special actions, want suspend", len(spec.SpecialActions))
	}
	if sa := spec.SpecialActions[0]; sa.Kind != game.SpecialActionSuspend || sa.Counters != 4 || sa.Cost != "{G}" {
		t.Errorf("special action %+v, want suspend 4—{G}", sa)
	}
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	// No mana cost: it can never be cast from hand (CR 118.6).
	id := handCardForTest(me, "Crashing Footfalls", "Sorcery", s82BnsFootfalls)
	// CR 118.6 reads the Scryfall-stamped Layout: the fixture has to look imported.
	stampLayout(g, me, id, "normal")
	floatForTest(g, me, "GGGGGGGG")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("cast Crashing Footfalls from hand")
	}
}

func TestS82BnsCrashingFootfallsResolvesIntoTwoTramplingRhinos(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Crashing Footfalls", "Sorcery", s82BnsFootfalls, nil)
	passPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Rhino" || c.Controller != me.ID {
			continue
		}
		n++
		if !c.IsToken() || c.Power != 4 || c.Toughness != 4 || len(c.Colors) != 1 || c.Colors[0] != "G" {
			t.Errorf("Rhino %d/%d %v, want a 4/4 green token", c.Power, c.Toughness, c.Colors)
		}
		assertKeywords(t, g, c.InstanceID, "trample")
	}
	if n != 2 {
		t.Errorf("%d Rhinos, want 2", n)
	}
}

// --- Akroma's Memorial ------------------------------------------------

func TestS82BnsAkromasMemorialGivesYourCreaturesEverythingAndNoOthers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 2, 2)
	b12Push(g, me.ID, "Akroma's Memorial", "Legendary Artifact", s82BnsAkromasMemorial, 0, 0)
	all := []string{"flying", "first strike", "vigilance", "trample", "haste", "protection from black", "protection from red"}
	for _, kw := range all {
		if !hasKeywordOnBattlefield(t, g, mine, kw) {
			t.Errorf("your creature lacks %q", kw)
		}
		if hasKeywordOnBattlefield(t, g, theirs, kw) {
			t.Errorf("an opponent's creature has %q", kw)
		}
	}
	// A creature that arrives later has it at once.
	later := b12Creature(g, me.ID, "Late Bear", "Creature — Bear", 1, 1)
	if !hasKeywordOnBattlefield(t, g, later, "protection from red") {
		t.Error("a creature that arrived later lacks protection from red")
	}
}

// --- Eye of Ugin ------------------------------------------------------

func s82BnsEldraziCost(t *testing.T, g *game.Game, typeLine, cost string, colors []string, pool string) error {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	floatForTest(g, me, pool)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Test Eldrazi", TypeLine: typeLine, ManaCost: cost,
		Colors: colors, Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
}

func TestS82BnsEyeOfUginMakesColorlessEldraziCostTwoLess(t *testing.T) {
	g := newCatalogGame(t)
	pushCatalogPermanent(g, g.Seats[0].ID, "Eye of Ugin", "Legendary Land", s82BnsEyeOfUgin, false)
	if err := s82BnsEldraziCost(t, g, "Creature — Eldrazi", "{4}", nil, "CC"); err != nil {
		t.Errorf("a colorless Eldrazi costing {4} should cost {2}: %v", err)
	}
}

func TestS82BnsEyeOfUginDoesNotDiscountOthers(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, cost string
		colors               []string
	}{
		{"a colorless non-Eldrazi", "Artifact Creature — Golem", "{4}", nil},
		{"a coloured Eldrazi", "Creature — Eldrazi", "{3}{G}", []string{"G"}},
	} {
		g := newCatalogGame(t)
		pushCatalogPermanent(g, g.Seats[0].ID, "Eye of Ugin", "Legendary Land", s82BnsEyeOfUgin, false)
		pool := "CC"
		if len(tc.colors) > 0 {
			pool = "CCG"
		}
		if err := s82BnsEldraziCost(t, g, tc.typeLine, tc.cost, tc.colors, pool); err == nil {
			t.Errorf("%s was discounted by Eye of Ugin", tc.name)
		}
	}
}

func TestS82BnsEyeOfUginDiscountOnlyForItsControllersSpells(t *testing.T) {
	g := newCatalogGame(t)
	pushCatalogPermanent(g, g.Seats[1].ID, "Eye of Ugin", "Legendary Land", s82BnsEyeOfUgin, false)
	if err := s82BnsEldraziCost(t, g, "Creature — Eldrazi", "{4}", nil, "CC"); err == nil {
		t.Error("an opponent's Eye of Ugin discounted my Eldrazi")
	}
}

func TestS82BnsEyeOfUginSearchesForAColorlessCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	eye := pushCatalogPermanent(g, me.ID, "Eye of Ugin", "Legendary Land", s82BnsEyeOfUgin, false)
	pushLibraryCardForTest(me, game.Card{Name: "Green Bear", TypeLine: "Creature — Bear", Colors: []string{"G"}, ManaCost: "{1}{G}"})
	pushLibraryCardForTest(me, game.Card{Name: "Sol Ring", TypeLine: "Artifact", ManaCost: "{1}"})
	needle := pushLibraryCardForTest(me, game.Card{Name: "Walking Ballista", TypeLine: "Artifact Creature — Construct", ManaCost: "{X}{X}"})
	advanceToMain(t, g)
	floatForTest(g, me, "CCCCCCC")
	if err := g.ActivateCatalogAbility(me.ID, eye, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if searchChoiceFor(g, me.ID) != nil {
		answerSearchByID(t, g, me.ID, needle)
	}
	if !me.Hand.Contains(needle) {
		t.Error("the colorless creature is not in hand")
	}
	if c, _ := g.LookupCardForEffect(eye); !c.Tapped {
		t.Error("Eye of Ugin was not tapped for its ability")
	}
	// {7} is needed.
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if err := g.ActivateCatalogAbility(me.ID, eye, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with no mana and a tapped land")
	}
}

// --- Kalamax, the Stormsire -------------------------------------------

func s82BnsKalamaxOnBattlefield(g *game.Game, owner uuid.UUID, tapped bool) uuid.UUID {
	id := apaPush(g, owner, owner, game.Card{Name: "Kalamax, the Stormsire", OracleID: s82BnsKalamax,
		TypeLine: "Legendary Creature — Elemental Dinosaur", Power: 4, Toughness: 4})
	if tapped {
		g.WithWriteLock(func() { findBattlefieldCardForTest(g, id).Tapped = true })
	}
	return id
}

func s82BnsCastInstant(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	floatForTest(g, me, "U")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Opt Test", TypeLine: "Instant", ManaCost: "{U}", Colors: []string{"U"},
		Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast instant: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func TestS82BnsKalamaxCopiesTheFirstInstantEachTurnWhileTappedAndGrows(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	k := s82BnsKalamaxOnBattlefield(g, me.ID, true)
	advanceToMain(t, g)
	s82BnsCastInstant(t, g, me)
	if got := plusOneCounters(g, k); got != 1 {
		t.Fatalf("after the first instant: %d counters, want 1 (it copied it)", got)
	}
	// Only the first instant each turn.
	s82BnsCastInstant(t, g, me)
	if got := plusOneCounters(g, k); got != 1 {
		t.Errorf("after the second instant: %d counters, want still 1", got)
	}
}

func TestS82BnsKalamaxDoesNothingWhileUntapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	k := s82BnsKalamaxOnBattlefield(g, me.ID, false)
	advanceToMain(t, g)
	s82BnsCastInstant(t, g, me)
	if got := plusOneCounters(g, k); got != 0 {
		t.Errorf("untapped Kalamax has %d counters, want 0", got)
	}
}

func TestS82BnsKalamaxIgnoresSorceries(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	k := s82BnsKalamaxOnBattlefield(g, me.ID, true)
	advanceToMain(t, g)
	floatForTest(g, me, "U")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Sorcery Test", TypeLine: "Sorcery", ManaCost: "{U}", Colors: []string{"U"},
		Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast sorcery: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := plusOneCounters(g, k); got != 0 {
		t.Errorf("a sorcery was copied: %d counters", got)
	}
}

// --- Codie, Vociferous Codex ------------------------------------------

func TestS82BnsCodieCantCastPermanentSpellsOnlyForItsController(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Codie, Vociferous Codex", "Legendary Artifact Creature — Book Construct", s82BnsCodie, false)
	advanceToMain(t, g)
	for _, tc := range []struct{ typeLine string }{{"Creature — Bear"}, {"Artifact"}, {"Enchantment"}} {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Permanent", TypeLine: tc.typeLine, Owner: me.ID, Controller: me.ID})
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
			t.Errorf("cast a %s spell under Codie", tc.typeLine)
		}
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Sorcery", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("a sorcery should still be castable: %v", err)
	}
}

func TestS82BnsCodieManaThenNextSpellExilesToALesserInstantOrSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	codie := pushCatalogPermanent(g, me.ID, "Codie, Vociferous Codex", "Legendary Artifact Creature — Book Construct", s82BnsCodie, false)
	// Library, TOP first (pushLibraryCardForTest pushes to the bottom); the walk
	// reads from the top: Bear, Big Sorcery, then the hit.
	bear := pushLibraryCardForTest(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"})
	bigSorcery := pushLibraryCardForTest(me, game.Card{Name: "Big Sorcery", TypeLine: "Sorcery", ManaCost: "{4}"})
	hit := pushLibraryCardForTest(me, game.Card{Name: "Cheap Instant", TypeLine: "Instant", ManaCost: "{1}"})
	deep := pushLibraryCardForTest(me, game.Card{Name: "Deep Filler", TypeLine: "Creature — Bear", ManaCost: "{1}"})
	advanceToMain(t, g)
	floatForTest(g, me, "CCCC")
	activateManaFor(t, g, me.ID, codie, 0, game.ManaAbilityParams{})
	if got := me.ManaPool.Total(); got != 5 {
		t.Fatalf("pool has %d mana, want {W}{U}{B}{R}{G}", got)
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Spell Three", TypeLine: "Sorcery", ManaCost: "{3}",
		Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !inExile(g, hit) {
		t.Fatal("the lesser instant was not exiled")
	}
	// A Bear (not an instant or sorcery) and a sorcery of equal-or-greater value
	// were passed over: they are back in the library, not in exile.
	for _, pass := range []uuid.UUID{bear, bigSorcery} {
		if inExile(g, pass) || !me.Library.Contains(pass) {
			t.Errorf("a passed-over card is not back in the library")
		}
	}
	if !me.Library.Contains(deep) {
		t.Error("the card below the hit was disturbed")
	}
	// The passed-over cards are on the BOTTOM: the untouched card is still on top.
	if top, _ := me.Library.Top(); top.InstanceID != deep {
		t.Error("a passed-over card was not put on the bottom of the library")
	}
	// "Until end of turn, you may cast that card without paying its mana cost."
	if err := g.CastSpell(me.ID, hit, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("free cast of the exiled instant: %v", err)
	}
	// The ability was one-shot: another spell this turn exiles nothing more.
	passPriorityAroundTable(t, g)
	before := me.Library.Size()
	floatForTest(g, me, "CC")
	id2 := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id2, Name: "Spell Two", TypeLine: "Sorcery", ManaCost: "{2}",
		Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id2, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("second cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Library.Size(); got != before {
		t.Errorf("library %d, want %d: the delayed trigger fired twice", got, before)
	}
}

func TestS82BnsCodieWithNoLesserHitBottomsEverythingItExiled(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	codie := pushCatalogPermanent(g, me.ID, "Codie, Vociferous Codex", "Legendary Artifact Creature — Book Construct", s82BnsCodie, false)
	var ids []uuid.UUID
	for _, n := range []string{"One", "Two", "Three"} {
		ids = append(ids, pushLibraryCardForTest(me, game.Card{Name: n, TypeLine: "Instant", ManaCost: "{5}"}))
	}
	advanceToMain(t, g)
	floatForTest(g, me, "CCCC")
	activateManaFor(t, g, me.ID, codie, 0, game.ManaAbilityParams{})
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Spell Three", TypeLine: "Sorcery", ManaCost: "{3}",
		Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, c := range ids {
		if inExile(g, c) || !me.Library.Contains(c) {
			t.Errorf("a card exiled by a fruitless walk did not return to the library")
		}
	}
}

// --- Lord Windgrace ---------------------------------------------------

func s82BnsWindgraceOnBattlefield(g *game.Game, owner uuid.UUID, loyalty int) uuid.UUID {
	return pushCatalogWalkerWithLine(g, owner, "Lord Windgrace", "Legendary Planeswalker — Windgrace", s82BnsWindgrace, loyalty)
}

func TestS82BnsWindgracePlusTwoLootsAndDrawsExtraForALand(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		drawn          int
	}{
		{"a nonland", "Creature — Bear", 1},
		{"a land", "Land", 2},
	} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		w := s82BnsWindgraceOnBattlefield(g, me.ID, 5)
		for i := 0; i < 4; i++ {
			pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Creature — Bear"})
		}
		pitch := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: pitch, Name: "Pitch", TypeLine: tc.typeLine, Owner: me.ID, Controller: me.ID})
		advanceToMain(t, g)
		hand := me.Hand.Size()
		if err := g.ActivateCatalogAbility(me.ID, w, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("%s: +2: %v", tc.name, err)
		}
		passPriorityAroundTable(t, g)
		answerDiscard(t, g, me.ID, pitch)
		passPriorityAroundTable(t, g)
		if !me.Graveyard.Contains(pitch) {
			t.Errorf("%s: the card was not discarded", tc.name)
		}
		if got := me.Hand.Size(); got != hand-1+tc.drawn {
			t.Errorf("%s: hand %d, want %d", tc.name, got, hand-1+tc.drawn)
		}
		if got := loyaltyCount(g, w); got != 7 {
			t.Errorf("%s: loyalty %d, want 7", tc.name, got)
		}
	}
}

func TestS82BnsWindgraceMinusThreeReturnsUpToTwoLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	w := s82BnsWindgraceOnBattlefield(g, me.ID, 5)
	a, b, c := uuid.New(), uuid.New(), uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: a, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: b, Name: "Swamp", TypeLine: "Basic Land — Swamp", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: c, Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, w, 1, game.ActivateAbilityParams{Strict: true, Targets: cardTargets(a, c)}); err == nil {
		t.Error("targeted a nonland card")
	}
	if err := g.ActivateCatalogAbility(me.ID, w, 1, game.ActivateAbilityParams{Strict: true, Targets: cardTargets(a, b)}); err != nil {
		t.Fatalf("-3: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !onBattlefieldByID(g, a) || !onBattlefieldByID(g, b) {
		t.Error("the lands did not return to the battlefield")
	}
	if got := loyaltyCount(g, w); got != 2 {
		t.Errorf("loyalty %d, want 2", got)
	}
}

func TestS82BnsWindgraceMinusElevenDestroysNonlandsAndMakesSixCats(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	w := s82BnsWindgraceOnBattlefield(g, me.ID, 11)
	a := pushVanillaCreature(g, opp.ID, "Wolf", 2, 2)
	b := pushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	land := pushCatalogPermanent(g, opp.ID, "Island", "Basic Land — Island", "", false)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, w, 2, game.ActivateAbilityParams{Strict: true, Targets: cardTargets(land)}); err == nil {
		t.Error("targeted a land")
	}
	if err := g.ActivateCatalogAbility(me.ID, w, 2, game.ActivateAbilityParams{Strict: true, Targets: cardTargets(a, b)}); err != nil {
		t.Fatalf("-11: %v", err)
	}
	passPriorityAroundTable(t, g)
	if onBattlefieldByID(g, a) || onBattlefieldByID(g, b) {
		t.Error("a targeted permanent survived")
	}
	if !onBattlefieldByID(g, land) {
		t.Error("the land was destroyed")
	}
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Cat Warrior" || c.Controller != me.ID {
			continue
		}
		n++
		if !c.IsToken() || c.Power != 2 || c.Toughness != 2 || len(c.Colors) != 1 || c.Colors[0] != "G" {
			t.Errorf("Cat Warrior %d/%d %v, want a 2/2 green token", c.Power, c.Toughness, c.Colors)
		}
		assertKeywords(t, g, c.InstanceID, "forestwalk")
	}
	if n != 6 {
		t.Errorf("%d Cat Warriors, want 6", n)
	}
}

func TestS82BnsWindgraceMinusElevenWithNoTargetsStillMakesTheCats(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	w := s82BnsWindgraceOnBattlefield(g, me.ID, 11)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, w, 2, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("-11: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countBattlefieldByName(g, "Cat Warrior"); got != 6 {
		t.Errorf("%d Cat Warriors, want 6", got)
	}
}
