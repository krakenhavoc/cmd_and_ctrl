package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Meeting of the Five — Sorcery {3}{W}{U}{B}{R}{G} (Streets of New Capenna):
//
//	"Exile the top ten cards of your library. You may cast spells with
//	 exactly three colors from among them this turn. Add {W}{W}{U}{U}{B}{B}
//	 {R}{R}{G}{G}. Spend this mana only to cast spells with exactly three
//	 colors."
//
// The exile is an impulse exile with a cast-only permission for the turn
// (the zero Duration) over spells with exactly three colors
// (PermissionFilter.ColorCountExactly, Niv-Mizzet, Supreme's), read off the
// card's effective colours (CR 105.2, a hybrid card is every colour of its
// symbols, CR 202.2d). The ten are exiled first, then the mana is added, in
// the printed order. The mana is ordinary mana carrying
// game.ManaRestrictColorCountExactly(3): it pays only for a cast spell with
// exactly three colors, from any zone, and not for an activation or a
// spell of any other colour count. It empties at the end of the phase
// (CR 106.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "60fc4c2d-1bf9-4c9a-a240-ab9b1c9c6817",
		Name:         "Meeting of the Five",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			g := ctx.Game
			if _, err := g.ExileTopWithPermissionForEffect(item.Controller, item.Controller, 10, game.CastPermission{
				CastOnly: true,
				Filter:   game.PermissionFilter{ColorCountExactly: 3},
				Label:    "Cast a spell with exactly three colors from among the cards exiled with Meeting of the Five",
			}); err != nil {
				return err
			}
			return g.AddManaWithOptionsForEffect(item.Controller, item.SourceCardID,
				"{W}{W}{U}{U}{B}{B}{R}{R}{G}{G}",
				game.AddManaOptions{Restrictions: []string{game.ManaRestrictCast, game.ManaRestrictColorCountExactly(3)}})
		},
	})
}
