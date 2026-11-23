package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dragon Whelp — Creature — Dragon {2}{R}{R}, 2/3:
//
//	"Flying
//	 {R}: This creature gets +1/+0 until end of turn. If this ability has
//	 been activated four or more times this turn, sacrifice this creature
//	 at the beginning of the next end step."
//
// Shivan Dragon's pump with Farrelite Priest's rider: the count is the
// per-object, per-ability, per-turn activation record
// (game.ActivatedThisTurn), read as the ability resolves, so the fourth
// activation (and any later one) schedules the sacrifice. The delayed
// trigger is bound to the object, so a flickered Whelp is not sacrificed.
//
// No simplification.
func init() {
	const label = "{R}: This creature gets +1/+0 until end of turn. If this ability has been activated four or more times this turn, sacrifice this creature at the beginning of the next end step."
	rider := effects.SacrificeAtNextEndStepIfActivatedRider(4, label, "Dragon Whelp — sacrifice this creature")
	effects.Register(effects.Spec{
		OracleID:        "705a1985-ed39-4a4b-812e-a677170b596e",
		Name:            "Dragon Whelp",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   label,
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := (effects.BoostUntilEOT{Target: ctx.Source(), Power: 1, Label: "Dragon Whelp — +1/+0"}).Apply(ctx); err != nil {
					return err
				}
				return rider(g, item.Controller, ctx.Source())
			},
		}},
	})
}
