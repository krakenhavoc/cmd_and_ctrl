package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Marshal of the Lost — Creature — Orc Warrior {2}{W}{B}, 3/3 (Tarkir:
// Dragonstorm, #4463):
//
//	"Deathtouch
//	 Whenever you attack, target creature gets +X/+X until end of turn,
//	 where X is the number of attacking creatures."
//
// "Whenever you attack" is Sparring Regimen's trigger: once per
// declaration of attackers, whatever the number of attackers (CR 603.2c),
// and whether or not the Marshal is one of them. The target is any
// creature, picked as the trigger goes on the stack. X is counted once,
// as the trigger resolves (CR 608.2h, and the printed ruling): every
// creature attacking then, whoever controls it, so a Warrior a mobilize
// trigger made before this one resolved counts and an attacker removed
// from combat in response does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "af762373-e159-4a36-83b5-a82b8b16fa41",
		Name:            "Marshal of the Lost",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.OncePerBatch(effects.Targeting(
				effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.AttackDeclaredByYou(ev, source.Controller)
				}, "Marshal of the Lost — target creature gets +X/+X, where X is the number of attacking creatures",
					marshalOfTheLostPump),
				effects.TargetCreature("target creature"),
			)), effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0))),
		},
	})
}

// marshalOfTheLostPump is the trigger's resolution: X read now.
func marshalOfTheLostPump(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	x := len(effects.AttackingCreatures(g))
	return effects.BoostUntilEOT{
		Target:    id,
		Power:     x,
		Toughness: x,
		Label:     "Marshal of the Lost — +X/+X until end of turn",
	}.Apply(ctx)
}
