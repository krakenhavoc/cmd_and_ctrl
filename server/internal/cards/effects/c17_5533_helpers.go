package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// c17_5533_helpers.go — shared bodies for the Commander 2017 batch (#5533).
// Every package-level name carries the c17 prefix so another batch cannot
// collide with it.

// Token rows for this batch, added from this file's init so the shared
// table's column alignment is not rewritten (#4452). The keys read like the
// printed text, as the table's do.
func init() {
	for key, card := range map[string]game.Card{
		// Hungry Lynx.
		"1/1 black Rat with deathtouch": {Name: "Rat", TypeLine: "Token Creature — Rat", Power: 1, Toughness: 1, Colors: []string{"B"}, Keywords: []string{"deathtouch"}},
		// Wasitora, Nekoru Queen.
		"3/3 black and red and green Cat Dragon with flying": {Name: "Cat Dragon", TypeLine: "Token Creature — Cat Dragon", Power: 3, Toughness: 3, Colors: []string{"B", "R", "G"}, Keywords: []string{"flying"}},
	} {
		if _, dup := TokenTable[key]; dup {
			panic("c17_5533_helpers: token key already in the table: " + key)
		}
		TokenTable[key] = card
	}
}

// c17CurseWhenAttacked is the pair of abilities a Commander 2017 Curse
// prints: "Whenever enchanted player is attacked, you <do X>. Each opponent
// attacking that player does the same." The first is Curse of Verbosity's
// own trigger, batched so one declaration is one "is attacked" (CR 506.3);
// the second is eachAttackingOpponentDoesTheSame, the attacking opponent's
// copy of the effect under their control. Curse of Bounty, Curse of
// Disturbance and Curse of Vitality.
func C17CurseWhenAttacked(mine, theirs string, effect Effect) []game.TriggeredAbility {
	you := OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return source.AttachedTo.Kind == game.TargetPlayer && source.AttachedTo.ID == ev.Target
	}, mine, effect))
	return []game.TriggeredAbility{you, EachAttackingOpponentDoesTheSame(theirs, effect)}
}

// c17UntapAllNonlandYouControl is "untap all nonland permanents you
// control" for the item's controller, who is the attacking opponent on
// Curse of Bounty's second ability.
func C17UntapAllNonlandYouControl(g *game.Game, item *game.StackItem) error {
	return B16UntapAllYouControlMatching(NewContext(g, item), item.Controller, func(c game.Card) bool { return !c.IsLand() })
}

// c17ARatDied is Hungry Lynx's "Whenever a Rat dies": any player's creature
// that was a Rat as it last existed on the battlefield (CR 603.10a). The
// word is read from the source's text (CR 612.2).
func C17ARatDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := DiedCreature(ev, g)
	return ok && LeftAsSubtype(ev, dead, source.TextSubtype("Rat"))
}

// c17ADragonYouControlDiedWhileInGraveyard is Boneyard Scourge's "Whenever a
// Dragon you control dies while this card is in your graveyard": the dying
// creature was a Dragon its controller controlled as it last existed, and
// the Scourge was already in the graveyard (furiousForebearWasInGraveyard:
// not its own death, nor a death in the batch that put it there).
func C17ADragonYouControlDiedWhileInGraveyard(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
	dead, ok := DiedCreature(ev, g)
	if !ok || LeftUnderControlOf(ev, dead) != source.Controller || !LeftAsSubtype(ev, dead, source.TextSubtype("Dragon")) {
		return false
	}
	return FuriousForebearWasInGraveyard(ev, source, ch, g)
}

// c17AttachAllEquipmentToSource is Balan's "Attach all Equipment you
// control to Balan": every Equipment the controller controls as the ability
// resolves, moved onto the source while it is still the same permanent
// (CR 400.7). AttachForEffect leaves alone an Equipment that can't be
// attached to it (CR 701.3b).
func C17AttachAllEquipmentToSource(g *game.Game, item *game.StackItem) error {
	if !SourceIsStillThisPermanent(g, item) {
		return nil
	}
	var gear []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.HasSubtype("Equipment") {
			gear = append(gear, c.InstanceID)
		}
	}
	host := game.TargetRef{Kind: game.TargetCard, ID: item.SourceCardID}
	for _, id := range gear {
		if err := g.AttachForEffect(id, host); err != nil {
			return err
		}
	}
	return nil
}

// c17ThenPutFoundHammerOntoBattlefield is Nazahn's "If you reveal a card
// named Hammer of Nazahn this way, put it onto the battlefield. Otherwise,
// put that card into your hand": the search puts the card into the hand and
// a Hammer of Nazahn is then put onto the battlefield from it, Guidelight
// Pathmaker's shape (nothing happens between the two moves). An on-disk key.
var C17ThenPutFoundHammerOntoBattlefield = game.RegisterTail("then/c17-put-found-hammer-of-nazahn-onto-battlefield",
	func(g *game.Game, o game.TailOutcome) error {
		for _, id := range o.IDs {
			c, ok := g.LookupCardForEffect(id)
			if !ok || c.Name != "Hammer of Nazahn" {
				continue
			}
			if _, err := g.PutFromHandOntoBattlefieldForEffect(id, game.HandEntryOptions{Controller: o.Tail.Player}); err != nil {
				return err
			}
		}
		return nil
	})

// c17WasitoraSacrificed is Wasitora's "If the player can't, you create a
// 3/3 black, red, and green Cat Dragon creature token with flying", read off
// the finished sacrifice run (Invoke Despair's shape): nothing sacrificed
// means they couldn't. Carry is the Queen's controller, then the damaged
// player. An on-disk key: never renamed, never reused.
var C17WasitoraSacrificed = game.RegisterPromptRunThen("prompt-run/c17-wasitora-sacrificed", func(g *game.Game, r game.RunDone) error {
	if len(r.Carry) < 2 {
		return nil
	}
	me, victim := r.Carry[0], r.Carry[1]
	if len(r.Sacrifices().By(victim)) > 0 {
		return nil
	}
	return CreateToken{Controller: me, Template: TokenCard("3/3 black and red and green Cat Dragon with flying"), N: 1}.Apply(KeyedContext(g, me, r.Source))
})
