package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b618_cards_test.go — S82 OLD batch 618 (#5607): Eighth Edition
// instants, sorceries, card draw and library. Every test plays the real
// catalog card through the engine.

const (
	oldB618Affinity = "09a1ab1b-d9f0-4a2a-a448-3190f85006e4"
	oldB618Plow     = "23b36aef-961f-4a78-9652-8594ad73102e"
	oldB618Hibern   = "633b541b-e37f-451a-be73-3398e61e0f75"
	oldB618Persec   = "676c4cfa-f929-43aa-95cc-df81a596c2a2"
	oldB618Twiddle  = "773ad2ef-5acc-49ea-8d85-056330e87039"
	oldB618Routes   = "120fa67e-e5e0-4c23-9a78-d6171d357aee"
	oldB618Slay     = "64418e1f-4fb1-4029-a31e-135b2d2c0ab9"
	oldB618Balance  = "c57ee68e-6832-4d4c-b710-fc03a2c10a9c"
	oldB618Portal   = "d4b10505-2b7a-4099-aacb-ff2359b24310"
)

func TestOldB618CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB618Affinity: "Natural Affinity", oldB618Plow: "Plow Under", oldB618Hibern: "Hibernation",
		oldB618Persec: "Persecute", oldB618Twiddle: "Twiddle", oldB618Routes: "Trade Routes",
		oldB618Slay: "Slay", oldB618Balance: "Balance of Power", oldB618Portal: "Planar Portal",
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

func oldB618IsCreature(t *testing.T, g *game.Game, id uuid.UUID) bool {
	return effectstest.ContainsString(effectstest.EffectiveTypes(t, g, id), "Creature")
}

func TestOldB618NaturalAffinityAnimatesEveryLandUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	mine := effectstest.PushLandFor(g, g.Seats[0].ID, "Forest", "Basic Land — Forest")
	theirs := effectstest.PushLandFor(g, g.Seats[1].ID, "Island", "Basic Land — Island")
	effectstest.CastCatalogSpell(t, g, "Natural Affinity", "Instant", oldB618Affinity, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"my land": mine, "their land": theirs} {
		if !oldB618IsCreature(t, g, id) {
			t.Errorf("%s is not a creature", name)
		}
		if !effectstest.ContainsString(effectstest.EffectiveTypes(t, g, id), "Land") {
			t.Errorf("%s stopped being a land", name)
		}
		if p, tough := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id); p != 2 || tough != 2 {
			t.Errorf("%s is %d/%d, want 2/2", name, p, tough)
		}
	}
	later := effectstest.PushLandFor(g, g.Seats[0].ID, "Forest", "Basic Land — Forest")
	if oldB618IsCreature(t, g, later) {
		t.Error("a land that entered after resolution was animated (CR 611.2c)")
	}
	effectstest.B39NextTurnOf(t, g, 1)
	if oldB618IsCreature(t, g, mine) || oldB618IsCreature(t, g, theirs) {
		t.Error("the lands stayed creatures after end of turn")
	}
}

func TestOldB618PlowUnderPutsTwoLandsOnTopOfTheirOwnersLibraries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	theirs := effectstest.PushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	effectstest.CastCatalogSpell(t, g, "Plow Under", "Sorcery", oldB618Plow,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range []struct {
		name string
		p    *game.Player
		id   uuid.UUID
	}{{"mine", me, mine}, {"theirs", opp, theirs}} {
		if g.Battlefield.Contains(c.id) {
			t.Errorf("%s is still on the battlefield", c.name)
		}
		if n := len(c.p.Library.Cards); n == 0 || c.p.Library.Cards[n-1].InstanceID != c.id {
			t.Errorf("%s is not on top of its owner's library", c.name)
		}
	}
}

func TestOldB618PlowUnderRefusesANonland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	land := effectstest.PushLandFor(g, g.Seats[0].ID, "Forest", "Basic Land — Forest")
	bear := effectstest.B12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Plow Under", "Sorcery", oldB618Plow,
		[]game.TargetRef{{Kind: game.TargetCard, ID: land}, {Kind: game.TargetCard, ID: bear}}); err == nil {
		t.Error("Plow Under accepted a creature as a target")
	}
}

func TestOldB618HibernationReturnsOnlyGreenPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	green := effectstest.B12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	oppGreen := effectstest.B12Creature(g, opp.ID, "Elf", "Creature — Elf", 1, 1)
	red := effectstest.B12Creature(g, opp.ID, "Goblin", "Creature — Goblin", 1, 1)
	land := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			switch g.Battlefield.Cards[i].InstanceID {
			case green, oppGreen:
				g.Battlefield.Cards[i].Colors = []string{"G"}
			case red:
				g.Battlefield.Cards[i].Colors = []string{"R"}
			}
		}
	})
	effectstest.CastCatalogSpell(t, g, "Hibernation", "Instant", oldB618Hibern, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(green) || !me.Hand.Contains(green) {
		t.Error("my green creature did not return to my hand")
	}
	if g.Battlefield.Contains(oppGreen) || !opp.Hand.Contains(oppGreen) {
		t.Error("their green creature did not return to its owner's hand")
	}
	if !g.Battlefield.Contains(red) || !g.Battlefield.Contains(land) {
		t.Error("a nongreen permanent was returned")
	}
}

func TestOldB618PersecuteDiscardsEveryCardOfTheChosenColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() { opp.Hand.Cards = nil })
	push := func(name string, colors ...string) uuid.UUID {
		id := uuid.New()
		opp.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Sorcery", Colors: colors, Owner: opp.ID, Controller: opp.ID})
		return id
	}
	red1, red2 := push("Red A", "R"), push("Red B", "R")
	gold := push("Gold", "R", "W")
	blue := push("Blue", "U")
	land := push("Plain Land")
	effectstest.CastCatalogSpell(t, g, "Persecute", "Sorcery", oldB618Persec, effectstest.B16TargetPlayer(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "R")
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{red1, red2, gold} {
		if opp.Hand.Contains(id) || !opp.Graveyard.Contains(id) {
			t.Errorf("a red card %s was not discarded", id)
		}
	}
	if !opp.Hand.Contains(blue) || !opp.Hand.Contains(land) {
		t.Error("a card that is not red was discarded")
	}
}

func TestOldB618PersecuteWithNoMatchingCardDiscardsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() { opp.Hand.Cards = nil })
	id := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Blue", TypeLine: "Sorcery", Colors: []string{"U"}, Owner: opp.ID, Controller: opp.ID})
	effectstest.CastCatalogSpell(t, g, "Persecute", "Sorcery", oldB618Persec, effectstest.B16TargetPlayer(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "G")
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(id) {
		t.Error("a blue card was discarded to a green Persecute")
	}
}

func TestOldB618TwiddleTapsOrUntapsTheChoice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Twiddle", "Instant", oldB618Twiddle, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	if !effectstest.B16Tapped(t, g, bear) {
		t.Fatal("the creature was not tapped")
	}
	effectstest.CastCatalogSpell(t, g, "Twiddle", "Instant", oldB618Twiddle, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	if effectstest.B16Tapped(t, g, bear) {
		t.Error("the creature was not untapped")
	}
}

func TestOldB618TwiddleRefusesAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	ench := effectstest.B12Permanent(g, g.Seats[1].ID, "Pacifism", "Enchantment — Aura")
	if err := effectstest.CastCatalogSpellErr(t, g, "Twiddle", "Instant", oldB618Twiddle, effectstest.B16TargetCard(ench)); err == nil {
		t.Error("Twiddle targeted an enchantment")
	}
}

func TestOldB618SlayDestroysAGreenCreatureEvenWithAShieldAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	elf := effectstest.B12Creature(g, g.Seats[1].ID, "Elf", "Creature — Elf", 1, 1)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == elf {
				g.Battlefield.Cards[i].Colors = []string{"G"}
			}
		}
	})
	effectstest.P1Regenerate(t, g, elf)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Slay", "Instant", oldB618Slay, effectstest.B16TargetCard(elf))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(elf) {
		t.Error("the regeneration shield saved the creature from Slay")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, got)
	}
}

func TestOldB618SlayRefusesANonGreenCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.B12Creature(g, g.Seats[1].ID, "Goblin", "Creature — Goblin", 1, 1)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].Colors = []string{"R"}
			}
		}
	})
	if err := effectstest.CastCatalogSpellErr(t, g, "Slay", "Instant", oldB618Slay, effectstest.B16TargetCard(bear)); err == nil {
		t.Error("Slay targeted a red creature")
	}
}

func oldB618SetHand(p *game.Player, n int) {
	p.Hand.Cards = nil
	for i := 0; i < n; i++ {
		effectstest.HandCardForTest(p, "Filler", "Sorcery", "")
	}
}

func TestOldB618BalanceOfPowerDrawsTheDifference(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F")
	g.WithWriteLock(func() { oldB618SetHand(me, 1); oldB618SetHand(opp, 4) })
	// The spell is cast from the hand, so the caster holds one card fewer
	// as it resolves: 1 filler + the spell → 1 filler after casting.
	effectstest.CastCatalogSpell(t, g, "Balance of Power", "Sorcery", oldB618Balance, effectstest.B16TargetPlayer(opp.ID))
	before := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if got, want := me.Hand.Size()-before, 4-before; got != want {
		t.Errorf("drew %d, want %d (the difference of %d and %d)", got, want, 4, before)
	}
}

func TestOldB618BalanceOfPowerDrawsNothingWhenTheOpponentHasNoMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B", "C")
	g.WithWriteLock(func() { oldB618SetHand(me, 3); oldB618SetHand(opp, 2) })
	effectstest.CastCatalogSpell(t, g, "Balance of Power", "Sorcery", oldB618Balance, effectstest.B16TargetPlayer(opp.ID))
	before := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before {
		t.Errorf("hand %d → %d, want no draw", before, got)
	}
}

func oldB618Activate(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, idx int, mana string, params game.ActivateAbilityParams) error {
	t.Helper()
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if mana != "" {
		effectstest.DcMana(t, g, me.ID, mana)
	}
	params.Strict = true
	err := g.ActivateCatalogAbility(me.ID, id, idx, params)
	if err == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	return err
}

func TestOldB618TradeRoutesBouncesALandYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	routes := effectstest.B12Push(g, me.ID, "Trade Routes", "Enchantment", oldB618Routes, 0, 0)
	land := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	theirs := effectstest.PushLandFor(g, g.Seats[1].ID, "Island", "Basic Land — Island")
	if err := oldB618Activate(t, g, me, routes, 0, "{C}", game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(theirs)}); err == nil {
		t.Fatal("bounced a land another player controls")
	}
	if err := oldB618Activate(t, g, me, routes, 0, "{C}", game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(land)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(land) || !me.Hand.Contains(land) {
		t.Error("the land did not return to its owner's hand")
	}
}

func TestOldB618TradeRoutesDiscardsALandToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	routes := effectstest.B12Push(g, me.ID, "Trade Routes", "Enchantment", oldB618Routes, 0, 0)
	effectstest.SeedLibrary(me, "A", "B", "C")
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	spell := effectstest.HandCardForTest(me, "Some Spell", "Sorcery", "")
	land := effectstest.HandCardForTest(me, "Forest", "Basic Land — Forest", "")
	if err := oldB618Activate(t, g, me, routes, 1, "{C}", game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{spell}}); err == nil {
		t.Fatal("discarded a nonland card to pay")
	}
	if err := oldB618Activate(t, g, me, routes, 1, "{C}", game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{land}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Graveyard.Contains(land) {
		t.Error("the land was not discarded")
	}
	if me.Hand.Size() != 2 || !me.Hand.Contains(spell) {
		t.Errorf("hand has %d cards, want the spell and one drawn card", me.Hand.Size())
	}
}

func TestOldB618PlanarPortalTutorsAnyCardForSix(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	portal := effectstest.B12Push(g, me.ID, "Planar Portal", "Artifact", oldB618Portal, 0, 0)
	ids := effectstest.SeedLibrary(me, "Alpha", "Beta", "Gamma")
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.DcMana(t, g, me.ID, "{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, portal, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated Planar Portal for five")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.DcMana(t, g, me.ID, "{C}{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, portal, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, ids[1])
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[1]) {
		t.Error("the chosen card is not in hand")
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, portal); !tapped {
		t.Error("Planar Portal was not tapped")
	}
}
