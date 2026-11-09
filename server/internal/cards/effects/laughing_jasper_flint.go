package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Laughing Jasper Flint — Legendary Creature — Lizard Rogue {1}{B}{R}, 3/3
// (Outlaws of Thunder Junction, #4468):
//
//	"Creatures you control but don't own are Mercenaries in addition to
//	 their other types.
//	 At the beginning of your upkeep, exile the top X cards of target
//	 opponent's library, where X is the number of outlaws you control. Until
//	 end of turn, you may cast spells from among those cards, and mana of
//	 any type can be spent to cast those spells."
//
// The static is a layer-4 type addition on the creatures its controller
// controls and does not own (control and ownership read through the
// layers, so a stolen creature qualifies and one's own does not). X is
// counted as the trigger resolves, over the effective subtypes of the
// controller's permanents (Assassin, Mercenary, Pirate, Rogue, Warlock), so
// the Lizard Rogue herself counts and a creature this makes a Mercenary
// counts too. Only spells may be cast from the exiled cards: no land is
// played (the printed "cast"), and any type of mana pays.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bd7c5a8b-cf62-41e1-a664-b16e3e5d142b",
		Name:         "Laughing Jasper Flint",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label: "Creatures you control but don't own are Mercenaries in addition to their other types.",
			Layer: game.Layer4Type,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller && target.Owner != source.Controller
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
				merc := source.TextSubtype("Mercenary")
				for _, st := range c.Subtypes {
					if st == merc {
						return
					}
				}
				c.Subtypes = append(c.Subtypes, merc)
			},
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(AtYourUpkeep("Laughing Jasper Flint — exile the top X cards of target opponent's library; you may cast spells from among them this turn, and mana of any type can be spent",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					opp := TargetedPlayer(ctx)
					if opp == uuid.Nil {
						return nil
					}
					x := len(MatchingBattlefield(ctx, And(YouControl(), b40Outlaw())))
					if x <= 0 {
						return nil
					}
					return ExileTopWithPermission{
						From:     opp,
						GrantTo:  item.Controller,
						N:        x,
						CastOnly: true,
						AnyColor: true,
						AnyType:  true,
					}.Apply(ctx)
				}),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}
