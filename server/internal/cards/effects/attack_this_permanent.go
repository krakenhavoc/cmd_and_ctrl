package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attack_this_permanent.go — "attacks <this planeswalker> if able"
// during one player's next turn (#2567, ADR 0045 amendment
// 2026-10-22). Two shapes, one requirement:
//
//   - Gideon Jura's "+2: During target opponent's next turn, creatures
//     that player controls attack Gideon Jura if able." is
//     PlayersCreaturesAttackThisIfAble: a record whose affected set is
//     the live rule "creatures that player controls"
//     (game.ScopeYourCreatures with the target opponent as its "you",
//     Incite War's shape), lasting until the end of that player's next
//     turn. The set is not locked (CR 611.2c: a requirement changes no
//     characteristic and no control), so a creature that player gains
//     before that turn must attack too — the card's ruling. Another
//     opponent's creatures are never in the set, so in a four-player
//     game the turns in between ask nothing of anyone.
//   - Gideon, Battle-Forged's "+2: Up to one target creature an
//     opponent controls attacks Gideon during its controller's next
//     turn if able." is CreatureAttacksThisDuringItsControllersNextTurn:
//     a record pinned to that creature, lasting until the end of its
//     controller's next turn with "its controller" read live
//     (game.UntilEndOfItsControllersNextTurn).
//
// The requirement itself is game.ModAttackPermanentIfAble: it names the
// planeswalker as the object it was when the ability was activated
// (CR 400.7), so a Gideon that has left — or left and come back — is
// owed nothing, and it is held back until a turn after the one it was
// made in, so an instant-speed activation during that player's own turn
// does not reach the rest of that turn. CR 508.1d's maximisation counts
// it like every other requirement: a creature that cannot attack Gideon
// for free may attack anything else or stay home.

// PlayersCreaturesAttackThisIfAble is "During <Player>'s next turn,
// creatures that player controls attack <this permanent> if able"
// (Gideon Jura).
type PlayersCreaturesAttackThisIfAble struct {
	// Player is the player whose creatures, and whose next turn.
	Player uuid.UUID
	// Label is attribution for logs and tests.
	Label string
}

// Apply registers the requirement as a data record. A source with no
// object to name (the ability has no source) registers nothing.
func (a PlayersCreaturesAttackThisIfAble) Apply(ctx *Context) error {
	if a.Player == uuid.Nil {
		return nil
	}
	ref, ok := ctx.SourceRef()
	if !ok || ref.ID == uuid.Nil {
		return nil
	}
	g := ctx.Game
	g.RegisterScopedRuleEffectForEffect(ctx.Source(), game.ScopeYourCreatures, a.Player,
		[]game.Mod{game.AttackPermanentIfAbleMod(ref, g.Turn.Seq)},
		g.UntilEndOfYourNextTurnDuration(a.Player),
		eotLabel(a.Label, "creatures that player controls attack this permanent if able"))
	return nil
}

// CreatureAttacksThisDuringItsControllersNextTurn is "<Creature>
// attacks <this permanent> during its controller's next turn if able"
// (Gideon, Battle-Forged).
type CreatureAttacksThisDuringItsControllersNextTurn struct {
	// Creature is the creature that must attack.
	Creature uuid.UUID
	// Label is attribution for logs and tests.
	Label string
}

// Apply registers the requirement pinned to the creature. A creature no
// longer on the battlefield registers nothing.
func (a CreatureAttacksThisDuringItsControllersNextTurn) Apply(ctx *Context) error {
	if a.Creature == uuid.Nil {
		return nil
	}
	ref, ok := ctx.SourceRef()
	if !ok || ref.ID == uuid.Nil {
		return nil
	}
	g := ctx.Game
	d, ok := g.UntilEndOfItsControllersNextTurnDuration(a.Creature)
	if !ok {
		return nil
	}
	return ScopedEffectFor{
		Target:   a.Creature,
		Mods:     []game.Mod{game.AttackPermanentIfAbleMod(ref, g.Turn.Seq)},
		Duration: d,
		Label:    eotLabel(a.Label, "that creature attacks this permanent if able"),
	}.Apply(ctx)
}
