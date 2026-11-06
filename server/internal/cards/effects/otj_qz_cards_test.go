package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// otj_qz_cards_test.go — Outlaws of Thunder Junction (OTJ), the Q–Z
// batch (#4468): the cards that need no new engine seam. Each test plays
// the real catalog entry through the engine and asserts what the printed
// text says.

const (
	otjQZRavenOfFellOmensOracle  = "eec7d4f4-4141-49ed-8978-aefacc0a6609"
	otjQZRecklessLackeyOracle    = "73733fa5-d292-40ec-8f17-05ba5feb40ab"
	otjQZRedrockSentinelOracle   = "51a6819d-c8fb-4f52-ab0c-4ae845728ed2"
	otjQZScorchingShotOracle     = "e13113ba-1923-493c-a48d-ca320b7adbde"
	otjQZSeizeTheSecretsOracle   = "db070312-9a27-435f-8b23-e538441103ec"
	otjQZSterlingKeykeeperOracle = "f893d3d6-efef-4394-8e15-e01deed72b4f"
	otjQZSterlingHoundOracle     = "3178c1f9-f644-4926-ac9a-d40e4db79955"
	otjQZSterlingSupplierOracle  = "ab23cca2-b59e-4d34-80de-98bd08985275"
	otjQZTreasureDredgerOracle   = "9a691e96-fb47-4c23-a063-3b287c51147b"
	otjQZVaultPlundererOracle    = "80e5d768-cc1f-48f8-b0fe-ed974e4d1a56"
	otjQZVoraciousVarmintOracle  = "bd785a7e-df65-48e6-843b-7831f3c1e5d3"
	otjQZWantedGriffinOracle     = "c1348afe-4dc7-41bb-9d3f-1e8751abd7db"
)

func TestOtjQZEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		otjQZRavenOfFellOmensOracle:  "Raven of Fell Omens",
		otjQZRecklessLackeyOracle:    "Reckless Lackey",
		otjQZRedrockSentinelOracle:   "Redrock Sentinel",
		otjQZScorchingShotOracle:     "Scorching Shot",
		otjQZSeizeTheSecretsOracle:   "Seize the Secrets",
		otjQZSterlingKeykeeperOracle: "Sterling Keykeeper",
		otjQZSterlingHoundOracle:     "Sterling Hound",
		otjQZSterlingSupplierOracle:  "Sterling Supplier",
		otjQZTreasureDredgerOracle:   "Treasure Dredger",
		otjQZVaultPlundererOracle:    "Vault Plunderer",
		otjQZVoraciousVarmintOracle:  "Voracious Varmint",
		otjQZWantedGriffinOracle:     "Wanted Griffin",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Raven of Fell Omens -------------------------------------------------------

func TestOtjQZRavenOfFellOmensDrainsEachOpponentOnTheFirstCrimeEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	raven := b12Push(g, me.ID, "Raven of Fell Omens", "Creature — Bird", otjQZRavenOfFellOmensOracle, 1, 2)
	if !hasEffectiveKeyword(t, g, raven, "flying") {
		t.Error("flying is printed")
	}
	mine := b16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	advanceToMain(t, g)

	start := me.Life
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetCard(mine))
	passPriorityAroundTable(t, g)
	if me.Life != start {
		t.Fatalf("targeting your own creature is no crime: your life is %d", me.Life)
	}

	before := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		before[s.ID] = s.Life
	}
	otjB1BoltOpponent(t, g, opp.ID)
	if me.Life != before[me.ID]+1 {
		t.Errorf("you gain 1 life: %d → %d", before[me.ID], me.Life)
	}
	for _, s := range g.Seats[1:] {
		want := before[s.ID] - 1
		if s.ID == opp.ID {
			want -= 3 // the Bolt itself
		}
		if s.Life != want {
			t.Errorf("opponent %s: life %d → %d, want %d", s.ID, before[s.ID], s.Life, want)
		}
	}

	life := me.Life
	otjB1BoltOpponent(t, g, opp.ID)
	if me.Life != life {
		t.Error("only once each turn: a second crime drained again")
	}
}

// --- Reckless Lackey -----------------------------------------------------------

func TestOtjQZRecklessLackeySacrificesItselfToDrawAndMakeATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Drawn")
	lackey := b12Push(g, me.ID, "Reckless Lackey", "Creature — Goblin Pirate", otjQZRecklessLackeyOracle, 1, 2)
	for _, kw := range []string{"first strike", "haste"} {
		if !hasEffectiveKeyword(t, g, lackey, kw) {
			t.Errorf("%s is printed", kw)
		}
	}
	advanceToMain(t, g)
	b06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, lackey, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without the {R}")
	}
	b06AddMana(me, "R")
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, lackey, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(lackey) || !me.Graveyard.Contains(lackey) {
		t.Fatal("the Lackey is sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if got := b16CountNamed(g, "Treasure"); got != 1 {
		t.Errorf("%d Treasures, want 1", got)
	}
}

// --- Redrock Sentinel ----------------------------------------------------------

func TestOtjQZRedrockSentinelSacrificesALandToDrawAndMakeATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Drawn")
	sentinel := b12Push(g, me.ID, "Redrock Sentinel", "Artifact Creature — Golem", otjQZRedrockSentinelOracle, 2, 4)
	if !hasEffectiveKeyword(t, g, sentinel, "defender") {
		t.Error("defender is printed")
	}
	land := b12Permanent(g, me.ID, "My Mountain", "Basic Land — Mountain")
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	advanceToMain(t, g)
	b06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, sentinel, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{rock}}); err == nil {
		t.Fatal("an artifact paid \"Sacrifice a land\"")
	}
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, sentinel, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{land}})
	if g.Battlefield.Contains(land) {
		t.Error("the land is sacrificed as the cost")
	}
	if !b16Tapped(t, g, sentinel) {
		t.Error("{T}: the Sentinel taps to activate")
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if got := b16CountNamed(g, "Treasure"); got != 1 {
		t.Errorf("%d Treasures, want 1", got)
	}
}

// --- Scorching Shot ------------------------------------------------------------

func TestOtjQZScorchingShotDealsFiveToACreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	five := b12Creature(g, opp.ID, "Their Five", "Creature — Giant", 5, 5)
	six := b12Creature(g, opp.ID, "Their Six", "Creature — Giant", 6, 6)
	if err := castCatalogSpellErr(t, g, "Scorching Shot", "Sorcery", otjQZScorchingShotOracle, b16TargetPlayer(opp.ID)); err == nil {
		t.Fatal("a player is not a target creature")
	}
	castCatalogSpell(t, g, "Scorching Shot", "Sorcery", otjQZScorchingShotOracle, b16TargetCard(five))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(five) {
		t.Error("5 damage to a 5/5: it dies")
	}
	castCatalogSpell(t, g, "Scorching Shot", "Sorcery", otjQZScorchingShotOracle, b16TargetCard(six))
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, six)
	if !ok {
		t.Fatal("5 damage to a 6/6: it survives")
	}
	if c.DamageMarked != 5 {
		t.Errorf("the 6/6 has %d damage, want 5", c.DamageMarked)
	}
}

// --- Seize the Secrets ---------------------------------------------------------

func TestOtjQZSeizeTheSecretsCostsOneLessAfterACrimeAndDrawsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	advanceToMain(t, g)
	if got := selfPricedMV(t, g, me, otjQZSeizeTheSecretsOracle, "Sorcery", "{2}{U}"); got != 3 {
		t.Errorf("no crime: %d, want 3", got)
	}
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetCard(mine))
	passPriorityAroundTable(t, g)
	if got := selfPricedMV(t, g, me, otjQZSeizeTheSecretsOracle, "Sorcery", "{2}{U}"); got != 3 {
		t.Errorf("targeting your own creature is no crime: %d, want 3", got)
	}
	otjB1BoltOpponent(t, g, opp.ID)
	if got := selfPricedMV(t, g, me, otjQZSeizeTheSecretsOracle, "Sorcery", "{2}{U}"); got != 2 {
		t.Errorf("after a crime: %d, want 2", got)
	}

	seedLibrary(me, "One", "Two", "Three")
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Seize the Secrets", "Sorcery", otjQZSeizeTheSecretsOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}

	advanceToMainOf(t, g, 1)
	advanceToMainOf(t, g, 0)
	if got := selfPricedMV(t, g, me, otjQZSeizeTheSecretsOracle, "Sorcery", "{2}{U}"); got != 3 {
		t.Errorf("a crime on an earlier turn does not count: %d, want 3", got)
	}
}

// --- Sterling Keykeeper -------------------------------------------------------

func TestOtjQZSterlingKeykeeperTapsANonMountCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	keeper := b12Push(g, me.ID, "Sterling Keykeeper", "Creature — Human Mercenary", otjQZSterlingKeykeeperOracle, 2, 2)
	mount := b12Creature(g, opp.ID, "Their Horse", "Creature — Horse Mount", 2, 2)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	b06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, keeper, 0, game.ActivateAbilityParams{Strict: true, Targets: b16TargetCard(mount)}); err == nil {
		t.Fatal("a Mount is not a non-Mount creature")
	}
	b16Activate(t, g, me.ID, keeper, 0, game.ActivateAbilityParams{Strict: true, Targets: b16TargetCard(bear)})
	if !b16Tapped(t, g, keeper) {
		t.Error("{T}: the Keykeeper taps to activate")
	}
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, bear) {
		t.Error("the Bear is tapped")
	}
	if b16Tapped(t, g, mount) {
		t.Error("the Mount was tapped")
	}
}

// --- Sterling Hound ------------------------------------------------------------

func TestOtjQZSterlingHoundSurveilsTwoOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	msc18Library(me, 10)
	advanceToMain(t, g)
	msc17Play(t, g, me, game.Card{Name: "Sterling Hound", TypeLine: "Artifact Creature — Dog",
		ManaCost: "{3}", OracleID: otjQZSterlingHoundOracle, Power: 3, Toughness: 2})
	prompt := surveilChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no surveil prompt when the Hound entered")
	}
	if len(prompt.ScryCards) != 2 {
		t.Errorf("surveil 2 looks at %d cards", len(prompt.ScryCards))
	}
}

// --- Sterling Supplier ---------------------------------------------------------

func TestOtjQZSterlingSupplierPutsACounterOnAnotherCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "My Other Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	supplier := castHoldCreature(t, g, "Sterling Supplier", "Creature — Bird Soldier", otjQZSterlingSupplierOracle, 3, 4)
	if !hasEffectiveKeyword(t, g, supplier, "flying") {
		t.Error("flying is printed")
	}
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if hasID(p.PickTargetCards, supplier) || hasID(p.PickTargetCards, theirs) || !hasID(p.PickTargetCards, bear) {
		t.Fatalf("only another creature you control is offered: %v", p.PickTargetCards)
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if pw, tt := ptOf(t, g, bear); pw != 3 || tt != 3 {
		t.Errorf("the Bear is %d/%d, want 3/3", pw, tt)
	}
	if pw, tt := ptOf(t, g, supplier); pw != 3 || tt != 4 {
		t.Errorf("the Supplier is %d/%d, want an uncountered 3/4", pw, tt)
	}
}

// --- Treasure Dredger ----------------------------------------------------------

func TestOtjQZTreasureDredgerPaysOneLifeForATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dredger := b12Push(g, me.ID, "Treasure Dredger", "Creature — Human Rogue", otjQZTreasureDredgerOracle, 2, 2)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, dredger, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without the {1}")
	}
	b06AddMana(me, "C")
	life := me.Life
	b16Activate(t, g, me.ID, dredger, 0, game.ActivateAbilityParams{})
	if me.Life != life-1 {
		t.Errorf("pay 1 life: %d → %d", life, me.Life)
	}
	if !b16Tapped(t, g, dredger) {
		t.Error("{T}: the Dredger taps to activate")
	}
	passPriorityAroundTable(t, g)
	if got := b16CountNamed(g, "Treasure"); got != 1 {
		t.Errorf("%d Treasures, want 1", got)
	}
}

// --- Vault Plunderer -----------------------------------------------------------

func TestOtjQZVaultPlundererMakesTargetPlayerDrawAndLoseOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(opp, "Their Drawn")
	castHoldCreature(t, g, "Vault Plunderer", "Creature — Human Rogue", otjQZVaultPlundererOracle, 3, 1)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if !hasID(p.PickTargetPlayers, me.ID) || !hasID(p.PickTargetPlayers, opp.ID) {
		t.Fatalf("any player may be the target: %v", p.PickTargetPlayers)
	}
	hand, life, mine := opp.Hand.Size(), opp.Life, me.Life
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("pick the opponent: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := opp.Hand.Size() - hand; got != 1 {
		t.Errorf("the opponent drew %d, want 1", got)
	}
	if opp.Life != life-1 {
		t.Errorf("the opponent loses 1 life: %d → %d", life, opp.Life)
	}
	if me.Life != mine {
		t.Errorf("your life changed: %d → %d", mine, me.Life)
	}
}

// --- Voracious Varmint ---------------------------------------------------------

func TestOtjQZVoraciousVarmintSacrificesItselfToDestroyAnArtifactOrEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	varmint := b12Push(g, me.ID, "Voracious Varmint", "Creature — Varmint", otjQZVoraciousVarmintOracle, 2, 2)
	if !hasEffectiveKeyword(t, g, varmint, "vigilance") {
		t.Error("vigilance is printed")
	}
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	b06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, varmint, 0, game.ActivateAbilityParams{Targets: b16TargetCard(bear)}); err == nil {
		t.Fatal("a creature is not an artifact or enchantment")
	}
	b16Activate(t, g, me.ID, varmint, 0, game.ActivateAbilityParams{Targets: b16TargetCard(rock)})
	if g.Battlefield.Contains(varmint) {
		t.Error("the Varmint is sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !opp.Graveyard.Contains(rock) {
		t.Error("the artifact is destroyed")
	}
}

// --- Wanted Griffin ------------------------------------------------------------

func TestOtjQZWantedGriffinMakesAMercenaryWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	griffin := b12Push(g, me.ID, "Wanted Griffin", "Creature — Griffin", otjQZWantedGriffinOracle, 3, 2)
	if !hasEffectiveKeyword(t, g, griffin, "flying") {
		t.Error("flying is printed")
	}
	if len(otjMPMercenaries(g, me.ID)) != 0 {
		t.Fatal("a Mercenary before the Griffin died")
	}
	p1Destroy(t, g, griffin)
	passPriorityAroundTable(t, g)
	otjMPOneMercenary(t, g, me.ID)
}
