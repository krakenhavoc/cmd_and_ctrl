package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gravity Well — Enchantment {1}{G}{G} (Rise of the Eldrazi, S82 OLD b485):
//
//	"Whenever a creature with flying attacks, it loses flying until end of
//	 turn."
//
// One trigger per attacking creature with flying, whoever controls it
// (Barbed Foliage's attack read). "With flying" is judged as the creature
// attacks (CR 603.2), through the layered view, so a creature that gains
// flying in response is still affected. The loss is a layer-6
// ability-removing effect with its own timestamp (CR 613.1f), so a flying
// granted later in the turn puts it back, and it ends at cleanup (CR 514.2).
// An attacker that has left the battlefield by resolution loses nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "33f757ad-c636-4f17-ba9e-40856f83efd0",
		Name:         "Gravity Well",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventAttack || ev.CardID == uuid.Nil {
					return false
				}
				attacker, ok := g.LookupCardForEffect(ev.CardID)
				return ok && game.HasKeyword(&attacker, "flying")
			}, "Gravity Well — it loses flying until end of turn", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if info, ok := ctx.TriggeringPermanent(); !ok || info.Left {
					return nil
				}
				return effects.UntilEndOfTurn(ctx, ctx.Trigger().Event.CardID, nil,
					"Gravity Well — loses flying until end of turn", game.RemoveKeywordsMod("flying"))
			}),
		},
	})
}
