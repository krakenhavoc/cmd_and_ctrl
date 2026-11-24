package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Academic Probation — Sorcery — Lesson {1}{W} (Strixhaven: School of Mages,
// S82 batch 047, #5502):
//
//	"Choose one —
//	 • Choose a nonland card name. Opponents can't cast spells with the
//	   chosen name until your next turn.
//	 • Choose target nonland permanent. Until your next turn, it can't
//	   attack or block, and its activated abilities can't be activated."
//
// The first bullet asks for the name as the spell resolves (Tamiyo's
// prompt) and the bans wait on it as a registered tail told the name as its
// Word: one named cast ban for each opponent (Reflector Mage's
// CastBanNamed), lasting until the caster's next turn begins (CR 611.2b).
// The name is free text (CR 201.2), so "nonland" is enforced where the
// engine can: a land card is never cast, so naming one bans nothing.
//
// The second bullet is Edifice of Authority's restriction on a targeted
// nonland permanent: it can't attack or block and neither its activated
// abilities (loyalty abilities included, CR 606.1) nor its mana abilities
// can be activated, until the caster's next turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "400b2e17-ee48-4e22-a8a6-a607c215d1a2",
		Name:         "Academic Probation",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Choose a nonland card name. Opponents can't cast spells with the chosen name until your next turn.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					ctx.Game.QueueCardNameChoiceKeyedForEffect(ctx.Controller(), ctx.Source(),
						"Academic Probation — choose a nonland card name",
						ctx.ItemTail(academicProbationNamed))
					return nil
				}),
			effects.ModeDoing("Choose target nonland permanent. Until your next turn, it can't attack or block, and its activated abilities can't be activated.",
				effects.TargetPermanent("target nonland permanent", effects.Nonland()),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetCard || !effects.OnBattlefield(ctx.Game, t.ID) {
						return nil
					}
					return effects.ScopedEffectFor{
						Target:   t.ID,
						Mods:     []game.Mod{game.AddRestrictionsMod(game.CantAttackOrBlock | game.CantActivate | game.CantActivateMana)},
						Duration: effects.DurationUntilYourNextTurn(ctx, ctx.Controller()),
						Label:    "Academic Probation — can't attack or block, and its activated abilities can't be activated, until your next turn",
					}.Apply(ctx)
				}),
		),
	})
}

// academicProbationNamed is the ban once the name (the tail's Word) is
// known: every opponent can't cast spells with it until the caster's next
// turn. An on-disk key.
var academicProbationNamed = effects.CardTail("then/academic-probation-named", func(ctx *effects.Context, o game.TailOutcome) error {
	name := o.Tail.Word
	if name == "" {
		return nil
	}
	d := effects.DurationUntilYourNextTurn(ctx, ctx.Controller())
	for _, opp := range ctx.Opponents() {
		if err := (effects.RestrictCasting{
			Player: opp,
			Rule: game.CastBanRule{
				Kind:   game.CastBanNamed,
				Filter: game.PermissionFilter{Named: name},
			},
			Label:    "Academic Probation — can't cast spells named " + name,
			Duration: d,
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
