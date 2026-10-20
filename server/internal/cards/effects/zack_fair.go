package effects

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zack Fair — Legendary Creature — Human Soldier {W}, 0/1 (#3656):
//
//	"Zack Fair enters with a +1/+1 counter on it.
//	 {1}, Sacrifice Zack Fair: Target creature you control gains
//	 indestructible until end of turn. Put Zack Fair's counters on that
//	 creature and attach an Equipment that was attached to Zack Fair to
//	 that creature."
//
// The entry counter is the self-replacement every "enters with" creature
// uses. The ability reads Zack's counters from the sacrificed
// permanent's last-known information (Context.SacrificedPermanent), so
// the counters he had when he was sacrificed move over, and a target
// that left in response gets nothing (CR 608.2b).
//
// "Attach an Equipment that was attached to Zack Fair" reads the
// sacrificed permanent's last-known attachments
// (PermanentInfo.Attachments, #3780): the state-based action that
// unattaches an Equipment from a host that left (CR 704.5n) has run
// before this resolves, so the Equipment's own link is gone. Of those
// objects, only an Equipment still on the battlefield as the same object
// (CR 400.7) that could legally be attached to the target counts (CR
// 701.3b). One candidate is attached; several make the controller pick
// one (the printed text says "an Equipment"). A target that left in
// response gets nothing, indestructible included (CR 608.2b), so the
// Equipment stays where it is.
func init() {
	Register(Spec{
		OracleID:     "e84d146a-1c67-49df-b062-622794e45c42",
		Name:         "Zack Fair",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters("+1/+1", 1, "Zack Fair enters with a +1/+1 counter on it."),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice Zack Fair: Target creature you control gains indestructible until end of turn. Put Zack Fair's counters on that creature and attach an Equipment that was attached to Zack Fair to that creature.",
			Cost:    Plus(ManaCost("{1}"), SacrificeThis()),
			Targets: TargetCreature("target creature you control", YouControl()),
			Effect: Do(
				// The Equipment moves after a prompt a counter asks (CR 608.2c, #4155).
				// Zack's counters and Equipment are read from the payment record.
				Instruction(func(ctx *Context) error {
					target, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					if err := (GrantKeywordUntilEOT{Target: target, Keywords: []string{"indestructible"}, Label: "Zack Fair — indestructible"}).Apply(ctx); err != nil {
						return err
					}
					zack, ok := ctx.SacrificedPermanent()
					if !ok {
						return nil
					}
					kinds := make([]string, 0, len(zack.Counters))
					for kind, n := range zack.Counters {
						if n > 0 {
							kinds = append(kinds, kind)
						}
					}
					sort.Strings(kinds)
					for _, kind := range kinds {
						if err := (AddCounter{Target: target, Kind: kind, N: zack.Counters[kind]}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
				Instruction(func(ctx *Context) error {
					target, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					zack, ok := ctx.SacrificedPermanent()
					if !ok {
						return nil
					}
					return zackFairMoveEquipment(ctx, zack, target)
				}),
			),
		}},
	})
}

// zackFairMoveEquipment attaches one Equipment that was attached to
// Zack Fair as he left to `target`.
func zackFairMoveEquipment(ctx *Context, zack game.PermanentInfo, target uuid.UUID) error {
	g := ctx.Game
	var candidates []uuid.UUID
	for _, id := range g.LiveAttachmentsForEffect(zack.Attachments) {
		card, ok := g.LookupCardForEffect(id)
		if !ok || !card.HasSubtype("Equipment") || !g.AttachLegalForEffect(id, target) {
			continue
		}
		candidates = append(candidates, id)
	}
	attach := func(g *game.Game, id uuid.UUID) error {
		return g.AttachForEffect(id, game.TargetRef{Kind: game.TargetCard, ID: target})
	}
	switch len(candidates) {
	case 0:
		return nil
	case 1:
		return attach(g, candidates[0])
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Item.SourceCardID,
		Question: "Zack Fair — choose an Equipment that was attached to him to attach to the target",
		Cards:    candidates,
		Min:      1,
		Max:      1,
		Zone:     game.ZoneBattlefield,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			if len(picked) != 1 {
				return nil
			}
			return attach(g, picked[0])
		},
	})
	return nil
}
