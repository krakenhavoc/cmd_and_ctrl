package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 set OLD, batch 584, Saviors of Kamigawa: Channel (#5601). Each test
// plays the real catalog entry through the engine.

const (
	oldB584JiwariOracle    = "03199a9e-3a52-42af-8027-50b455a054d4"
	oldB584WingsOracle     = "5fa542e5-d3a7-45d3-ac1d-32546fbfe784"
	oldB584ChillOracle     = "72e61b73-980c-4ee4-b529-a367897620ec"
	oldB584StalkerOracle   = "90a4c5ed-d268-4216-9b55-0a38406efd74"
	oldB584NourisherOracle = "acaed013-2e74-47e3-a687-084388b37a7b"
	oldB584WarderOracle    = "cf1a4b18-62fa-46f3-baeb-c0f699c9cdde"
	oldB584LightOracle     = "d43f813b-3607-4bc3-b996-3dbad6977bd8"
)

func TestOldB584EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		oldB584JiwariOracle:    "Jiwari, the Earth Aflame",
		oldB584WingsOracle:     "Shinen of Flight's Wings",
		oldB584ChillOracle:     "Shinen of Fear's Chill",
		oldB584StalkerOracle:   "Ghost-Lit Stalker",
		oldB584NourisherOracle: "Ghost-Lit Nourisher",
		oldB584WarderOracle:    "Ghost-Lit Warder",
		oldB584LightOracle:     "Shinen of Stars' Light",
	}
	for oracle, name := range want {
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

func oldB584Flyer(g *game.Game, owner uuid.UUID, toughness int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird", Power: 1, Toughness: toughness,
		Keywords: []string{"flying"}, Owner: owner, Controller: owner,
	})
}

func TestOldB584JiwariTapsToDealXDamageToAGroundCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jiwari := effectstest.B12Push(g, me.ID, "Jiwari, the Earth Aflame", "Legendary Creature — Spirit", oldB584JiwariOracle, 3, 3)
	bird := oldB584Flyer(g, opp.ID, 3)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.ApaMana(me, "R", "R", "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, jiwari, 0, game.ActivateAbilityParams{Strict: true, XValue: 3, Targets: effectstest.CardRefs(bird)}); err == nil {
		t.Fatal("targeted a creature with flying")
	}
	effectstest.ApaActivate(t, g, me, jiwari, 0, game.ActivateAbilityParams{XValue: 3, Targets: effectstest.CardRefs(bear)})
	if g.Battlefield.Contains(bear) {
		t.Error("the 2/2 survived X = 3")
	}
	if !g.Battlefield.Contains(bird) {
		t.Error("the flyer died")
	}
	if !effectstest.AlliTapped(t, g, jiwari) {
		t.Error("Jiwari did not tap")
	}
}

func TestOldB584JiwariChannelDamagesEveryCreatureWithoutFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mineBear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirBear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	bigBear := effectstest.B12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 5, 5)
	bird := oldB584Flyer(g, opp.ID, 2)
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PushCatalogHandCard(me, "Jiwari, the Earth Aflame", "Legendary Creature — Spirit", oldB584JiwariOracle)
	effectstest.FloatMana(t, g, me, "{C}{C}{R}{R}{R}")
	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{Strict: true, XValue: 2}); err != nil {
		t.Fatalf("channel: %v", err)
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the discard is a cost: the card must be in the graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mineBear) || g.Battlefield.Contains(theirBear) {
		t.Error("a 2-toughness ground creature survived X = 2")
	}
	if !g.Battlefield.Contains(bigBear) || !g.Battlefield.Contains(bird) {
		t.Error("a creature that should survive died")
	}
}

func TestOldB584ChannelKeywordGrantsWorkFromHandOnly(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, mana, kw string
	}{
		{"Shinen of Flight's Wings", oldB584WingsOracle, "{U}", "flying"},
		{"Shinen of Stars' Light", oldB584LightOracle, "{C}{W}", "first strike"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
			if effectstest.HasKeywordOnBattlefield(t, g, bear, tc.kw) {
				t.Fatalf("the bear already has %s", tc.kw)
			}
			id, err := effectstest.NeoB07Channel(t, g, tc.name, "Creature — Spirit", tc.oracle, tc.mana, 0,
				game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
			if err != nil {
				t.Fatalf("channel: %v", err)
			}
			if !me.Graveyard.Contains(id) {
				t.Error("the discard is a cost: the card must be in the graveyard")
			}
			effectstest.PassPriorityAroundTable(t, g)
			if !effectstest.HasKeywordOnBattlefield(t, g, bear, tc.kw) {
				t.Errorf("the target lacks %s", tc.kw)
			}
		})
	}
}

func TestOldB584ChannelIsRefusedFromTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	wings := effectstest.B12Push(g, me.ID, "Shinen of Flight's Wings", "Creature — Spirit", oldB584WingsOracle, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{U}")
	if err := g.ActivateCatalogAbility(me.ID, wings, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Fatal("channel was activated from the battlefield")
	}
}

func TestOldB584ShinenOfFearsChillCantBlockAndChannelStopsABlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	chill := effectstest.B12Push(g, me.ID, "Shinen of Fear's Chill", "Creature — Spirit", oldB584ChillOracle, 3, 2)
	if effectstest.AuraRestrictions(t, g, chill)&game.CantBlock == 0 {
		t.Error("Shinen of Fear's Chill can block")
	}
	blocker := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if effectstest.AuraRestrictions(t, g, blocker)&game.CantBlock != 0 {
		t.Fatal("the bear can't block before the channel")
	}
	id, err := effectstest.NeoB07Channel(t, g, "Shinen of Fear's Chill", "Creature — Spirit", oldB584ChillOracle, "{C}{B}", 0,
		game.ActivateAbilityParams{Targets: effectstest.CardTarget(blocker)})
	if err != nil {
		t.Fatalf("channel: %v", err)
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the discard is a cost: the card must be in the graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.AuraRestrictions(t, g, blocker)&game.CantBlock == 0 {
		t.Error("the target can still block")
	}
}

func TestOldB584NourisherBoostsAndChannelBoostsMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	nourisher := effectstest.B12Push(g, me.ID, "Ghost-Lit Nourisher", "Creature — Spirit", oldB584NourisherOracle, 2, 1)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.ApaMana(me, "G", "G", "G")
	effectstest.ApaActivate(t, g, me, nourisher, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("bear is %d/%d, want 4/4", p, tough)
	}

	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	bear = effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	id, err := effectstest.NeoB07Channel(t, g, "Ghost-Lit Nourisher", "Creature — Spirit", oldB584NourisherOracle, "{C}{C}{C}{G}", 1,
		game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
	if err != nil {
		t.Fatalf("channel: %v", err)
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the discard is a cost: the card must be in the graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 6 || tough != 6 {
		t.Errorf("bear is %d/%d, want 6/6", p, tough)
	}
}

func oldB584Counter(t *testing.T, row int, mana string, oppMana string, pay, fromHand bool) (g *game.Game, bolt uuid.UUID, opp *game.Player) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	opp = g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bolt = uuid.New()
	opp.Hand.PushTop(game.Card{
		InstanceID: bolt, Name: "Their Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Keywords: []string{"flash"},
		Owner: opp.ID, Controller: opp.ID,
	})
	if err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{}); err != nil {
		t.Fatalf("opponent CastSpell: %v", err)
	}
	var src uuid.UUID
	if fromHand {
		src = effectstest.PushCatalogHandCard(me, "Ghost-Lit Warder", "Creature — Spirit", oldB584WarderOracle)
	} else {
		src = effectstest.B12Push(g, me.ID, "Ghost-Lit Warder", "Creature — Spirit", oldB584WarderOracle, 1, 1)
	}
	effectstest.FloatMana(t, g, me, mana)
	if oppMana != "" {
		effectstest.FloatMana(t, g, opp, oppMana)
	}
	if err := g.ActivateCatalogAbility(me.ID, src, row, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bolt)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, opp.ID) {
		effectstest.AnswerPayUnless(t, g, opp.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
	}
	return g, bolt, opp
}

func TestOldB584WarderCountersUnlessItsControllerPays(t *testing.T) {
	_, bolt, opp := oldB584Counter(t, 0, "{C}{C}{C}{U}", "", false, false)
	if !opp.Graveyard.Contains(bolt) {
		t.Error("the unpaid spell was not countered")
	}
	g, bolt, opp := oldB584Counter(t, 0, "{C}{C}{C}{U}", "{C}{C}", true, false)
	if opp.Graveyard.Contains(bolt) || !g.Battlefield.Contains(bolt) {
		t.Error("the spell was countered although {2} was paid")
	}
}

func TestOldB584WarderChannelAsksForFour(t *testing.T) {
	_, bolt, opp := oldB584Counter(t, 1, "{C}{C}{C}{U}", "", false, true)
	if !opp.Graveyard.Contains(bolt) {
		t.Error("the unpaid spell was not countered")
	}
	g, bolt, opp := oldB584Counter(t, 1, "{C}{C}{C}{U}", "{C}{C}{C}{C}", true, true)
	if opp.Graveyard.Contains(bolt) || !g.Battlefield.Contains(bolt) {
		t.Error("the spell was countered although {4} was paid")
	}
}

func TestOldB584StalkerMakesAPlayerDiscardAtSorcerySpeed(t *testing.T) {
	for _, tc := range []struct {
		name    string
		channel bool
		n       int
		mana    string
	}{
		{"tap row", false, 2, "{C}{C}{C}{C}{B}"},
		{"channel row", true, 4, "{C}{C}{C}{C}{C}{B}{B}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			opp.Hand.Cards = nil
			for i := 0; i < 5; i++ {
				opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
			}
			effectstest.AdvanceToMain(t, g)
			var src uuid.UUID
			row := 0
			if tc.channel {
				src = effectstest.PushCatalogHandCard(me, "Ghost-Lit Stalker", "Creature — Spirit", oldB584StalkerOracle)
				row = 1
			} else {
				src = effectstest.B12Push(g, me.ID, "Ghost-Lit Stalker", "Creature — Spirit", oldB584StalkerOracle, 1, 1)
			}
			effectstest.FloatMana(t, g, me, tc.mana)
			if err := g.ActivateCatalogAbility(me.ID, src, row, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.DiscardOwed(g, opp.ID); got != tc.n {
				t.Fatalf("owes %d discards, want %d", got, tc.n)
			}
			effectstest.DiscardFromHand(t, g, opp.ID)
			effectstest.PassPriorityAroundTable(t, g)
			if opp.Hand.Size() != 5-tc.n {
				t.Errorf("hand is %d cards, want %d", opp.Hand.Size(), 5-tc.n)
			}
		})
	}
}

func TestOldB584StalkerIsRefusedOutsideSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stalker := effectstest.B12Push(g, me.ID, "Ghost-Lit Stalker", "Creature — Spirit", oldB584StalkerOracle, 1, 1)
	effectstest.B13OpponentCasts(t, g, opp, "Their Bolt", "Instant", "", "{R}", nil)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{B}")
	if err := g.ActivateCatalogAbility(me.ID, stalker, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err == nil {
		t.Fatal("activated at instant speed with a spell on the stack")
	}
}
