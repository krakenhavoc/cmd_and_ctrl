package testsex

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestFinSZ2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		effectstest.FinSZ2SahaginOracle:   "Sahagin",
		effectstest.FinSZ2ShantottoOracle: "Shantotto, Tactician Magician",
		effectstest.FinSZ2UltrosOracle:    "Ultros, Obnoxious Octopus",
		effectstest.FinSZ2PrimaVistaOracl: "The Prima Vista",
		effectstest.FinSZ2SephirothOracle: "Sephiroth, Planet's Heir",
		effectstest.FinSZ2SanctumOracle:   "Stuck in Summoner's Sanctum",
		effectstest.FinSZ2SleepOracle:     "Sleep Magic",
		effectstest.FinSZ2SwallowedOracle: "Swallowed by Leviathan",
		effectstest.FinSZ2GreeterOracle:   "Town Greeter",
		effectstest.FinSZ2ZellOracle:      "Zell Dincht",
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

// finSZ2CastPaid casts a plain spell from the active seat's hand, paid
// strictly from `pool`, so the payment record holds that much mana.
func finSZ2CastPaid(t *testing.T, g *game.Game, typeLine, cost, pool string) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.SocB17Hand(me, "Test Spell", typeLine, cost, effectstest.SocB17PlainSorceryOracle)
	effectstest.FloatForTest(g, me, pool)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast %s: %v", cost, err)
	}
	return id
}

// --- Sahagin ---------------------------------------------------------------------

func TestFinSZ2SahaginGrowsAndSlipsPastOnlyOnFourMana(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, cost, pool string
		want                       bool
	}{
		{"three mana", "Sorcery", "{2}{U}", "UUU", false},
		{"four mana", "Sorcery", "{3}{U}", "UUUU", true},
		{"four mana creature", "Creature — Bear", "{3}{G}", "GGGG", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			sahagin := effectstest.PushCatalogPermanent(g, me.ID, "Sahagin", "Creature — Merfolk Warrior", effectstest.FinSZ2SahaginOracle, false)
			finSZ2CastPaid(t, g, tc.typeLine, tc.cost, tc.pool)
			effectstest.PassPriorityAroundTable(t, g)
			counters := effectstest.CountersOn(g, sahagin, game.CounterPlusOne)
			unblockable := effectstest.RestrictionsOf(t, g, sahagin)&game.CantBeBlocked != 0
			if tc.want && (counters != 1 || !unblockable) {
				t.Errorf("counters %d, can't be blocked %v; want one counter and unblockable", counters, unblockable)
			}
			if !tc.want && (counters != 0 || unblockable) {
				t.Errorf("counters %d, can't be blocked %v; want neither", counters, unblockable)
			}
		})
	}
}

// --- Shantotto, Tactician Magician -----------------------------------------------

func TestFinSZ2ShantottoPumpsByManaSpentAndDrawsAtFour(t *testing.T) {
	for _, tc := range []struct {
		name, cost, pool string
		power, draws     int
	}{
		{"three mana", "{2}{R}", "RRR", 3, 0},
		{"four mana", "{3}{R}", "RRRR", 4, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			shantotto := effectstest.B12Push(g, me.ID, "Shantotto, Tactician Magician", "Legendary Creature — Dwarf Wizard", effectstest.FinSZ2ShantottoOracle, 0, 4)
			finSZ2CastPaid(t, g, "Sorcery", tc.cost, tc.pool)
			hand := me.Hand.Size()
			effectstest.PassPriorityAroundTable(t, g)
			if p := effectstest.EffectivePower(t, g, shantotto); p != tc.power {
				t.Errorf("power %d, want %d", p, tc.power)
			}
			if got := me.Hand.Size() - hand; got != tc.draws {
				t.Errorf("drew %d, want %d", got, tc.draws)
			}
		})
	}
}

// --- Ultros, Obnoxious Octopus ---------------------------------------------------

func TestFinSZ2UltrosTapsAndStunsOnFourMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
	ultros := effectstest.B12Push(g, me.ID, "Ultros, Obnoxious Octopus", "Legendary Creature — Octopus", effectstest.FinSZ2UltrosOracle, 2, 1)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	finSZ2CastPaid(t, g, "Sorcery", "{3}{U}", "UUUU")
	effectstest.AnswerPickTarget(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedNow(g, victim) || effectstest.CountersOn(g, victim, game.CounterStun) != 1 {
		t.Errorf("the target: tapped %v, %d stun counters; want tapped with one", effectstest.TappedNow(g, victim), effectstest.CountersOn(g, victim, game.CounterStun))
	}
	if got := effectstest.CountersOn(g, ultros, game.CounterPlusOne); got != 0 {
		t.Errorf("four mana put %d +1/+1 counters on Ultros, want 0", got)
	}
}

func TestFinSZ2UltrosGetsEightCountersOnEightMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	// No creature an opponent controls, so the tap trigger has no target
	// and is removed; only the eight-mana trigger is left.
	ultros := effectstest.B12Push(g, me.ID, "Ultros, Obnoxious Octopus", "Legendary Creature — Octopus", effectstest.FinSZ2UltrosOracle, 2, 1)
	finSZ2CastPaid(t, g, "Sorcery", "{7}{U}", "UUUUUUUU")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, ultros, game.CounterPlusOne); got != 8 {
		t.Errorf("Ultros has %d +1/+1 counters, want 8", got)
	}
}

// --- The Prima Vista -------------------------------------------------------------

func TestFinSZ2PrimaVistaAnimatesOnFourMana(t *testing.T) {
	for _, tc := range []struct {
		name, cost, pool string
		want             bool
	}{
		{"three mana", "{2}{U}", "UUU", false},
		{"four mana", "{3}{U}", "UUUU", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			ship := effectstest.B12Push(g, me.ID, "The Prima Vista", "Legendary Artifact — Vehicle", effectstest.FinSZ2PrimaVistaOracl, 5, 3)
			finSZ2CastPaid(t, g, "Sorcery", tc.cost, tc.pool)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.S58p6View(t, g, ship).IsCreature(); got != tc.want {
				t.Errorf("the Vehicle is a creature: %v, want %v", got, tc.want)
			}
		})
	}
}

func TestFinSZ2PrimaVistaCrewsTwoAndFlies(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	ship := s.PushVehicle("The Prima Vista", effectstest.FinSZ2PrimaVistaOracl, 5, 3)
	pilot := s.PushCreature("Pilot", "Creature — Human", 2, 2)
	s.CrewAndResolve(t, ship, 0, pilot)
	if v := effectstest.S58p6View(t, s.G, ship); !v.IsCreature() || v.CurrentPower() != 5 || v.CurrentToughness() != 3 {
		t.Errorf("crewed: creature %v, %d/%d, want a 5/3 creature", v.IsCreature(), v.CurrentPower(), v.CurrentToughness())
	}
	if !effectstest.HasEffectiveKeyword(t, s.G, ship, "flying") {
		t.Error("The Prima Vista has no flying")
	}
}

// --- Sephiroth, Planet's Heir ----------------------------------------------------

func TestFinSZ2SephirothShrinksOpponentsAndGrowsOnTheirDeaths(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
	small1 := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	small2 := effectstest.B12Creature(g, g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)].ID, "Other Bear", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	seph := effectstest.CastCatalogSpell(t, g, "Sephiroth, Planet's Heir", "Legendary Creature — Human Avatar Soldier", effectstest.FinSZ2SephirothOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small1) || g.Battlefield.Contains(small2) {
		t.Fatal("an opponent's 2/2 survived -2/-2")
	}
	if p, tough := effectstest.EffectivePower(t, g, big), effectstest.EffectiveToughness(t, g, big); p != 1 || tough != 1 {
		t.Errorf("the opponent's 3/3 is %d/%d, want 1/1", p, tough)
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("your own creature was shrunk")
	}
	if got := effectstest.CountersOn(g, seph, game.CounterPlusOne); got != 2 {
		t.Errorf("Sephiroth has %d +1/+1 counters after two opposing creatures died, want 2", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, seph, "vigilance") {
		t.Error("Sephiroth has no vigilance")
	}
}

// --- Stuck in Summoner's Sanctum -------------------------------------------------

func TestFinSZ2SanctumTapsLocksAndHoldsAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ring := effectstest.PushCatalogPermanent(g, me.ID, "Sol Ring", "Artifact", effectstest.FinSZ2SolRingOracle, false)
	aura := effectstest.CastCatalogSpell(t, g, "Stuck in Summoner's Sanctum", effectstest.AuraTypeLine, effectstest.FinSZ2SanctumOracle, effectstest.FinABCard(ring))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.AttachmentHostOf(t, g, aura); got.Kind != game.TargetCard || got.ID != ring {
		t.Fatalf("the Aura is attached to %+v, want the artifact", got)
	}
	if !effectstest.TappedNow(g, ring) {
		t.Fatal("the entry trigger did not tap the enchanted artifact")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(ring) })
	if err := g.ActivateManaAbility(me.ID, ring, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("the enchanted artifact's mana ability was activated")
	}
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(ring) })
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if !effectstest.TappedNow(g, ring) {
		t.Error("the enchanted artifact untapped during its controller's untap step")
	}
}

// --- Sleep Magic -----------------------------------------------------------------

func TestFinSZ2SleepMagicHoldsUntilTheCreatureIsDamaged(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.PushPermanentForTest(g, me.ID, "Host", "", "Creature — Bear")
	aura := effectstest.CastCatalogSpell(t, g, "Sleep Magic", effectstest.AuraTypeLine, effectstest.FinSZ2SleepOracle, effectstest.FinABCard(host))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedNow(g, host) {
		t.Fatal("the entry trigger did not tap the enchanted creature")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if !effectstest.TappedNow(g, host) {
		t.Fatal("the enchanted creature untapped during its controller's untap step")
	}
	if !g.Battlefield.Contains(aura) {
		t.Fatal("the Aura left before any damage")
	}
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.Nil, host, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("the Aura is still on the battlefield after its creature was dealt damage")
	}
}

// --- Swallowed by Leviathan ------------------------------------------------------

func TestFinSZ2SwallowedCountsTheGraveyardAfterTheSurveil(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushGraveyardCardTyped(me, "Old Card", "Instant")
	effectstest.SeedLibrary(me, "Top A", "Top B", "Third")
	spell := effectstest.CastCatalogSpell(t, g, "Divination", "Sorcery", effectstest.DivinationDelveOracle, nil)
	effectstest.CastCatalogSpell(t, g, "Swallowed by Leviathan", "Instant", effectstest.FinSZ2SwallowedOracle, effectstest.FinABCard(spell))
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("surveil = %+v, want a look at two cards", c)
	}
	if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("no counter-unless-pays prompt after the surveil")
	}
	tax := effectstest.AnswerPayUnless(t, g, me.ID, false)
	// One card already there plus the two surveilled away.
	if tax.PayCost != "{3}" {
		t.Errorf("the price is %q, want {3}", tax.PayCost)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Stack.Contains(spell) || !me.Graveyard.Contains(spell) {
		t.Error("the target spell was not countered")
	}
}

// --- Town Greeter ----------------------------------------------------------------

func TestFinSZ2TownGreeterTakesATownAndGainsTwo(t *testing.T) {
	for _, tc := range []struct {
		name     string
		pickTown bool
		gain     int
	}{
		{"a Town", true, 2},
		{"a basic land", false, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.SeedLibraryTop(me, "Bear", "Creature — Bear", "{1}{G}")
			forest := effectstest.SeedLibraryTop(me, "Forest", "Basic Land — Forest", "")
			town := effectstest.SeedLibraryTop(me, "Starting Town", "Land — Town", "")
			effectstest.SeedLibraryTop(me, "Shock", "Instant", "{R}")
			life := me.Life
			effectstest.CastCatalogSpell(t, g, "Town Greeter", "Creature — Human Citizen", effectstest.FinSZ2GreeterOracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			c := effectstest.ChooseCardsChoiceFor(g, me.ID)
			if c == nil || len(c.ChooseCards) != 2 {
				t.Fatalf("offer = %+v, want the two milled lands", c)
			}
			pick := forest
			if tc.pickTown {
				pick = town
			}
			effectstest.AnswerChooseCards(t, g, me.ID, pick)
			if !me.Hand.Contains(pick) {
				t.Fatal("the chosen land is not in your hand")
			}
			if me.Life != life+tc.gain {
				t.Errorf("life %d → %d, want +%d", life, me.Life, tc.gain)
			}
		})
	}
}

// --- Zell Dincht -----------------------------------------------------------------

func TestFinSZ2ZellCountsLandsAndReturnsOneAtYourEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	zell := effectstest.B12Push(g, me.ID, "Zell Dincht", "Legendary Creature — Human Monk", effectstest.FinSZ2ZellOracle, 0, 3)
	land := effectstest.B12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	effectstest.B12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	effectstest.B12Permanent(g, effectstest.FinABOpp(g).ID, "Their Mountain", "Basic Land — Mountain")
	if p := effectstest.EffectivePower(t, g, zell); p != 2 {
		t.Errorf("Zell's power is %d with two lands, want 2", p)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	for i := 0; i < 8 && effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	effectstest.AnswerOwnPermanents(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, land) != game.ZoneHand {
		t.Errorf("the chosen land is in %v, want its owner's hand", effectstest.B12ZoneOf(g, land))
	}
	if p := effectstest.EffectivePower(t, g, zell); p != 1 {
		t.Errorf("Zell's power is %d with one land, want 1", p)
	}
}
