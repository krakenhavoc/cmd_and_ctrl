package testso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ons_batch1_cards_test.go — Onslaught, batch 1 (#5610, S82): thirteen
// A–C cards built from seams the engine already has. Each test plays
// the real catalog entry.

const (
	ons1AccursedCentaurOracle   = "b1e6921e-e460-4e58-82b0-d98ee31c279d"
	ons1AkromasVengeanceOracle  = "ae19dce5-c9fc-4fbe-99b5-dc6a12124a68"
	ons1AnuridMurkdiverOracle   = "e7dc3d8b-2c79-463d-b862-9f01ccea05fd"
	ons1AphettoAlchemistOracle  = "867aa0b7-b814-4d33-a677-f6368c8e93c0"
	ons1AvenBrigadierOracle     = "910ae0ca-257d-4b45-b039-f26e7b2f3d5c"
	ons1BarkhideMaulerOracle    = "84a18b1f-5f9f-4e2b-87cd-c437bdf4d278"
	ons1BatteringCraghornOracle = "a084bca4-fe72-443e-adf4-ab5456c7e699"
	ons1BlisteringFirecatOracle = "b3595359-0d40-495e-85da-6a56fbce7357"
	ons1CabalArchonOracle       = "7d80f016-5a4a-4843-9aec-0fe64f7f7ae8"
	ons1CentaurGladeOracle      = "369c3b6c-6715-4b6a-89d2-e7542400ba8f"
	ons1ChargingSlatebackOracle = "67457ecb-5ff2-4f5d-9039-b9e80f54bba8"
	ons1CraftyPathmageOracle    = "c4020178-71d9-4363-a0bd-b9a7eecd5d02"
	ons1CrudeRampartOracle      = "c9823e5f-d0f8-4d28-8b3b-55f94ea9d76f"
)

func TestOns1CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		ons1AccursedCentaurOracle:   "Accursed Centaur",
		ons1AkromasVengeanceOracle:  "Akroma's Vengeance",
		ons1AnuridMurkdiverOracle:   "Anurid Murkdiver",
		ons1AphettoAlchemistOracle:  "Aphetto Alchemist",
		ons1AvenBrigadierOracle:     "Aven Brigadier",
		ons1BarkhideMaulerOracle:    "Barkhide Mauler",
		ons1BatteringCraghornOracle: "Battering Craghorn",
		ons1BlisteringFirecatOracle: "Blistering Firecat",
		ons1CabalArchonOracle:       "Cabal Archon",
		ons1CentaurGladeOracle:      "Centaur Glade",
		ons1ChargingSlatebackOracle: "Charging Slateback",
		ons1CraftyPathmageOracle:    "Crafty Pathmage",
		ons1CrudeRampartOracle:      "Crude Rampart",
	} {
		def, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if def.Name != name || def.Completeness != effects.CompletenessFull {
			t.Errorf("%s: name %q, completeness %v", name, def.Name, def.Completeness)
		}
	}
}

// Every morph in the batch casts face down for the keyword's {3} and
// turns face up for the card's own printed cost (CR 702.37).
func TestOns1MorphCostsAreThePrintedOnes(t *testing.T) {
	for _, tc := range []struct{ name, oracle, faceUp string }{
		{"Aphetto Alchemist", ons1AphettoAlchemistOracle, "{U}"},
		{"Battering Craghorn", ons1BatteringCraghornOracle, "{1}{R}{R}"},
		{"Blistering Firecat", ons1BlisteringFirecatOracle, "{R}{R}"},
		{"Charging Slateback", ons1ChargingSlatebackOracle, "{4}{R}"},
		{"Crude Rampart", ons1CrudeRampartOracle, "{4}{W}"},
	} {
		ac := game.AlternativeCostByKey(tc.oracle, "morph")
		if ac == nil || ac.FaceDown == nil {
			t.Errorf("%s offers no morph cast", tc.name)
			continue
		}
		if ac.ManaCost != effects.MorphCastCost || ac.FaceDown.Kind != game.FaceDownMorphed || ac.FaceDown.FaceUpCounter {
			t.Errorf("%s morph = %q %q counter=%v", tc.name, ac.ManaCost, ac.FaceDown.Kind, ac.FaceDown.FaceUpCounter)
		}
		if ac.FaceDown.FaceUpCost != tc.faceUp {
			t.Errorf("%s turns face up for %q, want %q", tc.name, ac.FaceDown.FaceUpCost, tc.faceUp)
		}
	}
}

func TestOns1AccursedCentaurSacrificesTheCreatureYouChoose(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	centaur := effectstest.CastCatalogSpell(t, g, "Accursed Centaur", "Creature — Zombie Centaur", ons1AccursedCentaurOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SacrificeChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no sacrifice prompt with two creatures to choose from")
	}
	opts := effectstest.SacrificeOptions(c)
	if !effectstest.EdictOffers(opts, bear) || !effectstest.EdictOffers(opts, centaur) {
		t.Fatalf("the prompt offers %v, want both the Bears and the Centaur", opts)
	}
	effectstest.AnswerSacrifice(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the chosen creature was not sacrificed")
	}
	if !g.Battlefield.Contains(centaur) {
		t.Error("the Centaur left although the Bears were chosen")
	}
}

func TestOns1AccursedCentaurAloneSacrificesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bears", "Creature — Bear", 2, 2)
	centaur := effectstest.CastCatalogSpell(t, g, "Accursed Centaur", "Creature — Zombie Centaur", ons1AccursedCentaurOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, centaur)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(centaur) {
		t.Error("the Centaur survived its own trigger with no other creature")
	}
	if !g.Battlefield.Contains(theirs) {
		t.Error("an opponent's creature was sacrificed")
	}
}

func TestOns1AkromasVengeanceDestroysArtifactsCreaturesAndEnchantments(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	aura := effectstest.B12Permanent(g, opp.ID, "Their Pact", "Enchantment")
	bear := effectstest.B12Creature(g, me.ID, "My Bears", "Creature — Bear", 2, 2)
	golem := effectstest.B12Creature(g, opp.ID, "Their Golem", "Artifact Creature — Golem", 3, 3)
	land := effectstest.SeedLand(g, opp.ID, "Plains", "Basic Land — Plains", "")
	effectstest.CastCatalogSpell(t, g, "Akroma's Vengeance", "Sorcery", ons1AkromasVengeanceOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"artifact": rock, "enchantment": aura, "creature": bear, "artifact creature": golem} {
		if g.Battlefield.Contains(id) {
			t.Errorf("the %s survived", name)
		}
	}
	if !g.Battlefield.Contains(land) {
		t.Error("a land was destroyed")
	}
}

func TestOns1AkromasVengeanceCyclesForThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.Ons1Cycle(t, g, "Akroma's Vengeance", "Sorcery", ons1AkromasVengeanceOracle, "{C}{C}{C}")
}

func TestOns1BarkhideMaulerCyclesForTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.Ons1Cycle(t, g, "Barkhide Mauler", "Creature — Beast", ons1BarkhideMaulerOracle, "{C}{C}")
}

func TestOns1AnuridMurkdiverHasSwampwalk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	diver := effectstest.B12Push(g, me.ID, "Anurid Murkdiver", "Creature — Zombie Frog Beast", ons1AnuridMurkdiverOracle, 4, 3)
	bear := effectstest.B12Creature(g, them.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	effectstest.AttackIntoBlocks(t, g, them.ID, diver)
	if r := effectstest.BlockRefusal(t, g, bear, diver); r != "" {
		t.Fatalf("blocked with no Swamp: refusal %q", r)
	}
	effectstest.SeedLand(g, them.ID, "Swamp", "Basic Land — Swamp", "")
	if r := effectstest.BlockRefusal(t, g, bear, diver); r != game.BlockReasonLandwalk {
		t.Errorf("refusal %q with a Swamp, want %q", r, game.BlockReasonLandwalk)
	}
}

func TestOns1AphettoAlchemistUntapsAnArtifactOrCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	alchemist := effectstest.B12Push(g, me.ID, "Aphetto Alchemist", "Creature — Human Wizard", ons1AphettoAlchemistOracle, 1, 2)
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Tapped Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID, Tapped: true,
	})
	land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Tapped Plains", TypeLine: "Basic Land — Plains", Owner: me.ID, Controller: me.ID, Tapped: true,
	})
	if err := g.ActivateCatalogAbility(me.ID, alchemist, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(land)}); err == nil {
		t.Fatal("the Alchemist targeted a land")
	}
	if err := g.ActivateCatalogAbility(me.ID, alchemist, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(rock)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TmtB1Card(t, g, rock).Tapped {
		t.Error("the artifact is still tapped")
	}
	if !effectstest.TmtB1Card(t, g, alchemist).Tapped {
		t.Error("the Alchemist did not tap to pay")
	}
}

func TestOns1AphettoAlchemistHasNoAbilityFaceDown(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.S36CastFaceDown(t, g, "Aphetto Alchemist", "Creature — Human Wizard", ons1AphettoAlchemistOracle, "morph", 1, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Tapped Bears", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID, Tapped: true,
	})
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Error("a face-down Alchemist activated its untap (CR 708.2a)")
	}
	effectstest.S36TurnUp(t, g, me, id, "U")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("activate face up: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TmtB1Card(t, g, bear).Tapped {
		t.Error("the creature is still tapped")
	}
}

func TestOns1AvenBrigadierPumpsEveryOtherBirdAndSoldier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	brigadier := effectstest.B12Push(g, me.ID, "Aven Brigadier", "Creature — Bird Soldier", ons1AvenBrigadierOracle, 3, 5)
	bird := effectstest.B12Creature(g, opp.ID, "Their Bird", "Creature — Bird", 1, 1)
	soldier := effectstest.B12Creature(g, me.ID, "My Soldier", "Creature — Human Soldier", 1, 1)
	both := effectstest.B12Creature(g, me.ID, "My Aven", "Creature — Bird Soldier", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "My Bears", "Creature — Bear", 2, 2)
	for _, tc := range []struct {
		name   string
		id     uuid.UUID
		pw, th int
	}{
		{"the Brigadier itself", brigadier, 3, 5},
		{"an opponent's Bird", bird, 2, 2},
		{"a Soldier", soldier, 2, 2},
		{"a Bird Soldier", both, 3, 3},
		{"a Bear", bear, 2, 2},
	} {
		if p, th := effectstest.MkmMPPT(t, g, tc.id); p != tc.pw || th != tc.th {
			t.Errorf("%s is %d/%d, want %d/%d", tc.name, p, th, tc.pw, tc.th)
		}
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, brigadier, "flying") {
		t.Error("the Brigadier has no flying")
	}
}

func TestOns1BatteringCraghornTurnsUpWithFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.S36CastFaceDown(t, g, "Battering Craghorn", "Creature — Goat Beast", ons1BatteringCraghornOracle, "morph", 3, 1)
	if effectstest.EffectiveAbilitiesContain(t, g, id, "first strike") {
		t.Error("the face-down Craghorn has first strike (CR 708.2a)")
	}
	if p, th := effectstest.MkmMPPT(t, g, id); p != 2 || th != 2 {
		t.Errorf("face down it is %d/%d, want 2/2", p, th)
	}
	effectstest.S36TurnUp(t, g, me, id, "R", "R", "C")
	if !effectstest.EffectiveAbilitiesContain(t, g, id, "first strike") {
		t.Error("first strike did not come back face up")
	}
	if p, th := effectstest.MkmMPPT(t, g, id); p != 3 || th != 1 {
		t.Errorf("face up it is %d/%d, want 3/1", p, th)
	}
}

func TestOns1BlisteringFirecatIsSacrificedAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	cat := effectstest.B12Push(g, me.ID, "Blistering Firecat", "Creature — Elemental Cat", ons1BlisteringFirecatOracle, 7, 1)
	for _, kw := range []string{"trample", "haste"} {
		if !effectstest.EffectiveAbilitiesContain(t, g, cat, kw) {
			t.Errorf("the Firecat has no %s", kw)
		}
	}
	effectstest.AdvanceToStepInTurn(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cat) {
		t.Error("the Firecat survived the end step")
	}
}

func TestOns1BlisteringFirecatFaceDownSurvivesTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.S36CastFaceDown(t, g, "Blistering Firecat", "Creature — Elemental Cat", ons1BlisteringFirecatOracle, "morph", 7, 1)
	effectstest.AdvanceToStepInTurn(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Error("a face-down Firecat was sacrificed (it has no text, CR 708.2a)")
	}
}

func TestOns1CabalArchonSacrificesAClericToDrainTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	archon := effectstest.B12Push(g, me.ID, "Cabal Archon", "Creature — Human Cleric", ons1CabalArchonOracle, 2, 2)
	cleric := effectstest.B12Creature(g, me.ID, "My Cleric", "Creature — Human Cleric", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "My Bears", "Creature — Bear", 2, 2)
	target := []game.TargetRef{effectstest.PlayerTarget(opp.ID)}
	if err := g.ActivateCatalogAbility(me.ID, archon, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{bear}, Targets: target,
	}); err == nil {
		t.Fatal("a non-Cleric paid the sacrifice")
	}
	myLife, theirLife := me.Life, opp.Life
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, archon, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{cleric}, Targets: target, Strict: true,
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(cleric) {
		t.Error("the Cleric was not sacrificed to pay")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirLife-2 || me.Life != myLife+2 {
		t.Errorf("life: target %d -> %d, me %d -> %d; want -2 and +2", theirLife, opp.Life, myLife, me.Life)
	}
}

func TestOns1CabalArchonMaySacrificeItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	archon := effectstest.B12Push(g, me.ID, "Cabal Archon", "Creature — Human Cleric", ons1CabalArchonOracle, 2, 2)
	theirLife := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, archon, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{archon}, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(archon) || opp.Life != theirLife-2 {
		t.Errorf("archon on battlefield=%v, target life %d -> %d", g.Battlefield.Contains(archon), theirLife, opp.Life)
	}
}

func TestOns1CentaurGladeMakesACentaur(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	glade := effectstest.B12Push(g, me.ID, "Centaur Glade", "Enchantment", ons1CentaurGladeOracle, 0, 0)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}{C}{C}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, glade, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with one green mana for {2}{G}{G}")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, glade, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountTokensControlled(g, me.ID, "Centaur"); n != 1 {
		t.Fatalf("%d Centaur tokens, want 1", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.HasSubtype("Centaur") && (c.CurrentPower() != 3 || c.CurrentToughness() != 3) {
			t.Errorf("the Centaur is %d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
		}
	}
}

func TestOns1ChargingSlatebackCantBlockFaceUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	bear := effectstest.B12Creature(g, me.ID, "My Bears", "Creature — Bear", 2, 2)
	slateback := effectstest.B12Push(g, them.ID, "Charging Slateback", "Creature — Beast", ons1ChargingSlatebackOracle, 4, 3)
	effectstest.AttackIntoBlocks(t, g, them.ID, bear)
	if r := effectstest.BlockRefusal(t, g, slateback, bear); r != game.BlockReasonCantBlock {
		t.Errorf("refusal %q, want %q", r, game.BlockReasonCantBlock)
	}
}

func TestOns1ChargingSlatebackCanBlockFaceDown(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	// The defender's morphed Slateback, face down on the battlefield.
	c := game.Card{
		InstanceID: uuid.New(), Name: "Charging Slateback", TypeLine: "Creature — Beast", OracleID: ons1ChargingSlatebackOracle,
		Power: 4, Toughness: 3, Owner: them.ID, Controller: them.ID, KnownBy: map[uuid.UUID]bool{them.ID: true},
	}
	c.SetFaceDown(game.FaceDownMorphed)
	id := effectstest.PushBattlefieldCardWithTimestamp(g, c)
	bear := effectstest.B12Creature(g, me.ID, "My Bears", "Creature — Bear", 2, 2)
	effectstest.AttackIntoBlocks(t, g, them.ID, bear)
	if r := effectstest.BlockRefusal(t, g, id, bear); r != "" {
		t.Errorf("a face-down Slateback could not block: %q (it has no text, CR 708.2a)", r)
	}
}

func TestOns1CraftyPathmageMakesASmallCreatureUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	mage := effectstest.B12Push(g, me.ID, "Crafty Pathmage", "Creature — Human Wizard", ons1CraftyPathmageOracle, 1, 1)
	small := effectstest.B12Creature(g, me.ID, "My Bears", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 3, 3)
	wall := effectstest.B12Creature(g, them.ID, "Their Wall", "Creature — Wall", 0, 4)
	if err := g.ActivateCatalogAbility(me.ID, mage, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(big)}); err == nil {
		t.Fatal("the Pathmage targeted a power-3 creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, mage, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(small)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AttackIntoBlocks(t, g, them.ID, small, big)
	if r := effectstest.BlockRefusal(t, g, wall, small); r != game.BlockReasonCantBeBlocked {
		t.Errorf("refusal %q for the targeted creature, want %q", r, game.BlockReasonCantBeBlocked)
	}
	if r := effectstest.BlockRefusal(t, g, wall, big); r != "" {
		t.Errorf("the other attacker could not be blocked: %q", r)
	}
}

func TestOns1CrudeRampartHasDefenderOnlyFaceUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.S36CastFaceDown(t, g, "Crude Rampart", "Creature — Wall", ons1CrudeRampartOracle, "morph", 4, 5)
	if effectstest.EffectiveAbilitiesContain(t, g, id, "defender") {
		t.Error("the face-down Rampart has defender (CR 708.2a)")
	}
	effectstest.S36TurnUp(t, g, me, id, "W", "C", "C", "C", "C")
	if !effectstest.EffectiveAbilitiesContain(t, g, id, "defender") {
		t.Error("defender did not come back face up")
	}
	if p, th := effectstest.MkmMPPT(t, g, id); p != 4 || th != 5 {
		t.Errorf("face up it is %d/%d, want 4/5", p, th)
	}
}
