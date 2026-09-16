package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bitterblossom — Kindred Enchantment — Faerie for {1}{B}:
//
//	"At the beginning of your upkeep, you lose 1 life and create a
//	1/1 black Faerie Rogue creature token with flying."
//
// S19 sub-PR 5: a "your upkeep" trigger combining a life cost with
// a token. Mandatory; the trigger goes on the stack at upkeep and
// resolves to the life loss + token. The Faerie token's flying is
// real: FaerieRogueToken carries it on Card.Keywords, which
// printedCharacteristic folds into the layer engine (S21 sub-PR 1).
// The "cosmetic" note here outlived the fix; corrected in the #338
// sweep.
func init() {
	Register(Spec{
		OracleID:     "fb868840-09fa-49b1-85cb-b08ad065e972",
		Name:         "Bitterblossom",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The Faerie Rogue token is created colorless instead of black, so anything that cares about a creature's color doesn't see it."},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Bitterblossom — lose 1 life, create a Faerie Rogue", func(g *game.Game, item *game.StackItem) error {
				if err := g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -1); err != nil {
					return err
				}
				return CreateToken{
					Controller: item.Controller,
					Template:   FaerieRogueToken(),
					N:          1,
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
