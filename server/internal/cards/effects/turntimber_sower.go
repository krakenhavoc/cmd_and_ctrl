package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Turntimber Sower:
//
//	"Whenever one or more land cards are put into your graveyard from
//	 anywhere, create a 0/1 green Plant creature token.
//	 {G}, Sacrifice three creatures: Return target land card from your
//	 graveyard to your hand."
//
// The trigger is a "one or more" batch (CR 603.2c). The return is a
// target in your graveyard, re-judged as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "23d7d8ba-5237-4ffc-bd5f-b07ff9ec73af",
		Name:         "Turntimber Sower",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			cmdr2B13OneOrMoreLandCardsToYourGraveyard("Turntimber Sower — create a 0/1 green Plant creature token",
				Do(CreateToken{Template: TokenCard("0/1 green Plant"), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{G}, Sacrifice three creatures: Return target land card from your graveyard to your hand.",
			Cost:    Plus(ManaCost("{G}"), SacrificeN(3, "three creatures", Creature())),
			Targets: TargetCardInGraveyard("target land card in your graveyard", YouOwn(), Land()),
			Effect: Do(OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				return ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}.Apply(ctx)
			})),
		}},
	})
}
