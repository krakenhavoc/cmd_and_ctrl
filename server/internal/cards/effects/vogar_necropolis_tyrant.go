package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vogar, Necropolis Tyrant — Legendary Creature — Zombie Giant {3}{B}{B},
// 4/4 (Game Night: Free-for-All, #5485):
//
//	"Menace
//	 Whenever another creature dies during your turn, put a +1/+1 counter
//	 on Vogar.
//	 When Vogar dies, draw a card for each +1/+1 counter on it."
//
// Menace is the printed keyword. The counter trigger is Orca, Siege
// Demon's "another creature dies" (the types the dying permanent last
// had, CR 603.10a), any player's creature, narrowed to Vogar's
// controller's turn. "During your turn" is part of the trigger event,
// not an intervening "if" (Starfleet Security), so nothing is checked
// again as it resolves. The counter goes on Vogar only while Vogar is on
// the battlefield, so a creature that dies alongside Vogar adds nothing.
//
// The death trigger is Toothy, Imaginary Friend's leave trigger narrowed
// to dying: it counts the +1/+1 counters Vogar had as it last existed on
// the battlefield (CR 608.2h), since by the time it resolves Vogar is in
// the graveyard with none.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1d704c2f-cd24-4a04-a51a-da7984dcd358",
		Name:            "Vogar, Necropolis Tyrant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, vogarAnotherCreatureDiedDuringYourTurn,
				"Vogar, Necropolis Tyrant — put a +1/+1 counter on Vogar",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			WhenThisDies("Vogar, Necropolis Tyrant — draw a card for each +1/+1 counter on it",
				cmdr2B09DrawForEachCounterOnIt),
		},
	})
}

// vogarAnotherCreatureDiedDuringYourTurn is "another creature dies during
// your turn": a creature other than Vogar died while Vogar's controller
// is the active player.
func vogarAnotherCreatureDiedDuringYourTurn(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	return AnotherCreatureDied(ev, source, lki, g) && IsYourTurn(g, source.Controller)
}
