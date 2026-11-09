package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Esper Origins // Summon: Esper Maduin — Sorcery // Enchantment Creature —
// Saga Elemental {1}{G} (Final Fantasy, #4462):
//
//	Front: "Surveil 2. You gain 2 life. If this spell was cast from a
//	 graveyard, exile it, then put it onto the battlefield transformed under
//	 its owner's control with a finality counter on it. (If a creature with a
//	 finality counter on it would die, exile it instead.)
//	 Flashback {3}{G}"
//	Back: Summon: Esper Maduin, a 4/4 Saga creature, registered under
//	 "<oracle_id>#1" (summon_esper_maduin.go).
//
// The surveil is a prompt, so the life and the rest are its keyed tail and
// the spell waits on the stack for the answer (CR 608.2c, #3613); the tail
// carries whether the spell was cast from a graveyard (any graveyard, not
// only the flashback cast: a Yawgmoth's Will cast counts), read as the spell
// was announced. Then the spell card moves itself (CR 608.2n): it is exiled
// from the stack and returns to the battlefield as a new object with its
// back face up, under its owner's control, entering with a finality counter
// (CR 614.1c), through the ordinary exile-return door so every entry
// replacement sees it. The finality exile is the engine's (CR 122.1h). A
// spell cast from the hand just goes to the graveyard, where its flashback
// can be paid.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         esperOriginsOracleID,
		Name:             "Esper Origins",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{3}{G}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			n := 0
			if item.CastFromZone == game.ZoneGraveyard {
				n = 1
			}
			return Surveil{N: 2, ThenTail: withN(ctx.ItemTail(esperOriginsAfterSurveil), n)}.Apply(ctx)
		},
	})
}

// esperOriginsOracleID is shared with the back face, which registers under
// it plus "#1" (game.CatalogKey).
const esperOriginsOracleID = "947b40a2-e272-4ef6-85fb-3d767a21bc7a"

// esperOriginsAfterSurveil is the surveil answered: gain 2 life, then the
// spell, if it was cast from a graveyard, exiles itself and returns
// transformed. N is 1 when it was. An on-disk key.
var esperOriginsAfterSurveil = CardTail("esper-origins/after-surveil", func(ctx *Context, o game.TailOutcome) error {
	if err := (GainLife{Amount: 2}).Apply(ctx); err != nil {
		return err
	}
	if o.Tail.N != 1 {
		return nil
	}
	return ctx.Game.ExileSpellAndReturnTransformedWithCounterForEffect(ctx.Source(), uuid.Nil, game.CounterFinality, 1)
})
