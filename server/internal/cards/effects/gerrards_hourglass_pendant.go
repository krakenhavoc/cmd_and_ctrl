package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gerrard's Hourglass Pendant — Legendary Artifact {1}:
//
//	"Flash
//	 If a player would begin an extra turn, that player skips that turn
//	 instead.
//	 {4}, {T}, Exile Gerrard's Hourglass Pendant: Return to the
//	 battlefield tapped all artifact, creature, enchantment, and land
//	 cards in your graveyard that were put there from the battlefield
//	 this turn."
//
// ADR 0188 Decision 2 (#3560).
//
//   - Flash is a PrintedKeyword.
//   - The extra-turn clause is SkipExtraTurns(), shared with Ugin's
//     Nexus: it skips the Pendant controller's own extra turns too.
//   - The third ability exiles the Pendant as a COST (ExileThis, from the
//     battlefield: it is gone at announce and cannot be destroyed in
//     response to save the cost), so it never counts among the cards it
//     returns. The set is read once as the ability resolves, from your
//     own graveyard, by each card's newest arrival there
//     (putIntoGraveyardFromBattlefieldThisTurn: a card that left and came
//     back is a new object, CR 400.7). A planeswalker or battle card is
//     not on the list, and a token is not a card (CR 108.2). The cards
//     enter together, tapped, as one simultaneous entry (CR 603.6a), under
//     their owner's control, which is yours: it is your graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "01591d42-4ef8-471d-beda-d31186bba6a6",
		Name:            "Gerrard's Hourglass Pendant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Replacements:    []game.ReplacementEffect{SkipExtraTurns()},
		Activated: []ActivatedAbility{{
			Label:   "{4}, {T}, Exile Gerrard's Hourglass Pendant: Return to the battlefield tapped all artifact, creature, enchantment, and land cards in your graveyard that were put there from the battlefield this turn.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{4}"), TapCost(), ExileThis()),
			Effect:  returnPutThereFromTheBattlefieldThisTurnTapped,
		}},
	})
}

// returnPutThereFromTheBattlefieldThisTurnTapped is the Pendant's third
// ability: every artifact, creature, enchantment and land card in the
// controller's graveyard that was put there from the battlefield this
// turn returns to the battlefield tapped, together.
func returnPutThereFromTheBattlefieldThisTurnTapped(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ids := graveyardCardIDs(ctx, item.Controller, func(c game.Card) bool {
		if c.IsToken() {
			return false
		}
		if !c.IsArtifact() && !c.IsCreature() && !c.IsEnchantment() && !c.IsLand() {
			return false
		}
		return putIntoGraveyardFromBattlefieldThisTurn(g, c.InstanceID)
	})
	if len(ids) == 0 {
		return nil
	}
	return ReturnFromGraveyardTogether{Targets: ids, Controller: uuid.Nil, Tapped: true}.Apply(ctx)
}
