package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exiled_this_way_cards_test.go — #870 at the catalog level.
//
// #866 taught the exile BATCH to count what landed; this is the same
// rule on the cards that were still exiling one at a time and reading
// the answer off "the call returned no error". It is not the same
// answer: the single-card exile returns nil when the leg PAUSED on the
// CR 903.9 prompt, so Winds of Abandon fetched its victim a basic land
// for a creature that was still on the battlefield, and fetched a
// second one if they then sent their commander to the command zone.
//
// The rule is CR 400.7, the batch's: "exiled this way" is the object
// that ARRIVED in exile. A commander that takes CR 903.9's offer left,
// but not to exile. A leg the CR 614 window cancelled never left at
// all.
//
// Every one of these cards is now a caller of the SAME batch — there
// is no per-card exile path to keep in step — so what each test pins
// is the clause, not the plumbing: the payout waits for the prompt,
// and it is sized to what reached exile.

// --- Winds of Abandon: the issue ---------------------------------

// windsBoard seeds a table for the overloaded Winds of Abandon: the
// caster, an opponent with `theirs` creatures and enough basics to be
// prompted, and a third seat with one creature and two basics.
func windsOverload(t *testing.T, g *game.Game, caster *game.Player) {
	t.Helper()
	advanceToMain(t, g)
	id := handCardFull(caster, "Winds of Abandon", "Sorcery", "", b18WindsOfAbandonOracle, []string{"W"})
	if err := g.CastSpell(caster.ID, id, game.CastSpellParams{AlternativeCost: "overload"}); err != nil {
		t.Fatalf("overloaded cast: %v", err)
	}
	passPriorityAroundTable(t, g)
}

// TestWindsOfAbandonDoesNotFetchForACommanderThatTookTheCommandZone
// is the issue's own board: three creatures you don't control, one of
// them an opponent's commander. That opponent lost ONE creature to the
// exile, so they search for one basic land, not two.
func TestWindsOfAbandonDoesNotFetchForACommanderThatTookTheCommandZone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	commander := b36Commander(g, opp.ID, "Their Commander")
	theirBear := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	otherBear := b16Creature(g, other.ID, "Other Bear", "Creature — Bear", 2, 2, "G")
	seedSearchLibrary(opp,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Plains", "Basic Land — Plains"),
		searchTestLand("Island", "Basic Land — Island"),
	)
	seedSearchLibrary(other,
		searchTestLand("Swamp", "Basic Land — Swamp"),
		searchTestLand("Mountain", "Basic Land — Mountain"),
	)

	windsOverload(t, g, me)

	// The sweep is not over: the commander's owner is being asked
	// about the command zone, and "for each creature exiled this way"
	// cannot be counted until they answer.
	if searchChoiceFor(g, opp.ID) != nil || searchChoiceFor(g, other.ID) != nil {
		t.Fatal("no search is queued while the CR 903.9 prompt is open — the count is not knowable yet")
	}
	b36AcceptCommandZone(t, g, opp.ID)

	if !opp.Command.Contains(commander) {
		t.Fatal("the commander took the offer and is in the command zone")
	}
	if !g.Exile.Contains(theirBear) || !g.Exile.Contains(otherBear) {
		t.Fatal("the rest of the sweep still reaches exile")
	}
	pc := searchChoiceFor(g, opp.ID)
	if pc == nil || pc.Count != 1 {
		t.Fatalf("the commander's controller searches for %v, want a ONE-card prompt — their commander "+
			"went to the command zone, not to exile, so it was not exiled this way (CR 400.7)", pc)
	}
	if pc := searchChoiceFor(g, other.ID); pc == nil || pc.Count != 1 {
		t.Errorf("the third seat lost one creature and searches for %v, want one", pc)
	}
}

// TestWindsOfAbandonDoesNotFetchForAnExileTheWindowCancelled — the
// other way a leg fails to land, and the one that needs no prompt at
// all: a creature the CR 614 window kept on the battlefield was not
// exiled this way either.
func TestWindsOfAbandonDoesNotFetchForAnExileTheWindowCancelled(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	doomed := b16Creature(g, opp.ID, "Doomed", "Creature — Bear", 2, 2, "G")
	saved := b16Creature(g, opp.ID, "Saved", "Creature — Bear", 2, 2, "G")
	registerExitReplacement(t, g, saved, "it can't be exiled",
		func(ev *game.ReplacementEvent) { ev.Cancel() })
	seedSearchLibrary(opp,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Plains", "Basic Land — Plains"),
		searchTestLand("Island", "Basic Land — Island"),
	)

	windsOverload(t, g, me)

	if !g.Battlefield.Contains(saved) {
		t.Error("an exile the window cancelled leaves its creature on the battlefield")
	}
	if !g.Exile.Contains(doomed) {
		t.Error("the unprotected creature is exiled — control case broken")
	}
	pc := searchChoiceFor(g, opp.ID)
	if pc == nil || pc.Count != 1 {
		t.Fatalf("the controller searches for %v, want a ONE-card prompt — only one of the two "+
			"creatures reached exile", pc)
	}
}

// TestUndoAcrossWindsOfAbandonPausedLegReplays is the undo contract
// the continuation signs, on the card: rewind into the open CR 903.9
// prompt, answer it the OTHER way, and the search is sized to that
// answer rather than to the one that was undone.
func TestUndoAcrossWindsOfAbandonPausedLegReplays(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oppID := opp.ID
	commander := b36Commander(g, oppID, "Their Commander")
	b16Creature(g, oppID, "Their Bear", "Creature — Bear", 2, 2, "G")
	seedSearchLibrary(opp,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Plains", "Basic Land — Plains"),
		searchTestLand("Island", "Basic Land — Island"),
	)

	windsOverload(t, g, me)
	promptOpen := g.Clone()

	b36AcceptCommandZone(t, g, oppID)
	if pc := searchChoiceFor(g, oppID); pc == nil || pc.Count != 1 {
		t.Fatalf("accepting the command zone searches for %v, want one", pc)
	}

	g.WithWriteLock(func() { g.RestoreFrom(promptOpen) })
	if !g.Battlefield.Contains(commander) {
		t.Fatal("the rewind puts the commander back on the battlefield")
	}
	if searchChoiceFor(g, oppID) != nil {
		t.Fatal("the rewind takes the search back too")
	}

	b21DeclineCommandZone(t, g, oppID)
	if !g.Exile.Contains(commander) {
		t.Fatal("declining exiles the commander after all")
	}
	if pc := searchChoiceFor(g, oppID); pc == nil || pc.Count != 2 {
		t.Errorf("declining searches for %v, want TWO — both creatures reached exile on the replay", pc)
	}
}

// --- the other single-card read-backs ----------------------------

// TestCurseOfTheSwineMakesABoarOnlyForACreatureThatReachedExile —
// "for each creature exiled this way, its controller creates a 2/2
// green Boar". One Boar, not two.
func TestCurseOfTheSwineMakesABoarOnlyForACreatureThatReachedExile(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := g.Seats[0], g.Seats[1]
	commander := b36Commander(g, opp.ID, "Their Commander")
	bear := seedCreature(g, "Their Bear", opp.ID)

	castXSpell(t, g, "Curse of the Swine", "Sorcery", b07CurseOfTheSwineOracle, "{X}{U}{U}", 2,
		[]game.TargetRef{{Kind: game.TargetCard, ID: commander}, {Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)

	if n := countBattlefieldNamed(g, opp.ID, "Boar"); n != 0 {
		t.Fatalf("%d Boars while the CR 903.9 prompt is open, want none — the clause waits", n)
	}
	b36AcceptCommandZone(t, g, opp.ID)

	if !opp.Command.Contains(commander) || !g.Exile.Contains(bear) {
		t.Fatal("the commander is in the command zone and the Bear in exile")
	}
	if n := countBattlefieldNamed(g, opp.ID, "Boar"); n != 1 {
		t.Errorf("%d Boars, want 1 — the commander went to the command zone, not to exile", n)
	}
}

// TestBindingOfTheTitansGainsLifeOnlyForACardThatReachedExile — "for
// each creature card exiled this way, you gain 1 life", off a
// graveyard, which is a CR 903.9 zone too.
//
// The chapter body is driven directly: a Saga's chapter II is a
// targeted trigger, and what is under test is the clause rather than
// the lore counter.
func TestBindingOfTheTitansGainsLifeOnlyForACardThatReachedExile(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirCommander := b17GraveyardCard(opp, "Their Commander", "Legendary Creature — Human", "{2}{G}")
	markCommanderCard(t, g, opp, theirCommander)
	theirBear := b17GraveyardCard(opp, "Their Bear", "Creature — Bear", "{1}{G}")

	before := me.Life
	item := &game.StackItem{
		Controller:   me.ID,
		SourceCardID: uuid.New(),
		Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: theirCommander},
			{Kind: game.TargetCard, ID: theirBear},
		},
	}
	g.WithWriteLock(func() {
		if err := titansExileFromGraveyards(g, item); err != nil {
			t.Fatalf("titansExileFromGraveyards: %v", err)
		}
	})

	if me.Life != before {
		t.Fatalf("life %d → %d while the CR 903.9 prompt is open, want unchanged", before, me.Life)
	}
	b36AcceptCommandZone(t, g, opp.ID)

	if !opp.Command.Contains(theirCommander) || !g.Exile.Contains(theirBear) {
		t.Fatal("the commander card is in the command zone and the Bear in exile")
	}
	if want := before + 1; me.Life != want {
		t.Errorf("life %d → %d, want %d — one creature card reached exile", before, me.Life, want)
	}
}

// TestMariPutsNoHitCounterOnACardSheDidNotExile — "exile it with a hit
// counter on it" is one instruction, and the counter belongs to the
// card in exile.
func TestMariPutsNoHitCounterOnACardSheDidNotExile(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Mari, the Killing Quill", "Legendary Creature — Vampire Assassin", b24MariOracle, 3, 2)
	commander := b36Commander(g, opp.ID, "Their Commander")

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(commander) })
	// CR 903.9 asks once on the way to the graveyard. Declined, so
	// the card is in the graveyard and Mari's trigger sees a creature
	// an opponent controlled die.
	b21DeclineCommandZone(t, g, opp.ID)
	passPriorityAroundTable(t, g)

	if b12Counter(t, g, commander, "hit") != 0 {
		t.Fatal("no hit counter while the exile's own CR 903.9 prompt is open")
	}
	b36AcceptCommandZone(t, g, opp.ID)

	if !opp.Command.Contains(commander) {
		t.Fatal("the commander took the offer and is in the command zone")
	}
	if got := b12Counter(t, g, commander, "hit"); got != 0 {
		t.Errorf("hit counters = %d, want 0 — Mari exiled nothing, so there was nothing to mark", got)
	}
}

// TestGreenwardenReturnsNothingUntilItsOwnExileLands — the "if you do"
// gate. The old read-back looked in exile on the next line, found the
// Greenwarden still in the graveyard with its owner's prompt open, and
// dropped the second half of its own trigger on the floor.
func TestGreenwardenReturnsNothingUntilItsOwnExileLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	buried := b17GraveyardCard(me, "Dead Rock", "Artifact", "{1}")
	warden := castCatalogSpell(t, g, "Greenwarden of Murasa", "Creature — Elemental", b35GreenwardenOracle, nil)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)

	b18Kill(t, g, warden)
	// A Greenwarden that is also its deck's commander: the exile out
	// of the graveyard is a CR 903.9 move, so the "if you do" has to
	// wait for the answer.
	markCommanderCard(t, g, me, warden)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, buried)
	passPriorityAroundTable(t, g)

	if me.Hand.Contains(buried) {
		t.Fatal("nothing returns while the Greenwarden's own CR 903.9 prompt is open")
	}
	b21DeclineCommandZone(t, g, me.ID)

	if !g.Exile.Contains(warden) {
		t.Fatal("declining exiles the Greenwarden from the graveyard")
	}
	if !me.Hand.Contains(buried) {
		t.Error("and because it did, the chosen card comes back — the 'if you do' is satisfied late, not lost")
	}
}

// TestWaterbendersRestorationSchedulesOnlyWhatReachedExile — the
// delayed return's payload is the cards that are in exile to be
// returned, which is not the same as the targets the exile was
// attempted on.
func TestWaterbendersRestorationSchedulesOnlyWhatReachedExile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	commander := b36Commander(g, me.ID, "My Commander")
	probe := pushFlickerCreature(g, me.ID, "Probe", flickerProbeOracle)
	flickerProbeETBs = 0
	helpers := pushTapCostSoldiers(g, me.ID, 2)

	if err := castRestoration(t, g, game.CastSpellParams{
		XValue: 2,
		TapIDs: helpers,
		Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: commander},
			{Kind: game.TargetCard, ID: probe},
		},
	}); err != nil {
		t.Fatalf("CastSpell Waterbender's Restoration: %v", err)
	}
	passPriorityAroundTable(t, g)

	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("%d delayed returns scheduled while the CR 903.9 prompt is open, want none — "+
			"what is coming back is not known yet", len(g.DelayedTriggers))
	}
	b36AcceptCommandZone(t, g, me.ID)

	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("delayed-return queue holds %d triggers, want 1", len(g.DelayedTriggers))
	}
	if got := g.DelayedTriggers[0].Cards; len(got) != 1 || got[0] != probe {
		t.Errorf("the delayed return carries %v, want just the Probe — the commander went to the "+
			"command zone and is not coming back from exile", got)
	}
}

// markCommanderCard makes the named card in one of `owner`'s zones a
// commander, so a move to a CR 903.9 destination offers the command
// zone. Tests that need the prompt rather than a whole commander
// deck set the flag directly, the way b36Commander does on the way
// onto the battlefield.
func markCommanderCard(t *testing.T, g *game.Game, owner *game.Player, id uuid.UUID) {
	t.Helper()
	marked := false
	g.WithWriteLock(func() {
		for _, z := range []*game.Zone{owner.Graveyard, owner.Hand, owner.Library, g.Battlefield, g.Exile} {
			if z == nil {
				continue
			}
			for i := range z.Cards {
				if z.Cards[i].InstanceID == id {
					z.Cards[i].IsCommander = true
					marked = true
					return
				}
			}
		}
	})
	if !marked {
		t.Fatalf("no card %s in any of %s's zones to mark as a commander", id, owner.ID)
	}
}
