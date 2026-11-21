package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// b11Creature seeds a non-catalog creature with a type line, able to
// attack (EnteredBattlefieldAt is zero, so no summoning sickness).
func b11Creature(g *game.Game, owner uuid.UUID, name, typeLine string, power, toughness int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// b11BoltCreature kills a creature with toughness 3 or less through
// a real Lightning Bolt cast by the active seat.
func b11BoltCreature(t *testing.T, g *game.Game, target uuid.UUID) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: target}})
	effectstest.PassPriorityAroundTable(t, g)
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
		effectstest.B11MillikinOracle:            "Millikin",
		effectstest.B11GlaringFleshrakerOracle:   "Glaring Fleshraker",
		effectstest.B11IronMyrOracle:             "Iron Myr",
		effectstest.B11SilverMyrOracle:           "Silver Myr",
		effectstest.B11AbandonedAirTempleOracle:  "Abandoned Air Temple",
		effectstest.B11MagdaOracle:               "Magda, Brazen Outlaw",
		effectstest.B11ExemplarOfLightOracle:     "Exemplar of Light",
		effectstest.B11SecureTheWastesOracle:     "Secure the Wastes",
		effectstest.B11RosieCottonOracle:         "Rosie Cotton of South Lane",
		effectstest.B11FractureOracle:            "Fracture",
		effectstest.B11WightOfTheReliquaryOracle: "Wight of the Reliquary",
		effectstest.B11KeeperOfTheAccordOracle:   "Keeper of the Accord",
		effectstest.B11MahadiOracle:              "Mahadi, Emporium Master",
		effectstest.B11BalefireDragonOracle:      "Balefire Dragon",
		effectstest.B11MaestrosTheaterOracle:     "Maestros Theater",
		effectstest.B11PeerIntoTheAbyssOracle:    "Peer into the Abyss",
		effectstest.B11HauntedMireOracle:         "Haunted Mire",
		effectstest.B11SylvanScryingOracle:       "Sylvan Scrying",
		effectstest.B11RuinCrabOracle:            "Ruin Crab",
		effectstest.B11AmuletOfVigorOracle:       "Amulet of Vigor",
		effectstest.B11NukaColaOracle:            "Nuka-Cola Vending Machine",
		effectstest.B11RuneScarredDemonOracle:    "Rune-Scarred Demon",
		effectstest.B11DefenseOfTheHeartOracle:   "Defense of the Heart",
		effectstest.B11TendershootDryadOracle:    "Tendershoot Dryad",
		effectstest.B11GuildlessCommonsOracle:    "Guildless Commons",
		effectstest.B11BrokersHideoutOracle:      "Brokers Hideout",
		effectstest.B11BaneOfProgressAlreadyOID:  "Bane of Progress",
		effectstest.B11VorinclexAlreadyOID:       "Vorinclex, Monstrous Raider",
		effectstest.B11CleansingNovaAlreadyOID:   "Cleansing Nova",
		effectstest.B11KambalAlreadyOID:          "Kambal, Consul of Allocation",
	}
	if len(want) != 30 {
		t.Fatalf("the batch registers 26 cards plus four already on main, the table lists %d", len(want))
	}
	alreadyOnMain := map[string]bool{
		"Bane of Progress": true, "Vorinclex, Monstrous Raider": true,
		"Cleansing Nova": true, "Kambal, Consul of Allocation": true,
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
		if spec.Completeness == effects.CompletenessUnreviewed && !alreadyOnMain[name] {
			t.Errorf("%s ships without a Completeness declaration", name)
		}
	}
}

// --- Millikin and the Myr -----------------------------------------

func TestB11MillikinMillsACardForColorlessAndNeedsALibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	millikin := effectstest.PushCatalogPermanent(g, me.ID, "Millikin", "Artifact Creature — Construct", effectstest.B11MillikinOracle, false)
	lib, yard := me.Library.Size(), me.Graveyard.Size()

	// Not a mana ability (#3971, CR 605.1a): the ability stacks and the
	// {C} arrives when it resolves.
	if err := g.ActivateCatalogAbility(me.ID, millikin, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if len(g.StackMeta) != 1 || me.ManaPool.Total() != 0 {
		t.Fatalf("stack %d, pool %d: want the ability stacked and no mana yet", len(g.StackMeta), me.ManaPool.Total())
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
	if me.Library.Size() != lib-1 || me.Graveyard.Size() != yard+1 {
		t.Errorf("library %d → %d, graveyard %d → %d: want one card milled", lib, me.Library.Size(), yard, me.Graveyard.Size())
	}
	if c, _ := effectstest.BattlefieldCard(g, millikin); !c.Tapped {
		t.Error("Millikin should be tapped")
	}

	// With nothing to mill the cost cannot be paid, so the ability
	// cannot be activated and the Myr stays untapped.
	effectstest.B02bSetTapped(g, millikin, false)
	me.Library.Cards = nil
	if err := g.ActivateCatalogAbility(me.ID, millikin, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("Millikin activated with an empty library")
	}
	if c, _ := effectstest.BattlefieldCard(g, millikin); c.Tapped {
		t.Error("a refused activation must not tap the source")
	}
}

func TestB11MillikinIsSummoningSick(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	millikin := effectstest.PushCatalogPermanent(g, me.ID, "Millikin", "Artifact Creature — Construct", effectstest.B11MillikinOracle, true)
	if err := g.ActivateCatalogAbility(me.ID, millikin, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("a summoning-sick Millikin paid its {T}")
	}
}

func TestB11ManaMyrTapForTheirColour(t *testing.T) {
	for _, tc := range []struct{ name, oracle, color string }{
		{"Iron Myr", effectstest.B11IronMyrOracle, "R"},
		{"Silver Myr", effectstest.B11SilverMyrOracle, "U"},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		myr := effectstest.PushCatalogPermanent(g, me.ID, tc.name, "Artifact Creature — Myr", tc.oracle, false)
		if err := g.ActivateManaAbility(me.ID, myr, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("%s: activate: %v", tc.name, err)
		}
		if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != tc.color {
			t.Errorf("%s: pool %v, want [%s]", tc.name, got, tc.color)
		}
		sick := effectstest.PushCatalogPermanent(g, me.ID, tc.name, "Artifact Creature — Myr", tc.oracle, true)
		if err := g.ActivateManaAbility(me.ID, sick, 0, game.ManaAbilityParams{}); err == nil {
			t.Errorf("%s: a summoning-sick Myr tapped for mana", tc.name)
		}
	}
}

// --- Glaring Fleshraker -------------------------------------------

func TestB11GlaringFleshrakerSpawnsOnColorlessSpellsAndPingsOnColorlessCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B11Push(g, me.ID, "Glaring Fleshraker", "Creature — Eldrazi Drone", effectstest.B11GlaringFleshrakerOracle, 2, 2)
	before := effectstest.LifeOfOpponents(g)

	// A colorless spell: a Spawn, whose own entry is a colorless
	// creature entering, so every opponent takes 1.
	effectstest.CastCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Eldrazi Spawn"); n != 1 {
		t.Fatalf("a colorless spell made %d Spawn, want 1", n)
	}
	for i, opp := range g.Seats[1:] {
		if opp.Life != before[i]-1 {
			t.Errorf("opponent %d: %d → %d, want -1 from the Spawn entering", i+1, before[i], opp.Life)
		}
	}

	// A coloured spell makes nothing, and a coloured creature entering
	// pings nobody.
	effectstest.CastWithCost(t, g, "Bear", "Creature — Bear", "{1}{G}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Eldrazi Spawn"); n != 1 {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLandOnBattlefield(g, me.ID, "Command Tower", "Land")
	tapped := effectstest.PlayLandFromHand(t, g, "Abandoned Air Temple", effectstest.B11AbandonedAirTempleOracle)
	effectstest.Top100AssertEnteredTapped(t, g, tapped, "Abandoned Air Temple without a basic")

	effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	untapped := effectstest.PlayLandFromHand(t, g, "Abandoned Air Temple", effectstest.B11AbandonedAirTempleOracle)
	effectstest.Top100AssertEnteredUntapped(t, g, untapped, "Abandoned Air Temple with a basic Forest")

	spec, _ := effects.Lookup(effectstest.B11AbandonedAirTempleOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{W}" {
		t.Errorf("mana abilities %+v, want one producing {W}", spec.ManaAbilities)
	}
}

func TestB11AbandonedAirTemplePutsACounterOnEachOfYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	temple := effectstest.PushCatalogPermanent(g, me.ID, "Abandoned Air Temple", "Land", effectstest.B11AbandonedAirTempleOracle, false)
	a := effectstest.SeedCreature(g, "Bear A", me.ID)
	b := effectstest.SeedCreature(g, "Bear B", me.ID)
	theirs := effectstest.SeedCreature(g, "Their Bear", opp.ID)
	me.ManaPool.AddMana(game.ManaToken{Color: "W"})
	effectstest.B09AddColorless(me, 3)

	if err := g.ActivateCatalogAbility(me.ID, temple, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if c, _ := effectstest.BattlefieldCard(g, temple); !c.Tapped {
		t.Fatal("the ability has a tap cost")
	}
	if effectstest.CountersOn(g, a, "+1/+1") != 0 {
		t.Fatal("the counters must wait for the ability to resolve")
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if got := effectstest.CountersOn(g, id, "+1/+1"); got != 1 {
			t.Errorf("my creature has %d +1/+1 counters, want 1", got)
		}
	}
	if got := effectstest.CountersOn(g, theirs, "+1/+1"); got != 0 {
		t.Errorf("an opponent's creature got %d counters", got)
	}
}

// --- Magda, Brazen Outlaw -----------------------------------------

func TestB11MagdaPumpsOtherDwarvesAndPaysATreasurePerTappedDwarf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	magda := effectstest.B11Push(g, me.ID, "Magda, Brazen Outlaw", "Legendary Creature — Dwarf Berserker", effectstest.B11MagdaOracle, 2, 1)
	dwarf := b11Creature(g, me.ID, "Dwarf", "Creature — Dwarf Warrior", 2, 2)
	bear := b11Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirDwarf := b11Creature(g, opp.ID, "Their Dwarf", "Creature — Dwarf", 2, 2)

	if p := effectstest.EffectivePower(t, g, dwarf); p != 3 {
		t.Errorf("another Dwarf = %d power, want 3", p)
	}
	if p := effectstest.EffectivePower(t, g, magda); p != 2 {
		t.Errorf("Magda pumps herself: %d power, want 2", p)
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("a non-Dwarf = %d power, want 2", p)
	}
	if p := effectstest.EffectivePower(t, g, theirDwarf); p != 2 {
		t.Errorf("an opponent's Dwarf = %d power, want 2", p)
	}

	// Magda and the Dwarf attack — two Dwarves tapped, two Treasures;
	// the Bear is not a Dwarf.
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{magda, dwarf, bear} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	// #859: the declaration is announced at its lock-in — the
	// priority wrap inside declare_attackers — so the attack triggers
	// exist only after this.
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Treasure"); n != 2 {
		t.Fatalf("three attackers, two of them Dwarves, made %d Treasures, want 2", n)
	}
}

func TestB11MagdaIgnoresAVigilantDwarfAndAnOpponentsTap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B11Push(g, me.ID, "Magda, Brazen Outlaw", "Legendary Creature — Dwarf Berserker", effectstest.B11MagdaOracle, 2, 1)
	vigilant := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Watchful Dwarf", TypeLine: "Creature — Dwarf",
		Power: 2, Toughness: 2, Keywords: []string{"vigilance"}, Owner: me.ID, Controller: me.ID,
	})
	theirDwarf := b11Creature(g, opp.ID, "Their Dwarf", "Creature — Dwarf", 2, 2)

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(vigilant, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Treasure"); n != 0 {
		t.Errorf("a vigilance Dwarf attacking made %d Treasures; it never became tapped", n)
	}

	// An opponent's Dwarf tapping is not "a Dwarf you control".
	if err := g.TapCard(theirDwarf, true); err != nil {
		t.Fatalf("TapCard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Treasure"); n != 0 {
		t.Errorf("an opponent's Dwarf tapping made %d Treasures", n)
	}
}

func TestB11MagdaPaysForADwarfTappedOutsideCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B11Push(g, me.ID, "Magda, Brazen Outlaw", "Legendary Creature — Dwarf Berserker", effectstest.B11MagdaOracle, 2, 1)
	dwarf := b11Creature(g, me.ID, "Dwarf", "Creature — Dwarf", 2, 2)
	if err := g.TapCard(dwarf, true); err != nil {
		t.Fatalf("TapCard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("a Dwarf tapped outside combat made %d Treasures, want 1", n)
	}
	// #747: the five-Treasure tutor ships at its printed count; its
	// engine test is in sacrifice_n_cards_test.go.
	spec, _ := effects.Lookup(effectstest.B11MagdaOracle)
	if len(spec.Activated) != 1 || spec.Completeness != effects.CompletenessFull {
		t.Error("the five-Treasure tutor ships whole")
	}
}

// --- Exemplar of Light --------------------------------------------

func TestB11ExemplarOfLightGrowsOnLifegainAndDrawsOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	exemplar := effectstest.B11Push(g, me.ID, "Exemplar of Light", "Creature — Angel", effectstest.B11ExemplarOfLightOracle, 3, 3)
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, exemplar), "flying") {
		t.Error("printed flying did not reach the effective abilities")
	}

	// castCatalogSpell adds the card it then casts, so the hand is
	// neutral across a cast and the difference is exactly the draws.
	before := me.Hand.Size()
	effectstest.B11Helix(t, g, opp.ID)
	if got := effectstest.CountersOn(g, exemplar, "+1/+1"); got != 1 {
		t.Fatalf("after gaining 3: %d +1/+1 counters, want 1", got)
	}
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("first counter of the turn: hand %d → %d, want +1", before, got)
	}

	before = me.Hand.Size()
	effectstest.B11Helix(t, g, opp.ID)
	if got := effectstest.CountersOn(g, exemplar, "+1/+1"); got != 2 {
		t.Fatalf("after gaining 3 again: %d counters, want 2", got)
	}
	if got := me.Hand.Size(); got != before {
		t.Errorf("second counter of the turn drew: hand %d → %d", before, got)
	}

	// A new turn resets the once-per-turn clause. (Measure from the
	// main phase so the turn's draw step is not counted.)
	b11ToMyNextUpkeep(t, g)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	before = me.Hand.Size()
	effectstest.B11Helix(t, g, opp.ID)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("next turn's first counter: hand %d → %d, want +1", before, got)
	}
}

func TestB11ExemplarOfLightDoesNotDrawForAnOpponentsCounterOrARemoval(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	exemplar := effectstest.B11Push(g, me.ID, "Exemplar of Light", "Creature — Angel", effectstest.B11ExemplarOfLightOracle, 3, 3)

	// The opponent puts a counter on the Exemplar: that is their
	// placement, not yours. (This used to be staged as "the opponent
	// resolves something, then an unattributed counter lands", which
	// credited the opponent only because the old reading took the last
	// resolver in the log, long after that resolution had ended, #3942.)
	before := me.Hand.Size()
	if err := g.AddCounterBy(opp.ID, exemplar, "+1/+1", 1); err != nil {
		t.Fatalf("AddCounterBy: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before {
		t.Errorf("an opponent's counter drew: hand %d → %d", before, got)
	}

	// Removing a counter is not placing one.
	b11ToMyNextUpkeep(t, g)
	effectstest.CastCatalogSpell(t, g, "Nothing", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	before = me.Hand.Size()
	if err := g.AddCounter(exemplar, "+1/+1", -1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before {
		t.Errorf("a counter removal drew: hand %d → %d", before, got)
	}
}

// --- Secure the Wastes --------------------------------------------

func TestB11SecureTheWastesMakesXWarriors(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastXSpell(t, g, "Secure the Wastes", "Instant", effectstest.B11SecureTheWastesOracle, "{X}{W}", 3, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Warrior"); n != 3 {
		t.Fatalf("X=3 made %d Warriors, want 3", n)
	}
	w := effectstest.FindBattlefieldByName(g, "Warrior")
	if c, _ := effectstest.BattlefieldCard(g, w); !c.HasColor("W") || c.Power != 1 || c.Toughness != 1 {
		t.Errorf("the Warrior should be a 1/1 white token, got %+v", c)
	}
}

// --- Rosie Cotton of South Lane -----------------------------------

func TestB11RosieCottonMakesAFoodAndPutsTheCounterOnAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	rosie := effectstest.CastAndResolveCreature(t, g, "Rosie Cotton of South Lane", "Legendary Creature — Halfling Peasant", effectstest.B11RosieCottonOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Food"); n != 1 {
		t.Fatalf("Rosie made %d Food, want 1", n)
	}
	// Her own Food is a token you created, so the second ability asks
	// for a creature — and Rosie herself is not on offer.
	prompt := effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("the token trigger must ask which creature gets the counter")
	}
	if effectstest.HasID(prompt.PickTargetCards, rosie) {
		t.Error("Rosie offered herself as the target")
	}
	if !effectstest.HasID(prompt.PickTargetCards, bear) {
		t.Fatal("the other creature should be a legal target")
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, bear, "+1/+1"); got != 1 {
		t.Errorf("the Bear has %d +1/+1 counters, want 1", got)
	}
	if got := effectstest.CountersOn(g, rosie, "+1/+1"); got != 0 {
		t.Errorf("Rosie has %d counters, want 0", got)
	}
}

func TestB11RosieCottonAloneDropsTheTargetedTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastAndResolveCreature(t, g, "Rosie Cotton of South Lane", "Legendary Creature — Halfling Peasant", effectstest.B11RosieCottonOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Food"); n != 1 {
		t.Fatalf("Rosie made %d Food, want 1", n)
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("with no other creature the trigger has no legal target and must not prompt")
	}
}

// --- Fracture -----------------------------------------------------

func TestB11FractureDestroysAPlaneswalkerAndRefusesACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	// A planeswalker needs loyalty, or the state check buries it
	// before the spell is cast.
	walker := effectstest.PushWalkerForTest(g, opp.ID, "Jace", "", 3)
	effectstest.CastCatalogSpell(t, g, "Fracture", "Instant", effectstest.B11FractureOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: walker}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(walker) || !opp.Graveyard.Contains(walker) {
		t.Error("the planeswalker should be in its owner's graveyard")
	}

	bear := effectstest.SeedCreature(g, "Bear", opp.ID)
	if err := b09TryCast(t, g, "Fracture", "Instant", effectstest.B11FractureOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}}); err == nil {
		t.Error("a creature was accepted for 'target artifact, enchantment, or planeswalker'")
	}
}

// --- Wight of the Reliquary ---------------------------------------

func TestB11WightOfTheReliquaryGrowsWithDeadCreaturesAndRampsOffAnother(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	wight := effectstest.B11Push(g, me.ID, "Wight of the Reliquary", "Creature — Zombie Knight", effectstest.B11WightOfTheReliquaryOracle, 2, 2)
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, wight), "vigilance") {
		t.Error("printed vigilance did not reach the effective abilities")
	}
	if p := effectstest.EffectivePower(t, g, wight); p != 2 {
		t.Fatalf("empty graveyard: %d power, want 2", p)
	}
	effectstest.PushGraveyardCardWithTypeLine(g, me.ID, "Dead Bear", "Creature — Bear")
	effectstest.PushGraveyardCardWithTypeLine(g, me.ID, "Old Rock", "Artifact")
	if p := effectstest.EffectivePower(t, g, wight); p != 3 {
		t.Errorf("one creature card in the graveyard: %d power, want 3", p)
	}

	fodder := effectstest.SeedCreature(g, "Fodder", me.ID)
	effectstest.SeedSearchLibrary(me,
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
	if c, _ := effectstest.BattlefieldCard(g, wight); !c.Tapped {
		t.Error("the ability has a tap cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Bear Card") != uuid.Nil {
		t.Error("a non-land was offered")
	}
	if effectstest.SearchOptionNamed(g, c, "Bayou") == uuid.Nil {
		t.Error("any land card qualifies, not only basics")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Bayou")
	land, ok := effectstest.SearchFetchedCard(g, "Bayou")
	if !ok || !land.Tapped {
		t.Errorf("the land should be on the battlefield tapped, got %+v (found %v)", land, ok)
	}
	// The sacrificed creature is now a creature card in the graveyard.
	if p := effectstest.EffectivePower(t, g, wight); p != 4 {
		t.Errorf("two creature cards in the graveyard: %d power, want 4", p)
	}
}

// --- Keeper of the Accord -----------------------------------------

func TestB11KeeperOfTheAccordCatchesUpAtARicherOpponentsEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, bystander := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.B11Push(g, me.ID, "Keeper of the Accord", "Creature — Human Soldier", effectstest.B11KeeperOfTheAccordOracle, 3, 4)
	effectstest.SeedCreature(g, "Their Bear A", opp.ID)
	effectstest.SeedCreature(g, "Their Bear B", opp.ID)
	effectstest.SeedLandOnBattlefield(g, opp.ID, "Plains", "Basic Land — Plains")
	effectstest.SeedLandOnBattlefield(g, opp.ID, "Plains", "Basic Land — Plains")
	effectstest.SeedLandOnBattlefield(g, me.ID, "Plains", "Basic Land — Plains")
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Plains A", TypeLine: "Basic Land — Plains"},
		game.Card{Name: "Snowfield", TypeLine: "Land — Plains"},
		game.Card{Name: "Forest A", TypeLine: "Basic Land — Forest"},
	)

	// Opponent 1 has more creatures (2 vs 1) and more lands (2 vs 1):
	// both triggers fire at their end step, and the controller orders
	// them.
	effectstest.AdvanceToEndStepOf(t, g, 1)
	if !effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID) {
		t.Fatal("two Keeper triggers should ask for an order")
	}
	for i := 0; i < 16 && effectstest.SearchChoiceFor(g, me.ID) == nil && !effectstest.StackFullyEmpty(g); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the land trigger should offer a basic Plains search")
	}
	if effectstest.SearchOptionNamed(g, c, "Snowfield") != uuid.Nil || effectstest.SearchOptionNamed(g, c, "Forest A") != uuid.Nil {
		t.Error("only a BASIC Plains card is on offer")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Plains A")
	effectstest.PassPriorityAroundTable(t, g)
	if plains, ok := effectstest.SearchFetchedCard(g, "Plains A"); !ok || !plains.Tapped {
		t.Errorf("the Plains should be on the battlefield tapped, got %+v (found %v)", plains, ok)
	}
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Soldier"); n != 1 {
		t.Errorf("%d Soldiers, want 1", n)
	}

	// Opponent 2 controls nothing, so their end step is quiet — and
	// so is the controller's own.
	effectstest.AdvanceToEndStepOf(t, g, 2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Soldier"); n != 1 {
		t.Errorf("a poorer opponent's end step made a Soldier (%d total)", n)
	}
	_ = bystander
}

// --- Mahadi, Emporium Master --------------------------------------

func TestB11MahadiPaysATreasurePerCreatureThatDiedThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B11Push(g, me.ID, "Mahadi, Emporium Master", "Legendary Creature — Devil", effectstest.B11MahadiOracle, 3, 3)
	mine := effectstest.SeedCreature(g, "My Bear", me.ID)
	theirs := effectstest.SeedCreature(g, "Their Bear", opp.ID)
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(opp.ID, effects.RedGoblinToken(), 1)
	})
	goblin := effectstest.FindBattlefieldByName(g, "Goblin")
	rock := effectstest.SeedPermanentFor(g, opp.ID, "Rock", "Artifact")

	b11BoltCreature(t, g, mine)
	b11BoltCreature(t, g, theirs)
	b11BoltCreature(t, g, goblin)
	effectstest.CastCatalogSpell(t, g, "Fracture", "Instant", effectstest.B11FractureOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Treasure"); n != 3 {
		t.Fatalf("three creatures (one a token) and an artifact died: %d Treasures, want 3", n)
	}

	// Next turn, nothing has died.
	b11ToMyNextUpkeep(t, g)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Treasure"); n != 3 {
		t.Errorf("a turn with no deaths made Treasures (%d total)", n)
	}
}

// --- Balefire Dragon ----------------------------------------------

func TestB11BalefireDragonWipesTheConnectedPlayersCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim, bystander := g.Seats[0], g.Seats[1], g.Seats[2]
	dragon := effectstest.B11Push(g, me.ID, "Balefire Dragon", "Creature — Dragon", effectstest.B11BalefireDragonOracle, 6, 6)
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, dragon), "flying") {
		t.Error("printed flying did not reach the effective abilities")
	}
	small := b11Creature(g, victim.ID, "Small", "Creature — Bear", 2, 2)
	exact := b11Creature(g, victim.ID, "Exact", "Creature — Giant", 6, 6)
	big := b11Creature(g, victim.ID, "Big", "Creature — Wurm", 7, 7)
	spared := b11Creature(g, bystander.ID, "Spared", "Creature — Bear", 2, 2)
	before := victim.Life

	effectstest.AttackWith(t, g, victim.ID, dragon)
	if victim.Life != before-6 {
		t.Fatalf("victim %d → %d, want -6", before, victim.Life)
	}
	if effectstest.TriggerOnStack(g, dragon) == nil {
		t.Fatal("the rider is a trigger with a response window")
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{small, exact} {
		if g.Battlefield.Contains(id) {
			t.Errorf("a creature with toughness ≤ 6 survived 6 damage")
		}
	}
	if !g.Battlefield.Contains(big) || effectstest.DamageMarkedOn(g, big) != 6 {
		t.Error("the 7/7 should survive with 6 damage marked")
	}
	if !g.Battlefield.Contains(spared) || effectstest.DamageMarkedOn(g, spared) != 0 {
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
		{"Maestros Theater", effectstest.B11MaestrosTheaterOracle, []string{"Island A", "Swamp A", "Mountain A"}, "Forest A"},
		{"Brokers Hideout", effectstest.B11BrokersHideoutOracle, []string{"Forest A", "Plains A", "Island A"}, "Swamp A"},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.SeedSearchLibrary(me,
			game.Card{Name: "Island A", TypeLine: "Basic Land — Island"},
			game.Card{Name: "Swamp A", TypeLine: "Basic Land — Swamp"},
			game.Card{Name: "Mountain A", TypeLine: "Basic Land — Mountain"},
			game.Card{Name: "Forest A", TypeLine: "Basic Land — Forest"},
			game.Card{Name: "Plains A", TypeLine: "Basic Land — Plains"},
			game.Card{Name: "Bayou", TypeLine: "Land — Swamp Forest"},
		)
		before := me.Life
		land := effectstest.PlayLandFromHand(t, g, tc.name, tc.oracle)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(land) || !me.Graveyard.Contains(land) {
			t.Fatalf("%s should have sacrificed itself", tc.name)
		}
		c := effectstest.SearchChoiceFor(g, me.ID)
		if c == nil {
			t.Fatalf("%s: no search prompt", tc.name)
		}
		for _, name := range tc.offered {
			if effectstest.SearchOptionNamed(g, c, name) == uuid.Nil {
				t.Errorf("%s: %s should be offered", tc.name, name)
			}
		}
		if effectstest.SearchOptionNamed(g, c, tc.refused) != uuid.Nil || effectstest.SearchOptionNamed(g, c, "Bayou") != uuid.Nil {
			t.Errorf("%s: an off-colour or nonbasic land was offered", tc.name)
		}
		effectstest.AnswerSearchNamed(t, g, me.ID, tc.offered[0])
		fetched, ok := effectstest.SearchFetchedCard(g, tc.offered[0])
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
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.SeedSearchLibrary(opp,
		game.Card{Name: "a"}, game.Card{Name: "b"}, game.Card{Name: "c"}, game.Card{Name: "d"},
		game.Card{Name: "e"}, game.Card{Name: "f"}, game.Card{Name: "g"},
	)
	opp.Life = 21
	hand := opp.Hand.Size()

	effectstest.CastCatalogSpell(t, g, "Peer into the Abyss", "Sorcery", effectstest.B11PeerIntoTheAbyssOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	id := effectstest.PlayLandFromHand(t, g, "Haunted Mire", effectstest.B11HauntedMireOracle)
	effectstest.Top100AssertEnteredTapped(t, g, id, "Haunted Mire")
	spec, _ := effects.Lookup(effectstest.B11HauntedMireOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{B|G}" {
		t.Errorf("row produces %+v, want {B|G}", spec.ManaAbilities)
	}
}

// #1337: the bounce is a resolution-time choice (ReturnOneYouControl
// / ChoosePermanents), not a target picked when the trigger goes on
// the stack.
func TestB11GuildlessCommonsIsAColourlessKaroo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	other := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	commons := effectstest.PlayLandFromHand(t, g, "Guildless Commons", effectstest.B11GuildlessCommonsOracle)
	effectstest.Top100AssertEnteredTapped(t, g, commons, "Guildless Commons")
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if prompt == nil {
		t.Fatal("the bounce trigger must ask which land to return")
	}
	if err := g.ResolveOwnPermanents(prompt.ID, me.ID, []uuid.UUID{other}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	if g.Battlefield.Contains(other) || !me.Hand.Contains(other) {
		t.Error("the chosen land should be back in hand")
	}
	spec, _ := effects.Lookup(effectstest.B11GuildlessCommonsOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{C}{C}" {
		t.Errorf("row produces %+v, want {C}{C}", spec.ManaAbilities)
	}
}

// --- Sylvan Scrying -----------------------------------------------

func TestB11SylvanScryingTutorsAnyLandToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Forest A", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Cradle", TypeLine: "Legendary Land"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	effectstest.CastCatalogSpell(t, g, "Sylvan Scrying", "Sorcery", effectstest.B11SylvanScryingOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Bear") != uuid.Nil {
		t.Error("a non-land was offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Cradle")
	if !effectstest.B02bHandHasNamed(me, "Cradle") {
		t.Error("the chosen land did not reach the hand")
	}
}

// --- Ruin Crab ----------------------------------------------------

func TestB11RuinCrabMillsEachOpponentThreeOnLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B11Push(g, me.ID, "Ruin Crab", "Creature — Crab", effectstest.B11RuinCrabOracle, 0, 3)
	var libs, yards []int
	for _, opp := range g.Seats[1:] {
		libs = append(libs, opp.Library.Size())
		yards = append(yards, opp.Graveyard.Size())
	}
	myLib := me.Library.Size()

	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Amulet of Vigor", "Artifact", effectstest.B11AmuletOfVigorOracle, false)

	mire := effectstest.PlayLandFromHand(t, g, "Haunted Mire", effectstest.B11HauntedMireOracle)
	if c, _ := effectstest.BattlefieldCard(g, mire); !c.Tapped {
		t.Fatal("Haunted Mire enters tapped")
	}
	if effectstest.TriggerOnStack(g, effectstest.FindBattlefieldByName(g, "Amulet of Vigor")) == nil {
		t.Fatal("the untap is a trigger with a response window")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, mire); c.Tapped {
		t.Error("the Amulet should have untapped the Mire")
	}

	// A tapped token counts too; an opponent's tapped entry does not.
	g.WithWriteLock(func() {
		_, _ = g.CreateTokensForEffect(me.ID, effects.TreasureToken(), 1, game.TokenEntryOptions{Tapped: true})
		_, _ = g.CreateTokensForEffect(opp.ID, effects.TreasureToken(), 1, game.TokenEntryOptions{Tapped: true})
	})
	effectstest.PassPriorityAroundTable(t, g)
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
	forest := effectstest.PlayLandFromHand(t, g, "Forest", "")
	if effectstest.TriggersOnStackFrom(g, effectstest.FindBattlefieldByName(g, "Amulet of Vigor")) != 0 {
		t.Error("an untapped entry queued an Amulet trigger")
	}
	if c, _ := effectstest.BattlefieldCard(g, forest); c.Tapped {
		t.Error("the Forest should be untapped")
	}
}

// --- Nuka-Cola Vending Machine ------------------------------------

func TestB11NukaColaMakesFoodAndATappedTreasurePerFoodSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	machine := effectstest.PushCatalogPermanent(g, me.ID, "Nuka-Cola Vending Machine", "Artifact", effectstest.B11NukaColaOracle, false)
	effectstest.B09AddColorless(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, machine, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	food := effectstest.FindBattlefieldByName(g, "Food")
	if food == uuid.Nil {
		t.Fatal("no Food was made")
	}

	// Cracking the Food (its own {2}, {T}, sacrifice ability) is a
	// Food sacrificed: a tapped Treasure, and the life.
	before := me.Life
	effectstest.B09AddColorless(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, food, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("crack the Food: %v", err)
	}
	if g.Battlefield.Contains(food) {
		t.Fatal("the Food is sacrificed as a cost")
	}
	if effectstest.TriggerOnStack(g, machine) == nil {
		t.Fatal("the Treasure trigger should be on the stack above the Food's ability")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+3 {
		t.Errorf("life %d → %d, want +3", before, me.Life)
	}
	treasure := effectstest.FindBattlefieldByName(g, "Treasure")
	if treasure == uuid.Nil {
		t.Fatal("no Treasure was made")
	}
	if c, _ := effectstest.BattlefieldCard(g, treasure); !c.Tapped {
		t.Error("the Treasure enters tapped")
	}

	// Sacrificing a non-Food makes nothing.
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	if err := g.SacrificePermanent(me.ID, bear); err != nil {
		t.Fatalf("SacrificePermanent: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("a non-Food sacrifice made a Treasure (%d total)", n)
	}
}

// --- Rune-Scarred Demon -------------------------------------------

func TestB11RuneScarredDemonTutorsAnyCardOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Forest A", TypeLine: "Basic Land — Forest"},
	)
	demon := effectstest.CastAndResolveCreature(t, g, "Rune-Scarred Demon", "Creature — Demon", effectstest.B11RuneScarredDemonOracle)
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, demon), "flying") {
		t.Error("printed flying did not reach the effective abilities")
	}
	if effectstest.TriggerOnStack(g, demon) == nil {
		t.Fatal("the search is an ETB trigger with a response window")
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Bear") == uuid.Nil || effectstest.SearchOptionNamed(g, c, "Forest A") == uuid.Nil {
		t.Error("every card in the library is on offer")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Forest A")
	if !effectstest.B02bHandHasNamed(me, "Forest A") {
		t.Error("the chosen card did not reach the hand")
	}
}

// --- Defense of the Heart -----------------------------------------

func TestB11DefenseOfTheHeartFiresAgainstThreeCreaturesAndFetchesTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	defense := effectstest.B11Push(g, me.ID, "Defense of the Heart", "Enchantment", effectstest.B11DefenseOfTheHeartOracle, 0, 0)
	effectstest.SeedCreature(g, "Their Bear A", opp.ID)
	effectstest.SeedCreature(g, "Their Bear B", opp.ID)

	// Two creatures: the intervening-if fails and nothing happens.
	b11ToMyNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, me.ID) != nil || !g.Battlefield.Contains(defense) {
		t.Fatal("Defense of the Heart fired against two creatures")
	}

	effectstest.SeedCreature(g, "Their Bear C", opp.ID)
	// Seeded top-first under a filler card, which the draw step takes
	// on the way round, leaving more matches than the search's limit
	// so the chooser is asked.
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Filler", TypeLine: "Sorcery"},
		game.Card{Name: "Big A", TypeLine: "Creature — Wurm"},
		game.Card{Name: "Big B", TypeLine: "Creature — Dragon"},
		game.Card{Name: "Big C", TypeLine: "Creature — Giant"},
		game.Card{Name: "Rock", TypeLine: "Artifact"},
	)
	b11ToMyNextUpkeep(t, g)
	if effectstest.TriggerOnStack(g, defense) == nil {
		t.Fatal("three opposing creatures: the trigger should be on the stack")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(defense) || !me.Graveyard.Contains(defense) {
		t.Error("the enchantment should have been sacrificed")
	}
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if c.SearchMax != 2 {
		t.Errorf("search max %d, want 2", c.SearchMax)
	}
	if effectstest.SearchOptionNamed(g, c, "Rock") != uuid.Nil {
		t.Error("a non-creature was offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Big A", "Big C")
	for _, name := range []string{"Big A", "Big C"} {
		card, ok := effectstest.SearchFetchedCard(g, name)
		if !ok || card.Controller != me.ID {
			t.Errorf("%s should be on the battlefield under my control", name)
		} else if card.Tapped {
			t.Errorf("%s entered tapped; the printed card says nothing about tapped", name)
		}
	}
}

// --- Tendershoot Dryad --------------------------------------------

func TestB11TendershootDryadMakesASaprolingEveryUpkeepAndPumpsAtTen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B11Push(g, me.ID, "Tendershoot Dryad", "Creature — Dryad", effectstest.B11TendershootDryadOracle, 2, 2)

	// Each player's upkeep: the next opponent's, then the controller's.
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Saproling"); n != 1 {
		t.Fatalf("an opponent's upkeep made %d Saprolings, want 1", n)
	}
	b11ToMyNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Saproling"); n != 4 {
		t.Fatalf("after every seat's upkeep: %d Saprolings, want 4", n)
	}
	sap := effectstest.FindBattlefieldByName(g, "Saproling")
	if p := effectstest.EffectivePower(t, g, sap); p != 1 {
		t.Errorf("five permanents: Saproling power %d, want 1", p)
	}

	// Ten permanents: the Dryad, four Saprolings, five lands.
	for i := 0; i < 5; i++ {
		effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	}
	g.RunStateChecksForTest()
	if p := effectstest.EffectivePower(t, g, sap); p != 3 {
		t.Errorf("ten permanents: Saproling power %d, want 3", p)
	}

	// The city's blessing is kept (CR 702.131c, #2696): the bonus stays
	// when the board drops back under ten.
	g.WithWriteLock(func() {
		removed := 0
		for i := len(g.Battlefield.Cards) - 1; i >= 0 && removed < 5; i-- {
			if c := g.Battlefield.Cards[i]; c.Controller == me.ID && c.Name == "Forest" {
				_, _ = g.Battlefield.Remove(c.InstanceID)
				removed++
			}
		}
		g.BumpLayerVersionForTest()
	})
	if p := effectstest.EffectivePower(t, g, sap); p != 3 {
		t.Errorf("back under ten: Saproling power %d, want 3 (the blessing is kept)", p)
	}
}

// --- the once-per-turn and died-this-turn walks -------------------

func TestB11CreaturesDiedThisTurnResetsAtUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	g.WithWriteLock(func() {
		if n := effects.B11CreaturesDiedThisTurn(g); n != 0 {
			t.Errorf("fresh game: %d deaths, want 0", n)
		}
	})
	b11BoltCreature(t, g, bear)
	g.WithWriteLock(func() {
		if n := effects.B11CreaturesDiedThisTurn(g); n != 1 {
			t.Errorf("after a Bolt: %d deaths, want 1", n)
		}
	})
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	g.WithWriteLock(func() {
		if n := effects.B11CreaturesDiedThisTurn(g); n != 0 {
			t.Errorf("next turn: %d deaths, want 0", n)
		}
	})
}
