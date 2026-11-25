package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodmist Infiltrator — Creature — Vampire {2}{B}, 3/1 (Ravnica
// Allegiance):
//
//	"Whenever this creature attacks, you may sacrifice another creature. If you
//	 do, this creature can't be blocked this turn."
//
// The Meep's optional sacrifice with a registered "if you do": only a creature
// that really left lets the Infiltrator slip past blockers, even when it has
// already been declared as an attacker.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dcebce37-02eb-4b5d-86be-42b4a0287624",
		Name:         "Bloodmist Infiltrator",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Bloodmist Infiltrator — you may sacrifice another creature; if you do, this creature can't be blocked this turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MaySacrificeOne(g, item, "Bloodmist Infiltrator — sacrifice another creature?", effects.Creature(), true, bloodmistInfiltratorSacrificed)
				}),
		},
	})
}

var bloodmistInfiltratorSacrificed = effects.RegisterSacrificedOne("prompt-run/bloodmist-infiltrator-unblockable", func(ctx *effects.Context, _ uuid.UUID) error {
	return effects.RestrictUntilEOT{Target: ctx.Source(), Restrictions: game.CantBeBlocked,
		Label: "Bloodmist Infiltrator — can't be blocked this turn"}.Apply(ctx)
})
