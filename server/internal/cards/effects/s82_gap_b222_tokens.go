package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_gap_b222_tokens.go — the Pirate Pursued Whale gives each opponent
// (Commander Legends: Battle for Baldur's Gate, #5948). A catalog template in
// its own file; the slug is an on-disk identity: never renamed, never reused.

// S82GapB222PursuedPirateToken is "a 1/1 red Pirate creature token with 'This
// creature can't block' and 'Creatures you control attack each combat if
// able.'" The second ability is Goblin Spymaster's static, read for the
// token's controller.
func S82GapB222PursuedPirateToken() game.Card {
	return TokenFromCatalog(printedS82GapB222PursuedPirateToken)
}

func printedS82GapB222PursuedPirateToken() TokenTemplate {
	return TokenTemplate{
		Slug: "pirate-cant-block-attack-each-combat",
		Card: game.Card{
			Name:      "Pirate",
			TypeLine:  "Token Creature — Pirate",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"R"},
		},
		Static: []game.StaticAbility{
			RestrictSelf(game.CantBlock),
			Labelled(AttacksEachCombatWhere(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller
			}), "Creatures you control attack each combat if able."),
		},
		Text: "This creature can't block.\nCreatures you control attack each combat if able.",
	}
}
