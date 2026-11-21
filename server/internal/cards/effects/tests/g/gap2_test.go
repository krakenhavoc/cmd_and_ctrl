package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestGap2CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.G2SlashTheRanksOracle:     "Slash the Ranks",
		effectstest.G2OblivionRingOracle:      "Oblivion Ring",
		effectstest.G2UnstableObeliskOracle:   "Unstable Obelisk",
		effectstest.G2ReclaimOracle:           "Reclaim",
		effectstest.G2HelpingHandOracle:       "Helping Hand",
		effectstest.G2BantCharmOracle:         "Bant Charm",
		effectstest.G2SultaiCharmOracle:       "Sultai Charm",
		effectstest.G2GazeOfGraniteOracle:     "Gaze of Granite",
		effectstest.G2PileOnOracle:            "Pile On",
		effectstest.G2FecundityOracle:         "Fecundity",
		effectstest.G2GlenElendraLiegeOracle:  "Glen Elendra Liege",
		effectstest.G2CliffhavenVampireOracle: "Cliffhaven Vampire",
		effectstest.G2WaywardServantOracle:    "Wayward Servant",
		effectstest.G2BeastcallerSavantOracle: "Beastcaller Savant",
		effectstest.G2InkwellLeviathanOracle:  "Inkwell Leviathan",
		effectstest.G2ValleyMightcallerOracle: "Valley Mightcaller",
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
		} else if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
	}
}

// --- Slash the Ranks ------------------------------------------------

func TestGap2SlashTheRanksSparesEveryCommander(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Ogre", 4, 4)
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Walker",
		TypeLine: "Planeswalker — Test", Owner: opp.ID, Controller: opp.ID,
		Counters: map[string]int{game.CounterLoyalty: 4}})
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rock",
		TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID})
	myCmdr := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "My Commander",
		TypeLine: "Legendary Creature — Test", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID, IsCommander: true})
	theirCmdr := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Commander",
		TypeLine: "Legendary Creature — Test", Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID, IsCommander: true})

	effectstest.CastCatalogSpell(t, g, "Slash the Ranks", "Sorcery", effectstest.G2SlashTheRanksOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	for name, id := range map[string]uuid.UUID{"my creature": mine, "their creature": theirs, "their planeswalker": walker} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s survived", name)
		}
	}
	for name, id := range map[string]uuid.UUID{"an artifact": rock, "my commander": myCmdr, "their commander": theirCmdr} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s was destroyed", name)
		}
	}
}

// --- Oblivion Ring --------------------------------------------------

func TestGap2OblivionRingExilesUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Their Threat", "Creature — Bear", 4, 4)
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")

	ring := effectstest.CastCatalogSpell(t, g, "Oblivion Ring", "Enchantment", effectstest.G2OblivionRingOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("the opposing creature is not exiled")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("the land was exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(ring) })
	g.RunStateChecksForTest()
	back := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Their Threat" {
				back++
				if c.Controller != opp.ID {
					t.Error("it returns under its owner's control")
				}
			}
		}
	})
	if back != 1 {
		t.Fatalf("after the ring left, %d copies are on the battlefield, want 1", back)
	}
}

func TestGap2OblivionRingExcludesItself(t *testing.T) {
	spec, _ := effects.Lookup(effectstest.G2OblivionRingOracle)
	if len(spec.Triggered) == 0 || spec.Triggered[0].Targets == nil {
		t.Fatal("Oblivion Ring declares no target clause")
	}
	if !spec.Triggered[0].Targets.ExcludeSource {
		t.Error("the clause must exclude the ring itself (\"another\")")
	}
}

// --- Unstable Obelisk -----------------------------------------------

func TestGap2UnstableObeliskTapsForColorlessAndDestroysAPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	obelisk := effectstest.PushCatalogPermanent(g, me.ID, "Unstable Obelisk", "Artifact", effectstest.G2UnstableObeliskOracle, false)
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Their Ogre", 4, 4)

	if err := g.ActivateManaAbility(me.ID, obelisk, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Fatalf("pool %v, want [C]", got)
	}
	effectstest.P1g3Untap(g, obelisk)
	effectstest.P1g3Activate(t, g, obelisk, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(victim)})
	if g.Battlefield.Contains(victim) {
		t.Error("the target survived")
	}
	if g.Battlefield.Contains(obelisk) {
		t.Error("the Obelisk was not sacrificed")
	}
}

// --- Reclaim --------------------------------------------------------

func TestGap2ReclaimPutsYourCardOnTopAndRefusesAnotherGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushGraveyardCardTyped(me, "My Card", "Sorcery")
	theirs := effectstest.PushGraveyardCardTyped(opp, "Their Card", "Sorcery")

	if err := effectstest.CastCatalogSpellErr(t, g, "Reclaim", "Instant", effectstest.G2ReclaimOracle, effectstest.CardRefs(theirs)); err == nil {
		t.Fatal("Reclaim targeted a card in another player's graveyard")
	}
	effectstest.CastCatalogSpell(t, g, "Reclaim", "Instant", effectstest.G2ReclaimOracle, effectstest.CardRefs(mine))
	effectstest.PassPriorityAroundTable(t, g)
	if top, err := me.Library.Top(); err != nil || top.InstanceID != mine {
		t.Error("the card is not on top of the library")
	}
}

// --- Helping Hand ---------------------------------------------------

func TestGap2HelpingHandReturnsASmallCreatureTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	small := effectstest.PushGraveyardCreature(g, me.ID, "Small Bear", "{1}{G}")
	big := effectstest.PushGraveyardCreature(g, me.ID, "Big Ogre", "{3}{R}")

	if err := effectstest.CastCatalogSpellErr(t, g, "Helping Hand", "Sorcery", effectstest.G2HelpingHandOracle, effectstest.CardRefs(big)); err == nil {
		t.Fatal("Helping Hand targeted a mana value 4 creature")
	}
	effectstest.CastCatalogSpell(t, g, "Helping Hand", "Sorcery", effectstest.G2HelpingHandOracle, effectstest.CardRefs(small))
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, small)
	if !ok {
		t.Fatal("the creature did not return")
	}
	if !c.Tapped {
		t.Error("the creature entered untapped")
	}
}

// --- Bant Charm -----------------------------------------------------

func TestGap2BantCharmDestroysAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rock",
		TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID})
	effectstest.B02cCastModalSpell(t, g, "Bant Charm", "Instant", effectstest.G2BantCharmOracle, []int{0}, []game.TargetRef{effectstest.P1g3ModeCard(rock, 0)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
}

func TestGap2BantCharmTucksACreatureUnderItsOwnersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// Owned by the opponent, controlled by me: it goes to the owner's library.
	stolen := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Stolen Ogre",
		TypeLine: "Creature — Ogre", Power: 4, Toughness: 4, Owner: opp.ID, Controller: me.ID})
	size := opp.Library.Size()
	effectstest.B02cCastModalSpell(t, g, "Bant Charm", "Instant", effectstest.G2BantCharmOracle, []int{1}, []game.TargetRef{effectstest.P1g3ModeCard(stolen, 0)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(stolen) {
		t.Fatal("the creature stayed on the battlefield")
	}
	if opp.Library.Size() != size+1 {
		t.Fatalf("owner's library %d -> %d, want +1", size, opp.Library.Size())
	}
	if opp.Library.Cards[len(opp.Library.Cards)-1].InstanceID != stolen && opp.Library.Cards[0].InstanceID != stolen {
		t.Error("the creature is neither on top nor on the bottom of the library")
	}
}

func TestGap2BantCharmCountersAnInstantButNotASorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	them := g.Seats[1]
	effectstest.ToMainForCost(t, g)
	victim := effectstest.CastCatalogSpell(t, g, "Their Instant", "Instant", "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	effectstest.G2CastFrom(t, g, them, "Bant Charm", "Instant", effectstest.G2BantCharmOracle,
		game.CastSpellParams{Modes: []int{2}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Seats[0].Graveyard.Contains(victim) {
		t.Error("the countered instant should be in its owner's graveyard")
	}

	g2 := effectstest.NewCatalogGame(t)
	them2 := g2.Seats[1]
	effectstest.ToMainForCost(t, g2)
	sorcery := effectstest.CastCatalogSpell(t, g2, "Their Sorcery", "Sorcery", "", nil)
	if err := g2.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	id := uuid.New()
	them2.Hand.PushTop(game.Card{InstanceID: id, Name: "Bant Charm", TypeLine: "Instant",
		OracleID: effectstest.G2BantCharmOracle, Owner: them2.ID, Controller: them2.ID})
	if err := g2.CastSpell(them2.ID, id, game.CastSpellParams{Modes: []int{2},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: sorcery}}}); err == nil {
		t.Error("Bant Charm targeted a sorcery spell")
	}
}

// --- Sultai Charm ---------------------------------------------------

func TestGap2SultaiCharmDestroysOnlyAMonocoloredCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	mono := effectstest.B16Creature(g, opp.ID, "Mono Bear", "Creature — Bear", 2, 2, "G")
	gold := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Gold Bear",
		TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"G", "W"}, Owner: opp.ID, Controller: opp.ID})

	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Sultai Charm", TypeLine: "Instant",
		OracleID: effectstest.G2SultaiCharmOracle, Owner: active.ID, Controller: active.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Modes: []int{0},
		Targets: []game.TargetRef{effectstest.P1g3ModeCard(gold, 0)}}); err == nil {
		t.Fatal("Sultai Charm targeted a multicolored creature")
	}
	effectstest.G2CastFrom(t, g, active, "Sultai Charm", "Instant", effectstest.G2SultaiCharmOracle,
		game.CastSpellParams{Modes: []int{0}, Targets: []game.TargetRef{effectstest.P1g3ModeCard(mono, 0)}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mono) {
		t.Error("the monocolored creature survived")
	}
	if !g.Battlefield.Contains(gold) {
		t.Error("the multicolored creature died")
	}
}

func TestGap2SultaiCharmDestroysAnEnchantmentAndLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ench := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Aura Thing",
		TypeLine: "Enchantment", Owner: opp.ID, Controller: opp.ID})
	effectstest.B02cCastModalSpell(t, g, "Sultai Charm", "Instant", effectstest.G2SultaiCharmOracle, []int{1}, []game.TargetRef{effectstest.P1g3ModeCard(ench, 0)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ench) {
		t.Error("the enchantment survived")
	}

	hand := me.Hand.Size()
	effectstest.CastCatalogSpellWithModes(t, g, "Sultai Charm", "Instant", effectstest.G2SultaiCharmOracle, []int{2})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Fatalf("hand %d -> %d, want +2 before the discard", hand, me.Hand.Size())
	}
	if c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID); c == nil || c.ChooseMin != 1 || c.ChooseMax != 1 {
		t.Error("want a prompt to discard exactly one card after drawing two")
	}
}

// --- Gaze of Granite ------------------------------------------------

func TestGap2GazeOfGraniteDestroysNonlandPermanentsUpToX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	two := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Two Drop",
		TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Mind Stone",
		TypeLine: "Artifact", ManaCost: "{2}", Owner: me.ID, Controller: me.ID})
	four := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Four Drop",
		TypeLine: "Creature — Ogre", ManaCost: "{3}{R}", Power: 4, Toughness: 4, Owner: opp.ID, Controller: opp.ID})
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")

	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Gaze of Granite", TypeLine: "Sorcery", ManaCost: "{X}{B}{B}{G}",
		OracleID: effectstest.G2GazeOfGraniteOracle, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 2}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(two) || g.Battlefield.Contains(rock) {
		t.Error("a permanent with mana value 2 or less survived")
	}
	if !g.Battlefield.Contains(four) {
		t.Error("a mana value 4 creature was destroyed by X = 2")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("a land was destroyed")
	}
}

// --- Pile On --------------------------------------------------------

func TestGap2PileOnDestroysACreatureOrPlaneswalkerAndSurveils(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Their Ogre", 4, 4)
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rock",
		TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID})
	if err := effectstest.CastCatalogSpellErr(t, g, "Pile On", "Instant", effectstest.G2PileOnOracle, effectstest.CardRefs(rock)); err == nil {
		t.Fatal("Pile On targeted an artifact")
	}
	effectstest.CastCatalogSpell(t, g, "Pile On", "Instant", effectstest.G2PileOnOracle, effectstest.CardRefs(victim))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the target survived")
	}
	if c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceSurveil, me.ID); c == nil {
		t.Error("no surveil prompt after the destroy")
	}
	spec, _ := effects.Lookup(effectstest.G2PileOnOracle)
	if spec.TapCost == nil {
		t.Error("Pile On declares no convoke")
	}
}

// --- Fecundity ------------------------------------------------------

func TestGap2FecundityLetsTheDeadCreaturesControllerDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushPermanentForTest(g, me.ID, "Fecundity", effectstest.G2FecundityOracle, "Enchantment")
	// Owned by me, controlled by the opponent: the controller draws.
	stolen := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Stolen Bear",
		TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: opp.ID})
	myHand, oppHand := me.Hand.Size(), opp.Hand.Size()

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(stolen) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, opp.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != oppHand+1 {
		t.Errorf("controller's hand %d -> %d, want +1", oppHand, opp.Hand.Size())
	}
	if me.Hand.Size() != myHand {
		t.Errorf("owner's hand %d -> %d, want unchanged", myHand, me.Hand.Size())
	}
}

func TestGap2FecundityMayDeclineAndIgnoresNoncreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushPermanentForTest(g, me.ID, "Fecundity", effectstest.G2FecundityOracle, "Enchantment")
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rock",
		TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	hand := me.Hand.Size()

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(rock) })
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want unchanged after declining", hand, me.Hand.Size())
	}
}

// --- Glen Elendra Liege ---------------------------------------------

func TestGap2GlenElendraLiegeBuffsOtherBlueAndBlackCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	liege := effectstest.PushCatalogPermanent(g, me.ID, "Glen Elendra Liege", "Creature — Faerie Knight", effectstest.G2GlenElendraLiegeOracle, false)
	blue := effectstest.B16Creature(g, me.ID, "Blue Bear", "Creature — Bear", 2, 2, "U")
	black := effectstest.B16Creature(g, me.ID, "Black Bear", "Creature — Bear", 2, 2, "B")
	red := effectstest.B16Creature(g, me.ID, "Red Bear", "Creature — Bear", 2, 2, "R")
	both := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Murky Bear",
		TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"U", "B"}, Owner: me.ID, Controller: me.ID})
	theirs := effectstest.B16Creature(g, opp.ID, "Their Blue Bear", "Creature — Bear", 2, 2, "U")

	for name, want := range map[string]struct {
		id uuid.UUID
		p  int
	}{
		"blue": {blue, 3}, "black": {black, 3}, "red": {red, 2}, "blue-black": {both, 4}, "opponent's blue": {theirs, 2},
		"the Liege itself": {liege, 1},
	} {
		if got := effectstest.EffectivePower(t, g, want.id); got != want.p {
			t.Errorf("%s creature power %d, want %d", name, got, want.p)
		}
	}
	if !effectstest.HasEffectiveKeyword(t, g, liege, "flying") {
		t.Error("the Liege has no flying")
	}
}

// --- Cliffhaven Vampire ---------------------------------------------

func TestGap2CliffhavenVampireDrainsEachOpponentWhenYouGainLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Cliffhaven Vampire", "Creature — Vampire Warrior Ally", effectstest.G2CliffhavenVampireOracle, false)
	before := []int{g.Seats[1].Life, g.Seats[2].Life, g.Seats[3].Life}

	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 5) })
	effectstest.PassPriorityAroundTable(t, g)
	for i := 1; i < 4; i++ {
		if g.Seats[i].Life != before[i-1]-1 {
			t.Errorf("seat %d life %d, want %d", i, g.Seats[i].Life, before[i-1]-1)
		}
	}
	// Losing life, or an opponent gaining it, does nothing.
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, g.Seats[1].ID, 3) })
	effectstest.PassPriorityAroundTable(t, g)
	if g.Seats[2].Life != before[1]-1 {
		t.Errorf("an opponent's lifegain drained: seat 2 life %d", g.Seats[2].Life)
	}
}

// --- Wayward Servant ------------------------------------------------

func TestGap2WaywardServantDrainsForAnotherZombieOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Wayward Servant", "Creature — Zombie", effectstest.G2WaywardServantOracle, false)
	oppLife, myLife := opp.Life, me.Life

	effectstest.CastCatalogSpell(t, g, "Plain Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife || me.Life != myLife {
		t.Fatalf("a non-Zombie entering drained: opp %d, me %d", opp.Life, me.Life)
	}
	effectstest.CastCatalogSpell(t, g, "Other Zombie", "Creature — Zombie", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-1 || me.Life != myLife+1 {
		t.Errorf("opp %d (want %d), me %d (want %d)", opp.Life, oppLife-1, me.Life, myLife+1)
	}
}

func TestGap2WaywardServantDoesNotTriggerOnItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oppLife, myLife := opp.Life, me.Life
	effectstest.CastCatalogSpell(t, g, "Wayward Servant", "Creature — Zombie", effectstest.G2WaywardServantOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife || me.Life != myLife {
		t.Errorf("the Servant drained on its own entry: opp %d, me %d", opp.Life, me.Life)
	}
}

// --- Beastcaller Savant ---------------------------------------------

func TestGap2BeastcallerSavantManaPaysOnlyForCreatureSpells(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	savant := effectstest.PushCatalogPermanent(g, me.ID, "Beastcaller Savant", "Creature — Elf Shaman Ally", effectstest.G2BeastcallerSavantOracle, true)
	rats := effectstest.PushCatalogPermanent(g, me.ID, "Crypt Rats", "Creature — Rat", effectstest.CryptRatsOracle, false)

	// It has haste, so a summoning-sick Savant may still tap.
	if err := g.ActivateManaAbility(me.ID, savant, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("tap the Savant for {G}: %v", err)
	}
	cost, _ := game.ParseCost("{G}")
	ratsCard, _ := effectstest.BattlefieldCard(g, rats)
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(ratsCard)) {
		t.Error("Beastcaller Savant mana pays an ability — stronger than printed")
	}
	bear := effectstest.HandSpell(me, "Green Creature", "Creature — Bear", "{G}")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the Savant's {G} for a creature spell: %v", err)
	}
}

// --- Inkwell Leviathan ----------------------------------------------

func TestGap2InkwellLeviathanHasTrampleIslandwalkAndShroud(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Inkwell Leviathan", "Artifact Creature — Leviathan", effectstest.G2InkwellLeviathanOracle, false)
	for _, kw := range []string{"trample", "islandwalk", "shroud"} {
		if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
			t.Errorf("missing %s", kw)
		}
	}
	// Shroud applies to its own controller too.
	if err := effectstest.CastCatalogSpellErr(t, g, "Pile On", "Instant", effectstest.G2PileOnOracle, effectstest.CardRefs(id)); err == nil {
		t.Error("a spell targeted a shroud creature")
	}
}

// --- Valley Mightcaller ---------------------------------------------

func TestGap2ValleyMightcallerGrowsForEachListedTribe(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cc := effectstest.PushCatalogPermanent(g, me.ID, "Valley Mightcaller", "Creature — Frog Warrior", effectstest.G2ValleyMightcallerOracle, false)

	for i, tribe := range []string{"Frog", "Rabbit", "Raccoon", "Squirrel"} {
		effectstest.CastCatalogSpell(t, g, "A "+tribe, "Creature — "+tribe, "", nil)
		effectstest.PassPriorityAroundTable(t, g)
		if got := g2PlusOneCounters(g, cc); got != i+1 {
			t.Fatalf("after a %s: %d counters, want %d", tribe, got, i+1)
		}
	}
	effectstest.CastCatalogSpell(t, g, "A Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := g2PlusOneCounters(g, cc); got != 4 {
		t.Errorf("a Bear grew it: %d counters, want 4", got)
	}
}

func g2PlusOneCounters(g *game.Game, id uuid.UUID) int {
	c, _ := effectstest.BattlefieldCard(g, id)
	return c.Counters[game.CounterPlusOne]
}
