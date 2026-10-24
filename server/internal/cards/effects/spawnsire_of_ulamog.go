package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spawnsire of Ulamog — Creature — Eldrazi {10}, 7/11:
//
//	"Annihilator 1
//	 {4}: Create two 0/1 colorless Eldrazi Spawn creature tokens. They
//	 have "Sacrifice this token: Add {C}."
//	 {20}: Cast any number of Eldrazi spells from among cards you own
//	 outside the game without paying their mana costs."
//
// Annihilator 1 is the canonical keyword token (ADR 0113 §2), and the
// Spawn are the shared Eldrazi Spawn token.
//
// The {20} ability is PlayFromOutside (#3955, ADR 0198's 2026-10-16
// amendment). As it resolves the owner chooses any number of Eldrazi
// cards from outside the game (their sideboard, or the card database
// when the deck seated none), and each one is granted a free, any-timing
// (CR 608.2g), cast-only, pass-closed cast from the container. The casts
// are then ordinary cast_spell actions with `from_zone: "outside"`,
// taken from the sideboard tab, before the holder's next priority pass;
// Eldrazi that are not cast by then stay outside the game. Additional
// costs a card has are still paid (CR 118.9) and X is 0 (CR 107.3b).
//
// "Eldrazi spells" are Eldrazi creature cards and the like, never a
// land: a card whose type line has the Eldrazi type and is a land does
// not exist, and a land is not a spell anyway.
func init() {
	Register(Spec{
		OracleID:        "b90d2f4d-b4ea-40af-aea0-1ab4234ab80f",
		Name:            "Spawnsire of Ulamog",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"annihilator 1"},
		Activated: []ActivatedAbility{
			{
				Label:   "{4}: Create two 0/1 colorless Eldrazi Spawn creature tokens. They have \"Sacrifice this token: Add {C}.\"",
				Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
				Cost:    ManaCost("{4}"),
				Effect:  Do(CreateToken{Template: EldraziSpawnToken(), N: 2}),
			},
			{
				Label:   "{20}: Cast any number of Eldrazi spells from among cards you own outside the game without paying their mana costs.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    ManaCost("{20}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					// "Eldrazi" is a word of the source's text, read as
					// the ability resolves (CR 612.2, 608.2h; #3984).
					eldrazi := ctx.TextSubtype("Eldrazi")
					return PlayFromOutside{
						Pick:   spawnsireOfUlamogCast,
						Word:   eldrazi,
						Clause: "any number of Eldrazi cards",
					}.Apply(ctx)
				},
			},
		},
	})
}

// spawnsireOfUlamogCast is the {20} ability's keyed pick (#4256). The
// subtype is the card's word, which a text-changing effect may have
// changed, so it rides on the prompt.
var spawnsireOfUlamogCast = registerPlayFromOutside("choose-cards/spawnsire-of-ulamog-cast", "Spawnsire of Ulamog",
	func(c game.Card, eldrazi string) bool { return c.HasSubtype(eldrazi) && !c.IsLand() })
