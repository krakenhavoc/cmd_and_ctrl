package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// phasing_extras.go — the card half of #6201 (game/phasing_extras.go).
//
//	"Pay 1 life: Until your next upkeep, target permanent can't phase out."
//	    (Spatial Binding)        TargetCantPhaseOutUntilYourNextUpkeep
//	"{U}: Until your next upkeep, this creature can't phase out."
//	    (Ertai's Familiar)       ThisCantPhaseOutUntilYourNextUpkeep
//	"Whenever this creature phases out, discard a card." (Teferi's Imp)
//	                             WheneverThisPhasesOut

// TargetCantPhaseOutUntilYourNextUpkeep pins the record to each legal
// card target, "your" being the ability's controller.
func TargetCantPhaseOutUntilYourNextUpkeep(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		for _, t := range NewContext(g, item).LegalTargets() {
			if t.Kind == game.TargetCard {
				g.CantPhaseOutUntilYourNextUpkeepForEffect(item.SourceCardID, item.Controller, t.ID,
					name+" — can't phase out until your next upkeep")
			}
		}
		return nil
	}
}

// ThisCantPhaseOutUntilYourNextUpkeep pins the record to the ability's
// source.
func ThisCantPhaseOutUntilYourNextUpkeep(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		g.CantPhaseOutUntilYourNextUpkeepForEffect(item.SourceCardID, item.Controller, item.SourceCardID,
			name+" — can't phase out until your next upkeep")
		return nil
	}
}

// WheneverThisPhasesOut is "Whenever this <permanent> phases out,
// <effect>": harvested from the phased-out permanent itself, off what it
// looked like on the battlefield (CR 603.10b).
func WheneverThisPhasesOut(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventPhaseOut, Self, label, effect)
}

// WheneverThisPhasesIn is "Whenever this <permanent> phases in,
// <effect>".
func WheneverThisPhasesIn(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventPhaseIn, Self, label, effect)
}
