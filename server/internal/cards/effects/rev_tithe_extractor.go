package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rev, Tithe Extractor — Legendary Creature — Human Rogue {3}{B}, 3/3
// (Foundations Jumpstart, #5929):
//
//	"Whenever you attack, target creature gains deathtouch until end of turn.
//	 Whenever one or more creatures you control deal combat damage to a
//	 player, create a Treasure token, then look at the top card of that
//	 player's library and exile it face down. You may cast that card for as
//	 long as it remains exiled."
//
// The attack trigger is Bast's shape (once per declaration, any creature as
// the target). The damage trigger is one per player hit in a combat damage
// step (CR 603.2c), however many creatures connect. The exiled card is
// Predators' Hour's: exiled face down with the controller allowed to look at
// it and to cast it, though not play it as a land, for as long as it stays
// exiled; the mana is paid as usual (no any-colour rider).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8345eb05-c04e-4c1d-9776-5ab7d91a7842",
		Name:         "Rev, Tithe Extractor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(Targeting(
				On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclaredByYou(ev, source.Controller)
				}, "Rev — target creature gains deathtouch until end of turn", revDeathtouch),
				TargetCreature("target creature"))),
			WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(nil,
				"Rev — create a Treasure token, then exile the top card of that player's library face down; you may cast it for as long as it remains exiled",
				Do(
					// The exile waits for a prompt the Treasure's creation asks (CR 608.2c, #4155).
					CreateToken{Template: TreasureToken(), N: 1},
					Instruction(revExileTop),
				)),
		},
	})
}

func revDeathtouch(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return GrantKeywordUntilEOT{Target: id, Keywords: []string{"deathtouch"}, Label: "Rev — deathtouch"}.Apply(ctx)
}

func revExileTop(ctx *Context) error {
	if ctx.Item.Trigger == nil {
		return nil
	}
	return ExileTopWithPermission{
		From:        ctx.Item.Trigger.Event.Target,
		GrantTo:     ctx.Controller(),
		N:           1,
		CastOnly:    true,
		FaceDown:    true,
		WhileExiled: true,
	}.Apply(ctx)
}
