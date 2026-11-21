package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Weaver King — Legendary Creature — Horror Wizard {4}{U}{B}, 3/6
// (Mystery Booster Commander Edition):
//
//	"Shadow (This creature can block or be blocked by only creatures with
//	 shadow.)
//	 Whenever The Weaver King deals combat damage to a player, they mill
//	 that many cards. Then for each opponent, put a creature card from that
//	 player's graveyard that was put there from their library this turn onto
//	 the battlefield under your control."
//
// The damaged player mills as many cards as the combat damage dealt. Then
// each opponent of the controller, in turn order, gives the controller one
// pick: a creature card in that opponent's graveyard whose latest arrival
// this turn came out of their library (Captain N'ghathrod's provenance
// reading, so a creature that died or was discarded is not offered, and a
// card that was in the graveyard when the turn began is not either). The
// pick is made as the ability resolves, is mandatory, and an opponent with
// no such card is skipped. The creatures enter under the controller's
// control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ed1a6d06-436b-45c0-979c-7802246eeb9e",
		Name:            "The Weaver King",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"shadow"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("The Weaver King — that player mills that many cards; then for each opponent, put a creature card from that player's graveyard that was put there from their library this turn onto the battlefield under your control",
				effects.Do(effects.Instruction(weaverKingMill), effects.Instruction(func(ctx *effects.Context) error {
					return weaverKingAsk(ctx.Game, ctx.Controller(), ctx.Source(), weaverKingOpponents(ctx))
				}))),
		},
	})
}

// weaverKingOpponents is the controller's opponents still in the game, in
// turn order.
func weaverKingOpponents(ctx *effects.Context) []uuid.UUID {
	var out []uuid.UUID
	for _, opp := range ctx.Opponents() {
		if p := ctx.Game.PlayerByIDForEffect(opp); p != nil && !p.Eliminated {
			out = append(out, opp)
		}
	}
	return out
}

func weaverKingMill(ctx *effects.Context) error {
	ev := ctx.Trigger().Event
	if ev.Amount <= 0 || ctx.Game.PlayerByIDForEffect(ev.Target) == nil {
		return nil
	}
	return effects.MillCards{Player: ev.Target, N: ev.Amount}.Apply(ctx)
}

// weaverKingAsk asks for the first opponent in `opps` with a creature card
// to take, carrying the opponents after it on the prompt.
func weaverKingAsk(g *game.Game, me, source uuid.UUID, opps []uuid.UUID) error {
	ctx := effects.NewContext(g, &game.StackItem{Kind: game.StackItemSpell, Controller: me, Owner: me, SourceCardID: source})
	for i, opp := range opps {
		cands := effects.GraveyardCardIDs(ctx, opp, func(c game.Card) bool {
			return c.IsCreature() && effects.NghaPutThereFromLibraryThisTurn(g, uuid.Nil, c)
		})
		if len(cands) == 0 {
			continue
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:    me,
			FromPlayer: opp,
			Source:     source,
			Question:   "The Weaver King — put a creature card from this graveyard that was put there from their library this turn onto the battlefield under your control",
			Cards:      cands,
			Min:        1,
			Max:        1,
			ThenKey:    weaverKingPicked,
			LiveZone:   game.ZoneGraveyard,
			Carry:      opps[i+1:],
		})
		return nil
	}
	return nil
}

// weaverKingPicked puts the pick onto the battlefield under the chooser's
// control, then asks about the next opponent. An on-disk key: never
// renamed, never reused.
var weaverKingPicked game.ChooseCardsThen

func init() {
	weaverKingPicked = game.RegisterChooseCardsThen("choose-cards/the-weaver-king-reanimate", weaverKingPick)
}

func weaverKingPick(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) > 0 {
		ctx := effects.NewContext(g, &game.StackItem{Kind: game.StackItemSpell, Controller: r.Chooser, Owner: r.Chooser, SourceCardID: r.Source})
		if err := (effects.ReturnFromGraveyardTogether{Targets: r.Picked, Controller: r.Chooser}).Apply(ctx); err != nil {
			return err
		}
	}
	return weaverKingAsk(g, r.Chooser, r.Source, r.Carry)
}
