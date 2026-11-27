package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// untilFourLandsRef is "until they reveal four land cards" (Mirko Vosk, Mind
// Drinker): asked with every card that has landed so far, it stops once four
// of them are lands. An on-disk key.
var untilFourLandsRef = game.RegisterMillUntil("until/four-land-cards", func(landed []game.Card, _ int) bool {
	lands := 0
	for _, c := range landed {
		if c.IsLand() {
			lands++
		}
	}
	return lands >= 4
})

// S82OldB405MillUntilFourLands is "that player reveals cards from the top of
// their library until they reveal four land cards, then puts those cards into
// their graveyard": the run is measured read-only and announced as one reveal
// (Consuming Aberration's shape), then milled, the fourth land included. A
// library with fewer than four lands reveals and mills all of it.
func S82OldB405MillUntilFourLands(ctx *Context, player uuid.UUID) error {
	p := ctx.PlayerByID(player)
	if p == nil || p.Library == nil {
		return nil
	}
	run := make([]uuid.UUID, 0, 8)
	lands := 0
	for i := len(p.Library.Cards) - 1; i >= 0; i-- {
		c := p.Library.Cards[i]
		run = append(run, c.InstanceID)
		if c.IsLand() {
			lands++
			if lands == 4 {
				break
			}
		}
	}
	if err := (RevealCards{Player: player, Cards: run, Reason: "reveal from the top until four land cards"}).Apply(ctx); err != nil {
		return err
	}
	return MillToZone{Player: player, Until: game.MillUntil{Ref: untilFourLandsRef}}.Apply(ctx)
}

// S82OldB405EnchantedAttacksOrBlocksLoseLife is "Whenever enchanted creature
// attacks or blocks, its controller loses N life" (Sinister Possession; the
// sibling Contaminated Bond is the same shape for 3). The enchanted creature
// is the event's creature: an attack event names it, and a block event names
// the blocker, once per block declaration (Amount <= 1, so a creature that
// blocks two attackers triggers once). Its controller is read as it is now,
// or as it last existed if it has left (CR 608.2h).
func S82OldB405EnchantedAttacksOrBlocksLoseLife(label string, n int) game.TriggeredAbility {
	return OnAny([]game.EventKind{game.EventAttack, game.EventBlock}, OldB484EnchantedAttacksOrBlocks, label,
		func(g *game.Game, item *game.StackItem) error {
			ev := item.Trigger.Event
			loser := ev.Actor
			if c, ok := g.LookupCardForEffect(ev.CardID); ok && OnBattlefield(g, ev.CardID) {
				loser = c.Controller
			} else if info, ok := g.LastKnownPermanentForEffect(ev.CardID); ok {
				loser = info.Controller
			}
			return g.ChangePlayerLifeForEffect(item.SourceCardID, loser, -n)
		})
}
