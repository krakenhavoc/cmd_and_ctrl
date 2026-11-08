package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Call a Surprise Witness — Sorcery {1}{W} (Murders at Karlov Manor,
// #4469):
//
//	"Return target creature card with mana value 3 or less from your
//	 graveyard to the battlefield. Put a flying counter on it. It's a
//	 Spirit in addition to its other types."
//
// The return is keyed (a creature whose entry asks a question pauses
// it), and its tail does the rest to the permanent that entered: a
// flying counter put on it after it entered (CR 122.1b; not an "enters
// with" counter), then a layer-4 Spirit type for as long as it stays on
// the battlefield, pinned to that object (CR 611.2a, 400.7) and read
// through the spell's text (CR 612.2), as Ascent of the Worthy's III
// does. If the card is gone, nothing happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9038a457-ce01-4f6d-9e2b-5220f08b4290",
		Name:         "Call a Surprise Witness",
		Completeness: CompletenessFull,
		Targets: TargetCardInGraveyard("target creature card with mana value 3 or less from your graveyard",
			Creature(), ManaValueLE(3), YouOwn()),
		Purpose: ForTargets(Reanimates(0)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			ids := legalTargetCardIDs(ctx)
			if len(ids) == 0 {
				return nil
			}
			return ctx.Game.ReturnFromGraveyardToBattlefieldKeyedForEffect(ids[0], ctx.Controller(), false,
				ctx.ItemTail(callASurpriseWitnessRider))
		},
	})
}

// callASurpriseWitnessRider is the return's tail: a flying counter on
// the creature, and it is a Spirit in addition to its other types. An
// on-disk key.
var callASurpriseWitnessRider = EnteredTail("call-a-surprise-witness/flying-spirit", func(ctx *Context, entered uuid.UUID, _ game.TailOutcome) error {
	g := ctx.Game
	// A graveyard return keeps the card's ID.
	if entered == uuid.Nil || !onBattlefield(g, entered) {
		return nil
	}
	if err := (AddCounter{Target: entered, Kind: game.CounterFlying, N: 1}).Apply(ctx); err != nil {
		return err
	}
	return ScopedEffectFor{
		Target:   entered,
		Mods:     []game.Mod{game.AddSubtypesMod(ctx.TextSubtypes("Spirit")...)},
		Duration: g.PinnedTo(game.IndefiniteDuration(), entered),
		Label:    "Call a Surprise Witness — it's a Spirit in addition to its other types",
	}.Apply(ctx)
})
