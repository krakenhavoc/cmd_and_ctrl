package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Radiant Lotus — Artifact {6}:
//
//	"{T}, Sacrifice one or more artifacts: Choose a color. Target
//	 player adds three mana of the chosen color for each artifact
//	 sacrificed this way."
//
// The "one or more" half of #1213's variable-count sacrifice row, and
// the card that needed the payment RECORDED rather than recomputed.
//
// "For each artifact sacrificed this way" is read at resolution, and
// by then the artifacts are in graveyards — a resolution that counted
// the board would answer zero, and one that counted the graveyard
// would count everything that ever died. So the announcement writes
// down how many it paid (PaidCost.Sacrificed) and the effect reads it
// back with ctx.Sacrificed(), exactly as a variable counter cost's
// count has been read back since #789.
//
// The count is the ACTIVATOR's, not the card's: the floor is one and
// there is no printed ceiling, so the board is the only bound. The
// Lotus itself is an artifact you control and is a legal thing to
// sacrifice to its own cost, which is the line — tap it, eat it and
// two other rocks, and hand somebody nine mana.
//
// It is NOT a mana ability (CR 605.1a): it targets a player, so it
// uses the stack and can be responded to. The printed reminder text
// "(Activate only as an instant.)" is saying exactly that and adds no
// restriction of its own.
//
// TARGET PLAYER adds the mana, which is why the effect calls the
// engine's add-mana helper for that seat rather than for the
// controller. Handing an opponent nine mana is a real Commander play
// (a deal, a Tempt-style bribe, or a way to make a symmetrical payoff
// asymmetrical) and the card is written the way it is printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "307be184-176a-40a8-944c-48aa00cfdd29",
		Name:         "Radiant Lotus",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{T}, Sacrifice one or more artifacts: Choose a color. Target player adds three mana of the chosen color for each artifact sacrificed this way.",
			Cost: Plus(
				TapCost(),
				SacrificeOneOrMore("one or more artifacts", Artifact()),
			),
			Targets: TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				// The PAYMENT, not the board: the artifacts are in
				// graveyards by now (#1213).
				paid := ctx.Sacrificed()
				if paid <= 0 {
					return nil
				}
				var receiver game.TargetRef
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetPlayer {
						receiver = t
						break
					}
				}
				if receiver.Kind != game.TargetPlayer {
					return nil
				}
				// CR 105.4: the colour is the ABILITY's controller's
				// choice even though the mana lands on the target.
				// The purpose is declared (#780) so a chooser with no
				// other information knows what the answer buys.
				source, player := item.SourceCardID, receiver.ID
				ChooseColorThen(game.ColorForMana, g, item.Controller, source,
					"Radiant Lotus — choose a color",
					func(g *game.Game, color string) error {
						if color == "" {
							return nil
						}
						return g.AddManaForEffect(player, source,
							strings.Repeat("{"+color+"}", 3*paid))
					})
				return nil
			},
		}},
	})
}
