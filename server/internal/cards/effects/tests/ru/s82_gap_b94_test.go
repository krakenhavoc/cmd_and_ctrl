package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b94_test.go — set GAP, batch 94: Doctor Who enchantments and
// instants (S82, #5948).

const (
	gapOutOfTime     = "8491de1e-dc37-4a89-b4f7-56a2c668f0fc"
	gapCaves         = "5ce47fb9-4e4f-4419-af75-309d2743f5f9"
	gapGirl          = "39068bf1-cae4-4ac9-82ef-4e81bf59446e"
	gapNightDoctor   = "a9c1fa95-34a8-42ef-af2f-1394df23cad8"
	gapWarGames      = "d7cfe3a3-e8e6-446d-a855-46c845f657fa"
	gapRunForLife    = "c95bedcc-911f-4c0f-a148-1cf5787e82bd"
	gapClockspinning = "3e1b32a9-27a4-4550-a644-ebe254e1f512"
	gapDontBlink     = "5733c3fb-c533-456c-b30e-5d2b9e206b6b"
	gapFlatline      = "d9c4a3ca-b966-4fd0-b189-ffbcbfb412b1"
	gapSurge         = "4d720849-56cd-4b78-b9a7-7b9484702f0f"
	gapTimeLordRegen = "a9a17d40-1675-49d9-9cbd-f6fe1b075d0c"
)

func gapB94Registered(t *testing.T, names map[string]string) {
	t.Helper()
	for oracle, name := range names {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v spec=%+v", name, ok, spec)
		}
	}
}

func TestS82GapB94EveryCardIsRegisteredFull(t *testing.T) {
	gapB94Registered(t, map[string]string{
		gapOutOfTime: "Out of Time", gapCaves: "The Caves of Androzani", gapGirl: "The Girl in the Fireplace",
		gapNightDoctor: "The Night of the Doctor", gapWarGames: "The War Games", gapRunForLife: "Run for Your Life",
		gapClockspinning: "Clockspinning", gapDontBlink: "Don't Blink", gapFlatline: "Flatline",
		gapSurge: "Surge of Brilliance", gapTimeLordRegen: "Time Lord Regeneration",
	})
}

func gapB94TimeCounters(g *game.Game, id uuid.UUID) int {
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		return -1
	}
	return c.Counters[game.CounterTime]
}

// gapB94ToMyNextFirstMain walks to seat 0's next precombat main phase,
// settling the stack at each step.
func gapB94ToMyNextFirstMain(t *testing.T, g *game.Game) {
	t.Helper()
	step := func() {
		effectstest.PassPriorityAroundTable(t, g)
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	for i := 0; i < 400 && g.Turn.ActiveSeat == 0; i++ {
		step()
	}
	for i := 0; i < 400 && !(g.Turn.ActiveSeat == 0 && g.Turn.Step == game.StepPrecombatMain); i++ {
		step()
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// gapB94PickTargets answers every open target prompt of chooser with all of
// `want` that are offered (a "up to two" clause is one prompt).
func gapB94PickTargets(t *testing.T, g *game.Game, chooser uuid.UUID, want ...uuid.UUID) {
	t.Helper()
	for i := 0; i < 8; i++ {
		var ch *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoicePickTarget && c.Chooser == chooser {
				ch = c
				break
			}
		}
		if ch == nil {
			return
		}
		var refs []game.TargetRef
		for _, offered := range ch.PickTargetCards {
			for _, w := range want {
				if offered == w {
					refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: w})
				}
			}
		}
		if err := g.ResolvePickTargets(ch.ID, chooser, refs); err != nil {
			t.Fatalf("ResolvePickTargets: %v", err)
		}
	}
}

// gapB94DropGoads removes the goad records so a test can walk through combat
// steps without declaring the goaded Warriors as attackers.
func gapB94DropGoads(g *game.Game) {
	g.WithWriteLock(func() {
		kept := g.ScopedEffects[:0]
		for _, e := range g.ScopedEffects {
			goad := false
			for _, m := range e.Mods {
				goad = goad || m.Kind == game.ModGoad
			}
			if !goad {
				kept = append(kept, e)
			}
		}
		g.ScopedEffects = kept
	})
}

// gapB94AnswerOwnPermanents answers the open own-permanents prompt of chooser.
func gapB94AnswerOwnPermanents(t *testing.T, g *game.Game, chooser uuid.UUID, picks ...uuid.UUID) {
	t.Helper()
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, chooser)
	if pick == nil {
		t.Fatalf("no own-permanents prompt is open: %+v", g.PendingChoices)
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want 0..1 (\"you may exile a nontoken creature\")", pick.ChooseMin, pick.ChooseMax)
	}
	for _, id := range pick.ChooseCards {
		if c, ok := effectstest.BattlefieldCard(g, id); !ok || c.IsToken() {
			t.Errorf("a token or a non-battlefield card %v is offered", id)
		}
	}
	if err := g.ResolveOwnPermanents(pick.ID, chooser, picks); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
}

func gapB94Exiled(g *game.Game, id uuid.UUID) (game.Card, bool) {
	for _, c := range g.Exile.Cards {
		if c.InstanceID == id {
			return c, true
		}
	}
	return game.Card{}, false
}

// --- Out of Time ------------------------------------------------------

func TestS82GapB94OutOfTimeUntapsAndPhasesOutEveryCreatureThenReturnsThemAtTheLastCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, theirs).Tapped = true })
	rock := effectstest.B12Permanent(g, me.ID, "A Rock", "Artifact")
	clock := effectstest.CastCatalogSpell(t, g, "Out of Time", "Enchantment", gapOutOfTime, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"my creature": mine, "their creature": theirs} {
		if !g.IsPhasedOutForEffect(id) {
			t.Errorf("%s did not phase out", name)
		}
	}
	if g.IsPhasedOutForEffect(rock) {
		t.Error("a noncreature permanent phased out")
	}
	if got := gapB94TimeCounters(g, clock); got != 2 {
		t.Fatalf("%d time counters, want 2 (one per creature phased out)", got)
	}
	// Their tapped creature was untapped before it phased out.
	for _, c := range g.PhasedOutCardsForEffect() {
		if c.InstanceID == theirs && c.Tapped {
			t.Error("the tapped creature stayed tapped; the enchantment untaps all creatures first")
		}
	}
	gapB94ToMyNextFirstMain(t, g)
	if got := gapB94TimeCounters(g, clock); got != 1 {
		t.Fatalf("%d counters after an upkeep, want 1", got)
	}
	if !g.IsPhasedOutForEffect(mine) {
		t.Error("the creatures phased in before the enchantment left")
	}
	gapB94ToMyNextFirstMain(t, g)
	if _, here := effectstest.BattlefieldCard(g, clock); here {
		t.Error("Out of Time survived its last time counter")
	}
	if g.IsPhasedOutForEffect(mine) || g.IsPhasedOutForEffect(theirs) {
		t.Error("the creatures did not phase back in when Out of Time left")
	}
}

// --- The Caves of Androzani -----------------------------------------------

func TestS82GapB94CavesOfAndrozaniStunsTappedCreaturesThenAddsCountersThenSearches(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tapped1 := effectstest.B12Creature(g, opp.ID, "Tapped One", "Creature — Bear", 2, 2)
	tapped2 := effectstest.B12Creature(g, opp.ID, "Tapped Two", "Creature — Bear", 2, 2)
	untapped := effectstest.B12Creature(g, opp.ID, "Untapped", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		effectstest.BattlefieldCardPtr(t, g, tapped1).Tapped = true
		effectstest.BattlefieldCardPtr(t, g, tapped2).Tapped = true
	})
	spec, _ := effects.Lookup(gapCaves)
	effectstest.RfPermAExpectLegal(t, g, me.ID, spec.Triggered[0].Targets, map[uuid.UUID]bool{tapped1: true, tapped2: true, untapped: false})
	doctor := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: doctor, Name: "A Doctor", TypeLine: "Legendary Creature — Time Lord Doctor", Owner: me.ID, Controller: me.ID})
	effectstest.PushLibraryTyped(me, "Other", "Basic Land — Forest")
	saga := effectstest.CastCatalogSpell(t, g, "The Caves of Androzani", "Enchantment — Saga", gapCaves, nil)
	effectstest.PassPriorityAroundTable(t, g)
	gapB94PickTargets(t, g, me.ID, tapped1, tapped2)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{tapped1, tapped2} {
		if c, _ := effectstest.BattlefieldCard(g, id); c.Counters[game.CounterStun] != 2 {
			t.Errorf("a tapped creature has %d stun counters, want 2", c.Counters[game.CounterStun])
		}
	}
	if c, _ := effectstest.BattlefieldCard(g, untapped); c.Counters[game.CounterStun] != 0 {
		t.Error("an untapped creature was stunned")
	}
	// II: each of the opponent's turns untapped nothing (a stun counter is
	// removed instead of untapping, CR 122.1g), so both tapped creatures now
	// have one stun counter left, and each is asked about in battlefield order.
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0) // another stun counter on tapped1
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 1) // decline for tapped2
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, tapped1); c.Counters[game.CounterStun] != 2 {
		t.Errorf("the first creature has %d stun counters, want 2 (1 left + 1 added)", c.Counters[game.CounterStun])
	}
	if c, _ := effectstest.BattlefieldCard(g, tapped2); c.Counters[game.CounterStun] != 1 {
		t.Errorf("the declined creature has %d stun counters, want 1", c.Counters[game.CounterStun])
	}
	if c, _ := effectstest.BattlefieldCard(g, saga); c.Counters[game.CounterLore] != 2 {
		t.Errorf("the Saga itself was asked about: lore counters %d, want 2", c.Counters[game.CounterLore])
	}
	// III then IV.
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	// Only the first creature still has a stun counter (the other's last one
	// was removed in place of untapping), so there is one question.
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(doctor) {
		t.Error("chapter IV did not fetch the Doctor card")
	}
}

// --- The Girl in the Fireplace ----------------------------------------------

func TestS82GapB94GirlInTheFireplaceMakesTheNobleAndTheHorseThenTimeTravelsOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	doctor := effectstest.B12Creature(g, me.ID, "My Doctor", "Legendary Creature — Time Lord Doctor", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Not A Doctor", "Creature — Bear", 2, 2)
	attacker := effectstest.PushVanillaCreature(g, me.ID, "Attacker", 2, 2)
	saga := effectstest.CastCatalogSpell(t, g, "The Girl in the Fireplace", "Enchantment — Saga", gapGirl, nil)
	effectstest.PassPriorityAroundTable(t, g)
	noble := effectstest.FindBattlefieldByName(g, "Human Noble")
	if noble == uuid.Nil {
		t.Fatal("chapter I made no Human Noble")
	}
	if got := gapB94TimeCounters(g, noble); got != 3 {
		t.Fatalf("the Human Noble has %d time counters, want 3 (vanishing 3)", got)
	}
	if p, tt := effectstest.EffectivePower(t, g, noble), effectstest.EffectiveToughness(t, g, noble); p != 1 || tt != 1 {
		t.Errorf("the Human Noble is %d/%d, want 1/1", p, tt)
	}
	blocker := effectstest.B12Creature(g, opp.ID, "Hitter", "Creature — Ogre", 3, 3)
	if err := g.MarkCombatDamage(blocker, noble, 3); err != nil {
		t.Fatalf("MarkCombatDamage: %v", err)
	}
	if got := bnsMarked(g, noble); got != 0 {
		t.Errorf("the Human Noble took %d damage, want 0 (prevented)", got)
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := gapB94TimeCounters(g, noble); got != 2 {
		t.Errorf("the Human Noble has %d time counters after an upkeep, want 2", got)
	}
	horse := effectstest.FindBattlefieldByName(g, "Horse")
	if horse == uuid.Nil {
		t.Fatal("chapter II made no Horse")
	}
	if !effectstest.HasEffectiveKeyword(t, g, doctor, "horsemanship") {
		t.Error("a Doctor I control lacks horsemanship")
	}
	if effectstest.HasEffectiveKeyword(t, g, other, "horsemanship") {
		t.Error("a non-Doctor has horsemanship")
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if _, here := effectstest.BattlefieldCard(g, saga); here {
		t.Error("the Saga stayed after chapter III")
	}
	effectstest.AttackWith(t, g, opp.ID, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	before := gapB94TimeCounters(g, noble)
	effectstest.AnswerOptionPick(t, g, me.ID, 0) // put a time counter on the Human Noble
	effectstest.PassPriorityAroundTable(t, g)
	if got := gapB94TimeCounters(g, noble); got != before+1 {
		t.Errorf("time travel did not run for the creature that connected (counters %d -> %d)", before, got)
	}
}

// --- The Night of the Doctor ---------------------------------------------------

func TestS82GapB94NightOfTheDoctorDestroysAllCreaturesThenReturnsALegendWithAChosenCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	legend := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: legend, Name: "A Legend", TypeLine: "Legendary Creature — Time Lord Doctor", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	plain := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: plain, Name: "Plain Bear", TypeLine: "Creature — Bear", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "The Night of the Doctor", "Enchantment — Saga", gapNightDoctor, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Fatal("chapter I left a creature on the battlefield")
	}
	spec, _ := effects.Lookup(gapNightDoctor)
	rfCardInGraveyardLegal := func(id uuid.UUID) bool {
		var cards []uuid.UUID
		g.ReadSnapshot(func() { cards = g.LegalTargetsForEffect(game.SourceChooser(me.ID), spec.Triggered[1].Targets).Cards })
		for _, c := range cards {
			if c == id {
				return true
			}
		}
		return false
	}
	if !rfCardInGraveyardLegal(legend) || rfCardInGraveyardLegal(plain) {
		t.Errorf("legal targets: legend=%v plain=%v, want true and false", rfCardInGraveyardLegal(legend), rfCardInGraveyardLegal(plain))
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.HobAnswerTargets(t, g, me.ID, legend)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 2) // lifelink
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, legend)
	if !ok {
		t.Fatal("the legendary creature was not returned")
	}
	if c.Counters[game.CounterLifelink] != 1 || c.Counters[game.CounterFirstStrike] != 0 || c.Counters[game.CounterVigilance] != 0 {
		t.Errorf("counters %v, want one lifelink counter", c.Counters)
	}
}

// --- The War Games ----------------------------------------------------------------

func TestS82GapB94WarGamesMakesGoadedWarriorsGrowsThemThenExilesThemAll(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sacrifice := effectstest.B12Creature(g, me.ID, "My Nontoken", "Creature — Bear", 2, 2)
	saga := effectstest.CastCatalogSpell(t, g, "The War Games", "Enchantment — Saga", gapWarGames, nil)
	effectstest.PassPriorityAroundTable(t, g)
	warriors := func() (mine, theirs []uuid.UUID) {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Warrior" && c.IsToken() {
				switch c.Controller {
				case me.ID:
					mine = append(mine, c.InstanceID)
				case opp.ID:
					theirs = append(theirs, c.InstanceID)
				}
			}
		}
		return
	}
	mine, theirs := warriors()
	if len(mine) != 3 || len(theirs) != 3 {
		t.Fatalf("Warriors: mine %d theirs %d, want 3 each", len(mine), len(theirs))
	}
	for _, id := range append(append([]uuid.UUID{}, mine...), theirs...) {
		c, _ := effectstest.BattlefieldCard(g, id)
		if !c.Tapped {
			t.Error("a Warrior token did not enter tapped")
		}
		if !c.Goaded() {
			t.Error("a Warrior token is not goaded")
		}
	}
	gapB94DropGoads(g)
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range append(append([]uuid.UUID{}, mine...), theirs...) {
		if c, _ := effectstest.BattlefieldCard(g, id); c.Counters[game.CounterPlusOne] != 2 {
			t.Errorf("a Warrior has %d +1/+1 counters after II and III, want 2", c.Counters[game.CounterPlusOne])
		}
	}
	if c, _ := effectstest.BattlefieldCard(g, sacrifice); c.Counters[game.CounterPlusOne] != 0 {
		t.Error("a non-Warrior got a counter")
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	gapB94AnswerOwnPermanents(t, g, me.ID, sacrifice)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(sacrifice) {
		t.Error("the chosen creature was not exiled")
	}
	if m, o := warriors(); len(m)+len(o) != 0 {
		t.Errorf("%d Warriors remain after chapter IV's reflexive trigger", len(m)+len(o))
	}
	if _, here := effectstest.BattlefieldCard(g, saga); here {
		t.Error("the Saga stayed after chapter IV")
	}
}

func TestS82GapB94WarGamesChapterFourDoesNothingWhenYouDeclineTheExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	keep := effectstest.B12Creature(g, me.ID, "My Nontoken", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "The War Games", "Enchantment — Saga", gapWarGames, nil)
	effectstest.PassPriorityAroundTable(t, g)
	gapB94DropGoads(g)
	for i := 0; i < 3; i++ {
		effectstest.AdvanceToPrecombatMainOf(t, g, 0)
		effectstest.PassPriorityAroundTable(t, g)
	}
	gapB94AnswerOwnPermanents(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(keep) || effectstest.CountBattlefieldByName(g, "Warrior") != 12 {
		t.Errorf("declining exiled something: nontoken kept=%v Warriors=%d", g.Battlefield.Contains(keep), effectstest.CountBattlefieldByName(g, "Warrior"))
	}
}

// --- Run for Your Life ------------------------------------------------------------

func TestS82GapB94RunForYourLifeGrantsHasteAndRestrictsBlockingToHastyCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, me.ID, "Runner A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Runner B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, me.ID, "Bystander", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Run for Your Life", "Instant", gapRunForLife, nil); err == nil {
		t.Error("cast with no target")
	}
	effectstest.CastCatalogSpell(t, g, "Run for Your Life", "Instant", gapRunForLife, effectstest.CardTargets(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if !effectstest.HasEffectiveKeyword(t, g, id, "haste") {
			t.Errorf("a target did not gain haste")
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, c, "haste") {
		t.Error("the bystander gained haste")
	}
	slow := effectstest.B12Creature(g, opp.ID, "Slow Blocker", "Creature — Bear", 2, 2)
	fast := effectstest.B12Creature(g, opp.ID, "Fast Blocker", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		effectstest.BattlefieldCardPtr(t, g, fast).Keywords = append(effectstest.BattlefieldCardPtr(t, g, fast).Keywords, "haste")
	})
	card := func(id uuid.UUID) *game.Card { return effectstest.BattlefieldCardPtr(t, g, id) }
	g.ReadSnapshot(func() {
		if r := g.BlockPairRefusalLocked(card(a), card(slow)); r.Legal() {
			t.Error("a creature without haste blocked the target")
		}
		if r := g.BlockPairRefusalLocked(card(a), card(fast)); !r.Legal() {
			t.Errorf("a creature with haste could not block the target: %q", r.Reason)
		}
		if r := g.BlockPairRefusalLocked(card(c), card(slow)); !r.Legal() {
			t.Errorf("the bystander became unblockable: %q", r.Reason)
		}
	})
}

func TestS82GapB94RunForYourLifeEscapesFromTheGraveyardForFourOtherCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.B12Creature(g, me.ID, "Runner A", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Run for Your Life", TypeLine: "Instant", OracleID: gapRunForLife, ManaCost: "{U}{R}", Owner: me.ID, Controller: me.ID})
	pay := effectstest.SeedGraveyardFodder(t, g, 3)
	effectstest.FloatMana(t, g, me, "{C}{C}{U}{R}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "escape", AltCostIDs: pay, Strict: true, Targets: effectstest.CardTargets(a)}); err == nil {
		t.Fatal("escaped with only three other cards")
	}
	pay = effectstest.SeedGraveyardFodder(t, g, 1)
	pay = append(pay, effectstest.SeedGraveyardFodder(t, g, 3)...)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "escape", AltCostIDs: pay, Strict: true, Targets: effectstest.CardTargets(a)}); err != nil {
		t.Fatalf("escape cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, a, "haste") {
		t.Error("the escaped spell did not resolve")
	}
}

// --- Flatline -------------------------------------------------------------------------

func TestS82GapB94FlatlineSetsOnlyTheOpponentsCreaturesToZeroOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Giant", "Creature — Giant", 5, 5)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)
	pumped := effectstest.B12Creature(g, opp.ID, "Pumped", "Creature — Giant", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterByForEffect(opp.ID, pumped, game.CounterPlusOne, 2) })
	effectstest.CastCatalogSpell(t, g, "Flatline", "Instant", gapFlatline, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.EffectivePower(t, g, mine), effectstest.EffectiveToughness(t, g, mine); p != 5 || tt != 5 {
		t.Errorf("my creature is %d/%d, want 5/5", p, tt)
	}
	if p, tt := effectstest.EffectivePower(t, g, theirs), effectstest.EffectiveToughness(t, g, theirs); p != 0 || tt != 1 {
		t.Errorf("their creature is %d/%d, want 0/1", p, tt)
	}
	var pp, pt int
	g.ReadSnapshot(func() {
		c := effectstest.BattlefieldCardPtr(t, g, pumped)
		_, pp, pt = c.PowerToughness()
	})
	if pp != 2 || pt != 3 {
		t.Errorf("their creature with two +1/+1 counters is %d/%d, want 2/3 (base 0/1 plus counters)", pp, pt)
	}
	late := effectstest.B12Creature(g, opp.ID, "Latecomer", "Creature — Giant", 4, 4)
	if p, tt := effectstest.EffectivePower(t, g, late), effectstest.EffectiveToughness(t, g, late); p != 4 || tt != 4 {
		t.Errorf("a creature that entered afterwards is %d/%d, want 4/4 (the set is locked as it resolves)", p, tt)
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, 1)
	if p, tt := effectstest.EffectivePower(t, g, theirs), effectstest.EffectiveToughness(t, g, theirs); p != 5 || tt != 5 {
		t.Errorf("after the turn their creature is %d/%d, want 5/5 again", p, tt)
	}
}

// --- Surge of Brilliance ---------------------------------------------------------------

func TestS82GapB94SurgeOfBrillianceDrawsForSpellsNotCastFromYourHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 6; i++ {
		effectstest.PushLibraryTyped(me, "Filler", "Basic Land — Forest")
	}
	effectstest.AdvanceToMain(t, g)
	// A spell cast from exile this turn, then Surge from the hand.
	exiled := uuid.New()
	g.WithWriteLock(func() {
		g.Exile.PushTop(game.Card{InstanceID: exiled, Name: "Exiled Bolt", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
		g.GrantCastPermissionOverCardForEffect(exiled, game.CastPermission{Player: me.ID, Zone: game.ZoneExile, Label: "test"})
	})
	if err := g.CastSpell(me.ID, exiled, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Surge of Brilliance", "Instant", gapSurge, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1 (one spell cast from exile this turn; Surge itself came from the hand)", got)
	}
}

func TestS82GapB94SurgeOfBrillianceDrawsNothingWhenEverythingCameFromYourHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushLibraryTyped(me, "Filler", "Basic Land — Forest")
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Surge of Brilliance", "Instant", gapSurge, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("drew %d cards, want 0", got)
	}
}

// --- Clockspinning ----------------------------------------------------------------------

func TestS82GapB94ClockspinningRemovesOrAddsTheChosenCounterOnAPermanent(t *testing.T) {
	for _, tc := range []struct {
		name         string
		pick         int
		wantP, wantT int
	}{
		{"remove a +1/+1 counter", 0, 1, 2},
		{"put another +1/+1 counter", 1, 3, 2},
		{"remove a time counter", 2, 2, 1},
		{"put another time counter", 3, 2, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			rock := effectstest.B12Permanent(g, me.ID, "Clock", "Artifact")
			g.WithWriteLock(func() {
				effectstest.BattlefieldCardPtr(t, g, rock).Counters = map[string]int{game.CounterPlusOne: 2, game.CounterTime: 2}
			})
			effectstest.CastCatalogSpell(t, g, "Clockspinning", "Instant", gapClockspinning, effectstest.CardTargets(rock))
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AnswerOptionPick(t, g, me.ID, tc.pick)
			effectstest.PassPriorityAroundTable(t, g)
			c, _ := effectstest.BattlefieldCard(g, rock)
			if c.Counters[game.CounterPlusOne] != tc.wantP || c.Counters[game.CounterTime] != tc.wantT {
				t.Errorf("counters %v, want +1/+1=%d time=%d", c.Counters, tc.wantP, tc.wantT)
			}
		})
	}
}

func TestS82GapB94ClockspinningTargetsASuspendedCardAndBuybackReturnsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	suspended := uuid.New()
	plain := uuid.New()
	g.WithWriteLock(func() {
		g.Exile.PushTop(game.Card{InstanceID: suspended, Name: "Suspended Spell", TypeLine: "Sorcery", ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
		_ = g.GainSuspendForEffect(me.ID, suspended, 2)
		g.Exile.PushTop(game.Card{InstanceID: plain, Name: "Plain Exiled Card", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	})
	spec, _ := effects.Lookup(gapClockspinning)
	var legal []uuid.UUID
	g.ReadSnapshot(func() { legal = g.LegalTargetsForEffect(game.SourceChooser(me.ID), spec.Targets).Cards })
	if !effectstest.HasID(legal, suspended) || effectstest.HasID(legal, plain) {
		t.Errorf("legal targets %v: want the suspended card and not the plain exiled card", legal)
	}
	id, err := effectstest.CastWithOptionalCosts(t, g, "Clockspinning", "Instant", gapClockspinning, effectstest.CardTargets(suspended), []int{0}, nil)
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0) // remove a time counter
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := gapB94Exiled(g, suspended)
	if !ok || c.Counters[game.CounterTime] != 1 {
		t.Errorf("the suspended card: in exile=%v counters=%v, want 1 time counter", ok, c.Counters)
	}
	effectstest.TmpB1AssertBuyback(t, me, id, true)
}

// --- Don't Blink -----------------------------------------------------------------------

func TestS82GapB94DontBlinkShufflesCreaturesEnteringFromExileOrCastFromExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	flickered := uuid.New()
	cast := uuid.New()
	plainSpell := uuid.New()
	g.WithWriteLock(func() {
		g.Exile.PushTop(game.Card{InstanceID: flickered, Name: "Flickered Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
		g.Exile.PushTop(game.Card{InstanceID: cast, Name: "Impulse Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
		g.Exile.PushTop(game.Card{InstanceID: plainSpell, Name: "Impulse Bolt", TypeLine: "Instant", ManaCost: "{R}", Owner: me.ID, Controller: me.ID})
		for _, id := range []uuid.UUID{cast, plainSpell} {
			g.GrantCastPermissionOverCardForEffect(id, game.CastPermission{Player: me.ID, Zone: game.ZoneExile, Label: "test"})
		}
	})
	libBefore := me.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Don't Blink", "Instant", gapDontBlink, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// A creature card put onto the battlefield from exile is shuffled in.
	g.WithWriteLock(func() { _, _ = g.ReturnFromExileToBattlefieldForEffect(flickered, me.ID, false) })
	if g.Battlefield.Contains(flickered) || !me.Library.Contains(flickered) {
		t.Errorf("the creature entered from exile anyway (battlefield=%v library=%v)", g.Battlefield.Contains(flickered), me.Library.Contains(flickered))
	}
	// A creature spell cast from exile is shuffled in as it resolves.
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.CastSpell(me.ID, cast, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cast) || !me.Library.Contains(cast) {
		t.Errorf("the creature spell cast from exile entered (battlefield=%v library=%v)", g.Battlefield.Contains(cast), me.Library.Contains(cast))
	}
	if got := me.Library.Size() - libBefore; got != 2 {
		t.Errorf("library grew by %d, want 2", got)
	}
	// A noncreature spell cast from exile is untouched.
	effectstest.FloatMana(t, g, me, "{R}")
	if err := g.CastSpell(me.ID, plainSpell, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("cast an instant from exile: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(plainSpell) {
		t.Error("an instant cast from exile did not go to the graveyard")
	}
}

func TestS82GapB94DontBlinkLeavesCreaturesCastFromHandAndEndsWithTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Don't Blink", "Instant", gapDontBlink, nil)
	effectstest.PassPriorityAroundTable(t, g)
	hand := effectstest.B38CastCreature(t, g, "Hand Bear", "Creature — Bear", "", 2, 2)
	if !g.Battlefield.Contains(hand) {
		t.Error("a creature cast from the hand did not enter")
	}
	token := effectstest.B12Creature(g, me.ID, "Token-ish", "Creature — Bear", 1, 1)
	if !g.Battlefield.Contains(token) {
		t.Error("a creature put onto the battlefield from nowhere was affected")
	}
	later := uuid.New()
	effectstest.AdvanceToPrecombatMainOf(t, g, 1)
	g.WithWriteLock(func() {
		g.Exile.PushTop(game.Card{InstanceID: later, Name: "Later Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	})
	g.WithWriteLock(func() { _, _ = g.ReturnFromExileToBattlefieldForEffect(later, me.ID, false) })
	if effectstest.CountOnBattlefield(g, "Later Bear", me.ID) != 1 {
		t.Error("the replacement lasted past the end of the turn")
	}
}

// --- Time Lord Regeneration ------------------------------------------------------------

func TestS82GapB94TimeLordRegenerationReturnsATimeLordWhenTheTargetDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	doctor := effectstest.B12Creature(g, me.ID, "The Doctor", "Legendary Creature — Time Lord Doctor", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "A Bear", "Creature — Bear", 2, 2)
	spec, _ := effects.Lookup(gapTimeLordRegen)
	effectstest.RfPermAExpectLegal(t, g, me.ID, spec.Targets, map[uuid.UUID]bool{doctor: true, bear: false})
	top := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: top, Name: "Next Time Lord", TypeLine: "Legendary Creature — Time Lord", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
	skipped := effectstest.PushLibraryTyped(me, "Skipped Spell", "Sorcery")
	effectstest.CastCatalogSpell(t, g, "Time Lord Regeneration", "Instant", gapTimeLordRegen, effectstest.CardTargets(doctor))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.KillCreature(t, g, me.ID, doctor)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(top) {
		t.Fatal("the Time Lord card was not put onto the battlefield")
	}
	if g.Battlefield.Contains(skipped) || !me.Library.Contains(skipped) {
		t.Error("the revealed non-Time Lord card should be on the bottom of the library")
	}
}

func TestS82GapB94TimeLordRegenerationLastsOnlyUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	doctor := effectstest.B12Creature(g, me.ID, "The Doctor", "Legendary Creature — Time Lord Doctor", 3, 3)
	top := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: top, Name: "Next Time Lord", TypeLine: "Legendary Creature — Time Lord", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Time Lord Regeneration", "Instant", gapTimeLordRegen, effectstest.CardTargets(doctor))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToPrecombatMainOf(t, g, 1)
	effectstest.KillCreature(t, g, me.ID, doctor)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(top) {
		t.Error("the granted ability outlasted the turn")
	}
}
