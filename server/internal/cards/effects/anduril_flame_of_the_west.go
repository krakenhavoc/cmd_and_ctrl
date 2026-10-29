package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Andúril, Flame of the West — Legendary Artifact — Equipment {3}
// (The Hobbit):
//
//	"Equipped creature gets +3/+1.
//	 Whenever equipped creature attacks, create two tapped 1/1 white
//	 Spirit creature tokens with flying. If that creature is legendary,
//	 instead create two of those tokens that are tapped and attacking.
//	 Equip {2}"
//
// The pump is a layer-7c static on the host. The attack trigger fires for
// the equipped creature only (the Staff of Titania window) and reads the
// "if that creature is legendary" when the ability RESOLVES (the clause sits
// inside the effect, it is no intervening if): the creature as it is now when it
// is still the object that attacked, and its last-known information
// (CR 608.2h) when it has left the battlefield or become a new object.
// Legendary tokens go in attacking (Hero of Bladehold's shape: their
// controller chooses whom they attack, the creature's own defender is
// the default, and they never "attacked", so nothing triggers); a
// nonlegendary attacker makes the same two Spirits tapped, not attacking.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e6521404-8474-4727-b2d0-537d15d8a63f",
		Name:         "Andúril, Flame of the West",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(3, 1),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source.IsAttachedTo(ev.CardID)
			}, "Andúril — create two tapped 1/1 white Spirit tokens with flying (tapped and attacking if the creature is legendary)",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if andurilCreatureIsLegendary(g, item) {
						return CreateTokensAttackingYourChoice{
							Template: TokenCard("1/1 white Spirit with flying"),
							N:        2,
							Tapped:   true,
							Prefer:   triggerAttackTarget(item),
							Label:    "Andúril",
						}.Apply(ctx)
					}
					spirit := TokenCard("1/1 white Spirit with flying")
					spirit.Tapped = true
					return CreateToken{Template: spirit, N: 2}.Apply(ctx)
				}),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}

// andurilCreatureIsLegendary reads "that creature is legendary": the
// attacker as it is now while it is still the object that attacked, its
// last-known supertypes once it is gone (CR 400.7, CR 608.2h).
func andurilCreatureIsLegendary(g *game.Game, item *game.StackItem) bool {
	if item == nil || item.Trigger == nil {
		return false
	}
	id := item.Trigger.Event.CardID
	o := item.Trigger.Object
	if c, ok := g.LookupCardForEffect(id); ok && (o == nil || g.ObjectEpochForEffect(id) == o.Epoch) {
		return c.IsLegendary()
	}
	if o == nil {
		return false
	}
	for _, s := range o.Supertypes {
		if strings.EqualFold(s, "legendary") {
			return true
		}
	}
	return false
}
