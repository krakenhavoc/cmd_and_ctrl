package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ultimate Magic: Meteor — Sorcery {5}{R} (#3656):
//
//	"Ultimate Magic: Meteor deals 7 damage to each creature. If this
//	 spell was cast from exile, for each opponent, choose an artifact
//	 or land that player controls. Destroy the chosen permanents.
//	 Foretell {5}{R}"
//
// Pyroclasm's damage sweep at 7 and foretell as one special action. The
// exile clause reads the spell's cast zone (CastFromZone), which a
// foretold cast and any other cast from exile both satisfy, as printed.
// The caster picks one artifact or land per opponent in seat order (an
// opponent with none is skipped), and all the chosen permanents are
// destroyed together once the last pick is in.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "31cc5d9d-9384-4a9b-ba7f-6a4edbcb7a35",
		Name:         "Ultimate Magic: Meteor",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 7}},
		Completeness: CompletenessFull,
		SpecialActions: []game.SpecialAction{
			Foretell("{5}{R}"),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := damageEachMatching(ctx, Creature(), 7); err != nil {
				return err
			}
			if item.CastFromZone != game.ZoneExile {
				return nil
			}
			return ultimateMagicMeteorChoose(ctx, item)
		},
	})
}

// ultimateMagicMeteorChoose walks the opponents, asking the caster for
// one artifact or land of each, then destroys the picks.
func ultimateMagicMeteorChoose(ctx *Context, item *game.StackItem) error {
	g, caster := ctx.Game, ctx.Controller()
	opps := ctx.Opponents()
	var chosen []uuid.UUID
	var step func(g *game.Game, i int) error
	step = func(g *game.Game, i int) error {
		for ; i < len(opps); i++ {
			var options []uuid.UUID
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == opps[i] && (c.IsArtifact() || c.IsLand()) {
					options = append(options, c.InstanceID)
				}
			}
			if len(options) == 0 {
				continue
			}
			next := i + 1
			g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:  caster,
				Source:   item.SourceCardID,
				Question: "Ultimate Magic: Meteor — choose an artifact or land that player controls",
				Cards:    options,
				Min:      1,
				Max:      1,
				Zone:     game.ZoneBattlefield,
				Then: func(g *game.Game, picked []uuid.UUID) error {
					chosen = append(chosen, picked...)
					return step(g, next)
				},
			})
			return nil
		}
		dctx := NewContext(g, item).asGroupMember()
		for _, id := range chosen {
			if err := (DestroyTarget{Target: id}).Apply(dctx); err != nil {
				return err
			}
		}
		return nil
	}
	return step(g, 0)
}
