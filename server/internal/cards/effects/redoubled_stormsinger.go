package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Redoubled Stormsinger — Creature — Orc Wizard {2}{R}, 3/3 (Secrets of
// Strixhaven Commander, #4834):
//
//	"First strike
//	 Whenever this creature attacks, for each creature token you control
//	 that entered this turn, create a tapped and attacking token that's a
//	 copy of that token. At the beginning of the next end step, sacrifice
//	 those tokens."
//
// The tokens to copy are fixed as the ability resolves: every creature
// token its controller controls that entered the battlefield this turn
// (Game.EnteredThisTurn, under anyone's control when it entered). Each is
// copied in turn through the tapped-and-attacking door Calamity, Galloping
// Inferno uses (CreateTokensAttackingYourChoice), so the copy's controller
// chooses what each one attacks, and a copy was never declared as an
// attacker, so nothing that watches attacks sees it (CR 508.4). The copy is
// of the token's copiable values (CR 707.2), so a token that was itself a
// copy is copied as what it copied (the 2025-04-04 rulings). One delayed
// trigger sacrifices every token made at the next end step (CR 603.7).
//
// The rounds ride a keyed tail: the tokens still to copy are the front of
// its Carry (its N of them) and the copies already made the back, so a table
// waiting on an attack question is a restore point.
//
// No simplification.
func init() {
	socB11StormsingerNext = AttackingTokensThen("attacking-tokens/redoubled-stormsinger-next",
		func(ctx *Context, created []uuid.UUID, tail game.KeyedTail) error {
			item := tail.Asked.StackItem()
			if item == nil || tail.N > len(tail.Carry) {
				return nil
			}
			remaining := append([]uuid.UUID(nil), tail.Carry[:tail.N]...)
			made := append(append([]uuid.UUID(nil), tail.Carry[tail.N:]...), created...)
			return socB11StormsingerRound(ctx.Game, item, remaining, made)
		})
	Register(Spec{
		OracleID:        "792c4d38-41e6-49eb-bd4e-de0a5ac8330f",
		Name:            "Redoubled Stormsinger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Redoubled Stormsinger — copy each creature token that entered this turn, tapped and attacking; sacrifice the copies at the next end step",
				socB11StormsingerAttacks),
		},
	})
}

// socB11StormsingerNext is the rest of the rounds after one copy lands. An
// on-disk key.
var socB11StormsingerNext game.TailRef

// socB11StormsingerAttacks gathers the tokens and starts the first round.
func socB11StormsingerAttacks(g *game.Game, item *game.StackItem) error {
	var tokens []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsToken() && c.IsCreature() && c.Controller == item.Controller && g.EnteredThisTurn(c.InstanceID) {
			tokens = append(tokens, c.InstanceID)
		}
	}
	return socB11StormsingerRound(g, item, tokens, nil)
}

// socB11StormsingerRound copies the next token that can still be copied,
// or, with none left, schedules the sacrifice of every copy made.
func socB11StormsingerRound(g *game.Game, item *game.StackItem, remaining, made []uuid.UUID) error {
	ctx := NewContext(g, item)
	for len(remaining) > 0 {
		id := remaining[0]
		remaining = remaining[1:]
		tmpl, ok := TokenCopyTemplate(g, id)
		if !ok {
			continue
		}
		tail := withN(withCarry(ctx.ItemTail(socB11StormsingerNext), append(append([]uuid.UUID(nil), remaining...), made...)...), len(remaining))
		return CreateTokensAttackingYourChoice{
			Template: tmpl,
			Copy:     true,
			N:        1,
			Tapped:   true,
			Prefer:   triggerAttackTarget(item),
			Label:    "Redoubled Stormsinger",
			Tail:     tail,
		}.Apply(ctx)
	}
	if len(made) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label: "Redoubled Stormsinger — sacrifice the tokens",
		Cards: made,
		Body:  sacrificeListedCardsBody,
	}.Apply(ctx)
}
