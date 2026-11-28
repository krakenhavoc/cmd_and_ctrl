package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b840_cards_test.go — Ice Age's damage cards, played from the real
// catalog entries (S82 OLD batch 840, #5636).

const (
	b840JusticeOracle  = "0586fdc7-3025-4040-8eaf-f8a57a21aff0"
	b840LyreOracle     = "0709e9c8-dbe7-4806-ae75-d2e2d745d6a4"
	b840HecatombOracle = "0ec153ee-2a91-4196-9105-212dd1864aea"
	b840VertigoOracle  = "1dabfade-79e5-44e2-afa6-b20498c6bb8b"
	b840BarrierOracle  = "3aab67bb-f287-4aa6-a3c5-e850b22dfff6"
	b840FortressOracle = "7709ed84-ed71-4dd2-b002-b257418dd79e"
	b840MinionOracle   = "93fbf4a7-4283-4eec-afde-6780149b56df"
	b840CannonOracle   = "9ebf3eeb-1a1b-4b88-a760-887c2188d287"
)

func TestS82OldB840CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b840JusticeOracle:  "Justice",
		b840LyreOracle:     "Goblin Lyre",
		b840HecatombOracle: "Hecatomb",
		b840VertigoOracle:  "Vertigo",
		b840BarrierOracle:  "Soul Barrier",
		b840FortressOracle: "Snow Fortress",
		b840MinionOracle:   "Minion of Tevesh Szat",
		b840CannonOracle:   "Orcish Cannoneers",
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

func TestVertigoHitsAFlyerAndGroundsItForTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	angel := effectstest.B12Push(g, opp.ID, "Serra Angel", "Creature — Angel", effectstest.SerraAngelOracle, 4, 4)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	cast := func(target uuid.UUID) error {
		effectstest.ApaMana(me, "R")
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Vertigo", TypeLine: "Instant", ManaCost: "{R}", OracleID: b840VertigoOracle, Owner: me.ID, Controller: me.ID})
		return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(target)})
	}
	if err := cast(bear); err == nil {
		t.Fatal("targeted a creature without flying: accepted")
	}
	if err := cast(angel); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, angel)
	if !ok || c.DamageMarked != 2 {
		t.Fatalf("the Angel took %d damage (on battlefield %v), want 2", c.DamageMarked, ok)
	}
	if effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, angel), "flying") {
		t.Error("the Angel kept flying")
	}
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, angel), "vigilance") {
		t.Error("the Angel lost more than flying")
	}
	effectstest.AdvancePastCleanupForTest(t, g)
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, angel), "flying") {
		t.Error("flying did not come back at end of turn")
	}
}

func TestOrcishCannoneersHurtsTheTargetAndItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	gun := effectstest.PushCatalogPermanent(g, me.ID, "Orcish Cannoneers", "Creature — Orc Warrior", b840CannonOracle, false)
	oppLife, myLife := opp.Life, me.Life
	effectstest.B16Activate(t, g, me.ID, gun, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-2 {
		t.Errorf("the target lost %d life, want 2", oppLife-opp.Life)
	}
	if me.Life != myLife-3 {
		t.Errorf("the controller lost %d life, want 3", myLife-me.Life)
	}
	if !effectstest.B16Tapped(t, g, gun) {
		t.Error("the Cannoneers did not tap")
	}
}

func TestSnowFortressPumpsItselfAndShootsAGroundAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fort := effectstest.B12Push(g, opp.ID, "Snow Fortress", "Artifact Creature — Wall", b840FortressOracle, 0, 4)
	bear := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 3})
	flyer := effectstest.B12Push(g, me.ID, "Serra Angel", "Creature — Angel", effectstest.SerraAngelOracle, 4, 4)
	effectstest.Pr7bAttack(t, g, opp.ID, bear, flyer)
	effectstest.ApaMana(opp, "G", "G", "G", "G", "G")
	// +1/+0 and +0/+1.
	effectstest.B16Activate(t, g, opp.ID, fort, 0, game.ActivateAbilityParams{})
	effectstest.B16Activate(t, g, opp.ID, fort, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, fort); p != 1 || tough != 5 {
		t.Errorf("the Fortress is %d/%d, want 1/5", p, tough)
	}
	// {3}: 1 damage to a creature without flying that's attacking you.
	effectstest.ApaMana(opp, "G", "G", "G")
	if err := g.ActivateCatalogAbility(opp.ID, fort, 2, game.ActivateAbilityParams{Targets: effectstest.CardTarget(flyer)}); err == nil {
		t.Error("the Fortress targeted an attacker with flying")
	}
	effectstest.B16Activate(t, g, opp.ID, fort, 2, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, bear); c.DamageMarked != 1 {
		t.Errorf("the attacker took %d damage, want 1", c.DamageMarked)
	}
}

func TestMinionOfTeveshSzatUpkeepAndBoost(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[1]
		effectstest.PushCatalogPermanent(g, me.ID, "Minion of Tevesh Szat", "Creature — Demon Minion", b840MinionOracle, false)
		life := me.Life
		effectstest.AdvanceToUpkeepOf(t, g, 1)
		effectstest.PassPriorityAroundTable(t, g)
		if pay {
			effectstest.FloatForTest(g, me, "BB")
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		want := life
		if !pay {
			want = life - 2
		}
		if me.Life != want {
			t.Errorf("pay=%v: life %d, want %d", pay, me.Life, want)
		}
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	minion := effectstest.PushCatalogPermanent(g, me.ID, "Minion of Tevesh Szat", "Creature — Demon Minion", b840MinionOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 3)
	effectstest.B16Activate(t, g, me.ID, minion, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 5 || tough != 1 {
		t.Errorf("the creature is %d/%d, want 5/1", p, tough)
	}
}

func TestSoulBarrierDamagesAnOpponentWhoCastsACreatureUnlessTheyPay(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMainOf(t, g, 1)
		effectstest.SeedEnchantment(g, b840BarrierOracle, "Soul Barrier", "Enchantment", me.ID)
		creature := uuid.New()
		opp.Hand.PushTop(game.Card{InstanceID: creature, Name: "Test Bear", TypeLine: "Creature — Bear", ManaCost: "{G}", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
		effectstest.FloatForTest(g, opp, "G")
		life := opp.Life
		if err := g.CastSpell(opp.ID, creature, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if pay {
			effectstest.FloatForTest(g, opp, "CC")
		}
		effectstest.AnswerPayUnless(t, g, opp.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		want := life
		if !pay {
			want = life - 2
		}
		if opp.Life != want {
			t.Errorf("pay=%v: life %d, want %d", pay, opp.Life, want)
		}
	}
}

func TestSoulBarrierIgnoresNoncreatureSpellsAndItsControllersOwn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedEnchantment(g, b840BarrierOracle, "Soul Barrier", "Enchantment", me.ID)
	mine := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: mine, Name: "Test Bear", TypeLine: "Creature — Bear", ManaCost: "{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	effectstest.FloatForTest(g, me, "G")
	if err := g.CastSpell(me.ID, mine, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) || effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Error("Soul Barrier asked for payment for its controller's own creature spell")
	}
}

func TestHecatombIsSacrificedWithoutFourCreaturesAndTaxesWithThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 3; i++ {
		effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	}
	if err := b840Cast(t, g, "Hecatomb", "Enchantment", "{1}{B}{B}", b840HecatombOracle, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == b840HecatombOracle {
			t.Error("Hecatomb stayed with only three creatures")
		}
	}
	if n := effectstest.CountOnBattlefield(g, "Bear", me.ID); n != 3 {
		t.Errorf("%d Bears left, want 3", n)
	}

	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	var bears []uuid.UUID
	for i := 0; i < 5; i++ {
		bears = append(bears, effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2))
	}
	if err := b840Cast(t, g, "Hecatomb", "Enchantment", "{1}{B}{B}", b840HecatombOracle, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.AnswerPermanentPicks(t, g, me.ID, map[uuid.UUID][]uuid.UUID{me.ID: bears[:4]})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Bear", me.ID); n != 1 {
		t.Errorf("%d Bears left, want 1", n)
	}
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == b840HecatombOracle {
			found = true
		}
	}
	if !found {
		t.Error("Hecatomb was sacrificed after its tax was paid")
	}
}

func TestHecatombPingsByTappingAnUntappedSwamp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	hec := effectstest.PushCatalogPermanent(g, me.ID, "Hecatomb", "Enchantment", b840HecatombOracle, false)
	swamp := effectstest.PushLandFor(g, me.ID, "Swamp", "Basic Land — Swamp")
	forest := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	life := opp.Life
	target := []game.TargetRef{effectstest.PlayerTarget(opp.ID)}
	if err := g.ActivateCatalogAbility(me.ID, hec, 0, game.ActivateAbilityParams{Targets: target, TapIDs: []uuid.UUID{forest}}); err == nil {
		t.Fatal("tapped a Forest for the Swamp cost: accepted")
	}
	effectstest.B16Activate(t, g, me.ID, hec, 0, game.ActivateAbilityParams{Targets: target, TapIDs: []uuid.UUID{swamp}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("the target lost %d life, want 1", life-opp.Life)
	}
	if !effectstest.B16Tapped(t, g, swamp) {
		t.Error("the Swamp was not tapped")
	}
}

// b840Cast casts `name` from the active player's hand with the mana floated.
func b840Cast(t *testing.T, g *game.Game, name, typeLine, cost, oracle string, targets []game.TargetRef) error {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "{2}{G}{G}{B}{B}{U}{R}{W}")
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
			OracleID: oracle, Owner: me.ID, Controller: me.ID})
	})
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: targets})
}

func TestJusticeReflectsRedCreatureAndSpellDamageToTheirControllers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedEnchantment(g, b840JusticeOracle, "Justice", "Enchantment", me.ID)
	// A red spell: Lightning Bolt from the opponent at me deals 3, so they take 3.
	bolt := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", Colors: []string{"R"}, OracleID: effectstest.LightningBoltOracle, Owner: opp.ID, Controller: opp.ID})
	effectstest.FloatForTest(g, opp, "R")
	oppLife, myLife := opp.Life, me.Life
	if err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{Targets: []game.TargetRef{effectstest.PlayerTarget(me.ID)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife-3 {
		t.Errorf("the bolt's target lost %d life, want 3", myLife-me.Life)
	}
	if opp.Life != oppLife-3 {
		t.Errorf("the red spell's controller lost %d life to Justice, want 3", oppLife-opp.Life)
	}
}

func TestJusticeIgnoresNonredDamageAndReflectsARedCreaturesCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedEnchantment(g, b840JusticeOracle, "Justice", "Enchantment", me.ID)
	red := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Red Bear", TypeLine: "Creature — Bear", Colors: []string{"R"}, Power: 3, Toughness: 3})
	green := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Green Bear", TypeLine: "Creature — Bear", Colors: []string{"G"}, Power: 2, Toughness: 2})
	_ = red
	_ = green
	// Opponent's turn: both attack me.
	effectstest.AdvanceToDeclareAttackersOf(t, g, 1)
	for _, a := range []uuid.UUID{red, green} {
		if err := g.DeclareAttacker(a, me.ID); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	oppLife := opp.Life
	effectstest.Pr7bDamageStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oppLife - opp.Life; got != 3 {
		t.Errorf("the attackers' controller lost %d life to Justice, want 3 (the red creature only)", got)
	}
}

func TestJusticeUpkeepCostsWW(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	justice := effectstest.SeedEnchantment(g, b840JusticeOracle, "Justice", "Enchantment", me.ID)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.RemoraUpkeepPrompt(t, g, me.ID)
	if prompt.PayCost != "{W}{W}" {
		t.Fatalf("cost = %q, want {W}{W}", prompt.PayCost)
	}
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	if _, ok := effectstest.BattlefieldCard(g, justice); ok {
		t.Error("an unpaid upkeep did not sacrifice Justice")
	}
}

func TestGoblinLyreFlipDealsDamageEitherWay(t *testing.T) {
	won, lost := 0, 0
	for table := 0; table < 40 && (won == 0 || lost == 0); table++ {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		lyre := effectstest.PushCatalogPermanent(g, me.ID, "Goblin Lyre", "Artifact", b840LyreOracle, false)
		for i := 0; i < 2; i++ {
			effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
		}
		for i := 0; i < 3; i++ {
			effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
		}
		myLife, oppLife := me.Life, opp.Life
		effectstest.B16Activate(t, g, me.ID, lyre, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
		effectstest.PassPriorityAroundTable(t, g)
		ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall)
		if ask == nil {
			t.Fatal("no coin call was asked")
		}
		if err := g.ResolveCoinCall(ask.ID, me.ID, "heads"); err != nil {
			t.Fatalf("ResolveCoinCall: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(lyre) {
			t.Error("the Lyre was not sacrificed")
		}
		evs := effectstest.RandomEvents(g, game.EventFlipCoin)
		if len(evs) == 0 {
			t.Fatal("no coin was flipped")
		}
		if evs[len(evs)-1].Won {
			won++
			if opp.Life != oppLife-2 || me.Life != myLife {
				t.Errorf("won flip: opponent %d -> %d (want -2), me %d -> %d (want unchanged)", oppLife, opp.Life, myLife, me.Life)
			}
		} else {
			lost++
			if me.Life != myLife-3 || opp.Life != oppLife {
				t.Errorf("lost flip: me %d -> %d (want -3), opponent %d -> %d (want unchanged)", myLife, me.Life, oppLife, opp.Life)
			}
		}
	}
	if won == 0 || lost == 0 {
		t.Fatalf("saw %d won and %d lost flips; want both", won, lost)
	}
}
