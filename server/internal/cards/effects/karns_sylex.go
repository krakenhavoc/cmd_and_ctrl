package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Karn's Sylex — Legendary Artifact {3} (DMU):
//
//	"Karn's Sylex enters tapped.
//	 Players can't pay life to cast spells or to activate abilities
//	 that aren't mana abilities.
//	 {X}, {T}, Exile Karn's Sylex: Destroy each nonland permanent with
//	 mana value X or less. Activate only as a sorcery."
//
// The tapped entry is the CR 614 self-replacement, so the Sylex cannot
// be activated the turn it arrives.
//
// The second line is PlayersCantPayLifeToCastOrActivate (#5660): every
// player, the Sylex's controller included, is refused the life half of
// a Phyrexian symbol, a "pay N life" additional or alternative cost, a
// targeting tax paid in life and a printed life cost on an activated
// ability. Mana abilities are exempt as printed (Mana Confluence still
// taps), and so is life paid for anything that is neither a cast nor an
// activation: a ward cost, a land entering, a "pay 2 life unless".
//
// The last ability exiles the Sylex as a cost (CR 602.2b), so it is gone
// before anyone can respond, and its own second line is gone with it.
// The sweep reads each permanent's mana value on the battlefield, where
// a token is 0, and destroys them at once through DestroyAllMatching.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c5e817f9-a003-45b5-8e16-1746e6096166",
		Name:         "Karn's Sylex",
		XMatters:     true,
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		Static:       []game.StaticAbility{PlayersCantPayLifeToCastOrActivate()},
		Activated: []ActivatedAbility{{
			Label: "{X}, {T}, Exile Karn's Sylex: Destroy each nonland permanent with mana value X or less. Activate only as a sorcery.",
			// ADR 0126 §6: only mana value X or less.
			Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy, Partial: true}},
			Cost:         Plus(ManaCost("{X}"), TapCost(), ExileThis()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return DestroyAllMatching{Match: And(Nonland(), ManaValueLE(ctx.X()))}.Apply(ctx)
			},
		}},
	})
}
