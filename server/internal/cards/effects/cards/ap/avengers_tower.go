package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Avengers Tower — Land (Marvel Super Heroes, #4455):
//
//	"{T}: Add {C}.
//	 {T}: Add one mana of any color. Spend this mana only to cast a Hero
//	 spell or to activate an ability of a Hero source.
//	 {4}, {T}: Look at the top three cards of your library. You may
//	 reveal a Hero card from among them and put it into your hand. Put
//	 the rest on the bottom of your library in any order."
//
// Two mana abilities, one plain and one restricted: the restriction is
// Castle Garenbrig's one ManaRestrictAnyOf tag, "cast a Hero spell" or
// "activate an ability of a Hero source" (CR 106.6), so a spell or an
// ability whose source is not a Hero can't spend it. The look is
// Duskwatch Recruiter's TakeFromLibraryToHand: private, the Hero taken
// is revealed, the rest go to the bottom in an order the player picks.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c5fc8e7c-a87e-4586-a13c-d30e0a3aafbf",
		Name:         "Avengers Tower",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			effects.ColorlessTapRow(),
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{W|U|B|R|G}",
				Label:    "{T}: Add one mana of any color (Hero spells and Hero abilities only).",
				Restrictions: []string{game.ManaRestrictAnyOf(
					[]string{game.ManaRestrictCast, game.ManaRestrictSubtype("Hero")},
					[]string{game.ManaRestrictActivate, game.ManaRestrictSubtype("Hero")},
				)},
			},
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}: Look at the top three cards of your library. You may reveal a Hero card from among them and put it into your hand. Put the rest on the bottom of your library in any order.",
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				player := ctx.Controller()
				return effects.TakeFromLibraryToHand{
					Player:   player,
					Cards:    g.LookAtTopOfLibraryForEffect(player, 3),
					Match:    effects.HasSubtype("Hero"),
					Max:      1,
					Optional: true,
					Label:    "Avengers Tower — you may reveal a Hero card from among them and put it into your hand",
					Key:      effects.TakeRevealRestBottomAny,
				}.Apply(ctx)
			},
		}},
	})
}
