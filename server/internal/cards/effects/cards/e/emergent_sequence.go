package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Emergent Sequence — Sorcery {1}{G} (Strixhaven: School of Mages, #5502):
//
//	"Search your library for a basic land card, put it onto the battlefield
//	 tapped, then shuffle. That land becomes a 0/0 green and blue Fractal
//	 creature that's still a land. Put a +1/+1 counter on it for each land
//	 you had enter the battlefield under your control this turn."
//
// The search is Fertilid's (a basic land, tapped, shuffle) with a keyed tail
// that runs once the library is settled and the land is on the battlefield, so
// it is a restore point. The tail makes the land what the text says, in the
// order the text says it (CR 608.2c): first an indefinite continuous effect
// pinned to that object (CR 611.2a, 400.7), adding the creature type and the
// Fractal subtype and setting the colours green and blue and the base power
// and toughness to 0/0, with no haste and no loss of its land type; then the
// counters. The count is the per-turn lands-entered tally as it stands after the
// land has entered, so the fetched land itself counts, tapped or not (CR
// 603.10 is not involved: nothing leaves). A search that finds nothing makes
// no creature and no counters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dc859ce2-a267-481f-80d5-2cc38935baa0",
		Name:         "Emergent Sequence",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Lands: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player: ctx.Controller(), Predicate: effects.IsBasicLand, Dest: game.ZoneBattlefield,
				Limit: 1, Shuffle: true, TappedOnEntry: true,
				Reason:   "Emergent Sequence — a basic land, tapped",
				ThenTail: ctx.Tail(emergentSequenceAnimate),
			}.Apply(ctx)
		},
	})
}

// emergentSequenceAnimate is the half of Emergent Sequence after the search:
// the land the search put onto the battlefield becomes a 0/0 green and blue
// Fractal creature that is still a land, then gets a +1/+1 counter for each
// land its controller had enter this turn. An on-disk key.
var emergentSequenceAnimate = effects.CardTail("old-b041/emergent-sequence-animate", func(ctx *effects.Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if !effects.OnBattlefield(ctx.Game, id) {
			continue
		}
		duration := ctx.Game.PinnedTo(game.IndefiniteDuration(), id)
		if err := (effects.ScopedEffectFor{
			Target: id,
			Mods: []game.Mod{
				game.AddTypesMod("Creature"),
				game.AddSubtypesMod(ctx.TextSubtypes("Fractal")...),
				game.SetColorsMod(ctx.TextColors("G", "U")...),
				game.SetBasePowerMod(0),
				game.SetBaseToughnessMod(0),
			},
			Duration: duration,
			Label:    "Emergent Sequence — becomes a 0/0 green and blue Fractal creature that's still a land",
		}).Apply(ctx); err != nil {
			return err
		}
		n := ctx.Game.TurnTallyFor(ctx.Controller()).LandsEntered
		if n <= 0 {
			continue
		}
		if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: n}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	return nil
})
