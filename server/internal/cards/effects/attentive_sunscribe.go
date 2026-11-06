package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Attentive Sunscribe — Artifact Creature — Gnome {1}{W}, 2/2 (The Lost
// Caverns of Ixalan, #4470):
//
//	"Whenever this creature becomes tapped, scry 1."
//
// Compassionate Healer's trigger without the life: "becomes tapped" is
// the tap event and the attack that taps it (ThisBecameTapped, #4534),
// so a vigilance attack is not one, and any other tap counts, such as
// crewing or paying a cost (CR 701.26a). The scry is the footprinted
// primitive (CR 701.22a); a rider is declared on the card's own Purpose,
// as Compassionate Healer does, since a triggered row's rider is not read
// (#4978).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0f85714b-e731-45a3-aa65-8cd59cd48280",
		Name:         "Attentive Sunscribe",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped,
				"Attentive Sunscribe — scry 1",
				Do(Scry{N: 1})),
		},
	})
}
