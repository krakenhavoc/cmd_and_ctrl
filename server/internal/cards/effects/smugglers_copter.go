package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Smuggler's Copter — Artifact — Vehicle, 3/3, for {2}:
//
//	"Flying
//	 Whenever this Vehicle attacks or blocks, you may draw a card.
//	 If you do, discard a card.
//	 Crew 1"
//
// SANDBOX SIMPLIFICATION, weaker than printed: the loot trigger
// fires on ATTACKS only. The engine emits no event for a blocker
// being declared — there is no EventBlock in events.go — so "or
// blocks" has nothing to watch. AGENTS.md §7's rule is that a
// trigger on an event the engine does not emit is plumbing work, not
// card work; here the attacks half stands on its own and the missing
// half only ever costs the controller a loot.
//
// The flying is real and does the work a Vehicle's flying does: the
// Copter is not a creature until it is crewed, so the keyword sits
// inert on an artifact until a crew ability resolves, at which point
// Layer 4 makes it a creature and Layer 6 has already put "flying"
// on its ability list.
func init() {
	Register(Spec{
		OracleID:        "49136bdc-bc50-49a2-999a-1ef9c16ea130",
		Name:            "Smuggler's Copter",
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:  "Crew 1",
			Cost:   CrewCost(1),
			Effect: CrewEffect("Smuggler's Copter"),
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclared(ev, source)
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Smuggler's Copter — draw a card, then discard a card?",
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Smuggler's Copter — loot",
					func(g *game.Game, item *game.StackItem) error {
						return lootOne(g, item, 1)
					})
			},
		}},
	})
}
