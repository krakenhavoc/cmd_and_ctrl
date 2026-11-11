package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vizier of Many Faces — Creature — Shapeshifter Cleric {2}{U}{U}, 0/0:
//
//	"You may have this creature enter as a copy of any creature on the
//	 battlefield, except if this creature was embalmed, the token has no
//	 mana cost, it's white, and it's a Zombie in addition to its other
//	 types.
//	 Embalm {3}{U}{U}"
//
// Clone's replacement with an except clause that applies only to the embalm
// token (CR 702.128a). An embalmed Vizier is the only token the card ever
// is with no mana cost: embalm blanks the cost (embalmException) and a
// token copy of any other Vizier keeps {2}{U}{U}, so a token with no mana
// cost is the embalm token. The "except" then lands on the copy the token
// enters as (CR 707.9b adds the Zombie type, it does not replace the types).
// A Vizier cast from a hand, or a non-embalm token, copies without any
// exception.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "55bceb83-bbba-477f-98ce-6e0a80b91080",
		Name:         "Vizier of Many Faces",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			EntersAsCopyOf(
				"Vizier of Many Faces",
				anyCreatureOnBattlefield,
				func(_ *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, src *game.Card) {
					if src == nil || !src.IsToken() || src.ManaCost != "" {
						return
					}
					v.ManaCost = ""
					v.Colors = []string{"W"}
					v.AddSubtype("Zombie")
				},
			),
		},
		Activated: []ActivatedAbility{Embalm("{3}{U}{U}")},
	})
}
