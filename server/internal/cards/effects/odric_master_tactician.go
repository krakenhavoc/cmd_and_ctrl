package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Odric, Master Tactician — Legendary Creature — Human Soldier
// {2}{W}{W}, 3/4:
//
//	"First strike
//	 Whenever Odric and at least three other creatures attack, you
//	 choose which creatures block this combat and how those creatures
//	 block."
//
// The trigger is battalion's shape with three others (Boros
// Strike-Captain's attackersYouControlOtherThan): Odric's own attack
// declaration with three other attackers in it. It resolves in the
// declare attackers step and makes a DeclarationControl for this
// combat, blockers only (game/declaration_control.go, #4154): in this
// combat's declare blockers step Odric's controller declares every
// defending player's blocks, under every restriction and requirement
// ("All blocking declarations must still be legal"), and may declare
// none ("You can decide that a creature won't block"). The engine has
// no cost to block, so the third ruling (the defender decides whether
// to pay one) has nothing to ask.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ade898df-14a5-460b-94f6-1f3f74d3ff95",
		Name:            "Odric, Master Tactician",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attackDeclared(ev, source) && attackersYouControlOtherThan(g, source.Controller, source.InstanceID) >= 3
			}, "Odric, Master Tactician — you choose which creatures block this combat and how they block", func(g *game.Game, item *game.StackItem) error {
				g.BeginDeclarationControlForEffect(game.DeclarationControl{
					Declarer:   item.Controller,
					Source:     item.SourceCardID,
					SourceName: "Odric, Master Tactician",
					Blockers:   true,
					ThisCombat: true,
				})
				return nil
			}),
		},
	})
}
