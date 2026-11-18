package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dancing Sword — Artifact — Equipment {1}{W} (Adventures in the Forgotten
// Realms Commander):
//
//	"Equipped creature gets +2/+1.
//	 When equipped creature dies, you may have this Equipment become a 2/1
//	 Construct artifact creature with flying and ward {1}. If you do, it isn't
//	 an Equipment.
//	 Equip {1}"
//
// "Dies" is Skullclamp's watch (the unattach SBA has not run when the trigger
// is harvested). The "you may" is asked as the ability resolves. The change
// has no duration (CR 611.2a), so it is one indefinite effect on that object
// at one timestamp (CR 613.7): layer 4 adds the Creature type and the
// Construct subtype and takes the Equipment subtype away, layer 6 gives
// flying and a ward {1} bundle (ADR 0093), layer 7b sets 2/1. It follows the
// Sword only while it stays the same object (CR 400.7); one that has left the
// battlefield since the trigger did nothing. No longer an Equipment, it can't
// stay attached to a creature (CR 301.5c, CR 704.5n).
//
// No simplification.
const dancingSwordWard = "dancing-sword/construct-ward"

func init() {
	Register(Spec{
		OracleID:     "da3b12d1-8867-4063-b50d-3b2f5ce0549f",
		Name:         "Dancing Sword",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(2, 1)},
		Grants: []AbilityGrant{{
			Key:       dancingSwordWard,
			Triggered: []game.TriggeredAbility{Ward(WardMana("{1}"), "Dancing Sword — ward {1}")},
			Text:      "Ward {1}",
		}},
		Triggered: []game.TriggeredAbility{
			Optional(
				On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return equippedCreatureDied(ev, source)
				}, "Dancing Sword — you may have this Equipment become a 2/1 Construct artifact creature with flying and ward {1}",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						if ctx.isNewSourceObjectAsThis(ctx.Source()) {
							return nil
						}
						if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneBattlefield {
							return nil
						}
						mods := []game.Mod{
							game.AddTypesMod("Creature"),
							game.AddSubtypesMod(ctx.TextSubtype("Construct")),
							game.LoseSubtypesMod("Equipment"),
							game.AddKeywordsMod("flying"),
							game.GrantAbilitiesMod(dancingSwordWard),
						}
						mods = append(mods, game.SetBasePTMods(2, 1)...)
						return ScopedEffectFor{
							Target:   item.SourceCardID,
							Mods:     mods,
							Duration: game.IndefiniteDuration(),
							Label:    "Dancing Sword — becomes a 2/1 Construct artifact creature with flying and ward {1}; it isn't an Equipment",
						}.Apply(ctx)
					}),
				"Dancing Sword — have this Equipment become a 2/1 Construct artifact creature?"),
		},
		Activated: []ActivatedAbility{EquipAbility("{1}")},
	})
}
