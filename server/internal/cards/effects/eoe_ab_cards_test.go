package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// eoe_ab_cards_test.go — Edge of Eternities, the cards from A to B
// (#4461). Each test plays the real catalog entry and asserts what its
// text says.

const (
	eoeABStalkerOracle    = "da9ec010-8b42-4927-b602-4e96bced4e99"
	eoeABAlpharaelOracle  = "e4be4071-7910-460f-90f3-6c4d8cc9db65"
	eoeABAnnulOracle      = "d08e9784-75f7-4164-ac48-d06160f8c56b"
	eoeABCharmOracle      = "edcb0248-e7c9-4df9-8dba-982048654fd2"
	eoeABReclaimerOracle  = "bb956492-e3ed-483a-9169-5b795a8e619e"
	eoeABGreenhouseOracle = "52c6afc6-aff1-46a8-b1ea-b183facec1ae"
	eoeABMicrosizerOracle = "048c3b0b-48ef-42f5-a2af-90aa1d133296"
	eoeABBoostersOracle   = "ffa3b20b-d77b-43d5-86ac-6a591df1877f"
	eoeABProspectorOracle = "bbe740f8-6e7c-4cfd-9134-9bf8d5f48cf5"
	eoeABQuietOracle      = "c96424e7-63c4-4f07-959f-6ad09f91279a"
	eoeABEngineerOracle   = "b395925b-2456-454e-9262-3b55c9f2799a"
	eoeABBurstOracle      = "3abd2342-f783-4871-ac78-1a54b35722c2"
	eoeABSpecialistOracle = "78de77ef-c325-4342-869f-042921229577"
	eoeABBladeOracle      = "33042d12-f08b-4b3e-b106-c424ef868c13"
	eoeABStingerOracle    = "ab282bb1-900e-40b8-9821-8e2de88927bc"
	eoeABBombardOracle    = "c0899c12-5462-4ff1-a9e5-d99cffa7c6cc"
	eoeABZealotOracle     = "31a0c9a2-c9bd-47f7-872b-a43fb64d873c"
	eoeABBroodguardOracle = "c90da077-de8c-4402-abfc-b63dd15853da"
)

// Every card in the batch, pinned by oracle id to its name and its
// declared completeness.
func TestEoeABEveryCardIsRegistered(t *testing.T) {
	for oracle, name := range map[string]string{
		eoeABStalkerOracle:    "All-Fates Stalker",
		eoeABAlpharaelOracle:  "Alpharael, Dreaming Acolyte",
		eoeABAnnulOracle:      "Annul",
		eoeABCharmOracle:      "Archenemy's Charm",
		eoeABReclaimerOracle:  "Astelli Reclaimer",
		eoeABGreenhouseOracle: "Atmospheric Greenhouse",
		eoeABMicrosizerOracle: "Atomic Microsizer",
		eoeABBoostersOracle:   "Auxiliary Boosters",
		eoeABProspectorOracle: "Beamsaw Prospector",
		eoeABQuietOracle:      "Beyond the Quiet",
		eoeABEngineerOracle:   "Biomechan Engineer",
		eoeABBurstOracle:      "Biosynthic Burst",
		eoeABSpecialistOracle: "Biotech Specialist",
		eoeABBladeOracle:      "Blade of the Swarm",
		eoeABStingerOracle:    "Blooming Stinger",
		eoeABBombardOracle:    "Bombard",
		eoeABZealotOracle:     "Brightspear Zealot",
		eoeABBroodguardOracle: "Broodguard Elite",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		want := CompletenessFull
		if spec.Name != name || spec.Completeness != want {
			t.Errorf("%s: name %q, completeness %s, want %s", name, spec.Name, spec.Completeness, want)
		}
	}
	for oracle, cost := range map[string]string{
		eoeABStalkerOracle:    "{1}{W}",
		eoeABReclaimerOracle:  "{2}{W}",
		eoeABBroodguardOracle: "{X}{G}",
	} {
		if ac := game.AlternativeCostByKey(oracle, "warp"); ac == nil || ac.ManaCost != cost {
			t.Errorf("%s: warp %+v, want %s", oracle, ac, cost)
		}
	}
}

// --- fixtures ------------------------------------------------------------------

// eoeABPickIfAsked answers the open target prompt with id when there is
// one: a trigger with a single legal target may take it without asking.
func eoeABPickIfAsked(t *testing.T, g *game.Game, chooser, id uuid.UUID) {
	t.Helper()
	if latestPickTarget(g, chooser) != nil {
		pickCard(t, g, chooser, id)
	}
}

// eoeABNamedCount counts the battlefield permanents named name that owner
// controls.
func eoeABNamedCount(g *game.Game, owner uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == owner {
			n++
		}
	}
	return n
}

// eoeABOfferedCards is the set of cards the open target prompt offers.
func eoeABOfferedCards(t *testing.T, g *game.Game) map[uuid.UUID]bool {
	t.Helper()
	ch := socB37OpenPickTarget(g)
	if ch == nil {
		t.Fatal("no target prompt")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range ch.PickTargetCards {
		offered[id] = true
	}
	return offered
}

// --- All-Fates Stalker -----------------------------------------------------------

func TestEoeABAllFatesStalkerExilesANonAssassinUntilItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	assassin := b12Creature(g, opp.ID, "Their Assassin", "Creature — Human Assassin", 2, 2)

	stalker := avB1ResolveCreature(t, g, me, "All-Fates Stalker", "Creature — Drix Assassin", eoeABStalkerOracle)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the entry trigger")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, cardRefs(assassin)); err == nil {
		t.Fatal("an Assassin is not a legal target")
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !g.Exile.Contains(bear) {
		t.Fatal("the non-Assassin should be in exile")
	}
	if !g.Battlefield.Contains(assassin) {
		t.Error("the Assassin stays")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(stalker) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if avB1Named(g, opp.ID, "Their Bear") == uuid.Nil {
		t.Fatal("the exiled creature returns when the Stalker leaves")
	}
}

// --- Alpharael, Dreaming Acolyte -------------------------------------------------

func TestEoeABAlpharaelDrawsTwoThenDiscardsTwoOrAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rock := handCardFull(me, "Mind Stone", "Artifact", "{2}", "", nil)
	bear := handCardFull(me, "Bear", "Creature — Bear", "{1}{G}", "", nil)
	hand := me.Hand.Size()

	avB1ResolveCreature(t, g, me, "Alpharael, Dreaming Acolyte", "Legendary Creature — Human Cleric", eoeABAlpharaelOracle)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Fatalf("hand %d → %d: Alpharael was added and cast (net 0), then drew two (+2)", hand, got)
	}
	c := discardChoiceFor(g, me.ID)
	if c == nil || c.ChooseMin != 1 || c.ChooseMax != 2 {
		t.Fatalf("a floor of one and a ceiling of two: %+v", c)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{bear}); err == nil {
		t.Fatal("one card that is not an artifact was accepted")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{rock}); err != nil {
		t.Fatalf("one artifact card is a complete discard: %v", err)
	}
	if me.Hand.Contains(rock) || !me.Hand.Contains(bear) {
		t.Error("only the artifact card was discarded")
	}
}

func TestEoeABAlpharaelHasDeathtouchOnlyDuringItsControllersTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := b12Push(g, me.ID, "Alpharael, Dreaming Acolyte", "Legendary Creature — Human Cleric", eoeABAlpharaelOracle, 2, 3)
	theirs := b12Push(g, opp.ID, "Alpharael, Dreaming Acolyte", "Legendary Creature — Human Cleric", eoeABAlpharaelOracle, 2, 3)
	if !effectiveAbilitiesContain(t, g, mine, "deathtouch") {
		t.Error("during your turn, Alpharael has deathtouch")
	}
	if effectiveAbilitiesContain(t, g, theirs, "deathtouch") {
		t.Error("an Alpharael whose controller is not the active player has no deathtouch")
	}
}

// --- Annul ---------------------------------------------------------------------

func TestEoeABAnnulCountersAnArtifactSpellButNotAnInstant(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	rock := avB1Hand(me, game.Card{Name: "Mind Stone", TypeLine: "Artifact", ManaCost: "{2}"})
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the artifact: %v", err)
	}
	shock := avB1Hand(me, game.Card{Name: "Shock", TypeLine: "Instant", OracleID: "a9d288b8-cdc1-4e55-a0c9-d6edfc95e65d"})
	if err := g.CastSpell(me.ID, shock, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("cast the instant: %v", err)
	}
	refused := avB1Hand(me, game.Card{Name: "Annul", TypeLine: "Instant", OracleID: eoeABAnnulOracle})
	if err := g.CastSpell(me.ID, refused, game.CastSpellParams{Targets: cardRefs(shock)}); err == nil {
		t.Fatal("an instant spell is not an artifact or enchantment spell")
	}
	annul := avB1Hand(me, game.Card{Name: "Annul", TypeLine: "Instant", OracleID: eoeABAnnulOracle})
	if err := g.CastSpell(me.ID, annul, game.CastSpellParams{Targets: cardRefs(rock)}); err != nil {
		t.Fatalf("Annul on the artifact spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !me.Graveyard.Contains(rock) {
		t.Error("the artifact spell is countered into its owner's graveyard")
	}
}

// --- Archenemy's Charm -------------------------------------------------------------

func TestEoeABArchenemysCharmExilesACreatureOrPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	walker := b12Permanent(g, opp.ID, "Their Walker", "Legendary Planeswalker — Jace")
	if err := g.AddCounter(walker, game.CounterLoyalty, 3); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	land := pushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	if _, err := castModalErr(t, g, "Archenemy's Charm", "Instant", eoeABCharmOracle, []int{0}, cardRefs(land)); err == nil {
		t.Fatal("a land is not a creature or planeswalker")
	}
	castModal(t, g, "Archenemy's Charm", "Instant", eoeABCharmOracle, []int{0}, cardRefs(walker))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(walker) || !g.Exile.Contains(walker) {
		t.Error("the planeswalker is exiled")
	}
}

func TestEoeABArchenemysCharmReturnsTwoCreatureOrPlaneswalkerCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushGraveyardPermanent(me, "Old Bear", "Creature — Bear", "{1}{G}")
	walker := pushGraveyardPermanent(me, "Old Walker", "Legendary Planeswalker — Jace", "{2}{U}{U}")
	spell := pushGraveyardPermanent(me, "Old Sorcery", "Sorcery", "{1}{B}")
	if _, err := castModalErr(t, g, "Archenemy's Charm", "Instant", eoeABCharmOracle, []int{1}, cardRefs(bear, spell)); err == nil {
		t.Fatal("a sorcery card is not a creature or planeswalker card")
	}
	castModal(t, g, "Archenemy's Charm", "Instant", eoeABCharmOracle, []int{1}, cardRefs(bear, walker))
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) || !me.Hand.Contains(walker) {
		t.Error("both cards return to hand")
	}
	if !me.Graveyard.Contains(spell) {
		t.Error("the sorcery stays in the graveyard")
	}
}

func TestEoeABArchenemysCharmGrowsYourCreatureAndGivesItLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if _, err := castModalErr(t, g, "Archenemy's Charm", "Instant", eoeABCharmOracle, []int{2}, cardRefs(theirs)); err == nil {
		t.Fatal("the counters go only on a creature you control")
	}
	castModal(t, g, "Archenemy's Charm", "Instant", eoeABCharmOracle, []int{2}, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if got := b12Counter(t, g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters %d, want 2", got)
	}
	if !effectiveAbilitiesContain(t, g, bear, "lifelink") {
		t.Error("it gains lifelink until end of turn")
	}
}

// --- Astelli Reclaimer ---------------------------------------------------------------

// eoeABCastReclaimer casts Astelli Reclaimer paying exactly its cost out
// of the pool, for the warp cost or the printed one.
func eoeABCastReclaimer(t *testing.T, g *game.Game, me *game.Player, warp bool) uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	id := avB1Hand(me, game.Card{Name: "Astelli Reclaimer", TypeLine: "Creature — Angel Warrior", ManaCost: "{3}{W}{W}",
		OracleID: eoeABReclaimerOracle, Power: 5, Toughness: 4})
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	params := game.CastSpellParams{Strict: true}
	n := 5
	if warp {
		params.AlternativeCost = "warp"
		n = 3
	}
	fillPoolColored(me, "W", n)
	if err := g.CastSpell(me.ID, id, params); err != nil {
		t.Fatalf("cast Astelli Reclaimer: %v", err)
	}
	passPriorityAroundTable(t, g)
	return id
}

func TestEoeABAstelliReclaimerReachesTheManaSpentForItsWarpCost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	three := pushGraveyardPermanent(me, "Three Relic", "Artifact", "{3}")
	four := pushGraveyardPermanent(me, "Four Relic", "Artifact", "{4}")
	pushGraveyardPermanent(me, "Old Bear", "Creature — Bear", "{1}")
	reclaimer := eoeABCastReclaimer(t, g, me, true)
	if !effectiveAbilitiesContain(t, g, reclaimer, "flying") {
		t.Error("printed flying is missing")
	}
	if socB37OpenPickTarget(g) != nil {
		offered := eoeABOfferedCards(t, g)
		if !offered[three] || offered[four] {
			t.Fatalf("three mana spent: the {3} relic and not the {4} one, got %v", offered)
		}
		pickCard(t, g, me.ID, three)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(three) || !me.Graveyard.Contains(four) {
		t.Error("the {3} relic returns to the battlefield and the {4} relic stays")
	}
}

func TestEoeABAstelliReclaimerReturnsOnlyANoncreatureNonlandPermanentCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	three := pushGraveyardPermanent(me, "Three Relic", "Artifact", "{3}")
	four := pushGraveyardPermanent(me, "Four Charm", "Enchantment", "{4}")
	six := pushGraveyardPermanent(me, "Six Relic", "Artifact", "{6}")
	bear := pushGraveyardPermanent(me, "Old Bear", "Creature — Bear", "{1}")
	land := pushGraveyardPermanent(me, "Old Plains", "Basic Land — Plains", "")
	spell := pushGraveyardPermanent(me, "Old Sorcery", "Sorcery", "{1}")
	theirs := pushGraveyardPermanent(opp, "Their Relic", "Artifact", "{1}")
	eoeABCastReclaimer(t, g, me, false)
	offered := eoeABOfferedCards(t, g)
	for id, want := range map[uuid.UUID]bool{three: true, four: true, six: false, bear: false, land: false, spell: false, theirs: false} {
		if offered[id] != want {
			t.Errorf("%s offered = %v, want %v", id, offered[id], want)
		}
	}
	pickCard(t, g, me.ID, four)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(four) {
		t.Error("five mana spent reaches the {4} enchantment")
	}
}

// --- Atmospheric Greenhouse ------------------------------------------------------------

func TestEoeABAtmosphericGreenhouseGrowsYourCreaturesAndFliesAtEight(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	ship := avB1Cast(t, g, me, game.Card{Name: "Atmospheric Greenhouse", TypeLine: "Artifact — Spacecraft",
		OracleID: eoeABGreenhouseOracle, Power: 5, Toughness: 4}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if got := b12Counter(t, g, mine, game.CounterPlusOne); got != 1 {
		t.Errorf("your creature has %d counters, want 1", got)
	}
	if got := b12Counter(t, g, theirs, game.CounterPlusOne); got != 0 {
		t.Errorf("an opponent's creature has %d counters, want 0", got)
	}
	if got := b12Counter(t, g, ship, game.CounterPlusOne); got != 0 {
		t.Errorf("the Spacecraft is not a creature and has %d counters, want 0", got)
	}
	if layeredCard(t, g, ship).IsCreature() {
		t.Error("below 8 charge counters it is not a creature")
	}
	if err := g.AddCounter(ship, game.CounterCharge, 8); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	at8 := layeredCard(t, g, ship)
	if !at8.IsCreature() || !game.HasKeyword(&at8, "flying") || !game.HasKeyword(&at8, "trample") {
		t.Error("at 8+ it is a creature with flying and trample")
	}
	if at8.CurrentPower() != 5 || at8.CurrentToughness() != 4 {
		t.Errorf("at 8+ it is %d/%d, want 5/4", at8.CurrentPower(), at8.CurrentToughness())
	}
	if spec, _ := Lookup(eoeABGreenhouseOracle); len(spec.Activated) != 1 {
		t.Error("the station ability is registered")
	}
}

// --- Atomic Microsizer -----------------------------------------------------------------

// eoeABEquip seeds Atomic Microsizer attached to host.
func eoeABEquip(g *game.Game, owner, host uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Atomic Microsizer", TypeLine: "Artifact — Equipment",
		OracleID: eoeABMicrosizerOracle, Owner: owner, Controller: owner,
		AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: host},
	})
}

func TestEoeABAtomicMicrosizerMakesTheAttackerAnUnblockableOneOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 4, 4)
	eoeABEquip(g, me.ID, attacker)
	wall := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 6)
	if got := layeredCard(t, g, attacker).CurrentPower(); got != 5 {
		t.Fatalf("equipped creature gets +1/+0: power %d, want 5", got)
	}
	declareAttack(t, g, opp.ID, attacker)
	pickCard(t, g, me.ID, attacker)
	passPriorityAroundTable(t, g)
	c := layeredCard(t, g, attacker)
	if c.CurrentPower() != 2 || c.CurrentToughness() != 1 {
		t.Errorf("base 1/1 with the Equipment's +1/+0 is 2/1, got %d/%d", c.CurrentPower(), c.CurrentToughness())
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if reason := blockRefusal(t, g, wall, attacker); reason == "" {
		t.Error("the chosen creature can't be blocked this turn")
	}
}

func TestEoeABAtomicMicrosizerShrinksAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	eoeABEquip(g, me.ID, attacker)
	giant := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 6, 6)
	declareAttack(t, g, opp.ID, attacker)
	pickCard(t, g, me.ID, giant)
	passPriorityAroundTable(t, g)
	if c := layeredCard(t, g, giant); c.CurrentPower() != 1 || c.CurrentToughness() != 1 {
		t.Errorf("the giant has base power and toughness 1/1: %d/%d", c.CurrentPower(), c.CurrentToughness())
	}
	if c := layeredCard(t, g, attacker); c.CurrentPower() != 3 {
		t.Errorf("the attacker is untouched: power %d, want 3", c.CurrentPower())
	}
}

// --- Auxiliary Boosters ---------------------------------------------------------------

func TestEoeABAuxiliaryBoostersMakesARobotAndEquipsIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	boosters := avB1Cast(t, g, me, game.Card{Name: "Auxiliary Boosters", TypeLine: "Artifact — Equipment",
		OracleID: eoeABBoostersOracle}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	robot := avB1Named(g, me.ID, "Robot")
	if robot == uuid.Nil {
		t.Fatal("no Robot token")
	}
	if !avB1Card(t, g, boosters).IsAttachedTo(robot) {
		t.Error("the Boosters are attached to the Robot")
	}
	c := layeredCard(t, g, robot)
	if !c.IsArtifact() || c.CurrentPower() != 3 || c.CurrentToughness() != 4 || !game.HasKeyword(&c, "flying") {
		t.Errorf("an equipped 2/2 Robot artifact is a 3/4 flier, got %d/%d", c.CurrentPower(), c.CurrentToughness())
	}
}

// --- Beamsaw Prospector ---------------------------------------------------------------

func TestEoeABBeamsawProspectorLeavesALanderWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	prospector := b12Push(g, me.ID, "Beamsaw Prospector", "Creature — Human Artificer", eoeABProspectorOracle, 2, 1)
	b15Destroy(t, g, prospector)
	if got := eoeABNamedCount(g, me.ID, "Lander"); got != 1 {
		t.Errorf("Landers %d, want 1", got)
	}
}

// --- Beyond the Quiet -----------------------------------------------------------------

func TestEoeABBeyondTheQuietExilesCreaturesAndSpacecraft(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	ship := b12Permanent(g, opp.ID, "Their Ship", "Artifact — Spacecraft")
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	ench := b12Permanent(g, me.ID, "My Enchantment", "Enchantment")
	land := pushLandFor(g, me.ID, "Plains", "Basic Land — Plains")
	castCatalogSpell(t, g, "Beyond the Quiet", "Sorcery", eoeABQuietOracle, nil)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, theirs, ship} {
		if !g.Exile.Contains(id) {
			t.Errorf("%s should be exiled", id)
		}
	}
	for _, id := range []uuid.UUID{rock, ench, land} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s is neither a creature nor a Spacecraft and stays", id)
		}
	}
}

// --- Biomechan Engineer ---------------------------------------------------------------

func TestEoeABBiomechanEngineerMakesALanderThenDrawsAndBuildsForEight(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	engineer := avB1ResolveCreature(t, g, me, "Biomechan Engineer", "Creature — Insect Artificer", eoeABEngineerOracle)
	passPriorityAroundTable(t, g)
	if got := eoeABNamedCount(g, me.ID, "Lander"); got != 1 {
		t.Fatalf("Landers %d, want 1", got)
	}
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, engineer, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d → %d, want two cards drawn", hand, got)
	}
	if got := eoeABNamedCount(g, me.ID, "Robot"); got != 1 {
		t.Errorf("Robots %d, want 1", got)
	}
}

// --- Biosynthic Burst -----------------------------------------------------------------

func TestEoeABBiosynthicBurstGrowsGrantsAndUntaps(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	tapForTest(t, g, bear)
	if err := castCatalogSpellErr(t, g, "Biosynthic Burst", "Instant", eoeABBurstOracle, cardRefs(theirs)); err == nil {
		t.Fatal("only a creature you control is a legal target")
	}
	castCatalogSpell(t, g, "Biosynthic Burst", "Instant", eoeABBurstOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if got := b12Counter(t, g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters %d, want 1", got)
	}
	for _, kw := range []string{"reach", "trample", "indestructible"} {
		if !effectiveAbilitiesContain(t, g, bear, kw) {
			t.Errorf("it gains %s until end of turn", kw)
		}
	}
	if layeredCard(t, g, bear).Tapped {
		t.Error("it is untapped")
	}
}

// --- Biotech Specialist ---------------------------------------------------------------

func TestEoeABBiotechSpecialistShootsAnOpponentWhenYouSacrificeAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	avB1ResolveCreature(t, g, me, "Biotech Specialist", "Creature — Insect Scientist", eoeABSpecialistOracle)
	passPriorityAroundTable(t, g)
	lander := avB1Named(g, me.ID, "Lander")
	if lander == uuid.Nil {
		t.Fatal("no Lander token")
	}

	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if err := g.SacrificePermanent(me.ID, bear); err != nil {
		t.Fatalf("sacrifice the bear: %v", err)
	}
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("sacrificing a creature that is not an artifact does not trigger")
	}

	life := opp.Life
	if err := g.SacrificePermanent(me.ID, lander); err != nil {
		t.Fatalf("sacrifice the Lander: %v", err)
	}
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt after sacrificing an artifact")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); err == nil {
		t.Fatal("you are not an opponent")
	}
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("opponent life %d → %d, want 2 damage", life, opp.Life)
	}
}

// --- Blade of the Swarm ---------------------------------------------------------------

func TestEoeABBladeOfTheSwarmCanGrowItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	blade := avB1ResolveCreature(t, g, me, "Blade of the Swarm", "Creature — Insect Assassin", eoeABBladeOracle)
	passPriorityAroundTable(t, g)
	if pick := modePickChoiceFor(g, me.ID); pick != nil {
		if err := g.ResolveModePick(pick.ID, me.ID, []int{0}); err != nil {
			t.Fatalf("ResolveModePick: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if got := b12Counter(t, g, blade, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters %d, want 2", got)
	}
}

func TestEoeABBladeOfTheSwarmTucksAnExiledCardWithWarp(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	warped := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: warped, Name: "Weftstalker Ardent", TypeLine: "Creature — Drix Artificer",
		ManaCost: "{2}{R}", OracleID: weftstalkerWarpOracle, Owner: opp.ID, Controller: opp.ID})
	plain := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: plain, Name: "Exiled Bear", TypeLine: "Creature — Bear",
		ManaCost: "{1}{G}", Owner: opp.ID, Controller: opp.ID})

	avB1ResolveCreature(t, g, me, "Blade of the Swarm", "Creature — Insect Assassin", eoeABBladeOracle)
	passPriorityAroundTable(t, g)
	pick := modePickChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no mode prompt")
	}
	if err := g.ResolveModePick(pick.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, cardRefs(plain)); err == nil {
			t.Fatal("an exiled card without warp is not a legal target")
		}
		pickCard(t, g, me.ID, warped)
	}
	passPriorityAroundTable(t, g)
	if g.Exile.Contains(warped) {
		t.Fatal("the card with warp left exile")
	}
	if bottom, err := opp.Library.Bottom(); err != nil || bottom.InstanceID != warped {
		t.Errorf("the card with warp is on the bottom of its owner's library: %v %v", bottom.Name, err)
	}
	if !g.Exile.Contains(plain) {
		t.Error("the card without warp stays in exile")
	}
}

// --- Blooming Stinger -----------------------------------------------------------------

func TestEoeABBloomingStingerLendsDeathtouchToAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "My Elk", "Creature — Elk", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	stinger := avB1ResolveCreature(t, g, me, "Blooming Stinger", "Creature — Plant Scorpion", eoeABStingerOracle)
	if !effectiveAbilitiesContain(t, g, stinger, "deathtouch") {
		t.Error("printed deathtouch is missing")
	}
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the entry trigger")
	}
	for _, bad := range []uuid.UUID{stinger, theirs} {
		if err := g.ResolvePickTargets(p.ID, me.ID, cardRefs(bad)); err == nil {
			t.Fatalf("%s is not another creature you control", bad)
		}
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !effectiveAbilitiesContain(t, g, bear, "deathtouch") {
		t.Error("the bear gains deathtouch until end of turn")
	}
}

// --- Bombard --------------------------------------------------------------------------

func TestEoeABBombardDealsFourToACreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	ogre := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 4)
	castCatalogSpell(t, g, "Bombard", "Instant", eoeABBombardOracle, cardRefs(ogre))
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(ogre) || !opp.Graveyard.Contains(ogre) {
		t.Error("4 damage kills a creature with toughness 4")
	}
}

// --- Brightspear Zealot ---------------------------------------------------------------

func TestEoeABBrightspearZealotHitsHarderAfterTwoSpells(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zealot := b12Push(g, me.ID, "Brightspear Zealot", "Creature — Human Soldier", eoeABZealotOracle, 2, 4)
	ogres := []uuid.UUID{
		b12Creature(g, opp.ID, "First Ogre", "Creature — Ogre", 3, 8),
		b12Creature(g, opp.ID, "Second Ogre", "Creature — Ogre", 3, 8),
	}
	if !effectiveAbilitiesContain(t, g, zealot, "vigilance") {
		t.Error("printed vigilance is missing")
	}
	for i, want := range []int{2, 4} {
		castCatalogSpell(t, g, "Bombard", "Instant", eoeABBombardOracle, cardRefs(ogres[i]))
		if got := layeredCard(t, g, zealot).CurrentPower(); got != want {
			t.Errorf("after spell %d the Zealot's power is %d, want %d", i+1, got, want)
		}
		passPriorityAroundTable(t, g)
	}
	if got := layeredCard(t, g, zealot).CurrentToughness(); got != 4 {
		t.Errorf("toughness %d, want 4: the bonus is +2/+0", got)
	}
}

// --- Broodguard Elite -----------------------------------------------------------------

func TestEoeABBroodguardElitePassesItsCountersOnWhenItsWarpEnds(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	elite := avB1Cast(t, g, me, game.Card{Name: "Broodguard Elite", TypeLine: "Creature — Insect Knight", ManaCost: "{X}{G}{G}",
		OracleID: eoeABBroodguardOracle}, game.CastSpellParams{AlternativeCost: "warp", XValue: 3})
	passPriorityAroundTable(t, g)
	if got := b12Counter(t, g, elite, game.CounterPlusOne); got != 3 {
		t.Fatalf("warped with X=3, the Elite enters with %d counters, want 3", got)
	}
	advanceThroughEndStep(t, g)
	eoeABPickIfAsked(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(elite) {
		t.Fatal("the warped Elite is exiled at the end step")
	}
	if got := b12Counter(t, g, bear, game.CounterPlusOne); got != 3 {
		t.Errorf("the bear gets the Elite's counters: %d, want 3", got)
	}
}
