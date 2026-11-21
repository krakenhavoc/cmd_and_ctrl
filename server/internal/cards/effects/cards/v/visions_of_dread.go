package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Visions of Dread — Sorcery {2}{B} (Midnight Hunt Commander):
//
//	"Target opponent puts a creature card of their choice from their
//	 graveyard onto the battlefield under your control.
//	 Flashback {8}{B}{B}. This spell costs {X} less to cast this way, where X
//	 is the greatest mana value of a commander you own on the battlefield or
//	 in the command zone."
//
// The TARGETED OPPONENT chooses (a ChooseCards prompt whose Chooser is that
// player, over their own graveyard), and the card enters under the spell
// controller's control. The choice is mandatory when they have a creature
// card; a spell whose target is gone does nothing (CR 608.2b). A card that
// left the graveyard before the choice settled is not put. The flashback
// discount is the Visions cycle's (visionsFlashbackDiscount).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:          "0b12fee3-1ad1-4dc3-bd19-e9ccce5cb829",
		Name:              "Visions of Dread",
		Completeness:      effects.CompletenessFull,
		CastableZones:     []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts:  []game.AlternativeCost{effects.VisionsFlashbackAlternative("{8}{B}{B}")},
		SelfCostModifiers: []game.CostModifier{effects.VisionsFlashbackDiscount()},
		Targets:           effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			legal := ctx.LegalTargets()
			if len(legal) == 0 || legal[0].Kind != game.TargetPlayer {
				return nil
			}
			opp := legal[0].ID
			p := ctx.Game.PlayerByIDForEffect(opp)
			if p == nil || p.Graveyard == nil {
				return nil
			}
			var cards []uuid.UUID
			for _, c := range p.Graveyard.Cards {
				if c.IsCreature() {
					cards = append(cards, c.InstanceID)
				}
			}
			if len(cards) == 0 {
				return nil
			}
			ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:    opp,
				FromPlayer: opp,
				Source:     item.SourceCardID,
				Question:   "Visions of Dread — choose a creature card in your graveyard; it enters the battlefield under their control",
				Cards:      cards,
				Min:        1,
				Max:        1,
				Zone:       game.ZoneGraveyard,
				ThenKey:    visionsOfDreadPicked,
				Asked:      game.NewAskedItem(item),
				LiveZone:   game.ZoneGraveyard,
			})
			return nil
		},
	})
}

// visionsOfDreadPicked puts the chosen card onto the battlefield under the
// spell controller's control. An on-disk key: never renamed, never reused.
var visionsOfDreadPicked = game.RegisterChooseCardsThen("choose-cards/visions-of-dread", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 || r.Asked == nil {
		return nil
	}
	z := g.FindCardZoneForEffect(r.Picked[0])
	if z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return g.ReturnFromGraveyardWithCountersKeyedForEffect(r.Picked[0], r.Asked.Controller, false, nil, game.KeyedTail{})
})
