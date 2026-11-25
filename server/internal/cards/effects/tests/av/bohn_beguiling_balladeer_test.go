package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const bohnOracle = "ab31b652-ddf2-480f-a955-e8b5c87728f9"

func bohnHandCard(me *game.Player, name, typeLine, oracle, mana string) uuid.UUID {
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: mana,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	return id
}

func bohnForetells(g *game.Game, me *game.Player, id uuid.UUID) []string {
	var out []string
	g.WithWriteLock(func() {
		card, ok := g.LookupCardForEffect(id)
		if !ok {
			return
		}
		for _, sa := range g.SpecialActionsOfferedLocked(me.ID, card, game.ZoneHand) {
			if sa.Kind == game.SpecialActionForetell {
				out = append(out, sa.Cost+"|"+sa.CastCost)
			}
		}
	})
	return out
}

// #6153 acceptance (CR 702.143a): with Bohn, a {4}{G} creature in hand is
// foretold for {2} and cast on a later turn for {2}{G}; a card that prints
// foretell keeps its own cost; without Bohn nothing is offered.
func TestBohnGivesForetellAtManaCostLessTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	beast := bohnHandCard(me, "Beast", "Creature — Beast", "", "{4}{G}")
	saw := bohnHandCard(me, "Saw It Coming", "Instant", effectstest.SawItComingOracle, "{1}{U}{U}")
	if got := bohnForetells(g, me, beast); len(got) != 0 {
		t.Fatalf("the Beast had foretell with no Bohn: %v", got)
	}
	effectstest.B12Push(g, me.ID, "Bohn, Beguiling Balladeer", "Legendary Creature — Human Bard", bohnOracle, 3, 3)
	if got := bohnForetells(g, me, beast); len(got) != 1 || got[0] != "{2}|{2}{G}" {
		t.Fatalf("the Beast offers %v, want foretell {2} with the foretell cost {2}{G}", got)
	}
	if got := bohnForetells(g, me, saw); len(got) != 1 || got[0] != "{2}|{1}{U}" {
		t.Errorf("Saw It Coming offers %v, want only its own foretell cost {1}{U}", got)
	}

	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.PerformSpecialAction(me.ID, beast, game.SpecialActionForetell, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("foretell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(beast) {
		t.Fatal("the foretold card is not in exile")
	}
	g.WithWriteLock(func() { g.Turn.Seq++ })
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "G"})
	if err := g.CastSpell(me.ID, beast, game.CastSpellParams{Strict: true, FromZone: "exile", AlternativeCost: game.AltCostKeyForetell}); err == nil {
		t.Fatal("cast the foretold Beast for {1}{G}")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.CastSpell(me.ID, beast, game.CastSpellParams{Strict: true, FromZone: "exile", AlternativeCost: game.AltCostKeyForetell}); err != nil {
		t.Fatalf("cast the foretold Beast for {2}{G}: %v", err)
	}
}

// The second spell I cast in a turn goads a creature an opponent controls;
// the first does not trigger, and my own creature is not a legal target.
func TestBohnGoadsOnTheSecondSpellEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Bohn, Beguiling Balladeer", "Legendary Creature — Human Bard", bohnOracle, 3, 3)
	theirs := effectstest.PushCreatureToBattlefieldForTest(g, opp.ID, "Their Bear")
	mine := effectstest.PushCreatureToBattlefieldForTest(g, me.ID, "My Bear")

	first := bohnHandCard(me, "Spell One", "Sorcery", "", "")
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{}); err != nil {
		t.Fatalf("first cast: %v", err)
	}
	if len(g.PendingTriggers) != 0 {
		t.Fatal("the first spell triggered Bohn")
	}
	effectstest.PassPriorityAroundTable(t, g)

	second := bohnHandCard(me, "Spell Two", "Sorcery", "", "")
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{}); err != nil {
		t.Fatalf("second cast: %v", err)
	}
	var walk *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget && c.Chooser == me.ID {
			walk = c
		}
	}
	if walk == nil {
		t.Fatal("the second spell did not ask for Bohn's target")
	}
	for _, id := range walk.PickTargetCards {
		if id == mine {
			t.Fatal("my own creature is offered as the goad target")
		}
	}
	if err := g.ResolvePickTarget(walk.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err != nil {
		t.Fatalf("target their creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	card, ok := effectstest.BattlefieldCard(g, theirs)
	if !ok || len(card.Goads) == 0 {
		t.Fatalf("their creature is not goaded: %+v", card.Goads)
	}
}
