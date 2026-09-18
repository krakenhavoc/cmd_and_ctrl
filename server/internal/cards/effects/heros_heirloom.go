package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hero's Heirloom — Artifact — Equipment for {2} (EDHREC rank 4008):
//
//	"Equipped creature gets +2/+1.
//	 As long as equipped creature is legendary, it has trample and
//	 haste.
//	 Equip {2}"
//
// Two mana down, two to equip, and a commander that came back from
// the command zone attacks the turn it lands. It is in the batch as
// the CONDITIONAL half of the grant builder: the +2/+1 is
// unconditional, the two keywords are not, and the condition is read
// off the HOST rather than off the Equipment.
//
// That distinction is the card. Attach the Heirloom to a 1/1 Soldier
// token and it is a +2/+1 stick; move it to the commander and trample
// and haste switch on with no new action — the layer engine
// recomputes the grant every pass, so the keywords follow the host,
// not the moment of attachment. Equally, a legendary creature that
// something makes non-legendary (a token copy printed "except it
// isn't legendary") loses the keywords while keeping the stats.
//
// "Legendary" is the EFFECTIVE supertype, not the printed one, so an
// effect that confers legendary status counts and one that strips it
// takes the keywords with it.
//
// Equip is sorcery-speed and targets a creature its controller
// controls, which is the shared EquipAbility; the Heirloom adds
// nothing of its own to that.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f707dfb5-dfa9-471b-ac29-5435c30069cc",
		Name:         "Hero's Heirloom",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(2, 1),
			GrantToAttachedWhile(func(host *game.Card, _ *game.Game, _ *game.Card) bool {
				return host.IsLegendary()
			}, "trample", "haste"),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}
