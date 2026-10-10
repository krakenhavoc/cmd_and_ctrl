package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// unstable_glyphbridge_test.go — #2719: Unstable Glyphbridge // Sandswirl
// Wanderglyph, Web of Inertia and Jace, Multiverse Architect's combat
// tax, the cards on the this-turn attack grant and the attacked-you
// cast ban.

const (
	sandswirlWanderglyphOracle = unstableGlyphbridgeOracleID + "#1"
	webOfInertiaOracle         = "a96d24f1-7670-4cb2-9971-5dd1b4c67957"
	jaceMultiverseOracle       = "3321c134-bf5b-4f63-8035-fca0bbdfa86b"
)

func TestGlyphbridgeCardsShipFull(t *testing.T) {
	for oracle, name := range map[string]string{
		unstableGlyphbridgeOracleID: "Unstable Glyphbridge",
		sandswirlWanderglyphOracle:  "Sandswirl Wanderglyph",
		webOfInertiaOracle:          "Web of Inertia",
		jaceMultiverseOracle:        "Jace, Multiverse Architect",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name {
			t.Errorf("%s: registered as %q (%v)", oracle, spec.Name, ok)
			continue
		}
		if spec.Completeness != CompletenessFull {
			t.Errorf("%s ships %q, want full", name, spec.Completeness)
		}
	}
}

// Cast, the Glyphbridge asks its caster about every board with a
// creature of power 2 or less, offers only those, and destroys every
// other creature.
func TestUnstableGlyphbridgeKeepsOneSmallCreaturePerPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	myBear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	myGiant := b12Creature(g, me.ID, "My Giant", "Creature — Giant", 5, 5)
	theirElf := b12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	theirOtherElf := b12Creature(g, opp.ID, "Their Other Elf", "Creature — Elf", 1, 1)
	thirdGiant := b12Creature(g, third.ID, "Third Giant", "Creature — Giant", 4, 4)

	castCatalogSpell(t, g, "Unstable Glyphbridge", "Artifact", unstableGlyphbridgeOracleID, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)

	mine := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	theirs := latestChoiceOfKindFor(g, game.PendingChoiceTheirPermanents, me.ID)
	if mine == nil || theirs == nil {
		t.Fatalf("the caster chooses on their own board and the opponent's: %+v", g.PendingChoices)
	}
	if mine.ChooseMin != 1 || mine.ChooseMax != 1 {
		t.Errorf("exactly one creature is chosen: %d..%d", mine.ChooseMin, mine.ChooseMax)
	}
	if err := g.ResolveOwnPermanents(mine.ID, me.ID, []uuid.UUID{myGiant}); err == nil {
		t.Error("a 5-power creature was accepted; the choice is power 2 or less")
	}
	for _, c := range append([]*game.PendingChoice(nil), g.PendingChoices...) {
		if c == nil {
			continue
		}
		if c.FromPlayer == third.ID {
			t.Fatal("a player with no creature of power 2 or less is not asked about")
		}
		switch c.Kind {
		case game.PendingChoiceOwnPermanents:
			if err := g.ResolveOwnPermanents(c.ID, me.ID, []uuid.UUID{myBear}); err != nil {
				t.Fatalf("ResolveOwnPermanents: %v", err)
			}
		case game.PendingChoiceTheirPermanents:
			if err := g.ResolveTheirPermanents(c.ID, me.ID, []uuid.UUID{theirElf}); err != nil {
				t.Fatalf("ResolveTheirPermanents: %v", err)
			}
		}
	}

	for _, id := range []uuid.UUID{myBear, theirElf} {
		if _, ok := battlefieldCard(g, id); !ok {
			t.Errorf("chosen creature %s was destroyed", id)
		}
	}
	for _, id := range []uuid.UUID{myGiant, theirOtherElf, thirdGiant} {
		if _, ok := battlefieldCard(g, id); ok {
			t.Errorf("unchosen creature %s survived", id)
		}
	}
	if len(battlefieldIDsNamed(g, "Unstable Glyphbridge")) != 1 {
		t.Error("the Glyphbridge is an artifact, not a creature; it stays")
	}
}

// "If you cast it": put onto the battlefield, nothing is chosen and
// nothing dies.
func TestUnstableGlyphbridgeNotCastDestroysNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	giant := b12Creature(g, me.ID, "Giant", "Creature — Giant", 5, 5)
	pushPermanentForTest(g, me.ID, "Unstable Glyphbridge", unstableGlyphbridgeOracleID, "Artifact")
	passPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 {
		t.Errorf("a Glyphbridge that was not cast asked something: %+v", g.PendingChoices)
	}
	if _, ok := battlefieldCard(g, giant); !ok {
		t.Error("a Glyphbridge that was not cast destroyed a creature")
	}
}

// Crafted, it is a 5/3 flying Sandswirl Wanderglyph.
func TestUnstableGlyphbridgeCraftsIntoAFlyingWanderglyph(t *testing.T) {
	g, me, _ := spendTable(t)
	bridge := pushCraftCard(g, me, transformRow(unstableGlyphbridgeOracleID, "Unstable Glyphbridge", "Artifact", "{3}{W}{W}",
		"Sandswirl Wanderglyph", "Artifact Creature — Golem", "5", "3", []string{"W"}))
	rock := pushGraveyardCardTyped(me, "Spent Rock", "Artifact")
	glyph := craftInto(t, g, me, bridge, 0, "{C}{C}{C}{W}{W}", "Sandswirl Wanderglyph", rock)
	if !effectiveAbilitiesContain(t, g, glyph.InstanceID, "flying") || effectivePower(t, g, glyph.InstanceID) != 5 {
		t.Error("Sandswirl Wanderglyph is not a 5/3 flier")
	}
}

// wanderglyphTable is a four-seat game with a Wanderglyph and a
// planeswalker under seat 0, a ready creature under seat 1, and a
// battle seat 0 controls.
func wanderglyphTable(t *testing.T) (g *game.Game, me, opp, third *game.Player, glyph, walker, battle, raider uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp, third = g.Seats[0], g.Seats[1], g.Seats[2]
	glyph = b12Push(g, me.ID, "Sandswirl Wanderglyph", "Artifact Creature — Golem", sandswirlWanderglyphOracle, 5, 3)
	walker = rfpbWalker(g, me.ID, "My Walker", "Legendary Planeswalker — Test", "", 3)
	battle = pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Battle", TypeLine: "Battle — Siege",
		Owner: me.ID, Controller: me.ID, ProtectorPlayerID: third.ID,
		Counters: map[string]int{game.CounterDefense: 3},
	})
	raider = b12Creature(g, opp.ID, "Raider", "Creature — Test", 2, 2)
	return
}

// oppCastsOnTheirTurn walks to seat 1's main phase and has them cast an
// instant, settling the stack.
func oppCastsOnTheirTurn(t *testing.T, g *game.Game, opp *game.Player) {
	t.Helper()
	advanceToStepOf(t, g, 1, game.StepPrecombatMain)
	id := handSpell(opp, "Opt", "Instant", "")
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("opponent CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
}

// An opponent who casts a spell on their own turn can't attack you or
// your planeswalkers this turn, but may still attack a battle you
// control and anyone else.
func TestSandswirlWanderglyphStopsASpellcastingOpponentAttackingYou(t *testing.T) {
	g, me, opp, third, _, walker, battle, raider := wanderglyphTable(t)
	oppCastsOnTheirTurn(t, g, opp)
	advanceToStepOf(t, g, 1, game.StepDeclareAttackers)

	for _, target := range []uuid.UUID{me.ID, walker} {
		var pe *game.PlayerCantAttackError
		if err := g.Clone().DeclareAttacker(raider, target); !errors.As(err, &pe) {
			t.Errorf("attack at %s: err = %v, want *PlayerCantAttackError", target, err)
		}
	}
	for _, target := range []uuid.UUID{battle, third.ID} {
		if err := g.Clone().DeclareAttacker(raider, target); err != nil {
			t.Errorf("attack at %s: %v (only you and your planeswalkers are protected)", target, err)
		}
	}

	// It lasts only this turn: on their next turn they may attack.
	advanceToStepOf(t, g, 2, game.StepPrecombatMain)
	advanceToStepOf(t, g, 1, game.StepDeclareAttackers)
	if err := g.Clone().DeclareAttacker(raider, me.ID); err != nil {
		t.Errorf("the next turn, with no spell cast: %v", err)
	}
}

// A spell cast on someone else's turn triggers nothing.
func TestSandswirlWanderglyphIgnoresSpellsCastOffTheirTurn(t *testing.T) {
	g, _, opp, _, _, _, _, _ := wanderglyphTable(t)
	advanceTo(t, g, game.StepPrecombatMain)
	id := handSpell(opp, "Opt", "Instant", "")
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("opponent CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	for _, s := range opp.Statics {
		if s.CantAttack.Protected != uuid.Nil {
			t.Fatalf("an opponent casting on my turn was restricted: %+v", s)
		}
	}
}

// An opponent who attacked you, or a planeswalker you controlled, can't
// cast spells for the rest of the turn; one who attacked a battle you
// control or another player can. The ban ends when the Wanderglyph
// leaves.
func TestSandswirlWanderglyphBansCastingAfterAttackingYou(t *testing.T) {
	for _, tc := range []struct {
		name   string
		target func(me, third *game.Player, walker, battle uuid.UUID) uuid.UUID
		banned bool
	}{
		{"you", func(me, _ *game.Player, _, _ uuid.UUID) uuid.UUID { return me.ID }, true},
		{"your planeswalker", func(_, _ *game.Player, walker, _ uuid.UUID) uuid.UUID { return walker }, true},
		{"your battle", func(_, _ *game.Player, _, battle uuid.UUID) uuid.UUID { return battle }, false},
		{"another player", func(_, third *game.Player, _, _ uuid.UUID) uuid.UUID { return third.ID }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, opp, third, glyph, walker, battle, raider := wanderglyphTable(t)
			advanceToStepOf(t, g, 1, game.StepDeclareAttackers)
			if err := g.DeclareAttacker(raider, tc.target(me, third, walker, battle)); err != nil {
				t.Fatalf("DeclareAttacker: %v", err)
			}
			lockInAttacks(t, g)
			spell := handSpell(opp, "Opt", "Instant", "")
			var err error
			g.ReadSnapshot(func() {
				c, _ := g.LookupCardForEffect(spell)
				err = g.CastGateLocked(opp.ID, c, game.ZoneHand, game.CastSpellParams{})
			})
			if banned := errors.Is(err, game.ErrCantCast); banned != tc.banned {
				t.Fatalf("cast gate after attacking %s: err = %v, want banned=%v", tc.name, err, tc.banned)
			}
			if !tc.banned {
				return
			}
			// The ruling: once the Wanderglyph has gone, they can cast.
			g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(glyph) })
			g.ReadSnapshot(func() {
				c, _ := g.LookupCardForEffect(spell)
				err = g.CastGateLocked(opp.ID, c, game.ZoneHand, game.CastSpellParams{})
			})
			if err != nil {
				t.Errorf("the Wanderglyph left; the cast gate still refuses: %v", err)
			}
		})
	}
}

// An attacked planeswalker that has since left still counts: it was
// yours when it was attacked.
func TestSandswirlWanderglyphRemembersAnAttackedPlaneswalker(t *testing.T) {
	g, _, opp, _, _, walker, _, raider := wanderglyphTable(t)
	advanceToStepOf(t, g, 1, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(raider, walker); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(walker) })
	spell := handSpell(opp, "Opt", "Instant", "")
	var err error
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(spell)
		err = g.CastGateLocked(opp.ID, c, game.ZoneHand, game.CastSpellParams{})
	})
	if !errors.Is(err, game.ErrCantCast) {
		t.Errorf("attacked a planeswalker that has since died: err = %v, want a cast ban", err)
	}
}

// --- Web of Inertia ----------------------------------------------------

// With an empty graveyard the opponent can't pay, so their creatures
// can't attack you this turn — but may attack your planeswalker (the
// ruling).
func TestWebOfInertiaRestrictsAnOpponentWithNothingToExile(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, me.ID, "Web of Inertia", webOfInertiaOracle, "Enchantment")
	walker := rfpbWalker(g, me.ID, "My Walker", "Legendary Planeswalker — Test", "", 3)
	raider := b12Creature(g, opp.ID, "Raider", "Creature — Test", 2, 2)

	advanceToStepOf(t, g, 1, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	advanceToStepOf(t, g, 1, game.StepDeclareAttackers)
	var pe *game.PlayerCantAttackError
	if err := g.Clone().DeclareAttacker(raider, me.ID); !errors.As(err, &pe) {
		t.Errorf("attack at the Web's controller: err = %v, want *PlayerCantAttackError", err)
	}
	if err := g.Clone().DeclareAttacker(raider, walker); err != nil {
		t.Errorf("attack at their planeswalker: %v", err)
	}
}

// Exiling a card from their graveyard keeps the attack open; declining
// closes it.
func TestWebOfInertiaAsksTheOpponentToExileACard(t *testing.T) {
	for _, exile := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		pushPermanentForTest(g, me.ID, "Web of Inertia", webOfInertiaOracle, "Enchantment")
		raider := b12Creature(g, opp.ID, "Raider", "Creature — Test", 2, 2)
		dead := pushGraveyardCardTyped(opp, "Dead Card", "Sorcery")

		advanceToStepOf(t, g, 1, game.StepBeginCombat)
		passPriorityAroundTable(t, g)
		if exile {
			answerChooseCards(t, g, opp.ID, dead)
		} else {
			answerChooseCards(t, g, opp.ID)
		}
		if got := g.Exile.Contains(dead); got != exile {
			t.Errorf("exile=%v: card in exile = %v", exile, got)
		}
		advanceToStepOf(t, g, 1, game.StepDeclareAttackers)
		err := g.Clone().DeclareAttacker(raider, me.ID)
		if exile && err != nil {
			t.Errorf("they exiled a card; attack refused: %v", err)
		}
		if !exile && err == nil {
			t.Error("they declined; the attack should be refused")
		}
	}
}

// --- Jace, Multiverse Architect's combat tax ----------------------------

// Declining the {2} stops attacks on Jaces only; paying it stops
// nothing.
func TestJaceMultiverseArchitectTaxesAttacksOnJaces(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		jace := rfpbWalker(g, me.ID, "Jace, Multiverse Architect", "Legendary Planeswalker — Jace", jaceMultiverseOracle, 4)
		other := rfpbWalker(g, me.ID, "Other Walker", "Legendary Planeswalker — Test", "", 3)
		raider := b12Creature(g, opp.ID, "Raider", "Creature — Test", 2, 2)

		advanceToStepOf(t, g, 1, game.StepBeginCombat)
		passPriorityAroundTable(t, g)
		choice := latestChoiceOfKindFor(g, game.PendingChoicePayUnless, opp.ID)
		if choice == nil {
			t.Fatalf("the opponent is not asked to pay {2}: %+v", g.PendingChoices)
		}
		if pay {
			floatMana(t, g, opp, "{C}{C}")
		}
		if err := g.ResolvePayUnless(choice.ID, opp.ID, pay); err != nil {
			t.Fatalf("ResolvePayUnless: %v", err)
		}
		advanceToStepOf(t, g, 1, game.StepDeclareAttackers)

		err := g.Clone().DeclareAttacker(raider, jace)
		if pay && err != nil {
			t.Errorf("paid {2}; attack on Jace refused: %v", err)
		}
		if !pay {
			var pe *game.PlayerCantAttackError
			if !errors.As(err, &pe) {
				t.Errorf("declined; attack on Jace: err = %v, want *PlayerCantAttackError", err)
			}
		}
		for _, target := range []uuid.UUID{me.ID, other} {
			if err := g.Clone().DeclareAttacker(raider, target); err != nil {
				t.Errorf("pay=%v: attack at %s refused: %v (only Jaces are protected)", pay, target, err)
			}
		}
	}
}
