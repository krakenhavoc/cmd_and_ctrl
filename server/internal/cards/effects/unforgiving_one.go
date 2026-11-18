package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Unforgiving One — Creature — Spirit {2}{B}, 2/3 (Kamigawa: Neon Dynasty):
//
//	"Menace
//	 Whenever this creature attacks, return target creature card with mana
//	 value X or less from your graveyard to the battlefield, where X is the
//	 number of modified creatures you control. (Equipment, Auras you control,
//	 and counters are modifications.)"
//
// Alesha's attack-time reanimation with the bound read off the board: the
// candidate's mana value is judged against the number of modified creatures
// the caster controls (CR 700.9, b07IsModified) as the target is chosen and
// again as the ability resolves (CR 608.2b), so a creature that stops being
// modified in response can leave the pick illegal. The card returns under
// its owner's control, yours.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9ec2114c-023b-45c3-a44c-6a9f6462bc85",
		Name:            "Unforgiving One",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacks("Unforgiving One — return target creature card with mana value X or less from your graveyard to the battlefield, where X is the number of modified creatures you control",
				returnFirstLegalGraveyardTargetToBattlefield),
				TargetCardInGraveyard("target creature card with mana value X or less from your graveyard, where X is the number of modified creatures you control",
					Creature(), YouOwn(), unforgivingOneManaValueAtMostModified)),
		},
	})
}

// unforgivingOneManaValueAtMostModified is "mana value X or less", X being
// the number of modified creatures the caster controls.
func unforgivingOneManaValueAtMostModified(g *game.Game, caster uuid.UUID, c game.Card) bool {
	mv, ok := g.ManaValueForEffect(c)
	return ok && mv <= s82NeoB11ModifiedCreaturesYouControl(g, caster)
}
