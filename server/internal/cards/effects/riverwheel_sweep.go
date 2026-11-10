package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Riverwheel Sweep — Sorcery {2/U}{2/R}{2/W} (Tarkir: Dragonstorm, #4463):
//
//	"Tap target creature. Put three stun counters on it. (If a permanent
//	 with a stun counter would become untapped, remove one from it
//	 instead.)
//	 Exile the top two cards of your library. Choose one of them. Until the
//	 end of your next turn, you may play that card."
//
// The tap and the three stun counters are one statement on the target (a
// target that is illegal as the spell resolves stops the whole spell, so
// nothing is exiled). The exile is a plain one with a continuation (Triple
// Triad's ExileTopKeyedForEffect), and once both cards have landed the
// caster chooses one of them (Exploratory Mission's grant on the chosen
// card alone): a play permission, a land included, that lasts until the end
// of the caster's next turn (ADR 0063's seat-turn window). The other card
// stays in exile with no permission. One card, or none, is not asked.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "917aef2b-0e6d-460e-8745-8da8423d1592",
		Name:         "Riverwheel Sweep",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if id, ok := b16FirstLegalTargetCard(ctx); ok {
				if err := (TapTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
				if err := (AddCounter{Target: id, Kind: game.CounterStun, N: 3}).Apply(ctx); err != nil {
					return err
				}
			} else {
				return nil
			}
			return ctx.Game.ExileTopKeyedForEffect(item.Controller, item.Controller, 2, ctx.ItemTail(riverwheelSweepExiled))
		},
	})
}

// riverwheelSweepExiled is "choose one of them": asked over the cards that
// landed in exile. An on-disk key: never renamed, never reused.
var riverwheelSweepExiled = CardTail("riverwheel-sweep/choose-one", func(ctx *Context, o game.TailOutcome) error {
	var exiled []uuid.UUID
	for _, id := range o.IDs {
		if z := ctx.Game.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneExile {
			exiled = append(exiled, id)
		}
	}
	switch len(exiled) {
	case 0:
		return nil
	case 1:
		return riverwheelSweepGrant(ctx, exiled[0])
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Riverwheel Sweep — choose one of the exiled cards; until the end of your next turn, you may play it",
		Cards:    exiled,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneExile,
		ThenKey:  riverwheelSweepChosen,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
})

var riverwheelSweepChosen = askedPick("choose-cards/riverwheel-sweep-play", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return riverwheelSweepGrant(ctx, r.Picked[0])
})

// riverwheelSweepGrant lets the controller play the chosen card until the
// end of their next turn, if it is still in exile.
func riverwheelSweepGrant(ctx *Context, id uuid.UUID) error {
	z := ctx.Game.FindCardZoneForEffect(id)
	c, ok := ctx.Game.LookupCardForEffect(id)
	if !ok || z == nil || z.Kind != game.ZoneExile {
		return nil
	}
	controller := ctx.Controller()
	window := ctx.Game.UntilEndOfYourNextTurnDuration(controller)
	perm := ExileTopWithPermission{Window: &window}.permission()
	perm.Player = controller
	perm.Zone = game.ZoneExile
	ctx.Game.GrantCastPermissionToCardsForEffect(perm, []game.Card{c})
	return nil
}
