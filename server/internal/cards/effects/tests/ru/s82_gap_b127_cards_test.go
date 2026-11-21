package testsru

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b127_cards_test.go — set GAP, batch 127 (#5948). Each test plays
// the real catalog entry through the engine and asserts what its text says.

const (
	gapB127TheodenOracle   = "e739e3ce-4c5e-4172-a9de-adf66ca6588c"
	gapB127EowynOracle     = "dcc349e9-399d-4f33-8131-d1801272466a"
	gapB127BillFernyOracle = "ebbe5d82-a7a4-4f4b-b038-095021d9a6a8"
	gapB127MarinerOracle   = "2c595ccb-6703-4670-a1e5-0e74d4b8d3c8"
	gapB127GoldberryOracle = "fca3545a-2956-49f9-8465-1c8937deb294"
	gapB127MeneldorOracle  = "ef9a5bd1-6ce2-4570-9d42-89d997e2fc6f"
	gapB127SarumanOracle   = "8ba24644-8dff-4f1e-a96a-7361baf46b5a"
	gapB127WatcherOracle   = "a9ddb3dc-f289-46a1-af17-3a8523a26b0d"
	gapB127FaramirOracle   = "8a8c7c20-beb7-4352-99ea-7f64a30b9f52"
	gapB127GwaihirOracle   = "b967fbb8-ddc8-485b-bd18-cbf3ef6b1bf8"
	gapB127PippinOracle    = "25afddd2-6e48-4742-b34e-e3b5def7c0fd"
	gapB127ImrahilOracle   = "4c66e0ca-2d8a-4f68-ad15-55d75959d6e8"
)

func TestGapB127EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB127TheodenOracle:   "Théoden, King of Rohan",
		gapB127EowynOracle:     "Éowyn, Fearless Knight",
		gapB127BillFernyOracle: "Bill Ferny, Bree Swindler",
		gapB127MarinerOracle:   "Elvish Mariner",
		gapB127GoldberryOracle: "Goldberry, River-Daughter",
		gapB127MeneldorOracle:  "Meneldor, Swift Savior",
		gapB127SarumanOracle:   "Saruman the White",
		gapB127WatcherOracle:   "The Watcher in the Water",
		gapB127FaramirOracle:   "Faramir, Prince of Ithilien",
		gapB127GwaihirOracle:   "Gwaihir the Windlord",
		gapB127PippinOracle:    "Pippin, Guard of the Citadel",
		gapB127ImrahilOracle:   "Prince Imrahil the Fair",
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

func gapB127Ref(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

// gapB127Main is a fresh table at the active seat's main phase, with that
// seat and the next.
func gapB127Main(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	return g, g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// gapB127Settle passes priority until the stack is empty, answering
// `chooser`'s target picks with `ref` and saying `may` to their "you may"
// prompts on the way. Any other prompt stops it, for the test to answer.
func gapB127Settle(t *testing.T, g *game.Game, chooser uuid.UUID, ref game.TargetRef, may bool) {
	t.Helper()
	for i := 0; i < 48; i++ {
		if p := effectstest.LatestPickTarget(g, chooser); p != nil {
			if err := g.ResolvePickTarget(p.ID, chooser, ref); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if c := effectstest.LatestTriggerPrompt(g, chooser); c != nil {
			if err := g.ResolveTriggerPrompt(c.ID, chooser, may); err != nil {
				t.Fatalf("ResolveTriggerPrompt: %v", err)
			}
			continue
		}
		if len(g.PendingTriggers) > 0 {
			g.RunStateChecksForTest()
			if effectstest.LatestPickTarget(g, chooser) != nil || effectstest.LatestTriggerPrompt(g, chooser) != nil {
				continue
			}
		}
		if len(g.PendingChoices) > 0 || effectstest.StackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			if errors.Is(err, game.ErrChoicePending) {
				continue
			}
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack did not settle")
}

// gapB127AnswerPicks answers the trigger's target prompts one clause at a
// time, in order, each with the next ref.
func gapB127AnswerPicks(t *testing.T, g *game.Game, chooser uuid.UUID, refs ...game.TargetRef) {
	t.Helper()
	for _, ref := range refs {
		p := effectstest.LatestPickTarget(g, chooser)
		if p == nil {
			t.Fatalf("no target prompt for %+v", ref)
		}
		if err := g.ResolvePickTarget(p.ID, chooser, ref); err != nil {
			t.Fatalf("ResolvePickTarget %+v: %v", ref, err)
		}
	}
}

func gapB127Draw(g *game.Game, who *game.Player, n int) {
	g.WithWriteLock(func() {
		_ = (effects.DrawCards{N: n}).Apply(effects.NewContext(g, &game.StackItem{Controller: who.ID, Owner: who.ID}))
	})
}

func gapB127Counters(g *game.Game, id uuid.UUID, kind string) int {
	return effectstest.CounterCount(g, id, kind)
}

func gapB127Tokens(g *game.Game, controller uuid.UUID, name string) int {
	n, _ := effectstest.CountTokensNamed(g, controller, name)
	return n
}

func gapB127Library(p *game.Player, cards ...game.Card) {
	for i := len(cards) - 1; i >= 0; i-- {
		cards[i].InstanceID = uuid.New()
		cards[i].Owner, cards[i].Controller = p.ID, p.ID
		p.Library.PushTop(cards[i])
	}
}

// --- Théoden, King of Rohan ---------------------------------------------------

func TestGapB127TheodenGivesDoubleStrikeWhenAHumanEnters(t *testing.T) {
	g, me, opp := gapB127Main(t)
	effectstest.B12Push(g, me.ID, "Théoden, King of Rohan", "Legendary Creature — Human Noble", gapB127TheodenOracle, 2, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Peasant", "Creature — Human Peasant", "", 1, 1)
	gapB127Settle(t, g, me.ID, gapB127Ref(theirs), false)
	if !effectstest.HasEffectiveKeyword(t, g, theirs, "double strike") {
		t.Error("the targeted creature did not gain double strike")
	}
}

func TestGapB127TheodenIgnoresANonHuman(t *testing.T) {
	g, me, _ := gapB127Main(t)
	effectstest.B12Push(g, me.ID, "Théoden, King of Rohan", "Legendary Creature — Human Noble", gapB127TheodenOracle, 2, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Wolf", "Creature — Wolf", "", 1, 1)
	gapB127Settle(t, g, me.ID, gapB127Ref(bear), false)
	if effectstest.HasEffectiveKeyword(t, g, bear, "double strike") {
		t.Error("a Wolf entering gave double strike")
	}
}

// --- Éowyn, Fearless Knight ------------------------------------------------------

func gapB127Colored(g *game.Game, owner uuid.UUID, name, typeLine string, power, toughness int, colors ...string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Colors: colors,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

func TestGapB127EowynExilesAGreaterPowerCreatureAndProtectsLegends(t *testing.T) {
	g, me, opp := gapB127Main(t)
	legend := effectstest.B12Creature(g, me.ID, "Legend", "Legendary Creature — Human", 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	small := gapB127Colored(g, opp.ID, "Small", "Creature — Bear", 3, 3, "G")
	big := gapB127Colored(g, opp.ID, "Big", "Creature — Dragon", 5, 5, "R", "W")
	effectstest.CastHoldCreature(t, g, "Éowyn, Fearless Knight", "Legendary Creature — Human Knight", gapB127EowynOracle, 3, 4)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, gapB127Ref(small)); err == nil {
		t.Error("a creature with power equal to Éowyn's was accepted")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, gapB127Ref(legend)); err == nil {
		t.Error("my own creature was accepted")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, gapB127Ref(big)); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("the greater-power creature was not exiled")
	}
	for _, color := range []string{"R", "W"} {
		if !effectstest.HasEffectiveKeyword(t, g, legend, game.ProtectionFromColor(color)) {
			t.Errorf("my legendary creature lacks protection from %s", color)
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, legend, game.ProtectionFromColor("G")) {
		t.Error("my legendary creature has protection from a color the victim is not")
	}
	if effectstest.HasEffectiveKeyword(t, g, plain, game.ProtectionFromColor("R")) {
		t.Error("a nonlegendary creature gained protection")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, g.Battlefield.Cards[len(g.Battlefield.Cards)-1].InstanceID, "haste") && !gapB127EowynHasHaste(t, g) {
		t.Error("Éowyn lacks haste")
	}
}

func gapB127EowynHasHaste(t *testing.T, g *game.Game) bool {
	t.Helper()
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Éowyn, Fearless Knight" {
			return effectstest.HasKeywordOnBattlefield(t, g, c.InstanceID, "haste")
		}
	}
	return false
}

func TestGapB127EowynWithNoLargerCreatureDoesNothing(t *testing.T) {
	g, me, opp := gapB127Main(t)
	legend := effectstest.B12Creature(g, me.ID, "Legend", "Legendary Creature — Human", 2, 2)
	small := gapB127Colored(g, opp.ID, "Small", "Creature — Bear", 3, 3, "G")
	effectstest.CastHoldCreature(t, g, "Éowyn, Fearless Knight", "Legendary Creature — Human Knight", gapB127EowynOracle, 3, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(small) {
		t.Error("a creature without greater power was exiled")
	}
	if effectstest.HasEffectiveKeyword(t, g, legend, game.ProtectionFromColor("G")) {
		t.Error("protection was granted with nothing exiled")
	}
}

// --- Bill Ferny, Bree Swindler ---------------------------------------------------

func gapB127Block(t *testing.T, g *game.Game, me, opp *game.Player, attacker, blocker uuid.UUID) {
	t.Helper()
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatal(err)
	}
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatal(err)
	}
}

func TestGapB127BillFernyMakesATreasureWhenBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bill := effectstest.B12Push(g, me.ID, "Bill Ferny, Bree Swindler", "Legendary Creature — Human Rogue", gapB127BillFernyOracle, 2, 1)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	gapB127Block(t, g, me, opp, bill, wall)
	// Choose the first mode.
	if p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID); p != nil {
		if err := g.ResolveModePick(p.ID, me.ID, []int{0}); err != nil {
			t.Fatalf("choose mode: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.HasSubtype("Treasure") {
			n++
		}
	}
	if n != 1 {
		t.Errorf("Treasures = %d, want 1", n)
	}
}

func TestGapB127BillFernyGivesAHorseAwayToLeaveCombatForThreeTreasures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bill := effectstest.B12Push(g, me.ID, "Bill Ferny, Bree Swindler", "Legendary Creature — Human Rogue", gapB127BillFernyOracle, 2, 1)
	horse := effectstest.B12Creature(g, me.ID, "Pony", "Creature — Horse", 1, 1)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	gapB127Block(t, g, me, opp, bill, wall)
	if p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID); p != nil {
		if err := g.ResolveModePick(p.ID, me.ID, []int{1}); err != nil {
			t.Fatalf("choose mode: %v", err)
		}
	}
	gapB127AnswerPicks(t, g, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}, gapB127Ref(horse))
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.GaLayered(t, g, horse); c.Controller != opp.ID {
		t.Error("the Horse was not given to the opponent")
	}
	if c := effectstest.GaLayered(t, g, bill); c.AttackingTarget != uuid.Nil {
		t.Error("Bill Ferny is still attacking")
	}
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.HasSubtype("Treasure") {
			n++
		}
	}
	if n != 3 {
		t.Errorf("Treasures = %d, want 3", n)
	}
}

// --- Elvish Mariner -------------------------------------------------------------------

func TestGapB127MarinerScriesWhenItAttacksAndTapsXPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mariner := effectstest.B12Push(g, me.ID, "Elvish Mariner", "Creature — Elf Pilot", gapB127MarinerOracle, 3, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, opp.ID, "Other", "Creature — Bear", 2, 2)
	land := effectstest.B12Permanent(g, opp.ID, "Plains", "Basic Land — Plains")
	gapB127Library(me, game.Card{Name: "1"}, game.Card{Name: "2"}, game.Card{Name: "3"})
	effectstest.DeclareAttack(t, g, opp.ID, mariner)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatal("attacking did not scry 1")
	}
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the scry trigger")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(blocker, other)); err == nil {
		t.Error("two targets were accepted for a scry 1")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(land)); err == nil {
		t.Error("a land was accepted as a nonland permanent")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(blocker)); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HmlTapped(t, g, blocker) {
		t.Error("the target was not tapped")
	}
	if effectstest.HmlTapped(t, g, other) {
		t.Error("an untargeted creature was tapped")
	}
}

func TestGapB127MarinerTapsTwoForAScryTwo(t *testing.T) {
	g, me, opp := gapB127Main(t)
	effectstest.B12Push(g, me.ID, "Elvish Mariner", "Creature — Elf Pilot", gapB127MarinerOracle, 3, 2)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	bb := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	gapB127Library(me, game.Card{Name: "1"}, game.Card{Name: "2"}, game.Card{Name: "3"})
	g.WithWriteLock(func() { g.ScryForEffect(me.ID, uuid.Nil, 2) })
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(a, bb)); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HmlTapped(t, g, a) || !effectstest.HmlTapped(t, g, bb) {
		t.Error("both targets should be tapped")
	}
}

// --- Goldberry, River-Daughter ----------------------------------------------------------

func TestGapB127GoldberryTakesOneCounterOfEachKindItLacks(t *testing.T) {
	g, me, _ := gapB127Main(t)
	gold := effectstest.B12Push(g, me.ID, "Goldberry, River-Daughter", "Legendary Creature — Nymph", gapB127GoldberryOracle, 1, 3)
	other := effectstest.PushCounterCreature(g, me.ID, "Hydra", game.CounterPlusOne, 3)
	g.WithWriteLock(func() {
		_ = (effects.AddCounter{Target: other, Kind: game.CounterStun, N: 2}).Apply(effects.NewContext(g, &game.StackItem{Controller: me.ID}))
		_ = (effects.AddCounter{Target: gold, Kind: game.CounterPlusOne, N: 1}).Apply(effects.NewContext(g, &game.StackItem{Controller: me.ID}))
	})
	if err := g.ActivateCatalogAbility(me.ID, gold, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(gold)}); err == nil {
		t.Error("Goldberry targeted herself")
	}
	if err := g.ActivateCatalogAbility(me.ID, gold, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(other)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if gapB127Counters(g, gold, game.CounterStun) != 1 || gapB127Counters(g, other, game.CounterStun) != 1 {
		t.Errorf("stun counters: Goldberry %d, other %d, want 1 and 1", gapB127Counters(g, gold, game.CounterStun), gapB127Counters(g, other, game.CounterStun))
	}
	if gapB127Counters(g, gold, game.CounterPlusOne) != 1 || gapB127Counters(g, other, game.CounterPlusOne) != 3 {
		t.Errorf("+1/+1 counters: Goldberry %d, other %d, want 1 and 3 (she already had that kind)", gapB127Counters(g, gold, game.CounterPlusOne), gapB127Counters(g, other, game.CounterPlusOne))
	}
}

func TestGapB127GoldberryMovesChosenCountersAndDraws(t *testing.T) {
	g, me, _ := gapB127Main(t)
	gold := effectstest.B12Push(g, me.ID, "Goldberry, River-Daughter", "Legendary Creature — Nymph", gapB127GoldberryOracle, 1, 3)
	other := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = (effects.AddCounter{Target: gold, Kind: game.CounterPlusOne, N: 3}).Apply(effects.NewContext(g, &game.StackItem{Controller: me.ID}))
	})
	hand := me.Hand.Size()
	effectstest.FloatForTest(g, me, "U")
	if err := g.ActivateCatalogAbility(me.ID, gold, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(other)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ResolvePayAmount(effectstest.PayAmountPrompt(g, me.ID).ID, me.ID, 2); err != nil {
		t.Fatalf("move two: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if gapB127Counters(g, gold, game.CounterPlusOne) != 1 || gapB127Counters(g, other, game.CounterPlusOne) != 2 {
		t.Errorf("+1/+1 counters: Goldberry %d, other %d, want 1 and 2", gapB127Counters(g, gold, game.CounterPlusOne), gapB127Counters(g, other, game.CounterPlusOne))
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want +1", hand, me.Hand.Size())
	}
}

// --- Meneldor, Swift Savior ----------------------------------------------------------------

func TestGapB127MeneldorBlinksACreatureYouOwnOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	meneldor := effectstest.B12Push(g, me.ID, "Meneldor, Swift Savior", "Legendary Creature — Bird Soldier", gapB127MeneldorOracle, 3, 3)
	if !effectstest.HasKeywordOnBattlefield(t, g, meneldor, "flying") {
		t.Fatal("Meneldor lacks flying")
	}
	loaded := effectstest.PushCounterCreature(g, me.ID, "Loaded", game.CounterPlusOne, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, meneldor)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for Meneldor's trigger")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, gapB127Ref(theirs)); err == nil {
		t.Error("a creature I do not own was accepted")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, gapB127Ref(loaded)); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	returned := uuid.Nil
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Loaded" && c.Controller == me.ID {
			returned = c.InstanceID
		}
	}
	if returned == uuid.Nil {
		t.Fatal("the creature did not return")
	}
	if got := gapB127Counters(g, returned, game.CounterPlusOne); got > 0 {
		t.Errorf("the blinked creature kept %d counters; it should be a new object", got)
	}
}

// --- Saruman the White ---------------------------------------------------------------------

func TestGapB127SarumanAmassesOnTheSecondSpellEachTurn(t *testing.T) {
	g, me, _ := gapB127Main(t)
	effectstest.B12Push(g, me.ID, "Saruman the White", "Legendary Creature — Avatar Wizard", gapB127SarumanOracle, 4, 4)
	effectstest.CastHoldCreature(t, g, "First", "Creature — Bear", "", 1, 1)
	if n := armyCounters(g, me.ID); n != 0 {
		t.Fatalf("Army counters = %d after the first spell, want none", n)
	}
	effectstest.CastHoldCreature(t, g, "Second", "Creature — Bear", "", 1, 1)
	if n := armyCounters(g, me.ID); n != 2 {
		t.Errorf("Army counters = %d after the second spell, want 2", n)
	}
	effectstest.CastHoldCreature(t, g, "Third", "Creature — Bear", "", 1, 1)
	if n := armyCounters(g, me.ID); n != 2 {
		t.Errorf("Army counters = %d after the third spell, want still 2", n)
	}
}

func armyCounters(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && game.IsArmy(c) {
			n += c.Counters[game.CounterPlusOne]
		}
	}
	return n
}

func TestGapB127SarumanHasWardTwo(t *testing.T) {
	spec, ok := effects.Lookup(gapB127SarumanOracle)
	if !ok {
		t.Fatal("not registered")
	}
	found := false
	for _, tr := range spec.Triggered {
		if tr.Key == "Saruman the White — ward {2}" {
			found = true
		}
	}
	if !found {
		t.Error("Saruman has no ward {2} trigger")
	}
}

// --- The Watcher in the Water ---------------------------------------------------------------

func TestGapB127WatcherEntersTappedWithNineStunCounters(t *testing.T) {
	g, me, _ := gapB127Main(t)
	id := effectstest.CastHoldCreature(t, g, "The Watcher in the Water", "Legendary Creature — Kraken", gapB127WatcherOracle, 9, 9)
	if !effectstest.HmlTapped(t, g, id) {
		t.Error("the Watcher did not enter tapped")
	}
	if got := gapB127Counters(g, id, game.CounterStun); got != 9 {
		t.Errorf("stun counters = %d, want 9", got)
	}
	_ = me
}

func TestGapB127WatcherMakesATentacleOnlyWhenYouDrawOnAnOpponentsTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "The Watcher in the Water", "Legendary Creature — Kraken", gapB127WatcherOracle, 9, 9)
	effectstest.AdvanceToMain(t, g)
	gapB127Draw(g, me, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := gapB127Tokens(g, me.ID, "Tentacle"); n != 0 {
		t.Fatalf("a draw on my own turn made %d Tentacles", n)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	gapB127Draw(g, me, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if n := gapB127Tokens(g, me.ID, "Tentacle"); n != 2 {
		t.Errorf("two draws on an opponent's turn made %d Tentacles, want 2", n)
	}
}

func TestGapB127WatcherUntapsAKrakenAndStunsWhenATentacleDies(t *testing.T) {
	g, me, opp := gapB127Main(t)
	watcher := effectstest.B12Push(g, me.ID, "The Watcher in the Water", "Legendary Creature — Kraken", gapB127WatcherOracle, 9, 9)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == watcher {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
		_ = g.CreateTokenForEffect(me.ID, effects.TokenCard("1/1 blue Tentacle"), 1)
	})
	var tentacle uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.HasSubtype("Tentacle") {
			tentacle = c.InstanceID
		}
	}
	target := effectstest.B12Creature(g, opp.ID, "Target", "Creature — Bear", 2, 2)
	gapB127Die(t, g, tentacle)
	g.RunStateChecksForTest()
	gapB127AnswerPicks(t, g, me.ID, gapB127Ref(watcher), gapB127Ref(target))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HmlTapped(t, g, watcher) {
		t.Error("the Kraken was not untapped")
	}
	if got := gapB127Counters(g, target, game.CounterStun); got != 1 {
		t.Errorf("stun counters on the target = %d, want 1", got)
	}
}

func gapB127Die(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
}

// --- Faramir, Prince of Ithilien ----------------------------------------------------------------

func gapB127Faramir(t *testing.T, g *game.Game, me, opp *game.Player) {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID)
	if p == nil {
		t.Fatal("no choose-an-opponent prompt")
	}
	idx := -1
	for i, o := range p.PickOptions {
		if o.Player == opp.ID {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatalf("the opponent is not offered: %+v", p.PickOptions)
	}
	if err := g.ResolveOptionPick(p.ID, me.ID, idx); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestGapB127FaramirDrawsIfTheOpponentDidNotAttackYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Faramir, Prince of Ithilien", "Legendary Creature — Human Noble", gapB127FaramirOracle, 3, 3)
	gapB127Library(me, game.Card{Name: "1"}, game.Card{Name: "2"}, game.Card{Name: "3"}, game.Card{Name: "4"})
	gapB127Faramir(t, g, me, opp)
	hand := me.Hand.Size()
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want +1: the opponent did not attack", hand, me.Hand.Size())
	}
	if n := gapB127Tokens(g, me.ID, "Human Soldier"); n != 0 {
		t.Errorf("%d tokens, want none", n)
	}
}

func TestGapB127FaramirMakesThreeSoldiersIfTheOpponentAttackedYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Faramir, Prince of Ithilien", "Legendary Creature — Human Noble", gapB127FaramirOracle, 3, 3)
	attacker := effectstest.B12Creature(g, opp.ID, "Attacker", "Creature — Bear", 1, 1)
	gapB127Library(me, game.Card{Name: "1"}, game.Card{Name: "2"}, game.Card{Name: "3"}, game.Card{Name: "4"})
	gapB127Faramir(t, g, me, opp)
	hand := me.Hand.Size()
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.DeclareAttack(t, g, me.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if n := gapB127Tokens(g, me.ID, "Human Soldier"); n != 3 {
		t.Errorf("%d Human Soldier tokens, want 3", n)
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want no draw", hand, me.Hand.Size())
	}
}

// --- Gwaihir the Windlord ----------------------------------------------------------------------

func gapB127CardsDrawn(g *game.Game, p *game.Player) int {
	n := 0
	g.ReadSnapshot(func() { n = g.TurnTallyFor(p.ID).CardsDrawn })
	return n
}

func TestGapB127GwaihirCostsTwoLessAfterDrawingTwoCards(t *testing.T) {
	g, me, _ := gapB127Main(t)
	price := func() int {
		return effectstest.SelfPricedMV(t, g, me, gapB127GwaihirOracle, "Legendary Creature — Bird Noble", "{4}{W}{U}")
	}
	for gapB127CardsDrawn(g, me) < 3 {
		drawn := gapB127CardsDrawn(g, me)
		want := 6
		if drawn >= 2 {
			want = 4
		}
		if got := price(); got != want {
			t.Fatalf("after %d cards drawn this turn it costs %d, want %d", drawn, got, want)
		}
		gapB127Draw(g, me, 1)
	}
	if got := price(); got != 4 {
		t.Errorf("after three cards drawn it costs %d, want 4", got)
	}
}

func TestGapB127GwaihirGivesOtherBirdsVigilance(t *testing.T) {
	g, me, opp := gapB127Main(t)
	gwaihir := effectstest.B12Push(g, me.ID, "Gwaihir the Windlord", "Legendary Creature — Bird Noble", gapB127GwaihirOracle, 4, 4)
	mine := effectstest.B12Creature(g, me.ID, "Hawk", "Creature — Bird", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Hawk", "Creature — Bird", 1, 1)
	for _, kw := range []string{"flying", "vigilance"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, gwaihir, kw) {
			t.Errorf("Gwaihir lacks %s", kw)
		}
	}
	if !effectstest.HasEffectiveKeyword(t, g, mine, "vigilance") {
		t.Error("my other Bird lacks vigilance")
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "vigilance") || effectstest.HasEffectiveKeyword(t, g, theirs, "vigilance") {
		t.Error("a non-Bird or an opponent's Bird has vigilance")
	}
}

// --- Pippin, Guard of the Citadel ----------------------------------------------------------------

func TestGapB127PippinGivesProtectionFromTheChosenCardType(t *testing.T) {
	g, me, _ := gapB127Main(t)
	pippin := effectstest.B12Push(g, me.ID, "Pippin, Guard of the Citadel", "Legendary Creature — Halfling Soldier", gapB127PippinOracle, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, pippin, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(pippin)}); err == nil {
		t.Error("Pippin targeted himself")
	}
	if err := g.ActivateCatalogAbility(me.ID, pippin, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID)
	if p == nil {
		t.Fatal("no card type prompt")
	}
	idx := -1
	for i, o := range p.PickOptions {
		if o.Label == "Instant" {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatal("Instant is not offered")
	}
	if err := g.ResolveOptionPick(p.ID, me.ID, idx); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, game.ProtectionFromCardType("Instant")) {
		t.Error("the Bear lacks protection from instants")
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, game.ProtectionFromCardType("Sorcery")) {
		t.Error("the Bear has protection from a type that was not chosen")
	}
	if !effectstest.HmlTapped(t, g, pippin) {
		t.Error("Pippin did not tap")
	}
}

func TestGapB127PippinHasVigilanceAndWardOne(t *testing.T) {
	g, me, _ := gapB127Main(t)
	pippin := effectstest.B12Push(g, me.ID, "Pippin, Guard of the Citadel", "Legendary Creature — Halfling Soldier", gapB127PippinOracle, 2, 2)
	if !effectstest.HasKeywordOnBattlefield(t, g, pippin, "vigilance") {
		t.Error("Pippin lacks vigilance")
	}
	spec, _ := effects.Lookup(gapB127PippinOracle)
	found := false
	for _, tr := range spec.Triggered {
		if tr.Key == "Pippin, Guard of the Citadel — ward {1}" {
			found = true
		}
	}
	if !found {
		t.Error("Pippin has no ward {1} trigger")
	}
}

// --- Prince Imrahil the Fair ----------------------------------------------------------------------

func TestGapB127ImrahilMakesATokenOnTheSecondDrawEachTurn(t *testing.T) {
	g, me, _ := gapB127Main(t)
	effectstest.B12Push(g, me.ID, "Prince Imrahil the Fair", "Legendary Creature — Human Noble", gapB127ImrahilOracle, 2, 2)
	// The draw step has already drawn this turn's first card (or not, on a
	// first turn): keep drawing and count only what the second draw made.
	for i := 0; i < 4; i++ {
		before := gapB127CardsDrawn(g, me)
		gapB127Draw(g, me, 1)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if gapB127CardsDrawn(g, me) >= 2 {
			want = 1
		}
		if before == 0 && gapB127CardsDrawn(g, me) != 1 {
			t.Fatalf("a draw changed the turn's tally from 0 to %d", gapB127CardsDrawn(g, me))
		}
		if got := gapB127Tokens(g, me.ID, "Human Soldier"); got != want {
			t.Fatalf("after %d cards drawn this turn there are %d tokens, want %d", gapB127CardsDrawn(g, me), got, want)
		}
	}
}
