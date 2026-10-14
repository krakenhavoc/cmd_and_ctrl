package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// raffine_a_shadowgrange_test.go — Shadowgrange Archfiend (#3212, slice raffine-a).

const shadowgrangeOracle = "8f3705ff-e8e5-48e7-ab68-df6849d78867"

func TestShadowgrangeArchfiendIsRegisteredAndWhole(t *testing.T) {
	spec, ok := Lookup(shadowgrangeOracle)
	if !ok {
		t.Fatal("not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 || spec.Madness != "{2}{B}" || spec.MadnessLife != 8 {
		t.Errorf("spec = %v / %v / madness %q life %d", spec.Completeness, spec.Caveats, spec.Madness, spec.MadnessLife)
	}
}

// Every opponent sacrifices a greatest-power creature of their own choice;
// the life is the greatest power among the creatures sacrificed.
func TestShadowgrangeArchfiendEachOpponentSacrificesAndYouGainTheGreatestPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a, b, c := g.Seats[1], g.Seats[2], g.Seats[3]
	aBig := pushVanillaCreature(g, a.ID, "A Giant", 5, 5)
	aSmall := pushVanillaCreature(g, a.ID, "A Bear", 2, 2)
	bOne := pushVanillaCreature(g, b.ID, "B Bear One", 2, 2)
	bTwo := pushVanillaCreature(g, b.ID, "B Bear Two", 2, 2)
	life := me.Life

	castAndResolveCreature(t, g, "Shadowgrange Archfiend", "Creature — Demon", shadowgrangeOracle)
	passPriorityAroundTable(t, g)

	// A's Giant is the only greatest-power creature: a forced pick,
	// collected unasked, that waits for B (ADR 0175, #3664).
	if pa := sacrificeChoiceFor(g, a.ID); pa != nil {
		t.Errorf("A has no choice to make and is asked anyway: %+v", pa)
	}
	if !g.Battlefield.Contains(aBig) {
		t.Error("A's Giant was sacrificed before B had chosen")
	}
	pb := sacrificeChoiceFor(g, b.ID)
	if pb == nil || len(sacrificeOptions(pb)) != 2 {
		t.Fatalf("B's two Bears tie for greatest and B chooses: %+v", pb)
	}
	if sacrificeChoiceFor(g, c.ID) != nil {
		t.Error("C has no creature and must not be asked")
	}
	answerSacrifice(t, g, b.ID, bTwo)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(aBig) || g.Battlefield.Contains(bTwo) || !g.Battlefield.Contains(aSmall) || !g.Battlefield.Contains(bOne) {
		t.Error("exactly the chosen creatures were sacrificed")
	}
	if me.Life != life+5 {
		t.Errorf("life %d → %d, want +5 (the greatest power sacrificed)", life, me.Life)
	}
}

// With no creature to sacrifice, nothing is gained.
func TestShadowgrangeArchfiendAgainstEmptyBoardsGainsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	castAndResolveCreature(t, g, "Shadowgrange Archfiend", "Creature — Demon", shadowgrangeOracle)
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d → %d, want unchanged", life, me.Life)
	}
}

// Madness is {2}{B} AND 8 life: looted into exile, accepted, the cast charges
// both, and a player with fewer than 8 life cannot make it.
func TestShadowgrangeArchfiendMadnessChargesEightLife(t *testing.T) {
	cast := func(t *testing.T, startLife int) (g *game.Game, me *game.Player, id uuid.UUID, err error) {
		g = newCatalogGame(t)
		me = g.Seats[g.Turn.ActiveSeat]
		g.WithWriteLock(func() { me.Hand.Cards = nil; me.Life = startLife })
		id = pushCatalogHandCard(me, "Shadowgrange Archfiend", "Creature — Demon", shadowgrangeOracle)
		spare := pushCatalogHandCard(me, "Spare", "Sorcery", "")
		castCatalogSpell(t, g, "Faithless Looting", "Sorcery", faithlessLootingOracl, nil)
		passPriorityAroundTable(t, g)
		answerChooseCards(t, g, me.ID, id, spare)
		passPriorityAroundTable(t, g)
		if !g.Exile.Contains(id) {
			t.Fatal("the Archfiend is not in exile")
		}
		offer := latestChoiceOfKind(g, game.PendingChoiceMayCast)
		if offer == nil {
			t.Fatal("no madness offer")
		}
		if e := g.ResolveMayCast(offer.ID, me.ID, true); e != nil {
			t.Fatalf("ResolveMayCast: %v", e)
		}
		if e := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{B}"); e != nil {
			t.Fatalf("mana: %v", e)
		}
		err = g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "exile", AlternativeCost: game.AltCostKeyMadness})
		return g, me, id, err
	}

	t.Run("enough life", func(t *testing.T) {
		g, me, id, err := cast(t, 20)
		if err != nil {
			t.Fatalf("the madness cast: %v", err)
		}
		if me.Life != 12 {
			t.Errorf("life = %d, want 12 (20 - 8)", me.Life)
		}
		if len(me.ManaPool) != 0 {
			t.Errorf("mana left in the pool: %d", len(me.ManaPool))
		}
		if g.Exile.Contains(id) {
			t.Error("the card is still in exile")
		}
	})
	t.Run("not enough life", func(t *testing.T) {
		_, me, _, err := cast(t, 7)
		if err == nil {
			t.Fatal("cast for the madness cost with 7 life")
		}
		if me.Life != 7 {
			t.Errorf("life = %d, want 7 untouched", me.Life)
		}
	})
}
