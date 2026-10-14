package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Embalmer's Tools — Artifact {2}:
//
//	"Activated abilities of creature cards in your graveyard cost {1}
//	 less to activate.
//	 Tap an untapped Zombie you control: Target player mills a card."
//
// The discount is the activation cost pass (ActivationCostsLess, #1184)
// narrowed to an ability whose source is a creature card in its
// controller's own graveyard: embalm, scavenge, unearth and the like.
// It spends generic mana only (CR 601.2f). The Tools tap no Zombie of
// their own: the second ability's cost is any untapped Zombie you
// control, summoning sickness or not (CR 302.6 is about {T}).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5bb4ee32-11c6-4ed4-a071-c3cf48070fff",
		Name:         "Embalmer's Tools",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			ActivationCostsLess(1, "Activated abilities of creature cards in your graveyard cost {1} less to activate.",
				ofACreatureCardInYourGraveyard()),
		},
		Activated: []ActivatedAbility{{
			Label:   "Tap an untapped Zombie you control: Target player mills a card.",
			Cost:    TapAnUntapped("an untapped Zombie you control", HasSubtype("Zombie")),
			Targets: TargetPlayer("target player"),
			Effect:  TargetPlayerMills(1),
		}},
	})
}

// ofACreatureCardInYourGraveyard passes on an activation whose source
// is a creature card in a graveyard, owned by the modifier's
// controller.
func ofACreatureCardInYourGraveyard() CostPredicate {
	return func(q game.CostQuery) bool {
		return q.FromZone == game.ZoneGraveyard && q.Card.IsCreature() && q.Card.Owner == q.Source.Controller
	}
}
