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
	// #3613, CR 608.2n: the free-cast offer is part of Tibalt's
	// Trickery's resolution, so the spell waits on the stack for it.
	if len(me.Graveyard.Cards) != 0 || !g.Stack.Contains(id) {
		t.Errorf("Tibalt's Trickery should still be on the stack while its offer is open (graveyard %d)", len(me.Graveyard.Cards))
	}
	// #4250, CR 608.2g: the free cast is made during that resolution, a
	// resolution_cast prompt over the hit.
	prompt := keywordCastPromptFor(g, opp.ID)
	if prompt == nil || prompt.Kind != game.PendingChoiceResolutionCast {
		t.Fatal("no resolution_cast prompt for the countered spell's controller")
	}
	if keywordCastCard(prompt) != hit.InstanceID || prompt.MayCastKeyword != game.MayCastKeywordExileUntil {
		t.Fatalf("prompt %+v, want the opponent's exile-until offer of Divination", prompt)
	}
	if prompt.Reason != "Tibalt's Trickery" {
		t.Errorf("prompt reason = %q, want the card's name", prompt.Reason)
	}

	answerMayCastPrompt(t, g, opp.ID, true)
	if len(me.Graveyard.Cards) != 1 || !me.Graveyard.Contains(id) {
		t.Errorf("the caster's graveyard has %d cards, want only Tibalt's Trickery", len(me.Graveyard.Cards))
	}
	if !g.Stack.Contains(hit.InstanceID) {
		t.Error("the hit should be on the stack, cast during the resolution")
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
	if !me.Graveyard.Contains(id) || g.ResolutionPaused() {
		t.Error("the decline did not let Tibalt's Trickery finish resolving")
	}
}

// A table paused on the offer is a restore point: the restored prompt
// casts the hit, or its decline bottoms it with the rest.
func TestTibaltsTrickeryOfferSurvivesARestoreMidPrompt(t *testing.T) {
	for _, cast := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		spell := batch01OpponentCasts(t, g, opp, "Zap", lightningBoltOracle, "{R}",
			[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
		opp.Library.Cards = nil
		hit := tibaltsLibCard(opp.ID, "Divination", "Sorcery")
		opp.Library.PushTop(hit)
		// Three lands over the hit: the mill takes one to three, so the
		// hit is always left for the walk.
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

		restored := restoreRoundTrip(t, g, true)
		ropp := restored.PlayerByIDForEffect(opp.ID)
		rc := keywordCastPromptFor(restored, ropp.ID)
		if rc == nil || keywordCastCard(rc) != hit.InstanceID || !restored.ResolutionPaused() {
			t.Fatalf("cast=%v: the restored table lost the offer: %+v", cast, rc)
		}
		answerMayCastPrompt(t, restored, ropp.ID, cast)
		if cast && !restored.Stack.Contains(hit.InstanceID) {
			t.Error("the restored cast did not reach the stack")
		}
		if !cast && !ropp.Library.Contains(hit.InstanceID) {
			t.Error("the restored decline did not bottom the hit")
		}
		if len(restored.Exile.Cards) != 0 || restored.ResolutionPaused() {
			t.Errorf("cast=%v: exile %d, paused %v after the restored answer", cast, len(restored.Exile.Cards), restored.ResolutionPaused())
		}
	}
}

// It needs a spell to target: with none on the stack the cast is refused.
func TestTibaltsTrickeryRefusedWithoutASpell(t *testing.T) {
	g := newCatalogGame(t)
	if err := castCatalogSpellErr(t, g, "Tibalt's Trickery", "Instant", tibaltsTrickeryOracle, nil); err == nil {
		t.Fatal("Tibalt's Trickery was cast with no target spell")
	}
}
