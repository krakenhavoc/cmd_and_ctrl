package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Farewell — Sorcery {4}{W}{W} (EDHREC rank 172):
//
//	"Choose one or more —
//	 • Exile all artifacts.
//	 • Exile all creatures.
//	 • Exile all enchantments.
//	 • Exile all graveyards."
//
// The most-played sweeper in the format because it is the most
// selective one: a Treasure deck exiles creatures and keeps its
// rocks, an enchantress board wipes everything but enchantments,
// and every mode is EXILE — no dies-triggers, no recursion, no
// indestructible (see merciless_eviction.go for why that decides
// games against the catalog's aristocrats payoffs).
//
// "Choose one or more" is ChooseN with Min 1 and Max 4 — every
// subset of the four is a legal cast. The three battlefield modes
// are Merciless Eviction's, resolved in printed order (CR 608.2c);
// the fourth walks every seat's graveyard, the caster's included,
// through the same helper Bojuka Bog uses.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4eb813fd-2d5a-4b02-8193-662681ef4e7d",
		Name:         "Farewell",
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose one or more", 1, 4,
			Mode("Exile all artifacts."),
			Mode("Exile all creatures."),
			Mode("Exile all enchantments."),
			Mode("Exile all graveyards."),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			sweeps := []CardPredicate{Artifact(), Creature(), Enchantment()}
			for i, match := range sweeps {
				if !ctx.HasMode(i) {
					continue
				}
				if err := (ExileAllMatching{Match: match}).Apply(ctx); err != nil {
					return err
				}
			}
			if ctx.HasMode(3) {
				for _, p := range ctx.Game.Seats {
					if p == nil {
						continue
					}
					if err := exileGraveyardForEffect(ctx.Game, item, p.ID); err != nil {
						return err
					}
				}
			}
			return nil
		},
	})
}
