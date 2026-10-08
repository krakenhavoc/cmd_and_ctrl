package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Superior Foes of Spider-Man — Creature — Human Rogue Villain {2}{R},
// 3/3:
//
//	"Trample
//	 Whenever you cast a spell with mana value 4 or greater, you may
//	 exile the top card of your library. If you do, you may play that
//	 card until you exile another card with this creature."
//
// Trample rides PrintedKeywords. The trigger is Spider Manifestation's
// (WheneverYouCast(ManaValueGE(4))): the spell's mana value on the
// stack. The "you may" is asked as the trigger resolves (CR 603.5),
// through MayChoice, so declining exiles nothing and leaves the earlier
// window open. A "yes" is #2539's window (exileTopUntilYouExileAnother),
// which closes the one this creature opened for the same player before.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a6a7af21-e52c-4f7f-a57f-bd365d075966",
		Name:            "Superior Foes of Spider-Man",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(ManaValueGE(4), "Superior Foes of Spider-Man — you may exile the top card of your library",
				func(g *game.Game, item *game.StackItem) error {
					return MayChoice{
						Question: "Superior Foes of Spider-Man — exile the top card of your library?",
						OnYes: func(ctx *Context) error {
							return exileTopUntilYouExileAnother(ctx.Game, ctx.Item)
						},
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
