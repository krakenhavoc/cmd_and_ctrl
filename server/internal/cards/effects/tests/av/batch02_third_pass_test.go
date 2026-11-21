package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestBatch02ThirdPassCardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.B02cDovinsVetoOracle:          "Dovin's Veto",
		effectstest.B02cRishkarsExpertiseOracle:   "Rishkar's Expertise",
		effectstest.B02cTakenumaOracle:            "Takenuma, Abandoned Mire",
		effectstest.B02cUnclaimedTerritoryOracle:  "Unclaimed Territory",
		effectstest.B02cGrandAbolisherOracle:      "Grand Abolisher",
		effectstest.B02cMorbidOpportunistOracle:   "Morbid Opportunist",
		effectstest.B02cFoundryInspectorOracle:    "Foundry Inspector",
		effectstest.B02cEverflowingChaliceOracle:  "Everflowing Chalice",
		effectstest.B02cBloomTenderOracle:         "Bloom Tender",
		effectstest.B02cUntimelyMalfunctionOracle: "Untimely Malfunction",
		effectstest.B02cForceOfNegationOracle:     "Force of Negation",
		effectstest.B02cEtaliOracle:               "Etali, Primal Storm",
		effectstest.B02cInspiringCallOracle:       "Inspiring Call",
		effectstest.B02cJetMedallionOracle:        "Jet Medallion",
		effectstest.B02cRuggedPrairieOracle:       "Rugged Prairie",
		effectstest.B02cFaerieMastermindOracle:    "Faerie Mastermind",
		effectstest.B02cRubyMedallionOracle:       "Ruby Medallion",
		effectstest.B02cCascadeBluffsOracle:       "Cascade Bluffs",
		effectstest.B02cExplorationOracle:         "Exploration",
		effectstest.B02cFloodedGroveOracle:        "Flooded Grove",
		effectstest.B02cDryadOfIlysianGroveOracle: "Dryad of the Ilysian Grove",
		effectstest.B02cUrzasIncubatorOracle:      "Urza's Incubator",
		effectstest.B02cDarkwaterCatacombsOracle:  "Darkwater Catacombs",
		effectstest.B02cAzusaOracle:               "Azusa, Lost but Seeking",
		effectstest.B02cFetidHeathOracle:          "Fetid Heath",
		effectstest.B02cSevinnesReclamationOracle: "Sevinne's Reclamation",
		effectstest.B02cSkycloudExpanseOracle:     "Skycloud Expanse",
		effectstest.B02cTwilightMireOracle:        "Twilight Mire",
	}
	if len(want) != 28 {
		t.Fatalf("the third pass ships 28 cards, the table lists %d", len(want))
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

// --- filter lands ---------------------------------------------------

// Rugged Prairie gets the thorough test; the other four filter lands
// (mechanically identical, different colours) get a lighter one
// below — the shape was proven once with Mystic Gate's own test.
func TestB02cRuggedPrairieFiltersOneHybridIntoTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.SeedPermanentWithOracle(g, me.ID, "Rugged Prairie", "Land", effectstest.B02cRuggedPrairieOracle)

	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("the {C} half: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}

	fresh := effectstest.SeedPermanentWithOracle(g, me.ID, "Rugged Prairie", "Land", effectstest.B02cRuggedPrairieOracle)
	me.ManaPool.EmptyPool()
	me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	if err := g.ActivateManaAbility(me.ID, fresh, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	// #3939: one choice, asked once per pick.
	picks := 0
	for c := effectstest.RiderLatestManaPick(g, me.ID); c != nil && picks < 8; c = effectstest.RiderLatestManaPick(g, me.ID) {
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
	if got := effectstest.Batch01PoolColors(me); len(got) != 2 {
		t.Errorf("pool %v, want two mana", got)
	}
}

func TestB02cRemainingFilterLandsProduceTheirColours(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, hybrid string
	}{
		{"Cascade Bluffs", effectstest.B02cCascadeBluffsOracle, "U"},
		{"Flooded Grove", effectstest.B02cFloodedGroveOracle, "G"},
		{"Fetid Heath", effectstest.B02cFetidHeathOracle, "W"},
		{"Twilight Mire", effectstest.B02cTwilightMireOracle, "B"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			land := effectstest.SeedPermanentWithOracle(g, me.ID, tc.name, "Land", tc.oracle)
			me.ManaPool.AddMana(game.ManaToken{Color: tc.hybrid})
			if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{}); err != nil {
				t.Fatalf("ActivateManaAbility: %v", err)
			}
			// #3939: one choice, asked once per pick.
			picks := 0
			for c := effectstest.RiderLatestManaPick(g, me.ID); c != nil && picks < 8; c = effectstest.RiderLatestManaPick(g, me.ID) {
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
			if got := effectstest.Batch01PoolColors(me); len(got) != 2 {
				t.Errorf("%s: pool %v, want two mana", tc.name, got)
			}
		})
	}
}

// --- simple tap-and-a-tax duals --------------------------------------

func TestB02cDarkwaterCatacombsProducesUB(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.SeedPermanentWithOracle(g, me.ID, "Darkwater Catacombs", "Land", effectstest.B02cDarkwaterCatacombsOracle)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	colors := effectstest.Batch01PoolColors(me)
	if len(colors) != 2 {
		t.Fatalf("pool %v, want [U B] plus nothing left of the {1}", colors)
	}
}

func TestB02cSkycloudExpanseProducesWU(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.SeedPermanentWithOracle(g, me.ID, "Skycloud Expanse", "Land", effectstest.B02cSkycloudExpanseOracle)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if colors := effectstest.Batch01PoolColors(me); len(colors) != 2 {
		t.Errorf("pool %v, want two mana", colors)
	}
}

// --- cost reducers ----------------------------------------------------

func TestB02cJetMedallionDiscountsBlackSpellsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Jet Medallion", "Artifact", effectstest.B02cJetMedallionOracle, false)

	if got := effectstest.PriceInHand(t, g, me, "My Grasp", "Instant", "{1}{B}"); got != 1 {
		t.Errorf("own black spell: %d, want 1", got)
	}
	if got := effectstest.PriceInHand(t, g, them, "Their Grasp", "Instant", "{1}{B}"); got != 2 {
		t.Errorf("opponent's black spell: %d, want 2 (undiscounted)", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "My Bolt", "Instant", "{R}"); got != 1 {
		t.Errorf("own red spell: %d, want 1 (untouched)", got)
	}
}

func TestB02cRubyMedallionDiscountsRedSpellsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Ruby Medallion", "Artifact", effectstest.B02cRubyMedallionOracle, false)
	// CR 601.2f: a reduction spends only the GENERIC component, so
	// {R} (no generic pip) is untouched and {1}{R} loses its {1}.
	if got := effectstest.PriceInHand(t, g, me, "My Bolt", "Instant", "{1}{R}"); got != 1 {
		t.Errorf("own red spell: %d, want 1", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "My Grasp", "Instant", "{1}{B}"); got != 2 {
		t.Errorf("own black spell: %d, want 2 (untouched)", got)
	}
}

func TestB02cFoundryInspectorDiscountsYourArtifactsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Foundry Inspector", "Artifact Creature — Construct", effectstest.B02cFoundryInspectorOracle, false)

	if got := effectstest.PriceInHand(t, g, me, "My Chalice", "Artifact", "{2}"); got != 1 {
		t.Errorf("own artifact spell: %d, want 1", got)
	}
	if got := effectstest.PriceInHand(t, g, them, "Their Chalice", "Artifact", "{2}"); got != 2 {
		t.Errorf("opponent's artifact spell: %d, want 2 (undiscounted)", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "My Bear", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("own creature spell: %d, want 2 (untouched)", got)
	}
}

func TestB02cUrzasIncubatorDiscountsTheChosenType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushNamedTribePermanent(t, g, me.ID, "Urza's Incubator", "Artifact", effectstest.B02cUrzasIncubatorOracle, "Goblin")

	if got := effectstest.PriceInHand(t, g, me, "Goblin Chieftain", "Creature — Goblin Warrior", "{2}{R}"); got != 1 {
		t.Errorf("a Goblin creature spell: %d, want 1", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Elvish Mystic", "Creature — Elf Druid", "{G}"); got != 1 {
		t.Errorf("a non-Goblin creature spell: %d, want 1 (untouched)", got)
	}
}

// --- choose-a-creature-type lands and permanents ----------------------

func TestB02cUnclaimedTerritoryRestrictsItsColouredMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.PushNamedTribePermanent(t, g, me.ID, "Unclaimed Territory", "Land", effectstest.B02cUnclaimedTerritoryOracle, "Elf")

	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("the colourless half: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
	me.ManaPool.EmptyPool()

	fresh := effectstest.PushNamedTribePermanent(t, g, me.ID, "Unclaimed Territory", "Land", effectstest.B02cUnclaimedTerritoryOracle, "Elf")
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
	if me.ManaPool.Total() != 1 || len(me.ManaPool.Units()[0].Restrictions) == 0 {
		t.Errorf("the picked mana must carry the chosen-type restriction: %+v", me.ManaPool)
	}
}

func TestB02cExplorationGrantsExactlyOneExtraLandDrop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedPermanentWithOracle(g, me.ID, "Exploration", "Enchantment", effectstest.B02cExplorationOracle)
	if err := effectstest.B02cPlayLandRaw(t, g, "Forest"); err != nil {
		t.Fatalf("the turn's ordinary land drop: %v", err)
	}
	if err := effectstest.B02cPlayLandRaw(t, g, "Island"); err != nil {
		t.Fatalf("Exploration's extra land drop was refused: %v", err)
	}
	if err := effectstest.B02cPlayLandRaw(t, g, "Swamp"); err == nil {
		t.Error("a THIRD land in one turn should still be refused — Exploration grants only one extra")
	}
}

func TestB02cAzusaGrantsTwoExtraLandDrops(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedPermanentWithOracle(g, me.ID, "Azusa, Lost but Seeking", "Legendary Creature — Human Monk", effectstest.B02cAzusaOracle)
	if err := effectstest.B02cPlayLandRaw(t, g, "Forest"); err != nil {
		t.Fatalf("the turn's ordinary land drop: %v", err)
	}
	if err := effectstest.B02cPlayLandRaw(t, g, "Island"); err != nil {
		t.Fatalf("Azusa's first extra land drop was refused: %v", err)
	}
	if err := effectstest.B02cPlayLandRaw(t, g, "Swamp"); err != nil {
		t.Fatalf("Azusa's second extra land drop was refused: %v", err)
	}
	if err := effectstest.B02cPlayLandRaw(t, g, "Mountain"); err == nil {
		t.Error("a FOURTH land in one turn should still be refused — Azusa grants only two extra")
	}
}

func TestB02cDryadOfTheIlysianGroveGrantsAllBasicTypes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedPermanentWithOracle(g, me.ID, "Dryad of the Ilysian Grove", "Enchantment Creature — Nymph Dryad", effectstest.B02cDryadOfIlysianGroveOracle)
	land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Command Tower", TypeLine: "Land",
		Owner: me.ID, Controller: me.ID,
	})
	subtypes := effectstest.EffectiveSubtypes(t, g, land)
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

	if err := effectstest.B02cPlayLandRaw(t, g, "Forest"); err != nil {
		t.Fatalf("the turn's ordinary land drop: %v", err)
	}
	if err := effectstest.B02cPlayLandRaw(t, g, "Island"); err != nil {
		t.Fatalf("Dryad's extra land drop was refused: %v", err)
	}
}

// --- straightforward spells and permanents -----------------------------

func TestB02cDovinsVetoCountersANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	victim := effectstest.CastCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
	spec, ok := effects.Lookup(effectstest.B02cDovinsVetoOracle)
	if !ok || !spec.CantBeCountered {
		t.Fatalf("Dovin's Veto must declare CantBeCountered")
	}
	effectstest.CastCatalogSpell(t, g, "Dovin's Veto", "Instant", effectstest.B02cDovinsVetoOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(victim) {
		t.Error("the targeted noncreature spell should have been countered into its owner's graveyard")
	}
}

func TestB02cMorbidOpportunistTriggersOnceEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedPermanentWithOracle(g, me.ID, "Morbid Opportunist", "Creature — Human Rogue", effectstest.B02cMorbidOpportunistOracle)
	first := effectstest.PushVanillaCreature(g, opp.ID, "First Victim", 2, 2)
	second := effectstest.PushVanillaCreature(g, opp.ID, "Second Victim", 2, 2)
	handBefore := me.Hand.Size()

	ctx := effectstest.CtxFor(g, &game.StackItem{Controller: me.ID})
	if err := (effects.DestroyTarget{Target: first}).Apply(ctx); err != nil {
		t.Fatalf("destroy first: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != handBefore+1 {
		t.Fatalf("hand = %d, want %d (one draw for the first death)", me.Hand.Size(), handBefore+1)
	}

	if err := (effects.DestroyTarget{Target: second}).Apply(ctx); err != nil {
		t.Fatalf("destroy second: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != handBefore+1 {
		t.Fatalf("hand = %d, want %d — the ability triggers only once each turn", me.Hand.Size(), handBefore+1)
	}
}

func TestB02cEverflowingChaliceEntersWithACounterPerKickAndScales(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	id := effectstest.PushCatalogHandCard(me, "Everflowing Chalice", "Artifact", effectstest.B02cEverflowingChaliceOracle)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Strict:        true,
		OptionalCosts: []int{0, 0},
	}); err != nil {
		t.Fatalf("CastSpell with multikicker x2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := g.LookupCardForEffect(id)
	if !ok || c.Counters[game.CounterCharge] != 2 {
		t.Fatalf("charge counters = %v, want 2", c.Counters)
	}
	me.ManaPool.EmptyPool()
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if colors := effectstest.Batch01PoolColors(me); len(colors) != 2 {
		t.Errorf("pool %v, want two colourless mana (one per charge counter)", colors)
	}
}

func TestB02cBloomTenderAddsOneManaPerColourControlled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tender := effectstest.SeedPermanentWithOracle(g, me.ID, "Bloom Tender", "Creature — Elf Druid", effectstest.B02cBloomTenderOracle)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Red Guy", TypeLine: "Creature — Human",
		Colors: []string{"R"}, Owner: me.ID, Controller: me.ID,
	})
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Green Guy", TypeLine: "Creature — Human",
		Colors: []string{"G"}, Owner: me.ID, Controller: me.ID,
	})
	if err := g.ActivateManaAbility(me.ID, tender, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	colors := effectstest.Batch01PoolColors(me)
	if len(colors) != 2 {
		t.Fatalf("pool %v, want [G R] (Bloom Tender is green itself)", colors)
	}
}

func TestB02cUntimelyMalfunctionDestroysTargetArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	art := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Rock", TypeLine: "Artifact",
		Owner: me.ID, Controller: me.ID,
	})
	effectstest.B02cCastModalSpell(t, g, "Untimely Malfunction", "Instant", effectstest.B02cUntimelyMalfunctionOracle,
		[]int{0}, []game.TargetRef{{Kind: game.TargetCard, ID: art}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(art) {
		t.Error("the artifact was not destroyed")
	}
}

func TestB02cUntimelyMalfunctionChangesTheTargetOfASingleTargetSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victimA, victimB := g.Seats[0].ID, g.Seats[1].ID, g.Seats[2].ID
	bolt := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victimA}})
	effectstest.B02cCastModalSpell(t, g, "Untimely Malfunction", "Instant", effectstest.B02cUntimelyMalfunctionOracle,
		[]int{1}, []game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	effectstest.PassPriorityAroundTable(t, g)

	prompt := effectstest.LatestRetarget(g, me)
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

func TestB02cUntimelyMalfunctionRedirectsAnAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	shredder := effectstest.B12Push(g, opp.ID, "Codex Shredder", "Artifact", effectstest.B22CodexShredderOracle, 0, 0)
	if err := g.ActivateCatalogAbility(opp.ID, shredder, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	ability := effectstest.AcItemLabelled(g, "{T}: Target player mills a card.")
	if ability == uuid.Nil {
		t.Fatal("the activation is not on the stack")
	}
	mine, theirs := me.Library.Size(), third.Library.Size()

	effectstest.B02cCastModalSpell(t, g, "Untimely Malfunction", "Instant", effectstest.B02cUntimelyMalfunctionOracle,
		[]int{1}, []game.TargetRef{{Kind: game.TargetCard, ID: ability}})
	effectstest.PassPriorityAroundTable(t, g)

	prompt := effectstest.LatestRetarget(g, me.ID)
	if prompt == nil {
		t.Fatal("Untimely Malfunction opened no retarget prompt for the ability")
	}
	if err := g.ResolveRetarget(prompt.ID, me.ID,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: third.ID}}); err != nil {
		t.Fatalf("ResolveRetarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != mine {
		t.Error("the mill still landed on the original target")
	}
	if third.Library.Size() != theirs-1 {
		t.Errorf("the redirected ability did not mill the new target: %d -> %d", theirs, third.Library.Size())
	}
}

func TestB02cUntimelyMalfunctionPreventsBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	blocker := effectstest.PushVanillaCreature(g, me.ID, "Their Blocker", 2, 2)
	effectstest.B02cCastModalSpell(t, g, "Untimely Malfunction", "Instant", effectstest.B02cUntimelyMalfunctionOracle,
		[]int{2}, []game.TargetRef{{Kind: game.TargetCard, ID: blocker}})
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := g.LookupCardForEffect(blocker)
	if !ok || !game.Restricted(&c, game.CantBlock) {
		t.Error("the targeted creature must not be able to block this turn")
	}
}

func TestB02cFaerieMastermindDrawsOffAnOpponentsSecondDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedPermanentWithOracle(g, me.ID, "Faerie Mastermind", "Creature — Faerie Rogue", effectstest.B02cFaerieMastermindOracle)
	effectstest.PushLibraryCardForTest(opp, game.Card{InstanceID: uuid.New(), Name: "Filler 1", TypeLine: "Land"})
	effectstest.PushLibraryCardForTest(opp, game.Card{InstanceID: uuid.New(), Name: "Filler 2", TypeLine: "Land"})
	handBefore := me.Hand.Size()

	if err := g.DrawCard(opp.ID); err != nil {
		t.Fatalf("first draw: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != handBefore {
		t.Fatalf("hand = %d, want %d — the FIRST draw of the turn must not trigger", me.Hand.Size(), handBefore)
	}
	if err := g.DrawCard(opp.ID); err != nil {
		t.Fatalf("second draw: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != handBefore+1 {
		t.Fatalf("hand = %d, want %d — the opponent's SECOND draw must trigger", me.Hand.Size(), handBefore+1)
	}
}

func TestB02cInspiringCallDrawsPerCounteredCreatureAndGrantsIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.PushVanillaCreature(g, me.ID, "Counted A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "Uncounted B", 2, 2)
	if err := g.AddCounterForEffect(a, game.CounterPlusOne, 1); err != nil {
		t.Fatalf("AddCounterForEffect: %v", err)
	}
	handBefore := me.Hand.Size()

	effectstest.CastCatalogSpell(t, g, "Inspiring Call", "Instant", effectstest.B02cInspiringCallOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if me.Hand.Size() != handBefore+1 {
		t.Fatalf("hand = %d, want %d (one counted creature)", me.Hand.Size(), handBefore+1)
	}
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, a), "indestructible") {
		t.Errorf("the countered creature must gain indestructible")
	}
	if effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, b), "indestructible") {
		t.Errorf("the UNcountered creature must not gain indestructible")
	}
}

func TestB02cGrandAbolisherLocksOutOpponentsDuringYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedPermanentWithOracle(g, me.ID, "Grand Abolisher", "Creature — Human Cleric", effectstest.B02cGrandAbolisherOracle)
	signet := effectstest.PushCatalogPermanent(g, opp.ID, "Boros Signet", "Artifact", "41c84665-1f99-40ab-aaca-1188649eb263", false)
	effectstest.ToMainForCost(t, g)
	oppBolt := effectstest.HandCardFull(opp, "Their Bolt", "Instant", "{R}", "", []string{"R"})
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	pitch := effectstest.HandCardFull(opp, "Their Blue Card", "Instant", "{U}", "", []string{"U"})
	fon := effectstest.HandCardFull(opp, "Force of Negation", "Instant", "{1}{U}{U}", effectstest.B02cForceOfNegationOracle, []string{"U"})
	victim := effectstest.CastCatalogSpell(t, g, "My Sorcery", "Sorcery", "", nil)

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
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(victim) {
		t.Error("the countered spell must not land in its owner's graveyard — Force of Negation exiles it")
	}
	if !g.Exile.Contains(victim) {
		t.Error("the countered spell must be exiled (#1273: CounterTarget.Dest, ZoneExile)")
	}
}

func TestB02cEtaliExilesEachLibraryAndOffersFreeCasts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	etali := effectstest.PushDiesCreatureForTest(g, me.ID, "Etali, Primal Storm", effectstest.B02cEtaliOracle, "Legendary Creature — Elder Dinosaur", 6, 6)
	effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "My Top", TypeLine: "Sorcery", ManaCost: "{1}{R}"})
	effectstest.PushLibraryCardForTest(opp, game.Card{InstanceID: uuid.New(), Name: "Their Top", TypeLine: "Instant", ManaCost: "{U}"})

	// attackWith declares the attacker AND advances into the combat
	// damage step, which is what actually locks the declaration in
	// and emits EventAttack (commitAttackDeclarationLocked, at the
	// step's first priority boundary) — a bare DeclareAttacker only
	// stages it.
	effectstest.AttackWith(t, g, opp.ID, etali)
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushVanillaCreature(g, me.ID, "Big Guy", 5, 5)
	cheap := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: cheap, Name: "Cheap Spell", TypeLine: "Sorcery", ManaCost: "{3}",
		Owner: me.ID, Controller: me.ID,
	})
	handBefore := me.Hand.Size()

	effectstest.CastCatalogSpell(t, g, "Rishkar's Expertise", "Sorcery", effectstest.B02cRishkarsExpertiseOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	// castCatalogSpell both adds Rishkar's Expertise to the hand and
	// casts it, netting zero; resolution then draws 5 (the greatest
	// power among creatures controlled).
	if me.Hand.Size() != handBefore+5 {
		t.Fatalf("hand = %d, want %d (drew 5)", me.Hand.Size(), handBefore+5)
	}
	// #4250: the free cast is offered inside the resolution.
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, cheap)
	perm := g.CastPermissionOnCardByIDForEffect(cheap)
	if perm == nil || perm.Cost != "{0}" {
		t.Errorf("the offered card must carry a free-cast permission: %+v", perm)
	}
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, cheap, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the free cast: %v", err)
	}
	effectstest.CastInsideTheResolution(t, g, cheap)
}

func TestB02cSevinnesReclamationReanimatesAndFlashbackOffersACopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	bear := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: bear, Name: "Graveyard Bear", TypeLine: "Creature — Bear",
		ManaCost: "{2}", Owner: me.ID, Controller: me.ID,
	})
	sevinne := effectstest.HandCardFull(me, "Sevinne's Reclamation", "Sorcery", "{2}{W}", effectstest.B02cSevinnesReclamationOracle, []string{"W"})
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.CastSpell(me.ID, sevinne, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatalf("the permanent card was not returned to the battlefield")
	}
}

func TestB02cTakenumaChannelsMillsAndReturnsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogHandCard(me, "Takenuma, Abandoned Mire", "Legendary Land", effectstest.B02cTakenumaOracle)
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
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != libBefore-3 {
		t.Errorf("library = %d, want %d (milled three)", me.Library.Size(), libBefore-3)
	}
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
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
