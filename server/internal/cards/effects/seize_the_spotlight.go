package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Seize the Spotlight — Sorcery {2}{R} (Edea steal-and-sac deck,
// #1565):
//
//	"Each opponent chooses fame or fortune. For each player who chose
//	 fame, gain control of a creature that player controls until end
//	 of turn. Untap those creatures and they gain haste until end of
//	 turn. For each player who chose fortune, you draw a card and
//	 create a Treasure token."
//
// Torment of Hailfire's per-opponent option pick (#568), asked one
// opponent at a time in APNAP order so each hears the choices before
// theirs (CR 101.4), then the two halves in printed order:
//
//   - **Fame.** For each player who chose fame, the CASTER chooses one
//     creature that player controls (a resolution-time pick over that
//     board, ChoosePermanents: it does not target, so hexproof does
//     not stop it). Each is then Act of Treason'd: gain control until
//     end of turn, untap, haste until end of turn. A player who chose
//     fame with no creature gives nothing.
//   - **Fortune.** One card and one Treasure per player who chose it.
//
// An opponent who leaves the game before answering is skipped, and the
// run goes on.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3d1b09a3-f151-4218-974b-02bb6159247a",
		Name:         "Seize the Spotlight",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return seizeTheSpotlightAsk(ctx, ctx.Opponents(), nil, 0)
		},
	})
}

// seizeTheSpotlightAsk asks the next opponent, accumulating the fame
// seats and the fortune count, and runs the effect once everyone has
// answered.
func seizeTheSpotlightAsk(ctx *Context, remaining, fame []uuid.UUID, fortune int) error {
	for len(remaining) > 0 {
		opp := remaining[0]
		rest := remaining[1:]
		if p := ctx.Game.PlayerByIDForEffect(opp); p == nil || p.Eliminated {
			remaining = rest
			continue
		}
		return PickOption{
			Player:   opp,
			Question: "Seize the Spotlight — choose fame or fortune",
			Options: []game.ChoiceOption{
				{Label: "Fame — its caster gains control of a creature you control until end of turn"},
				{Label: "Fortune — its caster draws a card and creates a Treasure token"},
			},
			Then: func(ctx *Context, index int) error {
				switch index {
				case 0:
					return seizeTheSpotlightAsk(ctx, rest, append(append([]uuid.UUID(nil), fame...), opp), fortune)
				case 1:
					return seizeTheSpotlightAsk(ctx, rest, fame, fortune+1)
				}
				return seizeTheSpotlightAsk(ctx, rest, fame, fortune)
			},
		}.Apply(ctx)
	}
	return seizeTheSpotlightResolve(ctx, fame, fortune)
}

// seizeTheSpotlightResolve runs the fame half, then the fortune half.
func seizeTheSpotlightResolve(ctx *Context, fame []uuid.UUID, fortune int) error {
	fortuneHalf := func(ctx *Context) error {
		for i := 0; i < fortune; i++ {
			if err := (DrawCards{Player: ctx.Controller(), N: 1}).Apply(ctx); err != nil {
				return err
			}
			if err := (CreateToken{Template: TreasureToken(), N: 1}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}
	if len(fame) == 0 {
		return fortuneHalf(ctx)
	}
	return ChoosePermanents{
		Question: "Seize the Spotlight — gain control of a creature that player controls",
		Of:       fame,
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			var out []uuid.UUID
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == of && c.IsCreature() {
					out = append(out, c.InstanceID)
				}
			}
			if len(out) == 0 {
				return nil, 0, 0
			}
			return out, 1, 1
		},
		Then: func(ctx *Context, picked game.PromptedPicks) error {
			for _, id := range picked.Cards() {
				if err := (GainControl{
					Target:   id,
					Duration: DurationUntilEndOfTurn(ctx),
					Label:    "Seize the Spotlight — gain control until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
				if err := (UntapTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
				if err := (GrantKeywordUntilEOT{
					Target:   id,
					Keywords: []string{"haste"},
					Label:    "Seize the Spotlight — haste until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return fortuneHalf(ctx)
		},
	}.Apply(ctx)
}
