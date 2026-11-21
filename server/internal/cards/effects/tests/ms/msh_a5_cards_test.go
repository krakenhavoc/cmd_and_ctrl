package testsms

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestMshA5EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		effectstest.MshA5JessicaOracle:        "Jessica Jones, Private Eye",
		effectstest.MshA5KidLokiOracle:        "Kid Loki",
		effectstest.MshA5TChallaOracle:        "King T'Challa",
		effectstest.MshA5TChallaOracle + "#1": "Black Panther, Hope Enduring",
		effectstest.MshA5CommandosOracle:      "Kree Commandos",
		effectstest.MshA5LokiOracle:           "Loki, God of Mischief",
		effectstest.MshA5ModokOracle:          "M.O.D.O.K.",
		effectstest.MshA5HydraOracle:          "Madame Hydra",
		effectstest.MshA5FantasticOracle:      "Mister Fantastic, Reed Richards",
		effectstest.MshA5MistyOracle:          "Misty Knight, Hero for Hire",
		effectstest.MshA5MockingOracle:        "Mockingbird, Ace Agent",
		effectstest.MshA5MonicaOracle:         "Monica Rambeau",
		effectstest.MshA5MonicaOracle + "#1":  "Photon, Living Light",
	}
	for oracle, name := range want {
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

func mshA5Keywords(t *testing.T, oracle string, want ...string) {
	t.Helper()
	spec, ok := effects.Lookup(oracle)
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
	effectstest.AssistCardSettle(t, g, p.ID, nil, true)
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
	effectstest.CastCatalogSpell(t, g, "Zap", "Instant", "", targets)
	effectstest.AssistCardSettle(t, g, g.Seats[g.Turn.ActiveSeat].ID, nil, true)
}

// --- Kree Commandos ----------------------------------------------------------------

func TestMshA5KreeCommandosHasFlyingVigilanceAndProwess(t *testing.T) {
	mshA5Keywords(t, effectstest.MshA5CommandosOracle, "flying", "vigilance", "prowess")
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B27Push(g, me.ID, "Kree Commandos", "Creature — Kree Soldier Villain", effectstest.MshA5CommandosOracle, "{2}{W}", 2, 1, "W")
	effectstest.CastNoncreature(t, g)
	effectstest.SettleProwess(t, g)
	if p, tg := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id); p != 3 || tg != 2 {
		t.Errorf("after a noncreature spell: %d/%d, want 3/2", p, tg)
	}
}

// --- Madame Hydra ------------------------------------------------------------------

func TestMshA5MadameHydraMakesAMenaceVillainForEachVillainSpellYouCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B27Push(g, me.ID, "Madame Hydra", "Legendary Creature — Human Villain", effectstest.MshA5HydraOracle, "{2}{B}{R}", 2, 3, "B", "R")
	effectstest.CastFromHand(t, g, game.Card{Name: "Henchman", TypeLine: "Creature — Human Villain", Power: 1, Toughness: 1}, game.CastSpellParams{})
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if got := mshA5Tokens(g, me.ID); got != 1 {
		t.Fatalf("%d tokens after a Villain spell, want 1", got)
	}
	for _, c := range g.Battlefield.Cards {
		if !c.IsToken() || c.Controller != me.ID {
			continue
		}
		if c.Power != 2 || c.Toughness != 1 || !c.HasSubtype("Villain") || !effectstest.HasKeywordOnBattlefield(t, g, c.InstanceID, "menace") {
			t.Errorf("the token is %d/%d %q with menace %v, want a 2/1 Villain with menace", c.Power, c.Toughness, c.TypeLine, effectstest.HasKeywordOnBattlefield(t, g, c.InstanceID, "menace"))
		}
	}
}

func TestMshA5MadameHydraIgnoresNonVillainsAndOpponentsSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B27Push(g, me.ID, "Madame Hydra", "Legendary Creature — Human Villain", effectstest.MshA5HydraOracle, "{2}{B}{R}", 2, 3, "B", "R")
	effectstest.CastFromHand(t, g, game.Card{Name: "Bystander", TypeLine: "Creature — Human Hero", Power: 1, Toughness: 1}, game.CastSpellParams{})
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	effectstest.CastFlashCreatureFrom(t, g, opp, game.Card{Name: "Their Henchman", TypeLine: "Creature — Human Villain", Power: 1, Toughness: 1})
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if got := mshA5Tokens(g, me.ID) + mshA5Tokens(g, opp.ID); got != 0 {
		t.Errorf("%d tokens after a Hero spell and an opponent's Villain spell, want 0", got)
	}
}

// --- Mister Fantastic --------------------------------------------------------------

func TestMshA5MisterFantasticDrawsOnceForABatchOfTokensAndMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B27Push(g, me.ID, "Mister Fantastic, Reed Richards", "Legendary Creature — Human Scientist Hero", effectstest.MshA5FantasticOracle, "{3}{U}", 2, 4, "U")
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("2/2 green Bear"), 2) })
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if got := me.Hand.Size(); got != hand+1 {
		t.Fatalf("hand %d -> %d: two tokens entering together draw one card", hand, got)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("2/2 green Bear"), 1) })
	effectstest.AssistCardSettle(t, g, me.ID, nil, false)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d: declining the \"may\" still drew", got)
	}
}

func TestMshA5MisterFantasticIgnoresNontokensAndOpponentsTokens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B27Push(g, me.ID, "Mister Fantastic, Reed Richards", "Legendary Creature — Human Scientist Hero", effectstest.MshA5FantasticOracle, "{3}{U}", 2, 4, "U")
	hand := me.Hand.Size()
	effectstest.CastFromHand(t, g, game.Card{Name: "Bystander", TypeLine: "Creature — Human Hero", Power: 1, Toughness: 1}, game.CastSpellParams{})
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.TokenCard("2/2 green Bear"), 1) })
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	// The creature cast left the hand, so a draw would leave it at hand.
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d -> %d after a nontoken creature and an opponent's token, want unchanged", hand, got)
	}
}

// --- Mockingbird, Ace Agent --------------------------------------------------------

func TestMshA5MockingbirdGrowsOnlyForSpellsAimedAtYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mockingbird := effectstest.B27Push(g, me.ID, "Mockingbird, Ace Agent", "Legendary Creature — Human Spy Hero", effectstest.MshA5MockingOracle, "{3}{W}", 2, 2, "W")
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mshA5Keywords(t, effectstest.MshA5MockingOracle, "double strike")

	mshA5Spell(t, g, mshA5Ref(theirs))
	mshA5Spell(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
	if got := effectstest.PlusOnes(g, mockingbird); got != 0 {
		t.Fatalf("%d counters after spells aimed at an opponent's creature and a player, want 0", got)
	}
	mshA5Spell(t, g, mshA5Ref(mine))
	if got := effectstest.PlusOnes(g, mockingbird); got != 1 {
		t.Errorf("%d counters after a spell aimed at my creature, want 1", got)
	}
	mshA5Spell(t, g, mshA5Ref(mockingbird))
	if got := effectstest.PlusOnes(g, mockingbird); got != 2 {
		t.Errorf("%d counters after a spell aimed at Mockingbird herself, want 2", got)
	}
}

// --- Kid Loki ----------------------------------------------------------------------

func TestMshA5KidLokiGrowsOnTheSecondCardYouDrawEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, other := g.Seats[1], g.Seats[2]
	loki := effectstest.B27Push(g, me.ID, "Kid Loki", "Legendary Creature — God Hero Villain", effectstest.MshA5KidLokiOracle, "{U}", 1, 1, "U")
	mshA5Draw(t, g, other)
	mshA5Draw(t, g, other)
	mshA5Draw(t, g, me)
	if got := effectstest.PlusOnes(g, loki); got != 0 {
		t.Fatalf("%d counters after one draw of mine and two of another player's, want 0", got)
	}
	mshA5Draw(t, g, me)
	if got := effectstest.PlusOnes(g, loki); got != 1 {
		t.Fatalf("%d counters after my second draw, want 1", got)
	}
	mshA5Draw(t, g, me)
	if got := effectstest.PlusOnes(g, loki); got != 1 {
		t.Errorf("%d counters after my third draw, want it to stay 1", got)
	}
}

func TestMshA5KidLokiGivesHexproofToCreaturesYouPutCountersOnThisTurnOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B27Push(g, me.ID, "Kid Loki", "Legendary Creature — God Hero Villain", effectstest.MshA5KidLokiOracle, "{U}", 1, 1, "U")
	boosted := effectstest.B12Creature(g, me.ID, "Boosted", "Creature — Bear", 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	byOpp := effectstest.B12Creature(g, me.ID, "Opp Boosted", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Charged", "Creature — Bear", 2, 2)
	if effectstest.EffectiveAbilitiesContain(t, g, boosted, "hexproof") {
		t.Fatal("a creature with no counter has hexproof")
	}
	g.WithWriteLock(func() {
		_ = g.AddCounterByForEffect(me.ID, boosted, game.CounterPlusOne, 1)
		_ = g.AddCounterByForEffect(opp.ID, byOpp, game.CounterPlusOne, 1)
		_ = g.AddCounterByForEffect(me.ID, theirs, game.CounterPlusOne, 1)
		_ = g.AddCounterByForEffect(me.ID, other, "charge", 1)
	})
	if !effectstest.EffectiveAbilitiesContain(t, g, boosted, "hexproof") {
		t.Error("a creature I put a +1/+1 counter on has no hexproof")
	}
	for name, id := range map[string]uuid.UUID{"untouched": plain, "countered by an opponent": byOpp, "an opponent's": theirs, "another counter kind": other} {
		if effectstest.EffectiveAbilitiesContain(t, g, id, "hexproof") {
			t.Errorf("the %s creature has hexproof", name)
		}
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	if effectstest.EffectiveAbilitiesContain(t, g, boosted, "hexproof") {
		t.Error("the hexproof outlasted the turn")
	}
}

// --- King T'Challa // Black Panther, Hope Enduring -------------------------------------

func mshA5TChallaCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   effectstest.MshA5TChallaOracle,
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
	return effectstest.PushBattlefieldCardWithTimestamp(g, c)
}

func TestMshA5KingTChallaDrawsWhenAnyPlayerDrawsTheirSecondCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, other := g.Seats[1], g.Seats[2]
	effectstest.PushBattlefieldCardWithTimestamp(g, mshA5TChallaCard(me.ID))
	mshA5Keywords(t, effectstest.MshA5TChallaOracle, "flash")
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	king := effectstest.PushBattlefieldCardWithTimestamp(g, mshA5TChallaCard(me.ID))
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{W}{U}")
	if err := g.ActivateCatalogAbility(me.ID, king, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("the transform was activated outside a main phase")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, king, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, king)
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
	g := effectstest.NewCatalogGame(t)
	panther := mshA5BlackPanther(g, g.Seats[0].ID)
	src := effectstest.PushVanillaCreature(g, g.Seats[1].ID, "Pinger", 5, 5)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(src, panther, 5) })
	g.RunStateChecksForTest()
	c, ok := effectstest.BattlefieldCard(g, panther)
	if !ok {
		t.Fatal("Black Panther died to damage that should be prevented")
	}
	if c.DamageMarked != 0 {
		t.Errorf("Black Panther has %d damage marked, want 0", c.DamageMarked)
	}
}

func TestMshA5BlackPantherDrawsForEachCombatDamageStepToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	panther := mshA5BlackPanther(g, me.ID)
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, panther)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
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
		OracleID:   effectstest.MshA5MonicaOracle,
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
	mshA5Keywords(t, effectstest.MshA5MonicaOracle, "flying", "prowess")
	mshA5Keywords(t, effectstest.MshA5MonicaOracle+"#1", "flying", "hexproof", "prowess")
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	monica := effectstest.PushBattlefieldCardWithTimestamp(g, mshA5MonicaCard(me.ID))
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{R}{W}{W}")
	effectstest.B16Activate(t, g, me.ID, monica, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, monica)
	if !ok || c.Name != "Photon, Living Light" {
		t.Fatalf("after the transform: %+v", c)
	}
	if c.CurrentPower() != 3 || c.CurrentToughness() != 5 {
		t.Errorf("Photon is %d/%d, want 3/5", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestMshA5PhotonPutsACounterOnEachOtherCreatureWhenYouCastANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	c := mshA5MonicaCard(me.ID)
	c.SetFace(1)
	photon := effectstest.PushBattlefieldCardWithTimestamp(g, c)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastFromHand(t, g, game.Card{Name: "Bystander", TypeLine: "Creature — Human Hero", Power: 1, Toughness: 1}, game.CastSpellParams{})
	effectstest.SettleProwess(t, g)
	if effectstest.PlusOnes(g, a) != 0 {
		t.Fatal("a creature spell put counters on the team")
	}
	effectstest.CastNoncreature(t, g)
	effectstest.SettleProwess(t, g)
	for name, id := range map[string]uuid.UUID{"Bear A": a, "Bear B": b} {
		if got := effectstest.PlusOnes(g, id); got != 1 {
			t.Errorf("%s has %d counters, want 1", name, got)
		}
	}
	if got := effectstest.PlusOnes(g, photon); got != 0 {
		t.Errorf("Photon put %d counters on herself, want 0", got)
	}
	if got := effectstest.PlusOnes(g, theirs); got != 0 {
		t.Errorf("an opponent's creature got %d counters, want 0", got)
	}
}

// --- M.O.D.O.K. --------------------------------------------------------------------

func mshA5Modok(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.B27Push(g, owner, "M.O.D.O.K.", "Legendary Artifact Creature — Villain", effectstest.MshA5ModokOracle, "{3}{B}{B}", 2, 2, "B")
}

func TestMshA5ModokShrinksOnlyYourOpponentsCreatures(t *testing.T) {
	mshA5Keywords(t, effectstest.MshA5ModokOracle, "flying", "lifelink")
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	modok := mshA5Modok(g, me.ID)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	small := effectstest.B12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	g.RunStateChecksForTest()
	if p, tg := effectstest.EffectivePower(t, g, big), effectstest.EffectiveToughness(t, g, big); p != 2 || tg != 2 {
		t.Errorf("their 3/3 is %d/%d, want 2/2", p, tg)
	}
	if p := effectstest.EffectivePower(t, g, mine); p != 2 {
		t.Errorf("my 2/2 is power %d, want 2", p)
	}
	if p := effectstest.EffectivePower(t, g, modok); p != 2 {
		t.Errorf("M.O.D.O.K. shrank himself to power %d", p)
	}
	if g.Battlefield.Contains(small) {
		t.Error("their 1/1 survived a -1/-1")
	}
}

func TestMshA5ModokPaysThreeLifeToConniveOnlyOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	modok := mshA5Modok(g, me.ID)
	spell := effectstest.HandCardForTest(me, "Spare Spell", "Sorcery", "")
	effectstest.AdvanceToMain(t, g)
	life, hand := me.Life, me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, modok, 0, game.ActivateAbilityParams{})
	if me.Life != life-3 {
		t.Errorf("life %d -> %d, want 3 paid", life, me.Life)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Fatalf("hand %d -> %d, want the connive draw", hand, got)
	}
	effectstest.AnswerDiscard(t, g, me.ID, spell)
	if got := effectstest.PlusOnes(g, modok); got != 1 {
		t.Errorf("%d counters after discarding a nonland card, want 1", got)
	}
	// Off your turn it cannot be activated, and costs nothing.
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	misty := effectstest.B27Push(g, me.ID, "Misty Knight, Hero for Hire", "Legendary Creature — Human Detective Hero", effectstest.MshA5MistyOracle, "{1}{R}", 3, 1, "R")
	effectstest.AdvanceToMain(t, g)
	// One card discarded earlier in the turn, the only one in hand at the time.
	me.Hand.Cards = nil
	effectstest.HandCardForTest(me, "Earlier Discard", "Sorcery", "")
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	paid := effectstest.HandCardForTest(me, "Paid Card", "Sorcery", "")
	effectstest.FloatMana(t, g, me, "{C}{C}")
	hand := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, misty, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{paid}})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, misty) {
		t.Error("the ability did not tap Misty Knight")
	}
	// Two cards discarded this turn (one earlier, one as the cost): draw two.
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("hand changed by %d: paid one card and drew two, want a net of 1", got)
	}
}

// --- Jessica Jones, Private Eye ------------------------------------------------------

func mshA5Jessica(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.B27Push(g, owner, "Jessica Jones, Private Eye", "Legendary Creature — Human Detective Hero", effectstest.MshA5JessicaOracle, "{2}{R}", 2, 3, "R")
}

func TestMshA5JessicaJonesStunsHerselfAndExilesTopXWhereXIsHerPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	jessica := mshA5Jessica(g, me.ID)
	ids := effectstest.SeedLibrary(me, "One", "Two", "Three", "Four")
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, jessica, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, jessica) {
		t.Error("the ability did not tap Jessica")
	}
	if got := effectstest.CountersOn(g, jessica, game.CounterStun); got != 1 {
		t.Errorf("%d stun counters, want 1", got)
	}
	for i, id := range ids[:2] {
		if !effectstest.B12ExiledPlayableBy(t, g, id, me.ID) {
			t.Errorf("card %d of the top two is not exiled and playable", i+1)
		}
	}
	if effectstest.B12ZoneOf(g, ids[2]) != game.ZoneLibrary {
		t.Error("a third card was exiled for a power-2 Jessica")
	}
}

func TestMshA5JessicaJonesPowerCountsCountersAndShowsUpAsX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	jessica := mshA5Jessica(g, me.ID)
	ids := effectstest.SeedLibrary(me, "One", "Two", "Three", "Four")
	mshA5PutPlusOne(g, jessica)
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, jessica, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for i, id := range ids[:3] {
		if effectstest.B12ZoneOf(g, id) != game.ZoneExile {
			t.Errorf("card %d of the top three is not exiled for a power-3 Jessica", i+1)
		}
	}
	if effectstest.B12ZoneOf(g, ids[3]) != game.ZoneLibrary {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B27Push(g, me.ID, "Loki, God of Mischief", "Legendary Creature — God Sorcerer Villain", effectstest.MshA5LokiOracle, "{1}{U}", 2, 1, "U")
	bomb := effectstest.PushCatalogPermanent(g, me.ID, "Goblin Bombardment", "Enchantment", effectstest.GoblinBombardmentOracle, false)
	f1 := effectstest.B12Creature(g, me.ID, "Fodder 1", "Creature — Goblin", 1, 1)
	f2 := effectstest.B12Creature(g, me.ID, "Fodder 2", "Creature — Goblin", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 4, 4)
	effectstest.AdvanceToMain(t, g)

	// A spell that targets is not an ability.
	hand := me.Hand.Size()
	mshA5Spell(t, g, mshA5Ref(theirs))
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if got := me.Hand.Size(); got != hand {
		t.Fatalf("hand %d -> %d after a targeting spell, want unchanged", hand, got)
	}
	if err := g.ActivateCatalogAbility(me.ID, bomb, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{f1}, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if got := me.Hand.Size(); got != hand+1 {
		t.Fatalf("hand %d -> %d after my ability targeted a player, want one more", hand, got)
	}
	if err := g.ActivateCatalogAbility(me.ID, bomb, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{f2}, Targets: []game.TargetRef{mshA5Ref(theirs)},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d: a second targeting ability the same turn drew again", got)
	}
}

func TestMshA5LokiIgnoresAnOpponentsAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B27Push(g, me.ID, "Loki, God of Mischief", "Legendary Creature — God Sorcerer Villain", effectstest.MshA5LokiOracle, "{1}{U}", 2, 1, "U")
	theirBomb := effectstest.PushCatalogPermanent(g, opp.ID, "Goblin Bombardment", "Enchantment", effectstest.GoblinBombardmentOracle, false)
	fodder := effectstest.B12Creature(g, opp.ID, "Their Fodder", "Creature — Goblin", 1, 1)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 4, 4)
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(opp.ID, theirBomb, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder}, Targets: []game.TargetRef{mshA5Ref(mine)},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d -> %d when an opponent's ability targeted my creature, want unchanged", hand, got)
	}
}
