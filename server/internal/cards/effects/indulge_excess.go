package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Indulge // Excess — Sorcery // Sorcery — Aftermath {2}{R} // {1}{R}
// (Streets of New Capenna Commander, #5491):
//
//	Indulge: "Whenever a creature you control attacks this turn, create a
//	          1/1 green and white Citizen creature token that's tapped and
//	          attacking."
//	Excess:  "Aftermath (Cast this spell only from your graveyard. Then
//	          exile it.)
//	          Create a Treasure token for each creature you controlled that
//	          dealt combat damage to a player this turn."
//
// Both halves register as ADR 0034 faces (ADR 0103): Indulge under the bare
// oracle ID, Excess under "#1"; aftermath is the engine's (CR 702.127a).
//
// Indulge sets up a repeating delayed trigger (CR 603.7b) that lasts the
// turn: it fires once for each creature its controller attacks with, in any
// combat this turn (CR 603.2c), and each Citizen enters tapped and attacking
// the player or planeswalker its controller chooses (CR 508.4). The tokens
// are not declared as attackers, so they trigger nothing. Cast after combat
// it does nothing unless another combat follows.
//
// Excess counts distinct creatures: as the spell resolves, each creature
// its controller controlled as it dealt combat damage to a player this turn,
// however many times or players, and whether or not it is still there.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "abcb5b99-bf5e-402f-97a2-abb1ab39dd72",
		Name:         "Indulge",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DelayedOnEvent{
				Label:     "Indulge — create a 1/1 green and white Citizen creature token that's tapped and attacking",
				On:        []game.EventKind{game.EventAttack},
				Condition: sncIndulgeAttackCondition,
				Body:      sncIndulgeCitizenBody,
				Repeats:   true,
			}.Apply(ctx)
		},
	})
	Register(Spec{
		OracleID:     "abcb5b99-bf5e-402f-97a2-abb1ab39dd72#1",
		Name:         "Excess",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			n := sncCreaturesYouControlledThatDealtCombatDamageToAPlayer(ctx.Game, ctx.Controller())
			if n <= 0 {
				return nil
			}
			return CreateToken{Template: TreasureToken(), N: n}.Apply(ctx)
		},
	})
}
