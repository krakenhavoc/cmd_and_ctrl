package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Paladin Class — Enchantment — Class {W} (Adventures in the Forgotten
// Realms):
//
//	"(Gain the next level as a sorcery to add its ability.)
//	 Spells your opponents cast during your turn cost {1} more to cast.
//	 {2}{W}: Level 2
//	 Creatures you control get +1/+1.
//	 {4}{W}: Level 3
//	 Whenever you attack, until end of turn, target attacking creature gets
//	 +1/+1 for each other attacking creature and gains double strike."
//
// One line per level (ADR 0071):
//
//   - Level 1 is a cost increase on every spell an opponent casts while it
//     is the Class's controller's turn, ungated (CR 716.3).
//   - Level 2 is the layer 7c anthem of a lord, gated at level 2.
//   - Level 3 is Battlesong Berserker's "whenever you attack" (one trigger
//     per declaration, OncePerBatch), gated at level 3, with the target
//     any attacking creature, whoever controls it. The count is taken as
//     the trigger resolves (CR 608.2h), over every attacking creature on
//     the battlefield other than the target; the boost and double strike
//     are pinned to the target, and nothing happens if it is no longer an
//     attacking creature (CR 608.2b).
//
// No simplification.
func init() {
	anthem := effects.TribalAnthem(effects.TribeFilter{YoursOnly: true}, 1, 1)
	anthem.ActiveWhen = effects.Level(2)
	effects.Register(effects.Spec{
		OracleID:     "ac9d52b2-78a5-4b17-8aaf-1c1cb936a508",
		Name:         "Paladin Class",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsMore(1, "Spells your opponents cast during your turn cost {1} more to cast.",
				effects.OpponentsSpell(), func(q game.CostQuery) bool {
					return q.Game != nil && effects.IsActivePlayer(q.Game, q.Source.Controller)
				}),
		},
		Static: []game.StaticAbility{anthem},
		Activated: []effects.ActivatedAbility{
			effects.LevelUp(2, effects.ManaCost("{2}{W}")),
			effects.LevelUp(3, effects.ManaCost("{4}{W}")),
		},
		Triggered: []game.TriggeredAbility{
			effects.AtLevel(3, effects.TriggerWithPurpose(effects.OncePerBatch(effects.Targeting(
				effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.AttackDeclaredByYou(ev, source.Controller)
				}, "Paladin Class — target attacking creature gets +1/+1 for each other attacking creature and gains double strike until end of turn",
					paladinClassPump),
				effects.TargetCreature("target attacking creature", effects.AttackingCreature()))),
				effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "double strike")))),
		},
	})
}

// paladinClassPump gives the still-legal target +1/+1 for each other
// attacking creature, and double strike, until end of turn.
func paladinClassPump(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	target, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if c, found := g.LookupCardForEffect(target); !found || c.AttackingTarget == uuid.Nil || !c.IsCreature() {
		return nil
	}
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID != target && c.AttackingTarget != uuid.Nil && c.IsCreature() {
			n++
		}
	}
	if n > 0 {
		if err := (effects.BoostUntilEOT{Target: target, Power: n, Toughness: n, Label: "Paladin Class — +1/+1 for each other attacking creature"}).Apply(ctx); err != nil {
			return err
		}
	}
	return effects.GrantKeywordUntilEOT{Target: target, Keywords: []string{"double strike"}, Label: "Paladin Class — double strike"}.Apply(ctx)
}
