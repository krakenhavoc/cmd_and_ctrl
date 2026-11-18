package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vampire Socialite — Creature — Vampire Noble {B}{R}, 2/2 (Innistrad:
// Midnight Hunt, #5497):
//
//	"Menace
//	 When this creature enters, if an opponent lost life this turn, put a
//	 +1/+1 counter on each other Vampire you control.
//	 As long as an opponent lost life this turn, each other Vampire you
//	 control enters with an additional +1/+1 counter on it."
//
// Menace is printed. The entry trigger is Bloodtithe Collector's
// intervening "if" (CR 603.4) with Cordial Vampire's counters, minus the
// Socialite itself. The last line is Renata's CR 614 entry replacement for
// other Vampires, applying only while an opponent has lost life this turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "91a45e96-2a93-4c48-b472-11904d2b856d",
		Name:            "Vampire Socialite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AllOf(Self, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return opponentLifeLostThisTurn(g, source.Controller) > 0
			}), "Vampire Socialite — a +1/+1 counter on each other Vampire you control (an opponent lost life this turn)",
				Do(If{
					Cond: avB34CondAnOpponentLostLife{},
					Then: []Applier{Instruction(func(ctx *Context) error {
						var ids []uuid.UUID
						for _, id := range s82MidB06VampiresYouControl(ctx.Game, ctx.Controller(), ctx.TextSubtype("Vampire")) {
							if id != ctx.Source() {
								ids = append(ids, id)
							}
						}
						return b13PutCounterOnEach(ctx, ids)
					})},
				})),
		},
		Replacements: []game.ReplacementEffect{
			s82MidB06OtherVampiresYouControlEnterWithACounter("Vampire Socialite: each other Vampire you control enters with an additional +1/+1 counter"),
		},
	})
}
