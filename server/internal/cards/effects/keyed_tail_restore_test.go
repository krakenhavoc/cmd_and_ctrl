package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// keyed_tail_restore_test.go — #4253 slice 3: a CR 616 prompt whose
// event carries a continuation is a restore point once that
// continuation is data (game/keyed_tail.go). Each test takes a STRICT
// restore (RestoreStrict, the boot path) while the prompt is open,
// answers it on the restored table, and checks what a player sees
// afterwards.

// searchEventsSince counts EventSearchLibrary from `from` on, and
// reports whether the last one says the library was shuffled.
func searchEventsSince(g *game.Game, from int) (n int, shuffled bool) {
	for _, ev := range g.Events[from:] {
		if ev.Kind == game.EventSearchLibrary {
			n++
			shuffled = ev.Label == game.SearchShuffledLabel
		}
	}
	return n, shuffled
}

// --- a library search's shuffle tail -------------------------------------

// Polluted Delta cracked for Watery Grave: the shockland's "pay 2 life?"
// is a restore point, and the restored answer finishes the fetch — the
// land enters as answered, the search event fires once and the library
// is shuffled, which is what the search's continuation still owed.
func TestAFetchedShocklandIsARestorePoint(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		delta := pushCatalogPermanent(g, me.ID, "Polluted Delta", "Land", pollutedDeltaOracle, false)
		grave := pushLibraryCardForTest(me, game.Card{
			Name: "Watery Grave", TypeLine: "Land — Island Swamp", OracleID: shocklandOracleIDs["Watery Grave"],
		})
		for i := 0; i < 3; i++ {
			stapleLibraryCard(me, "Mountain", "Basic Land — Mountain")
		}
		if err := g.ActivateCatalogAbility(me.ID, delta, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		passPriorityAroundTable(t, g)
		takeSearchPrompts(t, g)
		if entryPayLifeChoiceFor(g, me.ID) == nil {
			t.Fatal("the fetched Watery Grave did not ask")
		}
		before := len(g.Events)
		lifeBefore := me.Life

		g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveTail)
		me = g.Seats[0]
		if _, ok := battlefieldCard(g, grave); ok {
			t.Fatal("the restored table landed the Grave before the answer")
		}
		if n, _ := searchEventsSince(g, 0); n != 0 {
			t.Fatal("the restored table finished the search before the answer")
		}
		answerEntryPayLife(t, g, me.ID, pay)
		card, ok := battlefieldCard(g, grave)
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
		if n, shuffled := searchEventsSince(g, 0); n != 1 || !shuffled {
			t.Errorf("pay %v: %d search events (shuffled %v) after the answer, want one that shuffled", pay, n, shuffled)
		}
		_ = before
	}
}

// A two-card battlefield search whose takes are both shocklands: each
// take's question is a restore point in turn, the second asked only once
// the first has entered, and the search finishes once, after the last.
func TestATwoCardSearchIsARestorePointAtEachTake(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := seedSearchLibrary(me,
		game.Card{Name: "Blood Crypt", TypeLine: "Land — Swamp Mountain", OracleID: bloodCryptOracle},
		game.Card{Name: "Steam Vents", TypeLine: "Land — Island Mountain", OracleID: steamVentsOracle},
		game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"},
	)
	g.WithWriteLock(func() {
		if err := g.SearchLibraryForEffect(me.ID, func(c game.Card) bool { return c.HasSubtype("Mountain") && c.Name != "Mountain" },
			game.ZoneBattlefield, 2, false, true); err != nil {
			t.Fatalf("SearchLibraryForEffect: %v", err)
		}
	})
	takeSearchPrompts(t, g)

	var paid uuid.UUID
	for take := 0; take < 2; take++ {
		c := entryPayLifeChoiceFor(g, me.ID)
		if c == nil {
			t.Fatalf("take %d: no pay-life prompt", take)
		}
		if take == 0 {
			paid = c.Source
		}
		g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveTail)
		me = g.Seats[g.Turn.ActiveSeat]
		if n, _ := searchEventsSince(g, 0); n != 0 {
			t.Fatalf("take %d: the search finished before its last take", take)
		}
		answerEntryPayLife(t, g, me.ID, take == 0)
	}
	if c := entryPayLifeChoiceFor(g, me.ID); c != nil {
		t.Fatalf("a third question is open: %+v", c)
	}
	entered := 0
	for _, id := range ids[:2] {
		card, ok := battlefieldCard(g, id)
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
	if _, ok := battlefieldCard(g, ids[2]); ok {
		t.Error("the basic Mountain was not a take")
	}
	if n, shuffled := searchEventsSince(g, 0); n != 1 || !shuffled {
		t.Errorf("%d search events (shuffled %v), want one that shuffled", n, shuffled)
	}
}

// A search whose "then" is a registered tail keeps it across the
// restore: the tail is told the cards that entered.
var keyedSearchThenProbe = game.RegisterTail("test/search-then-probe", func(g *game.Game, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return g.AddCounterForEffect(o.IDs[0], game.CounterPlusOne, o.Tail.N)
})

func TestASearchsKeyedThenRunsAfterARestore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := seedSearchLibrary(me,
		game.Card{Name: "Blood Crypt", TypeLine: "Land — Swamp Mountain", OracleID: bloodCryptOracle},
	)
	g.WithWriteLock(func() {
		if err := g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
			Player: me.ID, Pred: func(c game.Card) bool { return c.Name == "Blood Crypt" },
			Dest: game.ZoneBattlefield, Limit: 1, Shuffle: true,
			ThenTail: game.KeyedTail{Ref: keyedSearchThenProbe, N: 2},
		}); err != nil {
			t.Fatalf("SearchLibraryThenForEffect: %v", err)
		}
	})
	takeSearchPrompts(t, g)
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	answerEntryPayLife(t, g, me.ID, true)
	card, ok := battlefieldCard(g, ids[0])
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	fountain := pushLibraryCardForTest(me, game.Card{
		Name: "Hallowed Fountain", TypeLine: "Land — Plains Island", OracleID: hallowedFountainOracle,
	})
	crypt := pushLibraryCardForTest(me, game.Card{
		Name: "Blood Crypt", TypeLine: "Land — Swamp Mountain", OracleID: bloodCryptOracle,
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
		c := entryPayLifeChoiceFor(g, me.ID)
		if c == nil {
			t.Fatalf("question %d is not open", q)
		}
		if q == 0 {
			paid = c.Source
		}
		g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveTail)
		me = g.Seats[0]
		if g.Battlefield.Contains(fountain) || g.Battlefield.Contains(crypt) {
			t.Fatalf("question %d: a card of the batch landed before the last answer", q)
		}
		answerEntryPayLife(t, g, me.ID, q == 0)
	}
	for _, id := range []uuid.UUID{fountain, crypt} {
		card, ok := battlefieldCard(g, id)
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

// libraryShuffles counts the plain shuffles (ShuffleLibraryForEffect).
func libraryShuffles(g *game.Game) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventSearchLibrary && ev.Label == "shuffle" {
			n++
		}
	}
	return n
}

// Sword of Hearth and Home puts a blinked Clone and a basic land onto
// the battlefield together, then shuffles. The Clone's "enter as a
// copy?" is a restore point, and the restored answer lands both and
// shuffles once: the Sword's registered tail.
func TestSwordOfHearthAndHomeIsARestorePointAtACloneQuestion(t *testing.T) {
	g, me, opp, bear, _, plains, _ := ebSwordSetup(t)
	clone := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Clone", TypeLine: "Creature — Shapeshifter", OracleID: oracleClone,
		Owner: me.ID, Controller: me.ID,
	})
	dealCombatDamageToPlayer(g, bear, opp.ID, 4)
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, clone)
	passPriorityAroundTable(t, g)
	if c := searchChoiceFor(g, me.ID); c != nil {
		answerSearchByID(t, g, me.ID, plains)
	}
	if copyPrompt(g) == nil {
		t.Fatal("the returning Clone did not ask")
	}
	before := libraryShuffles(g)

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveTail)
	if g.Battlefield.Contains(plains) {
		t.Fatal("the restored table landed the Plains before the Clone's answer")
	}
	bearName := ""
	if card, ok := battlefieldCard(g, bear); ok {
		bearName = card.Name
	}
	c := copyPrompt(g)
	if err := g.ResolveCopyTarget(c.ID, c.Chooser, bear); err != nil {
		t.Fatalf("ResolveCopyTarget: %v", err)
	}
	if !g.Battlefield.Contains(plains) {
		t.Error("the Plains did not enter with the Clone")
	}
	if n := len(battlefieldIDsNamed(g, bearName)); n != 2 {
		t.Errorf("battlefield has %d %s, want the bear and the Clone copying it", n, bearName)
	}
	if n := libraryShuffles(g) - before; n != 1 {
		t.Errorf("%d shuffles after the answer, want one", n)
	}
}

// --- created tokens ---------------------------------------------------------

// Doubling Season beside Academy Manufactor over one Treasure: the CR
// 616.1 ordering prompt on the creation instruction is a restore point,
// and either restored order makes two of each kind (CR 701.7b).
func TestATokenCreationOrderingPromptIsARestorePoint(t *testing.T) {
	for _, seasonFirst := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0].ID
		ds := pushTokenReplacementCard(g, doublingSeasonOracle, "Doubling Season", "Enchantment", me)
		am := pushTokenReplacementCard(g, academyManufactorOracle, "Academy Manufactor", "Artifact Creature — Assembly-Worker", me)
		makeTokens(t, g, me, TreasureToken(), 1)
		if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != game.PendingChoiceReplacementOrder {
			t.Fatalf("the creation did not ask for the order: %+v", g.PendingChoices)
		}

		g = strictRestoreWithFrame(t, g, game.ReplacementResumeCreateTokens)
		if n := onBattlefieldNamed(g, "Treasure"); n != 0 {
			t.Fatalf("the restored table made %d Treasures before the answer", n)
		}
		prompt := g.PendingChoices[0]
		first, second := ds, am
		if !seasonFirst {
			first, second = am, ds
		}
		order := []game.ReplacementEffectID{
			replacementIDForSource(t, g, prompt.ReplacementEffectIDs, first),
			replacementIDForSource(t, g, prompt.ReplacementEffectIDs, second),
		}
		if err := g.ResolveReplacementOrder(prompt.ID, me, order); err != nil {
			t.Fatalf("ResolveReplacementOrder: %v", err)
		}
		for _, name := range []string{"Clue", "Food", "Treasure"} {
			if got := onBattlefieldNamed(g, name); got != 2 {
				t.Errorf("season first %v: %ss = %d, want 2", seasonFirst, name, got)
			}
		}
	}
}

// keyedTokenThenProbe taps every token it is told about, so a test can
// read that the creation's keyed continuation ran with the tokens.
var keyedTokenThenProbe = game.RegisterTail("test/token-then-probe", func(g *game.Game, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if err := g.TapTargetForEffect(id); err != nil {
			return err
		}
	}
	return nil
})

// A creation whose continuation is a registered tail keeps it across a
// restore at the ordering prompt: the tail is told every token made.
func TestATokenCreationsKeyedThenRunsAfterARestore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	ds := pushTokenReplacementCard(g, doublingSeasonOracle, "Doubling Season", "Enchantment", me)
	am := pushTokenReplacementCard(g, academyManufactorOracle, "Academy Manufactor", "Artifact Creature — Assembly-Worker", me)
	g.WithWriteLock(func() {
		if err := g.CreateTokensKeyedForEffect(game.TokenCreation{
			Controller: me, Groups: []game.TokenGroup{{Template: TreasureToken(), Count: 1}},
		}, game.KeyedTail{Ref: keyedTokenThenProbe}); err != nil {
			t.Fatalf("CreateTokensKeyedForEffect: %v", err)
		}
	})
	g = strictRestoreWithFrame(t, g, game.ReplacementResumeCreateTokens)
	prompt := g.PendingChoices[0]
	order := []game.ReplacementEffectID{
		replacementIDForSource(t, g, prompt.ReplacementEffectIDs, ds),
		replacementIDForSource(t, g, prompt.ReplacementEffectIDs, am),
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
// frame), and the second is asked only after the first has entered.
func TestATokenEntryQuestionIsARestorePointForEachToken(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bears := seedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	clone := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Clone", TypeLine: "Creature — Shapeshifter", OracleID: oracleClone,
		Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() {
		tmpl, ok := TokenCopyTemplate(g, clone)
		if !ok {
			t.Fatal("TokenCopyTemplate failed")
		}
		if err := g.CreateTokenForEffect(me.ID, tmpl, 2); err != nil {
			t.Fatalf("CreateTokenForEffect: %v", err)
		}
	})
	for q := 0; q < 2; q++ {
		if copyPrompt(g) == nil {
			t.Fatalf("token %d did not ask", q)
		}
		g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveTail)
		if n := len(battlefieldIDsNamed(g, "Grizzly Bears")); n != 1+q {
			t.Fatalf("token %d: %d Grizzly Bears on the restored table, want %d", q, n, 1+q)
		}
		c := copyPrompt(g)
		if c == nil {
			t.Fatalf("token %d: the restored table has no question", q)
		}
		if err := g.ResolveCopyTarget(c.ID, c.Chooser, bears); err != nil {
			t.Fatalf("ResolveCopyTarget: %v", err)
		}
	}
	copies := 0
	for _, id := range battlefieldIDsNamed(g, "Grizzly Bears") {
		if id != bears && id != clone {
			copies++
		}
	}
	if copies != 2 {
		t.Errorf("%d new Grizzly Bears, want both token copies", copies)
	}
	if c := copyPrompt(g); c != nil {
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
		g := newCatalogGame(t)
		p0 := g.Seats[0].ID
		hs := seedReplacementPermanent(g, hardenedScalesOracle, "Hardened Scales", p0)
		ds := seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", p0)
		bear := seedCreature(g, "Bears", p0)
		g.WithWriteLock(func() {
			if err := g.MonstrosityForEffect(bear, 2); err != nil {
				t.Fatalf("MonstrosityForEffect: %v", err)
			}
		})

		g = strictRestoreWithFrame(t, g, game.ReplacementResumeCounterTail)
		if g.IsMonstrous(bear) {
			t.Fatal("the restored table made it monstrous before the order")
		}
		if tc.scalesFirst {
			answerOrderBySource(t, g, hs, ds)
		} else {
			answerOrderBySource(t, g, ds, hs)
		}
		if got := countersOn(g, bear, "+1/+1"); got != tc.want {
			t.Errorf("scales first %v: %d counters, want %d", tc.scalesFirst, got, tc.want)
		}
		if !g.IsMonstrous(bear) {
			t.Errorf("scales first %v: the creature did not become monstrous", tc.scalesFirst)
		}
	}
}

// keyedLifeThenProbe puts as many +1/+1 counters on Carry[0] as the life
// change moved, so a test can read the amount the tail was told.
var keyedLifeThenProbe = game.RegisterTail("test/life-then-probe", func(g *game.Game, o game.TailOutcome) error {
	return g.AddCounterForEffect(o.Tail.Carry[0], game.CounterPlusOne, o.Amount)
})

// A life gain whose continuation is a registered tail, under Rhox
// Faithmender and Alhammarret's Archive: the restored order gains ×4 and
// the tail is told the 12 that moved.
func TestAKeyedLifeTailRunsAfterARestoredOrder(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	faith := seedReplacementPermanent(g, rhoxFaithmenderOracle, "Rhox Faithmender", me.ID)
	archive := seedReplacementPermanent(g, alhammarretsArchiveOracle, "Alhammarret's Archive", me.ID)
	bear := seedCreature(g, "Bears", me.ID)
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeKeyedForEffect(uuid.Nil, me.ID, 3,
			game.KeyedTail{Ref: keyedLifeThenProbe, Carry: []uuid.UUID{bear}}); err != nil {
			t.Fatalf("ChangePlayerLifeKeyedForEffect: %v", err)
		}
	})
	g = strictRestoreWithFrame(t, g, game.ReplacementResumeLifeTail)
	answerOrderBySource(t, g, archive, faith)
	if got := countersOn(g, bear, "+1/+1"); got != 12 {
		t.Errorf("the tail was told %d, want 12", got)
	}
}

// lifeLossPlusOneProbeOracle is a test card: "If an opponent would lose
// life, they lose that much life plus 1 instead." Beside Bloodletter of
// Aclazotz it makes each opponent's loss during your turn an ordering
// prompt, which no two printed cards in the catalog do yet.
const lifeLossPlusOneProbeOracle = "test-4253-life-loss-plus-one"

func init() {
	Register(Spec{
		OracleID: lifeLossPlusOneProbeOracle,
		Name:     "Life Loss Probe",
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventChangeLife},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventLife && ev.LifeDelta < 0 && src != nil && ev.LifePlayer != src.Controller
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.LifeDelta--
				return nil
			},
			Label: "Life Loss Probe: lose one more",
		}},
	})
}

// Extort's drain (each opponent loses 1, you gain the life lost) under
// two different life-loss replacements: each opponent's ordering prompt
// is a restore point in turn, with the running total carried, and the
// gain once the last has answered is what they really lost.
func TestADrainIsARestorePointAtEachOpponentsOrder(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	blood := seedReplacementPermanent(g, bloodletterOracle, "Bloodletter of Aclazotz", me.ID)
	probe := seedReplacementPermanent(g, lifeLossPlusOneProbeOracle, "Life Loss Probe", me.ID)
	lifeBefore := me.Life
	g.WithWriteLock(func() {
		if err := b21DrainEachOpponentAndGainTheTotal(g, &game.StackItem{Controller: me.ID, SourceCardID: blood}, 1); err != nil {
			t.Fatalf("drain: %v", err)
		}
	})
	// Bloodletter first: (1×2)+1 = 3. The probe first: (1+1)×2 = 4.
	want := 0
	for q := 0; q < 3; q++ {
		if len(g.PendingChoices) != 1 {
			t.Fatalf("opponent %d: %d questions open, want the one ordering prompt", q, len(g.PendingChoices))
		}
		g = strictRestoreWithFrame(t, g, game.ReplacementResumeLifeTail)
		if q%2 == 0 {
			answerOrderBySource(t, g, blood, probe)
			want += 3
		} else {
			answerOrderBySource(t, g, probe, blood)
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

// --- a face-down exile batch (#4257, #4261) ----------------------------------

// exileOrderProbeOracles are two test cards, each "if a card named Needle
// would be put into exile, it is put into exile instead": two different
// declared effects, so a Needle's exile asks for their order (CR 616.1).
// They are catalog rows, so a restored table gathers them again.
var exileOrderProbeOracles = []string{"test-4253-exile-order-a", "test-4253-exile-order-b"}

func init() {
	for i, oracle := range exileOrderProbeOracles {
		Register(Spec{
			OracleID: oracle,
			Name:     "Exile Order Probe " + string(rune('A'+i)),
			Replacements: []game.ReplacementEffect{{
				Watches: []game.EventKind{game.EventZoneMove},
				AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) bool {
					if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneExile {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && c.Name == "Needle"
				},
				Replace: func(*game.ReplacementEvent, *game.Game, *game.Card) error { return nil },
				Label:   "Exile Order Probe " + string(rune('A'+i)),
			}},
		})
	}
}

// seedExileOrderProbes puts both probes on the battlefield under `who`.
func seedExileOrderProbes(g *game.Game, who uuid.UUID) {
	for i, oracle := range exileOrderProbeOracles {
		seedReplacementPermanent(g, oracle, "Exile Order Probe "+string(rune('A'+i)), who)
	}
}

// Necropotence's "exile the top card of your library face down" over a
// Needle: the ordering prompt is a restore point, and the restored answer
// exiles it face down and schedules its delivery, Necropotence's
// registered tail, which brings it back at the end step.
func TestANecropotenceExileOrderingPromptIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	necro := seedReplacementPermanent(g, necropotenceOracle, "Necropotence", me.ID)
	needle := pushLibraryTopForTest(me, "Needle")
	seedExileOrderProbes(g, me.ID)
	if err := g.ActivateCatalogAbility(me.ID, necro, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 1 {
		t.Fatalf("%d prompts open, want the exile's ordering prompt", len(g.PendingChoices))
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveTail)
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("%d delayed triggers before the answer", len(g.DelayedTriggers))
	}
	answerExileOrder(t, g)
	exiled, ok := findInZone(g.Exile, needle)
	if !ok || !exiled.FaceDown {
		t.Fatalf("the Needle is not exiled face down after the restored answer (found %v)", ok)
	}
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("%d delayed triggers after the restored answer, want 1", len(g.DelayedTriggers))
	}
	deliverAtOwnEndStep(t, g, g.Seats[0], needle)
}

// The plain batch (#4257): three cards off the top face down, the second a
// Needle. Its ordering prompt is a restore point with the first card in
// exile and the third still on the library, and the restored answer
// finishes the batch.
func TestAFaceDownExileBatchIsARestorePointMidBatch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushLibraryTopForTest(me, "Third")
	needle := pushLibraryTopForTest(me, "Needle")
	first := pushLibraryTopForTest(me, "First")
	third := me.Library.Cards[len(me.Library.Cards)-3].InstanceID
	seedExileOrderProbes(g, me.ID)
	libBefore := me.Library.Size()
	g.WithWriteLock(func() {
		if _, err := g.ExileTopFaceDownForEffect(me.ID, 3); err != nil {
			t.Fatalf("ExileTopFaceDownForEffect: %v", err)
		}
	})
	if len(g.PendingChoices) != 1 {
		t.Fatalf("%d prompts open, want the Needle's ordering prompt", len(g.PendingChoices))
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveTail)
	me = g.Seats[0]
	if _, ok := findInZone(g.Exile, first); !ok {
		t.Fatal("the first card is not in exile on the restored table")
	}
	if !me.Library.Contains(needle) || !me.Library.Contains(third) {
		t.Fatal("the restored table moved the Needle or the third card before the answer")
	}
	answerExileOrder(t, g)
	for _, id := range []uuid.UUID{first, needle, third} {
		c, ok := findInZone(g.Exile, id)
		if !ok || !c.FaceDown {
			t.Errorf("%s is not exiled face down (found %v)", id, ok)
		}
	}
	if me.Library.Size() != libBefore-3 {
		t.Errorf("library %d, want %d", me.Library.Size(), libBefore-3)
	}
}

// --- a keyword action with a "then" ------------------------------------------

// strictRestoreWithScryThen takes a strict restore and checks the open
// look-at was written with its keyed "then".
func strictRestoreWithScryThen(t *testing.T, g *game.Game) *game.Game {
	t.Helper()
	snap := g.CaptureSnapshot()
	found := false
	for _, c := range snap.PendingChoices {
		if (c.Kind == game.PendingChoiceScry || c.Kind == game.PendingChoiceSurveil) && c.PickThen != "" {
			found = true
		}
	}
	if !found {
		t.Fatalf("no look-at in the snapshot carries a keyed then: %+v", snap.Continuations)
	}
	return restoreRoundTrip(t, g, true)
}

// Preordain's "Scry 2, then draw a card": the scry prompt is a restore
// point, and the restored answer draws the card left on top.
func TestAPreordainScryIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Preordain", "Sorcery", preordainOracle, nil)
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) == nil {
		t.Fatal("Preordain did not scry")
	}

	g = strictRestoreWithScryThen(t, g)
	me = g.Seats[g.Turn.ActiveSeat]
	c := scryChoiceFor(g, me.ID)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Consider", "Instant", considerOracle, nil)
	passPriorityAroundTable(t, g)
	if surveilChoiceFor(g, me.ID) == nil {
		t.Fatal("Consider did not surveil")
	}

	g = strictRestoreWithScryThen(t, g)
	me = g.Seats[g.Turn.ActiveSeat]
	c := surveilChoiceFor(g, me.ID)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedReplacementPermanent(g, scryPlusOneProbeOracle, "Scrying Probe", me.ID)
	seedReplacementPermanent(g, scryDoubledProbeOracle, "Doubling Probe", me.ID)
	g.WithWriteLock(func() {
		if err := (Scry{N: 2, ThenTail: game.KeyedTail{Ref: thenDrawACard, Player: me.ID}}).Apply(
			NewContext(g, &game.StackItem{Controller: me.ID})); err != nil {
			t.Fatalf("Scry: %v", err)
		}
	})

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeKeywordActionTail)
	me = g.Seats[0]
	prompt := latestChoiceOfKindFor(g, game.PendingChoiceReplacementOrder, me.ID)
	if prompt == nil {
		t.Fatal("the restored table has no ordering prompt")
	}
	if err := g.ResolveReplacementOrder(prompt.ID, me.ID, replacementIDsByLabel(t, g, prompt.ReplacementEffectIDs,
		[]string{"Doubling Probe — scry twice as many", "Scrying Probe — scry one more"})); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}

	g = strictRestoreWithScryThen(t, g)
	me = g.Seats[0]
	c := scryChoiceFor(g, me.ID)
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
