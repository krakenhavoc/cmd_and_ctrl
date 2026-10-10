package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// boast.go — #2697, CR 702.142: the vocabulary a card file writes a
// boast ability in. The engine half is game/boast.go.
//
//	Activated: []ActivatedAbility{
//	    Boast("{1}{R}: Create a 2/1 red Dwarf Berserker creature token.",
//	        ManaCost("{1}{R}"), createDwarfBerserker),
//	},
//
// "Boast — [cost]: [effect]. (Activate only if this creature attacked
// this turn and only once each turn.)" is two activation instructions,
// and a card file must not assemble them: the constructor sets the one
// bit, and game.Game.BoastBlockLocked reads the attack record and the
// activation tally in the one place the engine, the bot enumerator and
// the view share. The ability is NOT a Condition — a Condition greys a
// row with a generic reason, and the client owes a boast row the
// specific one (it hasn't attacked; it has been used).

// boastPrefix is the printed ability word, with the dash the label and
// the oracle text both use.
const boastPrefix = "Boast — "

// Boast builds a boast ability. `text` is the printed line AFTER the
// ability word, "[cost]: [effect]." — the constructor prepends
// "Boast — " so the label reads as printed, which is what the oracle
// fixture compares.
func Boast(text string, cost game.AbilityCost, effect Effect) ActivatedAbility {
	return ActivatedAbility{
		Label:  boastPrefix + text,
		Cost:   cost,
		Boast:  true,
		Effect: effect,
	}
}

// BoastTargeting is Boast for an ability with a target clause
// ("Boast — {1}{R}: This creature deals 1 damage to any target.").
func BoastTargeting(text string, cost game.AbilityCost, targets *game.TargetSpec, effect Effect) ActivatedAbility {
	ab := Boast(text, cost, effect)
	ab.Targets = targets
	return ab
}

// YourCreaturesBoastTimes — "Creatures you control can boast twice
// during each of your turns rather than once" (Birgi, God of
// Storytelling). A per-controller replacement of the printed limit of
// one, not an addition to it, so two of these do not stack: the engine
// takes the largest limit that applies (game.Game.BoastLimitFor).
//
// Both halves of the clause are load-bearing: it reaches the
// controller's creatures only, and only during that player's own turn.
func YourCreaturesBoastTimes(label string, times int) game.BoastLimit {
	return game.BoastLimit{
		Label: label,
		Limit: times,
		Applies: func(g *game.Game, boaster game.Card, grantor game.Card) bool {
			return g != nil && boaster.Controller == grantor.Controller &&
				IsYourTurn(g, grantor.Controller)
		},
	}
}

// ABoastAbility — the activation event is of a boast ability. Reads
// the bit the announcement stamped (game.Event.Boast) rather than going
// back to the source's ability list, for the reason AnExhaustAbility
// does: a SacrificeSelf cost has already ended the object by the time
// a watcher runs.
func ABoastAbility(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Boast
}

// ABoastAbilityCost — the ability being priced is a boast ability
// (Dragonkin Berserker's "Boast abilities you activate cost {1} less").
func ABoastAbilityCost() CostPredicate {
	return func(q game.CostQuery) bool {
		return q.Ability != nil && q.Ability.Boast
	}
}

// ActivatedByTheModifiersController — "abilities YOU activate": the
// player paying for the ability is the controller of the permanent that
// contributes the modifier. Dragonkin Berserker's discount is its
// controller's alone; an opponent's boast is priced in full.
func ActivatedByTheModifiersController() CostPredicate {
	return func(q game.CostQuery) bool {
		return q.Controller == q.Source.Controller
	}
}

// createTheToken is "Create a <token> creature token" as an ability's
// whole effect, under the ability's controller. The key is a
// tokens_table.go row.
func createTheToken(key string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return CreateToken{Controller: item.Controller, Template: TokenCard(key), N: 1}.Apply(NewContext(g, item))
	}
}

// BoastAnswering is Boast with its ADR 0142 declaration: what the row
// can do in response, read by smart autopass and the bot. A boast row
// is declared by what its effect does, so it is not declared once in
// Boast.
func BoastAnswering(answers game.Answers, text string, cost game.AbilityCost, effect Effect) ActivatedAbility {
	ab := Boast(text, cost, effect)
	ab.Purpose.Answers = answers
	return ab
}
