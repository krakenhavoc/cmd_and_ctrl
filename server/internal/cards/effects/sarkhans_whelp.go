package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sarkhan's Whelp — Creature — Dragon {2}{R}, 2/2 (Jumpstart 2022, #5948):
//
//	"Flying
//	 Whenever you activate an ability of a Sarkhan planeswalker, this
//	 creature deals 1 damage to any target."
//
// Watches EventActivateAbility (a loyalty ability is an activated ability,
// CR 606.2) for an ability its controller activated whose source is a
// planeswalker with the Sarkhan subtype, read through the layered view. The
// Whelp is the source of the damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "482f5fc8-ca3b-4280-8cf7-a1c55f8b9f12",
		Name:            "Sarkhan's Whelp",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(
				On(game.EventActivateAbility, gapB158YouActivatedASarkhanAbility,
					"Sarkhan's Whelp — 1 damage to any target", sourceDealsDamageToEachLegalTarget(1)),
				TargetAny()), ForTargets(DamageToTarget(0, 1))),
		},
	})
}

func gapB158YouActivatedASarkhanAbility(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventActivateAbility || source == nil || ev.Actor != source.Controller {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsPlaneswalker() && c.HasSubtype("Sarkhan")
}
