package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// incubate_caveats_d_test.go — slice inc-d of #3568: the three
// incubate caveats cleared (Assimilate Essence's paid branch is tested
// beside the other Assimilate tests, incubate_cards_a_test.go).

const towashiFixtureOracle = "7a8f3c52-3b8e-4d29-9a43-2f3b6c0f1d11"

// towashiSeat puts Corruption of Towashi on the battlefield WITHOUT its
// enters trigger, so no Incubator muddies the board.
func towashiSeat(g *game.Game, p *game.Player) {
	pushCatalogPermanent(g, p.ID, "Corruption of Towashi", "Enchantment", corruptionOfTowashiOracle, false)
}

// graveyardDoubleFaced puts a real imported transform card into the
// seat's graveyard, front face up (CR 712.8a).
func graveyardDoubleFaced(p *game.Player) uuid.UUID {
	c := deck.ToGameCard(transformRow(towashiFixtureOracle,
		"Fixture Villager", "Creature — Human Werewolf", "{2}{G}",
		"Fixture Werewolf", "Creature — Werewolf", "3", "3", []string{"G"}), false)
	c.Owner, c.Controller = p.ID, p.ID
	p.Graveyard.PushTop(c)
	return c.InstanceID
}

// towashiAnswer settles the stack, orders any triggers that were put on
// it together, and answers the draw question.
func towashiAnswer(t *testing.T, g *game.Game, p *game.Player, yes bool) {
	t.Helper()
	for i := 0; i < 8 && latestConfirmFor(g, p.ID) == nil; i++ {
		passPriorityAroundTable(t, g)
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceTriggerOrder {
				if err := g.ResolveTriggerOrder(c.ID, c.Chooser, append([]uuid.UUID(nil), c.TriggerOrderIDs...)); err != nil {
					t.Fatalf("ResolveTriggerOrder: %v", err)
				}
				break
			}
		}
	}
	answerMayChoice(t, g, p.ID, yes)
	passPriorityAroundTable(t, g)
}

func TestCorruptionOfTowashiDrawsWhenAPermanentEntersTransformed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	towashiSeat(g, me)
	id := graveyardDoubleFaced(me)
	handBefore := me.Hand.Size()

	g.WithWriteLock(func() {
		got, err := g.ReturnFromGraveyardTransformedForEffect(id, uuid.Nil, false, nil)
		if err != nil || got == uuid.Nil {
			t.Fatalf("return transformed: %v (%v)", got, err)
		}
	})
	towashiAnswer(t, g, me, true)

	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Errorf("hand grew by %d, want 1: a permanent entering transformed draws", got)
	}
}

func TestCorruptionOfTowashiDrawsForADayboundPermanentEnteringAtNight(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	towashiSeat(g, me)
	g.WithWriteLock(func() { g.DayNight.Designation = game.DesignationNight })
	handBefore := me.Hand.Size()

	importAndCast(t, g, infestationExpertRow(), me)
	towashiAnswer(t, g, me, true)

	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Errorf("hand grew by %d, want 1: a daybound permanent entering at night enters transformed (CR 702.145b)", got)
	}
}

func TestCorruptionOfTowashiIgnoresAPermanentEnteringFrontFaceUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	towashiSeat(g, me)
	importToBattlefield(t, g, transformRow(towashiFixtureOracle,
		"Fixture Villager", "Creature — Human Werewolf", "{2}{G}",
		"Fixture Werewolf", "Creature — Werewolf", "3", "3", []string{"G"}), me)
	passPriorityAroundTable(t, g)
	if latestConfirmFor(g, me.ID) != nil {
		t.Error("a double-faced permanent entering on its front face asked about a draw")
	}
}

func TestCorruptionOfTowashiIgnoresAnOpponentsPermanentEnteringTransformed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	towashiSeat(g, me)
	id := graveyardDoubleFaced(opp)
	g.WithWriteLock(func() {
		if _, err := g.ReturnFromGraveyardTransformedForEffect(id, uuid.Nil, false, nil); err != nil {
			t.Fatalf("return transformed: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if latestConfirmFor(g, me.ID) != nil || latestConfirmFor(g, opp.ID) != nil {
		t.Error("\"a permanent you control\" triggered on an opponent's permanent")
	}
}

// A convert is not a transform (CR 701.28): the Incubator converting
// draws nothing, and transforming it does.
func TestCorruptionOfTowashiIgnoresAConvertAndCountsATransform(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	towashiSeat(g, me)
	g.WithWriteLock(func() { _, _ = g.CreateTokensForEffect(me.ID, IncubatorToken(), 1, game.TokenEntryOptions{}) })
	inc := theIncubator(t, g, me.ID)

	g.WithWriteLock(func() {
		if err := g.ConvertPermanentForEffect(inc.InstanceID); err != nil {
			t.Fatalf("convert: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if latestConfirmFor(g, me.ID) != nil {
		t.Fatal("a convert asked about a draw; it is not a transform")
	}
	g.WithWriteLock(func() {
		if err := g.TransformPermanentForEffect(inc.InstanceID); err != nil {
			t.Fatalf("transform: %v", err)
		}
	})
	handBefore := me.Hand.Size()
	towashiAnswer(t, g, me, true)
	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Errorf("hand grew by %d on a real transform, want 1", got)
	}
}

// "Do this only once each turn" limits the draw, not the trigger: a
// declined "may" leaves the turn's one draw, taking it uses it up, and a
// second trigger stacked BEFORE the first resolved does nothing.
func TestCorruptionOfTowashiDrawsOnlyOncePerTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	towashiSeat(g, me)
	turnOver := func() {
		id := uuid.New()
		pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: id, Name: "Test DFC", TypeLine: "Creature — Human Werewolf", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
		g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventTransform, CardID: id, Actor: me.ID}) })
	}
	handBefore := me.Hand.Size()

	turnOver()
	towashiAnswer(t, g, me, false)
	if got := me.Hand.Size() - handBefore; got != 0 {
		t.Fatalf("declining drew %d cards", got)
	}
	turnOver()
	towashiAnswer(t, g, me, true)
	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Fatalf("hand grew by %d after the second transform, want 1: the declined first one did not use the turn", got)
	}
	turnOver()
	passPriorityAroundTable(t, g)
	if latestConfirmFor(g, me.ID) != nil {
		t.Error("a third transform asked again; the turn's draw is spent")
	}

	// Two triggers stacked before either resolves: the second does nothing.
	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	towashiSeat(g2, me2)
	a, b := uuid.New(), uuid.New()
	for _, id := range []uuid.UUID{a, b} {
		pushBattlefieldCardWithTimestamp(g2, game.Card{InstanceID: id, Name: "Test DFC", TypeLine: "Creature — Human Werewolf", Power: 1, Toughness: 1, Owner: me2.ID, Controller: me2.ID})
	}
	before := me2.Hand.Size()
	g2.WithWriteLock(func() {
		g2.EmitEvent(game.Event{Kind: game.EventTransform, CardID: a, Actor: me2.ID})
		g2.EmitEvent(game.Event{Kind: game.EventTransform, CardID: b, Actor: me2.ID})
	})
	for i := 0; i < 4; i++ {
		passPriorityAroundTable(t, g2)
		if c := latestConfirmFor(g2, me2.ID); c != nil {
			answerMayChoice(t, g2, me2.ID, true)
		}
	}
	if got := me2.Hand.Size() - before; got != 1 {
		t.Errorf("two stacked triggers drew %d cards, want 1", got)
	}
}

// The face test, on its own: which entries count as "transformed".
func TestCardEnteredTransformedFollowsTheEntryFace(t *testing.T) {
	front := func() game.Card {
		return game.Card{Layout: game.LayoutTransform, ActiveFace: 0, Faces: []game.Face{{}, {}}}
	}
	back := front()
	back.ActiveFace = 1
	if cardEnteredTransformed(front()) {
		t.Error("a front-face entry is not transformed")
	}
	if !cardEnteredTransformed(back) {
		t.Error("a transform card entering on its back face entered transformed")
	}
	mdfc := back
	mdfc.Layout = game.LayoutModalDFC
	if cardEnteredTransformed(mdfc) {
		t.Error("an MDFC cast as its back face was cast, not transformed")
	}
	convert := back
	convert.Faces = []game.Face{{Keywords: []string{"more than meets the eye"}}, {}}
	if cardEnteredTransformed(convert) {
		t.Error("a More Than Meets the Eye cast enters converted, not transformed")
	}
}

// Brimaz's second ability, on the turn a Phyrexian died under its
// controller: the end step proliferates.
func TestBrimazBlightProliferatesAtTheEndStepAfterAPhyrexianDiedUnderYou(t *testing.T) {
	cases := []struct {
		name      string
		diesUnder int // seat index the creature dies under
		typeLine  string
		want      int // +1/+1 counters on my creature at the end
	}{
		{"my Phyrexian died", 0, "Creature — Phyrexian Horror", 2},
		{"my Bear died", 0, "Creature — Bear", 1},
		{"an opponent's Phyrexian died", 1, "Creature — Phyrexian Horror", 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			pushIncubateACreature(g, me.ID, "Brimaz, Blight of Oreskos", "Legendary Creature — Phyrexian Cat", brimazBlightOracle, 3, 4)
			grown := pushCounterCreature(g, me.ID, "Grown", game.CounterPlusOne, 1)
			victim := pushIncubateACreature(g, g.Seats[tc.diesUnder].ID, "Victim", tc.typeLine, "", 1, 1)
			advanceToMain(t, g)
			g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victim) })
			advanceToEndStep(t, g)
			passPriorityAroundTable(t, g)
			if p := proliferatePrompt(g); p != nil {
				answerProliferate(t, g, grown)
				passPriorityAroundTable(t, g)
			}
			if got := counterCount(g, grown, game.CounterPlusOne); got != tc.want {
				t.Errorf("+1/+1 counters = %d, want %d", got, tc.want)
			}
		})
	}
}

// The intervening if is checked again on resolution (CR 603.4): a
// trigger that loses its Phyrexian does nothing. The tally cannot lose
// one, so this proves the first check: nobody died, no prompt, no
// stack item.
func TestBrimazBlightDoesNotTriggerWhenNothingDied(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushIncubateACreature(g, me.ID, "Brimaz, Blight of Oreskos", "Legendary Creature — Phyrexian Cat", brimazBlightOracle, 3, 4)
	pushCounterCreature(g, me.ID, "Grown", game.CounterPlusOne, 1)
	advanceToEndStep(t, g)
	if !stackFullyEmpty(g) {
		t.Error("the end-step trigger went on the stack with no Phyrexian having died")
	}
}
