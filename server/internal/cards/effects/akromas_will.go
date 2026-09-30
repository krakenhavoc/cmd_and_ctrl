package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Akroma's Will — Instant {3}{W}:
//
//	"Choose one. If you control a commander as you cast this spell,
//	 you may choose both instead.
//	 • Creatures you control gain flying, vigilance, and double
//	   strike until end of turn.
//	 • Creatures you control gain lifelink, indestructible, and
//	   protection from each color until end of turn."
//
// The white finisher-or-fog, and the reason it sat blocked through
// four batches: the first bullet needed until-end-of-turn continuous
// effects (#279, S32) and the second needed protection to be a real
// quality rather than a badge (#662). Both landed, so both bullets
// are ordinary mass layer-6 grants now — `GrantKeywordUntilEOT` with
// a snapshotted affected set (CR 611.2c), so a creature cast after
// this resolves gets nothing.
//
// "Protection from each color" is five keywords, one per colour, and
// they are the same tokens a printed "protection from white" mints —
// `game.ProtectionFromColor`. Together they are DEBT for every
// coloured source: damage, enchanting and equipping, blocking, and
// targeting (CR 702.16b–e). Colourless sources go straight through,
// which is why an Ugin still kills the team and is the difference
// between this and "protection from everything".
//
// "If you control a commander as you cast this spell, you may choose
// both instead" is the conditional mode count #1590 built:
// `OrUpToIf(2, YouControlACommander)`, read at announce (CR 601.2b)
// and fixed from then on. Until then the card asked for one bullet
// always. The two grants touch the same snapshotted creatures and
// neither reads the other, so the order they run in is not
// observable.
func init() {
	Register(Spec{
		OracleID:     "fd949f82-fc10-4e37-8aa9-6c7569fe3c55",
		Name:         "Akroma's Will",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Creatures you control gain flying, vigilance, and double strike until end of turn.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return GrantKeywordUntilEOT{
						Match:    And(Creature(), YouControl()),
						Keywords: []string{"flying", "vigilance", "double strike"},
						Label:    "Akroma's Will — flying, vigilance and double strike",
					}.Apply(ctx)
				}),
			ModeDoing("Creatures you control gain lifelink, indestructible, and protection from each color until end of turn.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return GrantKeywordUntilEOT{
						Match:    And(Creature(), YouControl()),
						Keywords: akromasWillProtectiveKeywords(),
						Label:    "Akroma's Will — lifelink, indestructible and protection from each color",
					}.Apply(ctx)
				}),
		).OrUpToIf(2, YouControlACommander),
	})
}

// akromasWillProtectiveKeywords is the second bullet's grant:
// lifelink, indestructible, and one protection quality per colour
// (CR 702.16 — "protection from each color" is five separate
// abilities, not one).
func akromasWillProtectiveKeywords() []string {
	out := []string{"lifelink", "indestructible"}
	for _, c := range game.AllColors {
		out = append(out, game.ProtectionFromColor(c))
	}
	return out
}
