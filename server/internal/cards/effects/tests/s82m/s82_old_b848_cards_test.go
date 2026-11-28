package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b848_cards_test.go — Ice Age, batch 848 (#5636): each test plays the
// real catalog entry through the real engine paths.

const (
	b848FloodedWoodlands = "413172aa-b508-4e67-b68e-6a749d2e8130"
	b848Melting          = "599247f2-7f88-40ab-be08-e284e5f231d9"
	b848GhostlyFlame     = "6a461ca3-7923-4df4-9f15-4072545da117"
	b848Glaciers         = "a4a9c2bb-e6d0-4665-8b2f-e90c78784303"
	b848Mudslide         = "ac1d7dfe-1454-46a9-acf6-66ae86b6a6c5"
	b848Drought          = "ecba74bd-850e-4cbe-9b1c-b0d4d6f27d1c"
	b848Reclamation      = "ff942899-a70d-4fca-b125-9712aa635fff"
)

func TestS82OldB848CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b848FloodedWoodlands: "Flooded Woodlands", b848Melting: "Melting", b848GhostlyFlame: "Ghostly Flame",
		b848Glaciers: "Glaciers", b848Mudslide: "Mudslide", b848Drought: "Drought", b848Reclamation: "Reclamation",
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

func b848Tapped(g *game.Game, id uuid.UUID) bool {
	c := effectstest.FindBattlefieldCardByID(g, id)
	return c != nil && c.Tapped
}

func b848Colored(g *game.Game, owner *game.Player, name, manaCost string, power int) uuid.UUID {
	id := effectstest.B16Creature(g, owner.ID, name, "Creature — Test", power, 2, "C")
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardByID(g, id)
		c.ManaCost, c.Colors, c.SummonedThisTurn = manaCost, nil, false
	})
	g.BumpLayerVersionForTest()
	return id
}

func TestS82OldB848FloodedWoodlandsAndReclamationTaxTheirColorAttackersALandEach(t *testing.T) {
	for _, tc := range []struct{ oracle, name, cost, off string }{
		{b848FloodedWoodlands, "Flooded Woodlands", "{G}", "{R}"},
		{b848Reclamation, "Reclamation", "{B}", "{G}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			// The enchantment is the opponent's: it taxes my attackers too.
			effectstest.B12Push(g, opp.ID, tc.name, "Enchantment", tc.oracle, 0, 0)
			one := b848Colored(g, me, "One", tc.cost, 2)
			two := b848Colored(g, me, "Two", tc.cost, 2)
			plain := b848Colored(g, me, "Plain", tc.off, 2)
			var lands []uuid.UUID
			for i := 0; i < 3; i++ {
				lands = append(lands, effectstest.BlbB1Push(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0))
			}
			effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
			if err := g.DeclareAttacker(plain, opp.ID); err != nil {
				t.Fatalf("a creature of another color was taxed: %v", err)
			}
			if err := g.DeclareAttacker(one, opp.ID); err == nil {
				t.Fatal("a taxed creature attacked without sacrificing a land")
			}
			if err := g.DeclareAttackerWith(one, opp.ID, game.DeclareAttackersParams{Pay: game.CombatCostPayment{SacrificeIDs: []uuid.UUID{lands[0]}}}); err != nil {
				t.Fatalf("attacking with a land named: %v", err)
			}
			if g.Battlefield.Contains(lands[0]) {
				t.Error("the land was not sacrificed")
			}
			if err := g.DeclareAttackerWith(two, opp.ID, game.DeclareAttackersParams{Pay: game.CombatCostPayment{SacrificeIDs: []uuid.UUID{lands[1]}}}); err != nil {
				t.Fatalf("a second taxed attacker with its own land: %v", err)
			}
			if g.Battlefield.Contains(lands[1]) || !g.Battlefield.Contains(lands[2]) {
				t.Error("one land per taxed attacker was not what was sacrificed")
			}
		})
	}
}

func TestS82OldB848MeltingTakesSnowOffEveryLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.BlbB1Push(g, me.ID, "Snow-Covered Island", "Basic Snow Land — Island", "", 0, 0)
	theirs := effectstest.BlbB1Push(g, opp.ID, "Snow-Covered Swamp", "Basic Snow Land — Swamp", "", 0, 0)
	snowCreature := effectstest.B12Push(g, me.ID, "Snow Bear", "Snow Creature — Bear", "", 2, 2)
	effectstest.AdvanceToMain(t, g)
	for _, id := range []uuid.UUID{mine, theirs} {
		if !effectstest.LayeredCard(t, g, id).HasSupertype("snow") {
			t.Fatal("the snow land is not snow before Melting")
		}
	}
	effectstest.B12Push(g, me.ID, "Melting", "Enchantment", b848Melting, 0, 0)
	g.BumpLayerVersionForTest()
	for _, id := range []uuid.UUID{mine, theirs} {
		if effectstest.LayeredCard(t, g, id).HasSupertype("snow") {
			t.Error("a land is still snow under Melting")
		}
	}
	if !effectstest.LayeredCard(t, g, snowCreature).HasSupertype("snow") {
		t.Error("Melting took snow off a creature")
	}
}

func TestS82OldB848GhostlyFlameMakesBlackAndRedSourcesColorless(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := effectstest.PushTypedCard(g, me.ID, "Black Knight", "Creature — Knight", "{B}{B}")
	green := effectstest.PushTypedCard(g, me.ID, "Green Bear", "Creature — Bear", "{G}{G}")
	victim := effectstest.PushTypedCard(g, opp.ID, "White Knight", "Creature — Knight", "{W}{W}")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			switch c.InstanceID {
			case src:
				c.Colors = []string{"B"}
			case green:
				c.Colors = []string{"G"}
			case victim:
				c.Keywords = []string{"protection from black", "protection from green"}
				c.Toughness = 9
			}
		}
	})
	damage := func(source uuid.UUID) int {
		before := 0
		g.ReadSnapshot(func() { before = effectstest.FindBattlefieldCardByID(g, victim).DamageMarked })
		g.WithWriteLock(func() {
			g.RecomputeLayersIfStaleLocked()
			if err := g.DealDamageToCreatureForEffect(source, victim, 2); err != nil {
				t.Fatal(err)
			}
		})
		after := 0
		g.ReadSnapshot(func() { after = effectstest.FindBattlefieldCardByID(g, victim).DamageMarked })
		return after - before
	}
	if got := damage(src); got != 0 {
		t.Fatalf("protection from black let %d through before the Flame", got)
	}
	effectstest.B12Push(g, me.ID, "Ghostly Flame", "Enchantment", b848GhostlyFlame, 0, 0)
	g.BumpLayerVersionForTest()
	if got := damage(src); got != 2 {
		t.Errorf("under the Flame a black source dealt %d to a creature protected from black, want 2", got)
	}
	if got := damage(green); got != 0 {
		t.Errorf("a green source dealt %d through protection from green, want 0: the Flame only covers black and red", got)
	}
}

func TestS82OldB848GlaciersTurnsMountainsToPlainsAndTaxesTheUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	glaciers := effectstest.B12Push(g, me.ID, "Glaciers", "Enchantment", b848Glaciers, 0, 0)
	mine := effectstest.BlbB1Push(g, me.ID, "Mountain", "Basic Land — Mountain", "", 0, 0)
	theirs := effectstest.BlbB1Push(g, opp.ID, "Mountain", "Basic Land — Mountain", "", 0, 0)
	forest := effectstest.BlbB1Push(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	g.BumpLayerVersionForTest()
	for _, id := range []uuid.UUID{mine, theirs} {
		c := effectstest.LayeredCard(t, g, id)
		if !c.HasSubtype("Plains") || c.HasSubtype("Mountain") {
			t.Errorf("a Mountain is %v under Glaciers, want a Plains", c.Effective().Subtypes)
		}
	}
	if effectstest.LayeredCard(t, g, forest).HasSubtype("Plains") {
		t.Error("a Forest became a Plains")
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B06AddMana(me, "W", "U")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(glaciers) {
		t.Fatal("Glaciers was sacrificed after its tax was paid")
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(glaciers) {
		t.Error("Glaciers stayed after its tax was declined")
	}
	g.BumpLayerVersionForTest()
	if effectstest.LayeredCard(t, g, mine).HasSubtype("Plains") {
		t.Error("the Mountain stayed a Plains after Glaciers left")
	}
}

func TestS82OldB848MudslideHoldsNonflyersAndSellsUntapsAtUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Mudslide", "Enchantment", b848Mudslide, 0, 0)
	a := b848Colored(g, me, "Ground A", "{G}", 2)
	b := b848Colored(g, me, "Ground B", "{G}", 2)
	flyer := effectstest.Pr7bCreature(g, me.ID, "Flyer", 2, "{1}", []string{"U"}, "flying")
	for _, id := range []uuid.UUID{a, b, flyer} {
		effectstest.TapCard(g, id)
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	if !b848Tapped(g, a) || !b848Tapped(g, b) {
		t.Fatal("a creature without flying untapped in its untap step")
	}
	if b848Tapped(g, flyer) {
		t.Error("the flyer stayed tapped")
	}
	effectstest.PassPriorityAroundTable(t, g)
	// The upkeep trigger offers only my tapped nonflyers.
	ch := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if ch == nil {
		t.Fatal("no choice of creatures to untap")
	}
	if effectstest.ContainsID(ch.ChooseCards, flyer) {
		t.Error("an untapped flyer was offered")
	}
	// Choose both, pay {2}{2}.
	effectstest.AnswerChooseCards(t, g, me.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B06AddMana(me, "C", "C", "C", "C")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if b848Tapped(g, a) || b848Tapped(g, b) {
		t.Error("the chosen creatures stayed tapped after {2} for each was paid")
	}
}

func TestS82OldB848MudslideDeclinedPaymentUntapsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Mudslide", "Enchantment", b848Mudslide, 0, 0)
	a := b848Colored(g, me, "Ground A", "{G}", 2)
	effectstest.TapCard(g, a)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !b848Tapped(g, a) {
		t.Error("the creature untapped without payment")
	}
}

func TestS82OldB848DroughtMakesBlackSymbolsCostSwamps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	drought := effectstest.B12Push(g, me.ID, "Drought", "Enchantment", b848Drought, 0, 0)
	var swamps []uuid.UUID
	for i := 0; i < 3; i++ {
		swamps = append(swamps, effectstest.BlbB1Push(g, me.ID, "Swamp", "Basic Land — Swamp", "", 0, 0))
	}
	effectstest.AdvanceToMain(t, g)
	spell := game.Card{InstanceID: uuid.New(), Name: "Double Black", TypeLine: "Sorcery", ManaCost: "{B}{B}", Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(spell)
	effectstest.B06AddMana(me, "B", "B")
	if err := g.CastSpell(me.ID, spell.InstanceID, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("a spell with two black symbols was cast without sacrificing Swamps")
	}
	if err := g.CastSpell(me.ID, spell.InstanceID, game.CastSpellParams{Strict: true, SacrificeIDs: []uuid.UUID{swamps[0], swamps[1]}}); err != nil {
		t.Fatalf("casting with two Swamps sacrificed: %v", err)
	}
	if g.Battlefield.Contains(swamps[0]) || g.Battlefield.Contains(swamps[1]) || !g.Battlefield.Contains(swamps[2]) {
		t.Error("exactly two Swamps should have been sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	// A spell with no black symbol is free of it.
	green := game.Card{InstanceID: uuid.New(), Name: "Green Spell", TypeLine: "Sorcery", ManaCost: "{G}", Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(green)
	effectstest.B06AddMana(me, "G")
	if err := g.CastSpell(me.ID, green.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("a green spell was taxed: %v", err)
	}
	// The upkeep tax.
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(drought) {
		t.Error("Drought stayed after its tax was declined")
	}
}
