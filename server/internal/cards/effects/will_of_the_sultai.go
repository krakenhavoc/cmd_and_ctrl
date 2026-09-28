package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Will of the Sultai — Sorcery {4}{G}:
//
//	"Choose one. If you control a commander as you cast this spell,
//	 you may choose both instead.
//	 • Target player mills three cards. Return all land cards from
//	   your graveyard to the battlefield tapped.
//	 • Put X +1/+1 counters on target creature, where X is the number
//	   of lands you control. It gains trample until end of turn."
//
// The commander rider is the conditional mode count #1590 built:
// `OrUpToIf(2, YouControlACommander)`, read at announce (CR 601.2b) and
// fixed from then on.
//
// Printed order is the card (CR 608.2c): the lands come back BEFORE
// the second bullet counts "the number of lands you control", so with
// both chosen the counters include every land the first bullet
// returned — milled ones among them when the target player is you. The
// mill is the sequencing form (MillToZone.Then), so a milled commander
// pausing the run on CR 903.9's offer holds the return and the counters
// until it is answered, rather than letting them run a card early.
//
// The land return is not about the target: if the targeted player has
// become illegal the mill is skipped and the lands still come back (CR
// 608.2b — the spell does as much as it can). Each land returns under
// its owner's control, the caster's, through the CR 614 entry pipeline
// with EntersTapped, so it enters tapped rather than being tapped a
// beat later.
func init() {
	Register(Spec{
		OracleID:     "951a0e34-e610-477f-b441-6680df55a29b",
		Name:         "Will of the Sultai",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Target player mills three cards. Return all land cards from your graveyard to the battlefield tapped.",
				TargetPlayer("target player")),
			Mode("Put X +1/+1 counters on target creature, where X is the number of lands you control. It gains trample until end of turn.",
				TargetCreature("target creature")),
		).OrUpToIf(2, YouControlACommander),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if !ctx.HasMode(0) {
				return willOfTheSultaiCounters(ctx)
			}
			targets := OptionTargets(ctx, 0)
			if len(targets) == 0 {
				return willOfTheSultaiLandsThenCounters(ctx, nil)
			}
			return MillToZone{
				Player: targets[0].ID,
				N:      3,
				To:     game.ZoneGraveyard,
				Then:   willOfTheSultaiLandsThenCounters,
			}.Apply(ctx)
		},
	})
}

// willOfTheSultaiLandsThenCounters is the rest of the first bullet and
// then, if it was chosen, the second: the order the card prints them.
func willOfTheSultaiLandsThenCounters(ctx *Context, _ []uuid.UUID) error {
	if p := ctx.PlayerByID(ctx.Controller()); p != nil && p.Graveyard != nil {
		var lands []uuid.UUID
		for _, c := range p.Graveyard.Cards {
			if c.IsLand() {
				lands = append(lands, c.InstanceID)
			}
		}
		for _, id := range lands {
			if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield, Tapped: true}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	return willOfTheSultaiCounters(ctx)
}

// willOfTheSultaiCounters is the second bullet: X +1/+1 counters, X
// counted as it resolves, and trample until end of turn.
func willOfTheSultaiCounters(ctx *Context) error {
	if !ctx.HasMode(1) {
		return nil
	}
	for _, t := range OptionTargets(ctx, 1) {
		if n := b03LandsControlled(ctx.Game, ctx.Controller()); n > 0 {
			if err := (AddCounter{Target: t.ID, Kind: "+1/+1", N: n}).Apply(ctx); err != nil {
				return err
			}
		}
		if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"trample"}, Label: "Will of the Sultai — trample"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
