package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Éowyn, Lady of Rohan — Legendary Creature — Human Noble {2}{W}, 2/4
// (The Lord of the Rings: Tales of Middle-earth, #5948):
//
//	"At the beginning of combat on your turn, target creature gains your
//	 choice of first strike or vigilance until end of turn. If that
//	 creature is equipped, it gains first strike and vigilance until end of
//	 turn instead.
//	 Equip abilities you activate cost {1} less to activate."
//
// The trigger targets as it goes on the stack. As it resolves the target is
// checked again (CR 608.2b); whether it is equipped is read then, from the
// Equipment attached to it, and an equipped creature gets both keywords with
// no choice. Otherwise the controller picks one. The discount is Bureau
// Headmaster's equip-ability cost modifier.
//
// No simplification.
func init() {
	const label = "Éowyn, Lady of Rohan — target creature gains your choice of first strike or vigilance until end of turn (both if it is equipped)"
	Register(Spec{
		OracleID:     "1ee3753c-3b6e-4182-9305-2ab757f485f0",
		Name:         "Éowyn, Lady of Rohan",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventStepBegan, StepBegan(game.StepBeginCombat, true), label,
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					ts := ctx.LegalTargets()
					if len(ts) == 0 {
						return nil
					}
					if eowynIsEquipped(g, ts[0].ID) {
						return GrantKeywordUntilEOT{Target: ts[0].ID, Keywords: []string{"first strike", "vigilance"}, Label: label}.Apply(ctx)
					}
					return PickOption{
						Question: "Éowyn, Lady of Rohan — target creature gains your choice of first strike or vigilance",
						Options:  []game.ChoiceOption{{Label: "First strike"}, {Label: "Vigilance"}},
						ThenKey:  eowynKeywordChosen,
						WithItem: true,
					}.Apply(ctx)
				}), TargetCreature("target creature")),
		},
		CostModifiers: []game.CostModifier{
			EquipAbilitiesYouActivateCostLess(1, "Equip abilities you activate cost {1} less to activate."),
		},
	})
}

// eowynKeywordChosen is the first strike or vigilance pick, keyed: the
// trigger rides the prompt as data. An on-disk key: never renamed, never
// reused.
var eowynKeywordChosen = OptionPickThen("option-pick/eowyn-lady-of-rohan-keyword", func(ctx *Context, r game.OptionPicked) error {
	keywords := []string{"first strike", "vigilance"}
	if r.Index < 0 || r.Index >= len(keywords) {
		return nil
	}
	ts := ctx.LegalTargets()
	if len(ts) == 0 {
		return nil
	}
	return GrantKeywordUntilEOT{Target: ts[0].ID, Keywords: keywords[r.Index : r.Index+1],
		Label: "Éowyn, Lady of Rohan — gains " + keywords[r.Index] + " until end of turn"}.Apply(ctx)
})

// eowynIsEquipped reports whether some Equipment on the battlefield is
// attached to the creature.
func eowynIsEquipped(g *game.Game, creature uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.HasSubtype("Equipment") && c.IsAttachedTo(creature) {
			return true
		}
	}
	return false
}
