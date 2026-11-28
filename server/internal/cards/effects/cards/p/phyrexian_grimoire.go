package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Phyrexian Grimoire — Artifact — Book {3} (Tempest, S82 OLD b752, #5628):
//
//	"{4}, {T}: Target opponent chooses one of the top two cards of your graveyard.
//	 Exile that card and put the other one into your hand."
//
// The opponent is a target, and it is that opponent, not the Grimoire's controller,
// who picks (Burning-Rune Demon's ChooseCards prompt, chooser the other player):
// the top two cards of the controller's graveyard as the ability resolves
// (CR 608.2h), the topmost the last in the pile. The pick is exiled and the other
// goes to its owner's, the controller's, hand. With one card there is nothing to
// choose: it is the one chosen, so it is exiled; with none nothing happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "090052a8-9e5b-4024-881f-5f8b65db7aa7",
		Name:         "Phyrexian Grimoire",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}: Target opponent chooses one of the top two cards of your graveyard. Exile that card and put the other one into your hand.",
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				opp := effects.TargetedPlayerOf(ctx)
				p := g.PlayerByIDForEffect(item.Controller)
				if opp == uuid.Nil || p == nil || p.Graveyard == nil || len(p.Graveyard.Cards) == 0 {
					return nil
				}
				cards := p.Graveyard.Cards
				var top []uuid.UUID
				for i := len(cards) - 1; i >= 0 && len(top) < 2; i-- {
					top = append(top, cards[i].InstanceID)
				}
				if len(top) == 1 {
					return phyrexianGrimoireSettle(ctx, top[0], nil)
				}
				g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
					Chooser:    opp,
					FromPlayer: item.Controller,
					Source:     ctx.Source(),
					Question:   "Phyrexian Grimoire — choose one of the top two cards of their graveyard: it is exiled, the other goes to their hand",
					Cards:      top,
					Min:        1,
					Max:        1,
					ThenKey:    phyrexianGrimoirePicked,
					Carry:      top,
					LiveZone:   game.ZoneGraveyard,
					Asked:      game.NewAskedItem(item),
				})
				return nil
			},
		}},
	})
}

// phyrexianGrimoirePicked exiles the chosen card and puts the other into the
// controller's hand. An on-disk key.
var phyrexianGrimoirePicked = effects.AskedPick("choose-cards/s82-old-b752-phyrexian-grimoire-picked", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) != 1 {
		return nil
	}
	var rest []uuid.UUID
	for _, id := range r.Carry {
		if id != r.Picked[0] {
			rest = append(rest, id)
		}
	}
	return phyrexianGrimoireSettle(ctx, r.Picked[0], rest)
})

// phyrexianGrimoireSettle returns each other card to the controller's hand and
// exiles the chosen card, each only if it is still in the graveyard.
func phyrexianGrimoireSettle(ctx *effects.Context, chosen uuid.UUID, rest []uuid.UUID) error {
	inGraveyard := func(id uuid.UUID) bool {
		z := ctx.Game.FindCardZoneForEffect(id)
		return z != nil && z.Kind == game.ZoneGraveyard
	}
	// The other cards go to the hand first and the exile is the last thing done,
	// so nothing is gated on a value read before a fire-and-forget exit.
	for _, id := range rest {
		if inGraveyard(id) {
			if err := (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	if inGraveyard(chosen) {
		return effects.ExileTarget{Target: chosen}.Apply(ctx)
	}
	return nil
}
