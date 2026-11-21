package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b195_test.go — S82 GAP batch 195 (#5948): Warhammer 40,000
// Commander creatures. Each test plays the real catalog entry.

const (
	gapB195TzaangorOracle    = "1d55a1c5-fd0b-44d2-9e7c-550052fef262"
	gapB195TyrantOracle      = "6f0d8900-8460-4b31-8b67-eb4e470c3317"
	gapB195ZoanthropeOracle  = "e7d79fbe-b02c-41c7-8b0a-6477feb1da7a"
	gapB195LocusOracle       = "2a64cee5-04e2-4cbf-8fa9-d55c854f9500"
	gapB195PatriarchOracle   = "785ef850-4cd7-4e7d-88f2-7b6368c91a9c"
	gapB195HeraldsOracle     = "c4ecfc13-a335-42bb-92db-133998b2af5d"
	gapB195LordOracle        = "b575fb16-c3c3-4cad-857f-122257a4b32f"
	gapB195SicarianOracle    = "65db1ec4-70b9-45f9-a2c2-01f6c8e8fd89"
	gapB195SisterOracle      = "fcc38b0b-d54a-4ecf-96f7-e09709227735"
	gapB195SuppressorOracle  = "b7c48f87-bdef-4cef-a1c4-5b4d0b7a26cd"
	gapB195CawlOracle        = "6b336228-92c0-4f14-a479-1f97aaeae018"
	gapB195LibrarianOracle   = "cb4f65b0-2841-42bf-8e18-d04a3b9f1a76"
	gapB195LightningBoltOrcl = effectstest.LightningBoltOracle
)

func TestGapB195CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB195TzaangorOracle:   "Tzaangor Shaman",
		gapB195TyrantOracle:     "Winged Hive Tyrant",
		gapB195ZoanthropeOracle: "Zoanthrope",
		gapB195LocusOracle:      "Genestealer Locus",
		gapB195PatriarchOracle:  "Genestealer Patriarch",
		gapB195HeraldsOracle:    "Heralds of Tzeentch",
		gapB195LordOracle:       "Lord of Change",
		gapB195SicarianOracle:   "Sicarian Infiltrator",
		gapB195SisterOracle:     "Sister of Silence",
		gapB195SuppressorOracle: "Vanguard Suppressor",
		gapB195CawlOracle:       "Belisarius Cawl",
		gapB195LibrarianOracle:  "Epistolary Librarian",
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

// gapB195Settle passes priority until the stack is empty, ordering triggers,
// answering every target prompt with `keep` and declining every may-cast offer.
func gapB195Settle(t *testing.T, g *game.Game, who uuid.UUID, keep game.TargetRef) {
	t.Helper()
	for i := 0; i < 96; i++ {
		if p := effectstest.TriggerOrderPromptFor(g, who); p != nil {
			order := append([]uuid.UUID(nil), p.TriggerOrderIDs...)
			if err := g.ResolveTriggerOrder(p.ID, who, order); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			continue
		}
		if p := effectstest.LatestPickTarget(g, who); p != nil {
			if err := g.ResolvePickTarget(p.ID, who, keep); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if effectstest.CountMayCastPrompts(g) > 0 {
			effectstest.AnswerAllMayCast(t, g, who, false)
			continue
		}
		if effectstest.StackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never emptied")
}

func gapB195Player(id uuid.UUID) game.TargetRef {
	return game.TargetRef{Kind: game.TargetPlayer, ID: id}
}

func gapB195Cast(t *testing.T, g *game.Game, name, typeLine, oracle, cost string, p, tough, x int) uuid.UUID {
	t.Helper()
	id, err := effectstest.SpmB4Cast(t, g, name, typeLine, oracle, cost, p, tough, x)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

// --- Tzaangor Shaman ---

func TestGapB195TzaangorShamanCopiesTheNextInstantAfterItConnects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	shaman := effectstest.B12Push(g, me.ID, "Tzaangor Shaman", "Creature — Mutant Shaman", gapB195TzaangorOracle, 3, 3)
	effectstest.AttackWith(t, g, opp.ID, shaman)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("%d delayed triggers after the Shaman connected, want 1", len(g.DelayedTriggers))
	}
	life := opp.Life
	effectstest.CastBoltAtAndKeepTheCopysTarget(t, g, me.ID, opp.ID, 1)
	if opp.Life != life-6 {
		t.Errorf("opponent life %d -> %d, want 3 from the Bolt and 3 from the copy", life, opp.Life)
	}
}

func TestGapB195TzaangorShamanCopiesOnlyOneSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	shaman := effectstest.B12Push(g, me.ID, "Tzaangor Shaman", "Creature — Mutant Shaman", gapB195TzaangorOracle, 3, 3)
	effectstest.AttackWith(t, g, opp.ID, shaman)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	life := opp.Life
	effectstest.CastBoltAtAndKeepTheCopysTarget(t, g, me.ID, opp.ID, 1)
	effectstest.CastBoltAtAndKeepTheCopysTarget(t, g, me.ID, opp.ID, 0)
	if opp.Life != life-9 {
		t.Errorf("opponent life %d -> %d, want 6 + 3 (only the first instant is copied)", life, opp.Life)
	}
}

func TestGapB195TzaangorShamanDoesNothingWithoutConnecting(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Tzaangor Shaman", "Creature — Mutant Shaman", gapB195TzaangorOracle, 3, 3)
	life := opp.Life
	effectstest.CastBoltAtAndKeepTheCopysTarget(t, g, me.ID, opp.ID, 0)
	if opp.Life != life-3 || len(g.DelayedTriggers) != 0 {
		t.Errorf("opponent life %d -> %d, %d delayed triggers; the Shaman never connected", life, opp.Life, len(g.DelayedTriggers))
	}
}

// --- Winged Hive Tyrant ---

func TestGapB195WingedHiveTyrantGivesFlyingAndHasteToCreaturesWithCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tyrant := effectstest.B12Push(g, me.ID, "Winged Hive Tyrant", "Creature — Tyranid", gapB195TyrantOracle, 4, 4)
	marked := effectstest.B12Creature(g, me.ID, "Marked", "Creature — Bear", 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.PutCounters(t, g, marked, game.CounterPlusOne, 1)
	effectstest.PutCounters(t, g, theirs, game.CounterPlusOne, 1)
	effectstest.PutCounters(t, g, tyrant, game.CounterPlusOne, 1)
	effectstest.AssertKeywords(t, g, tyrant, "flying", "haste")
	for _, kw := range []string{"flying", "haste"} {
		if !effectstest.HasEffectiveKeyword(t, g, marked, kw) {
			t.Errorf("my creature with a counter has no %s", kw)
		}
		if effectstest.HasEffectiveKeyword(t, g, plain, kw) {
			t.Errorf("my creature with no counter has %s", kw)
		}
		if effectstest.HasEffectiveKeyword(t, g, theirs, kw) {
			t.Errorf("an opponent's creature with a counter has %s", kw)
		}
	}
}

func TestGapB195WingedHiveTyrantStopsWhenTheCounterIsGone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Winged Hive Tyrant", "Creature — Tyranid", gapB195TyrantOracle, 4, 4)
	marked := effectstest.B12Creature(g, me.ID, "Marked", "Creature — Bear", 2, 2)
	effectstest.PutCounters(t, g, marked, game.CounterStun, 1)
	if !effectstest.HasEffectiveKeyword(t, g, marked, "flying") {
		t.Fatal("a stun counter did not count as a counter")
	}
	effectstest.SetCounters(g, marked, map[string]int{})
	if effectstest.HasEffectiveKeyword(t, g, marked, "flying") {
		t.Error("the grant outlived the counter")
	}
}

// --- Zoanthrope ---

func TestGapB195ZoanthropeEntersWithXCountersDrawsAtFiveAndShootsForX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B", "C")
	hand, life := me.Hand.Size(), opp.Life
	id := gapB195Cast(t, g, "Zoanthrope", "Creature — Tyranid", gapB195ZoanthropeOracle, "{X}{U}{R}", 0, 0, 5)
	gapB195Settle(t, g, me.ID, gapB195Player(opp.ID))
	if got := effectstest.PlusOnes(g, id); got != 5 {
		t.Errorf("Zoanthrope has %d counters, want 5", got)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (X = 5 draws a card)", me.Hand.Size(), hand+1)
	}
	if opp.Life != life-5 {
		t.Errorf("opponent life %d -> %d, want 5", life, opp.Life)
	}
	effectstest.AssertKeywords(t, g, id, "flying")
}

func TestGapB195ZoanthropeForXFourDoesNotDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B", "C")
	hand := me.Hand.Size()
	gapB195Cast(t, g, "Zoanthrope", "Creature — Tyranid", gapB195ZoanthropeOracle, "{X}{U}{R}", 0, 0, 4)
	gapB195Settle(t, g, me.ID, gapB195Player(opp.ID))
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d (X = 4 draws nothing)", me.Hand.Size(), hand)
	}
}

func TestGapB195ZoanthropeHasWard(t *testing.T) {
	spec, _ := effects.Lookup(gapB195ZoanthropeOracle)
	if len(spec.Triggered) != 3 {
		t.Errorf("%d triggered abilities, want the draw, the ward and the shot", len(spec.Triggered))
	}
}

// --- Genestealer Locus ---

func TestGapB195GenestealerLocusShrinksAttackersAtYouAndGrowsAttackersAtOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Genestealer Locus", "Creature — Tyranid Human", gapB195LocusOracle, 3, 3)
	attacker := effectstest.B12Creature(g, me.ID, "My Attacker", "Creature — Bear", 3, 3)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, attacker); p != 3 || tough != 4 {
		t.Errorf("my attacker at an opponent is %d/%d, want 3/4", p, tough)
	}
}

func TestGapB195GenestealerLocusShrinksACreatureAttackingYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Genestealer Locus", "Creature — Tyranid Human", gapB195LocusOracle, 3, 3)
	attacker := effectstest.B12Creature(g, opp.ID, "Their Attacker", "Creature — Bear", 3, 3)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.DeclareAttack(t, g, me.ID, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, attacker); p != 2 || tough != 3 {
		t.Errorf("a creature attacking me is %d/%d, want 2/3", p, tough)
	}
}

func TestGapB195GenestealerLocusGrowsACreatureAttackingAnotherOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.B12Push(g, me.ID, "Genestealer Locus", "Creature — Tyranid Human", gapB195LocusOracle, 3, 3)
	attacker := effectstest.B12Creature(g, opp.ID, "Their Attacker", "Creature — Bear", 3, 3)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.DeclareAttack(t, g, third.ID, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, attacker); p != 3 || tough != 4 {
		t.Errorf("a creature attacking another opponent of mine is %d/%d, want 3/4", p, tough)
	}
}

// --- Genestealer Patriarch ---

func TestGapB195GenestealerPatriarchInfectsAndThenCopiesTheDeadCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.AdvanceToMain(t, g)
	patriarch := effectstest.B12Push(g, me.ID, "Genestealer Patriarch", "Creature — Tyranid", gapB195PatriarchOracle, 4, 4)
	victim := effectstest.B12Creature(g, opp.ID, "Their Elf", "Creature — Elf Warrior", 2, 2)
	bystander := effectstest.B12Creature(g, third.ID, "Third Elf", "Creature — Elf", 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, patriarch)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no attack-trigger target prompt")
	}
	if effectstest.HasID(pick.PickTargetCards, bystander) || !effectstest.HasID(pick.PickTargetCards, victim) {
		t.Errorf("offered %v; want only creatures the defending player controls", pick.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, victim, effects.GapB195InfectionCounter); got != 1 {
		t.Fatalf("the target has %d infection counters, want 1", got)
	}
	effectstest.KillCreature(t, g, opp.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	var copyID uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Their Elf" && c.Controller == me.ID && effects.IsToken(c) {
			copyID = c.InstanceID
		}
	}
	if copyID == uuid.Nil {
		t.Fatal("no token copy of the dead creature under my control")
	}
	c, _ := effectstest.BattlefieldCard(g, copyID)
	if !c.HasSubtype("Tyranid") || !c.HasSubtype("Elf") || !c.HasSubtype("Warrior") {
		t.Errorf("the copy's types are %q, want Elf Warrior and Tyranid", c.TypeLine)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, copyID); p != 2 || tough != 2 {
		t.Errorf("the copy is %d/%d, want 2/2", p, tough)
	}
}

func TestGapB195GenestealerPatriarchIgnoresACreatureWithoutAnInfectionCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Genestealer Patriarch", "Creature — Tyranid", gapB195PatriarchOracle, 4, 4)
	plain := effectstest.B12Creature(g, opp.ID, "Plain Elf", "Creature — Elf", 2, 2)
	marked := effectstest.B12Creature(g, opp.ID, "Marked Elf", "Creature — Elf", 2, 2)
	effectstest.PutCounters(t, g, marked, game.CounterPlusOne, 1)
	effectstest.KillCreature(t, g, opp.ID, plain)
	effectstest.KillCreature(t, g, opp.ID, marked)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if effects.IsToken(c) && c.Controller == me.ID {
			t.Errorf("a creature without an infection counter was copied (%s)", c.Name)
		}
	}
}

// --- Heralds of Tzeentch / Lord of Change ---

func TestGapB195HeraldsOfTzeentchCascades(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedCheapLibrary(g, me, 3)
	effectstest.CastWithCost(t, g, "Heralds of Tzeentch", "Creature — Demon", "{4}{U}", gapB195HeraldsOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountMayCastPrompts(g); got != 1 {
		t.Fatalf("cascade offers = %d, want 1", got)
	}
	effectstest.AnswerAllMayCast(t, g, me.ID, false)
	effectstest.AssertKeywords(t, g, effectstest.B12Push(g, me.ID, "Heralds of Tzeentch", "Creature — Demon", gapB195HeraldsOracle, 3, 3), "flying")
}

func TestGapB195LordOfChangeDrawsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")
	hand := me.Hand.Size()
	id := gapB195Cast(t, g, "Lord of Change", "Creature — Demon", gapB195LordOracle, "{6}{U}", 6, 6, 0)
	gapB195Settle(t, g, me.ID, gapB195Player(g.Seats[1].ID))
	if me.Hand.Size() != hand+3 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+3)
	}
	effectstest.AssertKeywords(t, g, id, "flying")
}

// --- Sicarian Infiltrator / Vanguard Suppressor ---

func TestGapB195SicarianInfiltratorSquadDrawsForEachCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F")
	effectstest.ToMainForCost(t, g)
	id := effectstest.TmtB11Hand(g, me, "Sicarian Infiltrator", "Artifact Creature — Human Soldier", gapB195SicarianOracle, "{2}{U}", 1, 2)
	hand := me.Hand.Size()
	effectstest.FloatMana(t, g, me, "{U}{C}{C}{C}{C}{C}{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, OptionalCosts: []int{0, 0}}); err != nil {
		t.Fatalf("cast with squad paid twice: %v", err)
	}
	effectstest.SettleOrderingTriggers(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.TmtB1Named(g, "Sicarian Infiltrator")); got != 3 {
		t.Errorf("%d Sicarian Infiltrators, want the card and two token copies", got)
	}
	// The card left the hand (-1) and each of the three drew one.
	if me.Hand.Size() != hand-1+3 {
		t.Errorf("hand %d, want %d (each Infiltrator draws a card)", me.Hand.Size(), hand-1+3)
	}
}

func TestGapB195SicarianInfiltratorWithoutSquadIsOneBodyAndOneCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.ToMainForCost(t, g)
	id := effectstest.TmtB11Hand(g, me, "Sicarian Infiltrator", "Artifact Creature — Human Soldier", gapB195SicarianOracle, "{2}{U}", 1, 2)
	hand := me.Hand.Size()
	effectstest.FloatMana(t, g, me, "{U}{C}{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	effectstest.SettleOrderingTriggers(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.TmtB1Named(g, "Sicarian Infiltrator")); got != 1 {
		t.Errorf("%d Sicarian Infiltrators, want 1", got)
	}
	if me.Hand.Size() != hand-1+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand)
	}
}

func TestGapB195VanguardSuppressorSquadCopiesEachDrawWhenTheyConnect(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")
	effectstest.ToMainForCost(t, g)
	id := effectstest.TmtB11Hand(g, me, "Vanguard Suppressor", "Creature — Astartes Warrior", gapB195SuppressorOracle, "{3}{U}", 3, 2)
	effectstest.FloatMana(t, g, me, "{U}{C}{C}{C}{C}{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}}); err != nil {
		t.Fatalf("cast with squad paid once: %v", err)
	}
	effectstest.SettleOrderingTriggers(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	ids := effectstest.TmtB1Named(g, "Vanguard Suppressor")
	if len(ids) != 2 {
		t.Fatalf("%d Vanguard Suppressors, want the card and one token copy", len(ids))
	}
	effectstest.AssertKeywords(t, g, ids[0], "flying")
	// Fresh creatures are summoning-sick: take the combat with two others
	// standing in, and check the trigger on a creature that can attack.
	hand := me.Hand.Size()
	attacker := effectstest.B12Push(g, me.ID, "Veteran Suppressor", "Creature — Astartes Warrior", gapB195SuppressorOracle, 3, 2)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AttackWith(t, g, opp.ID, attacker)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() < hand+1 {
		t.Errorf("hand %d, want at least %d after the Suppressor connected", me.Hand.Size(), hand+1)
	}
}

// --- Sister of Silence ---

func gapB195SisterOnTop(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	effectstest.CastFlashCreatureFrom(t, g, me, game.Card{
		Name: "Sister of Silence", TypeLine: "Creature — Human Knight", OracleID: gapB195SisterOracle, ManaCost: "{4}{U}", Power: 3, Toughness: 3,
	})
}

func TestGapB195SisterOfSilenceCountersAnInstantSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := me.Life
	bolt := effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", gapB195LightningBoltOrcl, "{R}",
		[]game.TargetRef{gapB195Player(me.ID)})
	gapB195SisterOnTop(t, g, me)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	for i := 0; i < 16 && effectstest.LatestPickTarget(g, me.ID) == nil && !effectstest.StackFullyEmpty(g); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	effectstest.AnswerPickTarget(t, g, bolt)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("my life %d -> %d: the Bolt was not countered", life, me.Life)
	}
}

func TestGapB195SisterOfSilenceCannotTargetACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := uuid.New()
	bear = effectstest.CastFlashCreatureFrom(t, g, opp, game.Card{Name: "Their Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	gapB195SisterOnTop(t, g, me)
	for i := 0; i < 16 && effectstest.LatestPickTarget(g, me.ID) == nil && !effectstest.StackFullyEmpty(g); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
			t.Error("a creature spell was accepted as the Sister's target")
		}
	}
}

// --- Belisarius Cawl ---

func TestGapB195CawlTapsTwoArtifactsForAnAstartesWarrior(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	cawl := effectstest.B12Push(g, me.ID, "Belisarius Cawl", "Legendary Artifact Creature — Human", gapB195CawlOracle, 2, 4)
	a := effectstest.B12Permanent(g, me.ID, "Rock A", "Artifact")
	b := effectstest.B12Permanent(g, me.ID, "Rock B", "Artifact")
	c := effectstest.B12Permanent(g, me.ID, "Rock C", "Artifact")
	if err := g.ActivateCatalogAbility(me.ID, cawl, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{a}}); err == nil {
		t.Fatal("one artifact paid for two")
	}
	effectstest.B16Activate(t, g, me.ID, cawl, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{a, b}})
	effectstest.PassPriorityAroundTable(t, g)
	var token game.Card
	for _, x := range g.Battlefield.Cards {
		if x.Name == "Astartes Warrior" && x.Controller == me.ID {
			token = x
		}
	}
	if token.InstanceID == uuid.Nil {
		t.Fatal("no Astartes Warrior token")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, token.InstanceID); p != 2 || tough != 2 || !token.HasColor("W") {
		t.Errorf("the token is %d/%d colours %v, want a 2/2 white", p, tough, token.Colors)
	}
	effectstest.AssertKeywords(t, g, token.InstanceID, "vigilance")
	if !effectstest.B16Tapped(t, g, a) || !effectstest.B16Tapped(t, g, b) || effectstest.B16Tapped(t, g, c) {
		t.Error("exactly the two named artifacts should be tapped")
	}
	if !effectstest.B16Tapped(t, g, cawl) {
		t.Error("the Cawl's own {T} was not paid")
	}
}

func TestGapB195CawlLooksAtXCardsAndTakesAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	cawl := effectstest.B12Push(g, me.ID, "Belisarius Cawl", "Legendary Artifact Creature — Human", gapB195CawlOracle, 2, 4)
	one := effectstest.B12Creature(g, me.ID, "Tapper One", "Creature — Bear", 1, 1)
	two := effectstest.B12Creature(g, me.ID, "Tapper Two", "Creature — Bear", 1, 1)
	idle := effectstest.B12Creature(g, me.ID, "Idle", "Creature — Bear", 1, 1)
	effectstest.SeedLibrary(me, "Z", "Y", "X", "W", "V")
	rock := effectstest.SeedLibraryTop(me, "Library Rock", "Artifact", "{1}")
	hand := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, cawl, 1, game.ActivateAbilityParams{Strict: true, XValue: 2, TapIDs: []uuid.UUID{one, two}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChooseCards(g, me.ID) == nil {
		t.Fatal("no \"you may reveal an artifact card\" prompt")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(rock) || me.Hand.Size() != hand+1 {
		t.Errorf("the artifact among the top two was not put into my hand (hand %d, was %d)", me.Hand.Size(), hand)
	}
	if effectstest.B16Tapped(t, g, idle) {
		t.Error("a creature that was not named was tapped")
	}
}

// --- Epistolary Librarian ---

func TestGapB195LibrarianCastsASpellFreeForTheNumberOfAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	librarian := effectstest.B12Push(g, me.ID, "Epistolary Librarian", "Creature — Astartes Wizard", gapB195LibrarianOracle, 3, 4)
	other := effectstest.B12Creature(g, me.ID, "Other Attacker", "Creature — Bear", 2, 2)
	cheap := effectstest.HandCardFull(me, "Cheap Spell", "Sorcery", "{1}{G}", "", []string{"G"})
	big := effectstest.HandCardFull(me, "Big Spell", "Sorcery", "{3}{G}", "", []string{"G"})
	effectstest.DeclareAttack(t, g, opp.ID, librarian, other)
	effectstest.PassPriorityAroundTable(t, g)
	cheapID := cheap
	offer := effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, cheapID)
	for _, c := range offer.DiscardOptions {
		if c == big {
			t.Error("a mana value 4 spell was offered with two attackers")
		}
	}
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, cheapID, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("free cast: %v", err)
	}
}

func TestGapB195LibrarianOffersNothingWhenNoSpellIsCheapEnough(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	librarian := effectstest.B12Push(g, me.ID, "Epistolary Librarian", "Creature — Astartes Wizard", gapB195LibrarianOracle, 3, 4)
	effectstest.D2dSetHand(me, 0)
	effectstest.HandCardFull(me, "Big Spell", "Sorcery", "{3}{G}", "", []string{"G"})
	effectstest.DeclareAttack(t, g, opp.ID, librarian)
	effectstest.PassPriorityAroundTable(t, g)
	if o := effectstest.ResolutionCastOfferFor(g, me.ID); o != nil {
		t.Errorf("a cast offer was made with only a mana value 4 spell and one attacker: %+v", o.DiscardOptions)
	}
}
