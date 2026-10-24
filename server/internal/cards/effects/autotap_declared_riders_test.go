package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// autotap_declared_riders_test.go — #4189, the follow-up to #4185.
//
// A mana ability's hand-written Rider is a closure the auto-tapper cannot
// read, so a source that carries one is never planned and the move list
// does not count its mana. The riders that say what they do in data are:
// "This permanent deals N damage to you" (PainToYou), "You gain N life"
// (GainLife, #4185) and, here, the two that act on the opponents:
//
//   - OpponentsGainLife: Grove of the Burnwillows' "Each opponent gains
//     1 life";
//   - PingOpponents: Zhur-Taa Druid's "deals 1 damage to each opponent".
//
// A source that gives the table something is planned, but only after every
// source that does not (tapSource.Table). The sources whose rider spends a
// resource stay manual; the table at the end of the file pins that.

// opponentLife is the life of every seat but p, in seat order.
func opponentLife(g *game.Game, p *game.Player) []int {
	var out []int
	for _, o := range g.Seats {
		if o.ID != p.ID {
			out = append(out, o.Life)
		}
	}
	return out
}

func sameLife(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func autoCast(g *game.Game, p *game.Player, spell uuid.UUID) error {
	return g.CastSpell(p.ID, spell, game.CastSpellParams{Strict: true, AutoTap: true})
}

// --- the shapes -----------------------------------------------------

// The declared numbers reach the built ability, and the built ability still
// carries a Rider, so the activation path is unchanged.
func TestOpponentRidersAreDeclaredAsData(t *testing.T) {
	grove := game.ManaAbilitiesForCard(game.Card{OracleID: b32GroveOfTheBurnwillowsOracle})
	if len(grove) != 2 || grove[0].Rider != nil || grove[0].RiderOpponentLifeGain != 0 {
		t.Fatalf("Grove's {C} half should have no rider: %+v", grove)
	}
	if grove[1].Rider == nil || grove[1].RiderOpponentLifeGain != 1 || grove[1].RiderOpponentDamage != 0 || grove[1].RiderSelfDamage != 0 {
		t.Errorf("Grove's coloured half = %+v, want an opponent life gain of 1 and nothing else declared", grove[1])
	}
	druid := game.ManaAbilitiesForCard(game.Card{OracleID: b40ZhurTaaDruidOracle})
	if len(druid) != 1 || druid[0].Rider == nil || druid[0].RiderOpponentDamage != 1 || druid[0].RiderOpponentLifeGain != 0 {
		t.Errorf("Zhur-Taa Druid = %+v, want an opponent damage of 1 and nothing else declared", druid)
	}
}

// Each declaration builds the rider, so a spec that also writes one would
// say one thing to the engine and another to the planner: Register refuses
// any two.
func TestRegisterRefusesTwoRidersOnOneManaAbility(t *testing.T) {
	rider := func(*game.Game, uuid.UUID, uuid.UUID) error { return nil }
	base := ManaAbility{Cost: ManaAbilityCost{Tap: true}, Produced: "{G}"}
	with := func(edit func(*ManaAbility)) ManaAbility {
		ma := base
		edit(&ma)
		return ma
	}
	for name, ma := range map[string]ManaAbility{
		"Rider and OpponentsGainLife":         with(func(m *ManaAbility) { m.Rider, m.OpponentsGainLife = rider, 1 }),
		"Rider and PingOpponents":             with(func(m *ManaAbility) { m.Rider, m.PingOpponents = rider, 1 }),
		"PainToYou and PingOpponents":         with(func(m *ManaAbility) { m.PainToYou, m.PingOpponents = 1, 1 }),
		"GainLife and OpponentsGainLife":      with(func(m *ManaAbility) { m.GainLife, m.OpponentsGainLife = 1, 1 }),
		"OpponentsGainLife and PingOpponents": with(func(m *ManaAbility) { m.OpponentsGainLife, m.PingOpponents = 1, 1 }),
		"a negative OpponentsGainLife":        with(func(m *ManaAbility) { m.OpponentsGainLife = -1 }),
		"a negative PingOpponents":            with(func(m *ManaAbility) { m.PingOpponents = -1 }),
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Errorf("Register accepted %s", name)
				}
			}()
			Register(Spec{OracleID: "riders-" + name, Name: "Riders " + name, ManaAbilities: []ManaAbility{ma}})
		})
	}
}

// --- Grove of the Burnwillows ---------------------------------------

// Grove alone pays {R}: the enumerator offers the cast, the automatic
// payment taps the coloured half, and each opponent gains the 1 life the
// card prints. The caster's own life does not move.
func TestGrovePaysForASpellAndGivesEachOpponentLife(t *testing.T) {
	g, me, _ := spendTable(t)
	grove := pushCatalogPermanent(g, me.ID, "Grove of the Burnwillows", "Land", b32GroveOfTheBurnwillowsOracle, false)
	spell := handSpell(me, "Shock-ish", "Instant", "{R}")
	before, mine := opponentLife(g, me), me.Life

	if !castMove(g, me.ID, spell) {
		t.Fatal("the enumerator does not offer a {R} spell Grove of the Burnwillows can pay for")
	}
	if err := autoCast(g, me, spell); err != nil {
		t.Fatalf("auto-tapped cast off Grove: %v", err)
	}
	if !isTapped(g, grove) {
		t.Error("Grove did not tap")
	}
	after := opponentLife(g, me)
	for i := range after {
		if after[i] != before[i]+1 {
			t.Errorf("opponent %d life %d → %d, want +1", i+1, before[i], after[i])
		}
	}
	if me.Life != mine {
		t.Errorf("the caster's life moved: %d → %d", mine, me.Life)
	}
}

// A generic cost is paid by Grove's painless {C} half: nobody gains life.
func TestGrovePaysGenericWithItsColorlessHalf(t *testing.T) {
	g, me, _ := spendTable(t)
	pushCatalogPermanent(g, me.ID, "Grove of the Burnwillows", "Land", b32GroveOfTheBurnwillowsOracle, false)
	spell := handSpell(me, "Mind Stone-ish", "Artifact", "{1}")
	before := opponentLife(g, me)

	if err := autoCast(g, me, spell); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if after := opponentLife(g, me); !sameLife(before, after) {
		t.Errorf("opponents' life %v → %v; the {C} half pays {1} for nothing", before, after)
	}
}

// A Mountain beside Grove pays the {R}: the coloured half feeds the table,
// so it is the last source the planner reaches for.
func TestGroveIsSpentAfterAPlainMountain(t *testing.T) {
	g, me, _ := spendTable(t)
	grove := pushCatalogPermanent(g, me.ID, "Grove of the Burnwillows", "Land", b32GroveOfTheBurnwillowsOracle, false)
	mountain := basics(g, me, "Mountain", 1)[0]
	spell := handSpell(me, "Shock-ish", "Instant", "{R}")
	before := opponentLife(g, me)

	if err := autoCast(g, me, spell); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if !isTapped(g, mountain) || isTapped(g, grove) {
		t.Errorf("Mountain tapped %v, Grove tapped %v; the Mountain should pay and Grove stay untapped", isTapped(g, mountain), isTapped(g, grove))
	}
	if after := opponentLife(g, me); !sameLife(before, after) {
		t.Errorf("opponents' life %v → %v with a Mountain to pay", before, after)
	}
}

// The tier is not just "fewest colours first": Grove's coloured half
// makes two colours and Birds of Paradise five, and an untiered planner
// would spend the narrower source. Grove feeds the table, so Birds pays.
func TestGroveIsSpentAfterAMoreFlexibleSourceThatGivesNothing(t *testing.T) {
	g, me, _ := spendTable(t)
	grove := pushCatalogPermanent(g, me.ID, "Grove of the Burnwillows", "Land", b32GroveOfTheBurnwillowsOracle, false)
	birds := pushCatalogPermanent(g, me.ID, "Birds of Paradise", "Creature — Bird", "d3a0b660-358c-41bd-9cd2-41fbf3491b1a", false)
	spell := handSpell(me, "Shock-ish", "Instant", "{R}")
	before := opponentLife(g, me)

	if err := autoCast(g, me, spell); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if !isTapped(g, birds) || isTapped(g, grove) {
		t.Errorf("Birds tapped %v, Grove tapped %v; Birds should pay", isTapped(g, birds), isTapped(g, grove))
	}
	if after := opponentLife(g, me); !sameLife(before, after) {
		t.Errorf("opponents' life %v → %v with Birds to pay", before, after)
	}
}

// Both halves are one permanent: {R}{1} off Grove alone is two mana from
// one tap, which it cannot make. Grove is not booked twice.
func TestGroveIsNotBookedTwice(t *testing.T) {
	g, me, _ := spendTable(t)
	pushCatalogPermanent(g, me.ID, "Grove of the Burnwillows", "Land", b32GroveOfTheBurnwillowsOracle, false)
	spell := handSpell(me, "Two", "Instant", "{R}{1}")
	if castMove(g, me.ID, spell) {
		t.Error("{R}{1} offered off one Grove of the Burnwillows")
	}
	refusedForMana(t, autoCast(g, me, spell), "{R}{1} off one Grove")
}

// --- Zhur-Taa Druid -------------------------------------------------

func TestZhurTaaDruidPaysForASpellAndPingsEachOpponent(t *testing.T) {
	g, me, _ := spendTable(t)
	druid := pushCatalogPermanent(g, me.ID, "Zhur-Taa Druid", "Creature — Human Druid", b40ZhurTaaDruidOracle, false)
	spell := handSpell(me, "Growth-ish", "Instant", "{G}")
	before, mine := opponentLife(g, me), me.Life

	if !castMove(g, me.ID, spell) {
		t.Fatal("the enumerator does not offer a {G} spell Zhur-Taa Druid can pay for")
	}
	if err := autoCast(g, me, spell); err != nil {
		t.Fatalf("auto-tapped cast off the Druid: %v", err)
	}
	if !isTapped(g, druid) {
		t.Error("the Druid did not tap")
	}
	after := opponentLife(g, me)
	for i := range after {
		if after[i] != before[i]-1 {
			t.Errorf("opponent %d life %d → %d, want -1", i+1, before[i], after[i])
		}
	}
	if me.Life != mine {
		t.Errorf("the caster's life moved: %d → %d", mine, me.Life)
	}
}

// A Forest pays the {G} and the Druid stays out of it.
func TestZhurTaaDruidIsSpentAfterAForest(t *testing.T) {
	g, me, _ := spendTable(t)
	druid := pushCatalogPermanent(g, me.ID, "Zhur-Taa Druid", "Creature — Human Druid", b40ZhurTaaDruidOracle, false)
	forest := basics(g, me, "Forest", 1)[0]
	spell := handSpell(me, "Growth-ish", "Instant", "{G}")
	before := opponentLife(g, me)

	if err := autoCast(g, me, spell); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if !isTapped(g, forest) || isTapped(g, druid) {
		t.Errorf("Forest tapped %v, Druid tapped %v; the Forest should pay", isTapped(g, forest), isTapped(g, druid))
	}
	if after := opponentLife(g, me); !sameLife(before, after) {
		t.Errorf("opponents' life %v → %v with a Forest to pay", before, after)
	}
}

// CR 302.6: a Druid that came in this turn can't be tapped for mana, and
// the planner does not count it.
func TestZhurTaaDruidWithSummoningSicknessIsNotCounted(t *testing.T) {
	g, me, _ := spendTable(t)
	pushCatalogPermanent(g, me.ID, "Zhur-Taa Druid", "Creature — Human Druid", b40ZhurTaaDruidOracle, true)
	spell := handSpell(me, "Growth-ish", "Instant", "{G}")
	if castMove(g, me.ID, spell) {
		t.Error("a summoning-sick Druid's mana is counted")
	}
}

// The table tier sits before a source that costs life: with Grove and a
// painland and nothing else, Grove pays and nobody pays life.
func TestTableSourceIsSpentBeforeALifeCost(t *testing.T) {
	g, me, _ := spendTable(t)
	grove := pushCatalogPermanent(g, me.ID, "Grove of the Burnwillows", "Land", b32GroveOfTheBurnwillowsOracle, false)
	pushCatalogPermanent(g, me.ID, "Shivan Reef", "Land", shivanReefOracle, false)
	spell := handSpell(me, "Shock-ish", "Instant", "{R}")
	mine := me.Life

	if err := autoCast(g, me, spell); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if !isTapped(g, grove) || me.Life != mine {
		t.Errorf("Grove tapped %v, caster life %d → %d; Grove should pay and the painland stay out", isTapped(g, grove), mine, me.Life)
	}
}

// --- the Great Henge under a life-gain-to-loss replacement ----------

// The Henge gains its controller 2 life as it taps. Under Plague Drone's
// Rot Fly — Tainted Remedy's text — an opponent's gain is a loss, so the
// tap costs 2 life. The planner holds that to the bar every life-spending
// source meets: it never takes its player to 0 or below.
func TestGreatHengeUnderRotFlyNeverTakesTheLastLife(t *testing.T) {
	g, me, opp := spendTable(t)
	pushCatalogPermanent(g, opp.ID, "Plague Drone", "Creature — Demon", shPlagueDroneOracle, false)
	henge := pushCatalogPermanent(g, me.ID, "The Great Henge", "Legendary Artifact", theGreatHengeOracle, false)
	spell := handSpell(me, "Two Green", "Instant", "{G}{G}")

	g.WithWriteLock(func() { me.Life = 2 })
	if castMove(g, me.ID, spell) {
		t.Error("the enumerator offers a cast whose only payment would take its caster from 2 to 0")
	}
	refusedForMana(t, autoCast(g, me, spell), "cast at 2 life under Rot Fly")
	if me.Life != 2 || isTapped(g, henge) {
		t.Errorf("refused cast spent something: life %d, tapped %v", me.Life, isTapped(g, henge))
	}
	// The Henge is still there for a deliberate click, as a painland is.
	if err := g.ActivateManaAbility(me.ID, henge, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("hand-tapping the Henge: %v", err)
	}
}

// One more point of life and the same cast is fine: it pays, and the
// player ends on 1, the way a painland does.
func TestGreatHengeUnderRotFlyPaysWhileItLeavesALife(t *testing.T) {
	g, me, opp := spendTable(t)
	pushCatalogPermanent(g, opp.ID, "Plague Drone", "Creature — Demon", shPlagueDroneOracle, false)
	henge := pushCatalogPermanent(g, me.ID, "The Great Henge", "Legendary Artifact", theGreatHengeOracle, false)
	spell := handSpell(me, "Two Green", "Instant", "{G}{G}")

	g.WithWriteLock(func() { me.Life = 3 })
	if !castMove(g, me.ID, spell) {
		t.Fatal("the enumerator hides a cast that leaves its caster on 1 life")
	}
	if err := autoCast(g, me, spell); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if !isTapped(g, henge) || me.Life != 1 {
		t.Errorf("Henge tapped %v, life %d, want tapped and 1", isTapped(g, henge), me.Life)
	}
}

// Rot Fly is an opponent's: a Drone the caster controls does not touch the
// caster's own gain, so the Henge is an ordinary source at 1 life.
func TestGreatHengeUnderMyOwnRotFlyStillGainsLife(t *testing.T) {
	g, me, _ := spendTable(t)
	pushCatalogPermanent(g, me.ID, "Plague Drone", "Creature — Demon", shPlagueDroneOracle, false)
	henge := pushCatalogPermanent(g, me.ID, "The Great Henge", "Legendary Artifact", theGreatHengeOracle, false)
	spell := handSpell(me, "Two Green", "Instant", "{G}{G}")

	g.WithWriteLock(func() { me.Life = 1 })
	if err := autoCast(g, me, spell); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if !isTapped(g, henge) || me.Life != 3 {
		t.Errorf("Henge tapped %v, life %d, want tapped and 3", isTapped(g, henge), me.Life)
	}
}

// The guard is only about a gain that becomes a loss. A doubler turns the
// gain into a bigger gain, and the Henge stays an ordinary source at 1 life.
func TestGreatHengeUnderADoublerStillPlans(t *testing.T) {
	g, me, _ := spendTable(t)
	b12Push(g, me.ID, "Rhox Faithmender", "Creature — Rhino Monk", b14RhoxFaithmenderOracle, 1, 5)
	henge := pushCatalogPermanent(g, me.ID, "The Great Henge", "Legendary Artifact", theGreatHengeOracle, false)
	spell := handSpell(me, "Two Green", "Instant", "{G}{G}")

	g.WithWriteLock(func() { me.Life = 1 })
	if err := autoCast(g, me, spell); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if !isTapped(g, henge) || me.Life != 5 {
		t.Errorf("Henge tapped %v, life %d, want tapped and 5", isTapped(g, henge), me.Life)
	}
}

// A Forest beside the Henge pays a {G} without the Henge's loss.
func TestGreatHengeUnderRotFlyIsSpentAfterAForest(t *testing.T) {
	g, me, opp := spendTable(t)
	pushCatalogPermanent(g, opp.ID, "Plague Drone", "Creature — Demon", shPlagueDroneOracle, false)
	henge := pushCatalogPermanent(g, me.ID, "The Great Henge", "Legendary Artifact", theGreatHengeOracle, false)
	forest := basics(g, me, "Forest", 1)[0]
	spell := handSpell(me, "Green", "Instant", "{G}")
	mine := me.Life

	if err := autoCast(g, me, spell); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if !isTapped(g, forest) || isTapped(g, henge) || me.Life != mine {
		t.Errorf("Forest tapped %v, Henge tapped %v, life %d → %d; the Forest should pay", isTapped(g, forest), isTapped(g, henge), mine, me.Life)
	}
}

// --- the sources that stay manual -----------------------------------

// A source whose rider spends or changes a resource is not tapped on the
// player's behalf, and its mana is not counted by the move list either:
// every move the enumerator offers carries auto_tap and is accepted as it
// stands (the #544 invariant), so a cast the planner cannot pay is not
// offered. Each is still an ordinary mana source by hand.
func TestSourcesWhoseRiderUsesAResourceStayManual(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, cost string
		counters                     map[string]int
	}{
		{"Gemstone Mine", "Land", "0c828f10-4775-492f-9224-1e2814ad2cad", "{W}", map[string]int{"mining": 3}},
		{"Twitching Doll", "Artifact Creature — Spider Toy", "fd6e1967-237a-41f6-bbf4-2c869f9447c8", "{W}", nil},
		{"Temple of Cyclical Time", "Land", ojerPakpatiqOracleID + "#1", "{U}", map[string]int{"time": 3}},
		{"Runaway Steam-Kin", "Creature — Elemental", "eb414a8f-566f-4635-80a1-77b2e18ac8dc", "{R}", map[string]int{"+1/+1": 3}},
		{"Vexing Puzzlebox", "Artifact", "7267ab16-2157-4b86-93ea-ca2c0fee064a", "{W}", nil},
		{"Solar Array", "Artifact", "e8e2f273-5e74-4f16-8d49-2e86e9c9f2dc", "{W}", nil},
		{"Yuna, Grand Summoner", "Legendary Creature — Human Cleric", "ba72d4a9-6f36-4796-bd71-2342fa9b4a30", "{W}", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := spendTable(t)
			src := pushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
			for kind, n := range tc.counters {
				g.WithWriteLock(func() { _ = g.AddCounterForEffect(src, kind, n) })
			}
			for _, ab := range game.ManaAbilitiesForCard(*battlefieldCardFor(g, src)) {
				if ab.Rider == nil && ab.PreRider == nil {
					t.Fatalf("%s: the ability has no rider; the table is stale", tc.name)
				}
				if ab.RiderSelfDamage != 0 || ab.RiderSelfLifeGain != 0 || ab.RiderOpponentLifeGain != 0 || ab.RiderOpponentDamage != 0 {
					t.Fatalf("%s declares a rider the planner can read: %+v", tc.name, ab)
				}
			}
			spell := handSpell(me, "One", "Instant", tc.cost)
			if castMove(g, me.ID, spell) {
				t.Errorf("%s: the enumerator offers a cast only %s could pay for", tc.name, tc.name)
			}
			refusedForMana(t, autoCast(g, me, spell), tc.name+" auto-tapped")
			if isTapped(g, src) {
				t.Errorf("%s tapped by a refused cast", tc.name)
			}
			// By hand it is still a mana source.
			if err := g.ActivateManaAbility(me.ID, src, 0, game.ManaAbilityParams{}); err != nil {
				t.Fatalf("%s by hand: %v", tc.name, err)
			}
			activated := false
			for _, ev := range g.Events {
				if ev.Kind == game.EventManaAbilityActivated && ev.Source == src {
					activated = true
				}
			}
			if !activated {
				t.Errorf("%s by hand did not activate its mana ability", tc.name)
			}
		})
	}
}
