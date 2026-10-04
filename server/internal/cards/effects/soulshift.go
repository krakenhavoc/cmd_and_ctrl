package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soulshift.go — Soulshift N (CR 702.46a): "When this permanent is put
// into a graveyard from the battlefield, you may return target Spirit
// card with mana value N or less from your graveyard to your hand."
// Append-only, mechanic-named, in bushido.go's shape: the keyword is a
// triggered ability, so a card declares it as one.
//
// The "you may" is the CR 603.5 optional prompt (Optional, as
// Auramancer's), the target is chosen as the ability goes on the
// stack, and a graveyard with no Spirit card of mana value N or less puts nothing on
// the stack (CR 603.3d). "Spirit card" is the subtype on the card's
// type line, so a Kindred Spirit card counts. The permanent itself is a
// legal target when it qualifies, since it is in the graveyard when the
// ability triggers. Multiple instances trigger separately (CR 702.46b):
// a card with two prints Soulshift twice.

// Soulshift is "Soulshift n" for the card named `name`.
func Soulshift(n int, name string) game.TriggeredAbility {
	label := fmt.Sprintf("%s — soulshift %d", name, n)
	return Targeting(
		Optional(WhenThisDies(label, returnFirstLegalGraveyardTargetToHand),
			fmt.Sprintf("%s — return a Spirit card with mana value %d or less from your graveyard to your hand?", name, n)),
		TargetCardInGraveyard(
			fmt.Sprintf("target Spirit card with mana value %d or less from your graveyard", n),
			YouOwn(), HasSubtype("Spirit"), ManaValueLE(n)))
}
