package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// kindle_the_inner_flame_test.go — #4689: "Flashback—{1}{R}, Behold three
// Elementals" (CR 702.34a, CR 701.4a). The flashback cost names three
// distinct objects on alt_cost_ids, each an Elemental the caster controls
// or an Elemental card in their hand; the hand cards are revealed, nothing
// moves, and the card is exiled as it leaves the stack.

const (
	kindleOracle    = "c7ca6371-73bb-4717-a148-0804b8f4c1e6"
	kindleName      = "Kindle the Inner Flame"
	kindleTypeLine  = "Kindred Sorcery — Elemental"
	elementalType   = "Creature — Elemental"
	kindleFlashback = "flashback"
)

// kindleBoard is a Kindle in its owner's graveyard at a main phase, two
// Elemental creatures on the battlefield and one Elemental card in hand.
type kindleBoard struct {
	g          *game.Game
	me, opp    *game.Player
	kindle     uuid.UUID
	smoke      uuid.UUID // an Elemental creature I control, the copy's target
	ember      uuid.UUID // another Elemental creature I control
	inHand     uuid.UUID // an Elemental card in my hand
	notElement uuid.UUID // a Goblin card in my hand
}

func newKindleBoard(t *testing.T) *kindleBoard {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b := &kindleBoard{g: g, me: me, opp: opp}
	b.kindle = seedGraveyardCard(t, g, kindleName, kindleTypeLine, kindleOracle)
	b.smoke = apaPush(g, me.ID, me.ID, game.Card{Name: "Smoke Elemental", TypeLine: elementalType, Power: 3, Toughness: 3})
	b.ember = apaPush(g, me.ID, me.ID, game.Card{Name: "Ember Elemental", TypeLine: elementalType, Power: 1, Toughness: 1})
	b.inHand = handCard(me, "Flame Elemental", elementalType)
	b.notElement = handCard(me, "Goblin Friend", "Creature — Goblin")
	return b
}

func (b *kindleBoard) flashback(pay ...uuid.UUID) error {
	return castForAltCost(b.g, b.kindle, "graveyard", kindleFlashback, pay, []game.TargetRef{{Kind: game.TargetCard, ID: b.smoke}})
}

// refusedCleanly checks that a refused flashback left the table as it was.
func (b *kindleBoard) refusedCleanly(t *testing.T, what string) {
	t.Helper()
	if !b.me.Graveyard.Contains(b.kindle) {
		t.Fatalf("%s: the refused cast moved Kindle out of the graveyard", what)
	}
	if n := stackLen(b.g); n != 0 {
		t.Fatalf("%s: a refused cast left %d object(s) on the stack", what, n)
	}
	if got := revealedEvents(b.g, b.inHand); got != 0 {
		t.Fatalf("%s: the refused cast revealed the Elemental card", what)
	}
}

// Three Elementals split between the battlefield and the hand pay it: the
// hand card is revealed and stays, the permanents are neither tapped nor
// revealed, the token copy arrives with haste, and the card is exiled.
func TestKindleTheInnerFlameFlashbackBeholdsThreeElementals(t *testing.T) {
	b := newKindleBoard(t)
	if err := b.flashback(b.smoke, b.ember, b.inHand); err != nil {
		t.Fatalf("flashback beholding three Elementals: %v", err)
	}
	if !b.me.Hand.Contains(b.inHand) {
		t.Fatal("beholding moved the Elemental card out of the hand")
	}
	if got := revealedEvents(b.g, b.inHand); got != 1 {
		t.Fatalf("%d reveal events for the beheld card, want 1", got)
	}
	for _, id := range []uuid.UUID{b.smoke, b.ember} {
		if !onBattlefield(b.g, id) {
			t.Fatal("beholding moved a permanent")
		}
		if findBattlefieldCardForTest(b.g, id).Tapped {
			t.Error("beholding tapped a permanent")
		}
		if got := revealedEvents(b.g, id); got != 0 {
			t.Errorf("choosing a permanent revealed it %d times", got)
		}
	}
	item := b.g.StackMeta[b.kindle]
	if item == nil || item.AltCost != kindleFlashback || len(item.Paid.AltCostObjects) != 3 {
		t.Fatalf("stack item = %+v, want the flashback claim and the three beheld objects recorded", item)
	}
	passPriorityAroundTable(t, b.g)
	if got := b43TokensNamed(b.g, b.me.ID, "Smoke Elemental"); got != 1 {
		t.Fatalf("token copies of the target = %d, want 1", got)
	}
	for i := range b.g.Battlefield.Cards {
		if c := &b.g.Battlefield.Cards[i]; c.Name == "Smoke Elemental" && c.IsToken() && !hasKeywordForTest(c.Keywords, "haste") {
			t.Errorf("the token should have haste: %v", c.Keywords)
		}
	}
	if !b.g.Exile.Contains(b.kindle) {
		t.Error("the flashed-back Kindle was not exiled")
	}
}

// All three from the battlefield, and all three from the hand, pay it too.
func TestKindleTheInnerFlameFlashbackFromOneZone(t *testing.T) {
	b := newKindleBoard(t)
	third := apaPush(b.g, b.me.ID, b.me.ID, game.Card{Name: "Spark Elemental", TypeLine: elementalType, Power: 1, Toughness: 1})
	if err := b.flashback(b.smoke, b.ember, third); err != nil {
		t.Fatalf("three Elemental permanents: %v", err)
	}
	if got := revealedEvents(b.g, b.inHand); got != 0 {
		t.Errorf("an unnamed hand card was revealed")
	}

	b = newKindleBoard(t)
	two := handCard(b.me, "Second Flame", elementalType)
	three := handCard(b.me, "Third Flame", elementalType)
	if err := b.flashback(b.inHand, two, three); err != nil {
		t.Fatalf("three Elemental cards in hand: %v", err)
	}
	for _, id := range []uuid.UUID{b.inHand, two, three} {
		if !b.me.Hand.Contains(id) || revealedEvents(b.g, id) != 1 {
			t.Errorf("hand card %v: kept %v, revealed %d times; want kept and revealed once", id, b.me.Hand.Contains(id), revealedEvents(b.g, id))
		}
	}
}

// A changeling is every creature type in every zone, so it is an Elemental
// to behold, from the hand or the battlefield.
func TestKindleTheInnerFlameBeholdsAChangeling(t *testing.T) {
	b := newKindleBoard(t)
	mimic := uuid.New()
	b.me.Hand.PushTop(game.Card{InstanceID: mimic, Name: "Mimic", TypeLine: "Creature — Shapeshifter", Keywords: []string{"changeling"}, Owner: b.me.ID, Controller: b.me.ID})
	if err := b.flashback(b.smoke, b.ember, mimic); err != nil {
		t.Fatalf("a changeling card in hand as the third Elemental: %v", err)
	}
}

// Refusals, each leaving the table as it was: fewer than three, more than
// three, the same object twice, a non-Elemental, an opponent's Elemental
// permanent or card, the spell itself, and nothing at all.
func TestKindleTheInnerFlameBeholdRefusals(t *testing.T) {
	b := newKindleBoard(t)
	theirs := apaPush(b.g, b.opp.ID, b.opp.ID, game.Card{Name: "Their Elemental", TypeLine: elementalType, Power: 1, Toughness: 1})
	theirCard := handCard(b.opp, "Their Flame", elementalType)
	extra := handCard(b.me, "Extra Flame", elementalType)
	for name, pay := range map[string][]uuid.UUID{
		"nothing named":           nil,
		"two Elementals":          {b.smoke, b.inHand},
		"four Elementals":         {b.smoke, b.ember, b.inHand, extra},
		"the same one twice":      {b.smoke, b.smoke, b.inHand},
		"a Goblin card":           {b.smoke, b.ember, b.notElement},
		"an opponent's Elemental": {b.smoke, b.ember, theirs},
		"an opponent's card":      {b.smoke, b.ember, theirCard},
		"the spell itself":        {b.smoke, b.ember, b.kindle},
	} {
		if err := b.flashback(pay...); err == nil {
			t.Fatalf("%s: the flashback was accepted", name)
		}
		b.refusedCleanly(t, name)
	}
}

// With only two Elementals anywhere, the flashback offer is withheld (CR
// 601.2b: it can't be paid) and the cast is refused with nothing moved.
func TestKindleTheInnerFlameNeedsThreeElementals(t *testing.T) {
	b := newKindleBoard(t)
	b.g.WithWriteLock(func() { _ = b.g.DestroyPermanentForEffect(b.ember, game.DestroyOptions{}) })
	var payable bool
	b.g.ReadSnapshot(func() {
		alt := game.AlternativeCostByKey(kindleOracle, kindleFlashback)
		payable = b.g.AlternativeCostPayableLocked(b.me.ID, b.kindle, alt)
	})
	if payable {
		t.Fatal("the offer is payable with two Elementals")
	}
	if err := b.flashback(b.smoke, b.inHand); err == nil {
		t.Fatal("a flashback beholding two Elementals was accepted")
	}
	b.refusedCleanly(t, "two Elementals")
}

// The printed cost from the hand beholds nothing: an announced behold is
// refused there, and the hand cast resolves into the graveyard.
func TestKindleTheInnerFlameHandCastBeholdsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	target := apaPush(g, me.ID, me.ID, game.Card{Name: "Smoke Elemental", TypeLine: elementalType, Power: 3, Toughness: 3})
	kindle := handCardAtMain(t, g, kindleName, kindleTypeLine, kindleOracle)
	at := []game.TargetRef{{Kind: game.TargetCard, ID: target}}
	if err := castForAltCost(g, kindle, "", kindleFlashback, nil, at); err == nil {
		t.Fatal("flashback was claimed from the hand")
	}
	if err := g.CastSpell(me.ID, kindle, game.CastSpellParams{AltCostIDs: []uuid.UUID{target}, Targets: at}); err == nil {
		t.Fatal("a hand cast for the printed cost accepted behold picks")
	}
	if err := g.CastSpell(me.ID, kindle, game.CastSpellParams{Targets: at}); err != nil {
		t.Fatalf("hand cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(kindle) {
		t.Error("Kindle cast from hand did not reach the graveyard")
	}
}

// The flashback cast is a restore point: with the spell on the stack, a
// snapshot brings back the flashback claim, and the restored spell still
// makes the copy and is exiled.
func TestKindleTheInnerFlameFlashbackSurvivesASnapshot(t *testing.T) {
	b := newKindleBoard(t)
	if err := b.flashback(b.smoke, b.ember, b.inHand); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	restored := restoreRoundTrip(t, b.g, false)
	item := restored.StackMeta[b.kindle]
	if item == nil || item.AltCost != kindleFlashback || len(item.Paid.AltCostObjects) != 3 {
		t.Fatalf("restored stack item = %+v, want the flashback claim and its three objects", item)
	}
	passPriorityAroundTable(t, restored)
	if got := b43TokensNamed(restored, b.me.ID, "Smoke Elemental"); got != 1 {
		t.Errorf("token copies after the restore = %d, want 1", got)
	}
	if !restored.Exile.Contains(b.kindle) {
		t.Error("the restored flashed-back Kindle was not exiled")
	}
}

// A snapshot taken before the cast, while the player is still choosing
// what to behold, restores a table where the same flashback is accepted.
func TestKindleTheInnerFlameFlashbackAfterARestore(t *testing.T) {
	b := newKindleBoard(t)
	restored := restoreRoundTrip(t, b.g, false)
	if err := castForAltCost(restored, b.kindle, "graveyard", kindleFlashback, []uuid.UUID{b.smoke, b.ember, b.inHand},
		[]game.TargetRef{{Kind: game.TargetCard, ID: b.smoke}}); err != nil {
		t.Fatalf("flashback on the restored table: %v", err)
	}
}

// --- Register's guards --------------------------------------------------

func TestRegisterRefusesMalformedAltCostBehold(t *testing.T) {
	good := FlashbackBehold("{1}{R}", 3, "Elemental", "Elementals")
	if good.Label != "Flashback—{1}{R}, Behold three Elementals" || good.PayLabel != "three Elementals" || !good.ExileOnLeavingStack || good.FromZone != game.ZoneGraveyard {
		t.Fatalf("FlashbackBehold = %+v", good)
	}
	checkAltCostBehold("Kindle", good)
	for name, mutate := range map[string]func(*game.AlternativeCost){
		"not a behold": func(a *game.AlternativeCost) { a.Behold = &game.RevealCost{Subtype: "Elemental"} },
		"no type":      func(a *game.AlternativeCost) { a.Behold = &game.RevealCost{Behold: true} },
		"a colour": func(a *game.AlternativeCost) {
			a.Behold = &game.RevealCost{Subtype: "Elemental", Behold: true, Color: "R"}
		},
		"no count":    func(a *game.AlternativeCost) { a.BeholdCount = 0 },
		"count alone": func(a *game.AlternativeCost) { a.Behold = nil },
	} {
		t.Run(name, func(t *testing.T) {
			bad := good
			mutate(&bad)
			mustPanic(t, "build it with FlashbackBehold", func() { checkAltCostBehold("Bad", bad) })
		})
	}
	two := good
	two.ExileFromHand = CardInYourHand("a red card", OfColor("R"))
	if n := altCostCardComponents(two); n != 2 {
		t.Errorf("a behold beside a pitch counts %d card components, want 2", n)
	}
}
