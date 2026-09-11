package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// City of Brass — Land:
//
//	"Whenever this land becomes tapped, it deals 1 damage to you."
//	"{T}: Add one mana of any color."
//
// The one card in this batch whose pain is NOT a rider, and modelling
// it as one would have been wrong in the direction that matters.
// "Whenever this land becomes tapped" is a real CR 603 triggered
// ability: it uses the stack, players get priority before it
// resolves, and it fires on EVERY tap — an opponent's Icy Manipulator,
// a convoke-style tap, a "tap target land" effect — not only on
// tapping for mana. Folding it into the mana ability would have
// silently made City of Brass better than printed by dropping every
// tap that wasn't an activation, and a batch rule of this project is
// that stronger-than-printed is the one direction we don't ship.
//
// So the mana ability here is a plain five-colour pipe with no rider
// at all, and the damage rides EventTapCard. The two halves meet in
// the engine: ActivateManaAbility emits EventTapCard as it pays the
// tap cost, the harvester queues the trigger, and it resolves off the
// stack after the mana is already in the pool — exactly the printed
// sequencing.
//
// One consequence worth stating: the auto-tapper is free to plan City
// of Brass (its mana ability really is painless in isolation), and the
// damage still happens, because materializePlanLocked emits
// EventTapCard the same way. That is correct rather than a leak.
//
// "Any color" opts out of commander-identity narrowing — see
// mana_confluence.go for why that matters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f25351e3-539b-4bbc-b92d-6480acf4d722",
		Name:         "City of Brass",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:                    ManaAbilityCost{Tap: true},
			Produced:                "{W|U|B|R|G}",
			Label:                   "Add one mana of any color",
			IgnoreCommanderIdentity: true,
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventTapCard},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "City of Brass — 1 damage to you",
					func(g *game.Game, item *game.StackItem) error {
						// "it deals 1 damage to YOU" — the land's
						// controller, even when an opponent is the one
						// who tapped it. NewTriggeredItem already put
						// the controller on the item.
						return g.DealDamageToPlayerForEffect(item.SourceCardID, item.Controller, 1)
					})
			},
		}},
	})
}
