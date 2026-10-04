package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Saryth, the Viper's Fang — Legendary Creature — Human Warlock
// {2}{G}{G}, 3/4:
//
//	"Other tapped creatures you control have deathtouch.
//	 Other untapped creatures you control have hexproof.
//	 {1}, {T}: Untap another target creature or land you control."
//
// Two layer-6 grants keyed on Tapped, The Wandering Rescuer's shape:
// the layer listener recomputes on every tap and untap, so a creature
// that taps to attack trades hexproof for deathtouch in the same
// recompute. Neither cares whether Saryth itself is tapped (the
// 2021-09-24 ruling), and neither covers Saryth ("other").
//
// The activated ability is a creature's {T} ability, so Saryth can't
// use it the turn it arrives (CR 302.6). "Another" is object identity
// (Another).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4e07a0b3-f340-4ed1-a8a9-c25fe6f37fb3",
		Name:         "Saryth, the Viper's Fang",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			KeywordGrant(sarythOtherCreatureYouControl(true), "deathtouch"),
			KeywordGrant(sarythOtherCreatureYouControl(false), "hexproof"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Untap another target creature or land you control.",
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Targets: Another(TargetPermanent("another target creature or land you control", And(Or(Creature(), Land()), YouControl()))),
			Effect:  untapFirstLegalTarget,
		}},
	})
}

// sarythOtherCreatureYouControl matches the other creatures Saryth's
// controller controls that are tapped (true) or untapped (false).
func sarythOtherCreatureYouControl(tapped bool) func(target *game.Card, g *game.Game, source *game.Card) bool {
	return func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.InstanceID != source.InstanceID &&
			target.Controller == source.Controller &&
			target.Tapped == tapped && target.IsCreature()
	}
}
