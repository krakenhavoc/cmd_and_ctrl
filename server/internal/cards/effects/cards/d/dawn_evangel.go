package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dawn Evangel — Enchantment Creature — Human Cleric {2}{W}, 2/3:
//
//	"Whenever a creature dies, if an Aura you controlled was attached to it,
//	 return target creature card with mana value 2 or less from your graveyard
//	 to your hand."
//
// The Auras are read as they were attached when the creature died (CR 603.10a,
// CR 608.2h), and the intervening "if" is checked again as the trigger
// resolves (CR 603.4): it names the Auras' controller at that time.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4d6e38f6-480b-4f81-a5d4-342fc45e39b0",
		Name:         "Dawn Evangel",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventLTB, dawnEvangelCondition,
					"Dawn Evangel — return target creature card with mana value 2 or less from your graveyard to your hand",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						if item.Trigger == nil || !dawnEvangelAura(g, item.Trigger.Event, item.Controller) {
							return nil
						}
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}.Apply(ctx)
					}),
				effects.TargetCardInGraveyard("target creature card with mana value 2 or less from your graveyard",
					effects.Creature(), effects.ManaValueLE(2), effects.YouOwn()),
			),
		},
	})
}

// dawnEvangelAura is "an Aura you controlled was attached to it".
func dawnEvangelAura(g *game.Game, ev game.Event, you uuid.UUID) bool {
	if _, died := effects.DiedCreature(ev, g); !died {
		return false
	}
	for _, controller := range effects.SocB6AurasOnTheDead(g, ev) {
		if controller == you {
			return true
		}
	}
	return false
}

func dawnEvangelCondition(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return dawnEvangelAura(g, ev, source.Controller)
}
