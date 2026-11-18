package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82MidB09TapThreeCreatures is "Tap three untapped creatures you control" as
// an activated cost (Skaab Wrangler, Larder Zombie, Siege Zombie). It is a
// TapOthers component, not the {T} symbol, so the source may tap itself as one
// of the three and a creature that arrived this turn may pay (CR 302.6).
func s82MidB09TapThreeCreatures() game.AbilityCost {
	const label = "three untapped creatures you control"
	return game.AbilityCost{TapOthers: &game.TapOthersCost{
		Count:  3,
		Filter: TargetPermanent(label, Creature()),
		Label:  label,
	}}
}
