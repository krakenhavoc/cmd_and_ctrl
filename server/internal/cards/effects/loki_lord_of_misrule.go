package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Loki, Lord of Misrule — Legendary Creature — God Sorcerer Villain {3}{U},
// 3/4 (Marvel Super Heroes Commander, #4455):
//
//	"{U}, {T}: Choose target creature you control. Each creature you
//	 control other than the chosen creature becomes a copy of that
//	 creature until end of turn, except it isn't legendary. Activate only
//	 as a sorcery."
//
// Shuri's duration BecomeCopy (CR 707.2, 611.2) over a set instead of one
// permanent. The chosen creature is the ability's one target (checked as it
// resolves, CR 608.2b); the creatures that change are read as it resolves:
// every creature its controller controls except the chosen one, Loki
// himself included. The "except" clause removes the Legendary supertype
// from the copiable values.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "20ec130f-05b4-4fbf-a54f-5874ae7ef90f",
		Name:         "Loki, Lord of Misrule",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{U}, {T}: Choose target creature you control. Each creature you control other than the chosen creature becomes a copy of that creature until end of turn, except it isn't legendary. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{U}"), TapCost()),
			SorcerySpeed: true,
			Targets:      TargetCreature("target creature you control (the chosen creature)", YouControl()),
			Effect:       lokiLordOfMisruleEffect,
		}},
	})
}

// lokiLordOfMisruleEffect is the resolution body.
//
// Caller holds g.mu.
func lokiLordOfMisruleEffect(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ids := legalTargetIDs(ctx)
	if len(ids) == 0 {
		return nil
	}
	chosen := ids[0]
	var others []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsCreature() && c.InstanceID != chosen {
			others = append(others, c.InstanceID)
		}
	}
	return BecomeCopy{
		Targets: others,
		Of:      chosen,
		Except:  func(v *game.PrintedValues) { v.RemoveSupertype("Legendary") },
		Label:   "Loki, Lord of Misrule — each other creature becomes a copy of the chosen creature until end of turn",
	}.Apply(ctx)
}
