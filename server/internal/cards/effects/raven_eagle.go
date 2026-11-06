package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Raven Eagle — Creature — Bird Assassin {2}{B}, 2/3 (Avatar: The Last
// Airbender, #4459):
//
//	"Flying
//	 Whenever this creature enters or attacks, exile up to one target
//	 card from a graveyard. If a creature card is exiled this way,
//	 create a Clue token.
//	 Whenever you draw your second card each turn, each opponent loses
//	 1 life and you gain 1 life."
//
// The first line is one ability with two conditions (Sun Titan's
// WhenThisEntersOrAttacks) on an "up to one" target clause: any
// graveyard, any card, or no target at all. Whether the card is a
// creature card is read before the exile, and the Clue waits for the
// exile and is gated on it (ExileThenIfItWas): a card a replacement kept
// out of exile was not "exiled this way". Making the Clue is not
// investigating.
//
// The second is Thopter Fabricator's "your second card each turn"
// (b41DrewYourSecondCardThisTurn): once a turn, whether or not the Eagle
// was on the battlefield for the first draw, with a footprinted drain.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f75ac0a3-8795-48b4-91fd-4717900ef5b7",
		Name:            "Raven Eagle",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEntersOrAttacks("Raven Eagle — exile up to one target card from a graveyard; a creature card makes a Clue",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							if t.Kind == game.TargetCard {
								return ExileThenIfItWas{Target: t.ID, Was: WasCreatureCard, Keyed: ravenEagleIfCreature}.Apply(ctx)
							}
						}
						return nil
					}),
				TargetCardInGraveyard("up to one target card from a graveyard").WithCount(0, 1),
			),
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Raven Eagle — each opponent loses 1 life and you gain 1 life",
				Do(EachOpponentLosesLife{N: 1}, GainLife{Amount: 1})),
		},
	})
}

// ravenEagleIfCreature is the exile's keyed "if a creature card is
// exiled this way, create a Clue token". An on-disk key: never renamed,
// never reused.
var ravenEagleIfCreature = IfItWasTail("if-it-was/raven-eagle-clue", func(ctx *Context) error {
	return CreateToken{Controller: ctx.Controller(), Template: ClueToken(), N: 1}.Apply(ctx)
}, nil)
