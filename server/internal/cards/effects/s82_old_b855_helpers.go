package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b855_helpers.go — Fallen Empires, Auras, other creatures, card draw,
// damage, instants and lands (S82 OLD b855, #5637).

// S82OldB855RegenerateEnchantedCreature is "Regenerate enchanted creature" in an
// Aura's own ability whose cost sacrifices the Aura (Thrull Retainer): the
// creature it was attached to as it left, if it is still on the battlefield
// (CR 608.2h). AttachedHostFor would find nothing, the Aura having gone.
func S82OldB855RegenerateEnchantedCreature(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	host, ok := EnchantedCreatureAsItResolves(ctx)
	if !ok || !OnBattlefield(g, host) {
		return nil
	}
	return Regenerate{Target: host}.Apply(ctx)
}

// S82OldB855RainbowValeRider is the rest of Rainbow Vale's mana ability: "An
// opponent gains control of this land at the beginning of the next end step."
// It runs as the mana is added and creates the delayed triggered ability (CR
// 603.7), which names the land as the object it is now, so a land that has left
// and come back is not given away (CR 400.7).
func S82OldB855RainbowValeRider(g *game.Game, controller, source uuid.UUID) error {
	g.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
		Controller:   controller,
		SourceCardID: source,
		Label:        "Rainbow Vale — an opponent gains control of this land",
		At:           game.StepEnd,
		Cards:        []uuid.UUID{source},
		Body:         s82OldB855RainbowValeBody,
	})
	return nil
}

var s82OldB855RainbowValeBody = game.SimpleDelayedBody("rainbow-vale/an-opponent-gains-control", func(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var land uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			land = t.ID
			break
		}
	}
	if land == uuid.Nil || !OnBattlefield(g, land) {
		return nil
	}
	var opponents []uuid.UUID
	for _, p := range g.Seats {
		if p != nil && !p.Eliminated && p.ID != item.Controller {
			opponents = append(opponents, p.ID)
		}
	}
	switch len(opponents) {
	case 0:
		return nil
	case 1:
		return s82OldB855GiveLand(g, land, opponents[0])
	}
	g.QueueChoosePlayerForEffect(game.ChoosePlayerPrompt{
		Chooser:  item.Controller,
		Among:    opponents,
		Source:   land,
		Question: "Rainbow Vale — choose an opponent to gain control of it",
		ThenKey:  s82OldB855RainbowValeChosen,
		Asked:    game.NewAskedItem(item),
	})
	return nil
})

var s82OldB855RainbowValeChosen = OptionPickThen("option-pick/rainbow-vale-opponent", func(ctx *Context, r game.OptionPicked) error {
	if r.Option == nil || r.Option.Player == uuid.Nil || !OnBattlefield(ctx.Game, r.Source) {
		return nil
	}
	return s82OldB855GiveLand(ctx.Game, r.Source, r.Option.Player)
})

func s82OldB855GiveLand(g *game.Game, land, to uuid.UUID) error {
	g.GainControlForEffect(land, land, to, game.IndefiniteDuration(), "Rainbow Vale — an opponent gains control of it")
	return nil
}
