package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// glittering_cats_test.go — Glittering Lion and Glittering Lynx (#1859,
// ADR 0106 amendment of 2026-10-07): a permanent that loses one of its
// own non-keyword abilities until end of turn (CR 613.1f, 611.2a).

const (
	glitteringLion = "549e6de7-56e9-4f5c-8c88-30e446bc53bb"
	glitteringLynx = "890ffe31-642f-46e3-9f09-c744351653b5"
)

func glitteringHit(g *game.Game, id uuid.UUID, n int) {
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), id, n) })
}

func TestGlitteringCats(t *testing.T) {
	cases := []struct {
		name, oracle string
		mana         []string
		power, hit   int
	}{
		{"Glittering Lion", glitteringLion, []string{"C", "C", "C"}, 2, 1},
		{"Glittering Lynx", glitteringLynx, []string{"C", "C"}, 1, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, bob := g.Seats[0], g.Seats[1]
			cat := apaPush(g, me.ID, me.ID, apaCreature(tc.name, tc.oracle, tc.power, tc.power))

			// The prevention works: damage is prevented and marks nothing.
			glitteringHit(g, cat, tc.hit)
			if got := damageMarkedOn(g, cat); got != 0 || !onBattlefield(g, cat) {
				t.Fatalf("damage marked = %d with the prevention on, want 0", got)
			}

			// An opponent pays the price out of their own pool and the
			// ability resolves: the damage now goes through.
			apaMana(bob, tc.mana...)
			apaActivate(t, g, bob, cat, 0, game.ActivateAbilityParams{})
			if rows := game.ActivatedAbilitiesForCard(*apaLive(g, cat)); len(rows) != 1 {
				t.Fatalf("the activated row is gone: %d rows", len(rows))
			}
			glitteringHit(g, cat, tc.hit)
			if tc.power == 1 { // the Lynx is a 1/1: the damage is lethal
				if got := damageMarkedOn(g, cat); onBattlefield(g, cat) && got != tc.hit {
					t.Fatalf("a Lynx without its prevention marked %d, want %d or death", got, tc.hit)
				}
				return
			}
			if got := damageMarkedOn(g, cat); got != tc.hit {
				t.Fatalf("damage marked = %d after the removal, want %d", got, tc.hit)
			}

			// The prevention is back next turn.
			advancePastCleanupForTest(t, g)
			if got := damageMarkedOn(g, cat); got != 0 {
				t.Fatalf("damage lingered past cleanup: %d", got)
			}
			glitteringHit(g, cat, tc.hit)
			if got := damageMarkedOn(g, cat); got != 0 {
				t.Errorf("damage marked = %d on the next turn, want the prevention back", got)
			}
		})
	}
}

// The Lynx's prevention is also back next turn: a second Lynx whose cost
// was paid on the previous turn is vulnerable only until cleanup.
func TestGlitteringLynxPreventionReturns(t *testing.T) {
	g := newCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	lynx := apaPush(g, me.ID, me.ID, apaCreature("Glittering Lynx", glitteringLynx, 1, 1))
	apaMana(bob, "C", "C")
	apaActivate(t, g, bob, lynx, 0, game.ActivateAbilityParams{})
	advancePastCleanupForTest(t, g)
	glitteringHit(g, lynx, 5)
	if !onBattlefield(g, lynx) {
		t.Errorf("the Lynx died to damage dealt after its removal had ended")
	}
}

// Only the Glittering cats' own prevention goes: another creature on the
// table keeps whatever it has, and the cat keeps the row that did it.
func TestGlitteringRemovalIsPinnedToTheCat(t *testing.T) {
	g := newCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	a := apaPush(g, me.ID, me.ID, apaCreature("Glittering Lion", glitteringLion, 2, 2))
	b := apaPush(g, me.ID, me.ID, apaCreature("Glittering Lion", glitteringLion, 2, 2))
	apaMana(bob, "C", "C", "C")
	apaActivate(t, g, bob, a, 0, game.ActivateAbilityParams{})
	glitteringHit(g, a, 1)
	glitteringHit(g, b, 1)
	if got := damageMarkedOn(g, a); got != 1 {
		t.Errorf("the activated Lion marked %d, want 1", got)
	}
	if got := damageMarkedOn(g, b); got != 0 {
		t.Errorf("the other Lion marked %d, want its prevention untouched", got)
	}
}
