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
// is the clause, not the plumbing: the payout is sized to what reached
// exile.
//
// ADR 0115: a commander no longer pauses an exile. It is exiled like
// any other card, so it IS exiled this way and counts, and CR 903.9a
// asks its owner about the command zone afterwards. The commander
// tests below pin that order; the window-cancelled tests still pin
// "only what reached exile".

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

// TestWindsOfAbandonFetchesForAnExiledCommander is the issue's own
// board: three creatures you don't control, one of them an opponent's
// commander. Since ADR 0115 the commander is exiled with the rest
// (CR 903.9a asks its owner only afterwards), so that opponent lost TWO
// creatures to the exile and searches for two basic lands.
func TestWindsOfAbandonFetchesForAnExiledCommander(t *testing.T) {
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

	if !g.Exile.Contains(commander) || !g.Exile.Contains(theirBear) || !g.Exile.Contains(otherBear) {
		t.Fatal("the whole sweep reaches exile, the commander included")
	}
	pc := searchChoiceFor(g, opp.ID)
	if pc == nil || pc.Count != 2 {
		t.Fatalf("the commander's controller searches for %v, want a TWO-card prompt — their commander "+
			"was exiled this way like the Bear", pc)
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

// --- the other single-card read-backs ----------------------------

// TestCurseOfTheSwineMakesABoarForAnExiledCommander — "for each
// creature exiled this way, its controller creates a 2/2 green Boar".
// The commander is exiled with the Bear (ADR 0115), so two Boars, and
// the CR 903.9a answer that follows does not take one back.
func TestCurseOfTheSwineMakesABoarForAnExiledCommander(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := g.Seats[0], g.Seats[1]
	commander := b36Commander(g, opp.ID, "Their Commander")
	bear := seedCreature(g, "Their Bear", opp.ID)

	castXSpell(t, g, "Curse of the Swine", "Sorcery", b07CurseOfTheSwineOracle, "{X}{U}{U}", 2,
		[]game.TargetRef{{Kind: game.TargetCard, ID: commander}, {Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)

	if !g.Exile.Contains(commander) || !g.Exile.Contains(bear) {
		t.Fatal("the commander and the Bear are both exiled")
	}
	if n := countBattlefieldNamed(g, opp.ID, "Boar"); n != 2 {
		t.Fatalf("%d Boars, want 2 — both creatures were exiled this way", n)
	}
	answerCommanderReturn(t, g, opp.ID, true)
	if !opp.Command.Contains(commander) {
		t.Fatal("yes sends the exiled commander to the command zone")
	}
	if n := countBattlefieldNamed(g, opp.ID, "Boar"); n != 2 {
		t.Errorf("%d Boars after the answer, want 2", n)
	}
}

// TestBindingOfTheTitansGainsLifeForAnExiledCommanderCard — "for each
// creature card exiled this way, you gain 1 life", off a graveyard. A
// commander card is exiled out of it like any other (ADR 0115), so two
// life, whatever its owner answers CR 903.9a afterwards.
//
// The chapter body is driven directly: a Saga's chapter II is a
// targeted trigger, and what is under test is the clause rather than
// the lore counter.
func TestBindingOfTheTitansGainsLifeForAnExiledCommanderCard(t *testing.T) {
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

	if !g.Exile.Contains(theirCommander) || !g.Exile.Contains(theirBear) {
		t.Fatal("the commander card and the Bear are both exiled")
	}
	if want := before + 2; me.Life != want {
		t.Fatalf("life %d → %d, want %d — two creature cards reached exile", before, me.Life, want)
	}
	answerCommanderReturn(t, g, opp.ID, true)
	if !opp.Command.Contains(theirCommander) {
		t.Fatal("yes sends the commander card to the command zone")
	}
	if want := before + 2; me.Life != want {
		t.Errorf("life %d → %d after the answer, want %d", before, me.Life, want)
	}
}

// TestMariPutsNoHitCounterOnACardSheDidNotExile — "exile it with a hit
// counter on it" is one instruction, and the counter belongs to the
// card in exile.
//
// ADR 0115: the commander dies, so Mari triggers. CR 903.9a asks its
// owner before the trigger goes on the stack (CR 704.3). Sent home, the
// commander is a new object Mari's trigger cannot find (CR 400.7,
// CR 603.6c): nothing is exiled and nothing is marked. Left in the
// graveyard, Mari exiles it with a hit counter, and CR 903.9a asks
// again, because it was put into exile.
func TestMariPutsNoHitCounterOnACardSheDidNotExile(t *testing.T) {
	for _, tc := range []struct {
		name     string
		goesHome bool
	}{{"sent home before the trigger", true}, {"left in the graveyard", false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			b12Push(g, me.ID, "Mari, the Killing Quill", "Legendary Creature — Vampire Assassin", b24MariOracle, 3, 2)
			commander := b36Commander(g, opp.ID, "Their Commander")

			g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(commander) })
			if !opp.Graveyard.Contains(commander) {
				t.Fatal("the destroyed commander is in its owner's graveyard")
			}
			answerCommanderReturn(t, g, opp.ID, tc.goesHome)
			passPriorityAroundTable(t, g)

			if tc.goesHome {
				if !opp.Command.Contains(commander) {
					t.Fatal("the commander took the offer and is in the command zone")
				}
				if got := b12Counter(t, g, commander, "hit"); got != 0 {
					t.Errorf("hit counters = %d, want 0 — Mari exiled nothing, so there was nothing to mark", got)
				}
				return
			}
			if !g.Exile.Contains(commander) {
				t.Fatal("Mari exiles the commander left in the graveyard")
			}
			if got := b12Counter(t, g, commander, "hit"); got != 1 {
				t.Errorf("hit counters = %d, want 1 on the card Mari exiled", got)
			}
			if commanderReturnPromptFor(g, opp.ID) == nil {
				t.Error("a commander put into exile is offered the command zone again (CR 903.9a)")
			}
		})
	}
}

// TestGreenwardenReturnsTheCardOnceItsOwnExileLands — the "if you do"
// gate. A Greenwarden that is also its deck's commander exiles itself
// out of the graveyard like any other card (ADR 0115), so the "if you
// do" is satisfied at once and the chosen card comes back; CR 903.9a
// then asks its owner about the command zone.
func TestGreenwardenReturnsTheCardOnceItsOwnExileLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	buried := b17GraveyardCard(me, "Dead Rock", "Artifact", "{1}")
	warden := castCatalogSpell(t, g, "Greenwarden of Murasa", "Creature — Elemental", b35GreenwardenOracle, nil)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)

	b18Kill(t, g, warden)
	markCommanderCard(t, g, me, warden)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, buried)
	passPriorityAroundTable(t, g)

	if !g.Exile.Contains(warden) {
		t.Fatal("the Greenwarden exiles itself from the graveyard")
	}
	if !me.Hand.Contains(buried) {
		t.Fatal("and because it did, the chosen card comes back")
	}
	answerCommanderReturn(t, g, me.ID, true)
	if !me.Command.Contains(warden) {
		t.Error("yes sends the exiled Greenwarden to the command zone")
	}
}

// TestWaterbendersRestorationSchedulesAnExiledCommander — the delayed
// return's payload is the cards that are in exile to be returned. A
// commander is exiled like any other (ADR 0115), so it is scheduled
// with the Probe; CR 903.9a then asks its owner about the command zone.
func TestWaterbendersRestorationSchedulesAnExiledCommander(t *testing.T) {
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

	if !g.Exile.Contains(commander) || !g.Exile.Contains(probe) {
		t.Fatal("the commander and the Probe are both exiled")
	}
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("delayed-return queue holds %d triggers, want 1", len(g.DelayedTriggers))
	}
	if got := g.DelayedTriggers[0].Cards; len(got) != 2 {
		t.Errorf("the delayed return carries %v, want the commander and the Probe", got)
	}
	answerCommanderReturn(t, g, me.ID, true)
	if !me.Command.Contains(commander) {
		t.Error("yes sends the exiled commander to the command zone")
	}
}

// markCommanderCard makes the named card in one of `owner`'s zones a
// commander, so a move to a hand or a library offers the command zone
// (CR 903.9b) and one into a graveyard or exile is offered it afterwards
// (CR 903.9a). Tests that need the prompt rather than a whole commander
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
