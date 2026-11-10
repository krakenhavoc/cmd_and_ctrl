package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mardu Siegebreaker — Creature — Human Warrior {1}{R}{W}{B}, 4/4 (Tarkir:
// Dragonstorm, #4463):
//
//	"Deathtouch, haste
//	 When this creature enters, exile up to one other target creature you
//	 control until this creature leaves the battlefield.
//	 Whenever this creature attacks, for each opponent, create a tapped
//	 token that's a copy of the exiled card attacking that opponent. At the
//	 beginning of your next end step, sacrifice those tokens."
//
// The enters trigger is Aang's Iceberg's exile-until-this-leaves (CR
// 610.3) with ExileUntil.Linked, the seam this card needed: the exiled
// card is stamped as exiled with this ability of this object (CR 607.2a),
// so the attack trigger's "the exiled card" is exactly that card, and a
// second Siegebreaker's exile, or a card exiled by anything else, is not.
// A Siegebreaker that has left when the enters trigger resolves exiles
// nothing (CR 610.3b).
//
// The attack trigger reads that card (Game.ExiledWithRowForEffect) and,
// for each opponent still in the game, creates one token copy of it, in
// one creation, tapped and attacking that opponent (CR 508.4: never
// declared, so nothing that reads "whenever a creature attacks" sees them).
// The copy is of what is printed on the exiled card and nothing else (the
// ruling): the card as it sits in exile, never the last-known values of the
// permanent it was. If the card isn't a creature card the tokens are still
// made, tapped, and attack nothing. The exiled card gone (a token, or the
// Siegebreaker left) makes no copies. Whatever the creation made is
// sacrificed at the beginning of your next end step by Mobilize's delayed
// trigger.
//
// No simplification.
func init() {
	const exileLabel = "Mardu Siegebreaker — exile up to one other target creature you control until this creature leaves the battlefield"
	Register(Spec{
		OracleID:        "1f82e4b9-e67f-4f34-ac4c-57b0b300ee42",
		Name:            "Mardu Siegebreaker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch", "haste"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters(exileLabel, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return ExileUntil{Target: t.ID, ThisLeaves: true, Linked: true,
						Label: "Mardu Siegebreaker — the exiled card returns when Mardu Siegebreaker leaves the battlefield"}.Apply(ctx)
				}
				return nil
			}),
				Another(TargetCreature("up to one other target creature you control", YouControl()).WithCount(0, 1))),
			WheneverThisAttacks("Mardu Siegebreaker — for each opponent, create a tapped token that's a copy of the exiled card attacking that opponent",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					ref, ok := ctx.SourceRef()
					if !ok {
						return nil
					}
					exiled := g.ExiledWithRowForEffect(game.PermissionCardRef(ref), game.AbilitySlotTriggered, exileLabel)
					if len(exiled) == 0 {
						return nil
					}
					tmpl := tokenCopyTemplateFromCard(*exiled[len(exiled)-1])
					var groups []game.TokenGroup
					for _, opp := range ctx.Opponents() {
						if p := g.PlayerByIDForEffect(opp); p == nil || p.Eliminated {
							continue
						}
						groups = append(groups, game.TokenGroup{
							Template: copiedToken(tmpl), Count: 1,
							Entry: game.TokenEntryOptions{Tapped: true}, Attacking: opp,
						})
					}
					if len(groups) == 0 {
						return nil
					}
					return g.CreateTokensKeyedForEffect(game.TokenCreation{
						Controller: item.Controller, Source: item.SourceCardID, Groups: groups, Together: true,
					}, ctx.ItemTail(marduSiegebreakerSacrificeTokens))
				}),
		},
	})
}

// marduSiegebreakerSacrificeTokens is "at the beginning of your next end
// step, sacrifice those tokens": the tokens the creation made. An on-disk
// key: never renamed, never reused.
var marduSiegebreakerSacrificeTokens = CardTail("mardu-siegebreaker/sacrifice-the-tokens", func(ctx *Context, o game.TailOutcome) error {
	created := make([]uuid.UUID, 0, len(o.IDs))
	created = append(created, o.IDs...)
	if len(created) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label: "Mardu Siegebreaker — sacrifice those tokens",
		Cards: created,
		Body:  sacrificeListedCardsBody,
	}.Apply(ctx)
})
