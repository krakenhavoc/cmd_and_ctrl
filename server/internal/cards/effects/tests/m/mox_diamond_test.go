package testsm

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func etbEventsFor(g *game.Game, id uuid.UUID) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventETB && ev.CardID == id {
			n++
		}
	}
	return n
}

func TestMoxDiamondAsksBeforeItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	forest := effectstest.HandCard(me, "Forest", "Basic Land — Forest")
	effectstest.HandCard(me, "Lightning Bolt", "Instant")

	mox := effectstest.CastMox(t, g)

	c := effectstest.EntryChoiceOfKind(g, game.PendingChoiceEntryDiscardFromHand, me.ID)
	if c == nil {
		t.Fatal("Mox Diamond resolved without asking for a land card")
	}
	if c.ChooseMin != 0 || c.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want 0..1 — it is a \"may\"", c.ChooseMin, c.ChooseMax)
	}
	if len(c.ChooseCards) != 1 || c.ChooseCards[0] != forest {
		t.Errorf("candidates %v, want just the Forest (only a land card)", c.ChooseCards)
	}
	if _, ok := effectstest.BattlefieldCard(g, mox); ok {
		t.Error("Mox Diamond is on the battlefield before the question was answered (CR 614.12a)")
	}
	if err := g.PassPriority(); !errors.Is(err, game.ErrChoicePending) {
		t.Errorf("PassPriority with the prompt open: %v, want ErrChoicePending", err)
	}
}

func TestMoxDiamondDiscardingALandPutsItOntoTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	forest := effectstest.HandCard(me, "Forest", "Basic Land — Forest")

	mox := effectstest.CastMox(t, g)
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntryDiscardFromHand, me.ID, forest)

	card, ok := effectstest.BattlefieldCard(g, mox)
	if !ok {
		t.Fatal("Mox Diamond is not on the battlefield after the discard")
	}
	if card.Tapped {
		t.Error("Mox Diamond entered tapped")
	}
	if !me.Graveyard.Contains(forest) {
		t.Error("the discarded land is not in the graveyard")
	}
	if g.Stack.Contains(mox) {
		t.Error("Mox Diamond was left on the stack")
	}
	// Owner decision 4: an effect's discard, credited to the Mox.
	var discard *game.Event
	for i := range g.Events {
		if g.Events[i].Kind == game.EventDiscardCard && g.Events[i].CardID == forest {
			discard = &g.Events[i]
		}
	}
	if discard == nil {
		t.Fatal("no EventDiscardCard for the land")
	}
	if discard.DiscardCause != game.DiscardCauseEffect {
		t.Errorf("discard cause %q, want %q", discard.DiscardCause, game.DiscardCauseEffect)
	}
	if discard.Source != mox {
		t.Errorf("discard source %s, want Mox Diamond %s", discard.Source, mox)
	}
	if n := etbEventsFor(g, mox); n != 1 {
		t.Errorf("%d ETB events for the Mox, want 1", n)
	}
}

func TestMoxDiamondDeclinedGoesToTheGraveyardAndNeverEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	forest := effectstest.HandCard(me, "Forest", "Basic Land — Forest")

	mox := effectstest.CastMox(t, g)
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntryDiscardFromHand, me.ID)

	if _, ok := effectstest.BattlefieldCard(g, mox); ok {
		t.Fatal("a declined Mox Diamond entered the battlefield")
	}
	if !me.Graveyard.Contains(mox) {
		t.Fatal("a declined Mox Diamond is not in its owner's graveyard")
	}
	if g.Stack.Contains(mox) {
		t.Error("a declined Mox Diamond was stranded on the stack")
	}
	if !me.Hand.Contains(forest) {
		t.Error("the land left the hand though nothing was discarded")
	}
	// The 2008-05-01 ruling: it never enters, so nothing that watches
	// an entry sees it.
	if n := etbEventsFor(g, mox); n != 0 {
		t.Errorf("%d ETB events for a Mox that never entered", n)
	}
	if err := g.PassPriority(); errors.Is(err, game.ErrChoicePending) {
		t.Error("a prompt is still open after the decline")
	}
}

func TestMoxDiamondWithNoLandIsNotAskedAndGoesToTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	effectstest.HandCard(me, "Lightning Bolt", "Instant")

	mox := effectstest.CastMox(t, g)

	if c := effectstest.EntryChoiceOfKind(g, game.PendingChoiceEntryDiscardFromHand, me.ID); c != nil {
		t.Error("asked for a land card with none in hand")
	}
	if !me.Graveyard.Contains(mox) {
		t.Fatal("Mox Diamond with no land to discard is not in the graveyard")
	}
	if _, ok := effectstest.BattlefieldCard(g, mox); ok {
		t.Error("Mox Diamond entered for free")
	}
}

// TestMoxDiamondDiscardPausesOnLibraryOfLeng is owner decisions 3 and
// 4 together: the discard is an effect's, so Library of Leng offers its
// "may"; that pauses the discard, and the entry rides across the pause
// and finishes when Leng is answered.
func TestMoxDiamondDiscardPausesOnLibraryOfLeng(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushTokenReplacementCard(g, effectstest.LibraryOfLengOracle, "Library of Leng", "Artifact", me.ID)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	forest := effectstest.HandCard(me, "Forest", "Basic Land — Forest")

	mox := effectstest.CastMox(t, g)
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntryDiscardFromHand, me.ID, forest)

	// Paused on Leng: nothing has entered yet.
	if _, ok := effectstest.BattlefieldCard(g, mox); ok {
		t.Fatal("the Mox entered before Library of Leng's question was answered")
	}
	effectstest.AnswerLengPrompt(t, g, me.ID, true)

	if _, ok := effectstest.BattlefieldCard(g, mox); !ok {
		t.Fatal("the Mox did not enter once Leng was answered")
	}
	if !me.Library.Contains(forest) {
		t.Fatal("Library of Leng did not put the land into the library")
	}
	if top := me.Library.Cards[len(me.Library.Cards)-1]; top.InstanceID != forest {
		t.Error("the land is not on top of the library")
	}
}

// TestMoxDiamondUndoAcrossTheLengPrompt is the frozen copy: rewinding
// into Library of Leng's prompt and answering it the other way gives
// one Mox on the battlefield and the land in the graveyard, not a
// second entry or a lost one.
func TestMoxDiamondUndoAcrossTheLengPrompt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushTokenReplacementCard(g, effectstest.LibraryOfLengOracle, "Library of Leng", "Artifact", me.ID)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	forest := effectstest.HandCard(me, "Forest", "Basic Land — Forest")

	mox := effectstest.CastMox(t, g)
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntryDiscardFromHand, me.ID, forest)
	snap := g.Clone()

	effectstest.AnswerLengPrompt(t, g, me.ID, true)
	if _, ok := effectstest.BattlefieldCard(g, mox); !ok {
		t.Fatal("setup: the Mox did not enter")
	}

	g.WithWriteLock(func() { g.RestoreFrom(snap) })
	me = g.Seats[g.Turn.ActiveSeat]
	if _, ok := effectstest.BattlefieldCard(g, mox); ok {
		t.Fatal("undo left the Mox on the battlefield")
	}
	effectstest.AnswerLengPrompt(t, g, me.ID, false)

	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == mox {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("%d Mox Diamonds on the battlefield after the replayed answer, want 1", n)
	}
	if !me.Graveyard.Contains(forest) {
		t.Error("declining Leng on the replay did not put the land into the graveyard")
	}
	if g.Stack.Contains(mox) {
		t.Error("the Mox was left on the stack after the replay")
	}
}

// TestMoxDiamondDeclinedUnderRestInPeaceIsExiled is CR 616.2: "put it
// into its owner's graveyard" is the modified event, and Rest in
// Peace's "exile it instead" applies to it.
func TestMoxDiamondDeclinedUnderRestInPeaceIsExiled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushTokenReplacementCard(g, effectstest.RestInPeaceOracle, "Rest in Peace", "Enchantment", opp.ID)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	effectstest.HandCard(me, "Forest", "Basic Land — Forest")

	mox := effectstest.CastMox(t, g)
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntryDiscardFromHand, me.ID)

	if !g.Exile.Contains(mox) {
		t.Fatal("under Rest in Peace a declined Mox should be exiled")
	}
	if me.Graveyard.Contains(mox) || g.Stack.Contains(mox) {
		t.Error("the Mox is in the graveyard or still on the stack")
	}
}

// TestMoxDiamondOrderedWithKismetStillAsks is ADR 0098 gap 4: the
// chosen-order loop had no branch for a hand choice, so a Mox ordered
// beside another entry replacement fired its decline blind.
func TestMoxDiamondOrderedWithKismetStillAsks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	kismet := effectstest.PushTokenReplacementCard(g, effectstest.KismetOracle, "Kismet", "Enchantment", opp.ID)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	forest := effectstest.HandCard(me, "Forest", "Basic Land — Forest")

	mox := effectstest.CastMox(t, g)
	order := effectstest.B06ReplacementOrderFor(g, me.ID)
	if order == nil {
		t.Fatal("Mox Diamond under an opponent's Kismet raised no CR 616 ordering prompt")
	}
	var kismetEff, moxEff game.ReplacementEffectID
	for _, id := range order.ReplacementEffectIDs {
		if _, src := g.ReplacementOptionMetaForEffect(id); src == kismet {
			kismetEff = id
		} else {
			moxEff = id
		}
	}
	if err := g.ResolveReplacementOrder(order.ID, me.ID, []game.ReplacementEffectID{kismetEff, moxEff}); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
	if effectstest.EntryChoiceOfKind(g, game.PendingChoiceEntryDiscardFromHand, me.ID) == nil {
		t.Fatal("the ordered Mox did not ask for a land card — its decline was fired blind")
	}
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntryDiscardFromHand, me.ID, forest)
	card, ok := effectstest.BattlefieldCard(g, mox)
	if !ok {
		t.Fatal("the Mox did not enter")
	}
	if !card.Tapped {
		t.Error("Kismet applied first, so the Mox should enter tapped")
	}
}

// TestMoxDiamondMovedByHandGoesToTheGraveyard is owner decision 6: the
// sandbox move cannot pause, so the question takes its declined branch.
func TestMoxDiamondMovedByHandGoesToTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	effectstest.HandCard(me, "Forest", "Basic Land — Forest")
	mox := effectstest.HandCardForTest(me, "Mox Diamond", "Artifact", effectstest.MoxDiamondOracle)

	err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, mox)
	if err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	if _, ok := effectstest.BattlefieldCard(g, mox); ok {
		t.Fatal("a Mox moved by hand entered for free")
	}
	if !me.Graveyard.Contains(mox) {
		t.Fatal("a Mox moved by hand is not in its owner's graveyard")
	}
}

// TestMoxDiamondPutFromHandCannotDiscardASiblingOfTheEntry is CR
// 614.13a: a land entering at the same time is not a candidate.
func TestMoxDiamondPutFromHandCannotDiscardASiblingOfTheEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	sibling := effectstest.HandCard(me, "Forest", "Basic Land — Forest")
	spare := effectstest.HandCard(me, "Island", "Basic Land — Island")
	mox := effectstest.HandCardForTest(me, "Mox Diamond", "Artifact", effectstest.MoxDiamondOracle)

	var entered []uuid.UUID
	g.WithWriteLock(func() {
		_ = g.PutOntoBattlefieldTogetherKeyedForEffect([]game.BatchEntry{
			{CardID: mox, From: game.ZoneHand},
			{CardID: sibling, From: game.ZoneHand},
		}, game.ZoneEntryOptions{}, effectstest.HookTail(func(_ *game.Game, o game.TailOutcome) error {
			ids := o.IDs
			entered = ids
			return nil
		}))
	})
	c := effectstest.EntryChoiceOfKind(g, game.PendingChoiceEntryDiscardFromHand, me.ID)
	if c == nil {
		t.Fatal("the Mox put from hand did not ask")
	}
	if len(c.ChooseCards) != 1 || c.ChooseCards[0] != spare {
		t.Fatalf("candidates %v, want only the Island — the Forest is entering with it (CR 614.13a)", c.ChooseCards)
	}
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntryDiscardFromHand, me.ID, spare)
	if len(entered) != 2 {
		t.Errorf("entered %v, want the Mox and the Forest", entered)
	}
}

// TestMoxDiamondReanimatedWithoutALandStaysInTheGraveyard is the
// destination-is-the-source-zone case: nothing moves.
func TestMoxDiamondReanimatedWithoutALandStaysInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	mox := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: mox, Name: "Mox Diamond", TypeLine: "Artifact",
		OracleID: effectstest.MoxDiamondOracle, Owner: me.ID, Controller: me.ID})

	var err error
	g.WithWriteLock(func() { err = g.ReturnFromGraveyardUnderControlForEffect(mox, game.ZoneBattlefield, me.ID) })
	if err != nil {
		t.Fatalf("ReturnFromGraveyardUnderControlForEffect: %v", err)
	}
	if _, ok := effectstest.BattlefieldCard(g, mox); ok {
		t.Fatal("a reanimated Mox with no land entered for free")
	}
	if !me.Graveyard.Contains(mox) {
		t.Fatal("the Mox left the graveyard")
	}
}

// TestRevealLandOrderedWithKismetStillAsks is gap 4's other half, and
// a regression test for the reveal-lands: before the shared dispatcher
// an ordered reveal-land entered tapped without being asked.
func TestRevealLandOrderedWithKismetStillAsks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	kismet := effectstest.PushTokenReplacementCard(g, effectstest.KismetOracle, "Kismet", "Enchantment", opp.ID)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	island := effectstest.SeedRevealLandHand(me, "Basic Land — Island")[0]

	land := effectstest.PlayLandFromHand(t, g, "Choked Estuary", effectstest.RevealLandOracleIDs["Choked Estuary"])
	order := effectstest.B06ReplacementOrderFor(g, me.ID)
	if order == nil {
		t.Fatal("no CR 616 ordering prompt")
	}
	var kismetEff, landEff game.ReplacementEffectID
	for _, id := range order.ReplacementEffectIDs {
		if _, src := g.ReplacementOptionMetaForEffect(id); src == kismet {
			kismetEff = id
		} else {
			landEff = id
		}
	}
	if err := g.ResolveReplacementOrder(order.ID, me.ID, []game.ReplacementEffectID{landEff, kismetEff}); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
	if effectstest.EntryRevealChoiceFor(g, me.ID) == nil {
		t.Fatal("the ordered reveal-land did not ask — its decline was fired blind")
	}
	effectstest.AnswerEntryReveal(t, g, me.ID, island)
	if _, ok := effectstest.BattlefieldCard(g, land); !ok {
		t.Fatal("the reveal-land did not enter")
	}
}

func TestMoxDiamondIsRegisteredComplete(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.MoxDiamondOracle)
	if !ok {
		t.Fatal("Mox Diamond is not registered")
	}
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("completeness %q with %d caveats", spec.Completeness, len(spec.Caveats))
	}
}
