package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Everquill Phoenix — Creature — Phoenix {2}{R}{R}, 4/4 (Ikoria: Lair of
// Behemoths; reprinted in March of the Machine Commander, #5908):
//
//	"Mutate {3}{R} (If you cast this spell for its mutate cost, put it
//	 over or under target non-Human creature you own. They mutate into
//	 the creature on top plus all abilities from under it.)
//	 Flying
//	 Whenever this creature mutates, create a red artifact token named
//	 Feather with "{1}, Sacrifice this token: Return target Phoenix card
//	 from your graveyard to the battlefield tapped.""
//
// Mutate is the engine's (ADR 0223, #5910): the alternative cost
// targets a non-Human creature with the same owner as the spell, an
// illegal target makes it an ordinary creature spell (CR 702.140b), and
// a legal one merges, over or under as the caster chooses (CR
// 702.140c). The merged creature has every component's abilities (CR
// 702.140e), so flying and the trigger work from underneath too, and
// the trigger fires on the merge whichever card ended on top (CR
// 702.140d). It does not fire when the Phoenix is cast for its mana
// cost, or when the mutate target was illegal: nothing merged.
//
// The Feather is a red artifact token with no creature type, and its
// ability returns a Phoenix CARD from your graveyard, tapped (the
// Phoenix itself, once it has died, among them).
func init() {
	Register(Spec{
		OracleID:         "56ba78c1-d1eb-4cc1-b70d-98c35d2972ba",
		Name:             "Everquill Phoenix",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{Mutate("{3}{R}")},
		Triggered: []game.TriggeredAbility{
			On(game.EventMutate, ThisCreatureMutates,
				"Everquill Phoenix — create a Feather token",
				Do(CreateToken{Template: FeatherToken(), N: 1})),
		},
	})
}
