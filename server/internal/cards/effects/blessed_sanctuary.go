package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blessed Sanctuary — Enchantment {3}{W}{W} (Jumpstart Selesnya deck
// request, #4893):
//
//	"Prevent all noncombat damage that would be dealt to you and creatures
//	 you control.
//	 Whenever a nontoken creature you control enters, create a 2/2 white
//	 Unicorn creature token."
//
// The prevention is two plain prevention statics over noncombat damage, one
// for its controller and one for the creatures they control (Rune-Tail's
// Essence's shape), so combat damage is untouched and damage that can't be
// prevented still gets through (CR 615.12). The Unicorn trigger watches
// every nontoken creature entering under its controller's control; a token
// entering, the Unicorns included, does not fire it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1b70d7fa-ba2c-4945-b8fd-23623b88c036",
		Name:         "Blessed Sanctuary",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			PreventAllDamageDealtTo(PreventionStatic{
				To:     ToYou,
				Damage: NoncombatDamage,
				Label:  "Blessed Sanctuary — prevent all noncombat damage that would be dealt to you",
			}),
			PreventAllDamageDealtTo(PreventionStatic{
				To:     ToCreaturesYouControl,
				Damage: NoncombatDamage,
				Label:  "Blessed Sanctuary — prevent all noncombat damage that would be dealt to creatures you control",
			}),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature() && !IsToken(c)
			}, "Blessed Sanctuary — create a 2/2 white Unicorn", Do(CreateToken{Template: TokenCard("2/2 white Unicorn"), N: 1})),
		},
	})
}
