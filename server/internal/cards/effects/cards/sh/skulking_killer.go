package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skulking Killer — Creature — Vampire Assassin {3}{B}, 4/2 (Innistrad:
// Crimson Vow, #5495):
//
//	"When this creature enters, target creature an opponent controls gets
//	 -2/-2 until end of turn if that opponent controls no other
//	 creatures."
//
// The target is chosen whatever the opponent controls; the "if" is a
// resolution check, not an intervening one (it is not in the trigger
// condition), so it is read as the ability resolves: the target must still
// be legal (CR 608.2b) and its controller must control no creature other
// than it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e711a2d6-008d-4e7b-986e-f34805165679",
		Name:         "Skulking Killer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Skulking Killer — target creature an opponent controls gets -2/-2 until end of turn if that opponent controls no other creatures",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					victim, found := g.LookupCardForEffect(id)
					if !found {
						return nil
					}
					g.RecomputeLayersIfStaleLocked()
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == victim.Controller && c.InstanceID != id && c.IsCreature() {
							return nil
						}
					}
					return effects.BoostUntilEOT{Target: id, Power: -2, Toughness: -2, Label: "Skulking Killer — -2/-2 until end of turn"}.Apply(ctx)
				}),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
