package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
// "If you control seven or more Plains" is an intervening "if"
// (CR 603.4): checked when the trigger would go on the stack and again
// as it resolves, so losing a Plains in response stops the return. The
// second check is declared (InterveningIf), so it comes before the "you
// may" and a failed one asks nothing (#3907).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cc999cf2-c99b-4911-8c52-6cc4a99fcc7b",
		Name:         "Emeria, the Sky Ruin",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
		Triggered: []game.TriggeredAbility{effects.InterveningIf(effects.CondBoard{
			Name: "you control seven or more Plains",
			Test: func(ctx *effects.Context) bool {
				return effects.B08LandsWithSubtypeControlled(ctx.Game, ctx.Controller(), ctx.TextSubtype("Plains")) >= 7
			},
		}, game.TriggeredAbility{
			Watches: []game.EventKind{game.EventBeginUpkeep},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && effects.B08LandsWithSubtypeControlled(g, source.Controller, source.TextSubtype("Plains")) >= 7
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Emeria, the Sky Ruin — return a creature card from your graveyard to the battlefield?"},
			Targets:        effects.TargetCardInGraveyard("target creature card in your graveyard", effects.Creature(), effects.YouOwn()),
			Key:            "Emeria, the Sky Ruin — return a creature card to the battlefield",
			Effect:         effects.HobB8ReturnTargetGraveyardCardToBattlefield,
		})},
	})
}
