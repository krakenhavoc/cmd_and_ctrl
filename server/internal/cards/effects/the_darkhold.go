package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Darkhold — Legendary Artifact — Book {1}{B}{B}:
//
//	"{T}: Put a sin counter on The Darkhold. Draw a card for each sin
//	 counter on it, then you lose 1 life for each card in your hand.
//	 {6}{B}{B}, {T}: Each player's life total becomes the lowest life
//	 total among all players. Activate only as a sorcery."
//
// The first ability is three steps in printed order, each read as it
// happens: the counter goes on, the draw is the number of sin counters
// on The Darkhold now (including the one just placed), and the life loss
// counts the cards in the controller's hand AFTER those draws. The
// later steps wait for any prompt the earlier ones ask (CR 608.2c).
//
// The second ability finds the lowest life total among the players
// still in the game, then sets every one of them to it, each through
// the life-change path (CR 119.5) so a player already at that total
// changes nothing and any other sees a life gain.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3981970f-b332-49b7-9698-3296830ac399",
		Name:         "The Darkhold",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{T}: Put a sin counter on The Darkhold. Draw a card for each sin counter on it, then you lose 1 life for each card in your hand.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    TapCost(),
				Effect: Do(
					Instruction(func(ctx *Context) error {
						return AddCounter{Target: ctx.Item.SourceCardID, Kind: "sin", N: 1}.Apply(ctx)
					}),
					Instruction(func(ctx *Context) error {
						c, ok := ctx.Game.LookupCardForEffect(ctx.Item.SourceCardID)
						if !ok || c.Counters["sin"] <= 0 {
							return nil
						}
						return DrawCards{N: c.Counters["sin"]}.Apply(ctx)
					}),
					Instruction(func(ctx *Context) error {
						p := ctx.PlayerByID(ctx.Controller())
						if p == nil || p.Hand == nil || p.Hand.Size() == 0 {
							return nil
						}
						return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), p.ID, -p.Hand.Size())
					}),
				),
			},
			{
				Label:        "{6}{B}{B}, {T}: Each player's life total becomes the lowest life total among all players. Activate only as a sorcery.",
				Cost:         Plus(ManaCost("{6}{B}{B}"), TapCost()),
				SorcerySpeed: true,
				Effect:       darkholdLowestLifeBecomes,
			},
		},
	})
}

// darkholdLowestLifeBecomes is the second ability: every player still in
// the game has their life total become the lowest among them (CR 119.5).
func darkholdLowestLifeBecomes(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	lowest, found := 0, false
	for _, p := range g.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		if !found || p.Life < lowest {
			lowest, found = p.Life, true
		}
	}
	if !found {
		return nil
	}
	for _, p := range g.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		if err := b31LifeBecomes(ctx, p.ID, lowest); err != nil {
			return err
		}
	}
	return nil
}
