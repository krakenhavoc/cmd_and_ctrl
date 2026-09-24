package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch15_test.go — card-level coverage for the card-coverage
// roadmap's batch 15 (#308, `edhrec_rank` 1626–1726): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, land play or step change.
// Helpers from the earlier batch test files are reused by name; new
// ones are b15-prefixed.

const (
	b15WhirlwindOfThoughtOracle    = "6467cbb7-1e4e-482d-a20f-6cb9fc0f1ad1"
	b15StitchTogetherOracle        = "bc3d5911-3580-4132-9daf-2826495b5739"
	b15SoulOfTheHarvestOracle      = "6b9194dd-2296-4879-ac5e-f89b18431df6"
	b15ObscuraStorefrontOracle     = "dc31a6f8-6228-4a25-b937-5d8d78514333"
	b15WindsOfRathOracle           = "a6fd90dc-0ec3-4dce-a77a-4d04f5e254bf"
	b15PoisonTipArcherOracle       = "d0e810bb-5f38-4045-a718-30d423c05659"
	b15BloodMoneyOracle            = "75f5d372-4ff9-430c-8302-72472439e0d2"
	b15SilentClearingOracle        = "fd45063f-c83c-431c-9104-f139c497ec0d"
	b15InspiringOverseerOracle     = "d646e42b-5635-4798-b633-29c093b66a55"
	b15BotanicalSanctumOracle      = "88f8f683-738e-48f3-afff-c8f73f1033a2"
	b15RunawaySteamKinOracle       = "eb414a8f-566f-4635-80a1-77b2e18ac8dc"
	b15MoonsilverKeyOracle         = "373395b2-f04b-46c9-b9f4-4f6815b71009"
	b15TheGafferOracle             = "a177295e-3b58-4e46-a1cb-fc003a7a0848"
	b15RunAwayTogetherOracle       = "290faa28-450e-4797-9a8f-642d8af3f82a"
	b15UtterEndOracle              = "cba94732-1f1f-4ffd-9aba-45db990043fa"
	b15SmugglersShareOracle        = "17b29350-4f37-4552-8192-4856b15345f9"
	b15SunscorchRegentOracle       = "c8ff464f-0059-4055-a2f2-556fe6db8fbf"
	b15HorizonCanopyOracle         = "262a5d83-506c-4781-9bc9-1a2b5d83955c"
	b15HornetQueenOracle           = "3b1f8108-6911-49e9-8f78-f950bb58cb6c"
	b15OversoldCemeteryOracle      = "ed4cd3a1-688b-4c05-948d-39d3336e00c0"
	b15MirrorBoxOracle             = "3bed1944-58dc-4679-9aee-7be4d94fb55c"
	b15KherKeepOracle              = "79638767-fbc7-451a-b29f-d93f2ac6f102"
	b15GalaGreetersOracle          = "cce081eb-8820-415a-a7b2-3c5b9d4a2601"
	b15FateUnravelerOracle         = "7d66f67d-3148-4636-ac5a-2cf8a51d5e50"
	b15ThermoAlchemistOracle       = "228fbae1-423e-461d-b8c3-55786938a3cb"
	b15SimulacrumSynthesizerOracle = "eb7a1f21-a66d-415b-8520-710b44890bb6"
	b15TheLocustGodOracle          = "e025a714-02da-4b0c-8021-cf3e8dc9b19e"

	b15SolRingOracle = "6ad8011d-3471-4369-9d68-b264cc027487"
)

// b15CastInstant puts an instant with the given colours in the
// active seat's hand and casts it from a main phase, leaving it on
// the stack.
func b15CastInstant(t *testing.T, g *game.Game, name string, colors ...string) uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := handCardFull(me, name, "Instant", "", "", colors)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// b15Destroy destroys a battlefield permanent through the effect API
// and settles whatever it triggered.
func b15Destroy(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
	passPriorityAroundTable(t, g)
}

// b15Aura attaches a fixture Aura to a battlefield creature.
func b15Aura(g *game.Game, owner, host uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Pacifism", TypeLine: "Enchantment — Aura",
		Owner: owner, Controller: owner,
		AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: host},
	})
}

// b15Lives reads every opponent's life total for a seat-0 controller.
func b15Lives(g *game.Game) []int {
	out := make([]int, 0, len(g.Seats)-1)
	for _, p := range g.Seats[1:] {
		out = append(out, p.Life)
	}
	return out
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Botanical
// Sanctum is a row in fastlands.go, so a transposed row is invisible
// until someone plays that exact card.
func TestBatch15CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b15WhirlwindOfThoughtOracle:    "Whirlwind of Thought",
		b15StitchTogetherOracle:        "Stitch Together",
		b15SoulOfTheHarvestOracle:      "Soul of the Harvest",
		b15ObscuraStorefrontOracle:     "Obscura Storefront",
		b15WindsOfRathOracle:           "Winds of Rath",
		b15PoisonTipArcherOracle:       "Poison-Tip Archer",
		b15BloodMoneyOracle:            "Blood Money",
		b15SilentClearingOracle:        "Silent Clearing",
		b15InspiringOverseerOracle:     "Inspiring Overseer",
		b15BotanicalSanctumOracle:      "Botanical Sanctum",
		b15RunawaySteamKinOracle:       "Runaway Steam-Kin",
		b15MoonsilverKeyOracle:         "Moonsilver Key",
		b15TheGafferOracle:             "The Gaffer",
		b15RunAwayTogetherOracle:       "Run Away Together",
		b15UtterEndOracle:              "Utter End",
		b15SmugglersShareOracle:        "Smuggler's Share",
		b15SunscorchRegentOracle:       "Sunscorch Regent",
		b15HorizonCanopyOracle:         "Horizon Canopy",
		b15HornetQueenOracle:           "Hornet Queen",
		b15OversoldCemeteryOracle:      "Oversold Cemetery",
		b15MirrorBoxOracle:             "Mirror Box",
		b15KherKeepOracle:              "Kher Keep",
		b15GalaGreetersOracle:          "Gala Greeters",
		b15FateUnravelerOracle:         "Fate Unraveler",
		b15ThermoAlchemistOracle:       "Thermo-Alchemist",
		b15SimulacrumSynthesizerOracle: "Simulacrum Synthesizer",
		b15TheLocustGodOracle:          "The Locust God",
	}
	if len(want) != 27 {
		t.Fatalf("the batch registers 27 cards, the table lists %d", len(want))
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

// --- the lands -----------------------------------------------------

func TestB15BotanicalSanctumIsTheTenthFastland(t *testing.T) {
	spec, ok := Lookup(b15BotanicalSanctumOracle)
	if !ok {
		t.Fatal("Botanical Sanctum not registered")
	}
	if len(spec.Replacements) != 1 || len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{G|U}" {
		t.Errorf("Botanical Sanctum: want one enters-tapped-unless replacement and one {G|U} ability, got %d / %+v", len(spec.Replacements), spec.ManaAbilities)
	}
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	seedLandOnBattlefield(g, me.ID, "Island", "Basic Land — Island")
	early := b12PlayFromHand(t, g, "Botanical Sanctum", "Land", b15BotanicalSanctumOracle, game.CastSpellParams{})
	if b13Tapped(t, g, early) {
		t.Error("with two other lands Botanical Sanctum enters untapped")
	}
	late := b12PlayFromHand(t, g, "Botanical Sanctum", "Land", b15BotanicalSanctumOracle, game.CastSpellParams{})
	if !b13Tapped(t, g, late) {
		t.Error("with three other lands Botanical Sanctum enters tapped")
	}
}

func TestB15CanopyLandsPayLifeForManaAndCashInForACard(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, colour string
	}{
		{"Silent Clearing", b15SilentClearingOracle, "B"},
		{"Horizon Canopy", b15HorizonCanopyOracle, "G"},
	} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		land := b12Push(g, me.ID, tc.name, "Land", tc.oracle, 0, 0)
		before := me.Life
		if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("%s: ActivateManaAbility: %v", tc.name, err)
		}
		if me.Life != before-1 {
			t.Errorf("%s: life %d → %d, want -1", tc.name, before, me.Life)
		}
		if n := b10ResolveAllManaPicks(t, g, me.ID, tc.colour); n != 1 {
			t.Fatalf("%s: expected one colour pick, answered %d", tc.name, n)
		}
		if got := batch01PoolColors(me); len(got) != 1 || got[0] != tc.colour {
			t.Errorf("%s: pool %v, want [%s]", tc.name, got, tc.colour)
		}
		b08Untap(g, land)
		advanceToMain(t, g)
		me.ManaPool.AddMana(game.ManaToken{Color: "C"})
		hand := me.Hand.Size()
		if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("%s: ActivateCatalogAbility: %v", tc.name, err)
		}
		passPriorityAroundTable(t, g)
		if me.Hand.Size() != hand+1 {
			t.Errorf("%s: hand %d → %d, want +1", tc.name, hand, me.Hand.Size())
		}
		if g.Battlefield.Contains(land) {
			t.Errorf("%s: the land was sacrificed", tc.name)
		}
	}
}

func TestB15ObscuraStorefrontFetchesAnEsperBasicTappedAndGainsOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedSearchLibrary(me,
		searchTestLand("Plains", "Basic Land — Plains"),
		searchTestLand("Island", "Basic Land — Island"),
		searchTestLand("Swamp", "Basic Land — Swamp"),
		searchTestLand("Forest", "Basic Land — Forest"),
	)
	before := me.Life
	storefront := playLandFromHand(t, g, "Obscura Storefront", b15ObscuraStorefrontOracle)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(storefront) || !me.Graveyard.Contains(storefront) {
		t.Fatal("the Storefront should have sacrificed itself")
	}
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the search should ask which basic to take")
	}
	if searchOptionNamed(g, c, "Forest") != uuid.Nil {
		t.Error("a Forest is not a Plains, Island or Swamp")
	}
	answerSearchNamed(t, g, me.ID, "Island")
	island := findBattlefieldByName(g, "Island")
	if card, ok := battlefieldCard(g, island); !ok || !card.Tapped {
		t.Error("the fetched basic enters tapped")
	}
	if me.Life != before+1 {
		t.Errorf("life %d → %d, want +1", before, me.Life)
	}
}

func TestB15KherKeepMakesANamedKobold(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	keep := b12Push(g, me.ID, "Kher Keep", "Legendary Land", b15KherKeepOracle, 0, 0)
	if err := g.ActivateManaAbility(me.ID, keep, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for colorless: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
	b08Untap(g, keep)
	advanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "R"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, keep, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	if !b13Tapped(t, g, keep) {
		t.Error("the token ability taps the Keep")
	}
	passPriorityAroundTable(t, g)
	kobold := findBattlefieldByName(g, "Kobolds of Kher Keep")
	if kobold == uuid.Nil {
		t.Fatal("no Kobolds of Kher Keep token")
	}
	if p, tough := effectivePower(t, g, kobold), effectiveToughness(t, g, kobold); p != 0 || tough != 1 {
		t.Errorf("the Kobold is 0/1, got %d/%d", p, tough)
	}
}

// --- cast triggers -------------------------------------------------

func TestB15WhirlwindOfThoughtDrawsOffNoncreatureSpellsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Whirlwind of Thought", "Enchantment", b15WhirlwindOfThoughtOracle, 0, 0)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Shock", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("an instant: hand %d → %d, want +1 (the draw, net of the cast)", hand, me.Hand.Size())
	}
	hand = me.Hand.Size()
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("a creature spell must not draw: hand %d → %d", hand, me.Hand.Size())
	}
}

func TestB15SunscorchRegentGrowsOffOpposingSpells(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	regent := b12Push(g, me.ID, "Sunscorch Regent", "Creature — Dragon", b15SunscorchRegentOracle, 4, 3)
	life := me.Life
	b13OpponentCasts(t, g, opp, "Shock", "Instant", "", "{R}", nil)
	passPriorityAroundTable(t, g)
	if n := counterCount(g, regent, "+1/+1"); n != 1 {
		t.Errorf("an opponent's spell: %d counters, want 1", n)
	}
	if me.Life != life+1 {
		t.Errorf("life %d → %d, want +1", life, me.Life)
	}
	castCatalogSpell(t, g, "Shock", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if n := counterCount(g, regent, "+1/+1"); n != 1 {
		t.Errorf("the controller's own spell must not trigger it: %d counters", n)
	}
}

func TestB15ThermoAlchemistPingsTheTableAndUntapsOnInstants(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	alchemist := b12Push(g, me.ID, "Thermo-Alchemist", "Creature — Human Shaman", b15ThermoAlchemistOracle, 0, 3)
	before := b15Lives(g)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, alchemist, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("opponent %d: %d → %d, want -1", i+1, b, got)
		}
	}
	if !b13Tapped(t, g, alchemist) {
		t.Fatal("the ping taps the Alchemist")
	}
	castCatalogSpell(t, g, "Shock", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if b13Tapped(t, g, alchemist) {
		t.Error("an instant cast untaps the Alchemist")
	}
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	if err := g.ActivateCatalogAbility(me.ID, alchemist, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("second activation: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !b13Tapped(t, g, alchemist) {
		t.Error("a creature spell must not untap it")
	}
	if summoningSickOf(t, g, alchemist) {
		t.Error("a pushed fixture is not summoning sick")
	}
}

func TestB15RunawaySteamKinChargesOnRedSpellsAndCashesInForRRR(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	kin := b12Push(g, me.ID, "Runaway Steam-Kin", "Creature — Elemental", b15RunawaySteamKinOracle, 1, 1)
	for i := 1; i <= 2; i++ {
		b15CastInstant(t, g, "Lightning Bolt", "R")
		passPriorityAroundTable(t, g)
		if n := counterCount(g, kin, "+1/+1"); n != i {
			t.Fatalf("after %d red spells: %d counters, want %d", i, n, i)
		}
	}
	if err := g.ActivateManaAbility(me.ID, kin, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("two counters must not be enough to activate")
	}
	if n := counterCount(g, kin, "+1/+1"); n != 2 {
		t.Fatalf("a refused activation must not touch the counters: %d", n)
	}
	b15CastInstant(t, g, "Counterspell", "U")
	passPriorityAroundTable(t, g)
	if n := counterCount(g, kin, "+1/+1"); n != 2 {
		t.Errorf("a blue spell must not charge it: %d counters", n)
	}
	b15CastInstant(t, g, "Boros Charm", "R", "W")
	passPriorityAroundTable(t, g)
	b15CastInstant(t, g, "Lightning Bolt", "R")
	passPriorityAroundTable(t, g)
	if n := counterCount(g, kin, "+1/+1"); n != 3 {
		t.Fatalf("the intervening-if caps it at three: %d counters", n)
	}
	if err := g.ActivateManaAbility(me.ID, kin, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("with three counters the ability activates: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 3 || got[0] != "R" || got[1] != "R" || got[2] != "R" {
		t.Errorf("pool %v, want [R R R]", got)
	}
	if n := counterCount(g, kin, "+1/+1"); n != 0 {
		t.Errorf("the three counters are the cost: %d left", n)
	}
	if b13Tapped(t, g, kin) {
		t.Error("the ability has no tap cost")
	}
}

// --- entry triggers ------------------------------------------------

func TestB15InspiringOverseerGainsALifeAndDrawsOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life, hand := me.Life, me.Hand.Size()
	overseer := castCatalogSpell(t, g, "Inspiring Overseer", "Creature — Angel Cleric", b15InspiringOverseerOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d → %d, want +1", life, me.Life)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want +1 (the draw, net of the cast)", hand, me.Hand.Size())
	}
	if !hasEffectiveKeyword(t, g, overseer, "flying") {
		t.Error("printed flying did not reach the effective abilities")
	}
}

func TestB15HornetQueenBringsFourDeathtouchFlyers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	queen := castCatalogSpell(t, g, "Hornet Queen", "Creature — Insect", b15HornetQueenOracle, nil)
	passPriorityAroundTable(t, g)
	insects := battlefieldIDsNamed(g, "Insect")
	if len(insects) != 4 {
		t.Fatalf("%d Insects, want 4", len(insects))
	}
	for _, id := range insects {
		if !hasEffectiveKeyword(t, g, id, "flying") || !hasEffectiveKeyword(t, g, id, "deathtouch") {
			t.Error("each Insect has flying and deathtouch")
		}
		if controllerOf(t, g, id) != me.ID {
			t.Error("the Insects are the controller's")
		}
	}
	if !hasEffectiveKeyword(t, g, queen, "deathtouch") {
		t.Error("the Queen's own deathtouch did not reach the effective abilities")
	}
}

func TestB15SoulOfTheHarvestOffersADrawPerNontokenCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Soul of the Harvest", "Creature — Elemental", b15SoulOfTheHarvestOracle, 6, 6)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if latestTriggerPrompt(g, me.ID) == nil {
		t.Fatal("a nontoken creature entering should ask about the draw")
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want +1", hand, me.Hand.Size())
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 1) })
	passPriorityAroundTable(t, g)
	if latestTriggerPrompt(g, me.ID) != nil {
		t.Error("a token must not trigger it")
	}
}

// #764: Gala Greeters is a MODAL TRIGGER. The mode is chosen as the
// ability goes on the stack (CR 603.3c) through a mode_pick prompt —
// not at resolution, and not by the engine picking for the player.
// This is the untargeted half of the modal-trigger proof set; the
// targeted half is Glissa Sunslayer.
func TestB15GalaGreetersPromptsForItsModeAsItGoesOnTheStack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	greeters := b12Push(g, me.ID, "Gala Greeters", "Creature — Elf Druid", b15GalaGreetersOracle, 1, 1)
	life := me.Life

	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)

	// The prompt is open and the ability is NOT on the stack yet: the
	// mode is chosen as it is put there, so there is nothing to
	// respond to until the controller has answered.
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the alliance trigger asks for its mode")
	}
	if triggerOnStack(g, greeters) != nil {
		t.Error("CR 603.3c: the mode is chosen as the ability is put on the stack, not after")
	}
	if len(c.ModeOptionIndex) != 3 || c.ModeMin != 1 || c.ModeMax != 1 {
		t.Fatalf("three bullets, choose one: %+v", c)
	}
	if c.ModeOptionLabel[1] != "Create a tapped Treasure token." {
		t.Errorf("the bullets ride the prompt verbatim: %q", c.ModeOptionLabel[1])
	}

	// Take the Treasure first — the old shape could only ever give
	// the counter to the first creature of the turn.
	if err := g.ResolveModePick(c.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	if triggerOnStack(g, greeters) == nil {
		t.Fatal("answering puts the ability on the stack")
	}
	passPriorityAroundTable(t, g)
	treasure := findBattlefieldByName(g, "Treasure")
	if treasure == uuid.Nil || !b13Tapped(t, g, treasure) {
		t.Fatal("the chosen mode resolved: a tapped Treasure")
	}
	if counterCount(g, greeters, "+1/+1") != 0 || me.Life != life {
		t.Error("only the chosen bullet happens (CR 608.2c)")
	}

	// A second creature, and the controller takes a different bullet.
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	c = modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the second trigger asks again")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := counterCount(g, greeters, "+1/+1"); n != 1 {
		t.Errorf("the +1/+1 bullet: %d counters, want 1", n)
	}
}

// modePickChoiceFor is the open mode_pick prompt for a chooser, or
// nil.
func modePickChoiceFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoiceModePick && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

func TestB15SimulacrumSynthesizerScriesAndBuildsConstructsItSizes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	synth := castCatalogSpell(t, g, "Simulacrum Synthesizer", "Artifact", b15SimulacrumSynthesizerOracle, nil)
	passPriorityAroundTable(t, g)
	c := scryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("the entry scries 2, got %+v", c)
	}
	if err := g.ResolveScry(c.ID, me.ID, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	castWithCost(t, g, "Signet", "Artifact", "{2}", "")
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Construct") != uuid.Nil {
		t.Fatal("a mana value 2 artifact makes nothing")
	}
	castWithCost(t, g, "Big Rock", "Artifact", "{3}", "")
	passPriorityAroundTable(t, g)
	construct := findBattlefieldByName(g, "Construct")
	if construct == uuid.Nil {
		t.Fatal("a mana value 3 artifact makes a Construct")
	}
	// Synthesizer, Signet, Big Rock and the Construct itself.
	if p := effectivePower(t, g, construct); p != 4 {
		t.Errorf("the Construct is +1/+1 per artifact: power %d, want 4", p)
	}
	other := b12Push(g, me.ID, "Simulacrum Synthesizer", "Artifact", b15SimulacrumSynthesizerOracle, 0, 0)
	if p := effectivePower(t, g, construct); p != 5 {
		t.Errorf("a second Synthesizer is one more artifact, not a second sizing: power %d, want 5", p)
	}
	b15Destroy(t, g, synth)
	if p := effectivePower(t, g, construct); p != 4 {
		t.Errorf("the surviving Synthesizer takes over the sizing: power %d, want 4", p)
	}
	b15Destroy(t, g, other)
	if p := effectivePower(t, g, construct); p != 0 {
		t.Errorf("with no Synthesizer the Construct shrinks to 0/0 — the declared gap: power %d", p)
	}
	// The engine treats a printed 0/0 that never had a counter as a
	// placeholder and never sweeps it (CurrentToughness's convention),
	// so the shrunken Construct lingers rather than dying. Pinned so a
	// change in that convention shows up here.
	runStateChecksViaDraw(t, g)
	if !g.Battlefield.Contains(construct) {
		t.Error("a printed 0/0 that never had a counter is not swept by the toughness SBA")
	}
}

// #683: a Construct that has had counters and lost them all is no
// placeholder (Card.LostLastCounter), so when the last Synthesizer
// leaves and it shrinks to 0/0 it dies — the caveat's second clause.
// A Construct that never had a counter still lingers beside it.
func TestB15SynthesizerConstructThatLostItsCountersDiesWhenItShrinks(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	synth := b12Push(g, me.ID, "Simulacrum Synthesizer", "Artifact", b15SimulacrumSynthesizerOracle, 0, 0)
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, TokenCard("0/0 colorless Construct artifact"), 2)
	})
	var constructs []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Construct" {
			constructs = append(constructs, c.InstanceID)
		}
	}
	if len(constructs) != 2 {
		t.Fatalf("want 2 Constructs, got %d", len(constructs))
	}
	spent, fresh := constructs[0], constructs[1]

	gainAndLoseACounter(t, g, spent)
	if !g.Battlefield.Contains(spent) {
		t.Fatal("while the Synthesizer sizes it, a Construct survives losing its last counter")
	}
	if got := effectiveToughness(t, g, spent); got != 3 {
		t.Errorf("Synthesizer and two Constructs: toughness %d, want 3", got)
	}

	b15Destroy(t, g, synth)
	runStateChecksViaDraw(t, g)
	if g.Battlefield.Contains(spent) {
		t.Error("a Construct that lost its counters survived shrinking to 0/0")
	}
	if !g.Battlefield.Contains(fresh) {
		t.Error("a Construct that never had a counter should linger as a 0/0")
	}
	if spec, _ := Lookup(b15SimulacrumSynthesizerOracle); len(spec.Caveats) != 1 || !strings.Contains(spec.Caveats[0], "lost them all dies") {
		t.Errorf("the caveat must tell players the spent Construct dies: %v", spec.Caveats)
	}
}

// --- draw and death triggers ---------------------------------------

func TestB15FateUnravelerPingsAnOpponentPerCardDrawn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Fate Unraveler", "Enchantment Creature — Hag", b15FateUnravelerOracle, 3, 4)
	before := opp.Life
	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 2) })
	passPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("two draws: opponent %d → %d, want -2", before, opp.Life)
	}
	mine := me.Life
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	passPriorityAroundTable(t, g)
	if me.Life != mine {
		t.Errorf("the controller's own draw must not trigger it: %d → %d", mine, me.Life)
	}
}

func TestB15PoisonTipArcherDrainsTheTableWhenAnotherCreatureDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	archer := b12Push(g, me.ID, "Poison-Tip Archer", "Creature — Elf Archer", b15PoisonTipArcherOracle, 2, 3)
	if !hasEffectiveKeyword(t, g, archer, "reach") || !hasEffectiveKeyword(t, g, archer, "deathtouch") {
		t.Error("printed reach and deathtouch did not reach the effective abilities")
	}
	before := b15Lives(g)
	b15Destroy(t, g, seedCreature(g, "Their Bear", opp.ID))
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("an opponent's creature died: opponent %d %d → %d, want -1", i+1, b, got)
		}
	}
	b15Destroy(t, g, seedCreature(g, "My Bear", me.ID))
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-2 {
			t.Errorf("the controller's own creature counts too: opponent %d %d → %d, want -2", i+1, b, got)
		}
	}
	b15Destroy(t, g, archer)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-2 {
			t.Errorf("the Archer's own death is not \"another\": opponent %d %d → %d", i+1, b, got)
		}
	}
	if me.Life != 40 {
		t.Errorf("the controller loses nothing: %d", me.Life)
	}
}

func TestB15TheLocustGodSwarmsLootsAndComesBack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	god := b12Push(g, me.ID, "The Locust God", "Legendary Creature — God", b15TheLocustGodOracle, 4, 4)
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 2) })
	passPriorityAroundTable(t, g)
	insects := battlefieldIDsNamed(g, "Insect")
	if len(insects) != 2 {
		t.Fatalf("two draws: %d Insects, want 2", len(insects))
	}
	if !hasEffectiveKeyword(t, g, insects[0], "flying") || !hasEffectiveKeyword(t, g, insects[0], "haste") {
		t.Error("the Insects have flying and haste")
	}
	advanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "U"}, game.ManaToken{Color: "R"})
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, god, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("loot: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || discardOwed(g, me.ID) != 1 {
		t.Errorf("draw first, then owe a discard: hand %d → %d, pending %d", hand, me.Hand.Size(), discardOwed(g, me.ID))
	}
	// The looted draw's trigger waits behind the discard prompt — the
	// table does not move on while one is open (#651) — so it reaches
	// the stack only once the discard is paid.
	answerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
	passPriorityAroundTable(t, g)
	if n := len(battlefieldIDsNamed(g, "Insect")); n != 3 {
		t.Errorf("the looted draw is a draw: %d Insects, want 3", n)
	}
	b15Destroy(t, g, god)
	if !me.Graveyard.Contains(god) {
		t.Fatal("the God should be in the graveyard until the end step")
	}
	advanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(god) {
		t.Error("at the next end step the God returns to hand")
	}
}

// --- end-step and upkeep triggers ----------------------------------

func TestB15TheGafferDrawsAtAnyEndStepAfterThreeLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "The Gaffer", "Legendary Creature — Halfling Peasant", b15TheGafferOracle, 2, 3)
	advanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 2) })
	hand := me.Hand.Size()
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("two life is not three: hand %d → %d", hand, me.Hand.Size())
	}
	// An opponent's turn: two separate gains add up, and "each end
	// step" includes theirs.
	advanceToMainOf(t, g, 1)
	g.WithWriteLock(func() {
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 2)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -5)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1)
	})
	hand = me.Hand.Size()
	advanceToEndStepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("three life gained (a loss does not offset it): hand %d → %d, want +1", hand, me.Hand.Size())
	}
}

func TestB15SmugglersSharePaysForGreedyAndRampingOpponents(t *testing.T) {
	g := newCatalogGame(t)
	me, greedy, ramper, quiet := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	b12Push(g, me.ID, "Smuggler's Share", "Enchantment", b15SmugglersShareOracle, 0, 0)
	advanceToMain(t, g)
	forest := game.Card{Name: "Forest", TypeLine: "Token Land — Forest"}
	g.WithWriteLock(func() {
		_ = g.DrawNForEffect(greedy.ID, 2)
		_ = g.DrawNForEffect(quiet.ID, 1)
		_ = g.CreateTokenForEffect(ramper.ID, forest, 2)
		_ = g.CreateTokenForEffect(quiet.ID, forest, 1)
		_ = g.CreateTokenForEffect(me.ID, forest, 2)
	})
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("one opponent drew two: hand %d → %d, want +1", hand, me.Hand.Size())
	}
	if n := countBattlefieldNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("one opponent had two lands enter (my own do not count): %d Treasures, want 1", n)
	}
	// The next end step starts from zero.
	hand = me.Hand.Size()
	advanceToEndStepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || countBattlefieldNamed(g, me.ID, "Treasure") != 1 {
		t.Error("the tallies reset with the turn")
	}
}

func TestB15OversoldCemeteryReturnsACreatureCardWithFourInTheYard(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	b12Push(g, owner.ID, "Oversold Cemetery", "Enchantment", b15OversoldCemeteryOracle, 0, 0)
	dead := pushGraveyardCardForTest(owner, "Dead Bear")
	pushGraveyardCardForTest(owner, "Dead Two")
	pushGraveyardCardForTest(owner, "Dead Three")
	relic := batch01GraveyardCard(owner, "Sol Ring", "Artifact")

	advanceToUpkeepOf(t, g, 1)
	if len(g.PendingChoices) != 0 {
		t.Fatal("three creature cards and an artifact is not four — no prompt")
	}
	passPriorityAroundTable(t, g)

	pushGraveyardCardForTest(owner, "Dead Four")
	advanceToUpkeepOf(t, g, 2)
	advanceToUpkeepOf(t, g, 1)
	answerLatestTriggerPrompt(t, g, owner.ID, true)
	b04WaitForPick(t, g, owner.ID)
	if hasID(latestPickTarget(g, owner.ID).PickTargetCards, relic) {
		t.Error("an artifact card was offered for 'target creature card'")
	}
	pickCard(t, g, owner.ID, dead)
	passPriorityAroundTable(t, g)
	if !owner.Hand.Contains(dead) {
		t.Error("the creature card did not return to hand")
	}
}

func TestB15OversoldCemeteryChecksTheCountAgainOnResolution(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	b12Push(g, owner.ID, "Oversold Cemetery", "Enchantment", b15OversoldCemeteryOracle, 0, 0)
	dead := pushGraveyardCardForTest(owner, "Dead Bear")
	extra := pushGraveyardCardForTest(owner, "Dead Two")
	pushGraveyardCardForTest(owner, "Dead Three")
	pushGraveyardCardForTest(owner, "Dead Four")
	advanceToUpkeepOf(t, g, 1)
	answerLatestTriggerPrompt(t, g, owner.ID, true)
	b04WaitForPick(t, g, owner.ID)
	pickCard(t, g, owner.ID, dead)
	// In response, one of the other creature cards leaves the yard.
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(extra) })
	passPriorityAroundTable(t, g)
	if owner.Hand.Contains(dead) || !owner.Graveyard.Contains(dead) {
		t.Error("with three creature cards left the intervening-if fails and nothing returns")
	}
}

// --- removal and reanimation ---------------------------------------

func TestB15UtterEndExilesAnyNonlandPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := seedPermanentFor(g, opp.ID, "Rock", "Artifact")
	castCatalogSpell(t, g, "Utter End", "Instant", b15UtterEndOracle, []game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !exileHas(g, rock) {
		t.Error("the artifact should be in exile")
	}
	land := seedLandOnBattlefield(g, opp.ID, "Forest", "Basic Land — Forest")
	if err := g.CastSpell(me.ID, b13HandCard(me, "Utter End", "Instant", b15UtterEndOracle),
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}}}); err == nil {
		t.Error("a land was accepted as a target for 'target nonland permanent'")
	}
}

func TestB15RunAwayTogetherBouncesTwoCreaturesOfDifferentControllers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine, theirs := seedCreature(g, "My Bear", me.ID), seedCreature(g, "Their Bear", opp.ID)
	castCatalogSpell(t, g, "Run Away Together", "Instant", b15RunAwayTogetherOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(mine) || !opp.Hand.Contains(theirs) {
		t.Error("both creatures return to their owners' hands")
	}
	// The declared gap: a same-controller pair is accepted and does
	// nothing.
	a, b := seedCreature(g, "Bear A", opp.ID), seedCreature(g, "Bear B", opp.ID)
	castCatalogSpell(t, g, "Run Away Together", "Instant", b15RunAwayTogetherOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(a) || !g.Battlefield.Contains(b) {
		t.Error("two creatures with the same controller are not a legal pair")
	}
	// One target gone in response: the other still returns.
	c, d := seedCreature(g, "Bear C", me.ID), seedCreature(g, "Bear D", opp.ID)
	castCatalogSpell(t, g, "Run Away Together", "Instant", b15RunAwayTogetherOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: c}, {Kind: game.TargetCard, ID: d}})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(c) })
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(d) {
		t.Error("the surviving target is still returned")
	}
}

func TestB15StitchTogetherReturnsToHandOrWithThresholdToTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dead := pushGraveyardCardForTest(me, "Dead Bear")
	castCatalogSpell(t, g, "Stitch Together", "Sorcery", b15StitchTogetherOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: dead}})
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(dead) {
		t.Fatal("below threshold the card returns to hand")
	}
	// Threshold: the target plus six others is seven; the Stitch
	// Together resolving is on the stack, not in the graveyard.
	second := pushGraveyardCardForTest(me, "Dead Two")
	for i := 0; i < 6; i++ {
		pushGraveyardCardForTest(me, "Filler")
	}
	if me.Graveyard.Size() < 7 {
		t.Fatalf("graveyard holds %d, want at least 7", me.Graveyard.Size())
	}
	castCatalogSpell(t, g, "Stitch Together", "Sorcery", b15StitchTogetherOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: second}})
	passPriorityAroundTable(t, g)
	card, ok := battlefieldCard(g, second)
	if !ok {
		t.Fatal("at threshold the card returns to the battlefield instead")
	}
	if card.Controller != me.ID {
		t.Error("it returns under its owner's control")
	}
}

func TestB15WindsOfRathSparesEnchantedCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := seedCreature(g, "My Bear", me.ID)
	blessed := seedCreature(g, "Blessed Bear", me.ID)
	b15Aura(g, me.ID, blessed)
	theirs := seedCreature(g, "Their Bear", opp.ID)
	cursed := seedCreature(g, "Cursed Bear", opp.ID)
	b15Aura(g, me.ID, cursed)
	equipped := seedCreature(g, "Equipped Bear", opp.ID)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sword", TypeLine: "Artifact — Equipment",
		Owner: opp.ID, Controller: opp.ID,
		AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: equipped},
	})
	castCatalogSpell(t, g, "Winds of Rath", "Sorcery", b15WindsOfRathOracle, nil)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, theirs, equipped} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s: an unenchanted creature (an Equipment does not enchant) is destroyed", cardByID(g, id).Name)
		}
	}
	for _, id := range []uuid.UUID{blessed, cursed} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s: an enchanted creature survives, whoever controls the Aura", cardByID(g, id).Name)
		}
	}
}

func TestB15BloodMoneyPaysATappedTreasurePerNontokenCreatureDestroyed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedCreature(g, "My Bear", me.ID)
	seedCreature(g, "Their Bear", opp.ID)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, RedGoblinToken(), 2) })
	darksteel := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Darksteel Myr", TypeLine: "Artifact Creature — Myr",
		Power: 0, Toughness: 1, Keywords: []string{"indestructible"}, Owner: opp.ID, Controller: opp.ID,
	})
	castCatalogSpell(t, g, "Blood Money", "Sorcery", b15BloodMoneyOracle, nil)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldByName(g, "Goblin"); n != 0 {
		t.Errorf("the tokens die too: %d Goblins left", n)
	}
	if !g.Battlefield.Contains(darksteel) {
		t.Error("an indestructible creature survives (#446 — DestroyAllMatching honours it)")
	}
	treasures := battlefieldIDsNamed(g, "Treasure")
	if len(treasures) != 2 {
		t.Fatalf("two nontoken creatures died (tokens and the survivor pay nothing): %d Treasures, want 2", len(treasures))
	}
	for _, id := range treasures {
		if !b13Tapped(t, g, id) || controllerOf(t, g, id) != me.ID {
			t.Error("the Treasures are the caster's and enter tapped")
		}
	}
}

// A commander caught in Blood Money was destroyed and pays a Treasure
// (CR 903.9 replaces the zone change, not the destruction). The engine
// keeps it on the battlefield until its owner answers the command-zone
// prompt, so the count is not knowable while the prompt is open.
//
// #815: the card WAITS for it. "For each creature destroyed this way"
// runs from DestroyPermanentsThenForEffect's continuation, so no
// Treasure is made until the answer arrives — and then the number is
// the real one, whichever answer it was. Before #815 the Treasures
// were made on the spot and the commander was counted on the strength
// of the prompt having been queued, which was right here and wrong for
// a destruction the window cancelled outright.
func TestB15BloodMoneyPaysForACommanderCaughtInTheWipe(t *testing.T) {
	for _, tc := range []struct {
		name        string
		commandZone bool
	}{{"to the command zone", true}, {"to the graveyard", false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			seedCreature(g, "My Bear", me.ID)
			commander := b36Commander(g, me.ID, "My Commander")
			g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 1) })

			castCatalogSpell(t, g, "Blood Money", "Sorcery", b15BloodMoneyOracle, nil)
			passPriorityAroundTable(t, g)
			if n := len(battlefieldIDsNamed(g, "Treasure")); n != 0 {
				t.Errorf("%d Treasures while the CR 903.9 prompt is still open, want 0 — "+
					"the count is not knowable until it is answered", n)
			}

			if tc.commandZone {
				b36AcceptCommandZone(t, g, me.ID)
				if !me.Command.Contains(commander) {
					t.Error("the commander goes to the command zone")
				}
			} else {
				b21DeclineCommandZone(t, g, me.ID)
				if !me.Graveyard.Contains(commander) {
					t.Error("the commander goes to the graveyard")
				}
			}
			if n := len(battlefieldIDsNamed(g, "Treasure")); n != 2 {
				t.Errorf("after the CR 903.9 answer: %d Treasures, want 2", n)
			}
		})
	}
}

// Blood Money destroys every creature at the same time (CR 700.4), so
// a Zulaport Cutthroat caught in the wipe triggers for every creature
// its controller lost, itself included. Zulaport is pushed first, the
// battlefield position where the old one-at-a-time loop destroyed it
// before the others and it saw only its own death.
func TestB15BloodMoneyDeathsAreSimultaneousForAristocratsPayoffs(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Zulaport Cutthroat", "Creature — Human Rogue", zulaportOracle, false)
	pushWipeCreature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	pushWipeCreature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	pushWipeCreature(g, me.ID, "Bear C", "Creature — Bear", 2, 2)
	survivor := pushIndestructibleWipeCreature(g, me.ID, "Darksteel Myr")

	meBefore, oppBefore := me.Life, opp.Life
	castCatalogSpell(t, g, "Blood Money", "Sorcery", b15BloodMoneyOracle, nil)
	passPriorityAroundTable(t, g)

	if want := oppBefore - 4; opp.Life != want {
		t.Errorf("opponent life %d -> %d, want %d (four simultaneous deaths; the indestructible survivor is not one)", oppBefore, opp.Life, want)
	}
	if want := meBefore + 4; me.Life != want {
		t.Errorf("caster life %d -> %d, want %d", meBefore, me.Life, want)
	}
	if !g.Battlefield.Contains(survivor) {
		t.Error("the indestructible creature survives the sweep")
	}
	if n := len(battlefieldIDsNamed(g, "Treasure")); n != 4 {
		t.Errorf("four nontoken creatures died: %d Treasures, want 4", n)
	}
}

// A commander prompt splits the physical moves across two actions, but the
// wipe is still one simultaneous destruction. In particular, a Cutthroat
// processed before the commander has already left the battlefield when the
// owner answers; the carried LKI batch must be active around that resumed
// move itself, not only around the continuation that starts the next leg.
func TestB15BloodMoneyPausedCommanderStaysInTheSimultaneousDeathBatch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Zulaport Cutthroat", "Creature — Human Rogue", zulaportOracle, false)
	pushWipeCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	commander := b36Commander(g, me.ID, "My Commander")

	meBefore, oppBefore := me.Life, opp.Life
	castCatalogSpell(t, g, "Blood Money", "Sorcery", b15BloodMoneyOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(commander) {
		t.Fatal("the commander waits on its CR 903.9 choice")
	}

	// Declining makes this a graveyard death, which Zulaport watches.
	b21DeclineCommandZone(t, g, me.ID)
	passPriorityAroundTable(t, g)

	if want := oppBefore - 3; opp.Life != want {
		t.Errorf("opponent life %d -> %d, want %d (Cutthroat, Bear and resumed commander died together)", oppBefore, opp.Life, want)
	}
	if want := meBefore + 3; me.Life != want {
		t.Errorf("caster life %d -> %d, want %d", meBefore, me.Life, want)
	}
}

// --- statics -------------------------------------------------------

func TestB15MirrorBoxPumpsLegendsAndSameNamedNontokens(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Mirror Box", "Artifact", b15MirrorBoxOracle, 0, 0)
	legend := b12Creature(g, me.ID, "Legend", "Legendary Creature — Human Knight", 2, 2)
	elfA := b12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	elfB := b12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirElf := b12Creature(g, opp.ID, "Elf", "Creature — Elf", 1, 1)
	if p := effectivePower(t, g, legend); p != 3 {
		t.Errorf("a legendary creature gets +1/+1: power %d, want 3", p)
	}
	if p := effectivePower(t, g, elfA); p != 2 {
		t.Errorf("one other Elf you control with the same name: power %d, want 2", p)
	}
	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("a lone nonlegendary is untouched: power %d, want 2", p)
	}
	if p := effectivePower(t, g, theirElf); p != 1 {
		t.Errorf("an opponent's Elf is untouched: power %d, want 1", p)
	}
	token := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Elf", TypeLine: "Token Creature — Elf",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	if p := effectivePower(t, g, elfB); p != 3 {
		t.Errorf("a token with the same name counts as an OTHER creature: power %d, want 3", p)
	}
	if p := effectivePower(t, g, token); p != 1 {
		t.Errorf("the token itself is not a nontoken creature: power %d, want 1", p)
	}
}

// --- tutors --------------------------------------------------------

func TestB15MoonsilverKeyFindsAManaRockOrABasic(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	key := b12Push(g, me.ID, "Moonsilver Key", "Artifact", b15MoonsilverKeyOracle, 0, 0)
	seedSearchLibrary(me,
		game.Card{Name: "Sol Ring", TypeLine: "Artifact", OracleID: b15SolRingOracle},
		game.Card{Name: "Plain Rock", TypeLine: "Artifact"},
		searchTestLand("Forest", "Basic Land — Forest"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	advanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, key, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	if g.Battlefield.Contains(key) {
		t.Error("the Key is sacrificed as a cost, at announce")
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("two matches: the searcher chooses")
	}
	if searchOptionNamed(g, c, "Plain Rock") != uuid.Nil || searchOptionNamed(g, c, "Bear") != uuid.Nil {
		t.Error("an artifact with no mana ability, and a creature, are not offered")
	}
	if searchOptionNamed(g, c, "Forest") == uuid.Nil {
		t.Error("a basic land is offered")
	}
	answerSearchNamed(t, g, me.ID, "Sol Ring")
	if !b02bHandHasNamed(me, "Sol Ring") {
		t.Error("the rock goes to hand")
	}
}
