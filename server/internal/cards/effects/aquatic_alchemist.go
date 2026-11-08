package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aquatic Alchemist // Bubble Up — Creature — Elemental {1}{U}, 1/3 //
// Sorcery — Adventure {2}{U} (Wilds of Eldraine, #4471):
//
//	Aquatic Alchemist — "Whenever you cast your first instant or sorcery
//	 spell each turn, this creature gets +2/+0 until end of turn."
//	Bubble Up — "Put target instant or sorcery card from your graveyard
//	 on top of your library. (Then exile this card. You may cast the
//	 creature later from exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID
// and "#1"). The trigger is Arcane Bombardment's condition: the turn's
// cast tally is bumped before the cast event, so it fires only for the
// first instant or sorcery you cast this turn, counting one cast before
// the Alchemist arrived. The pump is the shared self-pump, read when the
// trigger resolves (CR 611.2c). Bubble Up is Mend the Wilds' put on top,
// over an instant or sorcery card; a card that left the graveyard in
// response is dropped by the legality re-check (CR 608.2b). The Adventure
// lifecycle is the engine's (CR 715.3d).
//
// No simplification.
const aquaticAlchemistOracleID = "a78b063a-7f74-465a-9670-34f927f4bfe9"

func init() {
	const label = "Aquatic Alchemist — +2/+0 until end of turn"
	Register(Spec{
		OracleID:     aquaticAlchemistOracleID,
		Name:         "Aquatic Alchemist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast,
				AllOf(YouCast(Or(Instant(), Sorcery())), youCastYourFirstInstantOrSorceryEachTurn),
				label, thisCreatureUntilEOT(label, 2, 0)),
		},
	})
	Register(Spec{
		OracleID:     aquaticAlchemistOracleID + "#1",
		Name:         "Bubble Up",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target instant or sorcery card from your graveyard", YouOwn(), Or(Instant(), Sorcery())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return PutChosenTargetOnTopOfLibrary(ctx.Game, item)
		},
	})
}
