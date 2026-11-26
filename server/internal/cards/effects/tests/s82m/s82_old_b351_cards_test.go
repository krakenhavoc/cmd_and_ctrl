package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b351_cards_test.go — Khans of Tarkir, batch 351 (#5554). Each test
// plays the real catalog entry and asserts what its text says.

const (
	oldB351Scaldkin    = "3fb2dbb3-e064-44bf-a8b6-10d58b1174d5"
	oldB351Mindswipe   = "4682c39e-c21a-4c8b-bc63-9cbac59dc563"
	oldB351Winterflame = "8d3a0b81-1895-47a8-87a3-3e4093cbd951"
	oldB351Windstorm   = "9969b822-2ead-4956-bac2-875ffd90e872"
	oldB351Firehoof    = "487de0df-0496-4b5d-8050-27b8e617582b"
	oldB351Knuckleblde = "5a4fabfe-a641-49b7-8134-9d9edeb8bf1b"
	oldB351Scion       = "9d49cef5-8f9c-46cd-aa5e-7861fc1dcaa4"
	oldB351PearlLake   = "b7ae125c-955b-4cfd-9aeb-e6572c0d4d3d"
	oldB351Goblinslide = "4885354f-0b77-47b6-b8f7-00753804436f"
	oldB351Herald      = "889d1056-4d3d-42f0-96f2-16529a4a94e3"
	oldB351Ascendancy  = "aa1de5ce-ddd4-4c1f-8c52-0311f4419e7b"
	oldB351Rite        = "f009ef70-d718-45cf-bb22-e560484bbfe5"
)

func TestOldB351CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB351Scaldkin: "Scaldkin", oldB351Mindswipe: "Mindswipe", oldB351Winterflame: "Winterflame", oldB351Windstorm: "Windstorm",
		oldB351Firehoof: "Firehoof Cavalry", oldB351Knuckleblde: "Savage Knuckleblade", oldB351Scion: "Scion of Glaciers",
		oldB351PearlLake: "Pearl Lake Ancient", oldB351Goblinslide: "Goblinslide", oldB351Herald: "Herald of Anafenza",
		oldB351Ascendancy: "Mardu Ascendancy", oldB351Rite: "Rite of the Serpent",
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

func oldB351Table(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	return g, me, opp
}

// oldB351Cast puts the card in the active player's hand, floats the mana and casts it strictly.
func oldB351Cast(t *testing.T, g *game.Game, me *game.Player, card game.Card, mana string, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	card.InstanceID, card.Owner, card.Controller = uuid.New(), me.ID, me.ID
	me.Hand.PushTop(card)
	if mana != "" {
		if err := g.AddManaForEffect(me.ID, uuid.Nil, mana); err != nil {
			t.Fatalf("AddManaForEffect: %v", err)
		}
	}
	params.Strict = true
	return card.InstanceID, g.CastSpell(me.ID, card.InstanceID, params)
}

func oldB351Activate(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, row int, mana string, params game.ActivateAbilityParams) error {
	t.Helper()
	effectstest.FloatForTest(g, me, mana)
	params.Strict = true
	return g.ActivateCatalogAbility(me.ID, id, row, params)
}

func TestOldB351ScaldkinSacrificesItselfToDealTwoToAnyTarget(t *testing.T) {
	g, me, opp := oldB351Table(t)
	id := effectstest.PoolPush(g, me.ID, "Scaldkin", oldB351Scaldkin, 2, 2)
	before := opp.Life
	if err := oldB351Activate(t, g, me, id, 0, "RRR", game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if effectstest.BattlefieldHas(g, id) {
		t.Error("the Scaldkin was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("opponent life %d, want %d", opp.Life, before-2)
	}
}

func TestOldB351MindswipeCountersUnlessXIsPaidAndAlwaysDealsX(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g, me, _ := oldB351Table(t)
		spell, err := oldB351Cast(t, g, me, game.Card{Name: "A Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"}, "{C}{G}", game.CastSpellParams{})
		if err != nil {
			t.Fatalf("cast the bear: %v", err)
		}
		for i := 0; i < 2; i++ {
			// Two {C} left over to pay {X} with.
			if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
				t.Fatalf("AddManaForEffect: %v", err)
			}
		}
		life := me.Life
		if _, err := oldB351Cast(t, g, me, game.Card{Name: "Mindswipe", TypeLine: "Instant", ManaCost: "{X}{U}{R}", OracleID: oldB351Mindswipe},
			"{C}{C}{U}{R}", game.CastSpellParams{Targets: effectstest.CardTarget(spell), XValue: 2}); err != nil {
			t.Fatalf("cast Mindswipe: %v", err)
		}
		effectstest.PassUntilTaxed(t, g, me.ID)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.BattlefieldHas(g, spell); got != pay {
			t.Errorf("pay=%v: the bear is on the battlefield = %v, want %v", pay, got, pay)
		}
		if me.Life != life-2 {
			t.Errorf("pay=%v: the spell's controller is at %d, want %d (X damage either way)", pay, me.Life, life-2)
		}
	}
}

func TestOldB351WinterflameChoosesOneOrBoth(t *testing.T) {
	g, me, opp := oldB351Table(t)
	tapTarget := effectstest.B12Creature(g, opp.ID, "Tap Target", "Creature — Bear", 2, 3)
	hitTarget := effectstest.B12Creature(g, opp.ID, "Hit Target", "Creature — Bear", 2, 2)
	if _, err := oldB351Cast(t, g, me, game.Card{Name: "Winterflame", TypeLine: "Instant", ManaCost: "{1}{U}{R}", OracleID: oldB351Winterflame},
		"{C}{U}{R}", game.CastSpellParams{Modes: []int{0, 1}, Targets: effectstest.CardTargets(tapTarget, hitTarget)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.PermanentTapped(g, tapTarget) {
		t.Error("the first bullet did not tap its target")
	}
	if effectstest.BattlefieldHas(g, hitTarget) {
		t.Error("the second bullet did not deal 2 damage to its target")
	}
}

func TestOldB351WinterflameDamageOnlyBullet(t *testing.T) {
	g, me, opp := oldB351Table(t)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	if _, err := oldB351Cast(t, g, me, game.Card{Name: "Winterflame", TypeLine: "Instant", ManaCost: "{1}{U}{R}", OracleID: oldB351Winterflame},
		"{C}{U}{R}", game.CastSpellParams{Modes: []int{1}, Targets: effectstest.CardTarget(victim)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, victim) {
		t.Error("the creature survived 2 damage")
	}
}

func TestOldB351WindstormDamagesEachFlyerIncludingYours(t *testing.T) {
	g, me, opp := oldB351Table(t)
	flyer := effectstest.AcPermanent(g, opp.ID, "Their Flyer", "Creature — Bird", 2, 3)
	mine := effectstest.AcPermanent(g, me.ID, "My Flyer", "Creature — Bird", 2, 3)
	grounded := effectstest.AcPermanent(g, opp.ID, "Their Bear", "Creature — Bear", 2, 3)
	for _, id := range []uuid.UUID{flyer, mine} {
		g.WithWriteLock(func() {
			effectstest.FindBattlefieldCardForTest(g, id).Keywords = append(effectstest.FindBattlefieldCardForTest(g, id).Keywords, "flying")
		})
	}
	if _, err := oldB351Cast(t, g, me, game.Card{Name: "Windstorm", TypeLine: "Instant", ManaCost: "{X}{G}", OracleID: oldB351Windstorm},
		"{C}{C}{C}{G}", game.CastSpellParams{XValue: 3}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, flyer) || effectstest.BattlefieldHas(g, mine) {
		t.Error("a creature with flying survived 3 damage")
	}
	if !effectstest.BattlefieldHas(g, grounded) {
		t.Error("a creature without flying died")
	}
}

func TestOldB351FirehoofCavalryPumpsAndGainsTrample(t *testing.T) {
	g, me, _ := oldB351Table(t)
	id := effectstest.PoolPush(g, me.ID, "Firehoof Cavalry", oldB351Firehoof, 1, 1)
	if err := oldB351Activate(t, g, me, id, 0, "RRRR", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, id); p != 3 {
		t.Errorf("power %d, want 3", p)
	}
	if !effectstest.AvB22Has(t, g, id, "trample") {
		t.Error("no trample")
	}
	effectstest.AdvanceTo(t, g, game.StepUpkeep)
	if p := effectstest.EffectivePower(t, g, id); p != 1 || effectstest.AvB22Has(t, g, id, "trample") {
		t.Errorf("after the turn: power %d, trample %v, want 1 and none", p, effectstest.AvB22Has(t, g, id, "trample"))
	}
}

func TestOldB351KnuckleBladeRowsPumpOncePerTurnBounceAndHaste(t *testing.T) {
	g, me, _ := oldB351Table(t)
	id := effectstest.PoolPush(g, me.ID, "Savage Knuckleblade", oldB351Knuckleblde, 4, 4)
	if err := oldB351Activate(t, g, me, id, 0, "GGG", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("pump: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, id); p != 6 {
		t.Errorf("power %d after the pump, want 6", p)
	}
	if err := oldB351Activate(t, g, me, id, 0, "GGG", game.ActivateAbilityParams{}); err == nil {
		t.Error("the pump was activated twice in one turn")
	}
	g.WithWriteLock(func() { me.ManaPool = nil })
	if err := oldB351Activate(t, g, me, id, 2, "R", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("haste: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.AvB22Has(t, g, id, "haste") {
		t.Error("no haste")
	}
	if err := oldB351Activate(t, g, me, id, 1, "UUU", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("bounce: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, id) {
		t.Error("the Knuckleblade is still on the battlefield")
	}
}

func TestOldB351ScionOfGlaciersShrinksItsToughnessEachTime(t *testing.T) {
	g, me, _ := oldB351Table(t)
	id := effectstest.PoolPush(g, me.ID, "Scion of Glaciers", oldB351Scion, 2, 5)
	for i := 0; i < 2; i++ {
		if err := oldB351Activate(t, g, me, id, 0, "U", game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if p, tough := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id); p != 4 || tough != 3 {
		t.Errorf("Scion is %d/%d, want 4/3", p, tough)
	}
}

func TestOldB351PearlLakeAncientReturnsThreeLandsToReturnItself(t *testing.T) {
	g, me, _ := oldB351Table(t)
	id := effectstest.PoolPush(g, me.ID, "Pearl Lake Ancient", oldB351PearlLake, 6, 7)
	lands := effectstest.AcLands(g, me.ID, 4, "Island")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, ReturnIDs: lands[:2]}); err == nil {
		t.Error("two lands paid a cost of three")
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, ReturnIDs: lands[:3]}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, id) {
		t.Error("the Ancient is still on the battlefield")
	}
	for i, l := range lands {
		if got := effectstest.BattlefieldHas(g, l); got != (i == 3) {
			t.Errorf("land %d on the battlefield = %v", i, got)
		}
	}
	if spec, _ := effects.Lookup(oldB351PearlLake); !spec.CantBeCountered {
		t.Error("the Ancient's spell can be countered")
	}
}

func TestOldB351GoblinslidePaysOneForAHastyGoblin(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g, me, _ := oldB351Table(t)
		effectstest.PushCatalogPermanent(g, me.ID, "Goblinslide", "Enchantment", oldB351Goblinslide, false)
		if _, err := oldB351Cast(t, g, me, game.Card{Name: "Divination", TypeLine: "Sorcery", ManaCost: "{2}{U}"}, "{C}{C}{U}", game.CastSpellParams{}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
			t.Fatalf("AddManaForEffect: %v", err)
		}
		effectstest.PassUntilTaxed(t, g, me.ID)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		n, _ := effectstest.CountTokensNamed(g, me.ID, "Goblin")
		want := 0
		if pay {
			want = 1
		}
		if n != want {
			t.Errorf("pay=%v: %d Goblin tokens, want %d", pay, n, want)
		}
	}
}

func TestOldB351GoblinslideIgnoresCreatureSpells(t *testing.T) {
	g, me, _ := oldB351Table(t)
	effectstest.PushCatalogPermanent(g, me.ID, "Goblinslide", "Enchantment", oldB351Goblinslide, false)
	if _, err := oldB351Cast(t, g, me, game.Card{Name: "A Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"}, "{C}{G}", game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("a creature spell asked for the {1}")
	}
}

func TestOldB351HeraldOfAnafenzaMakesAWarriorWhenYouActivateOutlast(t *testing.T) {
	g, me, _ := oldB351Table(t)
	id := effectstest.PoolPush(g, me.ID, "Herald of Anafenza", oldB351Herald, 1, 2)
	if err := oldB351Activate(t, g, me, id, 0, "WWW", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("outlast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Warrior"); n != 1 {
		t.Errorf("%d Warrior tokens, want 1", n)
	}
	if got := effectstest.FindBattlefieldCardForTest(g, id).Counters[game.CounterPlusOne]; got != 1 {
		t.Errorf("the Herald has %d +1/+1 counters, want 1", got)
	}
}

func TestOldB351AscendancyMakesATappedAttackingGoblinPerNontokenAttackerAndSacrificesForToughness(t *testing.T) {
	g, me, opp := oldB351Table(t)
	asc := effectstest.PushCatalogPermanent(g, me.ID, "Mardu Ascendancy", "Enchantment", oldB351Ascendancy, false)
	a := effectstest.B12Creature(g, me.ID, "Attacker One", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Attacker Two", "Creature — Bear", 2, 2)
	token := effectstest.PushToken(g, me.ID, effects.TokenCard("1/1 white Warrior"))
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.DeclareAttack(t, g, opp.ID, a, b, token)
	effectstest.PassPriorityAroundTable(t, g)
	n, tapped := effectstest.CountTokensNamed(g, me.ID, "Goblin")
	if n != 2 || tapped != 2 {
		t.Fatalf("%d Goblin tokens (%d tapped), want 2 tapped (one per nontoken attacker, none for the token)", n, tapped)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Goblin" && c.AttackingTarget != opp.ID {
			t.Errorf("a Goblin is attacking %s, want the player the creature attacked", c.AttackingTarget)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, asc, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, asc) {
		t.Error("the Ascendancy was not sacrificed")
	}
	if tough := effectstest.EffectiveToughness(t, g, a); tough != 5 {
		t.Errorf("an attacker has toughness %d, want 5", tough)
	}
}

func TestOldB351RiteOfTheSerpentDestroysAndMakesASnakeOnlyForACounteredCreature(t *testing.T) {
	for _, counter := range []bool{false, true} {
		g, me, opp := oldB351Table(t)
		victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
		if counter {
			effectstest.PfAddCounters(g, victim, "+1/+1", 1)
		}
		if _, err := oldB351Cast(t, g, me, game.Card{Name: "Rite of the Serpent", TypeLine: "Sorcery", ManaCost: "{4}{B}{B}", OracleID: oldB351Rite},
			"{C}{C}{C}{C}{B}{B}", game.CastSpellParams{Targets: effectstest.CardTarget(victim)}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.BattlefieldHas(g, victim) {
			t.Errorf("counter=%v: the creature survived", counter)
		}
		n, _ := effectstest.CountTokensNamed(g, me.ID, "Snake")
		want := 0
		if counter {
			want = 1
		}
		if n != want {
			t.Errorf("counter=%v: %d Snake tokens, want %d", counter, n, want)
		}
	}
}
