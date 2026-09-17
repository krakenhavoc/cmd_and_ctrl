package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch11_test.go — card-level coverage for the card-coverage
// roadmap's batch 11 (#304, `edhrec_rank` 1214–1315): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, land play or attack. Helpers from
// the earlier batch test files are reused by name; new ones are
// b11-prefixed.

const (
	b11MillikinOracle             = "fa4dffda-6f04-4d0b-829d-28a1a5794dee"
	b11GlaringFleshrakerOracle    = "0ffce5e0-6b1d-4d1a-9318-1a1a219df532"
	b11IronMyrOracle              = "6c5cbab6-ee27-46f5-97a7-df85698d1e9f"
	b11SilverMyrOracle            = "66e8f7f8-3a6d-46ba-837c-b9713ddf7f40"
	b11AbandonedAirTempleOracle   = "9575d7ce-f26d-4b90-87a3-6329e9799572"
	b11MagdaOracle                = "3d268e48-3004-4384-bf52-e63243fb5e02"
	b11ExemplarOfLightOracle      = "9ad730b5-8950-4215-9319-387e2970dd22"
	b11SecureTheWastesOracle      = "b2347910-d6c6-4681-8316-7ef27056485c"
	b11RosieCottonOracle          = "168d5711-4459-440f-8de4-aabffd47c44d"
	b11FractureOracle             = "f21d0319-0509-4ac1-b6e3-10955a26fd7a"
	b11WightOfTheReliquaryOracle  = "4507df69-6bf7-43d6-a609-c032b61835d5"
	b11KeeperOfTheAccordOracle    = "49b9a3ac-d245-431a-bf40-64d1b8dc2b91"
	b11MahadiOracle               = "1b3e841e-0f8f-467d-9983-f9b8d081a67f"
	b11BalefireDragonOracle       = "3d783beb-9ca6-4681-9276-fc3ad13b993f"
	b11MaestrosTheaterOracle      = "9464ddf2-4bcb-44f6-b945-89a132544de6"
	b11PeerIntoTheAbyssOracle     = "21fa2442-6eac-4dce-a9cc-76f0053fdb8f"
	b11HauntedMireOracle          = "b0b58a03-462c-4964-97c7-42bc777ec23e"
	b11SylvanScryingOracle        = "ee24bf27-484d-4e1c-998e-6a74e3d3f6c4"
	b11RuinCrabOracle             = "8afc00d4-a1c6-4329-af2c-a7f58a0c33e7"
	b11AmuletOfVigorOracle        = "dba16032-66c1-4ccb-9d65-d41ac550d182"
	b11NukaColaOracle             = "6cfb03e5-aca6-4fe2-a3f1-93e1f0cbf9e1"
	b11RuneScarredDemonOracle     = "14aefe99-fb60-4f1c-a71f-7ffbe94c8b13"
	b11DefenseOfTheHeartOracle    = "e7e1b166-9267-426d-897d-24903327b48d"
	b11TendershootDryadOracle     = "a336c10a-b5bd-47ff-ba2d-31e27af1e15a"
	b11GuildlessCommonsOracle     = "ee723c7c-ec9f-4ffb-8f36-cd7637eb1fae"
	b11BrokersHideoutOracle       = "bd002797-a545-4bee-88bf-b878436e7cca"
	b11BaneOfProgressAlreadyOID   = "51f9a6cc-8eb2-44ed-a2d9-913ac514ad67"
	b11VorinclexAlreadyOID        = "5a3fdf5a-bff8-4896-b288-3f43f9a72d9b"
	b11CleansingNovaAlreadyOID    = "aff34f28-f707-4458-8af3-1bd5b13a6b10"
	b11KambalAlreadyOID           = "4987c458-604a-4727-b360-170616e91e67"
	b11LightningHelixOracleForTst = "800c258a-cfc4-4a54-a667-065ea8dea69e"
)

// b11Creature seeds a non-catalog creature with a type line, able to
// attack (EnteredBattlefieldAt is zero, so no summoning sickness).
func b11Creature(g *game.Game, owner uuid.UUID, name, typeLine string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// b11Push seeds a catalog permanent through the timestamped push so
// its statics, replacements and triggers are live.
func b11Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// b11Helix casts Lightning Helix at a player for the active seat —
// three damage there, three life here.
func b11Helix(t *testing.T, g *game.Game, target uuid.UUID) {
	t.Helper()
	castCatalogSpell(t, g, "Lightning Helix", "Instant", b11LightningHelixOracleForTst,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: target}})
	passPriorityAroundTable(t, g)
}

// b11BoltCreature kills a creature with toughness 3 or less through
// a real Lightning Bolt cast by the active seat.
func b11BoltCreature(t *testing.T, g *game.Game, target uuid.UUID) {
	t.Helper()
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: target}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(target) {
		t.Fatalf("the Bolt did not kill %s", target)
	}
}

// b11ToMyNextUpkeep walks the cursor around the table to seat 0's
// next upkeep, passing priority through every stop.
func b11ToMyNextUpkeep(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 400; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		if g.Turn.Step == game.StepUpkeep && g.Turn.ActiveSeat == 0 {
			return
		}
	}
	t.Fatal("never reached seat 0's next upkeep")
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Six of them are
// rows in cycle tables or one-line cycle files (the two Myr, Haunted
// Mire, Guildless Commons, the two Overlook lands), so a transposed
// row is invisible until someone plays that exact card. Four of the
// issue's 38 were already on main from the S23 / S27 / S28 work —
// Bane of Progress, Cleansing Nova, Vorinclex, Kambal — pinned here so
// the table matches the issue.
func TestBatch11CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b11MillikinOracle:            "Millikin",
		b11GlaringFleshrakerOracle:   "Glaring Fleshraker",
		b11IronMyrOracle:             "Iron Myr",
		b11SilverMyrOracle:           "Silver Myr",
		b11AbandonedAirTempleOracle:  "Abandoned Air Temple",
		b11MagdaOracle:               "Magda, Brazen Outlaw",
		b11ExemplarOfLightOracle:     "Exemplar of Light",
		b11SecureTheWastesOracle:     "Secure the Wastes",
		b11RosieCottonOracle:         "Rosie Cotton of South Lane",
		b11FractureOracle:            "Fracture",
		b11WightOfTheReliquaryOracle: "Wight of the Reliquary",
		b11KeeperOfTheAccordOracle:   "Keeper of the Accord",
		b11MahadiOracle:              "Mahadi, Emporium Master",
		b11BalefireDragonOracle:      "Balefire Dragon",
		b11MaestrosTheaterOracle:     "Maestros Theater",
		b11PeerIntoTheAbyssOracle:    "Peer into the Abyss",
		b11HauntedMireOracle:         "Haunted Mire",
		b11SylvanScryingOracle:       "Sylvan Scrying",
		b11RuinCrabOracle:            "Ruin Crab",
		b11AmuletOfVigorOracle:       "Amulet of Vigor",
		b11NukaColaOracle:            "Nuka-Cola Vending Machine",
		b11RuneScarredDemonOracle:    "Rune-Scarred Demon",
		b11DefenseOfTheHeartOracle:   "Defense of the Heart",
		b11TendershootDryadOracle:    "Tendershoot Dryad",
		b11GuildlessCommonsOracle:    "Guildless Commons",
		b11BrokersHideoutOracle:      "Brokers Hideout",
		b11BaneOfProgressAlreadyOID:  "Bane of Progress",
		b11VorinclexAlreadyOID:       "Vorinclex, Monstrous Raider",
		b11CleansingNovaAlreadyOID:   "Cleansing Nova",
		b11KambalAlreadyOID:          "Kambal, Consul of Allocation",
	}
	if len(want) != 30 {
		t.Fatalf("the batch registers 26 cards plus four already on main, the table lists %d", len(want))
	}
	alreadyOnMain := map[string]bool{
		"Bane of Progress": true, "Vorinclex, Monstrous Raider": true,
		"Cleansing Nova": true, "Kambal, Consul of Allocation": true,
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
		if spec.Completeness == CompletenessUnreviewed && !alreadyOnMain[name] {
			t.Errorf("%s ships without a Completeness declaration", name)
		}
	}
}

// --- Millikin and the Myr -----------------------------------------

func TestB11MillikinMillsACardForColorlessAndNeedsALibrary(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	millikin := pushCatalogPermanent(g, me.ID, "Millikin", "Artifact Creature — Construct", b11MillikinOracle, false)
	lib, yard := me.Library.Size(), me.Graveyard.Size()

	if err := g.ActivateManaAbility(me.ID, millikin, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
	if me.Library.Size() != lib-1 || me.Graveyard.Size() != yard+1 {
		t.Errorf("library %d → %d, graveyard %d → %d: want one card milled", lib, me.Library.Size(), yard, me.Graveyard.Size())
	}
	if c, _ := battlefieldCard(g, millikin); !c.Tapped {
		t.Error("Millikin should be tapped")
	}

	// With nothing to mill the cost cannot be paid, so the ability
	// cannot be activated and the Myr stays untapped.
	b02bSetTapped(g, millikin, false)
	me.Library.Cards = nil
	if err := g.ActivateManaAbility(me.ID, millikin, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("Millikin activated with an empty library")
	}
	if c, _ := battlefieldCard(g, millikin); c.Tapped {
		t.Error("a refused activation must not tap the source")
	}
}

func TestB11MillikinIsSummoningSick(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	millikin := pushCatalogPermanent(g, me.ID, "Millikin", "Artifact Creature — Construct", b11MillikinOracle, true)
	if err := g.ActivateManaAbility(me.ID, millikin, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("a summoning-sick Millikin tapped for mana")
	}
}

func TestB11ManaMyrTapForTheirColour(t *testing.T) {
	for _, tc := range []struct{ name, oracle, color string }{
		{"Iron Myr", b11IronMyrOracle, "R"},
		{"Silver Myr", b11SilverMyrOracle, "U"},
	} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		myr := pushCatalogPermanent(g, me.ID, tc.name, "Artifact Creature — Myr", tc.oracle, false)
		if err := g.ActivateManaAbility(me.ID, myr, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("%s: activate: %v", tc.name, err)
		}
		if got := batch01PoolColors(me); len(got) != 1 || got[0] != tc.color {
			t.Errorf("%s: pool %v, want [%s]", tc.name, got, tc.color)
		}
		sick := pushCatalogPermanent(g, me.ID, tc.name, "Artifact Creature — Myr", tc.oracle, true)
		if err := g.ActivateManaAbility(me.ID, sick, 0, game.ManaAbilityParams{}); err == nil {
			t.Errorf("%s: a summoning-sick Myr tapped for mana", tc.name)
		}
	}
}

// --- Glaring Fleshraker -------------------------------------------

func TestB11GlaringFleshrakerSpawnsOnColorlessSpellsAndPingsOnColorlessCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b11Push(g, me.ID, "Glaring Fleshraker", "Creature — Eldrazi Drone", b11GlaringFleshrakerOracle, 2, 2)
	before := lifeOfOpponents(g)

	// A colorless spell: a Spawn, whose own entry is a colorless
	// creature entering, so every opponent takes 1.
	castCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Eldrazi Spawn"); n != 1 {
		t.Fatalf("a colorless spell made %d Spawn, want 1", n)
	}
	for i, opp := range g.Seats[1:] {
		if opp.Life != before[i]-1 {
			t.Errorf("opponent %d: %d → %d, want -1 from the Spawn entering", i+1, before[i], opp.Life)
		}
	}

	// A coloured spell makes nothing, and a coloured creature entering
	// pings nobody.
	castWithCost(t, g, "Bear", "Creature — Bear", "{1}{G}", "")
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Eldrazi Spawn"); n != 1 {
		t.Errorf("a green spell made a Spawn (%d total)", n)
	}
	for i, opp := range g.Seats[1:] {
		if opp.Life != before[i]-1 {
			t.Errorf("opponent %d: %d after a green creature, want %d", i+1, opp.Life, before[i]-1)
		}
	}
}

// --- Abandoned Air Temple -----------------------------------------

func TestB11AbandonedAirTempleEntersTappedWithoutABasic(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLandOnBattlefield(g, me.ID, "Command Tower", "Land")
	tapped := playLandFromHand(t, g, "Abandoned Air Temple", b11AbandonedAirTempleOracle)
	top100AssertEnteredTapped(t, g, tapped, "Abandoned Air Temple without a basic")

	seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	untapped := playLandFromHand(t, g, "Abandoned Air Temple", b11AbandonedAirTempleOracle)
	top100AssertEnteredUntapped(t, g, untapped, "Abandoned Air Temple with a basic Forest")

	spec, _ := Lookup(b11AbandonedAirTempleOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{W}" {
		t.Errorf("mana abilities %+v, want one producing {W}", spec.ManaAbilities)
	}
}

func TestB11AbandonedAirTemplePutsACounterOnEachOfYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	temple := pushCatalogPermanent(g, me.ID, "Abandoned Air Temple", "Land", b11AbandonedAirTempleOracle, false)
	a := seedCreature(g, "Bear A", me.ID)
	b := seedCreature(g, "Bear B", me.ID)
	theirs := seedCreature(g, "Their Bear", opp.ID)
	me.ManaPool.AddMana(game.ManaToken{Color: "W"})
	b09AddColorless(me, 3)

	if err := g.ActivateCatalogAbility(me.ID, temple, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if c, _ := battlefieldCard(g, temple); !c.Tapped {
		t.Fatal("the ability has a tap cost")
	}
	if countersOn(g, a, "+1/+1") != 0 {
		t.Fatal("the counters must wait for the ability to resolve")
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if got := countersOn(g, id, "+1/+1"); got != 1 {
			t.Errorf("my creature has %d +1/+1 counters, want 1", got)
		}
	}
	if got := countersOn(g, theirs, "+1/+1"); got != 0 {
		t.Errorf("an opponent's creature got %d counters", got)
	}
}

// --- Magda, Brazen Outlaw -----------------------------------------

func TestB11MagdaPumpsOtherDwarvesAndPaysATreasurePerTappedDwarf(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	magda := b11Push(g, me.ID, "Magda, Brazen Outlaw", "Legendary Creature — Dwarf Berserker", b11MagdaOracle, 2, 1)
	dwarf := b11Creature(g, me.ID, "Dwarf", "Creature — Dwarf Warrior", 2, 2)
	bear := b11Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirDwarf := b11Creature(g, opp.ID, "Their Dwarf", "Creature — Dwarf", 2, 2)

	if p := effectivePower(t, g, dwarf); p != 3 {
		t.Errorf("another Dwarf = %d power, want 3", p)
	}
	if p := effectivePower(t, g, magda); p != 2 {
		t.Errorf("Magda pumps herself: %d power, want 2", p)
	}
	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("a non-Dwarf = %d power, want 2", p)
	}
	if p := effectivePower(t, g, theirDwarf); p != 2 {
		t.Errorf("an opponent's Dwarf = %d power, want 2", p)
	}

	// Magda and the Dwarf attack — two Dwarves tapped, two Treasures;
	// the Bear is not a Dwarf.
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{magda, dwarf, bear} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Treasure"); n != 2 {
		t.Fatalf("three attackers, two of them Dwarves, made %d Treasures, want 2", n)
	}
}

func TestB11MagdaIgnoresAVigilantDwarfAndAnOpponentsTap(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b11Push(g, me.ID, "Magda, Brazen Outlaw", "Legendary Creature — Dwarf Berserker", b11MagdaOracle, 2, 1)
	vigilant := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Watchful Dwarf", TypeLine: "Creature — Dwarf",
		Power: 2, Toughness: 2, Keywords: []string{"vigilance"}, Owner: me.ID, Controller: me.ID,
	})
	theirDwarf := b11Creature(g, opp.ID, "Their Dwarf", "Creature — Dwarf", 2, 2)

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(vigilant, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Treasure"); n != 0 {
		t.Errorf("a vigilance Dwarf attacking made %d Treasures; it never became tapped", n)
	}

	// An opponent's Dwarf tapping is not "a Dwarf you control".
	if err := g.TapCard(theirDwarf, true); err != nil {
		t.Fatalf("TapCard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Treasure"); n != 0 {
		t.Errorf("an opponent's Dwarf tapping made %d Treasures", n)
	}
}

func TestB11MagdaPaysForADwarfTappedOutsideCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b11Push(g, me.ID, "Magda, Brazen Outlaw", "Legendary Creature — Dwarf Berserker", b11MagdaOracle, 2, 1)
	dwarf := b11Creature(g, me.ID, "Dwarf", "Creature — Dwarf", 2, 2)
	if err := g.TapCard(dwarf, true); err != nil {
		t.Fatalf("TapCard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("a Dwarf tapped outside combat made %d Treasures, want 1", n)
	}
	// #747: the five-Treasure tutor ships at its printed count; its
	// engine test is in sacrifice_n_cards_test.go.
	spec, _ := Lookup(b11MagdaOracle)
	if len(spec.Activated) != 1 || spec.Completeness != CompletenessFull {
		t.Error("the five-Treasure tutor ships whole")
	}
}

// --- Exemplar of Light --------------------------------------------

func TestB11ExemplarOfLightGrowsOnLifegainAndDrawsOncePerTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	exemplar := b11Push(g, me.ID, "Exemplar of Light", "Creature — Angel", b11ExemplarOfLightOracle, 3, 3)
	if !eotHasAbility(effectiveAbilities(t, g, exemplar), "flying") {
		t.Error("printed flying did not reach the effective abilities")
	}

	// castCatalogSpell adds the card it then casts, so the hand is
	// neutral across a cast and the difference is exactly the draws.
	before := me.Hand.Size()
	b11Helix(t, g, opp.ID)
	if got := countersOn(g, exemplar, "+1/+1"); got != 1 {
		t.Fatalf("after gaining 3: %d +1/+1 counters, want 1", got)
	}
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("first counter of the turn: hand %d → %d, want +1", before, got)
	}

	before = me.Hand.Size()
	b11Helix(t, g, opp.ID)
	if got := countersOn(g, exemplar, "+1/+1"); got != 2 {
		t.Fatalf("after gaining 3 again: %d counters, want 2", got)
	}
	if got := me.Hand.Size(); got != before {
		t.Errorf("second counter of the turn drew: hand %d → %d", before, got)
	}

	// A new turn resets the once-per-turn clause. (Measure from the
	// main phase so the turn's draw step is not counted.)
	b11ToMyNextUpkeep(t, g)
	advanceTo(t, g, game.StepPrecombatMain)
	before = me.Hand.Size()
	b11Helix(t, g, opp.ID)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("next turn's first counter: hand %d → %d, want +1", before, got)
	}
}

func TestB11ExemplarOfLightDoesNotDrawForAnOpponentsCounterOrARemoval(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	exemplar := b11Push(g, me.ID, "Exemplar of Light", "Creature — Angel", b11ExemplarOfLightOracle, 3, 3)

	// The opponent resolves something, then a counter lands on the
	// Exemplar: that is their placement, not yours.
	batch01OpponentCasts(t, g, opp, "Lightning Bolt", lightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	before := me.Hand.Size()
	if err := g.AddCounter(exemplar, "+1/+1", 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before {
		t.Errorf("an opponent's counter drew: hand %d → %d", before, got)
	}

	// Removing a counter is not placing one.
	b11ToMyNextUpkeep(t, g)
	castCatalogSpell(t, g, "Nothing", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	before = me.Hand.Size()
	if err := g.AddCounter(exemplar, "+1/+1", -1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before {
		t.Errorf("a counter removal drew: hand %d → %d", before, got)
	}
}

// --- Secure the Wastes --------------------------------------------

func TestB11SecureTheWastesMakesXWarriors(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castXSpell(t, g, "Secure the Wastes", "Instant", b11SecureTheWastesOracle, "{X}{W}", 3, nil)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Warrior"); n != 3 {
		t.Fatalf("X=3 made %d Warriors, want 3", n)
	}
	w := findBattlefieldByName(g, "Warrior")
	if c, _ := battlefieldCard(g, w); !c.HasColor("W") || c.Power != 1 || c.Toughness != 1 {
		t.Errorf("the Warrior should be a 1/1 white token, got %+v", c)
	}
}

// --- Rosie Cotton of South Lane -----------------------------------

func TestB11RosieCottonMakesAFoodAndPutsTheCounterOnAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := seedCreature(g, "Bear", me.ID)
	rosie := castAndResolveCreature(t, g, "Rosie Cotton of South Lane", "Legendary Creature — Halfling Peasant", b11RosieCottonOracle)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Food"); n != 1 {
		t.Fatalf("Rosie made %d Food, want 1", n)
	}
	// Her own Food is a token you created, so the second ability asks
	// for a creature — and Rosie herself is not on offer.
	prompt := latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("the token trigger must ask which creature gets the counter")
	}
	if hasID(prompt.PickTargetCards, rosie) {
		t.Error("Rosie offered herself as the target")
	}
	if !hasID(prompt.PickTargetCards, bear) {
		t.Fatal("the other creature should be a legal target")
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, bear, "+1/+1"); got != 1 {
		t.Errorf("the Bear has %d +1/+1 counters, want 1", got)
	}
	if got := countersOn(g, rosie, "+1/+1"); got != 0 {
		t.Errorf("Rosie has %d counters, want 0", got)
	}
}

func TestB11RosieCottonAloneDropsTheTargetedTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castAndResolveCreature(t, g, "Rosie Cotton of South Lane", "Legendary Creature — Halfling Peasant", b11RosieCottonOracle)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Food"); n != 1 {
		t.Fatalf("Rosie made %d Food, want 1", n)
	}
	if latestPickTarget(g, me.ID) != nil {
		t.Error("with no other creature the trigger has no legal target and must not prompt")
	}
}

// --- Fracture -----------------------------------------------------

func TestB11FractureDestroysAPlaneswalkerAndRefusesACreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	// A planeswalker needs loyalty, or the state check buries it
	// before the spell is cast.
	walker := pushWalkerForTest(g, opp.ID, "Jace", "", 3)
	castCatalogSpell(t, g, "Fracture", "Instant", b11FractureOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: walker}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(walker) || !opp.Graveyard.Contains(walker) {
		t.Error("the planeswalker should be in its owner's graveyard")
	}

	bear := seedCreature(g, "Bear", opp.ID)
	if err := b09TryCast(t, g, "Fracture", "Instant", b11FractureOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}}); err == nil {
		t.Error("a creature was accepted for 'target artifact, enchantment, or planeswalker'")
	}
}

// --- Wight of the Reliquary ---------------------------------------

func TestB11WightOfTheReliquaryGrowsWithDeadCreaturesAndRampsOffAnother(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	wight := b11Push(g, me.ID, "Wight of the Reliquary", "Creature — Zombie Knight", b11WightOfTheReliquaryOracle, 2, 2)
	if !eotHasAbility(effectiveAbilities(t, g, wight), "vigilance") {
		t.Error("printed vigilance did not reach the effective abilities")
	}
	if p := effectivePower(t, g, wight); p != 2 {
		t.Fatalf("empty graveyard: %d power, want 2", p)
	}
	pushGraveyardCardWithTypeLine(g, me.ID, "Dead Bear", "Creature — Bear")
	pushGraveyardCardWithTypeLine(g, me.ID, "Old Rock", "Artifact")
	if p := effectivePower(t, g, wight); p != 3 {
		t.Errorf("one creature card in the graveyard: %d power, want 3", p)
	}

	fodder := seedCreature(g, "Fodder", me.ID)
	seedSearchLibrary(me,
		game.Card{Name: "Forest A", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Bayou", TypeLine: "Land — Swamp Forest"},
		game.Card{Name: "Bear Card", TypeLine: "Creature — Bear"},
	)
	// The Wight cannot be its own "another creature".
	if err := g.ActivateCatalogAbility(me.ID, wight, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{wight},
	}); err == nil {
		t.Fatal("the Wight sacrificed itself to its own ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, wight, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Fatal("the sacrifice is a cost and is paid at announce")
	}
	if c, _ := battlefieldCard(g, wight); !c.Tapped {
		t.Error("the ability has a tap cost")
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Bear Card") != uuid.Nil {
		t.Error("a non-land was offered")
	}
	if searchOptionNamed(g, c, "Bayou") == uuid.Nil {
		t.Error("any land card qualifies, not only basics")
	}
	answerSearchNamed(t, g, me.ID, "Bayou")
	land, ok := searchFetchedCard(g, "Bayou")
	if !ok || !land.Tapped {
		t.Errorf("the land should be on the battlefield tapped, got %+v (found %v)", land, ok)
	}
	// The sacrificed creature is now a creature card in the graveyard.
	if p := effectivePower(t, g, wight); p != 4 {
		t.Errorf("two creature cards in the graveyard: %d power, want 4", p)
	}
}

// --- Keeper of the Accord -----------------------------------------

func TestB11KeeperOfTheAccordCatchesUpAtARicherOpponentsEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, bystander := g.Seats[0], g.Seats[1], g.Seats[2]
	b11Push(g, me.ID, "Keeper of the Accord", "Creature — Human Soldier", b11KeeperOfTheAccordOracle, 3, 4)
	seedCreature(g, "Their Bear A", opp.ID)
	seedCreature(g, "Their Bear B", opp.ID)
	seedLandOnBattlefield(g, opp.ID, "Plains", "Basic Land — Plains")
	seedLandOnBattlefield(g, opp.ID, "Plains", "Basic Land — Plains")
	seedLandOnBattlefield(g, me.ID, "Plains", "Basic Land — Plains")
	seedSearchLibrary(me,
		game.Card{Name: "Plains A", TypeLine: "Basic Land — Plains"},
		game.Card{Name: "Snowfield", TypeLine: "Land — Plains"},
		game.Card{Name: "Forest A", TypeLine: "Basic Land — Forest"},
	)

	// Opponent 1 has more creatures (2 vs 1) and more lands (2 vs 1):
	// both triggers fire at their end step, and the controller orders
	// them.
	advanceToEndStepOf(t, g, 1)
	if !answerAnyTriggerOrderPrompt(t, g, me.ID) {
		t.Fatal("two Keeper triggers should ask for an order")
	}
	for i := 0; i < 16 && searchChoiceFor(g, me.ID) == nil && !stackFullyEmpty(g); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the land trigger should offer a basic Plains search")
	}
	if searchOptionNamed(g, c, "Snowfield") != uuid.Nil || searchOptionNamed(g, c, "Forest A") != uuid.Nil {
		t.Error("only a BASIC Plains card is on offer")
	}
	answerSearchNamed(t, g, me.ID, "Plains A")
	passPriorityAroundTable(t, g)
	if plains, ok := searchFetchedCard(g, "Plains A"); !ok || !plains.Tapped {
		t.Errorf("the Plains should be on the battlefield tapped, got %+v (found %v)", plains, ok)
	}
	if n := countBattlefieldNamed(g, me.ID, "Soldier"); n != 1 {
		t.Errorf("%d Soldiers, want 1", n)
	}

	// Opponent 2 controls nothing, so their end step is quiet — and
	// so is the controller's own.
	advanceToEndStepOf(t, g, 2)
	passPriorityAroundTable(t, g)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Soldier"); n != 1 {
		t.Errorf("a poorer opponent's end step made a Soldier (%d total)", n)
	}
	_ = bystander
}

// --- Mahadi, Emporium Master --------------------------------------

func TestB11MahadiPaysATreasurePerCreatureThatDiedThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b11Push(g, me.ID, "Mahadi, Emporium Master", "Legendary Creature — Devil", b11MahadiOracle, 3, 3)
	mine := seedCreature(g, "My Bear", me.ID)
	theirs := seedCreature(g, "Their Bear", opp.ID)
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(opp.ID, RedGoblinToken(), 1)
	})
	goblin := findBattlefieldByName(g, "Goblin")
	rock := seedPermanentFor(g, opp.ID, "Rock", "Artifact")

	b11BoltCreature(t, g, mine)
	b11BoltCreature(t, g, theirs)
	b11BoltCreature(t, g, goblin)
	castCatalogSpell(t, g, "Fracture", "Instant", b11FractureOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	passPriorityAroundTable(t, g)

	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Treasure"); n != 3 {
		t.Fatalf("three creatures (one a token) and an artifact died: %d Treasures, want 3", n)
	}

	// Next turn, nothing has died.
	b11ToMyNextUpkeep(t, g)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Treasure"); n != 3 {
		t.Errorf("a turn with no deaths made Treasures (%d total)", n)
	}
}

// --- Balefire Dragon ----------------------------------------------

func TestB11BalefireDragonWipesTheConnectedPlayersCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, victim, bystander := g.Seats[0], g.Seats[1], g.Seats[2]
	dragon := b11Push(g, me.ID, "Balefire Dragon", "Creature — Dragon", b11BalefireDragonOracle, 6, 6)
	if !eotHasAbility(effectiveAbilities(t, g, dragon), "flying") {
		t.Error("printed flying did not reach the effective abilities")
	}
	small := b11Creature(g, victim.ID, "Small", "Creature — Bear", 2, 2)
	exact := b11Creature(g, victim.ID, "Exact", "Creature — Giant", 6, 6)
	big := b11Creature(g, victim.ID, "Big", "Creature — Wurm", 7, 7)
	spared := b11Creature(g, bystander.ID, "Spared", "Creature — Bear", 2, 2)
	before := victim.Life

	attackWith(t, g, victim.ID, dragon)
	if victim.Life != before-6 {
		t.Fatalf("victim %d → %d, want -6", before, victim.Life)
	}
	if triggerOnStack(g, dragon) == nil {
		t.Fatal("the rider is a trigger with a response window")
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{small, exact} {
		if g.Battlefield.Contains(id) {
			t.Errorf("a creature with toughness ≤ 6 survived 6 damage")
		}
	}
	if !g.Battlefield.Contains(big) || damageMarkedOn(g, big) != 6 {
		t.Error("the 7/7 should survive with 6 damage marked")
	}
	if !g.Battlefield.Contains(spared) || damageMarkedOn(g, spared) != 0 {
		t.Error("another player's creature was touched")
	}
}

// --- the Overlook lands -------------------------------------------

func TestB11OverlookLandsSacrificeForATappedBasicAndALife(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		offered      []string
		refused      string
	}{
		{"Maestros Theater", b11MaestrosTheaterOracle, []string{"Island A", "Swamp A", "Mountain A"}, "Forest A"},
		{"Brokers Hideout", b11BrokersHideoutOracle, []string{"Forest A", "Plains A", "Island A"}, "Swamp A"},
	} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		seedSearchLibrary(me,
			game.Card{Name: "Island A", TypeLine: "Basic Land — Island"},
			game.Card{Name: "Swamp A", TypeLine: "Basic Land — Swamp"},
			game.Card{Name: "Mountain A", TypeLine: "Basic Land — Mountain"},
			game.Card{Name: "Forest A", TypeLine: "Basic Land — Forest"},
			game.Card{Name: "Plains A", TypeLine: "Basic Land — Plains"},
			game.Card{Name: "Bayou", TypeLine: "Land — Swamp Forest"},
		)
		before := me.Life
		land := playLandFromHand(t, g, tc.name, tc.oracle)
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(land) || !me.Graveyard.Contains(land) {
			t.Fatalf("%s should have sacrificed itself", tc.name)
		}
		c := searchChoiceFor(g, me.ID)
		if c == nil {
			t.Fatalf("%s: no search prompt", tc.name)
		}
		for _, name := range tc.offered {
			if searchOptionNamed(g, c, name) == uuid.Nil {
				t.Errorf("%s: %s should be offered", tc.name, name)
			}
		}
		if searchOptionNamed(g, c, tc.refused) != uuid.Nil || searchOptionNamed(g, c, "Bayou") != uuid.Nil {
			t.Errorf("%s: an off-colour or nonbasic land was offered", tc.name)
		}
		answerSearchNamed(t, g, me.ID, tc.offered[0])
		fetched, ok := searchFetchedCard(g, tc.offered[0])
		if !ok || !fetched.Tapped {
			t.Errorf("%s: the basic should be on the battlefield tapped", tc.name)
		}
		if me.Life != before+1 {
			t.Errorf("%s: life %d → %d, want +1", tc.name, before, me.Life)
		}
	}
}

// --- Peer into the Abyss ------------------------------------------

func TestB11PeerIntoTheAbyssHalvesLibraryAndLifeRoundedUp(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	seedSearchLibrary(opp,
		game.Card{Name: "a"}, game.Card{Name: "b"}, game.Card{Name: "c"}, game.Card{Name: "d"},
		game.Card{Name: "e"}, game.Card{Name: "f"}, game.Card{Name: "g"},
	)
	opp.Life = 21
	hand := opp.Hand.Size()

	castCatalogSpell(t, g, "Peer into the Abyss", "Sorcery", b11PeerIntoTheAbyssOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)

	if got := opp.Hand.Size(); got != hand+4 {
		t.Errorf("seven cards in library: hand %d → %d, want +4", hand, got)
	}
	if opp.Library.Size() != 3 {
		t.Errorf("library has %d cards, want 3", opp.Library.Size())
	}
	if opp.Life != 10 {
		t.Errorf("21 life: now %d, want 10 (loses 11)", opp.Life)
	}
}

// --- Haunted Mire / Guildless Commons -----------------------------

func TestB11HauntedMireEntersTappedAndTapsForEither(t *testing.T) {
	g := newCatalogGame(t)
	id := playLandFromHand(t, g, "Haunted Mire", b11HauntedMireOracle)
	top100AssertEnteredTapped(t, g, id, "Haunted Mire")
	spec, _ := Lookup(b11HauntedMireOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{B|G}" {
		t.Errorf("row produces %+v, want {B|G}", spec.ManaAbilities)
	}
}

func TestB11GuildlessCommonsIsAColourlessKaroo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	other := seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	commons := playLandFromHand(t, g, "Guildless Commons", b11GuildlessCommonsOracle)
	top100AssertEnteredTapped(t, g, commons, "Guildless Commons")
	prompt := latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("the bounce trigger must ask which land to return")
	}
	pickCard(t, g, me.ID, other)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(other) || !me.Hand.Contains(other) {
		t.Error("the chosen land should be back in hand")
	}
	spec, _ := Lookup(b11GuildlessCommonsOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{C}{C}" {
		t.Errorf("row produces %+v, want {C}{C}", spec.ManaAbilities)
	}
}

// --- Sylvan Scrying -----------------------------------------------

func TestB11SylvanScryingTutorsAnyLandToHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedSearchLibrary(me,
		game.Card{Name: "Forest A", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Cradle", TypeLine: "Legendary Land"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	castCatalogSpell(t, g, "Sylvan Scrying", "Sorcery", b11SylvanScryingOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Bear") != uuid.Nil {
		t.Error("a non-land was offered")
	}
	answerSearchNamed(t, g, me.ID, "Cradle")
	if !b02bHandHasNamed(me, "Cradle") {
		t.Error("the chosen land did not reach the hand")
	}
}

// --- Ruin Crab ----------------------------------------------------

func TestB11RuinCrabMillsEachOpponentThreeOnLandfall(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b11Push(g, me.ID, "Ruin Crab", "Creature — Crab", b11RuinCrabOracle, 0, 3)
	var libs, yards []int
	for _, opp := range g.Seats[1:] {
		libs = append(libs, opp.Library.Size())
		yards = append(yards, opp.Graveyard.Size())
	}
	myLib := me.Library.Size()

	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	for i, opp := range g.Seats[1:] {
		if opp.Library.Size() != libs[i]-3 || opp.Graveyard.Size() != yards[i]+3 {
			t.Errorf("opponent %d: library %d → %d, graveyard %d → %d, want three milled",
				i+1, libs[i], opp.Library.Size(), yards[i], opp.Graveyard.Size())
		}
	}
	if me.Library.Size() != myLib {
		t.Error("the controller was milled")
	}
}

// --- Amulet of Vigor ----------------------------------------------

func TestB11AmuletOfVigorUntapsAPermanentThatEntersTapped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Amulet of Vigor", "Artifact", b11AmuletOfVigorOracle, false)

	mire := playLandFromHand(t, g, "Haunted Mire", b11HauntedMireOracle)
	if c, _ := battlefieldCard(g, mire); !c.Tapped {
		t.Fatal("Haunted Mire enters tapped")
	}
	if triggerOnStack(g, findBattlefieldByName(g, "Amulet of Vigor")) == nil {
		t.Fatal("the untap is a trigger with a response window")
	}
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, mire); c.Tapped {
		t.Error("the Amulet should have untapped the Mire")
	}

	// A tapped token counts too; an opponent's tapped entry does not.
	g.WithWriteLock(func() {
		_, _ = g.CreateTokensForEffect(me.ID, TreasureToken(), 1, game.TokenEntryOptions{Tapped: true})
		_, _ = g.CreateTokensForEffect(opp.ID, TreasureToken(), 1, game.TokenEntryOptions{Tapped: true})
	})
	passPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Treasure" {
			continue
		}
		if c.Controller == me.ID && c.Tapped {
			t.Error("my tapped Treasure should have been untapped")
		}
		if c.Controller == opp.ID && !c.Tapped {
			t.Error("an opponent's Treasure was untapped")
		}
	}

	// A permanent that enters untapped is left alone — no trigger.
	forest := playLandFromHand(t, g, "Forest", "")
	if triggersOnStackFrom(g, findBattlefieldByName(g, "Amulet of Vigor")) != 0 {
		t.Error("an untapped entry queued an Amulet trigger")
	}
	if c, _ := battlefieldCard(g, forest); c.Tapped {
		t.Error("the Forest should be untapped")
	}
}

// --- Nuka-Cola Vending Machine ------------------------------------

func TestB11NukaColaMakesFoodAndATappedTreasurePerFoodSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	machine := pushCatalogPermanent(g, me.ID, "Nuka-Cola Vending Machine", "Artifact", b11NukaColaOracle, false)
	b09AddColorless(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, machine, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	food := findBattlefieldByName(g, "Food")
	if food == uuid.Nil {
		t.Fatal("no Food was made")
	}

	// Cracking the Food (its own {2}, {T}, sacrifice ability) is a
	// Food sacrificed: a tapped Treasure, and the life.
	before := me.Life
	b09AddColorless(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, food, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("crack the Food: %v", err)
	}
	if g.Battlefield.Contains(food) {
		t.Fatal("the Food is sacrificed as a cost")
	}
	if triggerOnStack(g, machine) == nil {
		t.Fatal("the Treasure trigger should be on the stack above the Food's ability")
	}
	passPriorityAroundTable(t, g)
	if me.Life != before+3 {
		t.Errorf("life %d → %d, want +3", before, me.Life)
	}
	treasure := findBattlefieldByName(g, "Treasure")
	if treasure == uuid.Nil {
		t.Fatal("no Treasure was made")
	}
	if c, _ := battlefieldCard(g, treasure); !c.Tapped {
		t.Error("the Treasure enters tapped")
	}

	// Sacrificing a non-Food makes nothing.
	bear := seedCreature(g, "Bear", me.ID)
	if err := g.SacrificePermanent(me.ID, bear); err != nil {
		t.Fatalf("SacrificePermanent: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("a non-Food sacrifice made a Treasure (%d total)", n)
	}
}

// --- Rune-Scarred Demon -------------------------------------------

func TestB11RuneScarredDemonTutorsAnyCardOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedSearchLibrary(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Forest A", TypeLine: "Basic Land — Forest"},
	)
	demon := castAndResolveCreature(t, g, "Rune-Scarred Demon", "Creature — Demon", b11RuneScarredDemonOracle)
	if !eotHasAbility(effectiveAbilities(t, g, demon), "flying") {
		t.Error("printed flying did not reach the effective abilities")
	}
	if triggerOnStack(g, demon) == nil {
		t.Fatal("the search is an ETB trigger with a response window")
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Bear") == uuid.Nil || searchOptionNamed(g, c, "Forest A") == uuid.Nil {
		t.Error("every card in the library is on offer")
	}
	answerSearchNamed(t, g, me.ID, "Forest A")
	if !b02bHandHasNamed(me, "Forest A") {
		t.Error("the chosen card did not reach the hand")
	}
}

// --- Defense of the Heart -----------------------------------------

func TestB11DefenseOfTheHeartFiresAgainstThreeCreaturesAndFetchesTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	defense := b11Push(g, me.ID, "Defense of the Heart", "Enchantment", b11DefenseOfTheHeartOracle, 0, 0)
	seedCreature(g, "Their Bear A", opp.ID)
	seedCreature(g, "Their Bear B", opp.ID)

	// Two creatures: the intervening-if fails and nothing happens.
	b11ToMyNextUpkeep(t, g)
	passPriorityAroundTable(t, g)
	if searchChoiceFor(g, me.ID) != nil || !g.Battlefield.Contains(defense) {
		t.Fatal("Defense of the Heart fired against two creatures")
	}

	seedCreature(g, "Their Bear C", opp.ID)
	// Seeded top-first under a filler card, which the draw step takes
	// on the way round, leaving more matches than the search's limit
	// so the chooser is asked.
	seedSearchLibrary(me,
		game.Card{Name: "Filler", TypeLine: "Sorcery"},
		game.Card{Name: "Big A", TypeLine: "Creature — Wurm"},
		game.Card{Name: "Big B", TypeLine: "Creature — Dragon"},
		game.Card{Name: "Big C", TypeLine: "Creature — Giant"},
		game.Card{Name: "Rock", TypeLine: "Artifact"},
	)
	b11ToMyNextUpkeep(t, g)
	if triggerOnStack(g, defense) == nil {
		t.Fatal("three opposing creatures: the trigger should be on the stack")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(defense) || !me.Graveyard.Contains(defense) {
		t.Error("the enchantment should have been sacrificed")
	}
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if c.SearchMax != 2 {
		t.Errorf("search max %d, want 2", c.SearchMax)
	}
	if searchOptionNamed(g, c, "Rock") != uuid.Nil {
		t.Error("a non-creature was offered")
	}
	answerSearchNamed(t, g, me.ID, "Big A", "Big C")
	for _, name := range []string{"Big A", "Big C"} {
		card, ok := searchFetchedCard(g, name)
		if !ok || card.Controller != me.ID {
			t.Errorf("%s should be on the battlefield under my control", name)
		} else if card.Tapped {
			t.Errorf("%s entered tapped; the printed card says nothing about tapped", name)
		}
	}
}

// --- Tendershoot Dryad --------------------------------------------

func TestB11TendershootDryadMakesASaprolingEveryUpkeepAndPumpsAtTen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b11Push(g, me.ID, "Tendershoot Dryad", "Creature — Dryad", b11TendershootDryadOracle, 2, 2)

	// Each player's upkeep: the next opponent's, then the controller's.
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Saproling"); n != 1 {
		t.Fatalf("an opponent's upkeep made %d Saprolings, want 1", n)
	}
	b11ToMyNextUpkeep(t, g)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Saproling"); n != 4 {
		t.Fatalf("after every seat's upkeep: %d Saprolings, want 4", n)
	}
	sap := findBattlefieldByName(g, "Saproling")
	if p := effectivePower(t, g, sap); p != 1 {
		t.Errorf("five permanents: Saproling power %d, want 1", p)
	}

	// Ten permanents: the Dryad, four Saprolings, five lands.
	for i := 0; i < 5; i++ {
		seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	}
	g.WithWriteLock(func() { g.BumpLayerVersionForTest() })
	if p := effectivePower(t, g, sap); p != 3 {
		t.Errorf("ten permanents: Saproling power %d, want 3", p)
	}
}

// --- the once-per-turn and died-this-turn walks -------------------

func TestB11CreaturesDiedThisTurnResetsAtUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := seedCreature(g, "Bear", me.ID)
	g.WithWriteLock(func() {
		if n := b11CreaturesDiedThisTurn(g); n != 0 {
			t.Errorf("fresh game: %d deaths, want 0", n)
		}
	})
	b11BoltCreature(t, g, bear)
	g.WithWriteLock(func() {
		if n := b11CreaturesDiedThisTurn(g); n != 1 {
			t.Errorf("after a Bolt: %d deaths, want 1", n)
		}
	})
	advanceToUpkeepOf(t, g, 1)
	g.WithWriteLock(func() {
		if n := b11CreaturesDiedThisTurn(g); n != 0 {
			t.Errorf("next turn: %d deaths, want 0", n)
		}
	})
}
