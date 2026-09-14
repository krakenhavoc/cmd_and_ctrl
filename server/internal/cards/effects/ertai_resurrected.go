package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ertai Resurrected — Legendary Creature — Phyrexian Human Wizard
// {2}{U}{B}, 3/2 (EDHREC rank 3404):
//
//	"Flash
//	 When Ertai Resurrected enters, choose up to one —
//	 • Counter target spell, activated ability, or triggered
//	   ability. Its controller draws a card.
//	 • Destroy another target creature or planeswalker. Its
//	   controller draws a card."
//
// A flash body that is a Cancel or a Murder, with a card for the
// victim either way. Flash rides PrintedKeywords, so he can be cast
// with a spell on the stack and his entry trigger goes on the stack
// above it.
//
// The modal entry trigger has no mode prompt of its own — a
// triggered ability cannot carry a "choose one" — so the mode is
// the target: one clause spanning the stack and the battlefield
// (Aang, Swift Savior's two-zone shape), "up to one" so declining is
// a legal answer. Picking a spell is the counter mode; picking a
// creature or planeswalker is the destroy mode; picking nothing is
// "up to one" doing nothing. The victim's controller is read before
// the counter or the destruction moves the card, and draws whether
// or not an indestructible creature survives — the draw is not
// conditional on the removal, as printed. A target gone in response
// counters the trigger (CR 608.2b), so no draw then.
//
// Sandbox simplifications, declared, both weaker than printed:
//
//   - Only a spell can be countered. An activated or triggered
//     ability on the stack is a StackMeta item with no card, and a
//     target clause enumerates cards — Disallow's posture.
//   - The mode is chosen through the target prompt rather than a
//     mode picker.
func init() {
	Register(Spec{
		OracleID:     "3d038f7c-95fa-4b71-8f74-b9b4dd45cde0",
		Name:         "Ertai Resurrected",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Only a spell can be countered — an activated or triggered ability on the stack can't be picked.",
			"The mode is chosen through the target prompt: pick a spell to counter it, pick a creature or planeswalker to destroy it, or pick nothing.",
		},
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Targets:   b32TargetSpellOrAnotherCreatureOrPlaneswalker("up to one target spell, or another target creature or planeswalker"),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, b32ErtaiLabel, b32CounterOrDestroyChosenThenControllerDraws)
			},
		}},
	})
}
