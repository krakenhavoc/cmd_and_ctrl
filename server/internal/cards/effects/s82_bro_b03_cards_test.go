package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bro_b03_cards_test.go — S82 BRO batch 3 (#5482): the Powerstone
// makers. Every test plays the real catalog card through the engine.

const (
	broB03StoneUnit   = "31d9a771-09ae-4cd4-bd19-6012f839a9e8"
	broB03Opportunist = "cc9d3868-9814-4506-80b3-660bc462ff98"
	broB03Roc         = "0956c0f5-ff1d-43d5-88de-4c2d603cf798"
	broB03Engineer    = "f2ffc999-609f-4424-8d40-ebc042357a40"
	broB03Stoneseeker = "78d74d6a-e1aa-4d70-92c3-b5338959d2fb"
	broB03Genius      = "68d4d753-5e53-48a6-8ba2-56020893c239"
	broB03Prospector  = "37278877-e7d8-47b6-a65f-37c7f85e3a0a"
	broB03Geology     = "c7c1352a-0332-4495-a384-8d6fed01f91b"
	broB03Explosion   = "4bb75182-6dd4-42f7-9541-d484d0b57897"
	broB03Excavation  = "b13c91be-079a-46f1-8a99-b7dba99bb9c0"
	broB03Rivalry     = "7fa1430f-cd2e-4071-a161-e32a61e33714"
	broB03Repair      = "4f42517b-e62e-4b8c-8c56-cb32fb62c478"
	broB03Hall        = "fd809587-773c-47ec-8676-cefef6d1e38f"
)

func TestBroB03EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		broB03StoneUnit: "Stone Retrieval Unit", broB03Opportunist: "Argothian Opportunist", broB03Roc: "Koilos Roc",
		broB03Engineer: "Powerstone Engineer", broB03Stoneseeker: "Horned Stoneseeker", broB03Genius: "Junkyard Genius",
		broB03Prospector: "Great Desert Prospector", broB03Geology: "Geology Enthusiast", broB03Explosion: "Excavation Explosion",
		broB03Excavation: "Fallaji Excavation", broB03Rivalry: "Sibling Rivalry", broB03Repair: "Repair and Recharge",
		broB03Hall: "Hall of Tagsin",
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

// broB03Stones returns the Powerstones p controls, as (total, tapped).
func broB03Stones(g *game.Game, p uuid.UUID) (n, tapped int) {
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Powerstone" && c.Controller == p {
				n++
				if c.Tapped {
					tapped++
				}
			}
		}
	})
	return n, tapped
}

func broB03WantStones(t *testing.T, g *game.Game, p uuid.UUID, want int) {
	t.Helper()
	n, tapped := broB03Stones(g, p)
	if n != want || tapped != want {
		t.Errorf("Powerstones %d (tapped %d), want %d, all tapped", n, tapped, want)
	}
}

func TestBroB03EntersCreaturesMakeATappedPowerstone(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Stone Retrieval Unit", "Artifact Creature — Construct", broB03StoneUnit},
		{"Argothian Opportunist", "Creature — Human Scout", broB03Opportunist},
		{"Koilos Roc", "Creature — Bird", broB03Roc},
		{"Horned Stoneseeker", "Creature — Lizard", broB03Stoneseeker},
		{"Junkyard Genius", "Creature — Human Artificer", broB03Genius},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			castCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, nil)
			passPriorityAroundTable(t, g)
			broB03WantStones(t, g, me.ID, 1)
		})
	}
}

func TestBroB03PowerstoneEngineerMakesOneWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	id := b12Push(g, me.ID, "Powerstone Engineer", "Creature — Human Artificer", broB03Engineer, 2, 1)
	broB03WantStones(t, g, me.ID, 0)
	killCreature(t, g, me.ID, id)
	passPriorityAroundTable(t, g)
	broB03WantStones(t, g, me.ID, 1)
}

func TestBroB03HornedStoneseekerSacrificesAPowerstoneWhenItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	id := b12Push(g, me.ID, "Horned Stoneseeker", "Creature — Lizard", broB03Stoneseeker, 2, 2)
	pushToken(g, me.ID, PowerstoneToken())
	pushToken(g, me.ID, PowerstoneToken())
	exileCreature(t, g, id)
	passPriorityAroundTable(t, g)
	if pick := latestChooseCardsFor(g, me.ID); pick != nil {
		var stone uuid.UUID
		g.ReadSnapshot(func() {
			for _, c := range g.Battlefield.Cards {
				if c.Name == "Powerstone" {
					stone = c.InstanceID
				}
			}
		})
		answerChooseCards(t, g, me.ID, stone)
		g.SettleResolution()
	}
	if n, _ := broB03Stones(g, me.ID); n != 1 {
		t.Errorf("%d Powerstones left, want 1 (one sacrificed)", n)
	}
}

func TestBroB03HornedStoneseekerWithNoPowerstoneDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	id := b12Push(g, me.ID, "Horned Stoneseeker", "Creature — Lizard", broB03Stoneseeker, 2, 2)
	killCreature(t, g, me.ID, id)
	passPriorityAroundTable(t, g)
	if n, _ := broB03Stones(g, me.ID); n != 0 {
		t.Errorf("%d Powerstones, want 0", n)
	}
}

func TestBroB03JunkyardGeniusRalliesOtherCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	genius := b12Push(g, me.ID, "Junkyard Genius", "Creature — Human Artificer", broB03Genius, 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Rat", 1, 1)
	fillPoolColored(me, "B", 1)
	fillPoolColored(me, "R", 1)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, genius, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	if p := effectivePower(t, g, bear); p != 3 {
		t.Errorf("Bear power %d, want 3", p)
	}
	if !hasEffectiveKeyword(t, g, bear, "menace") || !hasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("Bear should have menace and haste")
	}
	if p := effectivePower(t, g, genius); p != 2 {
		t.Errorf("Junkyard Genius power %d, want 2 (other creatures only)", p)
	}
	if hasEffectiveKeyword(t, g, genius, "menace") {
		t.Error("Junkyard Genius gained menace")
	}
	// The sacrifice is a cost: without another creature or artifact it can't be activated.
	fillPoolColored(me, "B", 1)
	fillPoolColored(me, "R", 1)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	killCreature(t, g, me.ID, bear)
	if err := g.ActivateCatalogAbility(me.ID, genius, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with nothing else to sacrifice")
	}
}

func TestBroB03GreatDesertProspectorMakesOnePerOtherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Great Desert Prospector", "Creature — Human Artificer", broB03Prospector, nil)
	passPriorityAroundTable(t, g)
	broB03WantStones(t, g, me.ID, 2)
}

func TestBroB03GreatDesertProspectorAloneMakesNone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Great Desert Prospector", "Creature — Human Artificer", broB03Prospector, nil)
	passPriorityAroundTable(t, g)
	broB03WantStones(t, g, me.ID, 0)
}

func TestBroB03GeologyEnthusiastEndStepAndActivation(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	seedLibrary(me, "A", "B")
	id := b12Push(g, me.ID, "Geology Enthusiast", "Creature — Human Artificer", broB03Geology, 2, 5)
	hand := handSize(me)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"},
		game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if handSize(me)-hand != 1 {
		t.Errorf("drew %d cards, want 1", handSize(me)-hand)
	}
	if n := countersOn(g, id, "+1/+1"); n != 1 {
		t.Errorf("%d +1/+1 counters, want 1", n)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated without {6}")
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	broB03WantStones(t, g, me.ID, 1)
}

func TestBroB03GeologyEnthusiastIgnoresOpponentsEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToUpkeepOf(t, g, 1)
	b12Push(g, me.ID, "Geology Enthusiast", "Creature — Human Artificer", broB03Geology, 2, 5)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	broB03WantStones(t, g, me.ID, 0)
}

func TestBroB03ExcavationExplosionDamagesAndMakesAPowerstone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	life := opp.Life
	castCatalogSpell(t, g, "Excavation Explosion", "Sorcery", broB03Explosion,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d, want %d", opp.Life, life-3)
	}
	broB03WantStones(t, g, me.ID, 1)
}

func TestBroB03ExcavationExplosionKillsACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 3)
	castCatalogSpell(t, g, "Excavation Explosion", "Sorcery", broB03Explosion, b16TargetCard(victim))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("3 damage should kill a 3-toughness creature")
	}
	broB03WantStones(t, g, me.ID, 1)
}

func TestBroB03FallajiExcavationMakesThreeAndGainsThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	castCatalogSpell(t, g, "Fallaji Excavation", "Sorcery", broB03Excavation, nil)
	passPriorityAroundTable(t, g)
	broB03WantStones(t, g, me.ID, 3)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

func TestBroB03SiblingRivalryStealsAnArtifactOrCreature(t *testing.T) {
	for _, tc := range []struct{ name, typeLine string }{
		{"creature", "Creature — Bear"},
		{"artifact", "Artifact"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			advanceToMain(t, g)
			var theirs uuid.UUID
			if tc.name == "creature" {
				theirs = b12Creature(g, opp.ID, "Theirs", tc.typeLine, 2, 2)
			} else {
				theirs = b12Permanent(g, opp.ID, "Theirs", tc.typeLine)
			}
			castCatalogSpell(t, g, "Sibling Rivalry", "Sorcery", broB03Rivalry, b16TargetCard(theirs))
			passPriorityAroundTable(t, g)
			if controllerOf(t, g, theirs) != me.ID {
				t.Error("the target was not stolen")
			}
			if !hasEffectiveKeyword(t, g, theirs, "haste") {
				t.Error("the target should have haste")
			}
			broB03WantStones(t, g, me.ID, 1)
			advanceToUpkeepOf(t, g, 1)
			if controllerOf(t, g, theirs) != opp.ID {
				t.Error("control should revert at end of turn")
			}
		})
	}
}

func TestBroB03SiblingRivalryRefusesALand(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	advanceToMain(t, g)
	land := b12Permanent(g, opp.ID, "Plains", "Basic Land — Plains")
	id := uuid.New()
	me := g.Seats[0]
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Sibling Rivalry", TypeLine: "Sorcery", OracleID: broB03Rivalry, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: b16TargetCard(land), Strict: true}); err == nil {
		t.Error("targeted a land")
	}
}

func TestBroB03RepairAndRechargeReturnsAndMakesAPowerstone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	ring := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: ring, Name: "Ring", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	castCatalogSpell(t, g, "Repair and Recharge", "Sorcery", broB03Repair, b16TargetCard(ring))
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ring) {
		t.Error("the artifact did not return")
	}
	broB03WantStones(t, g, me.ID, 1)
}

func TestBroB03RepairAndRechargeRefusesACreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Repair and Recharge", TypeLine: "Sorcery", OracleID: broB03Repair, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: b16TargetCard(bear), Strict: true}); err == nil {
		t.Error("targeted a creature card")
	}
}

func TestBroB03HallOfTagsinManaAndPowerstone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	hall := b12Permanent(g, me.ID, "Hall of Tagsin", "Land")
	g.Battlefield.Cards[len(g.Battlefield.Cards)-1].OracleID = broB03Hall
	if n := len(game.ManaAbilitiesForCard(g.Battlefield.Cards[len(g.Battlefield.Cards)-1])); n != 2 {
		t.Fatalf("%d mana abilities, want 2", n)
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, hall, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	broB03WantStones(t, g, me.ID, 1)
	if err := g.ActivateCatalogAbility(me.ID, hall, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated a tapped land / without {4}")
	}
}
