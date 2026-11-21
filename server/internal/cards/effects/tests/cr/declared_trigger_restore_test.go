package testscr

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ward reads the targeting spell and its caster off the item's
// triggering event, not off a closure: a restored ward still charges
// the caster and still counters the spell on a decline.
func TestRestoredWardTriggerStillCharges(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	giant := effectstest.PushCatalogPermanent(g, me.ID, "Rimeshield Frost Giant",
		"Creature — Giant Warrior", effectstest.RimeshieldOracle, false)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	blade := effectstest.CastAtWardedCreature(t, g, opp, giant)
	for i := 0; i < 8 && effectstest.TriggerOnStack(g, giant) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	effectstest.CorpusRequireTriggeredStamp(t, effectstest.TriggerOnStack(g, giant), "own:")

	restored := effectstest.RestoreThroughJSON(t, g)
	rOpp := restored.Seats[1]
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(restored, rOpp.ID); i++ {
		if err := restored.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !effectstest.HasPayUnlessFor(restored, rOpp.ID) {
		t.Fatal("the restored ward trigger asked the caster for nothing")
	}
	effectstest.AnswerPayUnless(t, restored, rOpp.ID, false)
	effectstest.PassPriorityAroundTable(t, restored)
	if !restored.Battlefield.Contains(giant) || !rOpp.Graveyard.Contains(blade) {
		t.Error("a declined ward after a restore must still counter the spell")
	}
}

// A clause built by TargetsFrom is built again on restore, with the
// trigger's own source: Gixian Puppeteer's "another target creature
// card" comes back as a clause, and the reanimation resolves.
func TestRestoredTargetsFromTriggerRebuildsItsClause(t *testing.T) {
	g := effectstest.CorpusTargetsFromTrigger(t)
	restored := effectstest.RestoreThroughJSON(t, g)

	again := restored.CaptureSnapshot()
	if len(again.StackMeta) != 1 || !again.StackMeta[0].HasTargetSpec {
		t.Fatalf("the restored trigger lost its target clause: %+v", again.StackMeta)
	}
	var elf game.Card
	for _, c := range restored.Seats[restored.Turn.ActiveSeat].Graveyard.Cards {
		if c.Name == "Llanowar Elves" {
			elf = c
		}
	}
	effectstest.PassPriorityAroundTable(t, restored)
	if !restored.Battlefield.Contains(elf.InstanceID) {
		t.Error("the restored Puppeteer trigger did not return its target")
	}
}

// Storm names the spell it copies by the item's source, not a closure,
// so a storm trigger over a countered spell is a restore point.
func TestStormTriggerOverACounteredSpellIsData(t *testing.T) {
	g := effectstest.CorpusStormAfterCounter(t)
	snap := g.CaptureSnapshot()
	if !snap.Restorable() || len(snap.StackMeta) != 1 {
		t.Fatalf("restorable %v, %d stack items", snap.Restorable(), len(snap.StackMeta))
	}
	if it := snap.StackMeta[0]; it.Body != game.CatalogTriggeredBodyKey || it.Params == nil || it.Params.Ability == nil {
		t.Errorf("stack item %q is not stamped: body %q", it.Label, it.Body)
	}
}

// Cascade's mana value rides the item as Params.Amount, filled in by
// the keyword's Build; the effect is the row's.
func TestCascadeTriggerCarriesItsManaValueAsData(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	elf := effectstest.CastWithCost(t, g, "Bloodbraid Elf", "Creature — Elf Berserker", "{2}{R}{G}", effectstest.BloodbraidElfOracle)
	var item *game.StackItem
	for i := 0; i < 8 && item == nil; i++ {
		for _, it := range g.PendingTriggers {
			if it.SourceCardID == elf {
				item = it
			}
		}
		if item == nil {
			item = effectstest.TriggerOnStack(g, elf)
		}
		if item == nil {
			if err := g.PassPriority(); err != nil {
				t.Fatalf("PassPriority: %v", err)
			}
		}
	}
	if item == nil {
		t.Fatal("no cascade trigger")
	}
	effectstest.CorpusRequireTriggeredStamp(t, item, "own:")
	if item.Params.Amount != 4 || item.Label != "Bloodbraid Elf — cascade" {
		t.Errorf("cascade item: amount %d label %q, want 4 and the card's label", item.Params.Amount, item.Label)
	}
}
