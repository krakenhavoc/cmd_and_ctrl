package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b192_test.go — S82 GAP batch 192 (#5948): the Tyranids of Warhammer
// 40,000 Commander. Each test plays the real catalog entry.

const (
	gapB192HormagauntOracle = "cc965557-f69e-4689-8791-d3618e4ffbfa"
	gapB192LictorOracle     = "3fdf064d-a5be-4925-87b1-3547997f0840"
	gapB192NexosOracle      = "4b9a2d17-205a-49cf-98c3-fd833015da18"
	gapB192OldOneEyeOracle  = "86ef2ceb-b7e6-4451-92ec-8528ea191d3a"
	gapB192GenestealerOracl = "3d4c7f1a-deca-4bae-90c7-68d623da400e"
	gapB192SporocystOracle  = "ed58a617-cf86-4c2e-8044-1af37b1a1465"
	gapB192TermagantOracle  = "e7d8b8ea-07d7-4b6c-aa49-6a2ad185a304"
	gapB192TervigonOracle   = "a82842b8-cc2e-44fa-b869-7bccc16ca74f"
	gapB192ToxicreneOracle  = "fbfe98f0-110b-4d4b-9240-2767b08acf68"
	gapB192TyrantGuardOracl = "560e6fd6-c061-44ba-b5c5-814d6a14e469"
	gapB192JackalOracle     = "8be11473-6edf-4f45-b541-11858655fb5f"
	gapB192DeathleaperOracl = "8e22495d-a921-4c91-ada3-79098851e5d4"
)

func TestGapB192CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB192HormagauntOracle: "Hormagaunt Horde",
		gapB192LictorOracle:     "Lictor",
		gapB192NexosOracle:      "Nexos",
		gapB192OldOneEyeOracle:  "Old One Eye",
		gapB192GenestealerOracl: "Purestrain Genestealer",
		gapB192SporocystOracle:  "Sporocyst",
		gapB192TermagantOracle:  "Termagant Swarm",
		gapB192TervigonOracle:   "Tervigon",
		gapB192ToxicreneOracle:  "Toxicrene",
		gapB192TyrantGuardOracl: "Tyrant Guard",
		gapB192JackalOracle:     "Atalan Jackal",
		gapB192DeathleaperOracl: "Deathleaper, Terror Weapon",
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

// gapB192Named counts the permanents on the battlefield with a name.
func gapB192Named(g *game.Game, name string) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name {
				n++
			}
		}
	})
	return n
}

// gapB192Land puts a basic land card on top of a player's library and returns it.
func gapB192Land(p *game.Player) uuid.UUID {
	return effectstest.StapleLibraryCard(p, "Forest", "Basic Land — Forest")
}

// --- Ravenous (shared by five of the cards) -------------------------------

func TestGapB192RavenousEntersWithXCountersAndDrawsAtFive(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, cost string
		x, counters, draws int
	}{
		{"Hormagaunt Horde", gapB192HormagauntOracle, "{X}{G}", 2, 2, 0},
		{"Hormagaunt Horde", gapB192HormagauntOracle, "{X}{G}", 5, 5, 1},
		{"Tervigon", gapB192TervigonOracle, "{X}{1}{G}", 4, 4, 0},
		{"Termagant Swarm", gapB192TermagantOracle, "{X}{G}", 5, 5, 1},
		{"Tyrant Guard", gapB192TyrantGuardOracl, "{X}{2}{G}", 3, 3, 0},
		{"Tyrant Guard", gapB192TyrantGuardOracl, "{X}{2}{G}", 6, 6, 1},
		{"Sporocyst", gapB192SporocystOracle, "{X}{X}{G}", 5, 5, 1},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		// Sporocyst's enters trigger searches; an empty library keeps this about ravenous.
		hand := len(me.Hand.Cards)
		id := effectstest.CastXSpell(t, g, tc.name, "Creature — Tyranid", tc.oracle, tc.cost, tc.x, nil)
		// Sporocyst has a second enters trigger, so its controller orders the two.
		effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.TakeSearchPrompts(t, g)
		if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != tc.counters {
			t.Errorf("%s X=%d: %d +1/+1 counters, want %d", tc.name, tc.x, got, tc.counters)
		}
		if got := len(me.Hand.Cards) - hand; got != tc.draws {
			t.Errorf("%s X=%d: drew %d cards, want %d", tc.name, tc.x, got, tc.draws)
		}
	}
}

// --- Purestrain Genestealer ----------------------------------------------

func TestGapB192GenestealerEntersWithTwoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Purestrain Genestealer", TypeLine: "Creature — Tyranid", OracleID: gapB192GenestealerOracl, Power: 1, Toughness: 1})
	if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters, want 2", got)
	}
}

// gapB192Attack declares attackers and locks the declaration in, leaving the
// "whenever this attacks" triggers on the stack.
func gapB192Attack(t *testing.T, g *game.Game, defender uuid.UUID, attackers ...uuid.UUID) {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range attackers {
		if err := g.DeclareAttacker(id, defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
}

func gapB192Genestealer(g *game.Game, me *game.Player, counters int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Purestrain Genestealer", TypeLine: "Creature — Tyranid", OracleID: gapB192GenestealerOracl,
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterPlusOne: counters},
	})
}

func TestGapB192GenestealerRemovesACounterToFetchABasicTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gen := gapB192Genestealer(g, me, 2)
	forest := gapB192Land(me)
	gapB192Attack(t, g, opp.ID, gen)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	if got := effectstest.CounterCount(g, gen, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters left, want 1", got)
	}
	land, ok := effectstest.AangCardOnBF(g, forest)
	if !ok || !land.Tapped {
		t.Errorf("the Forest is on the battlefield: %v, tapped: %v; want both", ok, land.Tapped)
	}
}

func TestGapB192GenestealerDecliningKeepsTheCounterAndFetchesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gen := gapB192Genestealer(g, me, 2)
	forest := gapB192Land(me)
	gapB192Attack(t, g, opp.ID, gen)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, gen, game.CounterPlusOne); got != 2 {
		t.Errorf("%d counters, want 2 after declining", got)
	}
	if _, ok := effectstest.AangCardOnBF(g, forest); ok {
		t.Error("a land was fetched after declining")
	}
}

func TestGapB192GenestealerWithoutCountersIsNotOffered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gen := gapB192Genestealer(g, me, 0)
	forest := gapB192Land(me)
	gapB192Attack(t, g, opp.ID, gen)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		t.Error("asked to remove a counter it does not have")
	}
	if _, ok := effectstest.AangCardOnBF(g, forest); ok {
		t.Error("a land was fetched with no counter removed")
	}
}

// --- Atalan Jackal ---------------------------------------------------------

func TestGapB192JackalFetchesABasicTappedOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jackal := effectstest.B12Push(g, me.ID, "Atalan Jackal", "Creature — Human Tyranid Scout", gapB192JackalOracle, 2, 2)
	forest := gapB192Land(me)
	life := opp.Life
	effectstest.AttackWith(t, g, opp.ID, jackal)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	if opp.Life != life-2 {
		t.Fatalf("opponent life %d -> %d, want 2 combat damage", life, opp.Life)
	}
	land, ok := effectstest.AangCardOnBF(g, forest)
	if !ok || !land.Tapped {
		t.Errorf("Forest on the battlefield: %v, tapped: %v; want both", ok, land.Tapped)
	}
}

func TestGapB192JackalDoesNotSearchForAnotherCreaturesDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Atalan Jackal", "Creature — Human Tyranid Scout", gapB192JackalOracle, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	forest := gapB192Land(me)
	effectstest.AttackWith(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	if _, ok := effectstest.AangCardOnBF(g, forest); ok {
		t.Error("another creature's combat damage searched for a land")
	}
}

// --- Lictor ----------------------------------------------------------------

func gapB192EnterLictor(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Lictor", TypeLine: "Creature — Tyranid", OracleID: gapB192LictorOracle, Power: 3, Toughness: 3})
	effectstest.PassPriorityAroundTable(t, g)
}

func TestGapB192LictorMakesAWarriorAfterAnOpponentsCreatureEntered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.EnterCreature(t, g, opp.ID, "Their Bear", 2, 2)
	gapB192EnterLictor(t, g, me)
	if n := gapB192Named(g, "Tyranid Warrior"); n != 1 {
		t.Fatalf("%d Tyranid Warrior tokens, want 1", n)
	}
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Tyranid Warrior" {
				if c.Power != 3 || c.Toughness != 3 || !c.IsToken() || !game.HasKeyword(&c, "trample") || c.Controller != me.ID {
					t.Errorf("token %+v, want a 3/3 trample token under my control", c)
				}
			}
		}
	})
}

func TestGapB192LictorMakesNothingWhenOnlyMyCreaturesEntered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.EnterCreature(t, g, me.ID, "My Bear", 2, 2)
	gapB192EnterLictor(t, g, me)
	if n := gapB192Named(g, "Tyranid Warrior"); n != 0 {
		t.Errorf("%d Warrior tokens, want none", n)
	}
}

func TestGapB192LictorCountsAnOpponentsCreatureThatHasSinceLeft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.EnterCreature(t, g, opp.ID, "Their Bear", 2, 2)
	effectstest.ExileCreature(t, g, bear)
	gapB192EnterLictor(t, g, me)
	if n := gapB192Named(g, "Tyranid Warrior"); n != 1 {
		t.Errorf("%d Warrior tokens, want 1: the creature entered this turn even though it has left", n)
	}
}

// --- Nexos -----------------------------------------------------------------

func TestGapB192NexosBasicLandsPayOnlyCostsWithX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Nexos", "Creature — Human Tyranid Advisor", gapB192NexosOracle, 2, 2)
	forest := effectstest.B12Push(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	g.WithWriteLock(func() { g.BumpLayerVersionForTest() })
	effectstest.ToMainForCost(t, g)
	granted := -1
	for i, ab := range game.ManaAbilitiesForCard(effectstest.LayeredCard(t, g, forest)) {
		if ab.Produced == "{C}{C}" {
			granted = i
		}
	}
	if granted < 0 {
		t.Fatalf("the Forest has no {C}{C} ability: %v", effectstest.LtProduced(t, g, forest))
	}
	me.ManaPool.EmptyPool()
	if err := g.ActivateManaAbility(me.ID, forest, granted, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for {C}{C}: %v", err)
	}
	if me.ManaPool.Total() != 2 {
		t.Fatalf("pool %+v, want two mana", me.ManaPool)
	}
	effectstest.FillPoolColored(me, "G", 1)
	plain := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: plain, Name: "Plain", TypeLine: "Sorcery", ManaCost: "{2}{G}", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, plain, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("the {X}-only mana paid a cost without {X}")
	}
	xs := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: xs, Name: "X Spell", TypeLine: "Sorcery", ManaCost: "{X}{G}", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, xs, game.CastSpellParams{Strict: true, XValue: 2}); err != nil {
		t.Fatalf("the {X}-only mana could not pay an {X} spell: %v", err)
	}
}

func TestGapB192NexosDoesNotGrantItToAnOpponentsOrNonbasicLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Nexos", "Creature — Human Tyranid Advisor", gapB192NexosOracle, 2, 2)
	theirs := effectstest.B12Push(g, opp.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	nonbasic := effectstest.B12Push(g, me.ID, "Temple Garden", "Land — Forest Plains", "", 0, 0)
	for name, id := range map[string]uuid.UUID{"an opponent's basic land": theirs, "a nonbasic land": nonbasic} {
		for _, shape := range effectstest.LtProduced(t, g, id) {
			if shape == "{C}{C}" {
				t.Errorf("%s got Nexos's {C}{C} ability", name)
			}
		}
	}
}

// --- Old One Eye -----------------------------------------------------------

func TestGapB192OldOneEyeMakesAFiveFiveAndGivesTrampleToOthers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	eye := effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Old One Eye", TypeLine: "Legendary Creature — Tyranid", OracleID: gapB192OldOneEyeOracle, Power: 6, Toughness: 6})
	effectstest.PassPriorityAroundTable(t, g)
	if n := gapB192Named(g, "Tyranid"); n != 1 {
		t.Fatalf("%d Tyranid tokens, want 1", n)
	}
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Tyranid" && (c.Power != 5 || c.Toughness != 5 || !c.IsToken()) {
				t.Errorf("token %+v, want a 5/5 token", c)
			}
		}
	})
	if !effectstest.HasKeywordOf(g, mine, "trample") {
		t.Error("my other creature lacks trample")
	}
	if effectstest.HasKeywordOf(g, theirs, "trample") {
		t.Error("an opponent's creature has trample")
	}
	_ = eye
}

func gapB192OldOneEyeInGraveyard(g *game.Game, me *game.Player) uuid.UUID {
	return effectstest.DcGraveyardCard(g, me, "Old One Eye", "Legendary Creature — Tyranid", gapB192OldOneEyeOracle, 6, 6)
}

func gapB192HandCards(me *game.Player, n int) []uuid.UUID {
	var ids []uuid.UUID
	for i := 0; i < n; i++ {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Spare", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		ids = append(ids, id)
	}
	return ids
}

func TestGapB192OldOneEyeReturnsFromTheGraveyardForTwoDiscards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	eye := gapB192OldOneEyeInGraveyard(g, me)
	spare := gapB192HandCards(me, 2)
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID, spare...)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HandHasID(me, eye) {
		t.Errorf("Old One Eye is in %v, want the hand", effectstest.B12ZoneOf(g, eye))
	}
}

func TestGapB192OldOneEyeStaysWhenYouDiscardNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	eye := gapB192OldOneEyeInGraveyard(g, me)
	gapB192HandCards(me, 2)
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.B12ZoneOf(g, eye); z != game.ZoneGraveyard {
		t.Errorf("Old One Eye is in %v, want the graveyard", z)
	}
}

func TestGapB192OldOneEyeRefusesASingleDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	eye := gapB192OldOneEyeInGraveyard(g, me)
	spare := gapB192HandCards(me, 2)
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.DiscardChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no discard prompt")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, spare[:1]); err == nil {
		t.Error("discarding one card was accepted for a two-or-none clause")
	}
	if z := effectstest.B12ZoneOf(g, eye); z != game.ZoneGraveyard {
		t.Errorf("Old One Eye is in %v, want the graveyard", z)
	}
}

// --- Deathleaper -----------------------------------------------------------

func TestGapB192DeathleaperGivesDoubleStrikeToCreaturesThatEnteredThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	old := effectstest.B12Creature(g, me.ID, "Old Bear", "Creature — Bear", 2, 2)
	effectstest.B12Push(g, me.ID, "Deathleaper, Terror Weapon", "Legendary Creature — Tyranid", gapB192DeathleaperOracl, 3, 3)
	fresh := effectstest.EnterCreature(t, g, me.ID, "Fresh Bear", 2, 2)
	theirs := effectstest.EnterCreature(t, g, opp.ID, "Their Fresh Bear", 2, 2)
	if !effectstest.HasKeywordOf(g, fresh, "double strike") {
		t.Error("a creature that entered this turn lacks double strike")
	}
	if effectstest.HasKeywordOf(g, old, "double strike") {
		t.Error("a creature that did not enter this turn has double strike")
	}
	if effectstest.HasKeywordOf(g, theirs, "double strike") {
		t.Error("an opponent's creature has double strike")
	}
}

// --- Tervigon, Termagant Swarm ---------------------------------------------

func TestGapB192TervigonMakesTokensEqualToTheDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	terv := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Tervigon", TypeLine: "Creature — Tyranid", OracleID: gapB192TervigonOracle,
		Power: 0, Toughness: 0, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterPlusOne: 3},
	})
	effectstest.AttackWith(t, g, opp.ID, terv)
	effectstest.PassPriorityAroundTable(t, g)
	if n := gapB192Named(g, "Tyranid"); n != 3 {
		t.Errorf("%d Tyranid tokens, want 3", n)
	}
}

func TestGapB192TermagantSwarmDiesIntoTokensEqualToItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	swarm := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Termagant Swarm", TypeLine: "Creature — Tyranid", OracleID: gapB192TermagantOracle,
		Power: 0, Toughness: 0, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterPlusOne: 4},
	})
	effectstest.P1Destroy(t, g, swarm)
	effectstest.PassPriorityAroundTable(t, g)
	if n := gapB192Named(g, "Tyranid"); n != 4 {
		t.Errorf("%d Tyranid tokens, want 4 (its power as it died)", n)
	}
}

// --- Sporocyst -------------------------------------------------------------

func TestGapB192SporocystFetchesUpToXBasicLandsTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lands := []uuid.UUID{gapB192Land(me), gapB192Land(me), gapB192Land(me)}
	effectstest.CastXSpell(t, g, "Sporocyst", "Creature — Tyranid", gapB192SporocystOracle, "{X}{X}{G}", 2, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	n := 0
	for _, id := range lands {
		if c, ok := effectstest.AangCardOnBF(g, id); ok {
			n++
			if !c.Tapped {
				t.Error("a fetched land entered untapped")
			}
		}
	}
	if n != 2 {
		t.Errorf("%d basic lands fetched, want exactly X = 2", n)
	}
}

// --- Hormagaunt Horde ------------------------------------------------------

func TestGapB192HormagauntReturnsForALandfallPayment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	horde := effectstest.DcGraveyardCard(g, me, "Hormagaunt Horde", "Creature — Tyranid", gapB192HormagauntOracle, 1, 1)
	effectstest.FillPoolColored(me, "G", 3)
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HandHasID(me, horde) {
		t.Errorf("Hormagaunt Horde is in %v, want the hand", effectstest.B12ZoneOf(g, horde))
	}
}

func TestGapB192HormagauntStaysWhenYouCannotOrDoNotPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	horde := effectstest.DcGraveyardCard(g, me, "Hormagaunt Horde", "Creature — Tyranid", gapB192HormagauntOracle, 1, 1)
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.B12ZoneOf(g, horde); z != game.ZoneGraveyard {
		t.Errorf("Hormagaunt Horde is in %v with nothing paid, want the graveyard", z)
	}
}

func TestGapB192HormagauntIgnoresAnOpponentsLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	horde := effectstest.DcGraveyardCard(g, me, "Hormagaunt Horde", "Creature — Tyranid", gapB192HormagauntOracle, 1, 1)
	effectstest.FillPoolColored(me, "G", 3)
	effectstest.EnterCard(t, g, opp.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("an opponent's land offered the payment")
	}
	if z := effectstest.B12ZoneOf(g, horde); z != game.ZoneGraveyard {
		t.Errorf("Hormagaunt Horde is in %v, want the graveyard", z)
	}
}

// --- Tyrant Guard ----------------------------------------------------------

func TestGapB192TyrantGuardShieldsCreaturesWithCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guard := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Tyrant Guard", TypeLine: "Creature — Tyranid", OracleID: gapB192TyrantGuardOracl,
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	counted := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Counted Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterPlusOne: 1},
	})
	bare := effectstest.B12Creature(g, me.ID, "Bare Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Counted Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Owner: opp.ID, Controller: opp.ID, Counters: map[string]int{game.CounterPlusOne: 1},
	})
	if err := g.ActivateCatalogAbility(me.ID, guard, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Shieldwall: %v", err)
	}
	if z := effectstest.B12ZoneOf(g, guard); z != game.ZoneGraveyard {
		t.Fatalf("the Guard is in %v, want the graveyard (sacrificed as a cost)", z)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, kw := range []string{"hexproof", "indestructible"} {
		if !effectstest.HasKeywordOf(g, counted, kw) {
			t.Errorf("my creature with a counter lacks %s", kw)
		}
		if effectstest.HasKeywordOf(g, bare, kw) {
			t.Errorf("my creature with no counter has %s", kw)
		}
		if effectstest.HasKeywordOf(g, theirs, kw) {
			t.Errorf("an opponent's creature has %s", kw)
		}
	}
}

// --- Toxicrene -------------------------------------------------------------

func TestGapB192ToxicreneLandsTapOnlyForAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Toxicrene", "Creature — Tyranid", gapB192ToxicreneOracle, 2, 4)
	forest := effectstest.B12Push(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	theirs := effectstest.B12Push(g, opp.ID, "Island", "Basic Land — Island", "", 0, 0)
	for _, land := range []uuid.UUID{forest, theirs} {
		if got := effectstest.LtProduced(t, g, land); len(got) != 1 || got[0] != "{W|U|B|R|G}" {
			t.Errorf("a land taps for %v, want only any colour", got)
		}
	}
}

func TestGapB192ToxicreneLandsRecoverWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	crene := effectstest.B12Push(g, me.ID, "Toxicrene", "Creature — Tyranid", gapB192ToxicreneOracle, 2, 4)
	forest := effectstest.B12Push(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	effectstest.P1Destroy(t, g, crene)
	if got := effectstest.LtProduced(t, g, forest); len(got) != 1 || got[0] != "{G}" {
		t.Errorf("the Forest taps for %v after Toxicrene left, want only {G}", got)
	}
}
