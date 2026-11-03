package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Volcanic Torrent — Sorcery {4}{R} (Secrets of Strixhaven Commander,
// #4834):
//
//	"Cascade (When you cast this spell, exile cards from the top of your
//	 library until you exile a nonland card that costs less. You may cast it
//	 without paying its mana cost. Put the exiled cards on the bottom in a
//	 random order.)
//	 Volcanic Torrent deals X damage to each creature and planeswalker your
//	 opponents control, where X is the number of spells you've cast this
//	 turn."
//
// Cascade is the keyword trigger (CR 702.85); the cascaded spell resolves
// first, being on top of this one. X is the controller's cast tally for the
// turn read once as the spell resolves (CR 608.2h): Torrent itself, the free
// spell cascade cast and every spell cast earlier this turn all count, and a
// copy of a spell is not a cast. Damage is dealt to every opposing creature
// and planeswalker at once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ec2968fb-163d-4907-862b-b84a1f167833",
		Name:         "Volcanic Torrent",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Cascade()},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return damageEachMatching(ctx,
				And(Or(Creature(), Planeswalker()), OpponentControls()),
				ctx.Game.CastTallyFor(ctx.Controller()).Total)
		},
	})
}
