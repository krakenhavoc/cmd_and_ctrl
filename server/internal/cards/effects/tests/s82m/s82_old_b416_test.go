package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b416_test.go — Gatecrash, batch 416 (#5565): every card played
// through the engine.

const (
	oldB416Loyalist = "560a5452-0dc3-421d-bf16-8843a35dae6f"
	oldB416Angel    = "ad0f50d5-7b40-4fd8-8d13-4a0493398f53"
)

func TestOldB416AllRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		oldB416Loyalist: "Legion Loyalist",
		oldB416Angel:    "Deathpact Angel",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

// --- Deathpact Angel -----------------------------------------------------------

func TestOldB416DeathpactAngelLeavesAClericWithTheReturnAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	angel := effectstest.B12Push(g, me.ID, "Deathpact Angel", "Creature — Angel", oldB416Angel, 5, 5)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(angel) })
	effectstest.PassPriorityAroundTable(t, g)
	clerics := effectstest.SocB17Names(g, me.ID, "Cleric")
	if len(clerics) != 1 {
		t.Fatalf("%d Clerics, want 1", len(clerics))
	}
	view := effectstest.GaLayered(t, g, clerics[0].InstanceID)
	if p, tough := effectstest.PtOf(t, g, clerics[0].InstanceID); p != 1 || tough != 1 || len(view.Colors) != 2 {
		t.Errorf("the Cleric is %d/%d with colours %v, want 1/1 white and black", p, tough, view.Colors)
	}
	// It is a new object: it can't {T} until its controller has controlled it
	// since the start of their turn.
	effectstest.B06AddMana(me, "C", "C", "C", "W", "B", "B")
	if err := g.ActivateCatalogAbility(me.ID, clerics[0].InstanceID, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the fresh Cleric paid {T} while summoning sick")
	}
}

func oldB416ReadyCleric(g *game.Game, owner uuid.UUID) uuid.UUID {
	tok := effects.OldB416ClericToken()
	tok.InstanceID, tok.Owner, tok.Controller = uuid.New(), owner, owner
	return effectstest.PushBattlefieldCardWithTimestamp(g, tok)
}

func TestOldB416ClericReturnsDeathpactAngelFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	angel := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: angel, Name: "Deathpact Angel", TypeLine: "Creature — Angel", OracleID: oldB416Angel,
		Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
	other := effectstest.PushGraveyardCardForTest(me, "Some Other Card")
	cleric := oldB416ReadyCleric(g, me.ID)
	effectstest.B06AddMana(me, "C", "C", "C", "W", "B", "B")
	if err := g.ActivateCatalogAbility(me.ID, cleric, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(angel) {
		t.Error("Deathpact Angel did not return to the battlefield")
	}
	if g.Battlefield.Contains(cleric) {
		t.Error("the Cleric was not sacrificed")
	}
	if !me.Graveyard.Contains(other) {
		t.Error("another card left the graveyard")
	}
}

func TestOldB416ClericDoesNothingWithNoAngelInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	cleric := oldB416ReadyCleric(g, me.ID)
	effectstest.B06AddMana(me, "C", "C", "C", "W", "B", "B")
	if err := g.ActivateCatalogAbility(me.ID, cleric, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cleric) {
		t.Error("the Cleric was not sacrificed as a cost")
	}
}

func TestOldB416ClericAsksWhichAngelToReturnWhenThereAreSeveral(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	var angels []uuid.UUID
	for i := 0; i < 2; i++ {
		id := uuid.New()
		angels = append(angels, id)
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Deathpact Angel", TypeLine: "Creature — Angel", OracleID: oldB416Angel,
			Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
	}
	cleric := oldB416ReadyCleric(g, me.ID)
	effectstest.B06AddMana(me, "C", "C", "C", "W", "B", "B")
	if err := g.ActivateCatalogAbility(me.ID, cleric, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsPrompt(t, g)
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{angels[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(angels[1]) || g.Battlefield.Contains(angels[0]) {
		t.Error("the picked Angel did not return, or the other one did")
	}
}

// --- Legion Loyalist ---------------------------------------------------------------

func oldB416Opp(g *game.Game) *game.Player { return g.Seats[1] }

func TestOldB416LegionLoyalistBattalionGivesFirstStrikeTrampleAndBarsTokenBlockers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], oldB416Opp(g)
	loyalist := effectstest.B12Push(g, me.ID, "Legion Loyalist", "Creature — Goblin Soldier", oldB416Loyalist, 1, 1)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "Idle Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.TokenCard("1/1 white Soldier"), 1) })
	token := effectstest.SocB17Names(g, opp.ID, "Soldier")[0].InstanceID
	effectstest.Pr7bAttack(t, g, opp.ID, loyalist, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	late := effectstest.B12Creature(g, me.ID, "Late Bear", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{loyalist, a, b, idle} {
		for _, kw := range []string{"first strike", "trample"} {
			if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
				t.Errorf("a creature I controlled did not gain %s", kw)
			}
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, late, "trample") {
		t.Error("a creature that arrived after resolution gained trample")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "trample") {
		t.Error("an opponent's creature gained trample")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(token, a); err == nil {
		t.Error("a creature token blocked a creature that can't be blocked by tokens")
	}
	if err := g.DeclareBlocker(theirs, a); err != nil {
		t.Errorf("a nontoken creature could not block: %v", err)
	}
}

func TestOldB416LegionLoyalistNeedsTwoOtherAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], oldB416Opp(g)
	loyalist := effectstest.B12Push(g, me.ID, "Legion Loyalist", "Creature — Goblin Soldier", oldB416Loyalist, 1, 1)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	effectstest.Pr7bAttack(t, g, opp.ID, loyalist, a)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasEffectiveKeyword(t, g, a, "trample") || effectstest.HasEffectiveKeyword(t, g, loyalist, "first strike") {
		t.Error("battalion triggered with only one other attacker")
	}
}
