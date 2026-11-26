package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jaddi Lifestrider — Creature — Elemental {4}{G}, 2/8 (Rise of the Eldrazi,
// S82 OLD batch 481, #5581):
//
//	"When this creature enters, you may tap any number of untapped creatures
//	 you control. You gain 2 life for each creature tapped this way."
//
// Orphans of the Wheat's ask: the controller picks any number of their
// untapped creatures (none is allowed) and the life is counted from the
// creatures that really became tapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9c8027b7-b823-4214-936c-91b7a8986228",
		Name:         "Jaddi Lifestrider",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Jaddi Lifestrider — tap any number of untapped creatures you control; gain 2 life for each",
				jaddiLifestriderAsk),
		},
	})
}

// jaddiLifestriderAsk offers the controller's untapped creatures.
func jaddiLifestriderAsk(g *game.Game, item *game.StackItem) error {
	candidates := effects.UntappedCreaturesControlledBy(g, item.Controller)
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Jaddi Lifestrider — tap any number of untapped creatures you control",
		Cards:    candidates,
		Min:      0,
		// Max 0 means "all of them", which is what "any number" is.
		Max:      0,
		LiveZone: game.ZoneBattlefield,
		ThenKey:  jaddiLifestriderTap,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// jaddiLifestriderTap taps the picks and gains 2 life for each that really
// became tapped. An on-disk key: never renamed, never reused.
var jaddiLifestriderTap = effects.AskedPick("choose-cards/jaddi-lifestrider-tap", func(ctx *effects.Context, r game.CardsChosen) error {
	n := 0
	for _, id := range r.Picked {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || c.Tapped || !c.IsCreature() || c.Controller != ctx.Controller() {
			continue
		}
		if err := (effects.TapTarget{Target: id}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
		n++
	}
	if n == 0 {
		return nil
	}
	return effects.GainLife{Amount: 2 * n}.Apply(ctx)
})
