package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// S82OldB379ScourgeGrow is Scourge of Skola Vale's ability: put a number of
// +1/+1 counters on this creature equal to the sacrificed creature's
// toughness, as it last existed on the battlefield (counters and an anthem's
// bonus count; zero or less puts none). A Scourge that has left the
// battlefield is a new object and gets nothing (CR 400.7).
func S82OldB379ScourgeGrow(g *game.Game, item *game.StackItem) error {
	fed, ok := B17PermanentSacrificedToPay(g, item)
	if !ok || !SourceIsStillThisPermanent(g, item) {
		return nil
	}
	n := DepartedCreatureToughness(g, fed)
	if n <= 0 {
		return nil
	}
	return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: n}.Apply(NewContext(g, item))
}

// S82OldB379IsEnchanted is "as long as this creature is enchanted": an Aura
// (a bestowed card included) is attached to it, read through the layers.
func S82OldB379IsEnchanted(g *game.Game, source *game.Card) bool {
	return B18AurasAttachedTo(g, source.InstanceID) > 0
}
