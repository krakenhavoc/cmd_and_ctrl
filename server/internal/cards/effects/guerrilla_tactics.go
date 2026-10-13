package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Guerrilla Tactics — Instant {1}{R}:
//
//	"Guerrilla Tactics deals 2 damage to any target.
//	 When a spell or ability an opponent controls causes you to discard
//	 this card, Guerrilla Tactics deals 4 damage to any target."
//
// The spell is Lightning Bolt's shape with 2. The second ability is the
// family's "this card" trigger (ADR 0153 §4): it watches from the
// GRAVEYARD, on a discard an opponent's spell or ability caused (the
// stack item's controller, CR 113.8), with its own target clause chosen
// when it goes on the stack. "Guerrilla Tactics deals" names the card in
// the graveyard as the damage source, last-known information included.
//
// Caveat, declared: if madness or Rest in Peace sends the discarded card
// to exile instead of the graveyard, its ability does not trigger from
// exile (owner decision 6).
func init() {
	Register(Spec{
		OracleID:          "65ad90f5-4dae-494b-83fb-8052e77e0b12",
		Name:              "Guerrilla Tactics",
		Completeness:      CompletenessCaveats,
		Caveats:           []string{discardTriggerFromExileCaveat},
		DiscardByOpponent: game.DiscardByOpponentTrigger,
		Targets:           TargetAny(),
		Purpose:           ForTargets(DamageToTarget(0, 2)),
		OnResolve:         damageToFirstTarget(2),
		Triggered: []game.TriggeredAbility{
			Targeting(InGraveyard(On(game.EventDiscardCard, DiscardedByAnOpponent(true),
				"Guerrilla Tactics — discarded: 4 damage to any target",
				func(g *game.Game, item *game.StackItem) error {
					return damageToFirstTarget(4)(item, NewContext(g, item))
				})), TargetAny()),
		},
	})
}
