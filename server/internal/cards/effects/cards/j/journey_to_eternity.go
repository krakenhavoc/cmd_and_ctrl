package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Journey to Eternity // Atzal, Cave of Eternity (Rivals of Ixalan, S82 OLD
// batch 217, #5531).
//
// Front face, Legendary Enchantment — Aura {1}{B}{G}:
//
//	"Enchant creature you control
//	 When enchanted creature dies, return it to the battlefield under your
//	 control, then return this card to the battlefield transformed under your
//	 control."
//
// Back face, Legendary Land:
//
//	"(Transforms from Journey to Eternity.)
//	 {T}: Add one mana of any color.
//	 {3}{B}{G}, {T}: Return target creature card from your graveyard to the
//	 battlefield."
//
// The dies trigger is harvested while the Aura is still attached (the dies event
// is emitted before the state-based action that puts a fallen-off Aura into the
// graveyard) and resolves with the Aura in a graveyard. It first returns the
// creature that died as a new object (CR 400.7) under the Aura's controller's
// control, as Kaya's Ghostform does, and a card that has since left the
// graveyard is nothing to return. Then it returns this card from the graveyard
// transformed (Radiant Grace's return, without an attachment: the back face is a
// land), when it is still the object that went there. The two halves are
// independent: a creature that can't come back (a token) still leaves the Aura
// returning as the Land.
//
// Atzal's mana ability is the any-colour tap; the second ability is Rise
// Again's reanimation of a creature card from your graveyard behind a {T}
// cost. The Land can tap the turn it transforms (CR 302.6 is about creatures).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     journeyToEternityOracleID,
		Name:         "Journey to Eternity",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(effects.YouControl()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.WhenEnchantedCreatureDies(
				"Journey to Eternity — return it to the battlefield under your control, then return this card transformed",
				func(g *game.Game, item *game.StackItem) error {
					if err := effects.Cmdr2B09ReturnTheEventCardUnderYourControl(g, item); err != nil {
						return err
					}
					ctx := effects.NewContext(g, item)
					if !effects.SourceCardInGraveyardAfterLeaving(ctx) {
						return nil
					}
					return effects.ReturnFromGraveyard{
						Target:      item.SourceCardID,
						Dest:        game.ZoneBattlefield,
						Controller:  item.Controller,
						Transformed: true,
					}.Apply(ctx)
				}),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     journeyToEternityOracleID + "#1",
		Name:         "Atzal, Cave of Eternity",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}: Add one mana of any color",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{B}{G}, {T}: Return target creature card from your graveyard to the battlefield.",
			Cost:    effects.Plus(effects.ManaCost("{3}{B}{G}"), effects.TapCost()),
			Targets: effects.TargetCreatureInYourGraveyard(),
			Purpose: effects.ForTargets(effects.Reanimates(0)),
			Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
				_, _ = effects.ReanimateSingleTarget(ctx, ctx.Controller())
				return nil
			})),
		}},
	})
}

const journeyToEternityOracleID = "7d6ccd0b-df16-40b2-930b-bcde0b6ef73f"
