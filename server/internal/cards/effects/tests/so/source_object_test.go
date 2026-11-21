package testsso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mana Vault — the issue's own case. The tapped Vault's draw-step
// trigger is waiting; the Vault leaves and comes back (untapped, as a
// new object). "If this artifact is tapped" is about the Vault that
// triggered, which left tapped, so it still deals its 1 damage. Before
// #1418 the resolution read the new Vault and dealt nothing.
func TestManaVaultThatLeftAndCameBackIsJudgedAsTheVaultThatTriggered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[0]
	vault := effectstest.PushPermanentForTest(g, owner.ID, "Mana Vault", effectstest.ManaVaultOracle, "Artifact")
	if err := g.TapCard(vault, true); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	choice := effectstest.UntapStaticPendingPay(t, g, owner.ID)
	if err := g.ResolvePayUnless(choice.ID, owner.ID, false); err != nil {
		t.Fatalf("decline Mana Vault upkeep offer: %v", err)
	}
	if _, err := g.AdvanceStep(); err != nil || g.Turn.Step != game.StepDraw {
		t.Fatalf("advance to draw: turn=%+v err=%v", g.Turn, err)
	}
	if effectstest.TriggerOnStack(g, vault) == nil {
		t.Fatal("tapped Mana Vault did not trigger at draw step")
	}
	lifeBefore := owner.Life

	effectstest.FlickerInResponse(t, g, vault)
	if c, _ := effectstest.BattlefieldCard(g, vault); c.Tapped {
		t.Fatal("setup: the returned Vault should be untapped")
	}
	effectstest.PassPriorityAroundTable(t, g)

	if owner.Life != lifeBefore-1 {
		t.Errorf("the Vault that triggered left tapped and deals 1 (CR 400.7, #1418): life %d → %d", lifeBefore, owner.Life)
	}
	spec, _ := effects.Lookup(effectstest.ManaVaultOracle)
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Mana Vault is complete now: %v %v", spec.Completeness, spec.Caveats)
	}
}

// Batterskull's living weapon: "attach this Equipment to it" names the
// Equipment that entered. Flickered in response, the Batterskull on the
// battlefield is a new object, so nothing is attached and the Germ dies
// a 0/0 — the same answer an equip gets for a replayed Warhammer (#812).
func TestLivingWeaponDoesNotAttachAnEquipmentThatCameBackAsANewObject(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	skull := effectstest.CastCatalogSpell(t, g, "Batterskull", effectstest.EquipTypeLine, effectstest.BatterskullOracle, nil)
	effectstest.PassUntilOnBattlefield(t, g, skull)
	if effectstest.TriggerOnStack(g, skull) == nil {
		t.Fatal("Batterskull's living weapon trigger is not on the stack")
	}

	effectstest.FlickerInResponse(t, g, skull)
	effectstest.PassPriorityAroundTable(t, g)

	if host := effectstest.AttachmentHostOf(t, g, skull); host.ID != uuid.Nil {
		t.Errorf("the new Batterskull was attached to %+v — CR 400.7 says it has no memory of the trigger", host)
	}
	if germ := effectstest.FindBattlefieldByName(g, "Phyrexian Germ"); germ != uuid.Nil {
		t.Error("the Germ survived with nothing attached to it — it is a 0/0")
	}
}

// Hero's Blade: "you may attach this Equipment to it". The Blade is
// bounced and replayed while the trigger waits; the Blade on the
// battlefield is not "this Equipment", so the legend stays unequipped.
func TestHerosBladeDoesNotAttachAsANewObject(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	blade := effectstest.SeedEquipment(g, me.ID, "Hero's Blade", effectstest.HerosBladeOracle)
	legend := effectstest.SeedLegendaryCreature(g, me.ID, "Commander")
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, Actor: me.ID, CardID: legend})
	})
	g.RunStateChecksForTest()
	if effectstest.TriggerOnStack(g, blade) == nil {
		t.Fatal("Hero's Blade's trigger is not on the stack")
	}

	effectstest.FlickerInResponse(t, g, blade)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	if host := effectstest.AttachmentHostOf(t, g, blade); host.ID != uuid.Nil {
		t.Errorf("the new Hero's Blade was attached to %+v, want nothing (CR 400.7)", host)
	}
	if got := effectstest.EffectivePower(t, g, legend); got != 2 {
		t.Errorf("legend power %d, want an unequipped 2", got)
	}
}

// Uthros Research Craft: "put a charge counter on this". The draw is not
// about the Craft and still happens; the counter does not land on the
// new, counterless Craft that the bounce made.
func TestUthrosResearchCraftPutsNoCounterOnANewObject(t *testing.T) {
	g, me, craft, _ := effectstest.StationBoard(t, "Uthros Research Craft", "Artifact — Spacecraft", effectstest.UthrosResearchCraftOracle, 1)
	if err := g.AddCounter(craft, game.CounterCharge, 3); err != nil {
		t.Fatalf("to three: %v", err)
	}
	hand := len(me.Hand.Cards)
	effectstest.CastCatalogSpell(t, g, "Test Trinket", "Artifact", "uthros-test-trinket", nil)
	if effectstest.TriggerOnStack(g, craft) == nil {
		t.Fatal("the 3+ cast trigger is not on the stack")
	}

	effectstest.FlickerInResponse(t, g, craft)
	effectstest.PassPriorityAroundTable(t, g)

	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("hand = %d, want %d — the draw does not need the Craft", got, hand+1)
	}
	if got := effectstest.CounterOn(g, craft, game.CounterCharge); got != 0 {
		t.Errorf("charge counters on the new Craft = %d, want 0 (CR 400.7)", got)
	}
}

// The control for all four: a source that never moved reads live, so
// SourcePermanent is the permanent as it is now and not a record.
func TestSourcePermanentReadsALiveSourceLive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	src := effectstest.PushPermanentForTest(g, me.ID, "Relic", "", "Artifact")
	item := &game.StackItem{ID: uuid.New(), Kind: game.StackItemTriggered, Controller: me.ID, Owner: me.ID, SourceCardID: src}
	var info game.PermanentInfo
	var ok bool
	g.WithWriteLock(func() {
		ref, _ := g.PermanentRefForEffect(src)
		item.SourceObject = ref
		info, ok = effects.NewContext(g, item).SourcePermanent()
	})
	if !ok || info.Left {
		t.Errorf("SourcePermanent = %+v, %v; want the live permanent", info, ok)
	}
}
