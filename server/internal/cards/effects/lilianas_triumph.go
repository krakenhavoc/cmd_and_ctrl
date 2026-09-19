package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Liliana's Triumph — Instant {1}{B} (EDHREC rank 4405):
//
//	"Each opponent sacrifices a creature of their choice. If you
//	 control a Liliana planeswalker, each opponent also discards a
//	 card."
//
// A two-mana instant-speed edict that hits the whole table. In
// Commander the edict half is the half that matters — it answers a
// hexproof or indestructible commander that nothing else in black
// can touch, as long as it is the only creature its controller has.
// Roadmap batch 42 (#449), "no new machinery".
//
// "Of their choice" is the printed wording and the engine's default:
// each opponent gets their own prompt and picks their own creature,
// so the Triumph never reaches past the worst creature somebody is
// willing to lose. That is also why it is not removal you can aim.
//
// The Liliana clause is checked ON RESOLUTION, not on cast (CR
// 608.2 — an intervening-"if" clause would be checked twice, but this
// is a plain conditional inside the effect and is checked once, when
// the effect happens). A Liliana that died in response takes the
// discard with her.
//
// "A Liliana planeswalker" is the planeswalker TYPE, read off the
// layered subtypes, not the card name: every Liliana card that is a
// planeswalker has the type, and a card merely named Liliana that is
// not a planeswalker does not count.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c2e68ac5-cfff-4b76-8062-25a82fdf9a5c",
		Name:         "Liliana's Triumph",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (EachPlayerSacrifices{
				ExceptController: true,
				Match:            Creature(),
				Label:            "a creature",
			}).Apply(ctx); err != nil {
				return err
			}
			if !b42ControlsPlaneswalkerNamed(ctx.Game, ctx.Controller(), "Liliana") {
				return nil
			}
			for _, opp := range ctx.Opponents() {
				ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
					Player:   opp,
					Source:   ctx.Source(),
					N:        1,
					Question: "Liliana's Triumph — discard a card",
				})
			}
			return nil
		},
	})
}
