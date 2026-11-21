package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Robot Domination — Enchantment — Plan {3}{B} (Marvel Super Heroes, #4455):
//
//	"Whenever one or more creature cards are put into your graveyard from
//	 anywhere, you draw a card, lose 1 life, and put a plan counter on this
//	 enchantment.
//	 When the third plan counter is put on this enchantment, sacrifice it and
//	 create three 2/2 colorless Robot Villain artifact creature tokens."
//
// Dennick's "creature card put into a graveyard from anywhere" watch (a
// death, a discard, a mill, a countered spell), narrowed to the source
// controller's own graveyard and collapsed to once per batch. The third
// counter's trigger is one instruction after another, not a reflexive: the
// tokens come even when the enchantment is already gone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4f69d2f9-c16a-4752-9026-794a29d73af8",
		Name:         "Robot Domination",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.OnAny(
				[]game.EventKind{game.EventZoneMove, game.EventDiscardCard, game.EventMill, game.EventCounterSpell},
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if !effects.CreatureCardPutIntoAGraveyard(ev, g) {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && c.Owner == source.Controller
				},
				"Robot Domination — you draw a card, lose 1 life, and put a plan counter on this enchantment",
				mshB7RobotDominationPlan)),
			effects.On(game.EventCounterPlaced, effects.MshB2PlanCounterReached(3),
				"Robot Domination — sacrifice it and create three 2/2 colorless Robot Villain artifact creature tokens",
				mshB7RobotDominationFinish),
		},
	})
}

func mshB7RobotDominationPlan(g *game.Game, item *game.StackItem) error {
	if err := effects.B36DrawAndLoseOne(g, item); err != nil {
		return err
	}
	return effects.AddCounter{Target: item.SourceCardID, Kind: "plan", N: 1}.Apply(effects.NewContext(g, item))
}

func mshB7RobotDominationFinish(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if effects.OnBattlefield(g, item.SourceCardID) && !ctx.IsNewSourceObjectAsThis(item.SourceCardID) {
		if err := (effects.SacrificePermanent{Target: item.SourceCardID}).Apply(ctx); err != nil {
			return err
		}
	}
	return effects.CreateToken{Template: effects.TokenCard("2/2 colorless Robot Villain artifact"), N: 3}.Apply(ctx)
}
