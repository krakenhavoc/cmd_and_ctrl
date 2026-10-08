package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vona's Hunger — Instant {2}{B} (EDHREC rank 3917):
//
//	"Ascend (If you control ten or more permanents, you get the city's
//	 blessing for the rest of the game.)
//	 Each opponent sacrifices a creature of their choice. If you have the
//	 city's blessing, instead each opponent sacrifices half the creatures
//	 they control of their choice, rounded up."
//
// An edict that becomes a sweeper once the caster's board is wide. The
// count is worked out PER OPPONENT as the prompt is built: "half the
// creatures they control, rounded up" is ceil(n/2) of that opponent's
// creatures, so three creatures is two and a single creature is one.
// The choice is the opponent's own and is not a target, so hexproof
// creatures are fair game; an opponent with no creature is skipped.
//
// Ascend on an instant is a spell ability (CR 702.131a): the engine
// checks as the spell resolves, before this body runs, so a caster who
// reaches ten permanents with the spell's own resolution has already
// earned the "instead" clause.
//
// Each opponent is asked in turn order from the active player; the
// sacrifices happen as each opponent answers rather than all at once
// (the same posture EachPlayerSacrifices takes, game/sacrifice.go).
//
// No simplification beyond that ordering.
func init() {
	Register(Spec{
		OracleID:        "da522912-1bc2-4d70-b426-94d10edb2afe",
		Name:            "Vona's Hunger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordAscend},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			g := ctx.Game
			half := YouHaveTheCitysBlessing(g, item.Controller)
			n := len(g.Seats)
			for i := 0; i < n; i++ {
				p := g.Seats[(g.Turn.ActiveSeat+i)%n]
				if p == nil || p.Eliminated || p.ID == item.Controller {
					continue
				}
				if err := (ChoosePermanents{
					Player:     p.ID,
					Question:   vonasHungerQuestion(half),
					Candidates: vonasHungerCandidates(half),
					Sacrifice:  true,
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}

func vonasHungerQuestion(half bool) string {
	if half {
		return "Vona's Hunger — sacrifice half your creatures, rounded up"
	}
	return "Vona's Hunger — sacrifice a creature"
}

// vonasHungerCandidates offers every creature the chooser controls and
// demands exactly one, or half of them rounded up.
func vonasHungerCandidates(half bool) func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
	return func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
		var out []uuid.UUID
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == of && c.IsCreature() {
				out = append(out, c.InstanceID)
			}
		}
		if len(out) == 0 {
			return nil, 0, 0
		}
		n := 1
		if half {
			n = (len(out) + 1) / 2
		}
		return out, n, n
	}
}
