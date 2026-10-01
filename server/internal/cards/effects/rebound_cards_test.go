package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rebound_cards_test.go — the printed rebound cards (#1854, ADR 0107
// §3). The mechanic is pinned in game/rebound_test.go; what is pinned
// here is that a catalog declaration reaches it, and that the cards
// with rebound declare it.

// reboundOracleIDs is every printed rebound card in the catalog.
var reboundOracleIDs = map[string]string{
	"056c3b7d-b603-40b8-8404-18c2eb7e7129": "Staggershock",
	"7b58920e-3e90-44ed-a0e9-4cb1b7359a5b": "Artful Maneuver",
	"67f6e949-14d9-4253-9d18-4004cac2f900": "Prey's Vengeance",
	"4c7f3fbf-6b68-492f-893d-853c73cfa463": "Distortion Strike",
	"f256c828-01d2-4077-823d-8dccd15120bd": "Taigam's Strike",
	"77f53f6a-36ab-43d1-b547-4578fd724864": "Virulent Swipe",
	"9a6e5834-a89c-44af-b297-1b8ccbfbc907": "Great Teacher's Decree",
	"69501650-ed48-4ebf-9287-b0338c5bc5d5": "Trumpeting Herd",
	"def45f3a-dba0-4d08-b086-5236d6e0edb1": "Ojutai's Summons",
	"e8ca5c7e-7d8a-4e86-91ee-2d3576f9fd5d": "Void Squall",
	"0fd57894-b917-41c8-a394-360d1d31b236": "Ephemerate",
	"c96d67fb-359c-4439-aa1b-598d9bca2880": "Invisible Force Field",
	"686b44ec-3446-4e1f-a15f-9d8557db6d70": "Center Soul",
	"894015d5-ee78-45df-8d37-53df963c59a6": "Emerge Unscathed",
	"2c09d296-5d15-449a-9b98-cf5056a17b91": "Ojutai's Breath",
	"ad9d969e-def5-45a3-b65b-0c776f62ef0e": "Into the Time Vortex",
	"17a34f8d-a80f-4331-8be5-06cbb9d10d7b": "Terramorph",
	"0366ccfd-c717-4ea2-8176-86e184f920f4": "Faithless Salvaging",
	"ca29588c-f117-418e-be9e-fa2ee89862ca": "Blessed Reincarnation",
	"489a60f1-83f8-465b-918f-7d63d4f76d14": "Blossoming Calm",
	"a10ce333-48d7-499b-9355-f381f2395497": "Consuming Vapors",
	"67af9477-82b2-459e-8c09-b9a93a1d1c94": "Fantastic Elasticity",
	"51545f6e-d0e3-4b4c-bc73-c731f85e26b0": "Fevered Suspicion",
	"e715d33e-3d60-4623-b843-db0b906ae98b": "Flame On!",
	"4a75ef46-c5ae-4b1a-a681-d439b01f6731": "It's Clobberin' Time!",
	"a9a0088a-3c86-4baf-9754-79006f733ce1": "Nomads' Assembly",
	"070e3224-0f89-4716-b91b-0131eff6146f": "Profound Journey",
	"197582b9-4c86-41a3-ad0c-789a0db8e087": "Quantum Misalignment",
	"c6815bbb-24a5-4c9f-bf30-6190bc766d05": "Recurring Insight",
	"c2905e12-8af1-46f4-b888-be272a9b9748": "Sight Beyond Sight",
	"fa482ded-b24c-4f9e-948e-d0620e907b8b": "Surreal Memoir",
	"5fb8be5a-3666-4680-84e2-341cb269df07": "Survival Cache",
	"77d5f298-06b4-49d1-9b33-f1f176665ba2": "Transpose",
}

// Every card on the list declares the keyword, and so has the probe
// coverage reads.
func TestReboundCardsDeclareTheKeyword(t *testing.T) {
	for id, name := range reboundOracleIDs {
		spec, ok := Lookup(id)
		if !ok {
			t.Errorf("%s (%s) is not in the catalog", name, id)
			continue
		}
		found := false
		for _, kw := range spec.PrintedKeywords {
			if kw == game.KeywordRebound {
				found = true
			}
		}
		if !found {
			t.Errorf("%s does not declare rebound", name)
		}
	}
}

// The whole loop through a real catalog card: Staggershock cast from
// hand deals 2, is exiled, comes back at its controller's next upkeep
// for free and deals 2 again, then goes to the graveyard.
func TestStaggershockReboundsAndDealsDamageTwice(t *testing.T) {
	const oracle = "056c3b7d-b603-40b8-8404-18c2eb7e7129"
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	victim := g.Seats[(seat+1)%len(g.Seats)]
	advanceTo(t, g, game.StepPrecombatMain)
	id := handCardForTest(me, "Staggershock", "Instant", oracle)
	life := victim.Life
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}},
	}); err != nil {
		t.Fatalf("cast from hand: %v", err)
	}
	passPriorityAroundTable(t, g)
	if victim.Life != life-2 {
		t.Fatalf("first cast: victim at %d, want %d", victim.Life, life-2)
	}
	if !g.Exile.Contains(id) {
		t.Fatal("Staggershock was not exiled by rebound")
	}

	advanceToUpkeepOf(t, g, (seat+1)%len(g.Seats))
	advanceToUpkeepOf(t, g, seat)
	passPriorityAroundTable(t, g)
	offer := latestChoiceOfKind(g, game.PendingChoiceMayCast)
	if offer == nil {
		t.Fatal("no rebound offer at the controller's next upkeep")
	}
	if err := g.ResolveMayCast(offer.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveMayCast: %v", err)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Strict:   true,
		FromZone: "exile",
		Targets:  []game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}},
	}); err != nil {
		t.Fatalf("the free rebound cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if victim.Life != life-4 {
		t.Errorf("rebound cast: victim at %d, want %d", victim.Life, life-4)
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the rebound recast did not go to the graveyard")
	}
}

// castFromHandAtMain seeds a catalog card in the active seat's hand at
// its main phase and casts it.
func castFromHandAtMain(t *testing.T, g *game.Game, name, typeLine, oracle string, targets ...game.TargetRef) (*game.Player, uuid.UUID) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	advanceTo(t, g, game.StepPrecombatMain)
	id := handCardForTest(me, name, typeLine, oracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: targets}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return me, id
}

// Faithless Salvaging is a rummage: the draw waits for the discard.
func TestReboundFaithlessSalvagingDiscardsThenDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceTo(t, g, game.StepPrecombatMain)
	pitch := handCardForTest(me, "Pitch Me", "Sorcery", "")
	me, id := castFromHandAtMain(t, g, "Faithless Salvaging", "Instant", "0366ccfd-c717-4ea2-8176-86e184f920f4")
	library := me.Library.Size()
	passPriorityAroundTable(t, g)
	if me.Library.Size() != library {
		t.Fatal("Faithless Salvaging drew before the discard was chosen")
	}
	answerDiscard(t, g, me.ID, pitch)
	if !me.Graveyard.Contains(pitch) {
		t.Error("the chosen card was not discarded")
	}
	if me.Library.Size() != library-1 {
		t.Errorf("library %d after the discard, want %d", me.Library.Size(), library-1)
	}
	if !g.Exile.Contains(id) {
		t.Error("Faithless Salvaging was not exiled by rebound")
	}
}

// Center Soul's colour is chosen as it resolves and becomes the
// creature's protection.
func TestReboundCenterSoulGivesProtectionFromTheChosenColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushCatalogPermanent(g, me.ID, "Grizzly Bears", "Creature — Bear", "", false)
	_, id := castFromHandAtMain(t, g, "Center Soul", "Instant", "686b44ec-3446-4e1f-a15f-9d8557db6d70",
		game.TargetRef{Kind: game.TargetCard, ID: bear})
	for i := 0; i < 8 && pendingOfKind(g, game.PendingChoiceColor) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	answerColor(t, g, me.ID, "R")
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if c, _ := battlefieldCard(g, bear); !game.HasKeyword(&c, game.ProtectionFromColor("R")) {
		t.Error("the creature did not gain protection from red")
	}
	if !g.Exile.Contains(id) {
		t.Error("Center Soul was not exiled by rebound")
	}
}

// Distortion Strike pumps and makes the creature unblockable.
func TestReboundDistortionStrikePumpsAndMakesUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushCatalogPermanent(g, me.ID, "Grizzly Bears", "Creature — Bear", "", false)
	castFromHandAtMain(t, g, "Distortion Strike", "Sorcery", "4c7f3fbf-6b68-492f-893d-853c73cfa463",
		game.TargetRef{Kind: game.TargetCard, ID: bear})
	passPriorityAroundTable(t, g)
	eff := effectiveOf(t, g, bear)
	if eff.Power != 2 || eff.Toughness != 1 {
		t.Errorf("bear is %d/%d, want 2/1", eff.Power, eff.Toughness)
	}
	if eff.Restrictions&game.CantBeBlocked == 0 {
		t.Error("the creature can be blocked")
	}
}

// Into the Time Vortex cascades on the rebound cast too: cascade
// triggers on every cast (CR 702.85a), and the upkeep cast is one.
func TestReboundIntoTheTimeVortexCascadesOnItsReboundCast(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, id := castFromHandAtMain(t, g, "Into the Time Vortex", "Sorcery", "ad9d969e-def5-45a3-b65b-0c776f62ef0e")
	for i := 0; i < 32 && !g.Exile.Contains(id); i++ {
		if offer := latestChoiceOfKind(g, game.PendingChoiceMayCast); offer != nil {
			if err := g.ResolveMayCast(offer.ID, me.ID, false); err != nil {
				t.Fatalf("decline cascade: %v", err)
			}
			continue
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !g.Exile.Contains(id) {
		t.Fatal("Into the Time Vortex was not exiled by rebound")
	}
	cascades := countEvents(g, game.EventCast)
	advanceToUpkeepOf(t, g, (seat+1)%len(g.Seats))
	advanceToUpkeepOf(t, g, seat)
	passPriorityAroundTable(t, g)
	offer := latestChoiceOfKind(g, game.PendingChoiceMayCast)
	if offer == nil || offer.MayCastKeyword != game.MayCastKeywordRebound {
		t.Fatal("no rebound offer at the controller's next upkeep")
	}
	if err := g.ResolveMayCast(offer.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveMayCast: %v", err)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("the free rebound cast: %v", err)
	}
	if countEvents(g, game.EventCast) != cascades+1 {
		t.Fatal("the rebound cast was not a cast")
	}
	found := false
	for _, it := range g.StackMeta {
		if it != nil && it.SourceCardID == id && it.Kind == game.StackItemTriggered {
			found = true
		}
	}
	for _, it := range g.PendingTriggers {
		if it != nil && it.SourceCardID == id {
			found = true
		}
	}
	if !found {
		t.Error("the rebound cast did not trigger cascade")
	}
}
