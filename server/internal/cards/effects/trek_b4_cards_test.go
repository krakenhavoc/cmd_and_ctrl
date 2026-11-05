package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b4_cards_test.go — Star Trek, batch 4 (#4973). Each test plays the
// real catalog entry and asserts what its text says.

const (
	trekB4PureEnergyOracle = "f4831eee-fef6-4fef-a174-400a25d0837d"
	trekB4RutherfordOracle = "42258216-fe2b-42d0-a6ce-76ecf8104c10"
	trekB4SaavikOracle     = "53aaa919-d08f-4b76-8583-bc2d4635f991"
	trekB4PhasersOracle    = "4da9b548-898d-431b-8aa5-9cb7bf579219"
	trekB4ShieldsUpOracle  = "0e14940b-87e1-4139-8f30-3629b3ad3a59"
	trekB4ShuttleAceOracle = "4b5b0864-7ddd-497d-a7d6-a31c49bb9c7b"
	trekB4SuluOracle       = "d090131d-428c-4a98-afc5-595b9bc7adce"
	trekB4TacticalOracle   = "1397e7d5-4aa4-4c8d-8957-845579b775f7"
	trekB4ConvictionOracle = "beffaea5-beeb-4d4e-aca3-82c698caf6f8"
	trekB4CoffeeOracle     = "887614f9-f12f-47bc-ae56-1438cd7cad10"
)

func TestTrekB4EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB4PureEnergyOracle: "Pure Energy",
		trekB4RutherfordOracle: "Rutherford, Upgraded Ensign",
		trekB4SaavikOracle:     "Saavik, Stoic Student",
		trekB4PhasersOracle:    "Set Phasers to . . .",
		trekB4ShieldsUpOracle:  "Shields Up!",
		trekB4ShuttleAceOracle: "Shuttle Ace",
		trekB4SuluOracle:       "Sulu, Space Swashbuckler",
		trekB4TacticalOracle:   "Tactical Officer",
		trekB4ConvictionOracle: "Test of Conviction",
		trekB4CoffeeOracle:     "There's Coffee in that Nebula",
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

// --- fixtures ------------------------------------------------------------------

// trekB4ToMain walks the cursor to the active player's precombat main.
func trekB4ToMain(t *testing.T, g *game.Game) {
	t.Helper()
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}

// trekB4Hand puts a card in p's hand and returns its id.
func trekB4Hand(p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	p.Hand.PushTop(c)
	return c.InstanceID
}

// trekB4Cast casts a card from p's hand, failing the test on a refusal.
func trekB4Cast(t *testing.T, g *game.Game, p *game.Player, c game.Card, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	id := trekB4Hand(p, c)
	if err := g.CastSpell(p.ID, id, params); err != nil {
		t.Fatalf("casting %s: %v", c.Name, err)
	}
	return id
}

// trekB4Creature seeds a creature on the battlefield, catalog row or not.
func trekB4Creature(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// trekB4Top stacks cards on top of p's library, the first one on top.
func trekB4Top(p *game.Player, cards ...game.Card) []uuid.UUID {
	ids := make([]uuid.UUID, len(cards))
	for i := len(cards) - 1; i >= 0; i-- {
		c := cards[i]
		c.InstanceID = uuid.New()
		c.Owner, c.Controller = p.ID, p.ID
		p.Library.PushTop(c)
		ids[i] = c.InstanceID
	}
	return ids
}

func trekB4Plains(g *game.Game, owner uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		pushLandFor(g, owner, "Plains", "Basic Land — Plains")
	}
}

func trekB4Offered(p *game.PendingChoice) map[uuid.UUID]bool {
	out := map[uuid.UUID]bool{}
	for _, id := range p.ChooseCards {
		out[id] = true
	}
	return out
}

// trekB4Attack declares the given creatures as attackers against defender
// and settles the triggers the declaration makes.
func trekB4Attack(t *testing.T, g *game.Game, defender uuid.UUID, attackers ...uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range attackers {
		if err := g.DeclareAttacker(id, defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
}

// --- Pure Energy -----------------------------------------------------------------

func TestTrekB4PureEnergyDistributesTwoCountersFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	trekB4ToMain(t, g)
	a := trekB4Creature(g, me.ID, "A", "Creature — Bear", "", 2, 2)
	b := trekB4Creature(g, opp.ID, "B", "Creature — Bear", "", 2, 2)
	trekB4Plains(g, me.ID, 2)
	trekB4Cast(t, g, me, game.Card{Name: "Pure Energy", TypeLine: "Sorcery", ManaCost: "{1}{W}", OracleID: trekB4PureEnergyOracle},
		game.CastSpellParams{Targets: cardRefs(a, b), Distribution: map[uuid.UUID]int{a: 1, b: 1}, Strict: true, AutoTap: true})
	passPriorityAroundTable(t, g)
	if counterCount(g, a, game.CounterPlusOne) != 1 || counterCount(g, b, game.CounterPlusOne) != 1 {
		t.Errorf("counters %d and %d, want 1 and 1", counterCount(g, a, game.CounterPlusOne), counterCount(g, b, game.CounterPlusOne))
	}
}

func TestTrekB4PureEnergyBeamsUpFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	energy := seedGraveyardCard(t, g, "Pure Energy", "Sorcery", trekB4PureEnergyOracle)
	target := trekB4Creature(g, me.ID, "Target", "Creature — Bear", "", 2, 2)
	crew := trekB4Creature(g, me.ID, "Crew", "Creature — Human", "", 1, 1)
	trekB4Plains(g, me.ID, 3)
	if err := g.CastSpell(me.ID, energy, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: trekB4AltCostKeyBeamMeUp, AltCostIDs: []uuid.UUID{crew},
		Targets: cardRefs(target), Distribution: map[uuid.UUID]int{target: 2}, Strict: true, AutoTap: true,
	}); err != nil {
		t.Fatalf("beam me up: %v", err)
	}
	if !me.Hand.Contains(crew) {
		t.Error("the returned creature is not in its owner's hand with the spell on the stack")
	}
	passPriorityAroundTable(t, g)
	if n := counterCount(g, target, game.CounterPlusOne); n != 2 {
		t.Errorf("%d counters on the single target, want 2", n)
	}
	if !g.Exile.Contains(energy) || me.Graveyard.Contains(energy) {
		t.Error("a beamed-up Pure Energy was not exiled as it left the stack")
	}
}

func TestTrekB4BeamMeUpNeedsACreatureToReturn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	energy := seedGraveyardCard(t, g, "Pure Energy", "Sorcery", trekB4PureEnergyOracle)
	theirs := trekB4Creature(g, opp.ID, "Theirs", "Creature — Bear", "", 2, 2)
	trekB4Plains(g, me.ID, 3)
	params := game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: trekB4AltCostKeyBeamMeUp,
		Targets: cardRefs(theirs), Distribution: map[uuid.UUID]int{theirs: 2}, Strict: true, AutoTap: true,
	}
	if err := g.CastSpell(me.ID, energy, params); err == nil {
		t.Fatal("beam me up was claimed without returning a creature")
	}
	params.AltCostIDs = []uuid.UUID{theirs}
	if err := g.CastSpell(me.ID, energy, params); err == nil {
		t.Fatal("beam me up returned a creature the caster does not control")
	}
	if !me.Graveyard.Contains(energy) {
		t.Error("a refused cast moved Pure Energy out of the graveyard")
	}
}

// --- Rutherford, Upgraded Ensign --------------------------------------------------------

func TestTrekB4RutherfordTakesAnArtifactFromTheTopSix(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	trekB4ToMain(t, g)
	top := trekB4Top(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Sol Ring", TypeLine: "Artifact"},
		game.Card{Name: "Shock", TypeLine: "Instant"},
		game.Card{Name: "Shuttle", TypeLine: "Artifact — Spacecraft"},
		game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"},
		game.Card{Name: "Myr", TypeLine: "Artifact Creature — Myr"},
		game.Card{Name: "Seventh", TypeLine: "Artifact"},
	)
	trekB4Cast(t, g, me, game.Card{Name: "Rutherford, Upgraded Ensign", TypeLine: "Legendary Creature — Human Artificer",
		ManaCost: "{2}{W}", Power: 2, Toughness: 2, OracleID: trekB4RutherfordOracle}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	prompt := latestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("no Rutherford look prompt")
	}
	offered := trekB4Offered(prompt)
	want := []uuid.UUID{top[1], top[3], top[5]}
	if len(offered) != len(want) {
		t.Fatalf("offered %d cards, want the 3 artifact cards of the top six", len(offered))
	}
	for _, id := range want {
		if !offered[id] {
			t.Error("an artifact card among the six was not offered")
		}
	}
	if prompt.ChooseMin != 0 || prompt.ChooseMax != 1 {
		t.Errorf("bounds [%d,%d], want [0,1]", prompt.ChooseMin, prompt.ChooseMax)
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{top[3]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(top[3]) {
		t.Error("the chosen artifact is not in hand")
	}
	if c, err := me.Library.Top(); err != nil || c.InstanceID != top[6] {
		t.Error("the seventh card is not on top: the rest did not go to the bottom")
	}
}

func TestTrekB4RutherfordCountersOnASpacecraftSpellOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	trekB4ToMain(t, g)
	trekB4Creature(g, me.ID, "Rutherford, Upgraded Ensign", "Legendary Creature — Human Artificer", trekB4RutherfordOracle, 2, 2)
	bear := trekB4Creature(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)

	trekB4Cast(t, g, me, game.Card{Name: "Plain Relic", TypeLine: "Artifact"}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Fatal("a non-Spacecraft artifact spell asked for a target")
	}

	trekB4Cast(t, g, me, game.Card{Name: "Shuttlecraft", TypeLine: "Artifact — Spacecraft"}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		pickCard(t, g, me.ID, bear)
	}
	passPriorityAroundTable(t, g)
	if n := counterCount(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters after a Spacecraft spell, want 1", n)
	}
}

// --- Saavik, Stoic Student -------------------------------------------------------

func TestTrekB4SaavikCountsCreatureTypesAmongNonBorg(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	trekB4ToMain(t, g)
	trekB4Creature(g, me.ID, "Soldier", "Creature — Human Soldier", "", 1, 1)
	trekB4Creature(g, me.ID, "Ranger", "Creature — Elf Soldier", "", 1, 1) // Soldier counts once
	trekB4Creature(g, me.ID, "Drone", "Artifact Creature — Borg Drone", "", 1, 1)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Mimic", TypeLine: "Creature — Shapeshifter", Keywords: []string{"changeling"},
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	victim := trekB4Creature(g, opp.ID, "Attacker", "Creature — Ogre", "", 9, 9)
	bystander := trekB4Creature(g, opp.ID, "Bystander", "Creature — Ogre", "", 9, 9)
	helicarrierAttacker(g, victim, me.ID)

	trekB4Cast(t, g, me, game.Card{Name: "Saavik, Stoic Student", TypeLine: "Legendary Creature — Vulcan Advisor",
		ManaCost: "{4}{W}", Power: 2, Toughness: 2, OracleID: trekB4SaavikOracle}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		for _, id := range p.PickTargetCards {
			if id == bystander {
				t.Error("a creature that is neither attacking nor blocking was offered")
			}
		}
		pickCard(t, g, me.ID, victim)
	}
	passPriorityAroundTable(t, g)
	// Human, Soldier, Elf, Vulcan, Advisor. The changeling is a Borg, and
	// the Borg Drone is excluded, so neither adds a type.
	if got := findBattlefieldCardForTest(g, victim).DamageMarked; got != 5 {
		t.Errorf("damage %d, want 5", got)
	}
	if findBattlefieldCardForTest(g, bystander).DamageMarked != 0 {
		t.Error("the bystander was damaged")
	}
}

func TestTrekB4SaavikHasNoTargetOutsideCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	trekB4ToMain(t, g)
	idle := trekB4Creature(g, opp.ID, "Idle", "Creature — Ogre", "", 9, 9)
	trekB4Cast(t, g, me, game.Card{Name: "Saavik, Stoic Student", TypeLine: "Legendary Creature — Vulcan Advisor",
		ManaCost: "{4}{W}", Power: 2, Toughness: 2, OracleID: trekB4SaavikOracle}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Fatal("Saavik asked for a target with nothing attacking or blocking")
	}
	if findBattlefieldCardForTest(g, idle).DamageMarked != 0 {
		t.Error("a creature outside combat was damaged")
	}
}

// --- Set Phasers to . . . ------------------------------------------------------------

func TestTrekB4SetPhasersStunTapsAndStuns(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	trekB4ToMain(t, g)
	victim := trekB4Creature(g, opp.ID, "Victim", "Creature — Bear", "", 2, 2)
	trekB4Plains(g, me.ID, 1)
	trekB4Cast(t, g, me, game.Card{Name: "Set Phasers to . . .", TypeLine: "Instant", ManaCost: "{W}", OracleID: trekB4PhasersOracle},
		game.CastSpellParams{Modes: []int{0}, Targets: cardRefs(victim), Strict: true, AutoTap: true})
	passPriorityAroundTable(t, g)
	c := findBattlefieldCardForTest(g, victim)
	if !c.Tapped || c.Counters[game.CounterStun] != 1 {
		t.Errorf("tapped %v, stun %d; want tapped with one stun counter", c.Tapped, c.Counters[game.CounterStun])
	}
}

func TestTrekB4SetPhasersKillHitsOnlyAnAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	trekB4ToMain(t, g)
	idle := trekB4Creature(g, opp.ID, "Idle", "Creature — Bear", "", 3, 3)
	trekB4Plains(g, me.ID, 1)
	id := trekB4Hand(me, game.Card{Name: "Set Phasers to . . .", TypeLine: "Instant", ManaCost: "{W}", OracleID: trekB4PhasersOracle})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{1}, Targets: cardRefs(idle), Strict: true, AutoTap: true}); err == nil {
		t.Fatal("Kill targeted a creature that is not attacking")
	}

	attacker := trekB4Creature(g, opp.ID, "Attacker", "Creature — Bear", "", 3, 3)
	helicarrierAttacker(g, attacker, me.ID)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{1}, Targets: cardRefs(attacker), Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("Kill at an attacker: %v", err)
	}
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(attacker) {
		t.Error("3 damage did not kill the 3/3 attacker")
	}
}

// --- Shields Up! ----------------------------------------------------------------------

func TestTrekB4ShieldsUpProtectsAndCountersOnlyACreature(t *testing.T) {
	for _, c := range []struct {
		name, typeLine string
		counter        int
	}{
		{"a creature", "Creature — Bear", 1},
		{"an artifact", "Artifact", 0},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			trekB4ToMain(t, g)
			id := trekB4Creature(g, me.ID, "Mine", c.typeLine, "", 2, 2)
			trekB4Plains(g, me.ID, 2)
			trekB4Cast(t, g, me, game.Card{Name: "Shields Up!", TypeLine: "Instant", ManaCost: "{1}{W}", OracleID: trekB4ShieldsUpOracle},
				game.CastSpellParams{Targets: cardRefs(id), Strict: true, AutoTap: true})
			passPriorityAroundTable(t, g)
			assertKeywords(t, g, id, "hexproof", "indestructible")
			if n := counterCount(g, id, game.CounterPlusOne); n != c.counter && !(c.counter == 0 && n <= 0) {
				t.Errorf("%d counters, want %d", n, c.counter)
			}
		})
	}
}

func TestTrekB4ShieldsUpRefusesAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	trekB4ToMain(t, g)
	theirs := trekB4Creature(g, opp.ID, "Theirs", "Creature — Bear", "", 2, 2)
	trekB4Plains(g, me.ID, 2)
	id := trekB4Hand(me, game.Card{Name: "Shields Up!", TypeLine: "Instant", ManaCost: "{1}{W}", OracleID: trekB4ShieldsUpOracle})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: cardRefs(theirs), Strict: true, AutoTap: true}); err == nil {
		t.Fatal("Shields Up! targeted a creature the caster does not control")
	}
}

// --- Shuttle Ace ----------------------------------------------------------------------

func TestTrekB4ShuttleAceFliesOnlyOnItsControllersTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	trekB4ToMain(t, g)
	mine := trekB4Creature(g, me.ID, "Shuttle Ace", "Creature — Kelpien Pilot", trekB4ShuttleAceOracle, 4, 4)
	theirs := trekB4Creature(g, opp.ID, "Shuttle Ace", "Creature — Kelpien Pilot", trekB4ShuttleAceOracle, 4, 4)
	if !effectiveAbilitiesContain(t, g, mine, "flying") {
		t.Error("Shuttle Ace has no flying during its controller's turn")
	}
	if effectiveAbilitiesContain(t, g, theirs, "flying") {
		t.Error("an opponent's Shuttle Ace has flying during another player's turn")
	}
}

func TestTrekB4ShuttleAceGainsTwoWhenItBecomesTapped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	trekB4ToMain(t, g)
	ace := trekB4Creature(g, me.ID, "Shuttle Ace", "Creature — Kelpien Pilot", trekB4ShuttleAceOracle, 4, 4)
	before := me.Life
	trekB4Attack(t, g, opp.ID, ace)
	if me.Life != before+2 {
		t.Errorf("life %d → %d, want +2 when the Ace taps to attack", before, me.Life)
	}
}

// --- Sulu, Space Swashbuckler ----------------------------------------------------------

func TestTrekB4SuluDrawsOnlyAfterAnEquipmentOrSpacecraftEntered(t *testing.T) {
	for _, c := range []struct {
		name, entered string
		draws         int
	}{
		{"nothing entered", "", 0},
		{"a plain artifact entered", "Artifact", 0},
		{"an Equipment entered", "Artifact — Equipment", 1},
		{"a Spacecraft entered", "Artifact — Spacecraft", 1},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			trekB4ToMain(t, g)
			sulu := trekB4Creature(g, me.ID, "Sulu, Space Swashbuckler", "Legendary Creature — Human Pilot", trekB4SuluOracle, 2, 2)
			if c.entered != "" {
				trekB4Cast(t, g, me, game.Card{Name: "Gear", TypeLine: c.entered}, game.CastSpellParams{})
				passPriorityAroundTable(t, g)
			}
			before := me.Hand.Size()
			trekB4Attack(t, g, opp.ID, sulu)
			if got := me.Hand.Size() - before; got != c.draws {
				t.Errorf("drew %d, want %d", got, c.draws)
			}
		})
	}
}

// --- Tactical Officer ---------------------------------------------------------------------

func TestTrekB4TacticalOfficerGrowsAndUntapsAtTotalPowerEight(t *testing.T) {
	for _, c := range []struct {
		name    string
		big     int
		counter int
	}{
		{"total power 8", 6, 1},
		{"total power 7", 5, 0},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			trekB4ToMain(t, g)
			officer := trekB4Creature(g, me.ID, "Tactical Officer", "Creature — Human Officer", trekB4TacticalOracle, 2, 1)
			big := trekB4Creature(g, me.ID, "Big", "Creature — Giant", "", c.big, c.big)
			trekB4Attack(t, g, opp.ID, officer, big)
			o := findBattlefieldCardForTest(g, officer)
			if o.Counters[game.CounterPlusOne] != c.counter {
				t.Errorf("%d counters, want %d", o.Counters[game.CounterPlusOne], c.counter)
			}
			if o.Tapped == (c.counter == 1) {
				t.Errorf("tapped = %v after the declaration", o.Tapped)
			}
		})
	}
}

// --- Test of Conviction -----------------------------------------------------------------

func TestTrekB4TestOfConvictionBattleStations(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	trekB4ToMain(t, g)
	a := trekB4Creature(g, me.ID, "A", "Creature — Bear", "", 2, 2)
	b := trekB4Creature(g, me.ID, "B", "Creature — Bear", "", 2, 2)
	theirs := trekB4Creature(g, opp.ID, "Theirs", "Creature — Bear", "", 2, 2)
	trekB4Plains(g, me.ID, 2)
	trekB4Cast(t, g, me, game.Card{Name: "Test of Conviction", TypeLine: "Sorcery", ManaCost: "{1}{W}", OracleID: trekB4ConvictionOracle},
		game.CastSpellParams{Modes: []int{0}, Strict: true, AutoTap: true})
	passPriorityAroundTable(t, g)
	if counterCount(g, a, game.CounterPlusOne) != 1 || counterCount(g, b, game.CounterPlusOne) != 1 {
		t.Error("a creature you control has no counter")
	}
	if counterCount(g, theirs, game.CounterPlusOne) > 0 {
		t.Error("an opponent's creature got a counter")
	}
}

func TestTrekB4TestOfConvictionBeginNegotiations(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	trekB4ToMain(t, g)
	trekB4Plains(g, me.ID, 2)
	id := trekB4Hand(me, game.Card{Name: "Test of Conviction", TypeLine: "Sorcery", ManaCost: "{1}{W}", OracleID: trekB4ConvictionOracle})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{1}, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
		Strict: true, AutoTap: true}); err == nil {
		t.Fatal("Begin Negotiations targeted its own caster")
	}
	mine, theirs := me.Hand.Size(), opp.Hand.Size()
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{1}, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("Begin Negotiations: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - (mine - 1); got != 2 {
		t.Errorf("you drew %d, want 2", got)
	}
	if got := opp.Hand.Size() - theirs; got != 1 {
		t.Errorf("the opponent drew %d, want 1", got)
	}
}

// --- There's Coffee in that Nebula ---------------------------------------------------------

func TestTrekB4CoffeeOffersAPlainsOrACheapNonlandPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	trekB4ToMain(t, g)
	top := trekB4Top(me,
		game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Three", TypeLine: "Creature — Bear", ManaCost: "{2}{G}"},
		game.Card{Name: "Four", TypeLine: "Artifact", ManaCost: "{4}"},
		game.Card{Name: "Bolt", TypeLine: "Instant", ManaCost: "{R}"},
	)
	trekB4Plains(g, me.ID, 2)
	coffee := trekB4Cast(t, g, me, game.Card{Name: "There's Coffee in that Nebula", TypeLine: "Sorcery", ManaCost: "{1}{W}", OracleID: trekB4CoffeeOracle},
		game.CastSpellParams{Strict: true, AutoTap: true})
	passPriorityAroundTable(t, g)
	prompt := latestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("no look prompt")
	}
	offered := trekB4Offered(prompt)
	if len(offered) != 2 || !offered[top[0]] || !offered[top[2]] {
		t.Fatalf("offered %v, want only the Plains and the mana value 3 creature", prompt.ChooseCards)
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{top[2]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(top[2]) {
		t.Error("the chosen card is not in hand")
	}
	if !me.Graveyard.Contains(coffee) {
		t.Error("cast from hand, the sorcery did not go to the graveyard")
	}
}

func TestTrekB4CoffeeBeamsUpAndIsExiled(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	coffee := seedGraveyardCard(t, g, "There's Coffee in that Nebula", "Sorcery", trekB4CoffeeOracle)
	crew := trekB4Creature(g, me.ID, "Crew", "Creature — Human", "", 1, 1)
	trekB4Plains(g, me.ID, 3)
	if err := g.CastSpell(me.ID, coffee, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: trekB4AltCostKeyBeamMeUp, AltCostIDs: []uuid.UUID{crew}, Strict: true, AutoTap: true,
	}); err != nil {
		t.Fatalf("beam me up: %v", err)
	}
	passPriorityAroundTable(t, g)
	if prompt := latestChooseCards(g, me.ID); prompt != nil {
		if err := g.ResolveChooseCards(prompt.ID, me.ID, nil); err != nil {
			t.Fatalf("declining: %v", err)
		}
	}
	if !me.Hand.Contains(crew) {
		t.Error("the returned creature is not in hand")
	}
	if !g.Exile.Contains(coffee) {
		t.Error("the beamed-up sorcery was not exiled")
	}
}
