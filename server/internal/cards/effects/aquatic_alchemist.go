package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aquatic Alchemist // Bubble Up — Creature — Elemental {1}{U}, 1/3 //
// Sorcery — Adventure {2}{U} (Wilds of Eldraine, #4471):
//
//	Aquatic Alchemist — "Whenever you cast your first instant or sorcery
//	 spell each turn, this creature gets +2/+0 until end of turn."
//	Bubble Up — "Put target instant or sorcery card from your graveyard on
//	 top of your library. (Then exile this card. You may cast the creature
//	 later from exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID and
// "#1"). The trigger reads the per-turn instant-or-sorcery tally as
// Arcane Bombardment does, so an Adventure's instant or sorcery half
// counts and a creature spell does not. Bubble Up is a target in your
// graveyard re-judged as it resolves (CR 608.2b); the card goes on top of
// its owner's library.
//
// No simplification.
const aquaticAlchemistOracleID = "a78b063a-7f74-465a-9670-34f927f4bfe9"

func init() {
	Register(Spec{
		OracleID:     aquaticAlchemistOracleID,
		Name:         "Aquatic Alchemist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast,
				AllOf(YouCast(Or(Instant(), Sorcery())), youCastYourFirstInstantOrSorceryEachTurn),
				"Aquatic Alchemist — +2/+0 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return BoostUntilEOT{Target: item.SourceCardID, Power: 2, Label: "Aquatic Alchemist — +2/+0"}.Apply(NewContext(g, item))
				}),
		},
	})
	Register(Spec{
		OracleID:     aquaticAlchemistOracleID + "#1",
		Name:         "Bubble Up",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target instant or sorcery card from your graveyard", Or(Instant(), Sorcery()), YouOwn()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					return PutIntoLibrary{Card: t.ID}.Apply(ctx)
				}
			}
			return nil
		},
	})
}
