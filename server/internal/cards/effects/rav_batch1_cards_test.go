package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rav_batch1_cards_test.go — Ravnica: City of Guilds, batch 1 (#5600).
// Each test plays the real catalog entry and asserts what its text says.

const (
	ravB1BorosRecruitOracle     = "00abb43b-550f-42d2-9049-61ac56ff5d8d"
	ravB1BorosSwiftbladeOracle  = "377e4dad-866b-4e02-8c10-072bf7f000be"
	ravB1CourierHawkOracle      = "7fc904aa-377b-4671-8308-275205f56d81"
	ravB1CharOracle             = "e57d90ad-2f54-46ca-8237-aa383353a0de"
	ravB1CarrionHowlerOracle    = "e139176f-483f-4f05-af3d-0176d0048b62"
	ravB1CentaurSafeguardOracle = "468604f2-52e6-4bb1-8a4f-ad0851288cba"
	ravB1CivicWayfinderOracle   = "b9e53414-c21f-4479-b8c1-8b36c3e6e006"
	ravB1ClingingDarknessOracle = "5ad3abf3-1ae5-4c4c-9950-ba7bff0d75ad"
	ravB1ConvoluteOracle        = "cd201ded-cfaf-4a51-9fcc-6a61b3c910b0"
	ravB1EquenautOracle         = "4127794b-84d1-4efb-b0e9-543a02b07fa8"
	ravB1AutochthonWurmOracle   = "18379907-ddba-479e-8953-3adc1ed40458"
	ravB1PhalanxOracle          = "5a1bebe9-c940-4fac-8572-0833808eaa55"
	ravB1BrainspoilOracle       = "dd464a76-5f96-4e21-b7c6-4a476e7f9ce6"
	ravB1ClutchOracle           = "316cae8b-31b0-4b76-8b72-8122005e1b4e"
	ravB1StartingLife           = 40
)

func TestRavB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		ravB1BorosRecruitOracle:     "Boros Recruit",
		ravB1BorosSwiftbladeOracle:  "Boros Swiftblade",
		ravB1CourierHawkOracle:      "Courier Hawk",
		ravB1CharOracle:             "Char",
		ravB1CarrionHowlerOracle:    "Carrion Howler",
		ravB1CentaurSafeguardOracle: "Centaur Safeguard",
		ravB1CivicWayfinderOracle:   "Civic Wayfinder",
		ravB1ClingingDarknessOracle: "Clinging Darkness",
		ravB1ConvoluteOracle:        "Convolute",
		ravB1EquenautOracle:         "Conclave Equenaut",
		ravB1AutochthonWurmOracle:   "Autochthon Wurm",
		ravB1PhalanxOracle:          "Conclave Phalanx",
		ravB1BrainspoilOracle:       "Brainspoil",
		ravB1ClutchOracle:           "Clutch of the Undercity",
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

// ravB1ConvokeCast puts a creature card with its printed cost and body in
// the active seat's hand and casts it under the strict mana gate with an
// empty pool, paying only by tapping the given creatures (convoke).
func ravB1ConvokeCast(t *testing.T, g *game.Game, name, typeLine, manaCost, oracle string, power, toughness int, taps []uuid.UUID) (uuid.UUID, error) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: manaCost, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: active.ID, Controller: active.ID,
	})
	toMainForCost(t, g)
	return id, g.CastSpell(active.ID, id, game.CastSpellParams{Strict: true, TapIDs: taps})
}

// ravB1Tappers seeds n untapped creatures of one colour for the active
// seat to convoke with.
func ravB1Tappers(g *game.Game, owner uuid.UUID, n int, colour string) []uuid.UUID {
	ids := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		ids = append(ids, pushTapCostPermanent(g, owner, "Helper", "Creature — Soldier", colour))
	}
	return ids
}

// --- Boros Recruit, Boros Swiftblade, Courier Hawk ---------------------------

func TestRavB1KeywordCreaturesHaveTheirKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		keywords               []string
	}{
		{"Boros Recruit", "Creature — Goblin Soldier", ravB1BorosRecruitOracle, 1, 1, []string{"first strike"}},
		{"Boros Swiftblade", "Creature — Human Soldier", ravB1BorosSwiftbladeOracle, 1, 2, []string{"double strike"}},
		{"Courier Hawk", "Creature — Bird", ravB1CourierHawkOracle, 1, 2, []string{"flying", "vigilance"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			for _, kw := range tc.keywords {
				if !p1g2HasKeyword(g, id, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

func TestRavB1BorosSwiftbladeDealsDamageTwice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	blade := b12Push(g, me.ID, "Boros Swiftblade", "Creature — Human Soldier", ravB1BorosSwiftbladeOracle, 1, 2)
	attackWith(t, g, opp.ID, blade)
	advanceTo(t, g, game.StepEndCombat)
	if opp.Life != ravB1StartingLife-2 {
		t.Errorf("an unblocked 1/2 with double strike: life %d, want %d", opp.Life, ravB1StartingLife-2)
	}
}

func TestRavB1CourierHawkAttacksWithoutTapping(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hawk := b12Push(g, me.ID, "Courier Hawk", "Creature — Bird", ravB1CourierHawkOracle, 1, 2)
	attackWith(t, g, opp.ID, hawk)
	if tapCostTapped(g, hawk) {
		t.Error("vigilance: attacking tapped the Hawk")
	}
	if opp.Life != ravB1StartingLife-1 {
		t.Errorf("the Hawk connected for %d, want 1", ravB1StartingLife-opp.Life)
	}
}

// --- Char -----------------------------------------------------------------------

func TestRavB1CharHitsAPlayerAndItsCaster(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	castCatalogSpell(t, g, "Char", "Instant", ravB1CharOracle, playerRefs(opp.ID))
	passPriorityAroundTable(t, g)
	if opp.Life != ravB1StartingLife-4 {
		t.Errorf("the target's life %d, want %d", opp.Life, ravB1StartingLife-4)
	}
	if me.Life != ravB1StartingLife-2 {
		t.Errorf("the caster's life %d, want %d", me.Life, ravB1StartingLife-2)
	}
}

func TestRavB1CharKillsACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)
	castCatalogSpell(t, g, "Char", "Instant", ravB1CharOracle, cardRefs(victim))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("a 4/4 survived 4 damage")
	}
	if me.Life != ravB1StartingLife-2 {
		t.Errorf("the caster's life %d, want %d", me.Life, ravB1StartingLife-2)
	}
}

func TestRavB1CharWithItsTargetGoneDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)
	castCatalogSpell(t, g, "Char", "Instant", ravB1CharOracle, cardRefs(victim))
	hobDestroy(t, g, victim)
	passPriorityAroundTable(t, g)
	if me.Life != ravB1StartingLife {
		t.Errorf("a Char with no legal target still dealt the caster %d", ravB1StartingLife-me.Life)
	}
}

// --- Carrion Howler -------------------------------------------------------------

func TestRavB1CarrionHowlerTradesToughnessForPowerForLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	howler := b12Push(g, me.ID, "Carrion Howler", "Creature — Zombie Wolf", ravB1CarrionHowlerOracle, 2, 2)
	if err := hobB3Activate(t, g, me.ID, howler, ""); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != ravB1StartingLife-1 {
		t.Errorf("life %d, want %d", me.Life, ravB1StartingLife-1)
	}
	if p, tough := effectivePower(t, g, howler), effectiveToughness(t, g, howler); p != 4 || tough != 1 {
		t.Errorf("after one activation it is %d/%d, want 4/1", p, tough)
	}
	// No limit on activations: a second one leaves it 6/0, and it dies.
	if err := hobB3Activate(t, g, me.ID, howler, ""); err != nil {
		t.Fatalf("second activation: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(howler) {
		t.Error("a 6/0 is still on the battlefield")
	}
}

func TestRavB1CarrionHowlerCantPayLifeItDoesntHave(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	howler := b12Push(g, me.ID, "Carrion Howler", "Creature — Zombie Wolf", ravB1CarrionHowlerOracle, 2, 2)
	me.Life = 0
	if err := hobB3Activate(t, g, me.ID, howler, ""); err == nil {
		t.Fatal("paid 1 life at 0 life")
	}
}

// --- Centaur Safeguard ----------------------------------------------------------

func TestRavB1CentaurSafeguardMayGainThreeWhenItDies(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		centaur := b12Push(g, me.ID, "Centaur Safeguard", "Creature — Centaur Warrior", ravB1CentaurSafeguardOracle, 3, 1)
		hobDestroy(t, g, centaur)
		answerLatestTriggerPrompt(t, g, me.ID, yes)
		passPriorityAroundTable(t, g)
		want := ravB1StartingLife
		if yes {
			want += 3
		}
		if me.Life != want {
			t.Errorf("answered %v: life %d, want %d", yes, me.Life, want)
		}
	}
}

// --- Civic Wayfinder ------------------------------------------------------------

func TestRavB1CivicWayfinderFetchesABasicLand(t *testing.T) {
	jmpB6FetchesABasicToHand(t, "Civic Wayfinder", "Creature — Elf Druid Warrior", ravB1CivicWayfinderOracle, 2, 2)
}

func TestRavB1CivicWayfinderMayDecline(t *testing.T) {
	jmpB6DeclinesTheSearch(t, "Civic Wayfinder", "Creature — Elf Druid Warrior", ravB1CivicWayfinderOracle, 2, 2)
}

// --- Clinging Darkness ----------------------------------------------------------

func TestRavB1ClingingDarknessShrinksAndKills(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	big := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)
	castCatalogSpell(t, g, "Clinging Darkness", "Enchantment — Aura", ravB1ClingingDarknessOracle, cardRefs(big))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, big), effectiveToughness(t, g, big); p != 1 || tough != 4 {
		t.Errorf("the enchanted 5/5 is %d/%d, want 1/4", p, tough)
	}

	small := b12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 3, 1)
	castCatalogSpell(t, g, "Clinging Darkness", "Enchantment — Aura", ravB1ClingingDarknessOracle, cardRefs(small))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("a 3/1 given -4/-1 did not die")
	}
}

// --- Convolute ------------------------------------------------------------------

// ravB1ConvoluteBoard has seat 0 cast a sorcery and seat 1 answer it with
// Convolute, stopping at the tax prompt.
func ravB1ConvoluteBoard(t *testing.T) (g *game.Game, victim uuid.UUID, caster *game.Player) {
	t.Helper()
	g = newCatalogGame(t)
	caster, them := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	victim = castCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
	convolute := handCardFull(them, "Convolute", "Instant", "", ravB1ConvoluteOracle, []string{"U"})
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority to the responder: %v", err)
	}
	if err := g.CastSpell(them.ID, convolute, game.CastSpellParams{Targets: cardRefs(victim)}); err != nil {
		t.Fatalf("casting Convolute: %v", err)
	}
	passUntilTaxed(t, g, caster.ID)
	return g, victim, caster
}

func TestRavB1ConvoluteCountersUnlessItsControllerPaysFour(t *testing.T) {
	g, victim, caster := ravB1ConvoluteBoard(t)
	prompt := answerPayUnless(t, g, caster.ID, false)
	if prompt.PayCost != "{4}" {
		t.Errorf("the tax is %q, want {4}", prompt.PayCost)
	}
	if g.Stack.Contains(victim) {
		t.Error("a declined tax must counter the spell")
	}
}

func TestRavB1ConvoluteLetsTheSpellResolveWhenTheTaxIsPaid(t *testing.T) {
	g, victim, caster := ravB1ConvoluteBoard(t)
	for i := 0; i < 4; i++ {
		caster.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
	answerPayUnless(t, g, caster.ID, true)
	if !g.Stack.Contains(victim) {
		t.Error("a paid tax countered the spell anyway")
	}
}

func TestRavB1ConvoluteCantBePaidWithThree(t *testing.T) {
	g, victim, caster := ravB1ConvoluteBoard(t)
	for i := 0; i < 3; i++ {
		caster.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
	answerPayUnless(t, g, caster.ID, true)
	if g.Stack.Contains(victim) {
		t.Error("three mana paid a {4} tax")
	}
}

// --- Conclave Equenaut, Autochthon Wurm, Conclave Phalanx -------------------------

func TestRavB1ConclaveEquenautConvokesAndFlies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	taps := ravB1Tappers(g, me.ID, 6, "W")
	id, err := ravB1ConvokeCast(t, g, "Conclave Equenaut", "Creature — Human Soldier", "{4}{W}{W}", ravB1EquenautOracle, 3, 3, taps)
	if err != nil {
		t.Fatalf("convoke cast with six white creatures: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Equenaut never resolved")
	}
	if !p1g2HasKeyword(g, id, "flying") {
		t.Error("the Equenaut has no flying")
	}
}

func TestRavB1ConclaveEquenautNeedsEveryPip(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	taps := ravB1Tappers(g, me.ID, 5, "W")
	if _, err := ravB1ConvokeCast(t, g, "Conclave Equenaut", "Creature — Human Soldier", "{4}{W}{W}", ravB1EquenautOracle, 3, 3, taps); err == nil {
		t.Fatal("five creatures paid a six-mana cost")
	}
	// Two green creatures can't pay {W}{W}.
	g = newCatalogGame(t)
	me = g.Seats[0]
	taps = append(ravB1Tappers(g, me.ID, 4, "W"), ravB1Tappers(g, me.ID, 2, "G")...)
	if _, err := ravB1ConvokeCast(t, g, "Conclave Equenaut", "Creature — Human Soldier", "{4}{W}{W}", ravB1EquenautOracle, 3, 3, taps); err != nil {
		t.Fatalf("four white and two green creatures should pay {4}{W}{W}: %v", err)
	}
	g = newCatalogGame(t)
	me = g.Seats[0]
	taps = append(ravB1Tappers(g, me.ID, 1, "W"), ravB1Tappers(g, me.ID, 5, "G")...)
	if _, err := ravB1ConvokeCast(t, g, "Conclave Equenaut", "Creature — Human Soldier", "{4}{W}{W}", ravB1EquenautOracle, 3, 3, taps); err == nil {
		t.Fatal("one white creature paid {W}{W}")
	}
}

func TestRavB1AutochthonWurmConvokesFifteenAndTramples(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	taps := append(ravB1Tappers(g, me.ID, 8, "G"), ravB1Tappers(g, me.ID, 6, "W")...)
	if _, err := ravB1ConvokeCast(t, g, "Autochthon Wurm", "Creature — Wurm", "{10}{G}{G}{G}{W}{W}", ravB1AutochthonWurmOracle, 9, 14, taps); err == nil {
		t.Fatal("fourteen creatures paid a fifteen-mana cost")
	}
	g = newCatalogGame(t)
	me = g.Seats[0]
	taps = append(ravB1Tappers(g, me.ID, 8, "G"), ravB1Tappers(g, me.ID, 7, "W")...)
	id, err := ravB1ConvokeCast(t, g, "Autochthon Wurm", "Creature — Wurm", "{10}{G}{G}{G}{W}{W}", ravB1AutochthonWurmOracle, 9, 14, taps)
	if err != nil {
		t.Fatalf("convoke cast with fifteen creatures: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Wurm never resolved")
	}
	if !p1g2HasKeyword(g, id, "trample") {
		t.Error("the Wurm has no trample")
	}
}

func TestRavB1ConclavePhalanxGainsOnePerCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	taps := ravB1Tappers(g, me.ID, 5, "W")
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	id, err := ravB1ConvokeCast(t, g, "Conclave Phalanx", "Creature — Human Soldier", "{4}{W}", ravB1PhalanxOracle, 2, 4, taps)
	if err != nil {
		t.Fatalf("convoke cast with five white creatures: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Phalanx never resolved")
	}
	// The five tapped creatures and the Phalanx itself; not the opponent's.
	if me.Life != ravB1StartingLife+6 {
		t.Errorf("life %d, want %d", me.Life, ravB1StartingLife+6)
	}
}

// --- transmute (Brainspoil, Clutch of the Undercity) ---------------------------

// ravB1Transmutes puts a transmute card in the active seat's hand, floats
// the ability's mana, activates it, and checks the search offers the
// library card with the same mana value and not the one with another.
func ravB1Transmutes(t *testing.T, name, typeLine, manaCost, oracle string, mana []string, sameCost, otherCost string) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: manaCost,
		OracleID: oracle, Owner: me.ID, Controller: me.ID,
	})
	same := pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Same Value", TypeLine: "Sorcery", ManaCost: sameCost, Owner: me.ID, Controller: me.ID})
	pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Other Value", TypeLine: "Sorcery", ManaCost: otherCost, Owner: me.ID, Controller: me.ID})
	b10AddMana(me, mana...)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("transmute: %v", err)
	}
	if !me.Graveyard.Contains(id) {
		t.Fatalf("%s was not discarded to pay for transmute", name)
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Other Value") != uuid.Nil {
		t.Error("the search offered a card with another mana value")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{same}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !me.Hand.Contains(same) {
		t.Error("the card with the same mana value did not reach the hand")
	}
}

func TestRavB1BrainspoilTransmutesForManaValueFive(t *testing.T) {
	ravB1Transmutes(t, "Brainspoil", "Sorcery", "{3}{B}{B}", ravB1BrainspoilOracle,
		[]string{"C", "B", "B"}, "{4}{U}", "{3}{U}")
}

func TestRavB1ClutchOfTheUndercityTransmutesForManaValueFour(t *testing.T) {
	ravB1Transmutes(t, "Clutch of the Undercity", "Instant", "{1}{U}{U}{B}", ravB1ClutchOracle,
		[]string{"C", "U", "B"}, "{2}{G}{G}", "{4}{U}")
}

// --- Brainspoil -----------------------------------------------------------------

func TestRavB1BrainspoilDestroysThroughRegeneration(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Troll", "Creature — Troll", 2, 2)
	g.WithWriteLock(func() { _ = g.RegenerateForEffect(victim) })
	castCatalogSpell(t, g, "Brainspoil", "Sorcery", ravB1BrainspoilOracle, cardRefs(victim))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("a regeneration shield saved the creature; it can't be regenerated")
	}
}

func TestRavB1BrainspoilCantTargetAnEnchantedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 6, 6)
	pushAuraOnLand(t, g, me.ID, "Clinging Darkness", ravB1ClingingDarknessOracle, host)
	if err := castCatalogSpellErr(t, g, "Brainspoil", "Sorcery", ravB1BrainspoilOracle, cardRefs(host)); err == nil {
		t.Fatal("Brainspoil targeted an enchanted creature")
	}
}

func TestRavB1BrainspoilFizzlesOnACreatureEnchantedInResponse(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 6, 6)
	castCatalogSpell(t, g, "Brainspoil", "Sorcery", ravB1BrainspoilOracle, cardRefs(victim))
	pushAuraOnLand(t, g, me.ID, "Clinging Darkness", ravB1ClingingDarknessOracle, victim)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(victim) {
		t.Error("an enchanted creature is no longer a legal target, so it is not destroyed")
	}
}

// --- Clutch of the Undercity ----------------------------------------------------

func TestRavB1ClutchBouncesAndItsControllerLosesThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	castCatalogSpell(t, g, "Clutch of the Undercity", "Instant", ravB1ClutchOracle, cardRefs(victim))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) || !opp.Hand.Contains(victim) {
		t.Error("the permanent did not return to its owner's hand")
	}
	if opp.Life != ravB1StartingLife-3 {
		t.Errorf("its controller's life %d, want %d", opp.Life, ravB1StartingLife-3)
	}
	if me.Life != ravB1StartingLife {
		t.Errorf("the caster lost life: %d", me.Life)
	}
}

func TestRavB1ClutchOnYourOwnPermanentCostsYouThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Clutch of the Undercity", "Instant", ravB1ClutchOracle, cardRefs(mine))
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(mine) {
		t.Error("the permanent did not return to its owner's hand")
	}
	if me.Life != ravB1StartingLife-3 {
		t.Errorf("life %d, want %d", me.Life, ravB1StartingLife-3)
	}
}

func TestRavB1ClutchWithItsTargetGoneCostsNobodyLife(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	castCatalogSpell(t, g, "Clutch of the Undercity", "Instant", ravB1ClutchOracle, cardRefs(victim))
	hobDestroy(t, g, victim)
	passPriorityAroundTable(t, g)
	if opp.Life != ravB1StartingLife {
		t.Errorf("a fizzled Clutch still cost its target's controller %d", ravB1StartingLife-opp.Life)
	}
}
