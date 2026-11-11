package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const predatorsHourGrant = "predators-hour/combat-damage-exile-top-face-down"

// Predators' Hour — Sorcery {1}{B} (Outlaws of Thunder Junction Commander):
//
//	"Until end of turn, creatures you control gain menace and "Whenever this
//	 creature deals combat damage to a player, exile the top card of that
//	 player's library face down. You may look at and play that card for as
//	 long as it remains exiled, and you may spend mana as though it were mana
//	 of any color to cast that spell.""
//
// The creatures are the ones you control as it resolves (CR 611.2c). The
// quoted trigger is a catalog bundle granted until end of turn (Stir the
// Pride's shape), each creature's own ability, so "you" is that creature's
// controller. The exile is face down with the grant to that controller, who
// alone sees it: "spend mana as though it were mana of any color" is the
// AnyColor grant, and NOT mana of any type, so a {C} in the cost still needs
// colourless mana.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4d504d5f-02a4-4104-89b7-88884483f53a",
		Name:         "Predators' Hour",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: predatorsHourGrant,
			Triggered: []game.TriggeredAbility{
				WheneverThisDealsCombatDamageToAPlayer(
					"Predators' Hour — exile the top card of that player's library face down; you may play it",
					predatorsHourExile),
			},
			Text: "Whenever this creature deals combat damage to a player, exile the top card of that player's library face down. You may look at and play that card for as long as it remains exiled, and you may spend mana as though it were mana of any color to cast that spell.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			yours := And(Creature(), YouControl())
			if err := (GrantKeywordUntilEOT{
				Match:    yours,
				Keywords: []string{"menace"},
				Label:    "Predators' Hour — menace",
			}).Apply(ctx); err != nil {
				return err
			}
			return GrantAbilitiesFor{
				Match: yours,
				Keys:  []string{predatorsHourGrant},
				Label: "Predators' Hour — whenever this creature deals combat damage to a player, exile the top card of that player's library face down",
			}.Apply(ctx)
		},
	})
}

func predatorsHourExile(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || item.Controller == [16]byte{} {
		return nil
	}
	return ExileTopWithPermission{
		From:        item.Trigger.Event.Target,
		GrantTo:     item.Controller,
		N:           1,
		AnyColor:    true,
		FaceDown:    true,
		WhileExiled: true,
	}.Apply(NewContext(g, item))
}
