package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aberrant Mind Sorcerer — Creature — Human Elf Shaman Sorcerer {4}{U},
// 3/4 (Adventures in the Forgotten Realms, #5499):
//
//	"Psionic Spells — When this creature enters, choose target instant or
//	 sorcery card in your graveyard, then roll a d20.
//	 1—9 | You may put that card on top of your library.
//	 10—20 | Return that card to your hand."
//
// A card that has left the graveyard by the time the roll settles is not
// moved (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "753ead8c-f3f8-42fe-931d-79f5ee3ece54",
		Name:         "Aberrant Mind Sorcerer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Psionic Spells — choose target instant or sorcery card in your graveyard, then roll a d20",
				Do(RollDice{Sides: 20, N: 1, Then: aberrantMindSorcererTable})),
				TargetCardInGraveyard("target instant or sorcery card in your graveyard", Or(Instant(), Sorcery()), YouOwn())),
		},
	})
}

// aberrantMindSorcererTable is the results table, after the d20.
var aberrantMindSorcererTable = DiceTail("dice/aberrant-mind-sorcerer-table", func(ctx *Context, rolls []int) error {
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if rolls[0] >= 10 {
		return ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}.Apply(ctx)
	}
	return MayChoice{
		Question: "Aberrant Mind Sorcerer — put that card on top of your library?",
		Then:     aberrantMindSorcererTop,
		Carry:    []uuid.UUID{id},
	}.Apply(ctx)
})

// aberrantMindSorcererTop is the "you may" answered: a yes puts the card
// on top of its owner's library if it is still in the graveyard.
var aberrantMindSorcererTop = MayYes("confirm/aberrant-mind-sorcerer-top", func(ctx *Context, r game.ConfirmAnswer) error {
	if len(r.Carry) == 0 {
		return nil
	}
	if z := ctx.Game.FindCardZoneForEffect(r.Carry[0]); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return ctx.Game.TuckToLibraryForEffect(r.Carry[0], false)
})
