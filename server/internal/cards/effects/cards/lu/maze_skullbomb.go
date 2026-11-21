package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Maze Skullbomb — Artifact {1} (Phyrexia: All Will Be One, #5479):
//
//	"{1}, Sacrifice this artifact: Draw a card.
//	 {2}{G}, Sacrifice this artifact: Target creature you control gets
//	 +3/+3 and gains trample until end of turn. Draw a card. Activate
//	 only as a sorcery."
//
// See Basilica Skullbomb.
//
// No simplification.
func init() {
	effects.Register(effects.S82OneSkullbomb("3cb91e96-285d-4191-97cd-ee7ce5d86eb2", "Maze Skullbomb",
		"{2}{G}, Sacrifice this artifact: Target creature you control gets +3/+3 and gains trample until end of turn. Draw a card. Activate only as a sorcery.",
		"{2}{G}", effects.TargetCreature("target creature you control", effects.YouControl()),
		effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3, "trample")),
		func(ctx *effects.Context, id uuid.UUID) error {
			if err := (effects.BoostUntilEOT{Target: id, Power: 3, Toughness: 3, Label: "Maze Skullbomb — +3/+3"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: "Maze Skullbomb — trample"}.Apply(ctx)
		}))
}
