package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// sneak.go — #4586: sneak (CR 702.190), on Marvel Super Heroes and TMNT.
//
//	"Sneak {1}{B}{B} (You may cast this spell for {1}{B}{B} if you also
//	 return an unblocked attacker you control to hand during the declare
//	 blockers step. She enters tapped and attacking.)"
//
// An alternative cost (ADR 0025): the printed price, plus Daze's
// ReturnToHand component narrowed to ninjutsu's unblocked attacker
// (UnblockedAttacker, ninjutsu.go). AlternativeCost.Sneak adds the
// declare-blockers window and the tapped-and-attacking entry; the engine
// half is game/sneak.go. Give the offer to a card the way it is printed:
//
//	AlternativeCosts: []game.AlternativeCost{Sneak("{1}{B}{B}")},
//
// An instant or sorcery with sneak takes the same constructor: the
// window and the cost apply, and the entry clause has no permanent to
// apply to.

// sneakPayLabel is the cost's card clause as CR 702.190a words it.
const sneakPayLabel = "an unblocked creature you control"

// Sneak is "Sneak <cost>" (CR 702.190a). `cost` is the printed sneak
// cost in brace notation. The key "sneak" is an on-disk identity: it
// rides StackItem.AltCost into snapshots.
func Sneak(cost string) game.AlternativeCost {
	return game.AlternativeCost{
		Key:          "sneak",
		Label:        "Sneak " + cost,
		ManaCost:     cost,
		ReturnToHand: PermanentYouControl(sneakPayLabel, Creature(), UnblockedAttacker()),
		PayLabel:     sneakPayLabel,
		Sneak:        true,
	}
}
