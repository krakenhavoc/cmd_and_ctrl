package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wrath of God — "Destroy all creatures. They can't be regenerated."
//
// S23: rewritten onto DestroyAllMatching. The behavioural change is
// not the destruction — it is that the creatures now die as ONE
// event, so a Blood Artist caught in the wrath drains for every
// creature that died with it instead of for however many happened to
// be processed after it. See mass.go.
//
// "They can't be regenerated" is ENFORCED as of #667
// (CR 701.19c): a creature with a regeneration shield on it is
// destroyed anyway, and CR 701.19c leaves the shield unspent. The
// rider is where this card and Day of Judgment differ, which is why
// they no longer share one helper.
func init() {
	Register(Spec{
		OracleID:     "34515b16-c9a4-4f98-8c77-416a7a523407",
		Name:         "Wrath of God",
		Completeness: CompletenessFull,
		OnResolve:    wrathDestroyAllCreaturesNoRegen,
	})
}

// wrathDestroyAllCreatures is shared between Day of Judgment, Supreme
// Verdict and Vanquish the Horde — the printings of "destroy all
// creatures" that do NOT say the creatures can't be regenerated.
// Extracted as a package-local func so the per-card files stay thin.
//
// #667 split the riderful printings off into the sibling below. The
// two used to share this one body on the argument that "once the
// uncounterable and can't-be-regenerated riders are accounted for"
// the four cards are one effect — true, and the accounting is what
// changed: the regeneration rider is now a real difference between
// Wrath of God and Day of Judgment, and a shared helper could not
// express it.
func wrathDestroyAllCreatures(_ *game.StackItem, ctx *Context) error {
	return DestroyAllMatching{Match: Creature()}.Apply(ctx)
}

// wrathDestroyAllCreaturesNoRegen is the same sweep for the printings
// that DO print "They can't be regenerated" — Wrath of God and
// Damnation (CR 701.19c).
func wrathDestroyAllCreaturesNoRegen(_ *game.StackItem, ctx *Context) error {
	return DestroyAllMatching{Match: Creature(), CantBeRegenerated: true}.Apply(ctx)
}
