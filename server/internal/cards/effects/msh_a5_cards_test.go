package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msh_a5_cards_test.go — Marvel Super Heroes, half A, fifth batch (#4455).
// Each test plays the real catalog entry and asserts what its text says.

const (
	mshA5JessicaOracle   = "6677cd5a-0785-499b-ba99-f6bdcfb196d5"
	mshA5KidLokiOracle   = "19ef0329-08de-4835-8079-2e791e19f3a6"
	mshA5TChallaOracle   = "fe00278d-4385-4e3f-919c-12a794fd86f5"
	mshA5CommandosOracle = "09eb988a-e2eb-42e1-a786-01be63d2263f"
	mshA5LokiOracle      = "c112852f-61d5-4b94-b864-c7d0b5333b29"
	mshA5ModokOracle     = "fe7cc414-42d0-4dad-8516-7a0be0860a67"
	mshA5HydraOracle     = "d812c8e3-a475-4ad2-a93f-7208ad873676"
	mshA5FantasticOracle = "7fa5237e-dd79-4646-b935-cb8c6ee803ab"
	mshA5MistyOracle     = "5b090ba3-0ecb-4585-b5b7-9d6028c32487"
	mshA5MockingOracle   = "1fd00656-83ba-4feb-b089-224007936b99"
	mshA5MonicaOracle    = "8dd929cb-0fe1-4cb3-bc8d-96ca37f194f3"
)

func TestMshA5EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		mshA5JessicaOracle:        "Jessica Jones, Private Eye",
		mshA5KidLokiOracle:        "Kid Loki",
		mshA5TChallaOracle:        "King T'Challa",
		mshA5TChallaOracle + "#1": "Black Panther, Hope Enduring",
		mshA5CommandosOracle:      "Kree Commandos",
		mshA5LokiOracle:           "Loki, God of Mischief",
		mshA5ModokOracle:          "M.O.D.O.K.",
		mshA5HydraOracle:          "Madame Hydra",
		mshA5FantasticOracle:      "Mister Fantastic, Reed Richards",
		mshA5MistyOracle:          "Misty Knight, Hero for Hire",
		mshA5MockingOracle:        "Mockingbird, Ace Agent",
		mshA5MonicaOracle:         "Monica Rambeau",
		mshA5MonicaOracle + "#1":  "Photon, Living Light",
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

func mshA5Keywords(t *testing.T, oracle string, want ...string) {
	t.Helper()
	spec, ok := Lookup(oracle)
	if !ok {
		t.Fatalf("%s is not registered", oracle)
	}
	for _, kw := range want {
		if !slices.Contains(spec.PrintedKeywords, kw) {
			t.Errorf("%s lacks %s (has %v)", spec.Name, kw, spec.PrintedKeywords)
		}
	}
}

// mshA5Draw has p draw one card and settles any prompt or trigger it set off.
func mshA5Draw(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	g.WithWriteLock(func() { _ = g.DrawNForEffect(p.ID, 1) })
	assistCardSettle(t, g, p.ID, nil, true)
}

func mshA5Ref(id uuid.UUID) game.TargetRef {
	return game.TargetRef{Kind: game.TargetCard, ID: id}
}

// mshA5Tokens counts the tokens a player controls.
func mshA5Tokens(g *game.Game, owner uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Controller == owner {
			n++
		}
	}
	return n
}

// mshA5Spell casts a plain instant from the active seat at the targets and lets
// everything it set off resolve.
func mshA5Spell(t *testing.T, g *game.Game, targets ...game.TargetRef) {
	t.Helper()
	castCatalogSpell(t, g, "Zap", "Instant", "", targets)
	assistCardSettle(t, g, g.Seats[g.Turn.ActiveSeat].ID, nil, true)
}

// --- Kree Commandos ----------------------------------------------------------------

func TestMshA5KreeCommandosHasFlyingVigilanceAndProwess(t *testing.T) {
	mshA5Keywords(t, mshA5CommandosOracle, "flying", "vigilance", "prowess")
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b27Push(g, me.ID, "Kree Commandos", "Creature — Kree Soldier Villain", mshA5CommandosOracle, "{2}{W}", 2, 1, "W")
	castNoncreature(t, g)
	settleProwess(t, g)
	if p, tg := effectivePower(t, g, id), effectiveToughness(t, g, id); p != 3 || tg != 2 {
		t.Errorf("after a noncreature spell: %d/%d, want 3/2", p, tg)
	}
}

// --- Madame Hydra ------------------------------------------------------------------

func TestMshA5MadameHydraMakesAMenaceVillainForEachVillainSpellYouCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b27Push(g, me.ID, "Madame Hydra", "Legendary Creature — Human Villain", mshA5HydraOracle, "{2}{B}{R}", 2, 3, "B", "R")
	castFromHand(t, g, game.Card{Name: "Henchman", TypeLine: "Creature — Human Villain", Power: 1, Toughness: 1}, game.CastSpellParams{})
	assistCardSettle(t, g, me.ID, nil, true)
	if got := mshA5Tokens(g, me.ID); got != 1 {
		t.Fatalf("%d tokens after a Villain spell, want 1", got)
	}
	for _, c := range g.Battlefield.Cards {
		if !c.IsToken() || c.Controller != me.ID {
			continue
		}
		if c.Power != 2 || c.Toughness != 1 || !c.HasSubtype("Villain") || !hasKeywordOnBattlefield(t, g, c.InstanceID, "menace") {
			t.Errorf("the token is %d/%d %q with menace %v, want a 2/1 Villain with menace", c.Power, c.Toughness, c.TypeLine, hasKeywordOnBattlefield(t, g, c.InstanceID, "menace"))
		}
	}
}

func TestMshA5MadameHydraIgnoresNonVillainsAndOpponentsSpells(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Madame Hydra", "Legendary Creature — Human Villain", mshA5HydraOracle, "{2}{B}{R}", 2, 3, "B", "R")
	castFromHand(t, g, game.Card{Name: "Bystander", TypeLine: "Creature — Human Hero", Power: 1, Toughness: 1}, game.CastSpellParams{})
	assistCardSettle(t, g, me.ID, nil, true)
	castFlashCreatureFrom(t, g, opp, game.Card{Name: "Their Henchman", TypeLine: "Creature — Human Villain", Power: 1, Toughness: 1})
	assistCardSettle(t, g, me.ID, nil, true)
	if got := mshA5Tokens(g, me.ID) + mshA5Tokens(g, opp.ID); got != 0 {
		t.Errorf("%d tokens after a Hero spell and an opponent's Villain spell, want 0", got)
	}
}

// --- Mister Fantastic --------------------------------------------------------------

func TestMshA5MisterFantasticDrawsOnceForABatchOfTokensAndMayDecline(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b27Push(g, me.ID, "Mister Fantastic, Reed Richards", "Legendary Creature — Human Scientist Hero", mshA5FantasticOracle, "{3}{U}", 2, 4, "U")
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("2/2 green Bear"), 2) })
	assistCardSettle(t, g, me.ID, nil, true)
	if got := me.Hand.Size(); got != hand+1 {
		t.Fatalf("hand %d -> %d: two tokens entering together draw one card", hand, got)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("2/2 green Bear"), 1) })
	assistCardSettle(t, g, me.ID, nil, false)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d: declining the \"may\" still drew", got)
	}
}

func TestMshA5MisterFantasticIgnoresNontokensAndOpponentsTokens(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Mister Fantastic, Reed Richards", "Legendary Creature — Human Scientist Hero", mshA5FantasticOracle, "{3}{U}", 2, 4, "U")
	hand := me.Hand.Size()
	castFromHand(t, g, game.Card{Name: "Bystander", TypeLine: "Creature — Human Hero", Power: 1, Toughness: 1}, game.CastSpellParams{})
	assistCardSettle(t, g, me.ID, nil, true)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, TokenCard("2/2 green Bear"), 1) })
	assistCardSettle(t, g, me.ID, nil, true)
	// The creature cast left the hand, so a draw would leave it at hand.
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d -> %d after a nontoken creature and an opponent's token, want unchanged", hand, got)
	}
}

// --- Mockingbird, Ace Agent --------------------------------------------------------

func TestMshA5MockingbirdGrowsOnlyForSpellsAimedAtYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mockingbird := b27Push(g, me.ID, "Mockingbird, Ace Agent", "Legendary Creature — Human Spy Hero", mshA5MockingOracle, "{3}{W}", 2, 2, "W")
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mshA5Keywords(t, mshA5MockingOracle, "double strike")

	mshA5Spell(t, g, mshA5Ref(theirs))
	mshA5Spell(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
	if got := plusOnes(g, mockingbird); got != 0 {
		t.Fatalf("%d counters after spells aimed at an opponent's creature and a player, want 0", got)
	}
	mshA5Spell(t, g, mshA5Ref(mine))
	if got := plusOnes(g, mockingbird); got != 1 {
		t.Errorf("%d counters after a spell aimed at my creature, want 1", got)
	}
	mshA5Spell(t, g, mshA5Ref(mockingbird))
	if got := plusOnes(g, mockingbird); got != 2 {
		t.Errorf("%d counters after a spell aimed at Mockingbird herself, want 2", got)
	}
}

// --- Kid Loki ----------------------------------------------------------------------

func TestMshA5KidLokiGrowsOnTheSecondCardYouDrawEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, other := g.Seats[1], g.Seats[2]
	loki := b27Push(g, me.ID, "Kid Loki", "Legendary Creature — God Hero Villain", mshA5KidLokiOracle, "{U}", 1, 1, "U")
	mshA5Draw(t, g, other)
	mshA5Draw(t, g, other)
	mshA5Draw(t, g, me)
	if got := plusOnes(g, loki); got != 0 {
		t.Fatalf("%d counters after one draw of mine and two of another player's, want 0", got)
	}
	mshA5Draw(t, g, me)
	if got := plusOnes(g, loki); got != 1 {
		t.Fatalf("%d counters after my second draw, want 1", got)
	}
	mshA5Draw(t, g, me)
	if got := plusOnes(g, loki); got != 1 {
		t.Errorf("%d counters after my third draw, want it to stay 1", got)
	}
}

func TestMshA5KidLokiGivesHexproofToCreaturesYouPutCountersOnThisTurnOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Kid Loki", "Legendary Creature — God Hero Villain", mshA5KidLokiOracle, "{U}", 1, 1, "U")
	boosted := b12Creature(g, me.ID, "Boosted", "Creature — Bear", 2, 2)
	plain := b12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	byOpp := b12Creature(g, me.ID, "Opp Boosted", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "Charged", "Creature — Bear", 2, 2)
	if effectiveAbilitiesContain(t, g, boosted, "hexproof") {
		t.Fatal("a creature with no counter has hexproof")
	}
	g.WithWriteLock(func() {
		_ = g.AddCounterByForEffect(me.ID, boosted, game.CounterPlusOne, 1)
		_ = g.AddCounterByForEffect(opp.ID, byOpp, game.CounterPlusOne, 1)
		_ = g.AddCounterByForEffect(me.ID, theirs, game.CounterPlusOne, 1)
		_ = g.AddCounterByForEffect(me.ID, other, "charge", 1)
	})
	if !effectiveAbilitiesContain(t, g, boosted, "hexproof") {
		t.Error("a creature I put a +1/+1 counter on has no hexproof")
	}
	for name, id := range map[string]uuid.UUID{"untouched": plain, "countered by an opponent": byOpp, "an opponent's": theirs, "another counter kind": other} {
		if effectiveAbilitiesContain(t, g, id, "hexproof") {
			t.Errorf("the %s creature has hexproof", name)
		}
	}
	advanceToUpkeepOfSeat(t, g, 1)
	if effectiveAbilitiesContain(t, g, boosted, "hexproof") {
		t.Error("the hexproof outlasted the turn")
	}
}

// --- King T'Challa // Black Panther, Hope Enduring -------------------------------------

func mshA5TChallaCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   mshA5TChallaOracle,
		Layout:     game.LayoutModalDFC,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{Name: "King T'Challa", TypeLine: "Legendary Creature — Human Noble Hero", ManaCost: "{1}{W}{U}", Colors: []string{"W", "U"}, Power: 3, Toughness: 2},
			{Name: "Black Panther, Hope Enduring", TypeLine: "Legendary Creature — Human Warrior Hero", ManaCost: "{4}{W}{U}", Colors: []string{"W", "U"}, Power: 3, Toughness: 3},
		},
	}
	c.SetFace(0)
	return c
}

func mshA5BlackPanther(g *game.Game, owner uuid.UUID) uuid.UUID {
	c := mshA5TChallaCard(owner)
	c.SetFace(1)
	return pushBattlefieldCardWithTimestamp(g, c)
}

func TestMshA5KingTChallaDrawsWhenAnyPlayerDrawsTheirSecondCard(t *testing.T) {
	g := newCatalogGame(t)
	me, other := g.Seats[1], g.Seats[2]
	pushBattlefieldCardWithTimestamp(g, mshA5TChallaCard(me.ID))
	mshA5Keywords(t, mshA5TChallaOracle, "flash")
	hand := me.Hand.Size()
	mshA5Draw(t, g, other)
	if me.Hand.Size() != hand {
		t.Fatalf("hand %d -> %d after another player's first draw, want unchanged", hand, me.Hand.Size())
	}
	mshA5Draw(t, g, other)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d -> %d after another player's second draw, want one more", hand, me.Hand.Size())
	}
	mshA5Draw(t, g, other)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d after another player's third draw, want it unchanged", me.Hand.Size())
	}
	// The card T'Challa drew me was my first of the turn, so my next draw is my
	// second: it pays out one more card, and that one is my third.
	mshA5Draw(t, g, me)
	if me.Hand.Size() != hand+3 {
		t.Fatalf("hand %d after my second draw, want %d (the draw and T'Challa's)", me.Hand.Size(), hand+3)
	}
	mshA5Draw(t, g, me)
	if me.Hand.Size() != hand+4 {
		t.Errorf("hand %d after my fourth draw, want %d (no further trigger)", me.Hand.Size(), hand+4)
	}
}

func TestMshA5KingTChallaTransformsAtSorcerySpeedForSix(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	king := pushBattlefieldCardWithTimestamp(g, mshA5TChallaCard(me.ID))
	advanceTo(t, g, game.StepBeginCombat)
	floatMana(t, g, me, "{C}{C}{C}{C}{W}{U}")
	if err := g.ActivateCatalogAbility(me.ID, king, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("the transform was activated outside a main phase")
	}
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, king, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, king)
	if !ok || c.Name != "Black Panther, Hope Enduring" {
		t.Fatalf("after the transform: %+v", c)
	}
	if !mshA5HasKeyword(g, king, "double strike") || !mshA5HasKeyword(g, king, "flash") {
		t.Error("Black Panther lacks double strike or flash")
	}
}

// mshA5HasKeyword reads a keyword off the battlefield card as the layers leave it.
func mshA5HasKeyword(g *game.Game, id uuid.UUID, kw string) bool {
	var has bool
	g.ReadSnapshot(func() {
		if c, ok := g.LookupCardForEffect(id); ok {
			has = game.HasKeyword(&c, kw)
		}
	})
	return has
}

func TestMshA5BlackPantherPreventsAllDamageDealtToHim(t *testing.T) {
	g := newCatalogGame(t)
	panther := mshA5BlackPanther(g, g.Seats[0].ID)
	src := pushVanillaCreature(g, g.Seats[1].ID, "Pinger", 5, 5)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(src, panther, 5) })
	g.RunStateChecksForTest()
	c, ok := battlefieldCard(g, panther)
	if !ok {
		t.Fatal("Black Panther died to damage that should be prevented")
	}
	if c.DamageMarked != 0 {
		t.Errorf("Black Panther has %d damage marked, want 0", c.DamageMarked)
	}
}

func TestMshA5BlackPantherDrawsForEachCombatDamageStepToAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	panther := mshA5BlackPanther(g, me.ID)
	hand := me.Hand.Size()
	declareAttack(t, g, opp.ID, panther)
	advanceTo(t, g, game.StepEndCombat)
	assistCardSettle(t, g, me.ID, nil, true)
	// Double strike: the first-strike and the regular damage steps each deal
	// combat damage to the player, so he draws twice.
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d cards off two combat damage steps, want 2", got)
	}
}

// --- Monica Rambeau // Photon, Living Light ------------------------------------------

func mshA5MonicaCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   mshA5MonicaOracle,
		Layout:     game.LayoutModalDFC,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{Name: "Monica Rambeau", TypeLine: "Legendary Creature — Human Hero", ManaCost: "{R}{W}", Colors: []string{"R", "W"}, Power: 1, Toughness: 3},
			{Name: "Photon, Living Light", TypeLine: "Legendary Creature — Elemental Hero", ManaCost: "{2}{R}{W}{W}", Colors: []string{"R", "W"}, Power: 3, Toughness: 5},
		},
	}
	c.SetFace(0)
	return c
}

func TestMshA5MonicaTransformsIntoPhotonAtSorcerySpeed(t *testing.T) {
	mshA5Keywords(t, mshA5MonicaOracle, "flying", "prowess")
	mshA5Keywords(t, mshA5MonicaOracle+"#1", "flying", "hexproof", "prowess")
	g := newCatalogGame(t)
	me := g.Seats[0]
	monica := pushBattlefieldCardWithTimestamp(g, mshA5MonicaCard(me.ID))
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}{R}{W}{W}")
	b16Activate(t, g, me.ID, monica, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, monica)
	if !ok || c.Name != "Photon, Living Light" {
		t.Fatalf("after the transform: %+v", c)
	}
	if c.CurrentPower() != 3 || c.CurrentToughness() != 5 {
		t.Errorf("Photon is %d/%d, want 3/5", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestMshA5PhotonPutsACounterOnEachOtherCreatureWhenYouCastANoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	c := mshA5MonicaCard(me.ID)
	c.SetFace(1)
	photon := pushBattlefieldCardWithTimestamp(g, c)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castFromHand(t, g, game.Card{Name: "Bystander", TypeLine: "Creature — Human Hero", Power: 1, Toughness: 1}, game.CastSpellParams{})
	settleProwess(t, g)
	if plusOnes(g, a) != 0 {
		t.Fatal("a creature spell put counters on the team")
	}
	castNoncreature(t, g)
	settleProwess(t, g)
	for name, id := range map[string]uuid.UUID{"Bear A": a, "Bear B": b} {
		if got := plusOnes(g, id); got != 1 {
			t.Errorf("%s has %d counters, want 1", name, got)
		}
	}
	if got := plusOnes(g, photon); got != 0 {
		t.Errorf("Photon put %d counters on herself, want 0", got)
	}
	if got := plusOnes(g, theirs); got != 0 {
		t.Errorf("an opponent's creature got %d counters, want 0", got)
	}
}

// --- M.O.D.O.K. --------------------------------------------------------------------

func mshA5Modok(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b27Push(g, owner, "M.O.D.O.K.", "Legendary Artifact Creature — Villain", mshA5ModokOracle, "{3}{B}{B}", 2, 2, "B")
}

func TestMshA5ModokShrinksOnlyYourOpponentsCreatures(t *testing.T) {
	mshA5Keywords(t, mshA5ModokOracle, "flying", "lifelink")
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	modok := mshA5Modok(g, me.ID)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	big := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	small := b12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	g.RunStateChecksForTest()
	if p, tg := effectivePower(t, g, big), effectiveToughness(t, g, big); p != 2 || tg != 2 {
		t.Errorf("their 3/3 is %d/%d, want 2/2", p, tg)
	}
	if p := effectivePower(t, g, mine); p != 2 {
		t.Errorf("my 2/2 is power %d, want 2", p)
	}
	if p := effectivePower(t, g, modok); p != 2 {
		t.Errorf("M.O.D.O.K. shrank himself to power %d", p)
	}
	if g.Battlefield.Contains(small) {
		t.Error("their 1/1 survived a -1/-1")
	}
}

func TestMshA5ModokPaysThreeLifeToConniveOnlyOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	modok := mshA5Modok(g, me.ID)
	spell := handCardForTest(me, "Spare Spell", "Sorcery", "")
	advanceToMain(t, g)
	life, hand := me.Life, me.Hand.Size()
	b16Activate(t, g, me.ID, modok, 0, game.ActivateAbilityParams{})
	if me.Life != life-3 {
		t.Errorf("life %d -> %d, want 3 paid", life, me.Life)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Fatalf("hand %d -> %d, want the connive draw", hand, got)
	}
	answerDiscard(t, g, me.ID, spell)
	if got := plusOnes(g, modok); got != 1 {
		t.Errorf("%d counters after discarding a nonland card, want 1", got)
	}
	// Off your turn it cannot be activated, and costs nothing.
	advanceToUpkeepOfSeat(t, g, 1)
	life = me.Life
	if err := g.ActivateCatalogAbility(me.ID, modok, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("M.O.D.O.K. connived during an opponent's turn")
	}
	if me.Life != life {
		t.Errorf("life changed to %d by a refused activation", me.Life)
	}
}

// --- Misty Knight, Hero for Hire ----------------------------------------------------

func TestMshA5MistyKnightDrawsForEachCardDiscardedThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	misty := b27Push(g, me.ID, "Misty Knight, Hero for Hire", "Legendary Creature — Human Detective Hero", mshA5MistyOracle, "{1}{R}", 3, 1, "R")
	advanceToMain(t, g)
	// One card discarded earlier in the turn, the only one in hand at the time.
	me.Hand.Cards = nil
	handCardForTest(me, "Earlier Discard", "Sorcery", "")
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	paid := handCardForTest(me, "Paid Card", "Sorcery", "")
	floatMana(t, g, me, "{C}{C}")
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, misty, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{paid}})
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, misty) {
		t.Error("the ability did not tap Misty Knight")
	}
	// Two cards discarded this turn (one earlier, one as the cost): draw two.
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("hand changed by %d: paid one card and drew two, want a net of 1", got)
	}
}

// --- Jessica Jones, Private Eye ------------------------------------------------------

func mshA5Jessica(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b27Push(g, owner, "Jessica Jones, Private Eye", "Legendary Creature — Human Detective Hero", mshA5JessicaOracle, "{2}{R}", 2, 3, "R")
}

func TestMshA5JessicaJonesStunsHerselfAndExilesTopXWhereXIsHerPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	jessica := mshA5Jessica(g, me.ID)
	ids := seedLibrary(me, "One", "Two", "Three", "Four")
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, jessica, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, jessica) {
		t.Error("the ability did not tap Jessica")
	}
	if got := countersOn(g, jessica, game.CounterStun); got != 1 {
		t.Errorf("%d stun counters, want 1", got)
	}
	for i, id := range ids[:2] {
		if !b12ExiledPlayableBy(t, g, id, me.ID) {
			t.Errorf("card %d of the top two is not exiled and playable", i+1)
		}
	}
	if b12ZoneOf(g, ids[2]) != game.ZoneLibrary {
		t.Error("a third card was exiled for a power-2 Jessica")
	}
}

func TestMshA5JessicaJonesPowerCountsCountersAndShowsUpAsX(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	jessica := mshA5Jessica(g, me.ID)
	ids := seedLibrary(me, "One", "Two", "Three", "Four")
	mshA5PutPlusOne(g, jessica)
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, jessica, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	for i, id := range ids[:3] {
		if b12ZoneOf(g, id) != game.ZoneExile {
			t.Errorf("card %d of the top three is not exiled for a power-3 Jessica", i+1)
		}
	}
	if b12ZoneOf(g, ids[3]) != game.ZoneLibrary {
		t.Error("a fourth card was exiled")
	}
	// Tapped, she cannot do it again.
	if err := g.ActivateCatalogAbility(me.ID, jessica, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("a tapped Jessica activated her ability")
	}
}

func mshA5PutPlusOne(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, game.CounterPlusOne, 1) })
}

// --- Loki, God of Mischief -----------------------------------------------------------

func TestMshA5LokiDrawsOnceEachTurnWhenYourAbilityTargets(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Loki, God of Mischief", "Legendary Creature — God Sorcerer Villain", mshA5LokiOracle, "{1}{U}", 2, 1, "U")
	bomb := pushCatalogPermanent(g, me.ID, "Goblin Bombardment", "Enchantment", goblinBombardmentOracle, false)
	f1 := b12Creature(g, me.ID, "Fodder 1", "Creature — Goblin", 1, 1)
	f2 := b12Creature(g, me.ID, "Fodder 2", "Creature — Goblin", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 4, 4)
	advanceToMain(t, g)

	// A spell that targets is not an ability.
	hand := me.Hand.Size()
	mshA5Spell(t, g, mshA5Ref(theirs))
	assistCardSettle(t, g, me.ID, nil, true)
	if got := me.Hand.Size(); got != hand {
		t.Fatalf("hand %d -> %d after a targeting spell, want unchanged", hand, got)
	}
	if err := g.ActivateCatalogAbility(me.ID, bomb, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{f1}, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	assistCardSettle(t, g, me.ID, nil, true)
	if got := me.Hand.Size(); got != hand+1 {
		t.Fatalf("hand %d -> %d after my ability targeted a player, want one more", hand, got)
	}
	if err := g.ActivateCatalogAbility(me.ID, bomb, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{f2}, Targets: []game.TargetRef{mshA5Ref(theirs)},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	assistCardSettle(t, g, me.ID, nil, true)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d: a second targeting ability the same turn drew again", got)
	}
}

func TestMshA5LokiIgnoresAnOpponentsAbility(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Loki, God of Mischief", "Legendary Creature — God Sorcerer Villain", mshA5LokiOracle, "{1}{U}", 2, 1, "U")
	theirBomb := pushCatalogPermanent(g, opp.ID, "Goblin Bombardment", "Enchantment", goblinBombardmentOracle, false)
	fodder := b12Creature(g, opp.ID, "Their Fodder", "Creature — Goblin", 1, 1)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 4, 4)
	advanceToMain(t, g)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(opp.ID, theirBomb, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder}, Targets: []game.TargetRef{mshA5Ref(mine)},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	assistCardSettle(t, g, me.ID, nil, true)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d -> %d when an opponent's ability targeted my creature, want unchanged", hand, got)
	}
}
