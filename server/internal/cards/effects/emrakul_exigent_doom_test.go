package effects

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// emrakul_exigent_doom_test.go — Emrakul, the Exigent Doom's hand ability
// (#3520): the {3}, exile-from-hand cost, the land that taps for {C}{C}
// until the card is cast from exile, and the standing cast permission.

const emrakulExigentOracle = "4421ab7d-6d9b-4edd-b5a0-53a8ed84da6f"

// emrakulSetup puts Emrakul in me's hand, an untapped land on the
// battlefield and {3} in the pool, in a main phase.
func emrakulSetup(t *testing.T) (g *game.Game, me *game.Player, emrakul, land uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	emrakul = uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: emrakul, Name: "Emrakul, the Exigent Doom", TypeLine: "Legendary Creature — Eldrazi",
		ManaCost: "{10}", OracleID: emrakulExigentOracle, Power: 12, Toughness: 12,
		Owner: me.ID, Controller: me.ID,
	})
	land = pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Wastes", TypeLine: "Basic Land", Owner: me.ID, Controller: me.ID,
	})
	return
}

// emrakulActivate pays {3} and activates the hand ability on `target`.
func emrakulActivate(t *testing.T, g *game.Game, me *game.Player, emrakul, target uuid.UUID) error {
	t.Helper()
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}"); err != nil {
		t.Fatal(err)
	}
	return g.ActivateCatalogAbility(me.ID, emrakul, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}, Strict: true,
	})
}

func castEmrakulFromExile(g *game.Game, me *game.Player, emrakul uuid.UUID) error {
	return g.CastSpell(me.ID, emrakul, game.CastSpellParams{FromZone: "exile", Strict: true})
}

func TestEmrakulHandAbilityExilesItAndGrantsTheLandCC(t *testing.T) {
	g, me, emrakul, land := emrakulSetup(t)
	if err := emrakulActivate(t, g, me, emrakul, land); err != nil {
		t.Fatalf("activate from hand: %v", err)
	}
	if !inExile(g, emrakul) {
		t.Fatal("the exile is a cost: Emrakul is in exile as the ability goes on the stack")
	}
	if me.Hand.Contains(emrakul) {
		t.Fatal("Emrakul is still in hand")
	}
	passPriorityAroundTable(t, g)
	if n := grantedManaRowsOf(g, land); n != 1 {
		t.Fatalf("the land has %d granted mana abilities, want 1", n)
	}
	if err := tapGrantedMana(t, g, me.ID, land, ""); err != nil {
		t.Fatalf("tap the land for the granted mana: %v", err)
	}
	if got := poolSize(g, me.ID); got != 2 {
		t.Errorf("the pool holds %d mana, want {C}{C}", got)
	}
}

// The land pays for the cast from exile, and the grant is over once
// Emrakul is on the stack.
func TestEmrakulCastFromExileUsesTheGrantedManaThenTheGrantEnds(t *testing.T) {
	g, me, emrakul, land := emrakulSetup(t)
	if err := emrakulActivate(t, g, me, emrakul, land); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)

	// {10} is 8 from the pool and {C}{C} from the land: short of either
	// alone.
	floatForTest(g, me, "CCCCCCCC")
	if err := castEmrakulFromExile(g, me, emrakul); err == nil {
		t.Fatal("cast Emrakul with {C}x8 and the land untapped: ten mana were not paid")
	}
	if err := tapGrantedMana(t, g, me.ID, land, ""); err != nil {
		t.Fatalf("tap the granted ability before the cast: %v", err)
	}
	if err := castEmrakulFromExile(g, me, emrakul); err != nil {
		t.Fatalf("cast Emrakul from exile with {C}{C} from the land: %v", err)
	}
	if n := grantedManaRowsOf(g, land); n != 0 {
		t.Errorf("the land keeps %d granted mana abilities after Emrakul was cast", n)
	}
}

// "You may cast this card for as long as it remains exiled": no price
// but {10}, and not from anywhere else.
func TestEmrakulCastFromExileIsPricedAndGatedToExile(t *testing.T) {
	g, me, emrakul, land := emrakulSetup(t)
	if err := emrakulActivate(t, g, me, emrakul, land); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	var perm *game.CastPermission
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(emrakul)
		perm = g.CastPermissionOnCardForEffect(c, game.ZoneExile)
	})
	if perm == nil || perm.Player != me.ID || perm.Cost != "" || perm.Duration.Kind != game.WhileInZone {
		t.Fatalf("cast permission = %+v, want the holder's, at the printed cost, while exiled", perm)
	}
}

// Emrakul leaving exile without being cast leaves the land's ability.
func TestEmrakulLeavingExileUncastKeepsTheLandsAbility(t *testing.T) {
	g, me, emrakul, land := emrakulSetup(t)
	if err := emrakulActivate(t, g, me, emrakul, land); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneExile}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, emrakul); err != nil {
		t.Fatalf("move Emrakul out of exile: %v", err)
	}
	if n := grantedManaRowsOf(g, land); n != 1 {
		t.Errorf("the land has %d granted mana abilities, want 1 (ruling 2026-08-21)", n)
	}
	// And the cast permission went with the card.
	if err := castEmrakulFromExile(g, me, emrakul); err == nil {
		t.Error("Emrakul was cast from exile after leaving it")
	}
}

// Another card cast from exile does not end Emrakul's grant.
func TestEmrakulGrantSurvivesAnotherCardBeingCastFromExile(t *testing.T) {
	g, me, emrakul, land := emrakulSetup(t)
	if err := emrakulActivate(t, g, me, emrakul, land); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	other := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: other, Name: "Other", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(other); err != nil {
			t.Fatalf("exile the other card: %v", err)
		}
		g.GrantCastPermissionOverCardForEffect(other, game.CastPermission{
			Player: me.ID, Duration: game.WhileInZoneDuration(), CastOnly: true, Timing: game.TimingFlash,
		})
	})
	if err := g.CastSpell(me.ID, other, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("cast the other card from exile: %v", err)
	}
	if n := grantedManaRowsOf(g, land); n != 1 {
		t.Errorf("the land has %d granted mana abilities, want 1: another card's cast ended it", n)
	}
}

// A land that is no longer there when the ability resolves: nothing
// happens, and Emrakul stays exiled with no permission.
func TestEmrakulFizzlesWithNoPermissionWhenTheLandLeaves(t *testing.T) {
	g, me, emrakul, land := emrakulSetup(t)
	if err := emrakulActivate(t, g, me, emrakul, land); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(land); err != nil {
			t.Fatalf("destroy the land: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if !inExile(g, emrakul) {
		t.Fatal("Emrakul left exile")
	}
	if err := castEmrakulFromExile(g, me, emrakul); err == nil {
		t.Error("Emrakul is castable from exile although the ability did nothing")
	}
}

func TestEmrakulCantTargetANonland(t *testing.T) {
	g, me, emrakul, _ := emrakulSetup(t)
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	err := emrakulActivate(t, g, me, emrakul, bear)
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("target a creature: err = %v, want ErrIllegalTarget", err)
	}
}

// The window survives a snapshot: the restored game ends the grant on
// the cast, not before.
func TestEmrakulGrantAndPermissionSurviveARestore(t *testing.T) {
	g, me, emrakul, land := emrakulSetup(t)
	if err := emrakulActivate(t, g, me, emrakul, land); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)

	r := restoreRoundTrip(t, g, false)
	rme := r.PlayerByIDForEffect(me.ID)
	if n := grantedManaRowsOf(r, land); n != 1 {
		t.Fatalf("restored: the land has %d granted mana abilities, want 1", n)
	}
	floatForTest(r, rme, "CCCCCCCCCC")
	if err := castEmrakulFromExile(r, rme, emrakul); err != nil {
		t.Fatalf("restored: cast from exile: %v", err)
	}
	if n := grantedManaRowsOf(r, land); n != 0 {
		t.Errorf("restored: the grant outlived the cast (%d rows)", n)
	}
}

// What a bot picks from: the hand ability is a move with a land to
// target, and once exiled the cast from exile is a move.
func TestEmrakulEnumeratorOffersTheAbilityThenTheCast(t *testing.T) {
	g, me, emrakul, land := emrakulSetup(t)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}"); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Kind == legal.KindActivate && paramsHaveID(string(m.Params), emrakul) && paramsHaveID(string(m.Params), land) {
			found = true
		}
	}
	if !found {
		t.Error("the hand ability targeting the land is not an enumerated move")
	}
	if err := emrakulActivate(t, g, me, emrakul, land); err != nil {
		// The pool was topped up twice above; the second payment is
		// harmless to the assertion below.
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	floatForTest(g, me, "CCCCCCCCCC")
	cast := false
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Kind == legal.KindCast && paramsHaveID(string(m.Params), emrakul) {
			cast = true
		}
	}
	if !cast {
		t.Error("the cast from exile is not an enumerated move")
	}
}

func paramsHaveID(s string, id uuid.UUID) bool { return strings.Contains(s, id.String()) }

func TestEmrakulIsComplete(t *testing.T) {
	spec, ok := Lookup(emrakulExigentOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Emrakul, the Exigent Doom ships in full: %+v", spec)
	}
}
