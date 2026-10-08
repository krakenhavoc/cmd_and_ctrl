package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// awaken_cards_test.go — ADR 0135 §3 (#2411): awaken (CR 702.113) against
// the real catalog. The awaken cost is an alternative cost whose target
// statement adds "target land you control" after the spell's own (CR
// 702.113b), and its land gets N +1/+1 counters FIRST and then becomes a
// 0/0 Elemental creature with haste that is still a land (owner decision
// 4, CR 702.113a, CR 608.2c).

const (
	boilingEarthOracle      = "fb34b671-b61a-47e2-90fa-dbe5cf6d3743"
	clutchOfCurrentsOracle  = "2c6658c3-8c0e-46c9-8127-bc5b77eb0dab"
	coastalDiscoveryOracle  = "f7c84690-8c7c-41a6-b430-a9771030f903"
	earthenArmsOracle       = "2146fa14-dc6a-4e13-bed0-dda235784ee6"
	encirclingFissureOracle = "39d7569e-62b0-4f37-813a-0985632c66c9"
	miresMaliceOracle       = "1e7b18a3-43eb-4491-bb19-99f261a94719"
	onduRisingOracle        = "01081f81-f588-43ab-8d47-13593aa19ce1"
	partTheWaterveilOracle  = "54dd35ee-6f89-460d-9370-72bc3fa6840a"
	planarOutburstOracle    = "ff0d05b9-0c9c-4217-abd2-7e7757a0c4c9"
	risingMiasmaOracle      = "83fb7dfb-64a4-4320-b544-830c29f24df9"
	roilSpoutOracle         = "ebc1d343-26f1-4d85-ac3a-610da24c990e"
	ruinousPathOracle       = "7b4aa101-3c8a-44b3-9d87-91a4fe4fbae4"
	rushOfIceOracle         = "e9d62416-c4b1-4982-9593-174d46e975c7"
	scatterToTheWindsOracle = "13d600e6-86b4-4813-8f34-75f23717f433"
	sheerDropOracle         = "7c42123d-c8e2-4571-b07d-f782e9ad1f8b"
)

func TestAwakenCardsAreFull(t *testing.T) {
	for _, c := range []struct {
		oracle, cost string
		n, printed   int
	}{
		{boilingEarthOracle, "{6}{R}", 4, 0},
		{clutchOfCurrentsOracle, "{4}{U}", 3, 1},
		{coastalDiscoveryOracle, "{5}{U}", 4, 0},
		{earthenArmsOracle, "{6}{G}", 4, 1},
		{encirclingFissureOracle, "{4}{W}", 2, 1},
		{miresMaliceOracle, "{5}{B}", 3, 1},
		{onduRisingOracle, "{4}{W}", 4, 0},
		{partTheWaterveilOracle, "{6}{U}{U}{U}", 6, 0},
		{planarOutburstOracle, "{5}{W}{W}{W}", 4, 0},
		{risingMiasmaOracle, "{5}{B}{B}", 3, 0},
		{roilSpoutOracle, "{4}{W}{U}", 4, 1},
		{ruinousPathOracle, "{5}{B}{B}", 4, 1},
		{rushOfIceOracle, "{4}{U}", 3, 1},
		{scatterToTheWindsOracle, "{4}{U}{U}", 3, 1},
		{sheerDropOracle, "{5}{W}", 3, 1},
	} {
		spec, ok := Lookup(c.oracle)
		if !ok {
			t.Errorf("%s is not registered", c.oracle)
			continue
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v, caveats %v; want Full with none", spec.Name, spec.Completeness, spec.Caveats)
		}
		if got := spec.Targets.ClauseCount(); got != c.printed {
			t.Errorf("%s: %d printed clauses, want %d", spec.Name, got, c.printed)
		}
		offers := game.AlternativeCostsFor(c.oracle)
		if len(offers) != 1 {
			t.Errorf("%s offers %d alternative costs, want one", spec.Name, len(offers))
			continue
		}
		o := offers[0]
		if o.Key != "awaken" || o.ManaCost != c.cost || o.Purpose.AwakenLand != c.n ||
			o.Targets.ClauseCount() != c.printed+1 || !strings.HasPrefix(o.Label, "Awaken ") {
			t.Errorf("%s offers %+v, want awaken %d for %s with %d clauses", spec.Name, o, c.n, c.cost, c.printed+1)
		}
	}
}

// awakenTable is a main phase on seat 0's turn with an empty hand and one
// land of seat 0's.
func awakenTable(t *testing.T) (g *game.Game, me, opp *game.Player, land uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	advanceToMain(t, g)
	me, opp = g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	land = pushEarthbendLand(g, me.ID, "Swamp", "Basic Land — Swamp")
	return g, me, opp, land
}

func awakenRef(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

func castAwaken(g *game.Game, p *game.Player, card uuid.UUID, targets ...game.TargetRef) error {
	return g.CastSpell(p.ID, card, game.CastSpellParams{AlternativeCost: "awaken", Targets: targets})
}

// awakenedLand reads the land after the layers: its counters, its P/T, and
// whether it is an Elemental Land Creature with haste.
func awakenedLand(t *testing.T, g *game.Game, land uuid.UUID) (counters, power, toughness int, elementalLandCreature, hasty bool) {
	t.Helper()
	power, toughness, landCreature, hasty := earthbentBody(t, g, land)
	elemental := false
	g.ReadSnapshot(func() {
		if c := findBattlefieldCardForTest(g, land); c != nil {
			elemental = c.HasSubtype("Elemental")
		}
	})
	return countersOn(g, land, game.CounterPlusOne), power, toughness, landCreature && elemental, hasty
}

func isAnimated(t *testing.T, g *game.Game, land uuid.UUID) bool {
	t.Helper()
	_, _, creature, _ := earthbentBody(t, g, land)
	return creature
}

// Ruinous Path for its awaken cost destroys the creature and makes the land
// a 4/4 Elemental land creature with haste. For {1}{B}{B} it has no land
// target (CR 702.113b): a second target is refused and nothing is
// animated; and an awaken cast must name the land.
func TestRuinousPathAwakensTheLand(t *testing.T) {
	g, me, opp, land := awakenTable(t)
	victim := pr7Creature(g, opp.ID, "Victim", 2, "G")
	path := handCardOf(g, me, "Ruinous Path", "Sorcery", "{1}{B}{B}", ruinousPathOracle)

	if err := castAwaken(g, me, path, awakenRef(victim)); err == nil {
		t.Fatal("an awaken cast with no land target was accepted")
	}
	if err := g.CastSpell(me.ID, path, game.CastSpellParams{Targets: []game.TargetRef{awakenRef(victim), awakenRef(land)}}); err == nil {
		t.Fatal("a cast for the mana cost took the awaken land as a target (CR 702.113b)")
	}
	theirLand := pushEarthbendLand(g, opp.ID, "Swamp", "Basic Land — Swamp")
	if err := castAwaken(g, me, path, awakenRef(victim), awakenRef(theirLand)); err == nil {
		t.Fatal("awaken took an opponent's land")
	}
	if err := castAwaken(g, me, path, awakenRef(victim), awakenRef(land)); err != nil {
		t.Fatalf("Ruinous Path for its awaken cost: %v", err)
	}
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, victim) != nil {
		t.Error("the creature was not destroyed")
	}
	counters, p, tough, elemental, hasty := awakenedLand(t, g, land)
	if counters != 4 || p != 4 || tough != 4 || !elemental || !hasty {
		t.Errorf("land: %d counters, %d/%d, elemental land creature %v, haste %v; want 4, 4/4, true, true",
			counters, p, tough, elemental, hasty)
	}
}

func TestRuinousPathForItsManaCostDoesNotAwaken(t *testing.T) {
	g, me, opp, land := awakenTable(t)
	victim := pr7Creature(g, opp.ID, "Victim", 2, "G")
	path := handCardOf(g, me, "Ruinous Path", "Sorcery", "{1}{B}{B}", ruinousPathOracle)
	if err := g.CastSpell(me.ID, path, game.CastSpellParams{Targets: []game.TargetRef{awakenRef(victim)}}); err != nil {
		t.Fatalf("Ruinous Path for {1}{B}{B}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, victim) != nil {
		t.Error("the creature was not destroyed")
	}
	if isAnimated(t, g, land) || countersOn(g, land, game.CounterPlusOne) != 0 {
		t.Error("a cast for the mana cost awakened a land")
	}
}

// The awaken cost is paid under the strict gate with real mana, and the
// land named to awaken may tap for that mana (a tapped land is still a
// land you control).
func TestClutchOfCurrentsPaysItsAwakenCostWithMana(t *testing.T) {
	g, me, opp, swamp := awakenTable(t)
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, swamp).Tapped = true })
	var islands []uuid.UUID
	for i := 0; i < 5; i++ {
		islands = append(islands, pushEarthbendLand(g, me.ID, "Island", "Basic Land — Island"))
	}
	creature := pr7Creature(g, opp.ID, "Bounced", 2, "G")
	clutch := handCardOf(g, me, "Clutch of Currents", "Sorcery", "{U}", clutchOfCurrentsOracle)
	if err := g.CastSpell(me.ID, clutch, game.CastSpellParams{
		AlternativeCost: "awaken", Strict: true, AutoTap: true,
		Targets: []game.TargetRef{awakenRef(creature), awakenRef(islands[0])},
	}); err != nil {
		t.Fatalf("Clutch of Currents for {4}{U}: %v", err)
	}
	tapped := 0
	for _, id := range islands {
		if tappedForTest(t, g, id) {
			tapped++
		}
	}
	if tapped != 5 {
		t.Errorf("%d Islands tapped, want 5 for {4}{U}", tapped)
	}
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, creature) != nil {
		t.Error("the creature was not returned")
	}
	if counters, p, tough, elemental, _ := awakenedLand(t, g, islands[0]); counters != 3 || p != 3 || tough != 3 || !elemental {
		t.Errorf("land: %d counters, %d/%d, elemental %v; want a 3/3 Elemental", counters, p, tough, elemental)
	}
}

// CR 608.2b: a printed target that left still lets the land awaken, and
// it is never handed to the land.
func TestClutchOfCurrentsAwakensWhenItsCreatureIsGone(t *testing.T) {
	g, me, opp, land := awakenTable(t)
	creature := pr7Creature(g, opp.ID, "Gone", 2, "G")
	clutch := handCardOf(g, me, "Clutch of Currents", "Sorcery", "{U}", clutchOfCurrentsOracle)
	if err := castAwaken(g, me, clutch, awakenRef(creature), awakenRef(land)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.TuckToLibraryForEffect(creature, false); err != nil {
			t.Fatalf("tuck: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, land) == nil {
		t.Fatal("the land was bounced in the creature's place")
	}
	if counters, _, _, elemental, _ := awakenedLand(t, g, land); counters != 3 || !elemental {
		t.Errorf("land: %d counters, elemental %v; want 3, true", counters, elemental)
	}
}

// Coastal Discovery's only target is the awaken land: if it is gone, the
// spell does not resolve and draws nothing (CR 608.2b; the ruling).
func TestCoastalDiscoveryDoesNotResolveWithoutItsLand(t *testing.T) {
	g, me, _, land := awakenTable(t)
	disc := handCardOf(g, me, "Coastal Discovery", "Sorcery", "{3}{U}", coastalDiscoveryOracle)
	if err := castAwaken(g, me, disc, awakenRef(land)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.TuckToLibraryForEffect(land, false); err != nil {
			t.Fatalf("tuck: %v", err)
		}
	})
	before := len(me.Hand.Cards)
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != before {
		t.Errorf("hand %d → %d; an awaken Coastal Discovery whose land left drew cards", before, got)
	}
}

func TestCoastalDiscoveryDrawsAndAwakens(t *testing.T) {
	g, me, _, land := awakenTable(t)
	disc := handCardOf(g, me, "Coastal Discovery", "Sorcery", "{3}{U}", coastalDiscoveryOracle)
	if err := castAwaken(g, me, disc, awakenRef(land)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	before := len(me.Hand.Cards)
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != before+2 {
		t.Errorf("hand %d → %d, want two cards drawn", before, got)
	}
	if counters, _, _, elemental, hasty := awakenedLand(t, g, land); counters != 4 || !elemental || !hasty {
		t.Errorf("land: %d counters, elemental %v, haste %v", counters, elemental, hasty)
	}
}

// Counters first (CR 608.2c, owner decision 4): Hardened Scales asks "a
// creature you control", and the land is not one yet, so it adds nothing.
// Doubling Season asks "a permanent you control", so it doubles them.
func TestAwakenCountersGoOnBeforeTheLandIsACreature(t *testing.T) {
	t.Run("Hardened Scales", func(t *testing.T) {
		g, me, _, land := awakenTable(t)
		seedReplacementPermanent(g, hardenedScalesOracle, "Hardened Scales", me.ID)
		disc := handCardOf(g, me, "Coastal Discovery", "Sorcery", "{3}{U}", coastalDiscoveryOracle)
		if err := castAwaken(g, me, disc, awakenRef(land)); err != nil {
			t.Fatalf("cast: %v", err)
		}
		passPriorityAroundTable(t, g)
		if got := countersOn(g, land, game.CounterPlusOne); got != 4 {
			t.Errorf("+1/+1 counters %d, want 4: Hardened Scales must not see a land that is not yet a creature", got)
		}
		if !isAnimated(t, g, land) {
			t.Error("the land was not animated")
		}
	})
	t.Run("Doubling Season", func(t *testing.T) {
		g, me, _, land := awakenTable(t)
		seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", me.ID)
		disc := handCardOf(g, me, "Coastal Discovery", "Sorcery", "{3}{U}", coastalDiscoveryOracle)
		if err := castAwaken(g, me, disc, awakenRef(land)); err != nil {
			t.Fatalf("cast: %v", err)
		}
		passPriorityAroundTable(t, g)
		if got := countersOn(g, land, game.CounterPlusOne); got != 8 {
			t.Errorf("+1/+1 counters %d, want 8: Doubling Season doubles counters on a permanent you control", got)
		}
	})
}

// The animation has no duration (CR 611.2a): the land is still a hasty
// Elemental creature on a later turn, and awaken queues no return trigger.
func TestAnAwakenedLandStaysAnimated(t *testing.T) {
	g, me, _, land := awakenTable(t)
	disc := handCardOf(g, me, "Coastal Discovery", "Sorcery", "{3}{U}", coastalDiscoveryOracle)
	if err := castAwaken(g, me, disc, awakenRef(land)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := len(g.DelayedTriggers); n != 0 {
		t.Errorf("%d delayed triggers queued; awaken has no return trigger", n)
	}
	endTurn(t, g)
	endTurn(t, g)
	if counters, p, tough, elemental, hasty := awakenedLand(t, g, land); counters != 4 || p != 4 || tough != 4 || !elemental || !hasty {
		t.Errorf("two turns later: %d counters, %d/%d, elemental %v, haste %v", counters, p, tough, elemental, hasty)
	}
}

// Earthen Arms may name the same land twice (CR 601.2c): two counters from
// the spell, four from awaken.
func TestEarthenArmsOnItsOwnAwakenLand(t *testing.T) {
	g, me, _, land := awakenTable(t)
	arms := handCardOf(g, me, "Earthen Arms", "Sorcery", "{1}{G}", earthenArmsOracle)
	if err := castAwaken(g, me, arms, awakenRef(land), awakenRef(land)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if counters, p, tough, elemental, _ := awakenedLand(t, g, land); counters != 6 || p != 6 || tough != 6 || !elemental {
		t.Errorf("land: %d counters, %d/%d, elemental %v; want a 6/6 Elemental", counters, p, tough, elemental)
	}
}

// Planar Outburst destroys nonland creatures: an earlier awakened land
// survives it, and so does the land it awakens.
func TestPlanarOutburstSparesLandCreatures(t *testing.T) {
	g, me, opp, land := awakenTable(t)
	bear := pr7Creature(g, opp.ID, "Bear", 2, "G")
	old := pushEarthbendLand(g, me.ID, "Plains", "Basic Land — Plains")
	g.WithWriteLock(func() {
		if err := g.AwakenForEffect(me.ID, uuid.Nil, old, 2); err != nil {
			t.Fatalf("awaken: %v", err)
		}
	})
	outburst := handCardOf(g, me, "Planar Outburst", "Sorcery", "{3}{W}{W}", planarOutburstOracle)
	if err := castAwaken(g, me, outburst, awakenRef(land)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, bear) != nil {
		t.Error("a nonland creature survived")
	}
	if !isAnimated(t, g, old) {
		t.Error("an awakened land creature was destroyed by \"nonland creatures\"")
	}
	if !isAnimated(t, g, land) {
		t.Error("the land was not awakened")
	}
}

// Rising Miasma's -2/-2 is fixed to the creatures there as it resolves
// (CR 611.2c); the land becomes a creature after it and stays a 3/3.
func TestRisingMiasmaDoesNotShrinkItsOwnAwakenedLand(t *testing.T) {
	g, me, opp, land := awakenTable(t)
	bear := pr7Creature(g, opp.ID, "Bear", 2, "G")
	miasma := handCardOf(g, me, "Rising Miasma", "Sorcery", "{3}{B}", risingMiasmaOracle)
	if err := castAwaken(g, me, miasma, awakenRef(land)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, p, tough, _, _ := awakenedLand(t, g, land); p != 3 || tough != 3 {
		t.Errorf("land %d/%d, want 3/3", p, tough)
	}
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if c := findBattlefieldCardForTest(g, bear); c != nil && c.CurrentToughness() != 2 {
		t.Errorf("the 2/4 test creature is %d toughness, want 2", c.CurrentToughness())
	}
}

// Encircling Fissure's shield is against the target opponent's creatures
// only, and an opponent may cast it for its awaken cost on their own land.
func TestEncirclingFissurePreventsTheTargetOpponentsCombatDamage(t *testing.T) {
	for _, c := range []struct {
		name      string
		targetMe  bool
		wantLoss  int
		awakenOpp bool
	}{
		{"targets the attacker's controller", true, 0, true},
		{"targets another opponent", false, 3, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			g, me, opp, _ := awakenTable(t)
			attacker := sfPush(g, me.ID, "Attacker", "Creature — Test", 3, []string{"R"})
			pr7bAttack(t, g, opp.ID, attacker)
			target := g.Seats[2].ID
			if c.targetMe {
				target = me.ID
			}
			fissure := handCardOf(g, opp, "Encircling Fissure", "Instant", "{2}{W}", encirclingFissureOracle)
			params := game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: target}}}
			var oppLand uuid.UUID
			if c.awakenOpp {
				oppLand = pushEarthbendLand(g, opp.ID, "Plains", "Basic Land — Plains")
				params.AlternativeCost = "awaken"
				params.Targets = append(params.Targets, awakenRef(oppLand))
			}
			if err := g.CastSpell(opp.ID, fissure, params); err != nil {
				t.Fatalf("cast: %v", err)
			}
			passPriorityAroundTable(t, g)
			if lost := sfCombat(t, g)[opp.ID]; lost != c.wantLoss {
				t.Errorf("defender lost %d, want %d", lost, c.wantLoss)
			}
			if c.awakenOpp {
				if counters, _, _, elemental, _ := awakenedLand(t, g, oppLand); counters != 2 || !elemental {
					t.Errorf("the caster's land: %d counters, elemental %v; want 2, true", counters, elemental)
				}
			}
		})
	}
}

// Ondu Rising: each creature that attacks this turn gains lifelink, the
// awakened land among them (it has haste).
func TestOnduRisingGivesAttackersLifelink(t *testing.T) {
	g, me, opp, land := awakenTable(t)
	bear := pr7Creature(g, me.ID, "Bear", 2, "W")
	ondu := handCardOf(g, me, "Ondu Rising", "Sorcery", "{1}{W}", onduRisingOracle)
	if err := castAwaken(g, me, ondu, awakenRef(land)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	declareAttack(t, g, opp.ID, bear, land)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	for name, id := range map[string]uuid.UUID{"the creature": bear, "the awakened land": land} {
		c := findBattlefieldCardForTest(g, id)
		if c == nil || !game.HasKeyword(c, "lifelink") {
			t.Errorf("%s attacked and has no lifelink", name)
		}
	}
}
