package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shifting Woodland — Land:
//
//	"This land enters tapped unless you control a Forest.
//	 {T}: Add {G}.
//	 Delirium — {2}{G}{G}: This land becomes a copy of target permanent
//	 card in your graveyard until end of turn. Activate only if there
//	 are four or more card types among cards in your graveyard."
//
// The enters-tapped clause and the mana ability are the ordinary
// checkland shape (EntersTappedUnless, one Forest, not "another" —
// this land has no printed subtype of its own to exclude).
//
// The delirium ability is a duration copy (#1593, become_copy.go): the
// land becomes the graveyard card until the cleanup step and then is
// Shifting Woodland again. The copied values are read as the ability
// resolves and stored, so exiling the card afterwards does not end the
// copy. A copied creature has summoning sickness unless the land has
// been under its controller's control since the turn began (CR 302.6),
// which is the usual reason to activate it on an opponent's end step.
//
// Shipped in #1592 with the delirium ability caveated; #1593 lifted it.
func init() {
	const label = "Delirium — {2}{G}{G}: This land becomes a copy of target permanent card in your graveyard until end of turn. Activate only if there are four or more card types among cards in your graveyard."
	effects.Register(effects.Spec{
		OracleID:     "7c2a4fe5-43e8-4e20-bef2-0278d18afc4b",
		Name:         "Shifting Woodland",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.EntersTappedUnless(effects.OtherLandsWithSubtypeAtLeast("forest", 1))},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   label,
			Cost:    effects.ManaCost("{2}{G}{G}"),
			Targets: effects.TargetCardInGraveyard("target permanent card in your graveyard", effects.YouOwn(), effects.Permanent()),
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
				return effects.B16CardTypesInGraveyard(g, controller) >= 4
			},
			Effect: effects.SelfBecomesCopyOfTarget("Shifting Woodland — becomes a copy until end of turn"),
		}},
	})
}
