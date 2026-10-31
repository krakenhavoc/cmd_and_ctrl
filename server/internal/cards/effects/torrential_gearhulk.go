package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Torrential Gearhulk — Artifact Creature — Construct {4}{U}{U}, 5/6:
//
//	"Flash
//	 When this creature enters, you may cast target instant card from
//	 your graveyard without paying its mana cost. If that spell would
//	 be put into your graveyard, exile it instead."
//
// The cast is a grant, the way every "you may cast it" a resolution
// offers is (ADR 0099 §4): a free ({0}, so X is locked at 0 and
// additional costs are still paid), any-timing (CR 608.2g), cast-only,
// permission over the target card in the graveyard, offered during the
// resolution (#4250). The holder's next priority pass closes it only as the
// fallback, and the card simply stays where it is. "If that spell would be put into your graveyard, exile it
// instead" is the permission's ExileInsteadOfGraveyard (#3863): the
// spell is exiled whether it resolves, is countered or fizzles, and a
// spell returned to its owner's hand or put into a library goes there,
// as printed (flashback's wider "anywhere else" would exile it). The
// target is chosen as the trigger goes on the stack and re-checked as
// it resolves (CR 608.2b).
//
// No simplification.
func init() {
	enters := WhenThisEnters("Torrential Gearhulk — you may cast target instant card from your graveyard without paying its mana cost",
		freeCastTargetFromYourGraveyardExilingIt("Torrential Gearhulk"))
	enters.Targets = TargetCardInGraveyard("target instant card from your graveyard", Instant(), YouOwn())
	Register(Spec{
		OracleID:        "a6395447-677d-4c39-8eda-2d57e527c94e",
		Name:            "Torrential Gearhulk",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered:       []game.TriggeredAbility{enters},
	})
}
