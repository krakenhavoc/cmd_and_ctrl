package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Karvanista, Loyal Lupari // Lupari Shield — an adventurer card (Doctor Who):
//
//	Karvanista, Loyal Lupari {4}{G}, Legendary Creature — Alien Dog Soldier 5/5:
//	"Vigilance, trample, haste
//	 Whenever Karvanista attacks, put a +1/+1 counter on each Human you
//	 control."
//	Lupari Shield {1}{G}, Sorcery — Adventure:
//	"Humans you control gain indestructible until your next turn."
//
// The attack trigger counts the Humans as it resolves (Karvanista is not one).
// The grant is a scoped effect whose set is pinned when the spell resolves
// (CR 611.2c): the Humans on the battlefield then keep indestructible until the
// start of the caster's next turn; one that arrives later does not get it. The
// engine owns the adventure: the spell resolves into exile and the creature may
// be cast from there.
//
// No simplification.
const karvanistaOracleID = "6c370299-9803-4cb5-97ba-c17831797b54"

func init() {
	effects.Register(effects.Spec{
		OracleID:        karvanistaOracleID,
		Name:            "Karvanista, Loyal Lupari",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "trample", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Karvanista — put a +1/+1 counter on each Human you control", effects.GapB85PlusOneOnEachHuman),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     karvanistaOracleID + "#1",
		Name:         "Lupari Shield",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.ScopedEffectFor{
					Match:    effects.And(effects.Creature(), effects.YouControl(), effects.OfSubtype(ctx.TextSubtype("Human"))),
					Mods:     []game.Mod{game.AddKeywordsMod("indestructible")},
					Duration: effects.DurationUntilYourNextTurn(ctx, ctx.Controller()),
					Label:    "Lupari Shield — Humans you control gain indestructible until your next turn",
				}.Apply(ctx)
			}),
		),
	})
}
