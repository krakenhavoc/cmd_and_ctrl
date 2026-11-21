package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestUginEyeOfTheStormsCastTriggerExilesAColoredPermanent: the cast
// trigger fires as the spell is announced and resolves above Ugin.
func TestUginEyeOfTheStormsCastTriggerExilesAColoredPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.PushUginVictim(g, opp.ID, "A Green Bear", []string{"G"})
	ugin := effectstest.CastUginEyeOfTheStorms(t, g)

	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, victim)
	if effectstest.TriggerOnStack(g, ugin) == nil {
		t.Fatal("the cast trigger sits above Ugin on the stack")
	}
	if !g.Battlefield.Contains(victim) {
		t.Error("nothing is exiled until the trigger resolves")
	}
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.InExile(g, victim) {
		t.Error("the targeted coloured permanent is exiled")
	}
	if !g.Battlefield.Contains(ugin) {
		t.Error("Ugin resolves after his cast trigger")
	}
}

// TestUginEyeOfTheStormsCastTriggerSkipsColorlessPermanents: only a
// permanent that's one or more colors is a legal target.
func TestUginEyeOfTheStormsCastTriggerSkipsColorlessPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	colorless := effectstest.PushUginVictim(g, opp.ID, "A Colorless Golem", nil)
	colored := effectstest.PushUginVictim(g, opp.ID, "A Blue Drake", []string{"U"})
	effectstest.CastUginEyeOfTheStorms(t, g)

	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if effectstest.HasID(p.PickTargetCards, colorless) {
		t.Error("a colorless permanent is not a legal target")
	}
	if !effectstest.HasID(p.PickTargetCards, colored) {
		t.Error("a coloured permanent is a legal target")
	}
}

// TestUginEyeOfTheStormsTriggersOnAColorlessSpell is the second
// trigger, and its complement: a coloured spell does nothing.
func TestUginEyeOfTheStormsTriggersOnAColorlessSpell(t *testing.T) {
	for _, tc := range []struct {
		name    string
		colors  []string
		trigger bool
	}{
		{"a colorless spell", nil, true},
		{"a red spell", []string{"R"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			ugin := effectstest.PushCatalogWalker(g, me.ID, "Ugin, Eye of the Storms", effectstest.UginEyeOfTheStormsOracle, 7)
			victim := effectstest.PushUginVictim(g, opp.ID, "A Green Bear", []string{"G"})
			effectstest.ToMain(t, g)

			id := uuid.New()
			me.Hand.PushTop(game.Card{
				InstanceID: id, Name: "A Spell", TypeLine: "Sorcery",
				Colors: tc.colors, Owner: me.ID, Controller: me.ID,
			})
			if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			if !tc.trigger {
				if effectstest.LatestPickTarget(g, me.ID) != nil {
					t.Error("a coloured spell must not trigger Ugin")
				}
				return
			}
			effectstest.B04WaitForPick(t, g, me.ID)
			effectstest.PickCard(t, g, me.ID, victim)
			effectstest.PassPriorityAroundTable(t, g)
			if !effectstest.InExile(g, victim) {
				t.Error("the coloured permanent is exiled")
			}
			if effectstest.LoyaltyCount(g, ugin) != 7 {
				t.Errorf("the trigger costs no loyalty: %d", effectstest.LoyaltyCount(g, ugin))
			}
		})
	}
}

// TestUginEyeOfTheStormsPlusTwo: gain 3 life and draw a card.
func TestUginEyeOfTheStormsPlusTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ugin := effectstest.PushCatalogWalker(g, me.ID, "Ugin, Eye of the Storms", effectstest.UginEyeOfTheStormsOracle, 7)
	effectstest.ToMain(t, g)
	life, hand := me.Life, me.Hand.Size()

	if err := g.ActivateCatalogAbility(me.ID, ugin, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("+2: %v", err)
	}
	if effectstest.LoyaltyCount(g, ugin) != 9 {
		t.Errorf("loyalty = %d, want 9 — the cost is paid at announce", effectstest.LoyaltyCount(g, ugin))
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d -> %d, want +3", life, me.Life)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand delta %d, want +1", me.Hand.Size()-hand)
	}
}

// TestUginEyeOfTheStormsZeroAddsThreeColorless: the 0 is a loyalty
// ability, so it USES THE STACK (CR 605.1a) — the mana arrives when
// it resolves, not when it is activated.
func TestUginEyeOfTheStormsZeroAddsThreeColorless(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ugin := effectstest.PushCatalogWalker(g, me.ID, "Ugin, Eye of the Storms", effectstest.UginEyeOfTheStormsOracle, 7)
	effectstest.ToMain(t, g)

	if err := g.ActivateCatalogAbility(me.ID, ugin, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("0: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("pool = %+v, want empty until the ability resolves", me.ManaPool)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.ManaPool.Total() != 3 {
		t.Fatalf("pool = %+v, want {C}{C}{C}", me.ManaPool)
	}
	for _, m := range me.ManaPool.Units() {
		if m.Color != "C" {
			t.Errorf("pool = %+v, want colorless", me.ManaPool)
		}
	}
	if effectstest.LoyaltyCount(g, ugin) != 7 {
		t.Errorf("loyalty = %d, want 7 — a 0 costs nothing", effectstest.LoyaltyCount(g, ugin))
	}
}

// TestUginEyeOfTheStormsIsFull: #1230 closed the search-to-exile seam,
// so all three loyalty abilities are registered and nothing is
// declared as a caveat.
func TestUginEyeOfTheStormsIsFull(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.UginEyeOfTheStormsOracle)
	if !ok {
		t.Fatal("Ugin, Eye of the Storms is not registered")
	}
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Completeness = %v, Caveats = %+v, want Full and none", spec.Completeness, spec.Caveats)
	}
	if len(spec.Activated) != 3 {
		t.Errorf("%d loyalty abilities, want the +2, the 0 and the −11", len(spec.Activated))
	}
	if spec.StartingLoyalty != 0 {
		t.Error("starting loyalty is printed card data (ADR 0032) and must not be set here")
	}
}

// TestUginEyeOfTheStormsMinusElevenExilesAndGrantsFreeCast: the −11
// searches for any number of colorless nonland cards, exiles them
// (proving #1230's search-to-exile destination), shuffles, and grants
// a free cast on everything it found until end of turn.
func TestUginEyeOfTheStormsMinusElevenExilesAndGrantsFreeCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ugin := effectstest.PushCatalogWalker(g, me.ID, "Ugin, Eye of the Storms", effectstest.UginEyeOfTheStormsOracle, 11)
	effectstest.ToMain(t, g)

	colorless := effectstest.PushLibraryCardForTest(me, game.Card{
		Name: "Colorless Artifact", TypeLine: "Artifact",
		ManaCost: "{3}",
	})
	colored := effectstest.PushLibraryCardForTest(me, game.Card{
		Name: "Colored Sorcery", TypeLine: "Sorcery",
		ManaCost: "{1}{R}", Colors: []string{"R"},
	})
	colorlessLand := effectstest.PushLibraryCardForTest(me, game.Card{
		Name: "Colorless Land", TypeLine: "Land",
	})

	if err := g.ActivateCatalogAbility(me.ID, ugin, 2, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("−11: %v", err)
	}
	if effectstest.LoyaltyCount(g, ugin) != 0 {
		t.Errorf("loyalty = %d, want 0 — announced at 11, paid 11", effectstest.LoyaltyCount(g, ugin))
	}
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt after the −11")
	}
	if !effectstest.HasID(c.SearchCards, colorless) {
		t.Fatalf("search offers %v, want it to include the colorless nonland card %v", c.SearchCards, colorless)
	}
	if effectstest.HasID(c.SearchCards, colored) {
		t.Error("a colored card is never a legal match")
	}
	if effectstest.HasID(c.SearchCards, colorlessLand) {
		t.Error("a colorless LAND is never a legal match — nonland only")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{colorless}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !effectstest.InExile(g, colorless) {
		t.Error("the chosen colorless nonland card should be exiled")
	}
	if !me.Library.Contains(colored) {
		t.Error("a colored card is never a legal match")
	}
	if !me.Library.Contains(colorlessLand) {
		t.Error("a colorless LAND is never a legal match — nonland only")
	}
	if perm := g.CastPermissionOnCardByIDForEffect(colorless); perm == nil || perm.Cost != "{0}" {
		t.Errorf("the exiled card should carry a {0} cast permission, got %+v", perm)
	}
}
