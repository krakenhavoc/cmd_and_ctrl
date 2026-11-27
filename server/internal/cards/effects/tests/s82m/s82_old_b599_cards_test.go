package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b599_cards_test.go — Betrayers of Kamigawa, counters and
// keywords, batch 599 (#5602, S82 OLD). Every test plays the real catalog
// entry through the engine.

const (
	oldB599Skullmane = "74466fa6-34d9-42f9-8590-c7f1fa672bf5"
	oldB599Altar     = "8443b8c2-d0ab-4fb1-ad35-7e5e4cb345b6"
	oldB599Blademane = "9ca64b8e-e9d2-41b3-b229-6fcdc48a8371"
	oldB599Petalmane = "f73f8eca-55fe-418f-a244-29ef82cfc47c"
	oldB599Monstrous = "01906aae-fc53-40b0-a386-dae43e4f437f"
	oldB599Glasskite = "0a661f7a-5d56-4d73-bab7-22bfc5cef22e"
	oldB599Kaijin    = "cd8bab2c-b04f-43a2-9786-df5f7411bf16"
	oldB599Chisei    = "f91407fe-3c3b-4ffc-81de-cb8c5b0fe526"
)

func TestOldB599CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB599Skullmane: "Skullmane Baku", oldB599Altar: "Baku Altar", oldB599Blademane: "Blademane Baku",
		oldB599Petalmane: "Petalmane Baku", oldB599Monstrous: "Ashen Monstrosity", oldB599Glasskite: "Shimmering Glasskite",
		oldB599Kaijin: "Kaijin of the Vanishing Touch", oldB599Chisei: "Chisei, Heart of Oceans",
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

func TestOldB599PrintedKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Ashen Monstrosity", "Creature — Spirit", oldB599Monstrous, []string{"haste"}},
		{"Shimmering Glasskite", "Creature — Spirit", oldB599Glasskite, []string{"flying"}},
		{"Kaijin of the Vanishing Touch", "Creature — Spirit", oldB599Kaijin, []string{"defender"}},
		{"Chisei, Heart of Oceans", "Legendary Creature — Spirit", oldB599Chisei, []string{"flying"}},
	} {
		g := effectstest.NewCatalogGame(t)
		id := effectstest.B12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
		for _, kw := range tc.keywords {
			if !effectstest.HasKeywordOnBattlefield(t, g, id, kw) {
				t.Errorf("%s lacks %s", tc.name, kw)
			}
		}
	}
}

func oldB599Table(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	return g, g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func TestOldB599BakuPutAKiCounterOnASpiritOrArcaneSpell(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
	}{
		{"Skullmane Baku", "Creature — Spirit", oldB599Skullmane},
		{"Blademane Baku", "Creature — Spirit", oldB599Blademane},
		{"Petalmane Baku", "Creature — Spirit", oldB599Petalmane},
		{"Baku Altar", "Artifact", oldB599Altar},
	} {
		g, me, _ := oldB599Table(t)
		baku := effectstest.PushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
		effectstest.CastCatalogSpell(t, g, "Bear Spell", "Creature — Bear", "", nil)
		if effectstest.OwedMays(g, me.ID) != 0 {
			t.Errorf("%s: a Bear spell triggered it", tc.name)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.CastCatalogSpell(t, g, "An Arcane Spell", "Instant — Arcane", "", nil)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
		effectstest.PassPriorityAroundTable(t, g)
		if n := effectstest.Pr8Counters(g, baku, "ki"); n != 1 {
			t.Errorf("%s has %d ki counters after an Arcane spell, want 1", tc.name, n)
		}
		effectstest.CastCatalogSpell(t, g, "A Spirit Spell", "Instant — Spirit", "", nil)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
		effectstest.PassPriorityAroundTable(t, g)
		if n := effectstest.Pr8Counters(g, baku, "ki"); n != 1 {
			t.Errorf("%s has %d ki counters after declining, want 1", tc.name, n)
		}
	}
}

func TestOldB599SkullmaneBakuShrinksATargetByTheCountersRemoved(t *testing.T) {
	g, me, opp := oldB599Table(t)
	baku := effectstest.PushCatalogPermanent(g, me.ID, "Skullmane Baku", "Creature — Spirit", oldB599Skullmane, false)
	effectstest.PfAddCounters(g, baku, "ki", 3)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 4, 4)
	if err := g.ActivateCatalogAbility(me.ID, baku, 0, game.ActivateAbilityParams{Strict: true, CounterCounts: []int{4}, Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Fatal("four ki counters were removed from a creature with three")
	}
	effectstest.B06AddMana(me, "C")
	effectstest.P7Activate(t, g, me, baku, 0, game.ActivateAbilityParams{Strict: true, CounterCounts: []int{2}, Targets: effectstest.CardTarget(bear)})
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("the target is %d/%d, want 2/2", p, tough)
	}
	if n := effectstest.Pr8Counters(g, baku, "ki"); n != 1 {
		t.Errorf("%d ki counters left, want 1", n)
	}
	if !effectstest.IsTapped(g, baku) {
		t.Error("the Baku did not tap")
	}
}

func TestOldB599BlademaneBakuGainsTwoPowerPerCounterRemoved(t *testing.T) {
	g, me, _ := oldB599Table(t)
	baku := effectstest.PushCatalogPermanent(g, me.ID, "Blademane Baku", "Creature — Spirit", oldB599Blademane, false)
	effectstest.PfAddCounters(g, baku, "ki", 3)
	effectstest.B06AddMana(me, "C")
	effectstest.P7Activate(t, g, me, baku, 0, game.ActivateAbilityParams{Strict: true, CounterCounts: []int{3}})
	if p, tough := effectstest.EffectivePTOf(t, g, baku); p != 1+6 || tough != 1 {
		t.Errorf("the Baku is %d/%d, want 7/1", p, tough)
	}
	if effectstest.IsTapped(g, baku) {
		t.Error("the ability tapped the Baku")
	}
	effectstest.AdvancePastCleanupOf(t, g, 0)
	if p, _ := effectstest.EffectivePTOf(t, g, baku); p != 1 {
		t.Errorf("the bonus lasted past the turn: power %d", p)
	}
}

func TestOldB599PetalmaneBakuAddsOneColorManaPerCounterRemoved(t *testing.T) {
	g, me, _ := oldB599Table(t)
	baku := effectstest.PushCatalogPermanent(g, me.ID, "Petalmane Baku", "Creature — Spirit", oldB599Petalmane, false)
	effectstest.PfAddCounters(g, baku, "ki", 2)
	effectstest.B06AddMana(me, "C")
	before := me.ManaPool.Total()
	effectstest.ActivateManaFor(t, g, me.ID, baku, 0, game.ManaAbilityParams{CounterCounts: []int{2}})
	for i := 0; i < 4; i++ {
		var pick *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceMana && c.Chooser == me.ID {
				pick = c
			}
		}
		if pick == nil {
			break
		}
		if err := g.ResolveManaChoice(pick.ID, me.ID, "G"); err != nil {
			t.Fatalf("ResolveManaChoice: %v", err)
		}
	}
	if got := me.ManaPool.Total(); got != before-1+2 {
		t.Errorf("pool %d, want %d (paid {1}, got 2 for 2 counters)", got, before-1+2)
	}
	if got := effectstest.PoolCount(me, "G"); got != 2 {
		t.Errorf("%d green mana, want 2 of the one chosen colour", got)
	}
	if n := effectstest.Pr8Counters(g, baku, "ki"); n != 0 {
		t.Errorf("%d ki counters left, want 0", n)
	}
	if effectstest.IsTapped(g, baku) {
		t.Error("the ability tapped the Baku")
	}
}

func TestOldB599BakuAltarMakesASpiritForAKiCounter(t *testing.T) {
	g, me, _ := oldB599Table(t)
	altar := effectstest.PushCatalogPermanent(g, me.ID, "Baku Altar", "Artifact", oldB599Altar, false)
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, altar, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the ability activated with no ki counter to remove")
	}
	effectstest.PfAddCounters(g, altar, "ki", 1)
	effectstest.P7Activate(t, g, me, altar, 0, game.ActivateAbilityParams{Strict: true})
	if n := effectstest.Pr8Counters(g, altar, "ki"); n != 0 {
		t.Errorf("%d ki counters left, want 0", n)
	}
	if n := effectstest.CountBattlefieldByName(g, "Spirit"); n != 1 {
		t.Fatalf("%d Spirit tokens, want 1", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Spirit" {
			if c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 0 || !c.IsCreature() {
				t.Errorf("the token is %d/%d colors %v, want a 1/1 colorless creature", c.Power, c.Toughness, c.Colors)
			}
		}
	}
}

func TestOldB599AshenMonstrosityMustAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beast := effectstest.B12Push(g, me.ID, "Ashen Monstrosity", "Creature — Spirit", oldB599Monstrous, 7, 4)
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != beast {
		t.Fatalf("pass with the Monstrosity home = %v, want a refusal naming it", err)
	}
	if err := g.DeclareAttacker(beast, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
}

func TestOldB599ShimmeringGlasskiteCountersTheFirstTargetingEachTurn(t *testing.T) {
	g, _, opp := oldB599Table(t)
	kite := effectstest.B12Push(g, opp.ID, "Shimmering Glasskite", "Creature — Spirit", oldB599Glasskite, 2, 9)
	first := effectstest.CastCatalogSpell(t, g, "Slaying Fire", "Instant", effectstest.SlayingFireOracle, effectstest.CardTarget(kite))
	effectstest.PassPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(first); z == nil || z.Kind != game.ZoneGraveyard {
		t.Errorf("the first spell is in %v, want the graveyard (countered)", z)
	}
	if got := effectstest.LayeredCard(t, g, kite).DamageMarked; got != 0 {
		t.Errorf("the Glasskite has %d damage after a countered spell", got)
	}
	effectstest.CastCatalogSpell(t, g, "Slaying Fire", "Instant", effectstest.SlayingFireOracle, effectstest.CardTarget(kite))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LayeredCard(t, g, kite).DamageMarked; got != 3 {
		t.Errorf("the Glasskite has %d damage, want 3 from the second spell this turn", got)
	}
}

func TestOldB599KaijinReturnsWhatItBlocksAtEndOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.HmlCreature(g, me.ID, "Attacker", "", 1, 5)
	kaijin := effectstest.HmlCreature(g, opp.ID, "Kaijin of the Vanishing Touch", oldB599Kaijin, 0, 3)
	bystander := effectstest.HmlCreature(g, me.ID, "Bystander", "", 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(kaijin, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(attacker) {
		t.Fatal("the attacker left before end of combat")
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) || !effectstest.HandHasID(me, attacker) {
		t.Error("the blocked creature was not returned to its owner's hand at end of combat")
	}
	if !g.Battlefield.Contains(kaijin) || !g.Battlefield.Contains(bystander) {
		t.Error("the Kaijin or a bystander was returned")
	}
}

func oldB599ChiseiUpkeep(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	chisei := effectstest.B12Push(g, me.ID, "Chisei, Heart of Oceans", "Legendary Creature — Spirit", oldB599Chisei, 4, 4)
	effectstest.AdvanceToUpkeepOf(t, g, g.Turn.ActiveSeat)
	if g.Turn.ActiveSeat != 0 {
		effectstest.AdvanceToUpkeepOf(t, g, 0)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return chisei
}

func TestOldB599ChiseiIsSacrificedWithNoCounterToRemove(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	chisei := oldB599ChiseiUpkeep(t, g, me)
	if g.Battlefield.Contains(chisei) || !me.Graveyard.Contains(chisei) {
		t.Error("Chisei was not sacrificed with no counter on any permanent you control")
	}
}

func TestOldB599ChiseiStaysWhenYouRemoveACounterAndGoesWhenYouDont(t *testing.T) {
	for _, remove := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		holder := effectstest.B12Creature(g, me.ID, "Counter Bear", "Creature — Bear", 2, 2)
		other := effectstest.B12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
		effectstest.PfAddCounters(g, holder, game.CounterPlusOne, 2)
		effectstest.PfAddCounters(g, other, game.CounterPlusOne, 2)
		chisei := oldB599ChiseiUpkeep(t, g, me)
		pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
		if pick == nil {
			t.Fatal("no choose-cards prompt for the counter")
		}
		if !effectstest.HasID(pick.ChooseCards, holder) || effectstest.HasID(pick.ChooseCards, other) {
			t.Fatalf("the prompt offers %v, want only the permanent you control with a counter", pick.ChooseCards)
		}
		if remove {
			effectstest.AnswerChooseCards(t, g, me.ID, holder)
		} else {
			effectstest.AnswerChooseCards(t, g, me.ID)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := g.Battlefield.Contains(chisei); got != remove {
			t.Errorf("remove=%v: Chisei on the battlefield = %v", remove, got)
		}
		want := 2
		if remove {
			want = 1
		}
		if n := effectstest.Pr8Counters(g, holder, game.CounterPlusOne); n != want {
			t.Errorf("remove=%v: %d counters on the permanent, want %d", remove, n, want)
		}
		if n := effectstest.Pr8Counters(g, other, game.CounterPlusOne); n != 2 {
			t.Errorf("an opponent's counter was removed")
		}
	}
}

func TestOldB599ChiseiAsksWhichKindWhenThePermanentHoldsSeveral(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	holder := effectstest.B12Creature(g, me.ID, "Counter Bear", "Creature — Bear", 2, 2)
	effectstest.PfAddCounters(g, holder, game.CounterPlusOne, 1)
	effectstest.PfAddCounters(g, holder, "charge", 2)
	chisei := oldB599ChiseiUpkeep(t, g, me)
	effectstest.AnswerChooseCards(t, g, me.ID, holder)
	opt := effectstest.LatestOptionPickFor(g, me.ID)
	if opt == nil {
		t.Fatal("no counter-kind prompt for a permanent with two kinds")
	}
	// Kinds are offered in sorted order: "+1/+1" sorts before "charge".
	if len(opt.PickOptions) != 2 {
		t.Fatalf("the prompt offers %d kinds, want 2", len(opt.PickOptions))
	}
	idx := 0
	for i, o := range opt.PickOptions {
		if o.Label == "charge" {
			idx = i
		}
	}
	effectstest.AnswerOptionPick(t, g, me.ID, idx)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.Pr8Counters(g, holder, "charge"); n != 1 {
		t.Errorf("%d charge counters, want 1 (one removed)", n)
	}
	if n := effectstest.Pr8Counters(g, holder, game.CounterPlusOne); n != 1 {
		t.Errorf("%d +1/+1 counters, want 1 (untouched)", n)
	}
	if !g.Battlefield.Contains(chisei) {
		t.Error("Chisei was sacrificed after the counter was removed")
	}
}
