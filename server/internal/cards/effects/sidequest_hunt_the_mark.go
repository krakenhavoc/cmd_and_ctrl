package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sidequest: Hunt the Mark // Yiazmat, Ultimate Mark — Enchantment //
// Legendary Creature — Dragon {3}{B}{B} (Final Fantasy, #4462):
//
//	Front: "When this enchantment enters, destroy up to one target
//	 creature.
//	 At the beginning of your end step, if a creature died under an
//	 opponent's control this turn, create a Treasure token. Then if you
//	 control three or more Treasures, transform this enchantment."
//	Back (5/6): "{1}{B}, Sacrifice another creature or artifact: Yiazmat
//	 gains indestructible until end of turn. Tap it."
//
// The front is Storm the Vault's pair of an entry trigger and an end-step
// transform; the back face is registered under "<oracle_id>#1". The enters
// trigger targets up to one creature as it goes on the stack and destroys
// what is still legal. The end-step ability has an intervening "if"
// (CR 603.4): it does not trigger at all unless a creature died under an
// opponent's control this turn (the per-player tally of creatures that
// died, tokens included), and that is checked again as it resolves. The
// Treasure comes first, then the Treasure count of three or more (the new
// one included) transforms the enchantment in place (TransformThis,
// CR 701.27). The back face is Immersturm Predator's indestructible
// sacrifice outlet with a mana cost and "another creature or artifact".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     sidequestHuntTheMarkOracleID,
		Name:         "Sidequest: Hunt the Mark",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Sidequest: Hunt the Mark — destroy up to one target creature", destroyTheFirstLegalTarget),
				UpToOneTargetCreature("up to one target creature")),
			On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && anOpponentCreatureDiedThisTurn(g, source.Controller)
			}, "Sidequest: Hunt the Mark — create a Treasure token; then if you control three or more Treasures, transform it",
				// The Treasure waits for no prompt of its own, but the
				// transform waits for the token to exist (CR 608.2c, #4155).
				Do(
					Instruction(func(ctx *Context) error {
						// CR 603.4's second check, as the ability resolves.
						if !anOpponentCreatureDiedThisTurn(ctx.Game, ctx.Controller()) {
							return nil
						}
						return CreateToken{Template: TreasureToken(), N: 1}.Apply(ctx)
					}),
					Instruction(func(ctx *Context) error {
						if !anOpponentCreatureDiedThisTurn(ctx.Game, ctx.Controller()) {
							return nil
						}
						return sidequestHuntTheMarkTransform(ctx)
					}),
				)),
		},
	})
}

// sidequestHuntTheMarkOracleID is shared with the back face, which
// registers under it plus "#1" (game.CatalogKey).
const sidequestHuntTheMarkOracleID = "d74f9a18-8dab-4549-be3f-9697ff68d73c"

// anOpponentCreatureDiedThisTurn is "a creature died under an opponent's
// control this turn": some other player's per-turn tally of creatures that
// died while they controlled them is above zero.
func anOpponentCreatureDiedThisTurn(g *game.Game, you uuid.UUID) bool {
	for _, p := range g.Seats {
		if p.ID != you && g.TurnTallyFor(p.ID).CreaturesDied > 0 {
			return true
		}
	}
	return false
}

// sidequestHuntTheMarkTransform is "then if you control three or more
// Treasures, transform this enchantment".
func sidequestHuntTheMarkTransform(ctx *Context) error {
	if countControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.HasSubtype("Treasure") }) < 3 {
		return nil
	}
	return TransformThis{}.Apply(ctx)
}
