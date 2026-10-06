package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #2115, ADR 0116's 2026-10-05 amendment: the revealed-hand pick's
// variants, through the cards that ship on them. One test per card's
// own clause; the engine's rules are pinned in
// game/revealed_hand_pick_test.go.

const (
	appetiteForBrainsOracle      = "4ffc16e0-1778-4aba-ade3-58146c800f5f"
	aggressiveNegotiationsOracle = "c469133e-174d-476b-b135-bbf15e415e72"
	agonizingRemorseOracle       = "1267dfda-eb1a-4963-9fe3-fa619d924d7a"
	nightsnareOracle             = "b9efd1bd-d82a-43ad-aba7-df172b9751ce"
	tourachsCanticleOracle       = "738d2be9-c40b-4a9b-95ed-18d659080918"
	talarasBaneOracle            = "2a573cc3-0a2f-4f06-bbac-c762adb9bae4"
	reckonerShakedownOracle      = "15159091-84ec-4b8c-8d84-8a8bb8baf27d"
	bindingNegotiationOracle     = "86dc3a41-4fb9-4b1f-8279-2b5edbc7e170"
	extractTheTruthOracle        = "18207fe8-41e4-418d-b5f9-e1239c527e67"
	addleOracle                  = "3745aea4-4455-4859-9772-bcfd14e4067b"
	lastRitesOracle              = "3196e1f4-7f94-4eb6-ae2b-ed11ba392552"
	megrimOracle                 = "633ad9e2-9f55-4a1c-9248-661ad4b0e1dc"
)

// openVariant returns the one revealed_hand_pick prompt, or fails.
func openVariant(t *testing.T, g *game.Game) *game.PendingChoice {
	t.Helper()
	var found *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceRevealedHandPick {
			if found != nil {
				t.Fatal("more than one revealed-hand pick is open")
			}
			found = c
		}
	}
	if found == nil {
		t.Fatal("no revealed-hand pick is open")
	}
	return found
}

func noVariant(t *testing.T, g *game.Game) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && (c.Kind == game.PendingChoiceRevealedHandPick || c.Kind == game.PendingChoiceDiscardFromHand) {
			t.Fatalf("a revealed-hand pick is open: %+v", c)
		}
	}
}

// castAt casts a catalog spell at the given targets and resolves it.
func castAt(t *testing.T, g *game.Game, name, typeLine, oracle string, targets ...game.TargetRef) {
	t.Helper()
	castCatalogSpell(t, g, name, typeLine, oracle, targets)
	passPriorityAroundTable(t, g)
}

func rvDiscards(g *game.Game, card uuid.UUID) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventDiscardCard && ev.CardID == card {
			n++
		}
	}
	return n
}

func rvBattlefieldCard(g *game.Game, id uuid.UUID) *game.Card {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return &g.Battlefield.Cards[i]
		}
	}
	return nil
}

func rvBig() game.Card {
	return game.Card{Name: "Big Spell", TypeLine: "Sorcery", ManaCost: "{4}{G}{G}{G}", Colors: []string{"G"}}
}

// Appetite for Brains: only a card with mana value 4 or more — X is 0
// in a hand — and it is exiled, not discarded.
func TestAppetiteForBrainsExilesACardWithManaValueFourOrMore(t *testing.T) {
	g := newCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	ids := revealHand(victim, rhForest(), rhBolt(), rhFireball(), rvBig())

	castAt(t, g, "Appetite for Brains", "Sorcery", appetiteForBrainsOracle, pbPlayer(victim.ID))
	pick := openVariant(t, g)
	if !sameIDs(pick.DiscardOptions, ids[3:]) || pick.PickDestination != game.PickExile || pick.PickOptional {
		t.Fatalf("pick = options %v destination %q optional %v; want only the big spell, exile, mandatory",
			pick.DiscardOptions, pick.PickDestination, pick.PickOptional)
	}
	if err := g.ResolvePendingChoice(pick.ID, caster.ID, nil); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("choosing nothing: err = %v, want ErrInvalidParam", err)
	}
	if err := g.ResolvePendingChoice(pick.ID, caster.ID, ids[3:]); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if !g.Exile.Contains(ids[3]) || victim.Graveyard.Contains(ids[3]) || rvDiscards(g, ids[3]) != 0 {
		t.Error("the card was not exiled, or was discarded")
	}
}

// Aggressive Negotiations: the exile is not a discard (CR 701.9a). A
// madness card goes to exile with no madness cast offered, Megrim's
// "whenever an opponent discards a card" never sees it, and the second
// clause's counter is placed.
func TestAggressiveNegotiationsExileTriggersNoMadnessOrDiscard(t *testing.T) {
	g := newCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, caster.ID, "Megrim", "Enchantment", megrimOracle, false)
	bear := pushCreatureToBattlefieldForTest(g, caster.ID, "My Bear")
	g.WithWriteLock(func() { victim.Hand.Cards = nil })
	temper := pushCatalogHandCard(victim, "Fiery Temper", "Instant", fieryTemperOracle)
	pushCatalogHandCard(victim, "Forest", "Basic Land — Forest", "")
	life := victim.Life

	castAt(t, g, "Aggressive Negotiations", "Sorcery", aggressiveNegotiationsOracle,
		game.TargetRef{Kind: game.TargetPlayer, ID: victim.ID, Slot: 0},
		game.TargetRef{Kind: game.TargetCard, ID: bear, Slot: 1})
	pick := openVariant(t, g)
	if !sameIDs(pick.DiscardOptions, []uuid.UUID{temper}) {
		t.Fatalf("options = %v, want the Temper", pick.DiscardOptions)
	}
	if err := g.ResolvePendingChoice(pick.ID, caster.ID, []uuid.UUID{temper}); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	passPriorityAroundTable(t, g)

	if !g.Exile.Contains(temper) {
		t.Fatal("Fiery Temper is not in exile")
	}
	if rvDiscards(g, temper) != 0 {
		t.Error("exiling the card emitted a discard")
	}
	if offer := latestChoiceOfKind(g, game.PendingChoiceMayCast); offer != nil {
		t.Errorf("madness offered a cast for an exiled, undiscarded card: %+v", offer)
	}
	if victim.Life != life {
		t.Errorf("Megrim dealt damage: life %d, want %d", victim.Life, life)
	}
	if c := rvBattlefieldCard(g, bear); c == nil || c.Counters["+1/+1"] != 1 {
		t.Error("the creature did not get its +1/+1 counter")
	}
}

// Agonizing Remorse: a graveyard card may be chosen, land included,
// and exiled from there; the caster loses 1 life. With nothing to
// choose, only the life is lost.
func TestAgonizingRemorseTakesFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	revealHand(victim, rhForest())
	victim.Graveyard.Cards = nil
	grave := uuid.New()
	victim.Graveyard.PushTop(game.Card{InstanceID: grave, Name: "Swamp", TypeLine: "Basic Land — Swamp", Owner: victim.ID, Controller: victim.ID})
	life := caster.Life

	castAt(t, g, "Agonizing Remorse", "Sorcery", agonizingRemorseOracle, pbPlayer(victim.ID))
	pick := openVariant(t, g)
	if !sameIDs(pick.DiscardOptions, []uuid.UUID{grave}) || !pick.PickFromGraveyard {
		t.Fatalf("options = %v from graveyard %v, want the graveyard land", pick.DiscardOptions, pick.PickFromGraveyard)
	}
	if caster.Life != life-1 {
		t.Errorf("caster life = %d, want %d", caster.Life, life-1)
	}
	if err := g.ResolvePendingChoice(pick.ID, caster.ID, []uuid.UUID{grave}); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if !g.Exile.Contains(grave) || victim.Graveyard.Contains(grave) {
		t.Error("the graveyard card was not exiled")
	}

	g2 := newCatalogGame(t)
	caster2, victim2 := g2.Seats[0], g2.Seats[1]
	revealHand(victim2, rhForest())
	victim2.Graveyard.Cards = nil
	life2 := caster2.Life
	castAt(t, g2, "Agonizing Remorse", "Sorcery", agonizingRemorseOracle, pbPlayer(victim2.ID))
	noVariant(t, g2)
	if caster2.Life != life2-1 {
		t.Errorf("with nothing to exile: caster life = %d, want %d", caster2.Life, life2-1)
	}
}

// Nightsnare: a chosen card is discarded and nothing else happens. If
// you choose nothing — or there is nothing to choose — that player
// discards two cards of their own choice.
func TestNightsnareIfYouDoAndIfYouDont(t *testing.T) {
	t.Run("you do", func(t *testing.T) {
		g := newCatalogGame(t)
		caster, victim := g.Seats[0], g.Seats[1]
		ids := revealHand(victim, rhForest(), rhBolt(), pbDivination())
		castAt(t, g, "Nightsnare", "Sorcery", nightsnareOracle, pbPlayer(victim.ID))
		pick := openVariant(t, g)
		if !pick.PickOptional || !sameIDs(pick.DiscardOptions, ids[1:]) {
			t.Fatalf("pick = optional %v options %v", pick.PickOptional, pick.DiscardOptions)
		}
		if err := g.ResolvePendingChoice(pick.ID, caster.ID, ids[1:2]); err != nil {
			t.Fatalf("resolve: %v", err)
		}
		if !victim.Graveyard.Contains(ids[1]) {
			t.Error("the chosen card was not discarded")
		}
		if c := chooseCardsChoiceFor(g, victim.ID); c != nil {
			t.Errorf("the player was asked to discard two as well: %+v", c)
		}
	})
	t.Run("you don't", func(t *testing.T) {
		g := newCatalogGame(t)
		caster, victim := g.Seats[0], g.Seats[1]
		revealHand(victim, rhForest(), rhBolt(), pbDivination())
		castAt(t, g, "Nightsnare", "Sorcery", nightsnareOracle, pbPlayer(victim.ID))
		pick := openVariant(t, g)
		if err := g.ResolvePendingChoice(pick.ID, caster.ID, nil); err != nil {
			t.Fatalf("choose nothing: %v", err)
		}
		c := chooseCardsChoiceFor(g, victim.ID)
		if c == nil || c.ChooseMin != 2 || c.ChooseMax != 2 {
			t.Fatalf("the player's own discard of two: %+v", c)
		}
	})
	t.Run("nothing to choose", func(t *testing.T) {
		g := newCatalogGame(t)
		victim := g.Seats[1]
		revealHand(victim, rhForest(), rhForest(), rhForest())
		castAt(t, g, "Nightsnare", "Sorcery", nightsnareOracle, pbPlayer(victim.ID))
		noVariant(t, g)
		if c := chooseCardsChoiceFor(g, victim.ID); c == nil || c.ChooseMax != 2 {
			t.Fatalf("with no nonland card, the player discards two: %+v", c)
		}
	})
}

// Tourach's Canticle: the chosen card is discarded, then a card at
// random from what is left.
func TestTourachsCanticleDiscardsTheChosenCardThenOneAtRandom(t *testing.T) {
	g := newCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	ids := revealHand(victim, rhForest(), rhBolt(), pbDivination())
	castAt(t, g, "Tourach's Canticle", "Sorcery", tourachsCanticleOracle, pbPlayer(victim.ID))
	pick := openVariant(t, g)
	if !sameIDs(pick.DiscardOptions, ids) {
		t.Fatalf("options = %v, want the whole hand", pick.DiscardOptions)
	}
	if err := g.ResolvePendingChoice(pick.ID, caster.ID, ids[1:2]); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if !victim.Graveyard.Contains(ids[1]) {
		t.Error("the chosen card was not discarded")
	}
	if victim.Hand.Size() != 1 {
		t.Errorf("hand = %d, want 1 after the chosen and the random discard", victim.Hand.Size())
	}
	if victim.Hand.Contains(ids[1]) {
		t.Error("the chosen card is still in hand")
	}
}

// Talara's Bane: the life gain reads the chosen card in the hand — its
// own characteristic-defining ability applied, so a Tarmogoyf's
// toughness counts the graveyards (CR 113.6a) — and happens before the
// discard (CR 608.2c).
func TestTalarasBaneGainsTheChosenCreaturesToughnessBeforeTheDiscard(t *testing.T) {
	g := newCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	goyf := game.Card{Name: "Tarmogoyf", TypeLine: "Creature — Lhurgoyf", ManaCost: "{1}{G}",
		Colors: []string{"G"}, OracleID: tarmogoyfOracle, Power: 0, Toughness: 1}
	red := game.Card{Name: "Red Ogre", TypeLine: "Creature — Ogre", ManaCost: "{2}{R}", Colors: []string{"R"}, Power: 3, Toughness: 3}
	ids := revealHand(victim, rhForest(), pbBear(), red, goyf)
	// Two card types among the graveyards: a land and an instant.
	for _, gy := range []game.Card{
		{InstanceID: uuid.New(), Name: "Swamp", TypeLine: "Basic Land — Swamp"},
		{InstanceID: uuid.New(), Name: "Shock", TypeLine: "Instant"},
	} {
		gy.Owner, gy.Controller = caster.ID, caster.ID
		caster.Graveyard.PushTop(gy)
	}
	life := caster.Life
	before := len(g.Events)

	castAt(t, g, "Talara's Bane", "Sorcery", talarasBaneOracle, pbPlayer(victim.ID))
	pick := openVariant(t, g)
	if want := []uuid.UUID{ids[1], ids[3]}; !sameIDs(pick.DiscardOptions, want) {
		t.Fatalf("options = %v, want the green Bears and Tarmogoyf %v", pick.DiscardOptions, want)
	}
	if err := g.ResolvePendingChoice(pick.ID, caster.ID, ids[3:]); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if caster.Life != life+3 {
		t.Errorf("caster life = %d, want %d (Tarmogoyf is 2/3 with two types in graveyards)", caster.Life, life+3)
	}
	if !victim.Graveyard.Contains(ids[3]) {
		t.Fatal("Tarmogoyf was not discarded")
	}
	gained, discarded := -1, -1
	for i, ev := range g.Events[before:] {
		switch {
		case ev.Kind == game.EventChangeLife && ev.Target == caster.ID && ev.Amount > 0 && gained < 0:
			gained = i
		case ev.Kind == game.EventDiscardCard && ev.CardID == ids[3]:
			discarded = i
		}
	}
	if gained < 0 || discarded < 0 || gained > discarded {
		t.Errorf("life gain at %d, discard at %d: the gain is printed first", gained, discarded)
	}
}

// Talara's Bane under Tamiyo, Collector of Tales: the opponent can't be
// made to discard (#2178), but the card is still chosen and the life is
// still gained.
func TestTalarasBaneStillGainsWhenTheDiscardIsStopped(t *testing.T) {
	g := newCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	pushCatalogWalker(g, victim.ID, "Tamiyo, Collector of Tales", tamiyoCollectorOracle, 5)
	ids := revealHand(victim, pbBear())
	life := caster.Life

	castAt(t, g, "Talara's Bane", "Sorcery", talarasBaneOracle, pbPlayer(victim.ID))
	pick := openVariant(t, g)
	if err := g.ResolvePendingChoice(pick.ID, caster.ID, ids); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if caster.Life != life+2 {
		t.Errorf("caster life = %d, want %d", caster.Life, life+2)
	}
	if !victim.Hand.Contains(ids[0]) {
		t.Error("the protected player discarded the card")
	}
}

// Reckoner Shakedown: choosing nothing puts two counters on a creature
// or Vehicle you choose; choosing a card does not.
func TestReckonerShakedownIfYouDont(t *testing.T) {
	g := newCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	bear := pushCreatureToBattlefieldForTest(g, caster.ID, "My Bear")
	vehicle := pushCatalogPermanent(g, caster.ID, "Smuggler's Copter", "Artifact — Vehicle", "", false)
	pushCatalogPermanent(g, caster.ID, "Sol Ring", "Artifact", "", false)
	revealHand(victim, rhBolt())

	castAt(t, g, "Reckoner Shakedown", "Sorcery", reckonerShakedownOracle, pbPlayer(victim.ID))
	pick := openVariant(t, g)
	if err := g.ResolvePendingChoice(pick.ID, caster.ID, nil); err != nil {
		t.Fatalf("choose nothing: %v", err)
	}
	c := chooseCardsChoiceFor(g, caster.ID)
	if c == nil || !sameIDs(c.ChooseCards, []uuid.UUID{bear, vehicle}) && !sameIDs(c.ChooseCards, []uuid.UUID{vehicle, bear}) {
		t.Fatalf("the counter pick: %+v, want the creature and the Vehicle", c)
	}
	answerChooseCards(t, g, caster.ID, vehicle)
	if v := rvBattlefieldCard(g, vehicle); v == nil || v.Counters["+1/+1"] != 2 {
		t.Error("the Vehicle did not get two +1/+1 counters")
	}

	g2 := newCatalogGame(t)
	caster2, victim2 := g2.Seats[0], g2.Seats[1]
	pushCreatureToBattlefieldForTest(g2, caster2.ID, "My Bear")
	ids := revealHand(victim2, rhBolt())
	castAt(t, g2, "Reckoner Shakedown", "Sorcery", reckonerShakedownOracle, pbPlayer(victim2.ID))
	pick2 := openVariant(t, g2)
	if err := g2.ResolvePendingChoice(pick2.ID, caster2.ID, ids); err != nil {
		t.Fatalf("choose the card: %v", err)
	}
	if c := chooseCardsChoiceFor(g2, caster2.ID); c != nil {
		t.Errorf("a chosen card still asked for the counters: %+v", c)
	}
}

// Binding Negotiation: choosing nothing offers a face-up exiled card
// that player owns — not a face-down one, not someone else's — and the
// one chosen goes to its owner's graveyard.
func TestBindingNegotiationOtherwiseReturnsAnExiledCardToTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	theirs, hidden, mine := uuid.New(), uuid.New(), uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: theirs, Name: "Exiled", TypeLine: "Sorcery", Owner: victim.ID, Controller: victim.ID})
	g.Exile.PushTop(game.Card{InstanceID: hidden, Name: "Hidden", TypeLine: "Sorcery", Owner: victim.ID, Controller: victim.ID, FaceDown: true})
	g.Exile.PushTop(game.Card{InstanceID: mine, Name: "Mine", TypeLine: "Sorcery", Owner: caster.ID, Controller: caster.ID})
	revealHand(victim, rhBolt())

	castAt(t, g, "Binding Negotiation", "Sorcery", bindingNegotiationOracle, pbPlayer(victim.ID))
	pick := openVariant(t, g)
	if err := g.ResolvePendingChoice(pick.ID, caster.ID, nil); err != nil {
		t.Fatalf("choose nothing: %v", err)
	}
	c := chooseCardsChoiceFor(g, caster.ID)
	if c == nil || !sameIDs(c.ChooseCards, []uuid.UUID{theirs}) || c.ChooseMin != 0 {
		t.Fatalf("the exile pick: %+v, want only their face-up card, optional", c)
	}
	answerChooseCards(t, g, caster.ID, theirs)
	if !victim.Graveyard.Contains(theirs) || g.Exile.Contains(theirs) {
		t.Error("the exiled card did not go to its owner's graveyard")
	}
}

// Extract the Truth: the first mode's pick may choose nothing even
// with a legal card in the hand (its ruling); the second mode is an
// edict.
func TestExtractTheTruthModes(t *testing.T) {
	g := newCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	ids := revealHand(victim, rhForest(), pbBear(), pbPacifism(), rhBolt())
	pbModal(t, g, "Extract the Truth", "Sorcery", extractTheTruthOracle, 0, pbPlayer(victim.ID))
	pick := openVariant(t, g)
	if !pick.PickOptional || !sameIDs(pick.DiscardOptions, ids[1:3]) {
		t.Fatalf("pick = optional %v options %v, want the creature and the enchantment", pick.PickOptional, pick.DiscardOptions)
	}
	if err := g.ResolvePendingChoice(pick.ID, caster.ID, nil); err != nil {
		t.Fatalf("choose nothing: %v", err)
	}
	if victim.Hand.Size() != 4 {
		t.Errorf("hand = %d, want 4", victim.Hand.Size())
	}

	g2 := newCatalogGame(t)
	victim2 := g2.Seats[1]
	pushCatalogPermanent(g2, victim2.ID, "Pacifism", "Enchantment — Aura", "", false)
	pbModal(t, g2, "Extract the Truth", "Sorcery", extractTheTruthOracle, 1, pbPlayer(victim2.ID))
	if latestChoiceOfKindFor(g2, game.PendingChoiceSacrifice, victim2.ID) == nil {
		t.Error("the edict mode asked the opponent for nothing")
	}
}

// Addle: the colour is chosen first; the pick then offers only cards of
// that colour, and a multicoloured card is a card of each of its
// colours.
func TestAddleChoosesAColourThenACardOfIt(t *testing.T) {
	g := newCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	gold := game.Card{Name: "Boros Charm", TypeLine: "Instant", ManaCost: "{R}{W}", Colors: []string{"R", "W"}}
	ids := revealHand(victim, rhForest(), rhBolt(), pbBear(), gold)

	castAt(t, g, "Addle", "Sorcery", addleOracle, pbPlayer(victim.ID))
	noVariant(t, g)
	color := latestChoiceOfKindFor(g, game.PendingChoiceColor, caster.ID)
	if color == nil {
		t.Fatal("no colour was asked for")
	}
	if err := g.ResolveColorChoice(color.ID, caster.ID, "R"); err != nil {
		t.Fatalf("ResolveColorChoice: %v", err)
	}
	pick := openPick(t, g)
	if want := []uuid.UUID{ids[1], ids[3]}; !sameIDs(pick.DiscardOptions, want) || pick.DiscardLabel != "red card" {
		t.Fatalf("pick = options %v label %q, want the two red cards", pick.DiscardOptions, pick.DiscardLabel)
	}
}

// Last Rites: the number of cards you discarded is the number you
// choose. Discarding nothing still reveals the hand.
func TestLastRitesChoosesOneForEachCardDiscarded(t *testing.T) {
	g := newCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() { caster.Hand.Cards = nil })
	a := handCardForTest(caster, "Card A", "Sorcery", "")
	b := handCardForTest(caster, "Card B", "Sorcery", "")
	handCardForTest(caster, "Card C", "Sorcery", "")
	revealHand(victim, rhForest(), rhBolt(), pbBear(), pbDivination())

	castAt(t, g, "Last Rites", "Sorcery", lastRitesOracle, pbPlayer(victim.ID))
	ask := chooseCardsChoiceFor(g, caster.ID)
	if ask == nil || ask.ChooseMin != 0 || ask.ChooseMax != 3 {
		t.Fatalf("the discard: %+v, want any number of three", ask)
	}
	answerChooseCards(t, g, caster.ID, a, b)
	pick := openPick(t, g)
	if pick.Count != 2 || len(pick.DiscardOptions) != 3 {
		t.Fatalf("pick = count %d of %d options, want 2 of the 3 nonland cards", pick.Count, len(pick.DiscardOptions))
	}

	g2 := newCatalogGame(t)
	caster2, victim2 := g2.Seats[0], g2.Seats[1]
	g2.WithWriteLock(func() { caster2.Hand.Cards = nil })
	handCardForTest(caster2, "Card A", "Sorcery", "")
	revealHand(victim2, rhBolt())
	castAt(t, g2, "Last Rites", "Sorcery", lastRitesOracle, pbPlayer(victim2.ID))
	answerChooseCards(t, g2, caster2.ID)
	noVariant(t, g2)
	if !victim2.Hand.Cards[0].IsKnownTo(caster2.ID) {
		t.Error("discarding nothing did not reveal the hand")
	}
}

// concealingCurtainsCard is the card as the importer builds it: two
// faces, the front materialised.
func concealingCurtainsCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   concealingCurtainsOracle,
		Layout:     game.LayoutTransform,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{Name: "Concealing Curtains", TypeLine: "Creature — Wall", ManaCost: "{B}", Colors: []string{"B"}, Power: 0, Toughness: 4},
			{Name: "Revealing Eye", TypeLine: "Creature — Eye Horror", Colors: []string{"B"}, Power: 3, Toughness: 4},
		},
	}
	c.SetFace(0)
	return c
}

// revealingEyeOnTheBattlefield activates Concealing Curtains' transform
// and settles it, aiming the trigger at victim. Returns the permanent.
func revealingEyeOnTheBattlefield(t *testing.T, g *game.Game, me, victim *game.Player) uuid.UUID {
	t.Helper()
	advanceToMainOf(t, g, g.Turn.ActiveSeat)
	id := pushBattlefieldCardWithTimestamp(g, concealingCurtainsCard(me.ID))
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}{B}{B}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("transform: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := rvBattlefieldCard(g, id); c == nil || c.Name != "Revealing Eye" {
		t.Fatalf("the permanent did not transform: %+v", c)
	}
	pickPlayer(t, g, me.ID, victim.ID)
	passPriorityAroundTable(t, g)
	return id
}

// Revealing Eye: transforming into it triggers (CR 701.27e). Choosing a
// card discards it and the player draws a card; choosing nothing does
// neither.
func TestRevealingEyeTriggersOnTransformAndDrawsOnlyIfYouDo(t *testing.T) {
	t.Run("you do", func(t *testing.T) {
		g := newCatalogGame(t)
		me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		ids := revealHand(victim, rhForest(), rhBolt())
		pushLibraryCardForTest(victim, game.Card{Name: "Next Card", TypeLine: "Sorcery"})
		revealingEyeOnTheBattlefield(t, g, me, victim)
		pick := openVariant(t, g)
		if !pick.PickOptional || !sameIDs(pick.DiscardOptions, ids[1:]) {
			t.Fatalf("pick = optional %v options %v", pick.PickOptional, pick.DiscardOptions)
		}
		if err := g.ResolvePendingChoice(pick.ID, me.ID, ids[1:]); err != nil {
			t.Fatalf("resolve: %v", err)
		}
		if !victim.Graveyard.Contains(ids[1]) {
			t.Error("the chosen card was not discarded")
		}
		if victim.Hand.Size() != 2 {
			t.Errorf("hand = %d, want 2: the Forest and the card drawn", victim.Hand.Size())
		}
	})
	t.Run("you don't", func(t *testing.T) {
		g := newCatalogGame(t)
		me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		revealHand(victim, rhForest(), rhBolt())
		pushLibraryCardForTest(victim, game.Card{Name: "Next Card", TypeLine: "Sorcery"})
		library := victim.Library.Size()
		revealingEyeOnTheBattlefield(t, g, me, victim)
		pick := openVariant(t, g)
		if err := g.ResolvePendingChoice(pick.ID, me.ID, nil); err != nil {
			t.Fatalf("choose nothing: %v", err)
		}
		if victim.Hand.Size() != 2 || victim.Library.Size() != library {
			t.Errorf("hand %d library %d: choosing nothing discarded or drew", victim.Hand.Size(), victim.Library.Size())
		}
	})
}
