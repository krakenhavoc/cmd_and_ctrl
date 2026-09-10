package effects

import (
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// top100_test.go — the play-rate batch catalogued in
// docs/decklists/top-100-commander-staples.md: the twenty
// conditional-dual lands and the eleven singles that needed no new
// machinery.
//
// The assertions aim at what each card is FOR, not at "it resolved".
// A checkland that always enters untapped would pass a "did it enter"
// test and be a strictly better card than the one printed, so every
// conditional-dual test pins BOTH sides of its condition.

const (
	drownedCatacombOracle    = "819fc966-434e-470f-91e9-a38df974ad17"
	sunkenHollowOracle       = "cd2c90ac-2b04-461c-92f3-939871b6b6a3"
	morphicPoolOracle        = "bd004c9d-771e-4e63-a97d-a2259c096af8"
	pongifyOracle            = "05849bd6-8f38-4031-be2b-e2aa03beb8cc"
	abradeOracle             = "f9db72dc-9a5b-48a4-a86e-7464d9a2166a"
	feedTheSwarmOracle       = "5825997b-10d7-4a36-972c-a80ddd90b8ed"
	assassinsTrophyOracle    = "ac10d218-f9a6-4058-9cda-a15ca1b0b7b5"
	deadlyRollickOracle      = "0456ec64-2c81-4763-a352-8ff64a4c3d6b"
	fierceGuardianshipOracle = "d09c9cba-fdd2-479b-ad5d-d05181c3e3f9"
	commandersSphereOracle   = "0b67c4e2-f88b-4e01-85a1-9d5f5b8db13b"
	pathOfAncestryOracle     = "b473e293-59e3-4e04-acf2-622604aeb25f"
	garruksUprisingOracle    = "3127ae9b-a7a7-43ec-89d7-688f8445b33d"
	reanimateOracle          = "a044474a-cd72-4e9d-bd8d-a08f2de9cdc0"
	fabledPassageOracle      = "0c85b8f7-0bd0-4680-9ec5-d4b110460a54"
	myriadLandscapeOracle    = "2549bc57-9ffb-4053-9f10-f2a5f792b845"
)

// newDuelCatalogGame is newCatalogGame with two seats. The bond lands
// are the only cards in the catalog whose behaviour depends on the
// number of opponents, and "unless you have two or more opponents" is
// only observable when the table is small enough for it to be false.
func newDuelCatalogGame(t *testing.T) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i := 0; i < 2; i++ {
		deck := make([]game.Card, 20)
		for j := range deck {
			deck[j] = game.NewCard("basic-filler", uuid.Nil)
		}
		if _, err := g.AddPlayer(string(rune('A'+i)), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(11, 22))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	return g
}

// top100ActiveSeat returns the seat whose turn it is — the one
// playLandFromHand / castCatalogSpell act as.
func top100ActiveSeat(g *game.Game) *game.Player {
	return g.Seats[g.Turn.ActiveSeat]
}

// top100Opponent returns some seat other than the active one.
func top100Opponent(g *game.Game) *game.Player {
	me := top100ActiveSeat(g)
	for _, p := range g.Seats {
		if p != nil && p.ID != me.ID {
			return p
		}
	}
	return nil
}

// top100AssertEnteredTapped pins the enters-tapped REPLACEMENT rather
// than merely a tapped permanent: an OnETB tap would leave the card
// tapped too, but it would emit a tap event because the permanent
// really was untapped on the battlefield for a beat.
func top100AssertEnteredTapped(t *testing.T, g *game.Game, id uuid.UUID, name string) {
	t.Helper()
	card, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s isn't on the battlefield", name)
	}
	if !card.Tapped {
		t.Errorf("%s entered untapped; its condition was not met, so it must enter tapped", name)
	}
	if n := tapEventsFor(g, id); n != 0 {
		t.Errorf("%s: %d tap events — it should have ENTERED tapped, not been tapped after", name, n)
	}
}

func top100AssertEnteredUntapped(t *testing.T, g *game.Game, id uuid.UUID, name string) {
	t.Helper()
	card, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s isn't on the battlefield", name)
	}
	if card.Tapped {
		t.Errorf("%s entered tapped; its condition was met, so it must enter untapped", name)
	}
}

// --- checklands --------------------------------------------------

// The bare board: nothing to check against, so the land is tapped.
func TestChecklandEntersTappedWithNothingToCheck(t *testing.T) {
	g := newCatalogGame(t)
	id := playLandFromHand(t, g, "Drowned Catacomb", drownedCatacombOracle)
	top100AssertEnteredTapped(t, g, id, "Drowned Catacomb")
}

// The check is on the land TYPE, not on a basic — a Watery Grave is
// an Island and a Swamp. Getting this wrong (narrowing to basics)
// would break the exact mana bases the cycle exists to serve.
func TestChecklandEntersUntappedOffANonbasicDual(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	seedPermanentFor(g, me.ID, "Watery Grave", "Land — Island Swamp")

	id := playLandFromHand(t, g, "Drowned Catacomb", drownedCatacombOracle)
	top100AssertEnteredUntapped(t, g, id, "Drowned Catacomb")
}

func TestChecklandEntersUntappedOffABasic(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	seedPermanentFor(g, me.ID, "Swamp", "Basic Land — Swamp")

	id := playLandFromHand(t, g, "Drowned Catacomb", drownedCatacombOracle)
	top100AssertEnteredUntapped(t, g, id, "Drowned Catacomb")
}

// "unless YOU control" — an opponent's Island does nothing. Reading
// the whole battlefield instead of the controller's share would make
// every checkland untapped at a four-player table.
func TestChecklandIgnoresOpponentsLands(t *testing.T) {
	g := newCatalogGame(t)
	opp := top100Opponent(g)
	seedPermanentFor(g, opp.ID, "Island", "Basic Land — Island")

	id := playLandFromHand(t, g, "Drowned Catacomb", drownedCatacombOracle)
	top100AssertEnteredTapped(t, g, id, "Drowned Catacomb")
}

// A Forest is neither an Island nor a Swamp; the wrong type must not
// satisfy the check.
func TestChecklandRejectsTheWrongLandType(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	seedPermanentFor(g, me.ID, "Forest", "Basic Land — Forest")

	id := playLandFromHand(t, g, "Drowned Catacomb", drownedCatacombOracle)
	top100AssertEnteredTapped(t, g, id, "Drowned Catacomb")
}

// --- battle lands ------------------------------------------------

func TestBattleLandEntersTappedWithOneBasic(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	seedPermanentFor(g, me.ID, "Island", "Basic Land — Island")

	id := playLandFromHand(t, g, "Sunken Hollow", sunkenHollowOracle)
	top100AssertEnteredTapped(t, g, id, "Sunken Hollow")
}

func TestBattleLandEntersUntappedWithTwoBasics(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	seedPermanentFor(g, me.ID, "Island", "Basic Land — Island")
	seedPermanentFor(g, me.ID, "Swamp", "Basic Land — Swamp")

	id := playLandFromHand(t, g, "Sunken Hollow", sunkenHollowOracle)
	top100AssertEnteredUntapped(t, g, id, "Sunken Hollow")
}

// The battle lands count BASICS where the checklands count types.
// Two nonbasic duals carrying the right subtypes turn a checkland on
// and leave a battle land tapped — that inversion is the whole
// difference between the cycles.
func TestBattleLandCountsBasicsNotLandTypes(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	seedPermanentFor(g, me.ID, "Watery Grave", "Land — Island Swamp")
	seedPermanentFor(g, me.ID, "Underground Sea", "Land — Island Swamp")

	id := playLandFromHand(t, g, "Sunken Hollow", sunkenHollowOracle)
	top100AssertEnteredTapped(t, g, id, "Sunken Hollow")
}

// --- bond lands --------------------------------------------------

func TestBondLandEntersUntappedAtAFourPlayerTable(t *testing.T) {
	g := newCatalogGame(t)
	id := playLandFromHand(t, g, "Morphic Pool", morphicPoolOracle)
	top100AssertEnteredUntapped(t, g, id, "Morphic Pool")
}

// One opponent is not two. The bond lands are unplayable in a duel
// and that is the printed card, not a bug.
func TestBondLandEntersTappedInADuel(t *testing.T) {
	g := newDuelCatalogGame(t)
	id := playLandFromHand(t, g, "Morphic Pool", morphicPoolOracle)
	top100AssertEnteredTapped(t, g, id, "Morphic Pool")
}

// An eliminated seat is not an opponent (CR 800.4a), so a bond land
// played once the table is down to one live opponent enters tapped.
func TestBondLandIgnoresEliminatedSeats(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
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

	id := playLandFromHand(t, g, "Morphic Pool", morphicPoolOracle)
	top100AssertEnteredTapped(t, g, id, "Morphic Pool")
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
		for _, spec := range All() {
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
	g := newCatalogGame(t)
	caster := top100ActiveSeat(g)
	victim := top100Opponent(g)
	creature := seedPermanentFor(g, victim.ID, "Some Dragon", "Creature — Dragon")

	castCatalogSpell(t, g, "Pongify", "Instant", pongifyOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: creature}})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(creature) {
		t.Error("Pongify did not destroy its target")
	}
	if got := countBattlefieldNamed(g, victim.ID, "Ape"); got != 1 {
		t.Errorf("victim has %d Ape tokens, want 1", got)
	}
	if got := countBattlefieldNamed(g, caster.ID, "Ape"); got != 0 {
		t.Errorf("caster got %d Ape tokens, want 0 — the token is the drawback", got)
	}
}

// --- Abrade ------------------------------------------------------

func TestAbradeBurnModeDealsThreeToACreature(t *testing.T) {
	g := newCatalogGame(t)
	victim := top100Opponent(g)
	creature := seedPermanentFor(g, victim.ID, "Grizzly Bears", "Creature — Bear")

	castModal(t, g, "Abrade", "Instant", abradeOracle, []int{0},
		[]game.TargetRef{{Kind: game.TargetCard, ID: creature}})
	passPriorityAroundTable(t, g)

	// The seeded creature is a 2/2, so 3 damage is lethal and SBAs
	// route it to the graveyard.
	if g.Battlefield.Contains(creature) {
		t.Error("Abrade's burn mode did not kill a 2/2")
	}
}

func TestAbradeArtifactModeDestroysAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	victim := top100Opponent(g)
	rock := seedPermanentFor(g, victim.ID, "Sol Ring", "Artifact")

	castModal(t, g, "Abrade", "Instant", abradeOracle, []int{1},
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	passPriorityAroundTable(t, g)

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
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	victim := top100Opponent(g)
	rock := seedPermanentFor(g, victim.ID, "Sol Ring", "Artifact")

	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Abrade", TypeLine: "Instant", OracleID: abradeOracle,
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
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	victim := top100Opponent(g)
	// {3}{G}{G} is mana value 5.
	target := pushCostedPermanentForTest(g, victim.ID, "Big Thing", "Creature — Beast", "{3}{G}{G}")
	before := me.Life

	castCatalogSpell(t, g, "Feed the Swarm", "Sorcery", feedTheSwarmOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: target}})
	passPriorityAroundTable(t, g)

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
	g := newCatalogGame(t)
	victim := top100Opponent(g)
	target := pushCostedPermanentForTest(g, victim.ID, "Ghostly Prison", "Enchantment", "{2}{W}")

	castCatalogSpell(t, g, "Feed the Swarm", "Sorcery", feedTheSwarmOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: target}})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(target) {
		t.Error("Feed the Swarm did not destroy the enchantment")
	}
}

// --- Assassin's Trophy -------------------------------------------

func TestAssassinsTrophyDestroysAndReplacesWithAnUntappedBasic(t *testing.T) {
	g := newCatalogGame(t)
	victim := top100Opponent(g)
	target := seedPermanentFor(g, victim.ID, "Gaea's Cradle", "Legendary Land")
	basic := stapleLibraryCard(victim, "Forest", "Basic Land — Forest")

	castCatalogSpell(t, g, "Assassin's Trophy", "Instant", assassinsTrophyOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: target}})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(target) {
		t.Error("Assassin's Trophy did not destroy its target")
	}
	got, ok := battlefieldCard(g, basic)
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
	g := newCatalogGame(t)
	victim := top100Opponent(g)
	creature := seedPermanentFor(g, victim.ID, "Some Commander", "Legendary Creature — Human")

	castCatalogSpell(t, g, "Deadly Rollick", "Instant", deadlyRollickOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: creature}})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(creature) {
		t.Error("Deadly Rollick did not remove its target")
	}
	if !g.Exile.Contains(creature) {
		t.Error("Deadly Rollick EXILES — a graveyard is a different card")
	}
}

func TestFierceGuardianshipCountersANoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	opp := top100Opponent(g)
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
		OracleID: fierceGuardianshipOracle, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, guardID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: boltID}},
	}); err != nil {
		t.Fatalf("CastSpell Fierce Guardianship: %v", err)
	}
	passPriorityAroundTable(t, g)

	if !opp.Graveyard.Contains(boltID) {
		t.Error("the countered Bolt is not in its owner's graveyard")
	}
	if me.Life != before {
		t.Errorf("life %d -> %d: the countered Bolt still dealt damage", before, me.Life)
	}
}

// --- Commander's Sphere ------------------------------------------

func TestCommandersSphereSacrificesToDrawACard(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	sphere := pushCatalogPermanent(g, me.ID, "Commander's Sphere", "Artifact",
		commandersSphereOracle, false)
	before := me.Hand.Size()

	if err := g.ActivateCatalogAbility(me.ID, sphere, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	// The sacrifice is a cost, paid at announce.
	if g.Battlefield.Contains(sphere) {
		t.Error("the Sphere should be sacrificed as a cost, at announce")
	}
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d -> %d, want one more card", before, got)
	}
}

// The mana half is the identity-narrowed pipe, not a fixed colour —
// asserting the shape catches a copy-paste from Sol Ring.
func TestCommandersSphereTapsForCommanderIdentity(t *testing.T) {
	abs := game.ManaAbilitiesForCard(game.Card{OracleID: commandersSphereOracle})
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
	g := newCatalogGame(t)
	id := playLandFromHand(t, g, "Path of Ancestry", pathOfAncestryOracle)
	top100AssertEnteredTapped(t, g, id, "Path of Ancestry")
}

func TestPathOfAncestryTapsForCommanderIdentity(t *testing.T) {
	abs := game.ManaAbilitiesForCard(game.Card{OracleID: pathOfAncestryOracle})
	if len(abs) != 1 {
		t.Fatalf("%d mana abilities, want 1", len(abs))
	}
	if abs[0].Produced != "{W|U|B|R|G}" {
		t.Errorf("produced %q, want the full pipe set for identity narrowing", abs[0].Produced)
	}
}

// --- Garruk's Uprising -------------------------------------------

// top100CastCreature casts a plain non-catalog creature from a
// player's hand through the real stack path, so the resolution emits
// EventETB and the trigger harvester sees it.
func top100CastCreature(t *testing.T, g *game.Game, p *game.Player, name string, power, toughness int) uuid.UUID {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Beast",
		Power: power, Toughness: toughness, Owner: p.ID, Controller: p.ID,
	})
	// A creature spell needs sorcery speed, and the priority passes
	// that resolved the previous spell may have moved the step on.
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
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
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	top100SeedCreature(g, me.ID, "Big Thing", 5, 5)
	before := me.Hand.Size()

	castCatalogSpell(t, g, "Garruk's Uprising", "Enchantment", garruksUprisingOracle, nil)
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d -> %d, want one more — the intervening-if was satisfied", before, got)
	}
}

// No power-4 creature, no card. The intervening-if is the entire
// difference between this and a three-mana cantrip.
func TestGarruksUprisingDrawsNothingWithoutABigCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	top100SeedCreature(g, me.ID, "Small Thing", 3, 3)
	before := me.Hand.Size()

	castCatalogSpell(t, g, "Garruk's Uprising", "Enchantment", garruksUprisingOracle, nil)
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != before {
		t.Errorf("hand %d -> %d, want unchanged — no creature with power 4 or greater", before, got)
	}
}

func TestGarruksUprisingDrawsWhenABigCreatureArrives(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)

	castCatalogSpell(t, g, "Garruk's Uprising", "Enchantment", garruksUprisingOracle, nil)
	passPriorityAroundTable(t, g)
	before := me.Hand.Size()

	top100CastCreature(t, g, me, "Big Thing", 5, 5)
	passPriorityAroundTable(t, g)

	// top100CastCreature seeds the spell into hand and immediately
	// casts it, so the cast itself is hand-neutral. The whole delta
	// is the trigger's draw.
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d -> %d after a 5/5 entered: want one more (the trigger's draw)", before, got)
	}
}

// A 2/2 arriving must NOT draw — the power gate is the card.
func TestGarruksUprisingIgnoresSmallCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)

	castCatalogSpell(t, g, "Garruk's Uprising", "Enchantment", garruksUprisingOracle, nil)
	passPriorityAroundTable(t, g)
	before := me.Hand.Size()

	top100CastCreature(t, g, me, "Small Thing", 3, 3)
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != before {
		t.Errorf("hand %d -> %d after a 3/3 entered: want unchanged (below the power-4 gate)", before, got)
	}
}

func TestGarruksUprisingGrantsTrampleToYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	opp := top100Opponent(g)
	mine := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Mine", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Theirs", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Garruk's Uprising", TypeLine: "Enchantment",
		OracleID: garruksUprisingOracle, Owner: me.ID, Controller: me.ID,
	})

	if !containsString(effectiveAbilities(t, g, mine), "trample") {
		t.Errorf("your creature has no trample: %v", effectiveAbilities(t, g, mine))
	}
	if containsString(effectiveAbilities(t, g, theirs), "trample") {
		t.Error("the grant is to CREATURES YOU CONTROL; an opponent's creature got trample")
	}
}

// --- Reanimate ---------------------------------------------------

func TestReanimateReturnsACreatureAndChargesItsManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: dead, Name: "Big Dead Thing", TypeLine: "Creature — Wurm",
		ManaCost: "{4}{B}{B}", Power: 6, Toughness: 6,
		Owner: me.ID, Controller: me.ID,
	})
	before := me.Life

	castCatalogSpell(t, g, "Reanimate", "Sorcery", reanimateOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: dead}})
	passPriorityAroundTable(t, g)

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

// The target clause is narrowed to your own graveyard (see
// reanimate.go for why), so an opponent's creature card is not a
// legal pick. Pinning it stops the restriction from being quietly
// widened back to "a graveyard" without the control-change seam.
func TestReanimateRejectsAnOpponentsGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	opp := top100Opponent(g)
	theirs := uuid.New()
	opp.Graveyard.PushTop(game.Card{
		InstanceID: theirs, Name: "Their Wurm", TypeLine: "Creature — Wurm",
		ManaCost: "{4}{B}{B}", Power: 6, Toughness: 6,
		Owner: opp.ID, Controller: opp.ID,
	})

	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Reanimate", TypeLine: "Sorcery",
		OracleID: reanimateOracle, Owner: me.ID, Controller: me.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
	}); err == nil {
		t.Error("Reanimate accepted a card in an opponent's graveyard")
	}
}

// --- Fabled Passage ----------------------------------------------

func TestFabledPassageFetchesTappedBelowFourLands(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	passage := pushCatalogPermanent(g, me.ID, "Fabled Passage", "Land", fabledPassageOracle, false)
	basic := stapleLibraryCard(me, "Forest", "Basic Land — Forest")

	if err := g.ActivateCatalogAbility(me.ID, passage, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(passage) {
		t.Error("the Passage is sacrificed as a cost, at announce")
	}
	passPriorityAroundTable(t, g)

	got, ok := battlefieldCard(g, basic)
	if !ok {
		t.Fatal("Fabled Passage did not fetch the basic")
	}
	if !got.Tapped {
		t.Error("with fewer than four lands the fetched land stays TAPPED")
	}
}

func TestFabledPassageUntapsAtFourLands(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	// Three lands plus the Passage is three after the sacrifice; the
	// fetched land makes four.
	seedPermanentFor(g, me.ID, "Forest A", "Basic Land — Forest")
	seedPermanentFor(g, me.ID, "Forest B", "Basic Land — Forest")
	seedPermanentFor(g, me.ID, "Forest C", "Basic Land — Forest")
	passage := pushCatalogPermanent(g, me.ID, "Fabled Passage", "Land", fabledPassageOracle, false)
	basic := stapleLibraryCard(me, "Island", "Basic Land — Island")

	if err := g.ActivateCatalogAbility(me.ID, passage, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	got, ok := battlefieldCard(g, basic)
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
	abs := game.ActivatedAbilitiesForCard(game.Card{OracleID: fabledPassageOracle})
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
	g := newCatalogGame(t)
	id := playLandFromHand(t, g, "Myriad Landscape", myriadLandscapeOracle)
	top100AssertEnteredTapped(t, g, id, "Myriad Landscape")
}

// Two basics that SHARE a type, both tapped. The type is taken from
// the first basic the search would find; a Forest sitting alongside
// two Plains must not come along for the ride.
func TestMyriadLandscapeFetchesTwoBasicsOfOneType(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	landscape := pushCatalogPermanent(g, me.ID, "Myriad Landscape", "Land", myriadLandscapeOracle, false)
	// stapleLibraryCard pushes to the BOTTOM, and the search walks
	// bottom-first — so the last card seeded here is the one the
	// deterministic pick lands on, and it is the one that names the
	// shared land type.
	forest := stapleLibraryCard(me, "Forest", "Basic Land — Forest")
	plainsA := stapleLibraryCard(me, "Plains", "Basic Land — Plains")
	plainsB := stapleLibraryCard(me, "Plains", "Basic Land — Plains")

	if err := g.ActivateCatalogAbility(me.ID, landscape, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{plainsA, plainsB} {
		got, ok := battlefieldCard(g, id)
		if !ok {
			t.Fatal("a Plains was not fetched")
		}
		if !got.Tapped {
			t.Error("Myriad Landscape's lands arrive TAPPED")
		}
	}
	if _, wrong := battlefieldCard(g, forest); wrong {
		t.Error("the Forest was fetched — the two lands must share a land type")
	}
}

// A whiffed search still costs the land. Nothing arrives, nothing
// crashes.
func TestMyriadLandscapeWithNoBasicsFetchesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	landscape := pushCatalogPermanent(g, me.ID, "Myriad Landscape", "Land", myriadLandscapeOracle, false)
	dual := stapleLibraryCard(me, "Watery Grave", "Land — Island Swamp")

	if err := g.ActivateCatalogAbility(me.ID, landscape, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(landscape) {
		t.Error("the Landscape is sacrificed as a cost, at announce")
	}
	if _, wrong := battlefieldCard(g, dual); wrong {
		t.Error("a nonbasic was fetched — the predicate is basic-land only")
	}
}

func TestMyriadLandscapeCostIsTwoManaTapSacrifice(t *testing.T) {
	abs := game.ActivatedAbilitiesForCard(game.Card{OracleID: myriadLandscapeOracle})
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
