package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Doom Foretold — Enchantment {2}{W}{B} (Throne of Eldraine, S82 old sets):
//
//	"At the beginning of each player's upkeep, that player sacrifices a
//	 nonland, nontoken permanent of their choice. If that player can't, they
//	 discard a card, they lose 2 life, you draw a card, you gain 2 life, you
//	 create a 2/2 white Knight creature token with vigilance, then you
//	 sacrifice this enchantment."
//
// Invoke Despair's keyed sacrifice for the upkeep player, whose continuation
// reads what they really sacrificed: nothing means they couldn't (also when
// a replacement stopped the sacrifice), so the rest of the sentence runs in
// printed order for this enchantment's controller. The sacrifice of the
// enchantment is skipped when it has already left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e2baf24c-7055-431d-8e96-ba5ff2ef91e1",
		Name:         "Doom Foretold",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtEachUpkeep("Doom Foretold — that player sacrifices a nonland, nontoken permanent of their choice; if they can't, the rest follows",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					player := ctx.Trigger().Event.Actor
					return g.PlayerSacrificesThenKeyedDataForEffect(item.SourceCardID, player,
						effects.SacrificeSpec("a nonland, nontoken permanent", effects.Nonland(), effects.Not(effects.IsTokenPredicate())),
						"Doom Foretold — sacrifice a nonland, nontoken permanent", 1,
						doomForetoldAfter, game.RunData{Carry: []uuid.UUID{item.Controller, player}})
				}),
		},
	})
}

// doomForetoldAfter is "if that player can't, …". An on-disk key: never
// renamed, never reused.
var doomForetoldAfter = game.RegisterPromptRunThen("prompt-run/old-doom-foretold-after", func(g *game.Game, r game.RunDone) error {
	if len(r.Carry) < 2 {
		return nil
	}
	me, victim := r.Carry[0], r.Carry[1]
	if len(r.Sacrifices().By(victim)) > 0 {
		return nil
	}
	ctx := effects.KeyedContext(g, me, r.Source)
	if p := g.PlayerByIDForEffect(victim); p != nil && !p.Eliminated {
		if err := (effects.DiscardCards{Player: victim, N: 1}).Apply(ctx); err != nil {
			return err
		}
		if err := g.ChangePlayerLifeForEffect(r.Source, victim, -2); err != nil {
			return err
		}
	}
	if err := (effects.DrawCards{Player: me, N: 1}).Apply(ctx); err != nil {
		return err
	}
	if err := (effects.GainLife{Player: me, Amount: 2}).Apply(ctx); err != nil {
		return err
	}
	if err := (effects.CreateToken{Controller: me, Template: effects.TokenCard("2/2 white Knight with vigilance"), N: 1}).Apply(ctx); err != nil {
		return err
	}
	if z := g.FindCardZoneForEffect(r.Source); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	return effects.SacrificePermanent{Target: r.Source}.Apply(ctx)
})
