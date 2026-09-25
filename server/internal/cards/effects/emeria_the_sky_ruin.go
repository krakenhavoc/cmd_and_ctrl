package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Emeria, the Sky Ruin — Land (EDHREC rank 988):
//
//	"This land enters tapped.
//	 At the beginning of your upkeep, if you control seven or more
//	 Plains, you may return target creature card from your graveyard
//	 to the battlefield.
//	 {T}: Add {W}."
//
// The mono-white deck's free reanimation, once the Plains count is
// there. Enters tapped is the self-replacement; the mana is a plain
// {W}; the upkeep trigger is optional and targeted, so with seven
// Plains the controller is asked, then picks the creature card in
// the zone browser, and the card comes back under its owner's
// control — "from your graveyard", so that is the controller.
//
// "Seven or more Plains" reads effective subtypes, so any land with
// the Plains type counts (a Savannah, a Plateau), as printed.
//
// Land Tax's intervening-if posture: the Plains count is checked
// when the trigger would go on the stack and not again at
// resolution, the same declared corner every intervening-if card in
// the catalog takes.
func init() {
	Register(Spec{
		OracleID:     "cc999cf2-c99b-4911-8c52-6cc4a99fcc7b",
		Name:         "Emeria, the Sky Ruin",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The Plains count is checked only when the upkeep trigger goes on the stack, so losing a Plains in response won't stop the return."},
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBeginUpkeep},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && b08LandsWithSubtypeControlled(g, source.Controller, "Plains") >= 7
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Emeria, the Sky Ruin — return a creature card from your graveyard to the battlefield?"},
			Targets:        TargetCardInGraveyard("target creature card in your graveyard", Creature(), YouOwn()),
			Key:            "Emeria, the Sky Ruin — return a creature card to the battlefield",
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				return ReturnFromGraveyard{Target: item.Targets[0].ID, Dest: game.ZoneBattlefield}.Apply(NewContext(g, item))
			},
		}},
	})
}
