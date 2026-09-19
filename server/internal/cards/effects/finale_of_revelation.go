package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Finale of Revelation — Sorcery {X}{U}{U}:
//
//	"Draw X cards. If X is 10 or more, instead shuffle your graveyard
//	 into your library, draw X cards, untap up to five lands, and you
//	 have no maximum hand size for the rest of the game.
//	 Exile Finale of Revelation."
//
// One effect with a threshold branch on ctx.X(), not two modes — the
// "instead" makes the two branches mutually exclusive, so this is an
// ordinary `if`, not Spec.Modes.
//
// The exile is unconditional and happens in EITHER branch, and it's
// the resolving card exiling ITSELF — the same "the spell may have
// moved itself" shape Blue Sun's Zenith's self-tuck uses (CR 608.2m):
// Finale is still on the stack when OnResolve runs, so ExileTarget on
// ctx.Source() here is what the printed last sentence means, and the
// engine notices the card already left the stack and skips the
// ordinary graveyard routing.
//
// "You have no maximum hand size for the rest of the game" is a
// PLAYER-level grant with no battlefield dependency — unlike
// Spec.NoMaxHandSize (Reliquary Tower, Thought Vessel), which is
// derived from the battlefield at cleanup and lapses the moment its
// permanent does. This needed a new lock-free setter,
// Game.SetMaxHandSizeForEffect, mirroring the existing (but publicly
// locking, and therefore uncallable from inside a resolution)
// SetMaxHandSize.
//
// Sandbox simplification, declared, weaker than printed:
//
//   - "Untap up to five lands" doesn't say "target" — the choice is
//     made as the spell resolves, not announced — and there's no
//     resolution-time picker for "choose up to N of your own
//     permanents" in the engine yet. The high-X branch untaps up to
//     five of your own tapped lands automatically instead of letting
//     you choose which five, which can only ever be as good as or
//     worse than the printed choice.
func init() {
	Register(Spec{
		OracleID:     "755bd5d8-67f1-4f24-a4e8-d98edf2f2e03",
		Name:         "Finale of Revelation",
		Completeness: CompletenessCaveats,
		XMatters:     true,
		Caveats: []string{
			"When X is 10 or more, you can't choose which lands get untapped — the engine untaps up to five of your own tapped lands for you.",
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			controller := ctx.Controller()
			x := ctx.X()
			if x >= 10 {
				if err := finaleShuffleGraveyardIntoLibrary(ctx, controller); err != nil {
					return err
				}
				if err := (DrawCards{Player: controller, N: x}).Apply(ctx); err != nil {
					return err
				}
				if err := finaleUntapUpToFiveLands(ctx, controller); err != nil {
					return err
				}
				if err := ctx.Game.SetMaxHandSizeForEffect(controller, game.NoMaxHandSize); err != nil {
					return err
				}
			} else {
				if err := (DrawCards{Player: controller, N: x}).Apply(ctx); err != nil {
					return err
				}
			}
			return (ExileTarget{Target: ctx.Source()}).Apply(ctx)
		},
	})
}

// finaleShuffleGraveyardIntoLibrary is "shuffle your graveyard into
// your library" — every card presently in the controller's graveyard
// moves to their library, then the library is shuffled regardless of
// how many cards moved (an empty graveyard still leaves the library
// shuffled, the same as Blue Sun's Zenith shuffling itself in).
//
// The graveyard is snapshotted into an ID slice before anything
// moves — the same reason every mass primitive in mass.go does: the
// zone's slice shifts under a move made while walking it.
func finaleShuffleGraveyardIntoLibrary(ctx *Context, controller uuid.UUID) error {
	p := ctx.Game.PlayerByIDForEffect(controller)
	if p != nil && p.Graveyard != nil {
		ids := make([]uuid.UUID, 0, len(p.Graveyard.Cards))
		for _, c := range p.Graveyard.Cards {
			ids = append(ids, c.InstanceID)
		}
		for _, id := range ids {
			if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneLibrary}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	return ShuffleLibrary{Player: controller}.Apply(ctx)
}

// finaleUntapUpToFiveLands untaps up to five of the controller's own
// tapped lands. See the declared Caveat above: the printed clause
// lets the controller choose which five, and this picks the first
// five it finds on the battlefield instead.
func finaleUntapUpToFiveLands(ctx *Context, controller uuid.UUID) error {
	n := 0
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if n >= 5 {
			break
		}
		if c.Controller != controller || !c.IsLand() || !c.Tapped {
			continue
		}
		if err := (UntapTarget{Target: c.InstanceID}).Apply(ctx); err != nil {
			return err
		}
		n++
	}
	return nil
}
