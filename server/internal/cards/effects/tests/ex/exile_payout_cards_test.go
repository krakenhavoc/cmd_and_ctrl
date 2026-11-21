package testsex

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exile_payout_cards_test.go — #911 on the three cards that print the
// clause: Cling to Dust, Scavenging Ooze and Deluge of the Dead.
//
// "Exile target card from a graveyard. If it was a creature card, …"
// The type is read before the move (it is a question about the past),
// and the CLAUSE waits for the move and is gated on it (CR 400.7, ADR
// 0013 §5t): the card that ARRIVED in exile is the one the effect
// exiled. All three used to pay out on the strength of "the call
// returned no error", which is true of an exile the CR 614 window
// cancelled, one it redirected, and one that has not been answered yet.
//
// Three boards, once each:
//
//   - CANCELLED. A graveyard static that says the card can't leave —
//     the Rest in Peace family's shape, written here as a test
//     replacement so the proof does not wait on a particular card
//     landing in the catalog.
//   - REDIRECTED. "If a card would leave a graveyard, put it on the
//     bottom of its owner's library instead" — the Leyline shape. The
//     card left; it did not reach exile.
//   - A COMMANDER card in the graveyard. Before ADR 0115 CR 903.9 asked
//     its owner before the exile and nothing was knowable until they
//     answered. Now the card is exiled like any other, the clause pays
//     at once, and CR 903.9a asks the owner afterwards; both answers
//     are checked, and neither changes the payout.

// cantLeaveTheGraveyard is the "cards in graveyards can't be exiled"
// static, gated to one card.
func cantLeaveTheGraveyard(t *testing.T, g *game.Game, only uuid.UUID) {
	t.Helper()
	effectstest.RegisterExitReplacementFrom(t, g, game.ZoneGraveyard, only, "it can't be exiled",
		func(ev *game.ReplacementEvent) { ev.Cancel() })
}

// tuckedInsteadOfExiled is the redirect: the card leaves the graveyard,
// but for the bottom of its owner's library rather than exile.
func tuckedInsteadOfExiled(t *testing.T, g *game.Game, only uuid.UUID, owner uuid.UUID) {
	t.Helper()
	effectstest.RegisterExitReplacementFrom(t, g, game.ZoneGraveyard, only,
		"put it on the bottom of its owner's library instead",
		func(ev *game.ReplacementEvent) {
			ev.NewZone = game.ZoneLibrary
			ev.NewZoneOwner = owner
		})
}

// TestScavengingOozeDoesNotGrowForAnExileTheWindowCancelled is the
// issue's own board. The meal never left the graveyard, so the Ooze ate
// nothing.
func TestScavengingOozeDoesNotGrowForAnExileTheWindowCancelled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ooze := effectstest.PushOoze(g, me)
	prey := effectstest.Batch01GraveyardCard(opp, "Dead Bear", "Creature — Bear")
	cantLeaveTheGraveyard(t, g, prey)
	before := me.Life

	effectstest.OozeEats(t, g, me, ooze, prey)

	if !opp.Graveyard.Contains(prey) {
		t.Fatal("control: the cancelled exile leaves the card in its graveyard")
	}
	if got := effectstest.B12Counter(t, g, ooze, "+1/+1"); got != 0 {
		t.Errorf("+1/+1 counters = %d, want 0 — nothing was exiled, so there is no \"it\" that "+
			"was a creature card", got)
	}
	if me.Life != before {
		t.Errorf("life %d → %d, want unchanged", before, me.Life)
	}
}

// TestScavengingOozeDoesNotGrowForACardSentSomewhereElse — the card
// left the graveyard, which is not the same as reaching exile
// (CR 400.7).
func TestScavengingOozeDoesNotGrowForACardSentSomewhereElse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ooze := effectstest.PushOoze(g, me)
	prey := effectstest.Batch01GraveyardCard(opp, "Dead Bear", "Creature — Bear")
	tuckedInsteadOfExiled(t, g, prey, opp.ID)
	before := me.Life

	effectstest.OozeEats(t, g, me, ooze, prey)

	if !opp.Library.Contains(prey) {
		t.Fatalf("control: the redirected card is in its owner's library, not %s", effectstest.B12ZoneOf(g, prey))
	}
	if got := effectstest.B12Counter(t, g, ooze, "+1/+1"); got != 0 {
		t.Errorf("+1/+1 counters = %d, want 0 — the card went to a library, not to exile", got)
	}
	if me.Life != before {
		t.Errorf("life %d → %d, want unchanged", before, me.Life)
	}
}

// TestScavengingOozeGrowsForAnExiledCommanderCard — a commander card
// in the graveyard is exiled like any other creature card (ADR 0115),
// so the Ooze grows at once; its owner is then asked about the command
// zone (CR 903.9a), and neither answer takes the growth back.
func TestScavengingOozeGrowsForAnExiledCommanderCard(t *testing.T) {
	for _, tc := range []struct {
		name        string
		commandZone bool
	}{
		{"then to the command zone", true},
		{"left in exile", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			ooze := effectstest.PushOoze(g, me)
			prey := effectstest.B17GraveyardCard(opp, "Their Commander", "Legendary Creature — Human", "{2}{G}")
			effectstest.MarkCommanderCard(t, g, opp, prey)
			before := me.Life

			effectstest.OozeEats(t, g, me, ooze, prey)

			if !g.Exile.Contains(prey) {
				t.Fatal("the commander card is exiled before its owner is asked")
			}
			if got := effectstest.B12Counter(t, g, ooze, "+1/+1"); got != 1 {
				t.Fatalf("+1/+1 counters = %d, want 1 — a creature card was exiled", got)
			}

			effectstest.AnswerCommanderReturn(t, g, opp.ID, tc.commandZone)
			if tc.commandZone && !opp.Command.Contains(prey) {
				t.Fatal("yes puts the commander card in the command zone")
			}
			if !tc.commandZone && !g.Exile.Contains(prey) {
				t.Fatal("no leaves the commander card in exile")
			}

			if got := effectstest.B12Counter(t, g, ooze, "+1/+1"); got != 1 {
				t.Errorf("+1/+1 counters = %d, want 1", got)
			}
			if want := before + 1; me.Life != want {
				t.Errorf("life %d → %d, want %d — the life rides the same clause as the counter",
					before, me.Life, want)
			}
		})
	}
}

// TestScavengingOozeStillGrowsForACreatureCardThatReachedExile is the
// control: the fix must not cost the card its printed behaviour.
func TestScavengingOozeStillGrowsForACreatureCardThatReachedExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ooze := effectstest.PushOoze(g, me)
	prey := effectstest.Batch01GraveyardCard(opp, "Dead Bear", "Creature — Bear")
	before := me.Life

	effectstest.OozeEats(t, g, me, ooze, prey)

	if !g.Exile.Contains(prey) {
		t.Fatal("the card is exiled")
	}
	if got := effectstest.B12Counter(t, g, ooze, "+1/+1"); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	if me.Life != before+1 {
		t.Errorf("life %d → %d, want %d", before, me.Life, before+1)
	}
}

// TestClingToDustNeitherGainsNorDrawsWhenTheExileIsCancelled. Cling to
// Dust is the card where the gating question has teeth, because its
// clause has an "Otherwise". Both halves are branches of one
// conditional about the card that was exiled, so an exile that did not
// happen runs neither (ADR 0013 §5t).
func TestClingToDustNeitherGainsNorDrawsWhenTheExileIsCancelled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	prey := effectstest.Batch01GraveyardCard(opp, "Dead Bear", "Creature — Bear")
	cantLeaveTheGraveyard(t, g, prey)
	lifeBefore, handBefore := me.Life, me.Hand.Size()

	effectstest.CastCatalogSpell(t, g, "Cling to Dust", "Instant", effectstest.ClingToDustOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: prey}})
	effectstest.PassPriorityAroundTable(t, g)

	if !opp.Graveyard.Contains(prey) {
		t.Fatal("control: the cancelled exile leaves the card in its graveyard")
	}
	if me.Life != lifeBefore {
		t.Errorf("life %d → %d, want unchanged — no card was exiled, so there is no \"it\" that "+
			"was a creature card", lifeBefore, me.Life)
	}
	if me.Hand.Size() != handBefore {
		t.Errorf("hand %d → %d, want unchanged — \"Otherwise, you draw a card\" is the other branch "+
			"of the same conditional, not a separate sentence", handBefore, me.Hand.Size())
	}
}

// TestClingToDustStillDrawsForANoncreatureCardThatReachedExile is the
// control for the Otherwise branch: a real exile of a noncreature card
// still draws.
func TestClingToDustStillDrawsForANoncreatureCardThatReachedExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Top Card", TypeLine: "Instant"})
	prey := effectstest.Batch01GraveyardCard(opp, "Spent Bolt", "Instant")
	lifeBefore, handBefore := me.Life, me.Hand.Size()

	effectstest.CastCatalogSpell(t, g, "Cling to Dust", "Instant", effectstest.ClingToDustOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: prey}})
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Exile.Contains(prey) {
		t.Fatal("the noncreature card is exiled")
	}
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d → %d, want +1", handBefore, me.Hand.Size())
	}
	if me.Life != lifeBefore {
		t.Errorf("life %d → %d, want unchanged for a noncreature card", lifeBefore, me.Life)
	}
}

// TestClingToDustGainsForAnExiledCommanderCard — the card whose payout
// is the most visible. The commander card is exiled (ADR 0115), Cling
// to Dust gains three life at once, and the CR 903.9a answer that
// follows changes nothing.
func TestClingToDustGainsForAnExiledCommanderCard(t *testing.T) {
	for _, tc := range []struct {
		name        string
		commandZone bool
	}{
		{"then to the command zone", true},
		{"left in exile", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			prey := effectstest.B17GraveyardCard(opp, "Their Commander", "Legendary Creature — Human", "{2}{B}")
			effectstest.MarkCommanderCard(t, g, opp, prey)
			lifeBefore, handBefore := me.Life, me.Hand.Size()

			effectstest.CastCatalogSpell(t, g, "Cling to Dust", "Instant", effectstest.ClingToDustOracle,
				[]game.TargetRef{{Kind: game.TargetCard, ID: prey}})
			effectstest.PassPriorityAroundTable(t, g)

			if want := lifeBefore + 3; me.Life != want {
				t.Fatalf("life %d → %d before the CR 903.9a answer, want %d", lifeBefore, me.Life, want)
			}
			effectstest.AnswerCommanderReturn(t, g, opp.ID, tc.commandZone)

			if want := lifeBefore + 3; me.Life != want {
				t.Errorf("life %d → %d, want %d", lifeBefore, me.Life, want)
			}
			if me.Hand.Size() != handBefore {
				t.Errorf("hand %d → %d, want unchanged either way — a commander card is a creature "+
					"card, so the Otherwise branch is not the one in play", handBefore, me.Hand.Size())
			}
		})
	}
}

// --- Deluge of the Dead --------------------------------------------

// pushDeluge puts the BACK FACE on the battlefield directly, under its
// own composite catalog key ("<oracle_id>#1"): the transform path is
// siege_transform_test.go's subject, and what is under test here is the
// activated ability's clause.
func pushDeluge(g *game.Game, me *game.Player) uuid.UUID {
	return effectstest.B12Push(g, me.ID, "Deluge of the Dead", "Enchantment",
		effects.InvasionOfInnistradOracleID+"#1", 0, 0)
}

// delugeEats activates the enchantment's ability at `prey` and resolves
// it.
func delugeEats(t *testing.T, g *game.Game, me *game.Player, deluge, prey uuid.UUID) {
	t.Helper()
	effectstest.B10AddMana(me, "B", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, deluge, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: prey}},
	}); err != nil {
		t.Fatalf("activate Deluge of the Dead: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// TestDelugeOfTheDeadMakesNoZombieForAnExileTheWindowCancelled.
func TestDelugeOfTheDeadMakesNoZombieForAnExileTheWindowCancelled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	deluge := pushDeluge(g, me)
	prey := effectstest.Batch01GraveyardCard(opp, "Dead Bear", "Creature — Bear")
	cantLeaveTheGraveyard(t, g, prey)

	delugeEats(t, g, me, deluge, prey)

	if !opp.Graveyard.Contains(prey) {
		t.Fatal("control: the cancelled exile leaves the card in its graveyard")
	}
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Zombie"); n != 0 {
		t.Errorf("%d Zombies, want 0 — nothing was exiled", n)
	}
}

// TestDelugeOfTheDeadMakesAZombieForAnExiledCommanderCard — the
// commander card is exiled (ADR 0115), the Zombie is made at once, and
// the CR 903.9a answer that follows does not take it back.
func TestDelugeOfTheDeadMakesAZombieForAnExiledCommanderCard(t *testing.T) {
	for _, tc := range []struct {
		name        string
		commandZone bool
	}{
		{"then to the command zone", true},
		{"left in exile", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			deluge := pushDeluge(g, me)
			prey := effectstest.B17GraveyardCard(opp, "Their Commander", "Legendary Creature — Human", "{2}{B}")
			effectstest.MarkCommanderCard(t, g, opp, prey)

			delugeEats(t, g, me, deluge, prey)

			if n := effectstest.CountBattlefieldNamed(g, me.ID, "Zombie"); n != 1 {
				t.Fatalf("%d Zombies before the CR 903.9a answer, want 1", n)
			}
			effectstest.AnswerCommanderReturn(t, g, opp.ID, tc.commandZone)
			if n := effectstest.CountBattlefieldNamed(g, me.ID, "Zombie"); n != 1 {
				t.Errorf("%d Zombies after the answer, want 1", n)
			}
		})
	}
}
