package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Damping Sphere — Artifact {2}:
//
//	"If a land is tapped for two or more mana, it produces {C}
//	 instead of any other type and amount.
//	 Each spell a player casts costs {1} more to cast for each other
//	 spell that player has cast this turn."
//
// THE LAND HALF is a CR 614 replacement on the mana production event,
// and CR 106.12b says which event: the one the land's mana ability makes
// "while such an ability is resolving and producing" the mana. Since
// #3939 that is ONE event per activation, carrying the whole output
// (game/produce_mana.go), so the Sphere sees Ancient Tomb's {C}{C} and an
// Azorius Chancery's {W}{U} as two mana and turns each into one {C}. A
// filter land's two picks are asked first and produced together, so its
// {W}{U} is caught too. A basic land's one mana is left alone.
//
// "Two or more" is read off the event as it stands, so the order of
// replacements matters, as CR 616.1 says it does. On a Forest, Mana
// Reflection makes the mana {G}{G} first, and the Sphere then applies to
// two mana and makes it {C}. On Ancient Tomb both apply from the start,
// and the order decides between {C} and {C}{C}. A production cannot
// pause to ask (CR 605.3b), so the window applies them in the order it
// gathers them, as it does for every production replacement (ADR 0013
// §5ab Decision 2).
//
// Only a land TAPPED for mana counts (CR 106.12): the event's
// ManaFromTap, which a cracked Treasure, a spell's "Add" and a
// triggered mana ability's own output all leave false.
//
// THE STORM HALF is the catalog's first SCALING cost modifier: the
// amount is a function of the board rather than a printed number, which
// is the whole reason CostModifier.Amount is a hook. "Each OTHER spell
// that player has cast this turn" comes out right for free. The per-turn
// tally (Game.SpellsCastThisTurn) is bumped AFTER the spell reaches the
// stack, and CR 601.2f prices the cast before that, so the count this
// reads already excludes the spell being priced. The first spell of a
// turn is taxed nothing, the second {1}, the third {2}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fd50d76c-7654-47b9-a5b6-d075874e4357",
		Name:         "Damping Sphere",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventManaAdded},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) bool {
				if ev.Kind != game.RepEventProduceMana || !ev.ManaFromTap || len(ev.ManaColors) < 2 {
					return false
				}
				land, ok := g.LookupCardForEffect(ev.ManaSource)
				return ok && land.IsLand()
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.ManaColors = []string{"C"}
				return nil
			},
			// The affected player chooses the order (CR 616.1): the one
			// whose land is producing the mana.
			Controller: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) uuid.UUID {
				return ev.ManaPlayer
			},
			Label: "Damping Sphere: a land tapped for two or more mana produces {C}",
		}},
		CostModifiers: []game.CostModifier{
			CostsMoreEach(OtherSpellsCastThisTurn(),
				"Each spell a player casts costs {1} more to cast for each other spell that player has cast this turn."),
		},
	})
}
