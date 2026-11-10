package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Minamo — Plane — Kamigawa (Planechase Anthology, #4870):
//
//	"Whenever a player casts a spell, that player may draw a card.
//	 Whenever chaos ensues, each player may return a blue card from their
//	 graveyard to their hand."
//
// A plane (ADR 0216 §6). The first ability is the caster's own "may" and
// resolves before the spell (the ruling). The chaos ability does not
// target: starting with the active player and going in turn order, each
// player chooses a blue card in their own graveyard or none (a keyed option
// pick, so a table waiting on it is a restore point), and only after the
// last answer do all the chosen cards go to their owners' hands together.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6491f511-87e6-4504-84f5-fe5cf44768d7",
		Name:         "Minamo",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, planesR6APlayerCastsASpell,
				"Minamo — that player may draw a card",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return MayChoice{
						Player:   ctx.Trigger().Event.Actor,
						Question: "Minamo — draw a card?",
						Draws:    1,
						Word:     "1",
						Then:     mayDrawCards,
					}.Apply(ctx)
				}),
			WheneverChaosEnsues("Minamo — each player may return a blue card from their graveyard to their hand",
				func(g *game.Game, item *game.StackItem) error {
					return minamoAsk(NewContext(g, item), apnapPlayers(g), nil)
				}),
		},
	})
}

// minamoBlueCards is the blue cards in `player`'s graveyard, in graveyard
// order.
func minamoBlueCards(g *game.Game, player uuid.UUID) []game.Card {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return nil
	}
	var out []game.Card
	for _, c := range p.Graveyard.Cards {
		if c.HasColor("U") {
			out = append(out, c)
		}
	}
	return out
}

// minamoAsk asks the first player still to choose; once all have, the
// chosen cards return together. The prompt's Carry is the cards chosen so
// far then the players still to ask, Nums[0] how many of the former.
func minamoAsk(ctx *Context, order, chosen []uuid.UUID) error {
	for len(order) > 0 {
		who, rest := order[0], order[1:]
		blue := minamoBlueCards(ctx.Game, who)
		if len(blue) == 0 {
			order = rest
			continue
		}
		opts := []game.ChoiceOption{{Label: "Don't return a card"}}
		for _, c := range blue {
			opts = append(opts, game.ChoiceOption{Label: fmt.Sprintf("Return %s to your hand", c.Name), Cards: []uuid.UUID{c.InstanceID}})
		}
		return PickOption{
			Player:   who,
			From:     who,
			Question: "Minamo — return a blue card from your graveyard to your hand?",
			Options:  opts,
			ThenKey:  minamoAnswered,
			Carry:    append(append([]uuid.UUID(nil), chosen...), rest...),
			Nums:     []int{len(chosen)},
			WithItem: true,
		}.Apply(ctx)
	}
	for _, id := range chosen {
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// minamoAnswered records one player's choice and asks the next. An on-disk
// key: never renamed, never reused.
var minamoAnswered game.OptionPickThen

func init() { minamoAnswered = OptionPickThen("option-pick/minamo-return-blue-card", minamoAnswer) }

func minamoAnswer(ctx *Context, r game.OptionPicked) error {
	n := 0
	if len(r.Nums) > 0 {
		n = min(r.Nums[0], len(r.Carry))
	}
	chosen, rest := append([]uuid.UUID(nil), r.Carry[:n]...), r.Carry[n:]
	if r.Option != nil && r.Index > 0 && len(r.Option.Cards) == 1 {
		chosen = append(chosen, r.Option.Cards[0])
	}
	return minamoAsk(ctx, rest, chosen)
}
