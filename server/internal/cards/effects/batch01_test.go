package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch01_test.go — card-level coverage for the card-coverage
// roadmap's batch 01 (#294, `edhrec_rank` 9–237): the "no new
// machinery" group, the two until-end-of-turn cards #314 unblocked,
// and the two mana-from-a-spell cards the AddMana primitive
// unblocked. One test per observable behaviour, driven through a
// real cast or activation rather than by calling primitives.

const (
	arcaneDenialOracle            = "ab1cc360-b9de-48d9-9983-4dfe4a7d2a37"
	manaDrainOracle               = "74d3277a-38e5-4732-afed-084a56148f20"
	victimizeOracle               = "240e85d3-e495-4877-8609-4b4056c402f7"
	spectatorSeatingOracle        = "cf6d10ed-85c3-48f2-8ba0-2960e03b408b"
	vaultOfChampionsOracle        = "ebc5ac83-08d4-4d6b-b840-0c4ba71a38ab"
	farewellOracle                = "4eb813fd-2d5a-4b02-8193-662681ef4e7d"
	undergrowthStadiumOracle      = "7c69f718-acc8-4851-8e5d-0cbaaa86192c"
	dreamrootCascadeOracle        = "dd8538e6-cd5f-4a88-aff5-eb5e76ce8ddb"
	tirelessProvisionerOracle     = "ab8d5f5c-1976-4f77-8ed2-8d28ee666741"
	stormcarvedCoastOracle        = "4722105b-0085-4bb8-bca1-9de0d3eb5600"
	spireGardenOracle             = "45fe016e-1a09-410c-bbe3-4663ba06c5b7"
	buriedRuinOracle              = "3644f316-f9a3-46c9-9b1e-747f86cf4ead"
	impactTremorsOracle           = "9242cd3e-1a71-4700-8182-9c1005616033"
	phyrexianTowerOracle          = "1861e642-21d5-4232-89f3-b5557f2946c1"
	stormKilnArtistOracle         = "a145ff8c-5812-4bcb-bd16-9839dc25121d"
	strokeOfMidnightOracle        = "9a107e48-3d50-4941-95b1-10f2b29a4245"
	rapidHybridizationOracle      = "06692cd9-ac2f-4a32-8fd1-043ba3c0fe71"
	professionalFaceBreakerOracle = "04152e7a-969c-4858-841b-0a569a9fc1bf"
	bountifulPromenadeOracle      = "761cb262-f83b-4a99-9345-b773182a7671"
	fyndhornElvesOracle           = "df317532-7d36-40fd-938f-e972749c8792"
	ornithopterOfParadiseOracle   = "3a940dfa-a026-4969-8981-ef90cdfbe9ef"
	karplusanForestOracle         = "bd912666-f37f-4767-af6f-9e6d0fcccacf"
	terminateOracle               = "6257c2fd-005f-41e3-8a72-af76df1eb134"
	rockfallValeOracle            = "185c70c1-8403-4ae5-b45d-3679d4ee092a"
	mirkwoodBatsOracle            = "0636b6c3-0662-420a-b30d-f0a14e7c512d"
	brushlandOracle               = "5eb8b497-ec9a-4a89-ad29-1ec3ca82da7c"
	darkRitualOracle              = "53f7c868-b03e-4fc2-8dcf-a75bbfa3272b"
	returnOfTheWildspeakerOracle  = "2b76f9e9-cd28-4eaf-8674-215c34263f96"
	borosCharmOracle              = "2679d0dd-ba30-4a1c-b6a0-b3ac6c790496"
)

// batch01AdvanceToStepOf walks the cursor until `seat` is active and
// the turn is at `step`.
func batch01AdvanceToStepOf(t *testing.T, g *game.Game, seat int, step game.Step) {
	t.Helper()
	for i := 0; i < 400; i++ {
		if g.Turn.Step == step && g.Turn.ActiveSeat == seat {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep toward %s of seat %d: %v", step, seat, err)
		}
	}
	t.Fatalf("never reached %s of seat %d", step, seat)
}

// batch01OpponentCasts puts a card in a non-active player's hand and
// casts it at instant speed (the sorcery gate only applies to
// non-instants), returning its ID so a counterspell can target it.
func batch01OpponentCasts(t *testing.T, g *game.Game, opp *game.Player, name, oracle, manaCost string, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	id := uuid.New()
	opp.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Instant", OracleID: oracle, ManaCost: manaCost,
		Owner: opp.ID, Controller: opp.ID,
	})
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{Targets: targets}); err != nil {
		t.Fatalf("opponent CastSpell %s: %v", name, err)
	}
	return id
}

func batch01PoolColors(p *game.Player) []string {
	out := make([]string, 0, len(p.ManaPool))
	for _, tok := range p.ManaPool {
		out = append(out, tok.Color)
	}
	return out
}

func batch01GraveyardCard(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine,
		Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID,
	})
	return id
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. The land
// cycles are loops over tables, so a transposed row is invisible
// until someone plays that exact card.
func TestBatch01CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		arcaneDenialOracle:            "Arcane Denial",
		manaDrainOracle:               "Mana Drain",
		victimizeOracle:               "Victimize",
		spectatorSeatingOracle:        "Spectator Seating",
		vaultOfChampionsOracle:        "Vault of Champions",
		farewellOracle:                "Farewell",
		undergrowthStadiumOracle:      "Undergrowth Stadium",
		dreamrootCascadeOracle:        "Dreamroot Cascade",
		tirelessProvisionerOracle:     "Tireless Provisioner",
		stormcarvedCoastOracle:        "Stormcarved Coast",
		spireGardenOracle:             "Spire Garden",
		buriedRuinOracle:              "Buried Ruin",
		impactTremorsOracle:           "Impact Tremors",
		phyrexianTowerOracle:          "Phyrexian Tower",
		stormKilnArtistOracle:         "Storm-Kiln Artist",
		strokeOfMidnightOracle:        "Stroke of Midnight",
		rapidHybridizationOracle:      "Rapid Hybridization",
		professionalFaceBreakerOracle: "Professional Face-Breaker",
		bountifulPromenadeOracle:      "Bountiful Promenade",
		fyndhornElvesOracle:           "Fyndhorn Elves",
		ornithopterOfParadiseOracle:   "Ornithopter of Paradise",
		karplusanForestOracle:         "Karplusan Forest",
		terminateOracle:               "Terminate",
		rockfallValeOracle:            "Rockfall Vale",
		mirkwoodBatsOracle:            "Mirkwood Bats",
		brushlandOracle:               "Brushland",
		darkRitualOracle:              "Dark Ritual",
		returnOfTheWildspeakerOracle:  "Return of the Wildspeaker",
		borosCharmOracle:              "Boros Charm",
	}
	if len(want) != 29 {
		t.Fatalf("the batch is 29 cards, the table lists %d", len(want))
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
	for _, spec := range All() {
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
			if spec.ManaAbilities[1].Produced != want || spec.ManaAbilities[1].Rider == nil || spec.ManaAbilities[1].NarrowToCommanderIdentity {
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
	g := newCatalogGame(t)
	id := playLandFromHand(t, g, "Spectator Seating", spectatorSeatingOracle)
	top100AssertEnteredUntapped(t, g, id, "Spectator Seating")
}

func TestBountifulPromenadeEntersTappedInADuel(t *testing.T) {
	g := newDuelCatalogGame(t)
	id := playLandFromHand(t, g, "Bountiful Promenade", bountifulPromenadeOracle)
	top100AssertEnteredTapped(t, g, id, "Bountiful Promenade")
}

func TestDreamrootCascadeEntersTappedWithOneOtherLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	id := playLandFromHand(t, g, "Dreamroot Cascade", dreamrootCascadeOracle)
	top100AssertEnteredTapped(t, g, id, "Dreamroot Cascade")
}

func TestRockfallValeEntersUntappedWithTwoOtherLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	seedLandOnBattlefield(g, me.ID, "Mountain", "Basic Land — Mountain")
	id := playLandFromHand(t, g, "Rockfall Vale", rockfallValeOracle)
	top100AssertEnteredUntapped(t, g, id, "Rockfall Vale")
}

func TestKarplusanForestColoredHalfDealsOneDamage(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Life
	land := seedPermanentWithOracle(g, me.ID, "Karplusan Forest", "Land", karplusanForestOracle)

	if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.Life != before-1 {
		t.Errorf("life %d → %d, want %d", before, me.Life, before-1)
	}
	pick := riderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 2 {
		t.Fatalf("the colored half must ask R or G, got %+v", pick)
	}
}

func TestBrushlandColorlessHalfIsFree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Life
	land := seedPermanentWithOracle(g, me.ID, "Brushland", "Land", brushlandOracle)

	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.Life != before {
		t.Errorf("the {C} half hurt: %d → %d", before, me.Life)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
}

// --- mana dorks ----------------------------------------------------

func TestFyndhornElvesTapsForGreen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	elf := pushCatalogPermanent(g, me.ID, "Fyndhorn Elves", "Creature — Elf Druid", fyndhornElvesOracle, false)
	if err := g.ActivateManaAbility(me.ID, elf, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool %v, want [G]", got)
	}
}

func TestOrnithopterOfParadiseFliesAndFixes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	thopter := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ornithopter of Paradise",
		TypeLine: "Artifact Creature — Thopter", OracleID: ornithopterOfParadiseOracle,
		Power: 0, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	if !eotHasAbility(effectiveAbilities(t, g, thopter), "flying") {
		t.Error("printed flying did not reach the effective abilities")
	}
	if err := g.ActivateManaAbility(me.ID, thopter, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := riderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("with no commander the pick must offer all five colours, got %+v", pick)
	}
}

// --- removal -------------------------------------------------------

func TestTerminateDestroysACreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := seedCreature(g, "Bear", opp.ID)
	castCatalogSpell(t, g, "Terminate", "Instant", terminateOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Graveyard.Contains(bear) {
		t.Error("the creature should be in its owner's graveyard")
	}
}

func TestRapidHybridizationGivesTheVictimAFrogLizard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := seedCreature(g, "Bear", opp.ID)
	castCatalogSpell(t, g, "Rapid Hybridization", "Instant", rapidHybridizationOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature survived")
	}
	if countBattlefieldNamed(g, opp.ID, "Frog Lizard") != 1 {
		t.Error("the VICTIM should get the 3/3 Frog Lizard")
	}
	if countBattlefieldNamed(g, me.ID, "Frog Lizard") != 0 {
		t.Error("the caster must not get the token")
	}
}

func TestStrokeOfMidnightDestroysANonlandAndHandsBackAHuman(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	rock := seedPermanentFor(g, opp.ID, "Rock", "Artifact")
	castCatalogSpell(t, g, "Stroke of Midnight", "Instant", strokeOfMidnightOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
	if countBattlefieldNamed(g, opp.ID, "Human") != 1 {
		t.Error("its controller should get a 1/1 Human")
	}
}

func TestStrokeOfMidnightRefusesALand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := seedLandOnBattlefield(g, opp.ID, "Forest", "Basic Land — Forest")
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Stroke of Midnight", TypeLine: "Instant",
		OracleID: strokeOfMidnightOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}},
	}); err == nil {
		t.Error("a land was accepted as a target for 'target nonland permanent'")
	}
}

// --- go-wide payoffs -----------------------------------------------

func TestImpactTremorsPingsOncePerCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Impact Tremors", "Enchantment", impactTremorsOracle, false)
	before := lifeOfOpponents(g)

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 2) })
	passPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-2 {
			t.Errorf("opponent %d: %d -> %d, want -2 (two creatures, two triggers)", i+1, b, got)
		}
	}

	// A noncreature entering, or an opponent's creature, is silent.
	mid := lifeOfOpponents(g)
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, TreasureToken(), 1)
		_ = g.CreateTokenForEffect(opp.ID, RedGoblinToken(), 1)
	})
	passPriorityAroundTable(t, g)
	for i, b := range mid {
		if got := g.Seats[i+1].Life; got != b {
			t.Errorf("opponent %d: %d -> %d, want unchanged", i+1, b, got)
		}
	}
}

func TestMirkwoodBatsDrainsOnTokenCreatedAndTokenSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Mirkwood Bats", "Creature — Bat", mirkwoodBatsOracle, false)
	before := lifeOfOpponents(g)

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TreasureToken(), 1) })
	passPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("opponent %d after a token was created: %d -> %d, want -1", i+1, b, got)
		}
	}

	treasure := findBattlefieldByName(g, "Treasure")
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(treasure) })
	passPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-2 {
			t.Errorf("opponent %d after the token was sacrificed: %d -> %d, want -2", i+1, b, got)
		}
	}

	// A nontoken sacrifice and an opponent's token are both silent.
	bear := seedCreature(g, "Bear", me.ID)
	g.WithWriteLock(func() {
		_ = g.SacrificePermanentForEffect(bear)
		_ = g.CreateTokenForEffect(opp.ID, TreasureToken(), 1)
	})
	passPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-2 {
			t.Errorf("opponent %d: %d -> %d, want still -2", i+1, b, got)
		}
	}
}

func TestTirelessProvisionerMakesATreasureOnLandfall(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Tireless Provisioner", "Creature — Elf Scout", tirelessProvisionerOracle, false)

	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	if countBattlefieldNamed(g, me.ID, "Treasure") != 1 {
		t.Fatal("a land drop should make one Treasure")
	}

	// An opponent's land is not "a land YOU control".
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(opp.ID, game.Card{Name: "Forest", TypeLine: "Token Land — Forest"}, 1)
	})
	passPriorityAroundTable(t, g)
	if countBattlefieldNamed(g, me.ID, "Treasure") != 1 {
		t.Error("an opponent's land triggered my landfall")
	}
}

// --- Phyrexian Tower -----------------------------------------------

func TestPhyrexianTowerSacrificesACreatureForTwoBlack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tower := seedPermanentWithOracle(g, me.ID, "Phyrexian Tower", "Legendary Land", phyrexianTowerOracle)
	bear := seedCreature(g, "Bear", me.ID)

	if err := g.ActivateManaAbility(me.ID, tower, 1, game.ManaAbilityParams{
		SacrificeIDs: []uuid.UUID{bear},
	}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 2 || got[0] != "B" || got[1] != "B" {
		t.Errorf("pool %v, want [B B]", got)
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the creature was not sacrificed")
	}
	card, _ := battlefieldCard(g, tower)
	if !card.Tapped {
		t.Error("the Tower has a tap cost and must be tapped")
	}
}

func TestPhyrexianTowerBlackHalfNeedsACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tower := seedPermanentWithOracle(g, me.ID, "Phyrexian Tower", "Legendary Land", phyrexianTowerOracle)
	if err := g.ActivateManaAbility(me.ID, tower, 1, game.ManaAbilityParams{}); err == nil {
		t.Fatal("activated the sacrifice half with nothing to sacrifice")
	}
	card, _ := battlefieldCard(g, tower)
	if card.Tapped {
		t.Error("a refused activation must leave the Tower untapped")
	}
	if err := g.ActivateManaAbility(me.ID, tower, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("the colorless half: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
}

// --- Buried Ruin ---------------------------------------------------

func TestBuriedRuinReturnsAnArtifactCardToHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ruin := pushCatalogPermanent(g, me.ID, "Buried Ruin", "Land", buriedRuinOracle, false)
	relic := batch01GraveyardCard(me, "Sol Ring", "Artifact")
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})

	if err := g.ActivateCatalogAbility(me.ID, ruin, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: relic}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(ruin) {
		t.Error("the Ruin is sacrificed as a cost, at announce")
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(relic) {
		t.Error("the artifact card did not come back to hand")
	}
}

func TestBuriedRuinRefusesANonArtifactCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ruin := pushCatalogPermanent(g, me.ID, "Buried Ruin", "Land", buriedRuinOracle, false)
	bear := pushGraveyardCardForTest(me, "Bear")
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myBear := seedCreature(g, "My Bear", me.ID)
	theirBear := seedCreature(g, "Their Bear", opp.ID)
	rock := seedPermanentFor(g, opp.ID, "Rock", "Artifact")
	aura := seedPermanentFor(g, opp.ID, "Glory", "Enchantment")
	myDead := pushGraveyardCardForTest(me, "My Dead")
	theirDead := pushGraveyardCardForTest(opp, "Their Dead")

	castModal(t, g, "Farewell", "Sorcery", farewellOracle, []int{1, 3}, nil)
	passPriorityAroundTable(t, g)

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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	artist := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Storm-Kiln Artist", TypeLine: "Creature — Dwarf Shaman",
		OracleID: stormKilnArtistOracle, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	if p := effectivePower(t, g, artist); p != 2 {
		t.Fatalf("with no artifacts power = %d, want 2", p)
	}

	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)

	if countBattlefieldNamed(g, me.ID, "Treasure") != 1 {
		t.Fatal("casting an instant should make one Treasure")
	}
	if p := effectivePower(t, g, artist); p != 3 {
		t.Errorf("with one artifact power = %d, want 3", p)
	}

	// A creature spell is not an instant or sorcery.
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if countBattlefieldNamed(g, me.ID, "Treasure") != 1 {
		t.Error("a creature spell triggered magecraft")
	}
}

// --- Arcane Denial -------------------------------------------------

func TestArcaneDenialCountersAndBothPlayersDrawAtTheNextUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := batch01OpponentCasts(t, g, opp, "Lightning Bolt", lightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})

	castCatalogSpell(t, g, "Arcane Denial", "Instant", arcaneDenialOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	passPriorityAroundTable(t, g)

	if !opp.Graveyard.Contains(bolt) {
		t.Fatal("the Bolt was not countered")
	}
	if n := len(g.DelayedTriggers); n != 1 {
		t.Fatalf("delayed triggers = %d, want 1 (both draws on one item — see the card comment)", n)
	}

	// Nothing is drawn before the next upkeep.
	advanceToUpkeepOf(t, g, 1)
	oppBefore, meBefore := opp.Hand.Size(), me.Hand.Size()
	if len(g.PendingTriggers)+len(g.StackMeta) == 0 {
		t.Fatal("the draws should be on the stack at the next upkeep")
	}
	passPriorityAroundTable(t, g)

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

// --- Mana Drain ----------------------------------------------------

// TestManaDrainRefundsOnYourOwnPostcombatMainWhenCastDuringYourPrecombatMain
// pins the fix for the closed caveat (#1565): a Drain cast during your
// own precombat main pays out THIS turn's postcombat main, not the
// next turn's precombat main — the "your next main phase" reading a
// delayed trigger naming one step could not express before
// manaDrainNextMainPhaseStep picked between the two at resolution.
func TestManaDrainRefundsOnYourOwnPostcombatMainWhenCastDuringYourPrecombatMain(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell := batch01OpponentCasts(t, g, opp, "Big Bolt", lightningBoltOracle, "{2}{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})

	// Cast during MY OWN precombat main (castCatalogSpell always casts
	// as the active seat, and it's seat 0's turn 1).
	castCatalogSpell(t, g, "Mana Drain", "Instant", manaDrainOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	passPriorityAroundTable(t, g)

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
	batch01AdvanceToStepOf(t, g, 0, game.StepPostcombatMain)
	if len(g.DelayedTriggers) != 0 {
		t.Fatal("the refund did not fire on the caster's own postcombat main this turn")
	}
	passPriorityAroundTable(t, g)
	if got := batch01PoolColors(me); len(got) != 3 || got[0] != "C" || got[1] != "C" || got[2] != "C" {
		t.Errorf("pool %v, want [C C C] — the countered spell's mana value", got)
	}
}

// TestManaDrainRefundsOnYourOwnNextPrecombatMainWhenCastOnAnOpponentsTurn
// is the case that was already correct: countering a spell on someone
// else's turn waits for the caster's own NEXT precombat main, however
// many turns away — ControllerTurnOnly's whole reason to exist.
func TestManaDrainRefundsOnYourOwnNextPrecombatMainWhenCastOnAnOpponentsTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	batch01AdvanceToStepOf(t, g, 1, game.StepPrecombatMain)
	spell := batch01OpponentCasts(t, g, opp, "Big Bolt", lightningBoltOracle, "{2}{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})

	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Mana Drain", TypeLine: "Instant",
		OracleID: manaDrainOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: spell}},
	}); err != nil {
		t.Fatalf("CastSpell Mana Drain: %v", err)
	}
	passPriorityAroundTable(t, g)

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
	batch01AdvanceToStepOf(t, g, 1, game.StepPostcombatMain)
	if len(g.DelayedTriggers) != 1 {
		t.Fatal("the refund fired on an opponent's main phase")
	}

	// My own next precombat main is.
	batch01AdvanceToStepOf(t, g, 0, game.StepPrecombatMain)
	if len(g.DelayedTriggers) != 0 {
		t.Fatal("the refund did not fire on the caster's own next precombat main")
	}
	passPriorityAroundTable(t, g)
	if got := batch01PoolColors(me); len(got) != 3 || got[0] != "C" || got[1] != "C" || got[2] != "C" {
		t.Errorf("pool %v, want [C C C] — the countered spell's mana value", got)
	}
}

func TestManaDrainOnAZeroValueSpellSchedulesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell := batch01OpponentCasts(t, g, opp, "Free Bolt", lightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	castCatalogSpell(t, g, "Mana Drain", "Instant", manaDrainOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(spell) {
		t.Fatal("the spell was not countered")
	}
	if n := len(g.DelayedTriggers); n != 0 {
		t.Errorf("a mana value of 0 scheduled %d refunds", n)
	}
}

// --- Dark Ritual ---------------------------------------------------

func TestDarkRitualAddsThreeBlackMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Dark Ritual", "Instant", darkRitualOracle, nil)
	passPriorityAroundTable(t, g)
	if got := batch01PoolColors(me); len(got) != 3 || got[0] != "B" || got[1] != "B" || got[2] != "B" {
		t.Errorf("pool %v, want [B B B]", got)
	}
	if me.Graveyard.Size() != 1 {
		t.Errorf("Dark Ritual should be in the graveyard after resolving")
	}
}

// --- Victimize -----------------------------------------------------

func TestVictimizeSacrificesAtCastAndReturnsBothTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := pushGraveyardCardForTest(me, "Dead A")
	b := pushGraveyardCardForTest(me, "Dead B")
	fodder := seedCreature(g, "Fodder", me.ID)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Victimize", TypeLine: "Sorcery",
		OracleID: victimizeOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}},
		SacrificeIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrifice is a cost and is paid at announce")
	}
	passPriorityAroundTable(t, g)
	for _, want := range []uuid.UUID{a, b} {
		card, ok := battlefieldCard(g, want)
		if !ok {
			t.Errorf("%s did not return to the battlefield", want)
			continue
		}
		if !card.Tapped {
			t.Errorf("%s returned untapped", want)
		}
	}
}

func TestVictimizeNeedsExactlyTwoTargets(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := pushGraveyardCardForTest(me, "Dead A")
	fodder := seedCreature(g, "Fodder", me.ID)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Victimize", TypeLine: "Sorcery",
		OracleID: victimizeOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: a}},
		SacrificeIDs: []uuid.UUID{fodder},
	}); err == nil {
		t.Error("one target was accepted for 'two target creature cards'")
	}
	if !g.Battlefield.Contains(fodder) {
		t.Error("a refused cast must not pay the sacrifice")
	}
}

// --- Professional Face-Breaker -------------------------------------

func TestProfessionalFaceBreakerMakesOneTreasurePerCombatDamageStep(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	breaker := pushCatalogPermanent(g, me.ID, "Professional Face-Breaker", "Creature — Human Warrior", professionalFaceBreakerOracle, false)
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)

	attackWith(t, g, victim.ID, a, b)
	passPriorityAroundTable(t, g)

	if n := countBattlefieldNamed(g, me.ID, "Treasure"); n != 1 {
		t.Fatalf("two creatures connecting made %d Treasures, want 1 — 'one or more' is one trigger", n)
	}
	if !eotHasAbility(effectiveAbilities(t, g, breaker), "menace") {
		t.Error("printed menace did not reach the effective abilities")
	}

	// Crack the Treasure for an impulse card.
	top := seedLibrary(me, "Top Card")[0]
	treasure := findBattlefieldByName(g, "Treasure")
	if err := g.ActivateCatalogAbility(me.ID, breaker, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{treasure},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(top) {
		t.Fatal("the top card was not exiled")
	}
	perm := exiledPermission(g, top)
	if perm.Player != me.ID {
		t.Errorf("the play permission belongs to %s, want the controller", perm.Player)
	}
	if perm.CastOnly {
		t.Error("the printed text says PLAY — a land must not be stranded")
	}
}

func TestProfessionalFaceBreakerNeedsATreasureToCrack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	breaker := pushCatalogPermanent(g, me.ID, "Professional Face-Breaker", "Creature — Human Warrior", professionalFaceBreakerOracle, false)
	bear := seedCreature(g, "Bear", me.ID)
	if err := g.ActivateCatalogAbility(me.ID, breaker, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{bear},
	}); err == nil {
		t.Error("a creature was accepted for 'Sacrifice a Treasure'")
	}
}

// --- Return of the Wildspeaker -------------------------------------

func batch01WildspeakerBoard(g *game.Game, me uuid.UUID) (bear, human uuid.UUID) {
	bear = pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
		Power: 4, Toughness: 4, Owner: me, Controller: me,
	})
	human = pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Knight", TypeLine: "Creature — Human Knight",
		Power: 6, Toughness: 6, Owner: me, Controller: me,
	})
	return bear, human
}

func TestReturnOfTheWildspeakerDrawsTheGreatestNonHumanPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	batch01WildspeakerBoard(g, me.ID)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Titan", TypeLine: "Creature — Giant",
		Power: 9, Toughness: 9, Owner: g.Seats[1].ID, Controller: g.Seats[1].ID,
	})
	before := me.Hand.Size()

	castModal(t, g, "Return of the Wildspeaker", "Instant", returnOfTheWildspeakerOracle, []int{0}, nil)
	passPriorityAroundTable(t, g)

	// The 6/6 Human and the opponent's 9/9 do not count; the 4/4 Bear
	// does. castModal adds the spell to the hand and casting removes
	// it, so the difference is exactly the cards drawn.
	if got := me.Hand.Size() - before; got != 4 {
		t.Errorf("drew %d, want 4", got)
	}
}

func TestReturnOfTheWildspeakerPumpsNonHumansUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear, human := batch01WildspeakerBoard(g, me.ID)

	castModal(t, g, "Return of the Wildspeaker", "Instant", returnOfTheWildspeakerOracle, []int{1}, nil)
	passPriorityAroundTable(t, g)

	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 7 || tough != 7 {
		t.Errorf("Bear = %d/%d, want 7/7", p, tough)
	}
	if p := effectivePower(t, g, human); p != 6 {
		t.Errorf("Human = %d power, want 6 (not pumped)", p)
	}
	advanceToNextSeatsTurn(t, g)
	if p := effectivePower(t, g, bear); p != 4 {
		t.Errorf("Bear after cleanup = %d power, want 4", p)
	}
}

// --- Boros Charm ---------------------------------------------------

func TestBorosCharmDealsFourToAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	before := opp.Life
	castModal(t, g, "Boros Charm", "Instant", borosCharmOracle, []int{0},
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if opp.Life != before-4 {
		t.Errorf("life %d -> %d, want -4", before, opp.Life)
	}
}

func TestBorosCharmDamageModeRefusesACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := seedCreature(g, "Bear", opp.ID)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Boros Charm", TypeLine: "Instant",
		OracleID: borosCharmOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Modes:   []int{0},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Error("a creature was accepted for 'target player or planeswalker'")
	}
}

func TestBorosCharmGrantsDoubleStrikeUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := seedCreature(g, "Bear", me.ID)
	castModal(t, g, "Boros Charm", "Instant", borosCharmOracle, []int{2},
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if !eotHasAbility(effectiveAbilities(t, g, bear), "double strike") {
		t.Fatal("double strike was not granted")
	}
	advanceToNextSeatsTurn(t, g)
	if eotHasAbility(effectiveAbilities(t, g, bear), "double strike") {
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Glorious Anthem", TypeLine: "Enchantment",
		OracleID: gloriousAnthemOracle, Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 1) })
	goblin := findBattlefieldByName(g, "Goblin")
	if goblin == uuid.Nil {
		t.Fatal("no Goblin token")
	}
	if p, tough := effectivePower(t, g, goblin), effectiveToughness(t, g, goblin); p != 2 || tough != 2 {
		t.Errorf("Goblin under an Anthem = %d/%d, want 2/2", p, tough)
	}
	card, _ := battlefieldCard(g, goblin)
	if card.EnteredBattlefieldAt == 0 {
		t.Error("a token must get a battlefield-entry timestamp (CR 613 ordering)")
	}
}

// --- the 2026-09 slice: the cards the September seams unblocked ----
//
// Twenty-four more of #294's hundred, written against the primitives
// that landed after the first pass: conditional alternative costs,
// until-end-of-turn grants, protection as a quality, modal triggers,
// trigger doubling, abilities from the hand, the one-shot land-drop
// grant, layer-4 dependency ordering, and the produced-mana grammar's
// per-colour amounts.

const (
	optOracle                    = "713332c1-5bd8-400f-bfff-c1ca0697a043"
	brainstormOracle             = "36cd2364-d113-47d1-b2c4-b088d9eb88dd"
	growthSpiralOracle           = "34bcc217-dd91-45a0-90d7-a94d02f1f317"
	exploreOracle                = "b8f566ea-8283-4afa-9ac8-737e26419283"
	cropRotationOracle           = "28b46183-c62f-47b1-9fee-3ba148202cab"
	ashBarrensOracle             = "58257464-278e-45fa-8e0b-bcd9a7500bc1"
	flawlessManeuverOracle       = "4e183439-17d2-47ff-9d99-5e22821d91e3"
	akromasWillOracle            = "fd949f82-fc10-4e37-8aa9-6c7569fe3c55"
	jeskasWillOracle             = "0fd114c4-092b-4e28-b0dc-ef529f3bc73e"
	mysticSanctuaryOracle        = "17b60106-a4c7-410a-8ac3-ec8e74e29a7c"
	secludedCourtyardOracle      = "79ba18fd-f184-43c1-86df-56ee18ce806c"
	nykthosOracle                = "84dc18f0-8225-4b40-a165-b10321e41769"
	threeTreeCityOracle          = "da3b17a2-e1e1-44e9-b9b1-ae54a92037db"
	patchworkBannerOracle        = "4fb00dbe-1f82-4ba6-b18c-97e816d10d3a"
	yavimayaOracle               = "8dd5f5af-d2d8-4356-8617-8381081b930c"
	scuteSwarmOracle             = "aa854d50-444c-49d9-bfb1-5476b33c1c0b"
	blackMarketConnectionsOracle = "d2664f28-49e1-46f8-a863-b217e961a57c"
	heraldsHornOracle            = "c02c5547-b9c9-4b2d-9d12-e87bfba8f2d2"
	roamingThroneOracle          = "3640c29b-1534-4952-b297-619ade948431"
	theGreatHengeOracle          = "78427103-9543-41fb-b6d4-72963fe87275"
	boseijuWhoEnduresOracle      = "bf1341dd-41a3-49f6-87ec-63170dde4324"
	otawaraSoaringCityOracle     = "e9b6a394-691c-425a-9307-76d8edc7375e"
	commandBeaconOracle          = "7e8c2a18-e404-40ff-a9e0-ec3eeb6d576e"
	gemstoneCavernsOracle        = "c0adbddc-b070-4c5f-afe0-0474c72a9251"
)

// Every card the slice registered, pinned by oracle ID → name, the
// way TestBatch01CardsAreRegistered pins the first pass's.
func TestBatch01SeptemberSliceCardsAreRegistered(t *testing.T) {
	want := map[string]string{
		optOracle:                    "Opt",
		brainstormOracle:             "Brainstorm",
		growthSpiralOracle:           "Growth Spiral",
		exploreOracle:                "Explore",
		cropRotationOracle:           "Crop Rotation",
		ashBarrensOracle:             "Ash Barrens",
		flawlessManeuverOracle:       "Flawless Maneuver",
		akromasWillOracle:            "Akroma's Will",
		jeskasWillOracle:             "Jeska's Will",
		mysticSanctuaryOracle:        "Mystic Sanctuary",
		secludedCourtyardOracle:      "Secluded Courtyard",
		nykthosOracle:                "Nykthos, Shrine to Nyx",
		threeTreeCityOracle:          "Three Tree City",
		patchworkBannerOracle:        "Patchwork Banner",
		yavimayaOracle:               "Yavimaya, Cradle of Growth",
		scuteSwarmOracle:             "Scute Swarm",
		blackMarketConnectionsOracle: "Black Market Connections",
		heraldsHornOracle:            "Herald's Horn",
		roamingThroneOracle:          "Roaming Throne",
		theGreatHengeOracle:          "The Great Henge",
		boseijuWhoEnduresOracle:      "Boseiju, Who Endures",
		otawaraSoaringCityOracle:     "Otawara, Soaring City",
		commandBeaconOracle:          "Command Beacon",
		gemstoneCavernsOracle:        "Gemstone Caverns",
	}
	if len(want) != 24 {
		t.Fatalf("the slice is 24 cards, the table lists %d", len(want))
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
	}
}

// --- the cantrips --------------------------------------------------

// Opt scries before it draws, which is the whole card: the draw is in
// Scry.Then, so the hand does not grow until the scry is answered.
func TestOptScriesThenDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := me.Hand.Size()
	castCatalogSpell(t, g, "Opt", "Instant", optOracle, nil)
	passPriorityAroundTable(t, g)

	scry := latestChoiceOfKind(g, game.PendingChoiceScry)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := me.Hand.Size()
	castCatalogSpell(t, g, "Brainstorm", "Instant", brainstormOracle, nil)
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != before+3 {
		t.Fatalf("hand after the draw = %d, want %d", got, before+3)
	}
	pick := latestChooseCardsFor(g, me.ID)
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
	order := putInLibraryChoiceFor(g, me.ID)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest",
		Owner: me.ID, Controller: me.ID,
	})
	castCatalogSpell(t, g, "Growth Spiral", "Instant", growthSpiralOracle, nil)
	passPriorityAroundTable(t, g)

	pick := latestChooseCardsFor(g, me.ID)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := me.Hand.Size()
	castCatalogSpell(t, g, "Explore", "Sorcery", exploreOracle, nil)
	passPriorityAroundTable(t, g)

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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	victim := seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	want := pushLibraryCardForTest(me, game.Card{
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
		OracleID: cropRotationOracle, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{SacrificeIDs: []uuid.UUID{victim}}); err != nil {
		t.Fatalf("CastSpell Crop Rotation: %v", err)
	}
	if g.Battlefield.Contains(victim) {
		t.Error("the additional cost must sacrifice the land at announce (CR 601.2f)")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(want) {
		t.Error("the fetched land did not reach the battlefield")
	}
}

// Ash Barrens' basic landcycling is an ability that functions from
// the hand, with a discard-this cost.
func TestAshBarrensBasicLandcyclesFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := seedSearchLibrary(me, searchTestLand("Island", "Basic Land — Island"))
	basic := ids[0]

	id, _ := cycleFromHand(t, g, "Ash Barrens", "Land", ashBarrensOracle, "{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the discard-this cost must put Ash Barrens in the graveyard")
	}
	passPriorityAroundTable(t, g)

	// One basic in the library is fewer matches than the limit, so the
	// search chooser takes it without asking; a prompt appears only
	// when there is a real choice to make.
	if c := searchChoiceFor(g, me.ID); c != nil {
		if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{basic}); err != nil {
			t.Fatalf("ResolveSearchLibrary: %v", err)
		}
	}
	if !me.Hand.Contains(basic) {
		t.Error("basic landcycling did not put the basic into hand")
	}
}

// --- the free spell and the two modal instants ---------------------

// Flawless Maneuver is free while you control a commander, and it
// only touches creatures.
func TestFlawlessManeuverGrantsIndestructibleToCreaturesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sol Ring", TypeLine: "Artifact",
		Owner: me.ID, Controller: me.ID,
	})
	castCatalogSpell(t, g, "Flawless Maneuver", "Instant", flawlessManeuverOracle, nil)
	passPriorityAroundTable(t, g)

	if !hasAbility(effectiveAbilities(t, g, bear), "indestructible") {
		t.Error("the creature did not gain indestructible")
	}
	if hasAbility(effectiveAbilities(t, g, rock), "indestructible") {
		t.Error("\"creatures you control\" must not reach an artifact")
	}
	if cost := game.AlternativeCostByKey(flawlessManeuverOracle, "free"); cost == nil {
		t.Error("the commander free-cast offer is not declared")
	}
}

// Akroma's Will's second bullet grants protection from each colour —
// five separate qualities, not one.
func TestAkromasWillSecondBulletGrantsProtectionFromEachColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	castCatalogSpellWithModes(t, g, "Akroma's Will", "Instant", akromasWillOracle, []int{1})
	passPriorityAroundTable(t, g)

	abilities := effectiveAbilities(t, g, bear)
	for _, want := range []string{"lifelink", "indestructible"} {
		if !hasAbility(abilities, want) {
			t.Errorf("the creature did not gain %s", want)
		}
	}
	for _, c := range game.AllColors {
		if !hasAbility(abilities, game.ProtectionFromColor(c)) {
			t.Errorf("the creature did not gain %s", game.ProtectionFromColor(c))
		}
	}
}

// Akroma's Will asks for ONE bullet: the commander clause that would
// let a player take both is the declared simplification, so a
// two-mode announcement is refused rather than quietly allowed.
func TestAkromasWillRefusesBothModes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Akroma's Will", TypeLine: "Instant",
		OracleID: akromasWillOracle, Owner: me.ID, Controller: me.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0, 1}}); err == nil {
		t.Error("choosing both bullets must be refused while the commander clause is unimplemented")
	}
}

// Jeska's Will's ritual counts the target opponent's hand as it
// resolves.
func TestJeskasWillRitualCountsTheOpponentsHand(t *testing.T) {
	g := newCatalogGame(t)
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
		OracleID: jeskasWillOracle, Owner: me.ID, Controller: me.ID,
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
	passPriorityAroundTable(t, g)

	red := 0
	for _, tok := range me.ManaPool {
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
	g := newCatalogGame(t)
	before := g.Exile.Size()
	castCatalogSpellWithModes(t, g, "Jeska's Will", "Sorcery", jeskasWillOracle, []int{1})
	passPriorityAroundTable(t, g)
	if got := g.Exile.Size() - before; got != 3 {
		t.Errorf("exiled %d cards, want 3", got)
	}
}

// --- the lands -----------------------------------------------------

// Mystic Sanctuary enters untapped on the fourth Island and its
// trigger puts an instant back on top.
func TestMysticSanctuaryEntersUntappedAndRebuysAnInstant(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 3; i++ {
		seedLandOnBattlefield(g, me.ID, "Island", "Basic Land — Island")
	}
	bolt := batch01GraveyardCard(me, "Counterspell", "Instant")

	id := playLandFromHand(t, g, "Mystic Sanctuary", mysticSanctuaryOracle)
	card, ok := battlefieldCard(g, id)
	if !ok || card.Tapped {
		t.Fatal("Mystic Sanctuary must enter untapped with three other Islands")
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	if prompt := latestPickTarget(g, me.ID); prompt == nil {
		t.Fatal("no pick_target prompt after the \"you may\"")
	}
	pickCard(t, g, me.ID, bolt)
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(bolt) {
		t.Error("the instant is still in the graveyard")
	}
	if top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID; top != bolt {
		t.Error("the instant did not go on top of the library")
	}
}

// With only two other Islands it enters tapped and asks nothing.
func TestMysticSanctuaryEntersTappedBelowThreeIslands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 2; i++ {
		seedLandOnBattlefield(g, me.ID, "Island", "Basic Land — Island")
	}
	id := playLandFromHand(t, g, "Mystic Sanctuary", mysticSanctuaryOracle)
	card, ok := battlefieldCard(g, id)
	if !ok || !card.Tapped {
		t.Error("Mystic Sanctuary must enter tapped below three other Islands")
	}
}

// Secluded Courtyard's coloured mana is spendable on a creature
// SPELL of the named type and on an ABILITY of one — the "or" Cavern
// of Souls does not have.
func TestSecludedCourtyardManaWorksForCastsAndActivations(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := pushNamedTribePermanent(t, g, me.ID, "Secluded Courtyard", "Land", secludedCourtyardOracle, "Elf")

	spec, ok := Lookup(secludedCourtyardOracle)
	if !ok || len(spec.ManaAbilities) != 2 {
		t.Fatal("Secluded Courtyard must declare two mana abilities")
	}
	tags := spec.ManaAbilities[1].RestrictionsFunc(g, me.ID, id)
	for _, want := range []string{game.ManaRestrictType("Creature"), game.ManaRestrictSubtype("Elf")} {
		if !containsString(tags, want) {
			t.Errorf("restriction tags %v are missing %q", tags, want)
		}
	}
	if containsString(tags, game.ManaRestrictCast) {
		t.Error("a purpose tag would make the mana casts-only; the card also allows activations")
	}
}

// Nykthos reads devotion per colour into the produced-mana grammar.
func TestNykthosProducesPerColourDevotion(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := seedPermanentWithOracle(g, me.ID, "Nykthos, Shrine to Nyx", "Legendary Land", nykthosOracle)
	for i := 0; i < 2; i++ {
		pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid",
			ManaCost: "{G}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
		})
	}
	spec, ok := Lookup(nykthosOracle)
	if !ok || len(spec.ManaAbilities) != 2 {
		t.Fatal("Nykthos must declare two mana abilities")
	}
	if spec.ManaAbilities[1].Cost.Mana != "{2}" {
		t.Errorf("the devotion ability costs %q, want {2}", spec.ManaAbilities[1].Cost.Mana)
	}
	if got := spec.ManaAbilities[1].ProducedFunc(g, me.ID, id); got != "{G2}" {
		t.Errorf("produced = %q, want {G2} (devotion to green is two)", got)
	}
}

// Three Tree City mints N of ONE colour, N being the creatures of the
// named type.
func TestThreeTreeCityScalesWithTheNamedTribe(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := pushNamedTribePermanent(t, g, me.ID, "Three Tree City", "Legendary Land", threeTreeCityOracle, "Elf")
	for i := 0; i < 3; i++ {
		pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid",
			Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
		})
	}
	spec, _ := Lookup(threeTreeCityOracle)
	if got := spec.ManaAbilities[1].ProducedFunc(g, me.ID, id); got != OneColorOfAmount(3) {
		t.Errorf("produced = %q, want %q", got, OneColorOfAmount(3))
	}
}

// Patchwork Banner pumps the named type and nothing else.
func TestPatchworkBannerPumpsTheNamedType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushNamedTribePermanent(t, g, me.ID, "Patchwork Banner", "Artifact", patchworkBannerOracle, "Elf")
	elf := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	if got := effectivePower(t, g, elf); got != 2 {
		t.Errorf("the Elf is %d power, want 2", got)
	}
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("the Bear is %d power, want 2 — it is not the named type", got)
	}
}

// Yavimaya makes every land a Forest, its controller's and everyone
// else's, and the Forest really taps for {G}.
func TestYavimayaMakesEveryLandAForest(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := seedLand(g, me.ID, "Island", "Basic Land — Island", "")
	theirs := seedLand(g, opp.ID, "Swamp", "Basic Land — Swamp", "")
	yav := seedLand(g, me.ID, "Yavimaya, Cradle of Growth", "Legendary Land", yavimayaOracle)

	for _, tc := range []struct {
		name string
		id   uuid.UUID
	}{{"your Island", mine}, {"an opponent's Swamp", theirs}, {"Yavimaya itself", yav}} {
		if !containsString(effectiveSubtypes(t, g, tc.id), "Forest") {
			t.Errorf("%s is not a Forest", tc.name)
		}
	}
	// "In addition to": the Island keeps its own type.
	if !containsString(effectiveSubtypes(t, g, mine), "Island") {
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
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			seedPermanentWithOracle(g, me.ID, "Scute Swarm", "Creature — Insect", scuteSwarmOracle)
			for i := 0; i < tc.lands; i++ {
				seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
			}
			playLandFromHand(t, g, "Forest", "")
			passPriorityAroundTable(t, g)

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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedPermanentWithOracle(g, me.ID, "Black Market Connections", "Enchantment", blackMarketConnectionsOracle)
	life, hand := me.Life, me.Hand.Size()

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	mode := latestChoiceOfKind(g, game.PendingChoiceModePick)
	if mode == nil {
		t.Fatal("no mode prompt at the first main phase")
	}
	if err := g.ResolveModePick(mode.ID, me.ID, []int{0, 1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := me.Life; got != life-3 {
		t.Errorf("life %d → %d, want %d (1 + 2)", life, got, life-3)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand = %d, want %d", got, hand+1)
	}
	if findBattlefieldByName(g, "Treasure") == uuid.Nil {
		t.Error("no Treasure token")
	}
}

// Herald's Horn discounts creature spells of the named type only, and
// its upkeep offers the top card when it matches.
func TestHeraldsHornDiscountsAndFiltersTheTopCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	horn := pushNamedTribePermanent(t, g, me.ID, "Herald's Horn", "Artifact", heraldsHornOracle, "Elf")

	mods := game.CostModifiersForCard(game.Card{OracleID: heraldsHornOracle})
	if len(mods) != 1 {
		t.Fatalf("%d cost modifiers, want 1", len(mods))
	}
	hornCard, _ := battlefieldCard(g, horn)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	henge := seedPermanentWithOracle(g, me.ID, "The Great Henge", "Legendary Artifact", theGreatHengeOracle)
	life, hand := me.Life, me.Hand.Size()

	if err := g.ActivateManaAbility(me.ID, henge, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := me.Life; got != life+2 {
		t.Errorf("life %d → %d, want %d — the rider gains two", life, got, life+2)
	}
	if len(me.ManaPool) != 2 {
		t.Errorf("mana pool holds %d tokens, want 2", len(me.ManaPool))
	}

	bear := pushCatalogHandCard(me, "Bear", "Creature — Bear", "")
	g.WithWriteLock(func() {
		_, _ = g.PutFromHandOntoBattlefieldForEffect(bear, game.HandEntryOptions{Controller: me.ID})
	})
	passPriorityAroundTable(t, g)
	card, ok := battlefieldCard(g, bear)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for _, p := range []int{2, 5, 3} {
		pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Beater", TypeLine: "Creature — Beast",
			Power: p, Toughness: p, Owner: me.ID, Controller: me.ID,
		})
	}
	spec, _ := Lookup(theGreatHengeOracle)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	throne := pushNamedTribePermanent(t, g, me.ID, "Roaming Throne", "Artifact Creature — Golem", roamingThroneOracle, "Elf")
	if !containsString(effectiveSubtypes(t, g, throne), "Elf") {
		t.Error("the Throne is not the type it named")
	}
	if !containsString(effectiveSubtypes(t, g, throne), "Golem") {
		t.Error("\"in addition to its other types\" must keep Golem")
	}

	doublers := game.CatalogTriggerDoublers(roamingThroneOracle)
	if len(doublers) != 1 {
		t.Fatalf("%d trigger doublers, want 1", len(doublers))
	}
	throneCard, _ := battlefieldCard(g, throne)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	victim := seedLandOnBattlefield(g, opp.ID, "Gaea's Cradle", "Legendary Land")
	pushLibraryCardForTest(opp, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	id := pushCatalogHandCard(me, "Boseiju, Who Endures", "Legendary Land", boseijuWhoEnduresOracle)
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
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the nonbasic land was not destroyed")
	}
}

// Otawara's channel bounces anything, including your own permanent.
func TestOtawaraChannelBouncesItsTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	id := pushCatalogHandCard(me, "Otawara, Soaring City", "Legendary Land", otawaraSoaringCityOracle)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{U}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("channel: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature was not returned to hand")
	}
	if !me.Hand.Contains(bear) {
		t.Error("the creature did not arrive in its owner's hand")
	}
}

// Command Beacon trades itself for the commander in the command zone.
func TestCommandBeaconPutsTheCommanderInHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	beacon := seedPermanentWithOracle(g, me.ID, "Command Beacon", "Land", commandBeaconOracle)
	cmdr := uuid.New()
	me.Command.PushTop(game.Card{
		InstanceID: cmdr, Name: "Kenrith, the Returned King", TypeLine: "Legendary Creature — Human Noble",
		Owner: me.ID, Controller: me.ID, IsCommander: true,
	})
	if err := g.ActivateCatalogAbility(me.ID, beacon, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	// A commander headed for a hand gets the CR 903.9 offer, from the
	// command zone like anywhere else. Declining is what the card is
	// for.
	if offer := latestChoiceOfKind(g, game.PendingChoiceOptionalReplacement); offer != nil {
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := seedPermanentWithOracle(g, me.ID, "Gemstone Caverns", "Legendary Land", gemstoneCavernsOracle)
	spec, _ := Lookup(gemstoneCavernsOracle)
	if got := spec.ManaAbilities[0].ProducedFunc(g, me.ID, id); got != "{C}" {
		t.Errorf("produced without a luck counter = %q, want {C}", got)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, "luck", 1) })
	if got := spec.ManaAbilities[0].ProducedFunc(g, me.ID, id); got != "{W|U|B|R|G}" {
		t.Errorf("produced with a luck counter = %q, want the five-colour pick", got)
	}
}

// --- slice helpers ---------------------------------------------------

// castCatalogSpellWithModes is castCatalogSpell for a modal card: the
// modes are announced with the cast (CR 601.2b).
func castCatalogSpellWithModes(t *testing.T, g *game.Game, name, typeLine, oracleID string, modes []int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracleID,
		Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Modes: modes}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}
