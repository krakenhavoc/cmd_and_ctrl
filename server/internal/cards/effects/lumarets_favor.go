package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lumaret's Favor — Instant {1}{G} (Secrets of Strixhaven, #4456):
//
//	"Infusion — When you cast this spell, copy it if you gained life this
//	 turn. You may choose new targets for the copy.
//	 Target creature gets +2/+4 until end of turn."
//
// Infusion is an ability word. The cast trigger of the spell itself
// (WhenYouCastThisSpell) always triggers; "if you gained life this turn" is
// part of what it does, not an intervening if (it follows the effect, CR
// 603.4), so the turn's tally is read as the trigger resolves, a life gain
// in response to it included. The copy is Ancestral Communion's CopySpell
// on the spell itself, which asks for new targets, and a copy resolves the
// same pump on its own target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d514704b-4055-48ed-9da3-e4397fca7576",
		Name:         "Lumaret's Favor",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      game.Purpose{Targets: game.ForTargets(BenefitUntilEOT(0, 2, 4))},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 4, Label: "Lumaret's Favor — +2/+4 until end of turn"}.Apply(ctx)
			}
			return nil
		},
		Triggered: []game.TriggeredAbility{
			WhenYouCastThisSpell("Lumaret's Favor — Infusion: copy this spell if you gained life this turn; you may choose new targets for the copy",
				func(g *game.Game, item *game.StackItem) error {
					if b15LifeGainedThisTurn(g, item.Controller) <= 0 {
						return nil
					}
					return CopySpell{StackID: item.SourceCardID, ChooseNewTargets: true}.Apply(NewContext(g, item))
				}),
		},
	})
}
