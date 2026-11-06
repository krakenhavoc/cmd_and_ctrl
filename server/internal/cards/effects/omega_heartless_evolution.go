package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Omega, Heartless Evolution — Legendary Artifact Creature — Robot
// {5}{G}{U}, 8/8 (Final Fantasy, #4462):
//
//	"Wave Cannon — When Omega enters, for each opponent, tap up to one
//	 target nonland permanent that opponent controls. Put X stun counters
//	 on each of those permanents and you gain X life, where X is the
//	 number of nonbasic lands you control. (If a permanent with a stun
//	 counter would become untapped, remove one from it instead.)"
//
// The enters trigger is King Solomon's Frogs' per-opponent clause list
// (perOpponentClauses: one "up to one" clause per opponent seat, each
// pick bound to that player, so the CR 608.2b re-check asks whether THAT
// opponent still controls it). X is counted as the trigger resolves
// (b26NonbasicLandsControlled: a land without the Basic supertype). Each
// still-legal pick is tapped, then every one of them gets X stun
// counters, then you gain X life; with no pick chosen the life is still
// gained. The stun counter's untap replacement is the engine's.
//
// No simplification.
func init() {
	enters := WhenThisEnters("Wave Cannon — for each opponent, tap up to one target nonland permanent; put X stun counters on each and gain X life",
		omegaWaveCannon)
	enters.Targets = perOpponentClauses(game.MaxPlayers-1, "nonland permanent", Nonland())
	Register(Spec{
		OracleID:     "3173ca98-d46a-4de3-a95e-3eaad28d4017",
		Name:         "Omega, Heartless Evolution",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
	})
}

// omegaWaveCannon is the enters trigger's resolution.
func omegaWaveCannon(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var picks []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			picks = append(picks, t.ID)
		}
	}
	for _, id := range picks {
		if err := (TapTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	x := b26NonbasicLandsControlled(g, item.Controller)
	if x <= 0 {
		return nil
	}
	for _, id := range picks {
		if !onBattlefield(g, id) {
			continue
		}
		if err := (AddCounter{Target: id, Kind: game.CounterStun, N: x}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return GainLife{Amount: x}.Apply(ctx)
}
