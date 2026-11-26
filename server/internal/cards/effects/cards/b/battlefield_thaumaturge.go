package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Battlefield Thaumaturge — Creature — Human Wizard {1}{U}, 2/1 (Journey into
// Nyx, S82 OLD, #5558):
//
//	"Each instant and sorcery spell you cast costs {1} less to cast for each
//	 creature it targets.
//	 Heroic — Whenever you cast a spell that targets this creature, this creature
//	 gains hexproof until end of turn."
//
// Baral's instant-and-sorcery discount with the amount read off the spell's
// announced targets (CR 601.2c, before 601.2f totals the cost): one {1} for each
// DIFFERENT creature it targets, a creature targeted twice counting once. The
// reduction spends against generic mana only. The heroic trigger is Brigone's
// watch for a spell of yours that targets this creature, granting hexproof until
// end of turn.
//
// No simplification.
func init() {
	discount := effects.CostsLessEach(battlefieldThaumaturgeCreaturesTargeted,
		"Each instant and sorcery spell you cast costs {1} less to cast for each creature it targets.",
		effects.YourSpell(), effects.InstantOrSorcerySpell())
	discount.ReadsTargets = true
	effects.Register(effects.Spec{
		OracleID:      "70d2f5f7-f727-48e1-9532-8d365b1c1ae5",
		Name:          "Battlefield Thaumaturge",
		Completeness:  effects.CompletenessFull,
		CostModifiers: []game.CostModifier{discount},
		Triggered: []game.TriggeredAbility{
			effects.Heroic("Battlefield Thaumaturge — this creature gains hexproof until end of turn (heroic)",
				func(g *game.Game, item *game.StackItem) error {
					return effects.GrantKeywordUntilEOT{
						Target:   item.SourceCardID,
						Keywords: []string{"hexproof"},
						Label:    "Battlefield Thaumaturge — hexproof until end of turn",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// battlefieldThaumaturgeCreaturesTargeted counts the different creatures a spell
// being cast targets, as the cost is determined.
func battlefieldThaumaturgeCreaturesTargeted(q game.CostQuery) int {
	if q.Game == nil {
		return 0
	}
	seen := map[game.TargetRef]bool{}
	n := 0
	for _, t := range q.Targets {
		if t.Kind != game.TargetCard || seen[t] {
			continue
		}
		seen[t] = true
		if c, ok := q.Game.LookupCardForEffect(t.ID); ok && c.IsCreature() {
			n++
		}
	}
	return n
}
