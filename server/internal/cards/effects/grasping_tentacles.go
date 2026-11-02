package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grasping Tentacles — {1}{U}{B} Sorcery (Marvel's Spider-Man Eternal,
// #4460):
//
//	"Target opponent mills eight cards. You may put an artifact card from
//	 that player's graveyard onto the battlefield under your control."
//
// Hama, the Bloodbender's shape: the mill settles first (its continuation is
// a keyed tail carrying the milled player), then the caster is offered every
// artifact card in that player's graveyard, the milled ones included, with
// declining always legal. The pick is re-read when it is answered and the
// card enters under the caster's control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dac76b28-23e8-4138-8442-51579a590f16",
		Name:         "Grasping Tentacles",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			opp := TargetedPlayer(ctx)
			if ctx.Game.PlayerByIDForEffect(opp) == nil {
				return nil
			}
			tail := ctx.ItemTail(graspingTentaclesMilled)
			tail.Carry = []uuid.UUID{opp}
			return ctx.Game.MillToZoneKeyedForEffect(opp, 8, game.ZoneGraveyard, game.MillUntil{}, tail)
		},
	})
}

var (
	graspingTentaclesMilled = CardTail("grasping-tentacles/choose-artifact", func(ctx *Context, o game.TailOutcome) error {
		if len(o.Tail.Carry) == 0 {
			return nil
		}
		return graspingTentaclesOffer(ctx.Game, ctx.Item, o.Tail.Carry[0])
	})
	graspingTentaclesPicked = game.RegisterChooseCardsThen("grasping-tentacles/take-pick", func(g *game.Game, r game.CardsChosen) error {
		if len(r.Picked) == 0 || r.Asked == nil {
			return nil
		}
		id := r.Picked[0]
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			return nil
		}
		return ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield, Controller: r.Asked.Controller}.Apply(NewContext(g, r.Asked))
	})
)

func graspingTentaclesOffer(g *game.Game, item *game.StackItem, opp uuid.UUID) error {
	p := g.PlayerByIDForEffect(opp)
	if p == nil || p.Graveyard == nil {
		return nil
	}
	var candidates []uuid.UUID
	for _, c := range p.Graveyard.Cards {
		if c.HasCardType("artifact") {
			candidates = append(candidates, c.InstanceID)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    item.Controller,
		FromPlayer: opp,
		Source:     item.SourceCardID,
		Question:   "Grasping Tentacles — you may put an artifact card from that player's graveyard onto the battlefield under your control",
		Cards:      candidates,
		Min:        0,
		Max:        1,
		Zone:       game.ZoneGraveyard,
		ThenKey:    graspingTentaclesPicked,
		Asked:      game.NewAskedItem(item),
		LiveZone:   game.ZoneGraveyard,
	})
	return nil
}
