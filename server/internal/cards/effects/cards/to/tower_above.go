package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// towerAboveAttack is the bundle the spell hands its target until end of turn.
// An on-disk key: never renamed, never reused.
const towerAboveAttack = "tower-above/attack-block"

// Tower Above — Sorcery {2/G}{2/G}{2/G} (Shadowmoor, S82 OLD b530, #5591):
//
//	"({2/G} can be paid with any two mana or with {G}. This card's mana value
//	 is 6.)
//	 Until end of turn, target creature gets +4/+4 and gains trample, wither,
//	 and "When this creature attacks, target creature blocks it this turn if
//	 able.""
//
// Feign Death's duration grant: the attack trigger is a catalog bundle the
// spell gives the target until end of turn, so it is the creature's own ability
// (CR 613.6) and has its own target, chosen when it triggers. The block
// requirement names the attacker as the object it is then (Feral Contest's
// record). With its one target illegal on resolution the spell does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6f5d31d1-3493-4d1b-bbf9-1726e47507db",
		Name:         "Tower Above",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 4, 4, "trample", game.KeywordWither)),
		Targets:      effects.TargetCreature("target creature"),
		Grants: []effects.AbilityGrant{{
			Key: towerAboveAttack,
			Triggered: []game.TriggeredAbility{
				effects.Targeting(effects.On(game.EventAttack, effects.ThisAttacked, "Tower Above — target creature blocks it this turn if able",
					effects.OldB530TargetBlocksThisAttacker("Tower Above — blocks the attacker if able")),
					effects.TargetCreature("target creature")),
			},
			Text: "When this creature attacks, target creature blocks it this turn if able.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			target, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: target, Power: 4, Toughness: 4, Label: "Tower Above — +4/+4"}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.GrantKeywordUntilEOT{Target: target, Keywords: []string{"trample", game.KeywordWither},
				Label: "Tower Above — trample and wither"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantAbilitiesFor{Target: target, Keys: []string{towerAboveAttack},
				Label: "Tower Above — until end of turn, it makes a creature block it when it attacks"}.Apply(ctx)
		},
	})
}
