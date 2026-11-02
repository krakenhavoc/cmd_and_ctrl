package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// junk_shard_tokens_test.go — #4764: the Shard (CR 111.10e) and Junk
// (CR 111.10t) predefined tokens. Each is spawned the way the table
// spawner does it (by its CR 111.10 name), and then used, because a
// token with the right name and no ability is the failure #4738 found.

// A Shard: {2}, sacrifice it, scry 1, then draw the card left on top.
func TestASpawnedShardScriesThenDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tmpl, ok := Tokens().PredefinedToken("Shard")
	if !ok {
		t.Fatal("no Shard template")
	}
	if !tmpl.IsEnchantment() || tmpl.IsArtifact() {
		t.Fatalf("Shard type line %q, want an enchantment token", tmpl.TypeLine)
	}
	ids, err := g.SpawnCards(me.ID, me.ID, game.ZoneBattlefield, tmpl, 1)
	if err != nil || len(ids) != 1 {
		t.Fatalf("SpawnCards: %v (%d spawned)", err, len(ids))
	}
	shard := ids[0]
	top := putOnTopForTest(me, "Top Card", "Sorcery")
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	handBefore := me.Hand.Size()

	if err := g.ActivateCatalogAbility(me.ID, shard, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("crack the Shard: %v", err)
	}
	if g.Battlefield.Contains(shard) {
		t.Error("the Shard was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	scry := latestChoiceOfKind(g, game.PendingChoiceScry)
	if scry == nil {
		t.Fatal("no scry 1 was asked")
	}
	if len(scry.ScryCards) != 1 || scry.ScryCards[0] != top {
		t.Fatalf("scry looked at %v, want only the top card", scry.ScryCards)
	}
	if me.Hand.Size() != handBefore {
		t.Fatal("the draw happened before the scry was answered")
	}
	if err := g.ResolveScry(scry.ID, me.ID, nil, scry.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	if !me.Hand.Contains(top) {
		t.Errorf("the card kept on top was not drawn: hand %d -> %d", handBefore, me.Hand.Size())
	}
}

// A Junk: {T}, sacrifice it, exile the top card with permission to play
// it this turn. Only as a sorcery.
func TestASpawnedJunkExilesTheTopToPlay(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	tmpl, ok := Tokens().PredefinedToken("Junk")
	if !ok {
		t.Fatal("no Junk template")
	}
	ids, err := g.SpawnCards(me.ID, me.ID, game.ZoneBattlefield, tmpl, 1)
	if err != nil || len(ids) != 1 {
		t.Fatalf("SpawnCards: %v (%d spawned)", err, len(ids))
	}
	junk := ids[0]
	top := putOnTopForTest(me, "Top Card", "Sorcery")

	if err := g.ActivateCatalogAbility(me.ID, junk, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate the Junk: %v", err)
	}
	if g.Battlefield.Contains(junk) {
		t.Error("the Junk was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(top) {
		t.Fatal("the top card of the library was not exiled")
	}
	perm := g.CastPermissionOnCardByIDForEffect(top)
	if perm == nil || perm.Player != me.ID {
		t.Errorf("the exiled card is not playable by its controller: %+v", perm)
	}
}

// TestAJunkIsSorcerySpeed — not while a spell is on the stack.
func TestAJunkIsSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	junk := pushToken(g, me.ID, JunkToken())
	putOnTopForTest(me, "Top Card", "Sorcery")
	held := game.NewCard("Test Instant", me.ID)
	held.TypeLine = "Instant"
	me.Hand.PushTop(held)
	if err := g.CastSpell(me.ID, held.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the instant: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, junk, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("a Junk activated with a spell on the stack")
	}
	if !g.Battlefield.Contains(junk) {
		t.Error("a refused activation sacrificed the Junk")
	}
}
