package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Village Pillagers — Creature — Goblin Warrior {3}{R}{R}, 5/5 (Lorwyn
// Eclipsed Commander, #5368):
//
//	"Wither
//	 When this creature enters, it deals 1 damage to each creature your
//	 opponents control.
//	 Whenever a creature an opponent controls with a counter on it dies,
//	 you create a tapped Treasure token."
//
// Wither is the engine's damage tail (game/infect_wither_toxic.go): the
// enters damage is dealt by a source with wither, so it comes as -1/-1
// counters. The dies trigger reads the counters of any kind the creature last
// had (CR 603.10a) and the controller it died under (leftUnderControlOf),
// so a creature of mine that an opponent stole and that died under their
// control counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "54b44499-398b-47ad-b790-bd5b5994d1a6",
		Name:     "Village Pillagers",
		// Its enters damage has wither, so each opposing creature takes a
		// -1/-1 counter rather than marked damage: a -1/-1 sweep.
		Purpose:         game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1, OpponentsOnly: true}},
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordWither},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Village Pillagers — it deals 1 damage to each creature your opponents control",
				func(g *game.Game, item *game.StackItem) error {
					return damageEachMatching(NewContext(g, item), And(Creature(), OpponentControls()), 1)
				}),
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return cmdr2B25OpponentCreatureDiedWithACounter(ev, source, g, false)
			}, "Village Pillagers — create a tapped Treasure token",
				Do(CreateTokenAdvanced{Spec: Token(TreasureToken()).EntersTapped(), N: 1})),
		},
	})
}
