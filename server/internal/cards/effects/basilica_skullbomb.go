package effects

import "github.com/google/uuid"

// Basilica Skullbomb — Artifact {1} (Phyrexia: All Will Be One, #5479):
//
//	"{1}, Sacrifice this artifact: Draw a card.
//	 {2}{W}, Sacrifice this artifact: Target creature you control gets
//	 +2/+2 and gains flying until end of turn. Draw a card. Activate only
//	 as a sorcery."
//
// The Skullbomb shape is s82OneSkullbomb's: the sacrifice is a cost, the
// card is drawn after the pump, and an illegal target stops the whole
// ability (CR 608.2b).
//
// No simplification.
func init() {
	Register(s82OneSkullbomb("43057d84-b424-489c-abec-3fb4cb823d80", "Basilica Skullbomb",
		"{2}{W}, Sacrifice this artifact: Target creature you control gets +2/+2 and gains flying until end of turn. Draw a card. Activate only as a sorcery.",
		"{2}{W}", TargetCreature("target creature you control", YouControl()),
		ForTargets(BenefitUntilEOT(0, 2, 2, "flying")),
		func(ctx *Context, id uuid.UUID) error {
			if err := (BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Basilica Skullbomb — +2/+2"}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "Basilica Skullbomb — flying"}.Apply(ctx)
		}))
}
