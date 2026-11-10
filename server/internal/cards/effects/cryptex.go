package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cryptex — Artifact {2} (Murders at Karlov Manor, #4469):
//
//	"{T}, Collect evidence 3: Add one mana of any color. Put an unlock
//	 counter on this artifact. (To collect evidence 3, exile cards with
//	 total mana value 3 or greater from your graveyard.)
//	 Sacrifice this artifact: Surveil 3, then draw three cards. Activate
//	 only if this artifact has five or more unlock counters on it."
//
// The first ability is a mana ability (CR 605.1a) whose cost collects
// evidence (CR 701.59a, ManaCollectEvidenceN): the activator names
// graveyard cards with total mana value 3 or more as exile_ids, and with
// less than that in the graveyard it can't be activated (CR 701.59b).
// Paying it emits the collect evidence event, so "whenever you collect
// evidence" sees it. The counter goes on after the mana, in printed
// order. The auto-tapper never spends the graveyard on its own: the
// ability is a hand-clicked source only.
//
// The second is a CR 602 ability gated on the counters, the sacrifice
// its whole cost.
//
// No simplification.
func init() {
	const unlock = "unlock"
	Register(Spec{
		OracleID:     "8186a992-9065-4a7f-bb00-1c20e01e8fb6",
		Name:         "Cryptex",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true, ExileCards: ManaCollectEvidenceN(3)},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}, Collect evidence 3: Add one mana of any color. Put an unlock counter on this artifact.",
			Rider: func(g *game.Game, _ uuid.UUID, source uuid.UUID) error {
				return g.AddCounterForEffect(source, unlock, 1)
			},
		}},
		Activated: []ActivatedAbility{{
			Label:     "Sacrifice this artifact: Surveil 3, then draw three cards. Activate only if this artifact has five or more unlock counters on it.",
			Purpose:   game.Purpose{Draws: 3, Answers: game.AnswerValue},
			Cost:      SacrificeThis(),
			Condition: SourceHasCountersAtLeast(unlock, 5),
			Effect: Do(Instruction(func(ctx *Context) error {
				return Surveil{Player: ctx.Controller(), N: 3, ThenTail: ctx.TailN(thenDrawCards, 3)}.Apply(ctx)
			})),
		}},
	})
}
