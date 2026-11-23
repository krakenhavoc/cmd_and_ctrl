package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// b26Lives snapshots every seat's life total.
func b26Lives(g *game.Game) []int {
	out := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		out[i] = p.Life
	}
	return out
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Graypelt
// Refuge is a row in the gain-land table, so a transposed row is
// invisible until someone plays that exact card.
func TestBatch26CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.B26TraumatizeOracle:            "Traumatize",
		effectstest.B26DuskLegionDuelistOracle:     "Dusk Legion Duelist",
		effectstest.B26AerialExtortionistOracle:    "Aerial Extortionist",
		effectstest.B26BiteDownOracle:              "Bite Down",
		effectstest.B26TheEarthKingOracle:          "The Earth King",
		effectstest.B26AetherSpellbombOracle:       "Aether Spellbomb",
		effectstest.B26NightshadeDryadOracle:       "Nightshade Dryad",
		effectstest.B26PeerlessRecyclingOracle:     "Peerless Recycling",
		effectstest.B26ExpelTheInterlopersOracle:   "Expel the Interlopers",
		effectstest.B26ChampionOfThePerishedOracle: "Champion of the Perished",
		effectstest.B26AetherspoutsOracle:          "Aetherspouts",
		effectstest.B26CybermanPatrolOracle:        "Cyberman Patrol",
		effectstest.B26HeapedHarvestOracle:         "Heaped Harvest",
		effectstest.B26DwarvenMineOracle:           "Dwarven Mine",
		effectstest.B26ThrashingBrontodonOracle:    "Thrashing Brontodon",
		effectstest.B26ElvenAmbushOracle:           "Elven Ambush",
		effectstest.B26UlvenwaldHydraOracle:        "Ulvenwald Hydra",
		effectstest.B26CarmenTestOracle:            "Carmen, Cruel Skymarcher",
		effectstest.B26SylvanAnthemOracle:          "Sylvan Anthem",
		effectstest.B26RaiseThePastOracle:          "Raise the Past",
		effectstest.B26FellOracle:                  "Fell",
		effectstest.B26ZurTheEnchanterOracle:       "Zur the Enchanter",
		effectstest.B26PriceOfProgressOracle:       "Price of Progress",
		effectstest.B26GraypeltRefugeOracle:        "Graypelt Refuge",
		effectstest.B26GenesisChamberOracle:        "Genesis Chamber",
		effectstest.B26ChromaticStarOracle:         "Chromatic Star",
		effectstest.B26InspiringLeaderOracle:       "Inspiring Leader",
		effectstest.B26ValkyrieHarbingerOracle:     "Valkyrie Harbinger",
	}
	if len(want) != 28 {
		t.Fatalf("the batch registers 28 cards, the table lists %d", len(want))
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
	// The remaining declared skips must NOT be registered — each needs
	// a prompt, a target kind or a static the engine cannot express,
	// and a spec would ship the card stronger than printed or as
	// something other than itself.
	//
	// Tragic Arrogance came OFF this list with #1214, which built the
	// prompt kind it was skipped for: "the caster picks among another
	// player's permanents" is game.PendingChoiceTheirPermanents, and
	// the card ships full in tragic_arrogance.go.
	for _, skipped := range []string{
		// Deadpool, Trading Card came off this list with #3213: it ships
		// with the text-box exchange declared as a caveat.
		// Caesar, Legion's Emperor came off this list with the Fallout
		// GAP batch 61 (#5948): the sacrifice is maySacrificeOne and the
		// bullets are a reflexive modal trigger (caesar_legions_emperor.go).
		// Great Divide Guide came off this list with ADR 0093 and ships
		// in great_divide_guide.go.
		"8d35cef8-a52d-45fb-8f5f-cccea26826d0", // Wyll's Reversal — a die roll and target redirection
	} {
		if _, ok := effects.Lookup(skipped); ok {
			t.Errorf("%s is a declared skip and must not be registered", skipped)
		}
	}
}

// --- the spells ----------------------------------------------------

func TestB26TraumatizeMillsHalfRoundedDown(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	before := opp.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Traumatize", "Sorcery", effectstest.B26TraumatizeOracle, effectstest.B16TargetPlayer(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Library.Size(); got != before-before/2 {
		t.Errorf("library %d → %d, want half milled (%d)", before, got, before-before/2)
	}
	if got := opp.Graveyard.Size(); got != before/2 {
		t.Errorf("graveyard holds %d, want %d", got, before/2)
	}
}

func TestB26FellDestroysTargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	if legal := effectstest.LegalCards(g, g.Seats[0].ID, effectstest.B26FellOracle); !legal[bear] || legal[rock] {
		t.Fatal("a creature is a legal target; an artifact is not")
	}
	effectstest.CastCatalogSpell(t, g, "Fell", "Sorcery", effectstest.B26FellOracle, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Graveyard.Contains(bear) {
		t.Error("the creature is destroyed")
	}
}

func TestB26PriceOfProgressBurnsEachPlayerForTheirNonbasics(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLand(g, me.ID, "Command Tower", "Land", "")
	effectstest.SeedLand(g, me.ID, "Exotic Orchard", "Land", "")
	effectstest.SeedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
	effectstest.SeedLand(g, opp.ID, "Reflecting Pool", "Land", "")
	effectstest.SeedLand(g, opp.ID, "Snow-Covered Island", "Basic Snow Land — Island", "")
	before := b26Lives(g)
	effectstest.CastCatalogSpell(t, g, "Price of Progress", "Instant", effectstest.B26PriceOfProgressOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	want := []int{before[0] - 4, before[1] - 2, before[2], before[3]}
	for i, p := range g.Seats {
		if p.Life != want[i] {
			t.Errorf("seat %d: %d → %d, want %d", i, before[i], p.Life, want[i])
		}
	}
}

func TestB26ElvenAmbushMakesAnElfPerElfAndTheTokensCount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Creature(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	effectstest.B12Creature(g, me.ID, "Elvish Mystic", "Creature — Elf Druid", 1, 1)
	effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Elven Ambush", "Instant", effectstest.B26ElvenAmbushOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := b05CountTokensNamed(g, me.ID, "Elf Warrior"); n != 2 {
		t.Fatalf("two Elves, two tokens: got %d", n)
	}
	effectstest.CastCatalogSpell(t, g, "Elven Ambush", "Instant", effectstest.B26ElvenAmbushOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := b05CountTokensNamed(g, me.ID, "Elf Warrior"); n != 6 {
		t.Errorf("the tokens are Elves — four Elves, four more: got %d", n)
	}
}

func TestB26RaiseThePastReturnsCheapCreatureCardsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cheap := effectstest.B17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	free := effectstest.B17GraveyardCard(me, "Dead Memnite", "Artifact Creature — Construct", "{0}")
	big := effectstest.B17GraveyardCard(me, "Dead Wurm", "Creature — Wurm", "{2}{G}")
	rock := effectstest.B17GraveyardCard(me, "Dead Signet", "Artifact", "{2}")
	effectstest.CastCatalogSpell(t, g, "Raise the Past", "Sorcery", effectstest.B26RaiseThePastOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap) || !g.Battlefield.Contains(free) {
		t.Error("creature cards with mana value 2 or less return")
	}
	if g.Battlefield.Contains(big) || g.Battlefield.Contains(rock) {
		t.Error("a three-drop and a noncreature stay in the graveyard")
	}
	if c, _ := effectstest.BattlefieldCard(g, cheap); c.Controller != me.ID {
		t.Error("under their owner's — your — control")
	}
}

func TestB26PeerlessRecyclingReturnsAPermanentCardWithoutTheGift(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rock := effectstest.B17GraveyardCard(me, "Dead Signet", "Artifact", "{2}")
	bolt := effectstest.B17GraveyardCard(me, "Dead Bolt", "Instant", "{R}")
	if legal := effectstest.LegalCards(g, me.ID, effectstest.B26PeerlessRecyclingOracle); !legal[rock] || legal[bolt] {
		t.Fatal("a permanent card is a legal target; an instant is not")
	}
	effectstest.CastCatalogSpell(t, g, "Peerless Recycling", "Instant", effectstest.B26PeerlessRecyclingOracle, effectstest.B16TargetCard(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(rock) {
		t.Error("the permanent card returns to hand")
	}
	// ADR 0089: unpromised, one card; the gift's clause is the
	// two-card one (TestGiftPeerlessRecyclingPromisedReturnsTwo).
	if spec, _ := effects.Lookup(effectstest.B26PeerlessRecyclingOracle); spec.Targets.Max != 1 || spec.Gift == nil || spec.Gift.Targets.Max != 2 {
		t.Error("one card without the gift, two with it")
	}
}

func TestB26BiteDownBitesWithACreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Beast", "Creature — Beast", 4, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	wall := effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 6)
	walker := effectstest.PushWalkerForTest(g, opp.ID, "Their Walker", "", 5)
	effectstest.CastCatalogSpell(t, g, "Bite Down", "Instant", effectstest.B26BiteDownOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("4 damage from the Beast kills the Bear")
	}
	if effectstest.DamageMarkedOn(g, mine) != 0 {
		t.Error("a bite is one-sided — the Beast takes nothing")
	}
	// A planeswalker is a legal second target.
	effectstest.CastCatalogSpell(t, g, "Bite Down", "Instant", effectstest.B26BiteDownOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: walker}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoyaltyOf(g, walker); got != 1 {
		t.Errorf("the walker takes 4: loyalty %d, want 1", got)
	}
	// #764: each slot is its own CLAUSE, so a pair that fits the
	// wrong slots is refused at ANNOUNCE (CR 601.2c) — their Wall
	// cannot fill "target creature you control", and my Beast cannot
	// fill "target creature or planeswalker you don't control".
	if err := effectstest.CastCatalogSpellErr(t, g, "Bite Down", "Instant", effectstest.B26BiteDownOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: wall}, {Kind: game.TargetCard, ID: mine}}); err != game.ErrIllegalTarget {
		t.Errorf("wrong-slot pair: err %v, want ErrIllegalTarget", err)
	}
	if effectstest.DamageMarkedOn(g, mine) != 0 {
		t.Error("a refused announcement deals nothing")
	}
	if spec, _ := effects.Lookup(effectstest.B26BiteDownOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("per-slot clauses landed in #764 — the caveat is gone")
	}
}

func TestB26ExpelTheInterlopersDestroysAtOrAboveTheChosenPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myBig := effectstest.B12Creature(g, me.ID, "My Giant", "Creature — Giant", 5, 5)
	mySmall := effectstest.B12Creature(g, me.ID, "My Elf", "Creature — Elf", 1, 1)
	theirFour := effectstest.B12Creature(g, opp.ID, "Their Hill Giant", "Creature — Giant", 4, 3)
	theirThree := effectstest.B12Creature(g, opp.ID, "Their Knight", "Creature — Knight", 3, 3)
	spec, _ := effects.Lookup(effectstest.B26ExpelTheInterlopersOracle)
	if spec.Modes != nil || spec.Completeness != effects.CompletenessFull {
		t.Fatalf("the number is chosen as the spell resolves, not as a mode at cast: %+v", spec.Modes)
	}
	effectstest.CastCatalogSpell(t, g, "Expel the Interlopers", "Sorcery", effectstest.B26ExpelTheInterlopersOracle, nil)
	if effectstest.PayAmountPrompt(g, me.ID) != nil {
		t.Fatal("nothing is asked while the spell is on the stack: opponents do not learn the number")
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.PayAmountPrompt(g, me.ID)
	if c == nil {
		t.Fatal("the number is asked as the spell resolves")
	}
	if pa := c.PayAmount; pa.Min != 0 || pa.Max != 10 || pa.NoMax {
		t.Fatalf("prompt = %+v, want 0 through 10", *pa)
	} else if pa.Goal != 3 {
		t.Errorf("goal %d, want 3: it kills their 4 and 3 for the loss of only your 5", pa.Goal)
	}
	if err := g.ResolvePayAmount(c.ID, me.ID, 4); err != nil {
		t.Fatalf("answer 4: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(myBig) || g.Battlefield.Contains(theirFour) {
		t.Error("power 4 and 5 are destroyed, yours included")
	}
	if !g.Battlefield.Contains(mySmall) || !g.Battlefield.Contains(theirThree) {
		t.Error("power 1 and 3 survive a 4")
	}
	// Zero is Wrath of God.
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.CastCatalogSpell(t, g, "Expel the Interlopers", "Sorcery", effectstest.B26ExpelTheInterlopersOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ResolvePayAmount(effectstest.PayAmountPrompt(g, me.ID).ID, me.ID, 0); err != nil {
		t.Fatalf("answer 0: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mySmall) || g.Battlefield.Contains(theirThree) {
		t.Error("a 0 destroys every creature")
	}
}

func TestB26AetherspoutsTucksAttackersByTheirOwnersChoice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	// Seat 1 attacks seat 2 with two creatures; seat 0 flashes it in.
	raider := effectstest.PushVanillaCreature(g, a.ID, "Raider", 2, 2)
	brute := effectstest.PushVanillaCreature(g, a.ID, "Brute", 3, 3)
	homebody := effectstest.PushVanillaCreature(g, a.ID, "Homebody", 1, 1)
	effectstest.AangAdvanceToMain(t, g, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{raider, brute} {
		if err := g.DeclareAttacker(id, b.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	libBefore := a.Library.Size()
	effectstest.CastInPlace(t, g, me.ID, "Aetherspouts", effectstest.B26AetherspoutsOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(raider) || g.Battlefield.Contains(brute) {
		t.Fatal("both attackers leave the battlefield")
	}
	if !g.Battlefield.Contains(homebody) {
		t.Error("a creature that did not attack stays")
	}
	if a.Library.Size() != libBefore+2 || !a.Library.Contains(raider) || !a.Library.Contains(brute) {
		t.Fatal("the attackers are in their owner's library")
	}
	c := effectstest.PutInLibraryChoiceFor(g, a.ID)
	if c == nil || len(c.ScryCards) != 2 || c.LibraryPlacement != game.LibraryPlaceTopOrBottom {
		t.Fatalf("the owner chooses top or bottom for each, over exactly those two: %+v", c)
	}
	if !effectstest.HasID(c.ScryCards, raider) || !effectstest.HasID(c.ScryCards, brute) {
		t.Fatal("the prompt names the attackers and nothing else")
	}
	if effectstest.PutInLibraryChoiceFor(g, me.ID) != nil || effectstest.PutInLibraryChoiceFor(g, b.ID) != nil {
		t.Error("only the owner is asked")
	}
	if effectstest.ScryChoiceFor(g, a.ID) != nil {
		t.Error("#996: Aetherspouts prints no scry")
	}
	// Raider to the bottom, Brute kept on top.
	if err := g.ResolvePutInLibrary(c.ID, a.ID, []uuid.UUID{raider}, []uuid.UUID{brute}); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if n := effectstest.CountEvents(g, game.EventScry); n != 0 {
		t.Errorf("#996: Aetherspouts emitted %d scry events; it is not a scry", n)
	}
	if got := a.Library.Cards[len(a.Library.Cards)-1].InstanceID; got != brute {
		t.Error("the kept card is on top")
	}
	if got := a.Library.Cards[0].InstanceID; got != raider {
		t.Error("the other is on the bottom")
	}
}

// --- the artifacts and their abilities -----------------------------

func TestB26AetherSpellbombBouncesOrDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	bomb := effectstest.B12Push(g, me.ID, "Aether Spellbomb", "Artifact", effectstest.B26AetherSpellbombOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "U")
	effectstest.B16Activate(t, g, me.ID, bomb, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bear)})
	if g.Battlefield.Contains(bomb) {
		t.Error("sacrificed as the cost")
	}
	if !opp.Hand.Contains(bear) {
		t.Error("the creature returns to its owner's hand")
	}
	second := effectstest.B12Push(g, me.ID, "Aether Spellbomb", "Artifact", effectstest.B26AetherSpellbombOracle, 0, 0)
	effectstest.B06AddMana(me, "C")
	hand := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, second, 1, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(second) {
		t.Error("the draw sacrifices it too")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("drew %d, want 1", got-hand)
	}
}

func TestB26ThrashingBrontodonEatsItselfToDestroyAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dino := effectstest.B12Push(g, me.ID, "Thrashing Brontodon", "Creature — Dinosaur", effectstest.B26ThrashingBrontodonOracle, 3, 4)
	aura := effectstest.B12Permanent(g, opp.ID, "Their Rancor", "Enchantment — Aura")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, dino, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bear)}); err == nil {
		t.Fatal("a creature is not an artifact or enchantment")
	}
	effectstest.B16Activate(t, g, me.ID, dino, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(aura)})
	if g.Battlefield.Contains(dino) || !me.Graveyard.Contains(dino) {
		t.Error("the Brontodon is sacrificed as the cost")
	}
	if g.Battlefield.Contains(aura) {
		t.Error("the enchantment is destroyed")
	}
}

func TestB26NightshadeDryadTapsForColorlessOrAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dryad := effectstest.B12Push(g, me.ID, "Nightshade Dryad", "Creature — Dryad", effectstest.B26NightshadeDryadOracle, 1, 2)
	effectstest.AssertKeywords(t, g, dryad, "deathtouch")
	if err := g.ActivateManaAbility(me.ID, dryad, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility 0: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
	me.ManaPool.EmptyPool()
	effectstest.B22Untap(g, dryad)
	if err := g.ActivateManaAbility(me.ID, dryad, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility 1: %v", err)
	}
	pick := effectstest.RiderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("any colour is a five-way pick, not narrowed to a commander: %+v", pick)
	}
}

func TestB26ChromaticStarCracksForAnyColorAndDrawsWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	star := effectstest.B12Push(g, me.ID, "Chromatic Star", "Artifact", effectstest.B26ChromaticStarOracle, 0, 0)
	if err := g.ActivateManaAbility(me.ID, star, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("{1} is a cost — with an empty pool the Star stays")
	}
	if !g.Battlefield.Contains(star) {
		t.Fatal("a failed activation leaves the Star untouched")
	}
	effectstest.B06AddMana(me, "C")
	hand := me.Hand.Size()
	if err := g.ActivateManaAbility(me.ID, star, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := effectstest.RiderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("any colour: %+v", pick)
	}
	if g.Battlefield.Contains(star) || !me.Graveyard.Contains(star) {
		t.Fatal("sacrificed as the cost")
	}
	if effectstest.B10ResolveAllManaPicks(t, g, me.ID, "U") != 1 {
		t.Fatal("one pick")
	}
	if effectstest.TriggerOnStack(g, star) == nil && len(g.PendingTriggers) == 0 {
		t.Fatal("the draw is a trigger — it uses the stack")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("drew %d, want 1", got-hand)
	}
	// Destroyed rather than cracked: still a card.
	second := effectstest.B12Push(g, me.ID, "Chromatic Star", "Artifact", effectstest.B26ChromaticStarOracle, 0, 0)
	effectstest.B18Kill(t, g, second)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("a destroyed Star draws too: hand %d, want %d", got, hand+2)
	}
}

func TestB26HeapedHarvestFetchesOnEntryAndOnSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lib := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	harvest := effectstest.CastCatalogSpell(t, g, "Heaped Harvest", "Artifact — Food", effectstest.B26HeapedHarvestOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil || len(c.SearchCards) != 2 || c.SearchMax != 1 {
		t.Fatalf("the entry search offers the two basics, one to take: %+v", c)
	}
	effectstest.AnswerSearchByID(t, g, me.ID, lib[0])
	if !g.Battlefield.Contains(lib[0]) || !effectstest.B20Tapped(t, g, lib[0]) {
		t.Fatal("the basic enters the battlefield tapped")
	}
	// Crack it: the sacrifice trigger's search lands above the life.
	life := me.Life
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, harvest, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	if g.Battlefield.Contains(harvest) {
		t.Fatal("sacrificed as the cost")
	}
	if n := len(g.PendingTriggers) + effectstest.TriggersOnStackFrom(g, harvest); n != 1 {
		t.Fatalf("the sacrifice trigger sits above the ability: %d triggers", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c = effectstest.SearchChoiceFor(g, me.ID)
	if c == nil || len(c.SearchCards) != 1 {
		t.Fatalf("the sacrifice search offers the remaining basic: %+v", c)
	}
	effectstest.AnswerSearchFailToFind(t, g, me.ID)
	if g.Battlefield.Contains(lib[1]) {
		t.Error("a declined search fetches nothing")
	}
	// #730: the search trigger sits ABOVE the ability's life gain and
	// the table does not move past an unanswered prompt, so the 3 life
	// arrives only once the search has been answered.
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d → %d, want +3", life, me.Life)
	}
}

// --- the lands -----------------------------------------------------

func TestB26DwarvenMineEntersUntappedWithThreeOtherMountainsAndMakesADwarf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	first := effectstest.B12PlayFromHand(t, g, "Dwarven Mine", "Land — Mountain", effectstest.B26DwarvenMineOracle, game.CastSpellParams{})
	if !effectstest.B20Tapped(t, g, first) {
		t.Fatal("with no other Mountains it enters tapped")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := b05CountTokensNamed(g, me.ID, "Dwarf"); n != 0 {
		t.Fatalf("entered tapped: no Dwarf, got %d", n)
	}
	effectstest.SeedLand(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	effectstest.SeedLand(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	// The first Mine is a Mountain too: that is three others.
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	second := effectstest.B12PlayFromHand(t, g, "Dwarven Mine", "Land — Mountain", effectstest.B26DwarvenMineOracle, game.CastSpellParams{})
	if effectstest.B20Tapped(t, g, second) {
		t.Fatal("with three other Mountains it enters untapped")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := b05CountTokensNamed(g, me.ID, "Dwarf"); n != 1 {
		t.Errorf("entered untapped: one Dwarf, got %d", n)
	}
	if err := g.ActivateManaAbility(me.ID, second, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("a Mountain taps for {R}: pool %v", got)
	}
}

func TestB26GraypeltRefugeEntersTappedGainsOneAndTapsForGreenOrWhite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	land := effectstest.PlayLandFromHand(t, g, "Graypelt Refuge", effectstest.B26GraypeltRefugeOracle)
	if !effectstest.B20Tapped(t, g, land) {
		t.Fatal("enters tapped")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d → %d, want +1", life, me.Life)
	}
	effectstest.B22Untap(g, land)
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := effectstest.RiderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 2 || pick.ColorOptions[0] != "G" || pick.ColorOptions[1] != "W" {
		t.Fatalf("{G} or {W}, got %+v", pick)
	}
}

// --- the creatures and their triggers ------------------------------

func TestB26DuskLegionDuelistDrawsOncePerTurnOnCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	duelist := effectstest.B12Push(g, me.ID, "Dusk Legion Duelist", "Creature — Vampire Soldier", effectstest.B26DuskLegionDuelistOracle, 2, 2)
	effectstest.AssertKeywords(t, g, duelist, "vigilance")
	hand := me.Hand.Size()
	if err := g.AddCounter(duelist, "+1/+1", 2); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Fatalf("two counters at once is one trigger: drew %d", got-hand)
	}
	if err := g.AddCounter(duelist, "+1/+1", 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Fatalf("only once each turn: drew %d", got-hand)
	}
	if err := g.AddCounter(duelist, "+1/+1", -1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.AddCounter(duelist, "+1/+1", 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("a new turn, a new draw — anyone's turn: drew %d in total", got-hand)
	}
}

func TestB26AerialExtortionistExilesWithABuybackAndDrawsOnTheBuyback(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Signet", TypeLine: "Artifact", ManaCost: "{2}",
		Owner: opp.ID, Controller: opp.ID,
	})
	land := effectstest.SeedLand(g, opp.ID, "Their Island", "Basic Land — Island", "")
	bird := effectstest.B20CastCreature(t, g, me, "Aerial Extortionist", "Creature — Bird Soldier", effectstest.B26AerialExtortionistOracle, 4, 3)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil || p.PickTargetMin != 0 {
		t.Fatalf("the ETB asks for up to one nonland permanent: %+v", p)
	}
	if effectstest.HasID(p.PickTargetCards, land) || !effectstest.HasID(p.PickTargetCards, rock) {
		t.Fatal("a land is not a legal pick; the Signet is")
	}
	effectstest.PickCard(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, rock) {
		t.Fatal("the Signet is exiled")
	}
	perm := effectstest.ExiledPermission(g, rock)
	if perm.Player != opp.ID || perm.Duration.Kind != game.WhileInZone || !perm.CastOnly {
		t.Fatalf("its OWNER may cast it for as long as it stays exiled: %+v", perm)
	}
	// The owner buys it back on their turn — a cast from exile, which
	// is a cast from somewhere other than their hand: the Extortionist
	// draws.
	hand := me.Hand.Size()
	effectstest.AdvanceToPrecombatMainOf(t, g, 1)
	effectstest.B06AddMana(opp, "C", "C")
	if err := g.CastSpell(opp.ID, rock, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("the owner casts it from exile: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) {
		t.Error("the Signet is back")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("a cast from exile by another player draws: drew %d", got-hand)
	}
	// A commander cast from the command zone is the printed shape.
	cmd := uuid.New()
	opp.Command.PushTop(game.Card{
		InstanceID: cmd, Name: "Their Commander", TypeLine: "Legendary Creature — Human",
		Owner: opp.ID, Controller: opp.ID, IsCommander: true,
	})
	if err := g.CastSpell(opp.ID, cmd, game.CastSpellParams{FromZone: "command"}); err != nil {
		t.Fatalf("commander cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("a commander cast draws: drew %d in total", got-hand)
	}
	// Your own cast from anywhere does not, and a hand cast by an
	// opponent does not.
	effectstest.B13OpponentCasts(t, g, opp, "Their Bear", "Creature — Bear", "", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("a hand cast is not the trigger: drew %d in total", got-hand)
	}
	// Combat damage to a player asks again; with nothing chosen,
	// nothing happens.
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.AttackWith(t, g, opp.ID, bird)
	if p := effectstest.LatestPickTarget(g, me.ID); p == nil {
		t.Fatal("connecting asks for a target again")
	}
	effectstest.B26DeclinePickTarget(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) {
		t.Error("declined: nothing exiled")
	}
}

func TestB26TheEarthKingMakesABearAndRampsPerBigAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lib := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Bayou", TypeLine: "Land — Swamp Forest"},
	)
	king := effectstest.B20CastCreature(t, g, me, "The Earth King", "Legendary Creature — Human Noble Ally", effectstest.B26TheEarthKingOracle, 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	n, bear := effectstest.B14Tokens(g, me.ID, "Bear")
	if n != 1 || bear.Power != 4 || bear.Toughness != 4 || !bear.HasColor("G") {
		t.Fatalf("one 4/4 green Bear: %d, %+v", n, bear)
	}
	effectstest.B10Awake(g, bear.InstanceID)
	giant := effectstest.B12Creature(g, me.ID, "My Giant", "Creature — Giant", 5, 5)
	elf := effectstest.B12Creature(g, me.ID, "My Elf", "Creature — Elf", 1, 1)
	effectstest.B10Awake(g, king)
	effectstest.DeclareAttack(t, g, opp.ID, bear.InstanceID, giant, elf, king)
	if n := len(g.PendingTriggers) + effectstest.TriggersOnStackFrom(g, king); n != 1 {
		t.Fatalf("one or more big attackers is ONE trigger, got %d", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil || c.SearchMax != 2 || len(c.SearchCards) != 3 {
		t.Fatalf("two attackers with power 4 or more: up to two basics from the three: %+v", c)
	}
	if effectstest.HasID(c.SearchCards, lib[3]) {
		t.Fatal("a Bayou is not a basic land")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, lib[0], lib[1])
	for _, id := range lib[:2] {
		if !g.Battlefield.Contains(id) || !effectstest.B20Tapped(t, g, id) {
			t.Error("the basics enter tapped")
		}
	}
}

func TestB26ChampionOfThePerishedGrowsWithZombies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	champ := effectstest.B12Push(g, me.ID, "Champion of the Perished", "Creature — Zombie", effectstest.B26ChampionOfThePerishedOracle, 1, 1)
	effectstest.CastCatalogSpell(t, g, "Gravecrawler", "Creature — Zombie", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, champ, "+1/+1"); got != 1 {
		t.Fatalf("a Zombie entering: one counter, got %d", got)
	}
	effectstest.CastCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, champ, "+1/+1"); got != 1 {
		t.Errorf("a Bear is not a Zombie: still %d", got)
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, 1)
	effectstest.B13OpponentCasts(t, g, opp, "Their Zombie", "Creature — Zombie", "", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, champ, "+1/+1"); got != 1 {
		t.Errorf("an opponent's Zombie is not yours: still %d", got)
	}
}

func TestB26CybermanPatrolGivesArtifactCreaturesAfflictThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	patrol := effectstest.B12Push(g, me.ID, "Cyberman Patrol", "Artifact Creature — Cyberman", effectstest.B26CybermanPatrolOracle, 2, 2)
	myr := effectstest.B12Creature(g, me.ID, "My Myr", "Artifact Creature — Myr", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	wall1 := effectstest.B12Creature(g, opp.ID, "Wall One", "Creature — Wall", 0, 4)
	wall2 := effectstest.B12Creature(g, opp.ID, "Wall Two", "Creature — Wall", 0, 4)
	wall3 := effectstest.B12Creature(g, opp.ID, "Wall Three", "Creature — Wall", 0, 4)
	effectstest.DeclareAttack(t, g, opp.ID, patrol, myr, bear)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	life := opp.Life
	if err := g.DeclareBlocker(wall1, myr); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	if err := g.DeclareBlocker(wall2, myr); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	if err := g.DeclareBlocker(wall3, bear); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	if n := len(g.PendingTriggers) + effectstest.TriggersOnStackFrom(g, patrol); n != 0 {
		t.Fatalf("#830: a click stages the pairing and announces nothing, %d triggers", n)
	}
	effectstest.LockInBlocks(t, g)
	if n := len(g.PendingTriggers) + effectstest.TriggersOnStackFrom(g, patrol); n != 1 {
		t.Fatalf("the Myr double-blocked is one afflict, the Bear is not an artifact: %d triggers", n)
	}
	// #830: the lock-in happens inside the declare blockers step, so
	// the afflict resolves there — before combat damage, as printed.
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Fatalf("afflict resolves in the declare blockers step: %d to %d", life, opp.Life)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	// #730: the double-blocked Myr's CR 510.1c assignment prompt gates
	// the table — combat no longer resolves around it.
	if n := effectstest.B26AnswerDamageAssignments(t, g); n != 1 {
		t.Fatalf("answered %d damage assignments, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2-3 {
		t.Errorf("2 combat damage and 3 afflict: %d → %d", life, opp.Life)
	}
	// #914 closed the other half of the #388 timing gap — the
	// skip-ahead button resolves the afflict inside declare_blockers
	// too — so the card carries no caveat any more.
	if spec, _ := effects.Lookup(effectstest.B26CybermanPatrolOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("the afflict timing is correct on both routes now")
	}
}

func TestB26UlvenwaldHydraIsAsBigAsYourLandsAndFetchesOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
	effectstest.SeedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
	lib := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Bayou", TypeLine: "Land — Swamp Forest"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	hydra := effectstest.CastCatalogSpell(t, g, "Ulvenwald Hydra", "Creature — Hydra", effectstest.B26UlvenwaldHydraOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertKeywords(t, g, hydra, "reach")
	if p, tough := effectstest.EffectivePower(t, g, hydra), effectstest.EffectiveToughness(t, g, hydra); p != 2 || tough != 2 {
		t.Fatalf("two lands: 2/2, got %d/%d", p, tough)
	}
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil || len(c.SearchCards) != 1 || c.SearchCards[0] != lib[0] {
		t.Fatalf("the ETB offers the one land card, any land: %+v", c)
	}
	effectstest.AnswerSearchByID(t, g, me.ID, lib[0])
	if !g.Battlefield.Contains(lib[0]) || !effectstest.B20Tapped(t, g, lib[0]) {
		t.Fatal("the land enters tapped")
	}
	if p, tough := effectstest.EffectivePower(t, g, hydra), effectstest.EffectiveToughness(t, g, hydra); p != 3 || tough != 3 {
		t.Errorf("three lands: 3/3, got %d/%d", p, tough)
	}
}

func TestB26CarmenGrowsOnAnySacrificeAndReanimatesOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	carmen := effectstest.B12Push(g, me.ID, "Carmen, Cruel Skymarcher", "Legendary Creature — Vampire Soldier", effectstest.B26CarmenTestOracle, 2, 2)
	treasure := effectstest.PushToken(g, opp.ID, effects.TreasureToken())
	three := effectstest.B17GraveyardCard(me, "Dead Knight", "Creature — Knight", "{1}{W}{W}")
	four := effectstest.B17GraveyardCard(me, "Dead Wurm", "Creature — Wurm", "{3}{G}")
	spell := effectstest.B17GraveyardCard(me, "Dead Bolt", "Instant", "{R}")
	life := me.Life
	// An opponent cracking a Treasure is a player sacrificing a
	// permanent.
	if err := g.ActivateManaAbility(opp.ID, treasure, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("crack Treasure: %v", err)
	}
	effectstest.B10ResolveAllManaPicks(t, g, opp.ID, "B")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, carmen, "+1/+1"); got != 1 {
		t.Fatalf("one counter, got %d", got)
	}
	if me.Life != life+1 {
		t.Fatalf("and 1 life: %d → %d", life, me.Life)
	}
	// Power 3 now: the Knight is a legal target, the Wurm and the
	// instant are not.
	effectstest.DeclareAttack(t, g, opp.ID, carmen)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the attack asks for up to one permanent card")
	}
	if !effectstest.HasID(p.PickTargetCards, three) || effectstest.HasID(p.PickTargetCards, four) || effectstest.HasID(p.PickTargetCards, spell) {
		t.Fatalf("mana value at most Carmen's power (3), permanent cards only: %v", p.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, three)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(three) {
		t.Error("the Knight returns to the battlefield")
	}
	if c, _ := effectstest.BattlefieldCard(g, three); c.Controller != me.ID {
		t.Error("under your control")
	}
}

func TestB26SylvanAnthemPumpsGreenCreaturesAndScriesWhenOneEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Sylvan Anthem", "Enchantment", effectstest.B26SylvanAnthemOracle, 0, 0)
	green := effectstest.B16Creature(g, me.ID, "My Elf", "Creature — Elf", 1, 1, "G")
	red := effectstest.B16Creature(g, me.ID, "My Goblin", "Creature — Goblin", 1, 1, "R")
	theirs := effectstest.B16Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1, "G")
	if p := effectstest.EffectivePower(t, g, green); p != 2 {
		t.Errorf("a green creature you control gets +1/+1: power %d", p)
	}
	if p := effectstest.EffectivePower(t, g, red); p != 1 {
		t.Errorf("a red one does not: power %d", p)
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 1 {
		t.Errorf("an opponent's green creature does not: power %d", p)
	}
	b23Cast(t, g, "Llanowar Elves", "Creature — Elf Druid", "{G}", "", []string{"G"}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Fatal("a green creature entering under your control: scry 1")
	}
	effectstest.B26AnswerScryAllTop(t, g, me.ID)
	b23Cast(t, g, "Goblin Guide", "Creature — Goblin", "{R}", "", []string{"R"}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) != nil {
		t.Error("a red creature entering does not")
	}
}

func TestB26ZurFetchesACheapEnchantmentOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zur := effectstest.B12Push(g, me.ID, "Zur the Enchanter", "Legendary Creature — Human Wizard", effectstest.B26ZurTheEnchanterOracle, 1, 4)
	lib := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Rhystic Study", TypeLine: "Enchantment", ManaCost: "{2}{U}"},
		game.Card{Name: "Necropotence", TypeLine: "Enchantment", ManaCost: "{B}{B}{B}"},
		game.Card{Name: "Omniscience", TypeLine: "Enchantment", ManaCost: "{7}{U}{U}{U}"},
		game.Card{Name: "Sol Ring", TypeLine: "Artifact", ManaCost: "{1}"},
	)
	effectstest.DeclareAttack(t, g, opp.ID, zur)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil || len(c.SearchCards) != 2 || c.SearchMax != 1 {
		t.Fatalf("enchantment cards with mana value 3 or less, one to take: %+v", c)
	}
	if effectstest.HasID(c.SearchCards, lib[2]) || effectstest.HasID(c.SearchCards, lib[3]) {
		t.Fatal("a ten-drop and an artifact are not offered")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, lib[1])
	if !g.Battlefield.Contains(lib[1]) {
		t.Error("the enchantment is put onto the battlefield")
	}
	if effectstest.B20Tapped(t, g, lib[1]) {
		t.Error("untapped — the text says nothing about tapped")
	}
}

func TestB26GenesisChamberHandsEachNontokenCreaturesControllerAMyr(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	chamber := effectstest.B12Push(g, me.ID, "Genesis Chamber", "Artifact", effectstest.B26GenesisChamberOracle, 0, 0)
	effectstest.CastCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := b05CountTokensNamed(g, me.ID, "Myr"); n != 1 {
		t.Fatalf("your creature, your Myr: got %d", n)
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, 1)
	effectstest.B13OpponentCasts(t, g, opp, "Their Bear", "Creature — Bear", "", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := b05CountTokensNamed(g, opp.ID, "Myr"); n != 1 {
		t.Fatalf("their creature, their Myr: got %d", n)
	}
	if n := b05CountTokensNamed(g, me.ID, "Myr"); n != 1 {
		t.Fatalf("the Myr is a token and does not re-trigger: still %d", n)
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.B16Tap(g, chamber)
	effectstest.CastCatalogSpell(t, g, "Another Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := b05CountTokensNamed(g, me.ID, "Myr"); n != 1 {
		t.Errorf("a tapped Chamber makes nothing: got %d", n)
	}
}

func TestB26InspiringLeaderPumpsTokensWhileYourCommanderIsOut(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Inspiring Leader", "Legendary Enchantment — Background", effectstest.B26InspiringLeaderOracle, 0, 0)
	mine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Soldier", TypeLine: "Token Creature — Soldier",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Soldier", TypeLine: "Token Creature — Soldier",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})
	real := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if p := effectstest.EffectivePower(t, g, mine); p != 1 {
		t.Fatalf("no commander on the battlefield: the ability does not exist, power %d", p)
	}
	cmd := effectstest.B21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 3, 3)
	if p, tough := effectstest.EffectivePower(t, g, mine), effectstest.EffectiveToughness(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("your token gets +2/+2: %d/%d", p, tough)
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 1 {
		t.Errorf("an opponent's token does not: power %d", p)
	}
	if p := effectstest.EffectivePower(t, g, real); p != 2 {
		t.Errorf("a nontoken creature does not: power %d", p)
	}
	effectstest.LeaveBattlefield(t, g, me.ID, cmd)
	effectstest.AnswerCommanderReturn(t, g, me.ID, false)
	if p := effectstest.EffectivePower(t, g, mine); p != 1 {
		t.Errorf("the commander gone, the anthem is gone: power %d", p)
	}
}

func TestB26ValkyrieHarbingerMakesAnAngelAtEndStepAfterFourLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	valk := effectstest.B12Push(g, me.ID, "Valkyrie Harbinger", "Creature — Angel Cleric", effectstest.B26ValkyrieHarbingerOracle, 4, 5)
	effectstest.AssertKeywords(t, g, valk, "flying", "lifelink")
	effectstest.B14Gain(t, g, me.ID, 3)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := b05CountTokensNamed(g, me.ID, "Angel"); n != 0 {
		t.Fatalf("three life is not four: got %d Angels", n)
	}
	// An opponent's end step counts too, and the gain is this turn's.
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B14Gain(t, g, me.ID, 4)
	effectstest.AdvanceToEndStepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	n, angel := effectstest.B14Tokens(g, me.ID, "Angel")
	if n != 1 || angel.Power != 4 || angel.Toughness != 4 {
		t.Fatalf("one 4/4 Angel: %d, %+v", n, angel)
	}
	effectstest.AssertKeywords(t, g, angel.InstanceID, "flying", "vigilance")
}
