package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	dmuB04Phalanx    = "1ca8d2cb-58ba-476b-a4f2-1bf273a54cf3"
	dmuB04Monstros   = "ea55886c-48a1-4339-81fd-03dbaa116c28"
	dmuB04Necromass  = "7f6505a1-ab8b-468b-a5de-33451598b464"
	dmuB04Vinewall   = "26685121-ab1f-4366-840b-c5a80508cde6"
	dmuB04Tablet     = "399d19cd-e6fa-4959-8d2a-9f1cfd52dce7"
	dmuB04Soultender = "a3e4d14b-818a-4a4b-8462-1de9ad844941"
	dmuB04Conscript  = "e5762815-5fc3-4ae7-abee-69f27c5e4e90"
	dmuB04Veteran    = "ba21f18e-ee65-4c10-92c9-b9fb41510eba"
	dmuB04Connoisseu = "8a4d505e-b884-4b8b-93d5-495992f3858e"
	dmuB04Figment    = "ed416ae3-9291-4e1e-ad83-2f14138e9e47"
	dmuB04Baird      = "1d078108-f88c-4a32-a28f-bae5182b2f53"
)

func TestDmuB04CardsAreFullWithNoCaveats(t *testing.T) {
	want := map[string]string{
		dmuB04Phalanx: "Argivian Phalanx", dmuB04Monstros: "Molten Monstrosity",
		dmuB04Necromass: "Writhing Necromass", dmuB04Vinewall: "Floriferous Vinewall",
		dmuB04Tablet: "Inscribed Tablet", dmuB04Soultender: "Eerie Soultender",
		dmuB04Conscript: "Cult Conscript", dmuB04Veteran: "Valiant Veteran",
		dmuB04Connoisseu: "Sengir Connoisseur", dmuB04Figment: "Haunting Figment",
		dmuB04Baird: "Baird, Argivian Recruiter",
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

func TestDmuB04PrintedKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for _, c := range []struct {
		oracle, name, typeLine, kw string
	}{
		{dmuB04Phalanx, "Argivian Phalanx", "Creature — Human Kor Soldier", "vigilance"},
		{dmuB04Monstros, "Molten Monstrosity", "Creature — Hellion", "trample"},
		{dmuB04Necromass, "Writhing Necromass", "Creature — Zombie Giant", "deathtouch"},
		{dmuB04Vinewall, "Floriferous Vinewall", "Creature — Plant Wall", "defender"},
		{dmuB04Connoisseu, "Sengir Connoisseur", "Creature — Vampire", "flying"},
		{dmuB04Figment, "Haunting Figment", "Creature — Illusion", "vigilance"},
	} {
		id := effectstest.B12Push(g, me.ID, c.name, c.typeLine, c.oracle, 2, 2)
		if !effectstest.HasKeywordOnBattlefield(t, g, id, c.kw) {
			t.Errorf("%s lacks %s", c.name, c.kw)
		}
	}
}

// --- the three cost reductions ---------------------------------------------

func TestDmuB04ArgivianPhalanxHasAffinityForCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	if got := effectstest.SelfPricedMV(t, g, me, dmuB04Phalanx, "Creature — Human Kor Soldier", "{5}{W}"); got != 6 {
		t.Errorf("no creatures: it costs %d, want 6", got)
	}
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
	if got := effectstest.SelfPricedMV(t, g, me, dmuB04Phalanx, "Creature — Human Kor Soldier", "{5}{W}"); got != 4 {
		t.Errorf("two of my creatures: it costs %d, want 4", got)
	}
	for i := 0; i < 8; i++ {
		effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	}
	if got := effectstest.SelfPricedMV(t, g, me, dmuB04Phalanx, "Creature — Human Kor Soldier", "{5}{W}"); got != 1 {
		t.Errorf("ten creatures: it costs %d, want 1 (only the generic part is reduced)", got)
	}
}

func TestDmuB04MoltenMonstrosityCostsLessByTheGreatestPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	if got := effectstest.SelfPricedMV(t, g, me, dmuB04Monstros, "Creature — Hellion", "{7}{R}"); got != 8 {
		t.Errorf("no creatures: it costs %d, want 8", got)
	}
	effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 9, 9)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Ox", "Creature — Ox", 4, 4)
	if got := effectstest.SelfPricedMV(t, g, me, dmuB04Monstros, "Creature — Hellion", "{7}{R}"); got != 4 {
		t.Errorf("greatest power 4: it costs %d, want 4 (an opponent's creature is not counted)", got)
	}
	effectstest.B12Creature(g, me.ID, "Titan", "Creature — Giant", 12, 12)
	if got := effectstest.SelfPricedMV(t, g, me, dmuB04Monstros, "Creature — Hellion", "{7}{R}"); got != 1 {
		t.Errorf("greatest power 12: it costs %d, want 1 (only the generic part is reduced)", got)
	}
}

func TestDmuB04WrithingNecromassCostsLessPerCreatureCardInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	if got := effectstest.SelfPricedMV(t, g, me, dmuB04Necromass, "Creature — Zombie Giant", "{6}{B}"); got != 7 {
		t.Errorf("empty graveyard: it costs %d, want 7", got)
	}
	effectstest.PushCatalogGraveyardCard(me, "Gy Bear", "Creature — Bear", "", 2, 2)
	effectstest.PushCatalogGraveyardCard(me, "Gy Bear", "Creature — Bear", "", 2, 2)
	effectstest.PushCatalogGraveyardCard(me, "Gy Bolt", "Instant", "", 0, 0)
	effectstest.PushCatalogGraveyardCard(opp, "Their Bear", "Creature — Bear", "", 2, 2)
	if got := effectstest.SelfPricedMV(t, g, me, dmuB04Necromass, "Creature — Zombie Giant", "{6}{B}"); got != 5 {
		t.Errorf("two creature cards: it costs %d, want 5", got)
	}
	for i := 0; i < 8; i++ {
		effectstest.PushCatalogGraveyardCard(me, "Gy Bear", "Creature — Bear", "", 2, 2)
	}
	if got := effectstest.SelfPricedMV(t, g, me, dmuB04Necromass, "Creature — Zombie Giant", "{6}{B}"); got != 1 {
		t.Errorf("ten creature cards: it costs %d, want 1", got)
	}
}

// --- library digs -------------------------------------------------------------

func dmuB04TopLands(p *game.Player, spells, lands int) (landIDs []uuid.UUID) {
	effectstest.SeedLibrary(p, make([]string, spells)...)
	for i := 0; i < lands; i++ {
		landIDs = append(landIDs, effectstest.SeedLibraryTop(p, "Forest", "Basic Land — Forest", ""))
	}
	return landIDs
}

func TestDmuB04FloriferousVinewallTakesALandFromTheTopSix(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lands := dmuB04TopLands(me, 5, 1)
	before, hand := me.Library.Size(), me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Floriferous Vinewall", "Creature — Plant Wall", dmuB04Vinewall, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, lands[0])
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(lands[0]) {
		t.Fatal("the land did not reach the hand")
	}
	if me.Hand.Size() != hand+1 || me.Library.Size() != before-1 {
		t.Errorf("hand %d (was %d), library %d (was %d): want exactly the land moved", me.Hand.Size(), hand, me.Library.Size(), before)
	}
}

func TestDmuB04FloriferousVinewallMayDeclineAndLooksOnlyAtSix(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	// The land is the seventh card: out of reach. Six spells above it.
	lands := dmuB04TopLands(me, 0, 1)
	effectstest.SeedLibrary(me, make([]string, 6)...)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Floriferous Vinewall", "Creature — Plant Wall", dmuB04Vinewall, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(lands[0]) || me.Hand.Size() != hand {
		t.Errorf("a land below the top six reached the hand (hand %d, was %d)", me.Hand.Size(), hand)
	}
}

func TestDmuB04InscribedTabletTakesALandAndBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lands := dmuB04TopLands(me, 4, 1)
	tablet := effectstest.B12Push(g, me.ID, "Inscribed Tablet", "Artifact", dmuB04Tablet, 0, 0)
	effectstest.ToMainForCost(t, g)
	effectstest.FloatForTest(g, me, "C")
	before, hand := me.Library.Size(), me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, tablet, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	// A mandatory take with one land to take asks nothing.
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(lands[0]) {
		t.Fatal("the land did not reach the hand")
	}
	if me.Hand.Size() != hand+1 || me.Library.Size() != before-1 {
		t.Errorf("hand %d (was %d), library %d (was %d): want the land only", me.Hand.Size(), hand, me.Library.Size(), before)
	}
	if _, ok := effectstest.BattlefieldCard(g, tablet); ok {
		t.Error("the Tablet was not sacrificed")
	}
}

func TestDmuB04InscribedTabletDrawsWhenNoLandIsFound(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, make([]string, 5)...)
	tablet := effectstest.B12Push(g, me.ID, "Inscribed Tablet", "Artifact", dmuB04Tablet, 0, 0)
	effectstest.ToMainForCost(t, g)
	effectstest.FloatForTest(g, me, "C")
	before, hand := me.Library.Size(), me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, tablet, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Library.Size() != before-1 {
		t.Errorf("hand %d (was %d), library %d (was %d): want exactly one card drawn", me.Hand.Size(), hand, me.Library.Size(), before)
	}
}

func TestDmuB04InscribedTabletNeedsItsMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tablet := effectstest.B12Push(g, me.ID, "Inscribed Tablet", "Artifact", dmuB04Tablet, 0, 0)
	effectstest.ToMainForCost(t, g)
	if err := g.ActivateCatalogAbility(me.ID, tablet, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the Tablet activated with no mana to pay {1}")
	}
}

// --- graveyard abilities ------------------------------------------------------

func TestDmuB04EerieSoultenderMillsThreeOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Eerie Soultender", "Creature — Spirit Cleric", dmuB04Soultender, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Graveyard.Size(); got != 3 {
		t.Errorf("graveyard = %d after entry, want 3 milled", got)
	}
}

func TestDmuB04EerieSoultenderExilesItselfToReturnAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	effectstest.ToMainForCost(t, g)
	soul := effectstest.PushCatalogGraveyardCard(me, "Eerie Soultender", "Creature — Spirit Cleric", dmuB04Soultender, 3, 1)
	bear := effectstest.PushCatalogGraveyardCard(me, "Gy Bear", "Creature — Bear", "", 2, 2)
	spell := effectstest.PushCatalogGraveyardCard(me, "Gy Bolt", "Instant", "", 0, 0)
	effectstest.FloatForTest(g, me, "CCCCB")
	for _, bad := range []uuid.UUID{soul, spell} {
		if err := g.ActivateCatalogAbility(me.ID, soul, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bad)}); err == nil {
			t.Fatalf("%s was accepted; the target is another creature card", bad)
		}
	}
	effectstest.B16Activate(t, g, me.ID, soul, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Error("the other creature card did not return to hand")
	}
	if !g.Exile.Contains(soul) {
		t.Error("the Soultender was not exiled")
	}
}

func TestDmuB04ValiantVeteranPumpsOtherSoldiersOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	vet := effectstest.B12Push(g, me.ID, "Valiant Veteran", "Creature — Kor Soldier", dmuB04Veteran, 2, 2)
	soldier := effectstest.B12Creature(g, me.ID, "Soldier", "Creature — Human Soldier", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Soldier", "Creature — Human Soldier", 1, 1)
	for _, c := range []struct {
		name string
		id   uuid.UUID
		want int
	}{{"the Veteran", vet, 2}, {"my Soldier", soldier, 2}, {"my Bear", bear, 2}, {"their Soldier", theirs, 1}} {
		if p, tough := effectstest.PtOf(t, g, c.id); p != c.want || tough != c.want {
			t.Errorf("%s is %d/%d, want %d/%d", c.name, p, tough, c.want, c.want)
		}
	}
}

func TestDmuB04ValiantVeteranCountersEachSoldierFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.ToMainForCost(t, g)
	vet := effectstest.PushCatalogGraveyardCard(me, "Valiant Veteran", "Creature — Kor Soldier", dmuB04Veteran, 2, 2)
	s1 := effectstest.B12Creature(g, me.ID, "Soldier", "Creature — Human Soldier", 1, 1)
	s2 := effectstest.B12Creature(g, me.ID, "Soldier", "Creature — Human Soldier", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Soldier", "Creature — Human Soldier", 1, 1)
	effectstest.FloatForTest(g, me, "CCCWW")
	effectstest.B16Activate(t, g, me.ID, vet, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{s1, s2} {
		if n := effectstest.PlusOnes(g, id); n != 1 {
			t.Errorf("a Soldier has %d counters, want 1", n)
		}
	}
	for _, id := range []uuid.UUID{bear, theirs} {
		if n := effectstest.PlusOnes(g, id); n != 0 {
			t.Errorf("a non-Soldier or opposing creature has %d counters, want 0", n)
		}
	}
	if !g.Exile.Contains(vet) {
		t.Error("the Veteran was not exiled")
	}
}

func TestDmuB04ValiantVeteranRefusesWithoutItsMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	effectstest.ToMainForCost(t, g)
	vet := effectstest.PushCatalogGraveyardCard(me, "Valiant Veteran", "Creature — Kor Soldier", dmuB04Veteran, 2, 2)
	effectstest.FloatForTest(g, me, "CCCW")
	if err := g.ActivateCatalogAbility(me.ID, vet, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the Veteran activated with only four mana")
	}
}

func TestDmuB04CultConscriptEntersTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.CastCatalogSpell(t, g, "Cult Conscript", "Creature — Skeleton Warrior", dmuB04Conscript, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, id) {
		t.Error("Cult Conscript entered untapped")
	}
}

func TestDmuB04CultConscriptReturnsOnlyAfterANonSkeletonDied(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	effectstest.ToMainForCost(t, g)
	conscript := effectstest.PushCatalogGraveyardCard(me, "Cult Conscript", "Creature — Skeleton Warrior", dmuB04Conscript, 2, 1)
	effectstest.FloatForTest(g, me, "CB")
	if err := g.ActivateCatalogAbility(me.ID, conscript, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("it returned with no creature having died")
	}
	// A Skeleton dying under my control is not enough.
	skel := effectstest.B12Creature(g, me.ID, "Skeleton", "Creature — Skeleton", 1, 1)
	effectstest.KillCreature(t, g, me.ID, skel)
	if err := g.ActivateCatalogAbility(me.ID, conscript, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("it returned after only a Skeleton died")
	}
	// A non-Skeleton dying under an opponent's control is not enough.
	_, opp, _ := effectstest.BlbB1Seats(g)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	if err := g.ActivateCatalogAbility(me.ID, conscript, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("it returned after only an opponent's creature died")
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.B16Activate(t, g, me.ID, conscript, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, conscript)
	if !ok {
		t.Fatal("Cult Conscript did not return to the battlefield")
	}
	if !c.Tapped {
		t.Error("Cult Conscript returned untapped")
	}
}

// --- triggers and conditions --------------------------------------------------

func TestDmuB04SengirConnoisseurGrowsOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	sengir := effectstest.B12Push(g, me.ID, "Sengir Connoisseur", "Creature — Vampire", dmuB04Connoisseu, 3, 3)
	a := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, me.ID, "Ox", "Creature — Ox", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.KillCreature(t, g, me.ID, a)
	effectstest.KillCreature(t, g, opp.ID, b)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOnes(g, sengir); n != 1 {
		t.Fatalf("%d counters after two deaths, want 1", n)
	}
	effectstest.KillCreature(t, g, me.ID, c)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOnes(g, sengir); n != 1 {
		t.Errorf("%d counters after a second death batch the same turn, want still 1", n)
	}
}

func TestDmuB04SengirConnoisseurIgnoresItsOwnDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	sengir := effectstest.B12Push(g, me.ID, "Sengir Connoisseur", "Creature — Vampire", dmuB04Connoisseu, 3, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.KillCreature(t, g, me.ID, sengir)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if g.Stack.Size() != 0 {
		t.Error("Sengir Connoisseur triggered on its own death")
	}
}

func dmuB04FigmentAttack(t *testing.T, cast bool, spell string) (*game.Game, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	figment := effectstest.B12Push(g, me.ID, "Haunting Figment", "Creature — Illusion", dmuB04Figment, 2, 1)
	blocker := effectstest.B12Creature(g, opp.ID, "Wall of Stone", "Creature — Wall", 0, 8)
	effectstest.AdvanceToMain(t, g)
	if cast {
		effectstest.CastCatalogSpell(t, g, "Test Spell", spell, "", nil)
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(figment, opp.ID); err != nil {
		t.Fatalf("attack: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	return g, blocker, figment
}

func TestDmuB04HauntingFigmentCantBeBlockedAfterAnInstantOrSorcery(t *testing.T) {
	for _, spell := range []string{"Instant", "Sorcery"} {
		g, blocker, figment := dmuB04FigmentAttack(t, true, spell)
		effectstest.BrRefusal(t, g.DeclareBlocker(blocker, figment), game.BlockReasonCantBeBlocked)
	}
}

func TestDmuB04HauntingFigmentCanBeBlockedOtherwise(t *testing.T) {
	g, blocker, figment := dmuB04FigmentAttack(t, false, "")
	if err := g.DeclareBlocker(blocker, figment); err != nil {
		t.Errorf("with no instant or sorcery cast it may be blocked: %v", err)
	}
	g, blocker, figment = dmuB04FigmentAttack(t, true, "Creature — Bear")
	if err := g.DeclareBlocker(blocker, figment); err != nil {
		t.Errorf("a creature spell does not make it unblockable: %v", err)
	}
}

func TestDmuB04BairdMakesASoldierOnlyWithAPoweredUpCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	effectstest.B12Push(g, me.ID, "Baird, Argivian Recruiter", "Legendary Creature — Human Soldier", dmuB04Baird, 2, 2)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Soldier", me.ID); n != 0 {
		t.Fatalf("%d Soldier tokens with no creature over its base power, want 0", n)
	}

	g = effectstest.NewCatalogGame(t)
	me, _, _ = effectstest.BlbB1Seats(g)
	effectstest.B12Push(g, me.ID, "Baird, Argivian Recruiter", "Legendary Creature — Human Soldier", dmuB04Baird, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.PutCounters(t, g, bear, game.CounterPlusOne, 1)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Soldier", me.ID); n != 1 {
		t.Errorf("%d Soldier tokens with a +1/+1 counter on a creature, want 1", n)
	}
}

func TestDmuB04BairdReadsBaseNotCurrentPowerAndOnlyOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.B12Push(g, me.ID, "Baird, Argivian Recruiter", "Legendary Creature — Human Soldier", dmuB04Baird, 2, 2)
	effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Ox", "Creature — Ox", 2, 2)
	effectstest.PutCounters(t, g, theirs, game.CounterPlusOne, 1)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Soldier", me.ID); n != 0 {
		t.Errorf("%d Soldier tokens from an opponent's grown creature, want 0", n)
	}
}
