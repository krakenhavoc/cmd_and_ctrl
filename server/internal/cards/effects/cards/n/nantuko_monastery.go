package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nantuko Monastery — Land (Dominaria Remastered, #5948):
//
//	"{T}: Add {C}.
//	 Threshold — {G}{W}: This land becomes a 4/4 green and white Insect Monk
//	 creature with first strike until end of turn. It's still a land.
//	 Activate only if there are seven or more cards in your graveyard."
//
// Celestial Colonnade's animation: one continuous effect from the resolving
// ability pinned to the land (CR 611.2c); the activation condition is
// checked when the ability is activated (CR 602.5b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "c6de0ee9-785d-4cd8-8a7f-5bb715763131",
		Name:          "Nantuko Monastery",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{effects.PainlessColorless()},
		Activated: []effects.ActivatedAbility{{
			Label:     "Threshold — {G}{W}: This land becomes a 4/4 green and white Insect Monk creature with first strike until end of turn. It's still a land. Activate only if there are seven or more cards in your graveyard.",
			Purpose:   game.Purpose{Answers: game.AnswerAnimate},
			Cost:      effects.ManaCost("{G}{W}"),
			Condition: effects.GraveyardAtLeast(7, nil),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				return effects.ScopedEffectFor{
					Target: item.SourceCardID,
					Mods: append([]game.Mod{
						game.AddTypesMod("Creature"),
						game.AddSubtypesMod(ctx.TextSubtypes("Insect", "Monk")...),
						game.SetColorsMod(ctx.TextColors("G", "W")...),
						game.AddKeywordsMod("first strike"),
					}, game.SetBasePTMods(4, 4)...),
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    "Nantuko Monastery — a 4/4 green and white Insect Monk creature with first strike until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
