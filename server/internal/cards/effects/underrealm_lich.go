package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Underrealm Lich — Creature — Zombie Elf Shaman {3}{B}{G}, 4/3:
//
//	"If you would draw a card, instead look at the top three cards of
//	 your library, then put one into your hand and the rest into your
//	 graveyard.
//	 Pay 4 life: This creature gains indestructible until end of turn.
//	 Tap it."
//
// The replacement is mandatory and replaces EACH draw of yours, a draw
// step's and an effect's alike (#2168). It cancels the draw (so nothing
// is "drawn": Consecrated Sphinx and Sheoldred stay quiet, which is the
// rule) and its body looks at the top three — only the looker sees
// them — asks which one to take, and puts the rest into the graveyard
// through the CR 614 window, so Rest in Peace sees them. With fewer
// than three cards the player looks at what there is; with an empty
// library nothing is looked at and nobody loses, because nothing was
// drawn (CR 704.5b needs an attempt to DRAW). A "draw three" asks three
// times, one after the other (CR 121.6b).
//
// The second ability is an ordinary CR 602 ability with a life cost.
// "Tap it" is part of the effect, not the cost, so a tapped Lich still
// gains indestructible.
//
// A draw a doubler made into two is two draws, and the replacement runs for
// each.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e1bce9c3-300c-4a9d-abe0-a1f02d3a1105",
		Name:         "Underrealm Lich",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventDrawCard},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventDraw && ev.DrawCount > 0 && src != nil && ev.DrawPlayer == src.Controller
			},
			DrawInstead: game.RegisterDrawInstead("underrealm-lich", underrealmLichLook),
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Underrealm Lich: look at the top three, put one into your hand and the rest into your graveyard",
		}},
		Activated: []ActivatedAbility{{
			Label: "Pay 4 life: This creature gains indestructible until end of turn. Tap it.",
			Cost:  PayLife(4),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (GrantKeywordUntilEOT{
					Target:   item.SourceCardID,
					Keywords: []string{"indestructible"},
					Label:    "Underrealm Lich — indestructible",
				}).Apply(ctx); err != nil {
					return err
				}
				return TapTarget{Target: item.SourceCardID}.Apply(ctx)
			},
		}},
	})
}

// underrealmLichLook is the replacement's body: a look at three, a
// mandatory pick for the hand, the rest into the graveyard, and then
// the rest of the draw instruction.
func underrealmLichLook(g *game.Game, drawer, source uuid.UUID, done func(*game.Game) error) error {
	item := &game.StackItem{Controller: drawer, SourceCardID: source}
	return TakeFromLibraryToHand{
		Player: drawer,
		Cards:  g.LookAtTopOfLibraryForEffect(drawer, 3),
		Max:    1,
		Label:  "Underrealm Lich — put one of the top three cards into your hand (the rest go to your graveyard)",
		Then: func(g *game.Game, res TakeFromLibraryResult) error {
			if err := TakeRestIntoGraveyard(g, res); err != nil {
				return err
			}
			return done(g)
		},
	}.Apply(NewContext(g, item))
}
