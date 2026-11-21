package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Fugitive Doctor — Legendary Creature — Time Lord Doctor {3}{R}{G}, 4/4
// (Doctor Who):
//
//	"When The Fugitive Doctor enters, investigate.
//	 Whenever The Fugitive Doctor attacks, you may sacrifice a Clue. When you do,
//	 target instant or sorcery card in your graveyard gains flashback {2}{R}{G}
//	 until end of turn. (You may cast that card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// Ace's shape: the attack trigger asks "you may sacrifice a Clue" as it resolves
// and a Clue that really went raises the printed "when you do" as a CR 603.12
// reflexive trigger, which chooses its own target as it goes on the stack. The
// grant is Archmage's Newt's flashback permission with an explicit cost, lasting
// until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b3a03768-7676-4b33-a813-cc69dea77c4d",
		Name:         "The Fugitive Doctor",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("The Fugitive Doctor — investigate", Do(Investigate{N: 1})),
			WheneverThisAttacks("The Fugitive Doctor — you may sacrifice a Clue", func(g *game.Game, item *game.StackItem) error {
				return maySacrificeOne(g, item, "The Fugitive Doctor — sacrifice a Clue to give an instant or sorcery card in your graveyard flashback {2}{R}{G}?",
					OfSubtype("Clue"), false, fugitiveDoctorSacrificed)
			}),
		},
	})
}

var fugitiveDoctorSacrificed = RegisterSacrificedOne("prompt-run/the-fugitive-doctor-sacrifice", func(ctx *Context, _ uuid.UUID) error {
	return ReflexiveTrigger{
		Label: "The Fugitive Doctor — target instant or sorcery card in your graveyard gains flashback {2}{R}{G} until end of turn",
		Body:  fugitiveDoctorFlashbackBody,
	}.Apply(ctx)
})

var fugitiveDoctorFlashbackBody = game.ReflexiveBody("the-fugitive-doctor/flashback",
	simpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := b16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return GrantFlashbackToCard{Target: id, Cost: "{2}{R}{G}", Label: "Flashback {2}{R}{G} (The Fugitive Doctor)"}.Apply(ctx)
	}),
	func(_ uuid.UUID, _ game.EffectParams) *game.TargetSpec {
		return TargetCardInGraveyard("target instant or sorcery card in your graveyard", Or(Instant(), Sorcery()), YouOwn())
	})
