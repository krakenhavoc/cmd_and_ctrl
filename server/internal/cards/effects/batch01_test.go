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
			if spec.ManaAbilities[1].Produced != want || spec.ManaAbilities[1].Rider == nil || !spec.ManaAbilities[1].IgnoreCommanderIdentity {
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

func TestManaDrainRefundsOnlyOnYourOwnMainPhase(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell := batch01OpponentCasts(t, g, opp, "Big Bolt", lightningBoltOracle, "{2}{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})

	castCatalogSpell(t, g, "Mana Drain", "Instant", manaDrainOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	passPriorityAroundTable(t, g)

	if !opp.Graveyard.Contains(spell) {
		t.Fatal("the spell was not countered")
	}
	if n := len(g.DelayedTriggers); n != 1 || !g.DelayedTriggers[0].ControllerTurnOnly {
		t.Fatalf("want one delayed trigger gated to the controller's turn, got %d", n)
	}

	// The opponent's precombat main is NOT "your next main phase".
	batch01AdvanceToStepOf(t, g, 1, game.StepPrecombatMain)
	if len(g.DelayedTriggers) != 1 || len(g.PendingTriggers)+len(g.StackMeta) != 0 {
		t.Fatal("the refund fired on an opponent's main phase")
	}

	// The caster's own is.
	batch01AdvanceToStepOf(t, g, 0, game.StepPrecombatMain)
	if len(g.DelayedTriggers) != 0 {
		t.Fatal("the refund did not fire on the caster's main phase")
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
