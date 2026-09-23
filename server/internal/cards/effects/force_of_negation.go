package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Force of Negation — Instant {1}{U}{U} (EDHREC rank 266):
//
//	"If it's not your turn, you may exile a blue card from your hand
//	 rather than pay this spell's mana cost.
//	 Counter target noncreature spell. If that spell is countered
//	 this way, exile it instead of putting it into its owner's
//	 graveyard."
//
// The Pitch shape (alternative_cost.go) with a Condition instead of a
// life cost — Force of Will's exile-a-blue-card cost, gated on "not
// your turn" through the same isActivePlayer read
// OpponentsCantCastDuringYourTurn uses.
//
// DECLARED SIMPLIFICATION, weaker than printed: the countered spell
// goes to its owner's graveyard, not exile. `CounterTargetForEffect`
// routes every counter to the graveyard; the destination-taking
// counter (`counterSpellLocked` with a `*ZoneRef`) has no `*ForEffect`
// wrapper — the same open seam Venser, Shaper Savant's spell half
// records (docs/engine-seams.md, "Counter-to-hand / counter-to-zone
// effect surface"). The counter itself is not weaker: a noncreature
// spell caught by Force of Negation is still fully countered.
func init() {
	Register(Spec{
		OracleID:     "ac2173f9-f223-440a-9231-fd98762bdc6f",
		Name:         "Force of Negation",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The countered spell goes to its owner's graveyard instead of being exiled."},
		Targets:      TargetSpell("target noncreature spell", Noncreature()),
		AlternativeCosts: []game.AlternativeCost{
			{
				Key:           "pitch",
				Label:         "If it's not your turn, exile a blue card from your hand",
				ExileFromHand: CardInYourHand("a blue card from your hand", OfColor("U")),
				PayLabel:      "a blue card from your hand",
				Condition: func(g *game.Game, controller uuid.UUID) bool {
					return !isActivePlayer(g, controller)
				},
			},
		},
		OnResolve: counterTheTargetSpell,
	})
}
