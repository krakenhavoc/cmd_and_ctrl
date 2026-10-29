package effects

import (
	"strconv"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// confirm_closures_keyed_restore_test.go — #4298: the closure-form confirms
// that were not "you may" callers of MayChoice (the opening-hand offer, the
// explore question, The Creation of Avacyn's third chapter, and the cards
// that queued QueueConfirmKeyedForEffect themselves) are keyed, so a table waiting
// on one is a restore point. Each test takes a STRICT restore while the
// prompt is open (restoreAtConfirm fails on a table that is not a restore
// point), checks the restored prompt names its key and the data it carries,
// then answers it on the restored table. A loop (Tainted Pact's repeat,
// Sylvan Library's per-card payments, Tempt with Discovery's opponents) is
// restored at each of its prompts in turn.

// --- the engine's confirms ------------------------------------------------

// The opening-hand offer (CR 103.6, ADR 0133) is restored with the prompt
// still private to its chooser, and the card begins the game on the
// battlefield on the restored table.
func TestKeyedConfirmOpeningHandOfferIsARestorePoint(t *testing.T) {
	for _, accept := range []bool{true, false} {
		g := openingTable(t)
		ids := handWith(g, 2, leylineOfSanctity(), fillerCard("Spare"))
		for _, s := range []int{0, 1, 3} {
			handWith(g, s, fillerCard("Nothing special"))
		}
		keepAll(t, g)
		seat := g.Seats[2].ID
		g, ask := restoreAtConfirm(t, g, seat, "confirm/opening-hand-offer")
		if ask.Source != ids[0] {
			t.Errorf("the restored offer's source is %s, want the Leyline %s", ask.Source, ids[0])
		}
		if !ask.PrivateText() {
			t.Error("the restored offer is no longer private to its chooser: its words name a card in a hand")
		}
		if err := g.ResolveConfirm(ask.ID, seat, accept); err != nil {
			t.Fatalf("accept=%v: %v", accept, err)
		}
		if got := beganOnBattlefield(g, ids[0]); got != accept {
			t.Errorf("accept=%v: the Leyline began on the battlefield on the restored table: %v", accept, got)
		}
		if !accept && !seatOf(g, seat).Hand.Contains(ids[0]) {
			t.Error("declined: the Leyline left the hand")
		}
	}
}

func TestKeyedConfirmOpeningHandOfferWithAnUnknownKeyIsRefused(t *testing.T) {
	g := openingTable(t)
	handWith(g, 2, leylineOfSanctity(), fillerCard("Spare"))
	for _, s := range []int{0, 1, 3} {
		handWith(g, s, fillerCard("Nothing special"))
	}
	keepAll(t, g)
	refuseUnknownKey(t, g, "confirm/opening-hand-offer")
}

// Explore (CR 701.44a): the revealed nonland and the explorer ride the
// prompt; a yes puts the card into the graveyard and the explorer has
// explored (CR 701.44b), a no leaves it on top.
func TestKeyedConfirmExploreGraveyardIsARestorePoint(t *testing.T) {
	for _, accept := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		explorer := ctrlPushCreature(g, me.ID, "Explorer")
		top := seedLibraryTop(me, "Some Sorcery", "Sorcery", "{1}")
		epoch := 0
		if c, ok := g.LookupCardForEffect(explorer); ok {
			epoch = c.ObjectEpoch
		}
		g.WithWriteLock(func() {
			if err := g.ExploreForEffect(uuid.Nil, game.ObjectRef{ID: explorer, Epoch: epoch}, me.ID); err != nil {
				t.Fatalf("ExploreForEffect: %v", err)
			}
		})
		g, ask := restoreAtConfirm(t, g, me.ID, "confirm/explore-graveyard")
		if len(ask.OptionCarry) != 2 || ask.OptionCarry[0] != top || ask.OptionCarry[1] != explorer {
			t.Errorf("the restored prompt carries %v, want the revealed card and the explorer", ask.OptionCarry)
		}
		if n := len(randomEvents(g, game.EventExplored)); n != 0 {
			t.Fatalf("explored %d times before the question was answered", n)
		}
		me = seatOf(g, me.ID)
		if err := g.ResolveConfirm(ask.ID, me.ID, accept); err != nil {
			t.Fatal(err)
		}
		if accept != me.Graveyard.Contains(top) {
			t.Errorf("accept=%v: the revealed card is in the graveyard on the restored table: %v", accept, me.Graveyard.Contains(top))
		}
		if !accept && me.Library.Cards[me.Library.Size()-1].InstanceID != top {
			t.Error("declined: the card is not still on top")
		}
		evs := randomEvents(g, game.EventExplored)
		if len(evs) != 1 || evs[0].CardID != explorer || evs[0].Actor != me.ID {
			t.Errorf("explored events %+v, want one for the carried explorer", evs)
		}
		if c, _ := g.LookupCardForEffect(explorer); c.Counters[game.CounterPlusOne] != 1 {
			t.Errorf("the explorer has %d +1/+1 counters, want 1", c.Counters[game.CounterPlusOne])
		}
	}
}

func TestKeyedConfirmExploreGraveyardWithAnUnknownKeyIsRefused(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	explorer := ctrlPushCreature(g, me.ID, "Explorer")
	seedLibraryTop(me, "Some Sorcery", "Sorcery", "{1}")
	g.WithWriteLock(func() {
		_ = g.ExploreForEffect(uuid.Nil, game.ObjectRef{ID: explorer}, me.ID)
	})
	refuseUnknownKey(t, g, "confirm/explore-graveyard")
}

// The Creation of Avacyn's third chapter: the exiled creature rides the
// prompt, a yes puts it onto the battlefield and a no into its owner's hand.
func TestKeyedConfirmExiledCreatureChapterIsARestorePoint(t *testing.T) {
	for _, accept := range []bool{true, false} {
		g, me, _, target := avacynTable(t, game.Card{TypeLine: "Creature — Elf", ManaCost: "{2}{G}", Power: 2, Toughness: 2})
		answerSearchByID(t, g, me.ID, target)
		avacynNextChapter(t, g, me)
		avacynNextChapter(t, g, me)
		g, ask := restoreAtConfirm(t, g, me.ID, "confirm/exiled-creature-battlefield-or-hand")
		if len(ask.OptionCarry) != 1 || ask.OptionCarry[0] != target {
			t.Errorf("the restored prompt carries %v, want the exiled card %s", ask.OptionCarry, target)
		}
		me = seatOf(g, me.ID)
		if err := g.ResolveConfirm(ask.ID, me.ID, accept); err != nil {
			t.Fatal(err)
		}
		passPriorityAroundTable(t, g)
		onBattlefield := false
		for _, bc := range g.Battlefield.Cards {
			if bc.Name == "Wanted" && bc.Controller == me.ID {
				onBattlefield = true
			}
		}
		if accept != onBattlefield {
			t.Errorf("accept=%v: the creature is on the battlefield on the restored table: %v", accept, onBattlefield)
		}
		if !accept && !me.Hand.Contains(target) {
			t.Error("declined: the card did not go to its owner's hand")
		}
	}
}

func TestKeyedConfirmExiledCreatureChapterWithAnUnknownKeyIsRefused(t *testing.T) {
	g, me, _, target := avacynTable(t, game.Card{TypeLine: "Creature — Elf", ManaCost: "{2}{G}", Power: 2, Toughness: 2})
	answerSearchByID(t, g, me.ID, target)
	avacynNextChapter(t, g, me)
	avacynNextChapter(t, g, me)
	refuseUnknownKey(t, g, "confirm/exiled-creature-battlefield-or-hand")
}

// --- the cards ------------------------------------------------------------

// Ponder: the shuffle question, with the draw on both branches. A no keeps
// the arrangement, so the card put on top is the card drawn.
func TestKeyedConfirmPonderShuffleIsARestorePoint(t *testing.T) {
	for _, shuffle := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		seedLibrary(me, "Worst", "Middling", "Best", "L4", "L5", "L6", "L7", "L8")
		castCatalogSpell(t, g, "Ponder", "Sorcery", ponderOracle, nil)
		passPriorityAroundTable(t, g)
		c := lookAtTopChoiceFor(g, me.ID)
		if c == nil {
			t.Fatal("Ponder queued no look-at prompt")
		}
		if err := g.ResolveLookAtTop(c.ID, me.ID, []uuid.UUID{c.ScryCards[2], c.ScryCards[0], c.ScryCards[1]}); err != nil {
			t.Fatal(err)
		}
		best := c.ScryCards[2]
		g, ask := restoreAtConfirm(t, g, me.ID, ponderShuffle.Key())
		me = seatOf(g, me.ID)
		hand := me.Hand.Size()
		if err := g.ResolveConfirm(ask.ID, me.ID, shuffle); err != nil {
			t.Fatal(err)
		}
		if me.Hand.Size() != hand+1 {
			t.Fatalf("shuffle=%v: hand %d → %d on the restored table, want the draw on both branches", shuffle, hand, me.Hand.Size())
		}
		if drawn, _ := me.Hand.Top(); !shuffle && drawn.InstanceID != best {
			t.Errorf("kept the order but drew %q, want the card put on top", drawn.Name)
		}
	}
}

func TestKeyedConfirmPonderShuffleWithAnUnknownKeyIsRefused(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B", "C", "D")
	castCatalogSpell(t, g, "Ponder", "Sorcery", ponderOracle, nil)
	passPriorityAroundTable(t, g)
	c := lookAtTopChoiceFor(g, me.ID)
	if err := g.ResolveLookAtTop(c.ID, me.ID, c.ScryCards); err != nil {
		t.Fatal(err)
	}
	refuseUnknownKey(t, g, ponderShuffle.Key())
}

// Portent: the caster decides whether the library's owner shuffles; the
// owner is the prompt's carried ID, and a shuffle un-knows the library.
func TestKeyedConfirmPortentShuffleIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	lib := seedLibrary(opp, "A", "B", "C", "D")
	castCatalogSpell(t, g, "Portent", "Sorcery", portentOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	ask := putInLibraryChoiceFor(g, me.ID)
	if ask == nil {
		t.Fatal("no look prompt")
	}
	if err := g.ResolvePutInLibrary(ask.ID, me.ID, nil, lib[:3]); err != nil {
		t.Fatal(err)
	}
	g, q := restoreAtConfirm(t, g, me.ID, portentShuffle.Key())
	if len(q.OptionCarry) != 2 || q.OptionCarry[1] != opp.ID {
		t.Errorf("the restored prompt carries %v, want the caster and the library's owner", q.OptionCarry)
	}
	if err := g.ResolveConfirm(q.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	for _, id := range lib[:3] {
		if knownTo(g, id, me.ID) {
			t.Fatal("the library's owner shuffled, but the caster still knows a card")
		}
	}
}

// Sylvan Library: the chosen cards still to settle ride the prompt, so the
// loop survives a restore at each payment. Paying the first and putting the
// second back on top, each on a restored table.
func TestKeyedConfirmSylvanLibraryPaymentsAreRestorePoints(t *testing.T) {
	g := newCatalogGame(t)
	owner, pick := sylvanOpening(t, g)
	chosen := []uuid.UUID{pick.ChooseCards[0], pick.ChooseCards[1]}
	life := owner.Life
	if err := g.ResolveChooseCards(pick.ID, owner.ID, chosen); err != nil {
		t.Fatal(err)
	}
	g, first := restoreAtConfirm(t, g, owner.ID, sylvanLibraryAnswer.Key())
	if first.LifeCost != 4 || len(first.OptionCarry) != 3 || first.OptionCarry[1] != chosen[0] || first.OptionCarry[2] != chosen[1] {
		t.Fatalf("the restored first prompt declares life %d and carries %v, want 4 and the controller and both cards", first.LifeCost, first.OptionCarry)
	}
	owner = seatOf(g, owner.ID)
	if err := g.ResolveConfirm(first.ID, owner.ID, true); err != nil {
		t.Fatal(err)
	}
	if owner.Life != life-4 || !owner.Hand.Contains(chosen[0]) {
		t.Fatalf("paid: life %d → %d, kept in hand %v", life, owner.Life, owner.Hand.Contains(chosen[0]))
	}
	g, second := restoreAtConfirm(t, g, owner.ID, sylvanLibraryAnswer.Key())
	if len(second.OptionCarry) != 2 || second.OptionCarry[1] != chosen[1] {
		t.Fatalf("the restored second prompt carries %v, want the controller and the second card", second.OptionCarry)
	}
	owner = seatOf(g, owner.ID)
	if err := g.ResolveConfirm(second.ID, owner.ID, false); err != nil {
		t.Fatal(err)
	}
	if top, err := owner.Library.Top(); err != nil || top.InstanceID != chosen[1] {
		t.Error("the declined card is not on top of the library on the restored table")
	}
	if owner.Life != life-4 || owner.Hand.Contains(chosen[1]) || len(g.PendingChoices) != 0 {
		t.Errorf("life %d, second card in hand %v, %d prompts open; want the chain finished", owner.Life, owner.Hand.Contains(chosen[1]), len(g.PendingChoices))
	}
}

func TestKeyedConfirmSylvanLibraryWithAnUnknownKeyIsRefused(t *testing.T) {
	g := newCatalogGame(t)
	owner, pick := sylvanOpening(t, g)
	if err := g.ResolveChooseCards(pick.ID, owner.ID, pick.ChooseCards[:2]); err != nil {
		t.Fatal(err)
	}
	refuseUnknownKey(t, g, sylvanLibraryAnswer.Key())
}

// Tainted Pact: the names exiled so far ride the prompt, so the repeat is
// data. Declining twice with a restore between, then taking the third.
func TestKeyedConfirmTaintedPactRepeatIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gamma := sweepFTop(me, "Gamma", "Sorcery")
	beta := sweepFTop(me, "Beta", "Sorcery")
	alpha := sweepFTop(me, "Alpha", "Sorcery")
	castCatalogSpell(t, g, "Tainted Pact", "Instant", sweepFTaintedPactOracle, nil)
	passPriorityAroundTable(t, g)

	g, ask := restoreAtConfirm(t, g, me.ID, taintedPactAnswer.Key())
	if ask.PickMatchArg != "Alpha" || len(ask.OptionCarry) != 2 || ask.OptionCarry[1] != alpha {
		t.Fatalf("the first prompt holds word %q and carry %v, want Alpha and its card", ask.PickMatchArg, ask.OptionCarry)
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, false); err != nil {
		t.Fatal(err)
	}
	g, ask = restoreAtConfirm(t, g, me.ID, taintedPactAnswer.Key())
	if ask.PickMatchArg != "Alpha\nBeta" || ask.OptionCarry[1] != beta {
		t.Fatalf("the second prompt holds word %q and carry %v, want both names so far", ask.PickMatchArg, ask.OptionCarry)
	}
	me = seatOf(g, me.ID)
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	if !inExile(g, alpha) || !inHandOf(me, beta) || !me.Library.Contains(gamma) {
		t.Errorf("alpha exiled=%v beta in hand=%v gamma in library=%v", inExile(g, alpha), inHandOf(me, beta), me.Library.Contains(gamma))
	}
}

// The repeat ends on a second card with the same name even when the names
// seen so far came through a restore.
func TestKeyedConfirmTaintedPactStopsOnADuplicateAfterARestore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	alpha2 := sweepFTop(me, "Alpha", "Sorcery")
	beta := sweepFTop(me, "Beta", "Sorcery")
	alpha1 := sweepFTop(me, "Alpha", "Sorcery")
	under := me.Library.Cards[len(me.Library.Cards)-4].InstanceID
	castCatalogSpell(t, g, "Tainted Pact", "Instant", sweepFTaintedPactOracle, nil)
	passPriorityAroundTable(t, g)
	if err := g.ResolveConfirm(latestConfirmFor(g, me.ID).ID, me.ID, false); err != nil {
		t.Fatal(err)
	}
	g, ask := restoreAtConfirm(t, g, me.ID, taintedPactAnswer.Key())
	me = seatOf(g, me.ID)
	if err := g.ResolveConfirm(ask.ID, me.ID, false); err != nil {
		t.Fatal(err)
	}
	if latestConfirmFor(g, me.ID) != nil {
		t.Fatal("asked about a second Alpha: the process ends there")
	}
	for _, id := range []uuid.UUID{alpha1, beta, alpha2} {
		if !inExile(g, id) {
			t.Errorf("card %s not exiled on the restored table", id)
		}
	}
	if !me.Library.Contains(under) {
		t.Error("exiled past the duplicate")
	}
}

func TestKeyedConfirmTaintedPactWithAnUnknownKeyIsRefused(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	sweepFTop(me, "Alpha", "Sorcery")
	castCatalogSpell(t, g, "Tainted Pact", "Instant", sweepFTaintedPactOracle, nil)
	passPriorityAroundTable(t, g)
	refuseUnknownKey(t, g, taintedPactAnswer.Key())
}

// Lurking Predators: "otherwise, you may put that card on the bottom".
func TestKeyedConfirmLurkingPredatorsBottomIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Lurking Predators", "Enchantment", plLurkingOracle, 0, 0)
	rite := plTop(me, "Rite", "Sorcery", "")
	b10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
	for i := 0; i < 8 && confirmChoiceFor(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	g, ask := restoreAtConfirm(t, g, me.ID, lurkingPredatorsBottom.Key())
	if len(ask.OptionCarry) != 2 || ask.OptionCarry[1] != rite {
		t.Errorf("the restored prompt carries %v, want the revealed card", ask.OptionCarry)
	}
	me = seatOf(g, me.ID)
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	if me.Library.Cards[0].InstanceID != rite {
		t.Error("accepted: the card is not on the bottom on the restored table")
	}
}

// Obeka's "the player whose turn it is may end the turn": the question is
// asked of the active player, and a yes on the restored table ends the turn.
func TestKeyedConfirmEndTheTurnIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	mine := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	obeka := pushCatalogPermanent(g, mine.ID, "Obeka, Brute Chronologist", "Legendary Creature — Ogre Wizard", obekaOracle, false)
	advanceToStepOf(t, g, mine.Seat, game.StepPrecombatMain)
	turn := g.Turn.Seq
	if err := g.ActivateCatalogAbility(mine.ID, obeka, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	g, ask := restoreAtConfirm(t, g, mine.ID, activePlayerEndTheTurn.Key())
	if err := g.ResolveConfirm(ask.ID, mine.ID, true); err != nil {
		t.Fatal(err)
	}
	discardIfOwed(t, g)
	if g.Turn.Seq != turn+1 || g.Turn.ActiveSeat == mine.Seat {
		t.Errorf("yes did not end the turn on the restored table: turn %d seat %d step %s", g.Turn.Seq, g.Turn.ActiveSeat, g.Turn.Step)
	}
}

// Sephiroth's cell counter: the target rides the prompt, and the base-P/T
// sentence happens either way.
func TestKeyedConfirmSephirothCellCounterIsARestorePoint(t *testing.T) {
	for _, accept := range []bool{true, false} {
		g := newCatalogGame(t)
		seat := g.Turn.ActiveSeat
		me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
		advanceToMain(t, g)
		seph := pushCatalogPermanent(g, me.ID, "Sephiroth, Fallen Hero", "Legendary Creature — Human Avatar Soldier", sephirothFallenOracle, false)
		bear := seedBear(g, me.ID)
		declareAttack(t, g, opp.ID, seph)
		if pick := latestPickTarget(g, me.ID); pick != nil {
			pickCard(t, g, me.ID, bear)
		}
		passPriorityAroundTable(t, g)
		g, ask := restoreAtConfirm(t, g, me.ID, sephirothCounter.Key())
		if len(ask.OptionCarry) != 2 || ask.OptionCarry[1] != bear {
			t.Fatalf("the restored prompt carries %v, want the target", ask.OptionCarry)
		}
		if err := g.ResolveConfirm(ask.ID, me.ID, accept); err != nil {
			t.Fatal(err)
		}
		passPriorityAroundTable(t, g)
		want := 2
		if accept {
			want = 7
		}
		if got := effectivePower(t, g, bear); got != want {
			t.Errorf("accept=%v: Bear power %d on the restored table, want %d", accept, got, want)
		}
	}
}

// Lazav, Familiar Stranger: the exiled card rides the prompt and its printed
// values are read from it when the answer arrives.
func TestKeyedConfirmLazavCopyIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lazav := dcSeed(g, me.ID, "Lazav, Familiar Stranger", "Legendary Creature — Shapeshifter", oracleLazavFamiliarStranger, 1, 4)
	giant := dcGraveyardCard(g, opp, "Hill Giant", "Creature — Giant", "oracle-giant", 3, 3)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetPlayer(opp.ID))
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("Lazav must offer the graveyard exile")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{giant}); err != nil {
		t.Fatal(err)
	}
	g, ask := restoreAtConfirm(t, g, me.ID, lazavFamiliarStrangerCopy.Key())
	if len(ask.OptionCarry) != 2 || ask.OptionCarry[1] != giant {
		t.Fatalf("the restored prompt carries %v, want the exiled card", ask.OptionCarry)
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	if c := dcCard(t, g, lazav); c.Name != "Hill Giant" || c.Effective().Power != 3 {
		t.Errorf("Lazav did not become the Hill Giant on the restored table: %q %d/%d", c.Name, c.Effective().Power, c.Effective().Toughness)
	}
}

// Reckless Endeavor: the two rolls ride the prompt as its word, and each
// answer assigns them the other way round.
func TestKeyedConfirmRecklessEndeavorResultIsARestorePoint(t *testing.T) {
	for _, accept := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		creature := b12Push(g, g.Seats[1].ID, "survivor", "Creature", "", 1, 30)
		castCatalogSpell(t, g, "Reckless Endeavor", "Sorcery", "928bf0d9-89be-462d-aeb7-74a65e95535c", nil)
		passPriorityAroundTable(t, g)
		rolls := randomEvents(g, game.EventRollDie)
		if len(rolls) != 2 {
			t.Fatalf("rolls: %+v", rolls)
		}
		if rolls[0].Amount == rolls[1].Amount {
			t.Skip("the two dice came up equal: there is no result to choose")
		}
		damage, tokens := rolls[0].Amount, rolls[1].Amount
		g, ask := restoreAtConfirm(t, g, me.ID, recklessEndeavorChoice.Key())
		if want := strconv.Itoa(damage) + " " + strconv.Itoa(tokens); ask.PickMatchArg != want {
			t.Errorf("the restored prompt's word is %q, want %q", ask.PickMatchArg, want)
		}
		if err := g.ResolveConfirm(ask.ID, me.ID, accept); err != nil {
			t.Fatal(err)
		}
		if !accept {
			damage, tokens = tokens, damage
		}
		c, _ := g.LookupCardForEffect(creature)
		if c.DamageMarked != damage || randomTokenCount(g, "Treasure") != tokens {
			t.Errorf("accept=%v: damage=%d tokens=%d on the restored table, want %d/%d", accept, c.DamageMarked, randomTokenCount(g, "Treasure"), damage, tokens)
		}
	}
}

// Tempt with Discovery: the opponents still to ask ride the prompt, so the
// offer is restored at each opponent in turn and the caster is still paid
// for each search.
func TestKeyedConfirmTemptWithDiscoveryOffersAreRestorePoints(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1, opp2, opp3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	for _, p := range []*game.Player{me, opp1, opp2, opp3} {
		for i := 0; i < 2; i++ {
			b07LibraryLand(p, "Forest", "Basic Land — Forest")
		}
	}
	b07LibraryLand(me, "Cabal Coffers", "Land")
	b07LibraryLand(me, "Gaea's Cradle", "Legendary Land")
	castCatalogSpell(t, g, "Tempt with Discovery", "Sorcery", b07TemptWithDiscoveryOracle, nil)
	passPriorityAroundTable(t, g)
	search := func(who uuid.UUID) {
		t.Helper()
		c := searchChoiceFor(g, who)
		if c == nil {
			t.Fatalf("no search prompt for %s", who)
		}
		answerSearchByID(t, g, who, c.SearchCards[0])
	}
	search(me.ID)

	g, ask := restoreAtConfirm(t, g, opp1.ID, b07TemptWithDiscoveryAnswer.Key())
	if len(ask.OptionCarry) != 3 || ask.OptionCarry[1] != opp2.ID || ask.OptionCarry[2] != opp3.ID {
		t.Fatalf("the first offer carries %v, want the caster and the two opponents still to ask", ask.OptionCarry)
	}
	if err := g.ResolveConfirm(ask.ID, opp1.ID, true); err != nil {
		t.Fatal(err)
	}
	search(opp1.ID)
	search(me.ID)

	g, ask = restoreAtConfirm(t, g, opp2.ID, b07TemptWithDiscoveryAnswer.Key())
	if len(ask.OptionCarry) != 2 || ask.OptionCarry[1] != opp3.ID {
		t.Fatalf("the second offer carries %v, want the caster and the last opponent", ask.OptionCarry)
	}
	if err := g.ResolveConfirm(ask.ID, opp2.ID, false); err != nil {
		t.Fatal(err)
	}
	ask = latestConfirmFor(g, opp3.ID)
	if ask == nil || len(ask.OptionCarry) != 1 {
		t.Fatalf("declining did not ask the last opponent: %+v", ask)
	}
	if err := g.ResolveConfirm(ask.ID, opp3.ID, true); err != nil {
		t.Fatal(err)
	}
	search(opp3.ID)
	search(me.ID)
	lands := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsLand() {
			lands++
		}
	}
	if lands != 3 {
		t.Errorf("the caster has %d lands: the opening search and one for each opponent who searched, want 3", lands)
	}
}

func TestKeyedConfirmTemptWithDiscoveryWithAnUnknownKeyIsRefused(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for _, p := range g.Seats {
		b07LibraryLand(p, "Forest", "Basic Land — Forest")
	}
	castCatalogSpell(t, g, "Tempt with Discovery", "Sorcery", b07TemptWithDiscoveryOracle, nil)
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, searchChoiceFor(g, me.ID).SearchCards[0])
	refuseUnknownKey(t, g, b07TemptWithDiscoveryAnswer.Key())
}

// Channel Harm: "you may have it deal that much damage to target creature",
// asked as the damage is prevented. The target rides the prompt and the
// damage is its word.
func TestKeyedConfirmChannelHarmDealIsARestorePoint(t *testing.T) {
	for _, accept := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 6)
		castCatalogSpell(t, g, "Channel Harm", "Instant", channelHarmOracle,
			[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
		passPriorityAroundTable(t, g)
		src := pr7Creature(g, opp.ID, "Src", 3, "R")
		g.WithWriteLock(func() {
			if err := g.DealDamageToPlayerForEffect(src, me.ID, 3); err != nil {
				t.Fatal(err)
			}
		})
		g.RunStateChecksForTest()
		g, ask := restoreAtConfirm(t, g, me.ID, channelHarmDealAnswer.Key())
		if len(ask.OptionCarry) != 2 || ask.OptionCarry[1] != bear || ask.PickMatchArg != "3" {
			t.Fatalf("the restored prompt carries %v with word %q, want the target and 3", ask.OptionCarry, ask.PickMatchArg)
		}
		if err := g.ResolveConfirm(ask.ID, me.ID, accept); err != nil {
			t.Fatal(err)
		}
		want := 0
		if accept {
			want = 3
		}
		if got := damageMarkedOn(g, bear); got != want {
			t.Errorf("accept=%v: the bear has %d damage on the restored table, want %d", accept, got, want)
		}
	}
}

// testDeclineEndsLoopAnswer is a keyed answer for the test below. A "test/"
// key is outside the ledger.
var testDeclineEndsLoopAnswer = MayYes("test/decline-ends-loop", func(*Context, game.ConfirmAnswer) error { return nil })

// MayChoice.DeclineEndsLoop (Channel Harm's hint to a bot, #2040) reaches
// the keyed prompt and survives a restore.
func TestKeyedConfirmMayChoiceKeepsItsDeclineEndsLoopHint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	item := &game.StackItem{Kind: game.StackItemTriggered, Controller: me.ID, Owner: me.ID, SourceCardID: uuid.New()}
	g.WithWriteLock(func() {
		err := MayChoice{Question: "Loop?", Then: testDeclineEndsLoopAnswer, DeclineEndsLoop: true}.Apply(NewContext(g, item))
		if err != nil {
			t.Fatal(err)
		}
	})
	ask := latestConfirmFor(g, me.ID)
	if ask == nil || !ask.DeclineEndsLoop {
		t.Fatalf("the prompt does not carry the hint: %+v", ask)
	}
	g, ask = restoreAtConfirm(t, g, me.ID, testDeclineEndsLoopAnswer.Key())
	if !ask.DeclineEndsLoop {
		t.Error("the restored prompt lost the hint")
	}
}
