package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sadistic Shell Game — Sorcery {4}{B} (Duskmourn Commander):
//
//	"Starting with the next opponent in turn order, each player chooses a
//	 creature you don't control. Destroy the chosen creatures."
//
// Every player still in the game chooses, the opponents in turn order
// from the one after the caster and the caster last. Each pick is
// independent, so two players may choose the same creature (nothing says
// the creature "hasn't been chosen"); a creature chosen twice is
// destroyed once. A player with no creature to choose does not choose.
// Nothing is destroyed until the last player has chosen.
func init() {
	Register(Spec{
		OracleID:     "88c3fc46-f2f5-440c-80a6-fed5b42cd340",
		Name:         "Sadistic Shell Game",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			order := append(opponentsInTurnOrder(ctx.Game, ctx.Controller()), ctx.Controller())
			return sadisticShellGameAsk(ctx, order, nil)
		},
	})
}

func sadisticShellGameAsk(ctx *Context, players, chosen []uuid.UUID) error {
	g, controller := ctx.Game, ctx.Controller()
	if len(players) == 0 {
		for _, id := range chosen {
			if err := (DestroyTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}
	chooser, rest := players[0], players[1:]
	var candidates []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && c.Controller != controller {
			candidates = append(candidates, c.InstanceID)
		}
	}
	if len(candidates) == 0 {
		return sadisticShellGameAsk(ctx, rest, chosen)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  chooser,
		Source:   ctx.Source(),
		Question: "Sadistic Shell Game — choose a creature the caster doesn't control (it will be destroyed)",
		Cards:    candidates,
		Min:      1,
		Max:      1,
		ThenKey:  sadisticShellGamePick,
		LiveZone: game.ZoneBattlefield,
		Carry:    packIDLists(rest, chosen),
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
}

// sadisticShellGamePick is one player's choice answered, keyed (#4253):
// the next player is asked, or the chosen creatures are destroyed.
var sadisticShellGamePick game.ChooseCardsThen

func init() {
	sadisticShellGamePick = askedPick("choose-cards/sadistic-shell-game-pick", func(ctx *Context, r game.CardsChosen) error {
		rest, chosen := unpackIDPair(r.Carry)
		chosen = append([]uuid.UUID(nil), chosen...)
		for _, id := range r.Picked {
			if !slices.Contains(chosen, id) {
				chosen = append(chosen, id)
			}
		}
		return sadisticShellGameAsk(ctx, rest, chosen)
	})
}
