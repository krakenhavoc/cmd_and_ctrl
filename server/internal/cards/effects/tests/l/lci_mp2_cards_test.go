package testsl

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lci_mp2_cards_test.go — The Lost Caverns of Ixalan, cards M–P, second
// batch (#4470), plus Queen's Bay Paladin and Relic's Roar. Each test
// plays the real catalog entry and asserts what its text says.

const (
	lciMP2GlyphOracle     = "483a0caa-e912-41e9-aaa3-bcea9369ec06"
	lciMP2CaveDiverOracle = "2f1019df-3e76-446c-ba33-f9dfe56b10bf"
	lciMP2CollapseOracle  = "f0d1636e-6f80-44fe-869a-1749ab249815"
	lciMP2SirenOracle     = "a06a35a9-a23e-4a0a-90bc-84e7a52d2cef"
	lciMP2OltecOracle     = "b19c4bbf-258f-4436-8ef9-fc905c283776"
	lciMP2PetrifyOracle   = "ebec110e-aff3-48aa-abf6-309dd2ffb478"
	lciMP2HatOracle       = "b9080653-87cb-4443-a4a3-36637d5ad165"
	lciMP2PoeticOracle    = "9780d975-acbe-40c6-b444-1e1e6b772788"
	lciMP2HammerOracle    = "da989677-3647-46b8-aa56-b2cfb60f7a2a"
	lciMP2PaladinOracle   = "60b18d94-8ae5-4508-b4db-18f92c2d39fd"
	lciMP2RoarOracle      = "7887f060-3024-4dc4-aeb5-fb5546db6b48"
)

func TestLciMP2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		lciMP2GlyphOracle:     "Malamet Battle Glyph",
		lciMP2CaveDiverOracle: "Merfolk Cave-Diver",
		lciMP2CollapseOracle:  "Molten Collapse",
		lciMP2SirenOracle:     "Oaken Siren",
		lciMP2OltecOracle:     "Oltec Archaeologists",
		lciMP2PetrifyOracle:   "Petrify",
		lciMP2HatOracle:       "Pirate Hat",
		lciMP2PoeticOracle:    "Poetic Ingenuity",
		lciMP2HammerOracle:    "Pugnacious Hammerskull",
		lciMP2PaladinOracle:   "Queen's Bay Paladin",
		lciMP2RoarOracle:      "Relic's Roar",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Malamet Battle Glyph ---------------------------------------------------------

// A creature that entered this turn gets the counter and wins the fight a
// 2/2 would lose.
func TestLciMP2BattleGlyphCountersACreatureThatEnteredThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	fresh := effectstest.Msc17Play(t, g, me, game.Card{Name: "Fresh Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	theirs := effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 2, 3)
	effectstest.CastCatalogSpell(t, g, "Malamet Battle Glyph", "Sorcery", lciMP2GlyphOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: fresh}, {Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, fresh, game.CounterPlusOne); n != 1 {
		t.Errorf("+1/+1 counters on the creature that entered this turn: %d, want 1", n)
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("the 2/3 survived a fight with a 3/3")
	}
	if !g.Battlefield.Contains(fresh) {
		t.Error("the 3/3 died to a 2-power creature")
	}
}

// A creature that was already there gets no counter, so the 2/2 loses.
func TestLciMP2BattleGlyphGivesNoCounterToAnOlderCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	old := effectstest.B12Creature(g, me.ID, "Old Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 2, 3)
	effectstest.CastCatalogSpell(t, g, "Malamet Battle Glyph", "Sorcery", lciMP2GlyphOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: old}, {Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, old, game.CounterPlusOne); n != 0 {
		t.Errorf("+1/+1 counters on a creature that did not enter this turn: %d, want 0", n)
	}
	if g.Battlefield.Contains(old) || !g.Battlefield.Contains(theirs) {
		t.Error("the fight did not go 2/2 against 2/3")
	}
}

// --- Merfolk Cave-Diver -----------------------------------------------------------

// lciMP2Explore has `explorer` explore under `controller` with a land on
// top of the library, so the land goes to hand and no question is asked.
func lciMP2Explore(g *game.Game, controller *game.Player, explorer uuid.UUID) {
	effectstest.SeedLibraryTop(controller, "Forest", "Basic Land — Forest", "")
	g.WithWriteLock(func() {
		_ = g.ExploreForEffect(uuid.Nil, game.ObjectRef{ID: explorer}, controller.ID)
	})
}

func TestLciMP2CaveDiverPumpsAndCantBeBlockedWhenACreatureYouControlExplores(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	diver := effectstest.B12Push(g, me.ID, "Merfolk Cave-Diver", "Creature — Merfolk Scout", lciMP2CaveDiverOracle, 2, 4)
	scout := effectstest.B12Creature(g, me.ID, "Scout", "Creature — Human Scout", 1, 1)
	blocker := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	lciMP2Explore(g, me, scout)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.S58p6View(t, g, diver); c.CurrentPower() != 3 || c.CurrentToughness() != 4 {
		t.Errorf("Cave-Diver is %d/%d, want 3/4", c.CurrentPower(), c.CurrentToughness())
	}
	effectstest.BrAttack(t, g, diver)
	if err := g.DeclareBlocker(blocker, diver); err == nil {
		t.Error("a creature blocked the Cave-Diver after a creature its controller controls explored")
	}
}

func TestLciMP2CaveDiverIgnoresAnOpponentsExplore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	diver := effectstest.B12Push(g, me.ID, "Merfolk Cave-Diver", "Creature — Merfolk Scout", lciMP2CaveDiverOracle, 2, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Scout", "Creature — Human Scout", 1, 1)
	effectstest.AdvanceToMain(t, g)
	lciMP2Explore(g, opp, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.S58p6View(t, g, diver); c.CurrentPower() != 2 {
		t.Errorf("Cave-Diver has power %d after an opponent's creature explored, want 2", c.CurrentPower())
	}
}

// --- Molten Collapse --------------------------------------------------------------

func TestLciMP2MoltenCollapseIsChooseOneWithoutDescending(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	trinket := effectstest.PushTypedCard(g, opp.ID, "Trinket", "Artifact", "{1}")
	id := effectstest.HandCardFull(me, "Molten Collapse", "Sorcery", "{B}{R}", lciMP2CollapseOracle, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0, 1}, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: bear, Mode: 0},
		{Kind: game.TargetCard, ID: trinket, Mode: 1},
	}}); err == nil {
		t.Fatal("both bullets were accepted without descending this turn")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0}, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: bear, Mode: 0},
	}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the targeted creature was not destroyed")
	}
	if !g.Battlefield.Contains(trinket) {
		t.Error("the unchosen bullet destroyed the artifact")
	}
}

func TestLciMP2MoltenCollapseChoosesBothAfterYouDescended(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mine) })
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	trinket := effectstest.PushTypedCard(g, opp.ID, "Trinket", "Artifact", "{1}")
	relic := effectstest.PushTypedCard(g, opp.ID, "Relic", "Artifact", "{2}")
	id := effectstest.HandCardFull(me, "Molten Collapse", "Sorcery", "{B}{R}", lciMP2CollapseOracle, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0, 1}, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: bear, Mode: 0},
		{Kind: game.TargetCard, ID: relic, Mode: 1},
	}}); err == nil {
		t.Fatal("the second bullet accepted a mana value 2 artifact")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0, 1}, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: bear, Mode: 0},
		{Kind: game.TargetCard, ID: trinket, Mode: 1},
	}}); err != nil {
		t.Fatalf("both bullets after descending: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || g.Battlefield.Contains(trinket) {
		t.Error("both bullets should have destroyed their targets")
	}
	if !g.Battlefield.Contains(relic) {
		t.Error("the untargeted artifact was destroyed")
	}
}

// --- Oaken Siren ------------------------------------------------------------------

func TestLciMP2OakenSirenTapsForBlueOnlyForArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	siren := effectstest.B12Push(g, me.ID, "Oaken Siren", "Artifact Creature — Siren Pirate", lciMP2SirenOracle, 1, 2)
	for _, kw := range []string{"flying", "vigilance"} {
		if !effectstest.HasEffectiveKeyword(t, g, siren, kw) {
			t.Errorf("Oaken Siren has no %s", kw)
		}
	}
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if err := g.ActivateManaAbility(me.ID, siren, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Siren: %v", err)
	}
	if me.ManaPool.CountColor("U") != 1 || me.ManaPool.Total() != 1 {
		t.Fatalf("the Siren added %d mana, want {U}", me.ManaPool.Total())
	}
	cost, _ := game.ParseCost("{U}")
	rock := game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{U}"}
	bear := game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{U}"}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(rock)) || !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(rock)) {
		t.Error("the mana refused an artifact spell or an artifact's ability")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(bear)) || me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(bear)) {
		t.Error("the mana paid for a nonartifact spell or ability")
	}
}

// --- Oltec Archaeologists ---------------------------------------------------------

func lciMP2PlayArchaeologists(t *testing.T, g *game.Game, mode int) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Oltec Archaeologists", TypeLine: "Creature — Human Artificer Scout",
		OracleID: lciMP2OltecOracle, Power: 4, Toughness: 4})
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if pick == nil {
		t.Fatalf("no mode_pick prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(pick.ID, me.ID, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
}

func TestLciMP2OltecArchaeologistsReturnsAnArtifactCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	relic := effectstest.GyCard(me, "Old Relic", "Artifact", "{2}", "")
	bear := effectstest.GyCard(me, "Dead Bear", "Creature — Bear", "{1}{G}", "")
	lciMP2PlayArchaeologists(t, g, 0)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
		t.Error("a creature card was accepted as the artifact card")
	}
	effectstest.PickCard(t, g, me.ID, relic)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(relic) || me.Graveyard.Contains(relic) {
		t.Error("the artifact card did not return to hand")
	}
	if !me.Graveyard.Contains(bear) {
		t.Error("the creature card left the graveyard")
	}
}

func TestLciMP2OltecArchaeologistsScriesThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "One", "Two", "Three", "Four")
	lciMP2PlayArchaeologists(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	scry := effectstest.ScryChoiceFor(g, me.ID)
	if scry == nil {
		t.Fatal("no scry prompt")
	}
	if len(scry.ScryCards) != 3 {
		t.Errorf("scry looked at %d cards, want 3", len(scry.ScryCards))
	}
}

// --- Petrify ----------------------------------------------------------------------

func TestLciMP2PetrifyStopsACreatureAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	aura := effectstest.CastCatalogSpell(t, g, "Petrify", effectstest.AuraTypeLine, lciMP2PetrifyOracle, effectstest.FinABCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.AttachmentHostOf(t, g, aura); got.Kind != game.TargetCard || got.ID != bear {
		t.Fatalf("Petrify is attached to %+v, want the creature", got)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err == nil {
		t.Error("the enchanted creature attacked")
	}
}

func TestLciMP2PetrifyLocksAnArtifactsManaAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ring := effectstest.PushCatalogPermanent(g, me.ID, "Sol Ring", "Artifact", effectstest.FinSZ2SolRingOracle, false)
	if err := effectstest.CastCatalogSpellErr(t, g, "Petrify", effectstest.AuraTypeLine, lciMP2PetrifyOracle,
		effectstest.FinABCard(effectstest.PushTypedCard(g, me.ID, "Plain Land", "Land", ""))); err == nil {
		t.Error("Petrify was cast at a land")
	}
	effectstest.CastCatalogSpell(t, g, "Petrify", effectstest.AuraTypeLine, lciMP2PetrifyOracle, effectstest.FinABCard(ring))
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ActivateManaAbility(me.ID, ring, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("the enchanted artifact's mana ability was activated")
	}
}

// --- Pirate Hat -------------------------------------------------------------------

func TestLciMP2PirateHatEquipsPiratesForOneAndLootsOnTheAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	hat := effectstest.PushCatalogPermanent(g, me.ID, "Pirate Hat", "Artifact — Equipment", lciMP2HatOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	pirate := effectstest.B12Creature(g, me.ID, "Deckhand", "Creature — Human Pirate", 2, 2)
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	effectstest.AdvanceToMain(t, g)

	spec, _ := effects.Lookup(lciMP2HatOracle)
	if len(spec.Activated) != 2 || spec.Activated[0].Label != "Equip Pirate {1}" || spec.Activated[1].Label != "Equip {2}" {
		t.Fatalf("activated abilities %+v, want Equip Pirate {1} and Equip {2}", spec.Activated)
	}
	effectstest.FillPool(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, hat, 0, game.ActivateAbilityParams{
		Strict: true, Targets: effectstest.FinABCard(bear),
	}); err == nil {
		t.Fatal("Equip Pirate attached the Hat to a non-Pirate")
	}
	if err := g.ActivateCatalogAbility(me.ID, hat, 0, game.ActivateAbilityParams{
		Strict: true, Targets: effectstest.FinABCard(pirate),
	}); err != nil {
		t.Fatalf("Equip Pirate {1}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.AttachmentHostOf(t, g, hat); got.ID != pirate {
		t.Fatalf("the Hat is attached to %+v, want the Pirate", got)
	}
	if c := effectstest.S58p6View(t, g, pirate); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("the equipped Pirate is %d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}

	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, pirate)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d → %d, want one card drawn before the discard", hand, me.Hand.Size())
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d → %d, want the loot to net zero", hand, me.Hand.Size())
	}
}

// --- Poetic Ingenuity -------------------------------------------------------------

func TestLciMP2PoeticIngenuityMakesATreasurePerAttackingDinosaur(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, me.ID, "Poetic Ingenuity", "Enchantment", lciMP2PoeticOracle, 0, 0)
	a := effectstest.B12Creature(g, me.ID, "Raptor A", "Creature — Dinosaur", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Raptor B", "Creature — Dinosaur", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, a, b, bear)
	if n := effectstest.TriggersOnStackFrom(g, lciMP2FirstNamed(g, "Poetic Ingenuity")); n != 1 {
		t.Errorf("%d attack triggers for one declaration, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.TrekB33Tokens(g, me.ID, "Treasure")); n != 2 {
		t.Errorf("%d Treasures for two attacking Dinosaurs, want 2", n)
	}
}

func TestLciMP2PoeticIngenuityMakesOneDinosaurATurnForArtifactSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Poetic Ingenuity", "Enchantment", lciMP2PoeticOracle, 0, 0)
	effectstest.CastCatalogSpell(t, g, "Bauble", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	dinos := effectstest.TrekB33Tokens(g, me.ID, "Dinosaur")
	if len(dinos) != 1 {
		t.Fatalf("%d Dinosaur tokens after an artifact spell, want 1", len(dinos))
	}
	if c := effectstest.S58p6View(t, g, dinos[0].InstanceID); c.CurrentPower() != 3 || c.CurrentToughness() != 1 || !c.HasColor("R") {
		t.Errorf("token is %d/%d %v, want a 3/1 red Dinosaur", c.CurrentPower(), c.CurrentToughness(), c.Effective().Colors)
	}
	effectstest.CastCatalogSpell(t, g, "Second Bauble", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.TrekB33Tokens(g, me.ID, "Dinosaur")); n != 1 {
		t.Errorf("%d Dinosaur tokens after a second artifact spell this turn, want 1", n)
	}
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.TrekB33Tokens(g, me.ID, "Dinosaur")); n != 1 {
		t.Errorf("a creature spell made a Dinosaur: %d", n)
	}
}

// lciMP2FirstNamed is the first battlefield card with that name.
func lciMP2FirstNamed(g *game.Game, name string) uuid.UUID {
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			return c.InstanceID
		}
	}
	return uuid.Nil
}

// --- Pugnacious Hammerskull -------------------------------------------------------

func TestLciMP2HammerskullIsStunnedWhenItAttacksAlone(t *testing.T) {
	for _, tc := range []struct {
		name     string
		otherDin bool
		want     int
	}{
		{"no other Dinosaur", false, 1},
		{"another Dinosaur", true, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.ToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			ham := effectstest.B12Push(g, me.ID, "Pugnacious Hammerskull", "Creature — Dinosaur", lciMP2HammerOracle, 6, 6)
			effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
			if tc.otherDin {
				effectstest.B12Creature(g, me.ID, "Raptor", "Creature — Dinosaur", 2, 2)
			}
			effectstest.DeclareAttack(t, g, opp.ID, ham)
			effectstest.PassPriorityAroundTable(t, g)
			if n := effectstest.CountersOn(g, ham, game.CounterStun); n != tc.want {
				t.Errorf("stun counters %d, want %d", n, tc.want)
			}
		})
	}
}

// --- Queen's Bay Paladin ----------------------------------------------------------

func TestLciMP2PaladinReturnsAVampireWithAFinalityCounterOnTheAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	paladin := effectstest.PushCatalogPermanent(g, me.ID, "Queen's Bay Paladin", "Creature — Vampire Knight", lciMP2PaladinOracle, false)
	vamp := effectstest.PushGraveyardPermanent(me, "Old Vampire", "Creature — Vampire", "{2}{B}")
	ogre := effectstest.PushGraveyardPermanent(me, "Ogre", "Creature — Ogre", "{1}{R}")
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	life := me.Life
	if err := g.DeclareAttacker(paladin, opp.ID); err != nil {
		t.Fatalf("attack: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the attack trigger")
	}
	if effectstest.HasID(p.PickTargetCards, ogre) {
		t.Error("a non-Vampire card was offered")
	}
	effectstest.PickCard(t, g, me.ID, vamp)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(vamp) {
		t.Fatal("the Vampire did not come back")
	}
	if n := effectstest.CountersOn(g, vamp, game.CounterFinality); n != 1 {
		t.Errorf("finality counters %d, want 1", n)
	}
	if me.Life != life-3 {
		t.Errorf("life %d → %d, want -3 for a mana value 3 card", life, me.Life)
	}
	if !me.Graveyard.Contains(ogre) {
		t.Error("the Ogre left the graveyard")
	}
}

func TestLciMP2PaladinTriggersAsItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	vamp := effectstest.PushGraveyardPermanent(me, "Old Vampire", "Creature — Vampire", "{B}")
	life := me.Life
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Queen's Bay Paladin", TypeLine: "Creature — Vampire Knight",
		OracleID: lciMP2PaladinOracle, Power: 5, Toughness: 4})
	if effectstest.LatestPickTarget(g, me.ID) == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	effectstest.PickCard(t, g, me.ID, vamp)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(vamp) || effectstest.CountersOn(g, vamp, game.CounterFinality) != 1 {
		t.Error("the enters trigger did not return the Vampire with a finality counter")
	}
	if me.Life != life-1 {
		t.Errorf("life %d → %d, want -1", life, me.Life)
	}
}

// --- Relic's Roar -----------------------------------------------------------------

func TestLciMP2RelicsRoarMakesAnArtifactAFourThreeDinosaurUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Relic's Roar", "Instant", lciMP2RoarOracle, effectstest.FinABCard(rock))
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, rock)
	eff := c.Effective()
	if !slices.Contains(eff.Types, "Artifact") || !slices.Contains(eff.Types, "Creature") || !c.HasSubtype("Dinosaur") {
		t.Errorf("types %v %v, want a Dinosaur artifact creature", eff.Types, eff.Subtypes)
	}
	if eff.Power != 4 || eff.Toughness != 3 {
		t.Errorf("%d/%d, want 4/3", eff.Power, eff.Toughness)
	}
	effectstest.EndTurn(t, g)
	if effectstest.LayeredCard(t, g, rock).IsCreature() {
		t.Error("the artifact is still a creature after the turn")
	}
}
