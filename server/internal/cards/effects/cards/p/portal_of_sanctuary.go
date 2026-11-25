package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Portal of Sanctuary — Artifact {2}{U} (Core Set 2020):
//
//	"{1}, {T}: Return target creature you control and each Aura attached to it
//	 to their owners' hands. Activate only during your turn."
//
// The creature and every Aura attached to it, whoever controls the Aura, go
// to their owners' hands as one simultaneous move (CR 400.7: the Auras are
// read at resolution, before the creature leaves and they would fall off). A
// target that became illegal does nothing, so no Aura is returned alone. The
// window is DuringYourTurn (CR 602.1b), not a sorcery-speed gate.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c30491b5-08fe-4694-9ac8-53be7b4b92ce",
		Name:         "Portal of Sanctuary",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     "{1}, {T}: Return target creature you control and each Aura attached to it to their owners' hands. Activate only during your turn.",
			Purpose:   game.Purpose{Answers: game.AnswerSurvive},
			Cost:      effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets:   effects.TargetCreature("target creature you control", effects.YouControl()),
			Condition: effects.DuringYourTurn(),
			Effect:    portalOfSanctuaryBounce,
		}},
	})
}

func portalOfSanctuaryBounce(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok || !effects.OnBattlefield(g, id) {
		return nil
	}
	ids := []uuid.UUID{id}
	for i := range g.Battlefield.Cards {
		a := &g.Battlefield.Cards[i]
		if a.IsAura() && a.IsAttachedTo(id) {
			ids = append(ids, a.InstanceID)
		}
	}
	g.BounceCardsToHandForEffect(ids)
	return nil
}
