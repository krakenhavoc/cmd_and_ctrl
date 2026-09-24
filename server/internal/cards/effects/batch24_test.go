package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch24_test.go — card-level coverage for the card-coverage
// roadmap's batch 24 (#386, `edhrec_rank` 2536–2637): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, attack, land play or step change.
// Helpers from the earlier batch test files are reused by name; new
// ones are b24-prefixed.

const (
	b24MariOracle                = "5f1fdc23-9af0-41a1-aeba-7288f9642734"
	b24DebtToTheDeathlessOracle  = "6130f22d-7901-4f9f-b777-27bb0dacc063"
	b24NaturesWillOracle         = "9837287b-d821-4655-8f01-93fc6c7f0ecc"
	b24AgnaQelaOracle            = "22d0a848-2126-48f0-9050-38daaf93b1d0"
	b24IndulgentAristocratOracle = "c7054a2d-2e7b-4487-8fc3-f6a47a716fd3"
	b24MerchantScrollOracle      = "86cebe2a-95e7-4f22-99cc-e805aeaf347e"
	b24SpineOfIshSahOracle       = "02f062f5-8012-4440-ac12-49fc49822106"
	b24GoblinChieftainAlreadyOID = "368b4052-174e-4458-a6e6-eaf8093aa0fe"
	b24WitchOfTheMoorsOracle     = "7ad97dd9-1342-4ed4-bea9-1bb21d748e04"
	b24BugenhagenOracle          = "4228dd75-1fdd-48dc-9259-b841b2e46c64"
	b24KaervekOracle             = "c7b72d38-0aa0-4e17-9dd5-9276d7cb21ec"
	b24ScourForScrapOracle       = "da422b3e-1092-47ad-97d6-a7a3ef060a53"
	b24AjanisWelcomeOracle       = "4a782bf9-4051-4613-8852-33b0d85a0edd"
	b24MegrimOracle              = "633ad9e2-9f55-4a1c-9248-661ad4b0e1dc"
	b24SceneOfTheCrimeOracle     = "ba11a517-1dbd-4797-9f5e-46ce0f6c77c0"
	b24SurvivorsEncampmentOracle = "152e7e91-4eda-4e72-a9fb-bd5cb2e68239"
	b24SoulsFireOracle           = "62d7ed6e-c386-477e-b155-982c3790f842"
	b24AjanisPridemateOracle     = "95e94dea-5ac0-4d6f-adec-ca147aee861f"
	b24LichKnightsConquestOracle = "b0e533dd-baf2-482a-b9e1-872eb5e47439"
	b24VraskaJoinsUpOracle       = "c91b0dd5-4c63-49a5-95bf-c342b6ff2076"
	b24ChulaneOracle             = "ebf7ce9b-9e5e-4557-9e28-76556997f0ee"
	b24BloodSeekerOracle         = "41087db6-34c4-4e2b-9f54-5e4488ca9c0b"
	b24BurglarRatOracle          = "2f807301-37df-4724-871a-08e3512b07b3"
	b24FinalPartingOracle        = "a5852994-d816-4e62-8a03-254223714544"
	b24WillOfTheAbzanOracle      = "1ae29791-aa7c-4050-bf72-dd0f739b11b8"
	b24MoltenTributaryOracle     = "58c592ed-20fc-481b-909b-2315567e5f20"
	// #1210: both were declared skips for "no activation gate", the
	// gate landed, so they move up into `want`.
	b24LinvalaOracle               = "88dadc31-dfac-41b3-bf2a-65fa89e3c16d"
	b24CursedTotemOracle           = "6225a704-430a-4f56-ad87-0e8d87f285f5"
	b24ThroneOfTheGodPharaohOracle = "ea750169-1f6f-40c2-96e9-55719e103a63"
	b24WrathfulRaptorsOracle       = "1ba92b64-d821-4c24-aed6-fc90d0c63c10"
	b24EfficientConstructionOracle = "5af48f87-7b94-44de-90e3-91f10ced00d3"
	b24TectonicHazardOracle        = "9097243a-39fa-4e18-8316-c0e57699c783"
	b24DwynenOracle                = "30d0d75f-e94c-460b-b957-9f1d655c0f65"
	b24YshtolaNightsBlessedOracle  = "3268251a-8292-44f9-9267-c961b182f739"
)

// b24Lives snapshots every seat's life total.
func b24Lives(g *game.Game) []int {
	out := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		out[i] = p.Life
	}
	return out
}

// b24CastNoncreature puts a noncreature card with a real mana cost in
// the active seat's hand and casts it from a main phase.
func b24CastNoncreature(t *testing.T, g *game.Game, name, typeLine, manaCost string) uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := handCardFull(me, name, typeLine, manaCost, "", nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Molten
// Tributary is a row in the Dominaria United dual table, so a
// transposed row is invisible until someone plays that exact card.
// Goblin Chieftain was already on main from the S26 tribal lords and
// is pinned here so the table matches the issue's 40.
func TestBatch24CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b24MariOracle:                  "Mari, the Killing Quill",
		b24DebtToTheDeathlessOracle:    "Debt to the Deathless",
		b24NaturesWillOracle:           "Nature's Will",
		b24AgnaQelaOracle:              "Agna Qel'a",
		b24IndulgentAristocratOracle:   "Indulgent Aristocrat",
		b24MerchantScrollOracle:        "Merchant Scroll",
		b24SpineOfIshSahOracle:         "Spine of Ish Sah",
		b24GoblinChieftainAlreadyOID:   "Goblin Chieftain",
		b24WitchOfTheMoorsOracle:       "Witch of the Moors",
		b24BugenhagenOracle:            "Bugenhagen, Wise Elder",
		b24KaervekOracle:               "Kaervek the Merciless",
		b24ScourForScrapOracle:         "Scour for Scrap",
		b24AjanisWelcomeOracle:         "Ajani's Welcome",
		b24MegrimOracle:                "Megrim",
		b24SceneOfTheCrimeOracle:       "Scene of the Crime",
		b24SurvivorsEncampmentOracle:   "Survivors' Encampment",
		b24SoulsFireOracle:             "Soul's Fire",
		b24AjanisPridemateOracle:       "Ajani's Pridemate",
		b24LichKnightsConquestOracle:   "Lich-Knights' Conquest",
		b24VraskaJoinsUpOracle:         "Vraska Joins Up",
		b24ChulaneOracle:               "Chulane, Teller of Tales",
		b24BloodSeekerOracle:           "Blood Seeker",
		b24BurglarRatOracle:            "Burglar Rat",
		b24FinalPartingOracle:          "Final Parting",
		b24WillOfTheAbzanOracle:        "Will of the Abzan",
		b24MoltenTributaryOracle:       "Molten Tributary",
		b24ThroneOfTheGodPharaohOracle: "Throne of the God-Pharaoh",
		b24WrathfulRaptorsOracle:       "Wrathful Raptors",
		b24EfficientConstructionOracle: "Efficient Construction",
		b24TectonicHazardOracle:        "Tectonic Hazard",
		b24DwynenOracle:                "Dwynen, Gilt-Leaf Daen",
		b24YshtolaNightsBlessedOracle:  "Y'shtola, Night's Blessed",
		b24LinvalaOracle:               "Linvala, Keeper of Silence",
		b24CursedTotemOracle:           "Cursed Totem",
	}
	if len(want) != 34 {
		t.Fatalf("the batch registers 33 cards plus Goblin Chieftain, the table lists %d", len(want))
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
	// The six remaining declared skips must NOT be registered —
	// each needs a trigger mode, a cost or a gate the engine cannot
	// express, and a spec would ship the card stronger than printed.
	// Linvala and Cursed Totem came off this list with #1210's
	// activation gate and are in `want` above.
	for _, skipped := range []string{
		"3f8e5ff1-af89-427e-924c-19a44f9a3788", // Retreat to Kazandu — modal trigger
		"1690a782-9e54-4231-9a72-abae5ed7050c", // Dread Presence — modal trigger
		"5d8acafe-c13e-43ca-a89e-0cfd12d95662", // Titan of Industry — modal ETB, two targeted modes
		"451e8ece-7389-4d52-8bc9-450cb5e53e6a", // Ascend from Avernus — a spell exiling itself as it resolves
		"10c31317-71e8-42e0-85e0-3e64bd0c3dd3", // Vat of Rebirth — remove-counters cost
		"be6155de-c5b2-415c-ad83-142f9926462a", // Khalni Heart Expedition — remove-counters cost
	} {
		if _, ok := Lookup(skipped); ok {
			t.Errorf("%s is a declared skip and must not be registered", skipped)
		}
	}
}

func TestB24SurvivorsEncampmentPaysItsCreatureTapCost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := playLandFromHand(t, g, "Survivors' Encampment", b24SurvivorsEncampmentOracle)
	bear := pushVanillaCreature(g, me.ID, "Fresh Bear", 2, 2)

	if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{TapIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if !b20Tapped(t, g, land) || !b20Tapped(t, g, bear) {
		t.Error("the land and the chosen creature both pay the ability")
	}
	pick := riderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("mana choice = %+v, want five colors", pick)
	}
}

// --- the lands -----------------------------------------------------

func TestB24MoltenTributaryEntersTappedAndTapsForBlueOrRed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := playLandFromHand(t, g, "Molten Tributary", b24MoltenTributaryOracle)
	if !b20Tapped(t, g, land) {
		t.Fatal("enters tapped")
	}
	b22Untap(g, land)
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := riderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 2 || pick.ColorOptions[0] != "U" || pick.ColorOptions[1] != "R" {
		t.Fatalf("{U} or {R}, got %+v", pick)
	}
}

func TestB24AgnaQelaEntersTappedWithoutABasicAndLoots(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	first := playLandFromHand(t, g, "Agna Qel'a", b24AgnaQelaOracle)
	if !b20Tapped(t, g, first) {
		t.Fatal("with no basic land it enters tapped")
	}
	seedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
	advanceToPrecombatMainOf(t, g, 0)
	second := playLandFromHand(t, g, "Agna Qel'a", b24AgnaQelaOracle)
	if b20Tapped(t, g, second) {
		t.Fatal("with a basic land it enters untapped")
	}
	if err := g.ActivateManaAbility(me.ID, second, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "U" {
		t.Errorf("pool %v, want [U]", got)
	}
	me.ManaPool.EmptyPool()
	b22Untap(g, second)
	b06AddMana(me, "U", "C", "C")
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, second, 0, game.ActivateAbilityParams{})
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("drew %d, want 1 before the discard", got-hand)
	}
	if discardOwed(g, me.ID) != 1 {
		t.Errorf("then discards a card: owed %d", discardOwed(g, me.ID))
	}
	if !b20Tapped(t, g, second) {
		t.Error("the loot has a tap cost")
	}
}

func TestB24SceneOfTheCrimeEntersTappedTapsForManaAndCracks(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	scene := playLandFromHand(t, g, "Scene of the Crime", b24SceneOfTheCrimeOracle)
	if !b20Tapped(t, g, scene) {
		t.Fatal("enters tapped")
	}
	b22Untap(g, scene)
	if err := g.ActivateManaAbility(me.ID, scene, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
	me.ManaPool.EmptyPool()
	b22Untap(g, scene)
	bear := pushVanillaCreature(g, me.ID, "Fresh Bear", 2, 2)
	if err := g.ActivateManaAbility(me.ID, scene, 1, game.ManaAbilityParams{TapIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("ActivateManaAbility colored: %v", err)
	}
	if !b20Tapped(t, g, scene) || !b20Tapped(t, g, bear) {
		t.Error("the land and the chosen creature both pay the colored ability")
	}
	pick := riderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("colored ability choice = %+v, want five colors", pick)
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "G"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool %v, want [G]", got)
	}
	me.ManaPool.EmptyPool()
	// The crack needs no tap: a tapped Scene can still be sacrificed.
	b06AddMana(me, "C", "C")
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, scene, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(scene) {
		t.Error("sacrificed as the cost")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("drew %d, want 1", got-hand)
	}
	if spec, _ := Lookup(b24SceneOfTheCrimeOracle); spec.Completeness != CompletenessFull || len(spec.ManaAbilities) != 2 {
		t.Error("Scene should be full with both printed mana abilities")
	}
}

// --- the spells ----------------------------------------------------

func TestB24DebtToTheDeathlessDrainsTwiceX(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := b24Lives(g)
	castXSpell(t, g, "Debt to the Deathless", "Sorcery", b24DebtToTheDeathlessOracle, "{X}{W}{W}{B}{B}", 3, nil)
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		want := before[i] - 6
		if i == 0 {
			want = before[0] + 18
		}
		if p.Life != want {
			t.Errorf("seat %d: %d → %d, want %d", i, before[i], p.Life, want)
		}
	}
	_ = me
}

func TestB24TectonicHazardPingsEachOpponentAndTheirCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Elf", "Creature — Elf", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	theirBear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	before := b24Lives(g)
	castCatalogSpell(t, g, "Tectonic Hazard", "Sorcery", b24TectonicHazardOracle, nil)
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		want := before[i] - 1
		if i == 0 {
			want = before[0]
		}
		if p.Life != want {
			t.Errorf("seat %d: %d → %d, want %d", i, before[i], p.Life, want)
		}
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("an opponent's 1/1 dies")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("your own creatures are untouched")
	}
	if damageMarkedOn(g, theirBear) != 1 {
		t.Errorf("their 2/2 took 1, got %d", damageMarkedOn(g, theirBear))
	}
}

func TestB24MerchantScrollFindsABlueInstant(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedSearchLibrary(me,
		game.Card{Name: "Counterspell", TypeLine: "Instant", ManaCost: "{U}{U}", Colors: []string{"U"}},
		game.Card{Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", Colors: []string{"R"}},
		game.Card{Name: "Ponder", TypeLine: "Sorcery", ManaCost: "{U}", Colors: []string{"U"}},
	)
	castCatalogSpell(t, g, "Merchant Scroll", "Sorcery", b24MerchantScrollOracle, nil)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[0]) {
		t.Error("the one blue instant is found without a prompt")
	}
	if me.Hand.Contains(ids[1]) || me.Hand.Contains(ids[2]) {
		t.Error("a red instant and a blue sorcery are not blue instants")
	}
}

func TestB24FinalPartingTutorsOneToHandAndOneToGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedSearchLibrary(me,
		game.Card{Name: "A", TypeLine: "Instant"},
		game.Card{Name: "B", TypeLine: "Creature — Bear"},
		game.Card{Name: "C", TypeLine: "Land"},
	)
	castCatalogSpell(t, g, "Final Parting", "Sorcery", b24FinalPartingOracle, nil)
	passPriorityAroundTable(t, g)
	first := searchChoiceFor(g, me.ID)
	if first == nil || len(first.SearchCards) != 3 || first.SearchMax != 1 {
		t.Fatalf("the first search offers the whole library for one card, got %+v", first)
	}
	answerSearchByID(t, g, me.ID, ids[0])
	if !me.Hand.Contains(ids[0]) {
		t.Fatal("the first pick goes to hand")
	}
	second := searchChoiceFor(g, me.ID)
	if second == nil || len(second.SearchCards) != 2 || second.SearchMax != 1 {
		t.Fatalf("the second search offers the rest for one card, got %+v", second)
	}
	answerSearchByID(t, g, me.ID, ids[1])
	if !me.Graveyard.Contains(ids[1]) {
		t.Error("the second pick goes to the graveyard")
	}
	if me.Library.Size() != 1 || !me.Library.Contains(ids[2]) {
		t.Error("the third card stays in the library")
	}
	if searchChoiceFor(g, me.ID) != nil {
		t.Error("two cards, two prompts, no third")
	}
}

func TestB24ScourForScrapTutorsAndReturnsAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lib := seedSearchLibrary(me,
		game.Card{Name: "Sol Ring", TypeLine: "Artifact"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	dead := b17GraveyardCard(me, "Dead Signet", "Artifact", "{2}")
	deadBear := b17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	if err := b22TryModal(g, b24ScourForScrapOracle, []int{1}, b16TargetCard(deadBear)); err == nil {
		t.Fatal("a creature card is not an artifact card")
	}
	castModal(t, g, "Scour for Scrap", "Instant", b24ScourForScrapOracle, []int{0, 1}, b16TargetCard(dead))
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(lib[0]) {
		t.Error("the one artifact in the library is found")
	}
	if me.Hand.Contains(lib[1]) {
		t.Error("a Bear is not an artifact card")
	}
	if !me.Hand.Contains(dead) {
		t.Error("the targeted artifact card returns to hand")
	}
}

func TestB24SoulsFireBitesWithACreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Beast", "Creature — Beast", 4, 4)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	wall := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 6)
	castCatalogSpell(t, g, "Soul's Fire", "Instant", b24SoulsFireOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("4 damage from the Beast kills the Bear")
	}
	// #764: the first slot is its own CLAUSE, so an opponent's
	// creature there is refused at ANNOUNCE (CR 601.2c) rather than
	// discovered at resolution. Back the clause list out to one wide
	// spec and this cast is accepted and quietly does nothing.
	if err := castCatalogSpellErr(t, g, "Soul's Fire", "Instant", b24SoulsFireOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: wall}, {Kind: game.TargetCard, ID: mine}}); err != game.ErrIllegalTarget {
		t.Errorf("their Wall in slot 0: err %v, want ErrIllegalTarget", err)
	}
	if damageMarkedOn(g, mine) != 0 {
		t.Error("a refused announcement deals nothing")
	}
	// A player in the second slot: face damage equal to power.
	life := opp.Life
	castCatalogSpell(t, g, "Soul's Fire", "Instant", b24SoulsFireOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("life %d → %d, want -4", life, opp.Life)
	}
}

func TestB24LichKnightsConquestSacrificesThatManyAndReanimates(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	deadA := b17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	deadB := b17GraveyardCard(me, "Dead Golem", "Artifact Creature — Golem", "{4}")
	deadC := b17GraveyardCard(me, "Dead Wurm", "Creature — Wurm", "{5}{G}")
	treasure := pushToken(g, me.ID, TreasureToken())
	signet := b12Permanent(g, me.ID, "Signet", "Artifact")
	aura := b12Permanent(g, me.ID, "Rancor", "Enchantment — Aura")
	bear := b12Creature(g, me.ID, "Living Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Lich-Knights' Conquest", "Sorcery", b24LichKnightsConquestOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: deadA}, {Kind: game.TargetCard, ID: deadB}})
	passPriorityAroundTable(t, g)
	// #1019: "that many" is how many were SACRIFICED, so nothing comes
	// back while the prompts are open.
	if g.Battlefield.Contains(deadA) || g.Battlefield.Contains(deadB) {
		t.Fatal("the creature cards came back before a single Treasure had been chosen")
	}
	// Two sacrifice prompts, each over the artifacts, enchantments
	// and tokens — never the nontoken creature, and never the
	// artifact creature that is going to come back.
	prompts := 0
	for _, c := range g.PendingChoices {
		if c == nil || c.Kind != game.PendingChoiceSacrifice || c.Chooser != me.ID {
			continue
		}
		prompts++
		if hasID(c.SacrificeOptions, bear) || hasID(c.SacrificeOptions, deadB) {
			t.Error("a nontoken creature that is not an artifact is not a legal sacrifice, nor is the returned Golem")
		}
		if !hasID(c.SacrificeOptions, treasure) || !hasID(c.SacrificeOptions, signet) || !hasID(c.SacrificeOptions, aura) {
			t.Error("a token, an artifact and an enchantment are all legal")
		}
	}
	if prompts != 2 {
		t.Fatalf("two cards returned, two sacrifices owed: %d prompts", prompts)
	}
	answerSacrifice(t, g, me.ID, treasure)
	if g.Battlefield.Contains(deadA) || g.Battlefield.Contains(deadB) {
		t.Error("the run waits for BOTH prompts before it returns anything")
	}
	answerSacrifice(t, g, me.ID, aura)
	if g.Battlefield.Contains(treasure) || g.Battlefield.Contains(aura) || !g.Battlefield.Contains(signet) {
		t.Error("exactly the two chosen permanents are sacrificed")
	}
	if !g.Battlefield.Contains(deadA) || !g.Battlefield.Contains(deadB) {
		t.Fatal("both chosen creature cards return once the sacrifices have landed")
	}
	if g.Battlefield.Contains(deadC) {
		t.Error("only the chosen cards return")
	}
	if sacrificeChoiceFor(g, me.ID) != nil {
		t.Error("no third prompt")
	}
	// Capped by what can be sacrificed: one eligible permanent (the
	// returned Golem is an artifact, so it dies first), one creature
	// back.
	b18Kill(t, g, deadB)
	deadD := b17GraveyardCard(me, "Dead Drake", "Creature — Drake", "{2}{U}")
	advanceToPrecombatMainOf(t, g, 0)
	castCatalogSpell(t, g, "Lich-Knights' Conquest", "Sorcery", b24LichKnightsConquestOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: deadC}, {Kind: game.TargetCard, ID: deadD}})
	passPriorityAroundTable(t, g)
	answerSacrifice(t, g, me.ID, signet)
	if !g.Battlefield.Contains(deadC) || g.Battlefield.Contains(deadD) {
		t.Error("one eligible permanent: the first chosen card returns, the second does not")
	}
	if spec, _ := Lookup(b24LichKnightsConquestOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the choice order is a declared gap")
	}
}

func TestB24WillOfTheAbzanEdictsTheBiggestOrReanimates(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	aBig := b12Creature(g, a.ID, "A's Giant", "Creature — Giant", 5, 5)
	aSmall := b12Creature(g, a.ID, "A's Bear", "Creature — Bear", 2, 2)
	bOne := b12Creature(g, b.ID, "B's Bear", "Creature — Bear", 2, 2)
	bTwo := b12Creature(g, b.ID, "B's Other Bear", "Creature — Bear", 2, 2)
	dead := b17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	if err := b22TryModal(g, b24WillOfTheAbzanOracle, []int{0, 1},
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: a.ID}, {Kind: game.TargetCard, ID: dead}}); err == nil {
		t.Fatal("choose both is the declared gap — one mode only")
	}
	aLife, bLife := a.Life, b.Life
	castModal(t, g, "Will of the Abzan", "Sorcery", b24WillOfTheAbzanOracle, []int{0},
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: a.ID}, {Kind: game.TargetPlayer, ID: b.ID}})
	passPriorityAroundTable(t, g)
	// #1019: the loss is a clause about a PLAYER, so it happens either
	// way — but it happens AFTER the sacrifice it is printed after.
	if a.Life != aLife || b.Life != bLife {
		t.Errorf("the life loss landed while the sacrifice prompts were open: %d→%d, %d→%d",
			aLife, a.Life, bLife, b.Life)
	}
	pa := sacrificeChoiceFor(g, a.ID)
	if pa == nil || len(pa.SacrificeOptions) != 1 || pa.SacrificeOptions[0] != aBig {
		t.Fatalf("A must sacrifice the Giant, the only greatest-power creature: %+v", pa)
	}
	pb := sacrificeChoiceFor(g, b.ID)
	if pb == nil || len(pb.SacrificeOptions) != 2 {
		t.Fatalf("B's two Bears tie for greatest and B chooses: %+v", pb)
	}
	answerSacrifice(t, g, a.ID, aBig)
	answerSacrifice(t, g, b.ID, bTwo)
	if g.Battlefield.Contains(aBig) || !g.Battlefield.Contains(aSmall) || !g.Battlefield.Contains(bOne) || g.Battlefield.Contains(bTwo) {
		t.Error("exactly the chosen creatures are sacrificed")
	}
	if a.Life != aLife-3 || b.Life != bLife-3 {
		t.Errorf("each targeted opponent loses 3 once they have answered: %d→%d, %d→%d",
			aLife, a.Life, bLife, b.Life)
	}
	// Mode two: reanimate.
	advanceToPrecombatMainOf(t, g, 0)
	castModal(t, g, "Will of the Abzan", "Sorcery", b24WillOfTheAbzanOracle, []int{1}, b16TargetCard(dead))
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) {
		t.Error("the creature card returns to the battlefield")
	}
	if c, _ := battlefieldCard(g, dead); c.Controller != me.ID {
		t.Error("under its owner's — your — control")
	}
}

// --- the triggers --------------------------------------------------

func TestB24NaturesWillTapsTheirLandsUntapsYoursOncePerPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	will := b12Push(g, me.ID, "Nature's Will", "Enchantment", b24NaturesWillOracle, 0, 0)
	myLand := seedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
	b08Tap(g, myLand)
	aLand := seedLand(g, a.ID, "Island", "Basic Land — Island", "")
	bLand := seedLand(g, b.ID, "Swamp", "Basic Land — Swamp", "")
	bearA := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	bearB := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	bearC := b12Creature(g, me.ID, "Bear C", "Creature — Bear", 2, 2)
	// One declaration split across two defenders (CR 508.1): the
	// lock-in announces all three together (#859).
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, d := range []struct {
		attacker uuid.UUID
		defender uuid.UUID
	}{{bearA, a.ID}, {bearB, a.ID}, {bearC, b.ID}} {
		if err := g.DeclareAttacker(d.attacker, d.defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	advanceTo(t, g, game.StepCombatDamage)
	if n := len(g.PendingTriggers) + triggersOnStackFrom(g, will); n != 2 {
		t.Fatalf("two bears on A and one on B: one trigger per player, got %d", n)
	}
	// Two differing triggers: their controller orders them (CR 603.3b).
	passPriorityUntilTriggerOrderPrompt(t, g)
	answerTriggerOrderInOfferedOrder(t, g)
	passPriorityAroundTable(t, g)
	if !b20Tapped(t, g, aLand) || !b20Tapped(t, g, bLand) {
		t.Error("both damaged players' lands are tapped")
	}
	if b20Tapped(t, g, myLand) {
		t.Error("your lands untap")
	}
}

func TestB24ThroneOfTheGodPharaohDrainsForTappedCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Throne of the God-Pharaoh", "Legendary Artifact", b24ThroneOfTheGodPharaohOracle, 0, 0)
	tappedA := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	tappedB := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Untapped Bear", "Creature — Bear", 2, 2)
	b08Tap(g, tappedA)
	b08Tap(g, tappedB)
	before := b24Lives(g)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		want := before[i] - 2
		if i == 0 {
			want = before[0]
		}
		if p.Life != want {
			t.Errorf("seat %d: %d → %d, want %d", i, before[i], p.Life, want)
		}
	}
	// An opponent's end step is not yours.
	before = b24Lives(g)
	advanceToEndStepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.Life != before[i] {
			t.Errorf("seat %d changed on an opponent's end step", i)
		}
	}
}

func TestB24AjanisWelcomeAndPridemateGrowTogether(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Ajani's Welcome", "Enchantment", b24AjanisWelcomeOracle, 0, 0)
	pridemate := b12Push(g, me.ID, "Ajani's Pridemate", "Creature — Cat Soldier", b24AjanisPridemateOracle, 2, 2)
	life := me.Life
	castAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Fatalf("a creature entering under your control gains 1: %d → %d", life, me.Life)
	}
	if got := counterCount(g, pridemate, "+1/+1"); got != 1 {
		t.Errorf("the gain grows the Pridemate: %d counters, want 1", got)
	}
	// An opponent's creature entering is not yours.
	b13PlayAs(t, g, 1, "Their Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	if me.Life != life+1 || counterCount(g, pridemate, "+1/+1") != 1 {
		t.Error("an opponent's creature gains nothing")
	}
	_ = opp
}

func TestB24BurglarRatMakesOpponentsDiscardAndMegrimPunishes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Megrim", "Enchantment", b24MegrimOracle, 0, 0)
	castAndResolveCreature(t, g, "Burglar Rat", "Creature — Rat", b24BurglarRatOracle)
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		owed := discardOwed(g, p.ID)
		if i == 0 && owed != 0 {
			t.Errorf("you owe nothing, got %d", owed)
		}
		if i != 0 && owed != 1 {
			t.Errorf("seat %d owes one discard, got %d", i, owed)
		}
	}
	life := opp.Life
	answerDiscard(t, g, opp.ID, opp.Hand.Cards[0].InstanceID)
	// Every opponent was asked; the table is gated until all three
	// have answered, so Megrim's trigger cannot reach the stack until
	// then (#651).
	for _, p := range g.Seats[2:] {
		discardFromHand(t, g, p.ID)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("Megrim deals 2 to the discarding opponent: %d → %d", life, opp.Life)
	}
	// Your own discard is not an opponent's — and the cleanup-step
	// hand-size discard is still DiscardSelection's, which is now the
	// only thing that map carries (#651).
	g.DiscardPending = map[uuid.UUID]int{me.ID: 1}
	mine := me.Hand.Cards[0].InstanceID
	myLife := me.Life
	if err := g.DiscardSelection(me.ID, []uuid.UUID{mine}); err != nil {
		t.Fatalf("DiscardSelection: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != myLife {
		t.Error("Megrim does not punish its controller")
	}
}

func TestB24EfficientConstructionMakesAThopterPerArtifactSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Efficient Construction", "Enchantment", b24EfficientConstructionOracle, 0, 0)
	b24CastNoncreature(t, g, "Sol Ring", "Artifact", "{1}")
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Thopter"); n != 1 {
		t.Fatalf("an artifact spell makes a Thopter: %d", n)
	}
	b24CastNoncreature(t, g, "Ponder", "Sorcery", "{U}")
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Thopter"); n != 1 {
		t.Errorf("a sorcery is not an artifact spell: %d", n)
	}
	thopter := findBattlefieldByName(g, "Thopter")
	if !hasAbility(effectiveAbilities(t, g, thopter), "flying") {
		t.Error("the Thopter flies")
	}
}

func TestB24BloodSeekerMayDrainTheEnteringCreaturesController(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Blood Seeker", "Creature — Vampire Shaman", b24BloodSeekerOracle, 1, 1)
	life := opp.Life
	b13PlayAs(t, g, 1, "Their Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	if latestTriggerPromptFor(g, me.ID) == nil {
		t.Fatal("the Seeker's controller is asked")
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("that player loses 1: %d → %d", life, opp.Life)
	}
	// Declining costs nothing; your own creature asks nothing.
	b13PlayAs(t, g, 1, "Their Second Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Error("no is no")
	}
	advanceToPrecombatMainOf(t, g, 0)
	castAndResolveCreature(t, g, "My Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	if latestTriggerPromptFor(g, me.ID) != nil {
		t.Error("your own creature entering is not an opponent's")
	}
}

func TestB24KaervekPunishesEveryOpponentsSpellForItsManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kaervek := b12Push(g, me.ID, "Kaervek the Merciless", "Legendary Creature — Human Shaman", b24KaervekOracle, 5, 4)
	theirs := b12Creature(g, opp.ID, "Their Golem", "Creature — Golem", 4, 4)
	batch01OpponentCasts(t, g, opp, "Big Spell", "", "{2}{U}{U}", nil)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if p == nil || p.Source != kaervek {
		t.Fatal("Kaervek's controller picks any target")
	}
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("four damage from a four-drop kills the 4/4")
	}
	// Your own spells are not an opponent's.
	advanceToPrecombatMainOf(t, g, 0)
	b24CastNoncreature(t, g, "My Spell", "Sorcery", "{2}{U}{U}")
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("your own spell does not trigger Kaervek")
	}
}

func TestB24YshtolaDrawsOnBigLifeLossAndPunishesBigSpells(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	yshtola := b12Push(g, me.ID, "Y'shtola, Night's Blessed", "Legendary Creature — Cat Warlock", b24YshtolaNightsBlessedOracle, 2, 4)
	if !hasAbility(effectiveAbilities(t, g, yshtola), "vigilance") {
		t.Error("printed vigilance")
	}
	// A two-drop is not mana value 3 or greater.
	before := b24Lives(g)
	b24CastNoncreature(t, g, "Small Spell", "Instant", "{1}{U}")
	passPriorityAroundTable(t, g)
	if lives := b24Lives(g); lives[1] != before[1] || lives[0] != before[0] {
		t.Error("mana value 2 does nothing")
	}
	b24CastNoncreature(t, g, "Big Spell", "Sorcery", "{1}{U}{B}")
	passPriorityAroundTable(t, g)
	lives := b24Lives(g)
	for i := 1; i < len(lives); i++ {
		if lives[i] != before[i]-2 {
			t.Errorf("seat %d takes 2: %d → %d", i, before[i], lives[i])
		}
	}
	if lives[0] != before[0]+2 {
		t.Errorf("you gain 2 once: %d → %d", before[0], lives[0])
	}
	// A creature spell is not a noncreature spell.
	castAndResolveCreature(t, g, "Big Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	if opp.Life != lives[1] {
		t.Error("a creature spell does not trigger")
	}
	// End step: an opponent lost only 2 this turn — no draw.
	hand := me.Hand.Size()
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatal("2 life lost is not 4")
	}
	// Next turn (an opponent's): Bolt an opponent twice — 6 lost —
	// and the end step draws for Y'shtola's controller.
	advanceToPrecombatMainOf(t, g, 1)
	b24CastNoncreature(t, g, "Filler", "Instant", "{U}")
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -4) })
	hand = me.Hand.Size()
	advanceToEndStepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("a player lost 4 this turn — each end step, any player's — you draw: %d", got-hand)
	}
}

func TestB24BugenhagenDrawsForASevenPowerCreatureAndTapsForAnyColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bugenhagen := b12Push(g, me.ID, "Bugenhagen, Wise Elder", "Legendary Creature — Human Shaman", b24BugenhagenOracle, 1, 3)
	if !hasAbility(effectiveAbilities(t, g, bugenhagen), "reach") {
		t.Error("printed reach")
	}
	if err := g.ActivateManaAbility(me.ID, bugenhagen, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if pick := riderLatestManaPick(g, me.ID); pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("any colour, the printed width, got %+v", pick)
	}
	b10ResolveAllManaPicks(t, g, me.ID, "G")
	big := b12Creature(g, me.ID, "Big Beast", "Creature — Beast", 6, 6)
	// The walk to the next upkeep passes a cleanup step, so the hand is
	// emptied first to keep the max-hand-size discard out of the count.
	emptyHandToLibrary(g, me)
	advanceToPrecombatMainOf(t, g, 1)
	advanceToUpkeepOf(t, g, 0)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatal("a 6-power creature is not 7 or greater")
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(big, "+1/+1", 1) })
	emptyHandToLibrary(g, me)
	advanceToPrecombatMainOf(t, g, 1)
	advanceToUpkeepOf(t, g, 0)
	hand = me.Hand.Size()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("a 7-power creature (counters count) draws: %d", got-hand)
	}
}

func TestB24WitchOfTheMoorsEdictsAndReturnsWhenYouGainedLife(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	witch := b12Push(g, me.ID, "Witch of the Moors", "Creature — Human Warlock", b24WitchOfTheMoorsOracle, 4, 4)
	if !hasAbility(effectiveAbilities(t, g, witch), "deathtouch") {
		t.Error("printed deathtouch")
	}
	aBear := b12Creature(g, a.ID, "A's Bear", "Creature — Bear", 2, 2)
	b12Creature(g, b.ID, "B's Bear", "Creature — Bear", 2, 2)
	dead := b17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	// No life gained: no trigger.
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if sacrificeChoiceFor(g, a.ID) != nil || latestPickTarget(g, me.ID) != nil {
		t.Fatal("without lifegain the Witch is quiet")
	}
	// Gained life, a creature card in the graveyard: the targeted
	// twin fires.
	advanceToPrecombatMainOf(t, g, 0)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1) })
	advanceToEndStepOf(t, g, 0)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if p.PickTargetMin != 0 || !hasID(p.PickTargetCards, dead) {
		t.Fatalf("up to one creature card from your graveyard: %+v", p)
	}
	pickCard(t, g, me.ID, dead)
	passPriorityAroundTable(t, g)
	if sacrificeChoiceFor(g, a.ID) == nil || sacrificeChoiceFor(g, b.ID) == nil || sacrificeChoiceFor(g, me.ID) != nil {
		t.Fatal("each opponent — not you — sacrifices a creature of their choice")
	}
	if !me.Hand.Contains(dead) {
		t.Error("the chosen creature card returns to hand")
	}
	answerSacrifice(t, g, a.ID, aBear)
	answerSacrifice(t, g, b.ID, findBattlefieldByName(g, "B's Bear"))
	// Gained life, no creature card: the untargeted twin still
	// edicts.
	b12Creature(g, a.ID, "A's Second Bear", "Creature — Bear", 2, 2)
	advanceToPrecombatMainOf(t, g, 0)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1) })
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("nothing to target: no prompt")
	}
	if sacrificeChoiceFor(g, a.ID) == nil {
		t.Error("with an empty graveyard the edict still comes")
	}
}

func TestB24SpineOfIshSahDestroysAndComesBack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Permanent(g, opp.ID, "Their Signet", "Artifact")
	spine := castAndResolveCreature(t, g, "Spine of Ish Sah", "Artifact", b24SpineOfIshSahOracle)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if !hasID(p.PickTargetCards, theirs) || !hasID(p.PickTargetCards, spine) {
		t.Fatal("any permanent, the Spine itself included")
	}
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Fatal("the target is destroyed")
	}
	b18Kill(t, g, spine)
	if !me.Hand.Contains(spine) {
		t.Error("put into a graveyard from the battlefield: back to its owner's hand")
	}
	// Exiled, not died: nothing.
	advanceToPrecombatMainOf(t, g, 0)
	spine = castAndResolveCreature(t, g, "Spine of Ish Sah", "Artifact", b24SpineOfIshSahOracle)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, spine)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(spine) {
		t.Error("destroying itself still returns it")
	}
}

func TestB24IndulgentAristocratGrowsTheVampires(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	aristocrat := b12Push(g, me.ID, "Indulgent Aristocrat", "Creature — Vampire Noble", b24IndulgentAristocratOracle, 1, 1)
	if !hasAbility(effectiveAbilities(t, g, aristocrat), "lifelink") {
		t.Error("printed lifelink")
	}
	vampire := b12Creature(g, me.ID, "Other Vampire", "Creature — Vampire", 2, 2)
	human := b12Creature(g, me.ID, "Human", "Creature — Human", 2, 2)
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	advanceToMain(t, g)
	b06AddMana(me, "C", "C")
	b16Activate(t, g, me.ID, aristocrat, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}})
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature is sacrificed as the cost")
	}
	if counterCount(g, aristocrat, "+1/+1") != 1 || counterCount(g, vampire, "+1/+1") != 1 {
		t.Error("each Vampire you control — the Aristocrat included — gets a counter")
	}
	if counterCount(g, human, "+1/+1") != 0 {
		t.Error("a Human is not a Vampire")
	}
}

func TestB24ChulaneDrawsOnCreatureSpellsAndBouncesYourCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	chulane := b12Push(g, me.ID, "Chulane, Teller of Tales", "Legendary Creature — Human Druid", b24ChulaneOracle, 2, 4)
	if !hasAbility(effectiveAbilities(t, g, chulane), "vigilance") {
		t.Error("printed vigilance")
	}
	hand := me.Hand.Size()
	bear := castAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Fatalf("a creature spell draws: %d", got-hand)
	}
	hand = me.Hand.Size()
	b24CastNoncreature(t, g, "Ponder", "Sorcery", "{U}")
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Error("a noncreature spell draws nothing")
	}
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b06AddMana(me, "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, chulane, 0, game.ActivateAbilityParams{Targets: b16TargetCard(theirs)}); err == nil {
		t.Fatal("target creature YOU control")
	}
	b16Activate(t, g, me.ID, chulane, 0, game.ActivateAbilityParams{Targets: b16TargetCard(bear)})
	if !me.Hand.Contains(bear) {
		t.Error("the creature returns to its owner's hand")
	}
	if !b20Tapped(t, g, chulane) {
		t.Error("the bounce has a tap cost")
	}
	if spec, _ := Lookup(b24ChulaneOracle); spec.Completeness != CompletenessFull {
		t.Error("the land drop landed with #654; nothing is deferred any more")
	}
}

func TestB24DwynenPumpsElvesAndGainsPerAttackingElf(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dwynen := b12Push(g, me.ID, "Dwynen, Gilt-Leaf Daen", "Legendary Creature — Elf Warrior", b24DwynenOracle, 3, 4)
	if !hasAbility(effectiveAbilities(t, g, dwynen), "reach") {
		t.Error("printed reach")
	}
	elf := b12Creature(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirElf := b12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	if effectivePower(t, g, elf) != 2 || effectivePower(t, g, dwynen) != 3 || effectivePower(t, g, bear) != 2 || effectivePower(t, g, theirElf) != 1 {
		t.Error("other Elves YOU control get +1/+1")
	}
	life := me.Life
	declareAttack(t, g, opp.ID, dwynen, elf, bear)
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("Dwynen and one Elf attacking: gain 2, got %d → %d", life, me.Life)
	}
}

func TestB24VraskaJoinsUpGivesDeathtouchCountersAndDrawsOffLegends(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	legend := b21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	vraska := castAndResolveCreature(t, g, "Vraska Joins Up", "Legendary Enchantment", b24VraskaJoinsUpOracle)
	passPriorityAroundTable(t, g)
	if counterCount(g, bear, "deathtouch") != 1 || counterCount(g, legend, "deathtouch") != 1 {
		t.Fatal("each creature you control gets a deathtouch counter")
	}
	if counterCount(g, theirs, "deathtouch") != 0 {
		t.Error("an opponent's creature gets none")
	}
	if !hasAbility(effectiveAbilities(t, g, bear), "deathtouch") {
		t.Error("a deathtouch counter grants deathtouch")
	}
	hand := me.Hand.Size()
	attackWith(t, g, opp.ID, bear, legend)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("the legendary creature connecting draws one — the Bear none: %d", got-hand)
	}
	// The gap: Vraska leaving turns the counters inert.
	b18Kill(t, g, vraska)
	if hasAbility(effectiveAbilities(t, g, bear), "deathtouch") {
		t.Error("declared gap: the counter stops granting once Vraska has left")
	}
	if spec, _ := Lookup(b24VraskaJoinsUpOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the counter gap is declared")
	}
}

func TestB24WrathfulRaptorsReflectsDamageToADinosaur(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	raptors := b12Push(g, me.ID, "Wrathful Raptors", "Creature — Dinosaur", b24WrathfulRaptorsOracle, 5, 5)
	if !hasAbility(effectiveAbilities(t, g, raptors), "trample") {
		t.Error("printed trample")
	}
	dino := b12Creature(g, me.ID, "My Dinosaur", "Creature — Dinosaur", 3, 3)
	theirDino := b12Creature(g, opp.ID, "Their Dinosaur", "Creature — Dinosaur", 3, 3)
	theirBear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b12Bolt(t, g, game.TargetRef{Kind: game.TargetCard, ID: dino})
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if p == nil || p.Source != raptors {
		t.Fatal("the Raptors' controller picks the target")
	}
	if hasID(p.PickTargetCards, theirDino) || hasID(p.PickTargetCards, raptors) {
		t.Error("a Dinosaur — anyone's — is not a legal target")
	}
	if !hasID(p.PickTargetCards, theirBear) || !hasID(p.PickTargetPlayers, opp.ID) {
		t.Error("a non-Dinosaur creature and a player are")
	}
	pickCard(t, g, me.ID, theirBear)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirBear) {
		t.Error("the damaged Dinosaur deals 3 to the Bear")
	}
	// An opponent's Dinosaur being damaged is not yours.
	advanceToPrecombatMainOf(t, g, 0)
	b12Bolt(t, g, game.TargetRef{Kind: game.TargetCard, ID: theirDino})
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a Dinosaur you don't control does not trigger")
	}
}

func TestB24MariExilesWithHitCountersAndBountiesThem(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mari := b12Push(g, me.ID, "Mari, the Killing Quill", "Legendary Creature — Vampire Assassin", b24MariOracle, 3, 2)
	rogue := b12Creature(g, me.ID, "Rogue", "Creature — Human Rogue", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if !hasAbility(effectiveAbilities(t, g, mari), "deathtouch") || !hasAbility(effectiveAbilities(t, g, rogue), "deathtouch") {
		t.Error("Assassins and Rogues you control have deathtouch — Mari included")
	}
	if hasAbility(effectiveAbilities(t, g, bear), "deathtouch") {
		t.Error("a Bear is none of the three")
	}
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "My Fodder", "Creature — Bear", 1, 1)
	b18Kill(t, g, theirs)
	if !inExile(g, theirs) || b12Counter(t, g, theirs, "hit") != 1 {
		t.Fatal("an opponent's creature dying is exiled with a hit counter")
	}
	b18Kill(t, g, mine)
	if !me.Graveyard.Contains(mine) {
		t.Error("your own creature dying stays in the graveyard")
	}
	// The Rogue connects: yes to the bounty.
	hand := me.Hand.Size()
	attackWith(t, g, opp.ID, rogue, bear)
	if n := b06TriggerPromptsFor(g, me.ID); n != 1 {
		t.Fatalf("the Rogue's hit asks; the Bear's does not: %d prompts", n)
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if b12Counter(t, g, theirs, "hit") != 0 {
		t.Error("the hit counter is removed")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("draw a card: %d", got-hand)
	}
	if n := countBattlefieldNamed(g, me.ID, "Treasure"); n != 2 {
		t.Errorf("two Treasures: %d", n)
	}
	// No hit-countered card that player owns: yes does nothing.
	advanceToPrecombatMainOf(t, g, 0)
	hand = me.Hand.Size()
	attackWith(t, g, opp.ID, rogue)
	prompt := latestTriggerPromptFor(g, me.ID)
	if prompt == nil || !prompt.NoLegalTarget {
		t.Fatal("the prompt warns there is nothing to remove")
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || countBattlefieldNamed(g, me.ID, "Treasure") != 2 {
		t.Error("if you do — and you could not")
	}
}
