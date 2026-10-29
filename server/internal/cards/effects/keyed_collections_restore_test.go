package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// keyed_collections_restore_test.go — #4290, the second half: the
// remaining sacrifice-run callers, the choose_cards fan-outs, the
// permanent picks (ChoosePermanents) and ADR 0175's ordered collector are
// keyed, so a table waiting on any seat's leg is a restore point. Each
// test takes a STRICT restore (restoreRoundTrip with strict, which fails
// on a table that is not a restore point) while a prompt is open, with
// the first leg already answered where there is more than one, and
// finishes on the restored table. The *UnknownKeyIsRefused tests are the
// rollback case: a binary that does not know the key refuses the file.

// refuseUnknownRunKey rewrites every snapshot run naming `key` to an
// unregistered key and checks the restore is refused.
func refuseUnknownRunKey(t *testing.T, g *game.Game, key string) {
	t.Helper()
	snap := g.CaptureSnapshot()
	found := false
	for i := range snap.PromptRuns {
		if snap.PromptRuns[i].Key == key {
			snap.PromptRuns[i].Key = "prompt-run/never-registered"
			found = true
		}
	}
	if !found {
		t.Fatalf("the snapshot holds no run naming %q: %+v", key, snap.PromptRuns)
	}
	if _, err := snap.Restore(); !errors.Is(err, game.ErrUnknownEffectKey) {
		t.Fatalf("restore with an unknown run key: err = %v, want ErrUnknownEffectKey", err)
	}
}

// refuseUnknownCollectionKey is the same for an ordered collection.
func refuseUnknownCollectionKey(t *testing.T, g *game.Game, key string) {
	t.Helper()
	snap := g.CaptureSnapshot()
	found := false
	for i := range snap.PendingChoices {
		if o := snap.PendingChoices[i].Ordered; o != nil && o.Key == key {
			o.Key = "prompt-run/never-registered"
			found = true
		}
	}
	if !found {
		t.Fatalf("the snapshot holds no collection naming %q", key)
	}
	if _, err := snap.Restore(); !errors.Is(err, game.ErrUnknownEffectKey) {
		t.Fatalf("restore with an unknown collection key: err = %v, want ErrUnknownEffectKey", err)
	}
}

// --- sacrifice runs that read the asking item ---------------------------

// Victimize: the sacrifice is a keyed run's leg and the item's two targets
// ride the run as the asking item, so the table restores with the prompt
// open and the creatures still come back, tapped, after it.
func TestVictimizeSacrificeIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := pushGraveyardCardForTest(me, "Dead A")
	b := pushGraveyardCardForTest(me, "Dead B")
	fodder := seedCreature(g, "Fodder", me.ID)
	castCatalogSpell(t, g, "Victimize", "Sorcery", victimizeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	passPriorityAroundTable(t, g)
	if c := sacrificeChoiceFor(g, me.ID); c == nil || c.PickThen != victimizeReturn.Key() {
		t.Fatalf("the sacrifice prompt is %+v, want a leg of the keyed run", c)
	}

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	answerSacrifice(t, g, me.ID, fodder)
	if !me.Graveyard.Contains(fodder) {
		t.Error("the chosen creature was not sacrificed on the restored table")
	}
	for _, want := range []uuid.UUID{a, b} {
		card, ok := battlefieldCard(g, want)
		if !ok {
			t.Errorf("%s did not return to the battlefield after the restore", want)
			continue
		}
		if !card.Tapped {
			t.Errorf("%s returned untapped", want)
		}
	}
}

// The rollback case: the sacrifice leg names its run's key.
func TestVictimizeSacrificeWithAnUnknownKeyIsRefused(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := pushGraveyardCardForTest(me, "Dead A")
	b := pushGraveyardCardForTest(me, "Dead B")
	seedCreature(g, "Fodder", me.ID)
	castCatalogSpell(t, g, "Victimize", "Sorcery", victimizeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	passPriorityAroundTable(t, g)
	refuseUnknownKey(t, g, victimizeReturn.Key())
	refuseUnknownRunKey(t, g, victimizeReturn.Key())
}

// A target that left the graveyard while the table was restored is still
// re-checked against the card's clause (CR 608.2b): only the one still
// there returns.
func TestVictimizeRechecksItsTargetsAfterARestore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := pushGraveyardCardForTest(me, "Dead A")
	b := pushGraveyardCardForTest(me, "Dead B")
	fodder := seedCreature(g, "Fodder", me.ID)
	castCatalogSpell(t, g, "Victimize", "Sorcery", victimizeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	passPriorityAroundTable(t, g)
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	g.WithWriteLock(func() { me.Graveyard.Remove(a) })
	answerSacrifice(t, g, me.ID, fodder)
	if g.Battlefield.Contains(a) {
		t.Error("a target that left the graveyard returned")
	}
}

// Disciple of Bolas: the creature's power is filed as the step record
// under the asking item, which rides the run, so the gain and the draw
// are right after a restore taken with the prompt open.
func TestDiscipleOfBolasPayoutSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	beast := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Beast", TypeLine: "Creature — Beast",
		Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	castCatalogSpell(t, g, "Disciple of Bolas", "Creature — Human Wizard", discipleOfBolasOracle, nil)
	passPriorityAroundTable(t, g)
	if c := sacrificeChoiceFor(g, me.ID); c == nil || c.PickThen != discipleOfBolasPayout.Key() {
		t.Fatalf("the sacrifice prompt is %+v, want a leg of the keyed run", c)
	}

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	life, hand := me.Life, me.Hand.Size()
	answerSacrifice(t, g, me.ID, beast)
	if me.Life != life+5 || me.Hand.Size() != hand+5 {
		t.Errorf("life %d -> %d, hand %d -> %d on the restored table, want +5 of each", life, me.Life, hand, me.Hand.Size())
	}
}

// Bound: the sacrifice, and then the second choice (up to X cards from the
// graveyard, X read off the creature as it last was), are both restore
// points, each answered on a restored table.
func TestBoundBothPicksAreRestorePoints(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	green := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Two Colors", TypeLine: "Creature — Elf",
		Colors: []string{"G", "B"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	a := pushGraveyardCardForTest(me, "Gone A")
	b := pushGraveyardCardForTest(me, "Gone B")
	c := pushGraveyardCardForTest(me, "Gone C")
	castCatalogSpell(t, g, "Bound", "Instant", "df39fbee-6ddc-4733-a32e-cdaf369ecccf", nil)
	passPriorityAroundTable(t, g)
	if sc := sacrificeChoiceFor(g, me.ID); sc == nil || sc.PickThen != boundSacrificed.Key() {
		t.Fatalf("the sacrifice prompt is %+v, want a leg of the keyed run", sc)
	}

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	answerSacrifice(t, g, me.ID, green)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil || pick.PickThen != boundReturned.Key() || pick.ChooseMax != 2 {
		t.Fatalf("the second choice is %+v, want a keyed pick of up to two (the creature was two colors)", pick)
	}

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	answerChooseCards(t, g, me.ID, a, c)
	for _, id := range []uuid.UUID{a, c} {
		if !me.Hand.Contains(id) {
			t.Errorf("%s did not return to the hand on the restored table", id)
		}
	}
	if me.Hand.Contains(b) {
		t.Error("a card that was not chosen returned")
	}
}

// --- choose_cards fan-outs ---------------------------------------------

// Ill-Gotten Gains: every player picks up to three cards from their
// graveyard. One has answered, the table restores with the others open, and
// the cards return after the last answer.
func TestIllGottenGainsRunIsARestorePointWithALegAnswered(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for _, p := range g.Seats {
		p.Hand.Cards = nil
	}
	mine := s58HandCards(me, 2)
	theirs := s58HandCards(opp, 3)
	castCatalogSpell(t, g, "Ill-Gotten Gains", "Sorcery", s58IllGottenGains, nil)
	passPriorityAroundTable(t, g)
	first := pendingChooseCards(t, g, me.ID)
	if first.PickThen != game.ChooseCardsRunLeg.Key() {
		t.Fatalf("a leg is %+v, want the engine's keyed run leg", first)
	}
	answerChooseCards(t, g, me.ID, mine[0])
	if me.Hand.Contains(mine[0]) {
		t.Fatal("a card returned before the last player had answered")
	}

	g = restoreRoundTrip(t, g, true)
	me, opp = g.Seats[0], g.Seats[1]
	if c := chooseCardsChoiceFor(g, opp.ID); c == nil || len(c.ChooseCards) != 3 {
		t.Fatalf("the restored table's second leg is %+v, want the opponent's own three", c)
	}
	answerChooseCards(t, g, opp.ID, theirs[0], theirs[1])
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(mine[0]) || !opp.Hand.Contains(theirs[0]) || !opp.Hand.Contains(theirs[1]) {
		t.Error("the chosen cards did not return to their owners' hands on the restored table")
	}
	if opp.Hand.Contains(theirs[2]) {
		t.Error("a card that was not chosen returned")
	}
}

// A keyed leg keeps the closure leg's live-zone re-check (Zone, read as
// LiveZone): a card that left the graveyard is refused as an answer, on a
// restored table too.
func TestIllGottenGainsLegRefusesACardThatLeftTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for _, p := range g.Seats {
		p.Hand.Cards = nil
	}
	mine := s58HandCards(me, 2)
	castCatalogSpell(t, g, "Ill-Gotten Gains", "Sorcery", s58IllGottenGains, nil)
	passPriorityAroundTable(t, g)
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := pendingChooseCards(t, g, me.ID)
	g.WithWriteLock(func() { me.Graveyard.Remove(mine[0]) })
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{mine[0]}); !errors.Is(err, game.ErrCardNotFound) {
		t.Fatalf("a card that left the graveyard: err = %v, want ErrCardNotFound", err)
	}
}

func TestIllGottenGainsWithAnUnknownKeyIsRefused(t *testing.T) {
	g := newCatalogGame(t)
	for _, p := range g.Seats {
		p.Hand.Cards = nil
	}
	s58HandCards(g.Seats[0], 2)
	s58HandCards(g.Seats[1], 2)
	castCatalogSpell(t, g, "Ill-Gotten Gains", "Sorcery", s58IllGottenGains, nil)
	passPriorityAroundTable(t, g)
	refuseUnknownRunKey(t, g, illGottenGainsReturned.Key())
	refuseUnknownKey(t, g, game.ChooseCardsRunLeg.Key())
}

// Slaughter the Strong: each player keeps creatures with total power 4 or
// less. The set rule is registered with the leg, so it is enforced on the
// restored table, and the rest are sacrificed after the last seat.
func TestSlaughterTheStrongKeepsItsSetRuleAcrossARestore(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myBig := b12Creature(g, me.ID, "My Big", "Creature — Wurm", 5, 5)
	myTwo := b12Creature(g, me.ID, "My Two", "Creature — Bear", 2, 2)
	myOne := b12Creature(g, me.ID, "My One", "Creature — Elf", 1, 1)
	theirFour := b12Creature(g, opp.ID, "Their Four", "Creature — Beast", 4, 4)
	theirThree := b12Creature(g, opp.ID, "Their Three", "Creature — Beast", 3, 3)
	castCatalogSpell(t, g, "Slaughter the Strong", "Sorcery", b22SlaughterTheStrongOracle, nil)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, myTwo, myOne)
	if !g.Battlefield.Contains(myBig) {
		t.Fatal("a creature was sacrificed before every player had chosen")
	}

	g = restoreRoundTrip(t, g, true)
	opp = g.Seats[1]
	c := chooseCardsChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("the restored table lost the opponent's leg")
	}
	if err := g.ResolveChooseCards(c.ID, opp.ID, []uuid.UUID{theirFour, theirThree}); !errors.Is(err, game.ErrChoiceSetRejected) {
		t.Fatalf("total power 7: err = %v, want ErrChoiceSetRejected from the registered rule", err)
	}
	answerChooseCards(t, g, opp.ID, theirFour)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(myBig) || g.Battlefield.Contains(theirThree) {
		t.Error("the creatures not kept were not sacrificed")
	}
	for _, keep := range []uuid.UUID{myTwo, myOne, theirFour} {
		if !g.Battlefield.Contains(keep) {
			t.Errorf("%s was kept and should have survived", keep)
		}
	}
}

func TestSlaughterTheStrongWithAnUnknownKeyIsRefused(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Creature(g, me.ID, "My Big", "Creature — Wurm", 5, 5)
	b12Creature(g, me.ID, "My Two", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Slaughter the Strong", "Sorcery", b22SlaughterTheStrongOracle, nil)
	passPriorityAroundTable(t, g)
	refuseUnknownRunKey(t, g, slaughterKeep.then.Key())
	refuseUnknownKey(t, g, "choose-cards/slaughter-the-strong-keep-leg")
}

// Kozilek, the Broken Reality: two targets choose two cards each to
// manifest. One has answered, the table restores with the other open, and
// everything is manifested in one entry after the last, the caster drawing
// for each.
func TestKozilekManifestRunIsARestorePointWithALegAnswered(t *testing.T) {
	g := newCatalogGame(t)
	me, first, second := g.Seats[0], g.Seats[1], g.Seats[2]
	firstHand := fillHandForTest(g, first, 5)
	secondHand := fillHandForTest(g, second, 4)
	castKozilek(t, g, first.ID, second.ID)
	handBefore := me.Hand.Size()
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, first.ID, firstHand[0], firstHand[1])
	if cardInPlay(g, firstHand[0]) {
		t.Fatal("a card entered before the last target had answered")
	}

	g = restoreRoundTrip(t, g, true)
	me, first, second = g.Seats[0], g.Seats[1], g.Seats[2]
	if c := chooseCardsChoiceFor(g, second.ID); c == nil || c.ChooseMin != 2 || c.ChooseMax != 2 {
		t.Fatalf("the restored second leg is %+v, want exactly two", c)
	}
	answerChooseCards(t, g, second.ID, secondHand[2], secondHand[3])
	passPriorityAroundTable(t, g)
	assertManifestedBy(t, g, firstHand[0], first.ID)
	assertManifestedBy(t, g, firstHand[1], first.ID)
	assertManifestedBy(t, g, secondHand[2], second.ID)
	assertManifestedBy(t, g, secondHand[3], second.ID)
	if got := me.Hand.Size(); got != handBefore+4 {
		t.Errorf("caster's hand %d -> %d on the restored table, want +4 (one draw per manifested card)", handBefore, got)
	}
}

func TestKozilekManifestRunWithAnUnknownKeyIsRefused(t *testing.T) {
	g := newCatalogGame(t)
	first, second := g.Seats[1], g.Seats[2]
	fillHandForTest(g, first, 5)
	fillHandForTest(g, second, 4)
	castKozilek(t, g, first.ID, second.ID)
	passPriorityAroundTable(t, g)
	refuseUnknownRunKey(t, g, kozilekDrawPerManifest.run.Key())
}

// Explore the Vastlands: each player may reveal a land card and/or an
// instant or sorcery card from the top five. The pick's set rule (a slot
// each) is registered with the leg and enforced on the restored table, the
// taken cards are revealed and the rest bottomed, then everyone gains 3.
func TestExploreTheVastlandsTakesAreARestorePointWithAPlayerAnswered(t *testing.T) {
	g := newCatalogGame(t)
	seats := apnapSeats(g)
	tops := make([][]uuid.UUID, len(seats))
	for i, p := range seats {
		tops[i] = stackTop(p, vastFive)
	}
	life := make([]int, len(seats))
	for i, p := range seats {
		life[i] = p.Life
	}
	castExplore(t, g)
	answer := func(g *game.Game, i int, ids ...uuid.UUID) {
		t.Helper()
		c := latestChooseCards(g, apnapSeats(g)[i].ID)
		if c == nil {
			t.Fatalf("seat %d was not asked", i)
		}
		if err := g.ResolveChooseCards(c.ID, apnapSeats(g)[i].ID, ids); err != nil {
			t.Fatalf("seat %d answer: %v", i, err)
		}
	}
	answer(g, 0, tops[0][0], tops[0][1])
	if seats[0].Hand.Contains(tops[0][0]) {
		t.Fatal("a card was taken before every player had answered")
	}

	g = restoreRoundTrip(t, g, true)
	seats = apnapSeats(g)
	c := latestChooseCards(g, seats[1].ID)
	if c == nil || c.ChooseMax != 2 {
		t.Fatalf("the restored second leg is %+v, want up to two", c)
	}
	// Two cards that fit one slot are refused by the registered rule.
	if err := g.ResolveChooseCards(c.ID, seats[1].ID, []uuid.UUID{tops[1][1], tops[1][3]}); !errors.Is(err, game.ErrChoiceSetRejected) {
		t.Fatalf("an instant and a sorcery: err = %v, want ErrChoiceSetRejected", err)
	}
	answer(g, 1, tops[1][0], tops[1][3])
	answer(g, 2)
	answer(g, 3, tops[3][1])
	for i, taken := range [][]uuid.UUID{{tops[0][0], tops[0][1]}, {tops[1][0], tops[1][3]}, nil, {tops[3][1]}} {
		for _, id := range taken {
			if !seats[i].Hand.Contains(id) {
				t.Errorf("seat %d: a revealed card did not reach the hand on the restored table", i)
			}
			if !vastRevealed(g, id) {
				t.Errorf("seat %d: a card taken into hand was not revealed", i)
			}
		}
		if seats[i].Life != life[i]+3 {
			t.Errorf("seat %d life %d, want %d", i, seats[i].Life, life[i]+3)
		}
	}
}

func TestExploreTheVastlandsWithAnUnknownKeyIsRefused(t *testing.T) {
	g := newCatalogGame(t)
	for _, p := range apnapSeats(g) {
		stackTop(p, vastFive)
	}
	castExplore(t, g)
	refuseUnknownRunKey(t, g, exploreTheVastlandsTakes.run.Key())
	refuseUnknownKey(t, g, "choose-cards/explore-the-vastlands-take")
}

// --- ChoosePermanents ---------------------------------------------------

// Tragic Arrogance: one chooser walks every board. The first leg is
// answered, the table restores with the second open, and the set rule
// (a permanent of each type) is registered with the run and enforced on the
// restored table; everything not chosen is sacrificed after the last.
func TestTragicArroganceIsARestorePointWithALegAnswered(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	var myBear, myOtherBear, theirBear, theirOtherBear, theirRelic uuid.UUID
	g.WithWriteLock(func() {
		myBear = pushNamedPermanent(g, me.ID, "My Bear", "Creature — Bear")
		myOtherBear = pushNamedPermanent(g, me.ID, "My Other Bear", "Creature — Bear")
		theirBear = pushNamedPermanent(g, opp.ID, "Their Bear", "Creature — Bear")
		theirOtherBear = pushNamedPermanent(g, opp.ID, "Their Other Bear", "Creature — Bear")
		theirRelic = pushNamedPermanent(g, opp.ID, "Their Relic", "Artifact")
	})
	castCatalogSpell(t, g, "Tragic Arrogance", "Sorcery", tragicArroganceOracle, nil)
	passPriorityAroundTable(t, g)
	mine := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if mine == nil || mine.PickThen != tragicArroganceChosen.run.Key() {
		t.Fatalf("the caster's leg is %+v, want a leg of the keyed run", mine)
	}
	if err := g.ResolveOwnPermanents(mine.ID, me.ID, []uuid.UUID{myBear}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	if !g.Battlefield.Contains(myOtherBear) {
		t.Fatal("a permanent was sacrificed before the last leg was answered")
	}

	g = restoreRoundTrip(t, g, true)
	theirs := latestChoiceOfKindFor(g, game.PendingChoiceTheirPermanents, me.ID)
	if theirs == nil {
		t.Fatal("the restored table lost the second leg")
	}
	if err := g.ResolveTheirPermanents(theirs.ID, me.ID, []uuid.UUID{theirBear, theirOtherBear}); !errors.Is(err, game.ErrChoiceSetRejected) {
		t.Fatalf("two creatures: err = %v, want ErrChoiceSetRejected from the registered rule", err)
	}
	if err := g.ResolveTheirPermanents(theirs.ID, me.ID, []uuid.UUID{theirBear, theirRelic}); err != nil {
		t.Fatalf("ResolveTheirPermanents: %v", err)
	}
	for _, id := range []uuid.UUID{myBear, theirBear, theirRelic} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s was chosen and should have survived", id)
		}
	}
	if g.Battlefield.Contains(myOtherBear) || g.Battlefield.Contains(theirOtherBear) {
		t.Error("the creatures not chosen were not sacrificed on the restored table")
	}
}

func TestTragicArroganceWithAnUnknownKeyIsRefused(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() {
		pushNamedPermanent(g, me.ID, "My Bear", "Creature — Bear")
		pushNamedPermanent(g, me.ID, "My Other Bear", "Creature — Bear")
	})
	castCatalogSpell(t, g, "Tragic Arrogance", "Sorcery", tragicArroganceOracle, nil)
	passPriorityAroundTable(t, g)
	refuseUnknownKey(t, g, tragicArroganceChosen.run.Key())
	refuseUnknownRunKey(t, g, tragicArroganceChosen.run.Key())
}

// Promise of Loyalty asks each player in turn, the answer asking the next:
// every question is a restore point and the players still to ask, and the
// creatures chosen so far, travel on the run.
func TestPromiseOfLoyaltyChainSurvivesARestoreAtEveryQuestion(t *testing.T) {
	g := newCatalogGame(t)
	var keep []uuid.UUID
	var all [][]uuid.UUID
	for _, p := range g.Seats {
		a := pushVanillaCreature(g, p.ID, p.Name+" A", 2, 2)
		b := pushVanillaCreature(g, p.ID, p.Name+" B", 2, 2)
		all = append(all, []uuid.UUID{a, b})
	}
	castCatalogSpell(t, g, "Promise of Loyalty", "Sorcery", "e1d294e8-67b1-4a4a-86b8-f98f28a327d9", nil)
	passPriorityAroundTable(t, g)
	for range g.Seats {
		g = restoreRoundTrip(t, g, true)
		var c *game.PendingChoice
		for _, pc := range g.PendingChoices {
			if pc != nil && (pc.Kind == game.PendingChoiceOwnPermanents || pc.Kind == game.PendingChoiceTheirPermanents) {
				c = pc
				break
			}
		}
		if c == nil {
			t.Fatalf("no question after %d answers: %+v", len(keep), g.PendingChoices)
		}
		pick := c.ChooseCards[0]
		keep = append(keep, pick)
		var err error
		if c.Kind == game.PendingChoiceOwnPermanents {
			err = g.ResolveOwnPermanents(c.ID, c.Chooser, []uuid.UUID{pick})
		} else {
			err = g.ResolveTheirPermanents(c.ID, c.Chooser, []uuid.UUID{pick})
		}
		if err != nil {
			t.Fatalf("answer: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if len(keep) != len(g.Seats) {
		t.Fatalf("answered %d questions for %d seats", len(keep), len(g.Seats))
	}
	for i, p := range g.Seats {
		survivors := 0
		for _, id := range all[i] {
			if g.Battlefield.Contains(id) {
				survivors++
			}
		}
		if survivors != 1 {
			t.Errorf("%s kept %d creatures, want exactly the one chosen", p.Name, survivors)
		}
	}
}

// --- Plunge into Darkness -----------------------------------------------

// Entwined, both bullets: the sacrifice (a keyed ChoosePermanents) and the
// look's pick (a keyed take) are restore points, each answered on a
// restored table.
func TestPlungeIntoDarknessBothPicksAreRestorePoints(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	a := pushVanillaCreature(g, me.ID, "Fodder A", 1, 1)
	b := pushVanillaCreature(g, me.ID, "Fodder B", 1, 1)
	top := seedTypedLibrary(me, "Sorcery", "Deep", "Deeper", "Top Three", "Top Two", "Top One")
	life := me.Life
	if _, err := castEntwineSpell(g, me, "Plunge into Darkness", "Instant", "{1}{B}", ewbPlungeOracle,
		[]string{"B"}, "{C}{B}{B}", []int{0, 1}, true, nil); err != nil {
		t.Fatalf("entwined: %v", err)
	}
	passPriorityAroundTable(t, g)
	sac := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if sac == nil || sac.PickThen != plungeIntoDarknessSacrificed.run.Key() {
		t.Fatalf("the sacrifice pick is %+v, want a leg of the keyed run", sac)
	}

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	sac = latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if err := g.ResolveOwnPermanents(sac.ID, me.ID, []uuid.UUID{a, b}); err != nil {
		t.Fatalf("ResolveOwnPermanents on the restored table: %v", err)
	}
	if me.Life != life+6 {
		t.Fatalf("life %d after two sacrifices on the restored table, want %d", me.Life, life+6)
	}
	pay := payAmountPrompt(g, me.ID)
	if pay == nil {
		t.Fatal("the life bullet asks how much to pay")
	}
	if err := g.ResolvePayAmount(pay.ID, me.ID, 3); err != nil {
		t.Fatalf("pay 3: %v", err)
	}
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil || pick.PickThen != plungeIntoDarknessTake.Key() {
		t.Fatalf("the look's pick is %+v, want the keyed take", pick)
	}

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	looked := top[2:]
	answerChooseCards(t, g, me.ID, looked[1])
	g.SettleResolution()
	if !me.Hand.Contains(looked[1]) {
		t.Error("the chosen card is not in hand after the restore")
	}
	for _, id := range []uuid.UUID{looked[0], looked[2]} {
		if me.Library.Contains(id) || me.Hand.Contains(id) {
			t.Errorf("%s was not exiled", id)
		}
	}
	if !me.Library.Contains(top[0]) || !me.Library.Contains(top[1]) {
		t.Error("cards below the three stay in the library")
	}
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("both creatures were sacrificed")
	}
}

func TestPlungeIntoDarknessPicksWithAnUnknownKeyAreRefused(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	pushVanillaCreature(g, me.ID, "Fodder A", 1, 1)
	pushVanillaCreature(g, me.ID, "Fodder B", 1, 1)
	seedTypedLibrary(me, "Sorcery", "Deep", "Deeper", "Top Three", "Top Two", "Top One")
	if _, err := castEntwineSpell(g, me, "Plunge into Darkness", "Instant", "{1}{B}", ewbPlungeOracle,
		[]string{"B"}, "{C}{B}{B}", []int{0, 1}, true, nil); err != nil {
		t.Fatalf("entwined: %v", err)
	}
	passPriorityAroundTable(t, g)
	refuseUnknownKey(t, g, plungeIntoDarknessSacrificed.run.Key())
	refuseUnknownRunKey(t, g, plungeIntoDarknessSacrificed.run.Key())

	// The look, on its own table: unentwined, the second bullet alone.
	g = newCatalogGame(t)
	me, _ = cmcSeats(t, g)
	seedTypedLibrary(me, "Sorcery", "Deep", "Deeper", "Top Three", "Top Two", "Top One")
	if _, err := castEntwineSpell(g, me, "Plunge into Darkness", "Instant", "{1}{B}", ewbPlungeOracle,
		[]string{"B"}, "{C}{B}", []int{1}, false, nil); err != nil {
		t.Fatalf("second bullet: %v", err)
	}
	passPriorityAroundTable(t, g)
	pay := payAmountPrompt(g, me.ID)
	if pay == nil {
		t.Fatal("the life bullet asks how much to pay")
	}
	if err := g.ResolvePayAmount(pay.ID, me.ID, 2); err != nil {
		t.Fatalf("pay 2: %v", err)
	}
	if c := chooseCardsChoiceFor(g, me.ID); c == nil || c.PickThen != plungeIntoDarknessTake.Key() {
		t.Fatalf("the look's pick is %+v, want the keyed take", c)
	}
	refuseUnknownKey(t, g, plungeIntoDarknessTake.Key())
}

// --- the ordered collector ----------------------------------------------

// Syphon Flesh: "each other player sacrifices a creature of their choice;
// you create a 2/2 Zombie for each creature sacrificed this way". The
// opponents choose one at a time, the first has answered, the table
// restores with the next open and seeing the first pick, nothing has been
// sacrificed, and the Zombies come after the last.
func TestSyphonFleshCollectionIsARestorePointWithASeatAnswered(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bears := make([][]uuid.UUID, len(g.Seats))
	for i, p := range g.Seats[1:] {
		bears[i+1] = []uuid.UUID{
			pushVanillaCreature(g, p.ID, p.Name+" A", 2, 2),
			pushVanillaCreature(g, p.ID, p.Name+" B", 2, 2),
		}
	}
	castCatalogSpell(t, g, "Syphon Flesh", "Sorcery", "bc7c3151-ad07-429a-8af8-e4f24b7d9f6b", nil)
	passPriorityAroundTable(t, g)
	var order []*game.Player
	for _, p := range apnapSeats(g) {
		if p.ID != me.ID {
			order = append(order, p)
		}
	}
	indexOf := func(g *game.Game, id uuid.UUID) int {
		for i, p := range g.Seats {
			if p.ID == id {
				return i
			}
		}
		return -1
	}
	firstPrompt := chooseCardsChoiceFor(g, order[0].ID)
	if firstPrompt == nil {
		t.Fatalf("the first opponent was not asked: %+v", g.PendingChoices)
	}
	firstPick := bears[indexOf(g, order[0].ID)][0]
	answerChooseCards(t, g, order[0].ID, firstPick)
	if !g.Battlefield.Contains(firstPick) {
		t.Fatal("a pick left the battlefield before every opponent had chosen")
	}

	g = restoreRoundTrip(t, g, true)
	order = nil
	for _, p := range apnapSeats(g) {
		if p.ID != g.Seats[0].ID {
			order = append(order, p)
		}
	}
	second := chooseCardsChoiceFor(g, order[1].ID)
	if second == nil {
		t.Fatal("the restored table lost the second opponent's turn to choose")
	}
	if earlier := second.EarlierPermanentChoicesForEffect(); len(earlier) != 1 || earlier[0].Seat != order[0].ID || earlier[0].Cards[0] != firstPick {
		t.Errorf("the second opponent sees %+v, want the first opponent's pick", earlier)
	}
	picks := []uuid.UUID{firstPick}
	for i := 1; i < len(order); i++ {
		pick := bears[indexOf(g, order[i].ID)][1]
		picks = append(picks, pick)
		answerChooseCards(t, g, order[i].ID, pick)
	}
	passPriorityAroundTable(t, g)
	for _, id := range picks {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s was chosen and should have been sacrificed", id)
		}
	}
	zombies := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == g.Seats[0].ID && c.Name == "Zombie" {
			zombies++
		}
	}
	if zombies != len(picks) {
		t.Errorf("%d Zombies on the restored table, want %d", zombies, len(picks))
	}
}

func TestSyphonFleshCollectionWithAnUnknownKeyIsRefused(t *testing.T) {
	g := newCatalogGame(t)
	g.WithWriteLock(func() {
		for _, p := range g.Seats[1:] {
			pushNamedPermanent(g, p.ID, p.Name+" A", "Creature — Bear")
			pushNamedPermanent(g, p.ID, p.Name+" B", "Creature — Bear")
		}
	})
	castCatalogSpell(t, g, "Syphon Flesh", "Sorcery", "bc7c3151-ad07-429a-8af8-e4f24b7d9f6b", nil)
	passPriorityAroundTable(t, g)
	refuseUnknownCollectionKey(t, g, syphonFleshEdicted.run.Key())
}

// Rise of the Witch-king: each player sacrifices a creature and, if you
// did, you may return another permanent card. The caster rides the run as
// its carry.
func TestRiseOfTheWitchKingCollectionIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := pushVanillaCreature(g, me.ID, "Mine", 1, 1)
	pushVanillaCreature(g, me.ID, "Mine Too", 1, 1)
	opps := make([][]uuid.UUID, len(g.Seats))
	for i, p := range g.Seats[1:] {
		opps[i+1] = []uuid.UUID{
			pushVanillaCreature(g, p.ID, p.Name+" A", 1, 1),
			pushVanillaCreature(g, p.ID, p.Name+" B", 1, 1),
		}
	}
	gone := pushGraveyardCardForTest(me, "Gone Land")
	g.WithWriteLock(func() {
		for _, c := range me.Graveyard.Cards {
			if c.InstanceID == gone {
				c.TypeLine = "Land"
			}
		}
	})
	castCatalogSpell(t, g, "Rise of the Witch-king", "Sorcery", "3c86541c-3601-4a38-8872-39705e41303a", nil)
	passPriorityAroundTable(t, g)
	// Answer the first seat that is asked, then restore with the rest open.
	answered := 0
	for _, p := range apnapSeats(g) {
		if c := chooseCardsChoiceFor(g, p.ID); c != nil {
			pick := c.ChooseCards[0]
			if err := g.ResolveChooseCards(c.ID, p.ID, []uuid.UUID{pick}); err != nil {
				t.Fatalf("answer: %v", err)
			}
			answered++
			break
		}
	}
	if answered != 1 {
		t.Fatalf("no seat was asked: %+v", g.PendingChoices)
	}
	g = restoreRoundTrip(t, g, true)
	for range g.Seats {
		var c *game.PendingChoice
		for _, pc := range g.PendingChoices {
			if pc != nil && pc.Kind == game.PendingChoiceChooseCards && pc.PickThen != "" && len(pc.ChooseCards) > 0 {
				c = pc
				break
			}
		}
		if c == nil {
			break
		}
		if err := g.ResolveChooseCards(c.ID, c.Chooser, []uuid.UUID{c.ChooseCards[0]}); err != nil {
			t.Fatalf("answer on the restored table: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) && g.Seats[0].Graveyard.Contains(mine) {
		t.Error("a creature is in two zones")
	}
	count := 0
	for i, p := range g.Seats {
		survivors := 0
		for _, c := range g.Battlefield.Cards {
			if c.Controller == p.ID && c.IsCreature() {
				survivors++
			}
		}
		if i > 0 && survivors != 1 {
			t.Errorf("%s has %d creatures after the edict, want one of two sacrificed", p.Name, survivors)
		}
		count += survivors
	}
}
