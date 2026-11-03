package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mutagen_token_test.go — #4676: the Mutagen predefined token
// (CR 111.10v). It is spawned the way the table spawner does it (by its
// CR 111.10 name) and then used, because a token with the right name and
// no ability is the failure #4738 found.

// A Mutagen: {1}, {T}, sacrifice it, one +1/+1 counter on target
// creature. Any creature: here an opponent's.
func TestASpawnedMutagenPutsACounterOnTargetCreature(t *testing.T) {
	g, me, opp := spendTable(t)
	tmpl, ok := Tokens().PredefinedToken("Mutagen")
	if !ok {
		t.Fatal("no Mutagen template")
	}
	if !tmpl.IsArtifact() || tmpl.IsCreature() || !tmpl.HasSubtype("Mutagen") || len(tmpl.Colors) != 0 {
		t.Fatalf("Mutagen type line %q colors %v, want a colorless Mutagen artifact token", tmpl.TypeLine, tmpl.Colors)
	}
	ids, err := g.SpawnCards(me.ID, me.ID, game.ZoneBattlefield, tmpl, 1)
	if err != nil || len(ids) != 1 {
		t.Fatalf("SpawnCards: %v (%d spawned)", err, len(ids))
	}
	mutagen := ids[0]
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, mutagen, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
	}); err != nil {
		t.Fatalf("activate the Mutagen: %v", err)
	}
	if g.Battlefield.Contains(mutagen) {
		t.Error("the Mutagen was not sacrificed")
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("the {1} was not spent: pool %+v", me.ManaPool)
	}
	passPriorityAroundTable(t, g)
	if got := plusOneCounters(g, theirs); got != 1 {
		t.Errorf("the target has %d +1/+1 counters, want 1", got)
	}
}

// The cost is {1} and {T}: no mana, or a tapped Mutagen, and nothing
// happens (the token stays).
func TestAMutagenNeedsItsManaAndItsTap(t *testing.T) {
	g, me, _ := spendTable(t)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	target := []game.TargetRef{{Kind: game.TargetCard, ID: bear}}

	broke := pushToken(g, me.ID, MutagenToken())
	if err := g.ActivateCatalogAbility(me.ID, broke, 0, game.ActivateAbilityParams{Strict: true, Targets: target}); err == nil {
		t.Error("a Mutagen activated without {1}")
	}
	if !g.Battlefield.Contains(broke) {
		t.Error("a refused activation sacrificed the Mutagen")
	}

	tapped := pushToken(g, me.ID, MutagenToken())
	tapForTest(t, g, tapped)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, tapped, 0, game.ActivateAbilityParams{Strict: true, Targets: target}); err == nil {
		t.Error("a tapped Mutagen activated")
	}
	if !g.Battlefield.Contains(tapped) {
		t.Error("a refused activation sacrificed the tapped Mutagen")
	}
	if got := plusOneCounters(g, bear); got != 0 {
		t.Errorf("the bear has %d +1/+1 counters after two refused activations", got)
	}
}

// TestAMutagenIsSorcerySpeed — not while a spell is on the stack.
func TestAMutagenIsSorcerySpeed(t *testing.T) {
	g, me, _ := spendTable(t)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	mutagen := pushToken(g, me.ID, MutagenToken())
	held := game.NewCard("Test Instant", me.ID)
	held.TypeLine = "Instant"
	me.Hand.PushTop(held)
	if err := g.CastSpell(me.ID, held.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the instant: %v", err)
	}

	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, mutagen, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Fatal("a Mutagen activated with a spell on the stack")
	}
	if !g.Battlefield.Contains(mutagen) {
		t.Error("a refused activation sacrificed the Mutagen")
	}
}

// TestAMutagenWhoseTargetLeftDoesNothing — the creature is bounced in
// response; the counter goes nowhere and the token is still spent.
func TestAMutagenWhoseTargetLeftDoesNothing(t *testing.T) {
	g, me, _ := spendTable(t)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	other := pushVanillaCreature(g, me.ID, "Other Bear", 2, 2)
	mutagen := pushToken(g, me.ID, MutagenToken())
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, mutagen, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate the Mutagen: %v", err)
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, bear); err != nil {
		t.Fatalf("bounce the target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mutagen) {
		t.Error("the Mutagen came back")
	}
	if got := plusOneCounters(g, other); got != 0 {
		t.Errorf("the other creature got %d counters", got)
	}
}
