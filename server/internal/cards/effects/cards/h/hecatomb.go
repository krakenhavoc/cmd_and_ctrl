package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hecatomb — Enchantment {1}{B}{B} (Ice Age, S82 OLD batch 840, #5636):
//
//	"When this enchantment enters, sacrifice this enchantment unless you sacrifice
//	 four creatures.
//	 Tap an untapped Swamp you control: This enchantment deals 1 damage to any
//	 target."
//
// Primeval Force's tax over four creatures: a two-way choice when its controller
// has four creatures to give, and an unconditional sacrifice of Hecatomb when they
// have fewer. The four are chosen and sacrificed together (CR 701.21). The ability
// is paid by tapping an untapped Swamp, which is not the {T} symbol (CR 302.6), so
// a Swamp that arrived this turn may pay it; the damage is Hecatomb's, dealt by its
// last-known information if it has left (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0ec153ee-2a91-4196-9105-212dd1864aea",
		Name:         "Hecatomb",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Hecatomb — sacrifice this enchantment unless you sacrifice four creatures", hecatombTax),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap an untapped Swamp you control: This enchantment deals 1 damage to any target.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Cost:    effects.TapAnUntapped("an untapped Swamp you control", effects.Subtype("Swamp")),
			Targets: effects.TargetAny(),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(1),
		}},
	})
}

func hecatombTax(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if len(effects.PermanentsControlledByMatching(g, item.Controller, effects.Creature())) < 4 {
		return effects.SacrificePermanent{Target: item.SourceCardID}.Apply(ctx)
	}
	return effects.PickOption{
		Question: "Hecatomb — sacrifice four creatures, or sacrifice Hecatomb",
		Options: []game.ChoiceOption{
			{Label: "Sacrifice four creatures"},
			{Label: "Sacrifice Hecatomb"},
		},
		ThenKey: hecatombAnswer,
	}.Apply(ctx)
}

// hecatombAnswer is the tax's choice answered, keyed: index 0 asks for four
// creatures, any other sacrifices the enchantment. An on-disk key: never renamed,
// never reused.
var hecatombAnswer = effects.OptionPickThen("option-pick/hecatomb-four-creatures-or-this", func(ctx *effects.Context, r game.OptionPicked) error {
	if r.Index != 0 {
		return effects.SacrificePermanent{Target: ctx.Source()}.Apply(ctx)
	}
	return effects.ChoosePermanents{
		Question: "Hecatomb — sacrifice four creatures",
		Keyed:    hecatombCreatures,
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			return effects.PermanentsControlledByMatching(g, of, effects.Creature()), 4, 4
		},
	}.Apply(ctx)
})

// hecatombCreatures sacrifices the four chosen creatures and nothing else.
var hecatombCreatures = effects.RegisterPermanentsPick("prompt-run/hecatomb-four-creatures", effects.PermanentsDef{Sacrifice: true})
