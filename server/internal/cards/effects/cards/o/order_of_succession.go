package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Order of Succession — Sorcery {3}{U} (Commander 2013, S82 OLD b387):
//
//	"Choose left or right. Starting with you and proceeding in the chosen
//	 direction, each player chooses a creature controlled by the next player in
//	 that direction. Each player gains control of the creature they chose."
//
// The caster chooses the direction (Teyo's left or right: left is the next
// player in turn order, a seat that has left the game is skipped). The choices
// are then one prompt per player, in that order, each from the creatures the
// NEXT player in the direction controls (a player whose next player controls
// none chooses nothing); the picks so far ride the prompt as data. Only once
// every player has chosen does anyone gain control, so a creature chosen early
// is not yet moved when a later player chooses (a creature can be chosen by one
// player only, since each pool is a different player's). Control has no
// duration (CR 611.2a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1b95970c-e7eb-41c4-a8d2-9889b64b3c63",
		Name:         "Order of Succession",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			options := make([]game.ChoiceOption, 0, 2)
			for _, w := range []string{game.AttackDirectionLeftWord, game.AttackDirectionRightWord} {
				options = append(options, game.ChoiceOption{Label: w})
			}
			return effects.PickOption{
				Question: "Order of Succession — choose left or right (left is the next player in turn order, right the player before)",
				Options:  options,
				ThenKey:  orderOfSuccessionDirection,
			}.Apply(ctx)
		},
	})
}

// orderOfSuccessionDirection starts the round of picks. An on-disk key.
var orderOfSuccessionDirection = effects.OptionPickThen("option-pick/old-b387-order-of-succession-direction", func(ctx *effects.Context, r game.OptionPicked) error {
	if r.Option == nil {
		return nil
	}
	return orderOfSuccessionAsk(ctx, r.Option.Label, nil)
})

// orderOfSuccessionPlayers is the players in the order they choose: the
// caster, then the next player in the direction, and so on round the table.
func orderOfSuccessionPlayers(g *game.Game, start uuid.UUID, dir game.AttackDirection) []uuid.UUID {
	order := []uuid.UUID{start}
	cur := start
	for range g.Seats {
		next := g.NearestOpponentForEffect(cur, dir)
		if next == uuid.Nil || next == start {
			break
		}
		order = append(order, next)
		cur = next
	}
	return order
}

// orderOfSuccessionAsk asks the next player who has a creature to choose from,
// or, with every pick made, hands the creatures over. picks line up with the
// players: one id per player so far, uuid.Nil for a player with no choice.
func orderOfSuccessionAsk(ctx *effects.Context, word string, picks []uuid.UUID) error {
	g := ctx.Game
	dir := game.AttackDirectionOfWord(word)
	order := orderOfSuccessionPlayers(g, ctx.Controller(), dir)
	for k := len(picks); k < len(order); k++ {
		next := g.NearestOpponentForEffect(order[k], dir)
		var ids []uuid.UUID
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == next && c.IsCreature() {
				ids = append(ids, c.InstanceID)
			}
		}
		if next == uuid.Nil || len(ids) == 0 {
			picks = append(picks, uuid.Nil)
			continue
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:    order[k],
			FromPlayer: next,
			Source:     ctx.Source(),
			Question:   "Order of Succession — choose a creature controlled by the next player to gain control of",
			Cards:      ids,
			Min:        1,
			Max:        1,
			ThenKey:    orderOfSuccessionPick,
			Asked:      game.NewAskedItem(ctx.Item),
			Carry:      append([]uuid.UUID(nil), picks...),
			CarryWord:  word,
		})
		return nil
	}
	for i, id := range picks {
		if id == uuid.Nil || i >= len(order) {
			continue
		}
		if err := (effects.GainControl{
			Target:     id,
			Controller: order[i],
			Duration:   g.PinnedTo(game.IndefiniteDuration(), id),
			Label:      "Order of Succession — gain control",
		}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	return nil
}

// orderOfSuccessionPick records one player's pick and asks the next. An
// on-disk key. Registered from init, because the pick and the asking function
// refer to each other.
var orderOfSuccessionPick game.ChooseCardsThen

func init() {
	orderOfSuccessionPick = effects.AskedPick("choose-cards/old-b387-order-of-succession-pick", func(ctx *effects.Context, r game.CardsChosen) error {
		pick := uuid.Nil
		if len(r.Picked) == 1 {
			pick = r.Picked[0]
		}
		return orderOfSuccessionAsk(ctx, r.Word, append(append([]uuid.UUID(nil), r.Carry...), pick))
	})
}
