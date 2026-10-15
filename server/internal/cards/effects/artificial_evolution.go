package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Artificial Evolution — Instant {U}:
//
//	"Change the text of target spell or permanent by replacing all
//	 instances of one creature type with another. The new creature type
//	 can't be Wall. (This effect lasts indefinitely.)"
//
// The creature-type class of CR 612.2's change (text_words.go, ADR
// 0160's 2026-10-15 amendment). Both words come from the choose a
// creature type prompt; the second refuses Wall and the word being
// replaced. The change reaches the type line (a Goblin changed to an
// Elf is an Elf), a target clause that names a type through its source
// ("target Goblin"), and a lord's printed tribes (Goblin King becomes a
// lord of Elves). A type a lord or trigger CHOSE as it entered (Cavern
// of Souls, Vanquisher's Banner) is not text and stays.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c5ab6dc9-34af-434f-ae8e-b2389ee6c5ae",
		Name:         "Artificial Evolution",
		Completeness: CompletenessFull,
		Targets:      TargetSpellOrPermanent("target spell or permanent", nil, nil),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return ChangeCreatureTypeWord{Target: t.ID, Except: []string{"Wall"}}.Apply(ctx)
			}
			return nil
		},
	})
}
