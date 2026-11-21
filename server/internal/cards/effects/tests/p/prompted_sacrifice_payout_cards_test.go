package testsp

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
// The boards are exit_payout_cards_test.go's boards: a sacrificed
// commander (which, since ADR 0115, lands in the graveyard and is asked
// CR 903.9a afterwards), and a seat whose only creature leaves while
// somebody else is being asked.

// --- Rise of the Witch-king: the answer -----------------------------

// TestRiseOfTheWitchKingWaitsForEveryAnswer is the headline. Two seats
// are asked; the permanent comes back after the SECOND of them, not
// after the first and not while both questions are open.
func TestRiseOfTheWitchKingWaitsForEveryAnswer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	effectstest.PushVanillaCreature(g, me.ID, "My Spare", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.PushVanillaCreature(g, opp.ID, "Their Spare", 2, 2)
	rock := effectstest.B17GraveyardCard(me, "Sol Ring", "Artifact", "{1}")

	effectstest.CastCatalogSpell(t, g, "Rise of the Witch-king", "Sorcery",
		effectstest.B17RiseOfTheWitchKingOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.ReturnPickFor(g, me.ID) != nil {
		t.Fatal("the return was offered before anybody had chosen a creature")
	}
	effectstest.AnswerSacrifice(t, g, me.ID, mine)
	if effectstest.ReturnPickFor(g, me.ID) != nil {
		t.Error("the run waits for every asked seat, not only for the controller's own answer")
	}
	effectstest.AnswerSacrifice(t, g, opp.ID, theirs)
	if effectstest.ReturnPickFor(g, me.ID) == nil {
		t.Fatal("once every seat has answered, the controller chooses the permanent card")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, rock)
	if !g.Battlefield.Contains(rock) {
		t.Error("the chosen permanent comes back")
	}
}

// TestRiseOfTheWitchKingCantReturnTheCreatureYouSacrificed: the card is
// chosen on resolution, after the sacrifices (#2863), but "another
// permanent card" excludes the creature you just sacrificed — the
// official ruling (2023-06-16): "You cannot return the same permanent
// card that you sacrificed." With nothing else in your graveyard there
// is no offer at all.
func TestRiseOfTheWitchKingCantReturnTheCreatureYouSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	rock := effectstest.B17GraveyardCard(me, "Sol Ring", "Artifact", "{1}")

	effectstest.CastCatalogSpell(t, g, "Rise of the Witch-king", "Sorcery",
		effectstest.B17RiseOfTheWitchKingOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, mine)
	effectstest.AnswerSacrifice(t, g, opp.ID, theirs)

	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil || !effectstest.HasID(pick.ChooseCards, rock) {
		t.Fatalf("your other permanent card is on offer: %+v", pick)
	}
	if effectstest.HasID(pick.ChooseCards, mine) {
		t.Error("the creature you sacrificed is not \"another\" permanent card")
	}
	if effectstest.HasID(pick.ChooseCards, theirs) {
		t.Error("only YOUR graveyard is on offer")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, rock)
	if !g.Battlefield.Contains(rock) || g.Battlefield.Contains(mine) {
		t.Error("the Sol Ring returns; the sacrificed Bear stays in the graveyard")
	}

	// Only the sacrificed creature in the graveyard: nothing to offer.
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	bear := effectstest.PushVanillaCreature(g2, me2.ID, "Lone Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g2, "Rise of the Witch-king", "Sorcery",
		effectstest.B17RiseOfTheWitchKingOracle, nil)
	effectstest.PassPriorityAroundTable(t, g2)
	effectstest.AnswerSacrifice(t, g2, me2.ID, bear)
	if c := effectstest.ChooseCardsChoiceFor(g2, me2.ID); c != nil && !effectstest.IsSacrificePrompt(c) {
		t.Errorf("no other permanent card, so no offer: %+v", c.ChooseCards)
	}
	if g2.Battlefield.Contains(bear) {
		t.Error("the sacrificed Bear must not come back")
	}
}

// TestRiseOfTheWitchKingMayReturnNothing — "you may": the floor is
// zero, and declining returns nothing.
func TestRiseOfTheWitchKingMayReturnNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	rock := effectstest.B17GraveyardCard(me, "Sol Ring", "Artifact", "{1}")

	effectstest.CastCatalogSpell(t, g, "Rise of the Witch-king", "Sorcery",
		effectstest.B17RiseOfTheWitchKingOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, mine)
	effectstest.AnswerChooseCards(t, g, me.ID)
	if g.Battlefield.Contains(rock) || !me.Graveyard.Contains(rock) || !me.Graveyard.Contains(mine) {
		t.Error("declining returns nothing")
	}
}

// TestRiseOfTheWitchKingPaysForASacrificedCommander was the CR 903.9
// half: the commander used to stay ON the battlefield while its owner
// was asked, so a payout on the next line read "not sacrificed".
//
// Since ADR 0115 the commander is sacrificed into its owner's graveyard
// like any other creature, the return is offered at once, and the
// CR 903.9a answer that follows changes nothing.
func TestRiseOfTheWitchKingPaysForASacrificedCommander(t *testing.T) {
	for _, tc := range []struct {
		name        string
		commandZone bool
	}{
		{"then to the command zone: still sacrificed", true},
		{"left in the graveyard", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			mine := effectstest.PushVanillaCreature(g, me.ID, "My Commander", 2, 2)
			effectstest.MarkCommanderCard(t, g, me, mine)
			rock := effectstest.B17GraveyardCard(me, "Sol Ring", "Artifact", "{1}")

			effectstest.CastCatalogSpell(t, g, "Rise of the Witch-king", "Sorcery",
				effectstest.B17RiseOfTheWitchKingOracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AnswerSacrifice(t, g, me.ID, mine)

			if !me.Graveyard.Contains(mine) {
				t.Fatal("the sacrificed commander is in its owner's graveyard")
			}
			if effectstest.ChooseCardsChoiceFor(g, me.ID) == nil {
				t.Fatal("you sacrificed a creature this way, so you choose a permanent card to return")
			}
			effectstest.AnswerChooseCards(t, g, me.ID, rock)
			if !g.Battlefield.Contains(rock) {
				t.Fatal("the chosen permanent comes back")
			}
			effectstest.AnswerCommanderReturn(t, g, me.ID, tc.commandZone)
			if tc.commandZone && !me.Command.Contains(mine) {
				t.Fatal("yes puts the commander in the command zone")
			}
			if !tc.commandZone && !me.Graveyard.Contains(mine) {
				t.Fatal("no leaves the commander in the graveyard")
			}
			if !g.Battlefield.Contains(rock) {
				t.Error("the returned permanent stays")
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	spare := effectstest.PushVanillaCreature(g, me.ID, "My Spare", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.PushVanillaCreature(g, opp.ID, "Their Spare", 2, 2)
	rock := effectstest.B17GraveyardCard(me, "Sol Ring", "Artifact", "{1}")

	effectstest.CastCatalogSpell(t, g, "Rise of the Witch-king", "Sorcery",
		effectstest.B17RiseOfTheWitchKingOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SacrificeChoiceFor(g, me.ID) == nil {
		t.Fatal("the controller is asked")
	}

	// The controller's creatures die to something else, so their
	// prompt has no legal answer left and is withdrawn.
	effectstest.B18Kill(t, g, mine)
	effectstest.B18Kill(t, g, spare)
	if effectstest.SacrificeChoiceFor(g, me.ID) != nil {
		t.Fatal("a prompt with no legal answer left is withdrawn")
	}

	effectstest.AnswerSacrifice(t, g, opp.ID, theirs)
	if effectstest.ReturnPickFor(g, me.ID) != nil || g.Battlefield.Contains(rock) {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	swamp := effectstest.B12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	crypt := effectstest.B12Permanent(g, me.ID, "Blood Crypt", "Land — Swamp Mountain")
	basics := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Fetched Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Fetched Plains", TypeLine: "Basic Land — Plains"},
	)

	effectstest.CastCatalogSpell(t, g, "Planar Engineering", "Sorcery", effectstest.B33PlanarEngineeringOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

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
	effectstest.AnswerSacrifice(t, g, me.ID, swamp)
	if onBattlefield() != 0 {
		t.Error("the search waits for BOTH lands")
	}
	if c := effectstest.SacrificeChoiceFor(g, me.ID); c != nil {
		for _, id := range basics {
			if effectstest.HasID(c.SacrificeOptions, id) {
				t.Error("a fetched basic was offered as the second land to sacrifice")
			}
		}
	}
	effectstest.AnswerSacrifice(t, g, me.ID, crypt)
	effectstest.TakeSearchPrompts(t, g)
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
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	fodder := effectstest.PushVanillaCreature(g, opp.ID, "Fodder", 1, 1)
	effectstest.PushVanillaCreature(g, opp.ID, "Keeper", 1, 1)

	effectstest.CastCatalogSpell(t, g, "Cornered by Black Mages", "Sorcery", effectstest.B29CorneredByBlackMagesOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.FindBattlefieldByName(g, "Wizard") != uuid.Nil {
		t.Fatal("the Wizard arrived while the opponent was still choosing")
	}
	effectstest.AnswerSacrifice(t, g, opp.ID, fodder)
	if effectstest.FindBattlefieldByName(g, "Wizard") == uuid.Nil {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	effectstest.CastCatalogSpell(t, g, "Cornered by Black Mages", "Sorcery", effectstest.B29CorneredByBlackMagesOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.SacrificeChoiceFor(g, opp.ID) != nil {
		t.Fatal("an opponent with no creature is not asked (CR 701.21a)")
	}
	wizard := effectstest.FindBattlefieldByName(g, "Wizard")
	if wizard == uuid.Nil {
		t.Fatal("the Wizard still comes — the clause is not gated on the sacrifice")
	}
	if w, _ := effectstest.BattlefieldCard(g, wizard); w.Controller != me.ID {
		t.Error("the Wizard is the caster's")
	}
}
