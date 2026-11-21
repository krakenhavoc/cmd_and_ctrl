package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Killing Wave — Sorcery {X}{B} (Innistrad Remastered):
//
//	"For each creature, its controller sacrifices it unless they pay X life."
//
// The creatures are the ones on the battlefield as the spell resolves
// (CR 608.2h). Each player is asked once, in APNAP order, which of THEIR
// creatures they will pay X life for (OwnObjectsOnly: unlike Cleansing, an
// opponent can't pay for your creatures), up to as many as their life covers
// (CR 119.4). The creatures nobody paid for are sacrificed together. With X
// = 0 every payment is free, so every creature stays.
func init() {
	Register(Spec{
		OracleID:     "69d6b906-5461-4eba-8667-dbb8c0ce3fcb",
		Name:         "Killing Wave",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepSacrifice}},
		Completeness: CompletenessFull,
		XMatters:     true,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			x := ctx.X()
			if x <= 0 {
				return nil
			}
			return AnyPlayerPaysForEach{
				Objects: func(ctx *Context) []uuid.UUID {
					var creatures []uuid.UUID
					for _, c := range MatchingBattlefield(ctx, Creature()) {
						creatures = append(creatures, c.InstanceID)
					}
					return creatures
				},
				EachLife:       x,
				Noun:           "creature",
				Question:       "Killing Wave — pay X life for each of your creatures you want to keep?",
				Then:           killingWaveSacrificesTheUnsaved,
				OwnObjectsOnly: true,
			}.Apply(ctx)
		},
	})
}

var killingWaveSacrificesTheUnsaved = AnyPlayerEachConsequence("any-player-each/killing-wave-sacrifice",
	func(ctx *Context, r game.AnyPlayerEachAnswer) error {
		return ctx.Game.SacrificeAllKeyedForEffect(ctx.Source(), r.Unsaved, game.KeyedTail{})
	})
