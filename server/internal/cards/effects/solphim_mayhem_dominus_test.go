package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// solphim_mayhem_dominus_test.go — #2257. A player blocked an
// attacking Solphim, Mayhem Dominus (a 5/4) with a deathtouch creature,
// saw "4" in the damage badge and watched Solphim survive. The table's
// bots activate Solphim's "{1}{R/P}{R/P}, Discard two cards: Put an
// indestructible counter on Solphim" after it attacks, so these pin
// both halves of what the player saw: without the counter the block
// kills it, and with the counter it lives with the damage still marked
// and says so on the wire.

// pushSolphim seats Solphim at its printed 5/4 under `owner`.
func pushSolphim(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b12Push(g, owner, "Solphim, Mayhem Dominus", "Legendary Creature — Phyrexian Horror", solphimOracle, 5, 4)
}

// pushImportedBlocker seats a creature carrying printed keywords the
// way the deck importer stamps them (Card.Keywords) — the road a
// player-uploaded deathtouch creature takes when it has no catalog
// entry.
func pushImportedBlocker(g *game.Game, owner uuid.UUID, power, toughness int, keywords ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Imported Blocker", TypeLine: "Creature — Test",
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
		Keywords: keywords,
	})
}

// solphimBlockedBy attacks `defender` with Solphim, blocks it with
// `blocker`, and walks the turn into the combat damage step, where the
// damage is dealt and the state-based actions are checked.
func solphimBlockedBy(t *testing.T, g *game.Game, solphim, blocker uuid.UUID, defender uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(solphim, defender); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, solphim); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	advanceTo(t, g, game.StepCombatDamage)
}

// putIndestructibleCounterOnSolphim activates the ability the way a
// seat does: {1}{R}{R} from the pool and two cards from hand.
func putIndestructibleCounterOnSolphim(t *testing.T, g *game.Game, me *game.Player, solphim uuid.UUID) {
	t.Helper()
	advanceToMain(t, g)
	discards := []uuid.UUID{handCard(me, "Pitched A", "Sorcery"), handCard(me, "Pitched B", "Instant")}
	floatForTest(g, me, "RRC")
	b16Activate(t, g, me.ID, solphim, 0, game.ActivateAbilityParams{DiscardIDs: discards})
	for _, id := range discards {
		if !me.Graveyard.Contains(id) {
			t.Fatalf("discarded card %s is not in the graveyard — the cost was not paid", id)
		}
	}
	if got := counterCount(g, solphim, "indestructible"); got != 1 {
		t.Fatalf("Solphim has %d indestructible counters, want 1", got)
	}
}

// The general case, both ways the reporter's block can kill: a
// deathtouch blocker of any power, and four damage on a creature with
// toughness four.
func TestSolphimDiesWhenBlockedForLethal(t *testing.T) {
	for _, tc := range []struct {
		name            string
		power, tough    int
		keywords        []string
		blockerSurvives bool
	}{
		{"one-power deathtouch blocker", 1, 1, []string{"deathtouch"}, false},
		{"four-power deathtouch blocker", 4, 4, []string{"deathtouch"}, false},
		{"four damage on four toughness, no deathtouch", 4, 6, nil, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			solphim := pushSolphim(g, me.ID)
			blocker := pushImportedBlocker(g, opp.ID, tc.power, tc.tough, tc.keywords...)

			solphimBlockedBy(t, g, solphim, blocker, opp.ID)

			if _, ok := battlefieldCard(g, solphim); ok {
				t.Error("Solphim is still on the battlefield — the block was lethal (CR 704.5g / 704.5h)")
			}
			if !me.Graveyard.Contains(solphim) {
				t.Error("Solphim should be in its owner's graveyard")
			}
			if _, ok := battlefieldCard(g, blocker); ok != tc.blockerSurvives {
				t.Errorf("blocker on battlefield = %v, want %v", ok, tc.blockerSurvives)
			}
		})
	}
}

// The reported board: an indestructible counter on Solphim, a
// four-power deathtouch blocker. Solphim takes the four damage and the
// deathtouch, keeps the damage marked, and is not destroyed (CR 702.12b) —
// and every viewer's snapshot carries the three facts that explain
// it: the damage, the counter and the keyword it grants.
func TestSolphimWithAnIndestructibleCounterSurvivesTheDeathtouchBlock(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	solphim := pushSolphim(g, me.ID)
	blocker := pushImportedBlocker(g, opp.ID, 4, 4, "deathtouch")
	putIndestructibleCounterOnSolphim(t, g, me, solphim)

	solphimBlockedBy(t, g, solphim, blocker, opp.ID)

	c, ok := battlefieldCard(g, solphim)
	if !ok {
		t.Fatal("Solphim died — an indestructible counter should keep it on the battlefield")
	}
	if c.DamageMarked != 4 {
		t.Errorf("Solphim has %d damage marked, want 4 — indestructible does not remove damage", c.DamageMarked)
	}
	if c.MarkedLethalByDeathtouch {
		t.Error("the deathtouch mark should be consumed by the SBA check that Solphim survived (CR 704.5h, #2319)")
	}
	if _, ok := battlefieldCard(g, blocker); ok {
		t.Error("the blocker took five damage and should have died")
	}

	var view *protocol.CardView
	for _, v := range protocol.ViewOfGame(g).Battlefield.Cards {
		if v.InstanceID == solphim.String() {
			view = &v
		}
	}
	if view == nil {
		t.Fatal("Solphim is missing from the battlefield view")
	}
	if view.DamageMarked != 4 || view.Toughness != 4 {
		t.Errorf("view damage %d on toughness %d, want 4 on 4", view.DamageMarked, view.Toughness)
	}
	if view.Counters["indestructible"] != 1 {
		t.Errorf("view counters = %v, want one indestructible counter", view.Counters)
	}
	if !slices.Contains(view.Abilities, "indestructible") {
		t.Errorf("view abilities = %v, want indestructible from the counter (CR 122.1b)", view.Abilities)
	}
}
