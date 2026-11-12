package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Joyful Stormsculptor — Creature — Human Shaman {3}{U}{R}, 2/3 (March
// of the Machine, #5647):
//
//	"When this creature enters, create two 1/1 blue and red Elemental
//	 creature tokens.
//	 Whenever you cast a spell that has convoke, this creature deals 1
//	 damage to each opponent and each battle they protect."
//
// The second trigger asks whether the spell had convoke as it became
// cast (YouCastASpellThatHasConvoke): printed on it, given by a
// permanent, or given by a next-spell promise, with or without a
// creature tapped for it. The damage is one simultaneous event to every
// opponent and to every battle an opponent protects; damage to a battle
// removes that many defense counters (CR 120.3h). A battle the
// Stormsculptor's controller protects, or one with no protector, is not
// one "they protect".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "39867856-c1f3-442a-abe9-53f8f73777d0",
		Name:         "Joyful Stormsculptor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Joyful Stormsculptor — create two 1/1 blue and red Elemental tokens",
				Do(CreateToken{Template: TokenCard("1/1 blue and red Elemental"), N: 2})),
			On(game.EventCast, YouCastASpellThatHasConvoke,
				"Joyful Stormsculptor — 1 damage to each opponent and each battle they protect",
				func(g *game.Game, item *game.StackItem) error {
					return damageEachMatchingAndEachOpponent(NewContext(g, item), battleAnOpponentProtects, 1)
				}),
		},
	})
}

// battleAnOpponentProtects is "each battle they protect", "they" being
// the controller's opponents: a battle whose protector (CR 310.9) is a
// player other than `controller`.
func battleAnOpponentProtects(_ *game.Game, controller uuid.UUID, c game.Card) bool {
	return c.IsBattle() && c.ProtectorPlayerID != uuid.Nil && c.ProtectorPlayerID != controller
}
