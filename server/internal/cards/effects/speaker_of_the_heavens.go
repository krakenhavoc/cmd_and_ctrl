package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Speaker of the Heavens — Creature — Human Cleric {W}, 1/1 (EDHREC
// rank 3723):
//
//	"Vigilance, lifelink
//	 {T}: Create a 4/4 white Angel creature token with flying. Activate
//	 only if you have at least 7 life more than your starting life total
//	 and only as a sorcery."
//
// The card that prints both activation instructions, which is why
// SorcerySpeed and Condition are separate fields (#743): "only as a
// sorcery" is the timing flag (CR 602.5d), and "at least 7 life more
// than your starting life total" is the condition (CR 602.1b),
// LifeAtLeastAboveStarting(7) — 47 or more at a Commander table. A
// {T} ability on a creature, so summoning sickness applies.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e4bacf1b-0a45-4c73-857d-dffadb7e6fa5",
		Name:            "Speaker of the Heavens",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "lifelink"},
		Activated: []ActivatedAbility{{
			Label:        "{T}: Create a 4/4 white Angel creature token with flying. Activate only if you have at least 7 life more than your starting life total and only as a sorcery.",
			Cost:         TapCost(),
			SorcerySpeed: true,
			Condition:    LifeAtLeastAboveStarting(7),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return CreateToken{Controller: item.Controller, Template: TokenCard("4/4 white Angel with flying"), N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
