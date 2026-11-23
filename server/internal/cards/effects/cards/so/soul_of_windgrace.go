package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soul of Windgrace — Legendary Creature — Cat Avatar {1}{B}{R}{G}, 5/4
// (Foundations Commander):
//
//	"Whenever Soul of Windgrace enters or attacks, you may put a land
//	 card from a graveyard onto the battlefield tapped under your control.
//	 {G}, Discard a land card: You gain 3 life.
//	 {1}{R}, Discard a land card: Draw a card.
//	 {2}{B}, Discard a land card: Soul of Windgrace gains indestructible
//	 until end of turn. Tap it."
//
// One enters-or-attacks trigger (Sun Titan's shape) whose land is chosen as
// it resolves from every graveyard (Extract from Darkness's pick, with a
// floor of zero for the "you may") and enters tapped under the controller's
// control. The three activated abilities each discard a land card as a
// cost; the last grants indestructible and then taps the Soul, which works
// when it is already tapped (the ruling).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2d73a3bd-f052-4f48-9a72-12db03e27249",
		Name:         "Soul of Windgrace",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrAttacks("Soul of Windgrace — you may put a land card from a graveyard onto the battlefield tapped under your control",
				soulOfWindgraceTrigger),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{G}, Discard a land card: You gain 3 life.",
				Purpose: game.Purpose{LifeGain: 3, Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{G}"), effects.DiscardCardsMatching(1, "a land card", game.Card.IsLand)),
				Effect:  effects.Do(effects.GainLife{Amount: 3}),
			},
			{
				Label:   "{1}{R}, Discard a land card: Draw a card.",
				Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{1}{R}"), effects.DiscardCardsMatching(1, "a land card", game.Card.IsLand)),
				Effect:  effects.Do(effects.DrawCards{N: 1}),
			},
			{
				Label:   "{2}{B}, Discard a land card: Soul of Windgrace gains indestructible until end of turn. Tap it.",
				Purpose: game.Purpose{Answers: game.AnswerProtect},
				Cost:    effects.Plus(effects.ManaCost("{2}{B}"), effects.DiscardCardsMatching(1, "a land card", game.Card.IsLand)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"indestructible"}, Label: "Soul of Windgrace — indestructible"}).Apply(ctx); err != nil {
						return err
					}
					return effects.TapTarget{Target: item.SourceCardID}.Apply(ctx)
				},
			},
		},
	})
}

func soulOfWindgraceTrigger(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	candidates := effects.AllGraveyardsCardIDs(ctx, game.Card.IsLand)
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Soul of Windgrace — you may put a land card from a graveyard onto the battlefield tapped under your control",
		Cards:    candidates,
		Min:      0,
		Max:      1,
		ThenKey:  soulOfWindgracePicked,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// soulOfWindgracePicked puts the chosen land card onto the battlefield
// tapped under the controller's control. An on-disk key.
var soulOfWindgracePicked = effects.ItemChoice("choose-cards/soul-of-windgrace-land", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return effects.ReturnFromGraveyard{Target: r.Picked[0], Dest: game.ZoneBattlefield, Controller: ctx.Controller(), Tapped: true}.Apply(ctx)
})
