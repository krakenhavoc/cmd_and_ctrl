package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Majestic Heliopterus — Creature — Dinosaur {3}{W}, 2/2 (Lost Caverns of
// Ixalan Commander):
//
//	"Flying
//	 Whenever this creature attacks, another target Dinosaur you control
//	 gains flying until end of turn."
//
// The target is another Dinosaur its controller controls, never the
// Heliopterus itself, read through the layered view so a changeling counts.
func init() {
	const label = "Majestic Heliopterus — gains flying until end of turn"
	Register(Spec{
		OracleID:        "d3b49970-46b7-441b-bff8-8403d982b8a7",
		Name:            "Majestic Heliopterus",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacks("Majestic Heliopterus — another target Dinosaur you control gains flying until end of turn",
				Do(OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
					return GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: label}.Apply(ctx)
				}))),
				Another(TargetCreature("another target Dinosaur you control", YouControl(), Subtype("Dinosaur")))),
		},
	})
}
