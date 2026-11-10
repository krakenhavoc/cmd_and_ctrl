package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cosmium Confluence — Sorcery {4}{G} (The Lost Caverns of Ixalan, #4470):
//
//	"Choose three. You may choose the same mode more than once.
//	 • Search your library for a Cave card, put it onto the battlefield
//	   tapped, then shuffle.
//	 • Put three +1/+1 counters on a Cave you control. It becomes a 0/0
//	   Elemental creature with haste. It's still a land.
//	 • Destroy target enchantment."
//
// Fiery Confluence's repeating modal spell (CR 700.2d), the bullets in
// printed order. The search is a mandatory one for a Cave card (finding
// none is allowed, CR 701.23b) and each time it shuffles. The animate
// bullet does not target: the controller names the Cave as it resolves
// (a keyed pick), and the counters-then-animate is awaken's with the
// Elemental. Chosen again, the second pick may be the Cave the first one
// animated. Each destroy bullet names its own enchantment.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "798aaf78-16d9-474b-a963-f5e9c39aae6e",
		Name:         "Cosmium Confluence",
		Completeness: CompletenessFull,
		Modes: ChooseNRepeating("Choose three. You may choose the same mode more than once.", 3, 3,
			ModeWithPurpose(ModeSteps("Search your library for a Cave card, put it onto the battlefield tapped, then shuffle.", nil,
				Instruction(func(ctx *Context) error {
					return SearchLibrary{
						Player:        ctx.Controller(),
						Predicate:     func(c game.Card) bool { return c.HasSubtype("Cave") },
						Dest:          game.ZoneBattlefield,
						Limit:         1,
						Shuffle:       true,
						TappedOnEntry: true,
						Reason:        "Cosmium Confluence — a Cave card",
						Source:        ctx.Source(),
					}.Apply(ctx)
				})), game.Purpose{Lands: 1, Tutors: 1}),
			ModeSteps("Put three +1/+1 counters on a Cave you control. It becomes a 0/0 Elemental creature with haste. It's still a land.", nil,
				Instruction(func(ctx *Context) error {
					caves := cavesYouControl(ctx)
					if len(caves) == 0 {
						return nil
					}
					ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
						Chooser:  ctx.Controller(),
						Source:   ctx.Source(),
						Question: "Cosmium Confluence — put three +1/+1 counters on a Cave you control",
						Cards:    caves,
						Min:      1,
						Max:      1,
						ThenKey:  cosmiumConfluenceAnimate,
						LiveZone: game.ZoneBattlefield,
						Asked:    game.NewAskedItem(ctx.Item),
					})
					return nil
				})),
			ModeDoing("Destroy target enchantment.",
				TargetPermanent("target enchantment", Enchantment()),
				DestroyTheModesTarget),
		),
	})
}
