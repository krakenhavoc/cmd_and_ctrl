package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Molten
// Tributary is a row in the Dominaria United dual table, so a
// transposed row is invisible until someone plays that exact card.
// Goblin Chieftain was already on main from the S26 tribal lords and
// is pinned here so the table matches the issue's 40.
func TestBatch24CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.B24MariOracle:                  "Mari, the Killing Quill",
		effectstest.B24DebtToTheDeathlessOracle:    "Debt to the Deathless",
		effectstest.B24NaturesWillOracle:           "Nature's Will",
		effectstest.B24AgnaQelaOracle:              "Agna Qel'a",
		effectstest.B24IndulgentAristocratOracle:   "Indulgent Aristocrat",
		effectstest.B24MerchantScrollOracle:        "Merchant Scroll",
		effectstest.B24SpineOfIshSahOracle:         "Spine of Ish Sah",
		effectstest.B24GoblinChieftainAlreadyOID:   "Goblin Chieftain",
		effectstest.B24WitchOfTheMoorsOracle:       "Witch of the Moors",
		effectstest.B24BugenhagenOracle:            "Bugenhagen, Wise Elder",
		effectstest.B24KaervekOracle:               "Kaervek the Merciless",
		effectstest.B24ScourForScrapOracle:         "Scour for Scrap",
		effectstest.B24AjanisWelcomeOracle:         "Ajani's Welcome",
		effectstest.B24MegrimOracle:                "Megrim",
		effectstest.B24SceneOfTheCrimeOracle:       "Scene of the Crime",
		effectstest.B24SurvivorsEncampmentOracle:   "Survivors' Encampment",
		effectstest.B24SoulsFireOracle:             "Soul's Fire",
		effectstest.B24AjanisPridemateOracle:       "Ajani's Pridemate",
		effectstest.B24LichKnightsConquestOracle:   "Lich-Knights' Conquest",
		effectstest.B24VraskaJoinsUpOracle:         "Vraska Joins Up",
		effectstest.B24ChulaneOracle:               "Chulane, Teller of Tales",
		effectstest.B24BloodSeekerOracle:           "Blood Seeker",
		effectstest.B24BurglarRatOracle:            "Burglar Rat",
		effectstest.B24FinalPartingOracle:          "Final Parting",
		effectstest.B24WillOfTheAbzanOracle:        "Will of the Abzan",
		effectstest.B24MoltenTributaryOracle:       "Molten Tributary",
		effectstest.B24ThroneOfTheGodPharaohOracle: "Throne of the God-Pharaoh",
		effectstest.B24WrathfulRaptorsOracle:       "Wrathful Raptors",
		effectstest.B24EfficientConstructionOracle: "Efficient Construction",
		effectstest.B24TectonicHazardOracle:        "Tectonic Hazard",
		effectstest.B24DwynenOracle:                "Dwynen, Gilt-Leaf Daen",
		effectstest.B24YshtolaNightsBlessedOracle:  "Y'shtola, Night's Blessed",
		effectstest.B24LinvalaOracle:               "Linvala, Keeper of Silence",
		effectstest.B24CursedTotemOracle:           "Cursed Totem",
	}
	if len(want) != 34 {
		t.Fatalf("the batch registers 33 cards plus Goblin Chieftain, the table lists %d", len(want))
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
	// The remaining declared skips must NOT be registered —
	// each needs a trigger mode, a cost or a gate the engine cannot
	// express, and a spec would ship the card stronger than printed.
	// Dread Presence came off with the modal trigger (#5948). Linvala and Cursed Totem came off this list with #1210's
	// activation gate and are in `want` above, and Vat of Rebirth came
	// off with the remove-counters cost (#2952). Ascend from Avernus came off with
	// a spell's own exile step as it resolves (Reunion of the House's, #5948).
	// Khalni Heart Expedition came off this list with the remove-counters
	// cost (GAP2 b01, #5948).
}

func TestB24SurvivorsEncampmentPaysItsCreatureTapCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.PlayLandFromHand(t, g, "Survivors' Encampment", effectstest.B24SurvivorsEncampmentOracle)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Fresh Bear", 2, 2)

	if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{TapIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if !effectstest.B20Tapped(t, g, land) || !effectstest.B20Tapped(t, g, bear) {
		t.Error("the land and the chosen creature both pay the ability")
	}
	pick := effectstest.RiderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("mana choice = %+v, want five colors", pick)
	}
}

// --- the lands -----------------------------------------------------

func TestB24MoltenTributaryEntersTappedAndTapsForBlueOrRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.PlayLandFromHand(t, g, "Molten Tributary", effectstest.B24MoltenTributaryOracle)
	if !effectstest.B20Tapped(t, g, land) {
		t.Fatal("enters tapped")
	}
	effectstest.B22Untap(g, land)
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := effectstest.RiderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 2 || pick.ColorOptions[0] != "U" || pick.ColorOptions[1] != "R" {
		t.Fatalf("{U} or {R}, got %+v", pick)
	}
}

func TestB24AgnaQelaEntersTappedWithoutABasicAndLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	first := effectstest.PlayLandFromHand(t, g, "Agna Qel'a", effectstest.B24AgnaQelaOracle)
	if !effectstest.B20Tapped(t, g, first) {
		t.Fatal("with no basic land it enters tapped")
	}
	effectstest.SeedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	second := effectstest.PlayLandFromHand(t, g, "Agna Qel'a", effectstest.B24AgnaQelaOracle)
	if effectstest.B20Tapped(t, g, second) {
		t.Fatal("with a basic land it enters untapped")
	}
	if err := g.ActivateManaAbility(me.ID, second, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "U" {
		t.Errorf("pool %v, want [U]", got)
	}
	me.ManaPool.EmptyPool()
	effectstest.B22Untap(g, second)
	effectstest.B06AddMana(me, "U", "C", "C")
	hand := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, second, 0, game.ActivateAbilityParams{})
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("drew %d, want 1 before the discard", got-hand)
	}
	if effectstest.DiscardOwed(g, me.ID) != 1 {
		t.Errorf("then discards a card: owed %d", effectstest.DiscardOwed(g, me.ID))
	}
	if !effectstest.B20Tapped(t, g, second) {
		t.Error("the loot has a tap cost")
	}
}

func TestB24SceneOfTheCrimeEntersTappedTapsForManaAndCracks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	scene := effectstest.PlayLandFromHand(t, g, "Scene of the Crime", effectstest.B24SceneOfTheCrimeOracle)
	if !effectstest.B20Tapped(t, g, scene) {
		t.Fatal("enters tapped")
	}
	effectstest.B22Untap(g, scene)
	if err := g.ActivateManaAbility(me.ID, scene, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
	me.ManaPool.EmptyPool()
	effectstest.B22Untap(g, scene)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Fresh Bear", 2, 2)
	if err := g.ActivateManaAbility(me.ID, scene, 1, game.ManaAbilityParams{TapIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("ActivateManaAbility colored: %v", err)
	}
	if !effectstest.B20Tapped(t, g, scene) || !effectstest.B20Tapped(t, g, bear) {
		t.Error("the land and the chosen creature both pay the colored ability")
	}
	pick := effectstest.RiderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("colored ability choice = %+v, want five colors", pick)
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "G"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool %v, want [G]", got)
	}
	me.ManaPool.EmptyPool()
	// The crack needs no tap: a tapped Scene can still be sacrificed.
	effectstest.B06AddMana(me, "C", "C")
	hand := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, scene, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(scene) {
		t.Error("sacrificed as the cost")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("drew %d, want 1", got-hand)
	}
	if spec, _ := effects.Lookup(effectstest.B24SceneOfTheCrimeOracle); spec.Completeness != effects.CompletenessFull || len(spec.ManaAbilities) != 2 {
		t.Error("Scene should be full with both printed mana abilities")
	}
}

// --- the spells ----------------------------------------------------

func TestB24DebtToTheDeathlessDrainsTwiceX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := effectstest.B24Lives(g)
	effectstest.CastXSpell(t, g, "Debt to the Deathless", "Sorcery", effectstest.B24DebtToTheDeathlessOracle, "{X}{W}{W}{B}{B}", 3, nil)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Elf", "Creature — Elf", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	theirBear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	before := effectstest.B24Lives(g)
	effectstest.CastCatalogSpell(t, g, "Tectonic Hazard", "Sorcery", effectstest.B24TectonicHazardOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
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
	if effectstest.DamageMarkedOn(g, theirBear) != 1 {
		t.Errorf("their 2/2 took 1, got %d", effectstest.DamageMarkedOn(g, theirBear))
	}
}

func TestB24MerchantScrollFindsABlueInstant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Counterspell", TypeLine: "Instant", ManaCost: "{U}{U}", Colors: []string{"U"}},
		game.Card{Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", Colors: []string{"R"}},
		game.Card{Name: "Ponder", TypeLine: "Sorcery", ManaCost: "{U}", Colors: []string{"U"}},
	)
	effectstest.CastCatalogSpell(t, g, "Merchant Scroll", "Sorcery", effectstest.B24MerchantScrollOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	if !me.Hand.Contains(ids[0]) {
		t.Error("the one blue instant is found without a prompt")
	}
	if me.Hand.Contains(ids[1]) || me.Hand.Contains(ids[2]) {
		t.Error("a red instant and a blue sorcery are not blue instants")
	}
}

func TestB24FinalPartingTutorsOneToHandAndOneToGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "A", TypeLine: "Instant"},
		game.Card{Name: "B", TypeLine: "Creature — Bear"},
		game.Card{Name: "C", TypeLine: "Land"},
	)
	effectstest.CastCatalogSpell(t, g, "Final Parting", "Sorcery", effectstest.B24FinalPartingOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	first := effectstest.SearchChoiceFor(g, me.ID)
	if first == nil || len(first.SearchCards) != 3 || first.SearchMax != 1 {
		t.Fatalf("the first search offers the whole library for one card, got %+v", first)
	}
	effectstest.AnswerSearchByID(t, g, me.ID, ids[0])
	if !me.Hand.Contains(ids[0]) {
		t.Fatal("the first pick goes to hand")
	}
	second := effectstest.SearchChoiceFor(g, me.ID)
	if second == nil || len(second.SearchCards) != 2 || second.SearchMax != 1 {
		t.Fatalf("the second search offers the rest for one card, got %+v", second)
	}
	effectstest.AnswerSearchByID(t, g, me.ID, ids[1])
	if !me.Graveyard.Contains(ids[1]) {
		t.Error("the second pick goes to the graveyard")
	}
	if me.Library.Size() != 1 || !me.Library.Contains(ids[2]) {
		t.Error("the third card stays in the library")
	}
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Error("two cards, two prompts, no third")
	}
}

func TestB24ScourForScrapTutorsAndReturnsAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lib := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Sol Ring", TypeLine: "Artifact"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	dead := effectstest.B17GraveyardCard(me, "Dead Signet", "Artifact", "{2}")
	deadBear := effectstest.B17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	if err := effectstest.B22TryModal(g, effectstest.B24ScourForScrapOracle, []int{1}, effectstest.B16TargetCard(deadBear)); err == nil {
		t.Fatal("a creature card is not an artifact card")
	}
	effectstest.CastModal(t, g, "Scour for Scrap", "Instant", effectstest.B24ScourForScrapOracle, []int{0, 1}, effectstest.B16TargetCard(dead))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Beast", "Creature — Beast", 4, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	wall := effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 6)
	effectstest.CastCatalogSpell(t, g, "Soul's Fire", "Instant", effectstest.B24SoulsFireOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("4 damage from the Beast kills the Bear")
	}
	// #764: the first slot is its own CLAUSE, so an opponent's
	// creature there is refused at ANNOUNCE (CR 601.2c) rather than
	// discovered at resolution. Back the clause list out to one wide
	// spec and this cast is accepted and quietly does nothing.
	if err := effectstest.CastCatalogSpellErr(t, g, "Soul's Fire", "Instant", effectstest.B24SoulsFireOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: wall}, {Kind: game.TargetCard, ID: mine}}); err != game.ErrIllegalTarget {
		t.Errorf("their Wall in slot 0: err %v, want ErrIllegalTarget", err)
	}
	if effectstest.DamageMarkedOn(g, mine) != 0 {
		t.Error("a refused announcement deals nothing")
	}
	// A player in the second slot: face damage equal to power.
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Soul's Fire", "Instant", effectstest.B24SoulsFireOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("life %d → %d, want -4", life, opp.Life)
	}
}

func TestB24LichKnightsConquestSacrificesAnyNumberThenChoosesThatMany(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	deadA := effectstest.B17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	deadC := effectstest.B17GraveyardCard(me, "Dead Wurm", "Creature — Wurm", "{5}{G}")
	treasure := effectstest.PushToken(g, me.ID, effects.TreasureToken())
	signet := effectstest.B12Permanent(g, me.ID, "Signet", "Artifact")
	aura := effectstest.B12Permanent(g, me.ID, "Rancor", "Enchantment — Aura")
	golem := effectstest.B12Creature(g, me.ID, "Living Golem", "Artifact Creature — Golem", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Living Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Lich-Knights' Conquest", "Sorcery", effectstest.B24LichKnightsConquestOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	// #2863: "sacrifice any number" is one pick on resolution over the
	// artifacts, enchantments and tokens, with a floor of zero.
	sac := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if sac == nil {
		t.Fatalf("the caster chooses what to sacrifice as the spell resolves: %+v", g.PendingChoices)
	}
	if sac.ChooseMin != 0 || sac.ChooseMax != 4 {
		t.Errorf("ANY NUMBER of the four: bounds %d..%d", sac.ChooseMin, sac.ChooseMax)
	}
	for _, id := range []uuid.UUID{treasure, signet, aura, golem} {
		if !effectstest.HasID(sac.ChooseCards, id) {
			t.Errorf("%s is an artifact, enchantment or token and is on offer", id)
		}
	}
	if effectstest.HasID(sac.ChooseCards, bear) {
		t.Error("a nontoken creature that is not an artifact is not on offer")
	}
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("the creature cards were offered before the sacrifice")
	}
	if err := g.ResolveOwnPermanents(sac.ID, me.ID, []uuid.UUID{treasure, golem}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	if g.Battlefield.Contains(treasure) || g.Battlefield.Contains(golem) ||
		!g.Battlefield.Contains(signet) || !g.Battlefield.Contains(aura) {
		t.Error("exactly the two chosen permanents are sacrificed")
	}

	// "Return that many" is chosen now, from the graveyard as it is
	// after the sacrifice: the Golem just sacrificed is an artifact
	// creature CARD and is on offer; the Treasure was a token and is
	// gone.
	ret := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if ret == nil {
		t.Fatal("the caster chooses the creature cards on resolution")
	}
	if ret.ChooseMin != 2 || ret.ChooseMax != 2 {
		t.Errorf("THAT MANY is two, exactly: bounds %d..%d", ret.ChooseMin, ret.ChooseMax)
	}
	if !effectstest.HasID(ret.ChooseCards, golem) || !effectstest.HasID(ret.ChooseCards, deadA) || !effectstest.HasID(ret.ChooseCards, deadC) {
		t.Errorf("offer %v: want both dead creatures and the sacrificed Golem", ret.ChooseCards)
	}
	if effectstest.HasID(ret.ChooseCards, treasure) {
		t.Error("a sacrificed token is not a card in the graveyard")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, golem, deadA)
	if !g.Battlefield.Contains(golem) || !g.Battlefield.Contains(deadA) {
		t.Fatal("both chosen creature cards return, the sacrificed Golem among them")
	}
	if g.Battlefield.Contains(deadC) {
		t.Error("only the chosen cards return")
	}
}

// TestB24LichKnightsConquestSacrificingNothingReturnsNothing — a floor
// of zero is a real answer, and "that many" is then zero.
func TestB24LichKnightsConquestSacrificingNothingReturnsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dead := effectstest.B17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	signet := effectstest.B12Permanent(g, me.ID, "Signet", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Lich-Knights' Conquest", "Sorcery", effectstest.B24LichKnightsConquestOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	sac := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if sac == nil {
		t.Fatal("no sacrifice prompt")
	}
	if err := g.ResolveOwnPermanents(sac.ID, me.ID, nil); err != nil {
		t.Fatalf("ResolveOwnPermanents(empty): %v", err)
	}
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil || g.Battlefield.Contains(dead) || !g.Battlefield.Contains(signet) {
		t.Error("nothing sacrificed, nothing returned")
	}
}

// TestB24LichKnightsConquestReturnsAllWhenThereAreTooFew — three
// sacrificed, one creature card: the one comes back (CR 608.2).
func TestB24LichKnightsConquestReturnsAllWhenThereAreTooFew(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dead := effectstest.B17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	a := effectstest.PushToken(g, me.ID, effects.TreasureToken())
	b := effectstest.PushToken(g, me.ID, effects.TreasureToken())
	c := effectstest.B12Permanent(g, me.ID, "Signet", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Lich-Knights' Conquest", "Sorcery", effectstest.B24LichKnightsConquestOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	sac := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if sac == nil {
		t.Fatal("no sacrifice prompt")
	}
	if err := g.ResolveOwnPermanents(sac.ID, me.ID, []uuid.UUID{a, b, c}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	ret := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if ret == nil || ret.ChooseMin != 1 || ret.ChooseMax != 1 {
		t.Fatalf("one creature card on offer, and it must be taken: %+v", ret)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, dead)
	if !g.Battlefield.Contains(dead) {
		t.Error("the only creature card returns")
	}
}

func TestB24WillOfTheAbzanEdictsTheBiggestOrReanimates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	aBig := effectstest.B12Creature(g, a.ID, "A's Giant", "Creature — Giant", 5, 5)
	aSmall := effectstest.B12Creature(g, a.ID, "A's Bear", "Creature — Bear", 2, 2)
	bOne := effectstest.B12Creature(g, b.ID, "B's Bear", "Creature — Bear", 2, 2)
	bTwo := effectstest.B12Creature(g, b.ID, "B's Other Bear", "Creature — Bear", 2, 2)
	dead := effectstest.B17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	if err := effectstest.B22TryModal(g, effectstest.B24WillOfTheAbzanOracle, []int{0, 1},
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: a.ID}, {Kind: game.TargetCard, ID: dead}}); err == nil {
		t.Fatal("no commander: choose both is refused (#1590) — one mode only")
	}
	aLife, bLife := a.Life, b.Life
	effectstest.CastModal(t, g, "Will of the Abzan", "Sorcery", effectstest.B24WillOfTheAbzanOracle, []int{0},
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: a.ID}, {Kind: game.TargetPlayer, ID: b.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	// #1019: the loss is a clause about a PLAYER, so it happens either
	// way — but it happens AFTER the sacrifice it is printed after.
	if a.Life != aLife || b.Life != bLife {
		t.Errorf("the life loss landed while the sacrifice prompts were open: %d→%d, %d→%d",
			aLife, a.Life, bLife, b.Life)
	}
	// A's Giant is the only greatest-power creature, so A's pick is
	// forced and collected unasked (ADR 0175, #3664); it waits for B.
	if pa := effectstest.SacrificeChoiceFor(g, a.ID); pa != nil {
		t.Errorf("A has no choice to make and is asked anyway: %+v", pa)
	}
	if !g.Battlefield.Contains(aBig) {
		t.Error("A's Giant was sacrificed before B had chosen")
	}
	pb := effectstest.SacrificeChoiceFor(g, b.ID)
	if pb == nil || len(effectstest.SacrificeOptions(pb)) != 2 {
		t.Fatalf("B's two Bears tie for greatest and B chooses: %+v", pb)
	}
	effectstest.AnswerSacrifice(t, g, b.ID, bTwo)
	if g.Battlefield.Contains(aBig) || !g.Battlefield.Contains(aSmall) || !g.Battlefield.Contains(bOne) || g.Battlefield.Contains(bTwo) {
		t.Error("exactly the chosen creatures are sacrificed")
	}
	if a.Life != aLife-3 || b.Life != bLife-3 {
		t.Errorf("each targeted opponent loses 3 once they have answered: %d→%d, %d→%d",
			aLife, a.Life, bLife, b.Life)
	}
	// Mode two: reanimate.
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.CastModal(t, g, "Will of the Abzan", "Sorcery", effectstest.B24WillOfTheAbzanOracle, []int{1}, effectstest.B16TargetCard(dead))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) {
		t.Error("the creature card returns to the battlefield")
	}
	if c, _ := effectstest.BattlefieldCard(g, dead); c.Controller != me.ID {
		t.Error("under its owner's — your — control")
	}
}

// TestB24WillOfTheAbzanFirstModeAcceptsZeroTargets is #1654: "any
// number of target opponents" is CR 115.1d, so zero is a legal
// choice, not a minimum-one clause. The cast must be accepted and
// resolve without error or any sacrifice prompt.
func TestB24WillOfTheAbzanFirstModeAcceptsZeroTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	a, b := g.Seats[1], g.Seats[2]
	effectstest.B12Creature(g, a.ID, "A's Giant", "Creature — Giant", 5, 5)
	effectstest.B12Creature(g, b.ID, "B's Bear", "Creature — Bear", 2, 2)
	aLife, bLife := a.Life, b.Life

	effectstest.CastModal(t, g, "Will of the Abzan", "Sorcery", effectstest.B24WillOfTheAbzanOracle, []int{0}, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if a.Life != aLife || b.Life != bLife {
		t.Errorf("nobody was targeted, so nobody loses life: %d→%d, %d→%d", aLife, a.Life, bLife, b.Life)
	}
	if effectstest.SacrificeChoiceFor(g, a.ID) != nil || effectstest.SacrificeChoiceFor(g, b.ID) != nil {
		t.Error("nobody was targeted, so nobody is prompted to sacrifice")
	}
}

// --- the triggers --------------------------------------------------

func TestB24NaturesWillTapsTheirLandsUntapsYoursOncePerPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	will := effectstest.B12Push(g, me.ID, "Nature's Will", "Enchantment", effectstest.B24NaturesWillOracle, 0, 0)
	myLand := effectstest.SeedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
	b08Tap(g, myLand)
	aLand := effectstest.SeedLand(g, a.ID, "Island", "Basic Land — Island", "")
	bLand := effectstest.SeedLand(g, b.ID, "Swamp", "Basic Land — Swamp", "")
	bearA := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	bearB := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	bearC := effectstest.B12Creature(g, me.ID, "Bear C", "Creature — Bear", 2, 2)
	// One declaration split across two defenders (CR 508.1): the
	// lock-in announces all three together (#859).
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, d := range []struct {
		attacker uuid.UUID
		defender uuid.UUID
	}{{bearA, a.ID}, {bearB, a.ID}, {bearC, b.ID}} {
		if err := g.DeclareAttacker(d.attacker, d.defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	if n := len(g.PendingTriggers) + effectstest.TriggersOnStackFrom(g, will); n != 2 {
		t.Fatalf("two bears on A and one on B: one trigger per player, got %d", n)
	}
	// Two differing triggers: their controller orders them (CR 603.3b).
	effectstest.PassPriorityUntilTriggerOrderPrompt(t, g)
	effectstest.AnswerTriggerOrderInOfferedOrder(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B20Tapped(t, g, aLand) || !effectstest.B20Tapped(t, g, bLand) {
		t.Error("both damaged players' lands are tapped")
	}
	if effectstest.B20Tapped(t, g, myLand) {
		t.Error("your lands untap")
	}
}

func TestB24ThroneOfTheGodPharaohDrainsForTappedCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Throne of the God-Pharaoh", "Legendary Artifact", effectstest.B24ThroneOfTheGodPharaohOracle, 0, 0)
	tappedA := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	tappedB := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Untapped Bear", "Creature — Bear", 2, 2)
	b08Tap(g, tappedA)
	b08Tap(g, tappedB)
	before := effectstest.B24Lives(g)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
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
	before = effectstest.B24Lives(g)
	effectstest.AdvanceToEndStepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.Life != before[i] {
			t.Errorf("seat %d changed on an opponent's end step", i)
		}
	}
}

func TestB24AjanisWelcomeAndPridemateGrowTogether(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Ajani's Welcome", "Enchantment", effectstest.B24AjanisWelcomeOracle, 0, 0)
	pridemate := effectstest.B12Push(g, me.ID, "Ajani's Pridemate", "Creature — Cat Soldier", effectstest.B24AjanisPridemateOracle, 2, 2)
	life := me.Life
	effectstest.CastAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Fatalf("a creature entering under your control gains 1: %d → %d", life, me.Life)
	}
	if got := effectstest.CounterCount(g, pridemate, "+1/+1"); got != 1 {
		t.Errorf("the gain grows the Pridemate: %d counters, want 1", got)
	}
	// An opponent's creature entering is not yours.
	effectstest.B13PlayAs(t, g, 1, "Their Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 || effectstest.CounterCount(g, pridemate, "+1/+1") != 1 {
		t.Error("an opponent's creature gains nothing")
	}
	_ = opp
}

func TestB24BurglarRatMakesOpponentsDiscardAndMegrimPunishes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Megrim", "Enchantment", effectstest.B24MegrimOracle, 0, 0)
	effectstest.CastAndResolveCreature(t, g, "Burglar Rat", "Creature — Rat", effectstest.B24BurglarRatOracle)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		owed := effectstest.DiscardOwed(g, p.ID)
		if i == 0 && owed != 0 {
			t.Errorf("you owe nothing, got %d", owed)
		}
		if i != 0 && owed != 1 {
			t.Errorf("seat %d owes one discard, got %d", i, owed)
		}
	}
	life := opp.Life
	effectstest.AnswerDiscard(t, g, opp.ID, opp.Hand.Cards[0].InstanceID)
	// Every opponent was asked; the table is gated until all three
	// have answered, so Megrim's trigger cannot reach the stack until
	// then (#651).
	for _, p := range g.Seats[2:] {
		effectstest.DiscardFromHand(t, g, p.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife {
		t.Error("Megrim does not punish its controller")
	}
}

func TestB24EfficientConstructionMakesAThopterPerArtifactSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Efficient Construction", "Enchantment", effectstest.B24EfficientConstructionOracle, 0, 0)
	effectstest.B24CastNoncreature(t, g, "Sol Ring", "Artifact", "{1}")
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Thopter"); n != 1 {
		t.Fatalf("an artifact spell makes a Thopter: %d", n)
	}
	effectstest.B24CastNoncreature(t, g, "Ponder", "Sorcery", "{U}")
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Thopter"); n != 1 {
		t.Errorf("a sorcery is not an artifact spell: %d", n)
	}
	thopter := effectstest.FindBattlefieldByName(g, "Thopter")
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, thopter), "flying") {
		t.Error("the Thopter flies")
	}
}

func TestB24BloodSeekerMayDrainTheEnteringCreaturesController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Blood Seeker", "Creature — Vampire Shaman", effectstest.B24BloodSeekerOracle, 1, 1)
	life := opp.Life
	effectstest.B13PlayAs(t, g, 1, "Their Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPromptFor(g, me.ID) == nil {
		t.Fatal("the Seeker's controller is asked")
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("that player loses 1: %d → %d", life, opp.Life)
	}
	// Declining costs nothing; your own creature asks nothing.
	effectstest.B13PlayAs(t, g, 1, "Their Second Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Error("no is no")
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.CastAndResolveCreature(t, g, "My Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPromptFor(g, me.ID) != nil {
		t.Error("your own creature entering is not an opponent's")
	}
}

func TestB24KaervekPunishesEveryOpponentsSpellForItsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kaervek := effectstest.B12Push(g, me.ID, "Kaervek the Merciless", "Legendary Creature — Human Shaman", effectstest.B24KaervekOracle, 5, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Golem", "Creature — Golem", 4, 4)
	effectstest.Batch01OpponentCasts(t, g, opp, "Big Spell", "", "{2}{U}{U}", nil)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil || p.Source != kaervek {
		t.Fatal("Kaervek's controller picks any target")
	}
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("four damage from a four-drop kills the 4/4")
	}
	// Your own spells are not an opponent's.
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.B24CastNoncreature(t, g, "My Spell", "Sorcery", "{2}{U}{U}")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("your own spell does not trigger Kaervek")
	}
}

func TestB24YshtolaDrawsOnBigLifeLossAndPunishesBigSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	yshtola := effectstest.B12Push(g, me.ID, "Y'shtola, Night's Blessed", "Legendary Creature — Cat Warlock", effectstest.B24YshtolaNightsBlessedOracle, 2, 4)
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, yshtola), "vigilance") {
		t.Error("printed vigilance")
	}
	// A two-drop is not mana value 3 or greater.
	before := effectstest.B24Lives(g)
	effectstest.B24CastNoncreature(t, g, "Small Spell", "Instant", "{1}{U}")
	effectstest.PassPriorityAroundTable(t, g)
	if lives := effectstest.B24Lives(g); lives[1] != before[1] || lives[0] != before[0] {
		t.Error("mana value 2 does nothing")
	}
	effectstest.B24CastNoncreature(t, g, "Big Spell", "Sorcery", "{1}{U}{B}")
	effectstest.PassPriorityAroundTable(t, g)
	lives := effectstest.B24Lives(g)
	for i := 1; i < len(lives); i++ {
		if lives[i] != before[i]-2 {
			t.Errorf("seat %d takes 2: %d → %d", i, before[i], lives[i])
		}
	}
	if lives[0] != before[0]+2 {
		t.Errorf("you gain 2 once: %d → %d", before[0], lives[0])
	}
	// A creature spell is not a noncreature spell.
	effectstest.CastAndResolveCreature(t, g, "Big Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != lives[1] {
		t.Error("a creature spell does not trigger")
	}
	// End step: an opponent lost only 2 this turn — no draw.
	hand := me.Hand.Size()
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatal("2 life lost is not 4")
	}
	// Next turn (an opponent's): Bolt an opponent twice — 6 lost —
	// and the end step draws for Y'shtola's controller.
	effectstest.AdvanceToPrecombatMainOf(t, g, 1)
	effectstest.B24CastNoncreature(t, g, "Filler", "Instant", "{U}")
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -4) })
	hand = me.Hand.Size()
	effectstest.AdvanceToEndStepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("a player lost 4 this turn — each end step, any player's — you draw: %d", got-hand)
	}
}

func TestB24BugenhagenDrawsForASevenPowerCreatureAndTapsForAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bugenhagen := effectstest.B12Push(g, me.ID, "Bugenhagen, Wise Elder", "Legendary Creature — Human Shaman", effectstest.B24BugenhagenOracle, 1, 3)
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, bugenhagen), "reach") {
		t.Error("printed reach")
	}
	if err := g.ActivateManaAbility(me.ID, bugenhagen, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if pick := effectstest.RiderLatestManaPick(g, me.ID); pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("any colour, the printed width, got %+v", pick)
	}
	effectstest.B10ResolveAllManaPicks(t, g, me.ID, "G")
	big := effectstest.B12Creature(g, me.ID, "Big Beast", "Creature — Beast", 6, 6)
	// The walk to the next upkeep passes a cleanup step, so the hand is
	// emptied first to keep the max-hand-size discard out of the count.
	effectstest.EmptyHandToLibrary(g, me)
	effectstest.AdvanceToPrecombatMainOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatal("a 6-power creature is not 7 or greater")
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(big, "+1/+1", 1) })
	effectstest.EmptyHandToLibrary(g, me)
	effectstest.AdvanceToPrecombatMainOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	hand = me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("a 7-power creature (counters count) draws: %d", got-hand)
	}
}

func TestB24WitchOfTheMoorsEdictsAndReturnsWhenYouGainedLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	witch := effectstest.B12Push(g, me.ID, "Witch of the Moors", "Creature — Human Warlock", effectstest.B24WitchOfTheMoorsOracle, 4, 4)
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, witch), "deathtouch") {
		t.Error("printed deathtouch")
	}
	aBear := effectstest.B12Creature(g, a.ID, "A's Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, b.ID, "B's Bear", "Creature — Bear", 2, 2)
	dead := effectstest.B17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	// No life gained: no trigger.
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SacrificeChoiceFor(g, a.ID) != nil || effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("without lifegain the Witch is quiet")
	}
	// Gained life, a creature card in the graveyard: the targeted
	// twin fires.
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1) })
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p.PickTargetMin != 0 || !effectstest.HasID(p.PickTargetCards, dead) {
		t.Fatalf("up to one creature card from your graveyard: %+v", p)
	}
	effectstest.PickCard(t, g, me.ID, dead)
	effectstest.PassPriorityAroundTable(t, g)
	// Each opponent has one creature, so neither is asked (ADR 0175's
	// forced pick, #3664) and both go.
	if effectstest.SacrificeChoiceFor(g, me.ID) != nil {
		t.Fatal("each opponent — not you — sacrifices a creature of their choice")
	}
	if g.Battlefield.Contains(aBear) || effectstest.FindBattlefieldByName(g, "B's Bear") != uuid.Nil {
		t.Error("each opponent sacrifices their creature")
	}
	if !g.Battlefield.Contains(witch) {
		t.Error("the Witch's controller is not an opponent")
	}
	if !me.Hand.Contains(dead) {
		t.Error("the chosen creature card returns to hand")
	}
	// Gained life, no creature card: the untargeted twin still
	// edicts.
	second := effectstest.B12Creature(g, a.ID, "A's Second Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1) })
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("nothing to target: no prompt")
	}
	if g.Battlefield.Contains(second) {
		t.Error("with an empty graveyard the edict still comes")
	}
}

func TestB24SpineOfIshSahDestroysAndComesBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Signet", "Artifact")
	spine := effectstest.CastAndResolveCreature(t, g, "Spine of Ish Sah", "Artifact", effectstest.B24SpineOfIshSahOracle)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if !effectstest.HasID(p.PickTargetCards, theirs) || !effectstest.HasID(p.PickTargetCards, spine) {
		t.Fatal("any permanent, the Spine itself included")
	}
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Fatal("the target is destroyed")
	}
	effectstest.B18Kill(t, g, spine)
	if !me.Hand.Contains(spine) {
		t.Error("put into a graveyard from the battlefield: back to its owner's hand")
	}
	// Exiled, not died: nothing.
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	spine = effectstest.CastAndResolveCreature(t, g, "Spine of Ish Sah", "Artifact", effectstest.B24SpineOfIshSahOracle)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, spine)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(spine) {
		t.Error("destroying itself still returns it")
	}
}

func TestB24IndulgentAristocratGrowsTheVampires(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	aristocrat := effectstest.B12Push(g, me.ID, "Indulgent Aristocrat", "Creature — Vampire Noble", effectstest.B24IndulgentAristocratOracle, 1, 1)
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, aristocrat), "lifelink") {
		t.Error("printed lifelink")
	}
	vampire := effectstest.B12Creature(g, me.ID, "Other Vampire", "Creature — Vampire", 2, 2)
	human := effectstest.B12Creature(g, me.ID, "Human", "Creature — Human", 2, 2)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C")
	effectstest.B16Activate(t, g, me.ID, aristocrat, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}})
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature is sacrificed as the cost")
	}
	if effectstest.CounterCount(g, aristocrat, "+1/+1") != 1 || effectstest.CounterCount(g, vampire, "+1/+1") != 1 {
		t.Error("each Vampire you control — the Aristocrat included — gets a counter")
	}
	if effectstest.CounterCount(g, human, "+1/+1") != 0 {
		t.Error("a Human is not a Vampire")
	}
}

func TestB24ChulaneDrawsOnCreatureSpellsAndBouncesYourCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	chulane := effectstest.B12Push(g, me.ID, "Chulane, Teller of Tales", "Legendary Creature — Human Druid", effectstest.B24ChulaneOracle, 2, 4)
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, chulane), "vigilance") {
		t.Error("printed vigilance")
	}
	hand := me.Hand.Size()
	bear := effectstest.CastAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Fatalf("a creature spell draws: %d", got-hand)
	}
	hand = me.Hand.Size()
	effectstest.B24CastNoncreature(t, g, "Ponder", "Sorcery", "{U}")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Error("a noncreature spell draws nothing")
	}
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.B06AddMana(me, "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, chulane, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(theirs)}); err == nil {
		t.Fatal("target creature YOU control")
	}
	effectstest.B16Activate(t, g, me.ID, chulane, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bear)})
	if !me.Hand.Contains(bear) {
		t.Error("the creature returns to its owner's hand")
	}
	if !effectstest.B20Tapped(t, g, chulane) {
		t.Error("the bounce has a tap cost")
	}
	if spec, _ := effects.Lookup(effectstest.B24ChulaneOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("the land drop landed with #654; nothing is deferred any more")
	}
}

func TestB24DwynenPumpsElvesAndGainsPerAttackingElf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dwynen := effectstest.B12Push(g, me.ID, "Dwynen, Gilt-Leaf Daen", "Legendary Creature — Elf Warrior", effectstest.B24DwynenOracle, 3, 4)
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, dwynen), "reach") {
		t.Error("printed reach")
	}
	elf := effectstest.B12Creature(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirElf := effectstest.B12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	if effectstest.EffectivePower(t, g, elf) != 2 || effectstest.EffectivePower(t, g, dwynen) != 3 || effectstest.EffectivePower(t, g, bear) != 2 || effectstest.EffectivePower(t, g, theirElf) != 1 {
		t.Error("other Elves YOU control get +1/+1")
	}
	life := me.Life
	effectstest.DeclareAttack(t, g, opp.ID, dwynen, elf, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("Dwynen and one Elf attacking: gain 2, got %d → %d", life, me.Life)
	}
}

func TestB24VraskaJoinsUpGivesDeathtouchCountersAndDrawsOffLegends(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	legend := effectstest.B21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	vraska := effectstest.CastAndResolveCreature(t, g, "Vraska Joins Up", "Legendary Enchantment", effectstest.B24VraskaJoinsUpOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterCount(g, bear, "deathtouch") != 1 || effectstest.CounterCount(g, legend, "deathtouch") != 1 {
		t.Fatal("each creature you control gets a deathtouch counter")
	}
	if effectstest.CounterCount(g, theirs, "deathtouch") != 0 {
		t.Error("an opponent's creature gets none")
	}
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, bear), "deathtouch") {
		t.Error("a deathtouch counter grants deathtouch")
	}
	hand := me.Hand.Size()
	effectstest.AttackWith(t, g, opp.ID, bear, legend)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("the legendary creature connecting draws one — the Bear none: %d", got-hand)
	}
	// ADR 0101 (CR 122.1b): the engine reads the counters itself, so
	// they keep granting deathtouch once Vraska has left.
	effectstest.B18Kill(t, g, vraska)
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, bear), "deathtouch") {
		t.Error("the deathtouch counter keeps granting deathtouch after Vraska has left")
	}
	if spec, _ := effects.Lookup(effectstest.B24VraskaJoinsUpOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("with no gap left Vraska Joins Up is CompletenessFull")
	}
}

func TestB24WrathfulRaptorsReflectsDamageToADinosaur(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	raptors := effectstest.B12Push(g, me.ID, "Wrathful Raptors", "Creature — Dinosaur", effectstest.B24WrathfulRaptorsOracle, 5, 5)
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, raptors), "trample") {
		t.Error("printed trample")
	}
	dino := effectstest.B12Creature(g, me.ID, "My Dinosaur", "Creature — Dinosaur", 3, 3)
	theirDino := effectstest.B12Creature(g, opp.ID, "Their Dinosaur", "Creature — Dinosaur", 3, 3)
	theirBear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.B12Bolt(t, g, game.TargetRef{Kind: game.TargetCard, ID: dino})
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil || p.Source != raptors {
		t.Fatal("the Raptors' controller picks the target")
	}
	if effectstest.HasID(p.PickTargetCards, theirDino) || effectstest.HasID(p.PickTargetCards, raptors) {
		t.Error("a Dinosaur — anyone's — is not a legal target")
	}
	if !effectstest.HasID(p.PickTargetCards, theirBear) || !effectstest.HasID(p.PickTargetPlayers, opp.ID) {
		t.Error("a non-Dinosaur creature and a player are")
	}
	effectstest.PickCard(t, g, me.ID, theirBear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirBear) {
		t.Error("the damaged Dinosaur deals 3 to the Bear")
	}
	// An opponent's Dinosaur being damaged is not yours.
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.B12Bolt(t, g, game.TargetRef{Kind: game.TargetCard, ID: theirDino})
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a Dinosaur you don't control does not trigger")
	}
}

func TestB24MariExilesWithHitCountersAndBountiesThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mari := effectstest.B12Push(g, me.ID, "Mari, the Killing Quill", "Legendary Creature — Vampire Assassin", effectstest.B24MariOracle, 3, 2)
	rogue := effectstest.B12Creature(g, me.ID, "Rogue", "Creature — Human Rogue", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, mari), "deathtouch") || !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, rogue), "deathtouch") {
		t.Error("Assassins and Rogues you control have deathtouch — Mari included")
	}
	if effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, bear), "deathtouch") {
		t.Error("a Bear is none of the three")
	}
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Fodder", "Creature — Bear", 1, 1)
	effectstest.B18Kill(t, g, theirs)
	if !effectstest.InExile(g, theirs) || effectstest.B12Counter(t, g, theirs, "hit") != 1 {
		t.Fatal("an opponent's creature dying is exiled with a hit counter")
	}
	effectstest.B18Kill(t, g, mine)
	if !me.Graveyard.Contains(mine) {
		t.Error("your own creature dying stays in the graveyard")
	}
	// The Rogue connects: yes to the bounty.
	hand := me.Hand.Size()
	effectstest.AttackWith(t, g, opp.ID, rogue, bear)
	if n := effectstest.B06TriggerPromptsFor(g, me.ID); n != 1 {
		t.Fatalf("the Rogue's hit asks; the Bear's does not: %d prompts", n)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12Counter(t, g, theirs, "hit") != 0 {
		t.Error("the hit counter is removed")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("draw a card: %d", got-hand)
	}
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Treasure"); n != 2 {
		t.Errorf("two Treasures: %d", n)
	}
	// No hit-countered card that player owns: yes does nothing.
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	hand = me.Hand.Size()
	effectstest.AttackWith(t, g, opp.ID, rogue)
	prompt := effectstest.AwaitTriggerPrompt(t, g, me.ID)
	if prompt == nil || !prompt.NoLegalTarget {
		t.Fatal("the prompt warns there is nothing to remove")
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || effectstest.CountBattlefieldNamed(g, me.ID, "Treasure") != 2 {
		t.Error("if you do — and you could not")
	}
}
