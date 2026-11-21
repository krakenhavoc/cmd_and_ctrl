package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Osgood, Operation Double — Legendary Creature — Human Alien Shapeshifter
// {2}{U}{U}, 2/2 (Doctor Who, #5948):
//
//	"When you cast this spell, create a token that's a copy of it, except it
//	 isn't legendary.
//	 {T}: Add {C}. Spend this mana only to cast an artifact spell or activate
//	 an ability of an artifact.
//	 Paradox — Whenever you cast a spell from anywhere other than your hand,
//	 investigate."
//
// The cast trigger is WhenYouCastThisSpell: it fires while the creature
// spell is on the stack and its token copy is made from the spell's values,
// without the Legendary supertype (CR 707.2). The copy is a token, so it is
// never cast and has no cast trigger of its own, and being nonlegendary it
// does not meet the legend rule. Its mana ability is restricted as the
// printed one is (Ambulatory Reactor's restriction: an artifact spell, or an
// ability of an artifact source). Paradox reads the zone the spell was cast
// from, so a flashback, exile or command-zone cast investigates and a hand
// cast does not.
//
// No simplification.
func init() {
	const mana = "{T}: Add {C}. Spend this mana only to cast an artifact spell or activate an ability of an artifact."
	Register(Spec{
		OracleID:     "292aafef-8945-4827-90b2-1dd9764e517b",
		Name:         "Osgood, Operation Double",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    mana,
			Restrictions: []string{game.ManaRestrictAnyOf(
				[]string{game.ManaRestrictCast, game.ManaRestrictType("Artifact")},
				[]string{game.ManaRestrictActivate, game.ManaRestrictType("Artifact")},
			)},
		}},
		Triggered: []game.TriggeredAbility{
			WhenYouCastThisSpell("Osgood, Operation Double — create a token that's a copy of it, except it isn't legendary",
				func(g *game.Game, item *game.StackItem) error {
					return CreateTokenCopy{
						Controller: item.Controller,
						Copy:       item.SourceCardID,
						N:          1,
						Except: func(t *game.Card) {
							t.TypeLine = removeLegendaryFromTypeLine(t.TypeLine)
						},
					}.Apply(NewContext(g, item))
				}),
			On(game.EventCast, gapB89YouCastNotFromHand,
				"Osgood, Operation Double — investigate", Do(Investigate{N: 1})),
		},
	})
}
