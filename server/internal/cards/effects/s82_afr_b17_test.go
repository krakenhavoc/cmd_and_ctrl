package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_afr_b17_test.go — Adventures in the Forgotten Realms, S82 batch 17:
// the five creature lands, Circle of Dreams Druid, Blink Dog, Divine Smite,
// Dragon Turtle and Ray of Frost, each driven through the engine.

const (
	afrDenOracle      = "f451b8f0-1ff5-4e8d-9f30-9352d83ed687"
	afrHiveOracle     = "d17163d4-dd43-4de6-b7cf-576448160b7f"
	afrLairOracle     = "126e9140-2c05-4c00-8b01-5653456c736a"
	afrCaveOracle     = "1e4146d2-cfa0-4f5e-9761-3c83519b90c3"
	afrHallOracle     = "087c8c0e-a91c-4e3c-8387-9312db01f343"
	afrCircleOracle   = "1c857fa4-e823-494d-9ea4-7eb9e719c9b6"
	afrBlinkDogOracle = "02a7d95a-0a45-4aa3-b5b8-01c0b36462d8"
	afrSmiteOracle    = "3b06b242-caed-4c8f-b5ab-30e86061286e"
	afrTurtleOracle   = "a3270e64-38d2-49e5-8a1e-a5e81205776b"
	afrRayOracle      = "3de32046-454b-40be-b522-a3d1d5302b10"
)

func TestAFRB17CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		afrDenOracle: "Den of the Bugbear", afrHiveOracle: "Hive of the Eye Tyrant",
		afrLairOracle: "Lair of the Hydra", afrCaveOracle: "Cave of the Frost Dragon",
		afrHallOracle: "Hall of Storm Giants", afrCircleOracle: "Circle of Dreams Druid",
		afrBlinkDogOracle: "Blink Dog", afrSmiteOracle: "Divine Smite",
		afrTurtleOracle: "Dragon Turtle", afrRayOracle: "Ray of Frost",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name {
			t.Fatalf("Lookup(%s) = %q, %v", oracle, spec.Name, ok)
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %s, caveats %v", name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- the creature lands --------------------------------------------------

var afrLands = []struct{ name, oracle string }{
	{"Den of the Bugbear", afrDenOracle},
	{"Hive of the Eye Tyrant", afrHiveOracle},
	{"Lair of the Hydra", afrLairOracle},
	{"Cave of the Frost Dragon", afrCaveOracle},
	{"Hall of Storm Giants", afrHallOracle},
}

// "If you control two or more other lands, this land enters tapped."
func TestAFRB17LandsEnterTappedWithTwoOrMoreOtherLands(t *testing.T) {
	for _, l := range afrLands {
		t.Run(l.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			s58p6Push(g, me.ID, "Plains", "Basic Land — Plains", "", "", 0, 0)
			one := b12PlayFromHand(t, g, l.name, "Land", l.oracle, game.CastSpellParams{})
			if isTapped(g, one) {
				t.Error("entered tapped with only one other land")
			}
		})
		t.Run(l.name+" tapped", func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			s58p6Push(g, me.ID, "Plains", "Basic Land — Plains", "", "", 0, 0)
			s58p6Push(g, me.ID, "Island", "Basic Land — Island", "", "", 0, 0)
			id := b12PlayFromHand(t, g, l.name, "Land", l.oracle, game.CastSpellParams{})
			if !isTapped(g, id) {
				t.Error("entered untapped with two other lands")
			}
		})
	}
}

func afrAnimate(t *testing.T, g *game.Game, me *game.Player, land uuid.UUID, params game.ActivateAbilityParams, colour string, generic int) {
	t.Helper()
	fillPool(me, generic)
	fillPoolColored(me, colour, 1)
	s58p6Activate(t, g, me.ID, land, params)
	passPriorityAroundTable(t, g)
}

func afrPlayedLand(g *game.Game, me *game.Player, name, oracle string) uuid.UUID {
	return s58p6Push(g, me.ID, name, "Land", oracle, "", 0, 0)
}

func TestAFRB17DenAnimatesAndItsAttackMakesATappedAttackingGoblin(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	den := afrPlayedLand(g, me, "Den of the Bugbear", afrDenOracle)
	advanceTo(t, g, game.StepPrecombatMain)
	if s58p6View(t, g, den).IsCreature() {
		t.Fatal("an idle Den is a creature")
	}
	afrAnimate(t, g, me, den, game.ActivateAbilityParams{}, "R", 3)
	c := s58p6View(t, g, den)
	if !c.IsCreature() || !c.IsLand() || c.CurrentPower() != 3 || c.CurrentToughness() != 2 {
		t.Fatalf("creature=%v land=%v %d/%d, want a 3/2 creature land", c.IsCreature(), c.IsLand(), c.CurrentPower(), c.CurrentToughness())
	}
	if !slices.Contains(c.Effective().Subtypes, "Goblin") || !slices.Equal(c.EffectiveColors(), []string{"R"}) {
		t.Errorf("subtypes %v colours %v, want a red Goblin", c.Effective().Subtypes, c.EffectiveColors())
	}
	declareAttack(t, g, opp.ID, den)
	passPriorityAroundTable(t, g)
	tokens := 0
	for _, bc := range g.Battlefield.Cards {
		if bc.IsToken() && bc.Name == "Goblin" && bc.Controller == me.ID {
			tokens++
			if !bc.Tapped || bc.CurrentPower() != 1 || bc.CurrentToughness() != 1 || !slices.Equal(bc.EffectiveColors(), []string{"R"}) {
				t.Errorf("token tapped=%v %d/%d colours %v, want a tapped 1/1 red Goblin", bc.Tapped, bc.CurrentPower(), bc.CurrentToughness(), bc.EffectiveColors())
			}
			if bc.AttackingTarget == uuid.Nil {
				t.Error("the Goblin token is not attacking")
			}
		}
	}
	if tokens != 1 {
		t.Fatalf("%d Goblin tokens, want 1", tokens)
	}
	advanceToNextSeatsTurn(t, g)
	if c := s58p6View(t, g, den); c.IsCreature() {
		t.Error("Den is still a creature after the turn")
	}
}

func TestAFRB17HiveHasMenaceAndExilesFromTheDefendersGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	hive := afrPlayedLand(g, me, "Hive of the Eye Tyrant", afrHiveOracle)
	victim := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: victim, Name: "Doomed", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})
	mine := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: mine, Name: "Mine", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	advanceTo(t, g, game.StepPrecombatMain)
	afrAnimate(t, g, me, hive, game.ActivateAbilityParams{}, "B", 3)
	c := s58p6View(t, g, hive)
	if !c.IsCreature() || c.CurrentPower() != 3 || c.CurrentToughness() != 3 || !slices.Contains(c.Effective().Subtypes, "Beholder") {
		t.Fatalf("hive is %d/%d %v, want a 3/3 Beholder creature", c.CurrentPower(), c.CurrentToughness(), c.Effective().Subtypes)
	}
	if !afrHas(c, "menace") {
		t.Errorf("abilities %v, want menace", c.Effective().Abilities)
	}
	declareAttack(t, g, opp.ID, hive)
	if p := latestPickTarget(g, me.ID); p == nil || hasID(p.PickTargetCards, mine) || !hasID(p.PickTargetCards, victim) {
		t.Fatalf("pick prompt %+v, want only the defender's graveyard card offered", p)
	}
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(victim) {
		t.Error("the defender's graveyard card was not exiled")
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("the attacker's own graveyard card was exiled")
	}
}

func TestAFRB17LairBecomesAnXByXHydraAndXCantBeZero(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lair := afrPlayedLand(g, me, "Lair of the Hydra", afrLairOracle)
	advanceTo(t, g, game.StepPrecombatMain)
	fillPoolColored(me, "G", 1)
	if err := g.ActivateCatalogAbility(me.ID, lair, 0, game.ActivateAbilityParams{XValue: 0, Strict: true}); err == nil {
		t.Fatal("X = 0 was accepted")
	}
	me.ManaPool.EmptyPool()
	afrAnimate(t, g, me, lair, game.ActivateAbilityParams{XValue: 3}, "G", 3)
	c := s58p6View(t, g, lair)
	if !c.IsCreature() || !c.IsLand() || c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Fatalf("creature=%v land=%v %d/%d, want a 3/3 creature land", c.IsCreature(), c.IsLand(), c.CurrentPower(), c.CurrentToughness())
	}
	if !slices.Contains(c.Effective().Subtypes, "Hydra") || !slices.Equal(c.EffectiveColors(), []string{"G"}) {
		t.Errorf("subtypes %v colours %v, want a green Hydra", c.Effective().Subtypes, c.EffectiveColors())
	}
}

func TestAFRB17CaveBecomesAThreeFourFlyingDragon(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	cave := afrPlayedLand(g, me, "Cave of the Frost Dragon", afrCaveOracle)
	advanceTo(t, g, game.StepPrecombatMain)
	afrAnimate(t, g, me, cave, game.ActivateAbilityParams{}, "W", 4)
	c := s58p6View(t, g, cave)
	if !c.IsCreature() || !c.IsLand() || c.CurrentPower() != 3 || c.CurrentToughness() != 4 {
		t.Fatalf("creature=%v land=%v %d/%d, want a 3/4 creature land", c.IsCreature(), c.IsLand(), c.CurrentPower(), c.CurrentToughness())
	}
	if !afrHas(c, "flying") || !slices.Contains(c.Effective().Subtypes, "Dragon") || !slices.Equal(c.EffectiveColors(), []string{"W"}) {
		t.Errorf("abilities %v subtypes %v colours %v, want a white flying Dragon", c.Effective().Abilities, c.Effective().Subtypes, c.EffectiveColors())
	}
	advanceToNextSeatsTurn(t, g)
	if c := s58p6View(t, g, cave); c.IsCreature() || afrHas(c, "flying") {
		t.Error("Cave is still a flying creature after the turn")
	}
}

func TestAFRB17HallBecomesASevenSevenGiantWithWard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hall := afrPlayedLand(g, me, "Hall of Storm Giants", afrHallOracle)
	advanceTo(t, g, game.StepPrecombatMain)
	afrAnimate(t, g, me, hall, game.ActivateAbilityParams{}, "U", 5)
	c := s58p6View(t, g, hall)
	if !c.IsCreature() || !c.IsLand() || c.CurrentPower() != 7 || c.CurrentToughness() != 7 {
		t.Fatalf("creature=%v land=%v %d/%d, want a 7/7 creature land", c.IsCreature(), c.IsLand(), c.CurrentPower(), c.CurrentToughness())
	}
	if !slices.Contains(c.Effective().Subtypes, "Giant") || !slices.Equal(c.EffectiveColors(), []string{"U"}) {
		t.Errorf("subtypes %v colours %v, want a blue Giant", c.Effective().Subtypes, c.EffectiveColors())
	}
	if len(c.Effective().GrantedAbilities) != 1 || c.Effective().GrantedAbilities[0].Key != game.GrantKey(hallOfStormGiantsWard) {
		t.Errorf("granted %v, want the ward bundle", c.Effective().GrantedAbilities)
	}
	advanceToNextSeatsTurn(t, g)
	if c := s58p6View(t, g, hall); c.IsCreature() || len(c.Effective().GrantedAbilities) != 0 {
		t.Error("Hall kept its creature type or ward after the turn")
	}
}

// --- Circle of Dreams Druid ------------------------------------------------

func TestAFRB17CircleOfDreamsDruidAddsGPerCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	druid := s58p6Push(g, me.ID, "Circle of Dreams Druid", "Creature — Elf Druid", afrCircleOracle, "{G}{G}{G}", 2, 1)
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b12Creature(g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepPrecombatMain)
	me.ManaPool.EmptyPool()
	if err := g.ActivateManaAbility(me.ID, druid, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	units := me.ManaPool.Units()
	if len(units) != 2 {
		t.Fatalf("pool %v, want {G}{G} (the Druid and one other creature of mine)", units)
	}
	for _, u := range units {
		if u.Color != "G" {
			t.Errorf("pool %v, want green only", units)
		}
	}
}

// --- Blink Dog ---------------------------------------------------------------

func TestAFRB17BlinkDogPhasesItselfOutAndBackIn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dog := s58p6Push(g, me.ID, "Blink Dog", "Creature — Dog", afrBlinkDogOracle, "{2}{W}", 1, 1)
	advanceTo(t, g, game.StepPrecombatMain)
	if !afrHas(s58p6View(t, g, dog), "double strike") {
		t.Error("Blink Dog lacks double strike")
	}
	fillPool(me, 3)
	fillPoolColored(me, "W", 1)
	s58p6Activate(t, g, me.ID, dog, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if !phasedOutInCatalogGame(g, dog) {
		t.Fatal("Blink Dog did not phase out")
	}
	passTurnsTo(t, g, g.Turn.ActiveSeat)
	if phasedOutInCatalogGame(g, dog) {
		t.Error("Blink Dog did not phase in during its controller's untap step")
	}
}

func TestAFRB17BlinkDogRefusesWithoutThePool(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dog := s58p6Push(g, me.ID, "Blink Dog", "Creature — Dog", afrBlinkDogOracle, "{2}{W}", 1, 1)
	advanceTo(t, g, game.StepPrecombatMain)
	fillPool(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, dog, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("Teleport activated without {3}{W}")
	}
}

// --- Divine Smite ---------------------------------------------------------------

func TestAFRB17DivineSmitePhasesOutANonblackCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := s58p6Push(g, opp.ID, "Bear", "Creature — Bear", "", "", 2, 2)
	castCatalogSpell(t, g, "Divine Smite", "Instant", afrSmiteOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if !phasedOutInCatalogGame(g, bear) {
		t.Fatal("the nonblack creature did not phase out")
	}
}

func TestAFRB17DivineSmiteExilesABlackPermanentInstead(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: id, Name: "Black Bear", TypeLine: "Creature — Bear", Colors: []string{"B"},
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	castCatalogSpell(t, g, "Divine Smite", "Instant", afrSmiteOracle, []game.TargetRef{{Kind: game.TargetCard, ID: id}})
	passPriorityAroundTable(t, g)
	if b12ZoneOf(g, id) != game.ZoneExile {
		t.Errorf("the black creature is in %v, want exile", b12ZoneOf(g, id))
	}
}

func TestAFRB17DivineSmiteCantTargetYourOwnCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := s58p6Push(g, me.ID, "Mine", "Creature — Bear", "", "", 2, 2)
	if err := castCatalogSpellErrStrict(t, g, "Divine Smite", "Instant", afrSmiteOracle, []game.TargetRef{{Kind: game.TargetCard, ID: mine}}); err == nil {
		t.Error("Divine Smite targeted a creature its caster controls")
	}
}

// --- Dragon Turtle ----------------------------------------------------------------

func TestAFRB17DragonTurtleTapsItselfAndTheTargetAndBothStayTapped(t *testing.T) {
	g := newCatalogGame(t)
	mySeat := g.Turn.ActiveSeat
	me := g.Seats[mySeat]
	opp := g.Seats[(mySeat+1)%len(g.Seats)]
	victim := s58p6Push(g, opp.ID, "Bear", "Creature — Bear", "", "", 2, 2)
	turtle := b12PlayFromHand(t, g, "Dragon Turtle", "Creature — Dragon Turtle", afrTurtleOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if !tappedForTest(t, g, turtle) {
		t.Fatal("Dragon Turtle did not tap itself")
	}
	if !tappedForTest(t, g, victim) {
		t.Fatal("the target creature was not tapped")
	}
	oppSeat := (mySeat + 1) % len(g.Seats)
	advanceToUpkeepOf(t, g, oppSeat)
	if !tappedForTest(t, g, victim) {
		t.Error("the target untapped in its controller's next untap step")
	}
	advanceToUpkeepOf(t, g, mySeat)
	if !tappedForTest(t, g, turtle) {
		t.Error("the Turtle untapped in its controller's next untap step")
	}
	advanceToUpkeepOf(t, g, oppSeat)
	if tappedForTest(t, g, victim) {
		t.Error("the target stayed tapped past one skipped untap step")
	}
}

// --- Ray of Frost -------------------------------------------------------------------

func afrRedCreature(g *game.Game, owner uuid.UUID, colours ...string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: id, Name: "Host", TypeLine: "Creature — Bear", Colors: colours,
		Power: 2, Toughness: 2, Owner: owner, Controller: owner, Keywords: []string{"flying"}})
	return id
}

func TestAFRB17RayOfFrostOnARedCreatureTapsItAndStripsItsAbilities(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	host := afrRedCreature(g, opp.ID, "R")
	castCatalogSpell(t, g, "Ray of Frost", auraTypeLine, afrRayOracle, []game.TargetRef{{Kind: game.TargetCard, ID: host}})
	passPriorityAroundTable(t, g)
	if !tappedForTest(t, g, host) {
		t.Error("Ray of Frost did not tap a red creature")
	}
	if afrHas(s58p6View(t, g, host), "flying") {
		t.Error("a red enchanted creature kept its abilities")
	}
}

func TestAFRB17RayOfFrostOnANonRedCreatureOnlyKeepsItFromUntapping(t *testing.T) {
	g := newCatalogGame(t)
	mySeat := g.Turn.ActiveSeat
	opp := g.Seats[(mySeat+1)%len(g.Seats)]
	host := afrRedCreature(g, opp.ID, "G")
	castCatalogSpell(t, g, "Ray of Frost", auraTypeLine, afrRayOracle, []game.TargetRef{{Kind: game.TargetCard, ID: host}})
	passPriorityAroundTable(t, g)
	if tappedForTest(t, g, host) {
		t.Error("Ray of Frost tapped a green creature")
	}
	if !afrHas(s58p6View(t, g, host), "flying") {
		t.Error("a green enchanted creature lost its abilities")
	}
	if err := g.TapCard(host, true); err != nil {
		t.Fatal(err)
	}
	advanceToUpkeepOf(t, g, (mySeat+1)%len(g.Seats))
	if !tappedForTest(t, g, host) {
		t.Error("the enchanted creature untapped during its controller's untap step")
	}
}

// afrHas reads a keyword through the layered view.
func afrHas(c game.Card, kw string) bool { return game.HasKeyword(&c, kw) }
