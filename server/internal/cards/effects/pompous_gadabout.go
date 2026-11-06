package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pompous Gadabout — Creature — Human Citizen {2}{G}, 4/2 (Murders at
// Karlov Manor, #4469):
//
//	"During your turn, this creature has hexproof.
//	 This creature can't be blocked by creatures that don't have a name."
//
// The hexproof is Feisty Spikeling's first strike: a layer 6 grant on
// itself that holds only while its controller is the active player. The
// block rule is Legolas Greenleaf's CantBeBlockedBy with "creatures that
// don't have a name" as the forbidden blockers: a face-down creature has
// no name (CR 708.2a), and neither has any other creature whose name is
// empty.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0361b9b7-9097-4aae-aa06-c63a5fb47a76",
		Name:         "Pompous Gadabout",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return thisPermanent(target, g, source) && isActivePlayer(g, source.Controller)
			}, game.KeywordHexproof),
		},
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnSelf(), creatureWithoutAName, "creatures that don't have a name"),
		},
	})
}

// creatureWithoutAName is "creatures that don't have a name": a face-down
// permanent (CR 708.2a), or one whose name is empty.
func creatureWithoutAName(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	return c.FaceDown || strings.TrimSpace(c.Effective().Name) == ""
}
