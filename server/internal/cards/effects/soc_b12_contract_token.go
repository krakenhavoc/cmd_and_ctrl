package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// soc_b12_contract_token.go — Scriv, the Obligator's Contract (#4834): a
// white Aura enchantment token named Contract with enchant creature and
// the attack trigger printed on Scriv. It has no printing, so the text is
// the card's own, and it is not a Role (no Role rules apply to it).

// ContractToken is the Contract token's template (see
// printedContractToken).
func ContractToken() game.Card { return tokenFromCatalog(printedContractToken) }

// printedContractToken is the Contract as Scriv defines it. "You" in its
// trigger is the Contract's controller, which is the controller of the
// trigger's stack item (CR 109.5), so the opponents are measured from the
// player who made it. It is a trigger of the Aura, so it keeps working
// whoever controls the enchanted creature.
func printedContractToken() tokenTemplate {
	return tokenTemplate{
		Slug: "contract",
		Card: game.Card{
			Name:     "Contract",
			TypeLine: "Token Enchantment — Aura",
			Colors:   []string{"W"},
		},
		Triggered: []game.TriggeredAbility{
			WheneverEnchantedCreatureAttacks("Contract — +2/+0 if it is attacking one of your opponents, otherwise its controller loses 2 life",
				contractAttackEffect),
		},
		Text: "Enchant creature\nWhenever enchanted creature attacks, it gets +2/+0 until end of turn if it's attacking one of your opponents. Otherwise, its controller loses 2 life.",
	}
}

// contractAttackEffect is the Contract's trigger: the enchanted creature
// that attacked (the event's card) gets +2/+0 if it is attacking a player
// who is an opponent of the Contract's controller as the trigger
// resolves. Otherwise (it attacks a planeswalker or battle, or the
// Contract's controller themself, or it has been removed from combat) its
// controller loses 2 life; a creature that has left the battlefield is
// not attacking anyone, and the player who declared it loses the life.
func contractAttackEffect(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ev := item.Trigger.Event
	attacker, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !onBattlefield(g, ev.CardID) {
		return g.ChangePlayerLifeForEffect(item.SourceCardID, ev.Actor, -2)
	}
	if socB8AttackingAnOpponentOf(g, item.Controller, attacker) {
		return BoostUntilEOT{
			Target: attacker.InstanceID,
			Power:  2,
			Label:  "Contract — +2/+0",
		}.Apply(NewContext(g, item))
	}
	return g.ChangePlayerLifeForEffect(item.SourceCardID, attacker.Controller, -2)
}
