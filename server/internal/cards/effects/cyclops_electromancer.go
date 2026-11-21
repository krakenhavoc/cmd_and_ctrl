package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cyclops Electromancer — Creature — Cyclops Wizard {4}{R}, 4/2
// (Foundations Jumpstart):
//
//	"When this creature enters, it deals X damage to target creature an
//	 opponent controls, where X is the number of instant and sorcery cards
//	 in your graveyard."
//
// X is counted as the trigger resolves, in the controller's graveyard
// (Cosmic Epiphany's count, a split card once). The Cyclops is the damage
// source. A target that is no longer legal takes nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "638ecc20-a275-4810-9669-89a0730c5362",
		Name:         "Cyclops Electromancer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Cyclops Electromancer — deals X damage to target creature an opponent controls, X the number of instant and sorcery cards in your graveyard",
					CyclopsElectromancerBlast),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}

func CyclopsElectromancerBlast(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return DealDamage{
		Source: item.SourceCardID,
		Target: target,
		Amount: CmdrB17InstantSorceryCardsInGraveyard(g, item.Controller),
	}.Apply(ctx)
}
