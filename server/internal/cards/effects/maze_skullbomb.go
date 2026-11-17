package effects

import "github.com/google/uuid"

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
	Register(s82OneSkullbomb("3cb91e96-285d-4191-97cd-ee7ce5d86eb2", "Maze Skullbomb",
		"{2}{G}, Sacrifice this artifact: Target creature you control gets +3/+3 and gains trample until end of turn. Draw a card. Activate only as a sorcery.",
		"{2}{G}", TargetCreature("target creature you control", YouControl()),
		ForTargets(BenefitUntilEOT(0, 3, 3, "trample")),
		func(ctx *Context, id uuid.UUID) error {
			if err := (BoostUntilEOT{Target: id, Power: 3, Toughness: 3, Label: "Maze Skullbomb — +3/+3"}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: "Maze Skullbomb — trample"}.Apply(ctx)
		}))
}
