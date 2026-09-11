package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Imperial Recruiter — 1/1 Creature — Human Advisor for {2}{R}:
//
//	"When this creature enters, search your library for a creature
//	 card with power 2 or less, reveal it, put it into your hand,
//	 then shuffle."
//
// A tutor with a power ceiling, which in this deck means it finds
// Ragavan, Marauding Mako, Siren Stormtamer or the commander. The
// power check reads the printed value: a card in the library has no
// battlefield characteristics for the layer engine to modify, so
// there's nothing else it could mean.
func init() {
	Register(Spec{
		OracleID: "4d6a1391-817a-4ddc-840d-886b138eeb3f",
		Name:     "Imperial Recruiter",
		OnETB: func(card *game.Card, ctx *Context) error {
			return SearchLibrary{
				Player: card.Controller,
				Predicate: func(c game.Card) bool {
					return c.IsCreature() && c.Power <= 2
				},
				Dest:    game.ZoneHand,
				Limit:   1,
				Reveal:  true,
				Shuffle: true,
				Reason:  "Imperial Recruiter — a creature with power 2 or less",
			}.Apply(ctx)
		},
	})
}
