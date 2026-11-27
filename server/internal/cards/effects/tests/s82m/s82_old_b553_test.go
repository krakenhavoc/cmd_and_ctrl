package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b553_test.go — Lorwyn, Auras, batch 553 (#5593). Every test plays the
// real catalog entry through the engine.

const (
	oldB553SoaringHope = "0d09049a-6786-4b69-b448-35efaa926a71"
	oldB553Nettlevine  = "11132bbe-985d-44fa-a5b4-4b4e0dc3a7cf"
	oldB553Bubble      = "1fba543c-bd22-4f7b-9ff4-0631af8ee814"
	oldB553Triclopean  = "6441456a-cc20-48bb-95e9-c981fea8ea98"
	oldB553Nap         = "8b476e35-77df-4332-b29d-8bf6f1bbbc3e"
	oldB553ZephyrNet   = "cd510cf2-df54-4024-919b-e4377cdb2be1"
	oldB553Bear        = "Creature — Bear"
)

func TestOldB553EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB553SoaringHope: "Soaring Hope", oldB553Nettlevine: "Nettlevine Blight", oldB553Bubble: "Protective Bubble",
		oldB553Triclopean: "Triclopean Sight", oldB553Nap: "Glimmerdust Nap", oldB553ZephyrNet: "Zephyr Net",
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

func TestOldB553SoaringHopeGainsLifeGrantsFlyingAndGoesBackOnTheLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB553Bear, 2, 2)
	before := effectstest.LifeOf(g, me.ID)
	aura := effectstest.AuraCast(t, g, "Soaring Hope", oldB553SoaringHope, bear)
	if got := effectstest.LifeOf(g, me.ID); got != before+3 {
		t.Errorf("life %d, want %d", got, before+3)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the enchanted creature has no flying")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FloatForTest(g, me, "W")
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOfCard(g, aura) != game.ZoneLibrary {
		t.Errorf("the Aura is in %q, want the library", effectstest.ZoneOfCard(g, aura))
	}
	if top := me.Library.Cards[me.Library.Size()-1].InstanceID; top != aura {
		t.Error("the Aura is not on top of its owner's library")
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the creature kept flying after the Aura left")
	}
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated from the library")
	}
}

func TestOldB553SoaringHopeCantEnchantALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.SeedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Soaring Hope", effectstest.AuraTypeLine, oldB553SoaringHope, effectstest.CardTargets(land)); err == nil {
		t.Fatal("enchanted a land")
	}
}

func TestOldB553NettlevineBlightSacrificesTheHostAndMovesOn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB553Bear, 2, 2)
	elk := effectstest.B12Creature(g, me.ID, "Elk", oldB553Bear, 3, 3)
	land := effectstest.SeedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
	aura := effectstest.AuraCast(t, g, "Nettlevine Blight", oldB553Nettlevine, land)
	if !effectstest.AttachedToNow(t, g, aura, land) {
		t.Fatal("the Aura did not enchant the land")
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if p == nil {
		t.Fatal("no choice of a creature or land")
	}
	if err := g.ResolveOwnPermanents(p.ID, me.ID, []uuid.UUID{elk}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	if effectstest.ZoneOfCard(g, land) == game.ZoneBattlefield {
		t.Error("the enchanted land was not sacrificed")
	}
	if !effectstest.AttachedToNow(t, g, aura, elk) {
		t.Error("the Aura did not move to the chosen creature")
	}
	if effectstest.AttachedToNow(t, g, aura, bear) {
		t.Error("the Aura is on the wrong creature")
	}
}

func TestOldB553NettlevineBlightTriggersOnTheHostControllersEndStepOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB553Bear, 2, 2)
	effectstest.AuraCast(t, g, "Nettlevine Blight", oldB553Nettlevine, theirs)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOfCard(g, theirs) != game.ZoneBattlefield {
		t.Error("the host was sacrificed on the Aura controller's end step")
	}
	_ = me
}

func TestOldB553ProtectiveBubbleMakesTheHostUnblockableWithShroud(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB553Bear, 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other Bear", oldB553Bear, 2, 2)
	effectstest.AuraCast(t, g, "Protective Bubble", oldB553Bubble, bear)
	if effectstest.AuraRestrictions(t, g, bear)&game.CantBeBlocked == 0 {
		t.Error("the enchanted creature can be blocked")
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "shroud") {
		t.Error("the enchanted creature has no shroud")
	}
	if effectstest.AuraRestrictions(t, g, other)&game.CantBeBlocked != 0 || effectstest.HasEffectiveKeyword(t, g, other, "shroud") {
		t.Error("an unenchanted creature got the grant")
	}
}

func TestOldB553TriclopeanSightUntapsAndPumpsWithVigilance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB553Bear, 2, 2)
	effectstest.TapForTest(t, g, bear)
	effectstest.AuraCast(t, g, "Triclopean Sight", oldB553Triclopean, bear)
	if effectstest.TappedNow(g, bear) {
		t.Error("the enchanted creature is still tapped")
	}
	if p, tt := effectstest.PtOf(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("the creature is %d/%d, want 3/3", p, tt)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Error("the enchanted creature has no vigilance")
	}
	spec, _ := effects.Lookup(oldB553Triclopean)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
		t.Errorf("printed keywords %v, want flash", spec.PrintedKeywords)
	}
}

func TestOldB553GlimmerdustNapNeedsATappedCreatureAndKeepsItTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	untapped := effectstest.B12Creature(g, opp.ID, "Untapped Bear", oldB553Bear, 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Glimmerdust Nap", effectstest.AuraTypeLine, oldB553Nap, effectstest.CardTargets(untapped)); err == nil {
		t.Fatal("enchanted an untapped creature")
	}
	tapped := effectstest.B12Creature(g, opp.ID, "Tapped Bear", oldB553Bear, 2, 2)
	effectstest.TapForTest(t, g, tapped)
	effectstest.Enchant(t, g, "Glimmerdust Nap", oldB553Nap, tapped)
	effectstest.TapForTest(t, g, untapped)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	if !effectstest.TappedNow(g, tapped) {
		t.Error("the enchanted creature untapped")
	}
	if effectstest.TappedNow(g, untapped) {
		t.Error("the unenchanted creature stayed tapped")
	}
}

func TestOldB553ZephyrNetGivesDefenderAndFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB553Bear, 2, 2)
	effectstest.AuraCast(t, g, "Zephyr Net", oldB553ZephyrNet, bear)
	for _, kw := range []string{"defender", "flying"} {
		if !effectstest.HasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the enchanted creature has no %s", kw)
		}
	}
}
