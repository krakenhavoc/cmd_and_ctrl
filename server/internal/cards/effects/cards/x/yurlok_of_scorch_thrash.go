package cardsx

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Yurlok of Scorch Thrash — Legendary Creature — Lizard Shaman {1}{B}{R}{G},
// 4/4 (Commander Legends):
//
//	"Vigilance
//	 A player losing unspent mana causes that player to lose that much life.
//	 {1}, {T}: Each player adds {B}{R}{G}."
//
// Vigilance is the engine's. The second line is a mana-pool static over every
// player (game.ManaPoolLosesLife): when mana is lost as steps and phases end
// (CR 106.4), or by an instruction to lose it, its owner then loses that much
// life through the ordinary life-change window, so a replacement or "can't
// lose life" applies. The third is a mana ability (CR 605.1a: it targets
// nothing and adds mana): the activator's {B}{R}{G} is the ability's own, and
// each other player's is added by its rider in the same atomic resolution
// (CR 605.3b), in turn order from the activator.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7baaf462-b0e6-4e21-86a7-73a883c4c9f7",
		Name:            "Yurlok of Scorch Thrash",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		ManaPool:        []game.ManaPoolStatic{{Players: game.ManaPoolEachPlayer, Kind: game.ManaPoolLosesLife}},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true, Mana: "{1}"},
			Produced: "{B}{R}{G}",
			Label:    "{1}, {T}: Each player adds {B}{R}{G}.",
			Rider: func(g *game.Game, controller, source uuid.UUID) error {
				for _, p := range g.Seats {
					if p.ID == controller || p.Eliminated {
						continue
					}
					if err := g.AddManaForEffect(p.ID, source, "{B}{R}{G}"); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
