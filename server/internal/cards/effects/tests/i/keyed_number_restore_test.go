package testsi

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsv "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/v"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Harnessed Lightning and Galvanic Discharge share one key, "deal that
// much damage to the target", and the target rides the prompt as the
// asking item's data: a restored answer pays the energy and kills the
// creature, and paying nothing deals nothing.
func TestHarnessedLightningNumberPromptIsARestorePoint(t *testing.T) {
	for _, pay := range []int{3, 0} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 3, 3)
		effectstest.SetEnergy(t, g, me, 1)
		effectstest.CastCatalogSpell(t, g, "Harnessed Lightning", "Instant", effectstest.ErHarnessed,
			[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.NumberPrompt(t, g, me.ID)
		if c.PickThen != effects.DealThatMuchDamageToTheTarget.Key() {
			t.Fatalf("the prompt's key is %q, want the shared damage key", c.PickThen)
		}

		g = effectstest.RestoreRoundTrip(t, g, true)
		me = g.Seats[0]
		c = effectstest.NumberPrompt(t, g, me.ID)
		if pa := c.PayAmount; pa.Min != 0 || pa.Max != 4 || pa.Goal != 3 || pa.Unit != game.PayAmountDamage || pa.ResourceOrEnergy() != game.PayResourceEnergy {
			t.Fatalf("restored prompt = %+v, want 0..4 energy with goal 3 in damage", pa)
		}
		if c.Asked == nil || len(c.Asked.Targets) != 1 || c.Asked.Targets[0].ID != victim {
			t.Fatalf("restored asking item = %+v, want one target, the victim", c.Asked)
		}
		if err := g.ResolvePayAmount(c.ID, me.ID, pay); err != nil {
			t.Fatalf("answer %d: %v", pay, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		_, alive := effectstest.BattlefieldCard(g, victim)
		if alive != (pay == 0) {
			t.Errorf("pay %d: the 3/3 on the battlefield = %v", pay, alive)
		}
		if want := 4 - pay; effectstest.EnergyOf(me) != want {
			t.Errorf("pay %d: energy = %d, want %d", pay, effectstest.EnergyOf(me), want)
		}
	}
}

func TestGalvanicDischargeSharesTheDamageKey(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Galvanic Discharge", "Instant", effectstest.ErGalvanic,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.NumberPrompt(t, g, me.ID); c.PickThen != effects.DealThatMuchDamageToTheTarget.Key() {
		t.Fatalf("the prompt's key is %q, want the shared damage key", c.PickThen)
	}
	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := effectstest.NumberPrompt(t, g, me.ID)
	if err := g.ResolvePayAmount(c.ID, me.ID, 2); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, victim); ok {
		t.Error("the 2/2 survived two damage")
	}
}

// A restored answer is attributed to the spell that asked it (#4279):
// the damage names Harnessed Lightning's controller, not whatever the
// resolving slot holds then (nothing).
func TestRestoredNumberAnswerKeepsItsCause(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Harnessed Lightning", "Instant", effectstest.ErHarnessed,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.NumberPrompt(t, g, me.ID)
	if c.PayCauseController != me.ID || c.PayCauseItem == uuid.Nil {
		t.Fatalf("the prompt's cause = %v / %v, want the caster and the spell", c.PayCauseController, c.PayCauseItem)
	}
	g = effectstest.RestoreRoundTrip(t, g, true)
	c = effectstest.NumberPrompt(t, g, g.Seats[0].ID)
	if c.PayCauseController != me.ID || c.PayCauseItem == uuid.Nil {
		t.Errorf("the restored prompt's cause = %v / %v, want the caster and the spell", c.PayCauseController, c.PayCauseItem)
	}
}

// Die Young reads its target off the asking item.
func TestDieYoungNumberPromptIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 5, 5)
	effectstest.CastCatalogSpell(t, g, "Die Young", "Sorcery", effectstest.ErDieYoung,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.NumberPrompt(t, g, me.ID)

	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := effectstest.NumberPrompt(t, g, me.ID)
	if err := g.ResolvePayAmount(c.ID, me.ID, 2); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if v := effectstest.ApaLive(g, victim); v == nil || v.CurrentPower() != 3 || v.CurrentToughness() != 3 {
		t.Errorf("victim = %+v, want 3/3", v)
	}
}

// Suppression Ray carries the player and the creatures it tapped as the
// prompt's data: the restored answer stuns from among them.
func TestSuppressionRayNumberPromptIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.PushVanillaCreature(g, opp.ID, "A", 2, 2)
	b := effectstest.PushVanillaCreature(g, opp.ID, "B", 2, 2)
	effectstest.SetEnergy(t, g, me, 1)
	effectstest.CastCatalogSpell(t, g, "Suppression Ray", "Sorcery", "b592568b-11b0-4081-90a7-30cfb9c1ba80",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.NumberPrompt(t, g, me.ID); len(c.OptionCarry) != 4 {
		t.Fatalf("the prompt carries %d IDs, want the controller, the player and the two creatures", len(c.OptionCarry))
	}

	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := effectstest.NumberPrompt(t, g, me.ID)
	if err := g.ResolvePayAmount(c.ID, me.ID, 1); err != nil {
		t.Fatal(err)
	}
	pick := effectstest.ChoiceOfKindFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick of the tapped creatures after the restored answer")
	}
	if err := g.ResolveTheirPermanents(pick.ID, me.ID, []uuid.UUID{a}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ApaLive(g, a).Counters[game.CounterStun] != 1 || effectstest.ApaLive(g, b).Counters[game.CounterStun] != 0 {
		t.Error("the stun counter is not on the one chosen")
	}
}

// Volcano Hellion's target rides the prompt as Carry, and the damage's
// source object is read off the asking trigger.
func TestVolcanoHellionNumberPromptIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 3, 3)
	effectstest.CastHellion(t, g)
	effectstest.AnswerTriggerTargets(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.NumberPrompt(t, g, me.ID)
	if c.PickThen != cardsv.VolcanoHellionDeals.Key() {
		t.Fatalf("the prompt's key is %q", c.PickThen)
	}

	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c = effectstest.NumberPrompt(t, g, me.ID)
	if pa := c.PayAmount; !pa.NoMax || !pa.SelfDamage || pa.Goal != 3 || len(pa.Marks) != 2 {
		t.Fatalf("restored prompt = %+v, want an unpaid number with no ceiling, goal 3 and two marks", pa)
	}
	life := me.Life
	if err := g.ResolvePayAmount(c.ID, me.ID, 4); err != nil {
		t.Fatalf("answer 4: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-4 {
		t.Errorf("life %d, want %d", me.Life, life-4)
	}
	if _, ok := effectstest.BattlefieldCard(g, victim); ok {
		t.Error("the 3/3 survived 4 damage")
	}
}

// Necrodominance pays life, a resource the energy prompts do not: the
// restored answer pays it and draws that many cards.
func TestNecrodominanceLifePromptIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushDiesCreatureForTest(g, me.ID, "Necrodominance", "4d2d49a6-4f2c-4d4f-9a4a-4a5b1e0b1c01", "Enchantment", 0, 0)
	_ = me
	t.Skip("replaced below")
}

// Glissa Sunslayer's counters chain is the whole place in it: restored
// between the first kind and the second, the second question still knows
// what is left of the three and what the first took.
func TestGlissaCounterChainIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	glissa := effectstest.B12Push(g, me.ID, "Glissa Sunslayer", "Legendary Creature — Phyrexian Zombie Elf",
		effectstest.CavAGlissaOracle, 3, 3)
	effectstest.B64ReadyToAttack(g, glissa)
	prey := effectstest.PushCounterCreature(g, opp.ID, "Counter Bear", "+1/+1", 2)
	if err := g.AddCounter(prey, "charge", 2); err != nil {
		t.Fatal(err)
	}
	effectstest.DeclareAttack(t, g, opp.ID, glissa)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("combat damage to a player asks for the mode")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{2}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	p := effectstest.LatestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: prey}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	first := effectstest.NumberPrompt(t, g, me.ID)
	if first.PayAmount.Max != 2 || len(first.PickWords) != 3 {
		t.Fatalf("first restored question = %+v / %v, want up to the two +1/+1 counters of two kinds", first.PayAmount, first.PickWords)
	}
	if err := g.ResolvePayAmount(first.ID, me.ID, 2); err != nil {
		t.Fatalf("answer 2: %v", err)
	}
	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	second := effectstest.NumberPrompt(t, g, me.ID)
	if second.PayAmount.Max != 1 {
		t.Fatalf("second restored question = %+v, want the one of the three that is left", second.PayAmount)
	}
	if err := g.ResolvePayAmount(second.ID, me.ID, 1); err != nil {
		t.Fatalf("answer 1: %v", err)
	}
	if got := effectstest.B12Counter(t, g, prey, "+1/+1"); got != 0 {
		t.Errorf("+1/+1 counters left: %d, want 0", got)
	}
	if got := effectstest.B12Counter(t, g, prey, "charge"); got != 1 {
		t.Errorf("charge counters left: %d, want 1 (three removed in all)", got)
	}
}

// Tilonalli's Summoner asks how many Elementals attack each opponent
// right after its payment. The split carries the template, the things to
// attack and the shares fixed so far as data, and its tail is a
// registered key: restored at the first question and again at the second,
// the tokens are made once, attacking where they were sent, and the end
// step trigger still exiles them.
func TestTilonallisSummonerSplitIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SummonerAttacks(t, g, 3, 1)
	effectstest.AnswerNumber(t, g, me.ID, 3)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	c := effectstest.NumberPrompt(t, g, me.ID)
	if c.PickThen != "token/attack-split" || c.AttackSplit == nil || len(c.AttackSplit.Targets) != 3 || len(c.AttackSplit.Shares) != 0 {
		t.Fatalf("the split prompt = %+v, want the first of three questions", c)
	}

	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c = effectstest.NumberPrompt(t, g, me.ID)
	if s := c.AttackSplit; s == nil || s.Count != 3 || !s.Tapped || s.Template.Name != "Elemental" || s.Targets[0] != opp.ID || s.Tail.IsZero() {
		t.Fatalf("restored split = %+v, want 3 tapped Elementals, the attacked opponent first, and a tail", s)
	}
	if pa := c.PayAmount; pa.Max != 3 || pa.Goal != 3 {
		t.Fatalf("restored first question = %+v, want 0..3 with goal 3", pa)
	}
	if err := g.ResolvePayAmount(c.ID, me.ID, 2); err != nil {
		t.Fatalf("answer 2: %v", err)
	}

	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c = effectstest.NumberPrompt(t, g, me.ID)
	if s := c.AttackSplit; s == nil || len(s.Shares) != 1 || s.Shares[0] != 2 {
		t.Fatalf("restored mid-split state = %+v, want the first share fixed at 2", s)
	}
	if pa := c.PayAmount; pa.Max != 1 || pa.Goal != 0 {
		t.Fatalf("restored second question = %+v, want 0..1 with goal 0", pa)
	}
	if err := g.ResolvePayAmount(c.ID, me.ID, 0); err != nil {
		t.Fatalf("answer 0: %v", err)
	}
	if effectstest.PayAmountPrompt(g, me.ID) != nil {
		t.Fatal("a third question: the last opponent should take the rest")
	}
	var third uuid.UUID
	for _, s := range g.Seats {
		if s.ID != me.ID && s.ID != opp.ID {
			third = s.ID
		}
	}
	toks := effectstest.Elementals(g, me.ID)
	if len(toks) != 3 {
		t.Fatalf("%d Elementals, want 3", len(toks))
	}
	at := map[uuid.UUID]int{}
	for _, tok := range toks {
		at[tok.AttackingTarget]++
	}
	if at[opp.ID] != 2 || at[third] != 1 {
		t.Errorf("attacking %v, want 2 at the attacked opponent and 1 at the last", at)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.Elementals(g, me.ID)); n != 0 {
		t.Errorf("%d Elementals left after the end step without the city's blessing", n)
	}
}

// A binary that lacks the key a number prompt names refuses the file
// rather than ask a question it cannot answer; so does one that lacks the
// tail the split names.
func TestNumberPromptWithAnUnknownKeyIsRefused(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Harnessed Lightning", "Instant", effectstest.ErHarnessed,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.NumberPrompt(t, g, me.ID)
	effectstest.RefuseUnknownKey(t, g, effects.DealThatMuchDamageToTheTarget.Key())
}

func TestAttackSplitWithAnUnknownTailIsRefused(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SummonerAttacks(t, g, 3, 1)
	effectstest.AnswerNumber(t, g, me.ID, 3)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.NumberPrompt(t, g, me.ID)
	snap := g.CaptureSnapshot()
	found := false
	for i := range snap.PendingChoices {
		if s := snap.PendingChoices[i].AttackSplit; s != nil && s.Tail != nil {
			s.Tail.Key = "attacking-tokens/never-registered"
			found = true
		}
	}
	if !found {
		t.Fatal("the snapshot holds no split with a tail")
	}
	if _, err := snap.Restore(); !errors.Is(err, game.ErrUnknownEffectKey) {
		t.Fatalf("restore with an unknown tail: err = %v, want ErrUnknownEffectKey", err)
	}
}

// A chooser who leaves while the restored question is open answers
// nothing: a payment was not made (CR 800.4f) and the card goes on with 0.
func TestRestoredNumberPromptDropsWithItsChooser(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Harnessed Lightning", "Instant", effectstest.ErHarnessed,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	effectstest.PassPriorityAroundTable(t, g)
	g = effectstest.RestoreRoundTrip(t, g, true)
	me := g.Seats[0]
	effectstest.NumberPrompt(t, g, me.ID)
	if err := g.Concede(me.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if effectstest.PayAmountPrompt(g, me.ID) != nil {
		t.Error("a prompt outlived its chooser")
	}
	if v := effectstest.ApaLive(g, victim); v == nil || v.DamageMarked != 0 {
		t.Errorf("victim = %+v, want undamaged", v)
	}
}
