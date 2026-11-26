package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Unconventional Tactics — {2}{W} Sorcery (Hour of Devastation):
//
//	"Target creature gets +3/+3 and gains flying until end of turn.
//	 Whenever a Zombie you control enters, you may pay {W}. If you do, return
//	 this card from your graveyard to your hand."
//
// The second sentence is a triggered ability that works from the graveyard
// only (CR 113.6k): "you" is the card's owner while it sits there, and the
// return is for the same graveyard object that asked (CR 400.7).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "95ee46ec-4b60-4fea-a09b-d15c083916b3",
		Name:         "Unconventional Tactics",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3, "flying")),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: t.ID, Power: 3, Toughness: 3, Label: "Unconventional Tactics — +3/+3"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"flying"}, Label: "Unconventional Tactics — flying"}.Apply(ctx)
		},
		Triggered: []game.TriggeredAbility{
			effects.InGraveyard(effects.On(game.EventETB,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
					return ok && c.HasSubtype(source.TextSubtype("Zombie"))
				},
				"Unconventional Tactics — you may pay {W}; if you do, return it to your hand",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPayForGraveyardReturnToHand(g, item, "{W}",
						"Unconventional Tactics — pay {W} to return it from your graveyard to your hand?")
				})),
		},
	})
}
