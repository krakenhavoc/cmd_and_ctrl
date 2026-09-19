package effects

import (
	"testing"

	"github.com/google/uuid"

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
//   - PAUSED. A commander card in the graveyard, so CR 903.9 asks its
//     owner and nothing is knowable until they answer. Both answers
//     are checked: the command zone pays nothing, exile pays.

// cantLeaveTheGraveyard is the "cards in graveyards can't be exiled"
// static, gated to one card.
func cantLeaveTheGraveyard(t *testing.T, g *game.Game, only uuid.UUID) {
	t.Helper()
	registerExitReplacementFrom(t, g, game.ZoneGraveyard, only, "it can't be exiled",
		func(ev *game.ReplacementEvent) { ev.Cancel() })
}

// tuckedInsteadOfExiled is the redirect: the card leaves the graveyard,
// but for the bottom of its owner's library rather than exile.
func tuckedInsteadOfExiled(t *testing.T, g *game.Game, only uuid.UUID, owner uuid.UUID) {
	t.Helper()
	registerExitReplacementFrom(t, g, game.ZoneGraveyard, only,
		"put it on the bottom of its owner's library instead",
		func(ev *game.ReplacementEvent) {
			ev.NewZone = game.ZoneLibrary
			ev.NewZoneOwner = owner
		})
}

// --- Scavenging Ooze ----------------------------------------------

// oozeEats puts a Scavenging Ooze on the battlefield under `me` and
// activates its ability at `prey`, resolving the ability.
func oozeEats(t *testing.T, g *game.Game, me *game.Player, ooze, prey uuid.UUID) {
	t.Helper()
	b10AddMana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, ooze, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: prey}},
	}); err != nil {
		t.Fatalf("activate Scavenging Ooze: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func pushOoze(g *game.Game, me *game.Player) uuid.UUID {
	return b12Push(g, me.ID, "Scavenging Ooze", "Creature — Ooze", b14ScavengingOozeOracle, 2, 2)
}

// TestScavengingOozeDoesNotGrowForAnExileTheWindowCancelled is the
// issue's own board. The meal never left the graveyard, so the Ooze ate
// nothing.
func TestScavengingOozeDoesNotGrowForAnExileTheWindowCancelled(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ooze := pushOoze(g, me)
	prey := batch01GraveyardCard(opp, "Dead Bear", "Creature — Bear")
	cantLeaveTheGraveyard(t, g, prey)
	before := me.Life

	oozeEats(t, g, me, ooze, prey)

	if !opp.Graveyard.Contains(prey) {
		t.Fatal("control: the cancelled exile leaves the card in its graveyard")
	}
	if got := b12Counter(t, g, ooze, "+1/+1"); got != 0 {
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ooze := pushOoze(g, me)
	prey := batch01GraveyardCard(opp, "Dead Bear", "Creature — Bear")
	tuckedInsteadOfExiled(t, g, prey, opp.ID)
	before := me.Life

	oozeEats(t, g, me, ooze, prey)

	if !opp.Library.Contains(prey) {
		t.Fatalf("control: the redirected card is in its owner's library, not %s", b12ZoneOf(g, prey))
	}
	if got := b12Counter(t, g, ooze, "+1/+1"); got != 0 {
		t.Errorf("+1/+1 counters = %d, want 0 — the card went to a library, not to exile", got)
	}
	if me.Life != before {
		t.Errorf("life %d → %d, want unchanged", before, me.Life)
	}
}

// TestScavengingOozeWaitsForTheCommandZoneAnswer — a commander card in
// the graveyard pauses the exile on CR 903.9, and the Ooze cannot know
// what it ate until the answer. Both answers are the test.
func TestScavengingOozeWaitsForTheCommandZoneAnswer(t *testing.T) {
	for _, tc := range []struct {
		name        string
		commandZone bool
		wantCounter int
	}{
		{"to the command zone: nothing was exiled", true, 0},
		{"to exile: the Ooze grows", false, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			ooze := pushOoze(g, me)
			prey := b17GraveyardCard(opp, "Their Commander", "Legendary Creature — Human", "{2}{G}")
			markCommanderCard(t, g, opp, prey)
			before := me.Life

			oozeEats(t, g, me, ooze, prey)

			if got := b12Counter(t, g, ooze, "+1/+1"); got != 0 {
				t.Fatalf("+1/+1 counters = %d while the CR 903.9 prompt is open, want 0 — "+
					"the clause waits for the answer", got)
			}
			if me.Life != before {
				t.Fatalf("life moved to %d while the prompt is open, want %d", me.Life, before)
			}

			if tc.commandZone {
				b36AcceptCommandZone(t, g, opp.ID)
				if !opp.Command.Contains(prey) {
					t.Fatal("accepting puts the commander card in the command zone")
				}
			} else {
				b21DeclineCommandZone(t, g, opp.ID)
				if !g.Exile.Contains(prey) {
					t.Fatal("declining exiles the commander card")
				}
			}

			if got := b12Counter(t, g, ooze, "+1/+1"); got != tc.wantCounter {
				t.Errorf("+1/+1 counters = %d, want %d", got, tc.wantCounter)
			}
			if want := before + tc.wantCounter; me.Life != want {
				t.Errorf("life %d → %d, want %d — the life rides the same clause as the counter",
					before, me.Life, want)
			}
		})
	}
}

// TestScavengingOozeStillGrowsForACreatureCardThatReachedExile is the
// control: the fix must not cost the card its printed behaviour.
func TestScavengingOozeStillGrowsForACreatureCardThatReachedExile(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ooze := pushOoze(g, me)
	prey := batch01GraveyardCard(opp, "Dead Bear", "Creature — Bear")
	before := me.Life

	oozeEats(t, g, me, ooze, prey)

	if !g.Exile.Contains(prey) {
		t.Fatal("the card is exiled")
	}
	if got := b12Counter(t, g, ooze, "+1/+1"); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	if me.Life != before+1 {
		t.Errorf("life %d → %d, want %d", before, me.Life, before+1)
	}
}

// --- Cling to Dust -------------------------------------------------

const clingToDustOracle = "9c44e556-4c7a-48bf-a108-c584f69c2cfa"

// TestClingToDustNeitherGainsNorDrawsWhenTheExileIsCancelled. Cling to
// Dust is the card where the gating question has teeth, because its
// clause has an "Otherwise". Both halves are branches of one
// conditional about the card that was exiled, so an exile that did not
// happen runs neither (ADR 0013 §5t).
func TestClingToDustNeitherGainsNorDrawsWhenTheExileIsCancelled(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	prey := batch01GraveyardCard(opp, "Dead Bear", "Creature — Bear")
	cantLeaveTheGraveyard(t, g, prey)
	lifeBefore, handBefore := me.Life, me.Hand.Size()

	castCatalogSpell(t, g, "Cling to Dust", "Instant", clingToDustOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: prey}})
	passPriorityAroundTable(t, g)

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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushLibraryCardForTest(me, game.Card{Name: "Top Card", TypeLine: "Instant"})
	prey := batch01GraveyardCard(opp, "Spent Bolt", "Instant")
	lifeBefore, handBefore := me.Life, me.Hand.Size()

	castCatalogSpell(t, g, "Cling to Dust", "Instant", clingToDustOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: prey}})
	passPriorityAroundTable(t, g)

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

// TestClingToDustWaitsForTheCommandZoneAnswer — the pause, on the card
// whose payout is the most visible.
func TestClingToDustWaitsForTheCommandZoneAnswer(t *testing.T) {
	for _, tc := range []struct {
		name        string
		commandZone bool
		wantLife    int
	}{
		{"to the command zone: no life", true, 0},
		{"to exile: three life", false, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			prey := b17GraveyardCard(opp, "Their Commander", "Legendary Creature — Human", "{2}{B}")
			markCommanderCard(t, g, opp, prey)
			lifeBefore, handBefore := me.Life, me.Hand.Size()

			castCatalogSpell(t, g, "Cling to Dust", "Instant", clingToDustOracle,
				[]game.TargetRef{{Kind: game.TargetCard, ID: prey}})
			passPriorityAroundTable(t, g)

			if me.Life != lifeBefore || me.Hand.Size() != handBefore {
				t.Fatalf("life %d and hand %d moved while the CR 903.9 prompt is open",
					me.Life, me.Hand.Size())
			}

			if tc.commandZone {
				b36AcceptCommandZone(t, g, opp.ID)
			} else {
				b21DeclineCommandZone(t, g, opp.ID)
			}

			if want := lifeBefore + tc.wantLife; me.Life != want {
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
	return b12Push(g, me.ID, "Deluge of the Dead", "Enchantment",
		invasionOfInnistradOracleID+"#1", 0, 0)
}

// delugeEats activates the enchantment's ability at `prey` and resolves
// it.
func delugeEats(t *testing.T, g *game.Game, me *game.Player, deluge, prey uuid.UUID) {
	t.Helper()
	b10AddMana(me, "B", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, deluge, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: prey}},
	}); err != nil {
		t.Fatalf("activate Deluge of the Dead: %v", err)
	}
	passPriorityAroundTable(t, g)
}

// TestDelugeOfTheDeadMakesNoZombieForAnExileTheWindowCancelled.
func TestDelugeOfTheDeadMakesNoZombieForAnExileTheWindowCancelled(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	deluge := pushDeluge(g, me)
	prey := batch01GraveyardCard(opp, "Dead Bear", "Creature — Bear")
	cantLeaveTheGraveyard(t, g, prey)

	delugeEats(t, g, me, deluge, prey)

	if !opp.Graveyard.Contains(prey) {
		t.Fatal("control: the cancelled exile leaves the card in its graveyard")
	}
	if n := countBattlefieldNamed(g, me.ID, "Zombie"); n != 0 {
		t.Errorf("%d Zombies, want 0 — nothing was exiled", n)
	}
}

// TestDelugeOfTheDeadWaitsForTheCommandZoneAnswer.
func TestDelugeOfTheDeadWaitsForTheCommandZoneAnswer(t *testing.T) {
	for _, tc := range []struct {
		name        string
		commandZone bool
		wantZombies int
	}{
		{"to the command zone: no Zombie", true, 0},
		{"to exile: one Zombie", false, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			deluge := pushDeluge(g, me)
			prey := b17GraveyardCard(opp, "Their Commander", "Legendary Creature — Human", "{2}{B}")
			markCommanderCard(t, g, opp, prey)

			delugeEats(t, g, me, deluge, prey)

			if n := countBattlefieldNamed(g, me.ID, "Zombie"); n != 0 {
				t.Fatalf("%d Zombies while the CR 903.9 prompt is open, want 0", n)
			}

			if tc.commandZone {
				b36AcceptCommandZone(t, g, opp.ID)
			} else {
				b21DeclineCommandZone(t, g, opp.ID)
			}

			if n := countBattlefieldNamed(g, me.ID, "Zombie"); n != tc.wantZombies {
				t.Errorf("%d Zombies, want %d", n, tc.wantZombies)
			}
		})
	}
}

// TestDelugeOfTheDeadUndoAcrossTheCommandZonePromptReplays is the undo
// contract the continuation signs, on a card: rewind into the open
// prompt, answer the OTHER way, and the board follows THAT answer.
func TestDelugeOfTheDeadUndoAcrossTheCommandZonePromptReplays(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	meID, oppID := me.ID, opp.ID
	deluge := pushDeluge(g, me)
	prey := b17GraveyardCard(opp, "Their Commander", "Legendary Creature — Human", "{2}{B}")
	markCommanderCard(t, g, opp, prey)

	delugeEats(t, g, me, deluge, prey)
	promptOpen := g.Clone()

	b36AcceptCommandZone(t, g, oppID)
	if n := countBattlefieldNamed(g, meID, "Zombie"); n != 0 {
		t.Fatalf("%d Zombies after the command zone, want 0", n)
	}

	g.WithWriteLock(func() { g.RestoreFrom(promptOpen) })
	if !g.Seats[1].Graveyard.Contains(prey) {
		t.Fatal("the rewind puts the commander card back in its graveyard")
	}

	b21DeclineCommandZone(t, g, oppID)
	if !g.Exile.Contains(prey) {
		t.Fatal("the replayed answer exiles the commander card")
	}
	if n := countBattlefieldNamed(g, meID, "Zombie"); n != 1 {
		t.Errorf("%d Zombies on the replay, want 1", n)
	}
}
