package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Calix, Guided by Fate — Legendary Enchantment Creature — Human Druid
// {1}{G}{W}, 2/2 (March of the Machine: The Aftermath, #5476):
//
//	"Constellation — Whenever Calix or another enchantment you control
//	 enters, put a +1/+1 counter on target creature.
//	 Whenever Calix or an enchanted creature you control deals combat damage
//	 to a player, you may create a token that's a copy of a nonlegendary
//	 enchantment you control. Do this only once each turn."
//
// Constellation is Doomwake Giant's condition with a target chosen as the
// trigger goes on the stack (CR 603.3d) and re-checked as it resolves
// (CR 608.2b). The second ability is one trigger per creature that
// connects (Ellivere's "enchanted creature", any Aura attached). The "you
// may" is asked as it resolves (CR 603.5); the enchantment is CHOSEN, not
// targeted, so hexproof does not stop it. "Only once each turn" limits the
// DOING (Baron Strucker's tally): declining, or having no nonlegendary
// enchantment, does not use the turn up, and a second trigger from another
// connecting creature finds it spent. Calix is legendary, so never his own
// copy; an enchantment creature (a legendary one excepted) is a legal copy.
//
// No simplification.
const (
	calixLabel = "Calix, Guided by Fate — you may create a token that's a copy of a nonlegendary enchantment you control"
	calixDone  = "Calix, Guided by Fate — copied an enchantment this turn"
)

var calixCopy = effects.RegisterPermanentsPick("prompt-run/calix-guided-by-fate-copy", effects.PermanentsDef{
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		ids := picked.Cards()
		if len(ids) == 0 || effects.B28DoneThisTurn(ctx.Game, ctx.Source(), calixDone) {
			return nil
		}
		effects.B28NoteDoneThisTurn(ctx.Game, ctx.Source(), calixDone)
		return effects.CreateTokenCopy{Controller: ctx.Controller(), Copy: ids[0], N: 1}.Apply(ctx)
	},
})

func calixCopyEffect(g *game.Game, item *game.StackItem) error {
	if effects.B28DoneThisTurn(g, item.SourceCardID, calixDone) {
		return nil
	}
	return effects.ChoosePermanents{
		Question: "Calix, Guided by Fate — choose a nonlegendary enchantment you control; create a token copy of it",
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			var out []uuid.UUID
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == of && c.IsEnchantment() && !c.IsLegendary() {
					out = append(out, c.InstanceID)
				}
			}
			return out, 1, 1
		},
		Keyed: calixCopy,
	}.Apply(effects.NewContext(g, item))
}

func init() {
	effects.Register(effects.Spec{
		OracleID:     "1bcc8ca6-5fd0-4fc9-8f0c-f81d8abf9249",
		Name:         "Calix, Guided by Fate",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.B30SelfOrAnotherEnchantmentYouControlEntered(ev, source, g)
				}, "Calix, Guided by Fate — put a +1/+1 counter on target creature (constellation)",
					effects.PlusOneCounterOnTheTarget),
				effects.TargetCreature("target creature")),
			effects.Optional(effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.CombatDamageToPlayerBy(ev, source.Controller, g) &&
					(ev.Source == source.InstanceID || effects.IsEnchantedByAura(g, ev.Source)) &&
					!effects.B28DoneThisTurn(g, source.InstanceID, calixDone)
			}, calixLabel, calixCopyEffect),
				"Calix, Guided by Fate — create a token copy of a nonlegendary enchantment you control?"),
		},
	})
}
