package cardsd

import (
	"errors"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Death in Heaven (#6092): "I, II — Target player mills two cards, then
// exiles their graveyard. III — Put all creature cards exiled with this
// enchantment onto the battlefield face down under your control. They're
// 2/2 Cyberman artifact creatures."
//
// Chapters I and II exile with the link (CR 607.2a: the chapter ability
// exiles, chapter III refers to the cards "exiled with this enchantment"),
// so chapter III reads exactly the cards this Saga object exiled, whoever
// owns them. The return is game.ReturnFromExileFaceDownForEffect with the
// listed Cyberman body (CR 708.2).
func init() {
	const name = "Death in Heaven"
	millText := "target player mills two cards, then exiles their graveyard"
	effects.Register(effects.Spec{
		OracleID:     "f2bf5014-677a-4880-b926-d93457348bdc",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTriggerTargeting(1, effects.SagaChapterLabel(name, 1, millText), effects.TargetPlayer("target player"), deathInHeavenMillAndExile),
			effects.ChapterTriggerTargeting(2, effects.SagaChapterLabel(name, 2, millText), effects.TargetPlayer("target player"), deathInHeavenMillAndExile),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "put all creature cards exiled with this enchantment onto the battlefield face down under your control as 2/2 Cyberman artifact creatures"), deathInHeavenReturnCybermen),
		},
	})
}

func deathInHeavenMillAndExile(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	p := effects.TargetedPlayerOf(ctx)
	if p == uuid.Nil {
		return nil
	}
	if err := (effects.MillCards{Player: p, N: 2}).Apply(ctx); err != nil {
		return err
	}
	player := g.PlayerByIDForEffect(p)
	if player == nil {
		return nil
	}
	ids := make([]game.Card, 0, len(player.Graveyard.Cards))
	ids = append(ids, player.Graveyard.Cards...)
	for _, c := range ids {
		if err := (effects.ExileTarget{Target: c.InstanceID, Linked: true}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

func deathInHeavenReturnCybermen(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	src, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var creatures []game.Card
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(src)) {
		if c.IsCreature() {
			creatures = append(creatures, *c)
		}
	}
	for _, c := range creatures {
		if _, err := g.ReturnFromExileFaceDownForEffect(c.InstanceID, ctx.Controller(), false, effects.CybermanBody()); err != nil && !errors.Is(err, game.ErrCardNotFound) {
			return err
		}
	}
	return nil
}
