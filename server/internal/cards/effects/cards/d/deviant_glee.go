package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deviant Glee — Enchantment — Aura {B} (Return to Ravnica, S82 OLD):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+1 and has "{R}: This creature gains trample
//	 until end of turn.""
//
// Fearsome Temper's pump-and-grant shape: the +2/+1 is a layer-7c modify of
// the host, and the {R} ability is a granted activated ability (ADR 0093),
// activated by whoever controls the enchanted creature, with the creature
// itself gaining trample.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "13aeccd5-cca2-4249-904d-d5516f977440",
		Name:         "Deviant Glee",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: deviantGleeGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{R}: This creature gains trample until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    effects.ManaCost("{R}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"trample"}, Label: "Deviant Glee — trample until end of turn"}.Apply(effects.NewContext(g, item))
				},
			}},
			Text: "{R}: This creature gains trample until end of turn.",
		}},
		Static: []game.StaticAbility{effects.PumpAttached(2, 1), effects.GrantAbilitiesToAttached(deviantGleeGrant)},
	})
}

// deviantGleeGrant is the granted ability's bundle key: an on-disk identity,
// never renamed, never reused.
const deviantGleeGrant = "deviant-glee/gain-trample"
