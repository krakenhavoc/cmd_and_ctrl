package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ob Nixilis, Unshackled — Legendary Creature — Demon {4}{B}{B}, 4/4:
//
//	"Flying, trample
//	 Whenever an opponent searches their library, that player
//	 sacrifices a creature of their choice and loses 10 life.
//	 Whenever another creature dies, put a +1/+1 counter on Ob
//	 Nixilis."
//
// The search trigger is Wan Shi Tong's condition
// (AnOpponentSearchesTheirOwnLibrary): an opponent searching THEIR OWN
// library, so a Bribery-style search of somebody else's library does not
// count, and the shuffle that follows a search is not a second search.
// "That player" is the searcher, read off the triggering event. They
// pick the creature as the trigger resolves (SacrificeChoice), and the
// 10 life is lost once it has gone; a player with no creature still
// loses the life.
//
// The counter trigger is AnotherCreatureDied: any creature but Ob
// Nixilis himself, tokens included, whoever controlled it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "73021bd2-3b60-4fe4-8858-6eb525be1422",
		Name:            "Ob Nixilis, Unshackled",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{
			On(game.EventSearchLibrary, AnOpponentSearchesTheirOwnLibrary,
				"Ob Nixilis, Unshackled — that player sacrifices a creature and loses 10 life",
				obNixilisUnshackledPunish),
			On(game.EventLTB, AnotherCreatureDied,
				"Ob Nixilis, Unshackled — put a +1/+1 counter on Ob Nixilis",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}

// obNixilisUnshackledPunish is the search trigger's resolution.
func obNixilisUnshackledPunish(g *game.Game, item *game.StackItem) error {
	victim := item.Trigger.Event.Actor
	if p := g.PlayerByIDForEffect(victim); p == nil || p.Eliminated {
		return nil
	}
	source := item.SourceCardID
	return SacrificeChoice{
		Player:     victim,
		Candidates: creaturesControlledByPlayer(g, victim),
		Question:   "Ob Nixilis, Unshackled — sacrifice a creature",
		// Keyed (#4253 slice 10b): the loss is the sacrifice's tail.
		ThenTail: game.KeyedTail{Ref: thenChangeLife, Player: victim, Source: source, N: -10},
	}.Apply(NewContext(g, item))
}
