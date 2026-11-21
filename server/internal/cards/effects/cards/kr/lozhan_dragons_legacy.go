package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lozhan, Dragons' Legacy — Legendary Creature — Dragon Shaman {3}{U}{R}, 4/2
// (Commander Legends: Battle for Baldur's Gate):
//
//	"Flying
//	 Whenever you cast an Adventure or Dragon spell, Lozhan deals damage equal to that
//	 spell's mana value to any target that isn't a commander."
//
// Flying is printed. The trigger is a cast trigger over an Adventure spell (Storyteller
// Pixie's read: the showing half is an Adventure) or a spell with the Dragon creature type
// (the type is a word of Lozhan's text, CR 612.2), with its any-target clause narrowed to
// refuse a commander (Card.IsCommander, CR 903.3): a player, any other creature,
// planeswalker or battle is legal. The damage is the mana value of the spell as it was
// cast, X included (CR 202.3e), dealt by Lozhan as he last existed (CR 608.2h).
//
// No simplification.
func init() {
	targets := effects.TargetAny()
	targets.Label = "any target that isn't a commander"
	cardOK := targets.CardOK
	targets.CardOK = func(g *game.Game, caster uuid.UUID, c game.Card, z game.ZoneKind) bool {
		return !c.IsCommander && cardOK(g, caster, c, z)
	}
	effects.Register(effects.Spec{
		OracleID:        "be2075af-1cd8-42fe-a29b-a0fd2b047a44",
		Name:            "Lozhan, Dragons' Legacy",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverYouCast(effects.Or(effects.AdventureSpellCast(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
					return c.HasSubtypeWord("Dragon")
				}), "Lozhan, Dragons' Legacy — damage equal to that spell's mana value to any target that isn't a commander",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						if item.Trigger == nil || len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) {
							return nil
						}
						ref := ctx.Trigger().Object
						dmg := effects.DealDamage{Source: item.SourceCardID, Target: item.Targets[0].ID, Amount: item.Trigger.Event.Amount}
						if ref != nil {
							r := ref.Ref()
							dmg.SourceObject = &r
						}
						return dmg.Apply(ctx)
					}),
				targets),
		},
	})
}
