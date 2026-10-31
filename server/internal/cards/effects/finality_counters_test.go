package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// finality_counters_test.go — #4575: CR 122.1h is an engine rule. One or
// more finality counters on a permanent create a single replacement effect,
// "If this permanent would be put into a graveyard from the battlefield,
// exile it instead", so a counter that any effect puts on any permanent
// works, not only the ones Grim Repriser and Proctor of Potential carry.
// Grim Reaper, Lethal Legionnaire is the card that puts one on a permanent
// that does not print it.

const grimReaperOracle = "edf429a8-6a4d-4d2a-b0ca-c0c72a0698a3"

// addFinality puts one finality counter on a permanent the way an effect does.
func addFinality(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	var err error
	g.WithWriteLock(func() { err = g.AddCounterForEffect(id, game.CounterFinality, 1) })
	if err != nil {
		t.Fatalf("AddCounterForEffect: %v", err)
	}
	if got := allCountersOn(t, g, id)[game.CounterFinality]; got != 1 {
		t.Fatalf("finality counters = %d, want 1", got)
	}
}

// optionLabelled is the ID of the ordering prompt's option with this label.
// The engine's own replacements have no source card, so the label is the
// name of the effect.
func optionLabelled(t *testing.T, g *game.Game, order *game.PendingChoice, label string) game.ReplacementEffectID {
	t.Helper()
	for _, id := range order.ReplacementEffectIDs {
		var got string
		g.ReadSnapshot(func() { got, _ = g.ReplacementOptionMetaForEffect(id) })
		if got == label {
			return id
		}
	}
	t.Fatalf("no option on the prompt is labelled %q: %v", label, order.ReplacementEffectIDs)
	return 0
}

// leftTo reports where the last EventLTB for the card said it went.
func leftTo(g *game.Game, id uuid.UUID) (game.ZoneKind, bool) {
	var zone game.ZoneKind
	found := false
	for _, ev := range g.Events {
		if ev.Kind == game.EventLTB && ev.CardID == id {
			zone, found = ev.NewZone, true
		}
	}
	return zone, found
}

// A creature given a finality counter by another effect is exiled when it
// dies, whatever killed it: the rule is on the counter, not on a card.
func TestFinalityCounterFromAnotherEffectExilesAnyCreatureThatWouldDie(t *testing.T) {
	kills := map[string]func(g *game.Game, id uuid.UUID){
		"destroy":   func(g *game.Game, id uuid.UUID) { _ = g.DestroyPermanentForEffect(id) },
		"sacrifice": func(g *game.Game, id uuid.UUID) { _ = g.SacrificePermanentForEffect(id) },
	}
	for name, kill := range kills {
		g := newCatalogGame(t)
		opp := g.Seats[1]
		bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
		addFinality(t, g, bear)
		g.WithWriteLock(func() { kill(g, bear) })

		if !g.Exile.Contains(bear) || g.Battlefield.Contains(bear) || opp.Graveyard.Contains(bear) {
			t.Errorf("%s: the Bear is not in exile (graveyard %v)", name, opp.Graveyard.Contains(bear))
		}
		if zone, ok := leftTo(g, bear); !ok || zone != game.ZoneExile {
			t.Errorf("%s: left the battlefield to %q (seen %v), want exile: it never died (CR 700.4)", name, zone, ok)
		}
	}
}

// Without the counter the same creature dies as usual.
func TestACreatureWithoutAFinalityCounterStillGoesToTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })

	if !opp.Graveyard.Contains(bear) || g.Exile.Contains(bear) {
		t.Error("a Bear with no finality counter should go to the graveyard")
	}
	if zone, ok := leftTo(g, bear); !ok || zone != game.ZoneGraveyard {
		t.Errorf("left the battlefield to %q (seen %v), want the graveyard", zone, ok)
	}
}

// A permanent that is not a creature is covered too: CR 122.1h says
// "permanent".
func TestFinalityCounterAppliesToAnyPermanent(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	rock := pushCatalogPermanent(g, opp.ID, "Rock", "Artifact", "", false)
	addFinality(t, g, rock)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(rock) })
	if !g.Exile.Contains(rock) || opp.Graveyard.Contains(rock) {
		t.Error("an artifact with a finality counter is exiled when destroyed")
	}
}

// A wipe: each finality creature is exiled, the others die, and the wipe
// does not stop at the first exile.
func TestAWipeExilesEveryFinalityCreatureAndKillsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	withCounter := []uuid.UUID{
		b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2),
		b12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2),
		b12Creature(g, g.Seats[2].ID, "Bear C", "Creature — Bear", 2, 2),
	}
	for _, id := range withCounter {
		addFinality(t, g, id)
	}
	plain := []uuid.UUID{
		b12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1),
		b12Creature(g, opp.ID, "Ox", "Creature — Ox", 2, 4),
	}
	castCatalogSpell(t, g, "Wrath of God", "Sorcery", wrathOfGodOracle, nil)
	passPriorityAroundTable(t, g)

	for i, id := range withCounter {
		if !g.Exile.Contains(id) || g.Battlefield.Contains(id) {
			t.Errorf("finality creature %d is not in exile", i)
		}
		for _, s := range g.Seats {
			if s.Graveyard.Contains(id) {
				t.Errorf("finality creature %d reached a graveyard", i)
			}
		}
	}
	for i, id := range plain {
		if g.Exile.Contains(id) || g.Battlefield.Contains(id) {
			t.Errorf("plain creature %d should have died", i)
		}
		if !me.Graveyard.Contains(id) && !opp.Graveyard.Contains(id) {
			t.Errorf("plain creature %d is in no graveyard", i)
		}
	}
}

// A token with a finality counter is exiled, so it never dies; it then
// ceases to exist. The same token without one dies first.
func TestATokenWithAFinalityCounterIsExiledNotKilled(t *testing.T) {
	for _, counter := range []bool{true, false} {
		g := newCatalogGame(t)
		opp := g.Seats[1]
		tok := b06GreenCreature(t, g, opp.ID, "Elf Token")
		if tok == uuid.Nil {
			t.Fatal("the token did not enter")
		}
		if counter {
			addFinality(t, g, tok)
		}
		g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(tok) })

		if g.Battlefield.Contains(tok) {
			t.Errorf("counter=%v: the token is still on the battlefield", counter)
		}
		if counter && opp.Graveyard.Contains(tok) {
			t.Error("a token with a finality counter reached the graveyard")
		}
		want := game.ZoneGraveyard
		if counter {
			want = game.ZoneExile
		}
		if zone, ok := leftTo(g, tok); !ok || zone != want {
			t.Errorf("counter=%v: the token left to %q (seen %v), want %q", counter, zone, ok, want)
		}
	}
}

// Both a finality counter and Rest in Peace exile the Bear. The Bear's
// controller orders them (CR 616.1) and it is exiled either way; once one
// has sent it to exile the other has nothing left to replace (CR 616.1f), so
// no second prompt opens.
func TestFinalityCounterOrdersAgainstRestInPeaceAndTheBearIsExiledEitherWay(t *testing.T) {
	for _, finalityFirst := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		rip := pushCatalogPermanent(g, me.ID, "Rest in Peace", "Enchantment", restInPeaceOracle, false)
		bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
		addFinality(t, g, bear)
		order := destroyAndOrder(t, g, bear, opp.ID)

		first := optionFrom(t, g, order, rip)
		if finalityFirst {
			first = optionLabelled(t, g, order, "Finality counter")
		}
		answerOrder(t, g, order, opp.ID, first)

		if !g.Exile.Contains(bear) || g.Battlefield.Contains(bear) || opp.Graveyard.Contains(bear) {
			t.Errorf("finalityFirst=%v: the Bear is not in exile", finalityFirst)
		}
	}
}

// A redirect to somewhere other than exile ends the move for the counter
// too (CR 616.1f: only the replacements that would now apply are offered). A
// Darksteel Colossus with a finality counter, sacrificed, is its controller's
// choice: its own "shuffle it into your library instead" first leaves the
// counter nothing to replace; the counter first exiles it.
func TestFinalityCounterOrdersAgainstADarksteelColossusShuffle(t *testing.T) {
	for _, tc := range []struct {
		name      string
		colossus  bool
		toLibrary bool
	}{
		{"its own replacement first", true, true},
		{"the finality counter first", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			colossus := pushBattlefieldCardWithTimestamp(g, darksteelColossusCard(me.ID))
			addFinality(t, g, colossus)
			sacrificeNow(t, g, me.ID, colossus)
			prompt := latestChoiceOfKindFor(g, game.PendingChoiceReplacementOrder, me.ID)
			if prompt == nil || len(prompt.ReplacementEffectIDs) != 2 {
				t.Fatalf("want a CR 616 prompt between the Colossus and the counter, got %+v", prompt)
			}
			own := replacementIDForSource(t, g, prompt.ReplacementEffectIDs, colossus)
			counter := optionLabelled(t, g, prompt, "Finality counter")
			order := []game.ReplacementEffectID{counter, own}
			if tc.colossus {
				order = []game.ReplacementEffectID{own, counter}
			}
			if err := g.ResolveReplacementOrder(prompt.ID, me.ID, order); err != nil {
				t.Fatalf("ResolveReplacementOrder: %v", err)
			}
			passPriorityAroundTable(t, g)
			if got := me.Library.Contains(colossus); got != tc.toLibrary {
				t.Errorf("in library %v, want %v", got, tc.toLibrary)
			}
			if got := g.Exile.Contains(colossus); got == tc.toLibrary {
				t.Errorf("in exile %v, want %v", got, !tc.toLibrary)
			}
			if me.Graveyard.Contains(colossus) {
				t.Error("the Colossus reached the graveyard")
			}
		})
	}
}

// The finality counter against a regeneration shield. Finality first: the
// move is exiled, which ends the destruction, so the shield no longer
// applies and is not spent. Shield first: the destruction is replaced, the
// Bear stays tapped and the counter stays on it.
func TestFinalityCounterOrdersAgainstARegenerationShield(t *testing.T) {
	t.Run("finality first", func(t *testing.T) {
		g := newCatalogGame(t)
		opp := g.Seats[1]
		bear := shieldedBear(g, opp.ID)
		addFinality(t, g, bear)
		order := destroyAndOrder(t, g, bear, opp.ID)
		answerOrder(t, g, order, opp.ID, optionLabelled(t, g, order, "Finality counter"))
		if !g.Exile.Contains(bear) || g.Battlefield.Contains(bear) {
			t.Error("the Bear is not in exile: a shield cannot regenerate what was exiled instead")
		}
	})
	t.Run("shield first", func(t *testing.T) {
		g := newCatalogGame(t)
		opp := g.Seats[1]
		bear := shieldedBear(g, opp.ID)
		addFinality(t, g, bear)
		order := destroyAndOrder(t, g, bear, opp.ID)
		answerOrder(t, g, order, opp.ID, optionLabelled(t, g, order, "Regeneration shield"))
		if !g.Battlefield.Contains(bear) || g.Exile.Contains(bear) || opp.Graveyard.Contains(bear) {
			t.Fatal("the Bear did not stay on the battlefield")
		}
		if !isTapped(g, bear) {
			t.Error("a regenerated creature is tapped (CR 701.19a)")
		}
		if got := regenShieldsOn(t, g, bear); got != 0 {
			t.Errorf("shields = %d, want 0 (spent)", got)
		}
		if got := allCountersOn(t, g, bear)[game.CounterFinality]; got != 1 {
			t.Errorf("finality counters = %d, want 1: the Bear never moved", got)
		}
	})
}

// A restore while the order is open keeps it a real choice, and either
// answer on the restored table lands the same board as on one that never
// went through the disk.
func TestFinalityCounterOrderIsARestorePoint(t *testing.T) {
	for _, finalityFirst := range []bool{true, false} {
		g := newCatalogGame(t)
		opp := g.Seats[1]
		bear := shieldedBear(g, opp.ID)
		addFinality(t, g, bear)
		destroyAndOrder(t, g, bear, opp.ID)

		g = strictRestoreWithFrame(t, g, game.ReplacementResumeMove)
		opp = seatByID(g, opp.ID)
		order := b06ReplacementOrderFor(g, opp.ID)
		if order == nil || len(order.ReplacementEffectIDs) != 2 {
			t.Fatalf("the restored prompt lost an option: %+v", g.PendingChoices)
		}
		label := "Regeneration shield"
		if finalityFirst {
			label = "Finality counter"
		}
		answerOrder(t, g, order, opp.ID, optionLabelled(t, g, order, label))

		if finalityFirst {
			if !g.Exile.Contains(bear) || g.Battlefield.Contains(bear) {
				t.Error("finality first, restored: the Bear is not in exile")
			}
		} else if !g.Battlefield.Contains(bear) || g.Exile.Contains(bear) {
			t.Error("shield first, restored: the Bear did not stay")
		}
	}
}

// The rule is not state on the table: a restored table with a finality
// creature on it still exiles the creature when it dies.
func TestFinalityCounterRuleSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	addFinality(t, g, bear)

	g = restoreRoundTrip(t, g, true)
	opp = seatByID(g, opp.ID)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	if !g.Exile.Contains(bear) || opp.Graveyard.Contains(bear) {
		t.Error("the restored table sent a finality creature to the graveyard")
	}
}

// --- Grim Reaper, Lethal Legionnaire --------------------------------

// grimReaperAttack puts Grim Reaper on the battlefield with a creature card
// in the graveyard and attacks with it; the trigger is on the stack when it
// returns.
func grimReaperAttack(t *testing.T, g *game.Game) (reaper, dead uuid.UUID) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	reaper = pushCatalogPermanent(g, me.ID, "Grim Reaper, Lethal Legionnaire", "Legendary Creature — Human Villain", grimReaperOracle, false)
	dead = pushGraveyardPermanent(me, "Fallen Bear", "Creature — Bear", "{1}{G}")
	declareAttack(t, g, opp.ID, reaper)
	if n := triggersOnStackFrom(g, reaper); n != 1 {
		t.Fatalf("Grim Reaper's attack triggers on the stack = %d, want 1", n)
	}
	return reaper, dead
}

// The card's own text end to end: attack, pay {3}{B}, choose a creature
// card from your graveyard, and it returns tapped and attacking with a
// finality counter. When it dies it is exiled instead.
func TestGrimReaperReturnsACreatureTappedAndAttackingWithAFinalityCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	_, dead := grimReaperAttack(t, g)
	third := g.Seats[2]
	// A land card and an opponent's creature card are not legal targets.
	land := pushGraveyardPermanent(me, "Forest", "Basic Land — Forest", "")
	theirs := pushGraveyardPermanent(g.Seats[1], "Their Bear", "Creature — Bear", "{1}{G}")

	passPriorityAroundTable(t, g)
	floatMana(t, g, me, "{B}{B}{B}{B}")
	answerPayUnless(t, g, me.ID, true)
	pick := pickTargetPrompt(t, g)
	for _, id := range pick.PickTargetCards {
		if id == land || id == theirs {
			t.Errorf("target %v is not a creature card in my graveyard", id)
		}
	}
	answerPickTarget(t, g, dead)
	passPriorityAroundTable(t, g)

	// Three opponents: the controller names who it attacks.
	var refs []game.AttackTargetRef
	g.ReadSnapshot(func() { refs = g.AttackTargetsForEffect(me.ID) })
	idx := -1
	for i, r := range refs {
		if r.ID == third.ID {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatalf("the third seat is not an attack target: %v", refs)
	}
	answerOptionPick(t, g, me.ID, idx)

	c, ok := interceptorOnBattlefield(g, dead)
	if !ok {
		t.Fatal("the Bear did not return to the battlefield")
	}
	if me.Graveyard.Contains(dead) {
		t.Error("the Bear is still in the graveyard")
	}
	if !c.Tapped || c.AttackingTarget != third.ID || c.Controller != me.ID {
		t.Errorf("tapped=%v attacking=%v controller=%v, want tapped, attacking seat 2, mine", c.Tapped, c.AttackingTarget, c.Controller)
	}
	if got := c.Counters[game.CounterFinality]; got != 1 {
		t.Fatalf("finality counters = %d, want 1", got)
	}
	if len(poolColors(me)) != 0 {
		t.Errorf("pool = %v after paying {3}{B}", poolColors(me))
	}
	for _, ev := range g.Events {
		if ev.Kind == game.EventAttack && ev.CardID == dead {
			t.Error("the returned creature was announced as a declared attacker")
		}
	}

	// It would die: exiled instead, and its owner's graveyard stays empty of it.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(dead) })
	if !g.Exile.Contains(dead) || me.Graveyard.Contains(dead) || g.Battlefield.Contains(dead) {
		t.Error("the returned creature died instead of being exiled")
	}
}

// Declining the payment returns nothing, and the mana stays.
func TestGrimReaperDecliningThePaymentReturnsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	_, dead := grimReaperAttack(t, g)

	passPriorityAroundTable(t, g)
	floatMana(t, g, me, "{B}{B}{B}{B}")
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)

	if !me.Graveyard.Contains(dead) || g.Battlefield.Contains(dead) {
		t.Error("the creature card should still be in the graveyard")
	}
	if len(poolColors(me)) != 4 {
		t.Errorf("the mana was spent anyway: pool = %v", poolColors(me))
	}
}

// With no creature card in the graveyard the reflexive trigger has no
// target, so nothing returns.
func TestGrimReaperWithNoCreatureCardInTheGraveyardDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	reaper := pushCatalogPermanent(g, me.ID, "Grim Reaper, Lethal Legionnaire", "Legendary Creature — Human Villain", grimReaperOracle, false)
	land := pushGraveyardPermanent(me, "Forest", "Basic Land — Forest", "")
	declareAttack(t, g, opp.ID, reaper)
	passPriorityAroundTable(t, g)
	floatMana(t, g, me, "{B}{B}{B}{B}")
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)

	if !me.Graveyard.Contains(land) || g.Battlefield.Contains(land) {
		t.Error("a land card is not a creature card and must not return")
	}
}
