package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kellogg, Dangerous Mind — Legendary Creature — Human Mercenary
// {1}{B}{R}, 3/2 (Smaug Unchained deck, #4724):
//
//	"First strike, haste
//	 Whenever Kellogg attacks, create a Treasure token.
//	 Sacrifice five Treasures: Gain control of target creature for as
//	 long as you control Kellogg. Activate only as a sorcery."
//
// Three pieces the catalog already has:
//
//   - First strike and haste ride PrintedKeywords.
//   - The attack trigger is Captain Lannery Storm's (WheneverThisAttacks
//     over CreateToken).
//   - The activated ability is Magda, Brazen Outlaw's cost, SacrificeN of
//     five Treasures (read by subtype, so a Treasure that is also
//     something else still pays), at sorcery speed, over Rangers of
//     Ithilien's control change: GainControlOfTargetFor with
//     DurationWhileYouControlSource (CR 611.2b, layer 2). The steal ends
//     the moment the ability's controller stops controlling Kellogg,
//     whether he leaves the battlefield or changes hands. A Kellogg that
//     left the battlefield (or left and came back as a new object) before
//     the ability resolved takes nothing, because the duration has
//     already ended (CR 611.2b). "Target creature" is any creature,
//     your own included.
//
// No simplification.
func init() {
	const label = "Sacrifice five Treasures: Gain control of target creature for as long as you control Kellogg. Activate only as a sorcery."
	effects.Register(effects.Spec{
		OracleID:        "904b2603-00c3-44d2-a43c-c25439dd8038",
		Name:            "Kellogg, Dangerous Mind",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Kellogg, Dangerous Mind — create a Treasure",
				effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        label,
			Cost:         effects.SacrificeN(5, "five Treasures", effects.IsTreasure),
			SorcerySpeed: true,
			Targets:      effects.TargetCreature("target creature"),
			Effect: effects.GainControlOfTargetFor("Kellogg, Dangerous Mind — control for as long as you control Kellogg",
				func(ctx *effects.Context) (game.Duration, bool) {
					return effects.DurationWhileYouControlSource(ctx, ctx.Source(), ctx.Controller())
				}),
		}},
	})
}
