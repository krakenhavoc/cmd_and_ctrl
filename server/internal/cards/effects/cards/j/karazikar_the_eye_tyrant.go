package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Karazikar, the Eye Tyrant — Legendary Creature — Beholder {3}{B}{R}, 5/5
// (Adventures in the Forgotten Realms):
//
//	"Whenever you attack a player, tap target creature that player controls
//	 and goad it.
//	 Whenever an opponent attacks another one of your opponents, you and the
//	 attacking player each draw a card and lose 1 life."
//
// Both lines trigger once per player attacked in a declaration (CR 603.2c,
// OncePerBatchPerPlayer, the Horizon Explorer and Breena shapes); an attack
// on a planeswalker or battle is not an attack on a player. The first
// targets a creature the attacked player controls (ControlledByThatPlayer,
// judged against the attack event), taps it and goads it for its controller
// (Alela's goad, CR 701.15). The second fires for an opponent of Karazikar's
// controller who attacks a different opponent of theirs: neither Karazikar's
// controller nor the attacker may be the one attacked. Both then draw, then
// both lose 1 life. No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "47fde936-be71-4550-826b-b98d139cc57a",
		Name:         "Karazikar, the Eye Tyrant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatchPerPlayer(effects.Targeting(
				effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.B16YouAttackedAPlayer(ev, source, g)
				}, "Karazikar, the Eye Tyrant — tap target creature that player controls and goad it", karazikarTapAndGoad),
				effects.ControlledByThatPlayer(effects.TargetCreature("target creature that player controls")))),
			effects.OncePerBatchPerPlayer(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return karazikarOpponentAttacksAnotherOpponent(g, ev, source.Controller)
			}, "Karazikar, the Eye Tyrant — you and the attacking player each draw a card and lose 1 life", karazikarDrawAndLose)),
		},
	})
}

// karazikarOpponentAttacksAnotherOpponent is the second line's condition:
// the attacker is an opponent of `you`, and the player attacked is neither
// `you` nor that attacker.
func karazikarOpponentAttacksAnotherOpponent(g *game.Game, ev game.Event, you uuid.UUID) bool {
	defender := effects.AttackedPlayer(g, ev)
	return defender != uuid.Nil && ev.Actor != uuid.Nil && ev.Actor != you && defender != you && defender != ev.Actor
}

func karazikarTapAndGoad(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok || !effects.OnBattlefield(g, id) {
		return nil
	}
	if err := (effects.TapTarget{Target: id}).Apply(ctx); err != nil {
		return err
	}
	return effects.GoadAndScheduleClear(ctx, []uuid.UUID{id})
}

func karazikarDrawAndLose(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ctx := effects.NewContext(g, item)
	attacker := item.Trigger.Event.Actor
	players := []uuid.UUID{item.Controller, attacker}
	for _, p := range players {
		if err := (effects.DrawCards{Player: p, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	for _, p := range players {
		if g.PlayerByIDForEffect(p) == nil {
			continue
		}
		if err := g.ChangePlayerLifeForEffect(ctx.Source(), p, -1); err != nil {
			return err
		}
	}
	return nil
}
