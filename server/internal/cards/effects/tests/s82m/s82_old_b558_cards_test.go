package testss82m

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b558_cards_test.go — Coldsnap cumulative upkeep cards, S82 OLD batch
// 558, played through the real catalog entries. The permanent's controller is
// seat 1 and the upkeep walked to is theirs.

const (
	oldB558OwlKeeper  = "3acef3dc-61ee-48e0-9b2e-a8e29d5f4ac7"
	oldB558Fallen     = "3f60a7ce-c714-4e87-a2d9-3710f2d610fa"
	oldB558Ancient    = "4554ce06-4992-4689-8a40-558a1545b968"
	oldB558Whispers   = "4baf3db8-6f35-4101-9b21-15f4384c249a"
	oldB558Nishoba    = "66ae3349-ad37-47da-8ccc-3e1cdddb32a9"
	oldB558Shards     = "8fab68ad-169d-46d3-93c4-5bdee4eea2ce"
	oldB558Javelineer = "95921d1e-0487-4274-b797-8652e94fc4bc"
	oldB558Core       = "b9526bd5-ed83-466f-8819-e9a8487bbe77"
	oldB558End        = "cf032b80-00d5-42fc-9e1a-16072b5fed55"
)

func TestOldB558CardsAreFullWithNoCaveats(t *testing.T) {
	for id, name := range map[string]string{
		oldB558OwlKeeper: "Jötun Owl Keeper", oldB558Fallen: "Balduvian Fallen",
		oldB558Ancient: "Sheltering Ancient", oldB558Whispers: "Krovikan Whispers",
		oldB558Nishoba: "Arctic Nishoba", oldB558Shards: "Wall of Shards",
		oldB558Javelineer: "Kjeldoran Javelineer", oldB558Core: "Magmatic Core",
		oldB558End: "Hibernation's End",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// oldB558Upkeep walks to seat 1's upkeep with the permanent carrying `seed` age
// counters and returns the open payment prompt (the upkeep adds one more).
func oldB558Upkeep(t *testing.T, g *game.Game, id uuid.UUID, seed int) *game.PendingChoice {
	t.Helper()
	if seed > 0 {
		effectstest.AddCountersForTest(g, id, game.CounterAge, seed)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	return effectstest.RemoraUpkeepPrompt(t, g, g.Seats[1].ID)
}

func oldB558Mana(p *game.Player, colors ...string) {
	for _, c := range colors {
		p.ManaPool.AddMana(game.ManaToken{Color: c})
	}
}

func oldB558Birds(g *game.Game, owner uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.IsToken() && c.Name == "Bird" {
			n++
		}
	}
	return n
}

func TestOldB558OwlKeeperHybridUpkeepAndABirdPerAgeCounterWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	keeper := effectstest.B12Push(g, me.ID, "Jötun Owl Keeper", "Creature — Giant", oldB558OwlKeeper, 3, 3)
	prompt := oldB558Upkeep(t, g, keeper, 1)
	if prompt.PayCost != "{W/U}{W/U}" {
		t.Fatalf("cost with two age counters = %q, want {W/U}{W/U}", prompt.PayCost)
	}
	if err := g.ResolvePayUnless(prompt.ID, me.ID, false); err != nil {
		t.Fatalf("decline: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(keeper) {
		t.Fatal("declining did not sacrifice the Keeper")
	}
	if got := oldB558Birds(g, me.ID); got != 2 {
		t.Errorf("Birds = %d, want 2 (one per age counter)", got)
	}
}

func TestOldB558OwlKeeperPaidWithEitherColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	keeper := effectstest.B12Push(g, me.ID, "Jötun Owl Keeper", "Creature — Giant", oldB558OwlKeeper, 3, 3)
	prompt := oldB558Upkeep(t, g, keeper, 1)
	oldB558Mana(me, "W", "U")
	if err := g.ResolvePayUnless(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("pay {W}{U}: %v", err)
	}
	if !g.Battlefield.Contains(keeper) {
		t.Error("a paid upkeep sacrificed the Keeper")
	}
	if got := oldB558Birds(g, me.ID); got != 0 {
		t.Errorf("Birds = %d, want none while it lives", got)
	}
}

func TestOldB558BalduvianFallenPumpsPerBlackOrRedSpent(t *testing.T) {
	for _, tc := range []struct {
		name  string
		mana  string
		power int
	}{
		{"black", "B", 4},
		{"red", "R", 4},
		{"white", "W", 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[1]
			fallen := effectstest.B12Push(g, me.ID, "Balduvian Fallen", "Creature — Zombie", oldB558Fallen, 3, 5)
			prompt := oldB558Upkeep(t, g, fallen, 0)
			if prompt.PayCost != "{1}" {
				t.Fatalf("cost = %q, want {1}", prompt.PayCost)
			}
			oldB558Mana(me, tc.mana)
			if err := g.ResolvePayUnless(prompt.ID, me.ID, true); err != nil {
				t.Fatalf("pay: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.EffectivePower(t, g, fallen); got != tc.power {
				t.Errorf("power after paying with {%s} = %d, want %d", tc.mana, got, tc.power)
			}
		})
	}
}

func TestOldB558ShelteringAncientPutsACounterOnAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[2]
	ancient := effectstest.B12Push(g, me.ID, "Sheltering Ancient", "Creature — Treefolk", oldB558Ancient, 5, 5)
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Their Bear", "Creature — Bear", "", false)
	prompt := oldB558Upkeep(t, g, ancient, 0)
	if a := prompt.PayAction(); a == nil || a.Kind != game.PayActionCounterOnOpponentCreature || a.Count != 1 {
		t.Fatalf("payment = %+v, want one counter on an opponent's creature", a)
	}
	if err := g.ResolvePayUnlessWithCards(prompt.ID, me.ID, true, []uuid.UUID{theirs}); err != nil {
		t.Fatalf("pay: %v", err)
	}
	if c, ok := effectstest.BattlefieldCard(g, theirs); !ok || c.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("their creature's +1/+1 counters = %v, want 1", c.Counters)
	}
	if !g.Battlefield.Contains(ancient) {
		t.Error("a paid upkeep sacrificed the Ancient")
	}
}

func TestOldB558ShelteringAncientIsSacrificedWhenDeclined(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[2]
	ancient := effectstest.B12Push(g, me.ID, "Sheltering Ancient", "Creature — Treefolk", oldB558Ancient, 5, 5)
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Their Bear", "Creature — Bear", "", false)
	prompt := oldB558Upkeep(t, g, ancient, 0)
	if err := g.ResolvePayUnless(prompt.ID, me.ID, false); err != nil {
		t.Fatalf("decline: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ancient) {
		t.Error("declining did not sacrifice the Ancient")
	}
	if c, _ := effectstest.BattlefieldCard(g, theirs); c.Counters[game.CounterPlusOne] != 0 {
		t.Error("a declined upkeep still put a counter on their creature")
	}
}

func TestOldB558KrovikanWhispersStealsAndCostsTwoLifePerAgeCounterWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[2]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	aura := effectstest.PushCatalogPermanent(g, me.ID, "Krovikan Whispers", "Enchantment — Aura", oldB558Whispers, false)
	effectstest.BrAttach(t, g, aura, bear)
	if got := effectstest.LayeredCard(t, g, bear).Controller; got != me.ID {
		t.Fatalf("the creature's controller = %s, want the Aura's controller", got)
	}
	prompt := oldB558Upkeep(t, g, aura, 1)
	if prompt.PayCost != "{U/B}{U/B}" {
		t.Fatalf("cost with two age counters = %q, want {U/B}{U/B}", prompt.PayCost)
	}
	life := effectstest.LifeOf(g, me.ID)
	if err := g.ResolvePayUnless(prompt.ID, me.ID, false); err != nil {
		t.Fatalf("decline: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Fatal("declining did not sacrifice the Aura")
	}
	if got := effectstest.LifeOf(g, me.ID); got != life-4 {
		t.Errorf("life %d → %d, want -4 (2 per age counter)", life, got)
	}
	if got := effectstest.LayeredCard(t, g, bear).Controller; got != opp.ID {
		t.Error("the creature did not go back to its owner")
	}
}

func TestOldB558ArcticNishobaGainsTwoLifePerAgeCounterWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	nishoba := effectstest.B12Push(g, me.ID, "Arctic Nishoba", "Creature — Cat Warrior", oldB558Nishoba, 6, 6)
	prompt := oldB558Upkeep(t, g, nishoba, 2)
	if prompt.PayCost != "{G/W}{G/W}{G/W}" {
		t.Fatalf("cost with three age counters = %q, want {G/W}{G/W}{G/W}", prompt.PayCost)
	}
	life := effectstest.LifeOf(g, me.ID)
	if err := g.ResolvePayUnless(prompt.ID, me.ID, false); err != nil {
		t.Fatalf("decline: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != life+6 {
		t.Errorf("life %d → %d, want +6", life, got)
	}
}

func TestOldB558WallOfShardsLetsAnOpponentGainLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[2]
	wall := effectstest.B12Push(g, me.ID, "Wall of Shards", "Snow Creature — Wall", oldB558Shards, 1, 8)
	prompt := oldB558Upkeep(t, g, wall, 1)
	life := effectstest.LifeOf(g, opp.ID)
	if err := g.ResolvePayUnlessWithCards(prompt.ID, me.ID, true, []uuid.UUID{me.ID, opp.ID}); err == nil {
		t.Error("named the payer as the opponent")
	}
	if err := g.ResolvePayUnlessWithCards(prompt.ID, me.ID, true, []uuid.UUID{opp.ID, opp.ID}); err != nil {
		t.Fatalf("name an opponent twice for two age counters: %v", err)
	}
	if got := effectstest.LifeOf(g, opp.ID); got != life+2 {
		t.Errorf("their life %d → %d, want +2", life, got)
	}
	if !g.Battlefield.Contains(wall) {
		t.Error("a paid upkeep sacrificed the Wall")
	}
}

func TestOldB558JavelineerDealsDamageEqualToItsAgeCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jav := effectstest.B21Push(g, me.ID, "Kjeldoran Javelineer", "Creature — Human Soldier", oldB558Javelineer, 1, 2, "W")
	effectstest.AddCountersForTest(g, jav, game.CounterAge, 2)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, opp.ID, "Idle", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, me.ID, small)
	if err := g.ActivateCatalogAbility(me.ID, jav, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(idle)}); err == nil {
		t.Fatal("a creature that is neither attacking nor blocking is not a legal target")
	}
	effectstest.B16Activate(t, g, me.ID, jav, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(small)})
	if g.Battlefield.Contains(small) {
		t.Error("two age counters' damage did not kill the 2/2 attacker")
	}
}

func TestOldB558JavelineerWithNoAgeCountersDealsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jav := effectstest.B21Push(g, me.ID, "Kjeldoran Javelineer", "Creature — Human Soldier", oldB558Javelineer, 1, 2, "W")
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, me.ID, small)
	effectstest.B16Activate(t, g, me.ID, jav, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(small)})
	if c, ok := effectstest.BattlefieldCard(g, small); !ok || c.DamageMarked != 0 {
		t.Errorf("a Javelineer with no age counters dealt damage (%v)", c.DamageMarked)
	}
}

func TestOldB558MagmaticCoreDividesItsAgeCountersAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[2]
	core := effectstest.PushCatalogPermanent(g, me.ID, "Magmatic Core", "Enchantment", oldB558Core, false)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 5)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Wall", 0, 5)
	prompt := oldB558Upkeep(t, g, core, 2)
	oldB558Mana(me, "R", "R", "R")
	if err := g.ResolvePayUnless(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("pay: %v", err)
	}
	effectstest.AdvanceToEndStepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the end step trigger asks for targets")
	}
	refs := effectstest.CardRefs(a, b)
	if err := g.ResolvePickTargetsDivided(p.ID, me.ID, refs, map[uuid.UUID]int{a: 2, b: 2}); err == nil {
		t.Error("four damage was divided for three age counters")
	}
	if err := g.ResolvePickTargetsDivided(p.ID, me.ID, refs, map[uuid.UUID]int{a: 1, b: 2}); err != nil {
		t.Fatalf("a legal division of 3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if da, db := effectstest.E2Card(t, g, a).DamageMarked, effectstest.E2Card(t, g, b).DamageMarked; da != 1 || db != 2 {
		t.Errorf("damage = %d and %d, want 1 and 2", da, db)
	}
}

func TestOldB558HibernationsEndFetchesACreatureWithManaValueEqualToItsAgeCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	end := effectstest.PushCatalogPermanent(g, me.ID, "Hibernation's End", "Enchantment", oldB558End, false)
	for _, c := range []struct{ name, cost string }{{"Two Drop", "{1}{G}"}, {"Three Drop", "{2}{G}"}} {
		card := game.NewCard(c.name, me.ID)
		card.TypeLine, card.ManaCost, card.Power, card.Toughness = "Creature — Bear", c.cost, 2, 2
		me.Library.PushTop(card)
	}
	prompt := oldB558Upkeep(t, g, end, 1)
	oldB558Mana(me, "C", "C")
	if err := g.ResolvePayUnless(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("pay: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchNamed(t, g, me.ID, "Two Drop")
	effectstest.PassPriorityAroundTable(t, g)
	if len(effectstest.BattlefieldIDsNamed(g, "Two Drop")) != 1 {
		t.Error("the creature with mana value 2 did not enter the battlefield")
	}
	if len(effectstest.BattlefieldIDsNamed(g, "Three Drop")) != 0 {
		t.Error("a creature with another mana value entered")
	}
}
