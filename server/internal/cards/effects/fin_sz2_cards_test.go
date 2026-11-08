package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fin_sz2_cards_test.go — Final Fantasy (FIN), the cards S–Z, batch 2
// (#4462). Each test plays the real catalog entry and asserts what its
// text says. The fixtures finABOpp, finABCard and finABAttack are the
// A–B batch's.

const (
	finSZ2SahaginOracle   = "24310f49-2156-4c0d-b3b1-40ca3182474f"
	finSZ2ShantottoOracle = "381df197-5fb7-44ab-a31f-96c4993a0e53"
	finSZ2UltrosOracle    = "b6f4b524-1081-4263-b7e4-9aee93131b31"
	finSZ2PrimaVistaOracl = "a09d9754-a321-4018-a237-e613e915b642"
	finSZ2SephirothOracle = "6af71c5a-2c1b-4120-b3c9-c3c70d94a5c9"
	finSZ2SanctumOracle   = "84d308ce-c3c5-4612-accc-6e6cc68a33be"
	finSZ2SleepOracle     = "33d6f4f0-e1ad-4363-a012-46a1bd9daeb7"
	finSZ2SwallowedOracle = "1a8185ed-636b-49ad-bf24-2ab02c7fb91e"
	finSZ2GreeterOracle   = "9dc61ef6-da22-42a8-b348-dfd5722deadf"
	finSZ2ZellOracle      = "f207ce73-e4cd-4dfc-b3b9-5704f5148959"
	finSZ2SolRingOracle   = "6ad8011d-3471-4369-9d68-b264cc027487"
)

func TestFinSZ2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		finSZ2SahaginOracle:   "Sahagin",
		finSZ2ShantottoOracle: "Shantotto, Tactician Magician",
		finSZ2UltrosOracle:    "Ultros, Obnoxious Octopus",
		finSZ2PrimaVistaOracl: "The Prima Vista",
		finSZ2SephirothOracle: "Sephiroth, Planet's Heir",
		finSZ2SanctumOracle:   "Stuck in Summoner's Sanctum",
		finSZ2SleepOracle:     "Sleep Magic",
		finSZ2SwallowedOracle: "Swallowed by Leviathan",
		finSZ2GreeterOracle:   "Town Greeter",
		finSZ2ZellOracle:      "Zell Dincht",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// finSZ2CastPaid casts a plain spell from the active seat's hand, paid
// strictly from `pool`, so the payment record holds that much mana.
func finSZ2CastPaid(t *testing.T, g *game.Game, typeLine, cost, pool string) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	id := socB17Hand(me, "Test Spell", typeLine, cost, socB17PlainSorceryOracle)
	floatForTest(g, me, pool)
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
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			sahagin := pushCatalogPermanent(g, me.ID, "Sahagin", "Creature — Merfolk Warrior", finSZ2SahaginOracle, false)
			finSZ2CastPaid(t, g, tc.typeLine, tc.cost, tc.pool)
			passPriorityAroundTable(t, g)
			counters := countersOn(g, sahagin, game.CounterPlusOne)
			unblockable := restrictionsOf(t, g, sahagin)&game.CantBeBlocked != 0
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
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			shantotto := b12Push(g, me.ID, "Shantotto, Tactician Magician", "Legendary Creature — Dwarf Wizard", finSZ2ShantottoOracle, 0, 4)
			finSZ2CastPaid(t, g, "Sorcery", tc.cost, tc.pool)
			hand := me.Hand.Size()
			passPriorityAroundTable(t, g)
			if p := effectivePower(t, g, shantotto); p != tc.power {
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
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], finABOpp(g)
	ultros := b12Push(g, me.ID, "Ultros, Obnoxious Octopus", "Legendary Creature — Octopus", finSZ2UltrosOracle, 2, 1)
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	finSZ2CastPaid(t, g, "Sorcery", "{3}{U}", "UUUU")
	answerPickTarget(t, g, victim)
	passPriorityAroundTable(t, g)
	if !tappedNow(g, victim) || countersOn(g, victim, game.CounterStun) != 1 {
		t.Errorf("the target: tapped %v, %d stun counters; want tapped with one", tappedNow(g, victim), countersOn(g, victim, game.CounterStun))
	}
	if got := countersOn(g, ultros, game.CounterPlusOne); got != 0 {
		t.Errorf("four mana put %d +1/+1 counters on Ultros, want 0", got)
	}
}

func TestFinSZ2UltrosGetsEightCountersOnEightMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	// No creature an opponent controls, so the tap trigger has no target
	// and is removed; only the eight-mana trigger is left.
	ultros := b12Push(g, me.ID, "Ultros, Obnoxious Octopus", "Legendary Creature — Octopus", finSZ2UltrosOracle, 2, 1)
	finSZ2CastPaid(t, g, "Sorcery", "{7}{U}", "UUUUUUUU")
	passPriorityAroundTable(t, g)
	if got := countersOn(g, ultros, game.CounterPlusOne); got != 8 {
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
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			ship := b12Push(g, me.ID, "The Prima Vista", "Legendary Artifact — Vehicle", finSZ2PrimaVistaOracl, 5, 3)
			finSZ2CastPaid(t, g, "Sorcery", tc.cost, tc.pool)
			passPriorityAroundTable(t, g)
			if got := s58p6View(t, g, ship).IsCreature(); got != tc.want {
				t.Errorf("the Vehicle is a creature: %v, want %v", got, tc.want)
			}
		})
	}
}

func TestFinSZ2PrimaVistaCrewsTwoAndFlies(t *testing.T) {
	s := newCrewScene(t)
	ship := s.pushVehicle("The Prima Vista", finSZ2PrimaVistaOracl, 5, 3)
	pilot := s.pushCreature("Pilot", "Creature — Human", 2, 2)
	s.crewAndResolve(t, ship, 0, pilot)
	if v := s58p6View(t, s.g, ship); !v.IsCreature() || v.CurrentPower() != 5 || v.CurrentToughness() != 3 {
		t.Errorf("crewed: creature %v, %d/%d, want a 5/3 creature", v.IsCreature(), v.CurrentPower(), v.CurrentToughness())
	}
	if !hasEffectiveKeyword(t, s.g, ship, "flying") {
		t.Error("The Prima Vista has no flying")
	}
}

// --- Sephiroth, Planet's Heir ----------------------------------------------------

func TestFinSZ2SephirothShrinksOpponentsAndGrowsOnTheirDeaths(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], finABOpp(g)
	small1 := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	small2 := b12Creature(g, g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)].ID, "Other Bear", "Creature — Bear", 2, 2)
	big := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 3, 3)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	seph := castCatalogSpell(t, g, "Sephiroth, Planet's Heir", "Legendary Creature — Human Avatar Soldier", finSZ2SephirothOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small1) || g.Battlefield.Contains(small2) {
		t.Fatal("an opponent's 2/2 survived -2/-2")
	}
	if p, tough := effectivePower(t, g, big), effectiveToughness(t, g, big); p != 1 || tough != 1 {
		t.Errorf("the opponent's 3/3 is %d/%d, want 1/1", p, tough)
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("your own creature was shrunk")
	}
	if got := countersOn(g, seph, game.CounterPlusOne); got != 2 {
		t.Errorf("Sephiroth has %d +1/+1 counters after two opposing creatures died, want 2", got)
	}
	if !hasEffectiveKeyword(t, g, seph, "vigilance") {
		t.Error("Sephiroth has no vigilance")
	}
}

// --- Stuck in Summoner's Sanctum -------------------------------------------------

func TestFinSZ2SanctumTapsLocksAndHoldsAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ring := pushCatalogPermanent(g, me.ID, "Sol Ring", "Artifact", finSZ2SolRingOracle, false)
	aura := castCatalogSpell(t, g, "Stuck in Summoner's Sanctum", auraTypeLine, finSZ2SanctumOracle, finABCard(ring))
	passPriorityAroundTable(t, g)
	if got := attachmentHostOf(t, g, aura); got.Kind != game.TargetCard || got.ID != ring {
		t.Fatalf("the Aura is attached to %+v, want the artifact", got)
	}
	if !tappedNow(g, ring) {
		t.Fatal("the entry trigger did not tap the enchanted artifact")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(ring) })
	if err := g.ActivateManaAbility(me.ID, ring, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("the enchanted artifact's mana ability was activated")
	}
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(ring) })
	advanceToUpkeepOf(t, g, 0)
	if !tappedNow(g, ring) {
		t.Error("the enchanted artifact untapped during its controller's untap step")
	}
}

// --- Sleep Magic -----------------------------------------------------------------

func TestFinSZ2SleepMagicHoldsUntilTheCreatureIsDamaged(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	host := pushPermanentForTest(g, me.ID, "Host", "", "Creature — Bear")
	aura := castCatalogSpell(t, g, "Sleep Magic", auraTypeLine, finSZ2SleepOracle, finABCard(host))
	passPriorityAroundTable(t, g)
	if !tappedNow(g, host) {
		t.Fatal("the entry trigger did not tap the enchanted creature")
	}
	advanceToUpkeepOf(t, g, 0)
	if !tappedNow(g, host) {
		t.Fatal("the enchanted creature untapped during its controller's untap step")
	}
	if !g.Battlefield.Contains(aura) {
		t.Fatal("the Aura left before any damage")
	}
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.Nil, host, 1) })
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("the Aura is still on the battlefield after its creature was dealt damage")
	}
}

// --- Swallowed by Leviathan ------------------------------------------------------

func TestFinSZ2SwallowedCountsTheGraveyardAfterTheSurveil(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushGraveyardCardTyped(me, "Old Card", "Instant")
	seedLibrary(me, "Top A", "Top B", "Third")
	spell := castCatalogSpell(t, g, "Divination", "Sorcery", divinationDelveOracle, nil)
	castCatalogSpell(t, g, "Swallowed by Leviathan", "Instant", finSZ2SwallowedOracle, finABCard(spell))
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("surveil = %+v, want a look at two cards", c)
	}
	if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("no counter-unless-pays prompt after the surveil")
	}
	tax := answerPayUnless(t, g, me.ID, false)
	// One card already there plus the two surveilled away.
	if tax.PayCost != "{3}" {
		t.Errorf("the price is %q, want {3}", tax.PayCost)
	}
	passPriorityAroundTable(t, g)
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
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			seedLibraryTop(me, "Bear", "Creature — Bear", "{1}{G}")
			forest := seedLibraryTop(me, "Forest", "Basic Land — Forest", "")
			town := seedLibraryTop(me, "Starting Town", "Land — Town", "")
			seedLibraryTop(me, "Shock", "Instant", "{R}")
			life := me.Life
			castCatalogSpell(t, g, "Town Greeter", "Creature — Human Citizen", finSZ2GreeterOracle, nil)
			passPriorityAroundTable(t, g)
			c := chooseCardsChoiceFor(g, me.ID)
			if c == nil || len(c.ChooseCards) != 2 {
				t.Fatalf("offer = %+v, want the two milled lands", c)
			}
			pick := forest
			if tc.pickTown {
				pick = town
			}
			answerChooseCards(t, g, me.ID, pick)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	zell := b12Push(g, me.ID, "Zell Dincht", "Legendary Creature — Human Monk", finSZ2ZellOracle, 0, 3)
	land := b12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	b12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	b12Permanent(g, finABOpp(g).ID, "Their Mountain", "Basic Land — Mountain")
	if p := effectivePower(t, g, zell); p != 2 {
		t.Errorf("Zell's power is %d with two lands, want 2", p)
	}
	advanceTo(t, g, game.StepEnd)
	for i := 0; i < 8 && latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	answerOwnPermanents(t, g, me.ID, land)
	passPriorityAroundTable(t, g)
	if b12ZoneOf(g, land) != game.ZoneHand {
		t.Errorf("the chosen land is in %v, want its owner's hand", b12ZoneOf(g, land))
	}
	if p := effectivePower(t, g, zell); p != 1 {
		t.Errorf("Zell's power is %d with one land, want 1", p)
	}
}
