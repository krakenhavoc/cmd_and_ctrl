package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Guardian of the Gateless — Creature — Angel {4}{W}, 3/3:
//
//	"Flying
//	 This creature can block any number of creatures.
//	 Whenever this creature blocks, it gets +1/+1 until end of turn for
//	 each creature it's blocking."
//
// CanBlockAnyNumber on itself (#1706). "Whenever this creature blocks"
// has no object, so it triggers ONCE however many attackers the Angel
// blocks (CR 509.3a) — selfBlocksOnce — and the count is read as the
// trigger resolves (game.BlockingCountForEffect), so an attacker
// removed from combat in response no longer counts.
func init() {
	Register(Spec{
		OracleID:        "e133605f-9227-42c4-afe9-6613ab095433",
		Name:            "Guardian of the Gateless",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{CanBlockAnyNumber(selfOnly)},
		Triggered: []game.TriggeredAbility{
			On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return selfBlocksOnce(ev, source)
			}, "Guardian of the Gateless — +1/+1 for each creature it's blocking", guardianOfTheGatelessPump),
		},
	})
}

// guardianOfTheGatelessPump gives the Angel +1/+1 until end of turn per
// attacker it is still blocking.
func guardianOfTheGatelessPump(g *game.Game, item *game.StackItem) error {
	n := g.BlockingCountForEffect(item.SourceCardID)
	if n == 0 {
		return nil
	}
	return BoostUntilEOT{
		Target:    item.SourceCardID,
		Power:     n,
		Toughness: n,
		Label:     "Guardian of the Gateless — +1/+1 for each creature it's blocking",
	}.Apply(NewContext(g, item))
}
