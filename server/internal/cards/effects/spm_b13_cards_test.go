package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spm_b13_cards_test.go — Marvel's Spider-Man, batch 13 (#4460). Each test
// plays the real catalog entry and asserts what its text says.

const (
	spmB13PeterOracle   = "5be3243a-6840-4601-96c8-ffb756041d49"
	spmB13SpawnOracle   = "7c3ff360-461a-44f3-a8f2-eb061b149cf5"
	spmB13JanesOracle   = "f2f4e14d-1ce8-41b3-b907-03ff561b42f7"
	spmB13CloneOracle   = "58dee718-6820-495a-8c03-5cc130b84167"
	spmB13BlastOracle   = "c4b3dc5e-651f-4307-9d76-d728e5054a36"
	spmB13DevourerOracl = "f064fc04-28fb-423f-a4f3-bb567c71168f"
	spmB13EddieOracle   = "9132d7e6-7190-45e1-9c57-8fe2e5d0f2f2"
)

func TestSpmB13EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		spmB13PeterOracle:   "Spider-Man, Peter Parker",
		spmB13SpawnOracle:   "Symbiote Spawn",
		spmB13JanesOracle:   "The Mary Janes",
		spmB13CloneOracle:   "Twisted Spider-Clone",
		spmB13BlastOracle:   "Venom Blast",
		spmB13DevourerOracl: "Venom, Deadly Devourer",
		spmB13EddieOracle:   "Venom, Eddie Brock",
		magikOracle:         "Magik, Illyana Rasputin",
		magikOracle + "#1":  "Darkchild, Demon Uncaged",
	}
	for oracle, name := range want {
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

// --- Spider-Man, Peter Parker ---

func TestSpmB13PeterGainLifeGrowsATargetCreatureYouControlAndMakesItIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	peter := b12Push(g, me.ID, "Spider-Man, Peter Parker", "Legendary Creature — Spider Human Hero", spmB13PeterOracle, 4, 4)
	assertKeywords(t, g, peter, "flying")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	foe := b12Creature(g, g.Seats[1].ID, "Foe", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	b04WaitForPick(t, g, me.ID)
	// The opponent's creature is not a legal pick.
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: foe}); err == nil {
			t.Fatal("an opponent's creature was a legal target")
		}
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if plusOnes(g, bear) != 1 {
		t.Errorf("the Bear has %d counters, want 1", plusOnes(g, bear))
	}
	if !hasKeywordOnBattlefield(t, g, bear, "indestructible") {
		t.Error("the Bear did not gain indestructible")
	}
	if plusOnes(g, foe) != 0 || plusOnes(g, peter) != 0 {
		t.Error("a counter landed on something that was not the target")
	}
}

func TestSpmB13PeterTriggersOncePerLifeGainAndNotForAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	peter := b12Push(g, me.ID, "Spider-Man, Peter Parker", "Legendary Creature — Spider Human Hero", spmB13PeterOracle, 4, 4)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, 3) })
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Fatal("an opponent gaining life triggered Peter")
	}
	passPriorityAroundTable(t, g)
	// Losing life does not trigger either; two gains trigger twice.
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -2) })
	for i := 0; i < 2; i++ {
		g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1) })
		b04WaitForPick(t, g, me.ID)
		pickCard(t, g, me.ID, peter)
		passPriorityAroundTable(t, g)
	}
	if plusOnes(g, peter) != 2 {
		t.Errorf("Peter has %d counters, want 2 (one per life gain)", plusOnes(g, peter))
	}
}

// --- Symbiote Spawn ---

func TestSpmB13SymbioteSpawnDrainsTwoWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spawn := b12Push(g, me.ID, "Symbiote Spawn", "Creature — Symbiote Villain", spmB13SpawnOracle, 3, 2)
	life := me.Life
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats[1:] {
		lives[s.ID] = s.Life
	}
	killCreature(t, g, me.ID, spawn)
	passPriorityAroundTable(t, g)
	for _, s := range g.Seats[1:] {
		if s.Life != lives[s.ID]-2 {
			t.Errorf("seat %s life %d, want %d", s.ID, s.Life, lives[s.ID]-2)
		}
	}
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
}

func TestSpmB13SymbioteSpawnDoesNothingWhenAnotherCreatureDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Symbiote Spawn", "Creature — Symbiote Villain", spmB13SpawnOracle, 3, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	life := me.Life
	killCreature(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life moved %d -> %d when another creature died", life, me.Life)
	}
}

// --- The Mary Janes ---

func spmB13CastJanes(t *testing.T, g *game.Game, mana string) error {
	t.Helper()
	me := g.Seats[0]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "The Mary Janes", TypeLine: "Legendary Creature — Human Bard Performer",
		OracleID: spmB13JanesOracle, ManaCost: "{3}{R}", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	floatMana(t, g, me, mana)
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
}

func TestSpmB13MaryJanesCostsOneLessForEachCreatureThatAttacked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	declareAttack(t, g, opp.ID, a, b)
	advanceTo(t, g, game.StepPostcombatMain)
	// Two attackers: {3}{R} costs {1}{R}. One mana is not enough.
	if err := spmB13CastJanes(t, g, "{R}"); err == nil {
		t.Fatal("cast The Mary Janes for {R} after two creatures attacked")
	}
}

func TestSpmB13MaryJanesPaysTheReducedPrice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	declareAttack(t, g, opp.ID, a, b)
	advanceTo(t, g, game.StepPostcombatMain)
	if err := spmB13CastJanes(t, g, "{R}{C}"); err != nil {
		t.Fatalf("cast for {1}{R} after two attackers: %v", err)
	}
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "The Mary Janes") == uuid.Nil {
		t.Error("The Mary Janes did not resolve")
	}
}

func TestSpmB13MaryJanesCostsFullPriceWhenNothingAttacked(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	if err := spmB13CastJanes(t, g, "{R}{C}{C}"); err == nil {
		t.Fatal("cast The Mary Janes for three mana with no attackers")
	}
}

func TestSpmB13MaryJanesNeverCostsLessThanItsColour(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	var ids []uuid.UUID
	for i := 0; i < 5; i++ {
		ids = append(ids, b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2))
	}
	advanceToMain(t, g)
	declareAttack(t, g, opp.ID, ids...)
	advanceTo(t, g, game.StepPostcombatMain)
	if err := spmB13CastJanes(t, g, "{R}"); err != nil {
		t.Fatalf("cast for {R} after five attackers: %v", err)
	}
}

// --- Twisted Spider-Clone ---

func TestSpmB13TwistedSpiderCloneGrowsOnlyCreaturesWithCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	grown := b12Creature(g, me.ID, "Grown", "Creature — Bear", 2, 2)
	plain := b12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	setCountersOn(g, grown, map[string]int{game.CounterPlusOne: 2})
	setCountersOn(g, theirs, map[string]int{game.CounterPlusOne: 1})
	id, err := spmB4Cast(t, g, "Twisted Spider-Clone", "Creature — Spider Human", spmB13CloneOracle, "{3}{G}", 4, 4, 0)
	if err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if plusOnes(g, grown) != 3 {
		t.Errorf("the creature with counters has %d, want 3", plusOnes(g, grown))
	}
	if plusOnes(g, plain) != 0 {
		t.Errorf("a creature with no counter got %d", plusOnes(g, plain))
	}
	if plusOnes(g, theirs) != 1 {
		t.Errorf("an opponent's creature changed to %d", plusOnes(g, theirs))
	}
	if plusOnes(g, id) != 0 {
		t.Errorf("the Clone itself got %d counters", plusOnes(g, id))
	}
}

// --- Venom Blast ---

func spmB13PlusTwoPower(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return spmB4Live(t, g, id).CurrentPower()
}

func TestSpmB13VenomBlastGrowsTheCreatureThenItDealsDamageEqualToItsPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 4, 4)
	castCatalogSpell(t, g, "Venom Blast", "Sorcery", spmB13BlastOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: victim},
	})
	passPriorityAroundTable(t, g)
	if plusOnes(g, mine) != 2 {
		t.Errorf("the creature has %d counters, want 2", plusOnes(g, mine))
	}
	if got := spmB13PlusTwoPower(t, g, mine); got != 4 {
		t.Errorf("power %d, want 4", got)
	}
	// 4 damage (2 base + 2 counters) is lethal to a 4/4: the counters came first.
	if !opp.Graveyard.Contains(victim) {
		t.Error("the victim survived: the bite did not read the grown power")
	}
	if spmB4Live(t, g, mine).DamageMarked != 0 {
		t.Error("the biter took damage: it is a one-sided bite")
	}
}

func TestSpmB13VenomBlastWithNoSecondTargetOnlyPutsCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Venom Blast", "Sorcery", spmB13BlastOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine},
	})
	passPriorityAroundTable(t, g)
	if plusOnes(g, mine) != 2 {
		t.Errorf("the creature has %d counters, want 2", plusOnes(g, mine))
	}
}

func TestSpmB13VenomBlastRefusesTheSameCreatureTwiceAndAnOpposingFirstTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Venom Blast", TypeLine: "Sorcery", OracleID: spmB13BlastOracle,
		ManaCost: "{2}{G}{G}", Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	floatMana(t, g, me, "{G}{G}{C}{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: mine},
	}}); err == nil {
		t.Error("one creature was both targets")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: foe},
	}}); err == nil {
		t.Error("an opposing creature was the creature you control")
	}
}

// --- Venom, Deadly Devourer ---

func spmB13Devourer(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	return b12Push(g, g.Seats[0].ID, "Venom, Deadly Devourer", "Legendary Creature — Symbiote Villain", spmB13DevourerOracl, 4, 4)
}

func TestSpmB13DevourerHasKeywords(t *testing.T) {
	g := newCatalogGame(t)
	assertKeywords(t, g, spmB13Devourer(t, g), "vigilance", "menace")
}

func TestSpmB13DevourerExilesACreatureCardThenPutsItsToughnessInCountersOnASymbiote(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	venom := spmB13Devourer(t, g)
	spawn := b12Creature(g, me.ID, "Spawn", "Creature — Symbiote Villain", 3, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	dead := uuid.New()
	opp.Graveyard.PushTop(game.Card{
		InstanceID: dead, Name: "Big Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 5,
		Owner: opp.ID, Controller: opp.ID,
	})
	land := uuid.New()
	opp.Graveyard.PushTop(game.Card{
		InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID,
	})
	floatMana(t, g, me, "{C}{C}{C}")
	// The land is not a creature card.
	if err := g.ActivateCatalogAbility(me.ID, venom, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}}}); err == nil {
		t.Fatal("a land card was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, venom, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: dead}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	// The reflexive trigger asks for a Symbiote; the Bear is not one.
	b04WaitForPick(t, g, me.ID)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
			t.Fatal("a non-Symbiote was a legal target")
		}
	}
	pickCard(t, g, me.ID, spawn)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(dead) || opp.Graveyard.Contains(dead) {
		t.Error("the creature card was not exiled")
	}
	if plusOnes(g, spawn) != 5 {
		t.Errorf("the Symbiote has %d counters, want 5 (the Wall's toughness)", plusOnes(g, spawn))
	}
	if plusOnes(g, bear) != 0 {
		t.Error("the Bear got counters")
	}
}

func TestSpmB13DevourerCanTargetItselfAsTheSymbiote(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	venom := spmB13Devourer(t, g)
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: dead, Name: "Bear Cub", TypeLine: "Creature — Bear", Power: 1, Toughness: 3,
		Owner: me.ID, Controller: me.ID,
	})
	floatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, venom, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: dead}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	if latestPickTarget(g, me.ID) != nil {
		pickCard(t, g, me.ID, venom)
	}
	passPriorityAroundTable(t, g)
	if plusOnes(g, venom) != 3 {
		t.Errorf("Venom has %d counters, want 3", plusOnes(g, venom))
	}
}

func TestSpmB13DevourerRefusesTooLittleMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	venom := spmB13Devourer(t, g)
	me.Graveyard.PushTop(game.Card{
		InstanceID: uuid.New(), Name: "Bear Cub", TypeLine: "Creature — Bear", Power: 1, Toughness: 3,
		Owner: me.ID, Controller: me.ID,
	})
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, venom, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: me.Graveyard.Cards[0].InstanceID}}}); err == nil {
		t.Fatal("activated {3} with two mana")
	}
}

// --- Venom, Eddie Brock ---

func spmB13Eddie(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	return b12Push(g, g.Seats[0].ID, "Venom, Eddie Brock", "Legendary Creature — Symbiote Villain", spmB13EddieOracle, 6, 4)
}

func TestSpmB13EddieBrockGrowsWhenAnotherCreatureDiesAndDrawsForAVillain(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eddie := spmB13Eddie(t, g)
	assertKeywords(t, g, eddie, "menace")
	seedLibrary(me, "One", "Two", "Three")
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	killCreature(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if plusOnes(g, eddie) != 1 {
		t.Errorf("Venom has %d counters, want 1", plusOnes(g, eddie))
	}
	if me.Hand.Size() != hand {
		t.Errorf("a non-Villain drew a card: hand %d -> %d", hand, me.Hand.Size())
	}
	villain := b12Creature(g, opp.ID, "Goon", "Creature — Human Villain", 2, 2)
	killCreature(t, g, opp.ID, villain)
	passPriorityAroundTable(t, g)
	if plusOnes(g, eddie) != 2 {
		t.Errorf("Venom has %d counters, want 2", plusOnes(g, eddie))
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("a Villain died: hand %d -> %d, want +1", hand, me.Hand.Size())
	}
}

func TestSpmB13EddieBrockCountsYourOwnVillainAndIgnoresItsOwnDeath(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	eddie := spmB13Eddie(t, g)
	seedLibrary(me, "One", "Two", "Three")
	goon := b12Creature(g, me.ID, "Goon", "Creature — Human Villain", 2, 2)
	hand := me.Hand.Size()
	killCreature(t, g, me.ID, goon)
	passPriorityAroundTable(t, g)
	if plusOnes(g, eddie) != 1 || me.Hand.Size() != hand+1 {
		t.Errorf("own Villain died: counters %d, hand %d -> %d", plusOnes(g, eddie), hand, me.Hand.Size())
	}
	killCreature(t, g, me.ID, eddie)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("Venom's own death drew a card: hand %d", me.Hand.Size())
	}
}

func TestSpmB13EddieBrockReadsTheVillainTypeAsItDied(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eddie := spmB13Eddie(t, g)
	seedLibrary(me, "One", "Two", "Three")
	// A token Villain dies and ceases to exist; the type is still read.
	tok := b12Creature(g, opp.ID, "Villain Token", "Token Creature — Human Villain", 1, 1)
	hand := me.Hand.Size()
	killCreature(t, g, opp.ID, tok)
	passPriorityAroundTable(t, g)
	if plusOnes(g, eddie) != 1 || me.Hand.Size() != hand+1 {
		t.Errorf("token Villain died: counters %d, hand %d -> %d", plusOnes(g, eddie), hand, me.Hand.Size())
	}
}

// --- Magik, Illyana Rasputin // Darkchild, Demon Uncaged ---

func spmB13MagikCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(), OracleID: magikOracle, Layout: game.LayoutModalDFC,
		Owner: owner, Controller: owner,
		Faces: []game.Face{
			{Name: "Magik, Illyana Rasputin", TypeLine: "Legendary Creature — Mutant Sorcerer Hero", ManaCost: "{1}{B}{R}", Colors: []string{"B", "R"}, Power: 3, Toughness: 3},
			{Name: "Darkchild, Demon Uncaged", TypeLine: "Legendary Creature — Mutant Demon Sorcerer", ManaCost: "{3}{B}{R}", Colors: []string{"B", "R"}, Power: 4, Toughness: 4},
		},
	}
	c.SetFace(0)
	return c
}

func spmB13PushMagik(g *game.Game, owner uuid.UUID, face int) uuid.UUID {
	c := spmB13MagikCard(owner)
	c.SetFace(face)
	return pushBattlefieldCardWithTimestamp(g, c)
}

func TestSpmB13MagikExilesTheTopCardAtUpkeepAndYouMayPlayIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spmB13PushMagik(g, me.ID, 0)
	top := topCard(me, "Top Bear", "Creature — Bear", "{G}")
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(top) {
		t.Fatal("the top card was not exiled at upkeep")
	}
	if !permissionLive(g, exiledPermission(g, top), me.ID) {
		t.Error("the exiled card is not playable this turn")
	}
}

func TestSpmB13MagikDoesNotExileAtAnOpponentsUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spmB13PushMagik(g, me.ID, 0)
	top := topCard(me, "Top Bear", "Creature — Bear", "{G}")
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if g.Exile.Contains(top) {
		t.Error("Magik exiled a card at an opponent's upkeep")
	}
}

func TestSpmB13MagikMayTransformAfterTwoSacrifices(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	magik := spmB13PushMagik(g, me.ID, 0)
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 1, 1)
	b := b12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
	advanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(a) })
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(b) })
	advanceToEndStep(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	card := battlefieldCardFor(g, magik)
	if card == nil || card.ActiveFace != 1 || card.Name != "Darkchild, Demon Uncaged" {
		t.Fatalf("Magik did not transform: %+v", card)
	}
}

func TestSpmB13MagikMayDeclineToTransform(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	magik := spmB13PushMagik(g, me.ID, 0)
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 1, 1)
	b := b12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
	advanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(a) })
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(b) })
	advanceToEndStep(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if card := battlefieldCardFor(g, magik); card == nil || card.ActiveFace != 0 {
		t.Fatalf("Magik transformed after the player declined: %+v", card)
	}
}

func TestSpmB13MagikDoesNotTriggerAfterOnlyOneSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	magik := spmB13PushMagik(g, me.ID, 0)
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 1, 1)
	advanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(a) })
	advanceToEndStep(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt {
			t.Fatal("Magik asked to transform after one sacrifice")
		}
	}
	passPriorityAroundTable(t, g)
	if card := battlefieldCardFor(g, magik); card == nil || card.ActiveFace != 0 {
		t.Fatalf("Magik transformed after one sacrifice: %+v", card)
	}
}

func TestSpmB13DarkchildExilesUntilANonlandCardAndYouMayPlayThemAll(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spmB13PushMagik(g, me.ID, 1)
	// Library order: bottom ... top. The walk takes land, land, spell.
	deep := topCard(me, "Deep Bear", "Creature — Bear", "{G}")
	spell := topCard(me, "Spell", "Instant", "{U}")
	land2 := topCard(me, "Island Two", "Basic Land — Island", "")
	land1 := topCard(me, "Island One", "Basic Land — Island", "")
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"first land": land1, "second land": land2, "the nonland card": spell} {
		if !g.Exile.Contains(id) {
			t.Errorf("%s was not exiled", name)
			continue
		}
		if !permissionLive(g, exiledPermission(g, id), me.ID) {
			t.Errorf("%s is not playable this turn", name)
		}
	}
	if g.Exile.Contains(deep) {
		t.Error("the walk went past the first nonland card")
	}
}

func TestSpmB13DarkchildPaysFourLifeForIndestructibleAndTaps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dark := spmB13PushMagik(g, me.ID, 1)
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, dark, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life-4 {
		t.Errorf("life %d -> %d, want -4", life, me.Life)
	}
	if !hasKeywordOnBattlefield(t, g, dark, "indestructible") {
		t.Error("Darkchild did not gain indestructible")
	}
	if c := battlefieldCardFor(g, dark); c == nil || !c.Tapped {
		t.Error("Darkchild is not tapped")
	}
}
