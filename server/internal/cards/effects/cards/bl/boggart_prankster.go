package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boggart Prankster — Creature — Goblin Warrior {1}{B}, 1/3 (Lorwyn
// Eclipsed, #4458):
//
//	"Whenever you attack, target attacking Goblin you control gets +1/+0
//	 until end of turn."
//
// Bast, Panther Goddess's trigger at a fixed +1/+0: once per attack
// declaration of yours (OncePerBatch over attackDeclaredByYou), whoever
// attacks, aimed at an attacking Goblin you control as it goes on the
// stack. The Prankster need not attack itself. With no attacking Goblin
// to target the trigger is removed (CR 603.3d); a target that has left
// combat or the battlefield by resolution gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1bcac664-835a-4474-8ca4-263801bde08e",
		Name:         "Boggart Prankster",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.Targeting(
				effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.AttackDeclaredByYou(ev, source.Controller)
				}, "Boggart Prankster — target attacking Goblin you control gets +1/+0 until end of turn",
					boggartPranksterPump),
				effects.TargetCreature("target attacking Goblin you control", effects.AttackingCreature(), effects.YouControl(), effects.OfCreatureType("Goblin")))),
		},
	})
}

// boggartPranksterPump gives the still-legal attacking Goblin +1/+0 until
// end of turn.
func boggartPranksterPump(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return effects.BoostUntilEOT{Target: id, Power: 1, Label: "Boggart Prankster — +1/+0"}.Apply(ctx)
}
