package testsso

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestSocB17EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		effectstest.SocB17LocustSprayOracle: "Locust Spray",
		effectstest.SocB17PrismaticOracle:   "Prismatic Ending",
		effectstest.SocB17RepelOracle:       "Repel Calamity",
		effectstest.SocB17SleightOracle:     "Sleight of Hand",
		effectstest.SocB17SmallpoxOracle:    "Smallpox",
		effectstest.SocB17StargazeOracle:    "Stargaze",
		effectstest.SocB17TremorsOracle:     "Subterranean Tremors",
		effectstest.SocB17StallionOracle:    "Colorstorm Stallion",
		effectstest.SocB17GreatHallOracle:   "Great Hall of the Biblioplex",
		effectstest.SocB17InkwrightOracle:   "Informed Inkwright",
	}
	for oracle, name := range want {
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

// --- Locust Spray -------------------------------------------------------------

func TestSocB17LocustSprayShrinksUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	mouse := effectstest.PushVanillaCreature(g, opp.ID, "Mouse", 1, 1)

	effectstest.CastCatalogSpell(t, g, "Locust Spray", "Instant", effectstest.SocB17LocustSprayOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.S58p6View(t, g, bear)
	if c.CurrentPower() != 1 || c.CurrentToughness() != 1 {
		t.Fatalf("the Bear is %d/%d, want 1/1", c.CurrentPower(), c.CurrentToughness())
	}

	effectstest.CastCatalogSpell(t, g, "Locust Spray", "Instant", effectstest.SocB17LocustSprayOracle, effectstest.CardRefs(mouse))
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(mouse) {
		t.Error("a 1/1 survived -1/-1")
	}

	// The shrink ends with the turn.
	turn := g.Turn.Seq
	for i := 0; i < 60 && g.Turn.Seq == turn; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	c = effectstest.S58p6View(t, g, bear)
	if c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("the Bear is %d/%d next turn, want 2/2", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestSocB17LocustSprayCyclesForBlack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.SocB17Hand(me, "Locust Spray", "Instant", "{B}", effectstest.SocB17LocustSprayOracle)
	hand := me.Hand.Size()

	effectstest.FloatForTest(g, me, "R")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("cycled with only red mana")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FloatForTest(g, me, "B")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("cycling {B}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(id) || !me.Graveyard.Contains(id) {
		t.Error("the cycled card is not in the graveyard")
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d: one card out, one drawn", me.Hand.Size(), hand)
	}
}

// --- Repel Calamity -----------------------------------------------------------

func TestSocB17RepelCalamityDestroysBigPowerOrBigToughness(t *testing.T) {
	for _, tc := range []struct {
		name  string
		p, th int
	}{
		{"power 4", 4, 1},
		{"toughness 4", 1, 4},
		{"both", 6, 6},
	} {
		g := effectstest.NewCatalogGame(t)
		big := effectstest.PushVanillaCreature(g, g.Seats[1].ID, "Big", tc.p, tc.th)
		effectstest.CastCatalogSpell(t, g, "Repel Calamity", "Instant", effectstest.SocB17RepelOracle, effectstest.CardRefs(big))
		effectstest.PassPriorityAroundTable(t, g)
		g.RunStateChecksForTest()
		if g.Battlefield.Contains(big) {
			t.Errorf("%s: the creature survived", tc.name)
		}
	}
}

func TestSocB17RepelCalamityRefusesASmallCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	small := effectstest.PushVanillaCreature(g, g.Seats[1].ID, "Small", 3, 3)
	err := effectstest.CastCatalogSpellErr(t, g, "Repel Calamity", "Instant", effectstest.SocB17RepelOracle, effectstest.CardRefs(small))
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("cast at a 3/3 = %v, want ErrIllegalTarget", err)
	}
}

// --- Prismatic Ending ---------------------------------------------------------

// socB17CastPrismatic casts Prismatic Ending for X with exactly `pool`
// floating, at `target`.
func socB17CastPrismatic(t *testing.T, g *game.Game, x int, pool string, target uuid.UUID) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.SocB17Hand(me, "Prismatic Ending", "Sorcery", "{X}{W}", effectstest.SocB17PrismaticOracle)
	effectstest.FloatForTest(g, me, pool)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: x, Targets: effectstest.CardRefs(target)}); err != nil {
		t.Fatalf("cast Prismatic Ending X=%d with %s: %v", x, pool, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestSocB17PrismaticEndingCountsColorsSpent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	three := effectstest.SocB17Perm(g, opp.ID, "Three Drop", "Artifact", "", "{3}", 0, 0)
	socB17CastPrismatic(t, g, 2, "WUB", three)
	if g.Battlefield.Contains(three) || !g.Exile.Contains(three) {
		t.Error("three colours spent: the mana value 3 permanent was not exiled")
	}
}

func TestSocB17PrismaticEndingLeavesAPermanentThatCostsMoreThanTheColors(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	two := effectstest.SocB17Perm(g, opp.ID, "Two Drop", "Artifact", "", "{2}", 0, 0)
	// X = 2, but only white mana: one colour, so only mana value 1 or less.
	socB17CastPrismatic(t, g, 2, "WWW", two)
	if !g.Battlefield.Contains(two) || g.Exile.Contains(two) {
		t.Error("one colour spent: a mana value 2 permanent was exiled")
	}
}

func TestSocB17PrismaticEndingAtXZeroExilesMyFreeOrOnePermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	free := effectstest.SocB17Perm(g, opp.ID, "Free Thing", "Artifact", "", "", 0, 0)
	socB17CastPrismatic(t, g, 0, "W", free)
	if !g.Exile.Contains(free) {
		t.Error("X=0 with {W}: a mana value 0 permanent was not exiled")
	}

	g2 := effectstest.NewCatalogGame(t)
	one := effectstest.SocB17Perm(g2, g2.Seats[1].ID, "One Drop", "Artifact", "", "{1}", 0, 0)
	socB17CastPrismatic(t, g2, 0, "W", one)
	if !g2.Exile.Contains(one) {
		t.Error("X=0 with {W}: a mana value 1 permanent was not exiled")
	}

	g3 := effectstest.NewCatalogGame(t)
	two := effectstest.SocB17Perm(g3, g3.Seats[1].ID, "Two Drop", "Artifact", "", "{2}", 0, 0)
	socB17CastPrismatic(t, g3, 0, "W", two)
	if g3.Exile.Contains(two) {
		t.Error("X=0 with {W}: a mana value 2 permanent was exiled")
	}
}

func TestSocB17PrismaticEndingCannotTargetALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	land := effectstest.SocB17Perm(g, g.Seats[1].ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.SocB17Hand(me, "Prismatic Ending", "Sorcery", "{X}{W}", effectstest.SocB17PrismaticOracle)
	effectstest.FloatForTest(g, me, "W")
	err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 0, Targets: effectstest.CardRefs(land)})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("cast at a land = %v, want ErrIllegalTarget", err)
	}
}

// --- Sleight of Hand ----------------------------------------------------------

func TestSocB17SleightOfHandTakesOneOfTheTopTwoAndBottomsTheOther(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SeedLibrary(me, "A", "B", "C")
	hand := me.Hand.Size()

	effectstest.CastCatalogSpell(t, g, "Sleight of Hand", "Sorcery", effectstest.SocB17SleightOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no pick prompt")
	}
	if len(prompt.ChooseCards) != 2 || prompt.ChooseMin != 1 || prompt.ChooseMax != 1 {
		t.Fatalf("prompt over %d cards, %d..%d; want the top two, exactly one", len(prompt.ChooseCards), prompt.ChooseMin, prompt.ChooseMax)
	}
	for _, c := range prompt.ChooseCards {
		if c == ids[2] {
			t.Fatal("the third card was offered")
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, ids[1])
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[1]) || me.Hand.Size() != hand+1 {
		t.Error("B is not the one card added to the hand")
	}
	if bottom, err := me.Library.Bottom(); err != nil || bottom.InstanceID != ids[0] {
		t.Error("A is not on the bottom of the library")
	}
	if top, err := me.Library.Top(); err != nil || top.InstanceID != ids[2] {
		t.Error("C is not the next card")
	}
}

func TestSocB17SleightOfHandWithOneCardLeftTakesItWithoutAsking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Library.Cards = nil
	ids := effectstest.SeedLibrary(me, "Only")

	effectstest.CastCatalogSpell(t, g, "Sleight of Hand", "Sorcery", effectstest.SocB17SleightOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a library of one card asked for a pick")
	}
	if !me.Hand.Contains(ids[0]) {
		t.Error("the only card did not reach the hand")
	}
}

// --- Stargaze -----------------------------------------------------------------

func TestSocB17StargazeTakesXOfTwiceXAndLosesXLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F")
	hand, life := me.Hand.Size(), me.Life

	effectstest.CastXSpell(t, g, "Stargaze", "Sorcery", effectstest.SocB17StargazeOracle, "{X}{B}{B}", 2, nil)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no pick prompt")
	}
	if len(prompt.ChooseCards) != 4 || prompt.ChooseMin != 2 || prompt.ChooseMax != 2 {
		t.Fatalf("prompt over %d cards, %d..%d; want the top four, exactly two", len(prompt.ChooseCards), prompt.ChooseMin, prompt.ChooseMax)
	}
	if me.Life != life {
		t.Fatalf("life %d before the pick, want %d: the loss waits behind the pick", me.Life, life)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, ids[0], ids[2])
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[0]) || !me.Hand.Contains(ids[2]) || me.Hand.Size() != hand+2 {
		t.Error("A and C are not the two cards added to the hand")
	}
	if !me.Graveyard.Contains(ids[1]) || !me.Graveyard.Contains(ids[3]) {
		t.Error("B and D did not go to the graveyard")
	}
	if me.Library.Contains(ids[0]) || me.Library.Contains(ids[1]) || !me.Library.Contains(ids[4]) || !me.Library.Contains(ids[5]) {
		t.Error("the library is not E, F and below")
	}
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
}

func TestSocB17StargazeWithAShortLibraryTakesWhatThereIsAndStillLosesLife(t *testing.T) {
	// Fewer than twice X: the whole library is looked at (3 cards for X=2).
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Library.Cards = nil
	ids := effectstest.SeedLibrary(me, "A", "B", "C")
	life := me.Life
	effectstest.CastXSpell(t, g, "Stargaze", "Sorcery", effectstest.SocB17StargazeOracle, "{X}{B}{B}", 2, nil)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if prompt == nil || len(prompt.ChooseCards) != 3 || prompt.ChooseMin != 2 || prompt.ChooseMax != 2 {
		t.Fatalf("prompt %+v, want all three cards, take two", prompt)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, ids[0], ids[1])
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(ids[2]) || me.Life != life-2 {
		t.Errorf("the third card in the graveyard %v, life %d (want %d)", me.Graveyard.Contains(ids[2]), me.Life, life-2)
	}

	// Fewer than X: every card goes to the hand, none to the graveyard, and
	// the life is lost anyway.
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	me.Library.Cards = nil
	ids = effectstest.SeedLibrary(me, "A")
	life = me.Life
	effectstest.CastXSpell(t, g, "Stargaze", "Sorcery", effectstest.SocB17StargazeOracle, "{X}{B}{B}", 2, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a library smaller than X asked for a pick")
	}
	if !me.Hand.Contains(ids[0]) || me.Life != life-2 {
		t.Errorf("the card in hand %v, life %d (want %d)", me.Hand.Contains(ids[0]), me.Life, life-2)
	}
}

// --- Subterranean Tremors -----------------------------------------------------

func socB17Tremors(t *testing.T, g *game.Game, x int) {
	t.Helper()
	effectstest.CastXSpell(t, g, "Subterranean Tremors", "Sorcery", effectstest.SocB17TremorsOracle, "{X}{R}", x, nil)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
}

func TestSocB17TremorsHitsOnlyCreaturesWithoutFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushTribalCreature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	bird := effectstest.PushTribalCreature(g, opp.ID, "Bird", "Creature — Bird", 2, 2, "flying")
	mine := effectstest.PushTribalCreature(g, me.ID, "My Ox", "Creature — Ox", 3, 3)
	rock := effectstest.SocB17Perm(g, opp.ID, "Rock", "Artifact", "", "{1}", 0, 0)
	lives := []int{me.Life, opp.Life}

	socB17Tremors(t, g, 2)
	if g.Battlefield.Contains(bear) {
		t.Error("a 2/2 without flying survived X=2")
	}
	if !g.Battlefield.Contains(bird) || effectstest.DamageMarkedOn(g, bird) != 0 {
		t.Error("a flyer was hit")
	}
	if !g.Battlefield.Contains(mine) || effectstest.DamageMarkedOn(g, mine) != 2 {
		t.Error("your own ground creature did not take the damage")
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("an artifact died below X=4")
	}
	if me.Life != lives[0] || opp.Life != lives[1] {
		t.Error("a player took damage; only creatures do")
	}
	if len(effectstest.SocB17Names(g, me.ID, "Lizard")) != 0 {
		t.Error("a Lizard below X=8")
	}
}

func TestSocB17TremorsDestroysArtifactsAtFourButMakesNoLizard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.SocB17Perm(g, opp.ID, "Rock", "Artifact", "", "{1}", 0, 0)
	mine := effectstest.SocB17Perm(g, me.ID, "My Rock", "Artifact", "", "{2}", 0, 0)
	golem := effectstest.SocB17Perm(g, opp.ID, "Golem", "Artifact Creature — Golem", "", "{4}", 3, 3)
	_ = effectstest.PushTribalCreature(g, opp.ID, "Bird", "Creature — Bird", 2, 2, "flying")

	socB17Tremors(t, g, 3)
	if !g.Battlefield.Contains(rock) || !g.Battlefield.Contains(mine) {
		t.Fatal("artifacts died at X=3")
	}
	if g.Battlefield.Contains(golem) {
		t.Fatal("a 3/3 artifact creature survived X=3 damage")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	rock = effectstest.SocB17Perm(g2, opp2.ID, "Rock", "Artifact", "", "{1}", 0, 0)
	mine = effectstest.SocB17Perm(g2, me2.ID, "My Rock", "Artifact", "", "{2}", 0, 0)
	land := effectstest.SocB17Perm(g2, opp2.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	bird := effectstest.PushTribalCreature(g2, opp2.ID, "Bird", "Creature — Bird", 2, 2, "flying")
	socB17Tremors(t, g2, 4)
	if g2.Battlefield.Contains(rock) || g2.Battlefield.Contains(mine) {
		t.Error("an artifact survived X=4, yours or theirs")
	}
	if !g2.Battlefield.Contains(land) || !g2.Battlefield.Contains(bird) {
		t.Error("a land or a flyer was destroyed")
	}
	if len(effectstest.SocB17Names(g2, me2.ID, "Lizard")) != 0 {
		t.Error("a Lizard at X=4")
	}
}

func TestSocB17TremorsMakesAnEightEightLizardAtEight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rock := effectstest.SocB17Perm(g, g.Seats[1].ID, "Rock", "Artifact", "", "{1}", 0, 0)
	socB17Tremors(t, g, 8)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived X=8")
	}
	lizards := effectstest.SocB17Names(g, me.ID, "Lizard")
	if len(lizards) != 1 {
		t.Fatalf("%d Lizards, want 1", len(lizards))
	}
	l := effectstest.S58p6View(t, g, lizards[0].InstanceID)
	if !l.IsToken() || l.CurrentPower() != 8 || l.CurrentToughness() != 8 || !l.IsCreature() {
		t.Errorf("Lizard %d/%d token=%v creature=%v, want an 8/8 creature token", l.CurrentPower(), l.CurrentToughness(), l.IsToken(), l.IsCreature())
	}
	if colors := l.EffectiveColors(); len(colors) != 1 || colors[0] != "R" {
		t.Errorf("Lizard colours %v, want red", colors)
	}
	// The Lizard arrives after the damage, so it is not hit by it.
	if effectstest.DamageMarkedOn(g, l.InstanceID) != 0 {
		t.Error("the new Lizard took damage from the spell that made it")
	}
}

// --- Smallpox -----------------------------------------------------------------

func socB17SmallpoxTable(g *game.Game) (creatures, lands [][]uuid.UUID) {
	creatures = make([][]uuid.UUID, len(g.Seats))
	lands = make([][]uuid.UUID, len(g.Seats))
	for i, p := range g.Seats {
		for j := 0; j < 2; j++ {
			creatures[i] = append(creatures[i], effectstest.PushVanillaCreature(g, p.ID, "Fodder", 1, 1))
			lands[i] = append(lands[i], effectstest.SocB17Perm(g, p.ID, "Forest", "Basic Land — Forest", "", "", 0, 0))
		}
	}
	return creatures, lands
}

func TestSocB17SmallpoxDoesFourThingsInOrderToEveryPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Turn.ActiveSeat
	creatures, lands := socB17SmallpoxTable(g)
	lives := make([]int, len(g.Seats))
	hands := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		lives[i], hands[i] = p.Life, p.Hand.Size()
	}

	effectstest.CastCatalogSpell(t, g, "Smallpox", "Sorcery", effectstest.SocB17SmallpoxOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.Life != lives[i]-1 {
			t.Errorf("seat %d life %d, want %d", i, p.Life, lives[i]-1)
		}
	}

	// Discards first: nothing is sacrificed while any player still owes one.
	for offset := range g.Seats {
		p := g.Seats[(active+offset)%len(g.Seats)]
		if effectstest.DiscardChoiceFor(g, p.ID) == nil {
			t.Fatalf("seat %s was not asked to discard (offset %d)", p.Name, offset)
		}
		for i := range g.Seats {
			for _, id := range append(append([]uuid.UUID{}, creatures[i]...), lands[i]...) {
				if !g.Battlefield.Contains(id) {
					t.Fatalf("a permanent was sacrificed before the discards finished")
				}
			}
		}
		effectstest.DiscardFromHand(t, g, p.ID)
	}
	for i, p := range g.Seats {
		if p.Hand.Size() != hands[i]-1 {
			t.Errorf("seat %d hand %d, want %d", i, p.Hand.Size(), hands[i]-1)
		}
	}

	// Then the creatures, chosen in APNAP order, all leaving together.
	for offset := range g.Seats {
		seat := (active + offset) % len(g.Seats)
		p := g.Seats[seat]
		prompt := effectstest.SacrificeChoiceFor(g, p.ID)
		if prompt == nil {
			t.Fatalf("seat %s was not asked for a creature", p.Name)
		}
		for _, id := range prompt.ChooseCards {
			if c, ok := g.LookupCardForEffect(id); !ok || !c.IsCreature() {
				t.Fatalf("seat %s was offered %v, which is not a creature", p.Name, c.Name)
			}
		}
		effectstest.AnswerSacrifice(t, g, p.ID, creatures[seat][0])
		for i := range g.Seats {
			if !g.Battlefield.Contains(creatures[i][0]) {
				if offset < len(g.Seats)-1 {
					t.Fatalf("a creature was sacrificed before every player had chosen")
				}
			}
			if !g.Battlefield.Contains(lands[i][0]) || !g.Battlefield.Contains(lands[i][1]) {
				t.Fatalf("a land was sacrificed during the creature step")
			}
		}
	}
	for i := range g.Seats {
		if g.Battlefield.Contains(creatures[i][0]) || !g.Battlefield.Contains(creatures[i][1]) {
			t.Errorf("seat %d: the chosen creature survived or the other one died", i)
		}
	}

	// Then the lands.
	for offset := range g.Seats {
		seat := (active + offset) % len(g.Seats)
		p := g.Seats[seat]
		prompt := effectstest.SacrificeChoiceFor(g, p.ID)
		if prompt == nil {
			t.Fatalf("seat %s was not asked for a land", p.Name)
		}
		for _, id := range prompt.ChooseCards {
			if c, ok := g.LookupCardForEffect(id); !ok || !c.IsLand() {
				t.Fatalf("seat %s was offered %v, which is not a land", p.Name, c.Name)
			}
		}
		effectstest.AnswerSacrifice(t, g, p.ID, lands[seat][1])
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i := range g.Seats {
		if g.Battlefield.Contains(lands[i][1]) || !g.Battlefield.Contains(lands[i][0]) {
			t.Errorf("seat %d: the chosen land survived or the other one died", i)
		}
		if !g.Battlefield.Contains(creatures[i][1]) {
			t.Errorf("seat %d lost its second creature", i)
		}
	}
}

func TestSocB17SmallpoxSkipsWhatAPlayerDoesNotHave(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Turn.ActiveSeat
	creatures, lands := socB17SmallpoxTable(g)
	// Seat (active+1): an empty hand, one creature and no lands.
	quiet := (active + 1) % len(g.Seats)
	g.Seats[quiet].Hand.Cards = nil
	g.WithWriteLock(func() {
		g.Battlefield.Cards = effectstest.SocB17RemoveCards(g.Battlefield.Cards, creatures[quiet][1], lands[quiet][0], lands[quiet][1])
	})
	life := g.Seats[quiet].Life

	effectstest.CastCatalogSpell(t, g, "Smallpox", "Sorcery", effectstest.SocB17SmallpoxOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Seats[quiet].Life != life-1 {
		t.Errorf("the quiet seat life %d, want %d", g.Seats[quiet].Life, life-1)
	}
	for offset := range g.Seats {
		p := g.Seats[(active+offset)%len(g.Seats)]
		if p.ID == g.Seats[quiet].ID {
			if effectstest.DiscardChoiceFor(g, p.ID) != nil {
				t.Fatal("a player with an empty hand was asked to discard")
			}
			continue
		}
		effectstest.DiscardFromHand(t, g, p.ID)
	}
	for offset := range g.Seats {
		seat := (active + offset) % len(g.Seats)
		if seat == quiet {
			continue
		}
		effectstest.AnswerSacrifice(t, g, g.Seats[seat].ID, creatures[seat][0])
	}
	for offset := range g.Seats {
		seat := (active + offset) % len(g.Seats)
		if seat == quiet {
			continue
		}
		effectstest.AnswerSacrifice(t, g, g.Seats[seat].ID, lands[seat][1])
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(creatures[quiet][0]) {
		t.Error("the lone creature was not sacrificed")
	}
	for seat := range g.Seats {
		if seat == quiet {
			continue
		}
		if g.Battlefield.Contains(creatures[seat][0]) || g.Battlefield.Contains(lands[seat][1]) {
			t.Errorf("seat %d did not finish its sacrifices", seat)
		}
	}
}

// --- Colorstorm Stallion ------------------------------------------------------

func socB17Stallion(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.SocB17Perm(g, owner, "Colorstorm Stallion", effectstest.SocB17StallionTypeLine, effectstest.SocB17StallionOracle, "{1}{U}{R}", 3, 3)
}

func TestSocB17StallionPumpsOnEveryInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	stallion := socB17Stallion(g, me.ID)
	effectstest.SocB17CastSpellFor(t, g, "Instant", "{2}{R}", "RRR")
	c := effectstest.S58p6View(t, g, stallion)
	if c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Fatalf("Stallion is %d/%d, want 4/4", c.CurrentPower(), c.CurrentToughness())
	}
	if n := len(effectstest.SocB17Names(g, me.ID, "Colorstorm Stallion")); n != 1 {
		t.Errorf("%d Stallions after a three-mana spell, want 1", n)
	}
	effectstest.SocB17CastSpellFor(t, g, "Sorcery", "{3}{R}", "RRRR")
	c = effectstest.S58p6View(t, g, stallion)
	if c.CurrentPower() != 5 || c.CurrentToughness() != 5 {
		t.Errorf("Stallion is %d/%d after a second spell, want 5/5", c.CurrentPower(), c.CurrentToughness())
	}
	if n := len(effectstest.SocB17Names(g, me.ID, "Colorstorm Stallion")); n != 1 {
		t.Errorf("%d Stallions after a four-mana spell, want 1", n)
	}
}

func TestSocB17StallionCopiesItselfOnFiveOrMoreManaSpent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	stallion := socB17Stallion(g, me.ID)
	effectstest.SocB17CastSpellFor(t, g, "Sorcery", "{4}{R}", "RRRRR")
	stallions := effectstest.SocB17Names(g, me.ID, "Colorstorm Stallion")
	if len(stallions) != 2 {
		t.Fatalf("%d Stallions after a five-mana spell, want the original and a token", len(stallions))
	}
	var token game.Card
	for _, c := range stallions {
		if c.InstanceID != stallion {
			token = c
		}
	}
	tv := effectstest.S58p6View(t, g, token.InstanceID)
	if !tv.IsToken() || !tv.IsCreature() {
		t.Fatalf("the copy is token %v creature %v", tv.IsToken(), tv.IsCreature())
	}
	// The copy has the original's copiable values, not its +1/+1.
	if tv.CurrentPower() != 3 || tv.CurrentToughness() != 3 {
		t.Errorf("the token is %d/%d, want 3/3", tv.CurrentPower(), tv.CurrentToughness())
	}
	if tv.OracleID != effectstest.SocB17StallionOracle {
		t.Errorf("the token's oracle id %q: it is not a copy of the Stallion", tv.OracleID)
	}
	if c := effectstest.S58p6View(t, g, stallion); c.CurrentPower() != 4 {
		t.Errorf("the original is %d/%d, want 4/4", c.CurrentPower(), c.CurrentToughness())
	}

	// The token has Opus too: the next big spell pumps both and copies both.
	effectstest.SocB17CastSpellFor(t, g, "Sorcery", "{5}{R}", "RRRRRR")
	if n := len(effectstest.SocB17Names(g, me.ID, "Colorstorm Stallion")); n != 4 {
		t.Errorf("%d Stallions after a second five-plus spell, want 4", n)
	}
}

func TestSocB17StallionIgnoresCreatureSpellsAndOpponentsSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	stallion := socB17Stallion(g, me.ID)
	effectstest.SocB17CastSpellFor(t, g, "Creature — Bear", "{4}{R}", "RRRRR")
	c := effectstest.S58p6View(t, g, stallion)
	if c.CurrentPower() != 3 {
		t.Errorf("a creature spell pumped the Stallion to %d", c.CurrentPower())
	}
	if n := len(effectstest.SocB17Names(g, me.ID, "Colorstorm Stallion")); n != 1 {
		t.Errorf("a creature spell copied the Stallion (%d)", n)
	}
}

func TestSocB17StallionHasWardAndHaste(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.SocB17StallionOracle)
	if !ok {
		t.Fatal("not registered")
	}
	found := false
	for _, k := range spec.PrintedKeywords {
		if k == "haste" {
			found = true
		}
	}
	if !found {
		t.Errorf("printed keywords %v lack haste", spec.PrintedKeywords)
	}
	if len(spec.Triggered) != 2 {
		t.Errorf("%d triggered abilities, want ward and Opus", len(spec.Triggered))
	}
}

// --- Informed Inkwright -------------------------------------------------------

func socB17Inkwright(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.SocB17Perm(g, owner, "Informed Inkwright", effectstest.SocB17InkwrightTypeLine, effectstest.SocB17InkwrightOracle, "{1}{W}", 2, 2)
}

func TestSocB17InkwrightMakesAFlyingInklingWhenYourSpellTargetsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	socB17Inkwright(g, me.ID)
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)

	effectstest.CastCatalogSpell(t, g, "Locust Spray", "Instant", effectstest.SocB17LocustSprayOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	inklings := effectstest.SocB17Names(g, me.ID, "Inkling")
	if len(inklings) != 1 {
		t.Fatalf("%d Inklings, want 1", len(inklings))
	}
	v := effectstest.S58p6View(t, g, inklings[0].InstanceID)
	if !v.IsToken() || v.CurrentPower() != 1 || v.CurrentToughness() != 1 || !game.HasKeyword(&v, "flying") {
		t.Errorf("Inkling %d/%d token=%v flying=%v, want a 1/1 flying token", v.CurrentPower(), v.CurrentToughness(), v.IsToken(), game.HasKeyword(&v, "flying"))
	}
	if colors := v.EffectiveColors(); len(colors) != 2 {
		t.Errorf("Inkling colours %v, want white and black", colors)
	}
}

func TestSocB17InkwrightMakesTheInklingEvenIfTheSpellKillsItsTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	socB17Inkwright(g, me.ID)
	mouse := effectstest.PushVanillaCreature(g, opp.ID, "Mouse", 1, 1)
	effectstest.CastCatalogSpell(t, g, "Locust Spray", "Instant", effectstest.SocB17LocustSprayOracle, effectstest.CardRefs(mouse))
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(mouse) {
		t.Fatal("the Mouse survived")
	}
	if n := len(effectstest.SocB17Names(g, me.ID, "Inkling")); n != 1 {
		t.Errorf("%d Inklings, want 1", n)
	}
}

func TestSocB17InkwrightIgnoresSpellsThatDoNotTargetACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	socB17Inkwright(g, me.ID)

	// An instant with no target.
	effectstest.CastCatalogSpell(t, g, "Plain Instant", "Instant", effectstest.SocB17PlainSorceryOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// An instant that targets only a player.
	effectstest.CastXSpell(t, g, "Blaze", "Sorcery", effectstest.BlazeOracle, "{X}{R}", 1,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.SocB17Names(g, me.ID, "Inkling")); n != 0 {
		t.Errorf("%d Inklings from spells that target no creature, want 0", n)
	}
}

func TestSocB17InkwrightIgnoresACreatureSpellThatTargetsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	socB17Inkwright(g, me.ID)
	_ = effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	// A creature spell with no targets: not an instant or sorcery.
	effectstest.CastCatalogSpell(t, g, "Bear Cub", "Creature — Bear", effectstest.SocB17PlainSorceryOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.SocB17Names(g, me.ID, "Inkling")); n != 0 {
		t.Errorf("%d Inklings from a creature spell, want 0", n)
	}
}

// --- Great Hall of the Biblioplex ---------------------------------------------

func socB17GreatHall(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.SocB17Perm(g, owner, "Great Hall of the Biblioplex", "Land", effectstest.SocB17GreatHallOracle, "", 0, 0)
}

func TestSocB17GreatHallMakesColorlessOrPaysLifeForSpellMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	hall := socB17GreatHall(g, me.ID)
	life := me.Life

	if err := g.ActivateManaAbility(me.ID, hall, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("{T}: Add {C}: %v", err)
	}
	if me.Life != life {
		t.Errorf("the colourless ability cost %d life", life-me.Life)
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.UntapForTest(g, hall)

	if err := g.ActivateManaAbility(me.ID, hall, 1, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatalf("{T}, pay 1 life: any colour: %v", err)
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d", me.Life, life-1)
	}
	// The mana pays for an instant, not for a creature spell.
	creature := effectstest.SocB17Hand(me, "Test Bear", "Creature — Bear", "{U}", effectstest.SocB17PlainSorceryOracle)
	if err := g.CastSpell(me.ID, creature, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("the restricted mana cast a creature spell")
	}
	instant := effectstest.SocB17Hand(me, "Test Instant", "Instant", "{U}", effectstest.SocB17PlainSorceryOracle)
	if err := g.CastSpell(me.ID, instant, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the restricted mana could not cast an instant: %v", err)
	}
}

func TestSocB17GreatHallBecomesAPermanentTwoFourWizardWhoseCastsPumpIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	hall := socB17GreatHall(g, me.ID)

	effectstest.FloatForTest(g, me, "CCCC")
	if err := g.ActivateCatalogAbility(me.ID, hall, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("animated for four mana")
	}
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, hall, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{5}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.S58p6View(t, g, hall)
	eff := c.Effective()
	if !c.IsCreature() || !c.IsLand() {
		t.Fatalf("creature %v land %v, want both", c.IsCreature(), c.IsLand())
	}
	found := false
	for _, s := range eff.Subtypes {
		if s == "Wizard" {
			found = true
		}
	}
	if !found {
		t.Errorf("subtypes %v, want Wizard", eff.Subtypes)
	}
	if c.CurrentPower() != 2 || c.CurrentToughness() != 4 {
		t.Errorf("%d/%d, want 2/4", c.CurrentPower(), c.CurrentToughness())
	}

	// A second activation does nothing: it is already a creature.
	effectstest.FloatForTest(g, me, "CCCCC")
	if err := g.ActivateCatalogAbility(me.ID, hall, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("second {5}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.S58p6View(t, g, hall).Effective().GrantedAbilities); n != 1 {
		t.Errorf("%d granted abilities after a second activation, want 1", n)
	}

	// Casting an instant or sorcery gives it +1/+0 until end of turn, once.
	effectstest.SocB17CastSpellFor(t, g, "Instant", "{U}", "U")
	c = effectstest.S58p6View(t, g, hall)
	if c.CurrentPower() != 3 || c.CurrentToughness() != 4 {
		t.Errorf("after a cast it is %d/%d, want 3/4", c.CurrentPower(), c.CurrentToughness())
	}
	effectstest.SocB17CastSpellFor(t, g, "Creature — Bear", "{U}", "U")
	if c = effectstest.S58p6View(t, g, hall); c.CurrentPower() != 3 {
		t.Errorf("a creature spell pumped it to %d", c.CurrentPower())
	}

	// The pump ends with the turn; the animation does not.
	turn := g.Turn.Seq
	for i := 0; i < 60 && g.Turn.Seq == turn; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	c = effectstest.S58p6View(t, g, hall)
	if !c.IsCreature() || c.CurrentPower() != 2 || c.CurrentToughness() != 4 {
		t.Errorf("next turn: creature %v, %d/%d; want a permanent 2/4 creature", c.IsCreature(), c.CurrentPower(), c.CurrentToughness())
	}
}

func TestSocB17GreatHallDoesNotPumpOnTheOpponentsCasts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	// An animated Hall under the opponent's control; the active player casts.
	hall := socB17GreatHall(g, opp.ID)
	effectstest.FloatForTest(g, opp, "CCCCC")
	if err := g.ActivateCatalogAbility(opp.ID, hall, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("animate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.SocB17CastSpellFor(t, g, "Instant", "{U}", "U")
	if c := effectstest.S58p6View(t, g, hall); c.CurrentPower() != 2 {
		t.Errorf("an opponent's cast pumped the Hall to %d", c.CurrentPower())
	}
	_ = me
}
