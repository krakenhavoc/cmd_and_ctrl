package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scepter of Celebration — Artifact — Equipment {2}{G} (Streets of New Capenna
// Commander):
//
//	"Equipped creature gets +2/+0 and has trample.
//	 Whenever equipped creature deals combat damage to a player, create that
//	 many 1/1 green and white Citizen creature tokens.
//	 Equip {3}"
//
// Pollenbright Wings' trigger, read off the attachment: "that many" is the
// combat damage dealt, and the Equipment's controller gets the tokens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c930b1f3-9bc3-4708-8d05-02295483c9d8",
		Name:         "Scepter of Celebration",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(2, 0),
			GrantToAttached("trample"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			}, "Scepter of Celebration — create that many 1/1 green and white Citizen creature tokens",
				Do(CreateTokensForEach{
					Template: TokenCard("1/1 green and white Citizen"),
					Name:     "the combat damage dealt",
					Count:    func(ctx *Context) int { return ctx.Trigger().Event.Amount },
				})),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{3}"),
		},
	})
}
