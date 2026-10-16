package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Filigree Familiar — 2/2 Artifact Creature — Fox for {3}:
//
//	"When Filigree Familiar enters the battlefield, you gain 2 life.
//	When Filigree Familiar dies, draw a card."
//
// S19 sub-PR 4 shipped only the dies half ("draw a card"). The
// #338 stale-simplification sweep adds the ETB-lifegain half, which
// had been "deferred to a later batch" since before ETB triggers
// were routine — Archivist of Oghma and Authority of the Consuls
// were already doing exactly this shape.
//
// cardDied gates the dies trigger to graveyard-only: a bounced or
// exiled Familiar does not draw. The draw happens when the trigger
// resolves.
func init() {
	Register(Spec{
		OracleID:     "b544f690-e4bf-4a5b-984d-9256518fd574",
		Name:         "Filigree Familiar",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			// "When this enters the battlefield, you gain 2 life."
			WhenThisEnters("Filigree Familiar — gain 2 life", Do(GainLife{Amount: 2})),
			WhenThisDies("Filigree Familiar — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
