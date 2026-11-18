package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// afrB04InterrogateExiled is "choose one of them" over the cards that
// reached exile from the opponent's library (You Find Some Prisoners). An
// on-disk key: never renamed, never reused.
var afrB04InterrogateExiled = CardTail("then/s82-afr-b04-interrogate-choose-one", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "You Find Some Prisoners — choose one of them (you may play it until the end of your next turn, spending mana as though it were mana of any color)",
		Cards:    o.IDs,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneExile,
		ThenKey:  afrB04InterrogateGrant,
	})
	return nil
})

// afrB04InterrogateGrant gives the chosen card its window. An on-disk key.
var afrB04InterrogateGrant = game.RegisterChooseCardsThen("choose-cards/s82-afr-b04-interrogate-play", func(g *game.Game, r game.CardsChosen) error {
	for _, id := range r.Picked {
		g.GrantCastPermissionOverCardForEffect(id, game.CastPermission{
			Player:   r.Chooser,
			Zone:     game.ZoneExile,
			Duration: g.UntilEndOfYourNextTurnDuration(r.Chooser),
			AnyColor: true,
			Label:    "Play it until the end of your next turn; spend mana as though it were mana of any color",
		})
	}
	return nil
})

// afrB04Interrogate is "exile the top three cards of target opponent's
// library" as a modal bullet's body: the opponent is the bullet's own
// target, still legal as the spell resolves (CR 608.2b).
func afrB04Interrogate(item *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	return ctx.Game.ExileTopKeyedForEffect(t.ID, item.Controller, 3, ctx.ItemTail(afrB04InterrogateExiled))
}
