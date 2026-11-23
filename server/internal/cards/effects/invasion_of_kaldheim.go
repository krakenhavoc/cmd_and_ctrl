package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Invasion of Kaldheim // Pyre of the World Tree — Battle — Siege, defense 4,
// for {3}{R} (March of the Machine, S82 batch 23):
//
//	"(As a Siege enters, choose an opponent to protect it. You and others
//	  can attack it. When it's defeated, exile it, then cast it
//	  transformed.)
//	 When this Siege enters, exile all cards from your hand, then draw that
//	 many cards. Until the end of your next turn, you may play cards exiled
//	 this way."
//
// The exile is one simultaneous event; the draw and the permission both
// wait for it (CR 608.2c) and count the cards that actually reached exile
// (a commander sent to the command zone instead was not exiled "this
// way"). The permission is an ordinary play permission over exactly those
// cards, to the end of the controller's next turn, so a land may be played
// from it. The draw happens first and the cards are drawn from the
// library, never from the exile.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     InvasionOfKaldheimOracleID,
		Name:         "Invasion of Kaldheim",
		Completeness: CompletenessFull,
		Battle: &BattleSpec{
			Defense: 4,
			Subtype: BattleSubtypeSiege,
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Invasion of Kaldheim — exile all cards from your hand, then draw that many cards", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return g.ExileCardsKeyedForEffect(AllHandCardIDs(g, ctx.Controller()), ctx.ItemTail(invasionOfKaldheimDraw))
			}),
			DefeatedTrigger("Invasion of Kaldheim — defeated: exile it, then cast Pyre of the World Tree", SiegeDefeated()),
		},
	})
}

// invasionOfKaldheimOracleID is shared with the back face's spec, which
// registers under this ID plus "#1".
const InvasionOfKaldheimOracleID = "66914a34-59b3-48a0-9305-ebd956d34f02"

// invasionOfKaldheimDraw is "then draw that many cards" and the play window
// over the cards exiled this way. An on-disk key.
var invasionOfKaldheimDraw = CardTail("invasion-of-kaldheim/draw-and-play", func(ctx *Context, o game.TailOutcome) error {
	g, controller := ctx.Game, ctx.Controller()
	if len(o.IDs) == 0 {
		return nil
	}
	cards := make([]game.Card, 0, len(o.IDs))
	for _, id := range o.IDs {
		if c, ok := g.LookupCardForEffect(id); ok {
			cards = append(cards, c)
		}
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:   controller,
		Zone:     game.ZoneExile,
		Duration: g.UntilEndOfYourNextTurnDuration(controller),
	}, cards)
	return DrawCards{Player: controller, N: len(o.IDs)}.Apply(ctx)
})
