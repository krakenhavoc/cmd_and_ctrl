package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. The land
// cycles are loops over tables, so a transposed row is invisible
// until someone plays that exact card.
func TestBatch01CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.ArcaneDenialOracle:            "Arcane Denial",
		effectstest.ManaDrainOracle:               "Mana Drain",
		effectstest.VictimizeOracle:               "Victimize",
		effectstest.SpectatorSeatingOracle:        "Spectator Seating",
		effectstest.VaultOfChampionsOracle:        "Vault of Champions",
		effectstest.FarewellOracle:                "Farewell",
		effectstest.UndergrowthStadiumOracle:      "Undergrowth Stadium",
		effectstest.DreamrootCascadeOracle:        "Dreamroot Cascade",
		effectstest.TirelessProvisionerOracle:     "Tireless Provisioner",
		effectstest.StormcarvedCoastOracle:        "Stormcarved Coast",
		effectstest.SpireGardenOracle:             "Spire Garden",
		effectstest.BuriedRuinOracle:              "Buried Ruin",
		effectstest.ImpactTremorsOracle:           "Impact Tremors",
		effectstest.PhyrexianTowerOracle:          "Phyrexian Tower",
		effectstest.StormKilnArtistOracle:         "Storm-Kiln Artist",
		effectstest.StrokeOfMidnightOracle:        "Stroke of Midnight",
		effectstest.RapidHybridizationOracle:      "Rapid Hybridization",
		effectstest.ProfessionalFaceBreakerOracle: "Professional Face-Breaker",
		effectstest.BountifulPromenadeOracle:      "Bountiful Promenade",
		effectstest.FyndhornElvesOracle:           "Fyndhorn Elves",
		effectstest.OrnithopterOfParadiseOracle:   "Ornithopter of Paradise",
		effectstest.KarplusanForestOracle:         "Karplusan Forest",
		effectstest.TerminateOracle:               "Terminate",
		effectstest.RockfallValeOracle:            "Rockfall Vale",
		effectstest.MirkwoodBatsOracle:            "Mirkwood Bats",
		effectstest.BrushlandOracle:               "Brushland",
		effectstest.DarkRitualOracle:              "Dark Ritual",
		effectstest.ReturnOfTheWildspeakerOracle:  "Return of the Wildspeaker",
		effectstest.BorosCharmOracle:              "Boros Charm",
	}
	if len(want) != 29 {
		t.Fatalf("the batch is 29 cards, the table lists %d", len(want))
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
	}
}

// The ten new land rows, each producing its own printed pair — the
// loop-leak canary the Temple cycle established.
func TestBatch01LandRowsProduceTheirPrintedColours(t *testing.T) {
	conditional := map[string]string{
		"Spectator Seating":   "{R|W}",
		"Vault of Champions":  "{W|B}",
		"Undergrowth Stadium": "{B|G}",
		"Spire Garden":        "{R|G}",
		"Bountiful Promenade": "{G|W}",
		"Dreamroot Cascade":   "{G|U}",
		"Stormcarved Coast":   "{U|R}",
		"Rockfall Vale":       "{R|G}",
	}
	pain := map[string]string{
		"Karplusan Forest": "{R|G}",
		"Brushland":        "{G|W}",
	}
	found := 0
	for _, spec := range effects.All() {
		if want, ok := conditional[spec.Name]; ok {
			found++
			if len(spec.Replacements) != 1 {
				t.Errorf("%s: %d replacements, want 1 (enters tapped unless …)", spec.Name, len(spec.Replacements))
			}
			if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != want {
				t.Errorf("%s: mana abilities %+v, want one producing %s", spec.Name, spec.ManaAbilities, want)
			}
		}
		if want, ok := pain[spec.Name]; ok {
			found++
			if len(spec.ManaAbilities) != 2 {
				t.Errorf("%s: %d mana abilities, want 2", spec.Name, len(spec.ManaAbilities))
				continue
			}
			if spec.ManaAbilities[0].Produced != "{C}" || spec.ManaAbilities[0].Rider != nil {
				t.Errorf("%s: ability 0 must be the painless {C}", spec.Name)
			}
			if spec.ManaAbilities[1].Produced != want || spec.ManaAbilities[1].PainToYou != 1 || spec.ManaAbilities[1].NarrowToCommanderIdentity {
				t.Errorf("%s: ability 1 must be %s with a damage rider and no identity narrowing", spec.Name, want)
			}
		}
	}
	if found != len(conditional)+len(pain) {
		t.Errorf("found %d of %d land rows", found, len(conditional)+len(pain))
	}
}

// --- the land cycles, one row each through the real entry path ---

func TestSpectatorSeatingEntersUntappedAtAFourPlayerTable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.PlayLandFromHand(t, g, "Spectator Seating", effectstest.SpectatorSeatingOracle)
	effectstest.Top100AssertEnteredUntapped(t, g, id, "Spectator Seating")
}

func TestBountifulPromenadeEntersTappedInADuel(t *testing.T) {
	g := effectstest.NewDuelCatalogGame(t)
	id := effectstest.PlayLandFromHand(t, g, "Bountiful Promenade", effectstest.BountifulPromenadeOracle)
	effectstest.Top100AssertEnteredTapped(t, g, id, "Bountiful Promenade")
}

func TestDreamrootCascadeEntersTappedWithOneOtherLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	id := effectstest.PlayLandFromHand(t, g, "Dreamroot Cascade", effectstest.DreamrootCascadeOracle)
	effectstest.Top100AssertEnteredTapped(t, g, id, "Dreamroot Cascade")
}

func TestRockfallValeEntersUntappedWithTwoOtherLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.SeedLandOnBattlefield(g, me.ID, "Mountain", "Basic Land — Mountain")
	id := effectstest.PlayLandFromHand(t, g, "Rockfall Vale", effectstest.RockfallValeOracle)
	effectstest.Top100AssertEnteredUntapped(t, g, id, "Rockfall Vale")
}

func TestKarplusanForestColoredHalfDealsOneDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := me.Life
	land := effectstest.SeedPermanentWithOracle(g, me.ID, "Karplusan Forest", "Land", effectstest.KarplusanForestOracle)

	if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.Life != before-1 {
		t.Errorf("life %d → %d, want %d", before, me.Life, before-1)
	}
	pick := effectstest.RiderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 2 {
		t.Fatalf("the colored half must ask R or G, got %+v", pick)
	}
}

func TestBrushlandColorlessHalfIsFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := me.Life
	land := effectstest.SeedPermanentWithOracle(g, me.ID, "Brushland", "Land", effectstest.BrushlandOracle)

	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.Life != before {
		t.Errorf("the {C} half hurt: %d → %d", before, me.Life)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
}

// --- mana dorks ----------------------------------------------------

func TestFyndhornElvesTapsForGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	elf := effectstest.PushCatalogPermanent(g, me.ID, "Fyndhorn Elves", "Creature — Elf Druid", effectstest.FyndhornElvesOracle, false)
	if err := g.ActivateManaAbility(me.ID, elf, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool %v, want [G]", got)
	}
}

func TestOrnithopterOfParadiseFliesAndFixes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	thopter := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ornithopter of Paradise",
		TypeLine: "Artifact Creature — Thopter", OracleID: effectstest.OrnithopterOfParadiseOracle,
		Power: 0, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, thopter), "flying") {
		t.Error("printed flying did not reach the effective abilities")
	}
	if err := g.ActivateManaAbility(me.ID, thopter, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := effectstest.RiderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("with no commander the pick must offer all five colours, got %+v", pick)
	}
}

// --- removal -------------------------------------------------------

func TestTerminateDestroysACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.SeedCreature(g, "Bear", opp.ID)
	effectstest.CastCatalogSpell(t, g, "Terminate", "Instant", effectstest.TerminateOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Graveyard.Contains(bear) {
		t.Error("the creature should be in its owner's graveyard")
	}
}

func TestRapidHybridizationGivesTheVictimAFrogLizard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.SeedCreature(g, "Bear", opp.ID)
	effectstest.CastCatalogSpell(t, g, "Rapid Hybridization", "Instant", effectstest.RapidHybridizationOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature survived")
	}
	if effectstest.CountBattlefieldNamed(g, opp.ID, "Frog Lizard") != 1 {
		t.Error("the VICTIM should get the 3/3 Frog Lizard")
	}
	if effectstest.CountBattlefieldNamed(g, me.ID, "Frog Lizard") != 0 {
		t.Error("the caster must not get the token")
	}
}

func TestStrokeOfMidnightDestroysANonlandAndHandsBackAHuman(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	rock := effectstest.SeedPermanentFor(g, opp.ID, "Rock", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Stroke of Midnight", "Instant", effectstest.StrokeOfMidnightOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
	if effectstest.CountBattlefieldNamed(g, opp.ID, "Human") != 1 {
		t.Error("its controller should get a 1/1 Human")
	}
}

func TestStrokeOfMidnightRefusesALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.SeedLandOnBattlefield(g, opp.ID, "Forest", "Basic Land — Forest")
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Stroke of Midnight", TypeLine: "Instant",
		OracleID: effectstest.StrokeOfMidnightOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}},
	}); err == nil {
		t.Error("a land was accepted as a target for 'target nonland permanent'")
	}
}

// --- go-wide payoffs -----------------------------------------------

func TestImpactTremorsPingsOncePerCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Impact Tremors", "Enchantment", effectstest.ImpactTremorsOracle, false)
	before := effectstest.LifeOfOpponents(g)

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.RedGoblinToken(), 2) })
	effectstest.PassPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-2 {
			t.Errorf("opponent %d: %d -> %d, want -2 (two creatures, two triggers)", i+1, b, got)
		}
	}

	// A noncreature entering, or an opponent's creature, is silent.
	mid := effectstest.LifeOfOpponents(g)
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, effects.TreasureToken(), 1)
		_ = g.CreateTokenForEffect(opp.ID, effects.RedGoblinToken(), 1)
	})
	effectstest.PassPriorityAroundTable(t, g)
	for i, b := range mid {
		if got := g.Seats[i+1].Life; got != b {
			t.Errorf("opponent %d: %d -> %d, want unchanged", i+1, b, got)
		}
	}
}

func TestMirkwoodBatsDrainsOnTokenCreatedAndTokenSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Mirkwood Bats", "Creature — Bat", effectstest.MirkwoodBatsOracle, false)
	before := effectstest.LifeOfOpponents(g)

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TreasureToken(), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("opponent %d after a token was created: %d -> %d, want -1", i+1, b, got)
		}
	}

	treasure := effectstest.FindBattlefieldByName(g, "Treasure")
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(treasure) })
	effectstest.PassPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-2 {
			t.Errorf("opponent %d after the token was sacrificed: %d -> %d, want -2", i+1, b, got)
		}
	}

	// A nontoken sacrifice and an opponent's token are both silent.
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	g.WithWriteLock(func() {
		_ = g.SacrificePermanentForEffect(bear)
		_ = g.CreateTokenForEffect(opp.ID, effects.TreasureToken(), 1)
	})
	effectstest.PassPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-2 {
			t.Errorf("opponent %d: %d -> %d, want still -2", i+1, b, got)
		}
	}
}

func TestTirelessProvisionerMakesATreasureOnLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Tireless Provisioner", "Creature — Elf Scout", effectstest.TirelessProvisionerOracle, false)

	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	// The token is chosen as the trigger resolves: Treasure first, then Food.
	c := effectstest.LatestOptionPickFor(g, me.ID)
	if c == nil || len(c.PickOptions) != 2 || c.PickOptions[0].Label != "Create a Treasure token" {
		t.Fatalf("a Food-or-Treasure choice, Treasure first: %+v", c)
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	if effectstest.CountBattlefieldNamed(g, me.ID, "Treasure") != 1 {
		t.Fatal("a land drop should make one Treasure")
	}

	// An opponent's land is not "a land YOU control".
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(opp.ID, game.Card{Name: "Forest", TypeLine: "Token Land — Forest"}, 1)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldNamed(g, me.ID, "Treasure") != 1 || effectstest.LatestOptionPickFor(g, me.ID) != nil {
		t.Error("an opponent's land triggered my landfall")
	}
}

// The other answer is a Food token.
func TestTirelessProvisionerCanMakeAFoodOnLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Tireless Provisioner", "Creature — Elf Scout", effectstest.TirelessProvisionerOracle, false)
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	if effectstest.CountBattlefieldNamed(g, me.ID, "Food") != 1 || effectstest.CountBattlefieldNamed(g, me.ID, "Treasure") != 0 {
		t.Error("choosing Food makes a Food and no Treasure")
	}
	if spec, _ := effects.Lookup(effectstest.TirelessProvisionerOracle); spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Error("both tokens are offered: nothing is left to declare")
	}
}

// --- Phyrexian Tower -----------------------------------------------

func TestPhyrexianTowerSacrificesACreatureForTwoBlack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tower := effectstest.SeedPermanentWithOracle(g, me.ID, "Phyrexian Tower", "Legendary Land", effectstest.PhyrexianTowerOracle)
	bear := effectstest.SeedCreature(g, "Bear", me.ID)

	if err := g.ActivateManaAbility(me.ID, tower, 1, game.ManaAbilityParams{
		SacrificeIDs: []uuid.UUID{bear},
	}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 2 || got[0] != "B" || got[1] != "B" {
		t.Errorf("pool %v, want [B B]", got)
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the creature was not sacrificed")
	}
	card, _ := effectstest.BattlefieldCard(g, tower)
	if !card.Tapped {
		t.Error("the Tower has a tap cost and must be tapped")
	}
}

func TestPhyrexianTowerBlackHalfNeedsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tower := effectstest.SeedPermanentWithOracle(g, me.ID, "Phyrexian Tower", "Legendary Land", effectstest.PhyrexianTowerOracle)
	if err := g.ActivateManaAbility(me.ID, tower, 1, game.ManaAbilityParams{}); err == nil {
		t.Fatal("activated the sacrifice half with nothing to sacrifice")
	}
	card, _ := effectstest.BattlefieldCard(g, tower)
	if card.Tapped {
		t.Error("a refused activation must leave the Tower untapped")
	}
	if err := g.ActivateManaAbility(me.ID, tower, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("the colorless half: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
}

// --- Buried Ruin ---------------------------------------------------

func TestBuriedRuinReturnsAnArtifactCardToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ruin := effectstest.PushCatalogPermanent(g, me.ID, "Buried Ruin", "Land", effectstest.BuriedRuinOracle, false)
	relic := effectstest.Batch01GraveyardCard(me, "Sol Ring", "Artifact")
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})

	if err := g.ActivateCatalogAbility(me.ID, ruin, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: relic}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(ruin) {
		t.Error("the Ruin is sacrificed as a cost, at announce")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(relic) {
		t.Error("the artifact card did not come back to hand")
	}
}

func TestBuriedRuinRefusesANonArtifactCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ruin := effectstest.PushCatalogPermanent(g, me.ID, "Buried Ruin", "Land", effectstest.BuriedRuinOracle, false)
	bear := effectstest.PushGraveyardCardForTest(me, "Bear")
	if err := g.ActivateCatalogAbility(me.ID, ruin, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Error("a creature card was accepted for 'target artifact card'")
	}
	if !g.Battlefield.Contains(ruin) {
		t.Error("a refused activation must not sacrifice the Ruin")
	}
}

// --- Farewell ------------------------------------------------------

func TestFarewellExilesOnlyTheChosenTypesAndAllGraveyards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myBear := effectstest.SeedCreature(g, "My Bear", me.ID)
	theirBear := effectstest.SeedCreature(g, "Their Bear", opp.ID)
	rock := effectstest.SeedPermanentFor(g, opp.ID, "Rock", "Artifact")
	aura := effectstest.SeedPermanentFor(g, opp.ID, "Glory", "Enchantment")
	myDead := effectstest.PushGraveyardCardForTest(me, "My Dead")
	theirDead := effectstest.PushGraveyardCardForTest(opp, "Their Dead")

	effectstest.CastModal(t, g, "Farewell", "Sorcery", effectstest.FarewellOracle, []int{1, 3}, nil)
	effectstest.PassPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{myBear, theirBear, myDead, theirDead} {
		if !g.Exile.Contains(id) {
			t.Errorf("%s should be in exile", id)
		}
	}
	for _, id := range []uuid.UUID{rock, aura} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s was exiled but its type was not chosen", id)
		}
	}
	if me.Graveyard.Size() != 1 || opp.Graveyard.Size() != 0 {
		t.Errorf("graveyards should hold only Farewell itself: me=%d opp=%d", me.Graveyard.Size(), opp.Graveyard.Size())
	}
}

// --- Storm-Kiln Artist ---------------------------------------------

func TestStormKilnArtistMakesATreasureOnAnInstantAndGrows(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	artist := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Storm-Kiln Artist", TypeLine: "Creature — Dwarf Shaman",
		OracleID: effectstest.StormKilnArtistOracle, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	if p := effectstest.EffectivePower(t, g, artist); p != 2 {
		t.Fatalf("with no artifacts power = %d, want 2", p)
	}

	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.CountBattlefieldNamed(g, me.ID, "Treasure") != 1 {
		t.Fatal("casting an instant should make one Treasure")
	}
	if p := effectstest.EffectivePower(t, g, artist); p != 3 {
		t.Errorf("with one artifact power = %d, want 3", p)
	}

	// A creature spell is not an instant or sorcery.
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldNamed(g, me.ID, "Treasure") != 1 {
		t.Error("a creature spell triggered magecraft")
	}
}

// --- Arcane Denial -------------------------------------------------

func TestArcaneDenialCountersAndBothPlayersDrawAtTheNextUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})

	effectstest.CastCatalogSpell(t, g, "Arcane Denial", "Instant", effectstest.ArcaneDenialOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	effectstest.PassPriorityAroundTable(t, g)

	if !opp.Graveyard.Contains(bolt) {
		t.Fatal("the Bolt was not countered")
	}
	if n := len(g.DelayedTriggers); n != 1 {
		t.Fatalf("delayed triggers = %d, want 1 (both draws on one item — see the card comment)", n)
	}

	// Nothing is drawn before the next upkeep.
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	oppBefore, meBefore := opp.Hand.Size(), me.Hand.Size()
	if len(g.PendingTriggers)+len(g.StackMeta) == 0 {
		t.Fatal("the draws should be on the stack at the next upkeep")
	}
	effectstest.PassPriorityAroundTable(t, g)

	// "May draw up to two": a number prompt to the victim; the caster's
	// own card follows the answer.
	pa := effectstest.PayAmountPrompt(g, opp.ID)
	if pa == nil || pa.PayAmount.Min != 0 || pa.PayAmount.Max != 2 {
		t.Fatalf("the victim is asked for 0..2: %+v", pa)
	}
	if me.Hand.Size() != meBefore {
		t.Error("the caster draws after the victim's answer, not before")
	}
	if err := g.ResolvePayAmount(pa.ID, opp.ID, 2); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := opp.Hand.Size(); got != oppBefore+2 {
		t.Errorf("the countered spell's controller drew %d, want 2", got-oppBefore)
	}
	if got := me.Hand.Size(); got != meBefore+1 {
		t.Errorf("Arcane Denial's caster drew %d, want 1", got-meBefore)
	}
	if n := len(g.DelayedTriggers); n != 0 {
		t.Errorf("delayed triggers left queued: %d", n)
	}
}

// The victim may draw FEWER (the old caveat): zero and one are answers,
// and the caster still draws their card.
func TestArcaneDenialVictimMayDrawFewer(t *testing.T) {
	for _, n := range []int{0, 1} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		bolt := effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "",
			[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
		effectstest.CastCatalogSpell(t, g, "Arcane Denial", "Instant", effectstest.ArcaneDenialOracle,
			[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AdvanceToUpkeepOf(t, g, 1)
		oppBefore, meBefore := opp.Hand.Size(), me.Hand.Size()
		effectstest.PassPriorityAroundTable(t, g)
		pa := effectstest.PayAmountPrompt(g, opp.ID)
		if pa == nil {
			t.Fatal("no number prompt")
		}
		if err := g.ResolvePayAmount(pa.ID, opp.ID, n); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := opp.Hand.Size() - oppBefore; got != n {
			t.Errorf("chose %d, drew %d", n, got)
		}
		if got := me.Hand.Size() - meBefore; got != 1 {
			t.Errorf("chose %d: the caster drew %d, want 1", n, got)
		}
	}
	if spec, _ := effects.Lookup(effectstest.ArcaneDenialOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("the up-to-two choice landed")
	}
}

// --- Mana Drain ----------------------------------------------------

// TestManaDrainRefundsOnYourOwnPostcombatMainWhenCastDuringYourPrecombatMain
// pins the fix for the closed caveat (#1565): a Drain cast during your
// own precombat main pays out THIS turn's postcombat main, not the
// next turn's precombat main — the "your next main phase" reading a
// delayed trigger naming one step could not express before
// manaDrainNextMainPhaseStep picked between the two at resolution.
func TestManaDrainRefundsOnYourOwnPostcombatMainWhenCastDuringYourPrecombatMain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell := effectstest.Batch01OpponentCasts(t, g, opp, "Big Bolt", effectstest.LightningBoltOracle, "{2}{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})

	// Cast during MY OWN precombat main (castCatalogSpell always casts
	// as the active seat, and it's seat 0's turn 1).
	effectstest.CastCatalogSpell(t, g, "Mana Drain", "Instant", effectstest.ManaDrainOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	effectstest.PassPriorityAroundTable(t, g)

	if !opp.Graveyard.Contains(spell) {
		t.Fatal("the spell was not countered")
	}
	if n := len(g.DelayedTriggers); n != 1 || !g.DelayedTriggers[0].ControllerTurnOnly {
		t.Fatalf("want one delayed trigger gated to the controller's turn, got %d", n)
	}
	if got := g.DelayedTriggers[0].At; got != game.StepPostcombatMain {
		t.Fatalf("At = %s, want postcombat main (this turn's, not next turn's precombat)", got)
	}

	// This turn's postcombat main, same active seat: the refund fires.
	effectstest.Batch01AdvanceToStepOf(t, g, 0, game.StepPostcombatMain)
	if len(g.DelayedTriggers) != 0 {
		t.Fatal("the refund did not fire on the caster's own postcombat main this turn")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Batch01PoolColors(me); len(got) != 3 || got[0] != "C" || got[1] != "C" || got[2] != "C" {
		t.Errorf("pool %v, want [C C C] — the countered spell's mana value", got)
	}
}

// TestManaDrainRefundsOnYourOwnNextPrecombatMainWhenCastOnAnOpponentsTurn
// is the case that was already correct: countering a spell on someone
// else's turn waits for the caster's own NEXT precombat main, however
// many turns away — ControllerTurnOnly's whole reason to exist.
func TestManaDrainRefundsOnYourOwnNextPrecombatMainWhenCastOnAnOpponentsTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.Batch01AdvanceToStepOf(t, g, 1, game.StepPrecombatMain)
	spell := effectstest.Batch01OpponentCasts(t, g, opp, "Big Bolt", effectstest.LightningBoltOracle, "{2}{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})

	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Mana Drain", TypeLine: "Instant",
		OracleID: effectstest.ManaDrainOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: spell}},
	}); err != nil {
		t.Fatalf("CastSpell Mana Drain: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if !opp.Graveyard.Contains(spell) {
		t.Fatal("the spell was not countered")
	}
	if n := len(g.DelayedTriggers); n != 1 || !g.DelayedTriggers[0].ControllerTurnOnly {
		t.Fatalf("want one delayed trigger gated to the controller's turn, got %d", n)
	}
	if got := g.DelayedTriggers[0].At; got != game.StepPrecombatMain {
		t.Fatalf("At = %s, want precombat main (the caster's own next one)", got)
	}

	// The opponent's postcombat main this same turn is NOT "my next
	// main phase".
	effectstest.Batch01AdvanceToStepOf(t, g, 1, game.StepPostcombatMain)
	if len(g.DelayedTriggers) != 1 {
		t.Fatal("the refund fired on an opponent's main phase")
	}

	// My own next precombat main is.
	effectstest.Batch01AdvanceToStepOf(t, g, 0, game.StepPrecombatMain)
	if len(g.DelayedTriggers) != 0 {
		t.Fatal("the refund did not fire on the caster's own next precombat main")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Batch01PoolColors(me); len(got) != 3 || got[0] != "C" || got[1] != "C" || got[2] != "C" {
		t.Errorf("pool %v, want [C C C] — the countered spell's mana value", got)
	}
}

func TestManaDrainOnAZeroValueSpellSchedulesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell := effectstest.Batch01OpponentCasts(t, g, opp, "Free Bolt", effectstest.LightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	effectstest.CastCatalogSpell(t, g, "Mana Drain", "Instant", effectstest.ManaDrainOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(spell) {
		t.Fatal("the spell was not countered")
	}
	if n := len(g.DelayedTriggers); n != 0 {
		t.Errorf("a mana value of 0 scheduled %d refunds", n)
	}
}

// --- Dark Ritual ---------------------------------------------------

func TestDarkRitualAddsThreeBlackMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Dark Ritual", "Instant", effectstest.DarkRitualOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Batch01PoolColors(me); len(got) != 3 || got[0] != "B" || got[1] != "B" || got[2] != "B" {
		t.Errorf("pool %v, want [B B B]", got)
	}
	if me.Graveyard.Size() != 1 {
		t.Errorf("Dark Ritual should be in the graveyard after resolving")
	}
}

// --- Victimize -----------------------------------------------------

// #2863: the sacrifice is part of the effect, not a cost. The creature
// is still on the battlefield after the cast, the caster picks it as
// the spell resolves, and the targets come back only once it is gone.
func TestVictimizeSacrificesOnResolutionAndReturnsBothTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.PushGraveyardCardForTest(me, "Dead A")
	b := effectstest.PushGraveyardCardForTest(me, "Dead B")
	fodder := effectstest.SeedCreature(g, "Fodder", me.ID)
	effectstest.CastCatalogSpell(t, g, "Victimize", "Sorcery", effectstest.VictimizeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	if !g.Battlefield.Contains(fodder) {
		t.Fatal("nothing is sacrificed at cast")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SacrificeChoiceFor(g, me.ID) == nil {
		t.Fatal("the caster is asked to sacrifice a creature as the spell resolves")
	}
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Fatal("the targets came back before the sacrifice")
	}
	effectstest.AnswerSacrifice(t, g, me.ID, fodder)
	if !me.Graveyard.Contains(fodder) {
		t.Error("the chosen creature is sacrificed")
	}
	for _, want := range []uuid.UUID{a, b} {
		card, ok := effectstest.BattlefieldCard(g, want)
		if !ok {
			t.Errorf("%s did not return to the battlefield", want)
			continue
		}
		if !card.Tapped {
			t.Errorf("%s returned untapped", want)
		}
	}
}

// With no creature to sacrifice the spell is still castable (#2863);
// "if you do" is false, so nothing comes back.
func TestVictimizeWithNoCreatureResolvesAndReturnsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.PushGraveyardCardForTest(me, "Dead A")
	b := effectstest.PushGraveyardCardForTest(me, "Dead B")
	spell := effectstest.CastCatalogSpell(t, g, "Victimize", "Sorcery", effectstest.VictimizeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SacrificeChoiceFor(g, me.ID) != nil {
		t.Fatal("a caster with no creature is not asked (CR 701.21a)")
	}
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("no creature was sacrificed, so nothing returns")
	}
	if !me.Graveyard.Contains(spell) || !me.Graveyard.Contains(a) || !me.Graveyard.Contains(b) {
		t.Error("the spell resolves to the graveyard and the targets stay there")
	}
}

func TestVictimizeNeedsExactlyTwoTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.PushGraveyardCardForTest(me, "Dead A")
	if err := effectstest.CastCatalogSpellErr(t, g, "Victimize", "Sorcery", effectstest.VictimizeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}}); err == nil {
		t.Error("one target was accepted for 'two target creature cards'")
	}
}

// --- Professional Face-Breaker -------------------------------------

func TestProfessionalFaceBreakerMakesOneTreasurePerCombatDamageStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	breaker := effectstest.PushCatalogPermanent(g, me.ID, "Professional Face-Breaker", "Creature — Human Warrior", effectstest.ProfessionalFaceBreakerOracle, false)
	a := effectstest.PushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "Bear B", 2, 2)

	effectstest.AttackWith(t, g, victim.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Treasure"); n != 1 {
		t.Fatalf("two creatures connecting made %d Treasures, want 1 — 'one or more' is one trigger", n)
	}
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, breaker), "menace") {
		t.Error("printed menace did not reach the effective abilities")
	}

	// Crack the Treasure for an impulse card.
	top := effectstest.SeedLibrary(me, "Top Card")[0]
	treasure := effectstest.FindBattlefieldByName(g, "Treasure")
	if err := g.ActivateCatalogAbility(me.ID, breaker, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{treasure},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(top) {
		t.Fatal("the top card was not exiled")
	}
	perm := effectstest.ExiledPermission(g, top)
	if perm.Player != me.ID {
		t.Errorf("the play permission belongs to %s, want the controller", perm.Player)
	}
	if perm.CastOnly {
		t.Error("the printed text says PLAY — a land must not be stranded")
	}
}

func TestProfessionalFaceBreakerNeedsATreasureToCrack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	breaker := effectstest.PushCatalogPermanent(g, me.ID, "Professional Face-Breaker", "Creature — Human Warrior", effectstest.ProfessionalFaceBreakerOracle, false)
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	if err := g.ActivateCatalogAbility(me.ID, breaker, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{bear},
	}); err == nil {
		t.Error("a creature was accepted for 'Sacrifice a Treasure'")
	}
}

// --- Return of the Wildspeaker -------------------------------------

func batch01WildspeakerBoard(g *game.Game, me uuid.UUID) (bear, human uuid.UUID) {
	bear = effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
		Power: 4, Toughness: 4, Owner: me, Controller: me,
	})
	human = effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Knight", TypeLine: "Creature — Human Knight",
		Power: 6, Toughness: 6, Owner: me, Controller: me,
	})
	return bear, human
}

func TestReturnOfTheWildspeakerDrawsTheGreatestNonHumanPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	batch01WildspeakerBoard(g, me.ID)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Titan", TypeLine: "Creature — Giant",
		Power: 9, Toughness: 9, Owner: g.Seats[1].ID, Controller: g.Seats[1].ID,
	})
	before := me.Hand.Size()

	effectstest.CastModal(t, g, "Return of the Wildspeaker", "Instant", effectstest.ReturnOfTheWildspeakerOracle, []int{0}, nil)
	effectstest.PassPriorityAroundTable(t, g)

	// The 6/6 Human and the opponent's 9/9 do not count; the 4/4 Bear
	// does. castModal adds the spell to the hand and casting removes
	// it, so the difference is exactly the cards drawn.
	if got := me.Hand.Size() - before; got != 4 {
		t.Errorf("drew %d, want 4", got)
	}
}

func TestReturnOfTheWildspeakerPumpsNonHumansUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear, human := batch01WildspeakerBoard(g, me.ID)

	effectstest.CastModal(t, g, "Return of the Wildspeaker", "Instant", effectstest.ReturnOfTheWildspeakerOracle, []int{1}, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 7 || tough != 7 {
		t.Errorf("Bear = %d/%d, want 7/7", p, tough)
	}
	if p := effectstest.EffectivePower(t, g, human); p != 6 {
		t.Errorf("Human = %d power, want 6 (not pumped)", p)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 4 {
		t.Errorf("Bear after cleanup = %d power, want 4", p)
	}
}

// --- Boros Charm ---------------------------------------------------

func TestBorosCharmDealsFourToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	before := opp.Life
	effectstest.CastModal(t, g, "Boros Charm", "Instant", effectstest.BorosCharmOracle, []int{0},
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-4 {
		t.Errorf("life %d -> %d, want -4", before, opp.Life)
	}
}

func TestBorosCharmDamageModeRefusesACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.SeedCreature(g, "Bear", opp.ID)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Boros Charm", TypeLine: "Instant",
		OracleID: effectstest.BorosCharmOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Modes:   []int{0},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Error("a creature was accepted for 'target player or planeswalker'")
	}
}

func TestBorosCharmGrantsDoubleStrikeUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	effectstest.CastModal(t, g, "Boros Charm", "Instant", effectstest.BorosCharmOracle, []int{2},
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, bear), "double strike") {
		t.Fatal("double strike was not granted")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, bear), "double strike") {
		t.Error("double strike survived the cleanup step")
	}
}

// --- engine fixes this batch found ---------------------------------

// A token entering the battlefield used to leave the layer cache
// stale: CreateTokenForEffect emits no EventZoneMove, and that was
// the only entry event the layer listener invalidated on. The
// user-visible symptom is the oldest static in the catalog — a
// Goblin token made under Glorious Anthem stayed 1/1.
func TestTokenEnteringUnderAnAnthemIsPumpedImmediately(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Glorious Anthem", TypeLine: "Enchantment",
		OracleID: effectstest.GloriousAnthemOracle, Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.RedGoblinToken(), 1) })
	goblin := effectstest.FindBattlefieldByName(g, "Goblin")
	if goblin == uuid.Nil {
		t.Fatal("no Goblin token")
	}
	if p, tough := effectstest.EffectivePower(t, g, goblin), effectstest.EffectiveToughness(t, g, goblin); p != 2 || tough != 2 {
		t.Errorf("Goblin under an Anthem = %d/%d, want 2/2", p, tough)
	}
	card, _ := effectstest.BattlefieldCard(g, goblin)
	if card.EnteredBattlefieldAt == 0 {
		t.Error("a token must get a battlefield-entry timestamp (CR 613 ordering)")
	}
}

// Every card the slice registered, pinned by oracle ID → name, the
// way TestBatch01CardsAreRegistered pins the first pass's.
func TestBatch01SeptemberSliceCardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.OptOracle:                    "Opt",
		effectstest.BrainstormOracle:             "Brainstorm",
		effectstest.GrowthSpiralOracle:           "Growth Spiral",
		effectstest.ExploreOracle:                "Explore",
		effectstest.CropRotationOracle:           "Crop Rotation",
		effectstest.AshBarrensOracle:             "Ash Barrens",
		effectstest.FlawlessManeuverOracle:       "Flawless Maneuver",
		effectstest.AkromasWillOracle:            "Akroma's Will",
		effectstest.JeskasWillOracle:             "Jeska's Will",
		effectstest.MysticSanctuaryOracle:        "Mystic Sanctuary",
		effectstest.SecludedCourtyardOracle:      "Secluded Courtyard",
		effectstest.NykthosOracle:                "Nykthos, Shrine to Nyx",
		effectstest.ThreeTreeCityOracle:          "Three Tree City",
		effectstest.PatchworkBannerOracle:        "Patchwork Banner",
		effectstest.YavimayaOracle:               "Yavimaya, Cradle of Growth",
		effectstest.ScuteSwarmOracle:             "Scute Swarm",
		effectstest.BlackMarketConnectionsOracle: "Black Market Connections",
		effectstest.HeraldsHornOracle:            "Herald's Horn",
		effectstest.RoamingThroneOracle:          "Roaming Throne",
		effectstest.TheGreatHengeOracle:          "The Great Henge",
		effectstest.BoseijuWhoEnduresOracle:      "Boseiju, Who Endures",
		effectstest.OtawaraSoaringCityOracle:     "Otawara, Soaring City",
		effectstest.CommandBeaconOracle:          "Command Beacon",
		effectstest.GemstoneCavernsOracle:        "Gemstone Caverns",
	}
	if len(want) != 24 {
		t.Fatalf("the slice is 24 cards, the table lists %d", len(want))
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
	}
}

// --- the cantrips --------------------------------------------------

// Opt scries before it draws, which is the whole card: the draw is in
// Scry.Then, so the hand does not grow until the scry is answered.
func TestOptScriesThenDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Opt", "Instant", effectstest.OptOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	scry := effectstest.LatestChoiceOfKind(g, game.PendingChoiceScry)
	if scry == nil {
		t.Fatal("Opt queued no scry prompt")
	}
	if got := me.Hand.Size(); got != before {
		t.Errorf("hand %d before the scry is answered, want %d — the draw must wait", got, before)
	}
	if err := g.ResolveScry(scry.ID, me.ID, nil, scry.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand after Opt = %d, want %d", got, before+1)
	}
}

// Brainstorm draws three and puts two back in the order the player
// answers (#996: a put_in_library prompt after the pick).
func TestBrainstormDrawsThreeAndPutsTwoBackFirstPickOnTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Brainstorm", "Instant", effectstest.BrainstormOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != before+3 {
		t.Fatalf("hand after the draw = %d, want %d", got, before+3)
	}
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("Brainstorm queued no put-back prompt")
	}
	if pick.ChooseMin != 2 || pick.ChooseMax != 2 {
		t.Errorf("put-back bounds = %d..%d, want exactly two", pick.ChooseMin, pick.ChooseMax)
	}
	first, second := me.Hand.Cards[0].InstanceID, me.Hand.Cards[1].InstanceID
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{first, second}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	order := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if order == nil {
		t.Fatal("no order prompt after the pick")
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, nil, []uuid.UUID{first, second}); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand after putting two back = %d, want %d", got, before+1)
	}
	top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID
	if top != first {
		t.Error("the first card of the answer must end on top of the library")
	}
}

// Growth Spiral draws, then offers the land — and the land is PUT,
// so it does not spend the turn's land drop.
func TestGrowthSpiralDrawsThenPutsALandFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest",
		Owner: me.ID, Controller: me.ID,
	})
	effectstest.CastCatalogSpell(t, g, "Growth Spiral", "Instant", effectstest.GrowthSpiralOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("Growth Spiral queued no land prompt")
	}
	if pick.ChooseMin != 0 {
		t.Errorf("the prompt floor is %d, want 0 — \"you MAY put\"", pick.ChooseMin)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{land}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !g.Battlefield.Contains(land) {
		t.Error("the picked land did not reach the battlefield")
	}
}

// Explore banks a land play for the turn and draws.
func TestExploreGrantsALandPlayAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Explore", "Sorcery", effectstest.ExploreOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if got := g.ExtraLandDropsThisTurn[me.ID]; got != 1 {
		t.Errorf("extra land plays = %d, want 1", got)
	}
	// Explore was seeded into the hand, cast out of it, and drew one:
	// net one card up on the count taken before the seeding.
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand after Explore = %d, want %d", got, before+1)
	}
}

// --- the search-and-sacrifice spells -------------------------------

// Crop Rotation eats a land at announce and fetches any land.
func TestCropRotationSacrificesALandAndFetchesOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	victim := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	want := effectstest.PushLibraryCardForTest(me, game.Card{
		Name: "Gaea's Cradle", TypeLine: "Legendary Land",
	})

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	spell := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: spell, Name: "Crop Rotation", TypeLine: "Instant",
		OracleID: effectstest.CropRotationOracle, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{SacrificeIDs: []uuid.UUID{victim}}); err != nil {
		t.Fatalf("CastSpell Crop Rotation: %v", err)
	}
	if g.Battlefield.Contains(victim) {
		t.Error("the additional cost must sacrifice the land at announce (CR 601.2f)")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	if !g.Battlefield.Contains(want) {
		t.Error("the fetched land did not reach the battlefield")
	}
}

// Ash Barrens' basic landcycling is an ability that functions from
// the hand, with a discard-this cost.
func TestAshBarrensBasicLandcyclesFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SeedSearchLibrary(me, effectstest.SearchTestLand("Island", "Basic Land — Island"))
	basic := ids[0]

	id, _ := effectstest.CycleFromHand(t, g, "Ash Barrens", "Land", effectstest.AshBarrensOracle, "{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the discard-this cost must put Ash Barrens in the graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)

	// Basic landcycling searches for a basic land card, a stated
	// quality, so even one basic in the library asks: the cycler may
	// find nothing (CR 701.23b, #3053).
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("landcycling against one basic did not ask")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{basic}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !me.Hand.Contains(basic) {
		t.Error("basic landcycling did not put the basic into hand")
	}
}

// --- the free spell and the two modal instants ---------------------

// Flawless Maneuver is free while you control a commander, and it
// only touches creatures.
func TestFlawlessManeuverGrantsIndestructibleToCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sol Ring", TypeLine: "Artifact",
		Owner: me.ID, Controller: me.ID,
	})
	effectstest.CastCatalogSpell(t, g, "Flawless Maneuver", "Instant", effectstest.FlawlessManeuverOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, bear), "indestructible") {
		t.Error("the creature did not gain indestructible")
	}
	if effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, rock), "indestructible") {
		t.Error("\"creatures you control\" must not reach an artifact")
	}
	if cost := game.AlternativeCostByKey(effectstest.FlawlessManeuverOracle, "free"); cost == nil {
		t.Error("the commander free-cast offer is not declared")
	}
}

// Akroma's Will's second bullet grants protection from each colour —
// five separate qualities, not one.
func TestAkromasWillSecondBulletGrantsProtectionFromEachColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	effectstest.CastCatalogSpellWithModes(t, g, "Akroma's Will", "Instant", effectstest.AkromasWillOracle, []int{1})
	effectstest.PassPriorityAroundTable(t, g)

	abilities := effectstest.EffectiveAbilities(t, g, bear)
	for _, want := range []string{"lifelink", "indestructible"} {
		if !effectstest.HasAbility(abilities, want) {
			t.Errorf("the creature did not gain %s", want)
		}
	}
	for _, c := range game.AllColors {
		if !effectstest.HasAbility(abilities, game.ProtectionFromColor(c)) {
			t.Errorf("the creature did not gain %s", game.ProtectionFromColor(c))
		}
	}
}

// Akroma's Will asks for ONE bullet without a commander on the
// battlefield: the "choose both" rider's condition is false, so a
// two-mode announcement is refused rather than quietly allowed
// (#1590; the commander half is in conditional_mode_count_test.go).
func TestAkromasWillRefusesBothModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Akroma's Will", TypeLine: "Instant",
		OracleID: effectstest.AkromasWillOracle, Owner: me.ID, Controller: me.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0, 1}}); err == nil {
		t.Error("choosing both bullets must be refused without a commander")
	}
}

// Jeska's Will's ritual counts the target opponent's hand as it
// resolves.
func TestJeskasWillRitualCountsTheOpponentsHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[1]
	for i := 0; i < 2; i++ {
		opp.Hand.PushTop(game.Card{
			InstanceID: uuid.New(), Name: "Filler", TypeLine: "Instant",
			Owner: opp.ID, Controller: opp.ID,
		})
	}
	want := opp.Hand.Size()
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Jeska's Will", TypeLine: "Sorcery",
		OracleID: effectstest.JeskasWillOracle, Owner: me.ID, Controller: me.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Modes:   []int{0},
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("CastSpell Jeska's Will: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	red := 0
	for _, tok := range me.ManaPool.Units() {
		if tok.Color == "R" {
			red++
		}
	}
	if red != want {
		t.Errorf("red mana added = %d, want %d (one per card in the opponent's hand)", red, want)
	}
}

// The second bullet exiles three and makes them playable this turn.
func TestJeskasWillImpulseExilesThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	before := g.Exile.Size()
	effectstest.CastCatalogSpellWithModes(t, g, "Jeska's Will", "Sorcery", effectstest.JeskasWillOracle, []int{1})
	effectstest.PassPriorityAroundTable(t, g)
	if got := g.Exile.Size() - before; got != 3 {
		t.Errorf("exiled %d cards, want 3", got)
	}
}

// --- the lands -----------------------------------------------------

// Mystic Sanctuary enters untapped on the fourth Island and its
// trigger puts an instant back on top.
func TestMysticSanctuaryEntersUntappedAndRebuysAnInstant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 3; i++ {
		effectstest.SeedLandOnBattlefield(g, me.ID, "Island", "Basic Land — Island")
	}
	bolt := effectstest.Batch01GraveyardCard(me, "Counterspell", "Instant")

	id := effectstest.PlayLandFromHand(t, g, "Mystic Sanctuary", effectstest.MysticSanctuaryOracle)
	card, ok := effectstest.BattlefieldCard(g, id)
	if !ok || card.Tapped {
		t.Fatal("Mystic Sanctuary must enter untapped with three other Islands")
	}
	if prompt := effectstest.LatestPickTarget(g, me.ID); prompt == nil {
		t.Fatal("no pick_target prompt after the \"you may\"")
	}
	effectstest.PickCard(t, g, me.ID, bolt)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(bolt) {
		t.Error("the instant is still in the graveyard")
	}
	if top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID; top != bolt {
		t.Error("the instant did not go on top of the library")
	}
}

// With only two other Islands it enters tapped and asks nothing.
func TestMysticSanctuaryEntersTappedBelowThreeIslands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 2; i++ {
		effectstest.SeedLandOnBattlefield(g, me.ID, "Island", "Basic Land — Island")
	}
	id := effectstest.PlayLandFromHand(t, g, "Mystic Sanctuary", effectstest.MysticSanctuaryOracle)
	card, ok := effectstest.BattlefieldCard(g, id)
	if !ok || !card.Tapped {
		t.Error("Mystic Sanctuary must enter tapped below three other Islands")
	}
}

// Secluded Courtyard's coloured mana is spendable on a creature
// SPELL of the named type and on an ABILITY of one — the "or" Cavern
// of Souls does not have.
func TestSecludedCourtyardManaWorksForCastsAndActivations(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.PushNamedTribePermanent(t, g, me.ID, "Secluded Courtyard", "Land", effectstest.SecludedCourtyardOracle, "Elf")

	spec, ok := effects.Lookup(effectstest.SecludedCourtyardOracle)
	if !ok || len(spec.ManaAbilities) != 2 {
		t.Fatal("Secluded Courtyard must declare two mana abilities")
	}
	tags := spec.ManaAbilities[1].RestrictionsFunc(g, me.ID, id)
	for _, want := range []string{game.ManaRestrictType("Creature"), game.ManaRestrictSubtype("Elf")} {
		if !effectstest.ContainsString(tags, want) {
			t.Errorf("restriction tags %v are missing %q", tags, want)
		}
	}
	if effectstest.ContainsString(tags, game.ManaRestrictCast) {
		t.Error("a purpose tag would make the mana casts-only; the card also allows activations")
	}
}

// Nykthos reads devotion per colour into the produced-mana grammar.
func TestNykthosProducesPerColourDevotion(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.SeedPermanentWithOracle(g, me.ID, "Nykthos, Shrine to Nyx", "Legendary Land", effectstest.NykthosOracle)
	for i := 0; i < 2; i++ {
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid",
			ManaCost: "{G}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
		})
	}
	spec, ok := effects.Lookup(effectstest.NykthosOracle)
	if !ok || len(spec.ManaAbilities) != 2 {
		t.Fatal("Nykthos must declare two mana abilities")
	}
	if spec.ManaAbilities[1].Cost.Mana != "{2}" {
		t.Errorf("the devotion ability costs %q, want {2}", spec.ManaAbilities[1].Cost.Mana)
	}
	if got := effectstest.ProducedOf(spec.ManaAbilities[1], g, me.ID, id); got != "{G2}" {
		t.Errorf("produced = %q, want {G2} (devotion to green is two)", got)
	}
}

// Three Tree City mints N of ONE colour, N being the creatures of the
// named type.
func TestThreeTreeCityScalesWithTheNamedTribe(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.PushNamedTribePermanent(t, g, me.ID, "Three Tree City", "Legendary Land", effectstest.ThreeTreeCityOracle, "Elf")
	for i := 0; i < 3; i++ {
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid",
			Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
		})
	}
	spec, _ := effects.Lookup(effectstest.ThreeTreeCityOracle)
	if got := effectstest.ProducedOf(spec.ManaAbilities[1], g, me.ID, id); got != effects.OneColorOfAmount(3) {
		t.Errorf("produced = %q, want %q", got, effects.OneColorOfAmount(3))
	}
}

// Patchwork Banner pumps the named type and nothing else.
func TestPatchworkBannerPumpsTheNamedType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushNamedTribePermanent(t, g, me.ID, "Patchwork Banner", "Artifact", effectstest.PatchworkBannerOracle, "Elf")
	elf := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	if got := effectstest.EffectivePower(t, g, elf); got != 2 {
		t.Errorf("the Elf is %d power, want 2", got)
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Errorf("the Bear is %d power, want 2 — it is not the named type", got)
	}
}

// Yavimaya makes every land a Forest, its controller's and everyone
// else's, and the Forest really taps for {G}.
func TestYavimayaMakesEveryLandAForest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.SeedLand(g, me.ID, "Island", "Basic Land — Island", "")
	theirs := effectstest.SeedLand(g, opp.ID, "Swamp", "Basic Land — Swamp", "")
	yav := effectstest.SeedLand(g, me.ID, "Yavimaya, Cradle of Growth", "Legendary Land", effectstest.YavimayaOracle)

	for _, tc := range []struct {
		name string
		id   uuid.UUID
	}{{"your Island", mine}, {"an opponent's Swamp", theirs}, {"Yavimaya itself", yav}} {
		if !effectstest.ContainsString(effectstest.EffectiveSubtypes(t, g, tc.id), "Forest") {
			t.Errorf("%s is not a Forest", tc.name)
		}
	}
	// "In addition to": the Island keeps its own type.
	if !effectstest.ContainsString(effectstest.EffectiveSubtypes(t, g, mine), "Island") {
		t.Error("the Island lost its own land type — this is an add, not a set")
	}
}

// --- the creatures and permanents ----------------------------------

// Scute Swarm makes an Insect below six lands and a copy of itself at
// six.
func TestScuteSwarmCopiesItselfAtSixLands(t *testing.T) {
	for _, tc := range []struct {
		name      string
		lands     int
		wantSwarm int
	}{
		{"five lands", 4, 1},
		{"six lands", 5, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.SeedPermanentWithOracle(g, me.ID, "Scute Swarm", "Creature — Insect", effectstest.ScuteSwarmOracle)
			for i := 0; i < tc.lands; i++ {
				effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
			}
			effectstest.PlayLandFromHand(t, g, "Forest", "")
			effectstest.PassPriorityAroundTable(t, g)

			swarms, insects := 0, 0
			for _, c := range g.Battlefield.Cards {
				switch c.Name {
				case "Scute Swarm":
					swarms++
				case "Insect":
					insects++
				}
			}
			if swarms != tc.wantSwarm {
				t.Errorf("Scute Swarms = %d, want %d", swarms, tc.wantSwarm)
			}
			if want := 1 - (tc.wantSwarm - 1); insects != want {
				t.Errorf("plain Insects = %d, want %d", insects, want)
			}
		})
	}
}

// Black Market Connections asks at the first main phase and runs
// every bullet the controller takes.
func TestBlackMarketConnectionsRunsTheChosenBullets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedPermanentWithOracle(g, me.ID, "Black Market Connections", "Enchantment", effectstest.BlackMarketConnectionsOracle)
	life, hand := me.Life, me.Hand.Size()

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	mode := effectstest.LatestChoiceOfKind(g, game.PendingChoiceModePick)
	if mode == nil {
		t.Fatal("no mode prompt at the first main phase")
	}
	if err := g.ResolveModePick(mode.ID, me.ID, []int{0, 1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := me.Life; got != life-3 {
		t.Errorf("life %d → %d, want %d (1 + 2)", life, got, life-3)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand = %d, want %d", got, hand+1)
	}
	if effectstest.FindBattlefieldByName(g, "Treasure") == uuid.Nil {
		t.Error("no Treasure token")
	}
}

// Herald's Horn discounts creature spells of the named type only, and
// its upkeep offers the top card when it matches.
func TestHeraldsHornDiscountsAndFiltersTheTopCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	horn := effectstest.PushNamedTribePermanent(t, g, me.ID, "Herald's Horn", "Artifact", effectstest.HeraldsHornOracle, "Elf")

	mods := game.CostModifiersForCard(game.Card{OracleID: effectstest.HeraldsHornOracle})
	if len(mods) != 1 {
		t.Fatalf("%d cost modifiers, want 1", len(mods))
	}
	hornCard, _ := effectstest.BattlefieldCard(g, horn)
	elf := game.Card{Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid"}
	bear := game.Card{Name: "Bear", TypeLine: "Creature — Bear"}
	if !mods[0].AppliesTo(game.CostQuery{Game: g, Card: elf, Controller: me.ID, Source: hornCard}) {
		t.Error("the discount must apply to a creature spell of the named type")
	}
	if mods[0].AppliesTo(game.CostQuery{Game: g, Card: bear, Controller: me.ID, Source: hornCard}) {
		t.Error("the discount must not apply to a creature of another type")
	}
}

// The Great Henge's mana ability gains two life alongside its {G}{G},
// and its entry trigger grows the creature and draws.
func TestTheGreatHengeRiderAndEntryTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	henge := effectstest.SeedPermanentWithOracle(g, me.ID, "The Great Henge", "Legendary Artifact", effectstest.TheGreatHengeOracle)
	life, hand := me.Life, me.Hand.Size()

	if err := g.ActivateManaAbility(me.ID, henge, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := me.Life; got != life+2 {
		t.Errorf("life %d → %d, want %d — the rider gains two", life, got, life+2)
	}
	if me.ManaPool.Total() != 2 {
		t.Errorf("mana pool holds %d tokens, want 2", me.ManaPool.Total())
	}

	bear := effectstest.PushCatalogHandCard(me, "Bear", "Creature — Bear", "")
	g.WithWriteLock(func() {
		_, _ = g.PutFromHandOntoBattlefieldForEffect(bear, game.HandEntryOptions{Controller: me.ID})
	})
	effectstest.PassPriorityAroundTable(t, g)
	card, ok := effectstest.BattlefieldCard(g, bear)
	if !ok || card.Counters["+1/+1"] != 1 {
		t.Error("the entering creature did not get a +1/+1 counter")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand = %d, want %d — the trigger draws", got, hand+1)
	}
}

// The Henge's own cost reduction reads the GREATEST power, not the
// total.
func TestTheGreatHengeReducesByTheGreatestPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for _, p := range []int{2, 5, 3} {
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Beater", TypeLine: "Creature — Beast",
			Power: p, Toughness: p, Owner: me.ID, Controller: me.ID,
		})
	}
	spec, _ := effects.Lookup(effectstest.TheGreatHengeOracle)
	if len(spec.SelfCostModifiers) != 1 {
		t.Fatalf("%d self cost modifiers, want 1", len(spec.SelfCostModifiers))
	}
	if got := spec.SelfCostModifiers[0].Amount(game.CostQuery{Game: g, Controller: me.ID}); got != 5 {
		t.Errorf("reduction = %d, want 5 (the greatest power, not 2+5+3)", got)
	}
}

// Roaming Throne is the type it named, and doubles another creature
// of that type rather than its own abilities.
func TestRoamingThroneIsTheNamedTypeAndDoublesOthers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	throne := effectstest.PushNamedTribePermanent(t, g, me.ID, "Roaming Throne", "Artifact Creature — Golem", effectstest.RoamingThroneOracle, "Elf")
	if !effectstest.ContainsString(effectstest.EffectiveSubtypes(t, g, throne), "Elf") {
		t.Error("the Throne is not the type it named")
	}
	if !effectstest.ContainsString(effectstest.EffectiveSubtypes(t, g, throne), "Golem") {
		t.Error("\"in addition to its other types\" must keep Golem")
	}

	doublers := game.CatalogTriggerDoublers(effectstest.RoamingThroneOracle)
	if len(doublers) != 1 {
		t.Fatalf("%d trigger doublers, want 1", len(doublers))
	}
	throneCard, _ := effectstest.BattlefieldCard(g, throne)
	elf := game.Card{InstanceID: uuid.New(), Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid", Controller: me.ID}
	elfLKI := game.Characteristic{Controller: me.ID, Types: []string{"Creature"}, Subtypes: []string{"Elf"}}
	ability := &game.TriggeredAbility{}
	if !doublers[0].Applies(g, game.TriggerDoublingQuery{
		Doubler: throneCard, DoublerLKI: game.Characteristic{Controller: me.ID},
		Source: elf, SourceLKI: elfLKI, Ability: ability,
	}) {
		t.Error("another Elf you control must be doubled")
	}
	if doublers[0].Applies(g, game.TriggerDoublingQuery{
		Doubler: throneCard, DoublerLKI: game.Characteristic{Controller: me.ID},
		Source: throneCard, SourceLKI: game.Characteristic{Controller: me.ID}, Ability: ability,
	}) {
		t.Error("\"another\" must exclude the Throne itself")
	}
}

// --- the channel lands and the utility lands -----------------------

// Boseiju's channel is a hand ability that destroys and compensates.
func TestBoseijuChannelDestroysAndOffersABasic(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	victim := effectstest.SeedLandOnBattlefield(g, opp.ID, "Gaea's Cradle", "Legendary Land")
	effectstest.PushLibraryCardForTest(opp, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	id := effectstest.PushCatalogHandCard(me, "Boseiju, Who Endures", "Legendary Land", effectstest.BoseijuWhoEnduresOracle)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{G}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("channel: %v", err)
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the discard-this cost must bin Boseiju at announce")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the nonbasic land was not destroyed")
	}
}

// Otawara's channel bounces anything, including your own permanent.
func TestOtawaraChannelBouncesItsTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	id := effectstest.PushCatalogHandCard(me, "Otawara, Soaring City", "Legendary Land", effectstest.OtawaraSoaringCityOracle)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{U}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("channel: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature was not returned to hand")
	}
	if !me.Hand.Contains(bear) {
		t.Error("the creature did not arrive in its owner's hand")
	}
}

// Command Beacon trades itself for the commander in the command zone.
func TestCommandBeaconPutsTheCommanderInHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	beacon := effectstest.SeedPermanentWithOracle(g, me.ID, "Command Beacon", "Land", effectstest.CommandBeaconOracle)
	cmdr := uuid.New()
	me.Command.PushTop(game.Card{
		InstanceID: cmdr, Name: "Kenrith, the Returned King", TypeLine: "Legendary Creature — Human Noble",
		Owner: me.ID, Controller: me.ID, IsCommander: true,
	})
	if err := g.ActivateCatalogAbility(me.ID, beacon, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// A commander headed for a hand gets the CR 903.9b offer, from the
	// command zone like anywhere else. Declining is what the card is
	// for.
	if offer := effectstest.LatestChoiceOfKind(g, game.PendingChoiceOptionalReplacement); offer != nil {
		if err := g.ResolveOptionalReplacement(offer.ID, me.ID, false); err != nil {
			t.Fatalf("ResolveOptionalReplacement: %v", err)
		}
	}
	if g.Battlefield.Contains(beacon) {
		t.Error("the sacrifice cost did not eat the land")
	}
	if me.Command.Contains(cmdr) {
		t.Error("the commander is still in the command zone")
	}
	if !me.Hand.Contains(cmdr) {
		t.Error("the commander did not reach the hand")
	}
}

// Gemstone Caverns makes {C} bare and any colour with a luck counter.
func TestGemstoneCavernsNeedsALuckCounterForColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.SeedPermanentWithOracle(g, me.ID, "Gemstone Caverns", "Legendary Land", effectstest.GemstoneCavernsOracle)
	spec, _ := effects.Lookup(effectstest.GemstoneCavernsOracle)
	if got := effectstest.ProducedOf(spec.ManaAbilities[0], g, me.ID, id); got != "{C}" {
		t.Errorf("produced without a luck counter = %q, want {C}", got)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, "luck", 1) })
	if got := effectstest.ProducedOf(spec.ManaAbilities[0], g, me.ID, id); got != "{W|U|B|R|G}" {
		t.Errorf("produced with a luck counter = %q, want the five-colour pick", got)
	}
}
