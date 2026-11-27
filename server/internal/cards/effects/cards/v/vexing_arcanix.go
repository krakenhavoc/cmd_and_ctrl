package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vexing Arcanix — Artifact {4} (Eighth Edition, S82 OLD batch 610, #5607):
//
//	"{3}, {T}: Target player chooses a card name, then reveals the top card of
//	 their library. If that card has the chosen name, that player puts it into
//	 their hand. Otherwise, they put it into their graveyard and this artifact
//	 deals 2 damage to them."
//
// The target player (not the controller) is asked for a name by a keyed
// card-name prompt (Conundrum Sphinx's), private to them; the name is free text
// (CR 201.2) and a split or double-faced card matches under any face. Their
// top card is then revealed: a match goes to their hand, anything else to their
// graveyard (not a mill, so it is an ordinary move) and the Arcanix deals 2
// damage to them as the object it was. An empty library does nothing, and a
// target who has left the game is not asked.
//
// No simplification.
func init() {
	vexingArcanixNamed = game.RegisterTail("vexing-arcanix/named", vexingArcanixNamedBody)
	effects.Register(effects.Spec{
		OracleID:     "2fd5b7e5-496c-4f6d-85e0-256afb7544d1",
		Name:         "Vexing Arcanix",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Target player chooses a card name, then reveals the top card of their library.",
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				player, ok := effects.FirstLegalPlayerTarget(effects.NewContext(g, item))
				if !ok {
					return nil
				}
				g.QueueCardNameChoiceKeyedForEffect(player, item.SourceCardID,
					"Vexing Arcanix — choose a card name",
					game.KeyedTail{Ref: vexingArcanixNamed, Player: player, Source: item.SourceCardID})
				return nil
			},
		}},
	})
}

var vexingArcanixNamed game.TailRef

func vexingArcanixNamedBody(g *game.Game, o game.TailOutcome) error {
	player, source, name := o.Tail.Player, o.Tail.Source, o.Tail.Word
	revealed := g.RevealTopOfLibraryForEffect(player, source, 1, "Vexing Arcanix — the top card of the library")
	if len(revealed) == 0 {
		return nil
	}
	return effects.TakeFromLibraryToHand{
		Player: player,
		Cards:  revealed,
		Match:  func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return game.CardNameMatches(c, name) },
		All:    true,
		Label:  "Vexing Arcanix — it has the chosen name, so it goes into hand",
		Key:    vexingArcanixTake,
	}.Apply(effects.KeyedContext(g, player, source))
}

// vexingArcanixTake is the rest of the take: a card that did not have the
// name goes into its owner's graveyard and the Arcanix deals 2 damage to the
// player.
var vexingArcanixTake = effects.RegisterTake("choose-cards/vexing-arcanix-take", effects.TakeDef{Then: func(g *game.Game, res effects.TakeFromLibraryResult) error {
	if len(res.Rest) == 0 {
		return nil
	}
	if err := effects.RestIntoGraveyard(g, res.Rest); err != nil {
		return err
	}
	return g.DealDamageToPlayerForEffect(res.Source, res.Player, 2)
}})
