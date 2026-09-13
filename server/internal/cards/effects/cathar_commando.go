package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cathar Commando — Creature — Human Soldier {1}{W}, 3/1 (EDHREC
// rank 1416):
//
//	"Flash
//	 {1}, Sacrifice this creature: Destroy target artifact or
//	 enchantment."
//
// A Disenchant that is also a flash 3/1 — the white deck's answer
// that can block first. Flash is the printed keyword; the removal is
// a CR 602 activated ability with a mana-plus-sacrifice-self cost
// (Haywire Mite's shape) and a real target clause, so it uses the
// stack and the sacrifice pays at announce — dies-triggers on the
// Commando land above the ability and resolve first.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "774dce79-67e0-4820-8013-c7a7347993ce",
		Name:            "Cathar Commando",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice this creature: Destroy target artifact or enchantment.",
			Cost:    Plus(ManaCost("{1}"), SacrificeThis()),
			Targets: TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				return DestroyTarget{Target: item.Targets[0].ID}.Apply(NewContext(g, item))
			},
		}},
	})
}
