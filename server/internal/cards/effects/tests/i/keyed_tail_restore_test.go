package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- a library search's shuffle tail -------------------------------------

// Polluted Delta cracked for Watery Grave: the shockland's "pay 2 life?"
// is a restore point, and the restored answer finishes the fetch — the
// land enters as answered, the search event fires once and the library
// is shuffled, which is what the search's continuation still owed.
func TestAFetchedShocklandIsARestorePoint(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		delta := effectstest.PushCatalogPermanent(g, me.ID, "Polluted Delta", "Land", effectstest.PollutedDeltaOracle, false)
		grave := effectstest.PushLibraryCardForTest(me, game.Card{
			Name: "Watery Grave", TypeLine: "Land — Island Swamp", OracleID: effectstest.ShocklandOracleIDs["Watery Grave"],
		})
		for i := 0; i < 3; i++ {
			effectstest.StapleLibraryCard(me, "Mountain", "Basic Land — Mountain")
		}
		if err := g.ActivateCatalogAbility(me.ID, delta, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.TakeSearchPrompts(t, g)
		if effectstest.EntryPayLifeChoiceFor(g, me.ID) == nil {
			t.Fatal("the fetched Watery Grave did not ask")
		}
		before := len(g.Events)
		lifeBefore := me.Life

		g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeMoveTail)
		me = g.Seats[0]
		if _, ok := effectstest.BattlefieldCard(g, grave); ok {
			t.Fatal("the restored table landed the Grave before the answer")
		}
		if n, _ := effectstest.SearchEventsSince(g, 0); n != 0 {
			t.Fatal("the restored table finished the search before the answer")
		}
		effectstest.AnswerEntryPayLife(t, g, me.ID, pay)
		card, ok := effectstest.BattlefieldCard(g, grave)
		if !ok {
			t.Fatal("the restored answer did not land the Grave")
		}
		if card.Tapped == pay {
			t.Errorf("pay %v: the Grave entered tapped=%v", pay, card.Tapped)
		}
		wantLife := lifeBefore
		if pay {
			wantLife -= 2
		}
		if me.Life != wantLife {
			t.Errorf("pay %v: life %d, want %d", pay, me.Life, wantLife)
		}
		if n, shuffled := effectstest.SearchEventsSince(g, 0); n != 1 || !shuffled {
			t.Errorf("pay %v: %d search events (shuffled %v) after the answer, want one that shuffled", pay, n, shuffled)
		}
		_ = before
	}
}

// A two-card battlefield search whose takes are both shocklands: each
// take's question is a restore point in turn, the second asked only once
// the first has entered, and the search finishes once, after the last.
func TestATwoCardSearchIsARestorePointAtEachTake(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Blood Crypt", TypeLine: "Land — Swamp Mountain", OracleID: effectstest.BloodCryptOracle},
		game.Card{Name: "Steam Vents", TypeLine: "Land — Island Mountain", OracleID: effectstest.SteamVentsOracle},
		game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"},
	)
	g.WithWriteLock(func() {
		if err := g.SearchLibraryForEffect(me.ID, func(c game.Card) bool { return c.HasSubtype("Mountain") && c.Name != "Mountain" },
			game.ZoneBattlefield, 2, false, true); err != nil {
			t.Fatalf("SearchLibraryForEffect: %v", err)
		}
	})
	effectstest.TakeSearchPrompts(t, g)

	var paid uuid.UUID
	for take := 0; take < 2; take++ {
		c := effectstest.EntryPayLifeChoiceFor(g, me.ID)
		if c == nil {
			t.Fatalf("take %d: no pay-life prompt", take)
		}
		if take == 0 {
			paid = c.Source
		}
		g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeMoveTail)
		me = g.Seats[g.Turn.ActiveSeat]
		if n, _ := effectstest.SearchEventsSince(g, 0); n != 0 {
			t.Fatalf("take %d: the search finished before its last take", take)
		}
		effectstest.AnswerEntryPayLife(t, g, me.ID, take == 0)
	}
	if c := effectstest.EntryPayLifeChoiceFor(g, me.ID); c != nil {
		t.Fatalf("a third question is open: %+v", c)
	}
	entered := 0
	for _, id := range ids[:2] {
		card, ok := effectstest.BattlefieldCard(g, id)
		if !ok {
			continue
		}
		entered++
		if card.Tapped == (id == paid) {
			t.Errorf("%s entered tapped=%v, want the first paid and the second declined", card.Name, card.Tapped)
		}
	}
	if entered != 2 {
		t.Fatalf("%d of the two takes entered", entered)
	}
	if _, ok := effectstest.BattlefieldCard(g, ids[2]); ok {
		t.Error("the basic Mountain was not a take")
	}
	if n, shuffled := effectstest.SearchEventsSince(g, 0); n != 1 || !shuffled {
		t.Errorf("%d search events (shuffled %v), want one that shuffled", n, shuffled)
	}
}

func TestASearchsKeyedThenRunsAfterARestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Blood Crypt", TypeLine: "Land — Swamp Mountain", OracleID: effectstest.BloodCryptOracle},
	)
	g.WithWriteLock(func() {
		if err := g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
			Player: me.ID, Pred: func(c game.Card) bool { return c.Name == "Blood Crypt" },
			Dest: game.ZoneBattlefield, Limit: 1, Shuffle: true,
			ThenTail: game.KeyedTail{Ref: effectstest.KeyedSearchThenProbe, N: 2},
		}); err != nil {
			t.Fatalf("SearchLibraryThenForEffect: %v", err)
		}
	})
	effectstest.TakeSearchPrompts(t, g)
	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	effectstest.AnswerEntryPayLife(t, g, me.ID, true)
	card, ok := effectstest.BattlefieldCard(g, ids[0])
	if !ok {
		t.Fatal("the Crypt did not enter")
	}
	if card.Counters[game.CounterPlusOne] != 2 {
		t.Errorf("the keyed then put %d counters, want 2", card.Counters[game.CounterPlusOne])
	}
}

// --- a simultaneous entry (entry_batch.go) ---------------------------------

// Two shocklands put onto the battlefield together (Genesis Wave's
// shape, through the synchronous door): each card's question is a
// restore point, the second restored with the first's settled answer
// held in the batch, and neither lands until both have answered.
func TestAPutBatchIsARestorePointAtEachQuestion(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	fountain := effectstest.PushLibraryCardForTest(me, game.Card{
		Name: "Hallowed Fountain", TypeLine: "Land — Plains Island", OracleID: effectstest.HallowedFountainOracle,
	})
	crypt := effectstest.PushLibraryCardForTest(me, game.Card{
		Name: "Blood Crypt", TypeLine: "Land — Swamp Mountain", OracleID: effectstest.BloodCryptOracle,
	})
	lifeBefore := me.Life
	g.WithWriteLock(func() {
		if _, err := g.PutCardsFromLibraryOntoBattlefieldForEffect([]uuid.UUID{fountain, crypt},
			game.LibraryEntryOptions{Controller: me.ID}); err != nil {
			t.Fatalf("put: %v", err)
		}
	})
	var paid uuid.UUID
	for q := 0; q < 2; q++ {
		c := effectstest.EntryPayLifeChoiceFor(g, me.ID)
		if c == nil {
			t.Fatalf("question %d is not open", q)
		}
		if q == 0 {
			paid = c.Source
		}
		g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeMoveTail)
		me = g.Seats[0]
		if g.Battlefield.Contains(fountain) || g.Battlefield.Contains(crypt) {
			t.Fatalf("question %d: a card of the batch landed before the last answer", q)
		}
		effectstest.AnswerEntryPayLife(t, g, me.ID, q == 0)
	}
	for _, id := range []uuid.UUID{fountain, crypt} {
		card, ok := effectstest.BattlefieldCard(g, id)
		if !ok {
			t.Fatalf("%s did not enter", id)
		}
		if card.Tapped == (id == paid) {
			t.Errorf("%s entered tapped=%v, want the first paid and the second declined", card.Name, card.Tapped)
		}
	}
	if me.Life != lifeBefore-2 {
		t.Errorf("life %d, want %d", me.Life, lifeBefore-2)
	}
}

// Sword of Hearth and Home puts a blinked Clone and a basic land onto
// the battlefield together, then shuffles. The Clone's "enter as a
// copy?" is a restore point, and the restored answer lands both and
// shuffles once: the Sword's registered tail.
func TestSwordOfHearthAndHomeIsARestorePointAtACloneQuestion(t *testing.T) {
	g, me, opp, bear, _, plains, _ := effectstest.EbSwordSetup(t)
	clone := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Clone", TypeLine: "Creature — Shapeshifter", OracleID: effectstest.OracleClone,
		Owner: me.ID, Controller: me.ID,
	})
	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 4)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, clone)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.SearchChoiceFor(g, me.ID); c != nil {
		effectstest.AnswerSearchByID(t, g, me.ID, plains)
	}
	if effectstest.CopyPrompt(g) == nil {
		t.Fatal("the returning Clone did not ask")
	}
	before := effectstest.LibraryShuffles(g)

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeMoveTail)
	if g.Battlefield.Contains(plains) {
		t.Fatal("the restored table landed the Plains before the Clone's answer")
	}
	bearName := ""
	if card, ok := effectstest.BattlefieldCard(g, bear); ok {
		bearName = card.Name
	}
	c := effectstest.CopyPrompt(g)
	if err := g.ResolveCopyTarget(c.ID, c.Chooser, bear); err != nil {
		t.Fatalf("ResolveCopyTarget: %v", err)
	}
	if !g.Battlefield.Contains(plains) {
		t.Error("the Plains did not enter with the Clone")
	}
	if n := len(effectstest.BattlefieldIDsNamed(g, bearName)); n != 2 {
		t.Errorf("battlefield has %d %s, want the bear and the Clone copying it", n, bearName)
	}
	if n := effectstest.LibraryShuffles(g) - before; n != 1 {
		t.Errorf("%d shuffles after the answer, want one", n)
	}
}

// --- created tokens ---------------------------------------------------------

// Doubling Season beside Academy Manufactor over one Treasure: the CR
// 616.1 ordering prompt on the creation instruction is a restore point,
// and either restored order makes two of each kind (CR 701.7b).
func TestATokenCreationOrderingPromptIsARestorePoint(t *testing.T) {
	for _, seasonFirst := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0].ID
		ds := effectstest.PushTokenReplacementCard(g, effectstest.DoublingSeasonOracle, "Doubling Season", "Enchantment", me)
		am := effectstest.PushTokenReplacementCard(g, effectstest.AcademyManufactorOracle, "Academy Manufactor", "Artifact Creature — Assembly-Worker", me)
		effectstest.MakeTokens(t, g, me, effects.TreasureToken(), 1)
		if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != game.PendingChoiceReplacementOrder {
			t.Fatalf("the creation did not ask for the order: %+v", g.PendingChoices)
		}

		g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeCreateTokens)
		if n := effectstest.OnBattlefieldNamed(g, "Treasure"); n != 0 {
			t.Fatalf("the restored table made %d Treasures before the answer", n)
		}
		prompt := g.PendingChoices[0]
		first, second := ds, am
		if !seasonFirst {
			first, second = am, ds
		}
		order := []game.ReplacementEffectID{
			effectstest.ReplacementIDForSource(t, g, prompt.ReplacementEffectIDs, first),
			effectstest.ReplacementIDForSource(t, g, prompt.ReplacementEffectIDs, second),
		}
		if err := g.ResolveReplacementOrder(prompt.ID, me, order); err != nil {
			t.Fatalf("ResolveReplacementOrder: %v", err)
		}
		for _, name := range []string{"Clue", "Food", "Treasure"} {
			if got := effectstest.OnBattlefieldNamed(g, name); got != 2 {
				t.Errorf("season first %v: %ss = %d, want 2", seasonFirst, name, got)
			}
		}
	}
}

// A creation whose continuation is a registered tail keeps it across a
// restore at the ordering prompt: the tail is told every token made.
func TestATokenCreationsKeyedThenRunsAfterARestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0].ID
	ds := effectstest.PushTokenReplacementCard(g, effectstest.DoublingSeasonOracle, "Doubling Season", "Enchantment", me)
	am := effectstest.PushTokenReplacementCard(g, effectstest.AcademyManufactorOracle, "Academy Manufactor", "Artifact Creature — Assembly-Worker", me)
	g.WithWriteLock(func() {
		if err := g.CreateTokensKeyedForEffect(game.TokenCreation{
			Controller: me, Groups: []game.TokenGroup{{Template: effects.TreasureToken(), Count: 1}},
		}, game.KeyedTail{Ref: effectstest.KeyedTokenThenProbe}); err != nil {
			t.Fatalf("CreateTokensKeyedForEffect: %v", err)
		}
	})
	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeCreateTokens)
	prompt := g.PendingChoices[0]
	order := []game.ReplacementEffectID{
		effectstest.ReplacementIDForSource(t, g, prompt.ReplacementEffectIDs, ds),
		effectstest.ReplacementIDForSource(t, g, prompt.ReplacementEffectIDs, am),
	}
	if err := g.ResolveReplacementOrder(prompt.ID, me, order); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
	tapped := 0
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Tapped {
			tapped++
		}
	}
	if tapped != 6 {
		t.Errorf("the keyed continuation tapped %d tokens, want all 6", tapped)
	}
}

// Two token copies of Clone: each token's "enter as a copy?" is asked in
// turn, each is a restore point (the staged tokens are written with the
// frame), and, since the two are created together (#4867), both are asked
// before either enters: the frame is the simultaneous entry's.
func TestATokenEntryQuestionIsARestorePointForEachToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bears := effectstest.SeedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	clone := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Clone", TypeLine: "Creature — Shapeshifter", OracleID: effectstest.OracleClone,
		Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() {
		tmpl, ok := effects.TokenCopyTemplate(g, clone)
		if !ok {
			t.Fatal("TokenCopyTemplate failed")
		}
		if err := g.CreateTokenForEffect(me.ID, tmpl, 2); err != nil {
			t.Fatalf("CreateTokenForEffect: %v", err)
		}
	})
	for q := 0; q < 2; q++ {
		if effectstest.CopyPrompt(g) == nil {
			t.Fatalf("token %d did not ask", q)
		}
		g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeMoveTokenBatch)
		if n := len(effectstest.BattlefieldIDsNamed(g, "Grizzly Bears")); n != 1 {
			t.Fatalf("token %d: %d Grizzly Bears on the restored table, want 1: neither token enters before both are asked", q, n)
		}
		c := effectstest.CopyPrompt(g)
		if c == nil {
			t.Fatalf("token %d: the restored table has no question", q)
		}
		if err := g.ResolveCopyTarget(c.ID, c.Chooser, bears); err != nil {
			t.Fatalf("ResolveCopyTarget: %v", err)
		}
	}
	copies := 0
	for _, id := range effectstest.BattlefieldIDsNamed(g, "Grizzly Bears") {
		if id != bears && id != clone {
			copies++
		}
	}
	if copies != 2 {
		t.Errorf("%d new Grizzly Bears, want both token copies", copies)
	}
	if c := effectstest.CopyPrompt(g); c != nil {
		t.Errorf("a third question is open: %+v", c)
	}
}

// --- counter placements and life changes with a tail -------------------------

// Monstrosity 2 under Hardened Scales and Doubling Season: the ordering
// prompt over the counters is a restore point, and the restored answer
// places them and makes the creature monstrous (CR 701.37a), the
// placement's registered tail.
func TestAMonstrosityOrderingPromptIsARestorePoint(t *testing.T) {
	for _, tc := range []struct {
		scalesFirst bool
		want        int
	}{{true, 6}, {false, 5}} {
		g := effectstest.NewCatalogGame(t)
		p0 := g.Seats[0].ID
		hs := effectstest.SeedReplacementPermanent(g, effectstest.HardenedScalesOracle, "Hardened Scales", p0)
		ds := effectstest.SeedReplacementPermanent(g, effectstest.DoublingSeasonOracle, "Doubling Season", p0)
		bear := effectstest.SeedCreature(g, "Bears", p0)
		g.WithWriteLock(func() {
			if err := g.MonstrosityForEffect(bear, 2); err != nil {
				t.Fatalf("MonstrosityForEffect: %v", err)
			}
		})

		g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeCounterTail)
		if g.IsMonstrous(bear) {
			t.Fatal("the restored table made it monstrous before the order")
		}
		if tc.scalesFirst {
			effectstest.AnswerOrderBySource(t, g, hs, ds)
		} else {
			effectstest.AnswerOrderBySource(t, g, ds, hs)
		}
		if got := effectstest.CountersOn(g, bear, "+1/+1"); got != tc.want {
			t.Errorf("scales first %v: %d counters, want %d", tc.scalesFirst, got, tc.want)
		}
		if !g.IsMonstrous(bear) {
			t.Errorf("scales first %v: the creature did not become monstrous", tc.scalesFirst)
		}
	}
}

// A life gain whose continuation is a registered tail, under Rhox
// Faithmender and Alhammarret's Archive: the restored order gains ×4 and
// the tail is told the 12 that moved.
func TestAKeyedLifeTailRunsAfterARestoredOrder(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	faith := effectstest.SeedReplacementPermanent(g, effectstest.RhoxFaithmenderOracle, "Rhox Faithmender", me.ID)
	archive := effectstest.SeedReplacementPermanent(g, effectstest.AlhammarretsArchiveOracle, "Alhammarret's Archive", me.ID)
	bear := effectstest.SeedCreature(g, "Bears", me.ID)
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeKeyedForEffect(uuid.Nil, me.ID, 3,
			game.KeyedTail{Ref: effectstest.KeyedLifeThenProbe, Carry: []uuid.UUID{bear}}); err != nil {
			t.Fatalf("ChangePlayerLifeKeyedForEffect: %v", err)
		}
	})
	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeLifeTail)
	effectstest.AnswerOrderBySource(t, g, archive, faith)
	if got := effectstest.CountersOn(g, bear, "+1/+1"); got != 12 {
		t.Errorf("the tail was told %d, want 12", got)
	}
}

// Extort's drain (each opponent loses 1, you gain the life lost) under
// two different life-loss replacements: each opponent's ordering prompt
// is a restore point in turn, with the running total carried, and the
// gain once the last has answered is what they really lost.
func TestADrainIsARestorePointAtEachOpponentsOrder(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	blood := effectstest.SeedReplacementPermanent(g, effectstest.BloodletterOracle, "Bloodletter of Aclazotz", me.ID)
	probe := effectstest.SeedReplacementPermanent(g, effectstest.LifeLossPlusOneProbeOracle, "Life Loss Probe", me.ID)
	lifeBefore := me.Life
	g.WithWriteLock(func() {
		if err := effects.B21DrainEachOpponentAndGainTheTotal(g, &game.StackItem{Controller: me.ID, SourceCardID: blood}, 1); err != nil {
			t.Fatalf("drain: %v", err)
		}
	})
	// Bloodletter first: (1×2)+1 = 3. The probe first: (1+1)×2 = 4.
	want := 0
	for q := 0; q < 3; q++ {
		if len(g.PendingChoices) != 1 {
			t.Fatalf("opponent %d: %d questions open, want the one ordering prompt", q, len(g.PendingChoices))
		}
		g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeLifeTail)
		if q%2 == 0 {
			effectstest.AnswerOrderBySource(t, g, blood, probe)
			want += 3
		} else {
			effectstest.AnswerOrderBySource(t, g, probe, blood)
			want += 4
		}
	}
	me = g.Seats[g.Turn.ActiveSeat]
	if me.Life != lifeBefore+want {
		t.Errorf("life %d, want %d: the gain is the life the opponents really lost", me.Life, lifeBefore+want)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("questions left open: %+v", g.PendingChoices)
	}
}

// Necropotence's "exile the top card of your library face down" over a
// Needle: the ordering prompt is a restore point, and the restored answer
// exiles it face down and schedules its delivery, Necropotence's
// registered tail, which brings it back at the end step.
func TestANecropotenceExileOrderingPromptIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	necro := effectstest.SeedReplacementPermanent(g, effectstest.NecropotenceOracle, "Necropotence", me.ID)
	needle := effectstest.PushLibraryTopForTest(me, "Needle")
	effectstest.SeedExileOrderProbes(g, me.ID)
	if err := g.ActivateCatalogAbility(me.ID, necro, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 1 {
		t.Fatalf("%d prompts open, want the exile's ordering prompt", len(g.PendingChoices))
	}

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeMoveTail)
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("%d delayed triggers before the answer", len(g.DelayedTriggers))
	}
	effectstest.AnswerExileOrder(t, g)
	exiled, ok := effectstest.FindInZone(g.Exile, needle)
	if !ok || !exiled.FaceDown {
		t.Fatalf("the Needle is not exiled face down after the restored answer (found %v)", ok)
	}
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("%d delayed triggers after the restored answer, want 1", len(g.DelayedTriggers))
	}
	effectstest.DeliverAtOwnEndStep(t, g, g.Seats[0], needle)
}

// The plain batch (#4257): three cards off the top face down, the second a
// Needle. Its ordering prompt is a restore point with the first card in
// exile and the third still on the library, and the restored answer
// finishes the batch.
func TestAFaceDownExileBatchIsARestorePointMidBatch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushLibraryTopForTest(me, "Third")
	needle := effectstest.PushLibraryTopForTest(me, "Needle")
	first := effectstest.PushLibraryTopForTest(me, "First")
	third := me.Library.Cards[len(me.Library.Cards)-3].InstanceID
	effectstest.SeedExileOrderProbes(g, me.ID)
	libBefore := me.Library.Size()
	g.WithWriteLock(func() {
		if _, err := g.ExileTopFaceDownForEffect(me.ID, 3); err != nil {
			t.Fatalf("ExileTopFaceDownForEffect: %v", err)
		}
	})
	if len(g.PendingChoices) != 1 {
		t.Fatalf("%d prompts open, want the Needle's ordering prompt", len(g.PendingChoices))
	}

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeMoveTail)
	me = g.Seats[0]
	if _, ok := effectstest.FindInZone(g.Exile, first); !ok {
		t.Fatal("the first card is not in exile on the restored table")
	}
	if !me.Library.Contains(needle) || !me.Library.Contains(third) {
		t.Fatal("the restored table moved the Needle or the third card before the answer")
	}
	effectstest.AnswerExileOrder(t, g)
	for _, id := range []uuid.UUID{first, needle, third} {
		c, ok := effectstest.FindInZone(g.Exile, id)
		if !ok || !c.FaceDown {
			t.Errorf("%s is not exiled face down (found %v)", id, ok)
		}
	}
	if me.Library.Size() != libBefore-3 {
		t.Errorf("library %d, want %d", me.Library.Size(), libBefore-3)
	}
}

// Preordain's "Scry 2, then draw a card": the scry prompt is a restore
// point, and the restored answer draws the card left on top.
func TestAPreordainScryIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "Preordain", "Sorcery", effectstest.PreordainOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Fatal("Preordain did not scry")
	}

	g = effectstest.StrictRestoreWithScryThen(t, g)
	me = g.Seats[g.Turn.ActiveSeat]
	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("the restored scry is %+v, want one over two cards", c)
	}
	hand := me.Hand.Size()
	// Keep the second card on top, send the first to the bottom.
	keep, bottom := c.ScryCards[1], c.ScryCards[0]
	if err := g.ResolveScry(c.ID, me.ID, []uuid.UUID{bottom}, []uuid.UUID{keep}); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	if me.Hand.Size() != hand+1 || !me.Hand.Contains(keep) {
		t.Errorf("hand %d -> %d (has the kept card %v), want the kept card drawn", hand, me.Hand.Size(), me.Hand.Contains(keep))
	}
}

// Consider's "Surveil 1, then draw a card": the surveil prompt is a
// restore point, and the restored answer draws the next card.
func TestAConsiderSurveilIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "Consider", "Instant", effectstest.ConsiderOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SurveilChoiceFor(g, me.ID) == nil {
		t.Fatal("Consider did not surveil")
	}

	g = effectstest.StrictRestoreWithScryThen(t, g)
	me = g.Seats[g.Turn.ActiveSeat]
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("the restored surveil is %+v, want one over a card", c)
	}
	looked := c.ScryCards[0]
	next := me.Library.Cards[len(me.Library.Cards)-2].InstanceID
	hand := me.Hand.Size()
	if err := g.ResolveSurveil(c.ID, me.ID, []uuid.UUID{looked}, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if !me.Graveyard.Contains(looked) {
		t.Error("the surveilled card is not in the graveyard")
	}
	if me.Hand.Size() != hand+1 || !me.Hand.Contains(next) {
		t.Errorf("hand %d -> %d, want the next card drawn", hand, me.Hand.Size())
	}
}

// A scry with a keyed "then" under two scry replacements: the ordering
// prompt is a restore point (the keyword tail carries the then), so is
// the scry it queues, and the draw still happens after the answer.
func TestAScryWithAKeyedThenIsARestorePointAtBothPrompts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedReplacementPermanent(g, effectstest.ScryPlusOneProbeOracle, "Scrying Probe", me.ID)
	effectstest.SeedReplacementPermanent(g, effectstest.ScryDoubledProbeOracle, "Doubling Probe", me.ID)
	g.WithWriteLock(func() {
		if err := (effects.Scry{N: 2, ThenTail: game.KeyedTail{Ref: effects.ThenDrawACard, Player: me.ID}}).Apply(
			effects.NewContext(g, &game.StackItem{Controller: me.ID})); err != nil {
			t.Fatalf("Scry: %v", err)
		}
	})

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeKeywordActionTail)
	me = g.Seats[0]
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceReplacementOrder, me.ID)
	if prompt == nil {
		t.Fatal("the restored table has no ordering prompt")
	}
	if err := g.ResolveReplacementOrder(prompt.ID, me.ID, effectstest.ReplacementIDsByLabel(t, g, prompt.ReplacementEffectIDs,
		[]string{"Doubling Probe — scry twice as many", "Scrying Probe — scry one more"})); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}

	g = effectstest.StrictRestoreWithScryThen(t, g)
	me = g.Seats[0]
	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 5 {
		t.Fatalf("the restored scry is %+v, want one over five cards", c)
	}
	hand := me.Hand.Size()
	if err := g.ResolveScry(c.ID, me.ID, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	if me.Hand.Size() != hand+1 || !me.Hand.Contains(c.ScryCards[0]) {
		t.Errorf("hand %d -> %d, want the top card drawn", hand, me.Hand.Size())
	}
}
