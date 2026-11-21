package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsj "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/j"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b := &kindleBoard{g: g, me: me, opp: opp}
	b.kindle = effectstest.SeedGraveyardCard(t, g, effectstest.KindleName, effectstest.KindleTypeLine, effectstest.KindleOracle)
	b.smoke = effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Smoke Elemental", TypeLine: effectstest.ElementalType, Power: 3, Toughness: 3})
	b.ember = effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Ember Elemental", TypeLine: effectstest.ElementalType, Power: 1, Toughness: 1})
	b.inHand = effectstest.HandCard(me, "Flame Elemental", effectstest.ElementalType)
	b.notElement = effectstest.HandCard(me, "Goblin Friend", "Creature — Goblin")
	return b
}

func (b *kindleBoard) flashback(pay ...uuid.UUID) error {
	return effectstest.CastForAltCost(b.g, b.kindle, "graveyard", effectstest.KindleFlashback, pay, []game.TargetRef{{Kind: game.TargetCard, ID: b.smoke}})
}

// refusedCleanly checks that a refused flashback left the table as it was.
func (b *kindleBoard) refusedCleanly(t *testing.T, what string) {
	t.Helper()
	if !b.me.Graveyard.Contains(b.kindle) {
		t.Fatalf("%s: the refused cast moved Kindle out of the graveyard", what)
	}
	if n := effectstest.StackLen(b.g); n != 0 {
		t.Fatalf("%s: a refused cast left %d object(s) on the stack", what, n)
	}
	if got := effectstest.RevealedEvents(b.g, b.inHand); got != 0 {
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
	if got := effectstest.RevealedEvents(b.g, b.inHand); got != 1 {
		t.Fatalf("%d reveal events for the beheld card, want 1", got)
	}
	for _, id := range []uuid.UUID{b.smoke, b.ember} {
		if !effects.OnBattlefield(b.g, id) {
			t.Fatal("beholding moved a permanent")
		}
		if effectstest.FindBattlefieldCardForTest(b.g, id).Tapped {
			t.Error("beholding tapped a permanent")
		}
		if got := effectstest.RevealedEvents(b.g, id); got != 0 {
			t.Errorf("choosing a permanent revealed it %d times", got)
		}
	}
	item := b.g.StackMeta[b.kindle]
	if item == nil || item.AltCost != effectstest.KindleFlashback || len(item.Paid.AltCostObjects) != 3 {
		t.Fatalf("stack item = %+v, want the flashback claim and the three beheld objects recorded", item)
	}
	effectstest.PassPriorityAroundTable(t, b.g)
	if got := effectstest.B43TokensNamed(b.g, b.me.ID, "Smoke Elemental"); got != 1 {
		t.Fatalf("token copies of the target = %d, want 1", got)
	}
	for i := range b.g.Battlefield.Cards {
		if c := &b.g.Battlefield.Cards[i]; c.Name == "Smoke Elemental" && c.IsToken() && !effectstest.HasKeywordForTest(c.Keywords, "haste") {
			t.Errorf("the token should have haste: %v", c.Keywords)
		}
	}
	if !b.g.Exile.Contains(b.kindle) {
		t.Error("the flashed-back Kindle was not exiled")
	}
}

// #4950: the front half is Electroduplicate's, with Kindle's own grant. The
// token carries "At the beginning of the end step, sacrifice this token." as
// a copiable value (CR 707.9a), so a token copy of it made by Cackling
// Counterpart has the trigger too, and both are sacrificed at the end step.
// The card is Full: no caveat is left.
func TestKindleTheInnerFlameCopyOfTheTokenIsSacrificedAtEndStep(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.KindleOracle)
	if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Kindle is registered %v as %v with caveats %v, want Full with none", ok, spec.Completeness, spec.Caveats)
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	smoke := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Smoke Elemental", TypeLine: effectstest.ElementalType, Power: 3, Toughness: 3})
	effectstest.CastCatalogSpell(t, g, effectstest.KindleName, effectstest.KindleTypeLine, effectstest.KindleOracle, []game.TargetRef{{Kind: game.TargetCard, ID: smoke}})
	effectstest.PassPriorityAroundTable(t, g)
	tokens, _ := effectstest.CopiesOnBattlefield(g, "Smoke Elemental")
	if len(tokens) != 1 {
		t.Fatalf("%d token copies, want 1", len(tokens))
	}
	token := tokens[0].InstanceID
	if want := game.GrantKey(cardsj.KindleTheInnerFlameGrant); len(tokens[0].GrantedAbilities) != 1 || tokens[0].GrantedAbilities[0] != want {
		t.Fatalf("the token's granted abilities are %v, want [%s]", tokens[0].GrantedAbilities, want)
	}

	effectstest.CastCatalogSpell(t, g, "Cackling Counterpart", "Instant", effectstest.CacklingCounterpartOracle, []game.TargetRef{{Kind: game.TargetCard, ID: token}})
	effectstest.PassPriorityAroundTable(t, g)
	tokens, _ = effectstest.CopiesOnBattlefield(g, "Smoke Elemental")
	if len(tokens) != 2 {
		t.Fatalf("%d Smoke Elemental tokens after the second copy, want 2", len(tokens))
	}

	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassThroughEndStepTriggers(t, g)
	for _, c := range tokens {
		if effectstest.OnBattlefieldByID(g, c.InstanceID) || !effectstest.SacrificedInLog(g, c.InstanceID) {
			t.Errorf("token %v was not sacrificed at the end step", c.InstanceID)
		}
	}
	if !effectstest.OnBattlefieldByID(g, smoke) {
		t.Error("the original Elemental was sacrificed")
	}
}

// All three from the battlefield, and all three from the hand, pay it too.
func TestKindleTheInnerFlameFlashbackFromOneZone(t *testing.T) {
	b := newKindleBoard(t)
	third := effectstest.ApaPush(b.g, b.me.ID, b.me.ID, game.Card{Name: "Spark Elemental", TypeLine: effectstest.ElementalType, Power: 1, Toughness: 1})
	if err := b.flashback(b.smoke, b.ember, third); err != nil {
		t.Fatalf("three Elemental permanents: %v", err)
	}
	if got := effectstest.RevealedEvents(b.g, b.inHand); got != 0 {
		t.Errorf("an unnamed hand card was revealed")
	}

	b = newKindleBoard(t)
	two := effectstest.HandCard(b.me, "Second Flame", effectstest.ElementalType)
	three := effectstest.HandCard(b.me, "Third Flame", effectstest.ElementalType)
	if err := b.flashback(b.inHand, two, three); err != nil {
		t.Fatalf("three Elemental cards in hand: %v", err)
	}
	for _, id := range []uuid.UUID{b.inHand, two, three} {
		if !b.me.Hand.Contains(id) || effectstest.RevealedEvents(b.g, id) != 1 {
			t.Errorf("hand card %v: kept %v, revealed %d times; want kept and revealed once", id, b.me.Hand.Contains(id), effectstest.RevealedEvents(b.g, id))
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
	theirs := effectstest.ApaPush(b.g, b.opp.ID, b.opp.ID, game.Card{Name: "Their Elemental", TypeLine: effectstest.ElementalType, Power: 1, Toughness: 1})
	theirCard := effectstest.HandCard(b.opp, "Their Flame", effectstest.ElementalType)
	extra := effectstest.HandCard(b.me, "Extra Flame", effectstest.ElementalType)
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
		alt := game.AlternativeCostByKey(effectstest.KindleOracle, effectstest.KindleFlashback)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	target := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Smoke Elemental", TypeLine: effectstest.ElementalType, Power: 3, Toughness: 3})
	kindle := effectstest.HandCardAtMain(t, g, effectstest.KindleName, effectstest.KindleTypeLine, effectstest.KindleOracle)
	at := []game.TargetRef{{Kind: game.TargetCard, ID: target}}
	if err := effectstest.CastForAltCost(g, kindle, "", effectstest.KindleFlashback, nil, at); err == nil {
		t.Fatal("flashback was claimed from the hand")
	}
	if err := g.CastSpell(me.ID, kindle, game.CastSpellParams{AltCostIDs: []uuid.UUID{target}, Targets: at}); err == nil {
		t.Fatal("a hand cast for the printed cost accepted behold picks")
	}
	if err := g.CastSpell(me.ID, kindle, game.CastSpellParams{Targets: at}); err != nil {
		t.Fatalf("hand cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	restored := effectstest.RestoreRoundTrip(t, b.g, false)
	item := restored.StackMeta[b.kindle]
	if item == nil || item.AltCost != effectstest.KindleFlashback || len(item.Paid.AltCostObjects) != 3 {
		t.Fatalf("restored stack item = %+v, want the flashback claim and its three objects", item)
	}
	effectstest.PassPriorityAroundTable(t, restored)
	if got := effectstest.B43TokensNamed(restored, b.me.ID, "Smoke Elemental"); got != 1 {
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
	restored := effectstest.RestoreRoundTrip(t, b.g, false)
	if err := effectstest.CastForAltCost(restored, b.kindle, "graveyard", effectstest.KindleFlashback, []uuid.UUID{b.smoke, b.ember, b.inHand},
		[]game.TargetRef{{Kind: game.TargetCard, ID: b.smoke}}); err != nil {
		t.Fatalf("flashback on the restored table: %v", err)
	}
}

// --- Register's guards --------------------------------------------------

func TestRegisterRefusesMalformedAltCostBehold(t *testing.T) {
	good := effects.FlashbackBehold("{1}{R}", 3, "Elemental", "Elementals")
	if good.Label != "Flashback—{1}{R}, Behold three Elementals" || good.PayLabel != "three Elementals" || !good.ExileOnLeavingStack || good.FromZone != game.ZoneGraveyard {
		t.Fatalf("FlashbackBehold = %+v", good)
	}
	effects.CheckAltCostBehold("Kindle", good)
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
			effectstest.MustPanic(t, "build it with FlashbackBehold", func() { effects.CheckAltCostBehold("Bad", bad) })
		})
	}
	two := good
	two.ExileFromHand = effects.CardInYourHand("a red card", effects.OfColor("R"))
	if n := effects.AltCostCardComponents(two); n != 2 {
		t.Errorf("a behold beside a pitch counts %d card components, want 2", n)
	}
}
