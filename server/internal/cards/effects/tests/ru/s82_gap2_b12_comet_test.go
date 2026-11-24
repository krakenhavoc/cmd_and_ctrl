package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap2_b12_comet_test.go — S82 GAP2 batch 12: Comet, Stellar Pup,
// through the real loyalty activation, the dice door (with a fixed RNG
// key) and each row of the results table.

const gap2B12Comet = "e7c67480-1816-4071-9eab-a1d1b535a344"

// gap2B12CometID is Comet's instance ID: the roll stream is keyed by it.
var gap2B12CometID = uuid.MustParse("0c0e7000-0000-4000-8000-00000000c0e7")

func TestGap2B12CometIsFull(t *testing.T) {
	spec, ok := effects.Lookup(gap2B12Comet)
	if !ok {
		t.Fatal("Comet, Stellar Pup is not registered")
	}
	if spec.Name != "Comet, Stellar Pup" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
	}
}

// gap2B12CometTable is seat 0's main phase with Comet (loyalty 5) in play,
// under an RNG key whose first d6 rolls from Comet are ones `want` accepts.
func gap2B12CometTable(t *testing.T, n int, want func([]int) bool) (*game.Game, *game.Player) {
	t.Helper()
	key := gap2B12CometKey(t, n, want)
	g := effectstest.NewCatalogGame(t)
	g.SetRNGKeyForTest(key)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	me := g.Seats[0]
	g.Battlefield.PushTop(game.Card{
		InstanceID: gap2B12CometID, Name: "Comet, Stellar Pup", TypeLine: "Legendary Planeswalker — Comet", OracleID: gap2B12Comet,
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterLoyalty: 5},
	})
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: gap2B12CometID, OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield})
	})
	g.RunStateChecksForTest()
	return g, me
}

// gap2B12CometKey finds an RNG key under which seat 0's next n single d6
// rolls from Comet are ones `want` accepts, rolled one at a time as n
// activations roll them.
func gap2B12CometKey(t *testing.T, n int, want func([]int) bool) [32]byte {
	t.Helper()
	for k := 0; k < 256; k++ {
		g := effectstest.NewCatalogGame(t)
		g.SetRNGKeyForTest(effectstest.RandomModesKey(byte(k)))
		var rolls []int
		g.WithWriteLock(func() {
			for i := 0; i < n; i++ {
				r, _ := g.RollDiceForEffect(game.RandomDraw{Player: g.Seats[0].ID, Source: gap2B12CometID}, 6, 1)
				rolls = append(rolls, r...)
			}
		})
		if len(rolls) == n && want(rolls) {
			return effectstest.RandomModesKey(byte(k))
		}
	}
	t.Fatal("no key in 0..255 gives the wanted rolls")
	return [32]byte{}
}

func gap2B12CometActivate(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	effectstest.S82Activate(t, g, me.ID, gap2B12CometID, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
}

func gap2B12Loyalty(g *game.Game) int {
	c, _ := g.LookupCardForEffect(gap2B12CometID)
	return c.Counters[game.CounterLoyalty]
}

func TestGap2B12CometSixGrantsTwoMoreActivationsThenSquirrels(t *testing.T) {
	g, me := gap2B12CometTable(t, 2, func(r []int) bool { return r[0] == 6 && r[1] <= 2 })
	gap2B12CometActivate(t, g, me)
	if l := gap2B12Loyalty(g); l != 6 {
		t.Fatalf("loyalty %d after a 6, want 6", l)
	}
	// The second activation this turn is allowed by the 6.
	gap2B12CometActivate(t, g, me)
	if l := gap2B12Loyalty(g); l != 8 {
		t.Errorf("loyalty %d after a 1 or 2, want 8", l)
	}
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Squirrel" && c.Controller == me.ID {
			n++
			if !game.HasKeyword(&c, "haste") {
				t.Error("a Squirrel lacks haste")
			}
		}
	}
	if n != 2 {
		t.Errorf("%d Squirrels, want 2", n)
	}
	// One more is still allowed (two were granted); then none.
	var spent bool
	g.WithWriteLock(func() { spent = g.LoyaltySpentLocked(gap2B12CometID) })
	if spent {
		t.Error("the second granted activation is not open")
	}
}

func TestGap2B12CometWithoutASixActivatesOnce(t *testing.T) {
	g, me := gap2B12CometTable(t, 1, func(r []int) bool { return r[0] == 3 })
	gap2B12CometActivate(t, g, me)
	if err := g.ActivateCatalogAbility(me.ID, gap2B12CometID, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("a second activation was allowed without a 6")
	}
}

func TestGap2B12CometThreeReturnsACheapCard(t *testing.T) {
	g, me := gap2B12CometTable(t, 1, func(r []int) bool { return r[0] == 3 })
	cheap := uuid.New()
	pricey := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: cheap, Name: "Cheap", TypeLine: "Instant", ManaCost: "{1}{R}", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: pricey, Name: "Pricey", TypeLine: "Sorcery", ManaCost: "{2}{R}", Owner: me.ID, Controller: me.ID})
	gap2B12CometActivate(t, g, me)
	if l := gap2B12Loyalty(g); l != 4 {
		t.Errorf("loyalty %d after a 3, want 4", l)
	}
	pick := effectstest.LatestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatalf("no graveyard choice: %+v", g.PendingChoices)
	}
	if effectstest.HasID(pick.ChooseCards, pricey) {
		t.Error("a mana value 3 card was offered")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{cheap}); err != nil {
		t.Fatalf("choosing: %v", err)
	}
	if z := g.FindCardZoneForEffect(cheap); z == nil || z.Kind != game.ZoneHand {
		t.Errorf("the chosen card is in %+v, want the hand", z)
	}
}

func TestGap2B12CometFourOrFiveHitsACreatureThenMinusTwo(t *testing.T) {
	g, me := gap2B12CometTable(t, 1, func(r []int) bool { return r[0] == 4 || r[0] == 5 })
	shrouded := effectstest.PushDiesKeywordCreature(g, g.Seats[1].ID, "Hexproof Giant", 4, 5, "hexproof")
	gap2B12CometActivate(t, g, me)
	effectstest.AnswerConfirmFor(t, g, me, true)
	pick := effectstest.LatestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatalf("no creature choice: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{shrouded}); err != nil {
		t.Fatalf("choosing a hexproof creature (not a target): %v", err)
	}
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(shrouded) {
		t.Error("5 damage did not kill the 4/5")
	}
	if l := gap2B12Loyalty(g); l != 3 {
		t.Errorf("loyalty %d after a 4 or 5, want 3", l)
	}
}

func TestGap2B12CometFourOrFiveHitsAPlayer(t *testing.T) {
	g, me := gap2B12CometTable(t, 1, func(r []int) bool { return r[0] == 4 || r[0] == 5 })
	opp := g.Seats[2]
	effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	life := opp.Life
	gap2B12CometActivate(t, g, me)
	effectstest.AnswerConfirmFor(t, g, me, false)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	if opp.Life != life-5 {
		t.Errorf("life %d, want %d", opp.Life, life-5)
	}
	if l := gap2B12Loyalty(g); l != 3 {
		t.Errorf("loyalty %d, want 3", l)
	}
}
