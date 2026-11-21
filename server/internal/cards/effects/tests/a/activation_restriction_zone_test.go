package testsa

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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
		{"Cursed Totem", "Artifact", effectstest.ActCursedTotemOracle,
			"Magmakin Artillerist", "Creature — Elemental Pirate", effectstest.MagmakinArtilleristOracl, "{C}{R}"},
		{"Collector Ouphe", "Creature — Ouphe", effectstest.ActCollectorOupheOracle,
			"Runaway Boulder", "Artifact", effectstest.LciQRUZBoulderOracle, "{C}{C}"},
		{"Clarion Conqueror", "Creature — Dragon", effectstest.Sb11ConquerorOracle,
			"Runaway Boulder", "Artifact", effectstest.LciQRUZBoulderOracle, "{C}{C}"},
		{"Damping Matrix", "Artifact", effectstest.MrdDampingMatrixOracle,
			"Magmakin Artillerist", "Creature — Elemental Pirate", effectstest.MagmakinArtilleristOracl, "{C}{R}"},
	} {
		t.Run(tc.restrictor, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.B21Push(g, me.ID, tc.restrictor, tc.restrictorType, tc.restrictorOracle, 0, 0)
			golem := effectstest.B21Push(g, me.ID, "Cobalt Golem", "Artifact Creature — Golem", effectstest.MrdCobaltGolemOracle, 2, 3)

			if err := g.ActivateCatalogAbility(me.ID, golem, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrCantActivate) {
				t.Fatalf("an artifact creature's ability on the battlefield: err = %v, want ErrCantActivate", err)
			}

			id := effectstest.PushCatalogHandCard(me, tc.cycler, tc.cyclerType, tc.cyclerOracle)
			if err := g.AddManaForEffect(me.ID, uuid.Nil, tc.mana); err != nil {
				t.Fatalf("AddManaForEffect: %v", err)
			}
			if why := actHandCantActivateReason(t, g, me, id, 0); why != "" {
				t.Errorf("the view greys %s's cycling with %q, want no reason — a card in hand is not a permanent", tc.cycler, why)
			}
			if !effectstest.ActOffersActivationFrom(g, me.ID, id) {
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
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B21Push(g, me.ID, "Cursed Totem", "Artifact", effectstest.ActCursedTotemOracle, 0, 0)
	krenko := effectstest.ActPushKrenko(g, me.ID)

	if err := g.ActivateCatalogAbility(me.ID, krenko, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrCantActivate) {
		t.Fatalf("Krenko's {T}: err = %v, want ErrCantActivate", err)
	}
	// activateFromGraveyard fails the test if the activation is refused.
	id, _ := effectstest.ActivateFromGraveyard(t, g, "Reassembling Skeleton", "Creature — Skeleton Warrior",
		effectstest.ReassemblingSkeletonOracle, 1, 1, "{B}{B}", game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Skeleton did not return to the battlefield")
	}
}

// Cursed Totem and ninjutsu, which the issue names: the ninja is a
// creature card in hand, so the Totem does not stop it, while a
// creature on the battlefield is still covered.
func TestCursedTotemLeavesNinjutsuFromHandAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	ninja, attacker, me, _ := effectstest.NinjutsuBoard(t, g, "Ninja of the Deep Hours", effectstest.NinjaOfTheDeepHoursOracle, "{U}{U}")
	effectstest.B21Push(g, me.ID, "Cursed Totem", "Artifact", effectstest.ActCursedTotemOracle, 0, 0)
	krenko := effectstest.ActPushKrenko(g, me.ID)

	if err := g.ActivateCatalogAbility(me.ID, krenko, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrCantActivate) {
		t.Fatalf("Krenko's {T}: err = %v, want ErrCantActivate", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{attacker},
	}); err != nil {
		t.Fatalf("ninjutsu from hand under Cursed Totem: err = %v, want nil", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ninja) {
		t.Fatal("the ninja is not on the battlefield")
	}
}
