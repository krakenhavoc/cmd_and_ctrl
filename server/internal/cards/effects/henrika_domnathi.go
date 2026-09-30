package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// henrikaDomnathiOracle is shared by both faces: the front registers
// under it, the back under henrikaDomnathiOracle + "#1".
const henrikaDomnathiOracle = "0c0845c3-f4b5-444e-8f42-da0c7dbf2841"

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
	Register(Spec{
		OracleID:        henrikaDomnathiOracle,
		Name:            "Henrika Domnathi",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			henrikaDomnathiTrigger(),
		},
	})
	Register(Spec{
		OracleID:        henrikaDomnathiOracle + "#1",
		Name:            "Henrika, Infernal Seer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch", "lifelink"},
		Activated: []ActivatedAbility{{
			Label: "{1}{B}{B}: Each creature you control with flying, deathtouch, and/or lifelink gets +1/+0 until end of turn.",
			Cost:  ManaCost("{1}{B}{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return BoostUntilEOT{
					Match: And(Creature(), YouControl(),
						Or(HasKeyword("flying"), HasKeyword("deathtouch"), HasKeyword("lifelink"))),
					Power: 1,
					Label: "Henrika, Infernal Seer — +1/+0 until end of turn",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}

func henrikaDomnathiTrigger() game.TriggeredAbility {
	t := AtBeginningOfYourCombat(henrikaDomnathiLabel, func(_ *game.Game, _ *game.StackItem) error { return nil })
	t.Modes = ChooseOneNotChosen(
		ModeDoing("Each player sacrifices a creature of their choice.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				return EachPlayerSacrifices{Match: Creature(), Label: "a creature"}.Apply(ctx)
			}),
		ModeDoing("You draw a card and you lose 1 life.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return b36DrawAndLoseOne(ctx.Game, item)
			}),
		ModeDoing("Transform Henrika.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				return TransformThis{}.Apply(ctx)
			}),
	)
	return t
}
