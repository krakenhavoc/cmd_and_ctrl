package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pippinCardTypes are the card types a player may choose, in the order they
// are offered.
var pippinCardTypes = []string{"Artifact", "Battle", "Creature", "Enchantment", "Instant", "Kindred", "Land", "Planeswalker", "Sorcery"}

// Pippin, Guard of the Citadel — Legendary Creature — Halfling Soldier {W}{U},
// 2/2 (The Lord of the Rings: Tales of Middle-earth Commander):
//
//	"Vigilance, ward {1}
//	 {T}: Another target creature you control gains protection from the
//	 card type of your choice until end of turn."
//
// Vigilance is printed and ward {1} is the ward trigger. The activation is
// Mother of Runes's protection from a chosen quality, asked as it resolves
// (a keyed option pick over the card types, as ProtectionFromCardType
// words them) and applied to the target only if it is still on the
// battlefield. "Another" keeps Pippin off his own clause.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "25afddd2-6e48-4742-b34e-e3b5def7c0fd",
		Name:            "Pippin, Guard of the Citadel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{1}"), "Pippin, Guard of the Citadel — ward {1}"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Another target creature you control gains protection from the card type of your choice until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "protection"))},
			Cost:    effects.TapCost(),
			Targets: effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				options := make([]game.ChoiceOption, len(pippinCardTypes))
				for i, t := range pippinCardTypes {
					options[i] = game.ChoiceOption{Label: t}
				}
				return effects.PickOption{
					Question: "Pippin, Guard of the Citadel — choose a card type",
					Options:  options,
					ThenKey:  pippinProtect,
					Carry:    []uuid.UUID{id},
				}.Apply(ctx)
			},
		}},
	})
}

var pippinProtect = effects.OptionPickThen("option-pick/pippin-guard-of-the-citadel-card-type", func(ctx *effects.Context, r game.OptionPicked) error {
	if len(r.Carry) == 0 || r.Index < 0 || r.Index >= len(pippinCardTypes) || !effects.OnBattlefield(ctx.Game, r.Carry[0]) {
		return nil
	}
	tok := game.ProtectionFromCardType(pippinCardTypes[r.Index])
	if tok == "" {
		return nil
	}
	return effects.GrantKeywordUntilEOT{Target: r.Carry[0], Keywords: []string{tok}, Label: "Pippin, Guard of the Citadel — " + tok}.Apply(ctx)
})
