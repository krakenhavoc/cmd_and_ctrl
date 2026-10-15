package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mass_casualty_a_test.go — the Mass Casualty deck request (#3245),
// slice A: the cards existing seams carry, plus the Snuff Out review.

const (
	mcEclipsedSteppeOracle = "6216635f-8e6e-40a3-9659-ef6352ab92ce"
	mcVitoOracle           = "7336a541-3ae9-4ab2-acb8-f574c0a5b3a9"
	mcCullingOracle        = "6385fd0d-256d-4400-9319-afa8170024db"
	mcPriestOfGixOracle    = "d93f82ce-0eed-45cc-a7b1-50fd4cbb6152"
	mcArchonOracle         = "ceef2d5a-77ea-4e56-9806-fd1a2d5be400"
	mcBeaconOracle         = "bfcdf563-c6e5-4a54-bb63-4aba219b94e2"
	mcShallowGraveOracle   = "9843d8a6-bcbf-40c2-8c41-316e97db9c62"
	mcRobeOfStarsOracle    = "416f49e2-feb1-43e7-befc-b0e02181b202"
	mcAinokOracle          = "8c1fda45-09f1-4b58-a487-e11c868e1357"
	mcVoiceOracle          = "b12cbda1-ac09-49eb-96b3-a4df27bc37ea"
)

func TestMassCasualtySliceAIsRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		mcEclipsedSteppeOracle: "Eclipsed Steppe",
		mcVitoOracle:           "Vito, Thorn of the Dusk Rose",
		mcCullingOracle:        "Culling the Weak",
		mcPriestOfGixOracle:    "Priest of Gix",
		mcArchonOracle:         "Archon of Emeria",
		mcBeaconOracle:         "Beacon of Immortality",
		mcShallowGraveOracle:   "Shallow Grave",
		mcRobeOfStarsOracle:    "Robe of Stars",
		mcAinokOracle:          "Ainok Strike Leader",
		mcVoiceOracle:          "Voice of Victory",
		snuffOutOracle:         "Snuff Out",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name {
			t.Errorf("%s registered as %q", oracle, spec.Name)
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: Completeness = %v, Caveats = %v", name, spec.Completeness, spec.Caveats)
		}
	}
}

func mcPoolCount(p *game.Player, color string) int {
	n := 0
	for _, tok := range p.ManaPool.Units() {
		if tok.Color == color {
			n++
		}
	}
	return n
}

func TestCullingTheWeakSacrificesACreatureForFourBlack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := paPermanent(g, me.ID, "Bear", "Creature — Bear", "{1}{G}", 2, 2)
	paCast(t, g, "Culling the Weak", "Instant", "{B}", mcCullingOracle, game.CastSpellParams{
		SacrificeIDs: []uuid.UUID{bear},
	})
	if g.Battlefield.Contains(bear) {
		t.Fatal("the creature was not sacrificed as the spell was cast")
	}
	passPriorityAroundTable(t, g)
	if got := mcPoolCount(me, "B"); got != 4 {
		t.Errorf("black mana in the pool = %d, want 4", got)
	}
}

func TestPriestOfGixAddsThreeBlackWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Priest of Gix", "Creature — Phyrexian Human Cleric Minion", mcPriestOfGixOracle, nil)
	passPriorityAroundTable(t, g)
	if got := mcPoolCount(me, "B"); got != 3 {
		t.Errorf("black mana in the pool = %d, want 3", got)
	}
}

func TestVitoDrainsWhatYouGainAndGrantsLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vito := pushCatalogPermanent(g, me.ID, "Vito, Thorn of the Dusk Rose", "Legendary Creature — Vampire Cleric", mcVitoOracle, false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	before := opp.Life

	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 4) })
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != before-4 {
		t.Errorf("the opponent lost %d, want 4", before-opp.Life)
	}

	advanceTo(t, g, game.StepPrecombatMain)
	fillPoolColored(me, "B", 5)
	if err := g.ActivateCatalogAbility(me.ID, vito, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Vito: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{vito, bear} {
		if !containsString(effectiveAbilities(t, g, id), "lifelink") {
			t.Errorf("a creature Vito's controller controls has no lifelink")
		}
	}
	if containsString(effectiveAbilities(t, g, theirs), "lifelink") {
		t.Error("an opponent's creature gained lifelink")
	}
}

func TestArchonOfEmeriaTapsOpponentsNonbasicsAndAllowsOneSpell(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, active.ID, "Archon of Emeria", "Creature — Archon", mcArchonOracle, 2, 3)

	if b13Tapped(t, g, b12PlayFromHand(t, g, "My Tomb", "Land", "", game.CastSpellParams{})) {
		t.Error("the Archon's controller's own nonbasic land entered tapped")
	}
	if b13Tapped(t, g, b13PlayAs(t, g, 1, "Plains", "Basic Land — Plains", "")) {
		t.Error("an opponent's basic land entered tapped")
	}
	if !b13Tapped(t, g, b13PlayAs(t, g, 2, "Ancient Tomb", "Land", "")) {
		t.Error("an opponent's nonbasic land entered untapped")
	}

	if _, err := castWithOptionalCosts(t, g, "Burst Lightning", "Instant", burstLightningOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}, nil, nil); err != nil {
		t.Fatalf("first spell under the Archon was refused: %v", err)
	}
	passPriorityAroundTable(t, g)
	_, err := castWithOptionalCosts(t, g, "Burst Lightning", "Instant", burstLightningOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}, nil, nil)
	assertCantCast(t, err, "more than one spell")
}

func TestBeaconOfImmortalityDoublesLifeAndShufflesIntoTheLibrary(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Life = 17
	beacon := castCatalogSpell(t, g, "Beacon of Immortality", "Instant", mcBeaconOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	passPriorityAroundTable(t, g)
	if me.Life != 34 {
		t.Errorf("life = %d, want 34", me.Life)
	}
	if !me.Library.Contains(beacon) {
		t.Error("Beacon is not in its owner's library")
	}
	if me.Graveyard.Contains(beacon) {
		t.Error("Beacon went to the graveyard")
	}
}

func TestShallowGraveReturnsTheTopCreatureWithHasteThenExilesIt(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	older := pushGraveyardCreature(g, me.ID, "Older Bear", "{1}{G}")
	newer := pushGraveyardCreature(g, me.ID, "Newer Bear", "{1}{G}")
	land := pushGraveyardCardWithTypeLineID(me, "Swamp", "Basic Land — Swamp")

	castCatalogSpell(t, g, "Shallow Grave", "Instant", mcShallowGraveOracle, nil)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(newer) {
		t.Fatal("the topmost creature card did not return")
	}
	if g.Battlefield.Contains(older) || g.Battlefield.Contains(land) {
		t.Error("something other than the topmost creature card returned")
	}
	if !containsString(effectiveAbilities(t, g, newer), "haste") {
		t.Error("the returned creature has no haste")
	}
	advanceToEndStepOf(t, g, seat)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(newer) || !g.Exile.Contains(newer) {
		t.Error("the returned creature was not exiled at the end step")
	}
}

func TestRobeOfStarsPumpsAndPhasesTheCreatureOut(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	robe := pushCatalogPermanent(g, me.ID, "Robe of Stars", "Artifact — Equipment", mcRobeOfStarsOracle, false)
	activateEquipWithoutResolving(t, g, me.ID, robe, bear)
	passPriorityAroundTable(t, g)
	if got := effectiveToughness(t, g, bear); got != 5 {
		t.Fatalf("equipped toughness = %d, want 5", got)
	}

	fillPoolColored(me, "W", 2)
	if err := g.ActivateCatalogAbility(me.ID, robe, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Astral Projection: %v", err)
	}
	passPriorityAroundTable(t, g)
	var bearOut, robeOut bool
	g.ReadSnapshot(func() { bearOut, robeOut = g.IsPhasedOutForEffect(bear), g.IsPhasedOutForEffect(robe) })
	if !bearOut || !robeOut {
		t.Errorf("phased out: creature %v, Robe %v; want both", bearOut, robeOut)
	}
}

func TestAinokStrikeLeaderMakesAGoblinPerOpponentForItselfOrYourCommander(t *testing.T) {
	for _, tc := range []struct {
		name      string
		commander bool
	}{{"Ainok attacks", false}, {"your commander attacks alone", true}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			ainok := pushCatalogPermanent(g, me.ID, "Ainok Strike Leader", "Creature — Dog Warrior", mcAinokOracle, false)
			attacker := ainok
			if tc.commander {
				attacker = pushVanillaCreature(g, me.ID, "My Commander", 3, 3)
				g.WithWriteLock(func() {
					for i := range g.Battlefield.Cards {
						if g.Battlefield.Cards[i].InstanceID == attacker {
							g.Battlefield.Cards[i].IsCommander = true
						}
					}
				})
			}
			declareAttack(t, g, opp.ID, attacker)
			passPriorityAroundTable(t, g)

			targets := map[uuid.UUID]int{}
			for _, c := range g.Battlefield.Cards {
				if c.Name == "Goblin" && c.Controller == me.ID {
					if !c.Tapped || c.AttackingTarget == uuid.Nil {
						t.Error("a Goblin is not tapped and attacking")
					}
					targets[c.AttackingTarget]++
				}
			}
			if len(targets) != len(g.Seats)-1 {
				t.Errorf("Goblins attack %v, want one at each of %d opponents", targets, len(g.Seats)-1)
			}
		})
	}
}

func TestAinokStrikeLeaderStaysQuietWhenNeitherAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	pushCatalogPermanent(g, me.ID, "Ainok Strike Leader", "Creature — Dog Warrior", mcAinokOracle, false)
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Goblin"); n != 0 {
		t.Errorf("a plain creature attacking alone made %d Goblins, want 0", n)
	}
}

func TestAinokStrikeLeaderSacrificeMakesTokensIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ainok := pushCatalogPermanent(g, me.ID, "Ainok Strike Leader", "Creature — Dog Warrior", mcAinokOracle, false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	tokens := pushTokens(g, me.ID, RedGoblinToken(), 1)
	if err := g.ActivateCatalogAbility(me.ID, ainok, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the sacrifice: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ainok) {
		t.Error("Ainok was not sacrificed")
	}
	if !containsString(effectiveAbilities(t, g, tokens[0]), "indestructible") {
		t.Error("the creature token is not indestructible")
	}
	if containsString(effectiveAbilities(t, g, bear), "indestructible") {
		t.Error("a nontoken creature became indestructible")
	}
}

func TestVoiceOfVictoryMobilizesAndSacrificesTheWarriors(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	voice := pushCatalogPermanent(g, me.ID, "Voice of Victory", "Creature — Human Bard", mcVoiceOracle, false)
	declareAttack(t, g, opp.ID, voice)
	passPriorityAroundTable(t, g)
	// The defender is offered first; all the Warriors go at it.
	for c := latestChoiceOfKind(g, game.PendingChoicePayAmount); c != nil; c = latestChoiceOfKind(g, game.PendingChoicePayAmount) {
		if err := g.ResolvePayAmount(c.ID, me.ID, c.PayAmount.Max); err != nil {
			t.Fatalf("place the Warriors: %v", err)
		}
	}
	var warriors []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Warrior" && c.Controller == me.ID {
			if !c.Tapped || c.AttackingTarget != opp.ID {
				t.Errorf("a Warrior is tapped=%v attacking %v, want tapped and attacking the defender", c.Tapped, c.AttackingTarget)
			}
			warriors = append(warriors, c.InstanceID)
		}
	}
	if len(warriors) != 2 {
		t.Fatalf("mobilize 2 made %d Warriors", len(warriors))
	}
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	for _, id := range warriors {
		if g.Battlefield.Contains(id) {
			t.Error("a Warrior survived the end step")
		}
	}
}

func TestVoiceOfVictoryStopsOpponentsCastingOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushCatalogPermanent(g, active.ID, "Voice of Victory", "Creature — Human Bard", mcVoiceOracle, false)
	advanceToMain(t, g)
	bolt := handCardFull(opp, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle, []string{"R"})
	err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: active.ID}},
	})
	if err == nil {
		t.Fatal("an opponent cast a spell during the Voice's controller's turn")
	}
	if _, err := castWithOptionalCosts(t, g, "Burst Lightning", "Instant", burstLightningOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, nil, nil); err != nil {
		t.Errorf("the Voice's controller was refused: %v", err)
	}
}

func TestEclipsedSteppeEntersTappedWithoutTwoBasics(t *testing.T) {
	g := newCatalogGame(t)
	if !b13Tapped(t, g, b12PlayFromHand(t, g, "Eclipsed Steppe", "Land — Plains Swamp", mcEclipsedSteppeOracle, game.CastSpellParams{})) {
		t.Error("with no basics the land entered untapped")
	}
}

func TestSnuffOutDestroysANonblackCreatureOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	bear := freeSpellPermanent(g, them.ID, "Grizzly Bears", "Creature — Bear")
	castCatalogSpell(t, g, "Snuff Out", "Instant", snuffOutOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the nonblack creature survived")
	}

	zombie := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: zombie, Name: "Zombie", TypeLine: "Creature — Zombie",
		Colors: []string{"B"}, Power: 2, Toughness: 2, Owner: them.ID, Controller: them.ID})
	snuff := handCardFull(me, "Snuff Out", "Instant", "{3}{B}", snuffOutOracle, []string{"B"})
	if err := g.CastSpell(me.ID, snuff, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: zombie}},
	}); err == nil {
		t.Error("Snuff Out targeted a black creature")
	}
}
