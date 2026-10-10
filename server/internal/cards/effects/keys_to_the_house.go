package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Keys to the House — Artifact {1}:
//
//	"{1}, {T}, Sacrifice this artifact: Search your library for a basic
//	 land card, reveal it, put it into your hand, then shuffle.
//	 {3}, {T}, Sacrifice this artifact: Lock or unlock a door of target
//	 Room you control. Activate only as a sorcery."
//
// The first ability is the Evolving-Wilds-to-hand search (SearchLibrary,
// revealed, shuffled). The second targets a Room you control and
// LockOrUnlockADoor (rooms.go) lets the controller choose either door:
// a locked one is unlocked (CR 709.5f), an unlocked one locked
// (CR 709.5g). The cost is paid at announce, so the Keys are already
// gone when the ability resolves, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "62320290-ac8e-4f92-bb06-3368f66ae0a9",
		Name:         "Keys to the House",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{1}, {T}, Sacrifice this artifact: Search your library for a basic land card, reveal it, put it into your hand, then shuffle.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{1}"), TapCost(), SacrificeThis()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: IsBasicLand,
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Keys to the House — a basic land card",
					}.Apply(NewContext(g, item))
				},
			},
			{
				Label:        "{3}, {T}, Sacrifice this artifact: Lock or unlock a door of target Room you control. Activate only as a sorcery.",
				Cost:         Plus(ManaCost("{3}"), TapCost(), SacrificeThis()),
				Targets:      TargetPermanent("target Room you control", HasSubtype("Room"), YouControl()),
				SorcerySpeed: true,
				Effect:       lockOrUnlockTargetRoom,
			},
		},
	})
}

// lockOrUnlockTargetRoom is "lock or unlock a door of target Room you
// control" — Keys to the House, Marina Vendrell.
func lockOrUnlockTargetRoom(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		return LockOrUnlockADoor{Room: t.ID, Player: item.Controller}.Apply(ctx)
	}
	return nil
}
