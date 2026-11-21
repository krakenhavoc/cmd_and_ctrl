package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thorin, Mountain-king — Legendary Creature — Dwarf Noble {3}{R}, 3/4
// (The Hobbit, #4454):
//
//	"Trample
//	 When Thorin enters, attach any number of target Equipment you
//	 control to target creature you control. When one or more Equipment
//	 become attached to that creature this way, that creature deals
//	 damage equal to its power to up to one target creature."
//
// Iron Hills Stalwart's two clauses (AttachClauseTargets' shape), the first
// "any number" (WithCount(0, 0)). Each is re-checked as the trigger resolves
// (CR 608.2b). An Equipment that is already on the creature, or that could
// not legally be attached to it, does not "become attached" (CR 701.3b), so
// only the ones that moved count: if none did, the reflexive trigger is not
// raised. The reflexive trigger is one trigger however many Equipment moved,
// chooses its own "up to one target creature" as it goes on the stack, and
// reads the creature's power as it resolves, so an Equipment attached this
// way counts. The creature that was attached to, not Thorin, is the damage
// source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "94917b80-ac32-45ab-b4ee-157ade024bcd",
		Name:            "Thorin, Mountain-king",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Thorin, Mountain-king — attach any number of target Equipment you control to target creature you control", thorinMountainKingAttach),
				effects.Clauses(
					effects.TargetPermanent("any number of target Equipment you control", effects.HasSubtype("Equipment"), effects.YouControl()).WithCount(0, 0),
					effects.TargetCreature("target creature you control", effects.YouControl()),
				)),
		},
	})
}

// thorinMountainKingAttach attaches the legal Equipment targets to the
// creature, then raises the reflexive damage if any of them became attached.
func thorinMountainKingAttach(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	chosen, ok := ctx.ClauseTarget(1)
	if !ok {
		return nil
	}
	host := game.TargetRef{Kind: game.TargetCard, ID: chosen.ID}
	attached := 0
	for _, t := range ctx.ClauseTargets(0) {
		if !ctx.IsTargetLegal(t) || !g.AttachLegalForEffect(t.ID, host.ID) {
			continue
		}
		if c, found := g.LookupCardForEffect(t.ID); !found || c.AttachedTo == host {
			continue
		}
		if err := g.AttachForEffect(t.ID, host); err != nil {
			return err
		}
		attached++
	}
	if attached == 0 {
		return nil
	}
	return effects.ReflexiveTrigger{
		Label: "Thorin, Mountain-king — that creature deals damage equal to its power to up to one target creature",
		Cards: []uuid.UUID{host.ID},
		Body:  thorinMountainKingDamageBody,
	}.Apply(ctx)
}

// thorinMountainKingDamageBody is the reflexive trigger's registration: the
// creature is the payload, the victim the trigger's own "up to one" target.
var thorinMountainKingDamageBody = game.ReflexiveBody("thorin-mountain-king/damage", effects.SimpleBody(thorinMountainKingDamage),
	effects.ConstTargets(func() *game.TargetSpec { return effects.UpToOneTargetCreature("up to one target creature") }))

// thorinMountainKingDamage is the payload creature dealing damage equal to
// its power, as it last stood, to the chosen creature.
func thorinMountainKingDamage(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	creatures := ctx.PayloadCards()
	if len(creatures) == 0 {
		return nil
	}
	power := effects.B29PowerAsItLastStood(g, creatures[0])
	if power <= 0 {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		return effects.DealDamage{Source: creatures[0], Target: t.ID, Amount: power}.Apply(ctx)
	}
	return nil
}
