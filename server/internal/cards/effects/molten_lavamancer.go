package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Molten Lavamancer — Creature — Elemental Warlock Villain {2}{R}, 2/3
// (Marvel Super Heroes Commander, #4455):
//
//	"Prowess
//	 Whenever a source you control deals noncombat damage to one or more of
//	 your opponents during your turn, you create a 1/1 red Elemental
//	 creature token. This ability triggers only once each turn."
//
// Quest for Pure Flame's damage read: a damage event to a player other
// than you, from a source you control (a spell on the stack is its
// caster's, b34DamageSourceController), not combat damage, on your turn.
// "One or more" is OncePerBatch (damage dealt together is one trigger) and
// "only once each turn" is Dionus's per-object tally, which counts the
// trigger when it goes on the stack, so a second noncombat hit later in the
// turn does nothing.
//
// No simplification.
const moltenLavamancerLabel = "Molten Lavamancer — create a 1/1 red Elemental creature token"

func init() {
	Register(Spec{
		OracleID:        "5fc60833-6e03-4bfa-833f-7d51c696db93",
		Name:            "Molten Lavamancer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventDealDamage, moltenLavamancerDamage, moltenLavamancerLabel,
				Do(CreateToken{Template: TokenCard("1/1 red Elemental"), N: 1}))),
		},
	})
}

func moltenLavamancerDamage(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || ev.Combat || ev.Amount <= 0 || ev.Target == source.Controller ||
		g.PlayerByIDForEffect(ev.Target) == nil || !IsYourTurn(g, source.Controller) {
		return false
	}
	controller, ok := b34DamageSourceController(ev, g)
	return ok && controller == source.Controller && !b11TriggeredThisTurn(g, source.InstanceID, moltenLavamancerLabel)
}
