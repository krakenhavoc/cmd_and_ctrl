package testsst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sth_b2_cards_test.go — Stronghold, batch 2 (#5627). Each test plays the
// real catalog entry and asserts what its text says. The seeding and
// buyback helpers are Tempest batch 1's (tmp_b1_cards_test.go).

const (
	sthB2ChangeOfHeartOracle = "a8c0b41b-916c-4da9-9c39-a058c91a78f4"
	sthB2FanningOracle       = "1d5f0a1c-88e0-4603-97af-8ed47766812e"
	sthB2LabRatsOracle       = "c53d6d3e-64e3-48b9-afab-3bde3164888f"
	sthB2SiftOracle          = "6e060984-2015-4dc4-b53d-ab6bf2abdacc"
	sthB2BandageOracle       = "c664df64-0103-462b-a0f8-c483b152974a"
	sthB2TormentOracle       = "b53ec6e6-fcc9-4471-88c5-7ad0fbd7bbea"
	sthB2PrimalRageOracle    = "d6464ee4-23fc-4d68-bbda-3b53772015d1"
	sthB2AssassinOracle      = "35826db0-b6ff-47b5-be16-c5b878d9e842"
	sthB2RevenantOracle      = "05fc62d3-9b3c-409a-a623-9f81f395a5a0"
	sthB2SpikeWorkerOracle   = "d7fcf43a-191f-4f44-aade-0cf89079a0c5"
	sthB2SpikeColonyOracle   = "15902a15-090a-44e6-b8d2-59e0f26ed46f"
	sthB2SpikeFeederOracle   = "5b147eab-39dd-4c2a-8f59-d778037788aa"
	sthB2SpikeSoldierOracle  = "a34d1845-e02f-4795-b445-d73366c2baea"
	sthB2MorgueThrullOracle  = "fc759e09-34b2-4004-8474-a6500dd1280c"
	sthB2ArcherOracle        = "42641a05-a6b6-4856-a72c-54181b97c190"
	sthB2SwordOracle         = "3c756eda-4fc1-4766-99be-32d8c9f35262"
)

func TestSthB2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		sthB2ChangeOfHeartOracle: "Change of Heart",
		sthB2FanningOracle:       "Fanning the Flames",
		sthB2LabRatsOracle:       "Lab Rats",
		sthB2SiftOracle:          "Sift",
		sthB2BandageOracle:       "Bandage",
		sthB2TormentOracle:       "Torment",
		sthB2PrimalRageOracle:    "Primal Rage",
		sthB2AssassinOracle:      "Stronghold Assassin",
		sthB2RevenantOracle:      "Revenant",
		sthB2SpikeWorkerOracle:   "Spike Worker",
		sthB2SpikeColonyOracle:   "Spike Colony",
		sthB2SpikeFeederOracle:   "Spike Feeder",
		sthB2SpikeSoldierOracle:  "Spike Soldier",
		sthB2MorgueThrullOracle:  "Morgue Thrull",
		sthB2ArcherOracle:        "Skyshroud Archer",
		sthB2SwordOracle:         "Sword of the Chosen",
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

// sthB2Keyworded seeds a vanilla creature with printed keywords.
func sthB2Keyworded(g *game.Game, owner uuid.UUID, name, typeLine string, power, toughness int, keywords ...string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Power: power, Toughness: toughness, Keywords: keywords, Owner: owner, Controller: owner,
	})
}

// --- buyback spells -----------------------------------------------------------------

func TestSthB2ChangeOfHeartStopsTheTargetAttacking(t *testing.T) {
	for _, tc := range effectstest.TmpB1BuybackCases {
		t.Run(tc.Name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			caster, opp := effectstest.SeatsForCombat(g)
			bear := effectstest.B12Creature(g, caster.ID, "Bear", "Creature — Bear", 2, 2)
			other := effectstest.B12Creature(g, caster.ID, "Other Bear", "Creature — Bear", 2, 2)
			id, err := effectstest.CastWithOptionalCosts(t, g, "Change of Heart", "Instant", sthB2ChangeOfHeartOracle,
				effectstest.CardTarget(bear), tc.Optional, nil)
			if err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.TmpB1AssertBuyback(t, caster, id, tc.WantHand)
			effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
			if err := g.DeclareAttacker(bear, opp.ID); err == nil {
				t.Error("the target was declared as an attacker")
			}
			if err := g.DeclareAttacker(other, opp.ID); err != nil {
				t.Errorf("a creature that was not the target could not attack: %v", err)
			}
		})
	}
}

// sthB2CastFanning casts Fanning the Flames from the active seat's hand
// with X and the optional costs claimed.
func sthB2CastFanning(t *testing.T, g *game.Game, x int, target game.TargetRef, optional []int) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Fanning the Flames", TypeLine: "Sorcery", ManaCost: "{X}{R}{R}",
		OracleID: sthB2FanningOracle, Owner: active.ID, Controller: active.ID,
	})
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{
		XValue: x, Targets: []game.TargetRef{target}, OptionalCosts: optional,
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestSthB2FanningTheFlamesDealsXToAPlayer(t *testing.T) {
	for _, tc := range effectstest.TmpB1BuybackCases {
		t.Run(tc.Name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			caster, victim := effectstest.SeatsForCombat(g)
			life := victim.Life
			id := sthB2CastFanning(t, g, 3, effectstest.PlayerTarget(victim.ID), tc.Optional)
			if got := life - victim.Life; got != 3 {
				t.Errorf("the target lost %d life, want 3", got)
			}
			effectstest.TmpB1AssertBuyback(t, caster, id, tc.WantHand)
		})
	}
}

func TestSthB2FanningTheFlamesDealsXToACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.SeatsForCombat(g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	sthB2CastFanning(t, g, 2, effectstest.CardTarget(bear)[0], nil)
	if g.Battlefield.Contains(bear) {
		t.Error("the 2/2 survived 2 damage")
	}
}

func TestSthB2LabRatsCreatesARat(t *testing.T) {
	for _, tc := range effectstest.TmpB1BuybackCases {
		t.Run(tc.Name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			caster := g.Seats[g.Turn.ActiveSeat]
			id, err := effectstest.CastWithOptionalCosts(t, g, "Lab Rats", "Sorcery", sthB2LabRatsOracle, nil, tc.Optional, nil)
			if err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			var rats []game.Card
			for _, c := range g.Battlefield.Cards {
				if c.IsToken() && c.Name == "Rat" && c.Controller == caster.ID {
					rats = append(rats, c)
				}
			}
			if len(rats) != 1 {
				t.Fatalf("%d Rat tokens, want 1", len(rats))
			}
			if r := rats[0]; r.Power != 1 || r.Toughness != 1 || !r.HasColor("B") {
				t.Errorf("the Rat is %d/%d with colours %v, want a 1/1 black", r.Power, r.Toughness, r.Colors)
			}
			effectstest.TmpB1AssertBuyback(t, caster, id, tc.WantHand)
		})
	}
}

// --- Sift, Bandage --------------------------------------------------------------------

func TestSthB2SiftDrawsThreeThenDiscardsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Library.Cards = nil
	effectstest.SeedLibrary(me, "One", "Two", "Three", "Four")
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Sift", "Sorcery", sthB2SiftOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DiscardFromHand(t, g, me.ID)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("the hand grew by %d, want 2 (three drawn, one discarded)", got)
	}
	if me.Library.Size() != 1 {
		t.Errorf("library %d, want 1 left after three draws", me.Library.Size())
	}
}

func TestSthB2BandagePreventsOneAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "Drawn")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 4)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Bandage", "Instant", sthB2BandageOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("the hand grew by %d, want 1", got)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 3) })
	if got := effectstest.DamageMarkedOn(g, bear); got != 2 {
		t.Errorf("3 damage into a 1-shield: marked %d, want 2", got)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 1) })
	if got := effectstest.DamageMarkedOn(g, bear); got != 3 {
		t.Errorf("the shield is spent: marked %d, want 3", got)
	}
}

func TestSthB2BandageShieldsAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "Drawn")
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Bandage", "Instant", sthB2BandageOracle, []game.TargetRef{effectstest.PlayerTarget(me.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(uuid.New(), me.ID, 3) })
	if got := life - me.Life; got != 2 {
		t.Errorf("3 damage into a 1-shield: lost %d life, want 2", got)
	}
}

// --- Torment, Primal Rage ---------------------------------------------------------------

func TestSthB2TormentGivesMinusThreeMinusZero(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.SeatsForCombat(g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Torment", "Enchantment — Aura", sthB2TormentOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != -1 || tough != 2 {
		t.Errorf("the enchanted creature is %d/%d, want -1/2", p, tough)
	}
}

func TestSthB2PrimalRageGivesYourCreaturesTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.TmpB1Push(g, me.ID, "Primal Rage", "Enchantment", sthB2PrimalRageOracle, 0, 0)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, mine, "trample") {
		t.Error("a creature you control has no trample")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "trample") {
		t.Error("an opponent's creature has trample")
	}
}

// --- Stronghold Assassin, Morgue Thrull ------------------------------------------------

func TestSthB2StrongholdAssassinSacrificesToDestroyANonblackCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	assassin := effectstest.TmpB1Push(g, me.ID, "Stronghold Assassin", "Creature — Phyrexian Zombie Assassin", sthB2AssassinOracle, 2, 1, "B")
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Rat", 1, 1)
	black := effectstest.TmpB1Push(g, opp.ID, "Black Knight", "Creature — Human Knight", "", 2, 2, "B")
	green := effectstest.TmpB1Push(g, opp.ID, "Green Bear", "Creature — Bear", "", 2, 2, "G")
	if err := g.ActivateCatalogAbility(me.ID, assassin, 0, game.ActivateAbilityParams{
		Targets: effectstest.CardRefs(black), SacrificeIDs: []uuid.UUID{fodder},
	}); err == nil {
		t.Fatal("a black creature was accepted as the target")
	}
	if err := g.ActivateCatalogAbility(me.ID, assassin, 0, game.ActivateAbilityParams{
		Targets: effectstest.CardRefs(green), SacrificeIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	if !effectstest.IsTapped(g, assassin) {
		t.Error("the Assassin did not tap")
	}
	if g.Battlefield.Contains(green) {
		t.Error("the nonblack target survived")
	}
	if !g.Battlefield.Contains(black) {
		t.Error("the black creature was destroyed")
	}
}

func TestSthB2MorgueThrullSacrificesToMillThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Library.Cards = nil
	top := effectstest.SeedLibrary(me, "One", "Two", "Three", "Four")
	thrull := effectstest.TmpB1Push(g, me.ID, "Morgue Thrull", "Creature — Thrull", sthB2MorgueThrullOracle, 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, thrull, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(thrull) {
		t.Error("the Thrull was not sacrificed")
	}
	for _, id := range top[:3] {
		if !me.Graveyard.Contains(id) {
			t.Errorf("card %s of the top three was not milled", id)
		}
	}
	if me.Library.Size() != 1 || !me.Library.Contains(top[3]) {
		t.Errorf("library %d, want only the fourth card left", me.Library.Size())
	}
}

// --- Revenant ------------------------------------------------------------------------

func TestSthB2RevenantCountsCreatureCardsInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	grave := func(p *game.Player, typeLine string) {
		p.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	}
	grave(me, "Creature — Bear")
	grave(me, "Creature — Elf")
	grave(me, "Instant")
	grave(opp, "Creature — Bear")
	revenant := effectstest.TmpB1Push(g, me.ID, "Revenant", "Creature — Spirit", sthB2RevenantOracle, 0, 0)
	if p, tough := effectstest.PtOf(t, g, revenant); p != 2 || tough != 2 {
		t.Errorf("Revenant is %d/%d with two creature cards in your graveyard, want 2/2", p, tough)
	}
	// A creature of yours dying is the third creature card.
	golem := effectstest.B12Creature(g, me.ID, "Golem", "Artifact Creature — Golem", 3, 3)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(golem); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	if !me.Graveyard.Contains(golem) {
		t.Fatal("the destroyed creature is not in its owner's graveyard")
	}
	if p, tough := effectstest.PtOf(t, g, revenant); p != 3 || tough != 3 {
		t.Errorf("Revenant is %d/%d after a third creature card, want 3/3", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, revenant, "flying") {
		t.Error("Revenant has no flying")
	}
}

// --- the Spike cycle ---------------------------------------------------------------------

func TestSthB2SpikesEnterWithTheirCounters(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine string
		counters               int
	}{
		{"Spike Worker", sthB2SpikeWorkerOracle, "Creature — Spike", 2},
		{"Spike Colony", sthB2SpikeColonyOracle, "Creature — Spike", 4},
		{"Spike Feeder", sthB2SpikeFeederOracle, "Creature — Spike", 2},
		{"Spike Soldier", sthB2SpikeSoldierOracle, "Creature — Spike Soldier", 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.CastCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.CountersOn(g, id, game.CounterPlusOne); got != tc.counters {
				t.Errorf("%d +1/+1 counters, want %d", got, tc.counters)
			}
			if p, tough := effectstest.PtOf(t, g, id); p != tc.counters || tough != tc.counters {
				t.Errorf("%s is %d/%d, want %d/%d", tc.name, p, tough, tc.counters, tc.counters)
			}
		})
	}
}

func TestSthB2SpikeMovesACounterToTargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, spike := effectstest.SthB2Spike(t, g, "Spike Worker", "Creature — Spike", sthB2SpikeWorkerOracle, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DcMana(t, g, me.ID, "{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, spike, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := effectstest.CountersOn(g, spike, game.CounterPlusOne); got != 1 {
		t.Errorf("the Spike has %d counters after paying one, want 1", got)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the target has %d counters, want 1", got)
	}
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the target is %d/%d, want 3/3", p, tough)
	}
	if err := g.ActivateCatalogAbility(me.ID, spike, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err == nil {
		t.Error("the move was activated with an empty mana pool")
	}
}

func TestSthB2SpikeWithNoCounterCantMoveOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, spike := effectstest.SthB2Spike(t, g, "Spike Colony", "Creature — Spike", sthB2SpikeColonyOracle, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DcMana(t, g, me.ID, "{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, spike, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err == nil {
		t.Error("a Spike with no +1/+1 counter paid the removal cost")
	}
}

func TestSthB2SpikeFeederTradesACounterForTwoLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, spike := effectstest.SthB2Spike(t, g, "Spike Feeder", "Creature — Spike", sthB2SpikeFeederOracle, 2)
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, spike, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Life - life; got != 2 {
		t.Errorf("gained %d life, want 2", got)
	}
	if got := effectstest.CountersOn(g, spike, game.CounterPlusOne); got != 1 {
		t.Errorf("the Feeder has %d counters, want 1", got)
	}
}

func TestSthB2SpikeSoldierTradesACounterForPlusTwoPlusTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, spike := effectstest.SthB2Spike(t, g, "Spike Soldier", "Creature — Spike Soldier", sthB2SpikeSoldierOracle, 3)
	if err := g.ActivateCatalogAbility(me.ID, spike, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, spike); p != 4 || tough != 4 {
		t.Errorf("the Soldier is %d/%d, want 4/4 (two counters and +2/+2)", p, tough)
	}
}

// --- Skyshroud Archer, Sword of the Chosen ------------------------------------------------

func TestSthB2SkyshroudArcherShrinksOnlyAFlier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	archer := effectstest.TmpB1Push(g, me.ID, "Skyshroud Archer", "Creature — Elf Archer", sthB2ArcherOracle, 1, 1, "G")
	flier := sthB2Keyworded(g, opp.ID, "Bird", "Creature — Bird", 2, 2, "flying")
	walker := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, archer, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(walker)}); err == nil {
		t.Fatal("a creature without flying was accepted as the target")
	}
	if err := g.ActivateCatalogAbility(me.ID, archer, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(flier)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, flier); p != 1 || tough != 1 {
		t.Errorf("the flier is %d/%d, want 1/1", p, tough)
	}
	if !effectstest.IsTapped(g, archer) {
		t.Error("the Archer did not tap")
	}
}

func TestSthB2SwordOfTheChosenPumpsOnlyALegend(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	sword := effectstest.TmpB1Push(g, me.ID, "Sword of the Chosen", "Legendary Artifact", sthB2SwordOracle, 0, 0)
	legend := effectstest.B12Creature(g, me.ID, "Legend", "Legendary Creature — Human", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, sword, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Fatal("a nonlegendary creature was accepted as the target")
	}
	if err := g.ActivateCatalogAbility(me.ID, sword, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(legend)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, legend); p != 4 || tough != 4 {
		t.Errorf("the legend is %d/%d, want 4/4", p, tough)
	}
	if !effectstest.IsTapped(g, sword) {
		t.Error("the Sword did not tap")
	}
}
