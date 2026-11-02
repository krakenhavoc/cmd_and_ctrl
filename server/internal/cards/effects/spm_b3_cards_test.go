package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spm_b3_cards_test.go — Marvel's Spider-Man, batch 3 (#4460). Each test
// plays the real catalog entry and asserts what its text says.

const (
	spm3HydroOracle      = "a93a8edf-2879-4cae-b3b9-b46298270823"
	spm3ImpostorOracle   = "7be47440-b627-411e-a28f-d5a92e5e2fb7"
	spm3WebWatchOracle   = "456dbec4-1b0e-48c4-abae-5c8205651872"
	spm3JackalOracle     = "5998f07e-01f9-4d02-94dd-867c7166463f"
	spm3KapowOracle      = "d6eabbff-cc0c-4000-8feb-4826e1b6bc66"
	spm3KravenHunterOrcl = "3e2bfa3a-ae83-453a-8e3f-ca6205a9af12"
	spm3CatsOracle       = "976db753-9887-448a-8e67-ec7937b6505c"
	spm3PredatorOracle   = "12b8b4bf-ba85-419c-8005-1ec34b2b1deb"
	spm3OctopusOracle    = "8e51fdb8-4f83-4e94-b36c-40a501075ecb"
	spm3BrainOracle      = "b255e13d-01b2-4348-80bd-8d9a8f29be91"
)

func TestSpm3EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		spm3HydroOracle:      "Hydro-Man, Fluid Felon",
		spm3ImpostorOracle:   "Impostor Syndrome",
		spm3WebWatchOracle:   "Interdimensional Web Watch",
		spm3JackalOracle:     "Jackal, Genius Geneticist",
		spm3KapowOracle:      "Kapow!",
		spm3KravenHunterOrcl: "Kraven the Hunter",
		spm3CatsOracle:       "Kraven's Cats",
		spm3PredatorOracle:   "Kraven, Proud Predator",
		spm3OctopusOracle:    "Lady Octopus, Inspired Inventor",
		spm3BrainOracle:      "Living Brain, Mechanical Marvel",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// spm3Hand puts a card in p's hand and returns its id.
func spm3Hand(g *game.Game, p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	g.WithWriteLock(func() { p.Hand.PushTop(c) })
	return c.InstanceID
}

// spm3Cast casts a card from p's hand strictly (so a refusal is real) and
// settles the stack.
func spm3Cast(t *testing.T, g *game.Game, p *game.Player, c game.Card, targets ...uuid.UUID) uuid.UUID {
	t.Helper()
	id := spm3Hand(g, p, c)
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true, FromZone: "hand", Targets: jumpTargets(targets...)}); err != nil {
		t.Fatalf("cast %s: %v", c.Name, err)
	}
	passPriorityAroundTable(t, g)
	return id
}

func spm3Power(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return gaLayered(t, g, id).CurrentPower()
}

// --- Kapow! ---

func TestSpm3KapowCountersThenFights(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushVanillaCreature(g, me.ID, "Fighter", 3, 5)
	theirs := pushVanillaCreature(g, opp.ID, "Victim", 4, 4)
	castCatalogSpell(t, g, "Kapow!", "Sorcery", spm3KapowOracle, jumpTargets(mine, theirs))
	passPriorityAroundTable(t, g)
	c := gaLayered(t, g, mine)
	if c.Counters[game.CounterPlusOne] != 1 {
		t.Fatalf("counters = %d, want 1", c.Counters[game.CounterPlusOne])
	}
	if battlefieldHas(g, theirs) {
		t.Errorf("the counter must precede the fight: 4 damage kills the 4/4")
	}
	if c.DamageMarked != 4 {
		t.Errorf("my creature took %d, want 4", c.DamageMarked)
	}
}

func TestSpm3KapowRefusesTwoOfMyCreaturesAndKeepsTheCounterIfTheOtherTargetLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushVanillaCreature(g, me.ID, "Fighter", 2, 2)
	other := pushVanillaCreature(g, me.ID, "Ally", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Victim", 1, 1)
	advanceToMain(t, g)
	id := spm3Hand(g, me, game.Card{Name: "Kapow!", TypeLine: "Sorcery", OracleID: spm3KapowOracle})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "hand", Targets: jumpTargets(mine, other)}); err == nil {
		t.Fatal("Kapow! accepted my own creature as the fight target")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "hand", Targets: jumpTargets(mine, theirs)}); err != nil {
		t.Fatalf("legal cast: %v", err)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	passPriorityAroundTable(t, g)
	c := gaLayered(t, g, mine)
	if c.Counters[game.CounterPlusOne] != 1 || c.DamageMarked != 0 {
		t.Errorf("counters %d, damage %d: want the counter and no fight", c.Counters[game.CounterPlusOne], c.DamageMarked)
	}
}

// --- Kraven's Cats ---

func TestSpm3CatsPumpOncePerTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cats := b27Push(g, me.ID, "Kraven's Cats", "Creature — Cat Villain", spm3CatsOracle, "{1}{G}", 2, 2, "G")
	advanceToMain(t, g)
	floatMana(t, g, me, "{G}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, cats, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := spm3Power(t, g, cats); got != 4 {
		t.Fatalf("power after one pump = %d, want 4", got)
	}
	floatMana(t, g, me, "{G}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, cats, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("a second activation in one turn was accepted")
	}
}

// --- Kraven, Proud Predator ---

func TestSpm3PredatorPowerIsTheGreatestManaValueYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	k := b27Push(g, me.ID, "Kraven, Proud Predator", "Legendary Creature — Human Warrior Villain", spm3PredatorOracle, "{1}{R}{G}", 0, 4, "R", "G")
	assertKeywords(t, g, k, "vigilance")
	if got := spm3Power(t, g, k); got != 3 {
		t.Errorf("alone: power %d, want his own mana value 3", got)
	}
	b27Push(g, opp.ID, "Their Titan", "Creature — Giant", "", "{9}", 9, 9)
	if got := spm3Power(t, g, k); got != 3 {
		t.Errorf("an opponent's mana value 9 counted: power %d, want 3", got)
	}
	rock := b27Push(g, me.ID, "Big Rock", "Artifact", "", "{6}", 0, 0)
	if got := spm3Power(t, g, k); got != 6 {
		t.Errorf("with a mana value 6 artifact: power %d, want 6", got)
	}
	if c := gaLayered(t, g, k); c.CurrentToughness() != 4 {
		t.Errorf("toughness %d, want the printed 4", c.CurrentToughness())
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(rock) })
	if got := spm3Power(t, g, k); got != 3 {
		t.Errorf("after the rock left: power %d, want 3", got)
	}
}

// --- Living Brain ---

func TestSpm3LivingBrainAnimatesAndUntapsANonEquipmentArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b27Push(g, me.ID, "Living Brain, Mechanical Marvel", "Legendary Artifact Creature — Robot Villain", spm3BrainOracle, "{4}", 3, 3)
	rock := b27Push(g, me.ID, "Mind Stone", "Artifact", "", "{2}", 0, 0)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(rock) })
	advanceTo(t, g, game.StepBeginCombat)
	mscM2PickTargets(t, g, rock)
	c := gaLayered(t, g, rock)
	if !c.IsCreature() || !c.IsArtifact() || c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("target is %s %d/%d, want a 3/3 artifact creature", c.TypeLine, c.CurrentPower(), c.CurrentToughness())
	}
	if c.Tapped {
		t.Errorf("the target was not untapped")
	}
	advanceToUpkeepOf(t, g, 1)
	if c := gaLayered(t, g, rock); c.IsCreature() {
		t.Errorf("still a creature after end of turn")
	}
}

func TestSpm3LivingBrainCannotTargetEquipment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b27Push(g, me.ID, "Living Brain, Mechanical Marvel", "Legendary Artifact Creature — Robot Villain", spm3BrainOracle, "{4}", 3, 3)
	sword := b27Push(g, me.ID, "Test Sword", "Artifact — Equipment", "", "{1}", 0, 0)
	rock := b27Push(g, me.ID, "Mind Stone", "Artifact", "", "{2}", 0, 0)
	advanceTo(t, g, game.StepBeginCombat)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTargets(ch.ID, ch.Chooser, []game.TargetRef{{Kind: game.TargetCard, ID: sword}}); err == nil {
		t.Fatal("an Equipment was accepted as the target")
	}
	if err := g.ResolvePickTargets(ch.ID, ch.Chooser, []game.TargetRef{{Kind: game.TargetCard, ID: rock}}); err != nil {
		t.Fatalf("a Mind Stone was refused: %v", err)
	}
}

// --- Impostor Syndrome ---

func TestSpm3ImpostorSyndromeCopiesANontokenCreatureThatConnects(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Impostor Syndrome", "Enchantment", spm3ImpostorOracle, "{4}{U}{U}", 0, 0, "U")
	legend := b27Push(g, me.ID, "Test Legend", "Legendary Creature — Human", "", "{2}", 2, 2)
	attackWith(t, g, opp.ID, legend)
	passPriorityAroundTable(t, g)
	n, nonLegendaryTokens := 0, 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Test Legend" {
			n++
			if c.IsToken() && !c.IsLegendary() {
				nonLegendaryTokens++
			}
		}
	}
	if n != 2 || nonLegendaryTokens != 1 {
		t.Errorf("%d Test Legends, %d nonlegendary tokens: want 2 and 1", n, nonLegendaryTokens)
	}
}

func TestSpm3ImpostorSyndromeIgnoresTokensAndOpposingCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Impostor Syndrome", "Enchantment", spm3ImpostorOracle, "{4}{U}{U}", 0, 0, "U")
	tok := b27Push(g, me.ID, "Token Bear", "Token Creature — Bear", "", "", 2, 2)
	attackWith(t, g, opp.ID, tok)
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Token Bear"); got != 1 {
		t.Errorf("a token connecting made a copy: %d Token Bears", got)
	}
}

// --- Jackal, Genius Geneticist ---

func TestSpm3JackalCopiesAMatchingCreatureSpellAndGrows(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	jackal := b27Push(g, me.ID, "Jackal, Genius Geneticist", "Legendary Creature — Human Scientist Villain", spm3JackalOracle, "{G}{U}", 1, 1, "G", "U")
	assertKeywords(t, g, jackal, "trample")
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}")
	spm3Cast(t, g, me, game.Card{Name: "Legendary Bear", TypeLine: "Legendary Creature — Bear", ManaCost: "{1}", Power: 1, Toughness: 1})
	if got := b27CountNamed(g, me.ID, "Legendary Bear"); got != 2 {
		t.Fatalf("%d Legendary Bears, want the spell and its copy", got)
	}
	tokens := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Legendary Bear" && c.IsToken() {
			tokens++
			if c.IsLegendary() {
				t.Errorf("the copy is legendary")
			}
		}
	}
	if tokens != 1 {
		t.Errorf("%d token copies, want 1", tokens)
	}
	if got := plusOnes(g, jackal); got != 1 {
		t.Errorf("Jackal has %d counters, want 1", got)
	}
	// Power is now 2: a mana value 1 creature is no longer copied.
	floatMana(t, g, me, "{C}")
	spm3Cast(t, g, me, game.Card{Name: "Plain Bear", TypeLine: "Creature — Bear", ManaCost: "{1}", Power: 1, Toughness: 1})
	if got := b27CountNamed(g, me.ID, "Plain Bear"); got != 1 {
		t.Errorf("%d Plain Bears, want 1 (mana value 1 is not Jackal's power 2)", got)
	}
	if got := plusOnes(g, jackal); got != 1 {
		t.Errorf("Jackal grew without a match: %d", got)
	}
}

func TestSpm3JackalIgnoresNoncreatureSpellsAndOpponents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jackal := b27Push(g, me.ID, "Jackal, Genius Geneticist", "Legendary Creature — Human Scientist Villain", spm3JackalOracle, "{G}{U}", 1, 1, "G", "U")
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}")
	spm3Cast(t, g, me, game.Card{Name: "Artifact One", TypeLine: "Artifact", ManaCost: "{1}"})
	theirs := spm3Hand(g, opp, game.Card{Name: "Their Bear", TypeLine: "Creature — Bear", Keywords: []string{"flash"}, Power: 1, Toughness: 1})
	if err := g.CastSpell(opp.ID, theirs, game.CastSpellParams{Strict: true, FromZone: "hand"}); err != nil {
		t.Fatalf("opponent's flash creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, jackal); got != 0 {
		t.Errorf("Jackal grew on a noncreature spell or an opponent's spell: %d", got)
	}
}

// --- Hydro-Man, Fluid Felon ---

func spm3Hydro(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b27Push(g, owner, "Hydro-Man, Fluid Felon", "Legendary Creature — Elemental Villain", spm3HydroOracle, "{U}{U}", 2, 2, "U")
}

func TestSpm3HydroManGrowsOnYourBlueSpellsWhileACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	h := spm3Hydro(g, me.ID)
	advanceToMain(t, g)
	spm3Cast(t, g, me, game.Card{Name: "Red Bolt", TypeLine: "Instant", Colors: []string{"R"}})
	if got := spm3Power(t, g, h); got != 2 {
		t.Fatalf("a red spell pumped him: power %d", got)
	}
	castFree(t, g, opp, "Their Blue Instant", "Instant")
	spm3Cast(t, g, me, game.Card{Name: "Blue Bolt", TypeLine: "Instant", Colors: []string{"U"}})
	if got := spm3Power(t, g, h); got != 3 {
		t.Errorf("after one blue spell: power %d, want 3", got)
	}
	spm3Cast(t, g, me, game.Card{Name: "Blue Bolt Two", TypeLine: "Instant", Colors: []string{"U"}})
	if got := spm3Power(t, g, h); got != 4 {
		t.Errorf("after two blue spells: power %d, want 4", got)
	}
}

func TestSpm3HydroManBecomesAManaLandUntilYourNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	h := spm3Hydro(g, me.ID)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(h) })
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	c := gaLayered(t, g, h)
	if !c.IsLand() || c.IsCreature() {
		t.Fatalf("after my end step he is %q, want a land and not a creature", c.TypeLine)
	}
	if !c.IsLegendary() {
		t.Errorf("he lost his legendary supertype")
	}
	if c.Tapped {
		t.Errorf("he was not untapped")
	}
	// A blue spell while he is a land does not pump him (the intervening if).
	spm3Cast(t, g, me, game.Card{Name: "Blue Bolt", TypeLine: "Instant", Colors: []string{"U"}})
	if got := gaLayered(t, g, h).Power; got == 3 {
		t.Errorf("a blue spell pumped Hydro-Man while he is a land")
	}
	// He taps for {U} straight away (a land is not summoning sick).
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	tapForMana(t, g, me.ID, h)
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "U" {
		t.Errorf("pool %v, want [U]", got)
	}
	// Still a land through an opponent's turn, a creature again on mine.
	advanceToUpkeepOf(t, g, 1)
	if c := gaLayered(t, g, h); !c.IsLand() || c.IsCreature() {
		t.Errorf("on the opponent's turn he is %q, want still a land", c.TypeLine)
	}
	advanceToUpkeepOf(t, g, 0)
	if c := gaLayered(t, g, h); c.IsLand() || !c.IsCreature() {
		t.Errorf("on my next turn he is %q, want a creature again", c.TypeLine)
	}
}

// --- Interdimensional Web Watch ---

func TestSpm3WebWatchExilesTwoAndItsManaPaysOnlyForSpellsFromExile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		for _, n := range []string{"Exiled Rock", "Exiled Gem"} {
			me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: n, TypeLine: "Artifact", ManaCost: "{2}", Owner: me.ID, Controller: me.ID})
		}
	})
	advanceToMain(t, g)
	castFromHandForTest(t, g, me, "Interdimensional Web Watch", "Artifact", "", spm3WebWatchOracle, game.CastSpellParams{FromZone: "hand"})
	passPriorityAroundTable(t, g)
	var rock uuid.UUID
	exiled := 0
	for _, c := range g.Exile.Cards {
		if c.Name == "Exiled Rock" || c.Name == "Exiled Gem" {
			exiled++
			if c.Name == "Exiled Rock" {
				rock = c.InstanceID
			}
		}
	}
	if exiled != 2 {
		t.Fatalf("%d cards exiled, want 2", exiled)
	}
	watch := findBattlefieldByName(g, "Interdimensional Web Watch")
	if err := g.ActivateManaAbility(me.ID, watch, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	for c := riderLatestManaPick(g, me.ID); c != nil; c = riderLatestManaPick(g, me.ID) {
		if err := g.ResolveManaChoice(c.ID, me.ID, "U"); err != nil {
			t.Fatalf("ResolveManaChoice: %v", err)
		}
	}
	if got := batch01PoolColors(me); len(got) != 2 {
		t.Fatalf("pool %v, want two mana", got)
	}
	// The mana cannot pay for a spell from the hand.
	handSpell := spm3Hand(g, me, game.Card{Name: "Hand Rock", TypeLine: "Artifact", ManaCost: "{2}"})
	if err := g.CastSpell(me.ID, handSpell, game.CastSpellParams{Strict: true, FromZone: "hand"}); err == nil {
		t.Fatal("restricted mana paid for a spell cast from the hand")
	}
	// It does pay for the exiled card.
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("cast from exile with the Watch's mana: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("pool still holds %d mana", me.ManaPool.Total())
	}
}

// --- Lady Octopus, Inspired Inventor ---

func TestSpm3OctopusCountsOnlyYourFirstTwoDrawsEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lady := b27Push(g, me.ID, "Lady Octopus, Inspired Inventor", "Legendary Creature — Human Scientist Villain", spm3OctopusOracle, "{U}", 0, 2, "U")
	advanceToMain(t, g)
	// The turn's draw step drew the first card before Lady Octopus existed
	// (no trigger), so my next draw is the second and the one after is the
	// third.
	draws := func(p *game.Player, n int) {
		for i := 0; i < n; i++ {
			if err := g.DrawCard(p.ID); err != nil {
				t.Fatalf("DrawCard: %v", err)
			}
			passPriorityAroundTable(t, g)
		}
	}
	draws(opp, 3)
	before := countersOn(g, lady, "ingenuity")
	draws(me, 1)
	mid := countersOn(g, lady, "ingenuity")
	draws(me, 3)
	after := countersOn(g, lady, "ingenuity")
	if before != 0 || mid != 1 || after != 1 {
		t.Errorf("ingenuity counters: after opponent draws %d, my second draw %d, after two more draws %d; want 0, 1, 1", before, mid, after)
	}
}

func TestSpm3OctopusCastsAnArtifactFromHandFreeUpToItsCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lady := b27Push(g, me.ID, "Lady Octopus, Inspired Inventor", "Legendary Creature — Human Scientist Villain", spm3OctopusOracle, "{U}", 0, 2, "U")
	battlefieldCardPtr(t, g, lady).SummonedThisTurn = false
	advanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(lady, "ingenuity", 2) })
	cheap := spm3Hand(g, me, game.Card{Name: "Cheap Gadget", TypeLine: "Artifact", ManaCost: "{2}"})
	pricey := spm3Hand(g, me, game.Card{Name: "Pricey Gadget", TypeLine: "Artifact", ManaCost: "{3}"})
	bear := spm3Hand(g, me, game.Card{Name: "Cheap Bear", TypeLine: "Creature — Bear", ManaCost: "{1}"})
	if err := g.ActivateCatalogAbility(me.ID, lady, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	ch := latestChoiceOfKind(g, game.PendingChoiceResolutionCast)
	if ch == nil {
		t.Fatal("no free-cast offer was made")
	}
	for _, id := range ch.DiscardOptions {
		if id == pricey || id == bear {
			t.Errorf("offered %s: not an artifact with mana value 2 or less", id)
		}
	}
	if len(ch.DiscardOptions) != 1 || ch.DiscardOptions[0] != cheap {
		t.Fatalf("offered %v, want only the cheap artifact", ch.DiscardOptions)
	}
	if err := g.CastForResolutionChoice(ch.Chooser, ch.ID, cheap, game.CastSpellParams{FromZone: "hand"}); err != nil {
		t.Fatalf("free cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Cheap Gadget") == uuid.Nil {
		t.Errorf("the free artifact did not resolve onto the battlefield")
	}
}

// --- Kraven the Hunter ---

func spm3Hunter(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b27Push(g, owner, "Kraven the Hunter", "Legendary Creature — Human Warrior Villain", spm3KravenHunterOrcl, "{1}{B}{G}", 4, 3, "B", "G")
}

func TestSpm3KravenTheHunterDrawsWhenTheirBiggestCreatureDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	k := spm3Hunter(g, me.ID)
	small := b12Creature(g, opp.ID, "Small", "Creature — Bear", 1, 1)
	big := b12Creature(g, opp.ID, "Big", "Creature — Giant", 4, 4)
	hand := me.Hand.Size()
	b27Kill(g, small)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || plusOnes(g, k) != 0 {
		t.Fatalf("a creature without the greatest power triggered him (hand %d -> %d, counters %d)", hand, me.Hand.Size(), plusOnes(g, k))
	}
	b27Kill(g, big)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || plusOnes(g, k) != 1 {
		t.Errorf("hand %d -> %d, counters %d: want a card and a counter", hand, me.Hand.Size(), plusOnes(g, k))
	}
}

func TestSpm3KravenTheHunterTiesAndOwnCreaturesAndSimultaneousDeaths(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	k := spm3Hunter(g, me.ID)
	mine := b12Creature(g, me.ID, "My Giant", "Creature — Giant", 5, 5)
	b27Kill(g, mine)
	passPriorityAroundTable(t, g)
	if plusOnes(g, k) != 0 {
		t.Fatalf("my own creature dying triggered him")
	}
	// Two tied creatures: each qualifies, so each death fires.
	a := b12Creature(g, opp.ID, "Twin A", "Creature — Giant", 3, 3)
	b := b12Creature(g, opp.ID, "Twin B", "Creature — Giant", 3, 3)
	b27Kill(g, a)
	passPriorityAroundTable(t, g)
	b27Kill(g, b)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, k); got != 2 {
		t.Errorf("two tied giants dying one after the other gave %d counters, want 2", got)
	}
	// Dying together, only the greatest power qualifies.
	c := b12Creature(g, opp.ID, "Mid", "Creature — Bear", 2, 2)
	d := b12Creature(g, opp.ID, "Top", "Creature — Giant", 6, 6)
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{c, d}) })
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, k); got != 3 {
		t.Errorf("a 2/2 and a 6/6 dying together gave %d counters in total, want 3 (only the 6/6 counts)", got)
	}
}
