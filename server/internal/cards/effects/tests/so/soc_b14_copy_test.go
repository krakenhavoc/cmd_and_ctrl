package testsso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b14_copy_test.go — the batch 14 cards that copy, double or set values:
// Tanazir Quandrix, Twinflame, Thunderclap Drake, Surge to Victory and
// Unbound Flourishing.

func socB14SetCounters(g *game.Game, id uuid.UUID, n int) {
	effectstest.AddCountersForTest(g, id, game.CounterPlusOne, n)
}

// --- Tanazir Quandrix ---

func TestSocB14TanazirDoublesTheCountersOnTheCreatureYouPickAsItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.AuraBear(g, me.ID)
	theirs := effectstest.AuraBear(g, opp.ID)
	socB14SetCounters(g, bear, 3)
	socB14SetCounters(g, theirs, 3)
	effectstest.CastCatalogSpell(t, g, "Tanazir Quandrix", "Legendary Creature — Elder Dragon", socB14TanazirOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 6 {
		t.Errorf("the chosen creature has %d counters, want 3 doubled to 6", n)
	}
	if n := effectstest.CountersOn(g, theirs, game.CounterPlusOne); n != 3 {
		t.Errorf("an opponent's creature has %d counters, want it left at 3", n)
	}
}

func TestSocB14TanazirDoublingAddsNothingToACreatureWithoutCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.AuraBear(g, me.ID)
	effectstest.CastCatalogSpell(t, g, "Tanazir Quandrix", "Legendary Creature — Elder Dragon", socB14TanazirOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 0 {
		t.Errorf("a creature with no counters got %d", n)
	}
}

func TestSocB14TanazirHasFlyingAndTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tanazir := effectstest.B12Push(g, me.ID, "Tanazir Quandrix", "Legendary Creature — Elder Dragon", socB14TanazirOracle, 4, 4)
	for _, kw := range []string{"flying", "trample"} {
		if !effectstest.HasEffectiveKeyword(t, g, tanazir, kw) {
			t.Errorf("Tanazir lacks %s", kw)
		}
	}
}

func socB14TanazirAttacks(t *testing.T) (*game.Game, *game.Player, uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tanazir := effectstest.B12Push(g, me.ID, "Tanazir Quandrix", "Legendary Creature — Elder Dragon", socB14TanazirOracle, 4, 4)
	socB14SetCounters(g, tanazir, 1) // a 5/5
	bear := effectstest.AuraBear(g, me.ID)
	socB14SetCounters(g, bear, 1) // base 2/2 with a counter: 3/3
	theirs := effectstest.AuraBear(g, opp.ID)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(tanazir, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	return g, me, tanazir, bear, theirs
}

func TestSocB14TanazirSetsTheBaseSizeOfYourOtherCreaturesToHis(t *testing.T) {
	g, me, tanazir, bear, theirs := socB14TanazirAttacks(t)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	// Tanazir is 5/5 with his counter, so a 2/2 with a counter becomes 6/6.
	if p, tg := effectstest.SocB14PT(t, g, bear); p != 6 || tg != 6 {
		t.Errorf("the bear is %d/%d, want 6/6 (base 5/5 and its +1/+1 counter)", p, tg)
	}
	if p, tg := effectstest.SocB14PT(t, g, tanazir); p != 5 || tg != 5 {
		t.Errorf("Tanazir is %d/%d, want his own 5/5 untouched", p, tg)
	}
	if p, tg := effectstest.SocB14PT(t, g, theirs); p != 2 || tg != 2 {
		t.Errorf("an opponent's bear is %d/%d, want 2/2", p, tg)
	}
}

func TestSocB14TanazirsEffectEndsAtEndOfTurn(t *testing.T) {
	g, me, _, bear, _ := socB14TanazirAttacks(t)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if p, tg := effectstest.SocB14PT(t, g, bear); p != 3 || tg != 3 {
		t.Errorf("the bear is %d/%d next turn, want 3/3 again", p, tg)
	}
}

func TestSocB14TanazirsAttackAbilityIsOptional(t *testing.T) {
	g, me, _, bear, _ := socB14TanazirAttacks(t)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.SocB14PT(t, g, bear); p != 3 || tg != 3 {
		t.Errorf("the bear is %d/%d after declining, want 3/3", p, tg)
	}
}

// --- Twinflame ---

func socB14TwinflameTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	a := effectstest.B12Push(g, me.ID, "Test Dragon", "Creature — Dragon", "", 3, 3)
	b := effectstest.B12Push(g, me.ID, "Test Wolf", "Creature — Wolf", "", 2, 2)
	return g, me, a, b
}

func socB14Twinflame(me *game.Player) uuid.UUID {
	return effectstest.HandCardFull(me, "Twinflame", "Sorcery", "{1}{R}", socB14TwinflameOracle, []string{"R"})
}

func socB14Tokens(g *game.Game) map[string]int {
	out := map[string]int{}
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if effects.IsToken(c) {
				out[c.Name]++
			}
		}
	})
	return out
}

func TestSocB14TwinflameCopiesEachChosenCreatureWithHasteAndExilesThemAtTheEndStep(t *testing.T) {
	g, me, a, b := socB14TwinflameTable(t)
	effectstest.B06AddMana(me, "C", "R", "C", "C", "R")
	spell := socB14Twinflame(me)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b},
	}}); err != nil {
		t.Fatalf("two targets for {1}{R} plus {2}{R}: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left, want 0 (strive costs {2}{R} more)", me.ManaPool.Total())
	}
	effectstest.PassPriorityAroundTable(t, g)
	tokens := socB14Tokens(g)
	if tokens["Test Dragon"] != 1 || tokens["Test Wolf"] != 1 || len(tokens) != 2 {
		t.Fatalf("tokens %v, want one copy of each chosen creature", tokens)
	}
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if effects.IsToken(c) {
				if !effectstest.HasKw(c.Keywords, "haste") {
					t.Errorf("token %s has no haste", c.Name)
				}
				if c.Controller != me.ID {
					t.Errorf("token %s is not under my control", c.Name)
				}
			}
		}
	})
	if effectstest.FindBattlefieldByName(g, "Test Dragon") != a {
		t.Error("the original Dragon is gone")
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if tokens := socB14Tokens(g); len(tokens) != 0 {
		t.Errorf("tokens %v survived the end step", tokens)
	}
	if effectstest.FindBattlefieldByName(g, "Test Dragon") != a || effectstest.FindBattlefieldByName(g, "Test Wolf") != b {
		t.Error("an original was exiled with the tokens")
	}
}

func TestSocB14TwinflameStriveCostsMoreForEachTargetBeyondTheFirst(t *testing.T) {
	g, me, a, b := socB14TwinflameTable(t)
	spell := socB14Twinflame(me)
	effectstest.B06AddMana(me, "C", "R")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b},
	}}); err == nil {
		t.Fatal("two targets were castable for just {1}{R}")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: a},
	}}); err != nil {
		t.Fatalf("one target for {1}{R}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if tokens := socB14Tokens(g); tokens["Test Dragon"] != 1 || len(tokens) != 1 {
		t.Errorf("tokens %v, want one Dragon copy", tokens)
	}
}

func TestSocB14TwinflameWithNoTargetsMakesNothing(t *testing.T) {
	g, me, _, _ := socB14TwinflameTable(t)
	effectstest.B06AddMana(me, "C", "R")
	if err := g.CastSpell(me.ID, socB14Twinflame(me), game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("Twinflame with no targets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if tokens := socB14Tokens(g); len(tokens) != 0 {
		t.Errorf("tokens %v from a spell with no targets", tokens)
	}
}

func TestSocB14TwinflameCannotTargetAnOpponentsCreature(t *testing.T) {
	g, me, _, _ := socB14TwinflameTable(t)
	theirs := effectstest.AuraBear(g, g.Seats[1].ID)
	effectstest.B06AddMana(me, "C", "R")
	if err := g.CastSpell(me.ID, socB14Twinflame(me), game.CastSpellParams{Strict: true, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: theirs},
	}}); err == nil {
		t.Error("Twinflame accepted an opponent's creature")
	}
}

func TestSocB14TwinflameSkipsATargetThatLeftAndStillCopiesTheRest(t *testing.T) {
	g, me, a, b := socB14TwinflameTable(t)
	effectstest.B06AddMana(me, "C", "R", "C", "C", "R")
	if err := g.CastSpell(me.ID, socB14Twinflame(me), game.CastSpellParams{Strict: true, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b},
	}}); err != nil {
		t.Fatal(err)
	}
	effectstest.KillCreature(t, g, me.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	tokens := socB14Tokens(g)
	if tokens["Test Wolf"] != 1 || len(tokens) != 1 {
		t.Errorf("tokens %v, want only the Wolf copied (CR 608.2b)", tokens)
	}
}

// #4945: Twinflame's copies are made by one creation (CR 608.2f), so they
// enter together and each Soul Warden copy sees the other enter (CR 603.6a).
// The probe is token_copies_together_test.go's.
func TestSocB14TwinflameCreatesItsCopiesAtTheSameTime(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.TctSoulWarden(g, me.ID)
	b := effectstest.TctSoulWarden(g, me.ID)
	life := me.Life

	effectstest.CastCatalogSpell(t, g, "Twinflame", "Sorcery", socB14TwinflameOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b},
	})
	effectstest.TctSettle(t, g, me.ID)

	if got := len(effectstest.TctTokens(g, me.ID)); got != 2 {
		t.Fatalf("%d tokens, want 2 (one per target)", got)
	}
	// Each original sees both copies enter (2 x 2), and each copy sees the
	// other, because they enter together (2 x 1). One at a time is 5.
	if got := me.Life - life; got != 6 {
		t.Errorf("gained %d life, want 6: each Soul Warden copy sees the other enter (CR 603.6a)", got)
	}
}

// #4945: "except it has haste" is part of the copying process, so haste is
// one of the token's copiable values (CR 707.9a) and a copy of the token,
// made by a spell with no exception of its own, has haste too. Only
// Twinflame's token is exiled at the end step.
func TestSocB14TwinflameHasteIsCopiedByACopyOfTheToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	orig := effectstest.B12Push(g, me.ID, "Test Dragon", "Creature — Dragon", "", 3, 3)

	effectstest.CastCatalogSpell(t, g, "Twinflame", "Sorcery", socB14TwinflameOracle, []game.TargetRef{{Kind: game.TargetCard, ID: orig}})
	effectstest.TctSettle(t, g, me.ID)
	tokens := effectstest.TctTokens(g, me.ID)
	if len(tokens) != 1 {
		t.Fatalf("%d tokens after Twinflame, want 1", len(tokens))
	}
	twin := tokens[0].InstanceID
	if !effectstest.HasKw(tokens[0].Keywords, "haste") {
		t.Fatal("Twinflame's copy does not have haste as a copiable value")
	}

	effectstest.CastCatalogSpell(t, g, "Cackling Counterpart", "Instant", "9e2adca5-f39c-4a09-bcce-8238ebac2c4a", []game.TargetRef{{Kind: game.TargetCard, ID: twin}})
	effectstest.TctSettle(t, g, me.ID)
	tokens = effectstest.TctTokens(g, me.ID)
	if len(tokens) != 2 {
		t.Fatalf("%d tokens after Cackling Counterpart, want 2", len(tokens))
	}
	var copyOfCopy uuid.UUID
	for _, tok := range tokens {
		if tok.InstanceID != twin {
			copyOfCopy = tok.InstanceID
		}
	}
	if !effectstest.HasEffectiveKeyword(t, g, copyOfCopy, "haste") {
		t.Error("a copy of Twinflame's token has no haste (CR 707.9a)")
	}

	effectstest.AdvanceToEndStep(t, g)
	effectstest.TctSettle(t, g, me.ID)
	left := effectstest.TctTokens(g, me.ID)
	if len(left) != 1 || left[0].InstanceID != copyOfCopy {
		t.Errorf("tokens after the end step %v, want only the copy of the copy", left)
	}
}

// #4945: a copy whose entry asks a question (Adaptive Automaton's creature
// type) pauses the creation; once it is answered, the copies are still
// exiled at the next end step, and an undo of the answer lands the same way.
func TestSocB14TwinflameExilesItsCopiesAfterACopysEntryChoiceAndUndo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	const automaton = "53c730c6-2f8c-4af8-b400-b9d573a71e60"
	a := effectstest.PushCatalogPermanent(g, me.ID, "Adaptive Automaton", "Artifact Creature — Construct", automaton, false)
	b := effectstest.PushCatalogPermanent(g, me.ID, "Adaptive Automaton", "Artifact Creature — Construct", automaton, false)
	for i := range g.Battlefield.Cards {
		g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 2, 2
	}

	effectstest.CastCatalogSpell(t, g, "Twinflame", "Sorcery", socB14TwinflameOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PendingOfKind(g, game.PendingChoiceCreatureType) == nil {
		t.Fatal("the copy's entry did not ask for its creature type")
	}
	snapshot := g.Clone()
	for run := 0; run < 2; run++ {
		if run > 0 {
			g.WithWriteLock(func() { g.RestoreFrom(snapshot) })
		}
		for i := 0; i < 2 && effectstest.PendingOfKind(g, game.PendingChoiceCreatureType) != nil; i++ {
			effectstest.AnswerCreatureType(t, g, me.ID, "Construct")
		}
		tokens := effectstest.TctTokens(g, me.ID)
		if len(tokens) != 2 {
			t.Fatalf("run %d: %d tokens, want 2", run, len(tokens))
		}
		for _, tok := range tokens {
			if !effectstest.HasEffectiveKeyword(t, g, tok.InstanceID, "haste") {
				t.Errorf("run %d: a copy has no haste", run)
			}
		}
		effectstest.AdvanceTo(t, g, game.StepEnd)
		effectstest.PassPriorityAroundTable(t, g)
		if left := effectstest.TctTokens(g, me.ID); len(left) != 0 {
			t.Errorf("run %d: %d copies survived the end step", run, len(left))
		}
	}
}

// --- Thunderclap Drake ---

func TestSocB14ThunderclapDrakeDiscountsYourInstantsAndSorceries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	drake := effectstest.B12Push(g, me.ID, "Thunderclap Drake", "Creature — Drake", socB14DrakeOracle, 2, 1)
	if !effectstest.HasEffectiveKeyword(t, g, drake, "flying") {
		t.Error("the Drake lacks flying")
	}
	instant := effectstest.HandCardFull(me, "Test Instant", "Instant", "{1}{R}", "", []string{"R"})
	effectstest.B06AddMana(me, "R")
	if err := g.CastSpell(me.ID, instant, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("an instant costing {1}{R} should cost {R}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	// Colored mana is not reduced, and a creature spell gets nothing.
	coloured := effectstest.HandCardFull(me, "Test Sorcery", "Sorcery", "{U}{U}", "", []string{"U"})
	effectstest.B06AddMana(me, "U")
	if err := g.CastSpell(me.ID, coloured, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a {U}{U} sorcery cost {U}")
	}
	me.ManaPool = nil
	creature := effectstest.HandCardFull(me, "Test Creature", "Creature — Bear", "{1}{G}", "", []string{"G"})
	effectstest.B06AddMana(me, "G")
	if err := g.CastSpell(me.ID, creature, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a creature spell was discounted")
	}

}

func TestSocB14ThunderclapDrakeDoesNotDiscountAnOpponentsSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	effectstest.B12Push(g, opp.ID, "Thunderclap Drake", "Creature — Drake", socB14DrakeOracle, 2, 1)
	effectstest.B06AddMana(me, "R")
	instant := effectstest.HandCardFull(me, "Test Instant", "Instant", "{1}{R}", "", []string{"R"})
	if err := g.CastSpell(me.ID, instant, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("an opponent's Drake discounted my instant")
	}
}

func socB14DrakeActivate(t *testing.T, g *game.Game, me *game.Player, drake uuid.UUID) {
	t.Helper()
	effectstest.B06AddMana(me, "C", "C", "U")
	if err := g.ActivateCatalogAbility(me.ID, drake, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate the Drake: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldByName(g, "Thunderclap Drake") != uuid.Nil {
		t.Fatal("the Drake was not sacrificed")
	}
}

func socB14CommanderCasts(me *game.Player, casts ...int) {
	if me.CommanderCasts == nil {
		me.CommanderCasts = map[uuid.UUID]int{}
	}
	for _, n := range casts {
		me.CommanderCasts[uuid.New()] = n
	}
}

func TestSocB14ThunderclapDrakeCopiesTheNextInstantOncePerCommanderCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	drake := effectstest.PushCatalogPermanent(g, me.ID, "Thunderclap Drake", "Creature — Drake", socB14DrakeOracle, false)
	effectstest.AdvanceToMain(t, g)
	socB14CommanderCasts(me, 1, 2) // two commanders: cast once and twice
	socB14DrakeActivate(t, g, me, drake)
	before := victim.Life
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}})
	answered := socB14SettleKeepingTargets(t, g, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: victim.ID})
	if answered != 3 {
		t.Errorf("answered %d new-target prompts, want 3 copies", answered)
	}
	if got := victim.Life; got != before-12 {
		t.Errorf("victim at %d, want %d (the Bolt and three copies)", got, before-12)
	}
	// It was used up by that spell.
	before = victim.Life
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := victim.Life; got != before-3 {
		t.Errorf("a second Bolt dealt %d, want 3: the delayed trigger fires once", before-got)
	}
}

func TestSocB14ThunderclapDrakeMakesNoCopiesWithoutACommanderCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	drake := effectstest.PushCatalogPermanent(g, me.ID, "Thunderclap Drake", "Creature — Drake", socB14DrakeOracle, false)
	effectstest.AdvanceToMain(t, g)
	socB14DrakeActivate(t, g, me, drake)
	before := victim.Life
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := victim.Life; got != before-3 {
		t.Errorf("victim lost %d, want 3 and no copy", before-got)
	}
	if n := len(g.StackMeta); n != 0 {
		t.Errorf("%d items left on the stack", n)
	}
}

func TestSocB14ThunderclapDrakeWaitsForAnInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	drake := effectstest.PushCatalogPermanent(g, me.ID, "Thunderclap Drake", "Creature — Drake", socB14DrakeOracle, false)
	effectstest.AdvanceToMain(t, g)
	socB14CommanderCasts(me, 2)
	socB14DrakeActivate(t, g, me, drake)
	effectstest.CastCatalogSpell(t, g, "Test Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("%d delayed triggers after a creature spell, want the Drake's still waiting", len(g.DelayedTriggers))
	}
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Divination", "Sorcery", effectstest.DivinationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// Divination draws two, and so does each of the two copies.
	if got := me.Hand.Size() - hand; got != 6 {
		t.Errorf("hand grew by %d, want 6 (the sorcery and two copies)", got)
	}
}

// --- Surge to Victory ---

func socB14SurgeTable(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gy := effectstest.GyCard(me, "Divination", "Sorcery", "{2}{U}", effectstest.DivinationOracle)
	return g, me, opp, gy
}

func TestSocB14SurgeToVictoryExilesTheCardAndPumpsYourCreaturesByItsManaValue(t *testing.T) {
	g, me, opp, gy := socB14SurgeTable(t)
	bear := effectstest.AuraBear(g, me.ID)
	theirs := effectstest.AuraBear(g, opp.ID)
	effectstest.CastCatalogSpell(t, g, "Surge to Victory", "Sorcery", socB14SurgeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: gy}})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, gy) {
		t.Fatal("the targeted card is not exiled")
	}
	if p, tg := effectstest.SocB14PT(t, g, bear); p != 5 || tg != 2 {
		t.Errorf("the bear is %d/%d, want 5/2 (+3/+0 for mana value 3)", p, tg)
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 2 {
		t.Errorf("an opponent's bear has power %d, want 2", p)
	}
	late := effectstest.B12Creature(g, me.ID, "Late Bear", "Creature — Bear", 2, 2)
	if p := effectstest.EffectivePower(t, g, late); p != 2 {
		t.Errorf("a creature that entered afterwards has power %d, want 2 (CR 611.2c)", p)
	}
}

func TestSocB14SurgeToVictoryCannotTargetASorceryOfAnotherPlayerOrACreatureCard(t *testing.T) {
	g, me, opp, _ := socB14SurgeTable(t)
	theirs := effectstest.GyCard(opp, "Divination", "Sorcery", "{2}{U}", effectstest.DivinationOracle)
	creature := effectstest.GyCard(me, "Grizzly Bears", "Creature — Bear", "{1}{G}", "")
	for name, id := range map[string]uuid.UUID{"another player's card": theirs, "a creature card": creature} {
		if err := effectstest.CastCatalogSpellErr(t, g, "Surge to Victory", "Sorcery", socB14SurgeOracle,
			[]game.TargetRef{{Kind: game.TargetCard, ID: id}}); err == nil {
			t.Errorf("Surge to Victory accepted %s", name)
		}
	}
}

func TestSocB14SurgeToVictoryCopiesAndCastsTheCardForEachCreatureThatConnects(t *testing.T) {
	g, me, opp, gy := socB14SurgeTable(t)
	a := effectstest.AuraBear(g, me.ID)
	effectstest.CastCatalogSpell(t, g, "Surge to Victory", "Sorcery", socB14SurgeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: gy}})
	effectstest.PassPriorityAroundTable(t, g)
	b := effectstest.B12Creature(g, me.ID, "Late Bear", "Creature — Bear", 2, 2)
	hand, life := me.Hand.Size(), opp.Life
	effectstest.AttackWith(t, g, opp.ID, a, b)
	if want := life - 5 - 2; opp.Life != want {
		t.Fatalf("the opponent is at %d, want %d", opp.Life, want)
	}
	casts := 0
	for i := 0; i < 8; i++ {
		effectstest.PassPriorityAroundTable(t, g)
		offer := effectstest.ResolutionCastOfferFor(g, me.ID)
		if offer == nil {
			break
		}
		cp := offer.DiscardOptions[0]
		// Strict mana with an empty pool: only a free cast gets through.
		if err := effectstest.CastThroughResolutionCast(t, g, me.ID, cp, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("casting the copy without paying its mana cost: %v", err)
		}
		casts++
	}
	effectstest.PassPriorityAroundTable(t, g)
	if casts != 2 {
		t.Fatalf("cast %d copies, want one for each of the two creatures that connected", casts)
	}
	if got := me.Hand.Size() - hand; got != 4 {
		t.Errorf("hand grew by %d, want 4 (two Divinations)", got)
	}
	if !effectstest.InExile(g, gy) {
		t.Error("the original card left exile")
	}
	if n := len(effectstest.CardCopies(g)); n != 0 {
		t.Errorf("%d copies are left over", n)
	}
}

func TestSocB14SurgeToVictoryCopyMayBeDeclined(t *testing.T) {
	g, me, opp, gy := socB14SurgeTable(t)
	a := effectstest.AuraBear(g, me.ID)
	effectstest.CastCatalogSpell(t, g, "Surge to Victory", "Sorcery", socB14SurgeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: gy}})
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.AttackWith(t, g, opp.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ResolutionCastOfferFor(g, me.ID) == nil {
		t.Fatal("no cast was offered")
	}
	effectstest.DeclineResolutionCast(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand grew by %d after declining", me.Hand.Size()-hand)
	}
	if n := len(effectstest.CardCopies(g)); n != 0 {
		t.Errorf("%d copies after declining, want none (CR 707.12)", n)
	}
	if !effectstest.InExile(g, gy) {
		t.Error("the original left exile")
	}
}

func TestSocB14SurgeToVictoryIgnoresNoncombatDamage(t *testing.T) {
	g, me, opp, gy := socB14SurgeTable(t)
	effectstest.CastCatalogSpell(t, g, "Surge to Victory", "Sorcery", socB14SurgeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: gy}})
	effectstest.PassPriorityAroundTable(t, g)
	// Noncombat damage to a player, then an opponent's creature hitting me.
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(uuid.Nil, opp.ID, 3); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ResolutionCastOfferFor(g, me.ID) != nil {
		t.Error("noncombat damage triggered the copy")
	}
}

// --- Unbound Flourishing ---

func socB14FlourishingTable(t *testing.T, copies int) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < copies; i++ {
		effectstest.B12Push(g, me.ID, "Unbound Flourishing", "Enchantment", socB14FlourishOracle, 0, 0)
	}
	return g, me, opp
}

func TestSocB14UnboundFlourishingDoublesXForAPermanentSpellWithXInItsCost(t *testing.T) {
	g, me, _ := socB14FlourishingTable(t, 1)
	_ = me
	id := effectstest.CastXSpell(t, g, "Primo, the Unbounded", "Legendary Creature — Fractal Wolf", socB10PrimoOracle, "{X}{G}{G}{U}", 3, nil)
	if item := g.StackMeta[id]; item == nil || item.XValue != 3 {
		t.Fatalf("X on the stack = %+v before the trigger resolves, want 3", item)
	}
	effectstest.PassPriorityAroundTable(t, g)
	ids := socB10Named(g, "Primo, the Unbounded")
	if len(ids) != 1 {
		t.Fatal("Primo did not enter")
	}
	// X = 6 once doubled; Primo enters with twice X counters.
	if n := effectstest.CountersOn(g, ids[0], game.CounterPlusOne); n != 12 {
		t.Errorf("Primo has %d counters, want twice the doubled X = 12", n)
	}
}

func TestSocB14UnboundFlourishingTwoOfThemQuadrupleX(t *testing.T) {
	g, _, _ := socB14FlourishingTable(t, 2)
	effectstest.CastXSpell(t, g, "Primo, the Unbounded", "Legendary Creature — Fractal Wolf", socB10PrimoOracle, "{X}{G}{G}{U}", 2, nil)
	socB14SettleKeepingTargets(t, g, g.Seats[0].ID, game.TargetRef{})
	ids := socB10Named(g, "Primo, the Unbounded")
	if len(ids) != 1 {
		t.Fatal("Primo did not enter")
	}
	if n := effectstest.CountersOn(g, ids[0], game.CounterPlusOne); n != 16 {
		t.Errorf("Primo has %d counters, want twice X = 2*8 = 16 with X doubled twice", n)
	}
}

func TestSocB14UnboundFlourishingLeavesAPermanentSpellWithoutXAlone(t *testing.T) {
	g, me, _ := socB14FlourishingTable(t, 1)
	id := effectstest.CastCatalogSpell(t, g, "Test Bear", "Creature — Bear", "", nil)
	if item := g.StackMeta[id]; item == nil || item.XValue != 0 {
		t.Fatalf("a creature spell without X has X = %+v", item)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.DelayedTriggers) != 0 || me.Life != 40 {
		t.Error("something happened for a spell without X")
	}
}

func TestSocB14UnboundFlourishingCopiesAnXInstantWithTheSameX(t *testing.T) {
	g, me, opp := socB14FlourishingTable(t, 1)
	before := opp.Life
	id := effectstest.CastXSpell(t, g, "Blaze", "Sorcery", effectstest.BlazeOracle, "{X}{R}", 3,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	if item := g.StackMeta[id]; item == nil || item.XValue != 3 {
		t.Fatalf("X = %+v, want 3: only permanent spells have X doubled", item)
	}
	answered := socB14SettleKeepingTargets(t, g, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if answered != 1 {
		t.Errorf("answered %d new-target prompts, want 1 copy", answered)
	}
	if got := opp.Life; got != before-6 {
		t.Errorf("the opponent is at %d, want %d (Blaze for 3 and a copy for 3)", got, before-6)
	}
	if got := effectstest.GraveyardSize(g, me.ID); got != 1 {
		t.Errorf("graveyard %d, want only the original Blaze: a copy is not a card", got)
	}
}

func TestSocB14UnboundFlourishingCopiesXOfZeroToo(t *testing.T) {
	g, me, opp := socB14FlourishingTable(t, 1)
	effectstest.CastXSpell(t, g, "Blaze", "Sorcery", effectstest.BlazeOracle, "{X}{R}", 0,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	answered := socB14SettleKeepingTargets(t, g, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if answered != 1 {
		t.Errorf("answered %d new-target prompts, want 1 copy even at X = 0", answered)
	}
}

func TestSocB14UnboundFlourishingIgnoresASpellWithoutXAndAnOpponentsXSpell(t *testing.T) {
	g, me, opp := socB14FlourishingTable(t, 1)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	if answered := socB14SettleKeepingTargets(t, g, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); answered != 0 {
		t.Errorf("a Bolt without X was copied (%d prompts)", answered)
	}
	if opp.Life != 37 {
		t.Errorf("opponent at %d, want 37", opp.Life)
	}

	// An opponent's Flourishing copies nothing of mine.
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	effectstest.B12Push(g2, opp2.ID, "Unbound Flourishing", "Enchantment", socB14FlourishOracle, 0, 0)
	effectstest.CastXSpell(t, g2, "Blaze", "Sorcery", effectstest.BlazeOracle, "{X}{R}", 2,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp2.ID}})
	if answered := socB14SettleKeepingTargets(t, g2, me2.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp2.ID}); answered != 0 {
		t.Errorf("a spell I cast was copied by an opponent's Flourishing (%d prompts)", answered)
	}
	if opp2.Life != 38 {
		t.Errorf("opponent at %d, want 38", opp2.Life)
	}
}

func TestSocB14UnboundFlourishingTwoOfThemMakeTwoCopies(t *testing.T) {
	g, me, opp := socB14FlourishingTable(t, 2)
	before := opp.Life
	effectstest.CastXSpell(t, g, "Blaze", "Sorcery", effectstest.BlazeOracle, "{X}{R}", 2,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	answered := socB14SettleKeepingTargets(t, g, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if answered != 2 {
		t.Errorf("answered %d new-target prompts, want 2 copies", answered)
	}
	if got := opp.Life; got != before-6 {
		t.Errorf("the opponent is at %d, want %d", got, before-6)
	}
}

func TestSocB14UnboundFlourishingCopiesAnAbilityWithXInItsCost(t *testing.T) {
	g, me, opp := socB14FlourishingTable(t, 1)
	kite := effectstest.PushCatalogPermanent(g, me.ID, "Crimson Hellkite", "Creature — Dragon", effectstest.CrimsonHellkiteOracle, false)
	victim := effectstest.B12Push(g, opp.ID, "Big Target", "Creature — Giant", "", 1, 9)
	at := game.TargetRef{Kind: game.TargetCard, ID: victim}
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, kite, 0, game.ActivateAbilityParams{Strict: true, XValue: 2, Targets: []game.TargetRef{at}}); err != nil {
		t.Fatalf("activate the Hellkite for X = 2: %v", err)
	}
	if answered := socB14SettleKeepingTargets(t, g, me.ID, at); answered != 1 {
		t.Errorf("answered %d new-target prompts, want 1 copy", answered)
	}
	if c, ok := effectstest.BattlefieldCard(g, victim); !ok || c.DamageMarked != 4 {
		t.Errorf("the target has %d damage, want 2 + 2 from the copy with the same X", c.DamageMarked)
	}
}

func TestSocB14UnboundFlourishingIgnoresAnAbilityWithoutXInItsCost(t *testing.T) {
	g, me, _ := socB14FlourishingTable(t, 1)
	troyan := effectstest.PushCatalogPermanent(g, me.ID, "Troyan, Gutsy Explorer", "Legendary Creature — Vedalken Scout", socB14TroyanOracle, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "U")
	hand := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, troyan, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	// One loot, not two: the copy would draw again.
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want a single loot", got)
	}
}
