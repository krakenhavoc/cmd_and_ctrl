package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Consul's Lieutenant — Creature — Human Soldier {W}{W}, 2/1:
//
//	"First strike
//	 Renown 1 (When this creature deals combat damage to a player, if
//	 it isn't renowned, put a +1/+1 counter on it and it becomes
//	 renowned.)
//	 Whenever this creature attacks, if it's renowned, other attacking
//	 creatures you control get +1/+1 until end of turn."
//
// #2049: renown is the engine's keyword trigger (game/renown.go). The
// attack trigger's "if it's renowned" is an intervening if (CR 603.4):
// read as the Lieutenant is declared as an attacker, and again as the
// trigger resolves, from its last-known information if it has left.
// "Other attacking creatures you control" is locked in as it resolves
// (CR 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a38b672f-4739-4b7e-8958-2a455510656e",
		Name:            "Consul's Lieutenant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "renown 1"},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, sourceLKI game.Characteristic, g *game.Game) bool {
				return ThisAttacked(ev, source, sourceLKI, g) && ThisIsRenowned(source)
			}, "Consul's Lieutenant — other attacking creatures you control get +1/+1", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if !ThisWasRenowned(ctx) {
					return nil
				}
				return BoostUntilEOT{
					Match:     And(AttackingCreature(), YouControl(), OtherThan(item.SourceCardID)),
					Power:     1,
					Toughness: 1,
					Label:     "Consul's Lieutenant — +1/+1",
				}.Apply(ctx)
			}),
		},
	})
}
