package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Runed Arch — Artifact {3} (Ice Age, S82 OLD batch 846, #5636):
//
//	"This artifact enters tapped.
//	 {X}, {T}, Sacrifice this artifact: X target creatures with power 2 or less
//	 can't be blocked this turn."
//
// It enters tapped (CR 614.12), so the {T} waits for the next untap. The number
// of targets is the announced X (Mishra's Helix: game.TargetSpec.CountFromX),
// each a creature whose power is 2 or less when targeted and again as the
// ability resolves (CR 608.2b); a target that left or grew past that is
// skipped. Each one that is still legal gets Rogue's Passage's can't-be-blocked
// restriction until end of turn, pinned to that object (CR 611.2c).
//
// No simplification.
func init() {
	targets := effects.TargetCreature("X target creatures with power 2 or less", effects.PowerLE(2))
	targets.CountFromX = true
	effects.Register(effects.Spec{
		OracleID:     "32717625-46c6-45c5-ad0a-fb6a91f9878c",
		Name:         "Runed Arch",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}, {T}, Sacrifice this artifact: X target creatures with power 2 or less can't be blocked this turn.",
			Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    effects.Plus(effects.ManaCost("{X}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: targets,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (effects.RestrictUntilEOT{Target: t.ID, Restrictions: game.CantBeBlocked, Label: "Runed Arch — can't be blocked"}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
