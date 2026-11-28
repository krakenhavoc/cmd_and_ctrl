package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Avenging Angel — Creature — Angel {3}{W}{W}, 3/3 (Tempest, S82 OLD batch 768,
// #5628):
//
//	"Flying
//	 When this creature dies, you may put it on top of its owner's library."
//
// Dark Revenant's trigger made optional: the card is moved only if it is still
// that card in its owner's graveyard when the ability resolves (CR 400.7), so a
// graveyard hate card that exiled it in response leaves nothing to put back. The
// "may" is the trigger's own (CR 603.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3415a199-05ac-4627-be97-f842eb1415a3",
		Name:            "Avenging Angel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.WhenThisDies("Avenging Angel — you may put it on top of its owner's library", effects.PutThisDiedCardOnTopOfItsOwnersLibrary),
				"Avenging Angel — put it on top of its owner's library?"),
		},
	})
}
