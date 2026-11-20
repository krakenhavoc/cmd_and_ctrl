package testsex

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const forTheCommonGoodOracle = "de6e4aa2-23f1-46e5-a061-05522944e0c6"

func TestForTheCommonGoodIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(forTheCommonGoodOracle)
	if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("For the Common Good: %+v %v", spec, ok)
	}
}

func fcgCast(t *testing.T, g *game.Game, me *game.Player, x int, target uuid.UUID, pool string) (uuid.UUID, error) {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "For the Common Good", TypeLine: "Sorcery", OracleID: forTheCommonGoodOracle,
			ManaCost: "{X}{X}{G}", Colors: []string{"G"}, Owner: me.ID, Controller: me.ID})
	})
	effectstest.FloatForTest(g, me, pool)
	var targets []game.TargetRef
	if target != uuid.Nil {
		targets = []game.TargetRef{{Kind: game.TargetCard, ID: target}}
	}
	return id, g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: x, Targets: targets})
}

func TestForTheCommonGoodCopiesXTimesGrantsIndestructibleAndGainsLifePerToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	orig := effectstest.FcgToken(g, me.ID, "Goblin")
	other := effectstest.FcgToken(g, me.ID, "Other token")
	// A nontoken creature is neither copied nor granted anything.
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	life := me.Life
	if _, err := fcgCast(t, g, me, 2, orig, "GGGGG"); err != nil {
		t.Fatalf("cast X=2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	goblins := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsToken() && c.Name == "Goblin" {
			goblins++
		}
	}
	if goblins != 3 {
		t.Fatalf("Goblin tokens: %d, want the original plus two copies", goblins)
	}
	tokens := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller != me.ID {
			continue
		}
		if c.IsToken() {
			tokens++
			if !game.HasKeyword(&c, "indestructible") {
				t.Errorf("token %s is not indestructible", c.Name)
			}
		}
	}
	if tokens != 4 {
		t.Errorf("tokens: %d, want 4", tokens)
	}
	if got := me.Life - life; got != 4 {
		t.Errorf("life gained %d, want 4 (one per token)", got)
	}
	b, _ := effectstest.BattlefieldCard(g, bear)
	if game.HasKeyword(&b, "indestructible") {
		t.Error("a nontoken creature gained indestructible")
	}
	_ = other
}

func TestForTheCommonGoodNeedsATokenAndPaysTwoXPlusG(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if _, err := fcgCast(t, g, me, 1, bear, "GGG"); err == nil {
		t.Error("targeted a nontoken creature")
	}
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	tok := effectstest.FcgToken(g2, me2.ID, "Goblin")
	effectstest.AdvanceToMain(t, g2)
	if _, err := fcgCast(t, g2, me2, 2, tok, "GGGG"); err == nil {
		t.Error("X=2 cast for four mana")
	}
}
