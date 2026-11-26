package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b339_cards_test.go — OLD batch 339 (Fate Reforged: pump and combat
// tricks, bolster, activated abilities, creatures with keywords, Auras, #5552):
// twelve cards played through the engine.

const (
	oldB339WarFlareOracle  = "7146eecb-8a79-467c-936e-f577e481cf59"
	oldB339ShuYunOracle    = "7b22f31c-9caa-4afd-8c44-f88346bfa79c"
	oldB339RuthlessOracle  = "b60aff46-2d5a-47e2-9340-39d34a47c3e1"
	oldB339GeneralOracle   = "9fc6e1db-cf09-4b8c-be4f-bb02ffe7a188"
	oldB339MapOracle       = "b845e69b-5d49-40ff-9cf7-d5cf2ccec7be"
	oldB339AdvantageOracle = "e7b2f844-2ba2-4465-bb80-358a4d915f36"
	oldB339TorrentOracle   = "0335e7db-5e0e-4271-8ae7-437a1d2c16ac"
	oldB339BreakOracle     = "555c571e-ccfc-4a8e-8a40-7b1eab0fbbef"
	oldB339WardscaleOracle = "1fb53b6a-b456-4252-9614-12ac72968113"
	oldB339YasovaOracle    = "75dc0995-f89e-4a2c-b30b-9b5a07c21183"
	oldB339VengeanceOracle = "321c4586-d27e-45ae-ba1f-3efea334a697"
	oldB339FruitOracle     = "8d579f04-b494-41a1-ba7c-b9779f75b818"
)

func TestOldB339EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB339WarFlareOracle:  "War Flare",
		oldB339ShuYunOracle:    "Shu Yun, the Silent Tempest",
		oldB339RuthlessOracle:  "Ruthless Instincts",
		oldB339GeneralOracle:   "Dragonscale General",
		oldB339MapOracle:       "Map the Wastes",
		oldB339AdvantageOracle: "Abzan Advantage",
		oldB339TorrentOracle:   "Torrent Elemental",
		oldB339BreakOracle:     "Break Through the Line",
		oldB339WardscaleOracle: "Wardscale Dragon",
		oldB339YasovaOracle:    "Yasova Dragonclaw",
		oldB339VengeanceOracle: "Ancestral Vengeance",
		oldB339FruitOracle:     "Fruit of the First Tree",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name=%q completeness=%v caveats=%v, want Full with none", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB339Tap(g *game.Game, id uuid.UUID, tapped bool) {
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, id).Tapped = tapped })
}

func oldB339Plus(g *game.Game, id uuid.UUID) int {
	return effectstest.CounterCount(g, id, game.CounterPlusOne)
}

func TestOldB339WarFlarePumpsAndUntapsYourCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Push(g, me.ID, "A", "Creature — Bear", "", 2, 2)
	b := effectstest.B12Push(g, me.ID, "B", "Creature — Bear", "", 1, 1)
	theirs := effectstest.B12Push(g, opp.ID, "Theirs", "Creature — Bear", "", 2, 2)
	oldB339Tap(g, a, true)
	oldB339Tap(g, b, true)
	oldB339Tap(g, theirs, true)
	effectstest.CastCatalogSpell(t, g, "War Flare", "Instant", oldB339WarFlareOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID][2]int{a: {4, 3}, b: {3, 2}, theirs: {2, 2}} {
		c := effectstest.GaLayered(t, g, id)
		if c.CurrentPower() != want[0] || c.CurrentToughness() != want[1] {
			t.Errorf("%s is %d/%d, want %d/%d", c.Name, c.CurrentPower(), c.CurrentToughness(), want[0], want[1])
		}
	}
	if effectstest.GaLayered(t, g, a).Tapped || effectstest.GaLayered(t, g, b).Tapped {
		t.Error("one of my creatures is still tapped")
	}
	if !effectstest.GaLayered(t, g, theirs).Tapped {
		t.Error("an opponent's creature was untapped")
	}
}

func TestOldB339ShuYunGivesDoubleStrikeWhenYouPay(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.B12Push(g, me.ID, "Shu Yun, the Silent Tempest", "Legendary Creature — Human Monk", oldB339ShuYunOracle, 3, 2)
		bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
		effectstest.ToMainForCost(t, g)
		spell := effectstest.HandCardFull(me, "Test Spell", "Instant", "", "", nil)
		effectstest.FillPoolColored(me, "R", 2)
		if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.MscM2PickTargets(t, g, bear)
		effectstest.AnswerTriggerOrderInOfferedOrder(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.HasEffectiveKeyword(t, g, bear, "double strike"); got != pay {
			t.Errorf("pay=%v: double strike = %v", pay, got)
		}
	}
}

// oldB339CastRuthless casts Ruthless Instincts from the active player's hand at
// the current step; strict casts must be paid for and refuse illegal targets.
func oldB339CastRuthless(t *testing.T, g *game.Game, strict bool, modes []int, targets []game.TargetRef) error {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.HandCardFull(me, "Ruthless Instincts", "Instant", "{2}{G}", oldB339RuthlessOracle, []string{"G"})
	if strict {
		effectstest.Acrb1AddMana(t, g, me.ID, "{C}{C}{G}")
	}
	err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: modes, Targets: targets, Strict: strict})
	if err != nil && !strict {
		t.Fatalf("cast: %v", err)
	}
	return err
}

func TestOldB339RuthlessInstinctsBothBullets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	idle := effectstest.B12Push(g, me.ID, "Idle", "Creature — Bear", "", 2, 2)
	attacker := effectstest.B12Push(g, me.ID, "Attacker", "Creature — Bear", "", 2, 2)
	oldB339Tap(g, idle, true)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)

	// A creature that is attacking is not a legal target of the first bullet.
	if err := oldB339CastRuthless(t, g, true, []int{0}, effectstest.CardRefs(attacker)); err == nil {
		t.Error("an attacking creature was accepted for the nonattacking bullet")
	}
	if err := oldB339CastRuthless(t, g, true, []int{1}, effectstest.CardRefs(idle)); err == nil {
		t.Error("a nonattacking creature was accepted for the attacking bullet")
	}

	_ = oldB339CastRuthless(t, g, false, []int{0}, effectstest.CardRefs(idle))
	effectstest.PassPriorityAroundTable(t, g)
	for _, kw := range []string{"reach", "deathtouch"} {
		if !effectstest.HasEffectiveKeyword(t, g, idle, kw) {
			t.Errorf("the nonattacking creature lacks %s", kw)
		}
	}
	if effectstest.GaLayered(t, g, idle).Tapped {
		t.Error("the creature was not untapped")
	}

	_ = oldB339CastRuthless(t, g, false, []int{1}, effectstest.CardRefs(attacker))
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.GaLayered(t, g, attacker); c.CurrentPower() != 4 || c.CurrentToughness() != 4 || !effectstest.HasEffectiveKeyword(t, g, attacker, "trample") {
		t.Errorf("the attacker is %d/%d, want 4/4 with trample", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestOldB339DragonscaleGeneralBolstersByTappedCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	general := effectstest.B12Push(g, me.ID, "Dragonscale General", "Creature — Human Warrior", oldB339GeneralOracle, 2, 3)
	small := effectstest.B12Push(g, me.ID, "Small", "Creature — Bear", "", 1, 1)
	idle := effectstest.B12Push(g, me.ID, "Idle", "Creature — Bear", "", 2, 2)
	oldB339Tap(g, general, true)
	oldB339Tap(g, small, true)
	oldB339Tap(g, idle, false)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB339Plus(g, small); got != 2 {
		t.Errorf("the creature with the least toughness has %d counters, want 2 (two tapped creatures)", got)
	}
	if oldB339Plus(g, general) != 0 || oldB339Plus(g, idle) != 0 {
		t.Error("a creature that was not the least toughness got counters")
	}
}

func TestOldB339MapTheWastesFetchesATappedLandThenBolsters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	me.Library.Cards = nil
	forest := effectstest.PushLibraryCard(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.PushLibraryCard(me, game.Card{Name: "Filler", TypeLine: "Sorcery"})
	effectstest.ToMainForCost(t, g)
	effectstest.CastCatalogSpell(t, g, "Map the Wastes", "Sorcery", oldB339MapOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, forest)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(forest) || !effectstest.GaLayered(t, g, forest).Tapped {
		t.Error("the basic land did not enter tapped")
	}
	if oldB339Plus(g, bear) != 1 {
		t.Errorf("the bolstered creature has %d counters, want 1", oldB339Plus(g, bear))
	}
}

func TestOldB339AbzanAdvantageMakesThemSacrificeAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	e1 := effectstest.B12Permanent(g, opp.ID, "Ench One", "Enchantment")
	e2 := effectstest.B12Permanent(g, opp.ID, "Ench Two", "Enchantment")
	effectstest.CastCatalogSpell(t, g, "Abzan Advantage", "Instant", oldB339AdvantageOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, opp.ID, e2)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(e1) || g.Battlefield.Contains(e2) {
		t.Error("the chosen enchantment was not the one sacrificed")
	}
	if oldB339Plus(g, bear) != 1 {
		t.Errorf("my creature has %d counters, want the bolster's 1", oldB339Plus(g, bear))
	}
}

func TestOldB339AbzanAdvantageStillBolstersWithNothingToSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Abzan Advantage", "Instant", oldB339AdvantageOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if oldB339Plus(g, bear) != 1 {
		t.Errorf("my creature has %d counters, want 1 although they had no enchantment", oldB339Plus(g, bear))
	}
}

func TestOldB339TorrentElementalTapsTheDefendersCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	other := g.Seats[2]
	el := effectstest.B12Push(g, me.ID, "Torrent Elemental", "Creature — Elemental", oldB339TorrentOracle, 3, 5)
	d1 := effectstest.B12Push(g, opp.ID, "D1", "Creature — Bear", "", 2, 2)
	d2 := effectstest.B12Push(g, opp.ID, "D2", "Creature — Bear", "", 2, 2)
	bystander := effectstest.B12Push(g, other.ID, "Bystander", "Creature — Bear", "", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, el)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.GaLayered(t, g, d1).Tapped || !effectstest.GaLayered(t, g, d2).Tapped {
		t.Error("a creature the defending player controls is untapped")
	}
	if effectstest.GaLayered(t, g, bystander).Tapped {
		t.Error("a creature of a player who was not attacked was tapped")
	}
}

func TestOldB339TorrentElementalComesBackFromExileTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		g.Exile.PushTop(game.Card{InstanceID: id, Name: "Torrent Elemental", TypeLine: "Creature — Elemental", OracleID: oldB339TorrentOracle, Power: 3, Toughness: 5, Owner: me.ID, Controller: me.ID})
	})
	effectstest.Acrb1AddMana(t, g, me.ID, "{C}{C}{C}{B}{G}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate from exile: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) && effectstest.FindBattlefieldByName(g, "Torrent Elemental") == uuid.Nil {
		t.Fatal("the card did not come back from exile")
	}
	back := effectstest.FindBattlefieldByName(g, "Torrent Elemental")
	if !effectstest.GaLayered(t, g, back).Tapped {
		t.Error("it entered untapped")
	}
}

func TestOldB339TorrentElementalCannotReturnFromExileOutsideSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	id := uuid.New()
	g.WithWriteLock(func() {
		g.Exile.PushTop(game.Card{InstanceID: id, Name: "Torrent Elemental", TypeLine: "Creature — Elemental", OracleID: oldB339TorrentOracle, Owner: me.ID, Controller: me.ID})
	})
	effectstest.Acrb1AddMana(t, g, me.ID, "{C}{C}{C}{B}{G}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the ability was activated at the beginning of combat")
	}
}

func TestOldB339BreakThroughTheLineGivesHasteAndEvasionToASmallCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ench := effectstest.PushCatalogPermanent(g, me.ID, "Break Through the Line", "Enchantment", oldB339BreakOracle, false)
	small := effectstest.B12Push(g, me.ID, "Small", "Creature — Bear", "", 2, 2)
	big := effectstest.B12Push(g, me.ID, "Big", "Creature — Giant", "", 3, 3)
	effectstest.ToMainForCost(t, g)
	effectstest.Acrb1AddMana(t, g, me.ID, "{R}{R}")
	if err := g.ActivateCatalogAbility(me.ID, ench, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(big), Strict: true}); err == nil {
		t.Error("a creature with power 3 was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, ench, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(small)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, small, "haste") {
		t.Error("the creature did not gain haste")
	}
	if effectstest.RestrictionsOf(t, g, small)&game.CantBeBlocked == 0 {
		t.Error("the creature can be blocked")
	}
}

func TestOldB339WardscaleDragonStopsTheDefenderCastingWhileAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	dragon := effectstest.B12Push(g, me.ID, "Wardscale Dragon", "Creature — Dragon", oldB339WardscaleOracle, 4, 4)
	cast := func(p *game.Player) error {
		spell := effectstest.HandCardFull(p, "Their Bolt", "Instant", "{R}", "", []string{"R"})
		if err := g.AddManaForEffect(p.ID, uuid.Nil, "{R}"); err != nil {
			t.Fatalf("AddManaForEffect: %v", err)
		}
		return g.CastSpell(p.ID, spell, game.CastSpellParams{Strict: true})
	}
	effectstest.ToMainForCost(t, g)
	if err := cast(opp); err != nil {
		t.Fatalf("the player could not cast while the Dragon was not attacking: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, dragon)
	if err := cast(opp); err == nil {
		t.Error("the defending player cast a spell while the Dragon was attacking them")
	}
	if err := cast(third); err != nil {
		t.Errorf("a player who is not being attacked could not cast: %v", err)
	}
}

func TestOldB339YasovaStealsASmallerCreatureWhenYouPay(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.B12Push(g, me.ID, "Yasova Dragonclaw", "Legendary Creature — Human Warrior", oldB339YasovaOracle, 4, 2)
		small := effectstest.B12Push(g, opp.ID, "Small", "Creature — Bear", "", 3, 3)
		equal := effectstest.B12Push(g, opp.ID, "Equal", "Creature — Giant", "", 4, 4)
		oldB339Tap(g, small, true)
		effectstest.AdvanceTo(t, g, game.StepBeginCombat)
		prompt := effectstest.LatestPickTarget(g, me.ID)
		if prompt == nil {
			t.Fatal("no target prompt at the beginning of combat")
		}
		if effectstest.HasID(prompt.PickTargetCards, equal) || !effectstest.HasID(prompt.PickTargetCards, small) {
			t.Fatalf("targets %v: only a creature with power less than Yasova's is legal", prompt.PickTargetCards)
		}
		effectstest.MscM2PickTargets(t, g, small)
		effectstest.FillPoolColored(me, "R", 3)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		stolen := effectstest.ControllerOf(t, g, small) == me.ID
		if stolen != pay {
			t.Errorf("pay=%v: stolen=%v", pay, stolen)
		}
		if pay && (effectstest.GaLayered(t, g, small).Tapped || !effectstest.HasEffectiveKeyword(t, g, small, "haste")) {
			t.Error("the stolen creature was not untapped and given haste")
		}
	}
}

func TestOldB339AncestralVengeanceShrinksTheHostAndGrowsYours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Push(g, opp.ID, "Host", "Creature — Bear", "", 2, 2)
	mine := effectstest.B12Push(g, me.ID, "Mine", "Creature — Bear", "", 2, 2)
	aura := effectstest.CastCatalogSpell(t, g, "Ancestral Vengeance", "Enchantment — Aura", oldB339VengeanceOracle, effectstest.CardRefs(host))
	effectstest.AdResolveSpell(t, g, aura)
	effectstest.MscM2PickTargets(t, g, mine)
	if c := effectstest.GaLayered(t, g, host); c.CurrentPower() != 1 || c.CurrentToughness() != 1 {
		t.Errorf("the enchanted creature is %d/%d, want 1/1", c.CurrentPower(), c.CurrentToughness())
	}
	if oldB339Plus(g, mine) != 1 {
		t.Errorf("my creature has %d counters, want 1", oldB339Plus(g, mine))
	}
}

func TestOldB339AncestralVengeanceKillsAOneToughnessCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Push(g, opp.ID, "Host", "Creature — Saproling", "", 1, 1)
	mine := effectstest.B12Push(g, me.ID, "Mine", "Creature — Bear", "", 2, 2)
	aura := effectstest.CastCatalogSpell(t, g, "Ancestral Vengeance", "Enchantment — Aura", oldB339VengeanceOracle, effectstest.CardRefs(host))
	effectstest.AdResolveSpell(t, g, aura)
	effectstest.MscM2PickTargets(t, g, mine)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(host) {
		t.Error("the 1/1 survived the -1/-1")
	}
}

func TestOldB339FruitOfTheFirstTreeGainsAndDrawsItsToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.B12Push(g, me.ID, "Host", "Creature — Giant", "", 3, 4)
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F")
	aura := effectstest.CastCatalogSpell(t, g, "Fruit of the First Tree", "Enchantment — Aura", oldB339FruitOracle, effectstest.CardRefs(host))
	effectstest.AdResolveSpell(t, g, aura)
	life, hand := me.Life, me.Hand.Size()
	effectstest.P1Destroy(t, g, host)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+4 || me.Hand.Size() != hand+4 {
		t.Errorf("life %d → %d, hand %d → %d, want 4 life and 4 cards", life, me.Life, hand, me.Hand.Size())
	}
}
