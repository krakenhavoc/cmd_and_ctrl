package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Helm of the Host — Legendary Artifact — Equipment {4}:
//
//	"At the beginning of combat on your turn, create a token that's a
//	 copy of equipped creature, except the token isn't legendary. That
//	 token gains haste.
//	 Equip {5}"
//
// The token is minted through CreateTokenCopy, which carries the
// copied creature's oracle ID onto the new object — its own ETB
// triggers, statics and activated abilities all come along, exactly
// as printed. "Except the token isn't legendary" and "gains haste"
// are both TEMPLATE edits (Except): the haste half reuses
// reflectionOfKikiJikiHasteException verbatim (Reflection of
// Kiki-Jiki's own "except it has haste"), and the legendary half is
// removeLegendaryFromTypeLine, legendaryTypeLine's mirror image.
//
// With nothing equipped the trigger fires and does nothing — there is
// no host to copy, and a resolution that finds nothing to do is not
// an error (the same posture CreateTokenCopy already takes for a
// `Copy` that can no longer be found).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "83b43aba-bf9c-4da2-967d-9daa632e97d2",
		Name:         "Helm of the Host",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat(
				"Helm of the Host — create a token that's a copy of equipped creature, except it isn't legendary and has haste",
				helmOfTheHostCopyEquippedCreature),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{5}"),
		},
	})
}

// helmOfTheHostCopyEquippedCreature reads the current host off the
// Helm itself (not captured from Build — the equipped creature can
// change between turns) and mints the nonlegendary, hasty copy.
func helmOfTheHostCopyEquippedCreature(g *game.Game, item *game.StackItem) error {
	host := attachedHostFor(g, item.SourceCardID)
	if host == nil {
		return nil
	}
	return CreateTokenCopy{
		Controller: item.Controller,
		Copy:       host.InstanceID,
		N:          1,
		Except: func(t *game.Card) {
			t.TypeLine = removeLegendaryFromTypeLine(t.TypeLine)
			reflectionOfKikiJikiHasteException(t)
		},
	}.Apply(NewContext(g, item))
}
