package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// SP//dr, Piloted by Peni — Legendary Artifact Creature — Spider Hero
// {3}{W}{U}, 4/4 (Marvel's Spider-Man, #4460):
//
//	"Vigilance
//	 When SP//dr enters, put a +1/+1 counter on target creature.
//	 Whenever a modified creature you control deals combat damage to a
//	 player, draw a card. (Equipment, Auras you control, and counters are
//	 modifications.)"
//
// Two printed triggers, two registered. The enters trigger targets any
// creature (SP//dr itself included). The draw reads the dealer as
// modified (CR 700.9, b07IsModified) when the damage is dealt, and
// triggers once per creature that connects, as printed ("a modified
// creature", not "one or more").
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "722ab6e4-a2e0-4aaa-a050-7b6436e22b40",
		Name:            "SP//dr, Piloted by Peni",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("SP//dr — put a +1/+1 counter on target creature",
				onePlusOneCounterOnFirstLegalTarget), TargetCreature("target creature")),
			On(game.EventDealDamage, modifiedCreatureDealtCombatDamageToPlayer, "SP//dr — a modified creature dealt combat damage to a player: draw a card",
				Do(DrawCards{N: 1})),
		},
	})
}

func onePlusOneCounterOnFirstLegalTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return nothingIfGone(AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx))
		}
	}
	return nil
}
