package testss82m

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b559_cards_test.go — Coldsnap cumulative upkeep cards, S82 OLD batch
// 559, played through the real catalog entries. The permanent's controller is
// seat 1 and the upkeep walked to is theirs.

const (
	oldB559Etchings  = "cf938a60-23ec-4834-a3cf-391a7b04747e"
	oldB559Hulk      = "d3fb83af-8fc4-4f26-b2df-be5dd3ac56b2"
	oldB559Radiance  = "d7f28d00-2ec3-4d4e-b3a6-89d33c66da28"
	oldB559Phantasm  = "df680ccd-7e76-4a05-9bdf-4f437b4bfbb3"
	oldB559Plating   = "e9ee4257-5df8-4f4e-b821-b6313ded1802"
	oldB559Survivor  = "ea96c8c6-b77b-4fc0-8b11-6b31e1ed9554"
	oldB559Herald    = "ed9e39ad-996c-4a68-8d11-e1c7d35f556c"
	oldB559EarthenGo = "f52c0ef9-ca35-47bd-9359-01e7b97ee770"
)

func TestOldB559CardsAreFullWithNoCaveats(t *testing.T) {
	for id, name := range map[string]string{
		oldB559Etchings: "Phyrexian Etchings", oldB559Hulk: "Ronom Hulk",
		oldB559Radiance: "Freyalise's Radiance", oldB559Phantasm: "Phobian Phantasm",
		oldB559Plating: "Glacial Plating", oldB559Survivor: "Survivor of the Unseen",
		oldB559Herald: "Herald of Leshrac", oldB559EarthenGo: "Earthen Goo",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// oldB559Upkeep walks to seat 1's upkeep with the permanent carrying `seed` age
// counters and returns the open payment prompt (the upkeep adds one more).
func oldB559Upkeep(t *testing.T, g *game.Game, id uuid.UUID, seed int) *game.PendingChoice {
	t.Helper()
	if seed > 0 {
		effectstest.AddCountersForTest(g, id, game.CounterAge, seed)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	return effectstest.RemoraUpkeepPrompt(t, g, g.Seats[1].ID)
}

func oldB559Mana(p *game.Player, colors ...string) {
	for _, c := range colors {
		p.ManaPool.AddMana(game.ManaToken{Color: c})
	}
}

func TestOldB559EtchingsDrawsAndLosesLifePerAgeCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	etch := effectstest.PushCatalogPermanent(g, me.ID, "Phyrexian Etchings", "Enchantment", oldB559Etchings, false)
	for i := 0; i < 4; i++ {
		me.Library.PushTop(game.NewCard("Filler", me.ID))
	}
	prompt := oldB559Upkeep(t, g, etch, 1)
	if prompt.PayCost != "{B}{B}" {
		t.Fatalf("cost with two age counters = %q, want {B}{B}", prompt.PayCost)
	}
	oldB559Mana(me, "B", "B")
	if err := g.ResolvePayUnless(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("pay: %v", err)
	}
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	effectstest.AdvanceToEndStepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d → %d at the end step, want +2 (one per age counter)", hand, got)
	}
	// The next upkeep: decline, and lose 2 life for each of the three age counters.
	effectstest.LeaveUpkeepThenAdvanceTo(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	prompt = effectstest.RemoraUpkeepPrompt(t, g, me.ID)
	life := effectstest.LifeOf(g, me.ID)
	if err := g.ResolvePayUnless(prompt.ID, me.ID, false); err != nil {
		t.Fatalf("decline: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(etch) {
		t.Fatal("declining did not sacrifice the Etchings")
	}
	if got := effectstest.LifeOf(g, me.ID); got != life-6 {
		t.Errorf("life %d → %d, want -6 (2 per age counter)", life, got)
	}
}

func TestOldB559RonomHulkHasProtectionFromSnowAndAnUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	hulk := effectstest.B12Push(g, me.ID, "Ronom Hulk", "Creature — Beast", oldB559Hulk, 5, 6)
	if !effectstest.HasEffectiveKeyword(t, g, hulk, game.ProtectionFromSnow) {
		t.Error("the Hulk does not have protection from snow")
	}
	prompt := oldB559Upkeep(t, g, hulk, 0)
	if prompt.PayCost != "{1}" {
		t.Errorf("cost = %q, want {1}", prompt.PayCost)
	}
}

func TestOldB559FreyalisesRadianceKeepsSnowPermanentsTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[0]
	effectstest.PushPermanentForTest(g, g.Seats[0].ID, "Freyalise's Radiance", oldB559Radiance, "Enchantment")
	snow := effectstest.PushPermanentForTest(g, owner.ID, "Snow-Covered Forest", "", "Basic Snow Land — Forest")
	plain := effectstest.PushPermanentForTest(g, owner.ID, "Forest", "", "Basic Land — Forest")
	snowBear := effectstest.PushPermanentForTest(g, owner.ID, "Snow Bear", "", "Snow Creature — Bear")
	for _, id := range []uuid.UUID{snow, plain, snowBear} {
		if err := g.TapCard(id, true); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if !effectstest.UntapStaticTapped(t, g, snow) || !effectstest.UntapStaticTapped(t, g, snowBear) {
		t.Error("a snow permanent untapped")
	}
	if effectstest.UntapStaticTapped(t, g, plain) {
		t.Error("a nonsnow permanent stayed tapped")
	}
}

func TestOldB559PhobianPhantasmFliesWithFearAndCharges(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	ph := effectstest.B12Push(g, me.ID, "Phobian Phantasm", "Creature — Illusion", oldB559Phantasm, 3, 3)
	for _, kw := range []string{"flying", "fear"} {
		if !effectstest.HasEffectiveKeyword(t, g, ph, kw) {
			t.Errorf("no %s", kw)
		}
	}
	prompt := oldB559Upkeep(t, g, ph, 0)
	if prompt.PayCost != "{B}" {
		t.Errorf("cost = %q, want {B}", prompt.PayCost)
	}
}

func TestOldB559GlacialPlatingNeedsSnowManaAndPumpsPerAgeCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	aura := effectstest.PushCatalogPermanent(g, me.ID, "Glacial Plating", "Snow Enchantment — Aura", oldB559Plating, false)
	effectstest.BrAttach(t, g, aura, bear)
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Fatalf("power with no age counters = %d, want 2", p)
	}
	prompt := oldB559Upkeep(t, g, aura, 0)
	if prompt.PayCost != "{S}" {
		t.Fatalf("cost = %q, want {S}", prompt.PayCost)
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "G", Source: uuid.New(), SourceKinds: game.ManaSourceLand | game.ManaSourceSnow})
	if err := g.ResolvePayUnless(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("pay with snow mana: %v", err)
	}
	if !g.Battlefield.Contains(aura) {
		t.Fatal("a payment of snow mana sacrificed the Aura")
	}
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("enchanted creature %d/%d with one age counter, want 5/5", p, tough)
	}
}

func TestOldB559GlacialPlatingIsSacrificedWhenOnlyOrdinaryManaIsOffered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	aura := effectstest.PushCatalogPermanent(g, me.ID, "Glacial Plating", "Snow Enchantment — Aura", oldB559Plating, false)
	effectstest.BrAttach(t, g, aura, bear)
	prompt := oldB559Upkeep(t, g, aura, 0)
	oldB559Mana(me, "G")
	// An unfunded payment degrades to a decline: ordinary mana is not snow mana.
	if err := g.ResolvePayUnless(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("answer: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("ordinary mana paid {S}")
	}
}

func TestOldB559SurvivorDrawsTwoThenPutsOneBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	surv := effectstest.B21Push(g, me.ID, "Survivor of the Unseen", "Creature — Human Wizard", oldB559Survivor, 2, 1, "U")
	for i := 0; i < 3; i++ {
		me.Library.PushTop(game.NewCard("Filler", me.ID))
	}
	effectstest.AdvanceToMain(t, g)
	hand, library := me.Hand.Size(), me.Library.Size()
	effectstest.B16Activate(t, g, me.ID, surv, 0, game.ActivateAbilityParams{})
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no put-back prompt")
	}
	if pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Errorf("put-back bounds = %d..%d, want exactly one", pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{me.Hand.Cards[0].InstanceID}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d, want +1 (two drawn, one put back)", hand, got)
	}
	if got := me.Library.Size(); got != library-1 {
		t.Errorf("library %d → %d, want -1", library, got)
	}
}

func TestOldB559HeraldTakesLandsPumpsAndGivesThemBackWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[2]
	herald := effectstest.B12Push(g, me.ID, "Herald of Leshrac", "Creature — Avatar", oldB559Herald, 2, 4)
	mine := effectstest.PushPermanentForTest(g, me.ID, "Swamp", "", "Basic Land — Swamp")
	theirs := effectstest.PushPermanentForTest(g, opp.ID, "Forest", "", "Basic Land — Forest")
	other := effectstest.PushPermanentForTest(g, opp.ID, "Island", "", "Basic Land — Island")
	prompt := oldB559Upkeep(t, g, herald, 0)
	if a := prompt.PayAction(); a == nil || a.Kind != game.PayActionGainControlOfLand || a.Count != 1 {
		t.Fatalf("payment = %+v, want one land you don't control", a)
	}
	if err := g.ResolvePayUnlessWithCards(prompt.ID, me.ID, true, []uuid.UUID{mine}); err == nil {
		t.Error("named a land the payer already controls")
	}
	if err := g.ResolvePayUnlessWithCards(prompt.ID, me.ID, true, []uuid.UUID{theirs}); err != nil {
		t.Fatalf("pay: %v", err)
	}
	if got := effectstest.LayeredCard(t, g, theirs).Controller; got != me.ID {
		t.Fatalf("the Forest's controller = %s, want the Herald's controller", got)
	}
	if p, tough := effectstest.EffectivePower(t, g, herald), effectstest.EffectiveToughness(t, g, herald); p != 3 || tough != 5 {
		t.Errorf("Herald %d/%d with one land taken, want 3/5", p, tough)
	}
	// Decline the next upkeep: the Herald is sacrificed and the land goes home.
	effectstest.LeaveUpkeepThenAdvanceTo(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	prompt = effectstest.RemoraUpkeepPrompt(t, g, me.ID)
	if err := g.ResolvePayUnless(prompt.ID, me.ID, false); err != nil {
		t.Fatalf("decline: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(herald) {
		t.Fatal("declining did not sacrifice the Herald")
	}
	if got := effectstest.LayeredCard(t, g, theirs).Controller; got != opp.ID {
		t.Error("the Forest did not go back to its owner")
	}
	if got := effectstest.LayeredCard(t, g, mine).Controller; got != me.ID {
		t.Error("the Herald's controller lost their own land")
	}
	if got := effectstest.LayeredCard(t, g, other).Controller; got != opp.ID {
		t.Error("an untouched land changed hands")
	}
}

func TestOldB559EarthenGooGrowsWithItsAgeCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	goo := effectstest.B12Push(g, me.ID, "Earthen Goo", "Creature — Ooze", oldB559EarthenGo, 2, 2)
	prompt := oldB559Upkeep(t, g, goo, 1)
	if prompt.PayCost != "{R/G}{R/G}" {
		t.Fatalf("cost with two age counters = %q, want {R/G}{R/G}", prompt.PayCost)
	}
	oldB559Mana(me, "R", "G")
	if err := g.ResolvePayUnless(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("pay {R}{G}: %v", err)
	}
	if p, tough := effectstest.EffectivePower(t, g, goo), effectstest.EffectiveToughness(t, g, goo); p != 4 || tough != 4 {
		t.Errorf("Goo %d/%d with two age counters, want 4/4", p, tough)
	}
}
