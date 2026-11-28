package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b853_helpers.go — the Thallids of Fallen Empires (S82 OLD b853,
// #5637): every one grows a spore counter each upkeep and spends three of them.

// S82OldB853SporeCounter is the spore counter's name on the card.
const S82OldB853SporeCounter = "spore"

// S82OldB853SporeUpkeep is "At the beginning of your upkeep, put a spore
// counter on this creature."
func S82OldB853SporeUpkeep(name string) game.TriggeredAbility {
	return AtYourUpkeep(name+" — put a spore counter on this creature", PutACounterOnThis(S82OldB853SporeCounter))
}

// S82OldB853SporeCost is "Remove three spore counters from this creature".
func S82OldB853SporeCost() game.AbilityCost {
	return RemoveCountersFromThis(S82OldB853SporeCounter, 3)
}

// S82OldB853SaprolingRow is "Remove three spore counters from this creature:
// Create a 1/1 green Saproling creature token."
func S82OldB853SaprolingRow() ActivatedAbility {
	return ActivatedAbility{
		Label:   "Remove three spore counters from this creature: Create a 1/1 green Saproling creature token.",
		Purpose: game.Purpose{Tokens: 1, Answers: game.AnswerValue},
		Cost:    S82OldB853SporeCost(),
		Effect: func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			return CreateToken{Controller: ctx.Controller(), Template: TokenCard("1/1 green Saproling"), N: 1}.Apply(ctx)
		},
	}
}

// S82OldB853WasThrull reports whether a permanent, as it last existed, was a
// Thrull.
func S82OldB853WasThrull(info game.PermanentInfo) bool {
	for _, s := range info.Characteristic.Subtypes {
		if s == "Thrull" {
			return true
		}
	}
	return info.Characteristic.AllCreatureTypes
}
