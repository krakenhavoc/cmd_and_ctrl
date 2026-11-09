package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Item Shopkeep — Creature — Human Citizen {1}{R}, 2/2 (Final Fantasy,
// #4462):
//
//	"Whenever you attack, target attacking equipped creature gains menace
//	 until end of turn."
//
// Boggart Prankster's trigger (once per attack declaration of yours,
// whoever attacks, aimed as it goes on the stack) with an attacking
// equipped creature as the target: "equipped" is an Equipment attached to
// it (isEquippedCreature), judged as the trigger targets and again as it
// resolves (CR 608.2b). The Shopkeep need not attack. Once the target is
// still legal the grant is made whatever then happens to the Equipment.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7532f0ea-9d13-43ca-9b7d-2d22f2a66017",
		Name:         "Item Shopkeep",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(Targeting(
				On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclaredByYou(ev, source.Controller)
				}, "Item Shopkeep — target attacking equipped creature gains menace until end of turn",
					itemShopkeepMenace),
				TargetCreature("target attacking equipped creature", AttackingCreature(),
					func(g *game.Game, _ uuid.UUID, c game.Card) bool { return isEquippedCreature(g, c) }))),
		},
	})
}

// itemShopkeepMenace gives the still-legal target menace until end of turn.
func itemShopkeepMenace(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return GrantKeywordUntilEOT{
		Target:   id,
		Keywords: []string{"menace"},
		Label:    "Item Shopkeep — menace until end of turn",
	}.Apply(ctx)
}
