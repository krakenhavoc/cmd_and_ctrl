package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kitsa, Otterball Elite — Legendary Creature — Otter Wizard {1}{U}, 1/3
// (Bloomburrow, #4467):
//
//	"Vigilance
//	 Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)
//	 {T}: Draw a card, then discard a card.
//	 {2}, {T}: Copy target instant or sorcery spell you control. You may
//	 choose new targets for the copy. Activate only if Kitsa's power is 3
//	 or greater."
//
// Vigilance and prowess are printed keywords (game/prowess.go derives
// the trigger). The loot is Merfolk Looter's. The copy is Lithoform
// Engine's spell copy (copyTargetedSpell, with CR 707.10c's new-targets
// offer) behind a CR 602.1b activation condition: Kitsa's power as it is
// now, through the layers and counters, so a prowess or other pump this
// turn opens it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "16414460-11ce-481d-ab54-6f0e0ea7d83d",
		Name:            "Kitsa, Otterball Elite",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "prowess"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Draw a card, then discard a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.TapCost(),
				Effect:  func(g *game.Game, item *game.StackItem) error { return effects.LootOne(g, item, 1) },
			},
			{
				Label:     "{2}, {T}: Copy target instant or sorcery spell you control. You may choose new targets for the copy. Activate only if Kitsa's power is 3 or greater.",
				Cost:      effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
				Targets:   effects.TargetInstantOrSorcerySpell("target instant or sorcery spell you control", effects.YouControl()),
				Condition: kitsaPowerThreeOrGreater,
				Effect:    effects.CopyTargetedSpell,
			},
		},
	})
}

// kitsaPowerThreeOrGreater is the copy's "Activate only if Kitsa's power
// is 3 or greater": the source's current power, layers and counters
// included.
func kitsaPowerThreeOrGreater(g *game.Game, _, source uuid.UUID) bool {
	c, ok := g.LookupCardForEffect(source)
	return ok && c.PowerForComparison() >= 3
}
