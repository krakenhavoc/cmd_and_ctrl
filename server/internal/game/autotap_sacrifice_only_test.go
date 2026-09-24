package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// autotap_sacrifice_only_test.go — #1242. The auto-tapper planned a
// mana ability that sacrifices its source only when the cost ALSO had
// a {T}: a Treasure came through, a Gold token, an Eldrazi Spawn and an
// Eldrazi Scion did not, because their printed cost is the sacrifice
// alone. A board of Spawn read as unpayable to the cast preview, the
// strict gate and the legal-move enumerator.
//
// What these hold:
//
//   - the planner plans them, including when something else has
//     tapped them (a sacrifice never asks);
//   - the executor cracks them WITHOUT tapping them — no EventTapCard,
//     no CR 106.12a "tapped for mana", no triggered mana ability;
//   - a creature the cost eats is the last sacrifice the planner
//     reaches for, in both comparators;
//   - a permanent a cast or activation has already NAMED to pay
//     another cost component is not also spent on its mana.

// goldAbility is the printed Gold token: "Sacrifice this artifact: Add
// one mana of any color."
func goldAbility() []ManaAbilityShape {
	return []ManaAbilityShape{{
		SacrificeCost: true,
		Produced:      "{W|U|B|R|G}",
		Label:         "Sacrifice this artifact: Add one mana of any color",
	}}
}

// spawnAbility is the printed Eldrazi Spawn / Scion: "Sacrifice this
// creature: Add {C}."
func spawnAbility() []ManaAbilityShape {
	return []ManaAbilityShape{{
		SacrificeCost: true,
		Produced:      "{C}",
		Label:         "Sacrifice this creature: Add {C}",
	}}
}

func pushGold(g *Game, owner *Player) uuid.UUID {
	return pushIntrinsicPermanent(g, owner, "Gold", "Token Artifact — Gold", goldAbility(), nil)
}

func pushSpawn(g *Game, owner *Player) uuid.UUID {
	return pushIntrinsicPermanent(g, owner, "Eldrazi Spawn", "Token Creature — Eldrazi Spawn", spawnAbility(), nil)
}

func setTappedForTest(g *Game, id uuid.UUID) {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].Tapped = true
		}
	}
}

// The headline, the shape the issue was reported in: three Eldrazi
// Spawn and nothing else pay {3}.
func TestAutoTapPlansABoardOfEldraziSpawn(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	spawn := []uuid.UUID{pushSpawn(g, me), pushSpawn(g, me), pushSpawn(g, me)}

	plan, ok := g.AutoTapForCost(me.ID, costFor(t, "{3}"), 0)
	if !ok {
		t.Fatal("three Eldrazi Spawn should pay {3} — a sacrifice-only source is a mana source")
	}
	for _, id := range spawn {
		if !containsID(plan, id) {
			t.Errorf("Spawn %v missing from the plan %v", id, plan)
		}
	}
}

// A Gold is a five-colour source with no {T} and pays a coloured pip.
func TestAutoTapPlansAGoldForAColouredPip(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	gold := pushGold(g, me)

	plan, ok := g.AutoTapForCost(me.ID, costFor(t, "{G}"), 0)
	if !ok || len(plan) != 1 || plan[0] != gold {
		t.Fatalf("one Gold for {G}: ok=%v plan=%v, want the Gold", ok, plan)
	}
}

// A TAPPED Gold is still a source: nothing about "Sacrifice this
// artifact" asks it to be untapped. The planner used to skip every
// tapped permanent before it had even picked an ability.
func TestAutoTapPlansATappedGold(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	gold := pushGold(g, me)
	setTappedForTest(g, gold)

	plan, ok := g.AutoTapForCost(me.ID, costFor(t, "{1}"), 0)
	if !ok || len(plan) != 1 || plan[0] != gold {
		t.Fatalf("a tapped Gold for {1}: ok=%v plan=%v, want the Gold", ok, plan)
	}
}

// …while a tapped TREASURE is not — its cost owes a {T}. The tapped
// check moved; it did not go away.
func TestAutoTapStillSkipsATappedTreasure(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	treasure := pushTreasures(g, me, 1)[0]
	setTappedForTest(g, treasure)

	if plan, ok := g.AutoTapForCost(me.ID, costFor(t, "{1}"), 0); ok {
		t.Fatalf("planned %v off a tapped Treasure", plan)
	}
}

// CR 302.6 is about a {T} cost, and a Spawn has none: one made this
// turn is a mana source this turn, exactly as a hand-clicked one is.
func TestAutoTapPlansASummoningSickSpawn(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	spawn := pushSpawn(g, me)
	setSummonedThisTurn(g, spawn, true)

	plan, ok := g.AutoTapForCost(me.ID, costFor(t, "{1}"), 0)
	if !ok || len(plan) != 1 || plan[0] != spawn {
		t.Fatalf("a Spawn made this turn for {1}: ok=%v plan=%v", ok, plan)
	}
}

// The demand that replaced "must have a {T}": the ability has to cost
// the SOURCE something. A battlefield mana ability with neither a {T}
// nor a sacrifice of itself — here, a bare "Add {C}" — is not planned,
// because nothing would bound it.
func TestAutoTapRefusesAManaAbilityThatCostsTheSourceNothing(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	pushIntrinsicPermanent(g, me, "Free Rock", "Artifact",
		[]ManaAbilityShape{{Produced: "{C}", Label: "Add {C}"}}, nil)

	if plan, ok := g.AutoTapForCost(me.ID, costFor(t, "{1}"), 0); ok {
		t.Fatalf("planned %v off an ability that costs its source nothing", plan)
	}
}

// #1283: an exile-a-card cost is a decision — WHICH card leaves the
// hand — so the planner refuses it even when the ability also owes a
// {T} and would otherwise pass the demand above. (Cadaverous Bloom
// itself prints no {T}, so the demand alone keeps it out; this is the
// shape the exclusion exists for.)
func TestAutoTapRefusesAnExileACardCost(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	pushTypedCardToHandWithCost(me, "Spare Card", "Sorcery", "{4}")
	pushIntrinsicPermanent(g, me, "Tap-and-Pitch Rock", "Artifact",
		[]ManaAbilityShape{{
			TapCost:    true,
			ExileCards: &ExileCost{N: 1, Label: "a card"},
			Produced:   "{B}",
			Label:      "{T}, Exile a card from your hand: Add {B}",
		}}, nil)

	if plan, ok := g.AutoTapForCost(me.ID, costFor(t, "{B}"), 0); ok {
		t.Fatalf("planned %v — the planner would have chosen a card to exile", plan)
	}
}

// The executor's half. A Spawn is CRACKED, not tapped: it is
// sacrificed and its mana lands, and no EventTapCard is emitted — a
// "whenever this becomes tapped" watcher never sees a tap that never
// happened.
func TestMaterializePlanCracksASpawnWithoutTappingIt(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	spawn := pushSpawn(g, me)

	g.WithWriteLock(func() {
		g.materializePlanLocked(me, tapPlan{{CardID: spawn}}, costFor(t, "{1}"))
	})
	if g.Battlefield.Contains(spawn) {
		t.Error("the Spawn survived the payment")
	}
	if !hasEvent(g, EventSacrifice, spawn) {
		t.Error("no EventSacrifice for the Spawn")
	}
	if hasEvent(g, EventTapCard, spawn) {
		t.Error("the executor TAPPED a source whose cost prints no {T}")
	}
	if len(me.ManaPool) != 1 || me.ManaPool[0].Color != "C" {
		t.Fatalf("pool = %+v, want one {C}", me.ManaPool)
	}
}

// …and a tapped Gold is cracked too — the executor's tapped check is
// the planner's, through the same helper.
func TestMaterializePlanCracksATappedGold(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	gold := pushGold(g, me)
	setTappedForTest(g, gold)

	g.WithWriteLock(func() {
		g.materializePlanLocked(me, tapPlan{{CardID: gold}}, costFor(t, "{U}"))
	})
	if g.Battlefield.Contains(gold) {
		t.Fatal("the executor skipped a tapped Gold the planner may plan")
	}
	if len(me.ManaPool) != 1 || me.ManaPool[0].Color != "U" {
		t.Fatalf("pool = %+v, want one {U}", me.ManaPool)
	}
}

// End to end: a strict auto-tapped cast off three Spawn.
func TestCastAutoTapPaysAThreeCostSpellOffThreeSpawn(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	spawn := []uuid.UUID{pushSpawn(g, me), pushSpawn(g, me), pushSpawn(g, me)}
	spell := pushTypedCardToHandWithCost(me, "Icy Manipulator", "Artifact", "{3}")

	if err := g.CastSpell(me.ID, spell, CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("auto-tap cast off three Spawn: %v", err)
	}
	if !g.Stack.Contains(spell) {
		t.Error("the spell never reached the stack")
	}
	for _, id := range spawn {
		if g.Battlefield.Contains(id) {
			t.Errorf("Spawn %v survived a cast it paid for", id)
		}
	}
}

// --- the creature sub-tier -------------------------------------------

// The generic recruiter. A Spawn's {C} is the colourless tier and a
// Gold is the any-colour tier, so without the creature key the Spawn —
// a body on the board — is spent first. With it, the Gold goes.
func TestAutoTapCracksAGoldBeforeASpawnForGeneric(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	spawn := pushSpawn(g, me)
	gold := pushGold(g, me)

	plan, ok := g.AutoTapForCost(me.ID, costFor(t, "{1}"), 0)
	if !ok || len(plan) != 1 {
		t.Fatalf("Spawn + Gold for {1}: ok=%v plan=%v", ok, plan)
	}
	if plan[0] != gold {
		t.Errorf("plan = %v, want the Gold %v kept ahead of the Spawn %v", plan, gold, spawn)
	}
}

// The coloured pass. A creature that cracks for {R} is MORE restrictive
// than a Gold, so without the creature key the restrictiveness order
// would send the creature first.
func TestAutoTapCracksAGoldBeforeACreatureForAColouredPip(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	pushIntrinsicPermanent(g, me, "Red Spawn", "Token Creature — Eldrazi Spawn",
		[]ManaAbilityShape{{SacrificeCost: true, Produced: "{R}", Label: "Sacrifice: Add {R}"}}, nil)
	gold := pushGold(g, me)

	plan, ok := g.AutoTapForCost(me.ID, costFor(t, "{R}"), 0)
	if !ok || len(plan) != 1 || plan[0] != gold {
		t.Fatalf("creature + Gold for {R}: ok=%v plan=%v, want the Gold", ok, plan)
	}
}

// And the sub-tier is only a sub-tier: an untapped land still beats
// both, and the Spawn is reached when the Gold and the land are short.
func TestAutoTapReachesTheSpawnLast(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	mountain := pushBattlefieldForTest(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	gold := pushGold(g, me)
	spawn := pushSpawn(g, me)

	plan, ok := g.AutoTapForCost(me.ID, costFor(t, "{1}"), 0)
	if !ok || len(plan) != 1 || plan[0] != mountain {
		t.Fatalf("{1}: ok=%v plan=%v, want the Mountain alone", ok, plan)
	}
	plan, ok = g.AutoTapForCost(me.ID, costFor(t, "{3}"), 0)
	if !ok || len(plan) != 3 {
		t.Fatalf("{3}: ok=%v plan=%v, want all three", ok, plan)
	}
	if !containsID(plan, gold) || !containsID(plan, spawn) {
		t.Errorf("{3}: plan %v, want the Gold and the Spawn too", plan)
	}
}

// --- what the announcement already spent ----------------------------

// A Spawn named to an additional cost's sacrifice is not also a mana
// source for the same cast. Without the exclusion the plan cracks it
// for the {1}, and the sacrifice then finds nothing to pay with —
// after the mana was made.
func TestCastAutoTapDoesNotCrackTheSpawnNamedToTheSacrifice(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	spawn := pushSpawn(g, me)

	params := CastSpellParams{SacrificeIDs: []uuid.UUID{spawn}}
	var ok bool
	g.ReadSnapshot(func() {
		_, ok = g.autoTapLocked(me.ID, costFor(t, "{1}"), 0, CastAutoTapExclusions(params))
	})
	if ok {
		t.Fatal("the planner spent the Spawn the cast already named to its sacrifice")
	}
	// …and a second Spawn is still a source.
	other := pushSpawn(g, me)
	var plan tapPlan
	g.ReadSnapshot(func() {
		plan, ok = g.autoTapLocked(me.ID, costFor(t, "{1}"), 0, CastAutoTapExclusions(params))
	})
	if !ok || len(plan) != 1 || plan[0].CardID != other {
		t.Fatalf("plan = %v ok=%v, want the other Spawn", plan.cardIDs(), ok)
	}
}

// The activation path end to end: "{1}, Sacrifice a creature" naming
// the only Spawn, which is also the only possible {1}. The activation is
// refused as unpayable before anything moves, rather than cracking the
// Spawn for mana and then finding nothing to sacrifice.
func TestActivateAbilityDoesNotCrackTheSpawnItSacrifices(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[g.Turn.ActiveSeat]
	spawn := pushSpawn(g, me)
	altar := NewCard("Paid Altar", me.ID)
	altar.TypeLine = "Artifact"
	altar.ActivatedAbilities = []ActivatedAbilityShape{{
		Label: "{1}, Sacrifice a creature: You gain 1 life",
		Cost: AbilityCost{
			Mana:           "{1}",
			SacrificeOther: altarAbility()[0].SacrificeOther,
		},
		Effect: func(*Game, *StackItem) error { return nil },
	}}
	g.Battlefield.PushTop(altar)

	err := g.ActivateCatalogAbility(me.ID, altar.InstanceID, 0, ActivateAbilityParams{
		Strict: true, AutoTap: true, SacrificeIDs: []uuid.UUID{spawn},
	})
	var short *InsufficientManaError
	if !errors.As(err, &short) {
		t.Fatalf("ActivateCatalogAbility = %v, want an InsufficientManaError", err)
	}
	if !g.Battlefield.Contains(spawn) {
		t.Error("the refused activation cracked the Spawn anyway")
	}
}

// The activation twin: the source of an ability that sacrifices itself,
// and the permanents it names, are not mana for it.
func TestAbilityAutoTapExclusionsCoverTheSacrificedSource(t *testing.T) {
	src, named, pitched, tapped, exiled := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	got := AbilityAutoTapExclusions(src, AbilityCost{SacrificeSelf: true},
		[]uuid.UUID{tapped}, []uuid.UUID{named}, []uuid.UUID{pitched}, []uuid.UUID{exiled})
	for _, id := range []uuid.UUID{src, named, pitched, tapped, exiled} {
		if !got[id] {
			t.Errorf("exclusions %v miss %v", got, id)
		}
	}
	if AbilityAutoTapExclusions(src, AbilityCost{}, nil, nil, nil, nil) != nil {
		t.Error("an ability that spends nothing excluded something")
	}
}

// --- #1285: the plan, described --------------------------------------

// The preview's projection names where each planned source is and what
// paying with it costs: a land taps, a Gold is sacrificed without a
// tap, a Spirit Guide is exiled out of the HAND.
func TestAutoTapPlanDescribesEachSource(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	mountain := pushBattlefieldForTest(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	gold := pushGold(g, me)
	guide := seedSpiritGuide(me, "Simian Spirit Guide", "{R}")

	entries, ok := g.AutoTapPlanPreferringExcluding(me.ID, costFor(t, "{R}{R}{R}"), 0, nil, 0)
	if !ok || len(entries) != 3 {
		t.Fatalf("Mountain + Gold + Guide for {R}{R}{R}: ok=%v entries=%+v", ok, entries)
	}
	by := map[uuid.UUID]AutoTapPlanEntry{}
	for _, e := range entries {
		by[e.CardID] = e
	}
	if e := by[mountain]; e.Zone != ZoneBattlefield || !e.Taps || e.Sacrifices || e.Exiles || e.Name != "Mountain" {
		t.Errorf("Mountain described as %+v", e)
	}
	if e := by[gold]; e.Zone != ZoneBattlefield || e.Taps || !e.Sacrifices || e.Exiles {
		t.Errorf("Gold described as %+v, want sacrificed and not tapped", e)
	}
	if e := by[guide]; e.Zone != ZoneHand || e.Taps || e.Sacrifices || !e.Exiles || e.Name != "Simian Spirit Guide" {
		t.Errorf("Spirit Guide described as %+v, want exiled from the hand", e)
	}
}
