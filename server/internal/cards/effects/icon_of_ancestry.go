package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Icon of Ancestry — Artifact {3}:
//
//	"As this artifact enters, choose a creature type.
//	 Creatures you control of the chosen type get +1/+1.
//	 {3}, {T}: Look at the top three cards of your library. You may reveal a
//	 creature card of the chosen type from among them and put it into your
//	 hand. Put the rest on the bottom of your library in a random order."
func init() {
	Register(Spec{
		OracleID:     "2b4973cc-83b0-4986-a19e-b4252706527e",
		Name:         "Icon of Ancestry",
		Completeness: CompletenessFull,
		AsEnters:     ChooseCreatureTypeAsEnters("Icon of Ancestry"),
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Chosen: true, YoursOnly: true}, 1, 1),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}, {T}: Look at the top three cards of your library. You may reveal a creature card of the chosen type from among them and put it into your hand. Put the rest on the bottom of your library in a random order.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tutors: 1},
			Cost:    Plus(ManaCost("{3}"), TapCost()),
			Effect:  iconOfAncestryLook,
		}},
	})
}

func iconOfAncestryLook(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	tribe := g.NamedTribeOf(item.SourceCardID)
	player := item.Controller
	return TakeFromLibraryToHand{
		Player: player,
		Cards:  g.LookAtTopOfLibraryForEffect(player, 3),
		Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
			return tribe != "" && c.IsCreature() && c.HasSubtype(tribe)
		},
		Max:      1,
		Optional: true,
		Label:    "Icon of Ancestry — reveal a creature card of the chosen type and put it into your hand",
		Key:      takeRevealRestBottomRandom,
	}.Apply(ctx)
}
