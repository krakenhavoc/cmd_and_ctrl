package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hostility — Creature — Elemental Incarnation {3}{R}{R}{R}, 6/6:
//
//	"Haste
//	 If a spell you control would deal damage to an opponent, prevent
//	 that damage. Create a 3/1 red Elemental Shaman creature token with
//	 haste for each 1 damage prevented this way.
//	 When Hostility is put into a graveyard from anywhere, shuffle it
//	 into its owner's library."
//
// The prevention is an ADR 0108 §8 static (#1906) on damage whose
// source is a spell its controller controls, on the stack or resolving
// (FromASpellYouControl), dealt to any of that player's opponents. Its
// subject is the source, so the additional effect runs once per spell
// in a damage instance with the total that spell would have dealt: a
// spell dealing 3 to each of three opponents makes nine Shamans. Damage
// from an ability, or from a creature, is not a spell's. Under damage
// that can't be prevented nothing is prevented and no token is made
// (CR 615.12, "for each 1 damage prevented this way").
//
// The last line is the graveyard trigger (#2038,
// WhenThisIsPutIntoAGraveyardFromAnywhereShuffleItIn): it fires
// however Hostility reached a graveyard, and if Hostility has left the
// graveyard by the time it resolves, its owner still shuffles (CR
// 701.24c).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "be16e9d8-aa73-448b-919c-7d02e2e15911",
		Name:            "Hostility",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Replacements: []game.ReplacementEffect{
			PreventDamageASourceWouldDeal(PreventionStatic{
				To:    ToAnOpponent,
				From:  FromASpellYouControl(),
				Then:  hostilityShamansBody,
				Label: "Hostility — prevent a spell's damage to an opponent and create a 3/1 Elemental Shaman for each 1 prevented",
			}),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisIsPutIntoAGraveyardFromAnywhereShuffleItIn("Hostility — shuffle it into its owner's library"),
		},
	})
}

// The additional effect: a 3/1 red Elemental Shaman with haste for each
// 1 damage prevented, under Hostility's controller.
var hostilityShamansBody = game.DelayedBody("hostility/elemental-shamans", hostilityShamans)

func hostilityShamans(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	if p.Amount <= 0 {
		return nil
	}
	return CreateToken{Controller: item.Controller, Template: TokenCard("3/1 red Elemental Shaman with haste"), N: p.Amount}.Apply(NewContext(g, item))
}
