package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sword of Hearth and Home — Artifact — Equipment for {3}:
//
//	"Equipped creature gets +2/+2 and has protection from green and
//	 from white.
//	 Whenever equipped creature deals combat damage to a player, exile
//	 up to one target creature you own, then search your library for a
//	 basic land card. Put both cards onto the battlefield under your
//	 control, then shuffle.
//	 Equip {2}"
//
// "PUT BOTH CARDS ONTO THE BATTLEFIELD" is one event, and that is what
// waited on #1324. The creature comes out of exile and the land out of
// the library at the same time, so every permanent on the battlefield
// — both newcomers included — is checked against that one event for
// enters triggers (CR 603.6a). Doing the two entries one after the
// other is visible on the deck this card is in: with Loyal Warhound
// blinked, Warhound first would read the lands before the new basic
// arrived (its intervening-if, CR 603.4, sees one land too few), and
// land first would leave a landfall creature entering alongside it
// blind to it. Game.PutOntoBattlefieldTogetherThenForEffect takes the
// two cards from their two zones as one simultaneous entry
// (server/internal/game/entry_batch.go).
//
// The sentence's parts, in the order the card prints them:
//
//   - the exile is its own event and can pause (a commander's owner is
//     asked about the command zone, CR 903.9), so the search waits in
//     its continuation, and only a card that really reached exile is
//     put back. A token exiled this way ceases to exist (CR 111.8) and
//     is left out, which also keeps it from refusing the land's entry.
//   - the search finds the land and LEAVES it in the library (ToTop
//     with no shuffle, the idiom Intuition uses), because the put is
//     the next sentence and the shuffle the one after it. A player may
//     always fail to find (CR 701.23b).
//   - the exiled card returns as a new object (CR 400.7) under the
//     Sword's controller — "under your control" — even when an
//     opponent was controlling your creature.
//   - the shuffle runs after the entry has finished, which is after
//     any question a card of it asked (a fetched shockland's life).
//
// Protection from green and from white is ADR 0072's layer-6 grant,
// as on the other Swords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "913e6182-706a-4872-8c8a-e146b0ae0738",
		Name:         "Sword of Hearth and Home",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttached("protection from green", "protection from white"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			},
			Targets: TargetCreature("up to one target creature you own", YouOwn()).WithCount(0, 1),
			Key:     "Sword of Hearth and Home — exile a creature you own, then put it and a basic land onto the battlefield",
			Effect:  swordOfHearthAndHomeResolve,
		}},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}

// swordOfHearthAndHomeResolve is the trigger's effect. Package-level,
// and every continuation below captures only scalars, so an undo
// across any of its prompts resolves it against the restored game.
func swordOfHearthAndHomeResolve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	you, source := item.Controller, item.SourceCardID
	var targets []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard && t.ID != uuid.Nil {
			targets = append(targets, t.ID)
		}
	}
	return g.ExileCardsThenForEffect(targets, func(g *game.Game, exiled []uuid.UUID) error {
		return g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
			Player: you,
			Source: source,
			Pred:   IsBasicLand,
			Dest:   game.ZoneLibrary,
			Limit:  1,
			// ToTop with no shuffle: the search reports what it found
			// and leaves it in the library for the put. The shuffle is
			// printed after the put, and runs from its continuation.
			ToTop:  true,
			Reason: "Sword of Hearth and Home — search for a basic land card",
			Then: func(g *game.Game, found []uuid.UUID) error {
				return swordOfHearthAndHomePutBoth(g, you, exiled, found)
			},
		})
	})
}

// swordOfHearthAndHomePutBoth is "put both cards onto the battlefield
// under your control, then shuffle": one simultaneous entry of the
// exiled card and the found land, and the shuffle once it has finished.
func swordOfHearthAndHomePutBoth(g *game.Game, you uuid.UUID, exiled, found []uuid.UUID) error {
	var entries []game.BatchEntry
	for _, id := range exiled {
		c, ok := g.LookupCardForEffect(id)
		if z := g.FindCardZoneForEffect(id); !ok || z == nil || z.Kind != game.ZoneExile || !isPermanentCard(c) {
			continue
		}
		entries = append(entries, game.BatchEntry{CardID: id, From: game.ZoneExile})
	}
	for _, id := range found {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneLibrary {
			continue
		}
		entries = append(entries, game.BatchEntry{CardID: id, From: game.ZoneLibrary})
	}
	shuffle := func(g *game.Game, _ []uuid.UUID) error {
		return g.ShuffleLibraryForEffect(you)
	}
	if len(entries) == 0 {
		return shuffle(g, nil)
	}
	return g.PutOntoBattlefieldTogetherThenForEffect(entries, game.ZoneEntryOptions{Controller: you}, shuffle)
}
