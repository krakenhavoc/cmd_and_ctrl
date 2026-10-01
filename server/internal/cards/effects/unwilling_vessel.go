package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unwilling Vessel — Creature — Human Wizard {2}{U}, 3/2:
//
//	"Vigilance
//	 Eerie — Whenever an enchantment you control enters and whenever
//	 you fully unlock a Room, put a possession counter on this creature.
//	 When this creature dies, create an X/X blue Spirit creature token
//	 with flying, where X is the number of counters on this creature."
//
// Eerie is one ability with two conditions (Eerie, rooms.go). The dies
// trigger reads "counters on this creature" off the last-known
// information (CR 603.10): EVERY kind of counter it wore, not only
// possession counters, so a +1/+1 counter counts too. A Vessel that dies
// with none makes a 0/0 that dies at once, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d2593334-ce0b-43e1-97ad-b00950bfae2b",
		Name:            "Unwilling Vessel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			Eerie("Unwilling Vessel — put a possession counter on it (eerie)", func(g *game.Game, item *game.StackItem) error {
				if !onBattlefield(g, item.SourceCardID) {
					return nil
				}
				return AddCounter{Target: item.SourceCardID, Kind: "possession", N: 1}.Apply(NewContext(g, item))
			}),
			{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return cardDied(ev, source)
				},
				Key: "Unwilling Vessel — create an X/X blue Spirit with flying, X the counters on it",
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, "Unwilling Vessel — create an X/X blue Spirit with flying, X the counters on it")
					if info, ok := g.LastKnownPermanentForEffect(source.InstanceID); ok {
						for _, n := range info.Counters {
							if n > 0 {
								item.Params.Amount += n
							}
						}
					}
					return item
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					x := item.Params.Amount
					return CreateToken{
						Controller: item.Controller,
						Template: game.Card{
							Name: "Spirit", TypeLine: "Token Creature — Spirit",
							Power: x, Toughness: x, PrintedPTKnown: true,
							Colors: []string{"U"}, Keywords: []string{"flying"},
						},
						N: 1,
					}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
