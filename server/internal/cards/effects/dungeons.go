package effects

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dungeons.go is the catalog half of CR 309 (dungeons), ADR 0203
// (#4096). The engine half — the venture marker, the venture keyword
// action, the room choice and the CR 704.5t completion — is
// game/dungeon.go.
//
// A dungeon is a DungeonSpec: its rooms in printed order, the topmost
// first, each with its arrows, its printed text and its room ability.
// registerDungeon files the room graph with the engine
// (game.RegisterDungeon) and the room abilities as ordinary triggered
// abilities under the dungeon's catalog key (game.DungeonKey), into
// `defs` and not into `registry` — a dungeon is a rules object, not a
// catalog card, so the card census never counts it, exactly as the
// Ring emblem is filed.
//
// Each room ability is "When you move your venture marker into this
// room, [effect]" (CR 309.4c): a trigger watching EventVentureRoom for
// its own room key. It goes on the stack like any other trigger, is
// controlled by the dungeon's owner, and a room that targets declares
// its clause on the row, so CR 603.3d (no legal target: the ability is
// removed) and CR 608.2b (re-checked at resolution) come for free.
//
// The Undercity is filed here. The three Adventures in the Forgotten
// Realms dungeons (Lost Mine of Phandelver, Dungeon of the Mad Mage,
// Tomb of Annihilation) are rows of the same table, filed in
// dungeons_forgotten_realms.go (#4100): the dungeons a plain "venture
// into the dungeon" chooses among.

// DungeonRoomSpec is one room of a dungeon.
type DungeonRoomSpec struct {
	// Key names the room within its dungeon. An on-disk identity.
	Key string
	// Name is the printed room name.
	Name string
	// Text is the room ability's printed effect, verbatim.
	Text string
	// LeadsTo are the rooms its arrows point to, by key.
	LeadsTo []string
	// Targets is the room ability's target clause, nil when it has
	// none.
	Targets *game.TargetSpec
	// Effect is the room ability's effect.
	Effect Effect
}

// DungeonSpec is one dungeon card.
type DungeonSpec struct {
	Key, Name string
	// VentureIntoOnly is the Undercity's "You can't enter this dungeon
	// unless you 'venture into Undercity.'"
	VentureIntoOnly bool
	Rooms           []DungeonRoomSpec
}

// roomAbilityLabel is a room ability's stack label, and its row's
// catalog name (TriggeredAbility.Key): "Undercity — Forge". An on-disk
// identity, so never reworded.
func roomAbilityLabel(dungeon, room string) string { return dungeon + " — " + room }

// registerDungeon files a dungeon: its room graph with the engine and
// its room abilities in the catalog.
func registerDungeon(d DungeonSpec) {
	key := game.DungeonKey(d.Key)
	if _, ok := defs[key]; ok {
		panic(fmt.Sprintf("effects.registerDungeon: duplicate dungeon %q", d.Key))
	}
	rooms := make([]game.DungeonRoom, len(d.Rooms))
	triggers := make([]game.TriggeredAbility, len(d.Rooms))
	for i, r := range d.Rooms {
		if r.Effect == nil || strings.TrimSpace(r.Text) == "" {
			panic(fmt.Sprintf("effects.registerDungeon: %s room %q has no effect or no text", d.Name, r.Key))
		}
		rooms[i] = game.DungeonRoom{Key: r.Key, Name: r.Name, Text: r.Text, LeadsTo: r.LeadsTo}
		t := On(game.EventVentureRoom, enteredRoom(d.Key, r.Key), roomAbilityLabel(d.Name, r.Name), r.Effect)
		t.Targets = r.Targets
		triggers[i] = t
	}
	game.RegisterDungeon(game.DungeonDef{Key: d.Key, Name: d.Name, Rooms: rooms, VentureIntoOnly: d.VentureIntoOnly})
	checkTriggerZones(d.Name, "room ability", triggers)
	fileDef(key, &game.CardDef{Triggered: triggers})
}

// enteredRoom is a room ability's trigger condition: the venture marker
// of this dungeon object moved into this room. `source` is the dungeon
// card (game.dungeonObject), so the comparison is against its own
// object, not against any dungeon with a room of that name.
func enteredRoom(dungeon, room string) When {
	label := game.VentureRoomLabel(dungeon, room)
	return func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return source != nil && ev.Kind == game.EventVentureRoom && ev.CardID == source.InstanceID && ev.Label == label
	}
}

// The Undercity — the dungeon the initiative ventures into (CR 726.2),
// printed on the back of the initiative's reminder card:
//
//	You can't enter this dungeon unless you "venture into Undercity."
//	Secret Entrance — Search your library for a basic land card,
//	  reveal it, put it into your hand, then shuffle. (Leads to:
//	  Forge, Lost Well)
//	Forge — Put two +1/+1 counters on target creature. (Leads to:
//	  Trap!, Arena)
//	Lost Well — Scry 2. (Leads to: Arena, Stash)
//	Trap! — Target player loses 5 life. (Leads to: Archives)
//	Arena — Goad target creature. (Leads to: Archives, Catacombs)
//	Stash — Create a Treasure token. (Leads to: Catacombs)
//	Archives — Draw a card. (Leads to: Throne of the Dead Three)
//	Catacombs — Create a 4/1 black Skeleton creature token with
//	  menace. (Leads to: Throne of the Dead Three)
//	Throne of the Dead Three — Reveal the top ten cards of your
//	  library. Put a creature card from among them onto the
//	  battlefield with three +1/+1 counters on it. It gains hexproof
//	  until your next turn. Then shuffle.
//
// (Scryfall oracle text of "Undercity // The Initiative",
// 36b61021-72f0-4c22-a41d-9b2f093d7ca8, from the 2026-10-05 dump.)
func init() {
	registerDungeon(DungeonSpec{
		Key:             game.UndercityKey,
		Name:            "Undercity",
		VentureIntoOnly: true,
		Rooms: []DungeonRoomSpec{
			{
				Key: "secret-entrance", Name: "Secret Entrance",
				Text:    "Search your library for a basic land card, reveal it, put it into your hand, then shuffle.",
				LeadsTo: []string{"forge", "lost-well"},
				Effect:  undercitySecretEntrance,
			},
			{
				Key: "forge", Name: "Forge",
				Text:    "Put two +1/+1 counters on target creature.",
				LeadsTo: []string{"trap", "arena"},
				Targets: TargetCreature("target creature"),
				Effect:  undercityForge,
			},
			{
				Key: "lost-well", Name: "Lost Well",
				Text:    "Scry 2.",
				LeadsTo: []string{"arena", "stash"},
				Effect:  undercityLostWell,
			},
			{
				Key: "trap", Name: "Trap!",
				Text:    "Target player loses 5 life.",
				LeadsTo: []string{"archives"},
				Targets: TargetPlayer("target player"),
				Effect:  undercityTrap,
			},
			{
				Key: "arena", Name: "Arena",
				Text:    "Goad target creature.",
				LeadsTo: []string{"archives", "catacombs"},
				Targets: TargetCreature("target creature"),
				Effect:  GoadTarget,
			},
			{
				Key: "stash", Name: "Stash",
				Text:    "Create a Treasure token.",
				LeadsTo: []string{"catacombs"},
				Effect:  undercityStash,
			},
			{
				Key: "archives", Name: "Archives",
				Text:    "Draw a card.",
				LeadsTo: []string{"throne"},
				Effect:  undercityArchives,
			},
			{
				Key: "catacombs", Name: "Catacombs",
				Text:    "Create a 4/1 black Skeleton creature token with menace.",
				LeadsTo: []string{"throne"},
				Effect:  undercityCatacombs,
			},
			{
				Key: "throne", Name: "Throne of the Dead Three",
				Text:   "Reveal the top ten cards of your library. Put a creature card from among them onto the battlefield with three +1/+1 counters on it. It gains hexproof until your next turn. Then shuffle.",
				Effect: undercityThrone,
			},
		},
	})
}

// undercitySecretEntrance: search your library for a basic land card,
// reveal it, put it into your hand, then shuffle. A search for a card
// with a stated quality may find nothing (CR 701.23b).
func undercitySecretEntrance(g *game.Game, item *game.StackItem) error {
	return SearchLibrary{
		Player:    item.Controller,
		Predicate: IsBasicLand,
		Dest:      game.ZoneHand,
		Limit:     1,
		Reveal:    true,
		Shuffle:   true,
		Reason:    "Undercity — Secret Entrance: search for a basic land card",
		Source:    item.SourceCardID,
	}.Apply(NewContext(g, item))
}

// undercityForge: put two +1/+1 counters on target creature.
var undercityForge = dungeonCounterOnTarget(2)

// undercityLostWell: scry 2.
func undercityLostWell(g *game.Game, item *game.StackItem) error {
	return Scry{Player: item.Controller, N: 2}.Apply(NewContext(g, item))
}

// undercityTrap: target player loses 5 life.
func undercityTrap(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			return GainLife{Player: t.ID, Amount: -5}.Apply(ctx)
		}
	}
	return nil
}

// undercityStash: create a Treasure token.
func undercityStash(g *game.Game, item *game.StackItem) error {
	return CreateToken{Controller: item.Controller, Template: TreasureToken(), N: 1}.Apply(NewContext(g, item))
}

// undercityArchives: draw a card.
func undercityArchives(g *game.Game, item *game.StackItem) error {
	return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
}

// undercityCatacombs: create a 4/1 black Skeleton creature token with
// menace.
func undercityCatacombs(g *game.Game, item *game.StackItem) error {
	return CreateToken{Controller: item.Controller, Template: TokenCard("4/1 black Skeleton with menace"), N: 1}.Apply(NewContext(g, item))
}

// undercityThrone: reveal the top ten cards of your library, put a
// creature card from among them onto the battlefield with three +1/+1
// counters on it, it gains hexproof until your next turn, then shuffle.
//
// "A creature card" is not "you may": with a creature among the ten,
// one is put (asked when there are two or more). The counters ride the
// entry (CR 614.1c), so a Hardened Scales sees them. The revealed cards
// are public (every seat becomes a knower); the shuffle that follows
// takes that knowledge away again.
func undercityThrone(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	you := item.Controller
	var revealed []uuid.UUID
	if err := (RevealTopOfLibrary{
		Player:   you,
		N:        10,
		Reason:   "Undercity — Throne of the Dead Three",
		Revealed: &revealed,
	}).Apply(ctx); err != nil {
		return err
	}
	return PutFromLibraryOntoBattlefield{
		Player:   you,
		Cards:    revealed,
		Match:    Creature(),
		Max:      1,
		Counters: map[string]int{"+1/+1": 3},
		Label:    "Undercity — Throne of the Dead Three: put a creature card from among them onto the battlefield",
		Then: func(g *game.Game, res PutFromLibraryResult) error {
			c := NewContext(g, item)
			for _, id := range res.Entered {
				if err := (ScopedEffectFor{
					Target:   id,
					Mods:     []game.Mod{game.AddKeywordsMod("hexproof")},
					Duration: DurationUntilYourNextTurn(c, you),
					Label:    "Undercity — Throne of the Dead Three: hexproof until your next turn",
				}).Apply(c); err != nil {
					return err
				}
			}
			return ShuffleLibrary{Player: you}.Apply(c)
		},
	}.Apply(ctx)
}
