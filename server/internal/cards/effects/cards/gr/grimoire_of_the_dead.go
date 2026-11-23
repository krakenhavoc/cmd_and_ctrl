package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grimoire of the Dead — Legendary Artifact — Book {4} (Commander Starter
// Decks):
//
//	"{1}, {T}, Discard a card: Put a study counter on Grimoire of the Dead.
//	 {T}, Remove three study counters from Grimoire of the Dead and sacrifice
//	 it: Put all creature cards from all graveyards onto the battlefield
//	 under your control. They're black Zombies in addition to their other
//	 colors and types."
//
// The first ability is Rummaging Goblin's looter cost with a counter in
// place of the draw. The second pays its counters and the sacrifice as costs
// (CR 118.8, so the Grimoire is gone before the ability resolves) and is
// Rise of the Dark Realms' return: every graveyard's creature cards, read
// as the ability resolves, enter together under your control. Each one then
// gets Liliana, Death's Majesty's "black Zombie in addition" effect pinned to
// that object, as Ghouls' Night Out's riders do.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "922c6e1b-abcf-40e0-847e-9ebb7d0b6fd4",
		Name:         "Grimoire of the Dead",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}, {T}, Discard a card: Put a study counter on Grimoire of the Dead.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.DiscardACard()),
				Effect:  effects.Do(effects.CounterOnThis{Kind: "study", N: 1}),
			},
			{
				Label:   "{T}, Remove three study counters from Grimoire of the Dead and sacrifice it: Put all creature cards from all graveyards onto the battlefield under your control. They're black Zombies in addition to their other colors and types.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.TapCost(), effects.RemoveCountersFromThis("study", 3), effects.SacrificeThis()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.ReturnFromGraveyardTogether{
						Targets:    effects.AllGraveyardsCardIDs(ctx, game.Card.IsCreature),
						Controller: ctx.Controller(),
						ThenTail:   grimoireOfTheDeadZombies,
					}.Apply(ctx)
				},
			},
		},
	})
}

// grimoireOfTheDeadZombies makes each returned creature a black Zombie in
// addition to its other colors and types for as long as it stays on the
// battlefield. An on-disk key: never renamed, never reused.
var grimoireOfTheDeadZombies = effects.TogetherTail("grimoire-of-the-dead/black-zombies",
	func(ctx *effects.Context, entered []uuid.UUID, _ int) error {
		for _, id := range entered {
			if err := (effects.ScopedEffectFor{
				Target: id,
				Mods: []game.Mod{
					game.AddSubtypesMod(ctx.TextSubtypes("Zombie")...), game.AddColorsMod(ctx.TextColors("B")...),
				},
				Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), id),
				Label:    "Grimoire of the Dead — a black Zombie in addition to its other colors and types",
			}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
