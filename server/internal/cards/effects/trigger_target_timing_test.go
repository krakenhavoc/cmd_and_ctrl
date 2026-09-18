package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trigger_target_timing_test.go — #809, the catalog half. Persist
// returning Dualcaster Mage is the shape the catalog soak found: the
// Mage's ETB ("copy target instant or sorcery spell") is dispatched
// from inside Persist's own resolution, so it saw a stack that still
// held Persist. CR 608.2n puts Persist in the graveyard as the final
// step of that resolution, and CR 603.3d then has nothing left for
// the trigger to target.

const tt809ReverberateOracle = "a1f55890-31c5-4ed4-a2cd-7a4a9f05f8ca"

// tt809SeedMage puts a Dualcaster Mage card in the seat's graveyard —
// a legal Persist target (nonlegendary creature card you own).
func tt809SeedMage(g *game.Game, p *game.Player) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id,
		Name:       "Dualcaster Mage",
		TypeLine:   "Creature — Human Wizard",
		OracleID:   b05DualcasterMageOracle,
		Power:      2,
		Toughness:  2,
		Owner:      p.ID,
		Controller: p.ID,
	})
	return id
}

// The reproduction from #809. Persist is the only spell on the stack,
// so by the time its resolution is over there is no instant or
// sorcery spell for the Mage's ETB to copy. CR 603.3d removes the
// trigger; before this fix the prompt was queued offering Persist
// itself, every answer came back `game: illegal target`, and the #791
// choice gate turned that into a table nobody could move.
func TestPersistReturningDualcasterMageDoesNotWedgeTheTable(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mage := tt809SeedMage(g, me)

	castCatalogSpell(t, g, "Persist", "Sorcery", b12PersistOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mage}})
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(mage) {
		t.Fatal("Persist should have returned the Mage to the battlefield")
	}
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Errorf("CR 603.3d: nothing is left to copy, so the trigger must be removed — got a prompt offering %v", p.PickTargetCards)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("no prompt of any kind should be pending, got %d", len(g.PendingChoices))
	}
	// The wedge: #791 gates the table on any unanswered prompt.
	if err := g.PassPriority(); err != nil {
		t.Fatalf("the table is wedged: PassPriority: %v", err)
	}
}

// The same trigger with a spell genuinely left underneath. Reverberate
// copies Persist, the COPY resolves and returns the Mage, and the
// original Persist is still on the stack when the ETB is put there —
// so the prompt offers the original and not the copy, which ceased to
// exist on resolution (CR 707.10).
func TestDualcasterMageOffersOnlyTheSpellStillOnTheStack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mage := tt809SeedMage(g, me)

	persist := castCatalogSpell(t, g, "Persist", "Sorcery", b12PersistOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mage}})
	castCatalogSpell(t, g, "Reverberate", "Instant", tt809ReverberateOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: persist}})

	// Reverberate resolves and opens the CR 707.10c re-target prompt
	// for the copy; keep the same target.
	passPriorityAroundTable(t, g)
	retarget := latestPickTarget(g, me.ID)
	if retarget == nil {
		t.Fatal("Reverberate should ask whether to re-target the copy")
	}
	pickCard(t, g, me.ID, mage)

	// The copy resolves, returns the Mage, and its ETB triggers while
	// the copy is still on the stack with the original Persist below.
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mage) {
		t.Fatal("the Persist copy should have returned the Mage")
	}
	prompt := latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("the original Persist is still on the stack, so the ETB should prompt")
	}
	if !hasID(prompt.PickTargetCards, persist) {
		t.Errorf("the original Persist is still a spell on the stack and must be offered: %v", prompt.PickTargetCards)
	}
	if len(prompt.PickTargetCards) != 1 {
		t.Errorf("the copy ceased to exist on resolution and must not be offered: %v", prompt.PickTargetCards)
	}
	if err := g.ResolvePickTarget(prompt.ID, me.ID,
		game.TargetRef{Kind: game.TargetCard, ID: persist}); err != nil {
		t.Fatalf("answering with the surviving spell: %v", err)
	}
}
