package testsr

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	cardsco "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/co"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Appetite for Brains: only a card with mana value 4 or more — X is 0
// in a hand — and it is exiled, not discarded.
func TestAppetiteForBrainsExilesACardWithManaValueFourOrMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	ids := effectstest.RevealHand(victim, effectstest.RhForest(), effectstest.RhBolt(), effectstest.RhFireball(), effectstest.RvBig())

	effectstest.CastAt(t, g, "Appetite for Brains", "Sorcery", effectstest.AppetiteForBrainsOracle, effectstest.PbPlayer(victim.ID))
	pick := effectstest.OpenVariant(t, g)
	if !effectstest.SameIDs(pick.DiscardOptions, ids[3:]) || pick.PickDestination != game.PickExile || pick.PickOptional {
		t.Fatalf("pick = options %v destination %q optional %v; want only the big spell, exile, mandatory",
			pick.DiscardOptions, pick.PickDestination, pick.PickOptional)
	}
	if err := g.ResolvePendingChoice(pick.ID, caster.ID, nil); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("choosing nothing: err = %v, want ErrInvalidParam", err)
	}
	if err := g.ResolvePendingChoice(pick.ID, caster.ID, ids[3:]); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if !g.Exile.Contains(ids[3]) || victim.Graveyard.Contains(ids[3]) || effectstest.RvDiscards(g, ids[3]) != 0 {
		t.Error("the card was not exiled, or was discarded")
	}
}

// Aggressive Negotiations: the exile is not a discard (CR 701.9a). A
// madness card goes to exile with no madness cast offered, Megrim's
// "whenever an opponent discards a card" never sees it, and the second
// clause's counter is placed.
func TestAggressiveNegotiationsExileTriggersNoMadnessOrDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, caster.ID, "Megrim", "Enchantment", effectstest.MegrimOracle, false)
	bear := effectstest.PushCreatureToBattlefieldForTest(g, caster.ID, "My Bear")
	g.WithWriteLock(func() { victim.Hand.Cards = nil })
	temper := effectstest.PushCatalogHandCard(victim, "Fiery Temper", "Instant", effectstest.FieryTemperOracle)
	effectstest.PushCatalogHandCard(victim, "Forest", "Basic Land — Forest", "")
	life := victim.Life

	effectstest.CastAt(t, g, "Aggressive Negotiations", "Sorcery", effectstest.AggressiveNegotiationsOracle,
		game.TargetRef{Kind: game.TargetPlayer, ID: victim.ID, Slot: 0},
		game.TargetRef{Kind: game.TargetCard, ID: bear, Slot: 1})
	pick := effectstest.OpenVariant(t, g)
	if !effectstest.SameIDs(pick.DiscardOptions, []uuid.UUID{temper}) {
		t.Fatalf("options = %v, want the Temper", pick.DiscardOptions)
	}
	if err := g.ResolvePendingChoice(pick.ID, caster.ID, []uuid.UUID{temper}); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Exile.Contains(temper) {
		t.Fatal("Fiery Temper is not in exile")
	}
	if effectstest.RvDiscards(g, temper) != 0 {
		t.Error("exiling the card emitted a discard")
	}
	if offer := effectstest.LatestChoiceOfKind(g, game.PendingChoiceMayCast); offer != nil {
		t.Errorf("madness offered a cast for an exiled, undiscarded card: %+v", offer)
	}
	if victim.Life != life {
		t.Errorf("Megrim dealt damage: life %d, want %d", victim.Life, life)
	}
	if c := effectstest.RvBattlefieldCard(g, bear); c == nil || c.Counters["+1/+1"] != 1 {
		t.Error("the creature did not get its +1/+1 counter")
	}
}

// Agonizing Remorse: a graveyard card may be chosen, land included,
// and exiled from there; the caster loses 1 life. With nothing to
// choose, only the life is lost.
func TestAgonizingRemorseTakesFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	effectstest.RevealHand(victim, effectstest.RhForest())
	victim.Graveyard.Cards = nil
	grave := uuid.New()
	victim.Graveyard.PushTop(game.Card{InstanceID: grave, Name: "Swamp", TypeLine: "Basic Land — Swamp", Owner: victim.ID, Controller: victim.ID})
	life := caster.Life

	effectstest.CastAt(t, g, "Agonizing Remorse", "Sorcery", effectstest.AgonizingRemorseOracle, effectstest.PbPlayer(victim.ID))
	pick := effectstest.OpenVariant(t, g)
	if !effectstest.SameIDs(pick.DiscardOptions, []uuid.UUID{grave}) || !pick.PickFromGraveyard {
		t.Fatalf("options = %v from graveyard %v, want the graveyard land", pick.DiscardOptions, pick.PickFromGraveyard)
	}
	// The life loss is the instruction after the pick, so it waits for the
	// answer (CR 608.2c, #4155).
	if caster.Life != life {
		t.Errorf("caster life = %d with the pick open, want %d", caster.Life, life)
	}
	if err := g.ResolvePendingChoice(pick.ID, caster.ID, []uuid.UUID{grave}); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	g.SettleResolution()
	if caster.Life != life-1 {
		t.Errorf("caster life = %d after the pick, want %d", caster.Life, life-1)
	}
	if !g.Exile.Contains(grave) || victim.Graveyard.Contains(grave) {
		t.Error("the graveyard card was not exiled")
	}

	g2 := effectstest.NewCatalogGame(t)
	caster2, victim2 := g2.Seats[0], g2.Seats[1]
	effectstest.RevealHand(victim2, effectstest.RhForest())
	victim2.Graveyard.Cards = nil
	life2 := caster2.Life
	effectstest.CastAt(t, g2, "Agonizing Remorse", "Sorcery", effectstest.AgonizingRemorseOracle, effectstest.PbPlayer(victim2.ID))
	effectstest.NoVariant(t, g2)
	if caster2.Life != life2-1 {
		t.Errorf("with nothing to exile: caster life = %d, want %d", caster2.Life, life2-1)
	}
}

// Nightsnare: a chosen card is discarded and nothing else happens. If
// you choose nothing — or there is nothing to choose — that player
// discards two cards of their own choice.
func TestNightsnareIfYouDoAndIfYouDont(t *testing.T) {
	t.Run("you do", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		caster, victim := g.Seats[0], g.Seats[1]
		ids := effectstest.RevealHand(victim, effectstest.RhForest(), effectstest.RhBolt(), effectstest.PbDivination())
		effectstest.CastAt(t, g, "Nightsnare", "Sorcery", effectstest.NightsnareOracle, effectstest.PbPlayer(victim.ID))
		pick := effectstest.OpenVariant(t, g)
		if !pick.PickOptional || !effectstest.SameIDs(pick.DiscardOptions, ids[1:]) {
			t.Fatalf("pick = optional %v options %v", pick.PickOptional, pick.DiscardOptions)
		}
		if err := g.ResolvePendingChoice(pick.ID, caster.ID, ids[1:2]); err != nil {
			t.Fatalf("resolve: %v", err)
		}
		if !victim.Graveyard.Contains(ids[1]) {
			t.Error("the chosen card was not discarded")
		}
		if c := effectstest.ChooseCardsChoiceFor(g, victim.ID); c != nil {
			t.Errorf("the player was asked to discard two as well: %+v", c)
		}
	})
	t.Run("you don't", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		caster, victim := g.Seats[0], g.Seats[1]
		effectstest.RevealHand(victim, effectstest.RhForest(), effectstest.RhBolt(), effectstest.PbDivination())
		effectstest.CastAt(t, g, "Nightsnare", "Sorcery", effectstest.NightsnareOracle, effectstest.PbPlayer(victim.ID))
		pick := effectstest.OpenVariant(t, g)
		if err := g.ResolvePendingChoice(pick.ID, caster.ID, nil); err != nil {
			t.Fatalf("choose nothing: %v", err)
		}
		c := effectstest.ChooseCardsChoiceFor(g, victim.ID)
		if c == nil || c.ChooseMin != 2 || c.ChooseMax != 2 {
			t.Fatalf("the player's own discard of two: %+v", c)
		}
	})
	t.Run("nothing to choose", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		victim := g.Seats[1]
		effectstest.RevealHand(victim, effectstest.RhForest(), effectstest.RhForest(), effectstest.RhForest())
		effectstest.CastAt(t, g, "Nightsnare", "Sorcery", effectstest.NightsnareOracle, effectstest.PbPlayer(victim.ID))
		effectstest.NoVariant(t, g)
		if c := effectstest.ChooseCardsChoiceFor(g, victim.ID); c == nil || c.ChooseMax != 2 {
			t.Fatalf("with no nonland card, the player discards two: %+v", c)
		}
	})
}

// Tourach's Canticle: the chosen card is discarded, then a card at
// random from what is left.
func TestTourachsCanticleDiscardsTheChosenCardThenOneAtRandom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	ids := effectstest.RevealHand(victim, effectstest.RhForest(), effectstest.RhBolt(), effectstest.PbDivination())
	effectstest.CastAt(t, g, "Tourach's Canticle", "Sorcery", effectstest.TourachsCanticleOracle, effectstest.PbPlayer(victim.ID))
	pick := effectstest.OpenVariant(t, g)
	if !effectstest.SameIDs(pick.DiscardOptions, ids) {
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
	g := effectstest.NewCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	goyf := game.Card{Name: "Tarmogoyf", TypeLine: "Creature — Lhurgoyf", ManaCost: "{1}{G}",
		Colors: []string{"G"}, OracleID: effectstest.TarmogoyfOracle, Power: 0, Toughness: 1}
	red := game.Card{Name: "Red Ogre", TypeLine: "Creature — Ogre", ManaCost: "{2}{R}", Colors: []string{"R"}, Power: 3, Toughness: 3}
	ids := effectstest.RevealHand(victim, effectstest.RhForest(), effectstest.PbBear(), red, goyf)
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

	effectstest.CastAt(t, g, "Talara's Bane", "Sorcery", effectstest.TalarasBaneOracle, effectstest.PbPlayer(victim.ID))
	pick := effectstest.OpenVariant(t, g)
	if want := []uuid.UUID{ids[1], ids[3]}; !effectstest.SameIDs(pick.DiscardOptions, want) {
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
	g := effectstest.NewCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogWalker(g, victim.ID, "Tamiyo, Collector of Tales", effectstest.TamiyoCollectorOracle, 5)
	ids := effectstest.RevealHand(victim, effectstest.PbBear())
	life := caster.Life

	effectstest.CastAt(t, g, "Talara's Bane", "Sorcery", effectstest.TalarasBaneOracle, effectstest.PbPlayer(victim.ID))
	pick := effectstest.OpenVariant(t, g)
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
	g := effectstest.NewCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	bear := effectstest.PushCreatureToBattlefieldForTest(g, caster.ID, "My Bear")
	vehicle := effectstest.PushCatalogPermanent(g, caster.ID, "Smuggler's Copter", "Artifact — Vehicle", "", false)
	effectstest.PushCatalogPermanent(g, caster.ID, "Sol Ring", "Artifact", "", false)
	effectstest.RevealHand(victim, effectstest.RhBolt())

	effectstest.CastAt(t, g, "Reckoner Shakedown", "Sorcery", effectstest.ReckonerShakedownOracle, effectstest.PbPlayer(victim.ID))
	pick := effectstest.OpenVariant(t, g)
	if err := g.ResolvePendingChoice(pick.ID, caster.ID, nil); err != nil {
		t.Fatalf("choose nothing: %v", err)
	}
	c := effectstest.ChooseCardsChoiceFor(g, caster.ID)
	if c == nil || !effectstest.SameIDs(c.ChooseCards, []uuid.UUID{bear, vehicle}) && !effectstest.SameIDs(c.ChooseCards, []uuid.UUID{vehicle, bear}) {
		t.Fatalf("the counter pick: %+v, want the creature and the Vehicle", c)
	}
	effectstest.AnswerChooseCards(t, g, caster.ID, vehicle)
	if v := effectstest.RvBattlefieldCard(g, vehicle); v == nil || v.Counters["+1/+1"] != 2 {
		t.Error("the Vehicle did not get two +1/+1 counters")
	}

	g2 := effectstest.NewCatalogGame(t)
	caster2, victim2 := g2.Seats[0], g2.Seats[1]
	effectstest.PushCreatureToBattlefieldForTest(g2, caster2.ID, "My Bear")
	ids := effectstest.RevealHand(victim2, effectstest.RhBolt())
	effectstest.CastAt(t, g2, "Reckoner Shakedown", "Sorcery", effectstest.ReckonerShakedownOracle, effectstest.PbPlayer(victim2.ID))
	pick2 := effectstest.OpenVariant(t, g2)
	if err := g2.ResolvePendingChoice(pick2.ID, caster2.ID, ids); err != nil {
		t.Fatalf("choose the card: %v", err)
	}
	if c := effectstest.ChooseCardsChoiceFor(g2, caster2.ID); c != nil {
		t.Errorf("a chosen card still asked for the counters: %+v", c)
	}
}

// Binding Negotiation: choosing nothing offers a face-up exiled card
// that player owns — not a face-down one, not someone else's — and the
// one chosen goes to its owner's graveyard.
func TestBindingNegotiationOtherwiseReturnsAnExiledCardToTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	theirs, hidden, mine := uuid.New(), uuid.New(), uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: theirs, Name: "Exiled", TypeLine: "Sorcery", Owner: victim.ID, Controller: victim.ID})
	g.Exile.PushTop(game.Card{InstanceID: hidden, Name: "Hidden", TypeLine: "Sorcery", Owner: victim.ID, Controller: victim.ID, FaceDown: true})
	g.Exile.PushTop(game.Card{InstanceID: mine, Name: "Mine", TypeLine: "Sorcery", Owner: caster.ID, Controller: caster.ID})
	effectstest.RevealHand(victim, effectstest.RhBolt())

	effectstest.CastAt(t, g, "Binding Negotiation", "Sorcery", effectstest.BindingNegotiationOracle, effectstest.PbPlayer(victim.ID))
	pick := effectstest.OpenVariant(t, g)
	if err := g.ResolvePendingChoice(pick.ID, caster.ID, nil); err != nil {
		t.Fatalf("choose nothing: %v", err)
	}
	c := effectstest.ChooseCardsChoiceFor(g, caster.ID)
	if c == nil || !effectstest.SameIDs(c.ChooseCards, []uuid.UUID{theirs}) || c.ChooseMin != 0 {
		t.Fatalf("the exile pick: %+v, want only their face-up card, optional", c)
	}
	effectstest.AnswerChooseCards(t, g, caster.ID, theirs)
	if !victim.Graveyard.Contains(theirs) || g.Exile.Contains(theirs) {
		t.Error("the exiled card did not go to its owner's graveyard")
	}
}

// Extract the Truth: the first mode's pick may choose nothing even
// with a legal card in the hand (its ruling); the second mode is an
// edict.
func TestExtractTheTruthModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	ids := effectstest.RevealHand(victim, effectstest.RhForest(), effectstest.PbBear(), pbPacifism(), effectstest.RhBolt())
	effectstest.PbModal(t, g, "Extract the Truth", "Sorcery", effectstest.ExtractTheTruthOracle, 0, effectstest.PbPlayer(victim.ID))
	pick := effectstest.OpenVariant(t, g)
	if !pick.PickOptional || !effectstest.SameIDs(pick.DiscardOptions, ids[1:3]) {
		t.Fatalf("pick = optional %v options %v, want the creature and the enchantment", pick.PickOptional, pick.DiscardOptions)
	}
	if err := g.ResolvePendingChoice(pick.ID, caster.ID, nil); err != nil {
		t.Fatalf("choose nothing: %v", err)
	}
	if victim.Hand.Size() != 4 {
		t.Errorf("hand = %d, want 4", victim.Hand.Size())
	}

	g2 := effectstest.NewCatalogGame(t)
	victim2 := g2.Seats[1]
	effectstest.PushCatalogPermanent(g2, victim2.ID, "Pacifism", "Enchantment — Aura", "", false)
	effectstest.PbModal(t, g2, "Extract the Truth", "Sorcery", effectstest.ExtractTheTruthOracle, 1, effectstest.PbPlayer(victim2.ID))
	if effectstest.LatestChoiceOfKindFor(g2, game.PendingChoiceSacrifice, victim2.ID) == nil {
		t.Error("the edict mode asked the opponent for nothing")
	}
}

// Addle: the colour is chosen first; the pick then offers only cards of
// that colour, and a multicoloured card is a card of each of its
// colours.
func TestAddleChoosesAColourThenACardOfIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	gold := game.Card{Name: "Boros Charm", TypeLine: "Instant", ManaCost: "{R}{W}", Colors: []string{"R", "W"}}
	ids := effectstest.RevealHand(victim, effectstest.RhForest(), effectstest.RhBolt(), effectstest.PbBear(), gold)

	effectstest.CastAt(t, g, "Addle", "Sorcery", effectstest.AddleOracle, effectstest.PbPlayer(victim.ID))
	effectstest.NoVariant(t, g)
	color := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceColor, caster.ID)
	if color == nil {
		t.Fatal("no colour was asked for")
	}
	if err := g.ResolveColorChoice(color.ID, caster.ID, "R"); err != nil {
		t.Fatalf("ResolveColorChoice: %v", err)
	}
	pick := effectstest.OpenPick(t, g)
	if want := []uuid.UUID{ids[1], ids[3]}; !effectstest.SameIDs(pick.DiscardOptions, want) || pick.DiscardLabel != "red card" {
		t.Fatalf("pick = options %v label %q, want the two red cards", pick.DiscardOptions, pick.DiscardLabel)
	}
}

// Last Rites: the number of cards you discarded is the number you
// choose. Discarding nothing still reveals the hand.
func TestLastRitesChoosesOneForEachCardDiscarded(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() { caster.Hand.Cards = nil })
	a := effectstest.HandCardForTest(caster, "Card A", "Sorcery", "")
	b := effectstest.HandCardForTest(caster, "Card B", "Sorcery", "")
	effectstest.HandCardForTest(caster, "Card C", "Sorcery", "")
	effectstest.RevealHand(victim, effectstest.RhForest(), effectstest.RhBolt(), effectstest.PbBear(), effectstest.PbDivination())

	effectstest.CastAt(t, g, "Last Rites", "Sorcery", effectstest.LastRitesOracle, effectstest.PbPlayer(victim.ID))
	ask := effectstest.ChooseCardsChoiceFor(g, caster.ID)
	if ask == nil || ask.ChooseMin != 0 || ask.ChooseMax != 3 {
		t.Fatalf("the discard: %+v, want any number of three", ask)
	}
	effectstest.AnswerChooseCards(t, g, caster.ID, a, b)
	pick := effectstest.OpenPick(t, g)
	if pick.Count != 2 || len(pick.DiscardOptions) != 3 {
		t.Fatalf("pick = count %d of %d options, want 2 of the 3 nonland cards", pick.Count, len(pick.DiscardOptions))
	}

	g2 := effectstest.NewCatalogGame(t)
	caster2, victim2 := g2.Seats[0], g2.Seats[1]
	g2.WithWriteLock(func() { caster2.Hand.Cards = nil })
	effectstest.HandCardForTest(caster2, "Card A", "Sorcery", "")
	effectstest.RevealHand(victim2, effectstest.RhBolt())
	effectstest.CastAt(t, g2, "Last Rites", "Sorcery", effectstest.LastRitesOracle, effectstest.PbPlayer(victim2.ID))
	effectstest.AnswerChooseCards(t, g2, caster2.ID)
	effectstest.NoVariant(t, g2)
	if !victim2.Hand.Cards[0].IsKnownTo(caster2.ID) {
		t.Error("discarding nothing did not reveal the hand")
	}
}

// concealingCurtainsCard is the card as the importer builds it: two
// faces, the front materialised.
func concealingCurtainsCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   cardsco.ConcealingCurtainsOracle,
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
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	id := effectstest.PushBattlefieldCardWithTimestamp(g, concealingCurtainsCard(me.ID))
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}{B}{B}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("transform: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.RvBattlefieldCard(g, id); c == nil || c.Name != "Revealing Eye" {
		t.Fatalf("the permanent did not transform: %+v", c)
	}
	effectstest.PickPlayer(t, g, me.ID, victim.ID)
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

// Revealing Eye: transforming into it triggers (CR 701.27e). Choosing a
// card discards it and the player draws a card; choosing nothing does
// neither.
func TestRevealingEyeTriggersOnTransformAndDrawsOnlyIfYouDo(t *testing.T) {
	t.Run("you do", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		ids := effectstest.RevealHand(victim, effectstest.RhForest(), effectstest.RhBolt())
		effectstest.PushLibraryCardForTest(victim, game.Card{Name: "Next Card", TypeLine: "Sorcery"})
		revealingEyeOnTheBattlefield(t, g, me, victim)
		pick := effectstest.OpenVariant(t, g)
		if !pick.PickOptional || !effectstest.SameIDs(pick.DiscardOptions, ids[1:]) {
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
		g := effectstest.NewCatalogGame(t)
		me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		effectstest.RevealHand(victim, effectstest.RhForest(), effectstest.RhBolt())
		effectstest.PushLibraryCardForTest(victim, game.Card{Name: "Next Card", TypeLine: "Sorcery"})
		library := victim.Library.Size()
		revealingEyeOnTheBattlefield(t, g, me, victim)
		pick := effectstest.OpenVariant(t, g)
		if err := g.ResolvePendingChoice(pick.ID, me.ID, nil); err != nil {
			t.Fatalf("choose nothing: %v", err)
		}
		if victim.Hand.Size() != 2 || victim.Library.Size() != library {
			t.Errorf("hand %d library %d: choosing nothing discarded or drew", victim.Hand.Size(), victim.Library.Size())
		}
	})
}
