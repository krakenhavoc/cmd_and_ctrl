package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// mkm_facedown_followups_cards_test.go — #5241: three face-down pieces
// and the cards on them. Karlov Watchdog's "can't be turned face up"
// (game/turn_face_up_gate.go), Goblin Maskmaker's discount on face-down
// spells (NextSpellPromise.FaceDownOnly) and Oblivious Bookworm's
// "entered face down / turned face up this turn" (the turn tally). Each
// test plays the real catalog entry.

const (
	karlovWatchdogOracle    = "73f44500-fe83-40fa-b6f2-ed5673d0edcc"
	goblinMaskmakerOracle   = "ff4b9723-d5f1-40d2-aa35-fa3d2a6f4b5c"
	obliviousBookwormOracle = "300fd0a2-23e6-4da9-bfaa-ad626097d5b5"
)

func TestMkmFaceDownFollowUpCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		karlovWatchdogOracle:    "Karlov Watchdog",
		goblinMaskmakerOracle:   "Goblin Maskmaker",
		obliviousBookwormOracle: "Oblivious Bookworm",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- fixtures ------------------------------------------------------------------

// fdNightwatch casts Museum Nightwatch face down (disguise) for the
// active player and resolves it: a face-down permanent that enters.
func fdNightwatch(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	return mkmMPCastFaceDown(t, g, "Museum Nightwatch", "Creature — Centaur Soldier", mkmMPNightwatchOracle, 3, 2)
}

// fdSeatFaceDown puts a disguised Museum Nightwatch onto the battlefield
// under owner's control with no entry announced: a face-down permanent
// that has been there since an earlier turn, so the turn tally has not
// seen it enter. Only its controller knows what it is.
func fdSeatFaceDown(g *game.Game, owner uuid.UUID) uuid.UUID {
	c := game.Card{
		InstanceID: uuid.New(), Name: "Museum Nightwatch", TypeLine: "Creature — Centaur Soldier",
		OracleID: mkmMPNightwatchOracle, Power: 3, Toughness: 2, Owner: owner, Controller: owner,
	}
	c.SetFaceDown(game.FaceDownDisguised)
	// CR 708.5: its controller may look at it, and nobody else.
	c.KnownBy = map[uuid.UUID]bool{owner: true}
	return pushBattlefieldCardWithTimestamp(g, c)
}

// fdTurnUpTiming is the turn_face_up window for `player` on `id`.
func fdTurnUpTiming(g *game.Game, player, id uuid.UUID) bool {
	var ok bool
	g.WithWriteLock(func() {
		c, found := g.LookupCardForEffect(id)
		ok = found && g.SpecialActionTimingOKLocked(player, c, game.SpecialActionTurnFaceUp)
	})
	return ok
}

// fdTurnUpOffered reports whether the legal-move enumerator offers
// `seat` the turn_face_up special action on `id`.
func fdTurnUpOffered(g *game.Game, seat, id uuid.UUID) bool {
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Type == legal.TypeSpecialAction && m.Source == id {
			return true
		}
	}
	return false
}

// fdTurnUpRow is the turn_face_up row the view shows `viewer` on `id`.
func fdTurnUpRow(t *testing.T, g *game.Game, viewer, id uuid.UUID) protocol.SpecialActionView {
	t.Helper()
	v := protocol.ViewOfGameFor(g, viewer.String())
	for _, c := range v.Battlefield.Cards {
		if c.InstanceID != id.String() {
			continue
		}
		for _, sa := range c.SpecialActions {
			if sa.Kind == string(game.SpecialActionTurnFaceUp) {
				return sa
			}
		}
	}
	t.Fatalf("the view shows %s no turn-face-up row on %s", viewer, id)
	return protocol.SpecialActionView{}
}

// --- Karlov Watchdog -----------------------------------------------------------

// The ruling: an opponent can't turn their face-down permanent face up
// while it is the Watchdog's controller's turn. The engine refuses it
// before charging, the enumerator does not offer it, the view greys the
// row, and an effect that would turn it up does nothing.
func TestKarlovWatchdogStopsAnOpponentTurningFaceUpDuringYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := mkmMPOpponents(g)[0]
	theirs := fdSeatFaceDown(g, opp.ID)
	pushCatalogPermanent(g, me.ID, "Karlov Watchdog", "Creature — Dog", karlovWatchdogOracle, false)

	if fdTurnUpTiming(g, opp.ID, theirs) {
		t.Fatal("the turn-face-up window is open for an opponent during the Watchdog's controller's turn")
	}
	g.WithWriteLock(func() { opp.ManaPool.AddMana(manaTokens("W", "C")...) })
	err := g.PerformSpecialAction(opp.ID, theirs, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true})
	if !errors.Is(err, game.ErrSpecialActionTiming) {
		t.Fatalf("PerformSpecialAction = %v, want ErrSpecialActionTiming", err)
	}
	if !findCardAnywhere(t, g, theirs).FaceDownIsPermanent() {
		t.Fatal("the opponent's permanent was turned face up")
	}
	if got := opp.ManaPool.Total(); got != 2 {
		t.Errorf("the refused action spent mana: %d left, want 2", got)
	}

	// Give the opponent priority, so the enumerator would offer it.
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if fdTurnUpOffered(g, opp.ID, theirs) {
		t.Error("the enumerator offers the opponent a turn face up the engine refuses")
	}
	if row := fdTurnUpRow(t, g, opp.ID, theirs); row.Available {
		t.Error("the view's turn-face-up row is available, want greyed")
	}

	var turned bool
	g.WithWriteLock(func() { turned = g.TurnFaceUpForEffect(uuid.Nil, me.ID, theirs) })
	if turned || !findCardAnywhere(t, g, theirs).FaceDownIsPermanent() {
		t.Error("an effect turned the opponent's permanent face up under the Watchdog")
	}
}

// Only opponents, and only during your turn: your own face-down
// permanent turns up under your Watchdog, and an opponent's Watchdog
// does not stop you on your own turn.
func TestKarlovWatchdogStopsOnlyOpponentsOnlyDuringYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := mkmMPOpponents(g)[0]
	mine := fdNightwatch(t, g)
	pushCatalogPermanent(g, me.ID, "Karlov Watchdog", "Creature — Dog", karlovWatchdogOracle, false)
	if !fdTurnUpTiming(g, me.ID, mine) {
		t.Error("my own Watchdog stops me turning my permanent face up")
	}
	pushCatalogPermanent(g, opp.ID, "Karlov Watchdog", "Creature — Dog", karlovWatchdogOracle, false)
	if !fdTurnUpTiming(g, me.ID, mine) {
		t.Error("an opponent's Watchdog stops me on my own turn")
	}
	mkmMPTurnFaceUp(t, g, me, mine, "W", "C")
}

// The restriction is the Watchdog's static, read live: once it leaves,
// the opponent may turn the permanent face up, and the enumerator offers
// it again.
func TestKarlovWatchdogLeavingLiftsTheRestriction(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := mkmMPOpponents(g)[0]
	theirs := fdSeatFaceDown(g, opp.ID)
	dog := pushCatalogPermanent(g, me.ID, "Karlov Watchdog", "Creature — Dog", karlovWatchdogOracle, false)
	if fdTurnUpTiming(g, opp.ID, theirs) {
		t.Fatal("setup: the window is open with the Watchdog out")
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(dog); err != nil {
			t.Fatalf("destroy the Watchdog: %v", err)
		}
	})
	if !fdTurnUpTiming(g, opp.ID, theirs) {
		t.Fatal("the window stays shut after the Watchdog left")
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	g.WithWriteLock(func() { opp.ManaPool.AddMana(manaTokens("W", "C")...) })
	if !fdTurnUpOffered(g, opp.ID, theirs) {
		t.Error("the enumerator does not offer the turn face up once the Watchdog is gone")
	}
	mkmMPTurnFaceUp(t, g, opp, theirs)
}

// Nothing is stored: a restored table rebuilds the restriction from the
// Watchdog on the battlefield.
func TestKarlovWatchdogRestrictionSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := mkmMPOpponents(g)[0]
	theirs := fdSeatFaceDown(g, opp.ID)
	pushCatalogPermanent(g, me.ID, "Karlov Watchdog", "Creature — Dog", karlovWatchdogOracle, false)

	restored := restoreRoundTrip(t, g, true)
	if fdTurnUpTiming(restored, opp.ID, theirs) {
		t.Error("the restored table lets the opponent turn the permanent face up")
	}
}

// "Whenever you attack with three or more creatures": three attackers
// pump every creature you control, the Watchdog included.
func TestKarlovWatchdogPumpsYourCreaturesWhenYouAttackWithThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := mkmMPOpponents(g)[0]
	dog := b12Push(g, me.ID, "Karlov Watchdog", "Creature — Dog", karlovWatchdogOracle, 3, 2)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	home := b12Creature(g, me.ID, "Bear C", "Creature — Bear", 2, 2)

	declareAttack(t, g, opp.ID, dog, a, b)
	if got := triggersOnStackFrom(g, dog); got != 1 {
		t.Fatalf("%d Watchdog triggers for one three-creature attack, want 1", got)
	}
	passPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{dog: 4, a: 3, b: 3, home: 3} {
		if got := currentPowerFor(t, g, id); got != want {
			t.Errorf("power %d, want %d", got, want)
		}
	}
}

// Once per combat, not once per turn: an additional combat in which you
// attack with three again triggers it again.
func TestKarlovWatchdogTriggersAgainInAnAdditionalCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := mkmMPOpponents(g)[0]
	dog := b12Push(g, me.ID, "Karlov Watchdog", "Creature — Dog", karlovWatchdogOracle, 3, 2)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)

	declareAttack(t, g, opp.ID, dog, a, b)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		g.AddPhasesForEffect(uuid.Nil, game.PhaseAnchor{Kind: game.AnchorThisPhase}, game.PhaseKindCombat)
	})
	for i := 0; i < 60 && !(g.Turn.Step == game.StepDeclareAttackers && g.Turn.PhaseOrdinal == 2); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.PhaseOrdinal != 2 || g.Turn.Step != game.StepDeclareAttackers {
		t.Fatalf("never reached the additional combat (step %s, phase %d)", g.Turn.Step, g.Turn.PhaseOrdinal)
	}
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{dog, a, b} {
			findBattlefieldCardForTest(g, id).Tapped = false
		}
	})
	for _, id := range []uuid.UUID{dog, a, b} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker in the additional combat: %v", err)
		}
	}
	lockInAttacks(t, g)
	if got := triggersOnStackFrom(g, dog); got != 1 {
		t.Fatalf("%d Watchdog triggers in the additional combat, want 1", got)
	}
	passPriorityAroundTable(t, g)
	if got := currentPowerFor(t, g, a); got != 4 {
		t.Errorf("a bear's power after two triggers is %d, want 4", got)
	}
}

func TestKarlovWatchdogIgnoresAnAttackWithTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := mkmMPOpponents(g)[0]
	dog := b12Push(g, me.ID, "Karlov Watchdog", "Creature — Dog", karlovWatchdogOracle, 3, 2)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)

	declareAttack(t, g, opp.ID, dog, a)
	if got := triggersOnStackFrom(g, dog); got != 0 {
		t.Errorf("%d Watchdog triggers for a two-creature attack, want 0", got)
	}
}

// --- Goblin Maskmaker ----------------------------------------------------------

// fdMaskmakerAttacks attacks with the Maskmaker, resolves its trigger
// and walks on to the postcombat main phase.
func fdMaskmakerAttacks(t *testing.T, g *game.Game, me, opp *game.Player) {
	t.Helper()
	mask := b12Push(g, me.ID, "Goblin Maskmaker", "Creature — Goblin Citizen", goblinMaskmakerOracle, 1, 2)
	declareAttack(t, g, opp.ID, mask)
	if got := triggersOnStackFrom(g, mask); got != 1 {
		t.Fatalf("%d Maskmaker triggers, want 1", got)
	}
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepPostcombatMain)
}

// fdCastNightwatch puts Museum Nightwatch in the active player's hand
// and casts it, strictly, from `mana` alone.
func fdCastNightwatch(g *game.Game, p *game.Player, alt string, mana ...string) error {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Museum Nightwatch", TypeLine: "Creature — Centaur Soldier",
		OracleID: mkmMPNightwatchOracle, ManaCost: "{1}{W}", Power: 3, Toughness: 2, Owner: p.ID, Controller: p.ID,
	})
	g.WithWriteLock(func() { p.ManaPool.AddMana(manaTokens(mana...)...) })
	return g.CastSpell(p.ID, id, game.CastSpellParams{AlternativeCost: alt, Strict: true})
}

// After the Maskmaker attacks, a disguise cast's {3} costs {2} for the
// rest of the turn.
func TestGoblinMaskmakerMakesFaceDownSpellsCostOneLess(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := mkmMPOpponents(g)[0]
	fdMaskmakerAttacks(t, g, me, opp)
	if err := fdCastNightwatch(g, me, "disguise", "C", "C"); err != nil {
		t.Fatalf("a face-down cast for {2} after the Maskmaker attacked: %v", err)
	}
	passPriorityAroundTable(t, g)
	// A second face-down spell this turn is discounted too.
	if err := fdCastNightwatch(g, me, "disguise", "R", "R"); err != nil {
		t.Errorf("a second face-down cast for {2}: %v", err)
	}
}

// Without the attack the {3} is {3}; and a face-up spell is not
// discounted by the promise.
func TestGoblinMaskmakerDiscountsOnlyFaceDownSpellsAfterItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Goblin Maskmaker", "Creature — Goblin Citizen", goblinMaskmakerOracle, 1, 2)
	advanceToMain(t, g)
	if err := fdCastNightwatch(g, me, "disguise", "C", "C"); err == nil {
		t.Fatal("a face-down cast for {2} with no attack")
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	fdMaskmakerAttacks(t, g2, me2, mkmMPOpponents(g2)[0])
	if err := fdCastNightwatch(g2, me2, "", "W"); err == nil {
		t.Error("a face-up {1}{W} spell was cast for {W} under the Maskmaker's promise")
	}
}

// The promise is the player's for the turn, so a restored table keeps
// the discount.
func TestGoblinMaskmakerPromiseSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	fdMaskmakerAttacks(t, g, me, mkmMPOpponents(g)[0])

	restored := restoreRoundTrip(t, g, true)
	rme := restored.Seats[restored.Turn.ActiveSeat]
	if err := fdCastNightwatch(restored, rme, "disguise", "C", "C"); err != nil {
		t.Errorf("a face-down cast for {2} on the restored table: %v", err)
	}
}

// --- Oblivious Bookworm --------------------------------------------------------

// fdBookwormEndStep walks to the end step, resolves the Bookworm's
// trigger and answers yes. It returns the hand size before the draw.
func fdBookwormEndStep(t *testing.T, g *game.Game, me *game.Player) int {
	t.Helper()
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	ask := latestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil || ask.Chooser != me.ID {
		t.Fatal("no \"draw a card?\" prompt at the end step")
	}
	hand := me.Hand.Size()
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d after drawing, want %d", me.Hand.Size(), hand+1)
	}
	return hand
}

func TestObliviousBookwormLootsOnAnOrdinaryTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	msc18Library(me, 6)
	b12Push(g, me.ID, "Oblivious Bookworm", "Creature — Human Wizard", obliviousBookwormOracle, 2, 3)
	fdBookwormEndStep(t, g, me)
	if discardChoiceFor(g, me.ID) == nil {
		t.Error("no discard after drawing on a turn with nothing face down")
	}
}

// A permanent that entered face down under your control this turn
// spares the discard, even after it has been turned face up.
func TestObliviousBookwormSkipsTheDiscardAfterAFaceDownEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	msc18Library(me, 6)
	b12Push(g, me.ID, "Oblivious Bookworm", "Creature — Human Wizard", obliviousBookwormOracle, 2, 3)
	fdNightwatch(t, g)
	fdBookwormEndStep(t, g, me)
	if discardChoiceFor(g, me.ID) != nil {
		t.Error("asked for a discard after a permanent entered face down this turn")
	}
}

// Turning a permanent face up spares it too, for one that has been face
// down since an earlier turn.
func TestObliviousBookwormSkipsTheDiscardAfterYouTurnedAPermanentFaceUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	msc18Library(me, 6)
	b12Push(g, me.ID, "Oblivious Bookworm", "Creature — Human Wizard", obliviousBookwormOracle, 2, 3)
	old := fdSeatFaceDown(g, me.ID)
	var entered bool
	g.WithWriteLock(func() { entered = g.PermanentEnteredFaceDownThisTurn(me.ID) })
	if entered {
		t.Fatal("setup: the seated permanent counted as entering face down")
	}
	mkmMPTurnFaceUp(t, g, me, old, "W", "C")
	fdBookwormEndStep(t, g, me)
	if discardChoiceFor(g, me.ID) != nil {
		t.Error("asked for a discard after you turned a permanent face up this turn")
	}
}

// An opponent's face-down entry is theirs: it does not spare you.
func TestObliviousBookwormIgnoresAnOpponentsFaceDownEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := mkmMPOpponents(g)[0]
	msc18Library(me, 6)
	b12Push(g, me.ID, "Oblivious Bookworm", "Creature — Human Wizard", obliviousBookwormOracle, 2, 3)
	msc18Library(opp, 2)
	g.WithWriteLock(func() {
		if _, err := g.ManifestForEffect(opp.ID); err != nil {
			t.Fatalf("manifest: %v", err)
		}
	})
	var theirs bool
	g.WithWriteLock(func() { theirs = g.PermanentEnteredFaceDownThisTurn(opp.ID) })
	if !theirs {
		t.Fatal("setup: the opponent's manifest did not count as entering face down")
	}
	fdBookwormEndStep(t, g, me)
	if discardChoiceFor(g, me.ID) == nil {
		t.Error("an opponent's face-down entry spared your discard")
	}
}

// The tally is written to the snapshot, so a table restored between the
// face-down entry and the end step still spares the discard.
func TestObliviousBookwormTallySurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	msc18Library(me, 6)
	b12Push(g, me.ID, "Oblivious Bookworm", "Creature — Human Wizard", obliviousBookwormOracle, 2, 3)
	fdNightwatch(t, g)

	restored := restoreRoundTrip(t, g, true)
	rme := restored.Seats[restored.Turn.ActiveSeat]
	fdBookwormEndStep(t, restored, rme)
	if discardChoiceFor(restored, rme.ID) != nil {
		t.Error("the restored table asked for a discard after a face-down entry")
	}
}
