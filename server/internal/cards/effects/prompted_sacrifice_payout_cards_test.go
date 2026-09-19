package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// prompted_sacrifice_payout_cards_test.go — #1019 on the six catalog
// callers that had a clause after the prompt, resolved through the
// real priority loop.
//
// exit_payout_cards_test.go is the same file for #993's fire-and-forget
// exits; this is the half that arrives through a PROMPT. The two
// entry points queue a question per seat and return how many seats
// were asked, so a clause on the next line ran while every one of
// those questions was still on the table.
//
// What each card owed is not the same answer:
//
//   - Rise of the Witch-king owes the ANSWER. "If you sacrificed a
//     creature this way" was "was the controller handed a prompt",
//     which is a question about the question.
//   - Lich-Knights' Conquest owes the COUNT. "Return that many" was
//     how many prompts went up.
//   - Planar Engineering, Cornered by Black Mages, Will of the Abzan
//     and Priest of Forgotten Gods owe ORDER. Their second sentence is
//     about a player and happens either way (ADR 0013 §5m item 5), and
//     the run's continuation is what makes the sequence the printed
//     one.
//
// The boards are exit_payout_cards_test.go's boards: a CR 903.9 prompt
// held open by a commander, and a seat whose only creature leaves
// while somebody else is being asked.

// --- Rise of the Witch-king: the answer -----------------------------

// TestRiseOfTheWitchKingWaitsForEveryAnswer is the headline. Two seats
// are asked; the permanent comes back after the SECOND of them, not
// after the first and not while both questions are open.
func TestRiseOfTheWitchKingWaitsForEveryAnswer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	rock := b17GraveyardCard(me, "Sol Ring", "Artifact", "{1}")

	castCatalogSpell(t, g, "Rise of the Witch-king", "Sorcery",
		b17RiseOfTheWitchKingOracle, b16TargetCard(rock))
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(rock) {
		t.Fatal("the permanent came back before anybody had chosen a creature")
	}
	answerSacrifice(t, g, me.ID, mine)
	if g.Battlefield.Contains(rock) {
		t.Error("the run waits for every asked seat, not only for the controller's own answer")
	}
	answerSacrifice(t, g, opp.ID, theirs)
	if !g.Battlefield.Contains(rock) {
		t.Error("once every seat has answered, the permanent comes back")
	}
}

// TestRiseOfTheWitchKingWaitsForACommandZoneAnswer is the CR 903.9
// half, and the reason the answer routes through the sacrifice's own
// continuation: the commander is still ON the battlefield while its
// owner is asked, so a payout on the next line would read "not
// sacrificed" for a sacrifice that lands a beat later.
//
// Both answers pay. CR 701.17a's keyword action is the controller's
// move OFF the battlefield, and a replacement rewrites only where the
// permanent goes.
func TestRiseOfTheWitchKingWaitsForACommandZoneAnswer(t *testing.T) {
	for _, tc := range []struct {
		name        string
		commandZone bool
	}{
		{"to the command zone: still sacrificed", true},
		{"to the graveyard", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			mine := pushVanillaCreature(g, me.ID, "My Commander", 2, 2)
			markCommanderCard(t, g, me, mine)
			rock := b17GraveyardCard(me, "Sol Ring", "Artifact", "{1}")

			castCatalogSpell(t, g, "Rise of the Witch-king", "Sorcery",
				b17RiseOfTheWitchKingOracle, b16TargetCard(rock))
			passPriorityAroundTable(t, g)
			answerSacrifice(t, g, me.ID, mine)

			if g.Battlefield.Contains(rock) {
				t.Fatal("the permanent came back with the CR 903.9 prompt still open")
			}
			if !g.Battlefield.Contains(mine) {
				t.Fatal("a paused leg has moved nothing")
			}

			if tc.commandZone {
				b36AcceptCommandZone(t, g, me.ID)
				if !me.Command.Contains(mine) {
					t.Fatal("accepting puts the commander in the command zone")
				}
			} else {
				b21DeclineCommandZone(t, g, me.ID)
				if !me.Graveyard.Contains(mine) {
					t.Fatal("declining puts the commander in the graveyard")
				}
			}
			if !g.Battlefield.Contains(rock) {
				t.Error("you sacrificed a creature this way whichever zone it went to, " +
					"so the permanent comes back")
			}
		})
	}
}

// TestRiseOfTheWitchKingPaysNothingForASacrificeThatDidNotHappen is
// the other end, and it is the bug rather than the ordering: the
// controller is asked, their only creature leaves while an opponent is
// still choosing, the engine withdraws the prompt — and the clause is
// gated on the SACRIFICE, which never happened.
//
// Before #1019 the gate was "was a prompt queued", so this paid out.
func TestRiseOfTheWitchKingPaysNothingForASacrificeThatDidNotHappen(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	rock := b17GraveyardCard(me, "Sol Ring", "Artifact", "{1}")

	castCatalogSpell(t, g, "Rise of the Witch-king", "Sorcery",
		b17RiseOfTheWitchKingOracle, b16TargetCard(rock))
	passPriorityAroundTable(t, g)
	if sacrificeChoiceFor(g, me.ID) == nil {
		t.Fatal("the controller is asked")
	}

	// The controller's only creature dies to something else, so their
	// prompt has no legal answer left and is withdrawn.
	b18Kill(t, g, mine)
	if sacrificeChoiceFor(g, me.ID) != nil {
		t.Fatal("a prompt with no legal answer left is withdrawn")
	}

	answerSacrifice(t, g, opp.ID, theirs)
	if g.Battlefield.Contains(rock) {
		t.Error("nothing was sacrificed by the controller, so nothing comes back — " +
			"the gate is the sacrifice, not the question")
	}
}

// --- Planar Engineering: the order ----------------------------------

// TestPlanarEngineeringSearchesAfterTheLandsHaveGone. "Sacrifice two
// lands. Search …, then shuffle" — the search used to be queued
// alongside the sacrifice prompts, because the prompt had nothing to
// hang it on, so with a small library (which resolves the search
// synchronously) the basics arrived while the caster was still being
// asked which lands to sacrifice. Offered them, even: a basic that had
// already entered was on the second prompt's option list.
func TestPlanarEngineeringSearchesAfterTheLandsHaveGone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	swamp := b12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	crypt := b12Permanent(g, me.ID, "Blood Crypt", "Land — Swamp Mountain")
	basics := seedSearchLibrary(me,
		game.Card{Name: "Fetched Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Fetched Plains", TypeLine: "Basic Land — Plains"},
	)

	castCatalogSpell(t, g, "Planar Engineering", "Sorcery", b33PlanarEngineeringOracle, nil)
	passPriorityAroundTable(t, g)

	onBattlefield := func() int {
		n := 0
		for _, id := range basics {
			if g.Battlefield.Contains(id) {
				n++
			}
		}
		return n
	}
	if onBattlefield() != 0 {
		t.Fatal("the basics arrived beside the sacrifice prompts rather than after them")
	}
	answerSacrifice(t, g, me.ID, swamp)
	if onBattlefield() != 0 {
		t.Error("the search waits for BOTH lands")
	}
	if c := sacrificeChoiceFor(g, me.ID); c != nil {
		for _, id := range basics {
			if hasID(c.SacrificeOptions, id) {
				t.Error("a fetched basic was offered as the second land to sacrifice")
			}
		}
	}
	answerSacrifice(t, g, me.ID, crypt)
	if onBattlefield() != len(basics) {
		t.Errorf("%d of %d basics arrived once both lands had gone", onBattlefield(), len(basics))
	}
}

// --- Cornered by Black Mages: the order -----------------------------

// TestCorneredByBlackMagesMakesTheWizardAfterTheSacrifice. The token
// is not gated on the sacrifice — a target with no creature still gets
// you a Wizard, which the second half of this test pins — but it comes
// after it.
func TestCorneredByBlackMagesMakesTheWizardAfterTheSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	fodder := pushVanillaCreature(g, opp.ID, "Fodder", 1, 1)

	castCatalogSpell(t, g, "Cornered by Black Mages", "Sorcery", b29CorneredByBlackMagesOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)

	if findBattlefieldByName(g, "Wizard") != uuid.Nil {
		t.Fatal("the Wizard arrived while the opponent was still choosing")
	}
	answerSacrifice(t, g, opp.ID, fodder)
	if findBattlefieldByName(g, "Wizard") == uuid.Nil {
		t.Error("the Wizard comes once the sacrifice has landed")
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the opponent's creature is sacrificed")
	}
}

// TestCorneredByBlackMagesStillMakesTheWizardWithNothingToSacrifice is
// the control: the continuation runs even when nobody could be asked,
// so the clause that is not gated on the sacrifice still happens.
func TestCorneredByBlackMagesStillMakesTheWizardWithNothingToSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	castCatalogSpell(t, g, "Cornered by Black Mages", "Sorcery", b29CorneredByBlackMagesOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)

	if sacrificeChoiceFor(g, opp.ID) != nil {
		t.Fatal("an opponent with no creature is not asked (CR 701.21a)")
	}
	wizard := findBattlefieldByName(g, "Wizard")
	if wizard == uuid.Nil {
		t.Fatal("the Wizard still comes — the clause is not gated on the sacrifice")
	}
	if w, _ := battlefieldCard(g, wizard); w.Controller != me.ID {
		t.Error("the Wizard is the caster's")
	}
}
