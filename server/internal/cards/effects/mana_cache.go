package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mana Cache — Enchantment {1}{R}{R}:
//
//	"At the beginning of each player's end step, put a charge counter
//	 on this enchantment for each untapped land that player controls.
//	 Remove a charge counter from this enchantment: Add {C}. Any player
//	 may activate this ability but only during their turn before the
//	 end step."
//
// The mana ability is an any-player one (ADR 0106 §1 amendment
// 2026-10-16, #1864): the activator removes the counter (CR 602.1a: the
// self form of the cost names this enchantment, whoever controls it),
// the {C} goes into the activator's pool (CR 106.4), and "their turn"
// is the activator's (CR 109.5). It is still a mana ability (CR 605.1a:
// no target, it adds mana), so it uses no stack and is activated
// whenever its activator may activate a mana ability (CR 605.3a),
// within its own "their turn before the end step".
//
// The trigger is "each player's" end step: "that player" is the active
// player, the Actor of the end step's event, and the lands are counted
// once, as the trigger resolves (CR 608.2h). An enchantment that left
// and came back in response is a new object and gets nothing (#1432).
//
// The auto-tapper never spends the counters on its own: the ability has
// no {T}, and the charge counters are a resource the activator chooses
// to spend. Each player activates it from the board.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1c09f2d1-413b-4a92-aa39-99c1a2106814",
		Name:         "Mana Cache",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, AnyPlayer,
				"Mana Cache — put a charge counter on it for each untapped land that player controls",
				manaCacheCharge),
		},
		ManaAbilities: []ManaAbility{{
			Cost:      ManaAbilityCost{RemoveCounters: RemoveCountersFromThis(game.CounterCharge, 1).RemoveCounters},
			Produced:  "{C}",
			Label:     "Remove a charge counter from this enchantment: Add {C}. Any player may activate this ability but only during their turn before the end step.",
			AnyPlayer: true,
			Condition: DuringYourTurnBeforeTheEndStep(),
		}},
	})
}

// manaCacheCharge puts a charge counter on Mana Cache for each untapped
// land the player whose end step it is controls, counted now.
func manaCacheCharge(g *game.Game, item *game.StackItem) error {
	if sourceIsNewObject(g, item) { // #1432
		return nil
	}
	player := triggeringActor(item)
	if player == uuid.Nil {
		return nil
	}
	n := countControlled(g, player, func(c game.Card) bool { return c.IsLand() && !c.Tapped })
	if n == 0 {
		return nil
	}
	return AddCounter{Target: item.SourceCardID, Kind: game.CounterCharge, N: n}.Apply(NewContext(g, item))
}
