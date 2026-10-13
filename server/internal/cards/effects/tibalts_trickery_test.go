package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const tibaltsTrickeryOracle = "6623cb85-fb1e-4561-8b21-0977c25096c2"

func tibaltsLibCard(owner uuid.UUID, name, typeLine string) game.Card {
	return game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Owner: owner, Controller: owner}
}

// The countered spell's controller, not the caster, mills 1-3, walks
// to the first nonland card with a different name, and is the one asked
// whether to cast it. A land and a card named like the spell are passed
// over and bottomed.
func TestTibaltsTrickeryCountersThenWalksTheVictimsLibrary(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell := batch01OpponentCasts(t, g, opp, "Zap", lightningBoltOracle, "{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})

	// Bottom first, so the last pushed is the top. Three lands absorb
	// the mill of 1, 2 or 3; the walk then meets the rest in order.
	opp.Library.Cards = nil
	hit := tibaltsLibCard(opp.ID, "Divination", "Sorcery")
	mountain := tibaltsLibCard(opp.ID, "Mountain", "Basic Land — Mountain")
	sameName := tibaltsLibCard(opp.ID, "Zap", "Instant")
	for _, c := range []game.Card{
		hit, mountain, sameName,
		tibaltsLibCard(opp.ID, "Plains", "Basic Land — Plains"),
		tibaltsLibCard(opp.ID, "Island", "Basic Land — Island"),
		tibaltsLibCard(opp.ID, "Swamp", "Basic Land — Swamp"),
	} {
		opp.Library.PushTop(c)
	}

	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Tibalt's Trickery", TypeLine: "Instant",
		OracleID: tibaltsTrickeryOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: spell}},
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)

	if !opp.Graveyard.Contains(spell) {
		t.Fatal("the spell was not countered")
	}
	milled := len(opp.Graveyard.Cards) - 1
	if milled < 1 || milled > 3 {
		t.Fatalf("milled %d cards, want 1 to 3", milled)
	}
	if len(me.Graveyard.Cards) != 1 {
		t.Errorf("the caster's graveyard has %d cards, want only Tibalt's Trickery", len(me.Graveyard.Cards))
	}
	var prompt *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceMayCast {
			prompt = c
		}
	}
	if prompt == nil {
		t.Fatal("no may_cast prompt for the countered spell's controller")
	}
	if prompt.Chooser != opp.ID || prompt.MayCastCard != hit.InstanceID {
		t.Fatalf("prompt chooser %s card %s, want the opponent and Divination", prompt.Chooser, prompt.MayCastCard)
	}

	if err := g.ResolveMayCast(prompt.ID, opp.ID, true); err != nil {
		t.Fatalf("ResolveMayCast: %v", err)
	}
	if !g.Exile.Contains(hit.InstanceID) {
		t.Error("the accepted hit should wait in exile under its free cast")
	}
	// The same-name card and the Mountain are bottomed, not milled or
	// kept in exile.
	if !opp.Library.Contains(sameName.InstanceID) || !opp.Library.Contains(mountain.InstanceID) {
		t.Error("the passed-over cards should be back in the library")
	}
	if g.Exile.Contains(sameName.InstanceID) || g.Exile.Contains(mountain.InstanceID) {
		t.Error("the passed-over cards should have left exile")
	}
}

// Declining puts the hit on the bottom too, leaving exile empty.
func TestTibaltsTrickeryDeclineBottomsTheHit(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell := batch01OpponentCasts(t, g, opp, "Zap", lightningBoltOracle, "{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	opp.Library.Cards = nil
	hit := tibaltsLibCard(opp.ID, "Divination", "Sorcery")
	opp.Library.PushTop(hit)
	for i := 0; i < 3; i++ {
		opp.Library.PushTop(tibaltsLibCard(opp.ID, "Island", "Basic Land — Island"))
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Tibalt's Trickery", TypeLine: "Instant",
		OracleID: tibaltsTrickeryOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: spell}},
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerMayCastPrompt(t, g, opp.ID, false)
	if len(g.Exile.Cards) != 0 || !opp.Library.Contains(hit.InstanceID) {
		t.Errorf("declined: exile %d cards, hit in library %v", len(g.Exile.Cards), opp.Library.Contains(hit.InstanceID))
	}
}

// It needs a spell to target: with none on the stack the cast is refused.
func TestTibaltsTrickeryRefusedWithoutASpell(t *testing.T) {
	g := newCatalogGame(t)
	if err := castCatalogSpellErr(t, g, "Tibalt's Trickery", "Instant", tibaltsTrickeryOracle, nil); err == nil {
		t.Fatal("Tibalt's Trickery was cast with no target spell")
	}
}
