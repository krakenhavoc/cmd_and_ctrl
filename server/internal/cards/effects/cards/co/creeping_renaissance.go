package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Creeping Renaissance — {3}{G}{G} Sorcery:
//
//	"Choose a permanent type. Return all cards of the chosen type from
//	 your graveyard to your hand.
//	 Flashback {5}{G}{G}"
//
// The six permanent types (CR 110.4) are the options: artifact, battle,
// creature, enchantment, land, planeswalker. The choice is made as the spell
// resolves, then every card of that type in the controller's graveyard comes
// back, read as the answer lands (Crystal Chimes' return). A card with more
// than one type returns for any of them (CR 205.2b). Flashback is Coffin
// Purge's.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "e2375f13-fd45-4087-abd1-c884acc6c736",
		Name:             "Creeping Renaissance",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{5}{G}{G}")},
		Purpose:          game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var options []game.ChoiceOption
			for _, t := range creepingRenaissanceTypes {
				options = append(options, game.ChoiceOption{Label: t})
			}
			return effects.PickOption{
				Question: "Creeping Renaissance — choose a permanent type",
				Options:  options,
				ThenKey:  creepingRenaissanceThen,
			}.Apply(ctx)
		},
	})
}

var creepingRenaissanceTypes = []string{"artifact", "battle", "creature", "enchantment", "land", "planeswalker"}

var creepingRenaissanceThen = effects.OptionPickThen("option-pick/creeping-renaissance-type", func(ctx *effects.Context, r game.OptionPicked) error {
	if r.Option == nil {
		return nil
	}
	chosen := r.Option.Label
	return effects.ReturnWholeGraveyardToHand(ctx, func(c game.Card) bool { return c.HasCardType(chosen) })
})
