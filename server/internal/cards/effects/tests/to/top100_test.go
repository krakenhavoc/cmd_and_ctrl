package teststo

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- checklands --------------------------------------------------

// The bare board: nothing to check against, so the land is tapped.
func TestChecklandEntersTappedWithNothingToCheck(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.PlayLandFromHand(t, g, "Drowned Catacomb", effectstest.DrownedCatacombOracle)
	effectstest.Top100AssertEnteredTapped(t, g, id, "Drowned Catacomb")
}

// The check is on the land TYPE, not on a basic — a Watery Grave is
// an Island and a Swamp. Getting this wrong (narrowing to basics)
// would break the exact mana bases the cycle exists to serve.
func TestChecklandEntersUntappedOffANonbasicDual(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.SeedPermanentFor(g, me.ID, "Watery Grave", "Land — Island Swamp")

	id := effectstest.PlayLandFromHand(t, g, "Drowned Catacomb", effectstest.DrownedCatacombOracle)
	effectstest.Top100AssertEnteredUntapped(t, g, id, "Drowned Catacomb")
}

func TestChecklandEntersUntappedOffABasic(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.SeedPermanentFor(g, me.ID, "Swamp", "Basic Land — Swamp")

	id := effectstest.PlayLandFromHand(t, g, "Drowned Catacomb", effectstest.DrownedCatacombOracle)
	effectstest.Top100AssertEnteredUntapped(t, g, id, "Drowned Catacomb")
}

// "unless YOU control" — an opponent's Island does nothing. Reading
// the whole battlefield instead of the controller's share would make
// every checkland untapped at a four-player table.
func TestChecklandIgnoresOpponentsLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := effectstest.Top100Opponent(g)
	effectstest.SeedPermanentFor(g, opp.ID, "Island", "Basic Land — Island")

	id := effectstest.PlayLandFromHand(t, g, "Drowned Catacomb", effectstest.DrownedCatacombOracle)
	effectstest.Top100AssertEnteredTapped(t, g, id, "Drowned Catacomb")
}

// A Forest is neither an Island nor a Swamp; the wrong type must not
// satisfy the check.
func TestChecklandRejectsTheWrongLandType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.SeedPermanentFor(g, me.ID, "Forest", "Basic Land — Forest")

	id := effectstest.PlayLandFromHand(t, g, "Drowned Catacomb", effectstest.DrownedCatacombOracle)
	effectstest.Top100AssertEnteredTapped(t, g, id, "Drowned Catacomb")
}

// --- battle lands ------------------------------------------------

func TestBattleLandEntersTappedWithOneBasic(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.SeedPermanentFor(g, me.ID, "Island", "Basic Land — Island")

	id := effectstest.PlayLandFromHand(t, g, "Sunken Hollow", effectstest.SunkenHollowOracle)
	effectstest.Top100AssertEnteredTapped(t, g, id, "Sunken Hollow")
}

func TestBattleLandEntersUntappedWithTwoBasics(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.SeedPermanentFor(g, me.ID, "Island", "Basic Land — Island")
	effectstest.SeedPermanentFor(g, me.ID, "Swamp", "Basic Land — Swamp")

	id := effectstest.PlayLandFromHand(t, g, "Sunken Hollow", effectstest.SunkenHollowOracle)
	effectstest.Top100AssertEnteredUntapped(t, g, id, "Sunken Hollow")
}

// The battle lands count BASICS where the checklands count types.
// Two nonbasic duals carrying the right subtypes turn a checkland on
// and leave a battle land tapped — that inversion is the whole
// difference between the cycles.
func TestBattleLandCountsBasicsNotLandTypes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.SeedPermanentFor(g, me.ID, "Watery Grave", "Land — Island Swamp")
	effectstest.SeedPermanentFor(g, me.ID, "Underground Sea", "Land — Island Swamp")

	id := effectstest.PlayLandFromHand(t, g, "Sunken Hollow", effectstest.SunkenHollowOracle)
	effectstest.Top100AssertEnteredTapped(t, g, id, "Sunken Hollow")
}

// --- bond lands --------------------------------------------------

func TestBondLandEntersUntappedAtAFourPlayerTable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.PlayLandFromHand(t, g, "Morphic Pool", effectstest.MorphicPoolOracle)
	effectstest.Top100AssertEnteredUntapped(t, g, id, "Morphic Pool")
}

// One opponent is not two. The bond lands are unplayable in a duel
// and that is the printed card, not a bug.
func TestBondLandEntersTappedInADuel(t *testing.T) {
	g := effectstest.NewDuelCatalogGame(t)
	id := effectstest.PlayLandFromHand(t, g, "Morphic Pool", effectstest.MorphicPoolOracle)
	effectstest.Top100AssertEnteredTapped(t, g, id, "Morphic Pool")
}

// An eliminated seat is not an opponent (CR 800.4a), so a bond land
// played once the table is down to one live opponent enters tapped.
func TestBondLandIgnoresEliminatedSeats(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	killed := 0
	g.WithWriteLock(func() {
		for _, p := range g.Seats {
			if p == nil || p.ID == me.ID || killed >= 2 {
				continue
			}
			p.Eliminated = true
			killed++
		}
	})
	if killed != 2 {
		t.Fatalf("eliminated %d seats, want 2", killed)
	}

	id := effectstest.PlayLandFromHand(t, g, "Morphic Pool", effectstest.MorphicPoolOracle)
	effectstest.Top100AssertEnteredTapped(t, g, id, "Morphic Pool")
}

// --- the three cycles as data ------------------------------------

// The twenty lands are written as three loops over three tables, so
// the failure mode this guards against is a mistyped row or a leaked
// loop variable, not a bug in the logic. Distinct produced-mana
// strings per cycle is the loop-leak canary the Temple cycle uses.
func TestConditionalDualCyclesAreRegistered(t *testing.T) {
	for cycle, rows := range map[string]map[string]string{
		"checkland": {
			"Sulfur Falls":       "{U|R}",
			"Clifftop Retreat":   "{R|W}",
			"Dragonskull Summit": "{B|R}",
			"Isolated Chapel":    "{W|B}",
			"Glacial Fortress":   "{W|U}",
			"Hinterland Harbor":  "{G|U}",
			"Drowned Catacomb":   "{U|B}",
			"Woodland Cemetery":  "{B|G}",
			"Rootbound Crag":     "{R|G}",
			"Sunpetal Grove":     "{G|W}",
		},
		"battle land": {
			"Sunken Hollow":    "{U|B}",
			"Cinder Glade":     "{R|G}",
			"Smoldering Marsh": "{B|R}",
			"Canopy Vista":     "{G|W}",
			"Prairie Stream":   "{W|U}",
		},
		"bond land": {
			"Morphic Pool":         "{U|B}",
			"Rejuvenating Springs": "{G|U}",
			"Training Center":      "{U|R}",
			"Luxury Suite":         "{B|R}",
			"Sea of Clouds":        "{W|U}",
		},
	} {
		seen := map[string]string{}
		found := 0
		for _, spec := range effects.All() {
			want, ok := rows[spec.Name]
			if !ok {
				continue
			}
			found++
			if len(spec.Replacements) != 1 {
				t.Errorf("%s: %d replacements, want 1 (enters tapped unless …)",
					spec.Name, len(spec.Replacements))
			}
			if len(spec.ManaAbilities) != 1 {
				t.Errorf("%s: %d mana abilities, want 1", spec.Name, len(spec.ManaAbilities))
				continue
			}
			got := spec.ManaAbilities[0].Produced
			if got != want {
				t.Errorf("%s produces %s, want %s", spec.Name, got, want)
			}
			if prev, dupe := seen[got]; dupe {
				t.Errorf("%s produces %s, same as %s — the loop variable leaked",
					spec.Name, got, prev)
			}
			seen[got] = spec.Name
		}
		if found != len(rows) {
			t.Errorf("%s cycle: %d of %d registered", cycle, found, len(rows))
		}
	}
}

// --- Pongify -----------------------------------------------------

// The Ape goes to the creature's controller, not the caster — the
// same asymmetry Beast Within has, and the reason the card is a
// downside-carrying one-mana answer rather than a bomb.
func TestPongifyDestroysAndGivesVictimTheApe(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster := effectstest.Top100ActiveSeat(g)
	victim := effectstest.Top100Opponent(g)
	creature := effectstest.SeedPermanentFor(g, victim.ID, "Some Dragon", "Creature — Dragon")

	effectstest.CastCatalogSpell(t, g, "Pongify", "Instant", effectstest.PongifyOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: creature}})
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(creature) {
		t.Error("Pongify did not destroy its target")
	}
	if got := effectstest.CountBattlefieldNamed(g, victim.ID, "Ape"); got != 1 {
		t.Errorf("victim has %d Ape tokens, want 1", got)
	}
	if got := effectstest.CountBattlefieldNamed(g, caster.ID, "Ape"); got != 0 {
		t.Errorf("caster got %d Ape tokens, want 0 — the token is the drawback", got)
	}
}

// --- Abrade ------------------------------------------------------

func TestAbradeBurnModeDealsThreeToACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	victim := effectstest.Top100Opponent(g)
	creature := effectstest.SeedPermanentFor(g, victim.ID, "Grizzly Bears", "Creature — Bear")

	effectstest.CastModal(t, g, "Abrade", "Instant", effectstest.AbradeOracle, []int{0},
		[]game.TargetRef{{Kind: game.TargetCard, ID: creature}})
	effectstest.PassPriorityAroundTable(t, g)

	// The seeded creature is a 2/2, so 3 damage is lethal and SBAs
	// route it to the graveyard.
	if g.Battlefield.Contains(creature) {
		t.Error("Abrade's burn mode did not kill a 2/2")
	}
}

func TestAbradeArtifactModeDestroysAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	victim := effectstest.Top100Opponent(g)
	rock := effectstest.SeedPermanentFor(g, victim.ID, "Sol Ring", "Artifact")

	effectstest.CastModal(t, g, "Abrade", "Instant", effectstest.AbradeOracle, []int{1},
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(rock) {
		t.Error("Abrade's artifact mode did not destroy the artifact")
	}
	if !victim.Graveyard.Contains(rock) {
		t.Error("destroyed artifact did not reach its owner's graveyard")
	}
}

// The chosen mode's target clause gates the cast: the burn bullet
// cannot be pointed at an artifact.
func TestAbradeBurnModeRejectsAnArtifactTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	victim := effectstest.Top100Opponent(g)
	rock := effectstest.SeedPermanentFor(g, victim.ID, "Sol Ring", "Artifact")

	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Abrade", TypeLine: "Instant", OracleID: effectstest.AbradeOracle,
		Owner: me.ID, Controller: me.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Modes:   []int{0},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}},
	})
	if err == nil {
		t.Error("Abrade's burn mode accepted an artifact target")
	}
}

// --- Feed the Swarm ----------------------------------------------

func TestFeedTheSwarmDestroysAndChargesManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	victim := effectstest.Top100Opponent(g)
	// {3}{G}{G} is mana value 5.
	target := effectstest.PushCostedPermanentForTest(g, victim.ID, "Big Thing", "Creature — Beast", "{3}{G}{G}")
	before := me.Life

	effectstest.CastCatalogSpell(t, g, "Feed the Swarm", "Sorcery", effectstest.FeedTheSwarmOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: target}})
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(target) {
		t.Error("Feed the Swarm did not destroy its target")
	}
	if me.Life != before-5 {
		t.Errorf("life %d -> %d, want %d — the cost is the target's mana value",
			before, me.Life, before-5)
	}
}

// The enchantment half is the whole reason mono-black plays this.
func TestFeedTheSwarmKillsAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	victim := effectstest.Top100Opponent(g)
	target := effectstest.PushCostedPermanentForTest(g, victim.ID, "Ghostly Prison", "Enchantment", "{2}{W}")

	effectstest.CastCatalogSpell(t, g, "Feed the Swarm", "Sorcery", effectstest.FeedTheSwarmOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: target}})
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(target) {
		t.Error("Feed the Swarm did not destroy the enchantment")
	}
}

// --- Assassin's Trophy -------------------------------------------

func TestAssassinsTrophyDestroysAndReplacesWithAnUntappedBasic(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	victim := effectstest.Top100Opponent(g)
	target := effectstest.SeedPermanentFor(g, victim.ID, "Gaea's Cradle", "Legendary Land")
	basic := effectstest.StapleLibraryCard(victim, "Forest", "Basic Land — Forest")

	effectstest.CastCatalogSpell(t, g, "Assassin's Trophy", "Instant", effectstest.AssassinsTrophyOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: target}})
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(target) {
		t.Error("Assassin's Trophy did not destroy its target")
	}
	// S22: the "may" is real, so the VICTIM is prompted and here
	// accepts the replacement land.
	effectstest.AnswerSearchByID(t, g, victim.ID, basic)
	got, ok := effectstest.BattlefieldCard(g, basic)
	if !ok {
		t.Fatal("the victim did not get their replacement basic")
	}
	if got.Tapped {
		t.Error("the replacement land arrives UNTAPPED — that is the card's drawback")
	}
	if got.Controller != victim.ID {
		t.Errorf("the replacement land is controlled by %v, want the victim %v",
			got.Controller, victim.ID)
	}
}

// --- the free spells ---------------------------------------------

func TestDeadlyRollickExilesTargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	victim := effectstest.Top100Opponent(g)
	creature := effectstest.SeedPermanentFor(g, victim.ID, "Some Commander", "Legendary Creature — Human")

	effectstest.CastCatalogSpell(t, g, "Deadly Rollick", "Instant", effectstest.DeadlyRollickOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: creature}})
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(creature) {
		t.Error("Deadly Rollick did not remove its target")
	}
	if !g.Exile.Contains(creature) {
		t.Error("Deadly Rollick EXILES — a graveyard is a different card")
	}
}

func TestFierceGuardianshipCountersANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	opp := effectstest.Top100Opponent(g)
	before := me.Life

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	boltID := uuid.New()
	opp.Hand.PushTop(game.Card{
		InstanceID: boltID, Name: "Lightning Bolt", TypeLine: "Instant",
		OracleID: "4457ed35-7c10-48c8-9776-456485fdf070",
		Owner:    opp.ID, Controller: opp.ID,
	})
	if err := g.CastSpell(opp.ID, boltID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	}); err != nil {
		t.Fatalf("opponent CastSpell Bolt: %v", err)
	}

	guardID := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: guardID, Name: "Fierce Guardianship", TypeLine: "Instant",
		OracleID: effectstest.FierceGuardianshipOracle, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, guardID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: boltID}},
	}); err != nil {
		t.Fatalf("CastSpell Fierce Guardianship: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if !opp.Graveyard.Contains(boltID) {
		t.Error("the countered Bolt is not in its owner's graveyard")
	}
	if me.Life != before {
		t.Errorf("life %d -> %d: the countered Bolt still dealt damage", before, me.Life)
	}
}

// --- Commander's Sphere ------------------------------------------

func TestCommandersSphereSacrificesToDrawACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	sphere := effectstest.PushCatalogPermanent(g, me.ID, "Commander's Sphere", "Artifact",
		effectstest.CommandersSphereOracle, false)
	before := me.Hand.Size()

	if err := g.ActivateCatalogAbility(me.ID, sphere, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	// The sacrifice is a cost, paid at announce.
	if g.Battlefield.Contains(sphere) {
		t.Error("the Sphere should be sacrificed as a cost, at announce")
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d -> %d, want one more card", before, got)
	}
}

// The mana half is the identity-narrowed pipe, not a fixed colour —
// asserting the shape catches a copy-paste from Sol Ring.
func TestCommandersSphereTapsForCommanderIdentity(t *testing.T) {
	abs := game.ManaAbilitiesForCard(game.Card{OracleID: effectstest.CommandersSphereOracle})
	if len(abs) != 1 {
		t.Fatalf("%d mana abilities, want 1", len(abs))
	}
	if abs[0].Produced != "{W|U|B|R|G}" {
		t.Errorf("produced %q, want the full pipe set for identity narrowing", abs[0].Produced)
	}
	if !abs[0].TapCost || abs[0].SacrificeCost {
		t.Errorf("cost %+v, want tap only — the sacrifice is the OTHER ability", abs[0])
	}
}

// --- Path of Ancestry --------------------------------------------

func TestPathOfAncestryEntersTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.PlayLandFromHand(t, g, "Path of Ancestry", effectstest.PathOfAncestryOracle)
	effectstest.Top100AssertEnteredTapped(t, g, id, "Path of Ancestry")
}

func TestPathOfAncestryTapsForCommanderIdentity(t *testing.T) {
	abs := game.ManaAbilitiesForCard(game.Card{OracleID: effectstest.PathOfAncestryOracle})
	if len(abs) != 1 {
		t.Fatalf("%d mana abilities, want 1", len(abs))
	}
	if abs[0].Produced != "{W|U|B|R|G}" {
		t.Errorf("produced %q, want the full pipe set for identity narrowing", abs[0].Produced)
	}
}

// top100SeedCreature puts a creature of a given size on the
// battlefield. seedPermanentFor hardcodes 2/2, and the whole subject
// here is the power-4 threshold.
func top100SeedCreature(g *game.Game, owner uuid.UUID, name string, power, toughness int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Beast",
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
	return id
}

func TestGarruksUprisingDrawsOnEntryWithABigCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	top100SeedCreature(g, me.ID, "Big Thing", 5, 5)
	before := me.Hand.Size()

	effectstest.CastCatalogSpell(t, g, "Garruk's Uprising", "Enchantment", effectstest.GarruksUprisingOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d -> %d, want one more — the intervening-if was satisfied", before, got)
	}
}

// No power-4 creature, no card. The intervening-if is the entire
// difference between this and a three-mana cantrip.
func TestGarruksUprisingDrawsNothingWithoutABigCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	top100SeedCreature(g, me.ID, "Small Thing", 3, 3)
	before := me.Hand.Size()

	effectstest.CastCatalogSpell(t, g, "Garruk's Uprising", "Enchantment", effectstest.GarruksUprisingOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != before {
		t.Errorf("hand %d -> %d, want unchanged — no creature with power 4 or greater", before, got)
	}
}

func TestGarruksUprisingDrawsWhenABigCreatureArrives(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)

	effectstest.CastCatalogSpell(t, g, "Garruk's Uprising", "Enchantment", effectstest.GarruksUprisingOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	before := me.Hand.Size()

	effectstest.Top100CastCreature(t, g, me, "Big Thing", 5, 5)
	effectstest.PassPriorityAroundTable(t, g)

	// top100CastCreature seeds the spell into hand and immediately
	// casts it, so the cast itself is hand-neutral. The whole delta
	// is the trigger's draw.
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d -> %d after a 5/5 entered: want one more (the trigger's draw)", before, got)
	}
}

// A 2/2 arriving must NOT draw — the power gate is the card.
func TestGarruksUprisingIgnoresSmallCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)

	effectstest.CastCatalogSpell(t, g, "Garruk's Uprising", "Enchantment", effectstest.GarruksUprisingOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	before := me.Hand.Size()

	effectstest.Top100CastCreature(t, g, me, "Small Thing", 3, 3)
	effectstest.PassPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != before {
		t.Errorf("hand %d -> %d after a 3/3 entered: want unchanged (below the power-4 gate)", before, got)
	}
}

func TestGarruksUprisingGrantsTrampleToYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	opp := effectstest.Top100Opponent(g)
	mine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Mine", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Theirs", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Garruk's Uprising", TypeLine: "Enchantment",
		OracleID: effectstest.GarruksUprisingOracle, Owner: me.ID, Controller: me.ID,
	})

	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, mine), "trample") {
		t.Errorf("your creature has no trample: %v", effectstest.EffectiveAbilities(t, g, mine))
	}
	if effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, theirs), "trample") {
		t.Error("the grant is to CREATURES YOU CONTROL; an opponent's creature got trample")
	}
}

// --- Reanimate ---------------------------------------------------

func TestReanimateReturnsACreatureAndChargesItsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: dead, Name: "Big Dead Thing", TypeLine: "Creature — Wurm",
		ManaCost: "{4}{B}{B}", Power: 6, Toughness: 6,
		Owner: me.ID, Controller: me.ID,
	})
	before := me.Life

	effectstest.CastCatalogSpell(t, g, "Reanimate", "Sorcery", effectstest.ReanimateOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: dead}})
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(dead) {
		t.Fatal("Reanimate did not put the creature onto the battlefield")
	}
	if me.Graveyard.Contains(dead) {
		t.Error("the reanimated card is still in the graveyard")
	}
	if me.Life != before-6 {
		t.Errorf("life %d -> %d, want %d — the cost is the card's mana value",
			before, me.Life, before-6)
	}
}
func TestFabledPassageFetchesTappedBelowFourLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	passage := effectstest.PushCatalogPermanent(g, me.ID, "Fabled Passage", "Land", effectstest.FabledPassageOracle, false)
	basic := effectstest.StapleLibraryCard(me, "Forest", "Basic Land — Forest")

	if err := g.ActivateCatalogAbility(me.ID, passage, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(passage) {
		t.Error("the Passage is sacrificed as a cost, at announce")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)

	got, ok := effectstest.BattlefieldCard(g, basic)
	if !ok {
		t.Fatal("Fabled Passage did not fetch the basic")
	}
	if !got.Tapped {
		t.Error("with fewer than four lands the fetched land stays TAPPED")
	}
}

func TestFabledPassageUntapsAtFourLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	// Three lands plus the Passage is three after the sacrifice; the
	// fetched land makes four.
	effectstest.SeedPermanentFor(g, me.ID, "Forest A", "Basic Land — Forest")
	effectstest.SeedPermanentFor(g, me.ID, "Forest B", "Basic Land — Forest")
	effectstest.SeedPermanentFor(g, me.ID, "Forest C", "Basic Land — Forest")
	passage := effectstest.PushCatalogPermanent(g, me.ID, "Fabled Passage", "Land", effectstest.FabledPassageOracle, false)
	basic := effectstest.StapleLibraryCard(me, "Island", "Basic Land — Island")

	if err := g.ActivateCatalogAbility(me.ID, passage, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)

	got, ok := effectstest.BattlefieldCard(g, basic)
	if !ok {
		t.Fatal("Fabled Passage did not fetch the basic")
	}
	if got.Tapped {
		t.Error("with four or more lands the fetched land is untapped")
	}
}

// The cost is tap + sacrifice and nothing else — no life, unlike the
// Zendikar fetchlands. A stray component would make the card worse
// than printed; a missing one, better.
func TestFabledPassageCostIsTapAndSacrificeOnly(t *testing.T) {
	abs := game.ActivatedAbilitiesForCard(game.Card{OracleID: effectstest.FabledPassageOracle})
	if len(abs) != 1 {
		t.Fatalf("%d activated abilities, want 1", len(abs))
	}
	c := abs[0].Cost
	if !c.Tap || !c.SacrificeSelf {
		t.Errorf("cost %+v, want tap + sacrifice self", c)
	}
	if c.Life != 0 || c.Mana != "" || c.SacrificeOther != nil {
		t.Errorf("cost %+v has an extra component", c)
	}
}

// --- Myriad Landscape --------------------------------------------

func TestMyriadLandscapeEntersTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.PlayLandFromHand(t, g, "Myriad Landscape", effectstest.MyriadLandscapeOracle)
	effectstest.Top100AssertEnteredTapped(t, g, id, "Myriad Landscape")
}

// Two basics that SHARE a type, both tapped. S22: the controller
// picks the pair, and the engine enforces the share-a-type clause on
// the pair rather than deriving a type from library order.
func TestMyriadLandscapeFetchesTwoBasicsOfOneType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	landscape := effectstest.PushCatalogPermanent(g, me.ID, "Myriad Landscape", "Land", effectstest.MyriadLandscapeOracle, false)
	forest := effectstest.StapleLibraryCard(me, "Forest", "Basic Land — Forest")
	plainsA := effectstest.StapleLibraryCard(me, "Plains", "Basic Land — Plains")
	plainsB := effectstest.StapleLibraryCard(me, "Plains", "Basic Land — Plains")

	if err := g.ActivateCatalogAbility(me.ID, landscape, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	// A mismatched pair is refused — that is the whole clause.
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("Myriad Landscape did not prompt for a pick")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{forest, plainsA}); err == nil {
		t.Error("a Forest + Plains pair was accepted; they share no land type")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, plainsA, plainsB)

	for _, id := range []uuid.UUID{plainsA, plainsB} {
		got, ok := effectstest.BattlefieldCard(g, id)
		if !ok {
			t.Fatal("a Plains was not fetched")
		}
		if !got.Tapped {
			t.Error("Myriad Landscape's lands arrive TAPPED")
		}
	}
	if _, wrong := effectstest.BattlefieldCard(g, forest); wrong {
		t.Error("the Forest was fetched — the two lands must share a land type")
	}
}

// A whiffed search still costs the land. Nothing arrives, nothing
// crashes.
func TestMyriadLandscapeWithNoBasicsFetchesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	landscape := effectstest.PushCatalogPermanent(g, me.ID, "Myriad Landscape", "Land", effectstest.MyriadLandscapeOracle, false)
	dual := effectstest.StapleLibraryCard(me, "Watery Grave", "Land — Island Swamp")

	if err := g.ActivateCatalogAbility(me.ID, landscape, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(landscape) {
		t.Error("the Landscape is sacrificed as a cost, at announce")
	}
	if _, wrong := effectstest.BattlefieldCard(g, dual); wrong {
		t.Error("a nonbasic was fetched — the predicate is basic-land only")
	}
}

func TestMyriadLandscapeCostIsTwoManaTapSacrifice(t *testing.T) {
	abs := game.ActivatedAbilitiesForCard(game.Card{OracleID: effectstest.MyriadLandscapeOracle})
	if len(abs) != 1 {
		t.Fatalf("%d activated abilities, want 1", len(abs))
	}
	c := abs[0].Cost
	if !c.Tap || !c.SacrificeSelf || c.Mana != "{2}" {
		t.Errorf("cost %+v, want {2} + tap + sacrifice self", c)
	}
	if c.Life != 0 {
		t.Errorf("cost %+v pays life; Myriad Landscape does not", c)
	}
}
