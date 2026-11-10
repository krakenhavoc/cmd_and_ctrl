package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Transforming Flourish — Instant {2}{R} (Tarkir: Dragonstorm Commander):
//
//	"Demonstrate (When you cast this spell, you may copy it. If you do, choose
//	 an opponent to also copy it. Players may choose new targets for their
//	 copies.)
//	 Destroy target artifact or creature you don't control. If that permanent
//	 is destroyed this way, its controller exiles cards from the top of their
//	 library until they exile a nonland card, then they may cast that card
//	 without paying its mana cost."
//
// Demonstrate is the keyword's cast trigger (Replication Technique's). The
// destruction is the mass verb narrowed to the one target, so "destroyed this
// way" is exactly what the engine says went (a regenerated, indestructible or
// replaced one is not, and a commander sent to the command zone was), and the
// continuation is told the permanent as it last existed, with the controller it
// had (CR 603.10, 608.2h). That player, not the spell's controller, exiles from
// their own library until a nonland card (the other exiled cards stay exiled)
// and is offered the free cast during the resolution (CR 608.2g): {0} with X as
// 0, no alternative cost, and a decline leaves the card in exile.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9fb7e7ca-ad2f-4677-bb4d-50808b01b4f9",
		Name:         "Transforming Flourish",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Demonstrate()},
		Targets:      TargetPermanent("target artifact or creature you don't control", Or(Artifact(), Creature()), OpponentControls()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				id := t.ID
				return DestroyAllMatching{
					Match:    func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.InstanceID == id },
					ThenTail: ctx.ItemTail(transformingFlourishDestroyed),
				}.Apply(ctx)
			}
			return nil
		},
	})
}

// transformingFlourishDestroyed is "its controller exiles cards from the top
// of their library until they exile a nonland card, then they may cast that
// card". An on-disk key.
var transformingFlourishDestroyed = MassTail("mass/transforming-flourish-destroyed", func(ctx *Context, swept []game.Card, _ int) error {
	if len(swept) == 0 {
		return nil
	}
	player := swept[0].Controller
	if player == uuid.Nil {
		player = swept[0].Owner
	}
	return MillToZone{
		Player:   player,
		To:       game.ZoneExile,
		Until:    UntilNonland,
		ThenTail: withCarry(ctx.ItemTail(transformingFlourishCast), player),
	}.Apply(ctx)
})

// transformingFlourishCast offers the exiled nonland card to the player whose
// library it came from. An on-disk key.
var transformingFlourishCast = CardTail("then/transforming-flourish-cast", func(ctx *Context, o game.TailOutcome) error {
	player := tailCarry0(o)
	var hit []uuid.UUID
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && !c.IsLand() {
			hit = append(hit, id)
		}
	}
	if player == uuid.Nil || len(hit) == 0 {
		return nil
	}
	grantFreeCasts(ctx.Game, player, ctx.Source(), "Transforming Flourish", game.TimingFlash, game.LapseStaysInExile, hit)
	return nil
})
