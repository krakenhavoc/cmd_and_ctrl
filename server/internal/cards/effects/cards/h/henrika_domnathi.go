package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// henrikaDomnathiOracle is shared by both faces: the front registers
// under it, the back under henrikaDomnathiOracle + "#1".
const HenrikaDomnathiOracle = "0c0845c3-f4b5-444e-8f42-da0c7dbf2841"

// henrikaDomnathiLabel is the front face's trigger label, and with it
// the key its "hasn't been chosen" memory is kept under.
const henrikaDomnathiLabel = "Henrika Domnathi — beginning of combat"

// Henrika Domnathi // Henrika, Infernal Seer — a transforming
// legendary Vampire:
//
//	Henrika Domnathi {2}{B}{B}, 1/3
//	  "Flying
//	   At the beginning of combat on your turn, choose one that hasn't
//	   been chosen —
//	   • Each player sacrifices a creature of their choice.
//	   • You draw a card and you lose 1 life.
//	   • Transform Henrika."
//
//	Henrika, Infernal Seer, 3/4
//	  "Flying, deathtouch, lifelink
//	   {1}{B}{B}: Each creature you control with flying, deathtouch,
//	   and/or lifelink gets +1/+0 until end of turn."
//
// ChooseOneNotChosen (ADR 0097) with no duration: each bullet is
// available once for this object. Transforming is not a zone change
// (CR 712.18), so a Henrika that transformed and — through some other
// effect — transformed back would still remember; the back face has
// no such trigger to consult it anyway. A Henrika that leaves and
// returns is a new object and may choose all three again.
//
// The edict is each player's own choice, the controller included
// (EachPlayerSacrifices). "Transform Henrika" turns this permanent
// over in place (Transform, ADR 0079): counters, damage and its
// timestamp stay. The back face's pump reads keywords live, so a
// creature that has flying only through an effect counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        HenrikaDomnathiOracle,
		Name:            "Henrika Domnathi",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			henrikaDomnathiTrigger(),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        HenrikaDomnathiOracle + "#1",
		Name:            "Henrika, Infernal Seer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch", "lifelink"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}{B}: Each creature you control with flying, deathtouch, and/or lifelink gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{1}{B}{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.BoostUntilEOT{
					Match: effects.And(effects.Creature(), effects.YouControl(),
						effects.Or(effects.HasKeyword("flying"), effects.HasKeyword("deathtouch"), effects.HasKeyword("lifelink"))),
					Power: 1,
					Label: "Henrika, Infernal Seer — +1/+0 until end of turn",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

func henrikaDomnathiTrigger() game.TriggeredAbility {
	t := effects.AtBeginningOfYourCombat(henrikaDomnathiLabel, func(_ *game.Game, _ *game.StackItem) error { return nil })
	t.Modes = effects.ChooseOneNotChosen(
		effects.ModeDoing("Each player sacrifices a creature of their choice.", nil,
			func(_ *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.EachPlayerSacrifices{Match: effects.Creature(), Label: "a creature"}.Apply(ctx)
			}),
		effects.ModeDoing("You draw a card and you lose 1 life.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.B36DrawAndLoseOne(ctx.Game, item)
			}),
		effects.ModeDoing("Transform Henrika.", nil,
			func(_ *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.TransformThis{}.Apply(ctx)
			}),
	)
	return t
}
