package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "844c1024-db80-4287-9770-3b10e53d0fc8",
		Name:         "Glorious Sunrise",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:  gloriousSunriseManaGrant,
			Text: "{T}: Add {G}{G}{G}.",
			Mana: []ManaAbility{{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{G}{G}{G}",
				Label:    "{T}: Add {G}{G}{G}",
			}},
		}},
		Triggered: []game.TriggeredAbility{func() game.TriggeredAbility {
			t := AtBeginningOfYourCombat("Glorious Sunrise — choose one",
				func(*game.Game, *game.StackItem) error { return nil })
			t.Modes = ChooseOne(
				ModeDoing("Creatures you control get +1/+1 and gain trample until end of turn.", nil,
					func(_ *game.StackItem, ctx *Context, _ int) error {
						mine := And(Creature(), YouControl())
						if err := (BoostUntilEOT{Match: mine, Power: 1, Toughness: 1, Label: "Glorious Sunrise — +1/+1"}).Apply(ctx); err != nil {
							return err
						}
						return GrantKeywordUntilEOT{Match: mine, Keywords: []string{"trample"}, Label: "Glorious Sunrise — trample"}.Apply(ctx)
					}),
				ModeDoing("Target land gains \"{T}: Add {G}{G}{G}\" until end of turn.",
					TargetPermanent("target land", Land()),
					func(_ *game.StackItem, ctx *Context, occ int) error {
						t, ok := ModeTarget(ctx, occ)
						if !ok {
							return nil
						}
						return ScopedEffectFor{
							Target:   t.ID,
							Mods:     []game.Mod{game.GrantAbilitiesMod(gloriousSunriseManaGrant)},
							Duration: DurationUntilEndOfTurn(ctx),
							Label:    "Glorious Sunrise — land gains {T}: Add {G}{G}{G}",
						}.Apply(ctx)
					}),
				ModeDoing("Draw a card if you control a creature with power 3 or greater.", nil,
					func(_ *game.StackItem, ctx *Context, _ int) error {
						if !b24ControlsCreatureWithPowerAtLeast(ctx.Game, ctx.Controller(), 3) {
							return nil
						}
						return DrawCards{N: 1}.Apply(ctx)
					}),
				ModeDoing("You gain 3 life.", nil,
					func(_ *game.StackItem, ctx *Context, _ int) error {
						return GainLife{Amount: 3}.Apply(ctx)
					}),
			)
			return t
		}()},
	})
}

const gloriousSunriseManaGrant = "glorious-sunrise/land-mana"
