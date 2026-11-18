package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_neo_b18_test.go — the seven transforming Sagas of S82 batch 18
// (#5493), through the imported two-faced card.

func neoB18Row(oracleID, front, frontCost, back, backType, power, toughness string, colors []string) cards.Card {
	return transformRow(oracleID, front, "Enchantment — Saga", frontCost,
		back, backType, power, toughness, colors)
}

func neoB18BoseijuRow() cards.Card {
	return neoB18Row(boseijuReachesSkywardOracleID, "Boseiju Reaches Skyward", "{3}{G}",
		"Branch of Boseiju", "Enchantment Creature — Plant", "0", "0", []string{"G"})
}

func neoB18MichikoRow() cards.Card {
	return neoB18Row(michikosReignOfTruthOracleID, "Michiko's Reign of Truth", "{1}{W}",
		"Portrait of Michiko", "Enchantment Creature — Human Noble", "0", "0", []string{"W"})
}

func neoB18LongReachRow() cards.Card {
	return neoB18Row(theLongReachOfNightOracleID, "The Long Reach of Night", "{3}{B}",
		"Animus of Night's Reach", "Enchantment Creature — Spirit", "0", "4", []string{"B"})
}

func neoB18KondaRow() cards.Card {
	return neoB18Row(theFallOfLordKondaOracleID, "The Fall of Lord Konda", "{2}{W}",
		"Fragment of Konda", "Enchantment Creature — Human Noble", "1", "3", []string{"W"})
}

func neoB18BeholdRow() cards.Card {
	return neoB18Row(beholdTheUnspeakableOracleID, "Behold the Unspeakable", "{3}{U}{U}",
		"Vision of the Unspeakable", "Enchantment Creature — Spirit", "0", "0", []string{"U"})
}

func neoB18TributeRow() cards.Card {
	return neoB18Row(tributeToHorobiOracleID, "Tribute to Horobi", "{1}{B}",
		"Echo of Death's Wail", "Enchantment Creature — Spirit", "3", "3", []string{"B"})
}

func neoB18TeachingsRow() cards.Card {
	return neoB18Row(teachingsOfTheKirinOracleID, "Teachings of the Kirin", "{1}{G}",
		"Kirin-Touched Orochi", "Enchantment Creature — Snake Monk", "1", "1", []string{"G"})
}

// neoB18BackOnBattlefield puts the back face straight onto the battlefield
// under `me`, not summoning sick.
func neoB18BackOnBattlefield(t *testing.T, g *game.Game, me *game.Player, row cards.Card) uuid.UUID {
	t.Helper()
	id := importToHand(row, me)
	var entered uuid.UUID
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == id {
				me.Hand.Cards[i].SetFace(1)
			}
		}
		var err error
		entered, err = g.PutFromHandOntoBattlefieldForEffect(id, game.HandEntryOptions{})
		if err != nil {
			t.Fatalf("put the back face onto the battlefield: %v", err)
		}
		findBattlefieldCardForTest(g, entered).SummonedThisTurn = false
		g.RecomputeLayersIfStaleLocked()
	})
	return entered
}

// neoB18Settle answers a trigger-target prompt with `target` when one is
// asked, then lets the stack resolve.
func neoB18Settle(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID) {
	t.Helper()
	passPriorityAroundTable(t, g)
	if target != uuid.Nil && latestPickTarget(g, me.ID) != nil {
		pickTriggerTarget(t, g, me.ID, target)
		passPriorityAroundTable(t, g)
	}
}

// neoB18NextChapter runs the Saga's controller to their next precombat
// main, where the next lore counter is added, and resolves the chapter.
func neoB18NextChapter(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID) {
	t.Helper()
	seat := g.Turn.ActiveSeat
	advanceToPrecombatMainOf(t, g, seat)
	neoB18Settle(t, g, me, target)
}

func neoB18Back(g *game.Game, name string) *game.Card {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == name {
			return &g.Battlefield.Cards[i]
		}
	}
	return nil
}

func neoB18PT(g *game.Game, id uuid.UUID) (int, int) {
	c := battlefieldCardFor(g, id)
	return c.CurrentPower(), c.CurrentToughness()
}

// finalChapter checks chapter III left a transformed, new object.
func neoB18CheckFlip(t *testing.T, g *game.Game, me *game.Player, saga uuid.UUID, back string) *game.Card {
	t.Helper()
	if battlefieldCardFor(g, saga) != nil {
		t.Fatal("the original object is still on the battlefield")
	}
	c := neoB18Back(g, back)
	if c == nil {
		t.Fatalf("%s never reached the battlefield", back)
	}
	if c.ActiveFace != 1 || c.Controller != me.ID || c.Counters[game.CounterLore] != 0 {
		t.Errorf("came back face %d, controller %s, %d lore counters", c.ActiveFace, c.Controller, c.Counters[game.CounterLore])
	}
	assertFaceInvariant(t, g)
	return c
}

func TestBoseijuReachesSkyward(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	var forests []uuid.UUID
	for i := 0; i < 3; i++ {
		forests = append(forests, pushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"}))
	}
	pushLibraryCardForTest(me, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
	deadLand := pushGraveyardCardTyped(me, "Dead Land", "Land")
	pushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	handBefore := me.Hand.Size()

	saga := importAndCast(t, g, neoB18BoseijuRow(), me)
	passPriorityAroundTable(t, g)
	// Chapter I: a search prompt for up to two basic Forests.
	search := latestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, me.ID)
	if search == nil {
		t.Fatal("no search prompt for chapter I")
	}
	if err := g.ResolveSearchLibrary(search.ID, me.ID, forests[:2]); err != nil {
		t.Fatalf("search: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - handBefore; got != 2 {
		t.Fatalf("hand grew by %d after chapter I, want the two Forests", got)
	}
	for _, id := range forests[:2] {
		if !handHas(me, id) {
			t.Errorf("Forest %s is not in hand", id)
		}
	}

	// Chapter II: put the land card on top of the library.
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("chapter II asked for no target")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: deadLand}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if top := libraryTopIDs(me, 1); len(top) != 1 || top[0] != deadLand {
		t.Errorf("top of library is %v, want the land card from the graveyard", top)
	}

	// Chapter III. Without a land the 0/0 Branch would die at once.
	for i := 0; i < 2; i++ {
		apaPush(g, me.ID, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	}
	neoB18NextChapter(t, g, me, uuid.Nil)
	back := neoB18CheckFlip(t, g, me, saga, "Branch of Boseiju")
	lands := b03LandsControlled(g, me.ID)
	if p, tough := neoB18PT(g, back.InstanceID); p != lands || tough != lands || !game.HasKeyword(back, "reach") {
		t.Errorf("Branch is %d/%d with %d lands, want +1/+1 per land and reach", p, tough, lands)
	}
}

func TestBranchOfBoseijuCountsLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	branch := neoB18BackOnBattlefield(t, g, me, neoB18BoseijuRow())
	if p, _ := neoB18PT(g, branch); p != 0 {
		t.Fatalf("no lands: power %d, want 0", p)
	}
	for i := 0; i < 3; i++ {
		apaPush(g, me.ID, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	}
	apaPush(g, g.Seats[1].ID, g.Seats[1].ID, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if p, tough := neoB18PT(g, branch); p != 3 || tough != 3 {
		t.Errorf("three lands: %d/%d, want 3/3 (an opponent's land doesn't count)", p, tough)
	}
}

func TestMichikosReignOfTruth(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	apaPush(g, me.ID, me.ID, game.Card{Name: "Relic", TypeLine: "Artifact", ManaCost: "{1}"})
	apaPush(g, me.ID, me.ID, game.Card{Name: "Relic Aura", TypeLine: "Enchantment Artifact"})
	apaPush(g, g.Seats[1].ID, g.Seats[1].ID, game.Card{Name: "Their Relic", TypeLine: "Artifact"})
	bear := apaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})

	saga := importAndCast(t, g, neoB18MichikoRow(), me)
	neoB18Settle(t, g, me, bear)
	// The Saga is an enchantment: 3 permanents (relic, relic aura, Saga).
	if p, tough := neoB18PT(g, bear); p != 5 || tough != 5 {
		t.Errorf("after chapter I the Bear is %d/%d, want 5/5", p, tough)
	}
	neoB18NextChapter(t, g, me, bear)
	neoB18NextChapter(t, g, me, uuid.Nil)
	back := neoB18CheckFlip(t, g, me, saga, "Portrait of Michiko")
	// relic, relic aura and the Portrait itself.
	if p, tough := neoB18PT(g, back.InstanceID); p != 3 || tough != 3 {
		t.Errorf("Portrait is %d/%d, want 3/3", p, tough)
	}
}

func TestFallOfLordKonda(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	small := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Small", TypeLine: "Creature — Bear", ManaCost: "{3}", Power: 2, Toughness: 2})
	big := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Big", TypeLine: "Creature — Giant", ManaCost: "{2}{G}{G}", Power: 4, Toughness: 4})
	mine := apaPush(g, me.ID, me.ID, game.Card{Name: "Mine", TypeLine: "Creature — Giant", ManaCost: "{6}", Power: 6, Toughness: 6})
	stolen := apaPush(g, opp.ID, me.ID, game.Card{Name: "Stolen", TypeLine: "Creature — Bear", Power: 1, Toughness: 1})
	mineTaken := apaPush(g, me.ID, opp.ID, game.Card{Name: "Taken", TypeLine: "Artifact"})

	saga := importAndCast(t, g, neoB18KondaRow(), me)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		pick := latestPickTarget(g, me.ID)
		for _, bad := range []uuid.UUID{small, mine} {
			if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
				t.Errorf("targeted an illegal creature %s", bad)
			}
		}
	}
	neoB18Settle(t, g, me, big)
	if !g.Exile.Contains(big) {
		t.Error("chapter I did not exile the mana value 4 creature")
	}
	if !g.Battlefield.Contains(small) || !g.Battlefield.Contains(mine) {
		t.Error("chapter I exiled something it should not have")
	}

	neoB18NextChapter(t, g, me, uuid.Nil)
	if c := battlefieldCardFor(g, stolen); c == nil || c.Controller != opp.ID {
		t.Error("chapter II did not return the stolen creature to its owner")
	}
	if c := battlefieldCardFor(g, mineTaken); c == nil || c.Controller != me.ID {
		t.Error("chapter II did not return my permanent to me")
	}

	neoB18NextChapter(t, g, me, uuid.Nil)
	back := neoB18CheckFlip(t, g, me, saga, "Fragment of Konda")
	if p, tough := neoB18PT(g, back.InstanceID); p != 1 || tough != 3 || !game.HasKeyword(back, "defender") {
		t.Errorf("Fragment is %d/%d, want a 1/3 defender", p, tough)
	}
}

func TestFragmentOfKondaDrawsWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushLibraryCardForTest(me, game.Card{Name: "Top", TypeLine: "Land"})
	frag := neoB18BackOnBattlefield(t, g, me, neoB18KondaRow())
	hand := me.Hand.Size()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(frag); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards when the Fragment died, want 1", got)
	}
}

func TestBeholdTheUnspeakable(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	theirs := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Theirs", TypeLine: "Creature — Bear", Power: 3, Toughness: 3})
	mine := apaPush(g, me.ID, me.ID, game.Card{Name: "Mine", TypeLine: "Creature — Bear", Power: 3, Toughness: 3})
	for i := 0; i < 8; i++ {
		pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Land"})
	}

	saga := importAndCast(t, g, neoB18BeholdRow(), me)
	passPriorityAroundTable(t, g)
	if p, _ := neoB18PT(g, theirs); p != 1 {
		t.Errorf("their creature has power %d, want 1", p)
	}
	if p, _ := neoB18PT(g, mine); p != 3 {
		t.Errorf("my creature has power %d, want 3", p)
	}
	late := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Late", TypeLine: "Creature — Bear", Power: 3, Toughness: 3})
	if p, _ := neoB18PT(g, late); p != 3 {
		t.Errorf("a creature that arrived later has power %d, want 3", p)
	}

	// Chapter II with a full hand: scry 2, then draw two.
	for me.Hand.Size() < 3 {
		me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
	}
	hand := me.Hand.Size()
	advanceToPrecombatMainOf(t, g, seat)
	// Until my next turn: the shrink ended as my turn began.
	if p, _ := neoB18PT(g, theirs); p != 3 {
		t.Errorf("their creature has power %d on my next turn, want 3 again", p)
	}
	passPriorityAroundTable(t, g)
	scry := latestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID)
	if scry == nil {
		t.Fatal("chapter II with three cards in hand did not scry")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		// The turn's own draw only; the chapter's draw waits for the scry.
		t.Fatalf("hand grew by %d before the scry was answered, want just the draw step's 1", got)
	}
	if err := g.ResolveScry(scry.ID, me.ID, nil, scry.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("hand grew by %d, want the draw step's 1 and the chapter's 2", got)
	}

	// Chapter III with one card in hand would draw four; here just flip.
	neoB18NextChapter(t, g, me, uuid.Nil)
	neoB18CheckFlip(t, g, me, saga, "Vision of the Unspeakable")
}

func TestBeholdTheUnspeakableDrawsFourWithAnEmptyHand(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	for i := 0; i < 12; i++ {
		pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Land"})
	}
	importAndCast(t, g, neoB18BeholdRow(), me)
	passPriorityAroundTable(t, g)
	// The draw step puts exactly one card in hand before chapter II.
	me.Hand.Cards = nil
	hand := 0
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID) != nil {
		t.Fatal("scried with one or fewer cards in hand")
	}
	if got := me.Hand.Size() - hand; got != 5 {
		t.Errorf("hand grew by %d, want the draw step's 1 and four more", got)
	}
}

func TestVisionOfTheUnspeakableCountsHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	v := neoB18BackOnBattlefield(t, g, me, neoB18BeholdRow())
	for me.Hand.Size() < 4 {
		me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
	}
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	n := me.Hand.Size()
	if p, tough := neoB18PT(g, v); p != n || tough != n {
		t.Errorf("Vision is %d/%d with %d cards in hand, want %d/%d", p, tough, n, n, n)
	}
}

func handHas(p *game.Player, id uuid.UUID) bool {
	for _, c := range p.Hand.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

func TestTheLongReachOfNight(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	pitcher, sacker, empty := g.Seats[(seat+1)%4], g.Seats[(seat+2)%4], g.Seats[(seat+3)%4]
	pitcher.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Pitch", TypeLine: "Land", Owner: pitcher.ID, Controller: pitcher.ID})
	sacker.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Keep", TypeLine: "Land", Owner: sacker.ID, Controller: sacker.ID})
	pitchBear := apaPush(g, pitcher.ID, pitcher.ID, game.Card{Name: "Pitcher Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	sackBear := apaPush(g, sacker.ID, sacker.ID, game.Card{Name: "Sacker Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	emptyBear := apaPush(g, empty.ID, empty.ID, game.Card{Name: "Empty Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	myBear := apaPush(g, me.ID, me.ID, game.Card{Name: "My Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	for _, p := range []*game.Player{empty} {
		p.Hand.Cards = nil
	}

	saga := importAndCast(t, g, neoB18LongReachRow(), me)
	passPriorityAroundTable(t, g)
	// An opponent with no cards in hand can't discard, so sacrifices.
	// Each of the others is asked in turn.
	answerOptionPick(t, g, pitcher.ID, 1)
	answerDiscard(t, g, pitcher.ID, pitcher.Hand.Cards[0].InstanceID)
	answerOptionPick(t, g, sacker.ID, 0)
	answerSacrifice(t, g, sacker.ID, sackBear)
	answerSacrifice(t, g, empty.ID, emptyBear)
	passPriorityAroundTable(t, g)
	for _, tc := range []struct {
		name string
		id   uuid.UUID
		want bool
	}{{"pitcher's creature", pitchBear, true}, {"sacrificer's creature", sackBear, false}, {"hand-less opponent's creature", emptyBear, false}, {"my creature", myBear, true}} {
		if got := g.Battlefield.Contains(tc.id); got != tc.want {
			t.Errorf("%s on the battlefield = %v, want %v", tc.name, got, tc.want)
		}
	}
	if pitcher.Graveyard.Size() != 1 {
		t.Errorf("the discarding opponent has %d cards in the graveyard, want 1", pitcher.Graveyard.Size())
	}

	// Chapter II: the discarder has nothing left to discard.
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, pitcher.ID, 0)
	answerSacrifice(t, g, pitcher.ID, pitchBear)
	answerOptionPick(t, g, sacker.ID, 1)
	answerDiscard(t, g, sacker.ID, sacker.Hand.Cards[0].InstanceID)
	// Every opponent drew on the way round the table, so the last is asked too.
	answerOptionPick(t, g, empty.ID, 0)
	answerSacrifice(t, g, empty.ID, emptyBear)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(pitchBear) {
		t.Error("chapter II: the opponent with no cards in hand kept their creature")
	}
	if !g.Battlefield.Contains(myBear) {
		t.Error("chapter II took my creature")
	}
	if g.Battlefield.Contains(emptyBear) {
		t.Error("chapter II: the third opponent kept their creature")
	}
	neoB18NextChapter(t, g, me, uuid.Nil)
	neoB18CheckFlip(t, g, me, saga, "Animus of Night's Reach")
}

func TestAnimusOfNightsReachPumpsByDefendersCreatureCards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	animus := neoB18BackOnBattlefield(t, g, me, neoB18LongReachRow())
	for i := 0; i < 3; i++ {
		pushGraveyardCardTyped(opp, "Dead Bear", "Creature — Bear")
	}
	pushGraveyardCardTyped(opp, "Dead Spell", "Instant")
	for i := 0; i < 5; i++ {
		pushGraveyardCardTyped(other, "Other Bear", "Creature — Bear")
	}
	if c := battlefieldCardFor(g, animus); !game.HasKeyword(c, "menace") {
		t.Error("the Animus has no menace")
	}
	neoB18Attack(t, g, animus, opp)
	passPriorityAroundTable(t, g)
	if p, _ := neoB18PT(g, animus); p != 3 {
		t.Errorf("Animus power %d after attacking, want 0+3 (the defender's creature cards only)", p)
	}
}

func TestTributeToHorobi(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	saga := importAndCast(t, g, neoB18TributeRow(), me)
	passPriorityAroundTable(t, g)
	countRats := func(p *game.Player) int {
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Rat Rogue" && c.Controller == p.ID && c.IsToken() {
				n++
			}
		}
		return n
	}
	for _, p := range g.Seats {
		want := 1
		if p.ID == me.ID {
			want = 0
		}
		if got := countRats(p); got != want {
			t.Errorf("after chapter I %s controls %d Rat Rogue tokens, want %d", p.ID, got, want)
		}
	}
	neoB18NextChapter(t, g, me, uuid.Nil)
	for _, p := range g.Seats {
		want := 2
		if p.ID == me.ID {
			want = 0
		}
		if got := countRats(p); got != want {
			t.Errorf("after chapter II %s controls %d Rat Rogue tokens, want %d", p.ID, got, want)
		}
	}
	// A non-token Rat is not "a Rat token".
	nonToken := apaPush(g, g.Seats[1].ID, g.Seats[1].ID, game.Card{Name: "Real Rat", TypeLine: "Creature — Rat", Power: 1, Toughness: 1})
	neoB18NextChapter(t, g, me, uuid.Nil)
	echo := neoB18CheckFlip(t, g, me, saga, "Echo of Death's Wail")
	passPriorityAroundTable(t, g)
	if got := countRats(me); got != 3*2 {
		t.Errorf("I control %d Rat Rogue tokens after the Echo entered, want all 6", got)
	}
	if c := battlefieldCardFor(g, nonToken); c == nil || c.Controller != g.Seats[1].ID {
		t.Error("the Echo took a Rat that isn't a token")
	}
	if p, tough := neoB18PT(g, echo.InstanceID); p != 3 || tough != 3 || !game.HasKeyword(echo, "flying") || !game.HasKeyword(echo, "haste") {
		t.Errorf("Echo is %d/%d, want a 3/3 flying haste", p, tough)
	}
}

func TestEchoOfDeathsWailAttackSacrificeDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	echo := neoB18BackOnBattlefield(t, g, me, neoB18TributeRow())
	fodder := apaPush(g, me.ID, me.ID, game.Card{Name: "Fodder", TypeLine: "Creature — Bear", Power: 1, Toughness: 1})
	pushLibraryCardForTest(me, game.Card{Name: "Top", TypeLine: "Land"})
	hand := me.Hand.Size()
	neoB18Attack(t, g, echo, opp)
	passPriorityAroundTable(t, g)
	answerOwnPermanents(t, g, me.ID, fodder)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the other creature was not sacrificed")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}

func TestTeachingsOfTheKirin(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	for i := 0; i < 5; i++ {
		pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Land"})
	}
	bear := apaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	gy := me.Graveyard.Size()
	lib := me.Library.Size()

	saga := importAndCast(t, g, neoB18TeachingsRow(), me)
	passPriorityAroundTable(t, g)
	if got := me.Graveyard.Size() - gy; got != 3 {
		t.Errorf("milled %d cards, want 3", got)
	}
	if got := lib - me.Library.Size(); got != 3 {
		t.Errorf("library shrank by %d, want 3", got)
	}
	if got := countBattlefieldByName(g, "Spirit"); got != 1 {
		t.Errorf("%d Spirit tokens, want 1", got)
	}

	neoB18NextChapter(t, g, me, bear)
	if got := counterOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("Bear has %d +1/+1 counters after chapter II, want 1", got)
	}

	neoB18NextChapter(t, g, me, uuid.Nil)
	neoB18CheckFlip(t, g, me, saga, "Kirin-Touched Orochi")
}

func TestKirinTouchedOrochiAttackModes(t *testing.T) {
	for _, mode := range []int{0, 1} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		orochi := neoB18BackOnBattlefield(t, g, me, neoB18TeachingsRow())
		bear := apaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
		deadCreature := pushGraveyardCardTyped(opp, "Dead Bear", "Creature — Bear")
		deadSpell := pushGraveyardCardTyped(opp, "Dead Spell", "Instant")

		neoB18Attack(t, g, orochi, opp)
		answerMode(t, g, modePickChoiceFor(g, me.ID), me.ID, mode)
		if mode == 0 {
			neoB18Settle(t, g, me, deadCreature)
		} else {
			neoB18Settle(t, g, me, deadSpell)
		}
		if mode == 1 {
			neoB18Settle(t, g, me, bear)
		}
		passPriorityAroundTable(t, g)
		spirits := countBattlefieldByName(g, "Spirit")
		counters := counterOn(g, bear, game.CounterPlusOne)
		if mode == 0 {
			if !g.Exile.Contains(deadCreature) || spirits != 1 || counters != 0 {
				t.Errorf("mode 1: exiled=%v spirits=%d counters=%d, want exile, 1 Spirit, no counter", g.Exile.Contains(deadCreature), spirits, counters)
			}
		} else if !g.Exile.Contains(deadSpell) || spirits != 0 || counters != 1 {
			t.Errorf("mode 2: exiled=%v spirits=%d counters=%d, want exile, no Spirit, 1 counter", g.Exile.Contains(deadSpell), spirits, counters)
		}
	}
}

// neoB18Attack declares `attacker` at `opp` and locks the attack in, which
// puts the attack triggers on the stack.
func neoB18Attack(t *testing.T, g *game.Game, attacker uuid.UUID, opp *game.Player) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
}
