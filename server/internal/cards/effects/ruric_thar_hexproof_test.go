package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ruric_thar_hexproof_test.go — #3172: "Ruric Thar has hexproof as long
// as they haven't dealt combat damage yet", through the real combat
// damage step rather than a faked event.

func ruricTharOnTheBattlefield(g *game.Game, owner uuid.UUID) uuid.UUID {
	id := pushDiesCreatureForTest(g, owner, "Ruric Thar, Magecrusher", rfdRuricMagecrusherOracle,
		"Legendary Creature — Ogre Warrior", 7, 7)
	// The entry event is what runs the layer pass and stamps the static;
	// it also makes the creature summoning sick, so clear that to let it
	// attack as if it had been there since the turn began.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: id, OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield})
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	return id
}

func TestRuricTharMagecrusherLosesHexproofOnceItDealsCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ruric := ruricTharOnTheBattlefield(g, me.ID)

	if !effectiveAbilitiesContain(t, g, ruric, "hexproof") {
		t.Fatal("before it has dealt combat damage it has hexproof")
	}
	if creatureTargetableBy(g, opp.ID, ruric) {
		t.Fatal("an opponent can't target it while it has hexproof")
	}
	if !creatureTargetableBy(g, me.ID, ruric) {
		t.Fatal("its controller still can")
	}

	// Noncombat damage dealt by it does not count.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Actor: me.ID, Source: ruric, Target: opp.ID, Amount: 1})
	})
	if !effectiveAbilitiesContain(t, g, ruric, "hexproof") {
		t.Error("only COMBAT damage turns it off")
	}

	attackWith(t, g, opp.ID, ruric)
	if opp.Life != 40-7 {
		t.Fatalf("opponent life %d, want 33: Ruric Thar connected", opp.Life)
	}
	if effectiveAbilitiesContain(t, g, ruric, "hexproof") {
		t.Error("after dealing combat damage it has lost hexproof")
	}
	if !creatureTargetableBy(g, opp.ID, ruric) {
		t.Error("an opponent can target it now")
	}
}

// CR 400.7: a permanent that leaves is a new object when it comes back,
// and it has dealt nothing. The mark is cleared on the way out, so a
// replay of the card starts with the hexproof again.
func TestRuricTharMagecrusherMarkIsClearedWhenItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ruric := ruricTharOnTheBattlefield(g, me.ID)
	attackWith(t, g, opp.ID, ruric)
	if effectiveAbilitiesContain(t, g, ruric, "hexproof") {
		t.Fatal("setup: it connected")
	}
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(ruric); err != nil {
			t.Fatalf("bounce: %v", err)
		}
	})
	for _, c := range me.Hand.Cards {
		if c.InstanceID == ruric && c.DealtCombatDamage {
			t.Error("the card in hand still remembers dealing combat damage")
		}
	}
}
