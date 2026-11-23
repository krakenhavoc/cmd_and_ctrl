package testsl

import (
	"testing"

	"github.com/google/uuid"

	cardshu "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/hu"
	cardsj "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/j"
	cardsv "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/v"
	cardsx "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/x"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestJidoorOvertureMillsHalfThenTheLandIsPlayableFromExile(t *testing.T) {
	g, me, id := effectstest.SeatLandAdventure(t, cardsj.JidoorOracleID, "Jidoor, Aristocratic Capital", "Overture", "{4}{U}{U}", []string{"U"})
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	before := opp.Library.Size()
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Face:    1,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("cast Overture: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Library.Size(); got != before-before/2 {
		t.Errorf("library %d -> %d, want half milled", before, got)
	}
	effectstest.AssertLandPlayableFromExile(t, g, me, id)
}

func TestJidoorIsAPlainLandPlayFromHand(t *testing.T) {
	g, me, id := effectstest.SeatLandAdventure(t, cardsj.JidoorOracleID, "Jidoor, Aristocratic Capital", "Overture", "{4}{U}{U}", []string{"U"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("play Jidoor: %v", err)
	}
	if !g.Battlefield.Contains(id) || g.Stack.Contains(id) {
		t.Fatal("Jidoor should be on the battlefield as a land play, not on the stack")
	}
}

func TestLastingFaythMakesAHeroWithACounterPerLand(t *testing.T) {
	g, me, id := effectstest.SeatLandAdventure(t, cardsx.ZanarkandOracleID, "Zanarkand, Ancient Metropolis", "Lasting Fayth", "{4}{G}{G}", []string{"G"})
	for i := 0; i < 3; i++ {
		l := game.NewCard("Test Forest", me.ID)
		l.TypeLine = "Basic Land — Forest"
		g.Battlefield.PushTop(l)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast Lasting Fayth: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var hero *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Hero" {
			hero = &g.Battlefield.Cards[i]
		}
	}
	if hero == nil {
		t.Fatal("no Hero token")
	}
	if got := hero.CurrentPower(); got != 4 {
		t.Errorf("Hero power = %d, want 1 + 3 counters (three lands)", got)
	}
	effectstest.AssertLandPlayableFromExile(t, g, me, id)
}

func TestTakeATripToDrawsTwoAndHitsEachOpponent(t *testing.T) {
	g, me, id := effectstest.SeatLandAdventure(t, cardsv.ValueTownOracleID, "Value Town", "Take a Trip to...", "{4}{U}{R}", []string{"U", "R"})
	handBefore := me.Hand.Size()
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast Take a Trip to...: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// The spell left the hand (-1) and two cards were drawn (+2).
	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand %d -> %d, want +1 net (drew two, cast one)", handBefore, got)
	}
	for _, p := range g.Seats {
		want := lives[p.ID]
		if p.ID != me.ID {
			want -= 2
		}
		if p.Life != want {
			t.Errorf("%s life = %d, want %d", p.Name, p.Life, want)
		}
	}
	effectstest.AssertLandPlayableFromExile(t, g, me, id)
}

func TestFaithAndGriefReturnsUpToTwoArtifactsOrEnchantments(t *testing.T) {
	g, me, id := effectstest.SeatLandAdventure(t, cardshu.IshgardOracleID, "Ishgard, the Holy See", "Faith & Grief", "{3}{W}{W}", []string{"W"})
	mk := func(name, typeLine string) uuid.UUID {
		c := game.NewCard(name, me.ID)
		c.TypeLine = typeLine
		me.Graveyard.PushTop(c)
		return c.InstanceID
	}
	rock := mk("Test Rock", "Artifact")
	aura := mk("Test Ward", "Enchantment")
	bear := mk("Test Bear", "Creature — Bear")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Face: 1,
		Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: rock},
			{Kind: game.TargetCard, ID: aura},
		},
	}); err != nil {
		t.Fatalf("cast Faith & Grief: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(rock) || !me.Hand.Contains(aura) {
		t.Error("both targets should return to hand")
	}
	if !me.Graveyard.Contains(bear) {
		t.Error("the creature was never targetable and should stay put")
	}
	effectstest.AssertLandPlayableFromExile(t, g, me, id)

	// A creature card is not a legal target.
	g2, me2, id2 := effectstest.SeatLandAdventure(t, cardshu.IshgardOracleID, "Ishgard, the Holy See", "Faith & Grief", "{3}{W}{W}", []string{"W"})
	c := game.NewCard("Test Bear", me2.ID)
	c.TypeLine = "Creature — Bear"
	me2.Graveyard.PushTop(c)
	if err := g2.CastSpell(me2.ID, id2, game.CastSpellParams{
		Face:    1,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: c.InstanceID}},
	}); err == nil {
		t.Error("a creature card was accepted as a target")
	}
}
