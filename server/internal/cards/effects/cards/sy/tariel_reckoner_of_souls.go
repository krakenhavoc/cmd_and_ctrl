package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tariel, Reckoner of Souls — Legendary Creature — Angel {4}{R}{W}{B},
// 4/7 (Alara Reborn):
//
//	"Flying, vigilance
//	 {T}: Choose a creature card at random from target opponent's
//	 graveyard. Put that card onto the battlefield under your control."
//
// Flying and vigilance are printed keywords. The target is the opponent;
// as the ability resolves one creature card is chosen at random from their
// graveyard (the engine's keyed random stream, never a private RNG) and
// enters under the controller's control. An opponent with no creature
// card in their graveyard gives nothing; an illegal target leaves the
// ability to do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9c00b503-ae5e-4ce3-ad6a-92b541b2e3cd",
		Name:            "Tariel, Reckoner of Souls",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Choose a creature card at random from target opponent's graveyard. Put that card onto the battlefield under your control.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect:  tarielReanimateAtRandom,
		}},
	})
}

func tarielReanimateAtRandom(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var opp uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			opp = t.ID
		}
	}
	if opp == uuid.Nil {
		return nil
	}
	ids := effects.GraveyardIDs(g, opp, func(c game.Card) bool { return c.IsCreature() })
	pick := effects.RandomPick(ctx, ids, 1)
	if len(pick) == 0 {
		return nil
	}
	return effects.ReturnFromGraveyard{Target: pick[0], Dest: game.ZoneBattlefield, Controller: ctx.Controller()}.Apply(ctx)
}
