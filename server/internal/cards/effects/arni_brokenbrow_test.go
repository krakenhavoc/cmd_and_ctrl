package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// arni_brokenbrow_test.go — #2731, ADR 0197 §10: an activated ability
// whose whole effect is "you may" asks as it resolves. Arni Brokenbrow's
// boast is the card.

const arniBrokenbrowOracle = "3dce6c6c-599f-4e35-ae15-728c5e78dee1"

// arniTable seats Arni (3/3) for seat 0, beside a creature of
// otherPower when otherPower > 0, has Arni attack, and activates the
// boast. The ability is on the stack, unasked.
func arniTable(t *testing.T, otherPower int) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arni := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Arni Brokenbrow", TypeLine: "Legendary Creature — Human Berserker",
		OracleID: arniBrokenbrowOracle, Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	if otherPower > 0 {
		pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Big Friend", TypeLine: "Creature — Giant",
			Power: otherPower, Toughness: otherPower, Owner: me.ID, Controller: me.ID,
		})
	}
	attackWith(t, g, opp.ID, arni)
	fillPool(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, arni, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("boast: %v", err)
	}
	return g, me, arni
}

// arniItem is Arni's boast waiting on the stack, or nil.
func arniItem(g *game.Game, arni uuid.UUID) *game.StackItem {
	var out *game.StackItem
	g.ReadSnapshot(func() {
		for _, it := range g.StackMeta {
			if it != nil && it.Kind == game.StackItemActivated && it.SourceCardID == arni {
				out = it
			}
		}
	})
	return out
}

// The question is not asked at activation: the ability goes on the
// stack carrying it, and it is asked once the ability resolves.
func TestArniAsksAsTheBoastResolvesAndYesSetsBasePower(t *testing.T) {
	g, me, arni := arniTable(t, 4)
	if latestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("the \"you may\" was asked at activation")
	}
	item := arniItem(g, arni)
	if item == nil || item.Optional == nil {
		t.Fatalf("the boast is not on the stack with its \"you may\": %+v", item)
	}
	ask := awaitTriggerPrompt(t, g, me.ID)
	if ask.OptionalItem == uuid.Nil || ask.NoGain {
		t.Fatalf("prompt = %+v, want the resolving boast's question with no warning (5 > 3)", ask)
	}
	if ask.Source != arni {
		t.Errorf("prompt source = %s, want Arni", ask.Source)
	}
	if err := g.ResolveTriggerPrompt(ask.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveTriggerPrompt: %v", err)
	}
	if got := effectivePower(t, g, arni); got != 5 {
		t.Errorf("Arni's power = %d, want 5 (1 plus the 4-power creature)", got)
	}
	if len(g.PendingChoices) != 0 || !stackFullyEmpty(g) {
		t.Error("the boast did not finish resolving")
	}
}

// "No" leaves Arni as it was. With no other creature the set would make
// it 1, so the prompt warns that "yes" would not help.
func TestArniDeclinedKeepsItsPowerAndWarnsWhenYesWouldShrinkIt(t *testing.T) {
	g, me, arni := arniTable(t, 0)
	ask := awaitTriggerPrompt(t, g, me.ID)
	if !ask.NoGain {
		t.Error("alone, a \"yes\" makes Arni 1 power, and the prompt does not warn")
	}
	if err := g.ResolveTriggerPrompt(ask.ID, me.ID, false); err != nil {
		t.Fatalf("ResolveTriggerPrompt: %v", err)
	}
	if got := effectivePower(t, g, arni); got != 3 {
		t.Errorf("Arni's power = %d, want 3: declined", got)
	}
	if len(g.PendingChoices) != 0 || !stackFullyEmpty(g) {
		t.Error("the declined boast did not finish resolving")
	}
}

// A "yes" is still the player's to give when it shrinks Arni: the set
// is the printed effect, worse or not.
func TestArniYesAloneMakesItOne(t *testing.T) {
	g, me, arni := arniTable(t, 0)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	if got := effectivePower(t, g, arni); got != 1 {
		t.Errorf("Arni's power = %d, want 1 (no other creature)", got)
	}
}

// ADR 0127: the prompt is keyed by the row, so a standing "Always"
// answers it for the player as it resolves.
func TestArniStandingAnswerAppliesAtResolution(t *testing.T) {
	g, me, arni := arniTable(t, 6)
	key := arniItem(g, arni).Optional.AnswerKey
	want := arniBrokenbrowOracle + "|activated|Boast — {1}: You may have Arni's base power"
	if !strings.HasPrefix(key, want) {
		t.Fatalf("standing-answer key = %q, want the row's (%q…)", key, want)
	}
	if err := g.SetAutoAnswers(me.ID, map[string]game.AutoAnswer{key: game.AutoAnswerAlways}); err != nil {
		t.Fatal(err)
	}
	ask := awaitTriggerPrompt(t, g, me.ID)
	if ask.AutoAnswerKey != key || ask.AutoAnswerCard != "Arni Brokenbrow" {
		t.Fatalf("prompt key %q card %q, want %q and Arni", ask.AutoAnswerKey, ask.AutoAnswerCard, key)
	}
	if n := autoAnswerAll(t, g); n != 1 {
		t.Fatalf("answered %d automatically, want 1", n)
	}
	if got := effectivePower(t, g, arni); got != 7 {
		t.Errorf("Arni's power = %d, want 7 after the automatic yes", got)
	}
}

// Arni leaves before its boast resolves. The ability still resolves (CR
// 113.7a) and asks, warning that "yes" would do nothing, and a "yes"
// changes nothing.
func TestArniLeftBeforeTheBoastResolves(t *testing.T) {
	g, me, arni := arniTable(t, 4)
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(arni); err != nil {
			t.Fatalf("exile in response: %v", err)
		}
	})
	ask := awaitTriggerPrompt(t, g, me.ID)
	if !ask.NoGain {
		t.Error("with Arni gone the prompt does not warn")
	}
	if err := g.ResolveTriggerPrompt(ask.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveTriggerPrompt: %v", err)
	}
	if len(g.PendingChoices) != 0 || !stackFullyEmpty(g) {
		t.Error("the boast did not finish resolving")
	}
	if g.Battlefield.Contains(arni) {
		t.Error("Arni came back")
	}
}

// The open question is plain data: a restore point taken while it is
// asked comes back asking, and a "yes" there still sets the power.
func TestArniOpenQuestionIsARestorePoint(t *testing.T) {
	g, me, arni := arniTable(t, 4)
	awaitTriggerPrompt(t, g, me.ID)
	restored := restoreRoundTrip(t, g, true)
	answerLatestTriggerPrompt(t, restored, me.ID, true)
	if got := effectivePower(t, restored, arni); got != 5 {
		t.Errorf("restored: Arni's power = %d, want 5", got)
	}
}

// A copy of the boast (CR 707.10) has its text, "you may" and all, so
// it asks its own question as it resolves.
func TestArniCopiedBoastAsksToo(t *testing.T) {
	g, me, arni := arniTable(t, 4)
	item := arniItem(g, arni)
	g.WithWriteLock(func() {
		if err := g.CopyAbilityForEffect(item.ID, me.ID, false); err != nil {
			t.Fatalf("copy: %v", err)
		}
	})
	asked := 0
	for i := 0; i < 2; i++ {
		ask := awaitTriggerPrompt(t, g, me.ID)
		asked++
		if err := g.ResolveTriggerPrompt(ask.ID, me.ID, i == 0); err != nil {
			t.Fatalf("ResolveTriggerPrompt: %v", err)
		}
	}
	if asked != 2 {
		t.Fatalf("asked %d times, want 2", asked)
	}
	if got := effectivePower(t, g, arni); got != 5 {
		t.Errorf("Arni's power = %d, want 5 (the copy said yes)", got)
	}
}
