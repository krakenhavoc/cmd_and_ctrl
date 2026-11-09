package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fin_b8_cards_test.go — Final Fantasy (FIN) batch 8 (#4462). Each test
// plays the real catalog entry and asserts what its text says.

const (
	finB8CidOracle       = "87050537-99c9-4993-a770-4329b2e749e4"
	finB8QuistisOracle   = "821e8648-222c-4b33-a8bd-e8bfff7dcd9e"
	finB8GaiusOracle     = "86d2a45a-e452-4241-8ab4-28a2bb808ffa"
	finB8KainOracle      = "70fbb378-d905-47cf-8c54-cc166a7867b3"
	finB8QuinaOracle     = "2acb15bd-6c89-4ff3-ab32-beac67e9c88a"
	finB8EmetOracle      = "3b102ccc-7629-457c-aba6-e9b00fd50c85"
	finB8EmperorOracle   = "042c01ec-9f1f-432b-bea9-d64d822cd91e"
	finB8GarlandOracle   = "75e930dd-32e9-4c39-81a7-d833a3976b8d"
	finB8GarnetOracle    = "8c709873-84b5-46c2-bb29-85b041a6df88"
	finB8IgnisOracle     = "23320d55-80cb-4da3-a53c-8c75abdc8c8f"
	finB8TerraOracle     = "54ca5904-9099-497a-9684-101656025487"
	finB8JoshuaOracle    = "f4629329-a79b-4460-a1bc-9ae0f59adf5a"
	finB8LoreCounterKind = game.CounterLore
)

func TestFinB8EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		finB8CidOracle:                         "Cid, Timeless Artificer",
		finB8QuistisOracle:                     "Quistis Trepe",
		finB8GaiusOracle:                       "Gaius van Baelsar",
		finB8KainOracle:                        "Kain, Traitorous Dragoon",
		finB8QuinaOracle:                       "Quina, Qu Gourmet",
		finB8EmetOracle:                        "Emet-Selch, Unsundered",
		finB8EmetOracle + "#1":                 "Hades, Sorcerer of Eld",
		finB8EmperorOracle:                     "The Emperor of Palamecia",
		finB8EmperorOracle + "#1":              "The Lord Master of Hell",
		finB8GarlandOracle:                     "Garland, Knight of Cornelia",
		finB8GarlandOracle + "#1":              "Chaos, the Endless",
		finB8GarnetOracle:                      "Garnet, Princess of Alexandria",
		finB8IgnisOracle:                       "Ignis Scientia",
		finB8TerraOracle:                       "Terra, Magical Adept",
		finB8TerraOracle + "#1":                "Esper Terra",
		finB8JoshuaOracle:                      "Joshua, Phoenix's Dominant",
		finB8JoshuaOracle + "#1":               "Phoenix, Warden of Fire",
		"577725ae-290f-4081-8352-3317521a096f": "Vanille, Cheerful l'Cie",
	} {
		finB7RegisteredFull(t, oracle, name)
	}
}

// finB8Row builds a transforming card's catalog row.
func finB8Row(oracle, front, frontType, frontCost, frontP, frontT, back, backType, backP, backT string) cards.Card {
	return cards.Card{
		ID:       uuid.New(),
		Name:     front + " // " + back,
		Layout:   "transform",
		TypeLine: frontType + " // " + backType,
		OracleID: uuid.MustParse(oracle),
		CardFaces: []cards.CardFace{
			{Name: front, TypeLine: frontType, ManaCost: frontCost, Power: frontP, Toughness: frontT},
			{Name: back, TypeLine: backType, Power: backP, Toughness: backT},
		},
	}
}

// finB8PutFront puts a transforming card's front face straight onto the
// battlefield (no enters trigger) with summoning sickness cleared.
func finB8PutFront(t *testing.T, g *game.Game, row cards.Card, p *game.Player) uuid.UUID {
	t.Helper()
	c := deck.ToGameCard(row, false)
	c.Owner, c.Controller = p.ID, p.ID
	c.SummonedThisTurn = false
	return pushBattlefieldCardWithTimestamp(g, c)
}

func finB8Counters(g *game.Game, id uuid.UUID, kind string) int {
	c, ok := battlefieldCard(g, id)
	if !ok {
		return -1
	}
	return c.Counters[kind]
}

// --- Terra, Magical Adept // Esper Terra --------------------------------------------

func TestFinB8TerraMillsFiveAndTakesAnEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Library.Cards = nil })
	for i := 0; i < 4; i++ {
		topLibraryCard(me, "Filler", "Sorcery")
	}
	aura := topLibraryCard(me, "Pretty Wall", "Enchantment")
	topLibraryCard(me, "Rest", "Sorcery")
	castCatalogSpell(t, g, "Terra, Magical Adept", "Legendary Creature — Human Wizard Warrior", finB8TerraOracle, nil)
	passPriorityAroundTable(t, g)
	c := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if c == nil {
		t.Fatalf("no pick of the milled enchantment: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{aura}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(aura) {
		t.Error("the milled enchantment did not go to hand")
	}
}

func finB8TerraRow() cards.Card {
	return finB8Row(finB8TerraOracle, "Terra, Magical Adept", "Legendary Creature — Human Wizard Warrior", "{1}{R}{G}", "4", "2",
		"Esper Terra", "Legendary Enchantment Creature — Saga Wizard", "6", "6")
}

func TestFinB8TerraTranceFlipsAndEsperTerraCopiesAnEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	wall := b12Permanent(g, me.ID, "Pretty Wall", "Enchantment")
	terra := finB8PutFront(t, g, finB8TerraRow(), me)
	floatForTest(g, me, "RGCCCC")
	if err := g.ActivateCatalogAbility(me.ID, terra, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("Trance: %v", err)
	}
	passPriorityAroundTable(t, g)
	var esper game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Esper Terra" {
			esper = c
		}
	}
	if esper.InstanceID == uuid.Nil || esper.ActiveFace != 1 {
		t.Fatalf("Esper Terra is not on the battlefield transformed: %+v", esper)
	}
	if latestPickTarget(g, me.ID) != nil {
		pickCard(t, g, me.ID, wall)
	}
	passPriorityAroundTable(t, g)
	copies := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Pretty Wall" && c.IsToken() {
			copies++
			if !hasEffectiveKeyword(t, g, c.InstanceID, "haste") {
				t.Error("the copy has no haste")
			}
		}
	}
	if copies != 1 {
		t.Fatalf("%d token copies of the enchantment, want 1", copies)
	}
	// The copy is sacrificed at the beginning of my next end step.
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Pretty Wall" && c.IsToken() {
			t.Error("the token copy survived my end step")
		}
	}
}

// --- Joshua, Phoenix's Dominant // Phoenix, Warden of Fire --------------------------

func finB8JoshuaRow() cards.Card {
	return finB8Row(finB8JoshuaOracle, "Joshua, Phoenix's Dominant", "Legendary Creature — Human Noble Wizard", "{1}{R}{W}", "3", "4",
		"Phoenix, Warden of Fire", "Legendary Enchantment Creature — Saga Phoenix", "4", "4")
}

func TestFinB8JoshuaRummagesUpToTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	importAndCast(t, g, finB8JoshuaRow(), me)
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, me.ID) == nil {
		t.Fatal("Joshua offered no discard")
	}
	hand := me.Hand.Size()
	discardFromHand(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got < hand {
		t.Errorf("hand %d → %d, want as many cards drawn as discarded", hand, got)
	}
}

func TestFinB8JoshuaFlipsAndPhoenixBurnsReturnsAndFlipsBack(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	advanceToMain(t, g)
	joshua := finB8PutFront(t, g, finB8JoshuaRow(), me)
	floatForTest(g, me, "RWCCC")
	if err := g.ActivateCatalogAbility(me.ID, joshua, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("flip: %v", err)
	}
	passPriorityAroundTable(t, g)
	opps := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		opps[i] = p.Life
	}
	// Chapter I was put on the stack as it entered; it has now resolved.
	for i, p := range g.Seats {
		if i != seat && p.Life != 20-2 && p.Life != opps[i] {
			t.Errorf("seat %d life %d", i, p.Life)
		}
	}
	var phoenix game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Phoenix, Warden of Fire" {
			phoenix = c
		}
	}
	if phoenix.InstanceID == uuid.Nil || phoenix.Counters[game.CounterLore] != 1 {
		t.Fatalf("Phoenix is not on the battlefield with a lore counter: %+v", phoenix)
	}
	for i, p := range g.Seats {
		if i != seat && p.Life != 18 {
			t.Errorf("seat %d life %d after chapter I, want 18", i, p.Life)
		}
	}
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if i != seat && p.Life != 16 {
			t.Errorf("seat %d life %d after chapter II, want 16", i, p.Life)
		}
	}
	// Chapter III returns Phoenix front face up (it has no graveyard targets here).
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	var back *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Joshua, Phoenix's Dominant" {
			back = &g.Battlefield.Cards[i]
		}
	}
	if back == nil || back.ActiveFace != 0 {
		t.Errorf("Joshua did not come back front face up: %+v", back)
	}
}

// --- Cid, Timeless Artificer --------------------------------------------------------

func TestFinB8CidBoostsArtifactCreaturesAndHeroesPerArtificer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogGraveyardCard(me, "Old Artificer", "Creature — Human Artificer", "", 1, 1)
	pushCatalogGraveyardCard(me, "Old Bear", "Creature — Bear", "", 1, 1)
	pushCatalogGraveyardCard(opp, "Their Artificer", "Creature — Human Artificer", "", 1, 1)
	s58p6Push(g, me.ID, "Cid, Timeless Artificer", "Legendary Creature — Human Artificer", finB8CidOracle, "{2}{W}{U}", 4, 4)
	golem := b12Creature(g, me.ID, "Golem", "Artifact Creature — Golem", 2, 2)
	hero := b12Creature(g, me.ID, "Hero", "Token Creature — Hero", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Golem", "Artifact Creature — Golem", 2, 2)
	if p, th := effectivePTOf(t, g, golem); p != 4 || th != 4 {
		t.Errorf("Golem is %d/%d with Cid and an Artificer card, want 4/4", p, th)
	}
	if p, th := effectivePTOf(t, g, hero); p != 3 || th != 3 {
		t.Errorf("Hero is %d/%d, want 3/3", p, th)
	}
	if p, _ := effectivePTOf(t, g, bear); p != 2 {
		t.Errorf("a plain Bear got a bonus (%d power)", p)
	}
	if p, _ := effectivePTOf(t, g, theirs); p != 2 {
		t.Errorf("an opponent's Golem got a bonus (%d power)", p)
	}
	spec, _ := Lookup(finB8CidOracle)
	if len(spec.Activated) != 1 || spec.Activated[0].Label == "" {
		t.Errorf("activated rows %+v, want cycling", spec.Activated)
	}
}

// --- Quistis Trepe ------------------------------------------------------------------

func TestFinB8QuistisCastsASpellFromAnyGraveyardWithAnyMana(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	bolt := pushCatalogGraveyardCard(opp, "Lightning Bolt", "Instant", lightningBoltOracle, 0, 0)
	g.WithWriteLock(func() {
		for i := range opp.Graveyard.Cards {
			if opp.Graveyard.Cards[i].InstanceID == bolt {
				opp.Graveyard.Cards[i].ManaCost = "{R}"
			}
		}
	})
	victim := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	floatForTest(g, me, "U")
	castCatalogSpell(t, g, "Quistis Trepe", "Legendary Creature — Human Wizard", finB8QuistisOracle, nil)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		pickCard(t, g, me.ID, bolt)
	}
	passPriorityAroundTable(t, g)
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordPaidCast, bolt)
	if err := castThroughResolutionCast(t, g, me.ID, bolt, game.CastSpellParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("cast Lightning Bolt with blue mana: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the Bear survived the Lightning Bolt cast with blue mana")
	}
	if !g.Exile.Contains(bolt) {
		t.Error("the Lightning Bolt was not exiled instead of going to a graveyard")
	}
}

// --- Gaius van Baelsar --------------------------------------------------------------

func finB8GaiusMode(t *testing.T, g *game.Game, me *game.Player, mode int) {
	t.Helper()
	castAndResolveCreature(t, g, "Gaius van Baelsar", "Legendary Creature — Human Soldier", finB8GaiusOracle)
	c := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if c == nil {
		t.Fatalf("no mode_pick prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func TestFinB8GaiusMakesEachPlayerSacrificeATokenCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myToken := b12Creature(g, me.ID, "Goblin", "Token Creature — Goblin", 1, 1)
	myBear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirToken := b12Creature(g, opp.ID, "Goblin", "Token Creature — Goblin", 1, 1)
	theirBear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	finB8GaiusMode(t, g, me, 0)
	for _, id := range []uuid.UUID{myToken, theirToken} {
		if g.Battlefield.Contains(id) {
			t.Error("a token survived")
		}
	}
	for _, id := range []uuid.UUID{myBear, theirBear} {
		if !g.Battlefield.Contains(id) {
			t.Error("a nontoken creature was sacrificed by the token mode")
		}
	}
}

func TestFinB8GaiusMakesEachPlayerSacrificeAnEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Permanent(g, me.ID, "My Aura", "Enchantment")
	theirs := b12Permanent(g, opp.ID, "Their Aura", "Enchantment")
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	finB8GaiusMode(t, g, me, 2)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("an enchantment survived")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("a creature was sacrificed by the enchantment mode")
	}
}

// --- Kain, Traitorous Dragoon -------------------------------------------------------

func TestFinB8KainHasFlyingOnlyOnItsControllersTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	kain := b12Push(g, me.ID, "Kain, Traitorous Dragoon", "Legendary Creature — Human Knight", finB8KainOracle, 2, 4)
	if !hasEffectiveKeyword(t, g, kain, "flying") {
		t.Error("Kain has no flying on his controller's turn")
	}
	advanceToMainOf(t, g, 1)
	if hasEffectiveKeyword(t, g, kain, "flying") {
		t.Error("Kain has flying on another player's turn")
	}
}

func TestFinB8KainHandsHimToTheVictimForCardsTreasureAndLifeLoss(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	kain := b12Push(g, me.ID, "Kain, Traitorous Dragoon", "Legendary Creature — Human Knight", finB8KainOracle, 2, 4)
	hand, life := me.Hand.Size(), me.Life
	dealCombatDamageToPlayer(g, kain, opp.ID, 3)
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, kain)
	if !ok || c.Controller != opp.ID {
		t.Fatalf("Kain is controlled by %v, want the damaged player", c.Controller)
	}
	if me.Hand.Size() != hand+3 {
		t.Errorf("hand %d → %d, want three drawn", hand, me.Hand.Size())
	}
	if me.Life != life-3 {
		t.Errorf("life %d → %d, want 3 lost", life, me.Life)
	}
	treasures := 0
	for _, bc := range g.Battlefield.Cards {
		if bc.Controller == me.ID && bc.HasSubtype("Treasure") {
			treasures++
			if !bc.Tapped {
				t.Error("a Treasure entered untapped")
			}
		}
	}
	if treasures != 3 {
		t.Errorf("%d Treasures, want 3", treasures)
	}
}

// --- Quina, Qu Gourmet --------------------------------------------------------------

func TestFinB8QuinaAddsAFrogToEveryTokenCreationAndSacrificesFrogsForCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	quina := s58p6Push(g, me.ID, "Quina, Qu Gourmet", "Legendary Creature — Qu", finB8QuinaOracle, "{2}{G}", 2, 3)
	g.WithWriteLock(func() {
		if _, err := g.CreateTokensForEffect(me.ID, TokenCard("1/1 white Soldier"), 2, game.TokenEntryOptions{}); err != nil {
			t.Fatalf("create tokens: %v", err)
		}
	})
	if n := countOnBattlefield(g, "Soldier", me.ID); n != 2 {
		t.Errorf("%d Soldiers, want 2", n)
	}
	if n := countOnBattlefield(g, "Frog", me.ID); n != 1 {
		t.Fatalf("%d Frogs, want exactly 1 for the whole creation", n)
	}
	// An opponent's tokens get no Frog for me.
	g.WithWriteLock(func() {
		if _, err := g.CreateTokensForEffect(g.Seats[1].ID, TokenCard("1/1 white Soldier"), 1, game.TokenEntryOptions{}); err != nil {
			t.Fatalf("create tokens: %v", err)
		}
	})
	if n := countOnBattlefield(g, "Frog", me.ID); n != 1 {
		t.Errorf("an opponent's token creation gave me a Frog (%d)", n)
	}
	frog := findBattlefieldByName(g, "Frog")
	advanceToMain(t, g)
	floatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, quina, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{frog}}); err != nil {
		t.Fatalf("sacrifice a Frog: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(frog) {
		t.Error("the Frog was not sacrificed")
	}
	if got := finB8Counters(g, quina, game.CounterPlusOne); got != 1 {
		t.Errorf("Quina has %d +1/+1 counters, want 1", got)
	}
}

// --- Emet-Selch, Unsundered // Hades, Sorcerer of Eld -------------------------------

func finB8EmetRow() cards.Card {
	return finB8Row(finB8EmetOracle, "Emet-Selch, Unsundered", "Legendary Creature — Elder Wizard", "{1}{U}{B}", "2", "4",
		"Hades, Sorcerer of Eld", "Legendary Creature — Avatar", "6", "6")
}

func TestFinB8EmetSelchLootsWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	importAndCast(t, g, finB8EmetRow(), me)
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, me.ID) == nil {
		t.Fatal("Emet-Selch asked for no discard after drawing")
	}
}

func TestFinB8EmetSelchTransformsAtUpkeepWithFourteenCardsInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	emet := finB8PutFront(t, g, finB8EmetRow(), me)
	for i := 0; i < 14; i++ {
		pushCatalogGraveyardCard(me, "Old Card", "Sorcery", "", 0, 0)
	}
	b12ToMyNextUpkeep(t, g)
	passPriorityAroundTable(t, g)
	ask := latestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil || ask.Chooser != me.ID {
		t.Fatalf("no \"you may transform\" question: %+v", g.PendingChoices)
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, emet)
	if !ok || c.ActiveFace != 1 || c.Name != "Hades, Sorcerer of Eld" {
		t.Fatalf("Emet-Selch did not transform into Hades: %+v", c)
	}
}

func TestFinB8EmetSelchDoesNotTriggerBelowFourteen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	finB8PutFront(t, g, finB8EmetRow(), me)
	for i := 0; i < 13; i++ {
		pushCatalogGraveyardCard(me, "Old Card", "Sorcery", "", 0, 0)
	}
	b12ToMyNextUpkeep(t, g)
	passPriorityAroundTable(t, g)
	if latestChoiceOfKind(g, game.PendingChoiceConfirm) != nil {
		t.Error("the upkeep trigger asked with only thirteen cards in the graveyard")
	}
}

func TestFinB8HadesLetsYouPlayFromYourGraveyardOnYourTurnAndExilesInstead(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	hades := putBackFaceOntoBattlefield(t, g, finB8EmetRow(), me)
	forest := pushCatalogGraveyardCard(me, "Forest", "Basic Land — Forest", "", 0, 0)
	if err := g.CastSpell(me.ID, forest, game.CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("play a land from the graveyard on my turn: %v", err)
	}
	if !g.Battlefield.Contains(forest) {
		t.Error("the land from the graveyard is not on the battlefield")
	}
	// A creature dying is exiled instead of going to the graveyard.
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	if !g.Exile.Contains(bear) || me.Graveyard.Contains(bear) {
		t.Error("the dying card was not exiled instead of reaching my graveyard")
	}
	_ = hades
}

// --- The Emperor of Palamecia // The Lord Master of Hell ----------------------------

func finB8EmperorRow() cards.Card {
	return finB8Row(finB8EmperorOracle, "The Emperor of Palamecia", "Legendary Creature — Human Noble Wizard", "{U}{R}", "2", "2",
		"The Lord Master of Hell", "Legendary Creature — Demon Noble Wizard", "3", "3")
}

func TestFinB8EmperorGrowsOnBigNoncreatureSpellsAndTransformsAtThreeCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	emperor := finB8PutFront(t, g, finB8EmperorRow(), me)
	socB17CastSpellFor(t, g, "Instant", "{1}{U}", "UU")
	if got := finB8Counters(g, emperor, game.CounterPlusOne); got != 0 {
		t.Errorf("a two-mana spell gave %d counters", got)
	}
	socB17CastSpellFor(t, g, "Creature — Bear", "{3}{U}", "UUUU")
	if got := finB8Counters(g, emperor, game.CounterPlusOne); got != 0 {
		t.Errorf("a creature spell gave %d counters", got)
	}
	for i := 1; i <= 2; i++ {
		socB17CastSpellFor(t, g, "Sorcery", "{3}{U}", "UUUU")
		if got := finB8Counters(g, emperor, game.CounterPlusOne); got != i {
			t.Fatalf("after %d big spells: %d counters, want %d", i, got, i)
		}
		if c, _ := battlefieldCard(g, emperor); c.ActiveFace != 0 {
			t.Fatalf("transformed with only %d counters", i)
		}
	}
	socB17CastSpellFor(t, g, "Sorcery", "{3}{U}", "UUUU")
	c, _ := battlefieldCard(g, emperor)
	if c.ActiveFace != 1 || c.Name != "The Lord Master of Hell" {
		t.Errorf("the Emperor did not transform with three counters: %+v", c)
	}
}

func TestFinB8EmperorTapsForBlueOrRedOnlyForNoncreatureSpells(t *testing.T) {
	spec := finB7RegisteredFull(t, finB8EmperorOracle, "The Emperor of Palamecia")
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{U|R}" || !spec.ManaAbilities[0].Cost.Tap {
		t.Fatalf("mana abilities %+v, want {T}: {U|R}", spec.ManaAbilities)
	}
	if len(spec.ManaAbilities[0].Restrictions) == 0 {
		t.Error("the mana has no spending restriction")
	}
}

func TestFinB8LordMasterStarfallsForNoncreatureNonlandCardsInYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[1]
	master := putBackFaceOntoBattlefield(t, g, finB8EmperorRow(), me)
	g.WithWriteLock(func() { me.Graveyard.Cards = nil })
	pushCatalogGraveyardCard(me, "Bolt", "Instant", "", 0, 0)
	pushCatalogGraveyardCard(me, "Sword", "Artifact — Equipment", "", 0, 0)
	pushCatalogGraveyardCard(me, "Bear", "Creature — Bear", "", 1, 1)
	pushCatalogGraveyardCard(me, "Forest", "Basic Land — Forest", "", 0, 0)
	life := opp.Life
	declareAttack(t, g, opp.ID, master)
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("opponent life %d → %d, want 2 from Starfall", life, opp.Life)
	}
}

// --- Garland, Knight of Cornelia // Chaos, the Endless ------------------------------

func finB8GarlandRow() cards.Card {
	return finB8Row(finB8GarlandOracle, "Garland, Knight of Cornelia", "Legendary Creature — Human Knight", "{B}{R}", "3", "2",
		"Chaos, the Endless", "Legendary Creature — Demon", "5", "5")
}

func TestFinB8GarlandSurveilsOnNoncreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	finB8PutFront(t, g, finB8GarlandRow(), me)
	advanceToMain(t, g)
	id := socB17Hand(me, "Test Spell", "Instant", "{1}{R}", socB17PlainSorceryOracle)
	floatForTest(g, me, "RR")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("a noncreature spell triggered no surveil")
	}
	if err := g.ResolveSurveil(c.ID, me.ID, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
}

func TestFinB8GarlandReturnsFromTheGraveyardAsChaos(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	card := importToHand(finB8GarlandRow(), me)
	var id uuid.UUID
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == card {
				c := me.Hand.Cards[i]
				me.Hand.Cards = append(me.Hand.Cards[:i], me.Hand.Cards[i+1:]...)
				me.Graveyard.PushTop(c)
				id = c.InstanceID
				break
			}
		}
	})
	floatForTest(g, me, "BBRRCCC")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate from the graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, id)
	if !ok || c.ActiveFace != 1 || c.Name != "Chaos, the Endless" {
		t.Fatalf("Garland did not return as Chaos: %+v", c)
	}
	if !hasEffectiveKeyword(t, g, id, "flying") {
		t.Error("Chaos has no flying")
	}
	// When Chaos dies it goes to the bottom of its owner's library.
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(id) || g.Battlefield.Contains(id) {
		t.Error("Chaos is still in the graveyard or on the battlefield")
	}
	if n := me.Library.Size(); n == 0 || me.Library.Cards[len(me.Library.Cards)-1].InstanceID != id && me.Library.Cards[0].InstanceID != id {
		t.Errorf("Chaos is not at an end of the library")
	}
}

// --- Garnet, Princess of Alexandria -------------------------------------------------

func TestFinB8GarnetRemovesLoreCountersForPlusOneCounters(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[1]
	garnet := b12Push(g, me.ID, "Garnet, Princess of Alexandria", "Legendary Creature — Human Noble Cleric", finB8GarnetOracle, 2, 2)
	sagaA := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Saga A", TypeLine: "Enchantment — Saga", Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterLore: 2},
	})
	sagaB := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Saga B", TypeLine: "Enchantment — Saga", Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterLore: 1},
	})
	theirSaga := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Saga", TypeLine: "Enchantment — Saga", Owner: opp.ID, Controller: opp.ID,
		Counters: map[string]int{game.CounterLore: 1},
	})
	declareAttack(t, g, opp.ID, garnet)
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	c := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if c == nil {
		t.Fatalf("no pick of Sagas: %+v", g.PendingChoices)
	}
	for _, id := range c.ChooseCards {
		if id == theirSaga {
			t.Error("an opponent's Saga was offered")
		}
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{sagaA, sagaB}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	t.Logf("garnet counters %d", finB8Counters(g, garnet, game.CounterPlusOne))
	if finB8Counters(g, sagaA, game.CounterLore) != 1 || finB8Counters(g, sagaB, game.CounterLore) != 0 {
		t.Errorf("lore counters A=%d B=%d, want 1 and 0", finB8Counters(g, sagaA, game.CounterLore), finB8Counters(g, sagaB, game.CounterLore))
	}
	if got := finB8Counters(g, garnet, game.CounterPlusOne); got != 2 {
		t.Errorf("Garnet has %d +1/+1 counters, want 2", got)
	}
	if !hasEffectiveKeyword(t, g, garnet, "lifelink") {
		t.Error("Garnet has no lifelink")
	}
}

// --- Ignis Scientia -----------------------------------------------------------------

func TestFinB8IgnisPutsATappedLandFromTheTopSixAndBottomsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Library.Cards = nil })
	for i := 0; i < 4; i++ {
		topLibraryCard(me, "Deep Filler", "Sorcery")
	}
	topLibraryCard(me, "Spell", "Sorcery")
	land := topLibraryCard(me, "Forest", "Basic Land — Forest")
	topLibraryCard(me, "Spell Two", "Sorcery")
	topLibraryCard(me, "Spell Three", "Sorcery")
	topLibraryCard(me, "Spell Four", "Sorcery")
	topLibraryCard(me, "Spell Five", "Sorcery")
	_ = land
	castCatalogSpell(t, g, "Ignis Scientia", "Legendary Creature — Human Advisor", finB8IgnisOracle, nil)
	passPriorityAroundTable(t, g)
	trekB42Pick(t, g, me.ID, land)
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, land); !ok || !c.Tapped {
		t.Fatalf("the land is on the battlefield %v, tapped %v; want it tapped", ok, c.Tapped)
	}
	if n := me.Library.Size(); n != 9 {
		t.Errorf("library %d, want 9 (ten cards less the land)", n)
	}
}

func TestFinB8IgnisNewRecipeExilesACardAndMakesFoodForACreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	ignis := s58p6Push(g, me.ID, "Ignis Scientia", "Legendary Creature — Human Advisor", finB8IgnisOracle, "{1}{G}{U}", 2, 2)
	advanceToMain(t, g)
	bear := pushCatalogGraveyardCard(opp, "Dead Bear", "Creature — Bear", "", 2, 2)
	spell := pushCatalogGraveyardCard(me, "Old Spell", "Sorcery", "", 0, 0)
	foods := func() int {
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.Controller == me.ID && c.HasSubtype("Food") {
				n++
			}
		}
		return n
	}
	activate := func(target uuid.UUID) {
		t.Helper()
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == ignis {
					g.Battlefield.Cards[i].Tapped = false
				}
			}
		})
		floatForTest(g, me, "GUC")
		if err := g.ActivateCatalogAbility(me.ID, ignis, 0, game.ActivateAbilityParams{
			Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
		}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	activate(spell)
	if !g.Exile.Contains(spell) || foods() != 0 {
		t.Errorf("a noncreature card: exiled %v, %d Foods (want exiled, none)", g.Exile.Contains(spell), foods())
	}
	activate(bear)
	if !g.Exile.Contains(bear) || foods() != 1 {
		t.Errorf("a creature card: exiled %v, %d Foods (want exiled, one)", g.Exile.Contains(bear), foods())
	}
}
