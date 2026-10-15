package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// speed_slice_a_test.go — the 14 speed cards of #2711 slice speed-a, on
// ADR 0138's seam. Each test sets the controller's speed directly
// (SetSpeedForTest) and checks the card's ability is off below 4 where it
// says "Max speed —", and does its printed job where it does not.

const (
	howlsquadHeavyOracle       = "cd45a29f-73bc-435c-9d50-09a4ecaa77be"
	mendicantCoreOracle        = "b20d74a4-0e08-4779-972a-499f1a2be15a"
	hazoretGodseekerOracle     = "6c5d8599-75ca-4479-a02b-1b257b7aea57"
	momentumBreakerOracle      = "087d2e42-c69f-4cd2-b2e3-e0b64ac2893b"
	nestingBotOracle           = "7ffb8c2e-1737-4df3-b310-384a5000bfa5"
	kickoffCelebrationsOracle  = "94868faf-9e88-4221-8db6-3ad10aa02da5"
	zahurGlorysPastOracle      = "fb82cd81-0e71-43ab-88f9-b1fa158f5138"
	gasGuzzlerOracle           = "7c98626b-ba97-4d1e-9be7-291480ae8d22"
	endriderCatalyzerOracle    = "b4d970ed-f332-440c-8c2f-8a8a0191410b"
	farFortuneEndBossOracle    = "8b9734c1-7185-40f6-8582-9f87109e3a09"
	samutDrivingForceOracle    = "ee89c719-5ae8-4316-bfd0-cfe3684b5859"
	pointTheWayOracle          = "0ffe9c8b-e0cb-4a29-bc14-649dbec5d1f3"
	hourOfVictoryOracle        = "eb100b3b-bc3f-48f3-8f28-bb2d80225ea9"
	gastalThrillseekerOracle   = "d185ad9f-c7fc-41f0-b852-4fe2cde0a8ae"
	speedSliceAOpponentSeatIdx = 1
)

func speedSliceARestricted(t *testing.T, g *game.Game, id uuid.UUID, r game.Restriction) bool {
	t.Helper()
	var restricted, found bool
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				restricted, found = game.Restricted(&g.Battlefield.Cards[i], r), true
			}
		}
	})
	if !found {
		t.Fatalf("card %s is not on the battlefield", id)
	}
	return restricted
}

func speedSliceAOpponent(g *game.Game) *game.Player {
	return g.Seats[(g.Turn.ActiveSeat+speedSliceAOpponentSeatIdx)%len(g.Seats)]
}

func TestSpeedSliceACardsDeclareStartYourEngines(t *testing.T) {
	for _, oracle := range []string{
		howlsquadHeavyOracle, mendicantCoreOracle, hazoretGodseekerOracle, momentumBreakerOracle,
		nestingBotOracle, kickoffCelebrationsOracle, zahurGlorysPastOracle, gasGuzzlerOracle,
		endriderCatalyzerOracle, farFortuneEndBossOracle, samutDrivingForceOracle, pointTheWayOracle,
		hourOfVictoryOracle, gastalThrillseekerOracle,
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		found := false
		for _, kw := range spec.PrintedKeywords {
			if kw == game.KeywordStartYourEngines {
				found = true
			}
		}
		if !found {
			t.Errorf("%s does not declare start your engines!", spec.Name)
		}
	}
}

func TestGastalThrillseekerPingsAnOpponentAndGainsLife(t *testing.T) {
	g, me := speedGame(t)
	opp := speedSliceAOpponent(g)
	lifeMe, lifeOpp := me.Life, opp.Life
	id := castCatalogSpell(t, g, "Gastal Thrillseeker", "Creature — Lizard Berserker", gastalThrillseekerOracle, nil)
	passPriorityAroundTable(t, g)
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if got := lifeOpp - opp.Life; got != 1 {
		t.Errorf("opponent lost %d, want 1", got)
	}
	if got := me.Life - lifeMe; got != 1 {
		t.Errorf("you gained %d, want 1", got)
	}
	// The ping made an opponent lose life on your turn: speed 1 from the
	// keyword, then 2 from CR 702.179d.
	if me.Speed != 2 {
		t.Errorf("speed = %d, want 2", me.Speed)
	}
	abilities := effectiveAbilities(t, g, id)
	if eotHasAbility(abilities, "deathtouch") || eotHasAbility(abilities, "haste") {
		t.Error("deathtouch or haste below max speed")
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	abilities = effectiveAbilities(t, g, id)
	if !eotHasAbility(abilities, "deathtouch") || !eotHasAbility(abilities, "haste") {
		t.Errorf("no deathtouch and haste at max speed: %v", abilities)
	}
}

func TestNestingBotLeavesAServoAndIsBiggerAtMaxSpeed(t *testing.T) {
	g, me := speedGame(t)
	id := pushSpeedCard(t, g, me, "Nesting Bot", "Artifact Creature — Robot", nestingBotOracle, 1, 1)
	if p := effectivePower(t, g, id); p != 1 {
		t.Errorf("power below max speed = %d, want 1", p)
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	if p := effectivePower(t, g, id); p != 2 {
		t.Errorf("power at max speed = %d, want 2", p)
	}
	if tt := effectiveToughness(t, g, id); tt != 1 {
		t.Errorf("toughness at max speed = %d, want 1", tt)
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if got := countOnBattlefield(g, "Servo", me.ID); got != 1 {
		t.Errorf("Servo tokens = %d, want 1", got)
	}
}

func TestEndriderCatalyzerAddsRedRedOnlyAtMaxSpeed(t *testing.T) {
	g, me := speedGame(t)
	id := pushSpeedCard(t, g, me, "Endrider Catalyzer", "Creature — Human Warrior", endriderCatalyzerOracle, 3, 1)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("below max speed: err = %v, want ErrConditionNotMet", err)
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	activateManaFor(t, g, me.ID, id, 0, game.ManaAbilityParams{})
	if got := me.ManaPool.Total(); got != 2 {
		t.Errorf("pool holds %d mana, want 2", got)
	}
	for _, m := range me.ManaPool.Units() {
		if m.Color != "R" {
			t.Errorf("mana colour %q, want R", m.Color)
		}
	}
}

func TestKickoffCelebrationsRummagesOnEntryAndGrantsHasteAtMaxSpeed(t *testing.T) {
	g, me := speedGame(t)
	castCatalogSpell(t, g, "Kickoff Celebrations", "Enchantment", kickoffCelebrationsOracle, nil)
	passPriorityAroundTable(t, g)
	// Declining the discard draws nothing.
	hand := me.Hand.Size()
	if c := discardChoiceFor(g, me.ID); c != nil {
		answerDiscard(t, g, me.ID)
	}
	if me.Hand.Size() != hand {
		t.Errorf("declined: hand %d, want %d", me.Hand.Size(), hand)
	}
	col := pushCatalogPermanent(g, me.ID, "Kickoff Celebrations", "Enchantment", kickoffCelebrationsOracle, false)
	bear := pushSpeedCard(t, g, me, "Bear", "Creature — Bear", "", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, col, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("below max speed: err = %v, want ErrConditionNotMet", err)
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	if err := g.ActivateCatalogAbility(me.ID, col, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("at max speed: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(col) {
		t.Error("the enchantment was not sacrificed")
	}
	if !eotHasAbility(effectiveAbilities(t, g, bear), "haste") {
		t.Error("the creature did not gain haste")
	}
}

func TestKickoffCelebrationsDiscardsToDrawTwo(t *testing.T) {
	g, me := speedGame(t)
	castCatalogSpell(t, g, "Kickoff Celebrations", "Enchantment", kickoffCelebrationsOracle, nil)
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	pitch := me.Hand.Cards[0].InstanceID
	answerDiscard(t, g, me.ID, pitch)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand = %d, want %d (one discarded, two drawn)", got, hand+1)
	}
}

func TestPointTheWayFetchesSpeedBasicsTapped(t *testing.T) {
	g, me := speedGame(t)
	way := pushSpeedCard(t, g, me, "Point the Way", "Enchantment", pointTheWayOracle, 0, 0)
	var forests []uuid.UUID
	for i := 0; i < 4; i++ {
		forests = append(forests, pushLibraryCardForTest(me, game.Card{
			InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest",
			Owner: me.ID, Controller: me.ID,
		}))
	}
	g.SetSpeedForTest(me.ID, 3)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{G}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, way, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := searchChoiceFor(g, me.ID); c == nil || c.SearchMax != 3 {
		t.Fatalf("search prompt max = %v, want 3 (speed 3)", c != nil)
	}
	answerSearchByID(t, g, me.ID, forests[:3]...)
	passPriorityAroundTable(t, g)
	lands := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Forest" && c.Controller == me.ID {
			lands++
			if !c.Tapped {
				t.Error("a fetched land entered untapped")
			}
		}
	}
	if lands != 3 {
		t.Errorf("fetched %d lands, want 3 (speed 3)", lands)
	}
	if g.Battlefield.Contains(way) {
		t.Error("Point the Way was not sacrificed")
	}
}

func TestHourOfVictoryTutorNeedsMaxSpeedAndSorcerySpeed(t *testing.T) {
	g, me := speedGame(t)
	castCatalogSpell(t, g, "Hour of Victory", "Enchantment", hourOfVictoryOracle, nil)
	passPriorityAroundTable(t, g)
	if got := countOnBattlefield(g, "Zombie", me.ID); got != 1 {
		t.Fatalf("Zombie tokens = %d, want 1", got)
	}
	var hour uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == hourOfVictoryOracle {
			hour = c.InstanceID
		}
	}
	needle := pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Needle", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{B}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, hour, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("below max speed: err = %v, want ErrConditionNotMet", err)
	}
	if !g.Battlefield.Contains(hour) {
		t.Fatal("a refused activation paid its sacrifice")
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	if err := g.ActivateCatalogAbility(me.ID, hour, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("at max speed: %v", err)
	}
	if g.Battlefield.Contains(hour) {
		t.Error("Hour of Victory was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, needle)
	if !me.Hand.Contains(needle) {
		t.Error("the searched card is not in hand")
	}
}

func TestMomentumBreakerEdictsAndDiscardsForTheOpponentWhoCant(t *testing.T) {
	g, me := speedGame(t)
	a := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	c := g.Seats[(g.Turn.ActiveSeat+3)%len(g.Seats)]
	aBear := pushCatalogPermanent(g, a.ID, "A Bear", "Creature — Bear", "", false)
	// b has only a Vehicle that is not a creature; c has nothing.
	bVehicle := pushCatalogPermanent(g, b.ID, "B Cart", "Artifact — Vehicle", "", false)
	handB, handC := b.Hand.Size(), c.Hand.Size()
	castCatalogSpell(t, g, "Momentum Breaker", "Enchantment", momentumBreakerOracle, nil)
	passPriorityAroundTable(t, g)
	answerSacrifice(t, g, a.ID, aBear)
	answerSacrifice(t, g, b.ID, bVehicle)
	if g.Battlefield.Contains(aBear) || g.Battlefield.Contains(bVehicle) {
		t.Error("an opponent kept the permanent they should have sacrificed")
	}
	discardFromHand(t, g, c.ID)
	if got := c.Hand.Size(); got != handC-1 {
		t.Errorf("the opponent who couldn't sacrifice: hand %d, want %d", got, handC-1)
	}
	if got := b.Hand.Size(); got != handB {
		t.Errorf("an opponent who sacrificed discarded: hand %d, want %d", got, handB)
	}
	if me.Speed != 1 {
		t.Errorf("speed = %d, want 1", me.Speed)
	}
}

func TestMomentumBreakerSacrificeGainsLifeEqualToSpeed(t *testing.T) {
	g, me := speedGame(t)
	mb := pushSpeedCard(t, g, me, "Momentum Breaker", "Enchantment", momentumBreakerOracle, 0, 0)
	g.SetSpeedForTest(me.ID, 3)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, mb, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Life - life; got != 3 {
		t.Errorf("gained %d, want 3 (speed 3)", got)
	}
}

func TestGasGuzzlerEntersTappedAndDrawsOnlyAtMaxSpeed(t *testing.T) {
	g, me := speedGame(t)
	castCatalogSpell(t, g, "Gas Guzzler", "Creature — Vampire Rogue", gasGuzzlerOracle, nil)
	passPriorityAroundTable(t, g)
	var guzzler uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == gasGuzzlerOracle {
			guzzler = c.InstanceID
			if !c.Tapped {
				t.Error("Gas Guzzler entered untapped")
			}
		}
	}
	fodder := pushCatalogPermanent(g, me.ID, "Fodder", "Creature — Bear", "", false)
	params := game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, guzzler, 0, params); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("below max speed: err = %v, want ErrConditionNotMet", err)
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	if err := g.ActivateCatalogAbility(me.ID, guzzler, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{guzzler}}); err == nil {
		t.Fatal("Gas Guzzler sacrificed itself to its own ability")
	}
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, guzzler, 0, params); err != nil {
		t.Fatalf("at max speed: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the fodder was not sacrificed")
	}
}

func TestZahurSurveilsOncePerTurnAndMakesZombiesAtMaxSpeed(t *testing.T) {
	g, me := speedGame(t)
	zahur := pushSpeedCard(t, g, me, "Zahur, Glory's Past", "Legendary Creature — Zombie Cat Warrior", zahurGlorysPastOracle, 3, 2)
	a := pushCatalogPermanent(g, me.ID, "Fodder A", "Creature — Bear", "", false)
	b := pushCatalogPermanent(g, me.ID, "Fodder B", "Creature — Bear", "", false)
	pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Top", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
	if err := g.ActivateCatalogAbility(me.ID, zahur, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{a}}); err != nil {
		t.Fatalf("first surveil: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, zahur, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{b}}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("second surveil: err = %v, want ErrConditionNotMet", err)
	}
	passPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceSurveil {
			if err := g.ResolveSurveil(c.ID, me.ID, nil, c.ScryCards); err != nil {
				t.Fatalf("ResolveSurveil: %v", err)
			}
			break
		}
	}
	// Below max speed a dying nontoken creature makes nothing.
	zombies := countOnBattlefield(g, "Zombie", me.ID)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(b) })
	passPriorityAroundTable(t, g)
	if got := countOnBattlefield(g, "Zombie", me.ID) - zombies; got != 0 {
		t.Errorf("below max speed made %d Zombies", got)
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	token := pushCatalogPermanent(g, me.ID, "Zombie", "Token Creature — Zombie", "", false)
	zombies = countOnBattlefield(g, "Zombie", me.ID)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(token) })
	passPriorityAroundTable(t, g)
	if got := countOnBattlefield(g, "Zombie", me.ID) - zombies + 1; got != 0 {
		t.Errorf("a dying token made a Zombie (net %d)", got)
	}
	nontoken := pushCatalogPermanent(g, me.ID, "Nontoken Bear", "Creature — Bear", "", false)
	seen := map[uuid.UUID]bool{}
	for _, c := range g.Battlefield.Cards {
		seen[c.InstanceID] = true
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(nontoken) })
	passPriorityAroundTable(t, g)
	made := 0
	for _, c := range g.Battlefield.Cards {
		if seen[c.InstanceID] || c.Name != "Zombie" || c.Controller != me.ID {
			continue
		}
		made++
		if !c.Tapped {
			t.Error("the new Zombie entered untapped")
		}
		if c.Power != 2 || c.Toughness != 2 {
			t.Errorf("the new Zombie is %d/%d, want 2/2", c.Power, c.Toughness)
		}
	}
	if made != 1 {
		t.Errorf("at max speed a nontoken creature dying made %d Zombies, want 1", made)
	}
}

func TestHazoretCantAttackOrBlockBelowMaxSpeed(t *testing.T) {
	g, me := speedGame(t)
	opp := speedSliceAOpponent(g)
	hz := pushSpeedCard(t, g, me, "Hazoret, Godseeker", "Legendary Creature — God", hazoretGodseekerOracle, 5, 3)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(hz, opp.ID); !errors.Is(err, game.ErrCantAttack) {
		t.Fatalf("below max speed: err = %v, want ErrCantAttack", err)
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	if err := g.DeclareAttacker(hz, opp.ID); err != nil {
		t.Fatalf("at max speed: %v", err)
	}
}

func TestHazoretCantBlockBelowMaxSpeed(t *testing.T) {
	g, me := speedGame(t)
	hz := pushSpeedCard(t, g, me, "Hazoret, Godseeker", "Legendary Creature — God", hazoretGodseekerOracle, 5, 3)
	if !speedSliceARestricted(t, g, hz, game.CantBlock) {
		t.Error("Hazoret can block below max speed")
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	if speedSliceARestricted(t, g, hz, game.CantBlock) {
		t.Error("Hazoret still can't block at max speed")
	}
}

func TestHazoretMakesASmallCreatureUnblockable(t *testing.T) {
	g, me := speedGame(t)
	hz := pushSpeedCard(t, g, me, "Hazoret, Godseeker", "Legendary Creature — God", hazoretGodseekerOracle, 5, 3)
	small := pushSpeedCard(t, g, me, "Small Bear", "Creature — Bear", "", 2, 2)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, hz, 0, game.ActivateAbilityParams{Targets: cardRefs(small)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !speedSliceARestricted(t, g, small, game.CantBeBlocked) {
		t.Error("the target can still be blocked")
	}
}

func TestHazoretCantTargetABigCreature(t *testing.T) {
	g, me := speedGame(t)
	hz := pushSpeedCard(t, g, me, "Hazoret, Godseeker", "Legendary Creature — God", hazoretGodseekerOracle, 5, 3)
	big := pushSpeedCard(t, g, me, "Big Bear", "Creature — Bear", "", 3, 3)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, hz, 0, game.ActivateAbilityParams{Targets: cardRefs(big)}); err == nil {
		t.Error("a power 3 creature was a legal target")
	}
}

func TestHowlsquadHeavyMakesAHastyGoblinThatMustAttack(t *testing.T) {
	g, me := speedGame(t)
	heavy := pushSpeedCard(t, g, me, "Howlsquad Heavy", "Creature — Goblin Mercenary", howlsquadHeavyOracle, 2, 3)
	_ = heavy
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	var token uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Goblin" && c.Controller == me.ID {
			token = c.InstanceID
		}
	}
	if token == uuid.Nil {
		t.Fatal("no Goblin token at the beginning of combat")
	}
	if !eotHasAbility(effectiveAbilities(t, g, token), "haste") {
		t.Error("the Goblin token has no haste from Howlsquad Heavy")
	}
	if !hasRequirementRecord(g) {
		t.Error("the token has no attack requirement")
	}
	if eotHasAbility(effectiveAbilities(t, g, heavy), "haste") {
		t.Error("Howlsquad Heavy gave itself haste ('other Goblins')")
	}
}

func TestHowlsquadHeavyManaAtMaxSpeedCountsGoblins(t *testing.T) {
	g, me := speedGame(t)
	heavy := pushSpeedCard(t, g, me, "Howlsquad Heavy", "Creature — Goblin Mercenary", howlsquadHeavyOracle, 2, 3)
	pushSpeedCard(t, g, me, "Goblin A", "Creature — Goblin", "", 1, 1)
	pushSpeedCard(t, g, me, "Goblin B", "Creature — Goblin", "", 1, 1)
	pushSpeedCard(t, g, me, "Bear", "Creature — Bear", "", 2, 2)
	if err := g.ActivateManaAbility(me.ID, heavy, 0, game.ManaAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("below max speed: err = %v, want ErrConditionNotMet", err)
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	activateManaFor(t, g, me.ID, heavy, 0, game.ManaAbilityParams{})
	if got := me.ManaPool.Total(); got != 3 {
		t.Errorf("pool holds %d mana, want 3 (three Goblins)", got)
	}
}

func TestFarFortunePingsOnAttackAndAddsOneAtMaxSpeed(t *testing.T) {
	g, me := speedGame(t)
	opp := speedSliceAOpponent(g)
	ff := pushSpeedCard(t, g, me, "Far Fortune, End Boss", "Legendary Creature — Human Mercenary", farFortuneEndBossOracle, 4, 5)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	_ = ff
	life := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		life[s.ID] = s.Life
	}
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		if s.ID == me.ID {
			continue
		}
		if got := life[s.ID] - s.Life; got != 1 {
			t.Errorf("%s lost %d to the attack trigger below max speed, want 1", s.Name, got)
		}
	}
	// At max speed every damage source you control deals one more to an
	// opponent, Far Fortune's own trigger included.
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	before := opp.Life
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(ff, opp.ID, 2); err != nil {
			t.Fatalf("DealDamageForEffect: %v", err)
		}
	})
	if got := before - opp.Life; got != 3 {
		t.Errorf("2 damage at max speed dealt %d, want 3", got)
	}
	own := me.Life
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(ff, me.ID, 2); err != nil {
			t.Fatalf("DealDamageForEffect: %v", err)
		}
	})
	if got := own - me.Life; got != 2 {
		t.Errorf("2 damage to yourself dealt %d, want 2", got)
	}
}

func TestSamutPumpsOthersAndDiscountsNoncreatureSpellsBySpeed(t *testing.T) {
	g, me := speedGame(t)
	samut := pushSpeedCard(t, g, me, "Samut, the Driving Force", "Legendary Creature — Human Warrior Cleric", samutDrivingForceOracle, 4, 5)
	bear := pushSpeedCard(t, g, me, "Bear", "Creature — Bear", "", 2, 2)
	opp := speedSliceAOpponent(g)
	theirs := pushCatalogPermanent(g, opp.ID, "Their Bear", "Creature — Bear", "", false)
	if me.Speed != 1 {
		t.Fatalf("speed = %d, want 1", me.Speed)
	}
	if p := effectivePower(t, g, bear); p != 3 {
		t.Errorf("Bear power at speed 1 = %d, want 3", p)
	}
	if p := effectivePower(t, g, samut); p != 4 {
		t.Errorf("Samut pumped itself: power %d, want 4", p)
	}
	g.SetSpeedForTest(me.ID, 3)
	if p := effectivePower(t, g, bear); p != 5 {
		t.Errorf("Bear power at speed 3 = %d, want 5", p)
	}
	if p := effectivePower(t, g, theirs); p != 1 {
		t.Errorf("an opponent's creature was pumped: power %d", p)
	}
	// {2}{R} less three generic: the reduction stops at the generic part.
	if got := priceInHand(t, g, me, "Shock", "Instant", "{2}{R}"); got != 1 {
		t.Errorf("noncreature spell at speed 3: %d, want 1", got)
	}
	g.SetSpeedForTest(me.ID, 1)
	if got := priceInHand(t, g, me, "Shock", "Instant", "{2}{R}"); got != 2 {
		t.Errorf("noncreature spell at speed 1: %d, want 2", got)
	}
	if got := priceInHand(t, g, me, "Grizzly Bears", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("creature spell: %d, want 2 (no discount)", got)
	}
	if got := priceInHand(t, g, opp, "Shock", "Instant", "{2}{R}"); got != 3 {
		t.Errorf("an opponent's spell: %d, want 3", got)
	}
}

func TestMendicantCoreCountsArtifactsAndCopiesArtifactSpellsAtMaxSpeed(t *testing.T) {
	g, me := speedGame(t)
	core := pushSpeedCard(t, g, me, "Mendicant Core, Guidelight", "Legendary Artifact Creature — Robot", mendicantCoreOracle, 0, 3)
	if p := effectivePower(t, g, core); p != 1 {
		t.Errorf("power with just itself = %d, want 1", p)
	}
	pushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	if p := effectivePower(t, g, core); p != 2 {
		t.Errorf("power with two artifacts = %d, want 2", p)
	}
	// Below max speed an artifact spell is not copied.
	castCatalogSpell(t, g, "Plain Trinket", "Artifact", "", nil)
	if hasPayUnlessFor(g, me.ID) {
		t.Fatal("asked to copy below max speed")
	}
	passPriorityAroundTable(t, g)
	trinkets := countOnBattlefield(g, "Plain Trinket", me.ID)
	if trinkets != 1 {
		t.Fatalf("Plain Trinket permanents = %d, want 1", trinkets)
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	castCatalogSpell(t, g, "Second Trinket", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("not asked to pay {1} at max speed")
	}
	b06AddMana(me, "C")
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := countOnBattlefield(g, "Second Trinket", me.ID); got != 2 {
		t.Errorf("Second Trinket permanents = %d, want 2 (the spell and its token copy)", got)
	}
}

func TestMendicantCoreDeclinedCopyMakesNothing(t *testing.T) {
	g, me := speedGame(t)
	pushSpeedCard(t, g, me, "Mendicant Core, Guidelight", "Legendary Artifact Creature — Robot", mendicantCoreOracle, 0, 3)
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	castCatalogSpell(t, g, "Trinket", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if got := countOnBattlefield(g, "Trinket", me.ID); got != 1 {
		t.Errorf("Trinket permanents = %d, want 1", got)
	}
}
