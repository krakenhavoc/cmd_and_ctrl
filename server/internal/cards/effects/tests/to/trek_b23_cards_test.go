package teststo

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b23_cards_test.go — Star Trek, batch 23 (#4973). Each test plays
// the real catalog entry and asserts what its text says.

const (
	trekB23VoyagerOracle = "33e8904a-9285-4031-900c-0a5c8f0bf6c0"
	trekB23WarpOracle    = "5a824931-2b65-42b1-8653-d4313a6693d2"
	trekB23XenoOracle    = "a9df60bf-f327-4d50-ab4d-99324fc2dcfe"
	trekB23XindiOracle   = "5fdff691-686d-4ce0-bd12-eea676a648b4"
	trekB23AllGoodOracle = "61dc9770-1e29-49fa-a31b-b63b5f0496fb"
	trekB23CityOracle    = "47c9ef8c-7ba7-4702-bd8c-5fedc59c8166"
	trekB23BookerOracle  = "b80879d3-5abd-47e5-8b8c-61db120854e4"
	trekB23OdoOracle     = "74b58f31-7634-4b51-a900-6fbfccfb3cd2"
	trekB23DalOracle     = "7cecf129-3df9-4ebf-b194-28d1b34faaab"
)

func TestTrekB23EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB23VoyagerOracle: "U.S.S. Voyager, Lone Wanderer",
		trekB23WarpOracle:    "Warp Speed",
		trekB23XenoOracle:    "Xenobotanist",
		trekB23XindiOracle:   "Xindi Surveyors",
		trekB23AllGoodOracle: "All Good Things...",
		trekB23CityOracle:    "The City on the Edge of Forever",
		trekB23BookerOracle:  "Cleveland Booker, Empath",
		trekB23OdoOracle:     "Constable Odo",
		trekB23DalOracle:     "Dal R'El, Augment Orphan",
	} {
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

// --- fixtures -------------------------------------------------------------------

// trekB23Library stacks cards on top of p's library, first named on top.
func trekB23Library(p *game.Player, cards ...game.Card) []uuid.UUID {
	ids := make([]uuid.UUID, len(cards))
	for i := len(cards) - 1; i >= 0; i-- {
		c := cards[i]
		c.InstanceID = uuid.New()
		c.Owner, c.Controller = p.ID, p.ID
		p.Library.PushTop(c)
		ids[i] = c.InstanceID
	}
	return ids
}

// trekB23Hand puts a card of the given type line in p's hand.
func trekB23Hand(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

// trekB23Tokens lists the tokens named `name` that p controls.
func trekB23Tokens(g *game.Game, p uuid.UUID, name string) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == p && c.Name == name && effects.IsToken(c) {
			out = append(out, c)
		}
	}
	return out
}

// trekB23ExileFromHand answers the open "exile a card from your hand"
// prompt with pick.
func trekB23ExileFromHand(t *testing.T, g *game.Game, p uuid.UUID, pick uuid.UUID) {
	t.Helper()
	ch := effectstest.LatestChooseCards(g, p)
	if ch == nil || ch.ChooseMin != 1 || ch.ChooseMax != 1 {
		t.Fatalf("expected a must-pick one card of your hand: %+v", ch)
	}
	if err := g.ResolveChooseCards(ch.ID, p, []uuid.UUID{pick}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func trekB23PhasedOut(g *game.Game, id uuid.UUID) bool {
	out := false
	g.ReadSnapshot(func() { out = g.IsPhasedOutForEffect(id) })
	return out
}

// --- U.S.S. Voyager, Lone Wanderer ---------------------------------------------------

func TestTrekB23VoyagerPutsTwoLandsTappedAndStationsAtEight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lib := trekB23Library(me,
		game.Card{Name: "Spell A", TypeLine: "Instant"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Spell B", TypeLine: "Sorcery"},
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Third Land", TypeLine: "Basic Land — Plains"},
	)
	ship := effectstest.CastFromHand(t, g, game.Card{Name: "U.S.S. Voyager, Lone Wanderer", TypeLine: "Legendary Artifact — Spacecraft",
		OracleID: trekB23VoyagerOracle, ManaCost: "{3}{G}", Power: 3, Toughness: 3}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{lib[1], lib[3]} {
		c, ok := effectstest.BattlefieldCard(g, id)
		if !ok || !c.Tapped {
			t.Errorf("%s: on the battlefield %v, tapped %v; want both", id, ok, c.Tapped)
		}
	}
	if g.Battlefield.Contains(lib[4]) || !me.Library.Contains(lib[4]) {
		t.Error("the third land was revealed past the second land")
	}
	if top, err := me.Library.Top(); err != nil || top.InstanceID != lib[4] {
		t.Errorf("the library top is %v, want the unrevealed third land", top.Name)
	}
	for _, id := range []uuid.UUID{lib[0], lib[2]} {
		if !me.Library.Contains(id) {
			t.Errorf("revealed nonland %v did not go to the bottom of the library", id)
		}
	}

	if effectstest.LayeredCard(t, g, ship).IsCreature() {
		t.Fatal("an unstationed Spacecraft is a creature")
	}
	effectstest.AdvanceToMain(t, g)
	seven := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Seven", TypeLine: "Creature — Human", Power: 7, Toughness: 7, Owner: me.ID, Controller: me.ID})
	effectstest.Station(t, g, me, ship, seven)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LayeredCard(t, g, ship); c.IsCreature() {
		t.Fatal("Voyager is a creature at 7 charge counters")
	}
	one := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "One", TypeLine: "Creature — Human", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	effectstest.Station(t, g, me, ship, one)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, ship)
	if !c.IsCreature() || !game.HasKeyword(&c, "flying") || !game.HasKeyword(&c, "vigilance") || c.CurrentPower() != 3 {
		t.Errorf("at 8+ Voyager is a 3/3 flying, vigilant artifact creature; creature=%v flying=%v vigilance=%v power=%d",
			c.IsCreature(), game.HasKeyword(&c, "flying"), game.HasKeyword(&c, "vigilance"), c.CurrentPower())
	}
}

// --- Warp Speed -------------------------------------------------------------------------

func TestTrekB23WarpSpeedFindsAPlanetAndUntapsTaplands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lib := trekB23Library(me,
		game.Card{Name: "Some Nonbasic", TypeLine: "Land"},
		game.Card{Name: "Some Planet", TypeLine: "Land — Planet"},
	)
	effectstest.CastFromHand(t, g, game.Card{Name: "Warp Speed", TypeLine: "Enchantment", OracleID: trekB23WarpOracle, ManaCost: "{2}{G}"}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	var offered []uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceSearchLibrary {
				offered = append(offered, c.SearchCards...)
			}
		}
	})
	for _, id := range offered {
		if id == lib[0] {
			t.Error("a nonbasic land that is not a Planet was offered")
		}
	}
	effectstest.TakeSearchPrompts(t, g)
	if !me.Hand.Contains(lib[1]) {
		t.Fatal("the Planet card was not put into your hand")
	}

	effectstest.AdvanceToMain(t, g)
	temple := effectstest.PlayLandFromHand(t, g, "Temple of Silence", effectstest.TempleOfSilenceOracle)
	if c, ok := effectstest.BattlefieldCard(g, temple); !ok || c.Tapped {
		t.Errorf("the temple is on the battlefield %v and tapped %v; Warp Speed makes it enter untapped", ok, c.Tapped)
	}
}

func TestTrekB23WarpSpeedDrawsTwoForLessWithDifferentlyNamedLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	warp := effectstest.B27Push(g, me.ID, "Warp Speed", "Enchantment", trekB23WarpOracle, "{2}{G}", 0, 0, "G")
	// Five names, tapped so they pay nothing; a second Island adds no name.
	for _, n := range []string{"Island", "Island", "Plains", "Swamp", "Mountain", "Forest"} {
		effectstest.TmtB4Tap(g, effectstest.B12Permanent(g, me.ID, n, "Basic Land — "+n))
	}
	effectstest.AdvanceToMain(t, g)
	before := me.Hand.Size()
	// {7}{U}{U} less five is {2}{U}{U}.
	effectstest.FloatMana(t, g, me, "{C}{U}{U}")
	if err := g.ActivateCatalogAbility(me.ID, warp, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated for {1}{U}{U} with five differently named lands")
	}
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, warp, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate for {2}{U}{U}: %v", err)
	}
	if g.Battlefield.Contains(warp) {
		t.Error("the sacrifice cost left Warp Speed on the battlefield")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
}

// --- Xenobotanist -------------------------------------------------------------------------

func TestTrekB23XenobotanistTapsForAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	xeno := effectstest.PushCatalogPermanent(g, me.ID, "Xenobotanist", "Creature — Vulcan Scientist", trekB23XenoOracle, false)
	if err := g.ActivateManaAbility(me.ID, xeno, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := effectstest.RiderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("want a five-colour pick, got %+v", pick)
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "B"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if me.ManaPool.Total() != 1 || me.ManaPool.Units()[0].Color != "B" {
		t.Errorf("pool = %+v, want one {B}", me.ManaPool)
	}
	if !effectstest.LayeredCard(t, g, xeno).Tapped {
		t.Error("the {T} cost did not tap Xenobotanist")
	}
}

func TestTrekB23XenobotanistCantTapWhileSummoningSick(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	xeno := effectstest.PushCatalogPermanent(g, me.ID, "Xenobotanist", "Creature — Vulcan Scientist", trekB23XenoOracle, true)
	if err := g.ActivateManaAbility(me.ID, xeno, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("a summoning-sick Xenobotanist tapped for mana")
	}
}

// --- Xindi Surveyors -------------------------------------------------------------------------

func TestTrekB23XindiSurveyorsGrowsOnYourLandfallOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	xindi := effectstest.PushCatalogPermanent(g, me.ID, "Xindi Surveyors", "Creature — Xindi Scientist", trekB23XindiOracle, false)
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, xindi, game.CounterPlusOne); got != 1 {
		t.Fatalf("+1/+1 counters after your land = %d, want 1", got)
	}
	// An opponent's land entering is not "a land you control".
	effectstest.SeedLandOnBattlefield(g, opp.ID, "Their Forest", "Basic Land — Forest")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, xindi, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters after an opponent's land = %d, want still 1", got)
	}
}

// --- All Good Things... -----------------------------------------------------------------------

// trekB23AllGoodThingsRun casts the Saga and walks it through all four
// chapters, exiling the hand card exile(i) returns at chapter i (1-3).
// Returns the life before chapter IV resolved and after.
func trekB23AllGoodThingsRun(t *testing.T, exile func(g *game.Game, me *game.Player, chapter int) uuid.UUID) (*game.Game, *game.Player, uuid.UUID, int) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	var saga uuid.UUID
	for ch := 1; ch <= 3; ch++ {
		if ch > 1 {
			effectstest.AdvanceToPrecombatMainOf(t, g, seat)
		}
		pick := exile(g, me, ch)
		before := me.Hand.Size()
		if ch == 1 {
			// The Saga leaves the hand it was put into: net zero.
			saga = effectstest.CastCatalogSpell(t, g, "All Good Things...", "Enchantment — Saga", trekB23AllGoodOracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := me.Hand.Size() - before; got != 2 {
			t.Fatalf("chapter %d drew %d, want 2", ch, got)
		}
		trekB23ExileFromHand(t, g, me.ID, pick)
		if !g.Exile.Contains(pick) {
			t.Fatalf("chapter %d did not exile the chosen card", ch)
		}
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	life := me.Life
	effectstest.PassPriorityAroundTable(t, g)
	return g, me, saga, life
}

func TestTrekB23AllGoodThingsGainsTenWhenTheExiledCardsShareAType(t *testing.T) {
	g, me, saga, before := trekB23AllGoodThingsRun(t, func(_ *game.Game, me *game.Player, ch int) uuid.UUID {
		if ch == 2 {
			return trekB23Hand(me, "Artifact Creature", "Artifact Creature — Golem")
		}
		return trekB23Hand(me, "Plain Creature", "Creature — Bear")
	})
	if got := me.Life - before; got != 10 {
		t.Errorf("chapter IV gained %d life, want 10 (all three are creatures)", got)
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed after chapter IV")
	}
}

func TestTrekB23AllGoodThingsGainsNothingWithoutASharedType(t *testing.T) {
	_, me, _, before := trekB23AllGoodThingsRun(t, func(_ *game.Game, me *game.Player, ch int) uuid.UUID {
		if ch == 3 {
			return trekB23Hand(me, "A Land", "Land")
		}
		return trekB23Hand(me, "Plain Creature", "Creature — Bear")
	})
	if got := me.Life - before; got != 0 {
		t.Errorf("chapter IV gained %d life with a creature, a creature and a land exiled; want 0", got)
	}
}

func TestTrekB23AllGoodThingsIgnoresACardThatLeftExile(t *testing.T) {
	var land uuid.UUID
	_, me, _, before := trekB23AllGoodThingsRun(t, func(g *game.Game, me *game.Player, ch int) uuid.UUID {
		if ch == 3 {
			// Before chapter III, the odd one out leaves exile: it is a new
			// object and no longer "exiled with this Saga" (CR 400.7).
			if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneExile}, game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, land); err != nil {
				t.Fatalf("move: %v", err)
			}
		}
		if ch == 2 {
			land = trekB23Hand(me, "A Land", "Land")
			return land
		}
		return trekB23Hand(me, "Plain Creature", "Creature — Bear")
	})
	if got := me.Life - before; got != 10 {
		t.Errorf("chapter IV gained %d life; the land left exile, so the rest share creature: want 10", got)
	}
}

// --- The City on the Edge of Forever -------------------------------------------------------------

func TestTrekB23CityPhasesOutAndTheHumanHoldsThemOutUntilSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	relic := effectstest.B12Permanent(g, opp.ID, "Their Relic", "Artifact")
	land := effectstest.B12Permanent(g, me.ID, "My Land", "Land")
	saga := effectstest.CastCatalogSpell(t, g, "The City on the Edge of Forever", "Enchantment — Saga", trekB23CityOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{mine, theirs, relic} {
		if !trekB23PhasedOut(g, id) {
			t.Errorf("%v did not phase out", id)
		}
	}
	if trekB23PhasedOut(g, land) || !g.Battlefield.Contains(saga) {
		t.Error("a land or the Saga phased out")
	}
	humans := trekB23Tokens(g, me.ID, "Human")
	if len(humans) != 1 {
		t.Fatalf("%d Human tokens, want 1", len(humans))
	}
	h := effectstest.LayeredCard(t, g, humans[0].InstanceID)
	if !h.IsCreature() || h.CurrentPower() != 0 || h.CurrentToughness() != 1 || len(h.Colors) != 1 || h.Colors[0] != "W" {
		t.Errorf("the Human is %d/%d %v, want a 0/1 white creature", h.CurrentPower(), h.CurrentToughness(), h.Colors)
	}

	// Each player's untap step passes once: nothing phases in.
	life := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		life[p.ID] = p.Life
	}
	before := me.Hand.Size()
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	for _, id := range []uuid.UUID{mine, theirs, relic} {
		if !trekB23PhasedOut(g, id) {
			t.Errorf("%v phased in while the Human was on the battlefield", id)
		}
	}

	// Chapter II: you draw two and each player loses 2 life.
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 3 {
		t.Errorf("hand grew by %d over the draw step and chapter II, want 3", got)
	}
	for _, p := range g.Seats {
		if got := life[p.ID] - p.Life; got != 2 {
			t.Errorf("seat %s lost %d life, want 2", p.Name, got)
		}
	}

	// Chapter III: the Human is the only creature token; sacrificing it
	// gains 5 life.
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	lifeBefore := me.Life
	effectstest.PassPriorityAroundTable(t, g)
	ask := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if ask == nil || len(ask.ChooseCards) != 1 || ask.ChooseCards[0] != humans[0].InstanceID {
		t.Fatalf("chapter III should offer the Human, the only creature token: %+v", ask)
	}
	effectstest.AnswerOwnPermanents(t, g, me.ID, humans[0].InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(humans[0].InstanceID) {
		t.Fatal("the Human was not sacrificed")
	}
	if got := me.Life - lifeBefore; got != 5 {
		t.Errorf("chapter III gained %d, want 5", got)
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed after chapter III")
	}

	// With the Human gone, each permanent phases in during the next untap
	// step of the player it phased out under.
	effectstest.AdvanceToPrecombatMainOf(t, g, (seat+1)%len(g.Seats))
	if trekB23PhasedOut(g, theirs) || trekB23PhasedOut(g, relic) {
		t.Error("the opponent's permanents did not phase in at their untap step")
	}
	if !trekB23PhasedOut(g, mine) {
		t.Error("your creature phased in during another player's untap step")
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	if trekB23PhasedOut(g, mine) {
		t.Error("your creature did not phase in at your untap step")
	}
}

func TestTrekB23CityChapterThreeWithNoTokenGainsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.CastCatalogSpell(t, g, "The City on the Edge of Forever", "Enchantment — Saga", trekB23CityOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	humans := trekB23Tokens(g, me.ID, "Human")
	if len(humans) != 1 {
		t.Fatalf("%d Human tokens, want 1", len(humans))
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, humans[0].InstanceID); err != nil {
		t.Fatalf("move: %v", err)
	}
	// A nontoken creature is no answer to "sacrifice a creature token".
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	life := me.Life
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID); c != nil {
		t.Fatalf("a sacrifice was asked with no creature token: %+v", c.ChooseCards)
	}
	if me.Life != life || !g.Battlefield.Contains(bear) {
		t.Errorf("life %d -> %d, bear on battlefield %v; want no change", life, me.Life, g.Battlefield.Contains(bear))
	}
}

// --- Cleveland Booker, Empath -------------------------------------------------------------------

func TestTrekB23BookerMakesAWormThatCantBlockAndPumpsForEight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	booker := effectstest.PushCatalogPermanent(g, me.ID, "Cleveland Booker, Empath", "Legendary Creature — Kwejian Rogue", trekB23BookerOracle, false)
	effectstest.PlayLandFromHand(t, g, "Mountain", "")
	effectstest.PassPriorityAroundTable(t, g)
	worms := trekB23Tokens(g, me.ID, "Worm")
	if len(worms) != 1 {
		t.Fatalf("%d Worm tokens, want 1", len(worms))
	}
	w := effectstest.LayeredCard(t, g, worms[0].InstanceID)
	if w.CurrentPower() != 2 || w.CurrentToughness() != 2 || len(w.Colors) != 2 || !game.Restricted(&w, game.CantBlock) {
		t.Errorf("the Worm is %d/%d %v cant-block=%v, want a 2/2 red and green Worm that can't block",
			w.CurrentPower(), w.CurrentToughness(), w.Colors, game.Restricted(&w, game.CantBlock))
	}

	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, booker, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the {8} ability was activated for seven")
	}
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, booker, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate {8}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LayeredCard(t, g, worms[0].InstanceID).CurrentPower(); p != 4 {
		t.Errorf("the Worm's power is %d, want 4", p)
	}
	if p := effectstest.LayeredCard(t, g, booker).CurrentPower(); p != 3 {
		t.Errorf("Booker's power is %d, want 3 (1 printed by the fixture, +2)", p)
	}
}

// --- Constable Odo ---------------------------------------------------------------------------------

func TestTrekB23OdoTapsAndDoubleStunsAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{W}{U}")
	effectstest.CastFromHandForTest(t, g, me, "Constable Odo", "Legendary Creature — Shapeshifter Detective", "{3}{W/U}{W/U}", trekB23OdoOracle, game.CastSpellParams{Strict: true})
	trekB9ModePick(t, g, me.ID, 0)
	if err := effectstest.SgAnswerTargets(t, g, theirs); err != nil {
		t.Fatalf("targeting their bear: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, theirs)
	if !c.Tapped || c.Counters[game.CounterStun] != 2 {
		t.Errorf("their bear tapped=%v stun=%d, want tapped with two stun counters", c.Tapped, c.Counters[game.CounterStun])
	}
}

func TestTrekB23OdoCantTargetYourOwnCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{W}{U}")
	effectstest.CastFromHandForTest(t, g, me, "Constable Odo", "Legendary Creature — Shapeshifter Detective", "{3}{W/U}{W/U}", trekB23OdoOracle, game.CastSpellParams{Strict: true})
	trekB9ModePick(t, g, me.ID, 0)
	if err := effectstest.SgAnswerTargets(t, g, mine); err == nil {
		t.Fatal("Odo's first mode targeted a creature you control")
	}
}

func TestTrekB23OdoInvestigatesTwice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{W}{U}")
	effectstest.CastFromHandForTest(t, g, me, "Constable Odo", "Legendary Creature — Shapeshifter Detective", "{3}{W/U}{W/U}", trekB23OdoOracle, game.CastSpellParams{Strict: true})
	trekB9ModePick(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(trekB23Tokens(g, me.ID, "Clue")); n != 2 {
		t.Errorf("%d Clues, want 2", n)
	}
}

// --- Dal R'El, Augment Orphan -------------------------------------------------------------------------

func TestTrekB23DalRElAddsRedOrGreenOnLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Dal R'El, Augment Orphan", "Legendary Creature — Human Mutant Rogue", trekB23DalOracle, false)
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.RiderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 2 {
		t.Fatalf("landfall should ask {R} or {G}, got %+v", pick)
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "U"); err == nil {
		t.Error("{U} was accepted for {R} or {G}")
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "R"); err != nil {
		t.Fatalf("ResolveManaChoice R: %v", err)
	}
	if me.ManaPool.Total() != 1 || me.ManaPool.Units()[0].Color != "R" {
		t.Errorf("pool = %+v, want one {R}", me.ManaPool)
	}
}

func TestTrekB23DalRElSacrificesALandToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dal := effectstest.PushCatalogPermanent(g, me.ID, "Dal R'El, Augment Orphan", "Legendary Creature — Human Mutant Rogue", trekB23DalOracle, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, dal, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no land to sacrifice")
	}
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	before := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, dal, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{land}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(land) {
		t.Error("the land was not sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}
