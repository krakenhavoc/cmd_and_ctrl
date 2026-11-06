package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Extract a Confession — Sorcery {1}{B} (Murders at Karlov Manor, #5193):
//
//	"As an additional cost to cast this spell, you may collect evidence 6.
//	 (Exile cards with total mana value 6 or greater from your graveyard.)
//	 Each opponent sacrifices a creature of their choice. If evidence was
//	 collected, instead each opponent sacrifices a creature with the
//	 greatest power among creatures they control."
//
// The offer is OptionalCollectEvidence; the edict is Kaya's Guile's
// (EachPlayerSacrifices over the opponents), and with evidence collected
// (ctx.EvidenceCollected, CR 607.2i) each opponent's choice is narrowed to
// the creatures tied for their greatest power (GreatestPowerYouControl,
// read for the chooser, as Shadowgrange Archfiend's is). Nothing targets.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "96480e86-70a8-401d-abf8-a7d509311b73",
		Name:          "Extract a Confession",
		Completeness:  CompletenessFull,
		Purpose:       game.Purpose{Answers: game.AnswerRemove},
		OptionalCosts: []game.AdditionalCost{OptionalCollectEvidence(6)},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			edict := EachPlayerSacrifices{ExceptController: true, Match: Creature(), Label: "a creature"}
			if ctx.EvidenceCollected() {
				edict.Match = GreatestPowerYouControl()
				edict.Label = "a creature with the greatest power among creatures you control"
			}
			return edict.Apply(ctx)
		},
	})
}
