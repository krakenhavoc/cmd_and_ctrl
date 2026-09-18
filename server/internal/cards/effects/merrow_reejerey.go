package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Merrow Reejerey — 2/2 Creature — Merfolk Soldier for {2}{U} (EDHREC
// rank 4000):
//
//	"Other Merfolk creatures you control get +1/+1.
//	 Whenever you cast a Merfolk spell, you may tap or untap target
//	 permanent."
//
// The Merfolk lord that is also the Merfolk engine: every Merfolk the
// deck casts either unlocks a blocker or untaps a land, and in a
// tribal deck that is most of the turn. It is in the batch as the
// only card in the catalog whose trigger offers a CHOICE BETWEEN TWO
// ACTIONS on one target, which is a different shape from a modal
// spell (the choice is made on resolution, not at announce, and there
// is no mode picker on a trigger).
//
// # The two halves are independent
//
// The anthem says "other Merfolk CREATURES you control" — the
// Reejerey does not pump itself, and it does not pump the Merfolk
// deck across the table. The trigger says "a Merfolk SPELL", which is
// wider: a Merfolk enchantment or artifact would count, and so does a
// Merfolk creature spell that gets countered, because the trigger is
// on the CAST.
//
// Both read effective types, so a changeling is a Merfolk on both
// counts.
//
// # "Target permanent" is anybody's
//
// Tap an opposing blocker before attacking, or untap your own land
// to cast the next Merfolk. Nothing restricts it to creatures or to
// your side.
//
// # "You may tap or untap"
//
// The decline is the trigger's own optional prompt, asked when the
// trigger would go on the stack (the Curiosity posture in this
// catalog); the tap-or-untap pick is asked when it resolves, which is
// the moment the card's choice is actually made. Picking untap on an
// already-untapped permanent, or tap on a tapped one, is legal and
// does nothing — the card places no condition on the target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6b7e6ae4-2ee1-44bf-ac93-79fc87494515",
		Name:         "Merrow Reejerey",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Merfolk"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			Optional(
				Targeting(
					WheneverYouCast(HasSubtype("Merfolk"), "Merrow Reejerey — tap or untap target permanent",
						b38TapOrUntapTarget("Merrow Reejerey")),
					TargetPermanent("target permanent")),
				"Merrow Reejerey — tap or untap target permanent?"),
		},
	})
}
