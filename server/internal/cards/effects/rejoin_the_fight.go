package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rejoin the Fight — Sorcery {5}{B} (deck request #3655):
//
//	"Mill three cards. Then starting with the next opponent in turn
//	 order, each opponent chooses a creature card in your graveyard that
//	 hasn't been chosen. Return each card chosen this way to the
//	 battlefield under your control."
//
// The mill finishes first, so the milled creatures are choosable. Each
// opponent is asked in turn order starting after the caster, and each
// pick is hidden from none: the cards chosen so far are no longer
// offered. An opponent who has no creature card left to choose does not
// choose. Nothing returns until the last opponent has chosen, then the
// chosen cards enter together under the caster's control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d22ea32b-2128-4dcc-8b1c-c4c52f0b2f45",
		Name:         "Rejoin the Fight",
		Completeness: CompletenessFull,
		// The opponents' choices wait for a prompt the mill asks (CR 608.2c, #4155).
		OnResolve: ResolveSteps(
			MillCards{N: 3},
			Instruction(func(ctx *Context) error {
				return rejoinAsk(ctx, opponentsInTurnOrder(ctx.Game, ctx.Controller()), nil)
			}),
		),
	})
}

// rejoinAsk asks the next opponent to choose, then continues; once the
// list is empty the chosen cards return under the caster's control.
func rejoinAsk(ctx *Context, opponents, chosen []uuid.UUID) error {
	controller := ctx.Controller()
	if len(opponents) == 0 {
		for _, id := range chosen {
			if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
				continue
			}
			if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield, Controller: controller}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}
	chooser, rest := opponents[0], opponents[1:]
	var candidates []uuid.UUID
	if p := ctx.Game.PlayerByIDForEffect(controller); p != nil && p.Graveyard != nil {
		for _, c := range p.Graveyard.Cards {
			if c.IsCreature() && !slices.Contains(chosen, c.InstanceID) {
				candidates = append(candidates, c.InstanceID)
			}
		}
	}
	if len(candidates) == 0 {
		return rejoinAsk(ctx, rest, chosen)
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    chooser,
		FromPlayer: controller,
		Source:     ctx.Source(),
		Question:   "Rejoin the Fight — choose a creature card in their graveyard that hasn't been chosen",
		Cards:      candidates,
		Min:        1,
		Max:        1,
		LiveZone:   game.ZoneGraveyard,
		ThenKey:    rejoinPick,
		// The opponents still to ask and the cards chosen so far are the
		// carry; the item is the caster's spell (#4253).
		Carry: packIDLists(rest, chosen),
		Asked: game.NewAskedItem(ctx.Item),
	})
	return nil
}

// rejoinPick is one opponent's choice answered, keyed (#4253): the next
// opponent is asked, or the chosen cards return. Registered in init because
// it asks the next opponent, which names the key again.
var rejoinPick game.ChooseCardsThen

func init() {
	rejoinPick = askedPick("choose-cards/rejoin-the-fight-pick", func(ctx *Context, r game.CardsChosen) error {
		rest, chosen := unpackIDPair(r.Carry)
		return rejoinAsk(ctx, rest, append(append([]uuid.UUID(nil), chosen...), r.Picked...))
	})
}
