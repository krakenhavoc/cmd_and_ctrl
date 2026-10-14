package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Emrakul, the Promised End — Legendary Creature — Eldrazi {13}, 13/13:
//
//	"This spell costs {1} less to cast for each card type among cards
//	 in your graveyard.
//	 When you cast this spell, you gain control of target opponent
//	 during that player's next turn. After that turn, that player takes
//	 an extra turn.
//	 Flying, trample, protection from instants"
//
// The discount is a self cost modifier over the distinct card types
// among the cards in the caster's graveyard, the delirium count, read
// as the spell is priced and never counting the card being cast.
//
// The cast trigger targets an opponent as Emrakul is cast and resolves
// above it, so it happens even if Emrakul is countered (the ruling).
// It schedules a turn-long control (ADR 0167,
// game.BeginTurnControlForEffect): the caster plays the opponent's next
// turn they actually take (CR 723.1, 723.1b), sees what they see
// (CR 723.4) and makes their decisions (CR 723.5), and the opponent
// takes an extra turn directly after it (CR 500.7). The control ends
// if either player leaves; the extra turn never begins for a player who
// has left (CR 800.4a, 800.4k). Several Emrakuls over one player each
// give their own extra turn, and the newest controls the turn
// (CR 723.1a).
//
// Protection from instants is a printed keyword the engine enforces.
//
// No simplification.
func init() {
	cast := WhenYouCastThisSpell(
		"Emrakul, the Promised End — you gain control of target opponent during that player's next turn",
		emrakulThePromisedEndControl)
	cast.Targets = TargetPlayer("target opponent", Opponent())
	Register(Spec{
		OracleID:        "4e7a8817-1a66-45c3-ade9-eac79b40b89f",
		Name:            "Emrakul, the Promised End",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample", "protection from instants"},
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(CardTypesAmongCardsInYourGraveyard(),
				"This spell costs {1} less to cast for each card type among cards in your graveyard."),
		},
		Triggered: []game.TriggeredAbility{cast},
	})
}

// emrakulThePromisedEndControl schedules the control of the targeted
// opponent if they are still a legal target (CR 608.2b). The engine
// refuses a player who has left the game.
func emrakulThePromisedEndControl(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, ref := range ctx.LegalTargets() {
		if ref.Kind != game.TargetPlayer || ref.ID == uuid.Nil {
			continue
		}
		g.BeginTurnControlForEffect(game.TurnControl{
			Controller: item.Controller,
			Controlled: ref.ID,
			Source:     item.SourceCardID,
			SourceName: "Emrakul, the Promised End",
		})
		return nil
	}
	return nil
}
