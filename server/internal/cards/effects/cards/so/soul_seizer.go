package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soul Seizer // Ghastly Haunting — Creature — Spirit {3}{U}{U}, 1/3, flying
// (Dark Ascension, S82 OLD, #5571); Enchantment — Aura:
//
//	Front: "Flying
//	        When this creature deals combat damage to a player, you may
//	        transform it. If you do, attach it to target creature that player
//	        controls."
//	Back:  "Enchant creature
//	        You control enchanted creature."
//
// Mistblade Shinobi's combat-damage trigger (the target is a creature that
// player controls, chosen as the trigger goes on the stack, CR 603.3d) with
// the "you may" asked as it resolves. A "yes" turns the Seizer over (the same
// object, CR 712.18) and attaches the Aura face to the target as an
// instruction of the ability, not a cast: the Aura face's own enchant clause
// decides whether the target is a legal host, and an Aura that can't
// legally enchant it is put into the graveyard as a state-based action
// (CR 704.5m). While it stays, the Aura is Mind Control's control effect. A
// target that is gone or illegal by resolution leaves the Seizer face up;
// a Seizer that has left the battlefield is not transformed.
//
// No simplification.
func init() {
	const oracle = "98950aaf-d0b8-420b-9e6a-c7d07d0f8588"
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Soul Seizer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.WheneverThisDealsCombatDamageToAPlayer(
				"Soul Seizer — transform it and attach it to target creature that player controls",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					t, ok := effects.FirstLegalTarget(ctx)
					if !ok || t.Kind != game.TargetCard || !effects.SourceIsStillThisPermanent(g, item) {
						return nil
					}
					if err := (effects.TransformThis{}).Apply(ctx); err != nil {
						return err
					}
					// "If you do": only a Seizer that really turned over is attached.
					if c, ok := g.LookupCardForEffect(item.SourceCardID); !ok || !effects.TransformedPermanent(c) {
						return nil
					}
					return g.AttachForEffect(item.SourceCardID, t)
				}),
				effects.ControlledByThatPlayer(effects.TargetCreature("target creature that player controls"))),
				"Soul Seizer — transform it and attach it to target creature that player controls?"),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Ghastly Haunting",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static:       []game.StaticAbility{effects.ControlAttachedBySource()},
	})
}
