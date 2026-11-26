package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b533_cards_test.go — Shadowmoor: Conspire, batch 533 (#5591, S82
// OLD). Every test plays the real catalog entry through the engine.

const (
	oldB533Roar      = "04e9ac31-3377-4a4c-9337-53ee0ada882a"
	oldB533Discovery = "1ca875f0-50a7-4332-8643-c2b4bbda18b1"
	oldB533Plot      = "3efecb6c-7c21-4b3b-9612-e3916f91c680"
	oldB533Trail     = "566b4dcc-b7d7-4443-b2d0-954f1fc33a11"
	oldB533Giant     = "894a220b-1372-4a76-b38f-bf0d7001af41"
	oldB533Mine      = "9bd5dc7a-11b0-48eb-bf52-2956dd3989ab"
	oldB533Sluice    = "a30b659e-ae2b-4736-9b4f-a9805905bf75"
	oldB533Bark      = "baf3d4d9-0590-4d94-8f4e-1e726194b920"
	oldB533Tow       = "dd0cac88-0a78-408f-8eb0-080f79150180"
)

func TestOldB533CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB533Roar: "Traitor's Roar", oldB533Discovery: "Ghastly Discovery", oldB533Plot: "Disturbing Plot",
		oldB533Trail: "Burn Trail", oldB533Giant: "Giantbaiting", oldB533Mine: "Mine Excavation",
		oldB533Sluice: "Memory Sluice", oldB533Bark: "Barkshell Blessing", oldB533Tow: "Aethertow",
	} {
		c := effectstest.RequireFullNoCaveats(t, oracle, name)
		if len(c.OptionalCosts) != 1 || c.OptionalCosts[0].Key != game.ConspireKey {
			t.Errorf("%s does not carry exactly one conspire cost", name)
		}
	}
}

// oldB533Cast casts a spell from the active seat's hand with its mana floated.
func oldB533Cast(t *testing.T, g *game.Game, card game.Card, mana string, params game.CastSpellParams) error {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	card.InstanceID = uuid.New()
	card.Owner, card.Controller = me.ID, me.ID
	me.Hand.PushTop(card)
	if g.Turn.Step != game.StepDeclareAttackers && g.Turn.Step != game.StepDeclareBlockers {
		effectstest.AdvanceToMain(t, g)
	}
	effectstest.FloatForTest(g, me, mana)
	params.Strict = true
	if err := g.CastSpell(me.ID, card.InstanceID, params); err != nil {
		return err
	}
	return nil
}

// oldB533Settle resolves the stack, answering each pending target pick (the
// conspire copy's "new target") with retarget.
func oldB533Settle(t *testing.T, g *game.Game, who uuid.UUID, retarget game.TargetRef) {
	t.Helper()
	for i := 0; i < 64; i++ {
		if effectstest.StackFullyEmpty(g) && effectstest.LatestPickTarget(g, who) == nil {
			return
		}
		if p := effectstest.LatestPickTarget(g, who); p != nil {
			if err := g.ResolvePickTarget(p.ID, who, retarget); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack did not settle")
}

// oldB533Tappers puts two creatures of a colour onto seat's battlefield.
func oldB533Tappers(g *game.Game, seat *game.Player, color string) []uuid.UUID {
	var ids []uuid.UUID
	for _, n := range []string{"Tapper A", "Tapper B"} {
		ids = append(ids, effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: n,
			TypeLine: "Creature — Elf", ManaCost: "{" + color + "}", Colors: []string{color}, Power: 1, Toughness: 1,
			Owner: seat.ID, Controller: seat.ID}))
	}
	return ids
}

func TestOldB533BurnTrailDealsThreeAndConspireCopiesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	o1 := g.Seats[1]
	card := game.Card{Name: "Burn Trail", TypeLine: "Sorcery", OracleID: oldB533Trail, ManaCost: "{3}{R}", Colors: []string{"R"}}
	if err := oldB533Cast(t, g, card, "CCCR", game.CastSpellParams{Targets: []game.TargetRef{effectstest.PlayerTarget(o1.ID)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if o1.Life != 37 {
		t.Errorf("target life %d, want 37", o1.Life)
	}

	// Conspire: two red creatures are tapped; the copy picks a new target.
	g2 := effectstest.NewCatalogGame(t)
	me, o1, o2 := g2.Seats[0], g2.Seats[1], g2.Seats[2]
	tap := oldB533Tappers(g2, me, "R")
	if err := oldB533Cast(t, g2, card, "CCCR", game.CastSpellParams{
		Targets: []game.TargetRef{effectstest.PlayerTarget(o1.ID)}, OptionalCosts: []int{0}, ConspireIDs: tap,
	}); err != nil {
		t.Fatalf("conspired cast: %v", err)
	}
	if !effectstest.IsTapped(g2, tap[0]) || !effectstest.IsTapped(g2, tap[1]) {
		t.Fatal("conspire did not tap the two creatures")
	}
	oldB533Settle(t, g2, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: o2.ID})
	if o1.Life != 37 || o2.Life != 37 {
		t.Errorf("lives %d and %d, want 37 and 37 (the copy hit the new target)", o1.Life, o2.Life)
	}
}

func TestOldB533BurnTrailConspireRefusesCreaturesOfAnotherColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, o1 := g.Seats[0], g.Seats[1]
	tap := oldB533Tappers(g, me, "G")
	err := oldB533Cast(t, g, game.Card{Name: "Burn Trail", TypeLine: "Sorcery", OracleID: oldB533Trail, ManaCost: "{3}{R}", Colors: []string{"R"}},
		"CCCR", game.CastSpellParams{Targets: []game.TargetRef{effectstest.PlayerTarget(o1.ID)}, OptionalCosts: []int{0}, ConspireIDs: tap})
	if err == nil {
		t.Fatal("conspire accepted creatures that share no colour with the spell")
	}
}

func TestOldB533TraitorsRoarTapsAndTheCreatureHurtsItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	tapped := effectstest.B12Creature(g, opp.ID, "Sleeper", "Creature — Ogre", 4, 4)
	effectstest.TapForTest(t, g, tapped)
	card := game.Card{Name: "Traitor's Roar", TypeLine: "Sorcery", OracleID: oldB533Roar, ManaCost: "{4}{B/R}", Colors: []string{"B", "R"}}
	if err := oldB533Cast(t, g, card, "CCCCB", game.CastSpellParams{Targets: effectstest.CardTarget(tapped)}); err == nil {
		t.Fatal("targeted a tapped creature")
	}
	if err := oldB533Cast(t, g, card, "CCCCB", game.CastSpellParams{Targets: effectstest.CardTarget(ogre)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	_ = me
	if !effectstest.IsTapped(g, ogre) {
		t.Error("the Ogre is not tapped")
	}
	if opp.Life != 37 {
		t.Errorf("its controller has %d life, want 37", opp.Life)
	}
}

func TestOldB533GhastlyDiscoveryDrawsTwoThenDiscardsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C")
	before := me.Hand.Size()
	if err := oldB533Cast(t, g, game.Card{Name: "Ghastly Discovery", TypeLine: "Sorcery", OracleID: oldB533Discovery, ManaCost: "{2}{U}", Colors: []string{"U"}},
		"CCU", game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DiscardFromHand(t, g, me.ID)
	if me.Hand.Size() != before+1 {
		t.Errorf("hand %d, want %d (the spell left, two drawn, one discarded)", me.Hand.Size(), before+1)
	}
}

func TestOldB533DisturbingPlotReturnsAnyPlayersCreatureCardToItsOwner(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.GraveyardCardOf(g, opp, "Bear", "Creature — Bear", "{1}{G}", "")
	relic := effectstest.GraveyardCardOf(g, opp, "Relic", "Artifact", "{2}", "")
	card := game.Card{Name: "Disturbing Plot", TypeLine: "Sorcery", OracleID: oldB533Plot, ManaCost: "{1}{B}", Colors: []string{"B"}}
	if err := oldB533Cast(t, g, card, "CB", game.CastSpellParams{Targets: effectstest.CardTarget(relic)}); err == nil {
		t.Fatal("targeted a noncreature card")
	}
	if err := oldB533Cast(t, g, card, "CB", game.CastSpellParams{Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	_ = me
	if !opp.Hand.Contains(bear) || opp.Graveyard.Contains(bear) {
		t.Error("the creature card did not return to its owner's hand")
	}
}

func TestOldB533MineExcavationReturnsAnArtifactOrEnchantmentCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	relic := effectstest.GraveyardCardOf(g, opp, "Relic", "Artifact", "{2}", "")
	aura := effectstest.GraveyardCardOf(g, opp, "Aura", "Enchantment — Aura", "{1}{W}", "")
	bear := effectstest.GraveyardCardOf(g, opp, "Bear", "Creature — Bear", "{1}{G}", "")
	card := game.Card{Name: "Mine Excavation", TypeLine: "Sorcery", OracleID: oldB533Mine, ManaCost: "{1}{W}", Colors: []string{"W"}}
	if err := oldB533Cast(t, g, card, "CW", game.CastSpellParams{Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Fatal("targeted a creature card")
	}
	for _, id := range []uuid.UUID{relic, aura} {
		if err := oldB533Cast(t, g, card, "CW", game.CastSpellParams{Targets: effectstest.CardTarget(id)}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !opp.Hand.Contains(id) {
			t.Error("the card did not return to its owner's hand")
		}
	}
}

func TestOldB533MemorySluiceMillsFourAndConspireMillsAnotherPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, o1, o2 := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.SeedLibrary(o1, "A", "B", "C", "D", "E", "F", "G", "H")
	effectstest.SeedLibrary(o2, "A", "B", "C", "D", "E", "F", "G", "H")
	l1, l2 := o1.Library.Size(), o2.Library.Size()
	tap := oldB533Tappers(g, me, "U")
	card := game.Card{Name: "Memory Sluice", TypeLine: "Sorcery", OracleID: oldB533Sluice, ManaCost: "{U/B}", Colors: []string{"U", "B"}}
	if err := oldB533Cast(t, g, card, "U", game.CastSpellParams{Targets: []game.TargetRef{effectstest.PlayerTarget(o1.ID)}, OptionalCosts: []int{0}, ConspireIDs: tap}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	oldB533Settle(t, g, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: o2.ID})
	if o1.Library.Size() != l1-4 || o2.Library.Size() != l2-4 {
		t.Errorf("libraries %d and %d, want %d and %d", o1.Library.Size(), o2.Library.Size(), l1-4, l2-4)
	}
}

func TestOldB533BarkshellBlessingPumpsUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := oldB533Cast(t, g, game.Card{Name: "Barkshell Blessing", TypeLine: "Instant", OracleID: oldB533Bark, ManaCost: "{G/W}", Colors: []string{"G", "W"}},
		"G", game.CastSpellParams{Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PowerOf(t, g, bear); got != 4 {
		t.Errorf("power %d, want 4", got)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if got := effectstest.PowerOf(t, g, bear); got != 2 {
		t.Errorf("power %d after the turn, want 2", got)
	}
}

func TestOldB533AethertowPutsAnAttackerOnTopOfItsOwnersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, opp.ID, "Idle", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, atk)
	card := game.Card{Name: "Aethertow", TypeLine: "Instant", OracleID: oldB533Tow, ManaCost: "{3}{W/U}", Colors: []string{"W", "U"}}
	if err := oldB533Cast(t, g, card, "CCCW", game.CastSpellParams{Targets: effectstest.CardTarget(idle)}); err == nil {
		t.Fatal("targeted a creature that is neither attacking nor blocking")
	}
	if err := oldB533Cast(t, g, card, "CCCW", game.CastSpellParams{Targets: effectstest.CardTarget(atk)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(atk) {
		t.Fatal("the attacker is still on the battlefield")
	}
	if top, err := me.Library.Top(); err != nil || top.InstanceID != atk {
		t.Error("the attacker is not on top of its owner's library")
	}
}

func TestOldB533GiantbaitingMakesAHasteGiantExiledAtTheNextEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tap := oldB533Tappers(g, me, "R")
	if err := oldB533Cast(t, g, game.Card{Name: "Giantbaiting", TypeLine: "Sorcery", OracleID: oldB533Giant, ManaCost: "{2}{R/G}", Colors: []string{"R", "G"}},
		"CCR", game.CastSpellParams{OptionalCosts: []int{0}, ConspireIDs: tap}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	giants := effectstest.TokensNamed(g, "Giant Warrior")
	if len(giants) != 2 {
		t.Fatalf("%d Giant Warriors, want 2 (the spell and its conspire copy)", len(giants))
	}
	for _, c := range giants {
		if c.Power != 4 || c.Toughness != 4 || !effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "haste") {
			t.Errorf("token %s is not a 4/4 with haste", c.Name)
		}
	}
	effectstest.EndTheTurn(t, g)
	if n := len(effectstest.TokensNamed(g, "Giant Warrior")); n != 0 {
		t.Errorf("%d Giant Warriors survived the end step", n)
	}
}
