package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Field of Souls — Enchantment {2}{W}{W}:
//
//	"Whenever a nontoken creature is put into your graveyard from the
//	 battlefield, create a 1/1 white Spirit creature token with flying."
//
// A card goes to its owner's graveyard (CR 400.3), so "your graveyard" is
// judged on the card's owner, not on who controlled it as it died: your
// creature stolen by an opponent still pays, an opponent's creature you
// control does not. The creature is judged on its last-known types
// (CR 603.10a); a token is told apart by the card the event names.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4d7a5b14-8fce-41f2-a0d5-fff3d15f41f6",
		Name:         "Field of Souls",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && !c.IsToken() && c.Owner == source.Controller && leftAsType(ev, c, "creature")
			}, "Field of Souls — create a 1/1 white Spirit creature token with flying",
				Do(CreateToken{Template: TokenCard("1/1 white Spirit with flying"), N: 1})),
		},
	})
}
