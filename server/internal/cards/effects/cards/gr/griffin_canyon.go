package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Griffin Canyon — Land (Visions, S82 older sets batch 792, #5631):
//
//	"{T}: Add {C}.
//	 {T}: Untap target Griffin. If it's a creature, it gets +1/+1 until end of
//	 turn."
//
// "Griffin" is the creature type, which a permanent need not be a creature to
// carry (a Griffin Aura-turned-land, an animated artifact with the type): the
// target is any Griffin permanent, and the +1/+1 happens only if it is a creature
// as the ability resolves, read through the layers (CR 608.2h). The type is a word of
// the Canyon's text (CR 612.2). Tapping the Canyon for this is a cost, so the Canyon
// can target itself only if it is a Griffin, which it is not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ba642c8b-9ade-4501-8393-672fd53d4955",
		Name:         "Griffin Canyon",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Untap target Griffin. If it's a creature, it gets +1/+1 until end of turn.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPermanent("target Griffin", effects.HasSubtype("Griffin")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id := effects.FirstLegalCardTarget(ctx)
				if id == uuid.Nil {
					return nil
				}
				if err := (effects.UntapTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
				if c, ok := g.LookupCardForEffect(id); ok && c.IsCreature() {
					return effects.BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Griffin Canyon — +1/+1"}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
