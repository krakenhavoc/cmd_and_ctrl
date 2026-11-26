package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b181_helpers.go — shared pieces of the Ravnica Allegiance batch 181
// cards (#5521).

// OldB181IllusionToken is Mesmerizing Benthid's 0/2 blue Illusion creature
// token with "Whenever this token blocks a creature, that creature doesn't
// untap during its controller's next untap step."
func OldB181IllusionToken() game.Card { return TokenFromCatalog(printedOldB181IllusionToken) }

// printedOldB181IllusionToken is that Illusion as PRINTED. Its slug is an
// on-disk identity.
func printedOldB181IllusionToken() TokenTemplate {
	const label = "Illusion — the creature it blocks doesn't untap during its controller's next untap step"
	return TokenTemplate{
		Slug: "illusion-blocks-freeze",
		Card: game.Card{
			Name:      "Illusion",
			TypeLine:  "Token Creature — Illusion",
			Power:     0,
			Toughness: 2,
			Colors:    []string{"U"},
		},
		Triggered: []game.TriggeredAbility{WheneverThisBlocksItDoesntUntap(label)},
		Text:      "Whenever this creature blocks a creature, that creature doesn't untap during its controller's next untap step.",
	}
}
