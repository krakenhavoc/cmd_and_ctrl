package testsds

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestEveryEchoCardChargesItsPrintedCost(t *testing.T) {
	for name, want := range map[string]string{
		"Acridian": "{1}{G}", "Albino Troll": "{1}{G}", "Avalanche Riders": "{3}{R}",
		"Basalt Gargoyle": "{2}{R}", "Bone Shredder": "{2}{B}", "Citanul Centaurs": "{3}{G}",
		"Cradle Guard": "{1}{G}{G}", "Crater Hellion": "{4}{R}{R}", "Deepcavern Imp": "Discard a card",
		"Deranged Hermit": "{3}{G}{G}", "Extruder": "{4}", "Firemaw Kavu": "{5}{R}",
		"Flamecore Elemental": "{2}{R}{R}", "Ghitu Slinger": "{2}{R}", "Goblin Marshal": "{4}{R}{R}",
		"Goblin Patrol": "{R}", "Goblin War Buggy": "{1}{R}", "Hammerheim Deadeye": "{5}{R}",
		"Henchfiend of Ukor": "{1}{B}", "Herald of Serra": "{2}{W}{W}", "Hunting Moa": "{2}{G}",
		"Karmic Guide": "{3}{W}{W}", "Keldon Champion": "{2}{R}{R}", "Keldon Vandals": "{2}{R}",
		"Lightning Dragon": "{2}{R}{R}", "Mogg War Marshal": "{1}{R}", "Multani's Acolyte": "{G}{G}",
		"Orcish Hellraiser": "{R}", "Pouncing Jaguar": "{G}", "Radiant's Dragoons": "{3}{W}",
		"Rakdos Headliner": "Discard a card", "Raven Familiar": "{2}{U}", "Ring of Gix": "{3}",
		"Shah of Naar Isle": "{0}", "Shivan Raptor": "{2}{R}", "Simian Grunts": "{2}{G}",
		"Skizzik Surger": "Sacrifice two lands", "Stingscourger": "{3}{R}",
		"Subterranean Shambler": "{3}{R}", "Tectonic Fiend": "{4}{R}{R}", "Thran War Machine": "{4}",
		"Thran Weaponry": "{4}", "Ticking Gnomes": "{3}", "Timbermare": "{5}{G}",
		"Uktabi Drake": "{1}{G}{G}", "Urza's Blueprints": "{6}", "Viashino Outrider": "{2}{R}",
		"Vug Lizard": "{1}{R}{R}", "Winding Wurm": "{4}{G}", "Yavimaya Granger": "{2}{G}",
	} {
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[1]
			if effectstest.OracleByName(name) == "" {
				t.Fatalf("%s is not in the catalog", name)
			}
			id := effectstest.SeatEchoCard(g, me.ID, name, "Creature — Test")
			effectstest.AdvanceToUpkeepOf(t, g, 1)
			if effectstest.TriggerOnStack(g, id) == nil {
				t.Fatal("echo did not trigger")
			}
			effectstest.PassPriorityAroundTable(t, g)
			p := effectstest.EchoPrompt(g, me.ID)
			if p == nil {
				t.Fatal("no echo prompt")
			}
			if p.PayCost != want {
				t.Fatalf("echo cost = %q, want %q", p.PayCost, want)
			}
			if err := g.ResolvePayUnless(p.ID, me.ID, false); err != nil {
				t.Fatalf("decline: %v", err)
			}
			if _, ok := effectstest.BattlefieldCard(g, id); ok {
				t.Fatal("still on the battlefield after declining its echo")
			}
		})
	}
}

// Shah of Naar Isle: paying its echo {0} lets each opponent draw up to
// three; each chooses how many.
func TestShahOfNaarIsleOpponentsChooseHowManyToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	shah := effectstest.SeatEchoCard(g, me.ID, "Shah of Naar Isle", "Creature — Efreet")
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ResolvePayUnless(effectstest.EchoPrompt(g, me.ID).ID, me.ID, true); err != nil {
		t.Fatalf("pay echo {0}: %v", err)
	}
	if effectstest.TriggerOnStack(g, shah) == nil {
		t.Fatal("paying the echo did not trigger Shah of Naar Isle")
	}
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Hand.Size()
	}
	effectstest.PassPriorityAroundTable(t, g)
	// Turn order after seat 1: seats 2, 3, 0. Each picks 3, 0, 1.
	for i, seat := range []int{2, 3, 0} {
		opp := g.Seats[seat]
		var c *game.PendingChoice
		for _, pc := range g.PendingChoices {
			if pc != nil && pc.Chooser == opp.ID && pc.Kind == game.PendingChoiceOptionPick {
				c = pc
			}
		}
		if c == nil {
			t.Fatalf("seat %d was not asked how many cards to draw", seat)
		}
		if err := g.ResolveOptionPick(c.ID, opp.ID, []int{3, 0, 1}[i]); err != nil {
			t.Fatalf("seat %d answers: %v", seat, err)
		}
	}
	for seat, want := range map[int]int{2: 3, 3: 0, 0: 1, 1: 0} {
		p := g.Seats[seat]
		if got := p.Hand.Size() - before[p.ID]; got != want {
			t.Errorf("seat %d drew %d, want %d", seat, got, want)
		}
	}
}

// Vexing Sphinx draws a card for each age counter it had as it died.
func TestVexingSphinxDrawsForEachAgeCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	sphinx := effectstest.SeatEchoCard(g, me.ID, "Vexing Sphinx", "Creature — Sphinx")
	effectstest.PushHandCard(g, me)
	effectstest.PushHandCard(g, me)
	effectstest.PushHandCard(g, me)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.EchoPrompt(g, me.ID)
	if p == nil || p.PayCost != "Discard a card" {
		t.Fatalf("first upkeep prompt = %+v", p)
	}
	if err := g.ResolvePayUnlessWithCards(p.ID, me.ID, true, []uuid.UUID{me.Hand.Cards[0].InstanceID}); err != nil {
		t.Fatalf("discard one: %v", err)
	}
	effectstest.LeaveUpkeepThenAdvanceTo(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	p = effectstest.EchoPrompt(g, me.ID)
	if p == nil || p.PayCost != "Discard two cards" {
		t.Fatalf("second upkeep prompt = %+v, want discard two cards", p)
	}
	hand := me.Hand.Size()
	if err := g.ResolvePayUnless(p.ID, me.ID, false); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if _, ok := effectstest.BattlefieldCard(g, sphinx); ok {
		t.Fatal("the Sphinx survived its declined upkeep")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Fatalf("drew %d, want 2 (two age counters)", got)
	}
}

// Thran Weaponry pumps every creature while it stays tapped.
func TestThranWeaponryPumpsWhileTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	weaponry := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Thran Weaponry",
		OracleID: effectstest.OracleByName("Thran Weaponry"), TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: g.Seats[1].ID, Controller: g.Seats[1].ID})
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, weaponry, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.BattlefieldCardCopy(t, g, bear)
	if c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Fatalf("bear is %d/%d while Weaponry is tapped, want 4/4", c.CurrentPower(), c.CurrentToughness())
	}
	if err := g.UntapTargetForEffect(weaponry); err != nil {
		t.Fatal(err)
	}
	g.BumpLayerVersionForTest()
	c = effectstest.BattlefieldCardCopy(t, g, bear)
	if c.CurrentPower() != 2 {
		t.Fatalf("bear is %d/%d after Weaponry untapped, want 2/2", c.CurrentPower(), c.CurrentToughness())
	}
}

// Deranged Hermit's four Squirrels are 2/2 under its anthem.
func TestDerangedHermitMakesFourBigSquirrels(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.CastEchoCreature(t, g, "Deranged Hermit", "Creature — Elf", effectstest.OracleByName("Deranged Hermit"))
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Squirrel" {
			n++
			cc := effectstest.BattlefieldCardCopy(t, g, c.InstanceID)
			if cc.CurrentPower() != 2 || cc.CurrentToughness() != 2 {
				t.Errorf("Squirrel is %d/%d, want 2/2", cc.CurrentPower(), cc.CurrentToughness())
			}
		}
	}
	if n != 4 {
		t.Fatalf("%d Squirrels, want 4", n)
	}
}
