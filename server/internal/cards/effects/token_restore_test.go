package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// token_restore_test.go is #521's regression, end to end and with the
// real catalog wired up.
//
// The bug it pins: ADR 0041 refuses to write a restore point for a
// snapshot holding an un-rebuildable Go closure, and a token's
// abilities WERE closures on the instance. So one Treasure on the
// battlefield meant no restore point was written at all, for as long
// as it sat there — a deploy rewound the table past every turn the
// Treasure survived, potentially to turn one.
//
// The assertion that matters is therefore about a board, not about a
// data structure: the four tokens whose whole printed text is an
// ability are on the battlefield, the census is EMPTY, the snapshot
// IS a restore point, and each token still works on the other side of
// the restore.

// artifactTokenBoard puts one of each ability-bearing artifact token
// onto the battlefield under `controller` and returns their IDs by
// name.
func artifactTokenBoard(g *game.Game, controller uuid.UUID) map[string]uuid.UUID {
	return map[string]uuid.UUID{
		"Treasure": pushToken(g, controller, TreasureToken()),
		"Food":     pushToken(g, controller, FoodToken()),
		"Clue":     pushToken(g, controller, ClueToken()),
		"Blood":    pushToken(g, controller, BloodToken()),
	}
}

// TestATableFullOfTokensIsStillARestorePoint is the issue's first two
// acceptance criteria in one run.
func TestATableFullOfTokensIsStillARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]

	if base := g.CaptureSnapshot(); !base.Restorable() {
		t.Fatalf("setup: the empty board is already not a restore point: %+v", base.Continuations)
	}
	ids := artifactTokenBoard(g, me.ID)

	snap := g.CaptureSnapshot()
	if n := snap.Continuations.IntrinsicAbilityCards; n != 0 {
		t.Errorf("a Treasure, a Food, a Clue and a Blood censused %d cards, want 0: %+v",
			n, snap.Continuations)
	}
	if !snap.Restorable() {
		t.Fatalf("a board holding tokens is not a restore point: %+v", snap.Continuations)
	}

	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	you := restored.PlayerByID(me.ID)
	if you == nil {
		t.Fatal("restored game lost the seat")
	}

	// Every token came back as itself, with its printed ability
	// readable through the ordinary accessors — which is what the
	// client, the legal-move enumerator and the activation path all
	// read.
	for name, id := range ids {
		c := findRestoredCard(t, restored, id)
		if c.Name != name {
			t.Errorf("%s came back named %q", name, c.Name)
		}
		if c.TokenKey == "" {
			t.Errorf("%s came back with no token key, so nothing can find its abilities", name)
		}
		mana := game.ManaAbilitiesForCard(*c)
		activated := game.ActivatedAbilitiesForCard(*c)
		if len(mana)+len(activated) == 0 {
			t.Errorf("%s came back with no abilities at all — a blank artifact", name)
		}
		for _, a := range activated {
			if a.Effect == nil {
				t.Errorf("%s came back with an ability that has no effect closure", name)
			}
		}
	}

	// Crack the restored Treasure for mana.
	if err := restored.ActivateManaAbility(you.ID, ids["Treasure"], 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("crack the restored Treasure: %v", err)
	}
	if restored.Battlefield.Contains(ids["Treasure"]) {
		t.Error("the cracked Treasure should have left the battlefield")
	}
	var pending *game.PendingChoice
	for _, c := range restored.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceMana && c.Chooser == you.ID {
			pending = c
		}
	}
	if pending == nil {
		t.Fatal("the restored Treasure queued no colour choice")
	}
	if err := restored.ResolveManaChoice(pending.ID, you.ID, "R"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if len(you.ManaPool) != 1 || you.ManaPool[0].Color != "R" {
		t.Fatalf("mana pool after cracking the restored Treasure = %+v, want one {R}", you.ManaPool)
	}

	// Sacrifice the restored Clue to draw.
	you.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	handBefore := you.Hand.Size()
	if err := restored.ActivateCatalogAbility(you.ID, ids["Clue"], 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("crack the restored Clue: %v", err)
	}
	passPriorityAroundTable(t, restored)
	if you.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d -> %d after cracking the restored Clue, want +1", handBefore, you.Hand.Size())
	}

	// And the Food, because its cost has a tap in it and a restored
	// permanent's tap state is its own axis.
	you.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	lifeBefore := you.Life
	if err := restored.ActivateCatalogAbility(you.ID, ids["Food"], 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("crack the restored Food: %v", err)
	}
	passPriorityAroundTable(t, restored)
	if you.Life != lifeBefore+3 {
		t.Errorf("life %d -> %d after cracking the restored Food, want +3", lifeBefore, you.Life)
	}

	// And the Blood.
	you.ManaPool.AddMana(game.ManaToken{Color: "C"})
	handBefore = you.Hand.Size()
	if err := restored.ActivateCatalogAbility(you.ID, ids["Blood"], 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("crack the restored Blood: %v", err)
	}
	passPriorityAroundTable(t, restored)
	if you.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d -> %d after cracking the restored Blood, want +1", handBefore, you.Hand.Size())
	}
}

// TestEveryTokenTemplateWithAbilitiesIsRestorable walks the whole
// token catalog rather than the four the issue names. A template
// added tomorrow that carries a closure on the instance would put the
// game back to writing no restore points, and would do it silently —
// so the table is the test.
func TestEveryTokenTemplateWithAbilitiesIsRestorable(t *testing.T) {
	for slug, build := range tokenTemplates {
		t.Run(slug, func(t *testing.T) {
			printed := build()
			tmpl := tokenFromCatalog(slug)

			if len(tmpl.ManaAbilities) != 0 || len(tmpl.ActivatedAbilities) != 0 {
				t.Errorf("the template still carries closures on the instance: %d mana, %d activated",
					len(tmpl.ManaAbilities), len(tmpl.ActivatedAbilities))
			}
			if tmpl.TokenKey != game.TokenKey(slug) {
				t.Fatalf("template key = %q, want %q", tmpl.TokenKey, game.TokenKey(slug))
			}
			// The catalog hands back exactly what the printed
			// template declared, which is what makes the swap
			// invisible to every reader.
			if got, want := len(game.ManaAbilitiesForCard(tmpl)), len(printed.ManaAbilities); got != want {
				t.Errorf("catalog returns %d mana abilities, the printed template declares %d", got, want)
			}
			if got, want := len(game.ActivatedAbilitiesForCard(tmpl)), len(printed.ActivatedAbilities); got != want {
				t.Errorf("catalog returns %d activated abilities, the printed template declares %d", got, want)
			}

			g := newCatalogGame(t)
			pushToken(g, g.Seats[0].ID, tmpl)
			snap := g.CaptureSnapshot()
			if n := snap.Continuations.IntrinsicAbilityCards; n != 0 {
				t.Errorf("censused %d cards: %+v", n, snap.Continuations)
			}
			if !snap.Restorable() {
				t.Fatalf("a board holding this token is not a restore point: %+v", snap.Continuations)
			}
		})
	}
}

// TestTokenAbilitiesReachTheCatalogByKey is the seam the wire
// projection reads through, stated on its own: CatalogActivatedAbilities
// answers for a token key exactly as it does for an oracle ID.
func TestTokenAbilitiesReachTheCatalogByKey(t *testing.T) {
	if abs := game.CatalogActivatedAbilities(game.TokenKey("food")); len(abs) != 1 {
		t.Fatalf("CatalogActivatedAbilities(token:food) returned %d abilities, want 1", len(abs))
	}
	if ma := game.CatalogManaAbilities(game.TokenKey("treasure")); len(ma) != 1 {
		t.Fatalf("CatalogManaAbilities(token:treasure) returned %d abilities, want 1", len(ma))
	}
	// An unregistered key is "no entry", not a panic and not a
	// half-answer.
	if abs := game.CatalogActivatedAbilities(game.TokenKey("no-such-token")); abs != nil {
		t.Errorf("an unregistered token key answered with %+v", abs)
	}
}

func findRestoredCard(t *testing.T, g *game.Game, id uuid.UUID) *game.Card {
	t.Helper()
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return &g.Battlefield.Cards[i]
		}
	}
	t.Fatalf("restored battlefield lost card %s", id)
	return nil
}

// TestEveryActionInATreasureGameWritesARestorePoint is the issue's
// last acceptance criterion, and it is the one a player would have
// noticed: the writer (ws.writeRestorePointLocked) skips the write
// whenever Restorable() is false, so a Treasure that censused itself
// meant the file on disk stopped moving the moment one hit the
// battlefield and did not move again until it was spent. This walks a
// short game that makes Treasures the ordinary way, through
// CreateTokenForEffect and the entry pipeline, and demands a restore
// point at EVERY settled point — not only before the first token.
//
// "Settled" is the same condition the writer meets in practice: the
// stack is empty and no prompt is open. A game mid-prompt is
// legitimately unrestorable and always has been; that is a different
// counter and not what #521 is about.
func TestEveryActionInATreasureGameWritesARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]

	restorePoint := func(t *testing.T, after string) {
		t.Helper()
		snap := g.CaptureSnapshot()
		if !snap.Restorable() {
			t.Fatalf("no restore point would be written after %s: %+v", after, snap.Continuations)
		}
		if _, err := snap.RestoreStrict(); err != nil {
			t.Fatalf("after %s: RestoreStrict: %v", after, err)
		}
	}

	restorePoint(t, "the opening board")

	if err := g.CreateTokenForEffect(me.ID, TreasureToken(), 3); err != nil {
		t.Fatalf("create three Treasures: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Treasure"); n != 3 {
		t.Fatalf("battlefield holds %d Treasures, want 3", n)
	}
	restorePoint(t, "three Treasures entered")

	// A step boundary with the Treasures still out. This is the case
	// the bug was worst for: every action for the rest of the game
	// was un-persisted while they sat there.
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	restorePoint(t, "a step boundary with three Treasures out")

	// Spend one.
	treasure := findBattlefieldByName(g, "Treasure")
	if err := g.ActivateManaAbility(me.ID, treasure, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("crack a Treasure: %v", err)
	}
	var pending *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceMana && c.Chooser == me.ID {
			pending = c
		}
	}
	if pending == nil {
		t.Fatal("cracking the Treasure queued no colour choice")
	}
	if err := g.ResolveManaChoice(pending.ID, me.ID, "G"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	restorePoint(t, "one Treasure spent, two still out and {G} floating")

	// And with the last of them gone, for the symmetry: nothing about
	// this change depends on a token being present.
	for _, id := range []uuid.UUID{
		findBattlefieldByName(g, "Treasure"),
	} {
		if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("crack another Treasure: %v", err)
		}
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceMana && c.Chooser == me.ID {
				if err := g.ResolveManaChoice(c.ID, me.ID, "G"); err != nil {
					t.Fatalf("ResolveManaChoice: %v", err)
				}
			}
		}
	}
	restorePoint(t, "two Treasures spent")
}
