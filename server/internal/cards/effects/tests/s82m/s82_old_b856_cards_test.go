package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b856_cards_test.go — Fallen Empires, counterspells, token makers and
// enchantments, batch 856 (#5637, S82 OLD). Every test plays the real catalog entry
// through the engine.

const (
	oldB856Mage  = "8dee4577-0315-4c9f-ab1c-0c68e2ab6a76"
	oldB856Pit   = "ddc83b35-cf7c-495b-bd30-e409b622e299"
	oldB856Curse = "f6880f71-a05a-4752-ba33-ffbc450606bf"
)

func TestOldB856CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB856Mage: "Vodalian Mage", oldB856Pit: "Breeding Pit", oldB856Curse: "Thelon's Curse",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB856MageBoard(t *testing.T) (g *game.Game, victim, mage uuid.UUID, caster *game.Player) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	caster, them := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	mage = effectstest.B12Push(g, them.ID, "Vodalian Mage", "Creature — Merfolk Wizard", oldB856Mage, 1, 1)
	victim = effectstest.CastCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority to the responder: %v", err)
	}
	effectstest.B06AddMana(them, "U")
	if err := g.ActivateCatalogAbility(them.ID, mage, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(victim)}); err != nil {
		t.Fatalf("activate the Mage: %v", err)
	}
	effectstest.PassUntilTaxed(t, g, caster.ID)
	return g, victim, mage, caster
}

func TestOldB856VodalianMageCountersUnlessOneIsPaid(t *testing.T) {
	g, victim, mage, caster := oldB856MageBoard(t)
	prompt := effectstest.AnswerPayUnless(t, g, caster.ID, false)
	if prompt.PayCost != "{1}" {
		t.Errorf("the tax is %q, want {1}", prompt.PayCost)
	}
	if g.Stack.Contains(victim) {
		t.Error("a declined tax must counter the spell")
	}
	if !effectstest.IsTapped(g, mage) {
		t.Error("the Mage did not tap")
	}

	g, victim, _, caster = oldB856MageBoard(t)
	effectstest.FillPool(caster, 1)
	effectstest.AnswerPayUnless(t, g, caster.ID, true)
	if !g.Stack.Contains(victim) {
		t.Error("a paid tax countered the spell anyway")
	}
}

func TestOldB856BreedingPitSacrificesUnlessYouPayBB(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		pit := effectstest.B12Push(g, me.ID, "Breeding Pit", "Enchantment", oldB856Pit, 0, 0)
		effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
		effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
		if pay {
			effectstest.B06AddMana(me, "B", "B")
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.HasPayUnlessFor(g, me.ID) {
			t.Fatal("no pay prompt for the upkeep trigger")
		}
		prompt := effectstest.AnswerPayUnless(t, g, me.ID, pay)
		if prompt.PayCost != "{B}{B}" {
			t.Errorf("the payment is %q, want {B}{B}", prompt.PayCost)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := g.Battlefield.Contains(pit); got != pay {
			t.Errorf("paid %v: the Pit on the battlefield = %v", pay, got)
		}
	}
}

func TestOldB856BreedingPitMakesAThrullAtYourEndStepOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Breeding Pit", "Enchantment", oldB856Pit, 0, 0)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldByName(g, "Thrull"); n != 1 {
		t.Fatalf("%d Thrull tokens at my end step, want 1", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Thrull" && (c.Power != 0 || c.Toughness != 1 || len(c.Colors) != 1 || c.Colors[0] != "B" || !c.IsCreature()) {
			t.Errorf("the token is %d/%d colors %v, want a 0/1 black creature", c.Power, c.Toughness, c.Colors)
		}
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.AdvanceToEndStepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldByName(g, "Thrull"); n != 1 {
		t.Errorf("%d Thrull tokens after an opponent's end step, want still 1", n)
	}
}

func oldB856Tapped(g *game.Game, ids ...uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			for _, id := range ids {
				if g.Battlefield.Cards[i].InstanceID == id {
					g.Battlefield.Cards[i].Tapped = true
				}
			}
		}
	})
}

func oldB856IsTapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatalf("card %s not found", id)
	}
	return c.Tapped
}

func oldB856Choose(t *testing.T, g *game.Game, picks ...uuid.UUID) {
	t.Helper()
	for i := 0; i < 16; i++ {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceChooseCards {
				if err := g.ResolveChooseCards(c.ID, c.Chooser, picks); err != nil {
					t.Fatalf("ResolveChooseCards: %v", err)
				}
				return
			}
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("no choose_cards prompt")
}

func TestOldB856ThelonsCurseKeepsBlueCreaturesTappedUntilPaidFor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	effectstest.B12Push(g, me.ID, "Thelon's Curse", "Enchantment", oldB856Curse, 0, 0)
	blue := func(name string) uuid.UUID {
		id := effectstest.B12Creature(g, me.ID, name, "Creature — Merfolk", 1, 1)
		g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, id).Colors = []string{"U"} })
		return id
	}
	fishA, fishB, fishC := blue("Fish A"), blue("Fish B"), blue("Fish C")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	oldB856Tapped(g, fishA, fishB, fishC, bear)
	effectstest.B12ToMyNextUpkeep(t, g)
	for _, id := range []uuid.UUID{fishA, fishB, fishC} {
		if !oldB856IsTapped(t, g, id) {
			t.Fatal("a blue creature untapped in the untap step")
		}
	}
	if oldB856IsTapped(t, g, bear) {
		t.Fatal("a non-blue creature stayed tapped through the untap step")
	}
	effectstest.PassPriorityAroundTable(t, g)
	var offered []uuid.UUID
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards {
			offered = c.ChooseCards
		}
	}
	if len(offered) != 3 {
		t.Errorf("%d creatures offered, want the three blue ones", len(offered))
	}
	effectstest.FloatForTest(g, me, "UU")
	oldB856Choose(t, g, fishA, fishB)
	prompt := effectstest.AnswerPayUnless(t, g, me.ID, true)
	if prompt.PayCost != "{U}{U}" {
		t.Errorf("the payment is %q, want {U}{U}", prompt.PayCost)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if oldB856IsTapped(t, g, fishA) || oldB856IsTapped(t, g, fishB) {
		t.Error("a paid-for creature stayed tapped")
	}
	if !oldB856IsTapped(t, g, fishC) {
		t.Error("an unchosen creature untapped")
	}
}

func TestOldB856ThelonsCurseDecliningLeavesEverythingTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	effectstest.B12Push(g, me.ID, "Thelon's Curse", "Enchantment", oldB856Curse, 0, 0)
	fish := effectstest.B12Creature(g, me.ID, "Fish", "Creature — Merfolk", 1, 1)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, fish).Colors = []string{"U"} })
	oldB856Tapped(g, fish)
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FloatForTest(g, me, "U")
	oldB856Choose(t, g, fish)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !oldB856IsTapped(t, g, fish) {
		t.Error("the creature untapped without payment")
	}
}
