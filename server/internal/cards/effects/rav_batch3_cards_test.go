package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rav_batch3_cards_test.go — Ravnica: City of Guilds, batch 3 (#5600).
// Each test plays the real catalog entry and asserts what its text says.

const (
	ravB3SwineOracle        = "c00f082b-787f-4f96-8356-c10d35713bc4"
	ravB3BlessingOracle     = "938e738f-c982-4bbf-b15f-0c3dd3a9901e"
	ravB3CongregationOracle = "b63dad39-4dd5-4340-953e-b05df1a3c776"
	ravB3HouseGuardOracle   = "35c027cd-f754-4e41-9242-8a9d0481171c"
	ravB3InfiltratorOracle  = "d215127b-5f1d-4d3c-935b-c228c9c7f874"
	ravB3MachinationsOracle = "cab67380-4e1a-4210-a98b-d2a21022b4b7"
	ravB3DisembowelOracle   = "6fa90258-3c9f-4723-880f-31871c0772e0"
	ravB3GriffinOracle      = "11e84e48-0f18-41b6-9f2a-30c64aecf9a7"
	ravB3DizzyOracle        = "7de04284-dcaa-4b63-a734-fcfe2698f83c"
	ravB3DogpileOracle      = "a4364218-f746-42ff-8495-bf8ca69c3b62"
	ravB3DowsingOracle      = "c9cb34df-75b1-4dda-8f22-daadf59c5332"
	ravB3DrakeOracle        = "5d144825-594a-454b-9276-fe92e87c19c0"
	ravB3DriftOracle        = "6ee51134-c4ad-4101-8ce2-23dde9359cd4"
	ravB3DromadOracle       = "c7adac8f-8187-49a0-80b2-3281d42f1b92"
)

func TestRavB3EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		ravB3SwineOracle:        "Coalhauler Swine",
		ravB3BlessingOracle:     "Conclave's Blessing",
		ravB3CongregationOracle: "Congregation at Dawn",
		ravB3HouseGuardOracle:   "Dimir House Guard",
		ravB3InfiltratorOracle:  "Dimir Infiltrator",
		ravB3MachinationsOracle: "Dimir Machinations",
		ravB3DisembowelOracle:   "Disembowel",
		ravB3GriffinOracle:      "Divebomber Griffin",
		ravB3DizzyOracle:        "Dizzy Spell",
		ravB3DogpileOracle:      "Dogpile",
		ravB3DowsingOracle:      "Dowsing Shaman",
		ravB3DrakeOracle:        "Drake Familiar",
		ravB3DriftOracle:        "Drift of Phantasms",
		ravB3DromadOracle:       "Dromad Purebred",
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

// ravB3Seats is the active seat and the next one.
func ravB3Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// ravB3LibraryTop is the top n cards of a library, top-first.
func ravB3LibraryTop(p *game.Player, n int) []uuid.UUID {
	out := make([]uuid.UUID, 0, n)
	for i := len(p.Library.Cards) - 1; i >= 0 && len(out) < n; i-- {
		out = append(out, p.Library.Cards[i].InstanceID)
	}
	return out
}

// ravB3SameOrder reports whether two ID lists match in order.
func ravB3SameOrder(got, want []uuid.UUID) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

// --- Coalhauler Swine, Dromad Purebred ------------------------------------------

func TestRavB3CoalhaulerSwineDealsWhatItIsDealtToEachPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	swine := b27Push(g, me.ID, "Coalhauler Swine", "Creature — Boar Beast", ravB3SwineOracle, "{4}{R}{R}", 4, 4, "R")
	src := b16Creature(g, opp.ID, "Source", "Creature — Ogre", 5, 5, "R")
	cmdr2B08Damage(g, src, swine, 2)
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if p.Life != ravB2StartingLife-2 {
			t.Errorf("seat %s life %d, want %d", p.ID, p.Life, ravB2StartingLife-2)
		}
	}
	if !g.Battlefield.Contains(swine) {
		t.Error("the Swine died to 2 damage")
	}
}

func TestRavB3CoalhaulerSwineStillDealsItWhenTheDamageKillsIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	swine := b27Push(g, me.ID, "Coalhauler Swine", "Creature — Boar Beast", ravB3SwineOracle, "{4}{R}{R}", 4, 4, "R")
	src := b16Creature(g, opp.ID, "Source", "Creature — Ogre", 5, 5, "R")
	cmdr2B08Damage(g, src, swine, 5)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(swine) {
		t.Fatal("the Swine survived 5 damage")
	}
	for _, p := range g.Seats {
		if p.Life != ravB2StartingLife-5 {
			t.Errorf("seat %s life %d, want %d", p.ID, p.Life, ravB2StartingLife-5)
		}
	}
}

func TestRavB3DromadPurebredGainsOneLifePerDamageEvent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dromad := b27Push(g, me.ID, "Dromad Purebred", "Creature — Camel Beast", ravB3DromadOracle, "{4}{W}", 1, 5, "W")
	src := b16Creature(g, opp.ID, "Source", "Creature — Ogre", 5, 5, "R")
	cmdr2B08Damage(g, src, dromad, 3)
	passPriorityAroundTable(t, g)
	if me.Life != ravB2StartingLife+1 {
		t.Errorf("life %d, want %d", me.Life, ravB2StartingLife+1)
	}
	if opp.Life != ravB2StartingLife {
		t.Errorf("the opponent's life changed to %d", opp.Life)
	}
}

// --- Conclave's Blessing -------------------------------------------------------

func TestRavB3ConclavesBlessingConvokesAndCountsOtherCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ravB3Seats(g)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	taps := ravB1Tappers(g, me.ID, 4, "W")
	if _, err := castWithTapParams(t, g, "Conclave's Blessing", "Enchantment — Aura", "{3}{W}", ravB3BlessingOracle,
		game.CastSpellParams{Strict: true, TapIDs: taps, Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("convoke cast with four white creatures: %v", err)
	}
	passPriorityAroundTable(t, g)
	// Four other creatures: +0/+8.
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 2 || tough != 10 {
		t.Fatalf("the Bear is %d/%d, want 2/10", p, tough)
	}
	b12Creature(g, me.ID, "Another", "Creature — Elf", 1, 1)
	if tough := effectiveToughness(t, g, bear); tough != 12 {
		t.Errorf("with a fifth other creature, toughness %d, want 12", tough)
	}
}

// --- Congregation at Dawn ------------------------------------------------------

// ravB3CongregationBoard seeds three creature cards and a sorcery in the
// active seat's library, casts Congregation at Dawn and returns the search
// prompt.
func ravB3CongregationBoard(t *testing.T) (g *game.Game, me, opp *game.Player, creatures []uuid.UUID, c *game.PendingChoice) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = ravB3Seats(g)
	for _, name := range []string{"Alpha", "Beta", "Gamma"} {
		creatures = append(creatures, pushLibraryCardForTest(me, game.Card{
			InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Elf", Owner: me.ID, Controller: me.ID,
		}))
	}
	pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Not a Creature", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	castCatalogSpell(t, g, "Congregation at Dawn", "Instant", ravB3CongregationOracle, nil)
	passPriorityAroundTable(t, g)
	c = searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Not a Creature") != uuid.Nil {
		t.Error("the search offered a noncreature card")
	}
	return g, me, opp, creatures, c
}

func TestRavB3CongregationAtDawnPutsThreeOnTopInTheChosenOrder(t *testing.T) {
	g, me, opp, cs, c := ravB3CongregationBoard(t)
	if err := g.ResolveSearchLibrary(c.ID, me.ID, cs); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	order := latestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, me.ID)
	if order == nil {
		t.Fatal("no prompt to order the three cards")
	}
	want := []uuid.UUID{cs[2], cs[0], cs[1]}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, nil, want); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if got := ravB3LibraryTop(me, 3); !ravB3SameOrder(got, want) {
		t.Fatalf("library top %v, want %v", got, want)
	}
	for _, card := range me.Library.Cards[len(me.Library.Cards)-3:] {
		if !card.KnownBy[me.ID] {
			t.Errorf("the caster does not know where %s is", card.Name)
		}
		if card.KnownBy[opp.ID] {
			t.Errorf("an opponent knows where %s is after the caster ordered them", card.Name)
		}
	}
}

func TestRavB3CongregationAtDawnWithOneCardPutsItOnTop(t *testing.T) {
	g, me, opp, cs, c := ravB3CongregationBoard(t)
	if err := g.ResolveSearchLibrary(c.ID, me.ID, cs[1:2]); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if latestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, me.ID) != nil {
		t.Error("asked to order a single card")
	}
	top := me.Library.Cards[len(me.Library.Cards)-1]
	if top.InstanceID != cs[1] {
		t.Fatalf("the top card is %s, want Beta", top.Name)
	}
	if !top.KnownBy[opp.ID] {
		t.Error("the revealed card on top is not known to the table")
	}
}

// --- transmute (Dimir House Guard, Dimir Infiltrator, Dimir Machinations,
// Dizzy Spell, Drift of Phantasms) ----------------------------------------------

// ravB3Transmutes puts a transmute card in the active seat's hand, floats
// the ability's mana, activates its transmute row, and checks the search
// offers the library card with the same mana value and not the one with
// another.
func ravB3Transmutes(t *testing.T, name, typeLine, manaCost, oracle string, row int, mana []string, sameCost, otherCost string) {
	t.Helper()
	g := newCatalogGame(t)
	me, _ := ravB3Seats(g)
	advanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: manaCost,
		OracleID: oracle, Owner: me.ID, Controller: me.ID,
	})
	same := pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Same Value", TypeLine: "Sorcery", ManaCost: sameCost, Owner: me.ID, Controller: me.ID})
	pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Other Value", TypeLine: "Sorcery", ManaCost: otherCost, Owner: me.ID, Controller: me.ID})
	b10AddMana(me, mana...)
	if err := g.ActivateCatalogAbility(me.ID, id, row, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("transmute: %v", err)
	}
	if !me.Graveyard.Contains(id) {
		t.Fatalf("%s was not discarded to pay for transmute", name)
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Other Value") != uuid.Nil {
		t.Error("the search offered a card with another mana value")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{same}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !me.Hand.Contains(same) {
		t.Error("the card with the same mana value did not reach the hand")
	}
}

func TestRavB3TransmuteCards(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, cost, oracle string
		row                          int
		mana                         []string
		same, other                  string
	}{
		{"Dimir House Guard", "Creature — Skeleton", "{3}{B}", ravB3HouseGuardOracle, 1, []string{"C", "B", "B"}, "{2}{G}{G}", "{4}{U}"},
		{"Dimir Infiltrator", "Creature — Spirit", "{U}{B}", ravB3InfiltratorOracle, 0, []string{"C", "U", "B"}, "{1}{G}", "{G}"},
		{"Dimir Machinations", "Sorcery", "{2}{B}", ravB3MachinationsOracle, 0, []string{"C", "B", "B"}, "{2}{U}", "{3}{U}"},
		{"Dizzy Spell", "Instant", "{U}", ravB3DizzyOracle, 0, []string{"C", "U", "U"}, "{G}", "{1}{G}"},
		{"Drift of Phantasms", "Creature — Spirit", "{2}{U}", ravB3DriftOracle, 0, []string{"C", "U", "U"}, "{1}{G}{G}", "{1}{G}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ravB3Transmutes(t, tc.name, tc.typeLine, tc.cost, tc.oracle, tc.row, tc.mana, tc.same, tc.other)
		})
	}
}

// --- Dimir House Guard, Dimir Infiltrator, Drift of Phantasms ------------------

func TestRavB3DimirHouseGuardHasFearAndRegeneratesBySacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	guard := b12Push(g, me.ID, "Dimir House Guard", "Creature — Skeleton", ravB3HouseGuardOracle, 2, 3)
	if !p1g2HasKeyword(g, guard, "fear") {
		t.Error("the Guard has no fear")
	}
	land := b12Permanent(g, me.ID, "My Land", "Land")
	if err := ravB2Activate(t, g, me.ID, guard, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{land}}); err == nil {
		t.Fatal("a land paid \"sacrifice a creature\"")
	}
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Rat", 1, 1)
	ravB2MustActivate(t, g, me.ID, guard, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}})
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	hobDestroy(t, g, guard)
	if !g.Battlefield.Contains(guard) {
		t.Error("the regeneration shield did not save the Guard")
	}
}

func TestRavB3DimirInfiltratorCantBeBlocked(t *testing.T) {
	g := newCatalogGame(t)
	inf := b12Push(g, g.Seats[0].ID, "Dimir Infiltrator", "Creature — Spirit", ravB3InfiltratorOracle, 1, 3)
	if restrictionsOf(t, g, inf)&game.CantBeBlocked == 0 {
		t.Error("the Infiltrator can be blocked")
	}
}

func TestRavB3DriftOfPhantasmsHasDefenderAndFlying(t *testing.T) {
	g := newCatalogGame(t)
	drift := b12Push(g, g.Seats[0].ID, "Drift of Phantasms", "Creature — Spirit", ravB3DriftOracle, 0, 5)
	for _, kw := range []string{"defender", "flying"} {
		if !p1g2HasKeyword(g, drift, kw) {
			t.Errorf("the Drift has no %s", kw)
		}
	}
}

// --- Dimir Machinations --------------------------------------------------------

// ravB3Machinations casts Dimir Machinations at the next seat, whose
// library has four named cards on top, and returns the exile prompt.
func ravB3Machinations(t *testing.T) (g *game.Game, me, opp *game.Player, top []uuid.UUID, pick *game.PendingChoice) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = ravB3Seats(g)
	top = seedLibrary(opp, "One", "Two", "Three", "Four")
	castCatalogSpell(t, g, "Dimir Machinations", "Sorcery", ravB3MachinationsOracle, playerRefs(opp.ID))
	passPriorityAroundTable(t, g)
	pick = latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt to exile any of the cards looked at")
	}
	return g, me, opp, top, pick
}

func TestRavB3DimirMachinationsExilesOneAndOrdersTheRest(t *testing.T) {
	g, me, opp, top, pick := ravB3Machinations(t)
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{top[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !g.Exile.Contains(top[1]) {
		t.Fatal("the picked card was not exiled")
	}
	order := latestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, me.ID)
	if order == nil {
		t.Fatal("no prompt to put the rest back in any order")
	}
	want := []uuid.UUID{top[2], top[0], top[3]}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, nil, want[:2]); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if got := ravB3LibraryTop(opp, 3); !ravB3SameOrder(got, want) {
		t.Errorf("library top %v, want %v", got, want)
	}
}

func TestRavB3DimirMachinationsMayExileNone(t *testing.T) {
	g, me, opp, top, pick := ravB3Machinations(t)
	exiled := g.Exile.Size()
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("ResolveChooseCards(none): %v", err)
	}
	if g.Exile.Size() != exiled {
		t.Error("a card was exiled with none picked")
	}
	order := latestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, me.ID)
	if order == nil {
		t.Fatal("no prompt to put the three back in any order")
	}
	want := []uuid.UUID{top[2], top[1], top[0]}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, nil, want); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if got := ravB3LibraryTop(opp, 3); !ravB3SameOrder(got, want) {
		t.Errorf("library top %v, want %v", got, want)
	}
}

func TestRavB3DimirMachinationsMayExileAllThree(t *testing.T) {
	g, me, opp, top, pick := ravB3Machinations(t)
	if err := g.ResolveChooseCards(pick.ID, me.ID, top[:3]); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	for _, id := range top[:3] {
		if !g.Exile.Contains(id) {
			t.Errorf("card %v was not exiled", id)
		}
	}
	if latestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, me.ID) != nil {
		t.Error("asked to order an empty pile")
	}
	if got := ravB3LibraryTop(opp, 1); got[0] != top[3] {
		t.Error("the fourth card is not on top")
	}
}

// --- Disembowel ----------------------------------------------------------------

func TestRavB3DisembowelDestroysACreatureWithManaValueX(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := ravB3Seats(g)
	two := paPermanent(g, opp.ID, "Two Drop", "Creature — Bear", "{1}{G}", 2, 2)
	three := paPermanent(g, opp.ID, "Three Drop", "Creature — Bear", "{2}{G}", 3, 3)
	b06AddMana(me, "C", "C", "B")
	if _, err := castWithTapParams(t, g, "Disembowel", "Instant", "{X}{B}", ravB3DisembowelOracle,
		game.CastSpellParams{Strict: true, XValue: 2, Targets: cardRefs(three)}); err == nil {
		t.Fatal("X=2 targeted a mana value 3 creature")
	}
	if _, err := castWithTapParams(t, g, "Disembowel", "Instant", "{X}{B}", ravB3DisembowelOracle,
		game.CastSpellParams{Strict: true, XValue: 2, Targets: cardRefs(two)}); err != nil {
		t.Fatalf("cast with X=2 at a mana value 2 creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(two) {
		t.Error("the mana value 2 creature was not destroyed")
	}
	if !g.Battlefield.Contains(three) {
		t.Error("the other creature was destroyed")
	}
}

// --- Divebomber Griffin --------------------------------------------------------

func TestRavB3DivebomberGriffinHitsAnAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	griffin := b12Push(g, me.ID, "Divebomber Griffin", "Creature — Griffin", ravB3GriffinOracle, 3, 2)
	if !p1g2HasKeyword(g, griffin, "flying") {
		t.Error("the Griffin has no flying")
	}
	idle := b12Creature(g, opp.ID, "Idle Bear", "Creature — Bear", 3, 3)
	if err := ravB2Activate(t, g, me.ID, griffin, 0, game.ActivateAbilityParams{Targets: cardRefs(idle)}); err == nil {
		t.Fatal("the Griffin targeted a creature that is not attacking or blocking")
	}
	attacker := b12Creature(g, opp.ID, "Attacking Bear", "Creature — Bear", 3, 3)
	crtBlock(g, uuid.Nil, attacker, me.ID)
	ravB2MustActivate(t, g, me.ID, griffin, 0, game.ActivateAbilityParams{Targets: cardRefs(attacker)})
	if g.Battlefield.Contains(attacker) {
		t.Error("the attacking 3/3 survived 3 damage")
	}
	if !me.Graveyard.Contains(griffin) {
		t.Error("the Griffin was not sacrificed")
	}
}

// --- Dizzy Spell ---------------------------------------------------------------

func TestRavB3DizzySpellGivesMinusThreeMinusZero(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := ravB3Seats(g)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 4, 3)
	castCatalogSpell(t, g, "Dizzy Spell", "Instant", ravB3DizzyOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 1 || tough != 3 {
		t.Errorf("the 4/3 is %d/%d, want 1/3", p, tough)
	}
}

// --- Dogpile -------------------------------------------------------------------

func TestRavB3DogpileCountsAttackingCreaturesYouControl(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := ravB3Seats(g)
	for _, name := range []string{"First", "Second"} {
		crtBlock(g, uuid.Nil, b12Creature(g, me.ID, name, "Creature — Hound", 1, 1), opp.ID)
	}
	b12Creature(g, me.ID, "Home Guard", "Creature — Hound", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Attacker", "Creature — Hound", 1, 1)
	crtBlock(g, uuid.Nil, theirs, me.ID)
	castCatalogSpell(t, g, "Dogpile", "Instant", ravB3DogpileOracle, playerRefs(opp.ID))
	passPriorityAroundTable(t, g)
	if opp.Life != ravB2StartingLife-2 {
		t.Errorf("life %d, want %d (two attacking creatures you control)", opp.Life, ravB2StartingLife-2)
	}
}

func TestRavB3DogpileWithNoAttackersDealsNothing(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	_, opp := ravB3Seats(g)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Dogpile", "Instant", ravB3DogpileOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) || damageMarkedOn(g, bear) != 0 {
		t.Error("Dogpile dealt damage with no attacking creatures")
	}
}

// --- Dowsing Shaman ------------------------------------------------------------

func TestRavB3DowsingShamanReturnsAnEnchantmentCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	shaman := b12Push(g, me.ID, "Dowsing Shaman", "Creature — Centaur Shaman", ravB3DowsingOracle, 3, 4)
	aura := b17GraveyardCard(me, "Old Aura", "Enchantment — Aura", "{1}{W}")
	bear := b17GraveyardCard(me, "Grizzly Bears", "Creature — Bear", "{1}{G}")
	if err := ravB2Activate(t, g, me.ID, shaman, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)}, "C", "C", "G"); err == nil {
		t.Fatal("a creature card is not an enchantment card")
	}
	me.ManaPool = game.ManaPool{}
	ravB2MustActivate(t, g, me.ID, shaman, 0, game.ActivateAbilityParams{Targets: cardRefs(aura)}, "C", "C", "G")
	if !me.Hand.Contains(aura) {
		t.Error("the enchantment card did not return to hand")
	}
	if !tapCostTapped(g, shaman) {
		t.Error("the ability has a {T} cost")
	}
}

// --- Drake Familiar ------------------------------------------------------------

// ravB3CastDrake casts Drake Familiar from the active seat's hand with its
// mana floated and resolves it, leaving its enters trigger's question open.
func ravB3CastDrake(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	toMainForCost(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Drake Familiar", TypeLine: "Creature — Drake", ManaCost: "{1}{U}",
		OracleID: ravB3DrakeOracle, Power: 2, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	b06AddMana(me, "C", "U")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast Drake Familiar: %v", err)
	}
	passPriorityAroundTable(t, g)
	return id
}

func TestRavB3DrakeFamiliarReturnsAnyEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ravB3Seats(g)
	theirs := b12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	drake := ravB3CastDrake(t, g, me)
	if !p1g2HasKeyword(g, drake, "flying") {
		t.Error("the Drake has no flying")
	}
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt to return an enchantment")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{theirs}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !opp.Hand.Contains(theirs) {
		t.Error("the enchantment did not return to its owner's hand")
	}
	if !g.Battlefield.Contains(drake) {
		t.Error("the Drake was sacrificed after the return")
	}
}

func TestRavB3DrakeFamiliarIsSacrificedWhenNothingReturns(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ravB3Seats(g)
	drake := ravB3CastDrake(t, g, me)
	if g.Battlefield.Contains(drake) {
		t.Fatal("the Drake stayed with no enchantment on the battlefield")
	}

	g2 := newCatalogGame(t)
	me2, _ := ravB3Seats(g2)
	mine := b12Permanent(g2, me2.ID, "My Enchantment", "Enchantment")
	drake2 := ravB3CastDrake(t, g2, me2)
	pick := latestChooseCardsFor(g2, me2.ID)
	if pick == nil {
		t.Fatal("no prompt to return an enchantment")
	}
	if err := g2.ResolveChooseCards(pick.ID, me2.ID, nil); err != nil {
		t.Fatalf("ResolveChooseCards(none): %v", err)
	}
	passPriorityAroundTable(t, g2)
	if g2.Battlefield.Contains(drake2) || !g2.Battlefield.Contains(mine) {
		t.Error("declining should sacrifice the Drake and keep the enchantment")
	}
}
