package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Emergent Woodwurm — Creature — Wurm {6}{G}, 4/4 (March of the Machine
// Commander, #5477):
//
//	"Backup 3 (When this creature enters, put three +1/+1 counters on
//	 target creature. If that's another creature, it gains the following
//	 ability until end of turn.)
//	 Whenever this creature attacks, look at the top X cards of your
//	 library, where X is its power. You may put a permanent card with mana
//	 value X or less from among them onto the battlefield. Put the rest on
//	 the bottom of your library in a random order."
//
// Backup (CR 702.165) grants the attack trigger, a bundle in Spec.Grants
// that the Woodwurm also has itself. X is the power of the creature that
// has the ability as the ability resolves, its last known power if it has
// left (CR 608.2h), and the same X bounds both the look and the mana value.
// The put is Aang, at the Crossroads': a look, an optional put, the rest
// on the bottom in a random order.
//
// No simplification.
const emergentWoodwurmGrant = "emergent-woodwurm/attack-dig"

func init() {
	Register(Spec{
		OracleID:     "a74e41b4-6284-453c-a586-d575735dca65",
		Name:         "Emergent Woodwurm",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       emergentWoodwurmGrant,
			Triggered: []game.TriggeredAbility{emergentWoodwurmAttack()},
			Text:      "Whenever this creature attacks, look at the top X cards of your library, where X is its power. You may put a permanent card with mana value X or less from among them onto the battlefield. Put the rest on the bottom of your library in a random order.",
		}},
		Triggered: []game.TriggeredAbility{
			Backup("Emergent Woodwurm", 3, emergentWoodwurmGrant),
			emergentWoodwurmAttack(),
		},
	})
}

func emergentWoodwurmAttack() game.TriggeredAbility {
	return On(game.EventAttack, ThisAttacked,
		"Emergent Woodwurm — look at the top X cards of your library, where X is its power; you may put a permanent card with mana value X or less onto the battlefield",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			x := sourcePowerNow(ctx)
			if x <= 0 {
				return nil
			}
			return PutFromLibraryOntoBattlefield{
				Player:   item.Controller,
				Cards:    g.LookAtTopOfLibraryForEffect(item.Controller, x),
				Match:    ManaValueLE(x),
				Max:      1,
				Optional: true,
				Label:    "Emergent Woodwurm — you may put a permanent card with mana value X or less from among them onto the battlefield",
				ThenTail: PutRestOnBottomInRandomOrderTail,
			}.Apply(ctx)
		})
}
