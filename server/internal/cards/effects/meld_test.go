package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// meld_test.go — ADR 0145, #2699: The Mightstone and Weakstone and
// Urza, Lord Protector meld into Urza, Planeswalker.

const (
	urzaLordProtectorOracle = "df2af646-3e5b-43a3-8f3e-50565889f456"
	urzaPlaneswalkerOracle  = "759406d7-44ae-4260-9ef5-3bb2c92f751a"
	urzaPlaneswalkerPrint   = "40a01679-1a23-4b8f-9a4b-0e5d5e0f5a6c"
)

// urzaMeldPrint is the print the deck importer stamps on both halves.
func urzaMeldPrint() *game.MeldPrint {
	return &game.MeldPrint{
		ResultOracleID:   urzaPlaneswalkerOracle,
		ResultScryfallID: urzaPlaneswalkerPrint,
		Result: game.Face{
			Name:            "Urza, Planeswalker",
			TypeLine:        "Legendary Planeswalker — Urza",
			Colors:          []string{"W", "U"},
			StartingLoyalty: 7,
		},
		ResultNeedsEffect: true,
	}
}

// pushUrzaPair puts Urza, Lord Protector and The Mightstone and
// Weakstone onto the battlefield under owner, each carrying the meld
// print, and returns their IDs.
func pushUrzaPair(g *game.Game, owner uuid.UUID) (urza, stone uuid.UUID) {
	urza = pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Urza, Lord Protector", OracleID: urzaLordProtectorOracle,
		TypeLine: "Legendary Creature — Human Artificer", ManaCost: "{1}{W}{U}",
		Power: 2, Toughness: 4, Colors: []string{"W", "U"}, ColorIdentity: []string{"W", "U"},
		Owner: owner, Controller: owner, Layout: game.LayoutMeld, Meld: urzaMeldPrint(),
		ScryfallID: uuid.NewString(),
	})
	stone = pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Mightstone and Weakstone", OracleID: mightstoneOracle,
		TypeLine: "Legendary Artifact — Powerstone", ManaCost: "{5}",
		Owner: owner, Controller: owner, Layout: game.LayoutMeld, Meld: urzaMeldPrint(),
		ScryfallID: uuid.NewString(),
	})
	return urza, stone
}

// activateUrzaMeld activates Urza's {7} at sorcery speed and lets it
// resolve.
func activateUrzaMeld(t *testing.T, g *game.Game, me *game.Player, urza uuid.UUID) error {
	t.Helper()
	advanceToMainOf(t, g, g.Turn.ActiveSeat)
	for range 7 {
		me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
	if err := g.ActivateCatalogAbility(me.ID, urza, 0, game.ActivateAbilityParams{}); err != nil {
		return err
	}
	passPriorityAroundTable(t, g)
	return nil
}

func meldedUrza(t *testing.T, g *game.Game) game.Card {
	t.Helper()
	var found []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == urzaPlaneswalkerOracle {
			found = append(found, c)
		}
	}
	if len(found) != 1 {
		t.Fatalf("want one Urza, Planeswalker on the battlefield, got %d: %+v", len(found), g.Battlefield.Cards)
	}
	return found[0]
}

func TestUrzaLordProtectorMeldsWithTheMightstone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	urza, stone := pushUrzaPair(g, me.ID)
	if err := activateUrzaMeld(t, g, me, urza); err != nil {
		t.Fatalf("activate: %v", err)
	}
	m := meldedUrza(t, g)
	if !m.IsMelded() || len(m.MeldedFrom) != 2 {
		t.Fatalf("the permanent is not two cards: %+v", m.MeldedFrom)
	}
	if m.Name != "Urza, Planeswalker" || !m.IsPlaneswalker() || m.Controller != me.ID || m.Owner != me.ID {
		t.Errorf("melded permanent = %q %q under %v", m.Name, m.TypeLine, m.Controller)
	}
	if got := m.Counters[game.CounterLoyalty]; got != 7 {
		t.Errorf("loyalty %d, want 7 (CR 306.5b off the back face)", got)
	}
	if mv, ok := m.ParsedManaValue(); !ok || mv != 8 {
		t.Errorf("mana value %d (%v), want 8: {1}{W}{U} + {5} (CR 712.8g)", mv, ok)
	}
	for _, id := range []uuid.UUID{urza, stone} {
		if g.Battlefield.Contains(id) || g.Exile.Contains(id) {
			t.Errorf("card %v left behind outside the melded permanent", id)
		}
	}
	if len(g.Battlefield.Cards) != 1 {
		t.Errorf("battlefield holds %d permanents, want 1", len(g.Battlefield.Cards))
	}
	var meld *game.Event
	for i := range g.Events {
		if g.Events[i].Kind == game.EventMeld {
			meld = &g.Events[i]
		}
	}
	if meld == nil || meld.CardID != m.InstanceID {
		t.Fatalf("no EventMeld naming the melded permanent: %+v", meld)
	}
}

func TestUrzaPlaneswalkerActivatesLoyaltyTwiceEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	urza, _ := pushUrzaPair(g, me.ID)
	if err := activateUrzaMeld(t, g, me, urza); err != nil {
		t.Fatalf("activate: %v", err)
	}
	pw := meldedUrza(t, g).InstanceID
	zero := 2 // "0: Create two 1/1 colorless Soldier artifact creature tokens."
	for i := range 2 {
		if err := g.ActivateCatalogAbility(me.ID, pw, zero, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activation %d: %v", i+1, err)
		}
		passPriorityAroundTable(t, g)
	}
	soldiers := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Soldier" {
			soldiers++
		}
	}
	if soldiers != 4 {
		t.Errorf("%d Soldiers, want 4 from two activations", soldiers)
	}
	if err := g.ActivateCatalogAbility(me.ID, pw, zero, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrLoyaltyAlreadyActivated) {
		t.Errorf("a third activation: %v, want ErrLoyaltyAlreadyActivated", err)
	}
}

func TestUrzaPlaneswalkerPlusTwoDiscountsEverySpellThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	urza, _ := pushUrzaPair(g, me.ID)
	if err := activateUrzaMeld(t, g, me, urza); err != nil {
		t.Fatalf("activate: %v", err)
	}
	pw := meldedUrza(t, g).InstanceID
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, pw, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+2: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
	// Twice: the promise is not spent by the first spell it prices.
	for range 2 {
		if got := priceInHand(t, g, me, "Divination", "Sorcery", "{3}{U}"); got != 2 {
			t.Errorf("Divination costs %d, want 2", got)
		}
	}
	if got := priceInHand(t, g, me, "Bear", "Creature — Bear", "{3}{G}"); got != 4 {
		t.Errorf("a creature spell costs %d, want 4 (not discounted)", got)
	}
}

func TestMeldedPermanentLeavesAsTwoCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	urza, _ := pushUrzaPair(g, me.ID)
	if err := activateUrzaMeld(t, g, me, urza); err != nil {
		t.Fatalf("activate: %v", err)
	}
	m := meldedUrza(t, g)
	before := len(g.Events)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(m.InstanceID); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	names := map[string]bool{}
	for _, c := range me.Graveyard.Cards {
		names[c.Name] = true
		if c.IsMelded() {
			t.Errorf("%s is still melded in the graveyard", c.Name)
		}
	}
	if !names["Urza, Lord Protector"] || !names["The Mightstone and Weakstone"] || names["Urza, Planeswalker"] {
		t.Errorf("graveyard = %v, want the two meld cards", names)
	}
	if g.Battlefield.Contains(m.InstanceID) {
		t.Error("the melded permanent is still on the battlefield")
	}
	ltb, moves := 0, 0
	for _, ev := range g.Events[before:] {
		switch {
		case ev.Kind == game.EventLTB:
			ltb++
		case ev.Kind == game.EventZoneMove && ev.NewZone == game.ZoneGraveyard:
			moves++
		}
	}
	if ltb != 1 || moves != 2 {
		t.Errorf("%d leaves-the-battlefield and %d zone moves, want one permanent and two cards (CR 712.21)", ltb, moves)
	}
}

func TestTheOwnerArrangesAMeldedPermanentsCardsInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	urza, _ := pushUrzaPair(g, me.ID)
	if err := activateUrzaMeld(t, g, me, urza); err != nil {
		t.Fatalf("activate: %v", err)
	}
	m := meldedUrza(t, g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(m.InstanceID); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	ask := latestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID)
	if ask == nil || len(ask.PickOptions) != 2 {
		t.Fatalf("CR 712.21a: no order question for the owner: %+v", g.PendingChoices)
	}
	top := func() string { return me.Graveyard.Cards[len(me.Graveyard.Cards)-1].Name }
	before := top()
	if err := g.ResolveOptionPick(ask.ID, me.ID, 1); err != nil {
		t.Fatalf("answer: %v", err)
	}
	if after := top(); after == before {
		t.Errorf("top of graveyard is still %q after choosing the other order", after)
	}
}

// A restore point holding a melded permanent restores it as one object
// with its two cards, and it still leaves as two (ADR 0145 decision 5).
func TestARestoredMeldedPermanentStillLeavesAsTwoCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	urza, _ := pushUrzaPair(g, me.ID)
	if err := activateUrzaMeld(t, g, me, urza); err != nil {
		t.Fatalf("activate: %v", err)
	}
	r := restoreThroughJSON(t, g)
	m := meldedUrza(t, r)
	if len(m.MeldedFrom) != 2 || !m.MeldedFrom[0].IsMeldCard() || m.MeldedFrom[1].Name != "The Mightstone and Weakstone" {
		t.Fatalf("restored melded permanent lost its cards: %+v", m.MeldedFrom)
	}
	if mv, _ := m.ParsedManaValue(); mv != 8 {
		t.Errorf("restored mana value %d, want 8", mv)
	}
	owner := r.Seats[r.Turn.ActiveSeat]
	r.WithWriteLock(func() {
		if err := r.DestroyPermanentForEffect(m.InstanceID); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	if got := owner.Graveyard.Size(); got != 2 {
		t.Errorf("graveyard holds %d cards, want 2", got)
	}
}

func TestMeldedPermanentBouncesAsTwoCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	urza, stone := pushUrzaPair(g, me.ID)
	if err := activateUrzaMeld(t, g, me, urza); err != nil {
		t.Fatalf("activate: %v", err)
	}
	m := meldedUrza(t, g)
	hand := me.Hand.Size()
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(m.InstanceID); err != nil {
			t.Fatalf("bounce: %v", err)
		}
	})
	if got := me.Hand.Size() - hand; got != 2 {
		t.Fatalf("hand grew by %d, want 2", got)
	}
	if !me.Hand.Contains(stone) && !me.Hand.Contains(urza) {
		t.Error("neither card kept an ID it had before the meld")
	}
	for _, c := range me.Hand.Cards {
		if c.IsMelded() || c.OracleID == urzaPlaneswalkerOracle {
			t.Errorf("hand holds %q, melded=%v", c.Name, c.IsMelded())
		}
	}
}

func TestUrzaDoesNotMeldWithAMightstoneHeDoesNotOwn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	urza, _ := pushUrzaPair(g, me.ID)
	_, theirs := pushUrzaPair(g, opp.ID)
	// Mine to control, theirs to own.
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == theirs {
			g.Battlefield.Cards[i].Controller = me.ID
		}
	}
	// Remove my own Mightstone so the only candidate is the borrowed one.
	for _, c := range g.Battlefield.Cards {
		if c.Name == "The Mightstone and Weakstone" && c.Owner == me.ID {
			g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(c.InstanceID) })
		}
	}
	if err := activateUrzaMeld(t, g, me, urza); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !g.Battlefield.Contains(urza) || !g.Battlefield.Contains(theirs) {
		t.Error("an unmet condition exiled something")
	}
	for _, c := range g.Battlefield.Cards {
		if c.IsMelded() {
			t.Error("melded with a card its controller does not own")
		}
	}
}

func TestACopyOfTheMightstoneIsExiledButNotMelded(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	urza, stone := pushUrzaPair(g, me.ID)
	// A Clone that is a copy of the Mightstone: the name and type are
	// copied, the meld print is not (CR 701.42b).
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == stone {
			g.Battlefield.Cards[i].Meld = nil
			g.Battlefield.Cards[i].Layout = ""
		}
	}
	if err := activateUrzaMeld(t, g, me, urza); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !g.Exile.Contains(urza) || !g.Exile.Contains(stone) {
		t.Fatal("both should have been exiled (CR 701.42c: they stay in exile)")
	}
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == urzaPlaneswalkerOracle {
			t.Error("a pair that isn't one was melded")
		}
	}
}

func TestMeldedCommanderSplitsAndOnlyTheCommanderIsOwedTheCommandZone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	urza, stone := pushUrzaPair(g, me.ID)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == urza {
			g.Battlefield.Cards[i].IsCommander = true
		}
	}
	if err := activateUrzaMeld(t, g, me, urza); err != nil {
		t.Fatalf("activate: %v", err)
	}
	m := meldedUrza(t, g)
	if !m.IsCommander {
		t.Fatal("CR 903.3b: the melded permanent is the commander")
	}
	if m.MeldedFrom[0].Name != "Urza, Lord Protector" {
		t.Errorf("carrier is %q, want the commander card", m.MeldedFrom[0].Name)
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(m.InstanceID); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	inCommand := map[string]bool{}
	for _, c := range me.Command.Cards {
		inCommand[c.Name] = true
	}
	inGrave := map[string]bool{}
	for _, c := range me.Graveyard.Cards {
		inGrave[c.Name] = true
	}
	if inCommand["The Mightstone and Weakstone"] || inGrave["Urza, Planeswalker"] {
		t.Errorf("command %v, graveyard %v", inCommand, inGrave)
	}
	if !inGrave["The Mightstone and Weakstone"] {
		t.Errorf("the Mightstone should be in the graveyard: %v (stone id %v)", inGrave, stone)
	}
	if !inCommand["Urza, Lord Protector"] && !inGrave["Urza, Lord Protector"] {
		t.Errorf("Urza went nowhere: command %v, graveyard %v", inCommand, inGrave)
	}
}

func TestAMeldedPermanentCantBeTurnedFaceDown(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	urza, _ := pushUrzaPair(g, me.ID)
	if err := activateUrzaMeld(t, g, me, urza); err != nil {
		t.Fatalf("activate: %v", err)
	}
	m := meldedUrza(t, g)
	if game.CanTurnFaceDown(m) {
		t.Error("CR 712.16: a melded permanent can't be turned face down")
	}
}

func TestAFlickeredMeldedPermanentReturnsBothCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	urza, _ := pushUrzaPair(g, me.ID)
	if err := activateUrzaMeld(t, g, me, urza); err != nil {
		t.Fatalf("activate: %v", err)
	}
	m := meldedUrza(t, g)
	g.WithWriteLock(func() {
		if err := g.ExileCardThenForEffect(m.InstanceID, func(g *game.Game, exiled bool) error {
			if !exiled {
				t.Fatal("not exiled")
			}
			_, err := g.ReturnFromExileToBattlefieldForEffect(m.InstanceID, me.ID, false)
			return err
		}); err != nil {
			t.Fatalf("flicker: %v", err)
		}
	})
	names := map[string]bool{}
	for _, c := range g.Battlefield.Cards {
		names[c.Name] = true
		if c.IsMelded() {
			t.Error("a flickered melded permanent came back melded")
		}
	}
	if !names["Urza, Lord Protector"] || !names["The Mightstone and Weakstone"] {
		t.Errorf("battlefield = %v, want both cards back (CR 712.21c)", names)
	}
	if len(g.Exile.Cards) != 0 {
		t.Errorf("exile still holds %d cards", len(g.Exile.Cards))
	}
}

const (
	hanweirBattlementsOracle = "0e735ba6-7fd1-4d12-b20c-21525dc1e2b5"
	hanweirTownshipOracle    = "f4905c40-003e-4992-b8d7-3f07ba09c686"
)

func hanweirMeldPrint() *game.MeldPrint {
	return &game.MeldPrint{
		ResultOracleID:   hanweirTownshipOracle,
		ResultScryfallID: uuid.NewString(),
		Result: game.Face{
			Name: "Hanweir, the Writhing Township", TypeLine: "Legendary Creature — Eldrazi Ooze",
			Power: 7, Toughness: 4, Keywords: []string{"trample", "haste"},
		},
	}
}

func TestHanweirBattlementsMeldsWithTheGarrisonAndTheTownshipAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hanweir Garrison", OracleID: b41HanweirGarrisonOracle,
		TypeLine: "Creature — Human Soldier", ManaCost: "{2}{R}", Power: 2, Toughness: 3,
		Owner: me.ID, Controller: me.ID, Layout: game.LayoutMeld, Meld: hanweirMeldPrint(),
	})
	land := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hanweir Battlements", OracleID: hanweirBattlementsOracle,
		TypeLine: "Land", Owner: me.ID, Controller: me.ID, Layout: game.LayoutMeld, Meld: hanweirMeldPrint(),
	})
	advanceToMainOf(t, g, 0)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"},
		game.ManaToken{Color: "R"}, game.ManaToken{Color: "R"})
	if err := g.ActivateCatalogAbility(me.ID, land, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the meld: %v", err)
	}
	passPriorityAroundTable(t, g)
	var township game.Card
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == hanweirTownshipOracle {
			township = c
		}
	}
	if !township.IsMelded() {
		t.Fatalf("no Hanweir, the Writhing Township: %+v", g.Battlefield.Cards)
	}
	if mv, _ := township.ParsedManaValue(); mv != 3 {
		t.Errorf("mana value %d, want 3 ({2}{R} + a land's none)", mv)
	}
	declareAttack(t, g, opp.ID, township.InstanceID)
	passPriorityAroundTable(t, g)
	horrors := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Eldrazi Horror" && c.Controller == me.ID {
			horrors++
			if !c.Tapped || c.AttackingTarget != opp.ID || c.Power != 3 || c.Toughness != 2 {
				t.Errorf("token %+v is not a tapped, attacking 3/2", c)
			}
		}
	}
	if horrors != 2 {
		t.Errorf("%d Eldrazi Horrors, want 2", horrors)
	}
}

func TestABouncedMeldedCommanderSendsOnlyTheCommanderToTheCommandZone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	urza, _ := pushUrzaPair(g, me.ID)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == urza {
			g.Battlefield.Cards[i].IsCommander = true
		}
	}
	if err := activateUrzaMeld(t, g, me, urza); err != nil {
		t.Fatalf("activate: %v", err)
	}
	m := meldedUrza(t, g)
	hand := me.Hand.Size()
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(m.InstanceID); err != nil {
			t.Fatalf("bounce: %v", err)
		}
	})
	var offer *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOptionalReplacement && c.Chooser == me.ID {
			offer = c
		}
	}
	if offer == nil {
		t.Fatalf("CR 903.9b: no command-zone offer for the bounced commander: %+v", g.PendingChoices)
	}
	if err := g.ResolveOptionalReplacement(offer.ID, me.ID, true); err != nil {
		t.Fatalf("answer: %v", err)
	}
	if len(me.Command.Cards) != 1 || me.Command.Cards[0].Name != "Urza, Lord Protector" {
		t.Errorf("command zone = %+v, want Urza alone (CR 903.9c)", me.Command.Cards)
	}
	if me.Hand.Size()-hand != 1 || me.Hand.Cards[len(me.Hand.Cards)-1].Name != "The Mightstone and Weakstone" {
		t.Errorf("the Mightstone should have gone to the hand the bounce named")
	}
}
