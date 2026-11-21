package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glorious Sunrise — {3}{G}{G} Enchantment (Innistrad: Crimson Vow, S82):
//
//	"At the beginning of combat on your turn, choose one —
//	 • Creatures you control get +1/+1 and gain trample until end of turn.
//	 • Target land gains "{T}: Add {G}{G}{G}" until end of turn.
//	 • Draw a card if you control a creature with power 3 or greater.
//	 • You gain 3 life."
//
// Ferocification's modal beginning-of-combat trigger. Mode one is Lorehold
// Charm's pump. Mode two is Hydro-Man's keyed land mana grant, held for
// the turn (DurationUntilEndOfTurn): the land can tap for {G}{G}{G} as
// often as it can untap. Mode three checks the condition as it resolves
// (an unconditional draw otherwise would be stronger), the power read
// through the layers (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "844c1024-db80-4287-9770-3b10e53d0fc8",
		Name:         "Glorious Sunrise",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:  gloriousSunriseManaGrant,
			Text: "{T}: Add {G}{G}{G}.",
			Mana: []effects.ManaAbility{{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{G}{G}{G}",
				Label:    "{T}: Add {G}{G}{G}",
			}},
		}},
		Triggered: []game.TriggeredAbility{func() game.TriggeredAbility {
			t := effects.AtBeginningOfYourCombat("Glorious Sunrise — choose one",
				func(*game.Game, *game.StackItem) error { return nil })
			t.Modes = effects.ChooseOne(
				effects.ModeDoing("Creatures you control get +1/+1 and gain trample until end of turn.", nil,
					func(_ *game.StackItem, ctx *effects.Context, _ int) error {
						mine := effects.And(effects.Creature(), effects.YouControl())
						if err := (effects.BoostUntilEOT{Match: mine, Power: 1, Toughness: 1, Label: "Glorious Sunrise — +1/+1"}).Apply(ctx); err != nil {
							return err
						}
						return effects.GrantKeywordUntilEOT{Match: mine, Keywords: []string{"trample"}, Label: "Glorious Sunrise — trample"}.Apply(ctx)
					}),
				effects.ModeDoing("Target land gains \"{T}: Add {G}{G}{G}\" until end of turn.",
					effects.TargetPermanent("target land", effects.Land()),
					func(_ *game.StackItem, ctx *effects.Context, occ int) error {
						t, ok := effects.ModeTarget(ctx, occ)
						if !ok {
							return nil
						}
						return effects.ScopedEffectFor{
							Target:   t.ID,
							Mods:     []game.Mod{game.GrantAbilitiesMod(gloriousSunriseManaGrant)},
							Duration: effects.DurationUntilEndOfTurn(ctx),
							Label:    "Glorious Sunrise — land gains {T}: Add {G}{G}{G}",
						}.Apply(ctx)
					}),
				effects.ModeDoing("Draw a card if you control a creature with power 3 or greater.", nil,
					func(_ *game.StackItem, ctx *effects.Context, _ int) error {
						if !effects.B24ControlsCreatureWithPowerAtLeast(ctx.Game, ctx.Controller(), 3) {
							return nil
						}
						return effects.DrawCards{N: 1}.Apply(ctx)
					}),
				effects.ModeDoing("You gain 3 life.", nil,
					func(_ *game.StackItem, ctx *effects.Context, _ int) error {
						return effects.GainLife{Amount: 3}.Apply(ctx)
					}),
			)
			return t
		}()},
	})
}

const gloriousSunriseManaGrant = "glorious-sunrise/land-mana"
