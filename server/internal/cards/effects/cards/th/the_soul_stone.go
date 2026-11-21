package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Soul Stone — Legendary Artifact — Infinity Stone {1}{B}
// (#1600, tracker #887):
//
//	"Indestructible
//	 {T}: Add {B}.
//	 {6}{B}, {T}, Exile a creature you control: Harness The Soul
//	 Stone. (Once harnessed, its ∞ ability is active.)
//	 ∞ — At the beginning of your upkeep, return target creature card
//	 from your graveyard to the battlefield."
//
// The Mind Stone's shape (the_mind_stone.go) with the one cost
// component it was waiting on: "Exile a creature you control"
// (game.ExilePermanentsCost, ADR 0020's 2026-10-03 amendment). The
// creature is named at announce and exiled before the harness ability
// is on the stack, so its leaves-the-battlefield triggers resolve
// first; it is not sacrificed, so no dies or sacrifice trigger sees
// it. A commander named to the cost is offered the command zone first,
// and the Stone is harnessed either way.
//
// INDESTRUCTIBLE rides PrintedKeywords; {T}: ADD {B} is Sol Ring's
// shape. THE ∞ ABILITY is Emeria, the Sky Ruin's upkeep reanimation
// with no "if" and no "may", gated on Harnessed() so it does not exist
// at all until the harness ability has resolved (CR 702.186b). The card
// returns under its owner's control, which "from your graveyard" makes
// the controller. Activating the harness again pays the cost and does
// nothing (CR 701.64a), as printed.
//
// No simplification.
func init() {
	harness := effects.Harness("{6}{B}, {T}, Exile a creature you control: Harness The Soul Stone.",
		effects.Plus(effects.ManaCost("{6}{B}"), effects.TapCost(), effects.ExileACreatureYouControl()))
	// ADR 0142: the cost exiles a creature you control.
	harness.Purpose = game.Purpose{Answers: game.AnswerSacOutlet}
	effects.Register(effects.Spec{
		OracleID:        "92cfba68-12f6-4f97-9187-0f6a39656a0f",
		Name:            "The Soul Stone",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
		Activated: []effects.ActivatedAbility{harness},
		Triggered: []game.TriggeredAbility{theSoulStoneUpkeepReturn()},
	})
}

// theSoulStoneUpkeepReturn is "∞ — At the beginning of your upkeep,
// return target creature card from your graveyard to the battlefield."
func theSoulStoneUpkeepReturn() game.TriggeredAbility {
	t := effects.AtYourUpkeep("The Soul Stone — return a creature card from your graveyard to the battlefield",
		effects.ReturnFirstLegalGraveyardTargetToBattlefield)
	t.Targets = effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn())
	t.ActiveWhen = effects.Harnessed()
	return t
}
