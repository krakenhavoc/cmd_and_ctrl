package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Enchantment Alteration — Instant {U} (Urza's Saga, S82 OLD batch 729,
// #5624):
//
//	"Attach target Aura attached to a creature or land to another permanent
//	 of that type."
//
// The Aura is the only target (Simic Guildmage's second ability): the new
// host is chosen as the spell resolves, so hexproof on it does not matter.
// "Of that type" is the type of the permanent the Aura is on now: another
// creature for an Aura on a creature, another land for one on a land (a land
// creature counts as either). Only a host the Aura can legally enchant is
// offered (CR 701.3b, game.AttachLegalForEffect); with none, or when the Aura
// has since left or moved on, nothing happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c80935bf-c17f-4940-ad9f-3f4e3e8f71bc",
		Name:         "Enchantment Alteration",
		Completeness: effects.CompletenessFull,
		Targets: effects.BattlefieldSpec("permanent", "target Aura attached to a creature or land",
			func(g *game.Game, _ uuid.UUID, c game.Card) bool {
				if !c.IsAura() || c.AttachedTo.Kind != game.TargetCard {
					return false
				}
				host, ok := g.LookupCardForEffect(c.AttachedTo.ID)
				return ok && effects.OnBattlefield(g, host.InstanceID) && (host.IsCreature() || host.IsLand())
			}),
		OnResolve: enchantmentAlterationResolve,
	})
}

// enchantmentAlterationResolve offers the controller each other permanent of
// the type the Aura's host has now that the Aura can enchant.
func enchantmentAlterationResolve(_ *game.StackItem, ctx *effects.Context) error {
	g := ctx.Game
	auraID := effects.FirstLegalCardTarget(ctx)
	if auraID == uuid.Nil {
		return nil
	}
	aura, ok := g.LookupCardForEffect(auraID)
	if !ok || aura.AttachedTo.Kind != game.TargetCard {
		return nil
	}
	current, ok := g.LookupCardForEffect(aura.AttachedTo.ID)
	if !ok || !effects.OnBattlefield(g, current.InstanceID) {
		return nil
	}
	var hosts []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID == current.InstanceID {
			continue
		}
		sameType := (current.IsCreature() && c.IsCreature()) || (current.IsLand() && c.IsLand())
		if sameType && g.AttachLegalForEffect(auraID, c.InstanceID) {
			hosts = append(hosts, c.InstanceID)
		}
	}
	if len(hosts) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Enchantment Alteration — attach " + aura.Name + " to another permanent of that type",
		Cards:    hosts,
		Min:      1,
		Max:      1,
		ThenKey:  enchantmentAlterationPick,
		LiveZone: game.ZoneBattlefield,
		Carry:    []uuid.UUID{auraID},
	})
	return nil
}

// enchantmentAlterationPick attaches the Aura (the one carried ID) to the
// permanent picked.
var enchantmentAlterationPick = game.RegisterChooseCardsThen("choose-cards/enchantment-alteration-attach", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 || len(r.Carry) == 0 {
		return nil
	}
	return g.AttachForEffect(r.Carry[0], game.TargetRef{Kind: game.TargetCard, ID: r.Picked[0]})
})
