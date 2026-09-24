package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Virtue of Knowledge // Vantress Visions — adventure card (CR 715),
// oracle f0bbcabf…:
//
//	Virtue of Knowledge — Enchantment {4}{U}
//	  "If a permanent entering causes a triggered ability of a
//	   permanent you control to trigger, that ability triggers an
//	   additional time."
//	Vantress Visions — Instant — Adventure {1}{U}
//	  "Copy target activated or triggered ability you control. You may
//	   choose new targets for the copy."
//
// Both halves are compositions of machinery that already exists, and
// neither adds anything of its own:
//
//   - The enchantment is Panharmonicon with the filter taken off.
//     Panharmonicon narrows the entering object to "an artifact or
//     creature"; the Virtue says "a permanent", which is every object
//     that can enter the battlefield, so DoublesEntering gets a nil
//     filter. The "of a permanent you control" half is DoublesEntering's
//     own controller check, the same one Panharmonicon relies on. A land
//     entering therefore doubles a landfall trigger here, which is the
//     observable difference between the two cards and what the test pins.
//   - The Adventure half is Lithoform Engine's first ability as an
//     instant: the same AbilityOnStack clause with AnAbilityYouControl
//     (no TriggeredAbilityOnly — "activated OR triggered"), and the same
//     copyTargetedAbility body with CR 707.10c's re-target offer. A mana
//     ability never uses the stack (CR 605.3b), so it is not a legal
//     target and needs no predicate to exclude it.
//
// Two keys, one card: the enchantment keeps the bare oracle ID and the
// Adventure takes "<oracle_id>#1", as Bonecrusher Giant // Stomp does.
// What happens after Vantress Visions resolves — exile, and the
// permission to cast the enchantment later — is the engine's CR 715.3d
// branch (game/adventure.go).
//
// No simplification.

// virtueOfKnowledgeOracleID is shared by both faces of the card.
const virtueOfKnowledgeOracleID = "f0bbcabf-29e7-4c7e-893f-86b64d3620a9"

func init() {
	// Face 0 — the enchantment.
	doubler := DoublesEntering(nil)
	doubler.Label = "Virtue of Knowledge"
	Register(Spec{
		OracleID:        virtueOfKnowledgeOracleID,
		Name:            "Virtue of Knowledge",
		Completeness:    CompletenessFull,
		TriggerDoublers: []game.TriggerDoubler{doubler},
	})

	// Face 1 — the Adventure, an instant while it is on the stack
	// (CR 715.3a).
	Register(Spec{
		OracleID:     virtueOfKnowledgeOracleID + "#1",
		Name:         "Vantress Visions",
		Completeness: CompletenessFull,
		Targets:      AbilityOnStack("target activated or triggered ability you control", AnAbilityYouControl()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return copyTargetedAbility(ctx.Game, item)
		},
	})
}
