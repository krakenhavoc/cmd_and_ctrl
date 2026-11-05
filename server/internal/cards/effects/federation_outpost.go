package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Federation Outpost — Land (Star Trek, #4973):
//
//	"This land enters tapped unless you control two or more other lands.
//	 When this land enters untapped, choose one —
//	 • You gain 2 life.
//	 • Surveil 1.
//	 {T}: Add {W}."
//
// The slowland condition (otherLandsAtLeast(2)) as the enters-tapped
// self-replacement, then a modal enters trigger gated on the land being
// untapped once it is on the battlefield (selfEnteredUntapped, the
// Mystic Sanctuary reading): entered tapped, no trigger at all. The
// mode is chosen as the trigger goes on the stack (CR 700.2b).
//
// No simplification.
func init() {
	etb := On(game.EventETB, selfEnteredUntapped, "Federation Outpost — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	etb.Modes = ChooseOne(
		ModeDoing("You gain 2 life.", nil, func(item *game.StackItem, ctx *Context, _ int) error {
			return GainLife{Amount: 2}.Apply(ctx)
		}),
		ModeDoing("Surveil 1.", nil, func(item *game.StackItem, ctx *Context, _ int) error {
			return Surveil{Player: item.Controller, N: 1}.Apply(ctx)
		}),
	)
	Register(Spec{
		OracleID:     "a8bcc930-3efe-4595-b164-f5985e1dc289",
		Name:         "Federation Outpost",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{EntersTappedUnless(otherLandsAtLeast(2))},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
		Triggered: []game.TriggeredAbility{etb},
	})
}
