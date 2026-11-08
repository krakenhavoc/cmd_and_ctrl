package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rot-Curse Rakshasa — Creature — Demon {1}{B}, 5/5 (Tarkir: Dragonstorm,
// #4463):
//
//	"Trample
//	 Decayed (This creature can't block. When it attacks, sacrifice it at
//	 end of combat.)
//	 Renew — {X}{B}{B}, Exile this card from your graveyard: Put a decayed
//	 counter on each of X target creatures. Activate only as a sorcery."
//
// Decayed is the engine's keyword (game/decayed.go, CR 702.147). The renew
// ability is renew.go's, with exactly X targets (a CountFromX clause, as
// Deepwood Elder's). A decayed counter is a keyword counter (CR 122.1b):
// each creature that still has one has decayed, so it can't block and is
// sacrificed at end of combat after it attacks. A target gone by
// resolution gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	targets := TargetCreature("X target creatures")
	targets.CountFromX = true
	Register(Spec{
		OracleID:        "9fee2572-8864-475c-bade-a64e36ef69c4",
		Name:            "Rot-Curse Rakshasa",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", game.KeywordDecayed},
		Activated: []ActivatedAbility{
			Renew("{X}{B}{B}", "Put a decayed counter on each of X target creatures.", targets,
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (AddCounter{Target: t.ID, Kind: game.CounterDecayed, N: 1}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
