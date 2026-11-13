package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// activation_restriction_zone_test.go — #5759: "activated abilities of
// artifacts / creatures can't be activated" names a card type and no
// zone, so it means permanents on the battlefield (CR 109.2). An
// artifact or creature CARD in a hand or a graveyard is not one, and
// its cycling, ninjutsu or graveyard ability still works.
//
// Each test asserts the battlefield refusal on the same board first,
// so a restriction that stopped applying at all would fail here too
// rather than pass every hand and graveyard assertion.

// actOffersActivationFrom reports whether the enumerator offers `seat`
// an activation of `source`, the bot's half of the agreement (#544).
func actOffersActivationFrom(g *game.Game, seat, source uuid.UUID) bool {
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Type == "activate_ability" && m.Source == source {
			return true
		}
	}
	return false
}

// actHandCantActivateReason is the view's cant_activate stamp for row
// `index` of a card in `p`'s hand: the gate's printed reason, or "".
func actHandCantActivateReason(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID, index int) string {
	t.Helper()
	var card game.Card
	found := false
	for _, c := range p.Hand.Cards {
		if c.InstanceID == id {
			card, found = c, true
			break
		}
	}
	if !found {
		t.Fatalf("%s is not in the hand", id)
	}
	abilities := game.ActivatedAbilitiesForCard(card)
	if index >= len(abilities) {
		t.Fatalf("%s has %d activated abilities, want row %d", card.Name, len(abilities), index)
	}
	var reason string
	g.WithWriteLock(func() {
		reason = g.CantActivateReasonLocked(p.ID, card, game.ZoneHand,
			game.ActivationAbility{Label: abilities[index].Label})
	})
	return reason
}

// Every catalog card built on SourcesCantActivate, each against a card
// in hand of a type its clause names. The battlefield half is a Cobalt
// Golem, an artifact creature with a non-mana ability, so all four
// clauses reach it whatever their mana exemption.
func TestSourcesCantActivateLeavesCyclingFromHandAlone(t *testing.T) {
	for _, tc := range []struct {
		restrictor, restrictorType, restrictorOracle string
		cycler, cyclerType, cyclerOracle, mana       string
	}{
		{"Cursed Totem", "Artifact", actCursedTotemOracle,
			"Magmakin Artillerist", "Creature — Elemental Pirate", magmakinArtilleristOracl, "{C}{R}"},
		{"Collector Ouphe", "Creature — Ouphe", actCollectorOupheOracle,
			"Runaway Boulder", "Artifact", lciQRUZBoulderOracle, "{C}{C}"},
		{"Clarion Conqueror", "Creature — Dragon", sb11ConquerorOracle,
			"Runaway Boulder", "Artifact", lciQRUZBoulderOracle, "{C}{C}"},
		{"Damping Matrix", "Artifact", mrdDampingMatrixOracle,
			"Magmakin Artillerist", "Creature — Elemental Pirate", magmakinArtilleristOracl, "{C}{R}"},
	} {
		t.Run(tc.restrictor, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			b21Push(g, me.ID, tc.restrictor, tc.restrictorType, tc.restrictorOracle, 0, 0)
			golem := b21Push(g, me.ID, "Cobalt Golem", "Artifact Creature — Golem", mrdCobaltGolemOracle, 2, 3)

			if err := g.ActivateCatalogAbility(me.ID, golem, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrCantActivate) {
				t.Fatalf("an artifact creature's ability on the battlefield: err = %v, want ErrCantActivate", err)
			}

			id := pushCatalogHandCard(me, tc.cycler, tc.cyclerType, tc.cyclerOracle)
			if err := g.AddManaForEffect(me.ID, uuid.Nil, tc.mana); err != nil {
				t.Fatalf("AddManaForEffect: %v", err)
			}
			if why := actHandCantActivateReason(t, g, me, id, 0); why != "" {
				t.Errorf("the view greys %s's cycling with %q, want no reason — a card in hand is not a permanent", tc.cycler, why)
			}
			if !actOffersActivationFrom(g, me.ID, id) {
				t.Errorf("the enumerator does not offer cycling %s from hand", tc.cycler)
			}
			if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
				t.Fatalf("cycling %s from hand: err = %v, want nil", tc.cycler, err)
			}
			if !me.Graveyard.Contains(id) {
				t.Errorf("the cycled %s is not in the graveyard", tc.cycler)
			}
		})
	}
}

// Cursed Totem and a creature card's graveyard ability: Reassembling
// Skeleton returns itself, and once it is back on the battlefield it
// is a creature the Totem reaches.
func TestCursedTotemLeavesACreatureCardsGraveyardAbilityAlone(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	b21Push(g, me.ID, "Cursed Totem", "Artifact", actCursedTotemOracle, 0, 0)
	krenko := actPushKrenko(g, me.ID)

	if err := g.ActivateCatalogAbility(me.ID, krenko, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrCantActivate) {
		t.Fatalf("Krenko's {T}: err = %v, want ErrCantActivate", err)
	}
	// activateFromGraveyard fails the test if the activation is refused.
	id, _ := activateFromGraveyard(t, g, "Reassembling Skeleton", "Creature — Skeleton Warrior",
		reassemblingSkeletonOracle, 1, 1, "{B}{B}", game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Skeleton did not return to the battlefield")
	}
}

// Cursed Totem and ninjutsu, which the issue names: the ninja is a
// creature card in hand, so the Totem does not stop it, while a
// creature on the battlefield is still covered.
func TestCursedTotemLeavesNinjutsuFromHandAlone(t *testing.T) {
	g := newCatalogGame(t)
	ninja, attacker, me, _ := ninjutsuBoard(t, g, "Ninja of the Deep Hours", ninjaOfTheDeepHoursOracle, "{U}{U}")
	b21Push(g, me.ID, "Cursed Totem", "Artifact", actCursedTotemOracle, 0, 0)
	krenko := actPushKrenko(g, me.ID)

	if err := g.ActivateCatalogAbility(me.ID, krenko, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrCantActivate) {
		t.Fatalf("Krenko's {T}: err = %v, want ErrCantActivate", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{attacker},
	}); err != nil {
		t.Fatalf("ninjutsu from hand under Cursed Totem: err = %v, want nil", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ninja) {
		t.Fatal("the ninja is not on the battlefield")
	}
}
