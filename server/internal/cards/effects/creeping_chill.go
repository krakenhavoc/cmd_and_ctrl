package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Creeping Chill — Sorcery {3}{B} (Ravnica Remastered):
//
//	"Creeping Chill deals 3 damage to each opponent and you gain 3 life.
//	 When Creeping Chill is put into your graveyard from your library, you
//	 may exile it. If you do, Creeping Chill deals 3 damage to each opponent
//	 and you gain 3 life."
//
// The spell is the first sentence. The second is a trigger from the
// graveyard (Narcomoeba's shape): it asks, exiles the card, and only if the
// exile happened drains (CR 603.12-style "if you do"), through a keyed tail
// told the exiled card. A card that has left the graveyard by resolution is
// a new object and nothing is exiled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dc0d4e58-db39-4a05-b27c-452180c86d47",
		Name:         "Creeping Chill",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{DamageEachOpponent: 3, LifeGain: 3},
		OnResolve:    ResolveSteps(DamageEachOpponent{N: 3}, GainLife{Amount: 3}),
		Triggered: []game.TriggeredAbility{
			Optional(WhenThisIsPutIntoYourGraveyardFromYourLibrary(
				"Creeping Chill — exile it; if you do, 3 damage to each opponent and you gain 3 life",
				creepingChillExile), "Creeping Chill — exile it?"),
		},
	})
}

func creepingChillExile(g *game.Game, item *game.StackItem) error {
	id := item.SourceCardID
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return ExileTarget{Target: id, ThenTail: NewContext(g, item).ItemTail(creepingChillDrain)}.Apply(NewContext(g, item))
}

// creepingChillDrain is "if you do": it runs only when the card reached
// exile (the tail is told the card, and nothing when it did not).
var creepingChillDrain = CardTail("then/creeping-chill-drain", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return Do(DamageEachOpponent{N: 3}, GainLife{Amount: 3})(ctx.Game, ctx.Item)
})
