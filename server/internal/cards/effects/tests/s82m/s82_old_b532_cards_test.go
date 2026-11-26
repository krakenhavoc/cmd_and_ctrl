package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b532_cards_test.go — Shadowmoor, batch 532 (S82, #5591): -1/-1 and
// other counters. Each test plays the real catalog entry through the engine.

const (
	oldB532ChainbreakerOracle = "02765930-e15e-4cc1-ac0d-040334a9c5bc"
	oldB532WarcrawlerOracle   = "0acadb47-17b2-47d8-9aad-5436931bc38c"
	oldB532MedicOracle        = "17661062-928c-4e33-a4f3-92ddc9d8c817"
	oldB532RitualOracle       = "367ec7c5-8c76-4af6-b941-a3a8c2ff67ff"
	oldB532BonderOracle       = "3f03d0e6-c7c0-484c-bb65-b8cbf0bd802b"
	oldB532EffigyOracle       = "5fa3cc1a-09ee-4152-8c50-ea719510fc20"
	oldB532MorselOracle       = "8fe687aa-d134-4a29-9c93-64086c4a1977"
	oldB532TyrantOracle       = "ce855b35-4a31-4e42-9dbc-fea8d244fbe0"
	oldB532WretchOracle       = "df6d9787-b423-4067-adf4-c09dcee1329e"
)

func oldB532Table(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	return g, me, opp
}

func oldB532Minus(g *game.Game, id uuid.UUID) int {
	return effectstest.CountersOn(g, id, game.CounterMinusOne)
}

func oldB532AddMinus(g *game.Game, id uuid.UUID, n int) {
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, game.CounterMinusOne, n) })
}

func TestOldB532EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB532ChainbreakerOracle: "Chainbreaker", oldB532WarcrawlerOracle: "Wicker Warcrawler",
		oldB532MedicOracle: "Barrenton Medic", oldB532RitualOracle: "Scarscale Ritual",
		oldB532BonderOracle: "Leech Bonder", oldB532EffigyOracle: "Gnarled Effigy",
		oldB532MorselOracle: "Morselhoarder", oldB532TyrantOracle: "Grief Tyrant",
		oldB532WretchOracle: "Cinderhaze Wretch",
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

func TestOldB532CreaturesEnterWithTheirMinusCounters(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness, n    int
	}{
		{"Chainbreaker", "Artifact Creature — Scarecrow", oldB532ChainbreakerOracle, 3, 3, 2},
		{"Leech Bonder", "Creature — Merfolk Soldier", oldB532BonderOracle, 3, 3, 2},
		{"Morselhoarder", "Creature — Elemental", oldB532MorselOracle, 6, 4, 2},
		{"Grief Tyrant", "Creature — Horror", oldB532TyrantOracle, 8, 8, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := oldB532Table(t)
			effectstest.AdvanceToMain(t, g)
			id := uuid.New()
			me.Hand.PushTop(game.Card{InstanceID: id, Name: tc.name, TypeLine: tc.typeLine, OracleID: tc.oracle,
				Power: tc.power, Toughness: tc.toughness, Owner: me.ID, Controller: me.ID})
			if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
				t.Fatalf("cast: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if n := oldB532Minus(g, id); n != tc.n {
				t.Errorf("%d -1/-1 counters on entry, want %d", n, tc.n)
			}
			if p, tough := effectstest.PtOf(t, g, id); p != tc.power-tc.n || tough != tc.toughness-tc.n {
				t.Errorf("%d/%d on entry, want %d/%d", p, tough, tc.power-tc.n, tc.toughness-tc.n)
			}
		})
	}
}

func TestOldB532ChainbreakerRemovesAMinusCounterFromAnyCreature(t *testing.T) {
	g, me, opp := oldB532Table(t)
	cb := effectstest.B12Push(g, me.ID, "Chainbreaker", "Artifact Creature — Scarecrow", oldB532ChainbreakerOracle, 3, 3)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 3, 3)
	oldB532AddMinus(g, bear, 2)
	effectstest.B10AddMana(me, "C", "C", "C")
	effectstest.B16Activate(t, g, me.ID, cb, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if n := oldB532Minus(g, bear); n != 1 {
		t.Errorf("-1/-1 counters on the target %d, want 1", n)
	}
	if !effectstest.B16Tapped(t, g, cb) {
		t.Error("the Chainbreaker did not tap")
	}
}

func TestOldB532WickerWarcrawlerShrinksAtEndOfCombat(t *testing.T) {
	for _, how := range []string{"attacks", "blocks"} {
		t.Run(how, func(t *testing.T) {
			g, me, opp := oldB532Table(t)
			crawler := effectstest.B12Push(g, me.ID, "Wicker Warcrawler", "Artifact Creature — Scarecrow", oldB532WarcrawlerOracle, 6, 6)
			if how == "attacks" {
				effectstest.DeclareAttack(t, g, opp.ID, crawler)
			} else {
				attacker := effectstest.B12Creature(g, opp.ID, "Attacker", "Creature — Bear", 2, 2)
				g.WithWriteLock(func() { g.Turn.ActiveSeat = opp.Seat })
				effectstest.DeclareAttack(t, g, me.ID, attacker)
				effectstest.WoeS2Block(g, me.ID, crawler, attacker)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if n := oldB532Minus(g, crawler); n != 0 {
				t.Fatalf("%d counters before the end of combat step, want 0", n)
			}
			effectstest.AdvanceTo(t, g, game.StepEndCombat)
			effectstest.PassPriorityAroundTable(t, g)
			if n := oldB532Minus(g, crawler); n != 1 {
				t.Errorf("%d -1/-1 counters after end of combat, want 1", n)
			}
		})
	}
}

func TestOldB532MorselhoarderTurnsACounterIntoManaOfAnyColor(t *testing.T) {
	g, me, _ := oldB532Table(t)
	mh := effectstest.B12Push(g, me.ID, "Morselhoarder", "Creature — Elemental", oldB532MorselOracle, 6, 4)
	if err := g.ActivateManaAbility(me.ID, mh, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("the mana ability was paid with no -1/-1 counter on the creature")
	}
	oldB532AddMinus(g, mh, 2)
	if err := g.ActivateManaAbility(me.ID, mh, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if n := oldB532Minus(g, mh); n != 1 {
		t.Errorf("-1/-1 counters left %d, want 1 (one is the cost)", n)
	}
	if got := me.ManaPool.Total(); got != 1 {
		t.Errorf("mana in the pool %d, want 1", got)
	}
}

func TestOldB532CinderhazeWretchMakesYouDiscardOnlyOnYourTurn(t *testing.T) {
	g, me, opp := oldB532Table(t)
	wretch := effectstest.B12Push(g, me.ID, "Cinderhaze Wretch", "Creature — Elemental Shaman", oldB532WretchOracle, 3, 2)
	card := effectstest.PushHandCardOfType(opp, "Victim Card", "Sorcery")
	before := opp.Hand.Size()
	params := game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}
	g.WithWriteLock(func() { g.Turn.ActiveSeat = opp.Seat })
	if err := g.ActivateCatalogAbility(me.ID, wretch, 0, params); err == nil {
		t.Fatal("activated during an opponent's turn")
	}
	g.WithWriteLock(func() { g.Turn.ActiveSeat = me.Seat })
	effectstest.B16Activate(t, g, me.ID, wretch, 0, params)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, opp.ID, card)
	if opp.Hand.Size() != before-1 {
		t.Errorf("hand %d → %d, want one discarded", before, opp.Hand.Size())
	}
}

func TestOldB532UntappersPayACounterAndUntapEvenWhileTapped(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness       int
	}{
		{"Cinderhaze Wretch", "Creature — Elemental Shaman", oldB532WretchOracle, 3, 2},
		{"Barrenton Medic", "Creature — Kithkin Cleric", oldB532MedicOracle, 0, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := oldB532Table(t)
			id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			g.WithWriteLock(func() { _ = g.TapTargetForEffect(id) })
			effectstest.B16Activate(t, g, me.ID, id, 1, game.ActivateAbilityParams{Strict: true})
			effectstest.PassPriorityAroundTable(t, g)
			if effectstest.B16Tapped(t, g, id) {
				t.Error("still tapped after the untapper resolved")
			}
			if n := oldB532Minus(g, id); n != 1 {
				t.Errorf("%d -1/-1 counters, want 1 (the cost)", n)
			}
		})
	}
}

func TestOldB532BarrentonMedicPreventsTheNextOneDamage(t *testing.T) {
	g, me, _ := oldB532Table(t)
	medic := effectstest.B12Push(g, me.ID, "Barrenton Medic", "Creature — Kithkin Cleric", oldB532MedicOracle, 0, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 4)
	effectstest.B16Activate(t, g, me.ID, medic, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 3) })
	if got := effectstest.DamageMarkedOn(g, bear); got != 2 {
		t.Errorf("3 damage into a 1-shield: marked %d, want 2", got)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 1) })
	if got := effectstest.DamageMarkedOn(g, bear); got != 3 {
		t.Errorf("the shield is spent: marked %d, want 3", got)
	}
}

func TestOldB532GnarledEffigyPutsAMinusCounterOnATarget(t *testing.T) {
	g, me, opp := oldB532Table(t)
	effigy := effectstest.B12Push(g, me.ID, "Gnarled Effigy", "Artifact", oldB532EffigyOracle, 0, 0)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B10AddMana(me, "C", "C", "C", "C")
	effectstest.B16Activate(t, g, me.ID, effigy, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if n := oldB532Minus(g, bear); n != 1 {
		t.Errorf("-1/-1 counters on the target %d, want 1", n)
	}
	if p, tough := effectstest.PtOf(t, g, bear); p != 1 || tough != 1 {
		t.Errorf("the target is %d/%d, want 1/1", p, tough)
	}
}

func TestOldB532GriefTyrantSpreadsItsCountersWhenItDies(t *testing.T) {
	g, me, opp := oldB532Table(t)
	tyrant := effectstest.B12Push(g, me.ID, "Grief Tyrant", "Creature — Horror", oldB532TyrantOracle, 8, 8)
	oldB532AddMinus(g, tyrant, 3)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 5, 5)
	effectstest.MoveToGraveyard(t, g, me.ID, tyrant)
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := oldB532Minus(g, bear); n != 3 {
		t.Errorf("-1/-1 counters on the target %d, want 3 (one for each the Tyrant had)", n)
	}
}

func TestOldB532LeechBonderMovesACounterWhenTapped(t *testing.T) {
	g, me, opp := oldB532Table(t)
	bonder := effectstest.B12Push(g, me.ID, "Leech Bonder", "Creature — Merfolk Soldier", oldB532BonderOracle, 3, 3)
	from := effectstest.B12Creature(g, opp.ID, "From", "Creature — Bear", 4, 4)
	to := effectstest.B12Creature(g, opp.ID, "To", "Creature — Bear", 4, 4)
	oldB532AddMinus(g, from, 2)
	effectstest.B10AddMana(me, "U")
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: from, Slot: 0}, {Kind: game.TargetCard, ID: to, Slot: 1}}
	params := game.ActivateAbilityParams{Strict: true, Targets: targets}
	if err := g.ActivateCatalogAbility(me.ID, bonder, 0, params); err == nil {
		t.Fatal("{Q} was paid by an untapped creature")
	}
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(bonder) })
	effectstest.B16Activate(t, g, me.ID, bonder, 0, params)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB532Minus(g, from) != 1 || oldB532Minus(g, to) != 1 {
		t.Errorf("counters from %d, to %d; want one moved (1 and 1)", oldB532Minus(g, from), oldB532Minus(g, to))
	}
	if effectstest.B16Tapped(t, g, bonder) {
		t.Error("the Bonder is still tapped after paying {Q}")
	}
}

func TestOldB532ScarscaleRitualNeedsACounterOnACreatureAndDrawsTwo(t *testing.T) {
	g, me, _ := oldB532Table(t)
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Scarscale Ritual", TypeLine: "Sorcery", ManaCost: "{1}{U/B}",
		OracleID: oldB532RitualOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast with no creature to put the counter on")
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FlashIfFund(g, me, "U", "C")
	hand := me.Hand.Size() - 1
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, BlightIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if n := oldB532Minus(g, bear); n != 1 {
		t.Errorf("-1/-1 counters on the creature %d, want 1 (the cost)", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d, want %d (drew two)", me.Hand.Size(), hand+2)
	}
}
