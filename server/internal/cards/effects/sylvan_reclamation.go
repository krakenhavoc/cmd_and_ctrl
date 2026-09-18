package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sylvan Reclamation — Instant for {3}{G}{W}:
//
//	"Exile up to two target artifacts and/or enchantments.
//	 Basic landcycling {2}"
//
// S20 sub-PR 5's "up to N" card: Min 0, so it can be cast with no
// targets at all (and then does nothing).
//
// Basic landcycling arrived with #660. Typecycling (CR 702.29e) is a
// cycling ability whose effect searches instead of drawing: "{2},
// Discard this card: Search your library for a basic land card,
// reveal it, put it into your hand, then shuffle." It is an activated
// ability from hand, not a cast, and it is a cycling ability — so a
// "whenever you cycle a card" watcher on the battlefield sees it.
func init() {
	Register(Spec{
		OracleID:     "aeec8e85-6571-4da6-8a48-f5d3985ca10b",
		Name:         "Sylvan Reclamation",
		Completeness: CompletenessFull,
		Targets: TargetPermanent("up to two target artifacts and/or enchantments",
			Or(Artifact(), Enchantment())).WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if err := (ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
		Activated: []ActivatedAbility{BasicLandcycling("{2}")},
	})
}
