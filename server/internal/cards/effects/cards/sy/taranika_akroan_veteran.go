package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Taranika, Akroan Veteran — Legendary Creature — Human Soldier {1}{W}{W}, 3/3
// (Jumpstart 2022):
//
//	"Vigilance
//	 Whenever Taranika attacks, untap another target creature you control.
//	 Until end of turn, that creature has base power and toughness 4/4 and
//	 gains indestructible."
//
// The target is chosen as the trigger goes on the stack ("another" is object
// identity, Another). On resolution the creature untaps, takes base size 4/4
// for the turn as a layer 7b effect (Hostile Takeover's, so counters and pumps
// still apply on top) and gains indestructible. A target gone by resolution
// gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1e3c1920-7bf5-44f9-b1ff-b17d6420e69b",
		Name:            "Taranika, Akroan Veteran",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverThisAttacks("Taranika, Akroan Veteran — untap another target creature you control; it has base power and toughness 4/4 and gains indestructible until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							if t.Kind != game.TargetCard {
								continue
							}
							if err := (effects.UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
								return err
							}
							if err := (effects.ScopedEffectFor{
								Target:   t.ID,
								Mods:     []game.Mod{game.SetBasePowerMod(4), game.SetBaseToughnessMod(4)},
								Duration: effects.DurationUntilEndOfTurn(ctx),
								Label:    "Taranika — base power and toughness 4/4",
							}).Apply(ctx); err != nil {
								return err
							}
							return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"indestructible"},
								Label: "Taranika — indestructible"}.Apply(ctx)
						}
						return nil
					}),
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl())),
			),
		},
	})
}
