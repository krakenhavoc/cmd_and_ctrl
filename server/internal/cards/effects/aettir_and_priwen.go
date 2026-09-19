package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aettir and Priwen — Legendary Artifact — Equipment for {6}:
//
//	"Equipped creature has base power and toughness X/X, where X is
//	 your life total.
//	 Equip {5}"
//
// A lifegain deck's win condition attached to anything: the smallest
// creature on the board becomes as big as your life total, and stays
// that big as the total moves.
//
// TWO THINGS DECIDE WHETHER THIS CARD IS RIGHT, and both are in the
// helper rather than here:
//
//   - IT IS LAYER 7b, a SET, not 7c's modify. 7b runs first, so
//     +1/+1 counters and anthems apply ON TOP of the life total
//     rather than being erased by it — a 40-life player's creature
//     with two +1/+1 counters is a 42/42. Setting the base in 7c
//     would instead fight every other pump on the board, and a
//     counter placed before the Equipment arrived would silently
//     disappear.
//   - THE VALUE IS RE-READ EVERY RECOMPUTE, not captured at equip
//     time. Gain three life and the creature grows in the same beat;
//     take a hit and it shrinks, which is the real drawback — a
//     creature carrying Aettir at 4 life is a 4/4 and dies to things
//     it used to beat.
//
// "Your" is the EQUIPMENT's controller (CR 109.5), so a stolen
// carrier is still sized by whoever holds Aettir and Priwen.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "19d8937e-23fb-4758-a306-0e7af9bb44c4",
		Name:         "Aettir and Priwen",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			aettirBasePTFromLifeTotal(),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{5}"),
		},
	})
}

// aettirBasePTFromLifeTotal is the card's one static, with the
// invalidation hint the value needs.
//
// A life total is not on the battlefield, so no zone move, no
// counter and no tap tells the layer cache it has gone stale. The
// engine reads game.StaticAbility.DependsOnLifeTotal to know that a
// life change has to drop the cached resolution — without it the
// creature would keep whatever size it had when something unrelated
// last happened, which is a card that looks right in a screenshot
// and is wrong in a combat.
func aettirBasePTFromLifeTotal() game.StaticAbility {
	ab := SetAttachedBasePTPer(lifeTotalOfSourceController)
	ab.DependsOnLifeTotal = true
	return ab
}

// lifeTotalOfSourceController is "X/X, where X is your life total"
// as a SetAttachedBasePTPer value: the life total of the
// ATTACHMENT's controller, for both halves of the P/T.
//
// A seat that has somehow left the game reads as 0/0, which the
// state-based actions then resolve — that is the right answer for a
// creature whose sizing effect has no player behind it, and it is
// never reachable while the Equipment is on the battlefield under a
// live controller.
func lifeTotalOfSourceController(g *game.Game, source *game.Card) (int, int) {
	p := g.PlayerByIDForEffect(source.Controller)
	if p == nil {
		return 0, 0
	}
	return p.Life, p.Life
}
