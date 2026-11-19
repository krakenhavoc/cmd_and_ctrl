package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Travel Preparations — Sorcery {1}{G} (Innistrad Remastered):
//
//	"Put a +1/+1 counter on each of up to two target creatures.
//	 Flashback {1}{W}"
//
// Support's body for a spell (SupportEffect): a counter on each target still
// legal as the spell resolves (CR 608.2b).
func init() {
	Register(Spec{
		OracleID:         "8432729e-5fd4-4661-8b79-e41b1d03eb7b",
		Name:             "Travel Preparations",
		Purpose:          game.Purpose{Answers: game.AnswerPump},
		Completeness:     CompletenessFull,
		Targets:          SupportTargets(2, false),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{1}{W}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return SupportEffect(ctx.Game, item)
		},
	})
}
