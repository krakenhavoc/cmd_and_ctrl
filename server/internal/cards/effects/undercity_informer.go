package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Undercity Informer — Creature — Human Rogue {2}{B}, 2/3:
//
//	"{1}, Sacrifice a creature: Target player reveals cards from the
//	 top of their library until they reveal a land card, then puts
//	 those cards into their graveyard."
//
// Consuming Aberration's run (b14MillUntilLand) aimed at one target
// player: a real reveal, then the revealed cards, the land included,
// go to the graveyard in one mill. The Informer may sacrifice itself
// ("a creature", not "another").
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ccf5a0d4-69e8-4607-a76c-cfca336899e4",
		Name:         "Undercity Informer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice a creature: Target player reveals cards from the top of their library until they reveal a land card, then puts those cards into their graveyard.",
			Cost:    Plus(ManaCost("{1}"), SacrificeACreature()),
			Targets: TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetPlayer {
						return b14MillUntilLand(ctx, t.ID)
					}
				}
				return nil
			},
		}},
	})
}
