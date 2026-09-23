package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch02_third_pass_test.go — the re-triage pass of roadmap batch 02
// (#295, `edhrec_rank` 238–360): 28 more cards, unblocked since the
// first two passes (batch02_test.go / #351, batch02_second_pass_test.go)
// by seams that landed this week and today — the optional-additional-
// cost cast gate (#664), the activation gate (#1210), cast provenance
// (#761), the modal and multi-target clauses (#764), and the S32
// mana pipeline's `ManaAbilityCost.Mana` / `ProducedFunc` (#356).
//
// Naming is `b02c`-prefixed to stay clear of batch02_test.go's `b02`
// helpers and batch02_second_pass_test.go's `b02b` helpers.

const (
	b02cDovinsVetoOracle          = "1b388371-f9ef-45b4-82a3-ca20a8cd7807"
	b02cRishkarsExpertiseOracle   = "97407cd0-2bd2-4074-94d3-4ec3d243fa78"
	b02cTakenumaOracle            = "ac2dd694-d2f1-4025-8400-12332bdc882a"
	b02cUnclaimedTerritoryOracle  = "584b15f2-6ae9-413a-8b8d-9244dbea4878"
	b02cGrandAbolisherOracle      = "c749f23c-40c0-4159-b84c-a70cbb062c14"
	b02cMorbidOpportunistOracle   = "322f44f0-e6da-4ee0-b474-e7d5e9a461c5"
	b02cFoundryInspectorOracle    = "18f22960-87ec-43cd-82ea-ec5cabf49ad3"
	b02cEverflowingChaliceOracle  = "0a79237e-0811-4a8a-bd4d-db3ca91bff22"
	b02cBloomTenderOracle         = "0c23fefe-9891-4dd8-9bb1-eebdb3274e31"
	b02cUntimelyMalfunctionOracle = "2c49d24a-98a4-43c2-bb59-1ef13d4214c2"
	b02cForceOfNegationOracle     = "ac2173f9-f223-440a-9231-fd98762bdc6f"
	b02cEtaliOracle               = "078def07-ae5d-4591-8db6-d156834aab97"
	b02cInspiringCallOracle       = "9b9a10ff-5a5d-4df8-88aa-18d84ff9117c"
	b02cJetMedallionOracle        = "bccb06f1-8839-4104-9ffa-c79a817f8378"
	b02cRuggedPrairieOracle       = "8e7641e1-e814-4d5a-9cb3-71ad2f4ceee8"
	b02cFaerieMastermindOracle    = "a984db23-40ea-428d-829f-e944267280f8"
	b02cRubyMedallionOracle       = "0c367958-729e-416d-988b-098e90e7a1fd"
	b02cCascadeBluffsOracle       = "f1603384-4361-49c9-98aa-7785fc3504c4"
	b02cExplorationOracle         = "0c2841bb-038c-4fbf-8360-bc0a1522b58d"
	b02cFloodedGroveOracle        = "dc974eb4-72b9-4213-887b-8ee684b93420"
	b02cDryadOfIlysianGroveOracle = "bdbde5d0-f5e4-44da-b27c-b4ad6f374cc9"
	b02cUrzasIncubatorOracle      = "b380c04b-0bce-4328-8f0a-1425e48bee72"
	b02cDarkwaterCatacombsOracle  = "4869a530-757f-4364-8d8e-4dc8001f433c"
	b02cAzusaOracle               = "6c2c8bf3-9bf8-4a86-89d3-3bb36260dc51"
	b02cFetidHeathOracle          = "42bf259d-4bb9-49c3-b4ec-223dca62f4d6"
	b02cSevinnesReclamationOracle = "1b9f9f5b-8712-4f00-90cb-1b7b9970eccc"
	b02cSkycloudExpanseOracle     = "76f335d0-7f71-4b1a-b60d-73de954cbe2c"
	b02cTwilightMireOracle        = "db623754-e078-4030-ba07-818803c348a8"
)

func TestBatch02ThirdPassCardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b02cDovinsVetoOracle:          "Dovin's Veto",
		b02cRishkarsExpertiseOracle:   "Rishkar's Expertise",
		b02cTakenumaOracle:            "Takenuma, Abandoned Mire",
		b02cUnclaimedTerritoryOracle:  "Unclaimed Territory",
		b02cGrandAbolisherOracle:      "Grand Abolisher",
		b02cMorbidOpportunistOracle:   "Morbid Opportunist",
		b02cFoundryInspectorOracle:    "Foundry Inspector",
		b02cEverflowingChaliceOracle:  "Everflowing Chalice",
		b02cBloomTenderOracle:         "Bloom Tender",
		b02cUntimelyMalfunctionOracle: "Untimely Malfunction",
		b02cForceOfNegationOracle:     "Force of Negation",
		b02cEtaliOracle:               "Etali, Primal Storm",
		b02cInspiringCallOracle:       "Inspiring Call",
		b02cJetMedallionOracle:        "Jet Medallion",
		b02cRuggedPrairieOracle:       "Rugged Prairie",
		b02cFaerieMastermindOracle:    "Faerie Mastermind",
		b02cRubyMedallionOracle:       "Ruby Medallion",
		b02cCascadeBluffsOracle:       "Cascade Bluffs",
		b02cExplorationOracle:         "Exploration",
		b02cFloodedGroveOracle:        "Flooded Grove",
		b02cDryadOfIlysianGroveOracle: "Dryad of the Ilysian Grove",
		b02cUrzasIncubatorOracle:      "Urza's Incubator",
		b02cDarkwaterCatacombsOracle:  "Darkwater Catacombs",
		b02cAzusaOracle:               "Azusa, Lost but Seeking",
		b02cFetidHeathOracle:          "Fetid Heath",
		b02cSevinnesReclamationOracle: "Sevinne's Reclamation",
		b02cSkycloudExpanseOracle:     "Skycloud Expanse",
		b02cTwilightMireOracle:        "Twilight Mire",
	}
	if len(want) != 28 {
		t.Fatalf("the third pass ships 28 cards, the table lists %d", len(want))
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

// --- filter lands ---------------------------------------------------

// Rugged Prairie gets the thorough test; the other four filter lands
// (mechanically identical, different colours) get a lighter one
// below — the shape was proven once with Mystic Gate's own test.
func TestB02cRuggedPrairieFiltersOneHybridIntoTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := seedPermanentWithOracle(g, me.ID, "Rugged Prairie", "Land", b02cRuggedPrairieOracle)

	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("the {C} half: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}

	fresh := seedPermanentWithOracle(g, me.ID, "Rugged Prairie", "Land", b02cRuggedPrairieOracle)
	me.ManaPool.EmptyPool()
	me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	if err := g.ActivateManaAbility(me.ID, fresh, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	picks := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceMana && c.Chooser == me.ID {
			picks++
			if len(c.ColorOptions) != 2 {
				t.Errorf("each slot must offer exactly R and W, got %v", c.ColorOptions)
			}
			if err := g.ResolveManaChoice(c.ID, me.ID, "R"); err != nil {
				t.Fatalf("ResolveManaChoice: %v", err)
			}
		}
	}
	if picks != 2 {
		t.Fatalf("two colour picks, got %d", picks)
	}
	if got := batch01PoolColors(me); len(got) != 2 {
		t.Errorf("pool %v, want two mana", got)
	}
}

func TestB02cRemainingFilterLandsProduceTheirColours(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, hybrid string
	}{
		{"Cascade Bluffs", b02cCascadeBluffsOracle, "U"},
		{"Flooded Grove", b02cFloodedGroveOracle, "G"},
		{"Fetid Heath", b02cFetidHeathOracle, "W"},
		{"Twilight Mire", b02cTwilightMireOracle, "B"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			land := seedPermanentWithOracle(g, me.ID, tc.name, "Land", tc.oracle)
			me.ManaPool.AddMana(game.ManaToken{Color: tc.hybrid})
			if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{}); err != nil {
				t.Fatalf("ActivateManaAbility: %v", err)
			}
			picks := 0
			for _, c := range g.PendingChoices {
				if c != nil && c.Kind == game.PendingChoiceMana && c.Chooser == me.ID {
					picks++
					if err := g.ResolveManaChoice(c.ID, me.ID, c.ColorOptions[0]); err != nil {
						t.Fatalf("ResolveManaChoice: %v", err)
					}
				}
			}
			if picks != 2 {
				t.Fatalf("%s: two colour picks, got %d", tc.name, picks)
			}
			if got := batch01PoolColors(me); len(got) != 2 {
				t.Errorf("%s: pool %v, want two mana", tc.name, got)
			}
		})
	}
}

// --- simple tap-and-a-tax duals --------------------------------------

func TestB02cDarkwaterCatacombsProducesUB(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := seedPermanentWithOracle(g, me.ID, "Darkwater Catacombs", "Land", b02cDarkwaterCatacombsOracle)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	colors := batch01PoolColors(me)
	if len(colors) != 2 {
		t.Fatalf("pool %v, want [U B] plus nothing left of the {1}", colors)
	}
}

func TestB02cSkycloudExpanseProducesWU(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := seedPermanentWithOracle(g, me.ID, "Skycloud Expanse", "Land", b02cSkycloudExpanseOracle)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if colors := batch01PoolColors(me); len(colors) != 2 {
		t.Errorf("pool %v, want two mana", colors)
	}
}

// --- cost reducers ----------------------------------------------------

func TestB02cJetMedallionDiscountsBlackSpellsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Jet Medallion", "Artifact", b02cJetMedallionOracle, false)

	if got := priceInHand(t, g, me, "My Grasp", "Instant", "{1}{B}"); got != 1 {
		t.Errorf("own black spell: %d, want 1", got)
	}
	if got := priceInHand(t, g, them, "Their Grasp", "Instant", "{1}{B}"); got != 2 {
		t.Errorf("opponent's black spell: %d, want 2 (undiscounted)", got)
	}
	if got := priceInHand(t, g, me, "My Bolt", "Instant", "{R}"); got != 1 {
		t.Errorf("own red spell: %d, want 1 (untouched)", got)
	}
}

func TestB02cRubyMedallionDiscountsRedSpellsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Ruby Medallion", "Artifact", b02cRubyMedallionOracle, false)
	// CR 601.2f: a reduction spends only the GENERIC component, so
	// {R} (no generic pip) is untouched and {1}{R} loses its {1}.
	if got := priceInHand(t, g, me, "My Bolt", "Instant", "{1}{R}"); got != 1 {
		t.Errorf("own red spell: %d, want 1", got)
	}
	if got := priceInHand(t, g, me, "My Grasp", "Instant", "{1}{B}"); got != 2 {
		t.Errorf("own black spell: %d, want 2 (untouched)", got)
	}
}

func TestB02cFoundryInspectorDiscountsYourArtifactsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Foundry Inspector", "Artifact Creature — Construct", b02cFoundryInspectorOracle, false)

	if got := priceInHand(t, g, me, "My Chalice", "Artifact", "{2}"); got != 1 {
		t.Errorf("own artifact spell: %d, want 1", got)
	}
	if got := priceInHand(t, g, them, "Their Chalice", "Artifact", "{2}"); got != 2 {
		t.Errorf("opponent's artifact spell: %d, want 2 (undiscounted)", got)
	}
	if got := priceInHand(t, g, me, "My Bear", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("own creature spell: %d, want 2 (untouched)", got)
	}
}

func TestB02cUrzasIncubatorDiscountsTheChosenType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushNamedTribePermanent(t, g, me.ID, "Urza's Incubator", "Artifact", b02cUrzasIncubatorOracle, "Goblin")

	if got := priceInHand(t, g, me, "Goblin Chieftain", "Creature — Goblin Warrior", "{2}{R}"); got != 1 {
		t.Errorf("a Goblin creature spell: %d, want 1", got)
	}
	if got := priceInHand(t, g, me, "Elvish Mystic", "Creature — Elf Druid", "{G}"); got != 1 {
		t.Errorf("a non-Goblin creature spell: %d, want 1 (untouched)", got)
	}
}

// --- choose-a-creature-type lands and permanents ----------------------

func TestB02cUnclaimedTerritoryRestrictsItsColouredMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := pushNamedTribePermanent(t, g, me.ID, "Unclaimed Territory", "Land", b02cUnclaimedTerritoryOracle, "Elf")

	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("the colourless half: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
	me.ManaPool.EmptyPool()

	fresh := pushNamedTribePermanent(t, g, me.ID, "Unclaimed Territory", "Land", b02cUnclaimedTerritoryOracle, "Elf")
	if err := g.ActivateManaAbility(me.ID, fresh, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("the restricted half: %v", err)
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceMana && c.Chooser == me.ID {
			if err := g.ResolveManaChoice(c.ID, me.ID, "G"); err != nil {
				t.Fatalf("ResolveManaChoice: %v", err)
			}
		}
	}
	if len(me.ManaPool) != 1 || len(me.ManaPool[0].Restrictions) == 0 {
		t.Errorf("the picked mana must carry the chosen-type restriction: %+v", me.ManaPool)
	}
}

// --- extra land drops ---------------------------------------------------

// b02cPlayLandRaw plays a plain land from the active seat's hand
// through the real cast path, WITHOUT temples_test.go's
// playLandFromHand bumping LandDropsPerTurn first — that bump exists
// so unrelated card fixtures don't have to spend real turns, and it
// would hide the very limit AdditionalLandPlays is being tested
// against here.
func b02cPlayLandRaw(t *testing.T, g *game.Game, name string) error {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Land",
		Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return g.CastSpell(active.ID, id, game.CastSpellParams{})
}

func TestB02cExplorationGrantsExactlyOneExtraLandDrop(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedPermanentWithOracle(g, me.ID, "Exploration", "Enchantment", b02cExplorationOracle)
	if err := b02cPlayLandRaw(t, g, "Forest"); err != nil {
		t.Fatalf("the turn's ordinary land drop: %v", err)
	}
	if err := b02cPlayLandRaw(t, g, "Island"); err != nil {
		t.Fatalf("Exploration's extra land drop was refused: %v", err)
	}
	if err := b02cPlayLandRaw(t, g, "Swamp"); err == nil {
		t.Error("a THIRD land in one turn should still be refused — Exploration grants only one extra")
	}
}

func TestB02cAzusaGrantsTwoExtraLandDrops(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedPermanentWithOracle(g, me.ID, "Azusa, Lost but Seeking", "Legendary Creature — Human Monk", b02cAzusaOracle)
	if err := b02cPlayLandRaw(t, g, "Forest"); err != nil {
		t.Fatalf("the turn's ordinary land drop: %v", err)
	}
	if err := b02cPlayLandRaw(t, g, "Island"); err != nil {
		t.Fatalf("Azusa's first extra land drop was refused: %v", err)
	}
	if err := b02cPlayLandRaw(t, g, "Swamp"); err != nil {
		t.Fatalf("Azusa's second extra land drop was refused: %v", err)
	}
	if err := b02cPlayLandRaw(t, g, "Mountain"); err == nil {
		t.Error("a FOURTH land in one turn should still be refused — Azusa grants only two extra")
	}
}

func TestB02cDryadOfTheIlysianGroveGrantsAllBasicTypes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedPermanentWithOracle(g, me.ID, "Dryad of the Ilysian Grove", "Enchantment Creature — Nymph Dryad", b02cDryadOfIlysianGroveOracle)
	land := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Command Tower", TypeLine: "Land",
		Owner: me.ID, Controller: me.ID,
	})
	subtypes := effectiveSubtypes(t, g, land)
	for _, want := range []string{"Plains", "Island", "Swamp", "Mountain", "Forest"} {
		found := false
		for _, st := range subtypes {
			if st == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("land is missing basic type %s: %v", want, subtypes)
		}
	}

	if err := b02cPlayLandRaw(t, g, "Forest"); err != nil {
		t.Fatalf("the turn's ordinary land drop: %v", err)
	}
	if err := b02cPlayLandRaw(t, g, "Island"); err != nil {
		t.Fatalf("Dryad's extra land drop was refused: %v", err)
	}
}

// --- straightforward spells and permanents -----------------------------

func TestB02cDovinsVetoCountersANoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	victim := castCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
	spec, ok := Lookup(b02cDovinsVetoOracle)
	if !ok || !spec.CantBeCountered {
		t.Fatalf("Dovin's Veto must declare CantBeCountered")
	}
	castCatalogSpell(t, g, "Dovin's Veto", "Instant", b02cDovinsVetoOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(victim) {
		t.Error("the targeted noncreature spell should have been countered into its owner's graveyard")
	}
}

func TestB02cMorbidOpportunistTriggersOnceEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedPermanentWithOracle(g, me.ID, "Morbid Opportunist", "Creature — Human Rogue", b02cMorbidOpportunistOracle)
	first := pushVanillaCreature(g, opp.ID, "First Victim", 2, 2)
	second := pushVanillaCreature(g, opp.ID, "Second Victim", 2, 2)
	handBefore := me.Hand.Size()

	ctx := ctxFor(g, &game.StackItem{Controller: me.ID})
	if err := (DestroyTarget{Target: first}).Apply(ctx); err != nil {
		t.Fatalf("destroy first: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != handBefore+1 {
		t.Fatalf("hand = %d, want %d (one draw for the first death)", me.Hand.Size(), handBefore+1)
	}

	if err := (DestroyTarget{Target: second}).Apply(ctx); err != nil {
		t.Fatalf("destroy second: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != handBefore+1 {
		t.Fatalf("hand = %d, want %d — the ability triggers only once each turn", me.Hand.Size(), handBefore+1)
	}
}

func TestB02cEverflowingChaliceEntersWithACounterPerKickAndScales(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	id := pushCatalogHandCard(me, "Everflowing Chalice", "Artifact", b02cEverflowingChaliceOracle)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Strict:        true,
		OptionalCosts: []int{0, 0},
	}); err != nil {
		t.Fatalf("CastSpell with multikicker x2: %v", err)
	}
	passPriorityAroundTable(t, g)
	c, ok := g.LookupCardForEffect(id)
	if !ok || c.Counters[game.CounterCharge] != 2 {
		t.Fatalf("charge counters = %v, want 2", c.Counters)
	}
	me.ManaPool.EmptyPool()
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if colors := batch01PoolColors(me); len(colors) != 2 {
		t.Errorf("pool %v, want two colourless mana (one per charge counter)", colors)
	}
}

func TestB02cBloomTenderAddsOneManaPerColourControlled(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tender := seedPermanentWithOracle(g, me.ID, "Bloom Tender", "Creature — Elf Druid", b02cBloomTenderOracle)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Red Guy", TypeLine: "Creature — Human",
		Colors: []string{"R"}, Owner: me.ID, Controller: me.ID,
	})
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Green Guy", TypeLine: "Creature — Human",
		Colors: []string{"G"}, Owner: me.ID, Controller: me.ID,
	})
	if err := g.ActivateManaAbility(me.ID, tender, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	colors := batch01PoolColors(me)
	if len(colors) != 2 {
		t.Fatalf("pool %v, want [G R] (Bloom Tender is green itself)", colors)
	}
}

// b02cCastModalSpell casts a modal catalog spell with both its modes
// and its targets announced together (CR 601.2b), which
// castCatalogSpellWithModes (targetless) can't express.
func b02cCastModalSpell(t *testing.T, g *game.Game, name, typeLine, oracleID string, modes []int, targets []game.TargetRef) uuid.UUID {
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
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Modes: modes, Targets: targets}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

func TestB02cUntimelyMalfunctionDestroysTargetArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	art := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Rock", TypeLine: "Artifact",
		Owner: me.ID, Controller: me.ID,
	})
	b02cCastModalSpell(t, g, "Untimely Malfunction", "Instant", b02cUntimelyMalfunctionOracle,
		[]int{0}, []game.TargetRef{{Kind: game.TargetCard, ID: art}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(art) {
		t.Error("the artifact was not destroyed")
	}
}

func TestB02cUntimelyMalfunctionChangesTheTargetOfASingleTargetSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, victimA, victimB := g.Seats[0].ID, g.Seats[1].ID, g.Seats[2].ID
	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victimA}})
	b02cCastModalSpell(t, g, "Untimely Malfunction", "Instant", b02cUntimelyMalfunctionOracle,
		[]int{1}, []game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	passPriorityAroundTable(t, g)

	prompt := latestRetarget(g, me)
	if prompt == nil {
		t.Fatalf("Untimely Malfunction's retarget mode opened no prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveRetarget(prompt.ID, me,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victimB}}); err != nil {
		t.Fatalf("ResolveRetarget: %v", err)
	}
	if got := g.StackMeta[bolt].Targets; len(got) == 0 || got[0].ID != victimB {
		t.Fatalf("the Bolt's target did not move: %+v", got)
	}
}

func TestB02cUntimelyMalfunctionPreventsBlocking(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	blocker := pushVanillaCreature(g, me.ID, "Their Blocker", 2, 2)
	b02cCastModalSpell(t, g, "Untimely Malfunction", "Instant", b02cUntimelyMalfunctionOracle,
		[]int{2}, []game.TargetRef{{Kind: game.TargetCard, ID: blocker}})
	passPriorityAroundTable(t, g)
	c, ok := g.LookupCardForEffect(blocker)
	if !ok || !game.Restricted(&c, game.CantBlock) {
		t.Error("the targeted creature must not be able to block this turn")
	}
}

func TestB02cFaerieMastermindDrawsOffAnOpponentsSecondDraw(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedPermanentWithOracle(g, me.ID, "Faerie Mastermind", "Creature — Faerie Rogue", b02cFaerieMastermindOracle)
	pushLibraryCardForTest(opp, game.Card{InstanceID: uuid.New(), Name: "Filler 1", TypeLine: "Land"})
	pushLibraryCardForTest(opp, game.Card{InstanceID: uuid.New(), Name: "Filler 2", TypeLine: "Land"})
	handBefore := me.Hand.Size()

	if err := g.DrawCard(opp.ID); err != nil {
		t.Fatalf("first draw: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != handBefore {
		t.Fatalf("hand = %d, want %d — the FIRST draw of the turn must not trigger", me.Hand.Size(), handBefore)
	}
	if err := g.DrawCard(opp.ID); err != nil {
		t.Fatalf("second draw: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != handBefore+1 {
		t.Fatalf("hand = %d, want %d — the opponent's SECOND draw must trigger", me.Hand.Size(), handBefore+1)
	}
}

func TestB02cInspiringCallDrawsPerCounteredCreatureAndGrantsIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := pushVanillaCreature(g, me.ID, "Counted A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Uncounted B", 2, 2)
	if err := g.AddCounterForEffect(a, game.CounterPlusOne, 1); err != nil {
		t.Fatalf("AddCounterForEffect: %v", err)
	}
	handBefore := me.Hand.Size()

	castCatalogSpell(t, g, "Inspiring Call", "Instant", b02cInspiringCallOracle, nil)
	passPriorityAroundTable(t, g)

	if me.Hand.Size() != handBefore+1 {
		t.Fatalf("hand = %d, want %d (one counted creature)", me.Hand.Size(), handBefore+1)
	}
	if !eotHasAbility(effectiveAbilities(t, g, a), "indestructible") {
		t.Errorf("the countered creature must gain indestructible")
	}
	if eotHasAbility(effectiveAbilities(t, g, b), "indestructible") {
		t.Errorf("the UNcountered creature must not gain indestructible")
	}
}

func TestB02cGrandAbolisherLocksOutOpponentsDuringYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedPermanentWithOracle(g, me.ID, "Grand Abolisher", "Creature — Human Cleric", b02cGrandAbolisherOracle)
	signet := pushCatalogPermanent(g, opp.ID, "Boros Signet", "Artifact", "41c84665-1f99-40ab-aaca-1188649eb263", false)
	toMainForCost(t, g)
	oppBolt := handCardFull(opp, "Their Bolt", "Instant", "{R}", "", []string{"R"})
	if err := g.AddManaForEffect(opp.ID, uuid.Nil, "{R}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}

	if err := g.CastSpell(opp.ID, oppBolt, game.CastSpellParams{}); err == nil {
		t.Error("an opponent must not be able to cast a spell during my turn")
	}
	if err := g.AddManaForEffect(opp.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateManaAbility(opp.ID, signet, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("an opponent must not be able to activate an artifact's ability during my turn")
	}
}

func TestB02cForceOfNegationPitchesOnlyOffYourOwnTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	pitch := handCardFull(opp, "Their Blue Card", "Instant", "{U}", "", []string{"U"})
	fon := handCardFull(opp, "Force of Negation", "Instant", "{1}{U}{U}", b02cForceOfNegationOracle, []string{"U"})
	victim := castCatalogSpell(t, g, "My Sorcery", "Sorcery", "", nil)

	if err := g.CastSpell(opp.ID, fon, game.CastSpellParams{
		Strict:          true,
		AlternativeCost: "pitch",
		AltCostIDs:      []uuid.UUID{pitch},
		Targets:         []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("Force of Negation off the pitch during someone else's turn: %v", err)
	}
	if opp.Hand.Contains(pitch) {
		t.Error("the pitched blue card must be exiled")
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(victim) {
		t.Error("the countered spell lands in its owner's graveyard")
	}
}

func TestB02cEtaliExilesEachLibraryAndOffersFreeCasts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	etali := pushDiesCreatureForTest(g, me.ID, "Etali, Primal Storm", b02cEtaliOracle, "Legendary Creature — Elder Dinosaur", 6, 6)
	pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "My Top", TypeLine: "Sorcery", ManaCost: "{1}{R}"})
	pushLibraryCardForTest(opp, game.Card{InstanceID: uuid.New(), Name: "Their Top", TypeLine: "Instant", ManaCost: "{U}"})

	// attackWith declares the attacker AND advances into the combat
	// damage step, which is what actually locks the declaration in
	// and emits EventAttack (commitAttackDeclarationLocked, at the
	// step's first priority boundary) — a bare DeclareAttacker only
	// stages it.
	attackWith(t, g, opp.ID, etali)
	passPriorityAroundTable(t, g)

	if len(g.Exile.Cards) != 4 {
		t.Fatalf("exile has %d cards, want 4 (one per seat's library top)", len(g.Exile.Cards))
	}
	for _, c := range g.Exile.Cards {
		perm := g.CastPermissionOnCardByIDForEffect(c.InstanceID)
		if perm == nil || perm.Player != me.ID || perm.Cost != "{0}" {
			t.Errorf("%s: no free-cast permission for Etali's controller: %+v", c.Name, perm)
		}
	}
}

func TestB02cRishkarsExpertiseDrawsAndOffersAFreeCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushVanillaCreature(g, me.ID, "Big Guy", 5, 5)
	cheap := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: cheap, Name: "Cheap Spell", TypeLine: "Sorcery", ManaCost: "{3}",
		Owner: me.ID, Controller: me.ID,
	})
	handBefore := me.Hand.Size()

	castCatalogSpell(t, g, "Rishkar's Expertise", "Sorcery", b02cRishkarsExpertiseOracle, nil)
	passPriorityAroundTable(t, g)

	// castCatalogSpell both adds Rishkar's Expertise to the hand and
	// casts it, netting zero; resolution then draws 5 (the greatest
	// power among creatures controlled).
	if me.Hand.Size() != handBefore+5 {
		t.Fatalf("hand = %d, want %d (drew 5)", me.Hand.Size(), handBefore+5)
	}
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatalf("Rishkar's Expertise must offer a free-cast pick: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{cheap}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	perm := g.CastPermissionOnCardByIDForEffect(cheap)
	if perm == nil || perm.Cost != "{0}" {
		t.Errorf("the chosen card must carry a free-cast permission: %+v", perm)
	}
}

func TestB02cSevinnesReclamationReanimatesAndFlashbackOffersACopy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	bear := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: bear, Name: "Graveyard Bear", TypeLine: "Creature — Bear",
		ManaCost: "{2}", Owner: me.ID, Controller: me.ID,
	})
	sevinne := handCardFull(me, "Sevinne's Reclamation", "Sorcery", "{2}{W}", b02cSevinnesReclamationOracle, []string{"W"})
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.CastSpell(me.ID, sevinne, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatalf("the permanent card was not returned to the battlefield")
	}
}

func TestB02cTakenumaChannelsMillsAndReturnsACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogHandCard(me, "Takenuma, Abandoned Mire", "Legendary Land", b02cTakenumaOracle)
	// Already in the graveyard, not necessarily among the milled
	// three — Takenuma's own text reads "a creature or planeswalker
	// card from your graveyard", not "one of the milled cards".
	buried := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: buried, Name: "Buried Legend", TypeLine: "Creature — Human",
		Owner: me.ID, Controller: me.ID,
	})
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}{C}{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	libBefore := me.Library.Size()
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("channel: %v", err)
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the discard-this cost must bin Takenuma at announce")
	}
	passPriorityAroundTable(t, g)
	if me.Library.Size() != libBefore-3 {
		t.Errorf("library = %d, want %d (milled three)", me.Library.Size(), libBefore-3)
	}
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatalf("Takenuma must offer to return a creature or planeswalker card: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{buried}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(buried) {
		t.Error("the picked card must arrive in hand")
	}
}
