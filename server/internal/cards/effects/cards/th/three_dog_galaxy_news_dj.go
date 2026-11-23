package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Three Dog, Galaxy News DJ — Legendary Creature — Human Bard {1}{R}{W}, 1/5
// (Fallout, #5948):
//
//	"Whenever you attack, you may pay {2} and sacrifice an Aura attached to
//	 Three Dog. When you sacrifice an Aura this way, for each other attacking
//	 creature you control, create a token that's a copy of that Aura attached
//	 to that creature."
//
// "Whenever you attack" is once per declaration (OncePerBatch). The {2} is the
// ordinary "you may pay" (MayPay); the sacrifice is asked as its paid branch
// resolves, over the Auras attached to Three Dog then (maySacrificeOne, CR
// 701.21), and only an Aura that really went raises the CR 603.12 reflexive
// trigger. That carries the Aura as the object it was (CR 400.7), so its copiable
// values are read from its last-known record even if it was a token that has
// since ceased to exist (CR 707.2). As the reflexive trigger resolves a token
// copy is created for each other creature its controller has attacking, entering
// attached to it, unless the Aura could not enchant that creature (CR 303.4j);
// the tokens are never cast, so they trigger nothing that watches casts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cc504f00-f177-4def-956d-92a2217438bb",
		Name:         "Three Dog, Galaxy News DJ",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller)
			}, "Three Dog, Galaxy News DJ — you may pay {2} and sacrifice an Aura attached to it; when you do, copy that Aura onto each other attacking creature",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{2}",
						Question: "Three Dog — pay {2} and sacrifice an Aura attached to it?",
						Then:     threeDogPaid,
					}.Apply(effects.NewContext(g, item))
				})),
		},
	})
}

// threeDogPaid is the paid half: the sacrifice of an Aura attached to Three
// Dog. An on-disk key: never renamed, never reused.
var threeDogPaid = effects.PayUnlessBranches("pay-unless/three-dog-sacrifice-an-aura",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		item := ctx.Item
		if item == nil || !effects.OnBattlefield(ctx.Game, item.SourceCardID) {
			return nil
		}
		src := item.SourceCardID
		attachedAura := func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsAura() && c.IsAttachedTo(src) }
		return effects.MaySacrificeOne(ctx.Game, item, "Three Dog — sacrifice an Aura attached to it?", attachedAura, false, threeDogSacrificed)
	}, nil)

// threeDogSacrificed raises the reflexive trigger for the Aura that went.
var threeDogSacrificed = effects.RegisterSacrificedOne("prompt-run/three-dog-sacrificed-aura", func(ctx *effects.Context, sacrificed uuid.UUID) error {
	ref, ok := ctx.Game.PermanentRefForEffect(sacrificed)
	if !ok {
		return nil
	}
	return effects.ReflexiveTrigger{
		Label:  "Three Dog — create a token copy of that Aura attached to each other attacking creature you control",
		Body:   threeDogCopyBody,
		Params: game.EffectParams{Object: ref},
	}.Apply(ctx)
})

var threeDogCopyBody = game.ReflexiveBody("three-dog/copy-the-aura", func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	tmpl, ok := g.TokenCopyTemplateForObjectForEffect(p.Object)
	if !ok {
		return nil
	}
	aura, _ := g.LookupCardForEffect(p.Object.ID)
	var hosts []uuid.UUID
	for _, id := range effects.AttackingCreatures(g) {
		c, found := g.LookupCardForEffect(id)
		if found && c.Controller == item.Controller && id != item.SourceCardID {
			hosts = append(hosts, id)
		}
	}
	for _, host := range hosts {
		if !effects.OnBattlefield(g, host) || !effects.CanBeAttachedTo(g, item.Controller, aura, host) {
			continue
		}
		if _, err := g.CreateTokensForEffect(item.Controller, effects.CopiedToken(tmpl), 1, game.TokenEntryOptions{
			AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: host},
		}); err != nil {
			return err
		}
	}
	return nil
}, nil)
