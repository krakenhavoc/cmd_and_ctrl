package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gap2_test.go — card-level coverage for the play-rate slice cut from
// tracker #293's "Gap list 2/2" (removal, charms, a few creatures).

const (
	g2SlashTheRanksOracle     = "ba7f1b09-5727-484d-a502-3dcf4d618c56"
	g2OblivionRingOracle      = "bd9b9772-f5f9-4c6b-913e-7193bea5d0a7"
	g2UnstableObeliskOracle   = "060ae1bb-956a-4264-b405-a33151f31493"
	g2ReclaimOracle           = "797155cd-faf4-4321-8629-c8c352392748"
	g2HelpingHandOracle       = "170a5dbc-6a43-4715-9501-178b1eb7b08c"
	g2BantCharmOracle         = "324889d6-c857-41ed-bb60-408809fc9964"
	g2SultaiCharmOracle       = "46ed38d1-e642-4cea-99ed-a9c17fd982b1"
	g2GazeOfGraniteOracle     = "103d9ad0-d655-4bd5-a899-e9f8869e333d"
	g2PileOnOracle            = "361b0d7f-1e43-45c5-92c2-92baacaf326f"
	g2FecundityOracle         = "ffa64ac6-fe55-48b8-b015-849982cd7ad8"
	g2GlenElendraLiegeOracle  = "946bba74-0951-408c-b06f-167739b10934"
	g2CliffhavenVampireOracle = "1915a311-209b-4628-8122-af3055a78fed"
	g2WaywardServantOracle    = "2916f041-f080-450f-ae79-dc09076d971a"
	g2BeastcallerSavantOracle = "e28227eb-b3b5-42fa-b597-1fefd2a70186"
	g2InkwellLeviathanOracle  = "dbdd0963-b395-4cb8-ad9a-ed85a3e9f2e5"
	g2ValleyMightcallerOracle = "16e9c452-6288-4da8-813d-2eb6b7a538c3"
)

func TestGap2CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		g2SlashTheRanksOracle:     "Slash the Ranks",
		g2OblivionRingOracle:      "Oblivion Ring",
		g2UnstableObeliskOracle:   "Unstable Obelisk",
		g2ReclaimOracle:           "Reclaim",
		g2HelpingHandOracle:       "Helping Hand",
		g2BantCharmOracle:         "Bant Charm",
		g2SultaiCharmOracle:       "Sultai Charm",
		g2GazeOfGraniteOracle:     "Gaze of Granite",
		g2PileOnOracle:            "Pile On",
		g2FecundityOracle:         "Fecundity",
		g2GlenElendraLiegeOracle:  "Glen Elendra Liege",
		g2CliffhavenVampireOracle: "Cliffhaven Vampire",
		g2WaywardServantOracle:    "Wayward Servant",
		g2BeastcallerSavantOracle: "Beastcaller Savant",
		g2InkwellLeviathanOracle:  "Inkwell Leviathan",
		g2ValleyMightcallerOracle: "Valley Mightcaller",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
		} else if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
	}
}

// g2CastFrom casts a card from `seat`'s hand with the given params.
func g2CastFrom(t *testing.T, g *game.Game, seat *game.Player, name, typeLine, oracle string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	id := uuid.New()
	seat.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: seat.ID, Controller: seat.ID})
	if err := g.CastSpell(seat.ID, id, params); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

// --- Slash the Ranks ------------------------------------------------

func TestGap2SlashTheRanksSparesEveryCommander(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Ogre", 4, 4)
	walker := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Walker",
		TypeLine: "Planeswalker — Test", Owner: opp.ID, Controller: opp.ID,
		Counters: map[string]int{game.CounterLoyalty: 4}})
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rock",
		TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID})
	myCmdr := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "My Commander",
		TypeLine: "Legendary Creature — Test", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID, IsCommander: true})
	theirCmdr := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Commander",
		TypeLine: "Legendary Creature — Test", Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID, IsCommander: true})

	castCatalogSpell(t, g, "Slash the Ranks", "Sorcery", g2SlashTheRanksOracle, nil)
	passPriorityAroundTable(t, g)

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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Threat", "Creature — Bear", 4, 4)
	land := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")

	ring := castCatalogSpell(t, g, "Oblivion Ring", "Enchantment", g2OblivionRingOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
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
	spec, _ := Lookup(g2OblivionRingOracle)
	if len(spec.Triggered) == 0 || spec.Triggered[0].Targets == nil {
		t.Fatal("Oblivion Ring declares no target clause")
	}
	if !spec.Triggered[0].Targets.ExcludeSource {
		t.Error("the clause must exclude the ring itself (\"another\")")
	}
}

// --- Unstable Obelisk -----------------------------------------------

func TestGap2UnstableObeliskTapsForColorlessAndDestroysAPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	obelisk := pushCatalogPermanent(g, me.ID, "Unstable Obelisk", "Artifact", g2UnstableObeliskOracle, false)
	victim := pushVanillaCreature(g, opp.ID, "Their Ogre", 4, 4)

	if err := g.ActivateManaAbility(me.ID, obelisk, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Fatalf("pool %v, want [C]", got)
	}
	p1g3Untap(g, obelisk)
	p1g3Activate(t, g, obelisk, 0, game.ActivateAbilityParams{Targets: cardRefs(victim)})
	if g.Battlefield.Contains(victim) {
		t.Error("the target survived")
	}
	if g.Battlefield.Contains(obelisk) {
		t.Error("the Obelisk was not sacrificed")
	}
}

// --- Reclaim --------------------------------------------------------

func TestGap2ReclaimPutsYourCardOnTopAndRefusesAnotherGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushGraveyardCardTyped(me, "My Card", "Sorcery")
	theirs := pushGraveyardCardTyped(opp, "Their Card", "Sorcery")

	if err := castCatalogSpellErr(t, g, "Reclaim", "Instant", g2ReclaimOracle, cardRefs(theirs)); err == nil {
		t.Fatal("Reclaim targeted a card in another player's graveyard")
	}
	castCatalogSpell(t, g, "Reclaim", "Instant", g2ReclaimOracle, cardRefs(mine))
	passPriorityAroundTable(t, g)
	if top, err := me.Library.Top(); err != nil || top.InstanceID != mine {
		t.Error("the card is not on top of the library")
	}
}

// --- Helping Hand ---------------------------------------------------

func TestGap2HelpingHandReturnsASmallCreatureTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	small := pushGraveyardCreature(g, me.ID, "Small Bear", "{1}{G}")
	big := pushGraveyardCreature(g, me.ID, "Big Ogre", "{3}{R}")

	if err := castCatalogSpellErr(t, g, "Helping Hand", "Sorcery", g2HelpingHandOracle, cardRefs(big)); err == nil {
		t.Fatal("Helping Hand targeted a mana value 4 creature")
	}
	castCatalogSpell(t, g, "Helping Hand", "Sorcery", g2HelpingHandOracle, cardRefs(small))
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, small)
	if !ok {
		t.Fatal("the creature did not return")
	}
	if !c.Tapped {
		t.Error("the creature entered untapped")
	}
}

// --- Bant Charm -----------------------------------------------------

func TestGap2BantCharmDestroysAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rock",
		TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID})
	b02cCastModalSpell(t, g, "Bant Charm", "Instant", g2BantCharmOracle, []int{0}, []game.TargetRef{p1g3ModeCard(rock, 0)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
}

func TestGap2BantCharmTucksACreatureUnderItsOwnersLibrary(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// Owned by the opponent, controlled by me: it goes to the owner's library.
	stolen := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Stolen Ogre",
		TypeLine: "Creature — Ogre", Power: 4, Toughness: 4, Owner: opp.ID, Controller: me.ID})
	size := opp.Library.Size()
	b02cCastModalSpell(t, g, "Bant Charm", "Instant", g2BantCharmOracle, []int{1}, []game.TargetRef{p1g3ModeCard(stolen, 0)})
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	them := g.Seats[1]
	toMainForCost(t, g)
	victim := castCatalogSpell(t, g, "Their Instant", "Instant", "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	g2CastFrom(t, g, them, "Bant Charm", "Instant", g2BantCharmOracle,
		game.CastSpellParams{Modes: []int{2}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}})
	passPriorityAroundTable(t, g)
	if !g.Seats[0].Graveyard.Contains(victim) {
		t.Error("the countered instant should be in its owner's graveyard")
	}

	g2 := newCatalogGame(t)
	them2 := g2.Seats[1]
	toMainForCost(t, g2)
	sorcery := castCatalogSpell(t, g2, "Their Sorcery", "Sorcery", "", nil)
	if err := g2.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	id := uuid.New()
	them2.Hand.PushTop(game.Card{InstanceID: id, Name: "Bant Charm", TypeLine: "Instant",
		OracleID: g2BantCharmOracle, Owner: them2.ID, Controller: them2.ID})
	if err := g2.CastSpell(them2.ID, id, game.CastSpellParams{Modes: []int{2},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: sorcery}}}); err == nil {
		t.Error("Bant Charm targeted a sorcery spell")
	}
}

// --- Sultai Charm ---------------------------------------------------

func TestGap2SultaiCharmDestroysOnlyAMonocoloredCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	mono := b16Creature(g, opp.ID, "Mono Bear", "Creature — Bear", 2, 2, "G")
	gold := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Gold Bear",
		TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"G", "W"}, Owner: opp.ID, Controller: opp.ID})

	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Sultai Charm", TypeLine: "Instant",
		OracleID: g2SultaiCharmOracle, Owner: active.ID, Controller: active.ID})
	advanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Modes: []int{0},
		Targets: []game.TargetRef{p1g3ModeCard(gold, 0)}}); err == nil {
		t.Fatal("Sultai Charm targeted a multicolored creature")
	}
	g2CastFrom(t, g, active, "Sultai Charm", "Instant", g2SultaiCharmOracle,
		game.CastSpellParams{Modes: []int{0}, Targets: []game.TargetRef{p1g3ModeCard(mono, 0)}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mono) {
		t.Error("the monocolored creature survived")
	}
	if !g.Battlefield.Contains(gold) {
		t.Error("the multicolored creature died")
	}
}

func TestGap2SultaiCharmDestroysAnEnchantmentAndLoots(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ench := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Aura Thing",
		TypeLine: "Enchantment", Owner: opp.ID, Controller: opp.ID})
	b02cCastModalSpell(t, g, "Sultai Charm", "Instant", g2SultaiCharmOracle, []int{1}, []game.TargetRef{p1g3ModeCard(ench, 0)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ench) {
		t.Error("the enchantment survived")
	}

	hand := me.Hand.Size()
	castCatalogSpellWithModes(t, g, "Sultai Charm", "Instant", g2SultaiCharmOracle, []int{2})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Fatalf("hand %d -> %d, want +2 before the discard", hand, me.Hand.Size())
	}
	if c := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID); c == nil || c.ChooseMin != 1 || c.ChooseMax != 1 {
		t.Error("want a prompt to discard exactly one card after drawing two")
	}
}

// --- Gaze of Granite ------------------------------------------------

func TestGap2GazeOfGraniteDestroysNonlandPermanentsUpToX(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	two := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Two Drop",
		TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Mind Stone",
		TypeLine: "Artifact", ManaCost: "{2}", Owner: me.ID, Controller: me.ID})
	four := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Four Drop",
		TypeLine: "Creature — Ogre", ManaCost: "{3}{R}", Power: 4, Toughness: 4, Owner: opp.ID, Controller: opp.ID})
	land := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")

	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Gaze of Granite", TypeLine: "Sorcery", ManaCost: "{X}{B}{B}{G}",
		OracleID: g2GazeOfGraniteOracle, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 2}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushVanillaCreature(g, opp.ID, "Their Ogre", 4, 4)
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rock",
		TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID})
	if err := castCatalogSpellErr(t, g, "Pile On", "Instant", g2PileOnOracle, cardRefs(rock)); err == nil {
		t.Fatal("Pile On targeted an artifact")
	}
	castCatalogSpell(t, g, "Pile On", "Instant", g2PileOnOracle, cardRefs(victim))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the target survived")
	}
	if c := latestChoiceOfKindFor(g, game.PendingChoiceSurveil, me.ID); c == nil {
		t.Error("no surveil prompt after the destroy")
	}
	spec, _ := Lookup(g2PileOnOracle)
	if spec.TapCost == nil {
		t.Error("Pile On declares no convoke")
	}
}

// --- Fecundity ------------------------------------------------------

func TestGap2FecundityLetsTheDeadCreaturesControllerDraw(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, me.ID, "Fecundity", g2FecundityOracle, "Enchantment")
	// Owned by me, controlled by the opponent: the controller draws.
	stolen := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Stolen Bear",
		TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: opp.ID})
	myHand, oppHand := me.Hand.Size(), opp.Hand.Size()

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(stolen) })
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, opp.ID, true)
	passPriorityAroundTable(t, g)
	if opp.Hand.Size() != oppHand+1 {
		t.Errorf("controller's hand %d -> %d, want +1", oppHand, opp.Hand.Size())
	}
	if me.Hand.Size() != myHand {
		t.Errorf("owner's hand %d -> %d, want unchanged", myHand, me.Hand.Size())
	}
}

func TestGap2FecundityMayDeclineAndIgnoresNoncreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushPermanentForTest(g, me.ID, "Fecundity", g2FecundityOracle, "Enchantment")
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rock",
		TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	hand := me.Hand.Size()

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(rock) })
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want unchanged after declining", hand, me.Hand.Size())
	}
}

// --- Glen Elendra Liege ---------------------------------------------

func TestGap2GlenElendraLiegeBuffsOtherBlueAndBlackCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	liege := pushCatalogPermanent(g, me.ID, "Glen Elendra Liege", "Creature — Faerie Knight", g2GlenElendraLiegeOracle, false)
	blue := b16Creature(g, me.ID, "Blue Bear", "Creature — Bear", 2, 2, "U")
	black := b16Creature(g, me.ID, "Black Bear", "Creature — Bear", 2, 2, "B")
	red := b16Creature(g, me.ID, "Red Bear", "Creature — Bear", 2, 2, "R")
	both := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Murky Bear",
		TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"U", "B"}, Owner: me.ID, Controller: me.ID})
	theirs := b16Creature(g, opp.ID, "Their Blue Bear", "Creature — Bear", 2, 2, "U")

	for name, want := range map[string]struct {
		id uuid.UUID
		p  int
	}{
		"blue": {blue, 3}, "black": {black, 3}, "red": {red, 2}, "blue-black": {both, 4}, "opponent's blue": {theirs, 2},
		"the Liege itself": {liege, 1},
	} {
		if got := effectivePower(t, g, want.id); got != want.p {
			t.Errorf("%s creature power %d, want %d", name, got, want.p)
		}
	}
	if !hasEffectiveKeyword(t, g, liege, "flying") {
		t.Error("the Liege has no flying")
	}
}

// --- Cliffhaven Vampire ---------------------------------------------

func TestGap2CliffhavenVampireDrainsEachOpponentWhenYouGainLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Cliffhaven Vampire", "Creature — Vampire Warrior Ally", g2CliffhavenVampireOracle, false)
	before := []int{g.Seats[1].Life, g.Seats[2].Life, g.Seats[3].Life}

	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 5) })
	passPriorityAroundTable(t, g)
	for i := 1; i < 4; i++ {
		if g.Seats[i].Life != before[i-1]-1 {
			t.Errorf("seat %d life %d, want %d", i, g.Seats[i].Life, before[i-1]-1)
		}
	}
	// Losing life, or an opponent gaining it, does nothing.
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, g.Seats[1].ID, 3) })
	passPriorityAroundTable(t, g)
	if g.Seats[2].Life != before[1]-1 {
		t.Errorf("an opponent's lifegain drained: seat 2 life %d", g.Seats[2].Life)
	}
}

// --- Wayward Servant ------------------------------------------------

func TestGap2WaywardServantDrainsForAnotherZombieOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Wayward Servant", "Creature — Zombie", g2WaywardServantOracle, false)
	oppLife, myLife := opp.Life, me.Life

	castCatalogSpell(t, g, "Plain Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife || me.Life != myLife {
		t.Fatalf("a non-Zombie entering drained: opp %d, me %d", opp.Life, me.Life)
	}
	castCatalogSpell(t, g, "Other Zombie", "Creature — Zombie", "", nil)
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-1 || me.Life != myLife+1 {
		t.Errorf("opp %d (want %d), me %d (want %d)", opp.Life, oppLife-1, me.Life, myLife+1)
	}
}

func TestGap2WaywardServantDoesNotTriggerOnItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oppLife, myLife := opp.Life, me.Life
	castCatalogSpell(t, g, "Wayward Servant", "Creature — Zombie", g2WaywardServantOracle, nil)
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife || me.Life != myLife {
		t.Errorf("the Servant drained on its own entry: opp %d, me %d", opp.Life, me.Life)
	}
}

// --- Beastcaller Savant ---------------------------------------------

func TestGap2BeastcallerSavantManaPaysOnlyForCreatureSpells(t *testing.T) {
	g, me, _ := spendTable(t)
	savant := pushCatalogPermanent(g, me.ID, "Beastcaller Savant", "Creature — Elf Shaman Ally", g2BeastcallerSavantOracle, true)
	rats := pushCatalogPermanent(g, me.ID, "Crypt Rats", "Creature — Rat", cryptRatsOracle, false)

	// It has haste, so a summoning-sick Savant may still tap.
	if err := g.ActivateManaAbility(me.ID, savant, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("tap the Savant for {G}: %v", err)
	}
	cost, _ := game.ParseCost("{G}")
	ratsCard, _ := battlefieldCard(g, rats)
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(ratsCard)) {
		t.Error("Beastcaller Savant mana pays an ability — stronger than printed")
	}
	bear := handSpell(me, "Green Creature", "Creature — Bear", "{G}")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the Savant's {G} for a creature spell: %v", err)
	}
}

// --- Inkwell Leviathan ----------------------------------------------

func TestGap2InkwellLeviathanHasTrampleIslandwalkAndShroud(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogPermanent(g, me.ID, "Inkwell Leviathan", "Artifact Creature — Leviathan", g2InkwellLeviathanOracle, false)
	for _, kw := range []string{"trample", "islandwalk", "shroud"} {
		if !hasEffectiveKeyword(t, g, id, kw) {
			t.Errorf("missing %s", kw)
		}
	}
	// Shroud applies to its own controller too.
	if err := castCatalogSpellErr(t, g, "Pile On", "Instant", g2PileOnOracle, cardRefs(id)); err == nil {
		t.Error("a spell targeted a shroud creature")
	}
}

// --- Valley Mightcaller ---------------------------------------------

func TestGap2ValleyMightcallerGrowsForEachListedTribe(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cc := pushCatalogPermanent(g, me.ID, "Valley Mightcaller", "Creature — Frog Warrior", g2ValleyMightcallerOracle, false)

	for i, tribe := range []string{"Frog", "Rabbit", "Raccoon", "Squirrel"} {
		castCatalogSpell(t, g, "A "+tribe, "Creature — "+tribe, "", nil)
		passPriorityAroundTable(t, g)
		if got := g2PlusOneCounters(g, cc); got != i+1 {
			t.Fatalf("after a %s: %d counters, want %d", tribe, got, i+1)
		}
	}
	castCatalogSpell(t, g, "A Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if got := g2PlusOneCounters(g, cc); got != 4 {
		t.Errorf("a Bear grew it: %d counters, want 4", got)
	}
}

func g2PlusOneCounters(g *game.Game, id uuid.UUID) int {
	c, _ := battlefieldCard(g, id)
	return c.Counters[game.CounterPlusOne]
}
