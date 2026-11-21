package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b118_test.go — S82 set GAP, batch 118 (Tales of Middle-earth
// Commander creatures), through the real catalog.

const (
	gap118Eomer      = "4f8191e6-a6aa-4426-9014-eff4e0efee9e"
	gap118Archivist  = "a086f7a5-1319-4c5d-b23f-41bec2f81027"
	gap118Monstrous  = "c5f52c7f-5dba-4d4c-8acc-068732964031"
	gap118Boromir    = "5bb5e4e5-c3b4-4f67-83e1-bde9ce9f18c4"
	gap118Faramir    = "466af6da-52e8-4ac5-b659-7492e3594ff8"
	gap118Beregond   = "8ea50101-8392-471e-8523-b12045a7b95b"
	gap118Champions  = "464a930a-79d5-4436-864d-3ce9f29475ae"
	gap118Medic      = "36a3c910-95b8-45a2-9f84-5681336d868c"
	gap118Gilraen    = "27e83bae-0ffe-4cc7-a6c0-cbca25f2ba4e"
	gap118GreyHost   = "053dff47-25df-424b-8328-3ee00739b4fd"
	gap118Gwaihir    = "77572e70-4b8b-46ae-8c8d-f71b7dd456e0"
	gap118Stonehewer = "b3d99ea0-d61f-4394-8913-b4057c38ff3e"
)

func gap118Counters(g *game.Game, id uuid.UUID, kind string) int {
	n := -1
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				n = c.Counters[kind]
			}
		}
	})
	return n
}

func gap118Named(g *game.Game, owner uuid.UUID, name string) []uuid.UUID {
	var out []uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == owner && c.Name == name {
				out = append(out, c.InstanceID)
			}
		}
	})
	return out
}

func gap118Power(g *game.Game, id uuid.UUID) int {
	p := -99
	g.ReadSnapshot(func() {
		if c, ok := g.LookupCardForEffect(id); ok {
			p = c.PowerForComparison()
		}
	})
	return p
}

func gap118Confirm(t *testing.T, g *game.Game, who uuid.UUID, accept bool) {
	t.Helper()
	for i := 0; i < 16; i++ {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceConfirm && c.Chooser == who {
				if err := g.ResolveConfirm(c.ID, who, accept); err != nil {
					t.Fatalf("ResolveConfirm: %v", err)
				}
				return
			}
		}
		if effectstest.StackFullyEmpty(g) {
			break
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("no confirm prompt")
}

func TestGap118Registered(t *testing.T) {
	for oracle, name := range map[string]string{
		gap118Eomer: "Éomer, King of Rohan", gap118Archivist: "Archivist of Gondor",
		gap118Monstrous: "Monstrosity of the Lake", gap118Boromir: "Boromir, Gondor's Hope",
		gap118Faramir: "Faramir, Steward of Gondor", gap118Beregond: "Beregond of the Guard",
		gap118Champions: "Champions of Minas Tirith", gap118Medic: "Frontline Medic",
		gap118Gilraen: "Gilraen, Dúnedain Protector", gap118GreyHost: "Grey Host Reinforcements",
		gap118Gwaihir: "Gwaihir, Greatest of the Eagles", gap118Stonehewer: "Stonehewer Giant",
	} {
		effectstest.StdRegisteredFull(t, oracle, name)
	}
}

func TestGap118EomerEntersWithCountersCrownsAndDeals(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.BlbB1Push(g, me.ID, "Human One", "Creature — Human Soldier", "", 1, 1)
	effectstest.BlbB1Push(g, me.ID, "Human Two", "Creature — Human Soldier", "", 1, 1)
	effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	before := effectstest.LifeOf(g, opp.ID)
	if _, err := effectstest.BlbHCastWithMana(t, g, "Éomer, King of Rohan", "Legendary Creature — Human Noble", gap118Eomer, "{3}{R}{W}", 2, 2, "R", "W", "G", "G", "G"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}, true)
	ids := gap118Named(g, me.ID, "Éomer, King of Rohan")
	if len(ids) != 1 {
		t.Fatalf("Éomer is not on the battlefield")
	}
	if got := gap118Counters(g, ids[0], game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters, want 2 (two other Humans)", got)
	}
	if !effectstest.P1g2HasKeyword(g, ids[0], "double strike") {
		t.Error("lacks double strike")
	}
	if g.Monarch != opp.ID {
		t.Errorf("monarch = %v, want the targeted player", g.Monarch)
	}
	if got := effectstest.LifeOf(g, opp.ID); got != before-4 {
		t.Errorf("life %d -> %d, want -4 (its power with the counters)", before, got)
	}
}

func TestGap118ArchivistCrownsOnCommanderDamageOnlyWithNoMonarch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.BlbB1Push(g, me.ID, "Archivist of Gondor", "Creature — Human Advisor", gap118Archivist, 2, 3)
	cmd := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "My Commander", TypeLine: "Legendary Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, IsCommander: true})
	plain := effectstest.BlbB1Push(g, me.ID, "Plain", "Creature — Bear", "", 2, 2)
	effectstest.DealCombatDamageToPlayer(g, plain, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Monarch == me.ID {
		t.Fatal("a non-commander's damage crowned you")
	}
	effectstest.DealCombatDamageToPlayer(g, cmd, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Monarch != me.ID {
		t.Fatalf("monarch = %v, want you after your commander dealt combat damage", g.Monarch)
	}
	// With a monarch already (an opponent), nothing happens.
	g.WithWriteLock(func() { g.Monarch = opp.ID })
	effectstest.DealCombatDamageToPlayer(g, cmd, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Monarch != opp.ID {
		t.Errorf("the crown moved although there was a monarch: %v", g.Monarch)
	}
}

func TestGap118ArchivistDrawsAnExtraCardForTheMonarchAtTheirEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	effectstest.BlbB1Push(g, me.ID, "Archivist of Gondor", "Creature — Human Advisor", gap118Archivist, 2, 3)
	g.WithWriteLock(func() { g.Monarch = me.ID })
	before := me.Hand.Size()
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	// The rules' own monarch draw (CR 725.2) and Archivist's.
	if got := me.Hand.Size(); got != before+2 {
		t.Errorf("hand %d -> %d, want +2 (the monarch draw and Archivist's)", before, got)
	}
}

func TestGap118MonstrosityStunsOnlyWhenPaid(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp, _ := effectstest.BlbB1Seats(g)
		bear := effectstest.BlbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
		if _, err := effectstest.BlbHCastWithMana(t, g, "Monstrosity of the Lake", "Legendary Creature — Kraken", gap118Monstrous, "{4}{U}", 4, 6, "U", "U", "U", "U", "U"); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.B06AddMana(me, "U", "U", "U", "U", "U")
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		stunned := gap118Counters(g, bear, game.CounterStun)
		var tapped bool
		g.ReadSnapshot(func() { c, _ := g.LookupCardForEffect(bear); tapped = c.Tapped })
		if pay && (stunned != 1 || !tapped) {
			t.Errorf("paid: stun counters %d, tapped %v; want 1 and tapped", stunned, tapped)
		}
		if !pay && (stunned != 0 || tapped) {
			t.Errorf("declined: stun counters %d, tapped %v; want none", stunned, tapped)
		}
	}
	spec, _ := effects.Lookup(gap118Monstrous)
	if len(spec.Activated) != 1 || !spec.Activated[0].Cycling {
		t.Error("lacks Islandcycling")
	}
}

func TestGap118BoromirLooksAtSixAndTakesAHumanOrArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	var human, art, land uuid.UUID
	for i := 0; i < 6; i++ {
		c := game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID}
		switch i {
		case 1:
			c.Name, c.TypeLine = "A Human", "Creature — Human Soldier"
			human = c.InstanceID
		case 3:
			c.Name, c.TypeLine = "An Artifact", "Artifact"
			art = c.InstanceID
		case 4:
			c.Name, c.TypeLine = "A Land", "Land"
			land = c.InstanceID
		}
		me.Library.PushTop(c)
	}
	if _, err := effectstest.BlbHCastWithMana(t, g, "Boromir, Gondor's Hope", "Legendary Creature — Human Warrior", gap118Boromir, "{2}{W}{U}", 3, 4, "W", "U", "G", "G"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChooseCards(g, me.ID)
	if p == nil {
		t.Fatal("no look prompt")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range p.ChooseCards {
		offered[id] = true
	}
	if !offered[human] || !offered[art] || offered[land] || len(offered) != 2 {
		t.Errorf("offered %v, want exactly the Human and the artifact", offered)
	}
	if err := g.ResolveChooseCards(p.ID, me.ID, []uuid.UUID{art}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	if !me.Hand.Contains(art) || me.Hand.Contains(human) {
		t.Error("the artifact was not the one taken")
	}
}

func TestGap118FaramirCrownsOnABigLegendAndMakesHumansAtEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	effectstest.BlbB1Push(g, me.ID, "Faramir, Steward of Gondor", "Legendary Creature — Human Noble", gap118Faramir, 2, 2)
	// A legendary creature with mana value 3 does not crown.
	if _, err := effectstest.BlbHCastWithMana(t, g, "Small Legend", "Legendary Creature — Bear", "", "{2}{G}", 2, 2, "G", "G", "G"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Monarch == me.ID {
		t.Fatal("a mana value 3 legend crowned you")
	}
	if _, err := effectstest.BlbHCastWithMana(t, g, "Big Legend", "Legendary Creature — Bear", "", "{3}{G}", 4, 4, "G", "G", "G", "G"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Monarch != me.ID {
		t.Fatal("a mana value 4 legend did not crown you")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(gap118Named(g, me.ID, "Human Soldier")); n != 2 {
		t.Errorf("%d Human Soldier tokens at end step, want 2", n)
	}
}

func TestGap118FaramirMakesNothingWithoutTheCrown(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.BlbB1Push(g, me.ID, "Faramir, Steward of Gondor", "Legendary Creature — Human Noble", gap118Faramir, 2, 2)
	g.WithWriteLock(func() { g.Monarch = opp.ID })
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(gap118Named(g, me.ID, "Human Soldier")); n != 0 {
		t.Errorf("%d tokens without the crown, want none", n)
	}
}

func TestGap118BeregondPumpsOnAHumanAndNotOnABear(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	effectstest.BlbB1Push(g, me.ID, "Beregond of the Guard", "Legendary Creature — Human Soldier", gap118Beregond, 3, 3)
	mine := effectstest.BlbB1Push(g, me.ID, "Mine", "Creature — Bear", "", 2, 2)
	effectstest.CastCatalogSpell(t, g, "A Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := gap118Power(g, mine); p != 2 {
		t.Fatalf("a Bear entering pumped the team: power %d", p)
	}
	effectstest.CastCatalogSpell(t, g, "A Human", "Creature — Human Soldier", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := gap118Power(g, mine); p != 3 {
		t.Errorf("power %d after a Human entered, want 3", p)
	}
	if !effectstest.P1g2HasKeyword(g, mine, "vigilance") {
		t.Error("no vigilance after a Human entered")
	}
}

func TestGap118ChampionsMakeAnOpponentPayOrNotAttack(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp, _ := effectstest.BlbB1Seats(g)
		effectstest.BlbB1Push(g, me.ID, "Champions of Minas Tirith", "Creature — Human Soldier", gap118Champions, 4, 6)
		g.WithWriteLock(func() { g.Monarch = me.ID })
		effectstest.AdvanceToMainOf(t, g, 1)
		atk := effectstest.BlbB1Push(g, opp.ID, "Their Attacker", "Creature — Bear", "", 2, 2)
		hand := opp.Hand.Size()
		if hand == 0 {
			t.Fatal("no cards in the opponent's hand")
		}
		effectstest.AdvanceTo(t, g, game.StepBeginCombat)
		for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
			if err := g.PassPriority(); err != nil {
				t.Fatalf("PassPriority: %v", err)
			}
		}
		if !effectstest.HasPayUnlessFor(g, opp.ID) {
			t.Fatal("the opponent was not asked to pay")
		}
		if pay {
			colors := make([]string, hand)
			for i := range colors {
				colors[i] = "G"
			}
			effectstest.B06AddMana(opp, colors...)
		}
		effectstest.AnswerPayUnless(t, g, opp.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		err := g.DeclareAttacker(atk, me.ID)
		if pay && err != nil {
			t.Errorf("paid, but the attack was refused: %v", err)
		}
		if !pay && err == nil {
			t.Errorf("declined, but the creature could attack the Champions' controller")
		}
	}
}

func TestGap118ChampionsDoNothingWithoutTheCrown(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.BlbB1Push(g, me.ID, "Champions of Minas Tirith", "Creature — Human Soldier", gap118Champions, 4, 6)
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.BlbB1Push(g, opp.ID, "Their Attacker", "Creature — Bear", "", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	for i := 0; i < 8; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	if effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Error("asked to pay although the Champions' controller is not the monarch")
	}
}

func TestGap118ChampionsCrownOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	if _, err := effectstest.BlbHCastWithMana(t, g, "Champions of Minas Tirith", "Creature — Human Soldier", gap118Champions, "{5}{W}", 4, 6, "W", "G", "G", "G", "G", "G"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Monarch != me.ID {
		t.Errorf("monarch = %v, want the caster", g.Monarch)
	}
}

func TestGap118FrontlineMedicBattalionAndSacrificeCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	medic := effectstest.BlbB1Push(g, me.ID, "Frontline Medic", "Creature — Human Cleric", gap118Medic, 3, 3)
	a := effectstest.BlbB1Push(g, me.ID, "A", "Creature — Bear", "", 2, 2)
	b := effectstest.BlbB1Push(g, me.ID, "B", "Creature — Bear", "", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, medic, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{medic, a, b} {
		if !effectstest.P1g2HasKeyword(g, id, "indestructible") {
			t.Errorf("%s lacks indestructible after battalion", id)
		}
	}
}

func TestGap118FrontlineMedicCountersOnlyAnXSpellUnlessPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	medic := effectstest.BlbB1Push(g, me.ID, "Frontline Medic", "Creature — Human Cleric", gap118Medic, 3, 3)
	effectstest.AdvanceToMain(t, g)
	// A spell without {X} in its mana cost is not a legal target.
	plain := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: plain, Name: "Plain Spell", TypeLine: "Sorcery", ManaCost: "{1}{R}", Owner: me.ID, Controller: me.ID})
	effectstest.B06AddMana(me, "R", "R")
	if err := g.CastSpell(me.ID, plain, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast plain: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, medic, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: plain}}}); err == nil {
		t.Error("a spell without {X} in its mana cost was a legal target")
	}
	effectstest.PassPriorityAroundTable(t, g)
	// An X spell is.
	xs := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: xs, Name: "X Spell", TypeLine: "Sorcery", ManaCost: "{X}{R}", Owner: me.ID, Controller: me.ID})
	effectstest.B06AddMana(me, "R")
	if err := g.CastSpell(me.ID, xs, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast X spell: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, medic, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: xs}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(medic) {
		t.Error("the Medic was not sacrificed")
	}
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, me.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("the spell's controller was not asked to pay {3}")
	}
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(xs) {
		t.Error("the unpaid X spell was not countered")
	}
}

func TestGap118GilraenReturnsNowOrAtEndStepWithCounters(t *testing.T) {
	for _, now := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, _, _ := effectstest.BlbB1Seats(g)
		gil := effectstest.BlbB1Push(g, me.ID, "Gilraen, Dúnedain Protector", "Legendary Creature — Human Noble", gap118Gilraen, 2, 3)
		bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
		effectstest.AdvanceToMain(t, g)
		effectstest.B06AddMana(me, "W", "W")
		if err := g.ActivateCatalogAbility(me.ID, gil, 0, game.ActivateAbilityParams{Strict: true,
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		gap118Confirm(t, g, me.ID, now)
		effectstest.PassPriorityAroundTable(t, g)
		if now {
			back := gap118Named(g, me.ID, "Bear")
			if len(back) != 1 {
				t.Fatal("the creature did not come back at once")
			}
			if gap118Counters(g, back[0], game.CounterVigilance) != 0 {
				t.Error("an immediate return got counters")
			}
			continue
		}
		if len(gap118Named(g, me.ID, "Bear")) != 0 {
			t.Fatal("the creature returned before the end step")
		}
		effectstest.AdvanceTo(t, g, game.StepEnd)
		effectstest.PassPriorityAroundTable(t, g)
		back := gap118Named(g, me.ID, "Bear")
		if len(back) != 1 {
			t.Fatal("the creature did not return at the end step")
		}
		if gap118Counters(g, back[0], game.CounterVigilance) != 1 || gap118Counters(g, back[0], game.CounterLifelink) != 1 {
			t.Errorf("counters vigilance=%d lifelink=%d, want 1 each",
				gap118Counters(g, back[0], game.CounterVigilance), gap118Counters(g, back[0], game.CounterLifelink))
		}
	}
}

func TestGap118GreyHostExilesAGraveyardForCountersPerCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	c1 := effectstest.PushGraveyardPermanent(opp, "Dead One", "Creature — Bear", "{1}")
	c2 := effectstest.PushGraveyardPermanent(opp, "Dead Two", "Creature — Bear", "{1}")
	sp := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: sp, Name: "Old Spell", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	if _, err := effectstest.BlbHCastWithMana(t, g, "Grey Host Reinforcements", "Creature — Spirit Soldier", gap118GreyHost, "{3}{W}", 1, 1, "W", "G", "G", "G"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}, true)
	if opp.Graveyard.Size() != 0 || !g.Exile.Contains(c1) || !g.Exile.Contains(c2) || !g.Exile.Contains(sp) {
		t.Error("the whole graveyard was not exiled")
	}
	ids := gap118Named(g, me.ID, "Grey Host Reinforcements")
	if len(ids) != 1 {
		t.Fatal("not on the battlefield")
	}
	if got := gap118Counters(g, ids[0], game.CounterPlusOne); got != 2 {
		t.Errorf("%d counters, want 2 (two creature cards)", got)
	}
	if !effectstest.P1g2HasKeyword(g, ids[0], "flying") {
		t.Error("lacks flying")
	}
}

func TestGap118GwaihirGrantsFlyingAndMakesABirdAfterGainingThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	gw := effectstest.BlbB1Push(g, me.ID, "Gwaihir, Greatest of the Eagles", "Legendary Creature — Bird Noble", gap118Gwaihir, 5, 5)
	ally := effectstest.BlbB1Push(g, me.ID, "Ground Ally", "Creature — Bear", "", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, gw, ally)
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: ally}, true)
	if !effectstest.P1g2HasKeyword(g, ally, "flying") {
		t.Error("the target attacking creature did not gain flying")
	}
	// Two life is not enough; three is.
	effectstest.GainLifeNow(t, g, gw, me, 2)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(gap118Named(g, me.ID, "Bird")); n != 0 {
		t.Fatalf("%d Birds after gaining only 2 life", n)
	}
	effectstest.GainLifeNow(t, g, gw, me, 1)
	// The end step has begun; the trigger waits for the next one, so make a
	// fresh game state for the positive case.
	g2 := effectstest.NewCatalogGame(t)
	me2, _, _ := effectstest.BlbB1Seats(g2)
	effectstest.BlbB1Push(g2, me2.ID, "Gwaihir, Greatest of the Eagles", "Legendary Creature — Bird Noble", gap118Gwaihir, 5, 5)
	effectstest.GainLifeNow(t, g2, uuid.New(), me2, 3)
	effectstest.AdvanceTo(t, g2, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g2)
	birds := gap118Named(g2, me2.ID, "Bird")
	if len(birds) != 1 {
		t.Fatalf("%d Birds after gaining 3 life, want 1", len(birds))
	}
	if gap118Power(g2, birds[0]) != 3 || !effectstest.P1g2HasKeyword(g2, birds[0], "flying") {
		t.Error("the Bird is not a 3/3 with flying")
	}
	var key string
	var nTriggers int
	g2.ReadSnapshot(func() {
		c, _ := g2.LookupCardForEffect(birds[0])
		key = c.TokenKey
		nTriggers = len(game.TriggersForCard(c))
	})
	if key == "" || nTriggers != 1 {
		t.Errorf("the Bird has token key %q and %d triggers, want its one attack trigger", key, nTriggers)
	}
}

func TestGap118GwaihirsBirdGrantsFlyingWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me.ID, effects.S82GapB118GwaihirBirdToken(), 1); err != nil {
			t.Fatalf("CreateTokenForEffect: %v", err)
		}
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].Name == "Bird" {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	birds := gap118Named(g, me.ID, "Bird")
	if len(birds) != 1 {
		t.Fatalf("%d Birds, want 1", len(birds))
	}
	ally := effectstest.BlbB1Push(g, me.ID, "Ground Ally", "Creature — Bear", "", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, birds[0], ally)
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: ally}, true)
	if !effectstest.P1g2HasKeyword(g, ally, "flying") {
		t.Error("the Bird's attack trigger did not give the target attacking creature flying")
	}
}
