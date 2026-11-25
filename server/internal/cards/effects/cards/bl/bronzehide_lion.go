package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bronzehide Lion — Creature — Cat {G}{W}, 3/3 (Theros Beyond Death, #5514):
//
//	"{G}{W}: This creature gains indestructible until end of turn.
//	 When this creature dies, return it to the battlefield. It's an Aura
//	 enchantment with enchant creature you control and "{G}{W}: Enchanted
//	 creature gains indestructible until end of turn," and it loses all other
//	 abilities."
//
// The returned-Aura seam (returned_aura.go). Unlike Harold and Bob, the Aura
// HAS the ability itself, so its grant goes to the returned permanent
// (S82OldB133ReturnedAuraSelfForm) and its effect reads the creature the Aura
// is attached to. The Aura enchants a creature its controller controls as it
// enters (CR 303.4f); with none, it stays in the graveyard (CR 303.4g). Unlike
// the Troll there is no "if it was a creature" clause: the Aura dying returns
// nothing because it has lost the trigger.
//
// No simplification.
func init() {
	const grant = "bronzehide-lion/enchanted-creature-gains-indestructible"
	effects.Register(effects.Spec{
		OracleID:     "0db3b29e-e21c-4240-8e5e-959ba6bcf706",
		Name:         "Bronzehide Lion",
		Completeness: effects.CompletenessFull,
		ReturnedAura: effects.EnchantCreature(effects.YouControl()),
		Grants: []effects.AbilityGrant{{
			Key:  grant,
			Text: "{G}{W}: Enchanted creature gains indestructible until end of turn.",
			Activated: []effects.ActivatedAbility{{
				Label:   "{G}{W}: Enchanted creature gains indestructible until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    effects.ManaCost("{G}{W}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id := effects.EnchantedCreatureOf(ctx)
					if id == uuid.Nil {
						return nil
					}
					return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"indestructible"}, Label: "Bronzehide Lion"}.Apply(ctx)
				},
			}},
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}{W}: This creature gains indestructible until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{G}{W}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"indestructible"}, Label: "Bronzehide Lion"}.Apply(ctx)
			},
		}},
		Static: effects.S82OldB133ReturnedAuraSelfForm(grant),
		Triggered: []game.TriggeredAbility{
			// Printed without "if it was a creature": the Aura loses this trigger.
			effects.WhenThisDiesReturnItAsAnAura("Bronzehide Lion"),
		},
	})
}
