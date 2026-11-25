package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Swimmer in Nightmares — Creature — Nightmare Merfolk {2}{U}, 1/4 (Theros
// Beyond Death, S82 OLD, #5514):
//
//	"This creature gets +3/+0 as long as there are ten or more cards in a
//	 single graveyard.
//	 This creature can't be blocked as long as you control an Ashiok
//	 planeswalker."
//
// Mindwhisker's threshold static over any one graveyard: the +3/+0 is a
// layer 7 modification read live, so it applies the moment any player's
// graveyard reaches ten cards and ends when it falls below. The evasion is a
// block rule judged as blockers are declared (CR 509.1b), Waterlogged Hulk's
// shape: the attacker can't be blocked while its controller controls a
// planeswalker with the subtype Ashiok (a word of the card's text, so a
// changed type works the same).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "64ca54ee-1502-4fae-b919-7e4457e641c0",
		Name:         "Swimmer in Nightmares",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "This creature gets +3/+0 as long as there are ten or more cards in a single graveyard.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && effects.SomeGraveyardHasAtLeast(g, 10)
			},
			Apply: effects.PumpBy(3, 0),
		}},
		BlockRules: []game.BlockRule{effects.CantBeBlockedWhile(effects.OnSelf(),
			func(g *game.Game, controller uuid.UUID, _ game.Card) bool {
				return len(effects.PermanentsControlledByMatching(g, controller,
					effects.And(effects.Planeswalker(), effects.Subtype("Ashiok")))) > 0
			})},
	})
}
