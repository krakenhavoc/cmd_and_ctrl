package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b543_cards_test.go — Lorwyn clash cards, S82 OLD batch 543 (#5593).
// Every test plays the real catalog entry through the engine. A clash is
// rigged by what is on top of each library: the greater mana value wins.

const (
	oldB543Woodland  = "09e84398-7782-48d0-8411-9055d253395c"
	oldB543Captivate = "11a372c7-3c7c-431c-9532-f2bae93f1c33"
	oldB543SentryOak = "1b2e0a60-844e-4db7-91fd-323497ab3cbf"
	oldB543NathsElit = "2445fb97-a9c9-4078-8a8e-64b8c3951c4a"
	oldB543Broken    = "4a3fba0c-50b1-4139-834e-c8fc8e1315bb"
	oldB543Fistful   = "81641cf1-2dcf-4cd3-ab92-7f5e8e20cb84"
	oldB543Brawler   = "819351c6-b0d6-42dc-9bd9-d1a3851aad7c"
	oldB543Ringskip  = "8e353b5f-0692-430e-9558-59818275b13b"
)

func TestOldB543EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB543Woodland: "Woodland Guidance", oldB543Captivate: "Captivating Glance", oldB543SentryOak: "Sentry Oak",
		oldB543NathsElit: "Nath's Elite", oldB543Broken: "Broken Ambitions", oldB543Fistful: "Fistful of Force",
		oldB543Brawler: "Oaken Brawler", oldB543Ringskip: "Ringskipper",
	} {
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

// oldB543Top puts a card of the given mana cost on top of p's library.
func oldB543Top(p *game.Player, name, cost string) uuid.UUID {
	c := game.NewCard(name, p.ID)
	c.TypeLine = "Sorcery"
	c.ManaCost = cost
	p.Library.PushTop(c)
	return c.InstanceID
}

// oldB543Rig makes `me` win the clash (win true), lose it, or tie.
func oldB543Rig(me, opp *game.Player, outcome string) {
	switch outcome {
	case "win":
		oldB543Top(me, "My Five", "{3}{R}{R}")
		oldB543Top(opp, "Their One", "{U}")
	case "lose":
		oldB543Top(me, "My One", "{R}")
		oldB543Top(opp, "Their Five", "{3}{U}{U}")
	default:
		oldB543Top(me, "My Two", "{1}{R}")
		oldB543Top(opp, "Their Two", "{1}{U}")
	}
}

// oldB543Clash answers a clash with `opponent`: the controller picks them
// and both clashers keep their card on top.
func oldB543Clash(t *testing.T, g *game.Game, opponent uuid.UUID) {
	t.Helper()
	answered := false
	for i := 0; i < 8; i++ {
		var open *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceOptionPick {
				open = c
				break
			}
		}
		if open == nil {
			break
		}
		pick := 0
		if !answered {
			for j, o := range open.PickOptions {
				if o.Player == opponent {
					pick = j
				}
			}
			answered = true
		}
		if err := g.ResolveOptionPick(open.ID, open.Chooser, pick); err != nil {
			t.Fatalf("answer the clash question: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func oldB543Counters(g *game.Game, id uuid.UUID) int {
	return effectstest.Pr8Counters(g, id, game.CounterPlusOne)
}

// oldB543Enter casts the creature and settles its enters trigger with the
// outcome rigged, returning the permanent.
func oldB543Enter(t *testing.T, g *game.Game, name, typeLine, oracle, outcome string) uuid.UUID {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	oldB543Rig(me, opp, outcome)
	id := effectstest.CastCatalogSpell(t, g, name, typeLine, oracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	oldB543Clash(t, g, opp.ID)
	return id
}

func TestOldB543OakenBrawlerGrowsOnlyWhenItWinsTheClash(t *testing.T) {
	for outcome, want := range map[string]int{"win": 1, "lose": 0, "tie": 0} {
		g := effectstest.NewCatalogGame(t)
		id := oldB543Enter(t, g, "Oaken Brawler", "Creature — Treefolk Warrior", oldB543Brawler, outcome)
		if got := oldB543Counters(g, id); got != want {
			t.Errorf("after a %s the Brawler has %d +1/+1 counters, want %d", outcome, got, want)
		}
	}
}

func TestOldB543NathsEliteGrowsOnlyWhenItWinsAndIsALure(t *testing.T) {
	spec, ok := effects.Lookup(oldB543NathsElit)
	if !ok || len(spec.Static) != 1 {
		t.Fatalf("registered=%v with %d statics, want the one all-able-to-block static", ok, len(spec.Static))
	}
	for outcome, want := range map[string]int{"win": 1, "lose": 0} {
		g := effectstest.NewCatalogGame(t)
		id := oldB543Enter(t, g, "Nath's Elite", "Creature — Elf Warrior", oldB543NathsElit, outcome)
		if got := oldB543Counters(g, id); got != want {
			t.Errorf("after a %s the Elite has %d +1/+1 counters, want %d", outcome, got, want)
		}
	}
}

func oldB543Ringskipper(t *testing.T, outcome string) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	skipper := effectstest.PushCatalogPermanent(g, me.ID, "Ringskipper", "Creature — Faerie Wizard", oldB543Ringskip, false)
	oldB543Rig(me, opp, outcome)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(skipper); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	oldB543Clash(t, g, opp.ID)
	return g, me, skipper
}

func TestOldB543RingskipperReturnsToHandOnAWin(t *testing.T) {
	_, me, skipper := oldB543Ringskipper(t, "win")
	inHand := false
	for _, c := range me.Hand.Cards {
		if c.InstanceID == skipper {
			inHand = true
		}
	}
	if !inHand {
		t.Fatal("a won clash did not return Ringskipper to its owner's hand")
	}
}

func TestOldB543RingskipperStaysInTheGraveyardOnALoss(t *testing.T) {
	_, me, skipper := oldB543Ringskipper(t, "lose")
	for _, c := range me.Graveyard.Cards {
		if c.InstanceID == skipper {
			return
		}
	}
	t.Fatal("a lost clash moved Ringskipper out of the graveyard")
}

func TestOldB543SentryOakAttacksOnlyAfterWinning(t *testing.T) {
	for outcome, won := range map[string]bool{"win": true, "lose": false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		oak := effectstest.B12Push(g, me.ID, "Sentry Oak", "Creature — Treefolk Warrior", oldB543SentryOak, 3, 5)
		effectstest.AdvanceToStepInTurn(t, g, game.StepBeginCombat)
		oldB543Rig(me, opp, outcome)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
		effectstest.PassPriorityAroundTable(t, g)
		oldB543Clash(t, g, opp.ID)
		power, _ := effectstest.EffectivePTOf(t, g, oak)
		wantPower := 3
		if won {
			wantPower = 5
		}
		if power != wantPower {
			t.Errorf("after a %s Sentry Oak has power %d, want %d", outcome, power, wantPower)
		}
		if got := effectstest.P1g2HasKeyword(g, oak, "defender"); got == won {
			t.Errorf("after a %s Sentry Oak has defender = %v", outcome, got)
		}
	}
}

func TestOldB543FistfulOfForceAddsAnotherTwoAndTrampleOnAWin(t *testing.T) {
	for outcome, want := range map[string]int{"win": 8, "lose": 6} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 4, 4)
		oldB543Rig(me, opp, outcome)
		effectstest.CastCatalogSpell(t, g, "Fistful of Force", "Instant", oldB543Fistful, effectstest.CardRefs(bear))
		effectstest.PassPriorityAroundTable(t, g)
		oldB543Clash(t, g, opp.ID)
		if p, tough := effectstest.EffectivePTOf(t, g, bear); p != want || tough != want {
			t.Errorf("after a %s the Bear is %d/%d, want %d/%d", outcome, p, tough, want, want)
		}
		if got := effectstest.P1g2HasKeyword(g, bear, "trample"); got != (outcome == "win") {
			t.Errorf("after a %s the Bear's trample = %v", outcome, got)
		}
	}
}

func TestOldB543WoodlandGuidanceReturnsUntapsAndExilesItself(t *testing.T) {
	for outcome, untapped := range map[string]bool{"win": true, "lose": false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		forest := effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
		g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, forest).Tapped = true })
		back := effectstest.PushGraveyardCardForTest(me, "Old Card")
		oldB543Rig(me, opp, outcome)
		spell := effectstest.CastCatalogSpell(t, g, "Woodland Guidance", "Sorcery", oldB543Woodland, effectstest.CardRefs(back))
		effectstest.PassPriorityAroundTable(t, g)
		oldB543Clash(t, g, opp.ID)
		inHand := false
		for _, c := range me.Hand.Cards {
			if c.InstanceID == back {
				inHand = true
			}
		}
		if !inHand {
			t.Errorf("after a %s the targeted card did not return to hand", outcome)
		}
		if got := !effectstest.HmlTapped(t, g, forest); got != untapped {
			t.Errorf("after a %s the Forest untapped = %v, want %v", outcome, got, untapped)
		}
		for _, c := range me.Graveyard.Cards {
			if c.InstanceID == spell {
				t.Errorf("after a %s Woodland Guidance is in the graveyard, want it exiled", outcome)
			}
		}
	}
}

func TestOldB543CaptivatingGlanceGivesTheCreatureToTheClashWinner(t *testing.T) {
	for outcome, wantMe := range map[string]bool{"win": true, "lose": false} {
		g := effectstest.NewCatalogGame(t)
		me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
		effectstest.AdvanceToMain(t, g)
		host := effectstest.B12Creature(g, third.ID, "Host", "Creature — Bear", 2, 2)
		effectstest.Enchant(t, g, "Captivating Glance", oldB543Captivate, host)
		effectstest.AdvanceToEndStep(t, g)
		oldB543Rig(me, opp, outcome)
		effectstest.PassPriorityAroundTable(t, g)
		oldB543Clash(t, g, opp.ID)
		want := opp.ID
		if wantMe {
			want = me.ID
		}
		c := effectstest.LayeredCard(t, g, host)
		if c.Controller != want {
			t.Errorf("after a %s the creature is controlled by %s, want %s", outcome, c.Controller, want)
		}
	}
}

func TestOldB543BrokenAmbitionsCountersUnlessPaidAndMillsOnAWin(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	theirs := effectstest.CastFromHandForTest(t, g, opp, "Their Spell", "Instant", "", "", game.CastSpellParams{})
	oldB543Rig(me, opp, "win")
	effectstest.ApaMana(me, "U", "C", "C")
	before := opp.Library.Size()
	effectstest.CastFromHandForTest(t, g, me, "Broken Ambitions", "Instant", "{X}{U}", oldB543Broken, game.CastSpellParams{
		Targets: effectstest.CardTarget(theirs), XValue: 2,
	})
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 8; i++ {
		var next *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoicePayUnless {
				next = c
			}
		}
		if next == nil {
			break
		}
		if err := g.ResolvePayUnless(next.ID, next.Chooser, false); err != nil {
			t.Fatalf("decline to pay: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	oldB543Clash(t, g, opp.ID)
	if !effectstest.P7Countered(g, theirs) {
		t.Error("the spell was not countered when its controller declined to pay")
	}
	if got := before - opp.Library.Size(); got < 4 {
		t.Errorf("the spell's controller milled %d cards, want 4 after a won clash", got)
	}
}
