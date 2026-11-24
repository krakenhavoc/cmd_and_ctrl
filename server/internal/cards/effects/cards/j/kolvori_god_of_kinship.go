package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kolvori, God of Kinship // The Ringhart Crest — a modal double-faced card
// (oracle 490ac985…), one file for both faces; the back is the
// "<oracle>#1" row (the MDFC keyspace Esika, God of the Tree uses).
//
// Front, Legendary Creature — God {2}{G}{G}, 2/4:
//
//	"As long as you control three or more legendary creatures, Kolvori gets
//	 +4/+2 and has vigilance.
//	 {1}{G}, {T}: Look at the top six cards of your library. You may reveal a
//	 legendary creature card from among them and put it into your hand. Put
//	 the rest on the bottom of your library in a random order."
//
// Back, Legendary Artifact {1}{G}:
//
//	"As The Ringhart Crest enters, choose a creature type.
//	 {T}: Add {G}. Spend this mana only to cast a creature spell of the chosen
//	 type or a legendary creature spell."
//
// The static is Descend's shape: a layer 7c pump and a layer 6 keyword on the
// same live count of the legendary creatures its controller controls (Kolvori
// herself included), read every pass. The look is Icon of Ancestry's. The Crest's
// mana carries one OR restriction (ManaRestrictAnyOf): a cast of a creature
// spell of the chosen type, or a cast of a legendary creature spell; never an
// ability. With no type chosen yet only the legendary clause can use it.
//
// No simplification.
func init() {
	const oracle = "490ac985-79b5-438d-a2e7-f47683b29683"
	threeLegends := func(g *game.Game, you uuid.UUID) bool {
		n := 0
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == you && c.IsCreature() && c.IsLegendary() {
				n++
			}
		}
		return n >= 3
	}
	kw := effects.KeywordGrant(effects.WhileYouHave(threeLegends, effects.SelfCreature), "vigilance")
	kw.Label = "As long as you control three or more legendary creatures, Kolvori has vigilance."
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Kolvori, God of Kinship",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:     "As long as you control three or more legendary creatures, Kolvori gets +4/+2.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: effects.WhileYouHave(threeLegends, effects.SelfCreature),
				Apply:     effects.PumpBy(4, 2),
			},
			kw,
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}, {T}: Look at the top six cards of your library. You may reveal a legendary creature card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tutors: 1},
			Cost:    effects.Plus(effects.ManaCost("{1}{G}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				player := item.Controller
				return effects.TakeFromLibraryToHand{
					Player: player,
					Cards:  g.LookAtTopOfLibraryForEffect(player, 6),
					Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
						return c.IsCreature() && c.IsLegendary()
					},
					Max:      1,
					Optional: true,
					Label:    "Kolvori, God of Kinship — reveal a legendary creature card and put it into your hand",
					Key:      effects.TakeRevealRestBottomRandom,
				}.Apply(ctx)
			},
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "The Ringhart Crest",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseCreatureTypeAsEnters("The Ringhart Crest"),
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}. Spend this mana only to cast a creature spell of the chosen type or a legendary creature spell",
			RestrictionsFunc: func(g *game.Game, _ uuid.UUID, source uuid.UUID) []string {
				return []string{game.ManaRestrictAnyOf(
					[]string{effects.ManaRestrictCast, effects.ManaRestrictType("Creature"), effects.ManaRestrictSubtype(g.NamedTribeOf(source))},
					[]string{effects.ManaRestrictCast, effects.ManaRestrictType("Creature"), effects.ManaRestrictSupertype("Legendary")},
				)}
			},
		}},
	})
}
