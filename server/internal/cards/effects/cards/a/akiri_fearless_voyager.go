package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Akiri, Fearless Voyager — Legendary Creature — Kor Warrior {1}{R}{W}, 3/3
// (Commander Legends: Battle for Baldur's Gate):
//
//	"Whenever you attack a player with one or more equipped creatures, draw a
//	 card.
//	 {W}: You may unattach an Equipment from a creature you control. If you
//	 do, tap that creature and it gains indestructible until end of turn."
//
// The draw is Ordruun Mentor's "whenever you attack a player" (one trigger
// per player attacked in a declaration, CR 508.3e, 603.2c), counting only an
// attacker that is equipped as it is declared. The second ability picks at
// resolution (ChoosePermanents, not a target) an Equipment attached to a
// creature you control, whoever controls the Equipment; declining is the
// "may". Unattaching puts the Equipment on the battlefield unattached
// (CR 701.3d), then the creature it left is tapped and given indestructible.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c03ae426-2536-4c20-a73a-a24e816f76b0",
		Name:         "Akiri, Fearless Voyager",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatchPerPlayer(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !effects.B16YouAttackedAPlayer(ev, source, g) {
					return false
				}
				attacker, ok := g.LookupCardForEffect(ev.CardID)
				return ok && effects.IsEquippedCreature(g, attacker)
			}, "Akiri, Fearless Voyager — draw a card", effects.Do(effects.DrawCards{N: 1}))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}: You may unattach an Equipment from a creature you control. If you do, tap that creature and it gains indestructible until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    effects.ManaCost("{W}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ChoosePermanents{
					Question:   "Akiri, Fearless Voyager — unattach an Equipment from a creature you control? (tap that creature; it gains indestructible)",
					Candidates: akiriEquipmentOnYourCreatures,
					Keyed:      akiriUnattach,
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

// akiriEquipmentOnYourCreatures offers every Equipment attached to a creature
// `of` controls, up to one: the pick is optional ("you may").
func akiriEquipmentOnYourCreatures(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if !c.HasSubtype("Equipment") || c.AttachedTo.Kind != game.TargetCard {
			continue
		}
		host, ok := g.LookupCardForEffect(c.AttachedTo.ID)
		if ok && host.IsCreature() && host.Controller == of {
			out = append(out, c.InstanceID)
		}
	}
	return out, 0, 1
}

// akiriUnattach is the pick's rest: unattach the Equipment, and if it came
// off, tap the creature it was on and give it indestructible until end of
// turn. An on-disk key: never renamed, never reused.
var akiriUnattach = effects.RegisterPermanentsPick("prompt-run/akiri-fearless-voyager-unattach", effects.PermanentsDef{
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		ids := picked.Cards()
		if len(ids) == 0 {
			return nil
		}
		eq, ok := ctx.Game.LookupCardForEffect(ids[0])
		if !ok || eq.AttachedTo.Kind != game.TargetCard {
			return nil
		}
		host := eq.AttachedTo.ID
		if err := ctx.Game.UnattachForEffect(ids[0]); err != nil {
			return err
		}
		if err := (effects.TapTarget{Target: host}).Apply(ctx); err != nil {
			return err
		}
		return effects.GrantKeywordUntilEOT{
			Target:   host,
			Keywords: []string{"indestructible"},
			Label:    "Akiri, Fearless Voyager — indestructible until end of turn",
		}.Apply(ctx)
	},
})
