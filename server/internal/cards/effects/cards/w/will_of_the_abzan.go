package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Will of the Abzan — Sorcery {3}{B} (EDHREC rank 2605):
//
//	"Choose one. If you control a commander as you cast this spell,
//	 you may choose both instead.
//	 • Any number of target opponents each sacrifice a creature with
//	   the greatest power among creatures that player controls and
//	   lose 3 life.
//	 • Return target creature card from your graveyard to the
//	   battlefield."
//
// An edict that takes the biggest creature, or a reanimation, and
// with a commander out, both. The first mode targets any number of
// opponents — CR 115.1d, so zero is a legal choice and resolves as a
// no-op; each still-legal one chooses a creature with the greatest
// power among their own (GreatestPowerYouControl evaluated for the
// chooser through the sacrifice prompt, so ties are theirs to break —
// b24PlayersSacrificeGreatestPowerCreatureAndLoseLife). The targets
// choose in APNAP order and their creatures are sacrificed together
// (CR 101.4, #3664); then each loses 3 life, in announce order. The
// second returns the chosen creature
// card under its owner's control, which is the caster's: the clause
// says "your graveyard".
//
// The commander rider is the conditional mode count #1590 built:
// `OrUpToIf(2, YouControlACommander)`, read at announce (CR 601.2b)
// and fixed from then on. With both chosen, each bullet reads its own
// target group through OptionTargets (#764) and the body runs them in
// PRINTED order (CR 608.2c): the edict, then the reanimation.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1ae29791-aa7c-4050-bf72-dd0f739b11b8",
		Name:         "Will of the Abzan",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.Mode("Any number of target opponents each sacrifice a creature with the greatest power among creatures that player controls and lose 3 life.",
				effects.TargetPlayer("any number of target opponents", effects.Opponent()).WithCount(0, 0)),
			effects.Mode("Return target creature card from your graveyard to the battlefield.",
				effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature())),
		).OrUpToIf(2, effects.YouControlACommander),
		OnResolve: effects.ResolveSteps(
			// The return waits for a prompt the sacrifices ask (CR 608.2c, #4155).
			effects.Instruction(func(ctx *effects.Context) error {
				if !ctx.HasMode(0) {
					return nil
				}
				var players []uuid.UUID
				for _, t := range effects.OptionTargets(ctx, 0) {
					if t.Kind == game.TargetPlayer {
						players = append(players, t.ID)
					}
				}
				if len(players) == 0 {
					return nil
				}
				return effects.B24PlayersSacrificeGreatestPowerCreatureAndLoseLife(ctx.Game, ctx.Item, players, 3)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if !ctx.HasMode(1) {
					return nil
				}
				for _, t := range effects.OptionTargets(ctx, 1) {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
		),
	})
}
