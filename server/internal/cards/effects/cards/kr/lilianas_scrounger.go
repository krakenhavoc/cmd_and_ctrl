package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Liliana's Scrounger — Creature — Human Wizard {2}{B}, 3/2 (Core Set 2021,
// #5511):
//
//	"At the beginning of each end step, if a creature died this turn, you may
//	 put a loyalty counter on a Liliana planeswalker you control."
//
// "Each end step" is every player's. An intervening "if" (CR 603.4) over the
// table-wide tally of creatures that died this turn (any controller, tokens
// included), checked as the step begins and again as the trigger resolves. The
// "may" is a choice of up to one Liliana planeswalker you control (the Liliana
// subtype, not a name); choosing none, or controlling none, does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "595f1dee-8c3b-44a6-9e71-a4bb4334b309",
		Name:         "Liliana's Scrounger",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "a creature died this turn",
				Test: func(ctx *effects.Context) bool { return ctx.Game.TurnTally.CreaturesDied > 0 },
			}, effects.On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
				return effects.AnyPlayer(ev, source, ch, g) && g.TurnTally.CreaturesDied > 0
			}, "Liliana's Scrounger — you may put a loyalty counter on a Liliana planeswalker you control",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ChoosePermanents{
						Question: "Liliana's Scrounger — put a loyalty counter on a Liliana planeswalker you control?",
						Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
							var out []uuid.UUID
							for _, c := range g.BattlefieldCardsForEffect() {
								if c.Controller == of && c.IsPlaneswalker() && c.HasSubtype("Liliana") {
									out = append(out, c.InstanceID)
								}
							}
							return out, 0, 1
						},
						Keyed: lilianasScroungerCounter,
					}.Apply(effects.NewContext(g, item))
				})),
		},
	})
}

// lilianasScroungerCounter puts the loyalty counter on the chosen planeswalker.
// An on-disk key: never renamed, never reused.
var lilianasScroungerCounter = effects.RegisterPermanentsPick("prompt-run/s82-old-b108-lilianas-scrounger-loyalty", effects.PermanentsDef{
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		for _, id := range picked.Cards() {
			if !effects.OnBattlefield(ctx.Game, id) {
				continue
			}
			return effects.AddCounter{Target: id, Kind: game.CounterLoyalty, N: 1}.Apply(ctx)
		}
		return nil
	},
})
