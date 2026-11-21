package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardslu "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/lu"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const maralenFaeAscendantOracle = "114e1d2c-cefc-4e2a-922e-3b2ceed1871d"

func TestMaralenFaeAscendantIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(maralenFaeAscendantOracle)
	if !ok || spec.Name != "Maralen, Fae Ascendant" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// maralenCastCreature casts a creature from `me`'s hand and answers the
// Maralen trigger its entry causes by targeting `opp`.
func maralenCastCreature(t *testing.T, g *game.Game, me, opp *game.Player, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	id := effectstest.RfdHandCard(me, name, typeLine, "", oracle, power, toughness)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

// maralenFreeCast casts a card exiled with Maralen for its {0} price.
func maralenFreeCast(g *game.Game, me *game.Player, card uuid.UUID) error {
	return g.CastSpell(me.ID, card, game.CastSpellParams{
		FromZone: "exile", AlternativeCost: cardslu.MaralenFaeAscendantFreeCastKey, Strict: true,
	})
}

// Maralen entering exiles the top two cards of the target opponent's
// library, linked to her. One of them is cast free when its mana value is
// at most the number of Elves and Faeries you control, counted as it is
// cast; once that use is spent, no other is.
func TestMaralenFaeAscendantCastsOneExiledSpellFreeUnderTheCount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	deeper := effectstest.TopOfLibrary(opp, "Third Card", "Sorcery", "{B}")
	two := effectstest.TopOfLibrary(opp, "Two Drop", "Sorcery", "{1}{B}")
	one := effectstest.TopOfLibrary(opp, "One Drop", "Sorcery", "{B}")

	maralen := maralenCastCreature(t, g, me, opp, "Maralen, Fae Ascendant", "Legendary Creature — Elf Faerie Noble", maralenFaeAscendantOracle, 4, 5)
	for _, id := range []uuid.UUID{one, two} {
		if got := effectstest.CardInExile(t, g, id).ExiledWith.ID; got != maralen {
			t.Errorf("exiled card linked to %s, want Maralen", got)
		}
	}
	if effectstest.E2Zone(g, deeper) != game.ZoneLibrary {
		t.Error("only the top two cards are exiled")
	}

	// One Elf or Faerie (Maralen): mana value 2 is too much.
	if err := maralenFreeCast(g, me, two); err == nil {
		t.Fatal("cast a mana value 2 spell with one Elf or Faerie")
	}
	// A second Elf: the count is two as the spell is cast.
	effectstest.B16Creature(g, me.ID, "Elf Ally", "Creature — Elf Warrior", 1, 1)
	// The pool is empty: the {0} price is the whole cost.
	if err := maralenFreeCast(g, me, two); err != nil {
		t.Fatalf("free cast with two Elves and Faeries: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.E2Zone(g, two) != game.ZoneGraveyard {
		t.Errorf("the free spell is in %q, want its owner's graveyard", effectstest.E2Zone(g, two))
	}
	if err := maralenFreeCast(g, me, one); err == nil {
		t.Error("a second free cast in the same turn")
	}
	if effectstest.EnumeratedFor(g, me.ID, one) {
		t.Error("the spent permission is still offered")
	}
}

// Cards exiled with Maralen on an earlier turn are not offered; a card
// exiled this turn is, until Maralen leaves the battlefield.
func TestMaralenFaeAscendantOnlyThisTurnAndOnlyWhileSheStays(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	old := effectstest.TopOfLibrary(opp, "Old Drop", "Sorcery", "{B}")
	_ = effectstest.TopOfLibrary(opp, "Old Top", "Sorcery", "{B}")
	maralen := maralenCastCreature(t, g, me, opp, "Maralen, Fae Ascendant", "Legendary Creature — Elf Faerie Noble", maralenFaeAscendantOracle, 4, 5)
	if !effectstest.EnumeratedFor(g, me.ID, old) {
		t.Fatal("non-vacuity: a card exiled this turn is offered")
	}

	effectstest.AdvanceToUpkeepOf(t, g, (seat+1)%len(g.Seats))
	effectstest.AdvanceToUpkeepOf(t, g, seat)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := maralenFreeCast(g, me, old); err == nil {
		t.Fatal("cast a card Maralen exiled on an earlier turn")
	}

	// The opponent's library may have been drawn from; seed a fresh top.
	_ = effectstest.TopOfLibrary(opp, "New Second", "Sorcery", "{B}")
	fresh := effectstest.TopOfLibrary(opp, "New Drop", "Sorcery", "{B}")
	maralenCastCreature(t, g, me, opp, "Elf Scout", "Creature — Elf Scout", "", 1, 1)
	if !effectstest.EnumeratedFor(g, me.ID, fresh) {
		t.Fatal("another Elf entering exiles cards castable this turn")
	}
	effectstest.B27Kill(g, maralen)
	effectstest.PassPriorityAroundTable(t, g)
	if err := maralenFreeCast(g, me, fresh); err == nil {
		t.Error("cast a card exiled with Maralen after she left")
	}
}
