package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch26_test.go — card-level coverage for the card-coverage
// roadmap's batch 26 (#388, `edhrec_rank` 2738–2838): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, attack, land play or step change.
// Helpers from the earlier batch test files are reused by name; new
// ones are b26-prefixed.

const (
	b26TraumatizeOracle            = "e2ea7d01-6564-4a7f-b935-49a5e3978dac"
	b26DuskLegionDuelistOracle     = "30367a59-a812-41d4-a451-80d9071dd4a4"
	b26AerialExtortionistOracle    = "7e7d7f4d-2c22-4803-ba5e-5eb653c25b5e"
	b26BiteDownOracle              = "623903de-3c04-4745-9af3-d7ec9fb2574d"
	b26TheEarthKingOracle          = "d5770b0f-4493-42e9-a2d7-0e74a28da4ba"
	b26AetherSpellbombOracle       = "4b033a0a-c1ae-44d7-9662-72cbbfda024b"
	b26NightshadeDryadOracle       = "f8263274-8cb4-4eab-b541-c7655b555830"
	b26PeerlessRecyclingOracle     = "938c03fc-8adf-4c7a-8ae1-eca8401f7a83"
	b26ExpelTheInterlopersOracle   = "744b7e93-c893-4770-a3bb-6a45b7f8e4f7"
	b26ChampionOfThePerishedOracle = "f3654fbd-16a5-4953-84ac-534e8421032f"
	b26AetherspoutsOracle          = "48369aec-a991-4bef-8554-01c84302b063"
	b26CybermanPatrolOracle        = "1e51fab7-3ca5-4fbb-a1e9-b39c842895e8"
	b26HeapedHarvestOracle         = "bbfc5011-a9b7-442d-a443-974a5a64de46"
	b26DwarvenMineOracle           = "74ed0bd3-ac31-41a4-8220-d8e7c8c1c437"
	b26ThrashingBrontodonOracle    = "60bc63dc-ac9f-4a2f-aef5-c90d0aa31553"
	b26ElvenAmbushOracle           = "a2373025-20c1-4416-91f7-e51f68dbd146"
	b26UlvenwaldHydraOracle        = "40b85f70-78b6-427a-9fb9-c3a72b8528ed"
	b26CarmenTestOracle            = "84c24e62-26bd-4ca4-b6b9-d31e2065f13b"
	b26SylvanAnthemOracle          = "5ab5eefd-47bf-4cdf-bc29-b517e4f6adc0"
	b26RaiseThePastOracle          = "a69a24d0-ca58-4a44-8af1-a3bd1608d2f9"
	b26FellOracle                  = "8f968470-ce00-43a3-95f5-2e3fe987be19"
	b26ZurTheEnchanterOracle       = "d7950018-d744-48a8-81aa-0d8384703f48"
	b26PriceOfProgressOracle       = "e9da499c-fa43-4e94-8395-5c030ff39502"
	b26GraypeltRefugeOracle        = "60b36821-0fad-423c-98c4-f64d991719f3"
	b26GenesisChamberOracle        = "150ab025-5cc3-4468-a724-bbba8838445d"
	b26ChromaticStarOracle         = "fedbd40b-e3a5-449c-a8a3-b42e9da191a9"
	b26InspiringLeaderOracle       = "d46dcc75-55ff-4226-a392-a49755d269d2"
	b26ValkyrieHarbingerOracle     = "18a5e71e-eb75-4ac7-bedd-2220aaa24f78"
)

// b26Lives snapshots every seat's life total.
func b26Lives(g *game.Game) []int {
	out := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		out[i] = p.Life
	}
	return out
}

// b26DeclinePickTarget answers an open "up to one" pick_target prompt
// with nothing.
func b26DeclinePickTarget(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	p := latestPickTarget(g, chooser)
	if p == nil {
		t.Fatalf("no pick_target prompt for %s", chooser)
	}
	if err := g.ResolvePickTargets(p.ID, chooser, nil); err != nil {
		t.Fatalf("ResolvePickTargets (decline): %v", err)
	}
}

// b26AnswerScryAllTop answers a player's open scry prompt by keeping
// every card on top in the offered order.
func b26AnswerScryAllTop(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	c := scryChoiceFor(g, chooser)
	if c == nil {
		t.Fatalf("no scry prompt for %s", chooser)
	}
	if err := g.ResolveScry(c.ID, chooser, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Graypelt
// Refuge is a row in the gain-land table, so a transposed row is
// invisible until someone plays that exact card.
func TestBatch26CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b26TraumatizeOracle:            "Traumatize",
		b26DuskLegionDuelistOracle:     "Dusk Legion Duelist",
		b26AerialExtortionistOracle:    "Aerial Extortionist",
		b26BiteDownOracle:              "Bite Down",
		b26TheEarthKingOracle:          "The Earth King",
		b26AetherSpellbombOracle:       "Aether Spellbomb",
		b26NightshadeDryadOracle:       "Nightshade Dryad",
		b26PeerlessRecyclingOracle:     "Peerless Recycling",
		b26ExpelTheInterlopersOracle:   "Expel the Interlopers",
		b26ChampionOfThePerishedOracle: "Champion of the Perished",
		b26AetherspoutsOracle:          "Aetherspouts",
		b26CybermanPatrolOracle:        "Cyberman Patrol",
		b26HeapedHarvestOracle:         "Heaped Harvest",
		b26DwarvenMineOracle:           "Dwarven Mine",
		b26ThrashingBrontodonOracle:    "Thrashing Brontodon",
		b26ElvenAmbushOracle:           "Elven Ambush",
		b26UlvenwaldHydraOracle:        "Ulvenwald Hydra",
		b26CarmenTestOracle:            "Carmen, Cruel Skymarcher",
		b26SylvanAnthemOracle:          "Sylvan Anthem",
		b26RaiseThePastOracle:          "Raise the Past",
		b26FellOracle:                  "Fell",
		b26ZurTheEnchanterOracle:       "Zur the Enchanter",
		b26PriceOfProgressOracle:       "Price of Progress",
		b26GraypeltRefugeOracle:        "Graypelt Refuge",
		b26GenesisChamberOracle:        "Genesis Chamber",
		b26ChromaticStarOracle:         "Chromatic Star",
		b26InspiringLeaderOracle:       "Inspiring Leader",
		b26ValkyrieHarbingerOracle:     "Valkyrie Harbinger",
	}
	if len(want) != 28 {
		t.Fatalf("the batch registers 28 cards, the table lists %d", len(want))
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
	// The seven declared skips must NOT be registered — each needs a
	// prompt, a target kind or a static the engine cannot express,
	// and a spec would ship the card stronger than printed or as
	// something other than itself.
	for _, skipped := range []string{
		"4b96c22a-0d5b-44fd-b326-5f3ffcc3917b", // Deadpool, Trading Card — text-box exchange
		"8e62d05a-6efd-4764-bca8-97895e0cb613", // Caesar, Legion's Emperor — modal reflexive trigger
		"2993dc7d-723d-4a9b-94bd-4bb02a9f7243", // Tishana's Tidebinder — abilities on the stack can't be targeted
		"5ba73182-30a7-4bad-9cb6-c0feecc2db33", // Meekstone — the untap step has no per-permanent exception
		"79e69a91-d580-47fb-be76-1e32c50d2fa0", // Great Divide Guide — a mana ability granted by a static
		"8a29bd35-33ef-4317-9fe5-8aaff5d7d64d", // Tragic Arrogance — the caster picks among another player's permanents
		"8d35cef8-a52d-45fb-8f5f-cccea26826d0", // Wyll's Reversal — a die roll and target redirection
	} {
		if _, ok := Lookup(skipped); ok {
			t.Errorf("%s is a declared skip and must not be registered", skipped)
		}
	}
}

// --- the spells ----------------------------------------------------

func TestB26TraumatizeMillsHalfRoundedDown(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	before := opp.Library.Size()
	castCatalogSpell(t, g, "Traumatize", "Sorcery", b26TraumatizeOracle, b16TargetPlayer(opp.ID))
	passPriorityAroundTable(t, g)
	if got := opp.Library.Size(); got != before-before/2 {
		t.Errorf("library %d → %d, want half milled (%d)", before, got, before-before/2)
	}
	if got := opp.Graveyard.Size(); got != before/2 {
		t.Errorf("graveyard holds %d, want %d", got, before/2)
	}
}

func TestB26FellDestroysTargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	if legal := legalCards(g, g.Seats[0].ID, b26FellOracle); !legal[bear] || legal[rock] {
		t.Fatal("a creature is a legal target; an artifact is not")
	}
	castCatalogSpell(t, g, "Fell", "Sorcery", b26FellOracle, b16TargetCard(bear))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Graveyard.Contains(bear) {
		t.Error("the creature is destroyed")
	}
}

func TestB26PriceOfProgressBurnsEachPlayerForTheirNonbasics(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLand(g, me.ID, "Command Tower", "Land", "")
	seedLand(g, me.ID, "Exotic Orchard", "Land", "")
	seedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
	seedLand(g, opp.ID, "Reflecting Pool", "Land", "")
	seedLand(g, opp.ID, "Snow-Covered Island", "Basic Snow Land — Island", "")
	before := b26Lives(g)
	castCatalogSpell(t, g, "Price of Progress", "Instant", b26PriceOfProgressOracle, nil)
	passPriorityAroundTable(t, g)
	want := []int{before[0] - 4, before[1] - 2, before[2], before[3]}
	for i, p := range g.Seats {
		if p.Life != want[i] {
			t.Errorf("seat %d: %d → %d, want %d", i, before[i], p.Life, want[i])
		}
	}
}

func TestB26ElvenAmbushMakesAnElfPerElfAndTheTokensCount(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Creature(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	b12Creature(g, me.ID, "Elvish Mystic", "Creature — Elf Druid", 1, 1)
	b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Elven Ambush", "Instant", b26ElvenAmbushOracle, nil)
	passPriorityAroundTable(t, g)
	if n := b05CountTokensNamed(g, me.ID, "Elf Warrior"); n != 2 {
		t.Fatalf("two Elves, two tokens: got %d", n)
	}
	castCatalogSpell(t, g, "Elven Ambush", "Instant", b26ElvenAmbushOracle, nil)
	passPriorityAroundTable(t, g)
	if n := b05CountTokensNamed(g, me.ID, "Elf Warrior"); n != 6 {
		t.Errorf("the tokens are Elves — four Elves, four more: got %d", n)
	}
}

func TestB26RaiseThePastReturnsCheapCreatureCardsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cheap := b17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	free := b17GraveyardCard(me, "Dead Memnite", "Artifact Creature — Construct", "{0}")
	big := b17GraveyardCard(me, "Dead Wurm", "Creature — Wurm", "{2}{G}")
	rock := b17GraveyardCard(me, "Dead Signet", "Artifact", "{2}")
	castCatalogSpell(t, g, "Raise the Past", "Sorcery", b26RaiseThePastOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap) || !g.Battlefield.Contains(free) {
		t.Error("creature cards with mana value 2 or less return")
	}
	if g.Battlefield.Contains(big) || g.Battlefield.Contains(rock) {
		t.Error("a three-drop and a noncreature stay in the graveyard")
	}
	if c, _ := battlefieldCard(g, cheap); c.Controller != me.ID {
		t.Error("under their owner's — your — control")
	}
}

func TestB26PeerlessRecyclingReturnsAPermanentCardWithoutTheGift(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rock := b17GraveyardCard(me, "Dead Signet", "Artifact", "{2}")
	bolt := b17GraveyardCard(me, "Dead Bolt", "Instant", "{R}")
	if legal := legalCards(g, me.ID, b26PeerlessRecyclingOracle); !legal[rock] || legal[bolt] {
		t.Fatal("a permanent card is a legal target; an instant is not")
	}
	castCatalogSpell(t, g, "Peerless Recycling", "Instant", b26PeerlessRecyclingOracle, b16TargetCard(rock))
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(rock) {
		t.Error("the permanent card returns to hand")
	}
	if spec, _ := Lookup(b26PeerlessRecyclingOracle); spec.Completeness != CompletenessCaveats || spec.Targets.Max != 1 {
		t.Error("the gift is a declared gap — one card, never two")
	}
}

func TestB26BiteDownBitesWithACreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Beast", "Creature — Beast", 4, 4)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	wall := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 6)
	walker := pushWalkerForTest(g, opp.ID, "Their Walker", "", 5)
	castCatalogSpell(t, g, "Bite Down", "Instant", b26BiteDownOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("4 damage from the Beast kills the Bear")
	}
	if damageMarkedOn(g, mine) != 0 {
		t.Error("a bite is one-sided — the Beast takes nothing")
	}
	// A planeswalker is a legal second target.
	castCatalogSpell(t, g, "Bite Down", "Instant", b26BiteDownOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: walker}})
	passPriorityAroundTable(t, g)
	if got := loyaltyOf(g, walker); got != 1 {
		t.Errorf("the walker takes 4: loyalty %d, want 1", got)
	}
	// The per-slot clauses are checked at resolution: their Wall in
	// the first slot bites nothing.
	castCatalogSpell(t, g, "Bite Down", "Instant", b26BiteDownOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: wall}, {Kind: game.TargetCard, ID: mine}})
	passPriorityAroundTable(t, g)
	if damageMarkedOn(g, mine) != 0 {
		t.Error("their Wall is not a creature you control — nothing is dealt")
	}
	if spec, _ := Lookup(b26BiteDownOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the resolution-time slot check is a declared gap")
	}
}

func TestB26ExpelTheInterlopersDestroysAtOrAboveTheChosenPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myBig := b12Creature(g, me.ID, "My Giant", "Creature — Giant", 5, 5)
	mySmall := b12Creature(g, me.ID, "My Elf", "Creature — Elf", 1, 1)
	theirFour := b12Creature(g, opp.ID, "Their Hill Giant", "Creature — Giant", 4, 3)
	theirThree := b12Creature(g, opp.ID, "Their Knight", "Creature — Knight", 3, 3)
	spec, _ := Lookup(b26ExpelTheInterlopersOracle)
	if spec.Modes == nil || len(spec.Modes.Options) != 11 {
		t.Fatalf("eleven numbers, 0 through 10: got %+v", spec.Modes)
	}
	castModal(t, g, "Expel the Interlopers", "Sorcery", b26ExpelTheInterlopersOracle, []int{4}, nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(myBig) || g.Battlefield.Contains(theirFour) {
		t.Error("power 4 and 5 are destroyed, yours included")
	}
	if !g.Battlefield.Contains(mySmall) || !g.Battlefield.Contains(theirThree) {
		t.Error("power 1 and 3 survive a 4")
	}
	// Zero is Wrath of God.
	advanceToPrecombatMainOf(t, g, 0)
	castModal(t, g, "Expel the Interlopers", "Sorcery", b26ExpelTheInterlopersOracle, []int{0}, nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mySmall) || g.Battlefield.Contains(theirThree) {
		t.Error("a 0 destroys every creature")
	}
}

func TestB26AetherspoutsTucksAttackersByTheirOwnersChoice(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	// Seat 1 attacks seat 2 with two creatures; seat 0 flashes it in.
	raider := pushVanillaCreature(g, a.ID, "Raider", 2, 2)
	brute := pushVanillaCreature(g, a.ID, "Brute", 3, 3)
	homebody := pushVanillaCreature(g, a.ID, "Homebody", 1, 1)
	aangAdvanceToMain(t, g, 1)
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{raider, brute} {
		if err := g.DeclareAttacker(id, b.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	libBefore := a.Library.Size()
	castInPlace(t, g, me.ID, "Aetherspouts", b26AetherspoutsOracle)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(raider) || g.Battlefield.Contains(brute) {
		t.Fatal("both attackers leave the battlefield")
	}
	if !g.Battlefield.Contains(homebody) {
		t.Error("a creature that did not attack stays")
	}
	if a.Library.Size() != libBefore+2 || !a.Library.Contains(raider) || !a.Library.Contains(brute) {
		t.Fatal("the attackers are in their owner's library")
	}
	c := scryChoiceFor(g, a.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("the owner chooses top or bottom for each, as a scry over exactly those two: %+v", c)
	}
	if !hasID(c.ScryCards, raider) || !hasID(c.ScryCards, brute) {
		t.Fatal("the scry looks at the attackers and nothing else")
	}
	if scryChoiceFor(g, me.ID) != nil || scryChoiceFor(g, b.ID) != nil {
		t.Error("only the owner is asked")
	}
	// Raider to the bottom, Brute kept on top.
	if err := g.ResolveScry(c.ID, a.ID, []uuid.UUID{raider}, []uuid.UUID{brute}); err != nil {
		t.Fatalf("ResolveScry: %v", err)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	bomb := b12Push(g, me.ID, "Aether Spellbomb", "Artifact", b26AetherSpellbombOracle, 0, 0)
	advanceToMain(t, g)
	b06AddMana(me, "U")
	b16Activate(t, g, me.ID, bomb, 0, game.ActivateAbilityParams{Targets: b16TargetCard(bear)})
	if g.Battlefield.Contains(bomb) {
		t.Error("sacrificed as the cost")
	}
	if !opp.Hand.Contains(bear) {
		t.Error("the creature returns to its owner's hand")
	}
	second := b12Push(g, me.ID, "Aether Spellbomb", "Artifact", b26AetherSpellbombOracle, 0, 0)
	b06AddMana(me, "C")
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, second, 1, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(second) {
		t.Error("the draw sacrifices it too")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("drew %d, want 1", got-hand)
	}
}

func TestB26ThrashingBrontodonEatsItselfToDestroyAnEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dino := b12Push(g, me.ID, "Thrashing Brontodon", "Creature — Dinosaur", b26ThrashingBrontodonOracle, 3, 4)
	aura := b12Permanent(g, opp.ID, "Their Rancor", "Enchantment — Aura")
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	b06AddMana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, dino, 0, game.ActivateAbilityParams{Targets: b16TargetCard(bear)}); err == nil {
		t.Fatal("a creature is not an artifact or enchantment")
	}
	b16Activate(t, g, me.ID, dino, 0, game.ActivateAbilityParams{Targets: b16TargetCard(aura)})
	if g.Battlefield.Contains(dino) || !me.Graveyard.Contains(dino) {
		t.Error("the Brontodon is sacrificed as the cost")
	}
	if g.Battlefield.Contains(aura) {
		t.Error("the enchantment is destroyed")
	}
}

func TestB26NightshadeDryadTapsForColorlessOrAnyColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dryad := b12Push(g, me.ID, "Nightshade Dryad", "Creature — Dryad", b26NightshadeDryadOracle, 1, 2)
	assertKeywords(t, g, dryad, "deathtouch")
	if err := g.ActivateManaAbility(me.ID, dryad, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility 0: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
	me.ManaPool.EmptyPool()
	b22Untap(g, dryad)
	if err := g.ActivateManaAbility(me.ID, dryad, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility 1: %v", err)
	}
	pick := riderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("any colour is a five-way pick, not narrowed to a commander: %+v", pick)
	}
}

func TestB26ChromaticStarCracksForAnyColorAndDrawsWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	star := b12Push(g, me.ID, "Chromatic Star", "Artifact", b26ChromaticStarOracle, 0, 0)
	if err := g.ActivateManaAbility(me.ID, star, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("{1} is a cost — with an empty pool the Star stays")
	}
	if !g.Battlefield.Contains(star) {
		t.Fatal("a failed activation leaves the Star untouched")
	}
	b06AddMana(me, "C")
	hand := me.Hand.Size()
	if err := g.ActivateManaAbility(me.ID, star, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := riderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("any colour: %+v", pick)
	}
	if g.Battlefield.Contains(star) || !me.Graveyard.Contains(star) {
		t.Fatal("sacrificed as the cost")
	}
	if b10ResolveAllManaPicks(t, g, me.ID, "U") != 1 {
		t.Fatal("one pick")
	}
	if triggerOnStack(g, star) == nil && len(g.PendingTriggers) == 0 {
		t.Fatal("the draw is a trigger — it uses the stack")
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("drew %d, want 1", got-hand)
	}
	// Destroyed rather than cracked: still a card.
	second := b12Push(g, me.ID, "Chromatic Star", "Artifact", b26ChromaticStarOracle, 0, 0)
	b18Kill(t, g, second)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("a destroyed Star draws too: hand %d, want %d", got, hand+2)
	}
}

func TestB26HeapedHarvestFetchesOnEntryAndOnSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lib := seedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	harvest := castCatalogSpell(t, g, "Heaped Harvest", "Artifact — Food", b26HeapedHarvestOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil || len(c.SearchCards) != 2 || c.SearchMax != 1 {
		t.Fatalf("the entry search offers the two basics, one to take: %+v", c)
	}
	answerSearchByID(t, g, me.ID, lib[0])
	if !g.Battlefield.Contains(lib[0]) || !b20Tapped(t, g, lib[0]) {
		t.Fatal("the basic enters the battlefield tapped")
	}
	// Crack it: the sacrifice trigger's search lands above the life.
	life := me.Life
	b06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, harvest, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	if g.Battlefield.Contains(harvest) {
		t.Fatal("sacrificed as the cost")
	}
	if n := len(g.PendingTriggers) + triggersOnStackFrom(g, harvest); n != 1 {
		t.Fatalf("the sacrifice trigger sits above the ability: %d triggers", n)
	}
	passPriorityAroundTable(t, g)
	c = searchChoiceFor(g, me.ID)
	if c == nil || len(c.SearchCards) != 1 {
		t.Fatalf("the sacrifice search offers the remaining basic: %+v", c)
	}
	if me.Life != life+3 {
		t.Errorf("life %d → %d, want +3", life, me.Life)
	}
	answerSearchFailToFind(t, g, me.ID)
	if g.Battlefield.Contains(lib[1]) {
		t.Error("a declined search fetches nothing")
	}
}

// --- the lands -----------------------------------------------------

func TestB26DwarvenMineEntersUntappedWithThreeOtherMountainsAndMakesADwarf(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	first := b12PlayFromHand(t, g, "Dwarven Mine", "Land — Mountain", b26DwarvenMineOracle, game.CastSpellParams{})
	if !b20Tapped(t, g, first) {
		t.Fatal("with no other Mountains it enters tapped")
	}
	passPriorityAroundTable(t, g)
	if n := b05CountTokensNamed(g, me.ID, "Dwarf"); n != 0 {
		t.Fatalf("entered tapped: no Dwarf, got %d", n)
	}
	seedLand(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	seedLand(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	// The first Mine is a Mountain too: that is three others.
	advanceToPrecombatMainOf(t, g, 0)
	second := b12PlayFromHand(t, g, "Dwarven Mine", "Land — Mountain", b26DwarvenMineOracle, game.CastSpellParams{})
	if b20Tapped(t, g, second) {
		t.Fatal("with three other Mountains it enters untapped")
	}
	passPriorityAroundTable(t, g)
	if n := b05CountTokensNamed(g, me.ID, "Dwarf"); n != 1 {
		t.Errorf("entered untapped: one Dwarf, got %d", n)
	}
	if err := g.ActivateManaAbility(me.ID, second, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("a Mountain taps for {R}: pool %v", got)
	}
}

func TestB26GraypeltRefugeEntersTappedGainsOneAndTapsForGreenOrWhite(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	land := playLandFromHand(t, g, "Graypelt Refuge", b26GraypeltRefugeOracle)
	if !b20Tapped(t, g, land) {
		t.Fatal("enters tapped")
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d → %d, want +1", life, me.Life)
	}
	b22Untap(g, land)
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := riderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 2 || pick.ColorOptions[0] != "G" || pick.ColorOptions[1] != "W" {
		t.Fatalf("{G} or {W}, got %+v", pick)
	}
}

// --- the creatures and their triggers ------------------------------

func TestB26DuskLegionDuelistDrawsOncePerTurnOnCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	duelist := b12Push(g, me.ID, "Dusk Legion Duelist", "Creature — Vampire Soldier", b26DuskLegionDuelistOracle, 2, 2)
	assertKeywords(t, g, duelist, "vigilance")
	hand := me.Hand.Size()
	if err := g.AddCounter(duelist, "+1/+1", 2); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Fatalf("two counters at once is one trigger: drew %d", got-hand)
	}
	if err := g.AddCounter(duelist, "+1/+1", 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Fatalf("only once each turn: drew %d", got-hand)
	}
	if err := g.AddCounter(duelist, "+1/+1", -1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if err := g.AddCounter(duelist, "+1/+1", 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("a new turn, a new draw — anyone's turn: drew %d in total", got-hand)
	}
}

func TestB26AerialExtortionistExilesWithABuybackAndDrawsOnTheBuyback(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Signet", TypeLine: "Artifact", ManaCost: "{2}",
		Owner: opp.ID, Controller: opp.ID,
	})
	land := seedLand(g, opp.ID, "Their Island", "Basic Land — Island", "")
	bird := b20CastCreature(t, g, me, "Aerial Extortionist", "Creature — Bird Soldier", b26AerialExtortionistOracle, 4, 3)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil || p.PickTargetMin != 0 {
		t.Fatalf("the ETB asks for up to one nonland permanent: %+v", p)
	}
	if hasID(p.PickTargetCards, land) || !hasID(p.PickTargetCards, rock) {
		t.Fatal("a land is not a legal pick; the Signet is")
	}
	pickCard(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	if !inExile(g, rock) {
		t.Fatal("the Signet is exiled")
	}
	perm := exiledPermission(g, rock)
	if perm.Player != opp.ID || !perm.WhileExiled || !perm.CastOnly {
		t.Fatalf("its OWNER may cast it for as long as it stays exiled: %+v", perm)
	}
	// The owner buys it back on their turn — a cast from exile, which
	// is a cast from somewhere other than their hand: the Extortionist
	// draws.
	hand := me.Hand.Size()
	advanceToPrecombatMainOf(t, g, 1)
	b06AddMana(opp, "C", "C")
	if err := g.CastSpell(opp.ID, rock, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("the owner casts it from exile: %v", err)
	}
	passPriorityAroundTable(t, g)
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
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("a commander cast draws: drew %d in total", got-hand)
	}
	// Your own cast from anywhere does not, and a hand cast by an
	// opponent does not.
	b13OpponentCasts(t, g, opp, "Their Bear", "Creature — Bear", "", "", nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("a hand cast is not the trigger: drew %d in total", got-hand)
	}
	// Combat damage to a player asks again; with nothing chosen,
	// nothing happens.
	advanceToPrecombatMainOf(t, g, 0)
	attackWith(t, g, opp.ID, bird)
	if p := latestPickTarget(g, me.ID); p == nil {
		t.Fatal("connecting asks for a target again")
	}
	b26DeclinePickTarget(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) {
		t.Error("declined: nothing exiled")
	}
}

func TestB26TheEarthKingMakesABearAndRampsPerBigAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lib := seedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Bayou", TypeLine: "Land — Swamp Forest"},
	)
	king := b20CastCreature(t, g, me, "The Earth King", "Legendary Creature — Human Noble Ally", b26TheEarthKingOracle, 2, 2)
	passPriorityAroundTable(t, g)
	n, bear := b14Tokens(g, me.ID, "Bear")
	if n != 1 || bear.Power != 4 || bear.Toughness != 4 || !bear.HasColor("G") {
		t.Fatalf("one 4/4 green Bear: %d, %+v", n, bear)
	}
	b10Awake(g, bear.InstanceID)
	giant := b12Creature(g, me.ID, "My Giant", "Creature — Giant", 5, 5)
	elf := b12Creature(g, me.ID, "My Elf", "Creature — Elf", 1, 1)
	b10Awake(g, king)
	declareAttack(t, g, opp.ID, bear.InstanceID, giant, elf, king)
	if n := len(g.PendingTriggers) + triggersOnStackFrom(g, king); n != 1 {
		t.Fatalf("one or more big attackers is ONE trigger, got %d", n)
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil || c.SearchMax != 2 || len(c.SearchCards) != 3 {
		t.Fatalf("two attackers with power 4 or more: up to two basics from the three: %+v", c)
	}
	if hasID(c.SearchCards, lib[3]) {
		t.Fatal("a Bayou is not a basic land")
	}
	answerSearchByID(t, g, me.ID, lib[0], lib[1])
	for _, id := range lib[:2] {
		if !g.Battlefield.Contains(id) || !b20Tapped(t, g, id) {
			t.Error("the basics enter tapped")
		}
	}
}

func TestB26ChampionOfThePerishedGrowsWithZombies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	champ := b12Push(g, me.ID, "Champion of the Perished", "Creature — Zombie", b26ChampionOfThePerishedOracle, 1, 1)
	castCatalogSpell(t, g, "Gravecrawler", "Creature — Zombie", "", nil)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, champ, "+1/+1"); got != 1 {
		t.Fatalf("a Zombie entering: one counter, got %d", got)
	}
	castCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, champ, "+1/+1"); got != 1 {
		t.Errorf("a Bear is not a Zombie: still %d", got)
	}
	advanceToPrecombatMainOf(t, g, 1)
	b13OpponentCasts(t, g, opp, "Their Zombie", "Creature — Zombie", "", "", nil)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, champ, "+1/+1"); got != 1 {
		t.Errorf("an opponent's Zombie is not yours: still %d", got)
	}
}

func TestB26CybermanPatrolGivesArtifactCreaturesAfflictThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	patrol := b12Push(g, me.ID, "Cyberman Patrol", "Artifact Creature — Cyberman", b26CybermanPatrolOracle, 2, 2)
	myr := b12Creature(g, me.ID, "My Myr", "Artifact Creature — Myr", 1, 1)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	wall1 := b12Creature(g, opp.ID, "Wall One", "Creature — Wall", 0, 4)
	wall2 := b12Creature(g, opp.ID, "Wall Two", "Creature — Wall", 0, 4)
	wall3 := b12Creature(g, opp.ID, "Wall Three", "Creature — Wall", 0, 4)
	declareAttack(t, g, opp.ID, patrol, myr, bear)
	advanceTo(t, g, game.StepDeclareBlockers)
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
	if n := len(g.PendingTriggers) + triggersOnStackFrom(g, patrol); n != 1 {
		t.Fatalf("the Myr double-blocked is one afflict, the Bear is not an artifact: %d triggers", n)
	}
	// A trigger harvested off a block declaration is drained onto the
	// stack only when the step advances — after combat damage has been
	// dealt on entry to the damage step (see the card comment). The
	// unblocked Patrol connects for 2, then the afflict resolves for 3.
	passPriorityAroundTable(t, g)
	if opp.Life != life-2-3 {
		t.Errorf("2 combat damage and 3 afflict: %d → %d", life, opp.Life)
	}
	if spec, _ := Lookup(b26CybermanPatrolOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the post-damage timing is a declared gap")
	}
}

func TestB26UlvenwaldHydraIsAsBigAsYourLandsAndFetchesOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
	seedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
	lib := seedSearchLibrary(me,
		game.Card{Name: "Bayou", TypeLine: "Land — Swamp Forest"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	hydra := castCatalogSpell(t, g, "Ulvenwald Hydra", "Creature — Hydra", b26UlvenwaldHydraOracle, nil)
	passPriorityAroundTable(t, g)
	assertKeywords(t, g, hydra, "reach")
	if p, tough := effectivePower(t, g, hydra), effectiveToughness(t, g, hydra); p != 2 || tough != 2 {
		t.Fatalf("two lands: 2/2, got %d/%d", p, tough)
	}
	c := searchChoiceFor(g, me.ID)
	if c == nil || len(c.SearchCards) != 1 || c.SearchCards[0] != lib[0] {
		t.Fatalf("the ETB offers the one land card, any land: %+v", c)
	}
	answerSearchByID(t, g, me.ID, lib[0])
	if !g.Battlefield.Contains(lib[0]) || !b20Tapped(t, g, lib[0]) {
		t.Fatal("the land enters tapped")
	}
	if p, tough := effectivePower(t, g, hydra), effectiveToughness(t, g, hydra); p != 3 || tough != 3 {
		t.Errorf("three lands: 3/3, got %d/%d", p, tough)
	}
}

func TestB26CarmenGrowsOnAnySacrificeAndReanimatesOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	carmen := b12Push(g, me.ID, "Carmen, Cruel Skymarcher", "Legendary Creature — Vampire Soldier", b26CarmenTestOracle, 2, 2)
	treasure := pushToken(g, opp.ID, TreasureToken())
	three := b17GraveyardCard(me, "Dead Knight", "Creature — Knight", "{1}{W}{W}")
	four := b17GraveyardCard(me, "Dead Wurm", "Creature — Wurm", "{3}{G}")
	spell := b17GraveyardCard(me, "Dead Bolt", "Instant", "{R}")
	life := me.Life
	// An opponent cracking a Treasure is a player sacrificing a
	// permanent.
	if err := g.ActivateManaAbility(opp.ID, treasure, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("crack Treasure: %v", err)
	}
	b10ResolveAllManaPicks(t, g, opp.ID, "B")
	passPriorityAroundTable(t, g)
	if got := counterCount(g, carmen, "+1/+1"); got != 1 {
		t.Fatalf("one counter, got %d", got)
	}
	if me.Life != life+1 {
		t.Fatalf("and 1 life: %d → %d", life, me.Life)
	}
	// Power 3 now: the Knight is a legal target, the Wurm and the
	// instant are not.
	declareAttack(t, g, opp.ID, carmen)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the attack asks for up to one permanent card")
	}
	if !hasID(p.PickTargetCards, three) || hasID(p.PickTargetCards, four) || hasID(p.PickTargetCards, spell) {
		t.Fatalf("mana value at most Carmen's power (3), permanent cards only: %v", p.PickTargetCards)
	}
	pickCard(t, g, me.ID, three)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(three) {
		t.Error("the Knight returns to the battlefield")
	}
	if c, _ := battlefieldCard(g, three); c.Controller != me.ID {
		t.Error("under your control")
	}
}

func TestB26SylvanAnthemPumpsGreenCreaturesAndScriesWhenOneEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Sylvan Anthem", "Enchantment", b26SylvanAnthemOracle, 0, 0)
	green := b16Creature(g, me.ID, "My Elf", "Creature — Elf", 1, 1, "G")
	red := b16Creature(g, me.ID, "My Goblin", "Creature — Goblin", 1, 1, "R")
	theirs := b16Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1, "G")
	if p := effectivePower(t, g, green); p != 2 {
		t.Errorf("a green creature you control gets +1/+1: power %d", p)
	}
	if p := effectivePower(t, g, red); p != 1 {
		t.Errorf("a red one does not: power %d", p)
	}
	if p := effectivePower(t, g, theirs); p != 1 {
		t.Errorf("an opponent's green creature does not: power %d", p)
	}
	b23Cast(t, g, "Llanowar Elves", "Creature — Elf Druid", "{G}", "", []string{"G"}, nil)
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) == nil {
		t.Fatal("a green creature entering under your control: scry 1")
	}
	b26AnswerScryAllTop(t, g, me.ID)
	b23Cast(t, g, "Goblin Guide", "Creature — Goblin", "{R}", "", []string{"R"}, nil)
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) != nil {
		t.Error("a red creature entering does not")
	}
}

func TestB26ZurFetchesACheapEnchantmentOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zur := b12Push(g, me.ID, "Zur the Enchanter", "Legendary Creature — Human Wizard", b26ZurTheEnchanterOracle, 1, 4)
	lib := seedSearchLibrary(me,
		game.Card{Name: "Rhystic Study", TypeLine: "Enchantment", ManaCost: "{2}{U}"},
		game.Card{Name: "Necropotence", TypeLine: "Enchantment", ManaCost: "{B}{B}{B}"},
		game.Card{Name: "Omniscience", TypeLine: "Enchantment", ManaCost: "{7}{U}{U}{U}"},
		game.Card{Name: "Sol Ring", TypeLine: "Artifact", ManaCost: "{1}"},
	)
	declareAttack(t, g, opp.ID, zur)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil || len(c.SearchCards) != 2 || c.SearchMax != 1 {
		t.Fatalf("enchantment cards with mana value 3 or less, one to take: %+v", c)
	}
	if hasID(c.SearchCards, lib[2]) || hasID(c.SearchCards, lib[3]) {
		t.Fatal("a ten-drop and an artifact are not offered")
	}
	answerSearchByID(t, g, me.ID, lib[1])
	if !g.Battlefield.Contains(lib[1]) {
		t.Error("the enchantment is put onto the battlefield")
	}
	if b20Tapped(t, g, lib[1]) {
		t.Error("untapped — the text says nothing about tapped")
	}
}

func TestB26GenesisChamberHandsEachNontokenCreaturesControllerAMyr(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	chamber := b12Push(g, me.ID, "Genesis Chamber", "Artifact", b26GenesisChamberOracle, 0, 0)
	castCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if n := b05CountTokensNamed(g, me.ID, "Myr"); n != 1 {
		t.Fatalf("your creature, your Myr: got %d", n)
	}
	advanceToPrecombatMainOf(t, g, 1)
	b13OpponentCasts(t, g, opp, "Their Bear", "Creature — Bear", "", "", nil)
	passPriorityAroundTable(t, g)
	if n := b05CountTokensNamed(g, opp.ID, "Myr"); n != 1 {
		t.Fatalf("their creature, their Myr: got %d", n)
	}
	if n := b05CountTokensNamed(g, me.ID, "Myr"); n != 1 {
		t.Fatalf("the Myr is a token and does not re-trigger: still %d", n)
	}
	advanceToPrecombatMainOf(t, g, 0)
	b16Tap(g, chamber)
	castCatalogSpell(t, g, "Another Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if n := b05CountTokensNamed(g, me.ID, "Myr"); n != 1 {
		t.Errorf("a tapped Chamber makes nothing: got %d", n)
	}
}

func TestB26InspiringLeaderPumpsTokensWhileYourCommanderIsOut(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Inspiring Leader", "Legendary Enchantment — Background", b26InspiringLeaderOracle, 0, 0)
	mine := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Soldier", TypeLine: "Token Creature — Soldier",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Soldier", TypeLine: "Token Creature — Soldier",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})
	real := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if p := effectivePower(t, g, mine); p != 1 {
		t.Fatalf("no commander on the battlefield: the ability does not exist, power %d", p)
	}
	cmd := b21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 3, 3)
	if p, tough := effectivePower(t, g, mine), effectiveToughness(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("your token gets +2/+2: %d/%d", p, tough)
	}
	if p := effectivePower(t, g, theirs); p != 1 {
		t.Errorf("an opponent's token does not: power %d", p)
	}
	if p := effectivePower(t, g, real); p != 2 {
		t.Errorf("a nontoken creature does not: power %d", p)
	}
	leaveBattlefield(t, g, me.ID, cmd)
	b21DeclineCommandZone(t, g, me.ID)
	if p := effectivePower(t, g, mine); p != 1 {
		t.Errorf("the commander gone, the anthem is gone: power %d", p)
	}
}

func TestB26ValkyrieHarbingerMakesAnAngelAtEndStepAfterFourLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	valk := b12Push(g, me.ID, "Valkyrie Harbinger", "Creature — Angel Cleric", b26ValkyrieHarbingerOracle, 4, 5)
	assertKeywords(t, g, valk, "flying", "lifelink")
	b14Gain(t, g, me.ID, 3)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if n := b05CountTokensNamed(g, me.ID, "Angel"); n != 0 {
		t.Fatalf("three life is not four: got %d Angels", n)
	}
	// An opponent's end step counts too, and the gain is this turn's.
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	b14Gain(t, g, me.ID, 4)
	advanceToEndStepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	n, angel := b14Tokens(g, me.ID, "Angel")
	if n != 1 || angel.Power != 4 || angel.Toughness != 4 {
		t.Fatalf("one 4/4 Angel: %d, %+v", n, angel)
	}
	assertKeywords(t, g, angel.InstanceID, "flying", "vigilance")
}
