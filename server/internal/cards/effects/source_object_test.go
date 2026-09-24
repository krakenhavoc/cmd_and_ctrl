package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// source_object_test.go is the card-level proof for #1418: a triggered
// ability names the OBJECT its source was when it triggered, so "this
// permanent" at resolution is never a new object the same card has
// become (CR 400.7). The engine half is game/source_object_test.go.
//
// Every test here returns the source to the battlefield with a raw move
// after bouncing it, so it is the same card under the same instance ID,
// a new object, and — because the raw move emits nothing — without a
// second copy of its own enter trigger to muddy the count.

// flickerInResponse bounces `id` to its owner's hand and puts it
// straight back onto the battlefield while its trigger waits.
func flickerInResponse(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			t.Fatalf("flicker: %v is nowhere", id)
		}
		owner := g.PlayerByIDForEffect(c.Owner)
		before := c.ObjectEpoch
		if err := g.BounceToHandForEffect(id); err != nil {
			t.Fatalf("BounceToHandForEffect: %v", err)
		}
		if _, err := game.MoveCard(owner.Hand, g.Battlefield, id); err != nil {
			t.Fatalf("MoveCard back: %v", err)
		}
		g.BumpLayerVersionForTest()
		g.RecomputeLayersIfStaleLocked()
		if back, _ := g.LookupCardForEffect(id); back.ObjectEpoch == before {
			t.Fatal("setup: the returned card should be a new object")
		}
	})
}

// Mana Vault — the issue's own case. The tapped Vault's draw-step
// trigger is waiting; the Vault leaves and comes back (untapped, as a
// new object). "If this artifact is tapped" is about the Vault that
// triggered, which left tapped, so it still deals its 1 damage. Before
// #1418 the resolution read the new Vault and dealt nothing.
func TestManaVaultThatLeftAndCameBackIsJudgedAsTheVaultThatTriggered(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[0]
	vault := pushPermanentForTest(g, owner.ID, "Mana Vault", manaVaultOracle, "Artifact")
	if err := g.TapCard(vault, true); err != nil {
		t.Fatal(err)
	}
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	choice := untapStaticPendingPay(t, g, owner.ID)
	if err := g.ResolvePayUnless(choice.ID, owner.ID, false); err != nil {
		t.Fatalf("decline Mana Vault upkeep offer: %v", err)
	}
	if _, err := g.AdvanceStep(); err != nil || g.Turn.Step != game.StepDraw {
		t.Fatalf("advance to draw: turn=%+v err=%v", g.Turn, err)
	}
	if triggerOnStack(g, vault) == nil {
		t.Fatal("tapped Mana Vault did not trigger at draw step")
	}
	lifeBefore := owner.Life

	flickerInResponse(t, g, vault)
	if c, _ := battlefieldCard(g, vault); c.Tapped {
		t.Fatal("setup: the returned Vault should be untapped")
	}
	passPriorityAroundTable(t, g)

	if owner.Life != lifeBefore-1 {
		t.Errorf("the Vault that triggered left tapped and deals 1 (CR 400.7, #1418): life %d → %d", lifeBefore, owner.Life)
	}
	spec, _ := Lookup(manaVaultOracle)
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Mana Vault is complete now: %v %v", spec.Completeness, spec.Caveats)
	}
}

// Batterskull's living weapon: "attach this Equipment to it" names the
// Equipment that entered. Flickered in response, the Batterskull on the
// battlefield is a new object, so nothing is attached and the Germ dies
// a 0/0 — the same answer an equip gets for a replayed Warhammer (#812).
func TestLivingWeaponDoesNotAttachAnEquipmentThatCameBackAsANewObject(t *testing.T) {
	g := newCatalogGame(t)
	skull := castCatalogSpell(t, g, "Batterskull", equipTypeLine, batterskullOracle, nil)
	passUntilOnBattlefield(t, g, skull)
	if triggerOnStack(g, skull) == nil {
		t.Fatal("Batterskull's living weapon trigger is not on the stack")
	}

	flickerInResponse(t, g, skull)
	passPriorityAroundTable(t, g)

	if host := attachmentHostOf(t, g, skull); host.ID != uuid.Nil {
		t.Errorf("the new Batterskull was attached to %+v — CR 400.7 says it has no memory of the trigger", host)
	}
	if germ := findBattlefieldByName(g, "Phyrexian Germ"); germ != uuid.Nil {
		t.Error("the Germ survived with nothing attached to it — it is a 0/0")
	}
}

// Hero's Blade: "you may attach this Equipment to it". The Blade is
// bounced and replayed while the trigger waits; the Blade on the
// battlefield is not "this Equipment", so the legend stays unequipped.
func TestHerosBladeDoesNotAttachAsANewObject(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	blade := seedEquipment(g, me.ID, "Hero's Blade", herosBladeOracle)
	legend := seedLegendaryCreature(g, me.ID, "Commander")
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, Actor: me.ID, CardID: legend})
	})
	answerLatestTriggerPrompt(t, g, me.ID, true)
	if triggerOnStack(g, blade) == nil {
		t.Fatal("Hero's Blade's trigger is not on the stack")
	}

	flickerInResponse(t, g, blade)
	passPriorityAroundTable(t, g)

	if host := attachmentHostOf(t, g, blade); host.ID != uuid.Nil {
		t.Errorf("the new Hero's Blade was attached to %+v, want nothing (CR 400.7)", host)
	}
	if got := effectivePower(t, g, legend); got != 2 {
		t.Errorf("legend power %d, want an unequipped 2", got)
	}
}

// Uthros Research Craft: "put a charge counter on this". The draw is not
// about the Craft and still happens; the counter does not land on the
// new, counterless Craft that the bounce made.
func TestUthrosResearchCraftPutsNoCounterOnANewObject(t *testing.T) {
	g, me, craft, _ := stationBoard(t, "Uthros Research Craft", "Artifact — Spacecraft", uthrosResearchCraftOracle, 1)
	if err := g.AddCounter(craft, game.CounterCharge, 3); err != nil {
		t.Fatalf("to three: %v", err)
	}
	hand := len(me.Hand.Cards)
	castCatalogSpell(t, g, "Test Trinket", "Artifact", "uthros-test-trinket", nil)
	if triggerOnStack(g, craft) == nil {
		t.Fatal("the 3+ cast trigger is not on the stack")
	}

	flickerInResponse(t, g, craft)
	passPriorityAroundTable(t, g)

	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("hand = %d, want %d — the draw does not need the Craft", got, hand+1)
	}
	if got := counterOn(g, craft, game.CounterCharge); got != 0 {
		t.Errorf("charge counters on the new Craft = %d, want 0 (CR 400.7)", got)
	}
}

// The control for all four: a source that never moved reads live, so
// SourcePermanent is the permanent as it is now and not a record.
func TestSourcePermanentReadsALiveSourceLive(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	src := pushPermanentForTest(g, me.ID, "Relic", "", "Artifact")
	item := &game.StackItem{ID: uuid.New(), Kind: game.StackItemTriggered, Controller: me.ID, Owner: me.ID, SourceCardID: src}
	var info game.PermanentInfo
	var ok bool
	g.WithWriteLock(func() {
		ref, _ := g.PermanentRefForEffect(src)
		item.SourceObject = ref
		info, ok = NewContext(g, item).SourcePermanent()
	})
	if !ok || info.Left {
		t.Errorf("SourcePermanent = %+v, %v; want the live permanent", info, ok)
	}
}
