package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// borrowed_last_cards_test.go — ADR 0155 PR 4 (#3202), the last cards on
// the borrow seam: the resolved borrows (Quicksilver Elemental, Grell
// Philosopher, Havengul Lich), the library-top lenders (Conspicuous Snoop,
// Skill Borrower) and Scheming Fence. Each test asserts what a player
// gets: which permanent has which ability, what pays for it and what
// using it does.

const (
	SlQuicksilverOracle = "c88e3aff-c539-4e8b-8ee1-a43b3d0962b4"
	SlGrellOracle       = "d5dc00cb-1510-4ce9-8dd9-0722f65d34b4"
	SlLichOracle        = "1fb02bad-0d1f-437f-a269-f51cc19b280f"
	SlSnoopOracle       = "f5d1bd3c-0e65-4999-a810-881b2389b40e"
	SlSkillOracle       = "9e862e4a-254d-4d39-b2cf-4beed6db761f"
	SlFenceOracle       = "cef87a98-61a0-4a88-950f-867144f685e3"

	SlShadeOracle       = "9af1b6c4-295f-41d2-b3a0-0863798330d4" // Evernight Shade: {B}: +1/+1 until end of turn
	SlPuzzleknotOracle  = "3334b218-01e3-47a3-8a28-5cc2decbedea" // Woodweaver's Puzzleknot: {2}{G}, sac: gain 3, {E}{E}{E}
	SlSharpshooterOracl = "d81285b7-a718-411a-8be3-ecc0cfe0bcb0" // Goblin Sharpshooter: {T}: 1 damage
	SlLilianaOracle     = "0ba134d8-ee7d-48ec-8dc6-57942b8e9261" // Liliana of the Veil
)

// --- Scheming Fence -------------------------------------------------------

func SlChoose(t *testing.T, g *game.Game, chooser uuid.UUID, picks ...uuid.UUID) {
	t.Helper()
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoiceChooseCards && c.Chooser == chooser {
			if err := g.ResolveChooseCards(c.ID, chooser, picks); err != nil {
				t.Fatalf("ResolveChooseCards: %v", err)
			}
			return
		}
	}
	t.Fatal("no choose-cards prompt")
}
