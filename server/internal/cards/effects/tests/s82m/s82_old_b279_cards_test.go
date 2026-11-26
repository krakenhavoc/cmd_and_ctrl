package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b279_cards_test.go — Eternal Masters, batch 279 (#5542). Every test
// plays the real catalog entry through the engine.

const (
	b279ToothOracle     = "587368eb-068c-44a3-ba8c-5ad0f59f880f"
	b279TrenchesOracle  = "b43f40e6-c0ad-4a12-b75d-f2ba12629bfe"
	b279GlareOracle     = "bcfc112e-0482-4616-8ca9-32338a37fa82"
	b279BodyguardOracle = "c8c38570-0b42-4bfc-896f-f57f9a11ba9d"
	b279CommuneOracle   = "6042a806-6e6a-4f1d-9563-5932a17f32f9"
	b279QuietOracle     = "938f48a9-766e-41d3-a425-9c3a6adf675b"
	b279TidalOracle     = "c5f5b7e5-af1e-4c8a-8d44-3e2152955d4e"
	b279HumbleOracle    = "d2d41f12-18d8-473b-9308-3f02d7d791f0"
	b279TherapyOracle   = "018107d1-ebf7-4657-bc91-c8374af55cc5"
	b279MonkOracle      = "1f47ff5a-dd93-46de-a1ca-b412631218bf"
	b279TorrentOracle   = "3ec1b309-8c26-4dc4-b4c3-f97c2c3f8a68"
)

func TestB279EveryCardIsRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		b279ToothOracle: "Tooth and Claw", b279TrenchesOracle: "Goblin Trenches", b279GlareOracle: "Glare of Subdual",
		b279BodyguardOracle: "Benevolent Bodyguard", b279CommuneOracle: "Commune with the Gods", b279QuietOracle: "Quiet Speculation",
		b279TidalOracle: "Tidal Wave", b279HumbleOracle: "Humble", b279TherapyOracle: "Cabal Therapy",
		b279MonkOracle: "Monk Idealist", b279TorrentOracle: "Torrent of Souls",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

func b279Target(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func b279Grave(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

// --- Tooth and Claw ---

func TestB279ToothAndClawSacrificesTwoCreaturesForACarnivore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	tooth := effectstest.PushCatalogPermanent(g, me.ID, "Tooth and Claw", "Enchantment", b279ToothOracle, false)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, tooth, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{a}}); err == nil {
		t.Fatal("one creature paid for a two-creature sacrifice")
	}
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, tooth, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{a, b}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("the two creatures were not sacrificed")
	}
	toks := effectstest.TokensNamed(g, "Carnivore")
	if len(toks) != 1 {
		t.Fatalf("%d Carnivore tokens, want 1", len(toks))
	}
	if p, tough := effectstest.PtOf(t, g, toks[0].InstanceID); p != 3 || tough != 1 {
		t.Errorf("Carnivore is %d/%d, want 3/1", p, tough)
	}
	if c := effectstest.LayeredCard(t, g, toks[0].InstanceID); len(c.Colors) != 1 || c.Colors[0] != "R" {
		t.Errorf("Carnivore colors %v, want red", c.Colors)
	}
}

// --- Goblin Trenches ---

func TestB279GoblinTrenchesSacrificesALandForTwoGoblinSoldiers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	trenches := effectstest.PushCatalogPermanent(g, me.ID, "Goblin Trenches", "Enchantment", b279TrenchesOracle, false)
	effectstest.FillPool(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, trenches, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no land to sacrifice")
	}
	land := effectstest.PushCatalogPermanent(g, me.ID, "Plains", "Basic Land — Plains", "", false)
	if err := g.ActivateCatalogAbility(me.ID, trenches, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{land}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land was not sacrificed")
	}
	toks := effectstest.TokensNamed(g, "Goblin Soldier")
	if len(toks) != 2 {
		t.Fatalf("%d Goblin Soldier tokens, want 2", len(toks))
	}
	if c := effectstest.LayeredCard(t, g, toks[0].InstanceID); len(c.Colors) != 2 {
		t.Errorf("Goblin Soldier colors %v, want red and white", c.Colors)
	}
}

func TestB279GoblinTrenchesNeedsTwoMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	trenches := effectstest.PushCatalogPermanent(g, me.ID, "Goblin Trenches", "Enchantment", b279TrenchesOracle, false)
	land := effectstest.PushCatalogPermanent(g, me.ID, "Plains", "Basic Land — Plains", "", false)
	effectstest.FillPool(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, trenches, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{land}}); err == nil {
		t.Fatal("one mana paid for {2}")
	}
}

// --- Glare of Subdual ---

func TestB279GlareTapsAnUntappedCreatureToTapATarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	glare := effectstest.PushCatalogPermanent(g, me.ID, "Glare of Subdual", "Enchantment", b279GlareOracle, false)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, glare, 0, game.ActivateAbilityParams{Strict: true, Targets: b279Target(glare), TapIDs: []uuid.UUID{mine}}); err == nil {
		t.Fatal("an enchantment was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, glare, 0, game.ActivateAbilityParams{Strict: true, Targets: b279Target(theirs), TapIDs: []uuid.UUID{mine}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.LayeredCard(t, g, theirs).Tapped {
		t.Error("the target was not tapped")
	}
	if !effectstest.LayeredCard(t, g, mine).Tapped {
		t.Error("the creature that paid the cost is not tapped")
	}
	if err := g.ActivateCatalogAbility(me.ID, glare, 0, game.ActivateAbilityParams{Strict: true, Targets: b279Target(theirs), TapIDs: []uuid.UUID{mine}}); err == nil {
		t.Fatal("activated again with no untapped creature of mine")
	}
}

// --- Benevolent Bodyguard ---

func TestB279BodyguardSacrificesItselfToProtectAnotherCreatureFromAChosenColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	guard := effectstest.B12Push(g, me.ID, "Benevolent Bodyguard", "Creature — Human Cleric", b279BodyguardOracle, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, guard, 0, game.ActivateAbilityParams{Strict: true, Targets: b279Target(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	for i := 0; i < 8 && effectstest.PendingOfKind(g, game.PendingChoiceColor) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	effectstest.AnswerColor(t, g, me.ID, "R")
	if g.Battlefield.Contains(guard) {
		t.Error("the Bodyguard was not sacrificed")
	}
	c, _ := effectstest.BattlefieldCard(g, bear)
	if !game.HasKeyword(&c, game.ProtectionFromColor("R")) {
		t.Error("the Bear did not gain protection from red")
	}
	if game.HasKeyword(&c, game.ProtectionFromColor("U")) {
		t.Error("the Bear gained protection from a color that was not chosen")
	}
}

// --- Commune with the Gods ---

func b279Stack(g *game.Game, p *game.Player, cards ...game.Card) []uuid.UUID {
	ids := make([]uuid.UUID, len(cards))
	g.WithWriteLock(func() {
		p.Library.Cards = nil
		for i := len(cards) - 1; i >= 0; i-- {
			c := cards[i]
			c.InstanceID, c.Owner, c.Controller = uuid.New(), p.ID, p.ID
			ids[i] = c.InstanceID
			p.Library.PushTop(c)
		}
	})
	return ids
}

func b279Commune(t *testing.T, pick int) (g *game.Game, me *game.Player, ids []uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ids = b279Stack(g, me,
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"}, game.Card{Name: "Bolt", TypeLine: "Instant"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"}, game.Card{Name: "Pacifism", TypeLine: "Enchantment — Aura"},
		game.Card{Name: "Ritual", TypeLine: "Sorcery"}, game.Card{Name: "Deep", TypeLine: "Creature — Bear"})
	effectstest.B06AddMana(me, "G", "C")
	effectstest.CastCatalogSpell(t, g, "Commune with the Gods", "Sorcery", b279CommuneOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if pick >= 0 {
		effectstest.AnswerChooseCards(t, g, me.ID, ids[pick])
	} else {
		effectstest.AnswerChooseCards(t, g, me.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return g, me, ids
}

func TestB279CommuneTakesAnEnchantmentAndPutsTheRestOfTheFiveInTheGraveyard(t *testing.T) {
	_, me, ids := b279Commune(t, 3)
	if !me.Hand.Contains(ids[3]) {
		t.Error("the chosen enchantment did not reach the hand")
	}
	for _, i := range []int{0, 1, 2, 4} {
		if !me.Graveyard.Contains(ids[i]) {
			t.Errorf("revealed card %d is not in the graveyard", i)
		}
	}
	if !me.Library.Contains(ids[5]) {
		t.Error("the sixth card left the library")
	}
}

func TestB279CommuneMayTakeNothingAndStillMillsFive(t *testing.T) {
	_, me, ids := b279Commune(t, -1)
	for i := 0; i < 5; i++ {
		if !me.Graveyard.Contains(ids[i]) {
			t.Errorf("revealed card %d is not in the graveyard", i)
		}
	}
}

func TestB279CommuneCannotTakeAnInstant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ids := b279Stack(g, me, game.Card{Name: "Bolt", TypeLine: "Instant"}, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	effectstest.B06AddMana(me, "G", "C")
	effectstest.CastCatalogSpell(t, g, "Commune with the Gods", "Sorcery", b279CommuneOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, ids[1])
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(ids[0]) || !me.Hand.Contains(ids[1]) {
		t.Error("the creature should be taken and the instant left")
	}
	if !me.Graveyard.Contains(ids[0]) {
		t.Error("the instant did not go to the graveyard")
	}
}

// --- Quiet Speculation ---

func TestB279QuietSpeculationPutsUpToThreeFlashbackCardsInTheTargetsGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	fb := func(n string) game.Card {
		return game.Card{Name: n, TypeLine: "Sorcery", OracleID: b279TherapyOracle}
	}
	ids := b279Stack(g, opp, game.Card{Name: "Plain", TypeLine: "Sorcery"}, fb("Fb1"), fb("Fb2"), fb("Fb3"), fb("Fb4"))
	effectstest.B06AddMana(me, "U", "C")
	effectstest.CastCatalogSpell(t, g, "Quiet Speculation", "Sorcery", b279QuietOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, ids[1], ids[2], ids[3])
	effectstest.PassPriorityAroundTable(t, g)
	got := 0
	for _, id := range ids[1:4] {
		if opp.Graveyard.Contains(id) {
			got++
		}
	}
	if got != 3 {
		t.Errorf("%d of the three chosen flashback cards are in the opponent's graveyard, want 3", got)
	}
	if !opp.Library.Contains(ids[0]) || !opp.Library.Contains(ids[4]) {
		t.Error("a card that was not chosen left the library")
	}
}

// --- Tidal Wave ---

func TestB279TidalWaveMakesAFiveFiveWallThatIsSacrificedAtTheNextEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "U", "C", "C")
	effectstest.CastCatalogSpell(t, g, "Tidal Wave", "Instant", b279TidalOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	walls := effectstest.TokensNamed(g, "Wall")
	if len(walls) != 1 {
		t.Fatalf("%d Wall tokens, want 1", len(walls))
	}
	wall := walls[0].InstanceID
	if p, tough := effectstest.PtOf(t, g, wall); p != 5 || tough != 5 {
		t.Errorf("the Wall is %d/%d, want 5/5", p, tough)
	}
	effectstest.AssertKeywords(t, g, wall, "defender")
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wall) {
		t.Error("the Wall survived the end step")
	}
}

// --- Humble ---

func TestB279HumbleStripsAbilitiesAndSetsBasePowerAndToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	big := effectstest.B12Push(g, opp.ID, "Elder", "Creature — Dragon", "", 5, 5)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, big).Counters = map[string]int{game.CounterPlusOne: 2}
		g.BumpLayerVersionForTest()
	})
	effectstest.B06AddMana(me, "W", "C")
	effectstest.CastCatalogSpell(t, g, "Humble", "Instant", b279HumbleOracle, b279Target(big))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, big); p != 2 || tough != 3 {
		t.Errorf("the Dragon with two counters is %d/%d, want base 0/1 plus the counters (2/3)", p, tough)
	}
}

// --- Cabal Therapy ---

func b279Therapy(t *testing.T, name string) (g *game.Game, me, opp *game.Player, one, two, other, land uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	opp.Hand.Cards = nil
	one = effectstest.HandCard(opp, "Counterspell", "Instant")
	two = effectstest.HandCard(opp, "Counterspell", "Instant")
	other = effectstest.HandCard(opp, "Opt", "Instant")
	land = effectstest.HandCard(opp, "Island", "Basic Land — Island")
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "B")
	effectstest.CastCatalogSpell(t, g, "Cabal Therapy", "Sorcery", b279TherapyOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ActNameFor(t, g, me.ID, name)
	effectstest.PassPriorityAroundTable(t, g)
	return
}

func TestB279CabalTherapyDiscardsEveryCardWithTheNamedName(t *testing.T) {
	_, _, opp, one, two, other, land := b279Therapy(t, "counterspell")
	if !opp.Graveyard.Contains(one) || !opp.Graveyard.Contains(two) {
		t.Error("not every card with the named name was discarded")
	}
	if !opp.Hand.Contains(other) || !opp.Hand.Contains(land) {
		t.Error("a card with another name was discarded")
	}
}

func TestB279CabalTherapyNamingALandDiscardsNothing(t *testing.T) {
	_, _, opp, _, _, _, land := b279Therapy(t, "Island")
	if !opp.Hand.Contains(land) {
		t.Error("a land was discarded for a name that is not a nonland card name")
	}
}

func TestB279CabalTherapyFlashbackSacrificesACreatureAndIsExiled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	opp.Hand.Cards = nil
	card := effectstest.HandCard(opp, "Opt", "Instant")
	effectstest.AdvanceToMain(t, g)
	therapy := effectstest.PushCatalogGraveyardCard(me, "Cabal Therapy", "Sorcery", b279TherapyOracle, 0, 0)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	if err := g.CastSpell(me.ID, therapy, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "flashback",
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Fatal("flashed back with no creature to sacrifice")
	}
	if err := g.CastSpell(me.ID, therapy, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "flashback", AltCostIDs: []uuid.UUID{fodder},
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ActNameFor(t, g, me.ID, "Opt")
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature was not sacrificed")
	}
	if !opp.Graveyard.Contains(card) {
		t.Error("the named card was not discarded")
	}
	if !g.Exile.Contains(therapy) {
		t.Error("the flashed-back Cabal Therapy was not exiled")
	}
}

// --- Monk Idealist ---

func TestB279MonkIdealistReturnsAnEnchantmentCardNotAnotherType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ench := b279Grave(me, "Pacifism", "Enchantment — Aura")
	art := b279Grave(me, "Rock", "Artifact")
	effectstest.B06AddMana(me, "W", "C", "C")
	effectstest.CastCatalogSpell(t, g, "Monk Idealist", "Creature — Human Monk Cleric", b279MonkOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickIfAsked(t, g, me.ID, ench)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ench) {
		t.Error("the enchantment card did not return to hand")
	}
	if !me.Graveyard.Contains(art) {
		t.Error("the artifact card left the graveyard")
	}
}

// --- Torrent of Souls ---

func b279Torrent(t *testing.T, pool string) (g *game.Game, me *game.Player, dead, ally uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	dead = effectstest.PushCatalogGraveyardCard(me, "Dead Bear", "Creature — Bear", "", 2, 2)
	ally = effectstest.B12Creature(g, me.ID, "Ally", "Creature — Bear", 2, 2)
	effectstest.FloatForTest(g, me, pool)
	spell := effectstest.HandCardOf(g, me, "Torrent of Souls", "Sorcery", "{4}{B/R}", b279TorrentOracle)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: dead}, {Kind: game.TargetPlayer, ID: me.ID}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return
}

func TestB279TorrentWithBlackSpentReturnsTheCreatureOnly(t *testing.T) {
	g, _, dead, ally := b279Torrent(t, "BCCCC")
	if !g.Battlefield.Contains(dead) {
		t.Error("{B} was spent but the creature card did not return")
	}
	if p, _ := effectstest.PtOf(t, g, ally); p != 2 {
		t.Errorf("power %d, want 2: no {R} was spent", p)
	}
}

func TestB279TorrentWithRedSpentPumpsOnly(t *testing.T) {
	g, _, dead, ally := b279Torrent(t, "RCCCC")
	if g.Battlefield.Contains(dead) {
		t.Error("{R} only was spent but the creature card returned")
	}
	if p, _ := effectstest.PtOf(t, g, ally); p != 4 {
		t.Errorf("power %d, want 4 (+2/+0)", p)
	}
	effectstest.AssertKeywords(t, g, ally, "haste")
}

func TestB279TorrentWithBothSpentDoesBothAndPumpsTheReturnedCreature(t *testing.T) {
	g, _, dead, ally := b279Torrent(t, "BRCCC")
	if !g.Battlefield.Contains(dead) {
		t.Fatal("the creature card did not return")
	}
	if p, _ := effectstest.PtOf(t, g, dead); p != 4 {
		t.Errorf("the returned creature has power %d, want 4", p)
	}
	if p, _ := effectstest.PtOf(t, g, ally); p != 4 {
		t.Errorf("power %d, want 4", p)
	}
}
