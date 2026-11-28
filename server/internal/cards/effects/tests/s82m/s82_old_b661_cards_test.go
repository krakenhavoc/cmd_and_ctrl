package testss82m

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b661_cards_test.go — Invasion Auras, S82 OLD batch 661 (#5616).
// Every test plays the real catalog entry through the engine.

const (
	b661Shackles  = "05861ff2-ec85-43d1-b641-c5d2da819c03"
	b661Mourning  = "20602d11-c0ae-47ce-a7fd-998b12480e18"
	b661Scavenged = "2d3010d5-5c21-4342-ad79-a737b6731230"
	b661Wings     = "7d892d61-5a69-4500-9fb2-7cf50728aedd"
	b661Cursed    = "96e244ef-21b9-4795-a881-c3db23b60cc9"
	b661Leak      = "cb02daf6-9013-4ccc-bf1f-773b757ac468"
	b661WhipSilk  = "ce2c420b-e569-4715-b65b-4a097700799d"
	b661Pledge    = "e88b3066-eb10-4ade-8207-af4de40facec"
	b661Crown     = "f14fb965-5ebc-427a-a5b8-149f54c64171"
	b661Bear      = "Creature — Bear"
	b661AuraLine  = "Enchantment — Aura"
)

func b661Aura(t *testing.T, g *game.Game, name, oracle string, target uuid.UUID) uuid.UUID {
	t.Helper()
	id := effectstest.B12PlayFromHand(t, g, name, b661AuraLine, oracle, game.CastSpellParams{Targets: effectstest.CardRefs(target)})
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

// b661Returns activates ability idx of the Aura with the mana given and
// checks that it came back to its owner's hand.
func b661Returns(t *testing.T, g *game.Game, me *game.Player, aura uuid.UUID, idx int, mana string) {
	t.Helper()
	if err := g.ActivateCatalogAbility(me.ID, aura, idx, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the return was activated with no mana")
	}
	effectstest.AddPool(t, g, me.ID, mana)
	effectstest.ApaActivate(t, g, me, aura, idx, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(aura) || !me.Hand.Contains(aura) {
		t.Error("the Aura did not return to its owner's hand")
	}
}

func TestB661WhipSilkGivesReachAndReturns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", b661Bear, 2, 2)
	aura := b661Aura(t, g, "Whip Silk", b661WhipSilk, bear)
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "reach") {
		t.Error("the enchanted creature has no reach")
	}
	b661Returns(t, g, me, aura, 0, "{G}")
	if effectstest.HasKeywordOnBattlefield(t, g, bear, "reach") {
		t.Error("the creature kept reach after the Aura left")
	}
}

func TestB661ShimmeringWingsGivesFlyingAndReturns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", b661Bear, 2, 2)
	aura := b661Aura(t, g, "Shimmering Wings", b661Wings, bear)
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Error("the enchanted creature has no flying")
	}
	b661Returns(t, g, me, aura, 0, "{U}")
	if effectstest.HasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Error("the creature kept flying after the Aura left")
	}
}

func TestB661MourningShrinksPowerAndReturns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", b661Bear, 3, 3)
	aura := b661Aura(t, g, "Mourning", b661Mourning, bear)
	if p, tough := effectstest.PtOf(t, g, bear); p != 1 || tough != 3 {
		t.Errorf("the enchanted creature is %d/%d, want 1/3", p, tough)
	}
	b661Returns(t, g, me, aura, 0, "{B}")
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the creature is %d/%d after the Aura left, want 3/3", p, tough)
	}
}

func TestB661CrownOfFlamesPumpsUntilEndOfTurnAndReturns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", b661Bear, 2, 2)
	aura := b661Aura(t, g, "Crown of Flames", b661Crown, bear)
	for i := 0; i < 2; i++ {
		effectstest.AddPool(t, g, me.ID, "{R}")
		effectstest.ApaActivate(t, g, me, aura, 0, game.ActivateAbilityParams{})
	}
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("after two pumps the creature is %d/%d, want 4/2", p, tough)
	}
	b661Returns(t, g, me, aura, 1, "{R}")
	if p, _ := effectstest.PtOf(t, g, bear); p != 4 {
		t.Errorf("the until-end-of-turn pump ended with the Aura: power %d, want 4", p)
	}
}

func TestB661ShacklesKeepsTheCreatureTappedAndReturns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Theirs", b661Bear, 2, 2)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == theirs {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	aura := b661Aura(t, g, "Shackles", b661Shackles, theirs)
	effectstest.DftMPPassTurn(t, g)
	effectstest.AdvanceToMain(t, g)
	if !effectstest.DftCG2Tapped(g, theirs) {
		t.Error("the enchanted creature untapped during its controller's untap step")
	}
	_ = aura
	_ = me
}

func TestB661ShacklesReturnsToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", b661Bear, 2, 2)
	aura := b661Aura(t, g, "Shackles", b661Shackles, bear)
	b661Returns(t, g, me, aura, 0, "{W}")
}

func TestB661CursedFleshShrinksAndGivesFear(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", b661Bear, 2, 2)
	b661Aura(t, g, "Cursed Flesh", b661Cursed, bear)
	if p, tough := effectstest.PtOf(t, g, bear); p != 1 || tough != 1 {
		t.Errorf("the enchanted creature is %d/%d, want 1/1", p, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "fear") {
		t.Error("the enchanted creature has no fear")
	}
}

func TestB661CursedFleshKillsATougnessOneCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mouse := effectstest.B12Creature(g, me.ID, "Mouse", b661Bear, 1, 1)
	aura := b661Aura(t, g, "Cursed Flesh", b661Cursed, mouse)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(mouse) || g.Battlefield.Contains(aura) {
		t.Error("the 1/1 creature survived -1/-1 or the Aura stayed behind")
	}
}

func TestB661ScavengedWeaponryDrawsAndPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", b661Bear, 2, 2)
	effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})
	before := effectstest.HandSize(me)
	b661Aura(t, g, "Scavenged Weaponry", b661Scavenged, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandSize(me); got != before+1 {
		t.Errorf("hand = %d, want %d (the Aura left the hand, one card drawn)", got, before+1)
	}
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the enchanted creature is %d/%d, want 3/3", p, tough)
	}
}

func TestB661PledgeGivesProtectionFromEachColorAmongYourPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", b661Bear, 2, 2)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Red Thing", TypeLine: "Creature — Elemental", Colors: []string{"R"},
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	// The test harness casts the Aura without its printed color, so a white
	// permanent stands in for the Aura's own.
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "White Thing", TypeLine: "Creature — Elemental", Colors: []string{"W"},
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	other := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Blue Thing", TypeLine: "Creature — Elemental", Colors: []string{"U"},
		Power: 1, Toughness: 1, Owner: g.Seats[1].ID, Controller: g.Seats[1].ID,
	})
	_ = other
	aura := b661Aura(t, g, "Pledge of Loyalty", b661Pledge, bear)
	has := func(color string) bool {
		c := effectstest.LayeredCard(t, g, bear)
		for _, a := range c.Effective().Abilities {
			if strings.HasPrefix(strings.ToLower(a), "protection from "+color) {
				return true
			}
		}
		return false
	}
	if !g.Battlefield.Contains(aura) {
		t.Fatal("the Aura removed itself: it is white and grants protection from white")
	}
	if !has("red") || !has("white") {
		t.Error("no protection from red or white, colors among my permanents")
	}
	if has("blue") {
		t.Error("protection from blue came from an opponent's permanent")
	}
}

func b661RunLeak(t *testing.T, pay bool, color []string, cost, mana string, wantStays bool) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Host", TypeLine: "Creature — Elemental", Colors: color, ManaCost: cost,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() { g.Turn.ActiveSeat = 0 })
	effectstest.B12PlayFromHand(t, g, "Essence Leak", b661AuraLine, b661Leak, game.CastSpellParams{Targets: effectstest.CardRefs(host)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.WalkToUpkeepOfSeat0(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if len(color) == 0 || (color[0] != "R" && color[0] != "G") {
		if len(g.PendingChoices) != 0 {
			t.Fatal("a non-red, non-green permanent got the upkeep trigger")
		}
		return
	}
	if pay {
		effectstest.DcMana(t, g, me.ID, mana)
	}
	effectstest.AnswerPayUnless(t, g, me.ID, pay)
	if got := g.Battlefield.Contains(host); got != wantStays {
		t.Errorf("paid %v: host on the battlefield = %v, want %v", pay, got, wantStays)
	}
}

func TestB661EssenceLeakRedPermanentPaysOrIsSacrificed(t *testing.T) {
	b661RunLeak(t, true, []string{"R"}, "{1}{R}", "{R}{R}", true)
	b661RunLeak(t, false, []string{"R"}, "{1}{R}", "{R}{R}", false)
}

func TestB661EssenceLeakGreenPermanentPaysOrIsSacrificed(t *testing.T) {
	b661RunLeak(t, true, []string{"G"}, "{2}{G}", "{G}{G}{G}", true)
	b661RunLeak(t, false, []string{"G"}, "{2}{G}", "{G}{G}{G}", false)
}

func TestB661EssenceLeakIgnoresOtherColors(t *testing.T) {
	b661RunLeak(t, false, []string{"W"}, "{1}{W}", "{W}{W}", true)
}
