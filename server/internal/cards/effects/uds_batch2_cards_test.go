package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// uds_batch2_cards_test.go — Urza's Destiny, batch 2 (#5621). Each test
// plays the real catalog entry and asserts what its text says.

const (
	udsB2AncientSilverbackOracle = "5f12134c-d626-4a5e-8653-700681a491fc"
	udsB2DiseaseCarriersOracle   = "ccdd6376-9b74-4d13-9f80-87d0bd3d4a2a"
	udsB2HulkingOgreOracle       = "a95aa9ba-fa61-4a3d-a493-1ea790bd20ad"
	udsB2MetathranSoldierOracle  = "b723fcec-f208-4397-a178-e1e47b729cf1"
	udsB2MaskOfLawAndGraceOracle = "21474195-5d91-466d-8874-4818c900dea7"
	udsB2PlatedSpiderOracle      = "15163c99-9388-4225-b707-4e057b2adfbc"
	udsB2GoliathBeetleOracle     = "86ab4400-fbdd-4c18-a893-441286a9d7d0"
	udsB2GoblinBerserkerOracle   = "5f71f205-9551-416a-bba6-55bc65d11c85"
	udsB2WildColosOracle         = "cb6b8ce3-9f9d-418c-94b0-c4469a254938"
	udsB2TormentedAngelOracle    = "176eeafd-0482-4fa4-9174-10f8b9494d9b"
	udsB2VoiceOfDutyOracle       = "b5d5b9c9-accc-4dbe-aaf4-c806f3e1c61c"
	udsB2VoiceOfReasonOracle     = "372e2d6e-427c-4152-b936-27dcf536ca84"
	udsB2ElvishLookoutOracle     = "9e9aea99-1eee-4bb6-9043-a1ac2884d071"
	udsB2SquirmingMassOracle     = "2c526c61-6cdb-440c-b4c8-b6d1503cef73"
	udsB2MarkerBeetlesOracle     = "4b1614fd-2810-4d4d-a0d6-934f27adbe65"
	udsB2PlagueDogsOracle        = "a7cc69fd-0c67-44fc-be9d-d8f5b1bd54f5"

	// udsB2BoomerangOracle is Boomerang, "Return target permanent to its
	// owner's hand.", the targeted spell the shroud test aims.
	udsB2BoomerangOracle = "dc4a4996-108a-4aac-850f-2d9f76403446"
)

func TestUdsB2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		udsB2AncientSilverbackOracle: "Ancient Silverback",
		udsB2DiseaseCarriersOracle:   "Disease Carriers",
		udsB2HulkingOgreOracle:       "Hulking Ogre",
		udsB2MetathranSoldierOracle:  "Metathran Soldier",
		udsB2MaskOfLawAndGraceOracle: "Mask of Law and Grace",
		udsB2PlatedSpiderOracle:      "Plated Spider",
		udsB2GoliathBeetleOracle:     "Goliath Beetle",
		udsB2GoblinBerserkerOracle:   "Goblin Berserker",
		udsB2WildColosOracle:         "Wild Colos",
		udsB2TormentedAngelOracle:    "Tormented Angel",
		udsB2VoiceOfDutyOracle:       "Voice of Duty",
		udsB2VoiceOfReasonOracle:     "Voice of Reason",
		udsB2ElvishLookoutOracle:     "Elvish Lookout",
		udsB2SquirmingMassOracle:     "Squirming Mass",
		udsB2MarkerBeetlesOracle:     "Marker Beetles",
		udsB2PlagueDogsOracle:        "Plague Dogs",
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

// --- Printed keywords ------------------------------------------------------------

func TestUdsB2PrintedKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Plated Spider", "Creature — Spider", udsB2PlatedSpiderOracle, []string{"reach"}},
		{"Goliath Beetle", "Creature — Insect", udsB2GoliathBeetleOracle, []string{"trample"}},
		{"Goblin Berserker", "Creature — Goblin Berserker", udsB2GoblinBerserkerOracle, []string{"first strike", "haste"}},
		{"Wild Colos", "Creature — Goat Beast", udsB2WildColosOracle, []string{"haste"}},
		{"Tormented Angel", "Creature — Angel", udsB2TormentedAngelOracle, []string{"flying"}},
		{"Voice of Duty", "Creature — Angel", udsB2VoiceOfDutyOracle, []string{"flying", game.ProtectionFromColor("G")}},
		{"Voice of Reason", "Creature — Angel", udsB2VoiceOfReasonOracle, []string{"flying", game.ProtectionFromColor("U")}},
		{"Elvish Lookout", "Creature — Elf", udsB2ElvishLookoutOracle, []string{"shroud"}},
		{"Squirming Mass", "Creature — Horror", udsB2SquirmingMassOracle, []string{"fear"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			assertKeywords(t, g, id, tc.keywords...)
		})
	}
}

// Haste (CR 702.10b): Wild Colos and Goblin Berserker attack the turn
// they arrive; a creature without it does not.
func TestUdsB2HasteCreaturesAttackTheTurnTheyArrive(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Wild Colos", "Creature — Goat Beast", udsB2WildColosOracle},
		{"Goblin Berserker", "Creature — Goblin Berserker", udsB2GoblinBerserkerOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			seat := g.Turn.ActiveSeat
			me, victim := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
			hasty := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
			advanceToStepOf(t, g, seat, game.StepDeclareAttackers)
			makeSick(t, g, hasty)
			makeSick(t, g, bear)
			if err := g.DeclareAttacker(bear, victim.ID); !errors.Is(err, game.ErrSummoningSick) {
				t.Errorf("control: a creature without haste attacked the turn it arrived: %v", err)
			}
			if err := g.DeclareAttacker(hasty, victim.ID); err != nil {
				t.Errorf("%s could not attack the turn it arrived: %v", tc.name, err)
			}
		})
	}
}

// --- Evasion: flying, reach and protection from a colour ----------------------------

func TestUdsB2FlyersReachAndProtectionInCombat(t *testing.T) {
	type blocker struct {
		name, oracle, color string
		want                game.BlockReason
	}
	for _, tc := range []struct {
		name, oracle, color string
		blockers            []blocker
	}{
		{"Tormented Angel", udsB2TormentedAngelOracle, "W", []blocker{
			{"Bear", "", "G", game.BlockReasonFlying},
			{"Plated Spider", udsB2PlatedSpiderOracle, "G", ""},
		}},
		{"Voice of Duty", udsB2VoiceOfDutyOracle, "W", []blocker{
			{"Plated Spider", udsB2PlatedSpiderOracle, "G", game.BlockReasonProtection},
			{"Tormented Angel", udsB2TormentedAngelOracle, "W", ""},
		}},
		{"Voice of Reason", udsB2VoiceOfReasonOracle, "W", []blocker{
			{"Kingfisher", udsB1KingfisherOracle, "U", game.BlockReasonProtection},
			{"Plated Spider", udsB2PlatedSpiderOracle, "G", ""},
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, them := seatsForCombat(g)
			attacker := exoB1Push(g, me.ID, tc.name, "Creature — Angel", tc.oracle, 2, 2, tc.color)
			ids := make([]uuid.UUID, len(tc.blockers))
			for i, b := range tc.blockers {
				ids[i] = exoB1Push(g, them.ID, b.name, "Creature — Beast", b.oracle, 2, 2, b.color)
			}
			attackIntoBlocks(t, g, them.ID, attacker)
			for i, b := range tc.blockers {
				if got := blockRefusal(t, g, ids[i], attacker); got != b.want {
					t.Errorf("%s blocking %s: refusal %q, want %q", b.name, tc.name, got, b.want)
				}
			}
		})
	}
}

// --- Squirming Mass ----------------------------------------------------------------

func TestUdsB2SquirmingMassHasFear(t *testing.T) {
	g := newCatalogGame(t)
	me, them := seatsForCombat(g)
	mass := exoB1Push(g, me.ID, "Squirming Mass", "Creature — Horror", udsB2SquirmingMassOracle, 1, 1, "B")
	green := exoB1Push(g, them.ID, "Green Bear", "Creature — Bear", "", 2, 2, "G")
	black := exoB1Push(g, them.ID, "Black Bear", "Creature — Bear", "", 2, 2, "B")
	golem := exoB1Push(g, them.ID, "Golem", "Artifact Creature — Golem", "", 2, 2)
	attackIntoBlocks(t, g, them.ID, mass)
	if got := blockRefusal(t, g, green, mass); got != game.BlockReasonFear {
		t.Errorf("a green creature blocking: refusal %q, want %q (CR 702.36b)", got, game.BlockReasonFear)
	}
	for name, id := range map[string]uuid.UUID{"a black creature": black, "an artifact creature": golem} {
		if got := blockRefusal(t, g, id, mass); got != "" {
			t.Errorf("%s was refused: %q", name, got)
		}
	}
}

// --- Metathran Soldier -------------------------------------------------------------

func TestUdsB2MetathranSoldierCantBeBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me, them := seatsForCombat(g)
	soldier := b12Push(g, me.ID, "Metathran Soldier", "Creature — Metathran Soldier", udsB2MetathranSoldierOracle, 1, 1)
	other := b12Creature(g, me.ID, "Raider", "Creature — Human", 2, 2)
	bear := b12Creature(g, them.ID, "Bear", "Creature — Bear", 2, 2)
	attackIntoBlocks(t, g, them.ID, soldier, other)
	if got := blockRefusal(t, g, bear, soldier); got != game.BlockReasonCantBeBlocked {
		t.Errorf("blocking Metathran Soldier: refusal %q, want %q", got, game.BlockReasonCantBeBlocked)
	}
	if got := blockRefusal(t, g, bear, other); got != "" {
		t.Errorf("control: blocking another attacker was refused: %q", got)
	}
}

// --- Hulking Ogre ------------------------------------------------------------------

func TestUdsB2HulkingOgreCantBlock(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, me.ID, "Raider", "Creature — Human", 2, 2)
	ogre := b12Push(g, opp.ID, "Hulking Ogre", "Creature — Ogre", udsB2HulkingOgreOracle, 3, 3)
	other := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	brAttack(t, g, attacker)
	brRefusal(t, g.DeclareBlocker(ogre, attacker), game.BlockReasonCantBlock)
	if err := g.DeclareBlocker(other, attacker); err != nil {
		t.Fatalf("another creature can't block: %v", err)
	}
}

// --- Elvish Lookout ----------------------------------------------------------------

func TestUdsB2ElvishLookoutCantBeTargeted(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	elf := b12Push(g, me.ID, "Elvish Lookout", "Creature — Elf", udsB2ElvishLookoutOracle, 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Boomerang", "Instant", udsB2BoomerangOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: elf}}); err == nil {
		t.Error("its own controller's spell targeted a creature with shroud (CR 702.18a)")
	}
	if err := castCatalogSpellErr(t, g, "Boomerang", "Instant", udsB2BoomerangOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}}); err != nil {
		t.Errorf("control: targeting a Bear was refused: %v", err)
	}
}

// --- Mask of Law and Grace -----------------------------------------------------------

func TestUdsB2MaskOfLawAndGraceGivesProtectionFromBlackAndRed(t *testing.T) {
	g := newCatalogGame(t)
	me, them := seatsForCombat(g)
	bear := exoB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2, "G")
	auraCast(t, g, "Mask of Law and Grace", udsB2MaskOfLawAndGraceOracle, bear)
	for _, kw := range []string{game.ProtectionFromColor("B"), game.ProtectionFromColor("R")} {
		if !hasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the enchanted Bear lacks %s", kw)
		}
	}
	black := exoB1Push(g, them.ID, "Black Bear", "Creature — Bear", "", 2, 2, "B")
	red := exoB1Push(g, them.ID, "Red Bear", "Creature — Bear", "", 2, 2, "R")
	white := exoB1Push(g, them.ID, "White Bear", "Creature — Bear", "", 2, 2, "W")
	attackIntoBlocks(t, g, them.ID, bear)
	for name, id := range map[string]uuid.UUID{"black": black, "red": red} {
		if got := blockRefusal(t, g, id, bear); got != game.BlockReasonProtection {
			t.Errorf("a %s creature blocking: refusal %q, want %q", name, got, game.BlockReasonProtection)
		}
	}
	if got := blockRefusal(t, g, white, bear); got != "" {
		t.Errorf("a white creature was refused: %q", got)
	}
}

// --- Ancient Silverback --------------------------------------------------------------

func TestUdsB2AncientSilverbackRegenerates(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	ape := b12Push(g, me.ID, "Ancient Silverback", "Creature — Ape", udsB2AncientSilverbackOracle, 6, 5)
	if err := usgB2Activate(t, g, me, ape, "B", nil); err == nil {
		t.Fatal("activated for {B}; it costs {G}")
	}
	if err := usgB2Activate(t, g, me, ape, "G", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := regenShieldsOn(t, g, ape); n != 1 {
		t.Fatalf("shields = %d, want 1", n)
	}
	destroyIt(t, g, ape)
	c, ok := battlefieldCard(g, ape)
	if !ok {
		t.Fatal("the shield did not replace the destruction")
	}
	if !c.Tapped {
		t.Error("a regenerated creature is tapped")
	}
}

// --- Disease Carriers, Marker Beetles ------------------------------------------------

func TestUdsB2DiesTriggersChangeTheirTarget(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness, dp   int
	}{
		{"Disease Carriers", "Creature — Rat", udsB2DiseaseCarriersOracle, 2, 2, -2},
		{"Marker Beetles", "Creature — Insect", udsB2MarkerBeetlesOracle, 2, 3, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := jmpB7Seats(g)
			dier := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			target := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
			other := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
			advanceToMain(t, g)
			g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(dier) })
			assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: target}, true)
			if p, tough := effectivePTOf(t, g, target); p != 3+tc.dp || tough != 3+tc.dp {
				t.Errorf("the target is %d/%d, want %d/%d", p, tough, 3+tc.dp, 3+tc.dp)
			}
			if p, tough := effectivePTOf(t, g, other); p != 2 || tough != 2 {
				t.Errorf("another creature is %d/%d, want 2/2", p, tough)
			}
		})
	}
}

func TestUdsB2DiseaseCarriersKillsATwoTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := jmpB7Seats(g)
	rats := b12Push(g, me.ID, "Disease Carriers", "Creature — Rat", udsB2DiseaseCarriersOracle, 2, 2)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	killCreature(t, g, me.ID, rats)
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: bear}, true)
	if g.Battlefield.Contains(bear) || !opp.Graveyard.Contains(bear) {
		t.Error("a 2/2 survived -2/-2 (CR 704.5f)")
	}
}

func TestUdsB2DiesTriggersDeclareTheirDirection(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		want         game.TargetPurpose
	}{
		{"Disease Carriers", udsB2DiseaseCarriersOracle, HarmUntilEOT(0, -2, -2)},
		{"Marker Beetles", udsB2MarkerBeetlesOracle, BenefitUntilEOT(0, 1, 1)},
	} {
		spec, _ := Lookup(tc.oracle)
		got := spec.Triggered[0].Purpose.Targets.List()
		if len(got) != 1 || got[0].UntilEndOfTurn != tc.want.UntilEndOfTurn || got[0].Pump == nil ||
			got[0].Pump.Power != tc.want.Pump.Power || got[0].Pump.Toughness != tc.want.Pump.Toughness {
			t.Errorf("%s declares %+v, want %+v", tc.name, got, tc.want)
		}
	}
}

// --- Marker Beetles, Plague Dogs: the sacrifice-to-draw row ------------------------

func TestUdsB2SacrificeToDrawRowsAlsoTriggerTheDiesAbility(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "A", "B")
	beetles := b12Push(g, me.ID, "Marker Beetles", "Creature — Insect", udsB2MarkerBeetlesOracle, 2, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	if err := usgB2Activate(t, g, me, beetles, "C", nil); err == nil {
		t.Fatal("activated for {1}; it costs {2}")
	}
	if err := usgB2Activate(t, g, me, beetles, "C", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(beetles) || !me.Graveyard.Contains(beetles) {
		t.Fatal("the Beetles were not sacrificed as the cost")
	}
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: bear}, true)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if p, tough := effectivePTOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the Bear is %d/%d, want 3/3 from the dies trigger", p, tough)
	}
}

// --- Plague Dogs -----------------------------------------------------------------------

func TestUdsB2PlagueDogsShrinksEveryCreatureWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := jmpB7Seats(g)
	dogs := b12Push(g, me.ID, "Plague Dogs", "Creature — Phyrexian Zombie Dog", udsB2PlagueDogsOracle, 3, 3)
	mine := b12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	theirs := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	b27Kill(g, dogs)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || !me.Graveyard.Contains(mine) {
		t.Error("a 1/1 survived -1/-1 (CR 704.5f)")
	}
	if p, tough := effectivePTOf(t, g, theirs); p != 1 || tough != 1 {
		t.Errorf("an opponent's 2/2 is %d/%d, want 1/1", p, tough)
	}
	late := b12Creature(g, opp.ID, "Late", "Creature — Bear", 2, 2)
	if p, tough := effectivePTOf(t, g, late); p != 2 || tough != 2 {
		t.Errorf("a creature that arrived after resolution is %d/%d, want 2/2 (CR 611.2c)", p, tough)
	}
}

func TestUdsB2PlagueDogsDeclaresItsSweep(t *testing.T) {
	spec, _ := Lookup(udsB2PlagueDogsOracle)
	want := game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1}
	if got := spec.Triggered[0].Purpose.Sweep; got != want {
		t.Errorf("Sweep %+v, want %+v", got, want)
	}
}

func TestUdsB2PlagueDogsDrawsWhenSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "A", "B")
	dogs := b12Push(g, me.ID, "Plague Dogs", "Creature — Phyrexian Zombie Dog", udsB2PlagueDogsOracle, 3, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	if err := usgB2Activate(t, g, me, dogs, "CC", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	assistCardSettle(t, g, me.ID, nil, true)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if p, tough := effectivePTOf(t, g, bear); p != 1 || tough != 1 {
		t.Errorf("the Bear is %d/%d, want 1/1 from the dies trigger", p, tough)
	}
}
